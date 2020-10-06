package main

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

//onExpiration - store info about keys with set expiration date
type onExpiration struct {
	Mut               *sync.Mutex
	ByKeyMap          map[string]*ExpTime   //key - key from main database, value - expiration time
	ByTimeMap         map[*ExpTime][]string //key - expiration date, value - slice of keys from main database
	TimePriorityQueue *PriorityQueue        //priority queue
}

//ExpTime store expiration date and it's index in priority queue
type ExpTime struct {
	value time.Time
	index int
}

func (ExpKeys *onExpiration) String() string {
	ExpKeys.Mut.Lock()
	res := "On expiration:\nby key map:\n"
	for k, v := range ExpKeys.ByKeyMap {
		res = fmt.Sprint(res, "{", k, " : ", v.value, "}\n")
	}

	res = fmt.Sprint(res, "by time map:\n")
	for k, v := range ExpKeys.ByTimeMap {
		res = fmt.Sprint(res, "{", k.value, " : ", v, "}\n")
	}

	res = fmt.Sprint(res, "time priority queue:\n")
	res = fmt.Sprint(res, ExpKeys.TimePriorityQueue)
	ExpKeys.Mut.Unlock()
	return res
}

func newOnExpiration() *onExpiration {
	var TimePriorityQueue = make(PriorityQueue, 0)
	heap.Init(&TimePriorityQueue)
	return &onExpiration{&sync.Mutex{}, make(map[string]*ExpTime), make(map[*ExpTime][]string), &TimePriorityQueue}
}

//addExpirationForKey - add exparaiontion for key
func (ExpKeys *onExpiration) addExpirationForKey(key string, expTime time.Time) {
	timeItem := &ExpTime{value: expTime}

	ExpKeys.Mut.Lock()

	_, ok := ExpKeys.ByKeyMap[key]

	if ok {
		removeExpiredKey(key, ExpKeys)
	}

	ExpKeys.ByKeyMap[key] = timeItem
	ExpKeys.ByTimeMap[timeItem] = append(ExpKeys.ByTimeMap[timeItem], key)
	heap.Push(ExpKeys.TimePriorityQueue, timeItem)
	ExpKeys.Mut.Unlock()
}

//removeExpirationFromKey - remove expiration date for key
func (ExpKeys *onExpiration) removeExpirationFromKey(key string) {
	ExpKeys.Mut.Lock()
	removeExpiredKey(key, ExpKeys)
	ExpKeys.Mut.Unlock()
}

//getExpiredKeys - find all keys, that expired, delete them from *onExpiration base and return it as slice of string
func (ExpKeys *onExpiration) getExpiredKeys(tillTime time.Time) []string {
	ExpKeys.Mut.Lock()

	keysToReturn := make([]string, 0)

	for i := 0; i < ExpKeys.TimePriorityQueue.Len(); i++ {
		timeItem := ExpKeys.TimePriorityQueue.Peek()

		if timeItem.value.Before(tillTime) {
			keysToReturn = append(keysToReturn, ExpKeys.ByTimeMap[heap.Pop(ExpKeys.TimePriorityQueue).(*ExpTime)]...)
			delete(ExpKeys.ByTimeMap, timeItem)
			continue
		}
		break
	}

	for _, key := range keysToReturn {
		delete(ExpKeys.ByKeyMap, key)
	}

	ExpKeys.Mut.Unlock()
	return keysToReturn
}

//removeExpiredKey - remove key-element from *onExpiration base
func removeExpiredKey(key string, ExpKeys *onExpiration) {
	timeItem, ok := ExpKeys.ByKeyMap[key]

	if ok {
		delete(ExpKeys.ByKeyMap, key)

		keySlice := ExpKeys.ByTimeMap[timeItem]

		for i, k := range keySlice {
			if k == key {
				keySlice = removeByIndexFromSliceOfString(keySlice, i)
				break
			}
		}
		if len(keySlice) == 0 {
			delete(ExpKeys.ByTimeMap, timeItem)
			heap.Remove(ExpKeys.TimePriorityQueue, timeItem.index)
			return
		}
		ExpKeys.ByTimeMap[timeItem] = keySlice
	}
}

func removeByIndexFromSliceOfString(slice []string, index int) []string {
	length := len(slice)
	if length == 1 {
		return []string{}
	}
	last := slice[length-1]
	slice[index] = last
	return slice[0 : length-2]
}

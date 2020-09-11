package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

//onExparation - store info about keys with seted expiration date
type onExparation struct {
	Mut       *sync.Mutex
	ByKeyMap  map[string]time.Time   //key - key from main database, value - exparation time
	ByTimeMap map[time.Time][]string //key - exparation date, value - slice of keys from main database
	TimeSlice []time.Time            //sorted slice of exparation date
}

func (ExpKeys *onExparation) String() string {
	res := "On exparation:\nby key map:\n"
	for v, k := range ExpKeys.ByKeyMap {
		res = fmt.Sprint(res, "{", v, " : ", k, "}\n")
	}

	res = fmt.Sprint(res, "by time map:\n")
	for v, k := range ExpKeys.ByTimeMap {
		res = fmt.Sprint(res, "{", v, " : ", k, "}\n")
	}

	res = fmt.Sprint(res, "sorted by time slice:\n")
	for _, v := range ExpKeys.TimeSlice {
		res = fmt.Sprint(res, v, ", ")
	}
	return res
}

func newOnExparation() *onExparation {
	return &onExparation{&sync.Mutex{}, make(map[string]time.Time), make(map[time.Time][]string), make([]time.Time, 0)}
}

//addExparationForKey - add exparaiontion for key
func (ExpKeys *onExparation) addExparationForKey(key string, expTime time.Time) {
	ExpKeys.Mut.Lock()

	defer ExpKeys.Mut.Unlock()

	ExpKeys.TimeSlice = append(ExpKeys.TimeSlice, expTime)
	sortKeysToExpire(ExpKeys)

	_, ok := ExpKeys.ByKeyMap[key]

	if ok {
		removeExpiredKey(key, ExpKeys)

		ExpKeys.ByKeyMap[key] = expTime
		ExpKeys.ByTimeMap[expTime] = append(ExpKeys.ByTimeMap[expTime], key)
		return
	}

	ExpKeys.ByKeyMap[key] = expTime
	ExpKeys.ByTimeMap[expTime] = append(ExpKeys.ByTimeMap[expTime], key)
}

//removeExparationFromKey - remove exparation date for key
func (ExpKeys *onExparation) removeExparationFromKey(key string) {
	ExpKeys.Mut.Lock()
	removeExpiredKey(key, ExpKeys)
	ExpKeys.Mut.Unlock()
}

//getExpiredKeys - find all keys, that expired, delete them from *onExparation base and return it as slice of string
func (ExpKeys *onExparation) getExpiredKeys(tillTime time.Time) []string {
	ExpKeys.Mut.Lock()

	keysToReturn := make([]string, 0)
	indexToDeleteBefore := 0

	for _, timeValue := range ExpKeys.TimeSlice {

		if timeValue.Before(tillTime) {

			indexToDeleteBefore++
			keysToReturn = append(keysToReturn, ExpKeys.ByTimeMap[timeValue]...)
			delete(ExpKeys.ByTimeMap, timeValue)
			continue
		}
		break
	}

	ExpKeys.TimeSlice = ExpKeys.TimeSlice[indexToDeleteBefore:]

	for _, key := range keysToReturn {
		delete(ExpKeys.ByKeyMap, key)
	}

	ExpKeys.Mut.Unlock()
	return keysToReturn
}

//sortKeysToExpire - sort *onExparation TimeSlice
func sortKeysToExpire(ExpKeys *onExparation) {
	sort.Slice(ExpKeys.TimeSlice, func(i, j int) bool { return ExpKeys.TimeSlice[i].Before(ExpKeys.TimeSlice[j]) })
}

//removeExpiredKey - remove key-element from *onExparation base
func removeExpiredKey(key string, ExpKeys *onExparation) {
	expTime, ok := ExpKeys.ByKeyMap[key]

	if ok {
		delete(ExpKeys.ByKeyMap, key)
		removeFromTimeSlice(expTime, ExpKeys)

		keySlice := ExpKeys.ByTimeMap[expTime]

		for i, k := range keySlice {
			if k == key {
				keySlice = removeByIndexFromSliceOfString(keySlice, i)
				break
			}
		}
		if len(keySlice) == 0 {
			delete(ExpKeys.ByTimeMap, expTime)
			return
		}
		ExpKeys.ByTimeMap[expTime] = keySlice
	}
}

//removeFromTimeSlice - binary search element by value in array. Delete when find.
func removeFromTimeSlice(expTime time.Time, ExpKeys *onExparation) {
	start := 0
	end := len(ExpKeys.TimeSlice) - 1
	index := end / 2
	for expTime != ExpKeys.TimeSlice[index] {
		if expTime.Before(ExpKeys.TimeSlice[index]) {
			end = index - 1
			index = (start + end) / 2
			continue
		}
		start = index + 1
		index = (start + end) / 2
	}
	ExpKeys.TimeSlice = append(ExpKeys.TimeSlice[0:index], ExpKeys.TimeSlice[index+1:]...)
}

func removeByIndexFromSliceOfString(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}

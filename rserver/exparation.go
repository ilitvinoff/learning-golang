package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type onExparation struct {
	Mut       *sync.Mutex
	ByKeyMap  map[string]time.Time
	ByTimeMap map[time.Time][]string
	TimeSlice []time.Time
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

func newOnExparationStruct() *onExparation {
	return &onExparation{&sync.Mutex{}, make(map[string]time.Time), make(map[time.Time][]string), make([]time.Time, 0)}
}

func (ExpKeys *onExparation) addExparationForKey(key string, expTime time.Time) {
	ExpKeys.Mut.Lock()

	defer ExpKeys.Mut.Unlock()

	ExpKeys.TimeSlice = append(ExpKeys.TimeSlice, expTime)
	sortKeysToExpire(ExpKeys)

	_, ok := ExpKeys.ByKeyMap[key]

	if ok {
		removeExpiredKey(key, ExpKeys)
		ExpKeys.ByKeyMap[key] = expTime

		_, ok = ExpKeys.ByTimeMap[expTime]

		if ok {
			ExpKeys.ByTimeMap[expTime] = append(ExpKeys.ByTimeMap[expTime], key)
			return
		}
		ExpKeys.ByTimeMap[expTime] = append(ExpKeys.ByTimeMap[expTime], key)
		return
	}

	ExpKeys.ByKeyMap[key] = expTime
	_, ok = ExpKeys.ByTimeMap[expTime]

	if ok {
		ExpKeys.ByTimeMap[expTime] = append(ExpKeys.ByTimeMap[expTime], key)
		return
	}

	ExpKeys.ByTimeMap[expTime] = []string{key}
}

func (ExpKeys *onExparation) removeExparationFromKey(key string) {
	ExpKeys.Mut.Lock()
	removeExpiredKey(key, ExpKeys)
	ExpKeys.Mut.Unlock()
}

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

func sortKeysToExpire(ExpKeys *onExparation) {
	sort.Slice(ExpKeys.TimeSlice, func(i, j int) bool { return ExpKeys.TimeSlice[i].Before(ExpKeys.TimeSlice[j]) })
}

func removeExpiredKey(key string, ExpKeys *onExparation) {
	expTime, ok := ExpKeys.ByKeyMap[key]

	if ok {
		delete(ExpKeys.ByKeyMap, key)

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

func removeByIndexFromSliceOfString(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}

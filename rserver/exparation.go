package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type onExparation struct {
	*sync.Mutex
	byKeyMap  map[string]time.Time
	byTimeMap map[time.Time][]string
	timeSlice []time.Time
}

func (expKeys *onExparation) String() string {
	return fmt.Sprintf("bykeymap: %v\nbytimemap: %v\ntimeslice: %v", expKeys.byKeyMap, expKeys.byTimeMap, expKeys.timeSlice)
}

func newOnExparationStruct() *onExparation {
	return &onExparation{&sync.Mutex{}, make(map[string]time.Time), make(map[time.Time][]string), make([]time.Time, 0)}
}

func (expKeys *onExparation) addExparationForKey(key string, expTime time.Time) {
	expKeys.Lock()

	defer expKeys.Unlock()

	expKeys.timeSlice = append(expKeys.timeSlice, expTime)
	sortKeysToExpire(expKeys)

	_, ok := expKeys.byKeyMap[key]

	if ok {
		removeExpiredKey(key, expKeys)
		expKeys.byKeyMap[key] = expTime
		expKeys.byTimeMap[expTime] = append(expKeys.byTimeMap[expTime], key)
		return
	}

	expKeys.byKeyMap[key] = expTime
	_, ok = expKeys.byTimeMap[expTime]

	if ok {
		expKeys.byTimeMap[expTime] = append(expKeys.byTimeMap[expTime], key)
		return
	}

	expKeys.byTimeMap[expTime] = []string{key}
}

func (expKeys *onExparation) removeExparationFromKey(key string) {
	expKeys.Lock()
	removeExpiredKey(key, expKeys)
	expKeys.Unlock()
}

func (expKeys *onExparation) getExpiredKeys(tillTime time.Time) []string {
	expKeys.Lock()

	keysToReturn := make([]string, 0)
	indexToDeleteFromTimeSlice := 0

	for _, timeValue := range expKeys.timeSlice {

		if timeValue.Before(tillTime) {

			indexToDeleteFromTimeSlice++
			keysToReturn = append(keysToReturn, expKeys.byTimeMap[timeValue]...)
			delete(expKeys.byTimeMap, timeValue)
			continue
		}
		break
	}

	expKeys.timeSlice = expKeys.timeSlice[indexToDeleteFromTimeSlice:]

	for _, key := range keysToReturn {
		delete(expKeys.byKeyMap, key)
	}

	expKeys.Unlock()
	return keysToReturn
}

func sortKeysToExpire(expKeys *onExparation) {
	sort.Slice(expKeys.timeSlice, func(i, j int) bool { return expKeys.timeSlice[i].Before(expKeys.timeSlice[j]) })
}

func removeExpiredKey(key string, expKeys *onExparation) {
	expTime, ok := expKeys.byKeyMap[key]

	if ok {
		delete(expKeys.byKeyMap, key)

		keySlice := expKeys.byTimeMap[expTime]

		for i, k := range keySlice {
			if k == key {
				keySlice = removeByIndexFromSliceOfString(keySlice, i)
				break
			}
		}
		expKeys.byTimeMap[expTime] = keySlice
	}
}

func removeByIndexFromSliceOfString(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}

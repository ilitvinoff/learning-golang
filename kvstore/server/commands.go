package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	autosave = "autosave"
)

/*Commands Map - includes a list of custom commands for interacting with the database*/
var commands = map[string]func(*KVCache, *command) (string, error){
	//set - set's to KVCache.DataStore {key:value} pair. If key allready exists,
	//set Value.ExpireIsSet to false.
	//Return:
	//"Ok",nil - if successful,
	//"", error - if unsuccessful
	"set": func(KVCache *KVCache, cmd *command) (string, error) {
		err := validateArgsCount(cmd, 2)
		if err != nil {
			return "", err
		}

		KVCache.Mut.Lock()
		KVCache.DataStore[cmd.args[0]] = newValue(cmd.args[1], false)
		KVCache.ExpKeys.removeExpirationFromKey(cmd.args[0])
		KVCache.Mut.Unlock()

		return "true", nil
	},

	//get - return the value corresponding to the key from KVCache.DataStore,
	//Return:
	//Value.Value, nil - if successful,
	//"", error - if unsuccessful
	"get": func(KVCache *KVCache, cmd *command) (string, error) {
		err := validateArgsCount(cmd, 1)
		if err != nil {
			return "", err
		}

		KVCache.Mut.RLock()
		defer KVCache.Mut.RUnlock()

		result, ok := KVCache.DataStore[cmd.args[0]]

		if ok {
			return result.Value, nil
		}

		return "", fmt.Errorf("ERR: NO SUCH ELEMENT: key = %s;", cmd.args[0])
	},

	//getset - set to key new value and return the old one.
	//If there was no such key in database, add new pair{key:value} and return "",error.
	"getset": func(KVCache *KVCache, cmd *command) (string, error) {
		err := validateArgsCount(cmd, 2)
		if err != nil {
			return "", err
		}

		KVCache.Mut.Lock()
		defer KVCache.Mut.Unlock()

		Value, ok := KVCache.DataStore[cmd.args[0]]
		KVCache.DataStore[cmd.args[0]] = newValue(cmd.args[1], false)
		KVCache.ExpKeys.removeExpirationFromKey(cmd.args[0])

		if ok {
			return Value.Value, nil
		}

		return "", fmt.Errorf("ERR: No available Value for key: %s, is present. %s;", cmd.args[0], cmd)
	},

	//exists - check if key is presented in database.
	//return "true" - if is, "false" - if not, or error if it was occured.
	"exist": func(KVCache *KVCache, cmd *command) (string, error) {
		err := validateArgsCount(cmd, 1)
		if err != nil {
			return "", err
		}

		KVCache.Mut.RLock()
		_, ok := KVCache.DataStore[cmd.args[0]]
		KVCache.Mut.RUnlock()
		return strconv.FormatBool(ok), nil
	},

	//deleteElement - delete element from database by key.
	//return amount of deleted keys, or error if it was occured.
	"del": func(KVCache *KVCache, cmd *command) (string, error) {
		if len(cmd.args) < 1 {
			return "", fmt.Errorf("ERR: Not enough arguments. Command name: %s;", cmd.name)
		}

		counter := 0
		KVCache.Mut.Lock()
		for _, key := range cmd.args {
			_, ok := KVCache.DataStore[key]
			if ok {
				delete(KVCache.DataStore, key)
				KVCache.ExpKeys.removeExpirationFromKey(key)
				counter++
			}
		}
		KVCache.Mut.Unlock()

		return strconv.Itoa(counter), nil
	},

	//expire - set expiration date to element of database bu key.
	//This element will be deleted when expired.
	//Return "true"/"false",nil - when set/not set.
	//Return "",error - if error was occured.
	"ex": func(KVCache *KVCache, cmd *command) (string, error) {
		err := validateArgsCount(cmd, 2)
		if err != nil {
			return "", err
		}

		expTime, err := validateTimeDuration(cmd.args[1])
		if err != nil {
			return "", err
		}

		KVCache.Mut.Lock()
		defer KVCache.Mut.Unlock()
		Value, ok := KVCache.DataStore[cmd.args[0]]

		if ok {
			Value.ExpireIsSet = true
			KVCache.ExpKeys.addExpirationForKey(cmd.args[0], time.Now().Add(expTime).Truncate(time.Second))
			return "true", nil
		}

		return "false", nil
	},

	//saveData - save databse as json formated string to file.
	//Return "true",nil - if successful.
	//Return "",error - if it was occured.
	"save": func(KVCache *KVCache, cmd *command) (string, error) {
		err := validateArgsCount(cmd, 1)
		if err != nil {
			return "", err
		}

		file, err := os.Create(cmd.args[0])
		if err != nil {
			return "", fmt.Errorf("ERR: CAN'T CREATE FILE: %s. ERR: %s;", cmd.args[0], err)
		}
		defer file.Close()

		KVCache.Mut.RLock()
		data, err := json.Marshal(&KVCache)
		KVCache.Mut.RUnlock()

		if err != nil {
			return "", fmt.Errorf("ERR: ENCODE ERR: %s;", err)
		}

		_, err = file.Write(data)
		if err != nil {
			return "", fmt.Errorf("ERR: WRITING TO FILE ERR: %s;", err)
		}

		log.Println("MARSHAL AND SAVE: \n" + string(data))
		return "true", nil
	},

	"autosave": func(KVCache *KVCache, cmd *command) (string, error) {
		err := validateArgsCount(cmd, 1)
		if err != nil {
			return "", err
		}

		interval, err := validateTimeDuration(cmd.args[0])
		if err != nil {
			return "", err
		}

		if KVCache.autosaveIndicator {
			KVCache.autoSaveTimeDurationChan <- interval

			if interval == time.Duration(0) {
				KVCache.Mut.Lock()
				KVCache.autosaveIndicator = false
				KVCache.Mut.Unlock()
				return fmt.Sprint("Autosave is off."), nil
			}

			return fmt.Sprintf("Interval changed to - %v\n", interval), nil
		}

		KVCache.autoSaveTimeDurationChan <- interval

		go func() {
			KVCache.Mut.Lock()
			KVCache.autosaveIndicator = true
			KVCache.Mut.Unlock()

			for {
				el := <-KVCache.autoSaveTimeDurationChan

				if el == time.Duration(0) {
					KVCache.Mut.Lock()
					KVCache.autosaveIndicator = false
					KVCache.Mut.Unlock()
					return
				}

				file, err := os.Create(autosave)
				if err != nil {
					fmt.Printf("ERR:AUTOSAVING. CAN'T CREATE FILE: %s. ERR: %s;\n", cmd.args[0], err)
				}

				KVCache.Mut.RLock()
				data, err := json.Marshal(&KVCache)
				KVCache.Mut.RUnlock()

				if err != nil {
					fmt.Printf("ERR:AUTOSAVING. ENCODE ERR: %s;\n", err)
				}

				_, err = file.Write(data)
				if err != nil {
					fmt.Printf("ERR:AUTOSAVING. WRITING TO FILE ERR: %s;\n", err)
				}
				file.Close()

				if len(KVCache.autoSaveTimeDurationChan) > 0 {
					el = <-KVCache.autoSaveTimeDurationChan
				}

				KVCache.autoSaveTimeDurationChan <- time.Duration(el)
				time.Sleep(el)
			}
		}()

		if interval == time.Duration(0) {
			KVCache.Mut.Lock()
			KVCache.autosaveIndicator = false
			KVCache.Mut.Unlock()
			return fmt.Sprint("Autosave is off."), nil
		}

		return fmt.Sprintf("Autosave is on. Interval - %v", interval), nil
	},

	//restoreData - restore database with help of json formated string(from file).
	//Return "true",nil - if successful.
	//Return "",error - if it was occured.
	"restore": func(KVCache *KVCache, cmd *command) (string, error) {
		err := validateArgsCount(cmd, 1)
		if err != nil {
			return "", err
		}

		file, err := os.Open(cmd.args[0])
		if err != nil {
			return "", fmt.Errorf("ERR: CAN'T READ DATA FROM FILE: %s. ERR: %s;", cmd.args[0], err)
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)
		if err != nil {
			return "", fmt.Errorf("ERR: READING FROM FILE: %s. ERR: %s;", cmd.args[0], err)
		}

		newCache := newKVCache()
		err = json.Unmarshal(data, &newCache)
		if err != nil {
			return "", fmt.Errorf("ERR: UNMARSHAL ERR: %s;", err)
		}

		KVCache.Mut.Lock()
		KVCache.DataStore = newCache.DataStore
		KVCache.ExpKeys = newCache.ExpKeys
		for k := range KVCache.ExpKeys.ByTimeMap {
			KVCache.ExpKeys.TimePriorityQueue.Push(&ExpTime{value: k})
		}
		KVCache.Mut.Unlock()

		log.Println("UNMARSHALED AND RESTOREd: \n" + string(data))
		return "true", nil
	},

	//showAll - return all information from database as string.
	"showall": func(KVCache *KVCache, cmd *command) (string, error) {
		return KVCache.String(), nil
	},
}

// expirationWatcher - Checks if any keys have expired at the moment, and removes them, if any.
func (KVCache *KVCache) expirationWatcher() {
	for {

		go func() {
			keysToDelete := KVCache.ExpKeys.getExpiredKeys(time.Now())

			KVCache.Mut.Lock()
			for _, key := range keysToDelete {

				if _, ok := KVCache.DataStore[key]; ok && KVCache.DataStore[key].ExpireIsSet {
					delete(KVCache.DataStore, key)
				}
			}
			KVCache.Mut.Unlock()
		}()

		time.Sleep(time.Second)
	}
}

//Command - describes user command.
type command struct {
	name string
	args []string
}

//KVCache - main struct to store all possible information about our database.
type KVCache struct {
	Mut                      *sync.RWMutex
	DataStore                map[string]*Value //main database
	ExpKeys                  *onExpiration     //information about keys with a set expiration date
	autoSaveTimeDurationChan chan time.Duration
	autosaveIndicator        bool
}

//Value - describes value set to key in Rcache.DataStore
type Value struct {
	Mut         *sync.Mutex
	Value       string
	ExpireIsSet bool
}

//newRcache - creates and returns *Rcache instance
func newKVCache() *KVCache {
	return &KVCache{&sync.RWMutex{}, make(map[string]*Value), newOnExpiration(),
		make(chan time.Duration, 1), false}
}

//newValue - creates and returns *Value instance
func newValue(s string, ExpireIsSet bool) *Value {
	return &Value{&sync.Mutex{}, s, ExpireIsSet}
}

func (KVCache *KVCache) String() string {
	KVCache.Mut.RLock()
	defer KVCache.Mut.RUnlock()

	res := fmt.Sprint("=================================")
	res = "\nData store:\n"
	for k, v := range KVCache.DataStore {
		res = fmt.Sprint(res, "{", k, " : ", v, "}\n")
	}
	res = fmt.Sprint(res, KVCache.ExpKeys)
	res = fmt.Sprint(res, "=================================")

	return res
}

func (v *Value) String() string {
	v.Mut.Lock()
	defer v.Mut.Unlock()
	return fmt.Sprintf("<value: %s | expire_is_set: %v>", v.Value, v.ExpireIsSet)
}

func (cmd *command) String() string {
	return fmt.Sprintf("Command name: %s, args: %s;", cmd.name, cmd.args)
}

//validateArgsCount validate if arg's amount is correct
func validateArgsCount(cmd *command, n int) error {
	if len(cmd.args) != n {
		return fmt.Errorf("ERR: Invalid number of arguments. Should be %d, has: %d. %s;", n, len(cmd.args), cmd)
	}
	return nil
}

//validateTimeDuration validate if value set as expiration date for key is correct
func validateTimeDuration(Value string) (time.Duration, error) {
	n, err := strconv.Atoi(Value)
	if err != nil {
		return time.Duration(0), err
	}

	if n < 0 {
		return time.Duration(0), fmt.Errorf("ERR: USE POSITIVE VALUE TO SET TIME. NEITHER: %d", n)
	}

	return time.Duration(n) * time.Second, nil
}

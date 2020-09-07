package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

var commands = map[string]func(*rcashe, *command) (string, error){
	"set":    set,
	"get":    get,
	"getset": getset,
	"exists": exists,
	"del":    deleteElement,
	"ex":     expire,
}

const (
	//OK return as result of command when succesfull
	OK               = "Ok"
	ifErrMessage     = "ERR: Invalid number of arguments. Should be %d, has: %d. %s;"
	getsetErrMessage = "ERR: No available value for key: %s, is present. %s;"
)

type command struct {
	name string
	args []string
}

type rcashe struct {
	*sync.RWMutex
	datastore map[string]*value
	expKeys   *onExparation
}

type value struct {
	*sync.Mutex
	str           string
	expireIsSeted bool
}

func newRCashe() *rcashe {
	return &rcashe{&sync.RWMutex{}, make(map[string]*value), newOnExparationStruct()}
}

func newValue(s string, expireIsSeted bool) *value {
	return &value{&sync.Mutex{}, s, expireIsSeted}
}

func (cmd *command) String() string {
	return fmt.Sprintf("Command name: %s, args: %s", cmd.name, cmd.args)
}

func validateArgsAmount(cmd *command, n int) error {
	if len(cmd.args) != n {
		return fmt.Errorf(ifErrMessage, n, len(cmd.args), cmd)
	}
	return nil
}

func validateExparationValue(value string) (time.Duration, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return time.Second, err
	}

	if n < 0 {
		return time.Second, fmt.Errorf("ERR: Use positive value to set exparation. Neither: %d", n)
	}

	return time.Duration(n) * time.Second, nil
}

func set(rcashe *rcashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 2)
	if err != nil {
		return "", err
	}

	rcashe.RLock()
	value, ok := rcashe.datastore[cmd.args[0]]
	rcashe.RUnlock()

	if ok {
		value.Lock()
		value.str = cmd.args[1]
		value.expireIsSeted = false
		value.Unlock()

		return OK, nil
	}

	rcashe.datastore[cmd.args[0]] = newValue(cmd.args[1], false)

	return OK, nil
}

func get(rcashe *rcashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 1)
	if err != nil {
		return "", err
	}

	rcashe.RLock()
	defer rcashe.RUnlock()

	result, ok := rcashe.datastore[cmd.args[0]]

	if ok {
		return result.str, nil
	}

	return "", fmt.Errorf("ERR: No such element: key = %s;", cmd.args[0])
}

func getset(rcashe *rcashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 2)
	if err != nil {
		return "", err
	}

	rcashe.RLock()
	value, ok := rcashe.datastore[cmd.args[0]]
	rcashe.RUnlock()

	if ok {
		defer set(rcashe, cmd)
		return value.str, nil
	}

	set(rcashe, cmd)
	return "", fmt.Errorf(getsetErrMessage, cmd.args[0], cmd)
}

func exists(rcashe *rcashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 1)
	if err != nil {
		return "", err
	}

	rcashe.RLock()
	_, ok := rcashe.datastore[cmd.args[0]]
	rcashe.RUnlock()
	return strconv.FormatBool(ok), nil
}

func deleteElement(rcashe *rcashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 1)
	if err != nil {
		return "", err
	}

	counter := 0
	for _, key := range cmd.args {
		rcashe.Lock()
		_, ok := rcashe.datastore[key]
		if ok {
			delete(rcashe.datastore, key)
			counter++
		}
		rcashe.Unlock()
	}

	return strconv.Itoa(counter), nil
}

func expire(rcashe *rcashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 2)
	if err != nil {
		return "", err
	}

	expTime, err := validateExparationValue(cmd.args[1])
	if err != nil {
		return "", err
	}

	rcashe.RLock()
	value, ok := rcashe.datastore[cmd.args[0]]
	rcashe.RUnlock()

	if ok {
		value.Lock()
		value.expireIsSeted = true
		value.Unlock()

		rcashe.expKeys.addExparationForKey(cmd.args[0], time.Now().Add(expTime))

		return "true", nil
	}

	return "false", nil
}

func (rc *rcashe) exparationWatcher() {
	for {

		go func() {
			keysToDelete := rc.expKeys.getExpiredKeys(time.Now())

			rc.Lock()
			for _, key := range keysToDelete {

				if _, ok := rc.datastore[key]; ok && rc.datastore[key].expireIsSeted {
					delete(rc.datastore, key)
				}
			}
			rc.Unlock()
		}()

		time.Sleep(time.Second)
	}
}

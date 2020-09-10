package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"
)

var commands = map[string]func(*RCashe, *command) (string, error){
	"set":         set,
	"get":         get,
	"getset":      getset,
	"exist":       exists,
	"del":         deleteElement,
	"ex":          expire,
	"savedata":    saveData,
	"restoredata": restoreData,
	"showall":     showAll,
}

const (
	//OK return as result of command when succesfull
	OK               = "Ok"
	ifErrMessage     = "ERR: Invalid number of arguments. Should be %d, has: %d. %s;"
	getsetErrMessage = "ERR: No available Value for key: %s, is present. %s;"
)

type command struct {
	name string
	args []string
}

type RCashe struct {
	Mut       *sync.RWMutex
	DataStore map[string]*Value
	ExpKeys   *onExparation
}

type Value struct {
	Mut           *sync.Mutex
	Value         string
	ExpireIsSeted bool
}

func newRCashe() *RCashe {
	return &RCashe{&sync.RWMutex{}, make(map[string]*Value), newOnExparationStruct()}
}

func newValue(s string, ExpireIsSeted bool) *Value {
	return &Value{&sync.Mutex{}, s, ExpireIsSeted}
}

func (RCashe *RCashe) String() string {
	res := "\nData store:\n"
	for k, v := range RCashe.DataStore {
		res = fmt.Sprint(res, "{", k, " : ", v, "}\n")
	}
	res = fmt.Sprint(res, RCashe.ExpKeys)
	return res
}

func (v *Value) String() string {
	return fmt.Sprintf("{value: %s | expire_is_seted: %v}", v.Value, v.ExpireIsSeted)
}

func (cmd *command) String() string {
	return fmt.Sprintf("Command name: %s, args: %s;", cmd.name, cmd.args)
}

func validateArgsAmount(cmd *command, n int) error {
	if len(cmd.args) != n {
		return fmt.Errorf(ifErrMessage, n, len(cmd.args), cmd)
	}
	return nil
}

func validateExparationValue(Value string) (time.Duration, error) {
	n, err := strconv.Atoi(Value)
	if err != nil {
		return time.Second, err
	}

	if n < 0 {
		return time.Second, fmt.Errorf("ERR: USE POSITIVE VALUE TO SET EXPIRATION. NEITHER: %d", n)
	}

	return time.Duration(n) * time.Second, nil
}

func set(RCashe *RCashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 2)
	if err != nil {
		return "", err
	}

	RCashe.Mut.RLock()
	Value, ok := RCashe.DataStore[cmd.args[0]]
	RCashe.Mut.RUnlock()

	if ok {
		Value.Mut.Lock()
		Value.Value = cmd.args[1]
		Value.ExpireIsSeted = false
		Value.Mut.Unlock()

		return OK, nil
	}

	RCashe.DataStore[cmd.args[0]] = newValue(cmd.args[1], false)

	return OK, nil
}

func get(RCashe *RCashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 1)
	if err != nil {
		return "", err
	}

	RCashe.Mut.RLock()
	defer RCashe.Mut.RUnlock()

	result, ok := RCashe.DataStore[cmd.args[0]]

	if ok {
		return result.Value, nil
	}

	return "", fmt.Errorf("ERR: NO SUCH ELEMENT: key = %s;", cmd.args[0])
}

func getset(RCashe *RCashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 2)
	if err != nil {
		return "", err
	}

	RCashe.Mut.RLock()
	Value, ok := RCashe.DataStore[cmd.args[0]]
	RCashe.Mut.RUnlock()

	if ok {
		defer set(RCashe, cmd)
		return Value.Value, nil
	}

	set(RCashe, cmd)
	return "", fmt.Errorf(getsetErrMessage, cmd.args[0], cmd)
}

func exists(RCashe *RCashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 1)
	if err != nil {
		return "", err
	}

	RCashe.Mut.RLock()
	_, ok := RCashe.DataStore[cmd.args[0]]
	RCashe.Mut.RUnlock()
	return strconv.FormatBool(ok), nil
}

func deleteElement(RCashe *RCashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 1)
	if err != nil {
		return "", err
	}

	counter := 0
	for _, key := range cmd.args {
		RCashe.Mut.Lock()
		_, ok := RCashe.DataStore[key]
		if ok {
			delete(RCashe.DataStore, key)
			counter++
		}
		RCashe.Mut.Unlock()
	}

	return strconv.Itoa(counter), nil
}

func expire(RCashe *RCashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 2)
	if err != nil {
		return "", err
	}

	expTime, err := validateExparationValue(cmd.args[1])
	if err != nil {
		return "", err
	}

	RCashe.Mut.RLock()
	Value, ok := RCashe.DataStore[cmd.args[0]]
	RCashe.Mut.RUnlock()

	if ok {
		Value.Mut.Lock()
		Value.ExpireIsSeted = true
		Value.Mut.Unlock()

		RCashe.ExpKeys.addExparationForKey(cmd.args[0], time.Now().Add(expTime))

		return "true", nil
	}

	return "false", nil
}

func (RCashe *RCashe) exparationWatcher() {
	for {

		go func() {
			keysToDelete := RCashe.ExpKeys.getExpiredKeys(time.Now())

			RCashe.Mut.Lock()
			for _, key := range keysToDelete {

				if _, ok := RCashe.DataStore[key]; ok && RCashe.DataStore[key].ExpireIsSeted {
					delete(RCashe.DataStore, key)
				}
			}
			RCashe.Mut.Unlock()
		}()

		time.Sleep(time.Second)
	}
}

func saveData(RCashe *RCashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 1)
	if err != nil {
		return "", err
	}

	file, err := os.Create(cmd.args[0])
	if err != nil {
		return "", fmt.Errorf("ERR: CAN'T CREATE FILE: %s. ERR: %s", cmd.args[0], err)
	}

	RCashe.Mut.Lock()
	slice, err := json.Marshal(&RCashe)
	RCashe.Mut.Unlock()

	fmt.Println("MARSHAL: \n" + string(slice))

	if err != nil {
		return "", fmt.Errorf("ERR: ENCODE ERR: %s", err)
	}

	_, err = file.Write(slice)
	if err != nil {
		return "", fmt.Errorf("ERR: WRITING TO FILE ERR: %s", err)
	}
	file.Close()

	return "saved", nil
}

func showAll(RCashe *RCashe, cmd *command) (string, error) {
	return RCashe.String(), nil
}

func restoreData(RCashe *RCashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 1)
	if err != nil {
		return "", err
	}

	file, err := os.Open("store.gob")
	if err != nil {
		return "", fmt.Errorf("ERR: CAN'T READ DATA FROM FILE: %s", err)
	}

	slice, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("ERR: READING FROM FILE ERR: %s", err)
	}

	file.Close()

	newCashe := newRCashe()
	err = json.Unmarshal(slice, &newCashe)
	if err != nil {
		return "", fmt.Errorf("ERR: UNMARSHAL ERR: %s", err)
	}
	fmt.Println("\nNEWCASHE:\n" + newCashe.String())
	RCashe.Mut.Lock()
	RCashe.DataStore = newCashe.DataStore
	RCashe.ExpKeys = newCashe.ExpKeys
	RCashe.Mut.Unlock()

	return "restored", nil
}

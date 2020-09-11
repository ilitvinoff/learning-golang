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

/*Commands Map - includes a list of custom commands for interacting with the database*/
var commands = map[string]func(*RCashe, *command) (string, error){
	"set":     set,
	"get":     get,
	"getset":  getset,
	"exist":   exists,
	"del":     deleteElement,
	"ex":      expire,
	"save":    saveData,
	"restore": restoreData,
	"showall": showAll,
}

const (
	//OK return as result of command when succesfull
	OK               = "Ok"
	ifErrMessage     = "ERR: Invalid number of arguments. Should be %d, has: %d. %s;"
	getsetErrMessage = "ERR: No available Value for key: %s, is present. %s;"
)

//Command - describes user command.
type command struct {
	name string
	args []string
}

//RCashe - main struct to store all possible information about our database.
type RCashe struct {
	Mut       *sync.RWMutex
	DataStore map[string]*Value //main database
	ExpKeys   *onExparation     //information about keys with a set expiration date
}

//Value - describes value seted to key in Rcashe.DataStore
type Value struct {
	Mut           *sync.Mutex
	Value         string
	ExpireIsSeted bool
}

//newRcashe - creates and returns *Rcashe instance
func newRCashe() *RCashe {
	return &RCashe{&sync.RWMutex{}, make(map[string]*Value), newOnExparation()}
}

//newValue - creates and returns *Value instance
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

//validateArgsAmount validate if arg's amount is correct
func validateArgsAmount(cmd *command, n int) error {
	if len(cmd.args) != n {
		return fmt.Errorf(ifErrMessage, n, len(cmd.args), cmd)
	}
	return nil
}

//validateExparationValue validate if value seted as exparation date for key is correct
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

//set - set's to RCashe.DataStore {key:value} pair. If key allready exists,
//set Value.ExpireIsSeted to false.
//Return:
//"Ok",nil - if successful,
//"", error - if unsuccessful
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

		RCashe.ExpKeys.removeExparationFromKey(cmd.args[0])

		return OK, nil
	}

	RCashe.DataStore[cmd.args[0]] = newValue(cmd.args[1], false)

	return OK, nil
}

//get - return the value corresponding to the key from RCashe.DataStore,
//Return:
//Value.Value, nil - if successful,
//"", error - if unsuccessful
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

//getset - set to key new value and return the old one.
//If there was no such key in database, add new pair{key:value} and return "",error.
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

//exists - check if key is presented in database.
//return "true" - if is, "false" - if not, or error if it was occured.
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

//deleteElement - delete element from database by key.
//return amount of deleted keys, or error if it was occured.
func deleteElement(RCashe *RCashe, cmd *command) (string, error) {
	if len(cmd.args) < 2 {
		return "", fmt.Errorf("ERR: Not enough arguments. Command name: %s", cmd.name)
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

//expire - set exparation date to element of database bu key.
//This element will be deleted when expired.
//Return "true"/"false",nil - when seted/not seted.
//Return "",error - if error was occured.
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

		RCashe.ExpKeys.addExparationForKey(cmd.args[0], time.Now().Add(expTime).Truncate(1*time.Second))

		return "true", nil
	}

	return "false", nil
}

// exparationWatcher - Checks if any keys have expired at the moment, and removes them, if any.
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

//saveData - save databse as json formated string to file.
//Return "true",nil - if successful.
//Return "",error - if it was occured.
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

	if err != nil {
		return "", fmt.Errorf("ERR: ENCODE ERR: %s", err)
	}

	_, err = file.Write(slice)
	if err != nil {
		return "", fmt.Errorf("ERR: WRITING TO FILE ERR: %s", err)
	}
	file.Close()

	log.Println("MARSHAL AND SAVE: \n" + string(slice))
	return "true", nil
}

//showAll - return all information from database as string.
func showAll(RCashe *RCashe, cmd *command) (string, error) {
	return RCashe.String(), nil
}

//restoreData - restore database with help of json formated string(from file).
//Return "true",nil - if successful.
//Return "",error - if it was occured.
func restoreData(RCashe *RCashe, cmd *command) (string, error) {
	err := validateArgsAmount(cmd, 1)
	if err != nil {
		return "", err
	}

	file, err := os.Open(cmd.args[0])
	if err != nil {
		return "", fmt.Errorf("ERR: CAN'T READ DATA FROM FILE: %s. ERR: %s", cmd.args[0], err)
	}

	slice, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("ERR: READING FROM FILE: %s. ERR: %s", cmd.args[0], err)
	}

	file.Close()

	newCashe := newRCashe()
	err = json.Unmarshal(slice, &newCashe)
	if err != nil {
		return "", fmt.Errorf("ERR: UNMARSHAL ERR: %s", err)
	}

	RCashe.Mut.Lock()
	RCashe.DataStore = newCashe.DataStore
	RCashe.ExpKeys = newCashe.ExpKeys
	RCashe.Mut.Unlock()

	log.Println("UNMARSHAL AND RESTORE: \n" + string(slice))
	return "true", nil
}

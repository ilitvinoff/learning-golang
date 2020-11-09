package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Name string `json:"login"`
	Pass string `json:"pass"`
}

func (user *user) marshalUser() (string, error) {
	res, err := json.Marshal(user)
	return string(res), err
}

func unmarshalUser(userJSON string) (*user, error) {
	res := &user{}
	err := json.Unmarshal([]byte(userJSON), res)
	return res, err
}

func (user *user) Exists(env *envState) bool {
	var indicator = 0
	row := env.sqlDB.QueryRow("select exists(select 1 from todoDB.users where name = ?)", user.Name)

	err := row.Scan(&indicator)
	logIfError(err)
	return indicator == 1
}

func (user *user) register(env *envState) bool {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Pass), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERR: Crypt password error: {%v}; user: %v\n", err, user.Name)
		return false
	}

	if !user.Exists(env) {
		_, err := env.sqlDB.Exec("insert into todoDB.users (name,pass) values(?,?)", user.Name, string(hash))
		if err != nil {
			log.Printf("ERR: can't add user to database. User name: %v; query err: {%v}\n", user.Name, err)
			return false
		}
		return true
	}
	return false
}

func (user *user) login(env *envState) bool {
	hash := ""
	row := env.sqlDB.QueryRow("select pass from todoDB.users where name = ?", user.Name)

	err := row.Scan(&hash)
	if err != nil {
		log.Printf("ERR: can't scan pass for user: %v; scan err: {%v};\n", user.Name, err)
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(user.Pass))
	if err != nil {
		log.Printf("ERR: Decrypt password err: {%v}; user: %v\n", err, user.Name)
		return false
	}

	return true
}

func (user *user) String() string {
	return fmt.Sprintf("{name: %s; pass: %s}", user.Name, user.Pass)
}

type todoNote struct {
	ID     int    `json:"id"`
	UserID string `json:"userid"`
	Text   string `json:"text"`
	Mark   bool   `json:"checked"`
}

type noteSlice struct {
	Notes []todoNote `json:"items"`
}

func (user *user) getItems(env *envState) (string, error) {
	rows, err := env.sqlDB.Query("select * from todoDB.list where userid = ?", user.Name)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	list := noteSlice{}
	for rows.Next() {
		note := todoNote{}
		err := rows.Scan(&note.ID, &note.UserID, &note.Text, &note.Mark)
		if err != nil {
			return "", err
		}
		list.Notes = append(list.Notes, note)
	}

	res, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func (user *user) addItem(env *envState, jsonStr []byte) (string, error) {
	note := &todoNote{}
	err := json.Unmarshal(jsonStr, note)
	if err != nil {
		return "", err
	}

	env.mut.Lock()
	defer env.mut.Unlock()

	row := env.sqlDB.QueryRow("select AUTO_INCREMENT from information_schema.TABLES where TABLE_SCHEMA = \"todoDB\" and TABLE_NAME = \"list\"")

	var ID int
	err = row.Scan(&ID)
	if err != nil {
		return "", err
	}

	_, err = env.sqlDB.Exec("insert into todoDB.list (userid,text) values(?,?)", user.Name, note.Text)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("{id: %v}", ID), nil
}

func (user *user) changeItem(env *envState, jsonStr []byte) (string, error) {
	note := &todoNote{}
	err := json.Unmarshal(jsonStr, note)
	if err != nil {
		return "", err
	}

	_, err = env.sqlDB.Exec("update todoDB.list set text = ?, mark = ? where id = ?", note.Text, note.Mark, note.ID)
	if err != nil {
		return "", err
	}

	return "{ \"ok\" : true }", nil
}

func (user *user) deleteItem(env *envState, jsonStr []byte) (string, error) {
	note := &todoNote{}
	err := json.Unmarshal(jsonStr, note)
	if err != nil {
		return "", err
	}

	_, err = env.sqlDB.Exec("delete from todoDB.list where id = ?", note.ID)
	if err != nil {
		return "", err
	}

	return "{ \"ok\" : true }", nil
}

type sessionExpiration struct {
	id   string
	date time.Time
}

func (se *sessionExpiration) addSession(env *envState) error {
	var indicator int
	err := env.sqlDB.QueryRow("select exists(select 1 from todoDB.session where id = ?)", se.id).Scan(&indicator)
	if err != nil {
		return err
	}

	if indicator == 1 {
		err = se.updateSession(env)
		if err != nil {
			return err
		}
		return nil
	}

	_, err = env.sqlDB.Exec("insert into todoDB.session (id,expire) values(?,?)", se.id, se.date.Format(time.RFC3339))
	if err != nil {
		return err
	}

	return nil
}

func (se *sessionExpiration) deleteSession(env *envState) error {
	_, err := env.sqlDB.Exec("delete from todoDB.session where id = ?", se.id)
	if err != nil {
		return err
	}

	return nil
}

func (se *sessionExpiration) updateSession(env *envState) error {
	_, err := env.sqlDB.Exec("update todoDB.session set expire = ? where id = ?", se.date.Format(time.RFC3339), se.id)
	if err != nil {
		return err
	}

	return nil
}

func (se *sessionExpiration) getSessionExpTime(env *envState) (time.Time, error) {
	var expire time.Time

	err := env.sqlDB.QueryRow("select expire from todoDB.session where id = ?", se.id).Scan(&expire)
	if err != nil {
		return time.Now(), err
	}

	return expire, nil
}

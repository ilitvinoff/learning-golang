package main

import (
	"encoding/json"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Name string `json:"login"`
	Pass string `json:"pass"`
}

type okModel struct {
	Ok bool `json:"ok"`
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
	return env.redisDB.Exists(env.ctx, user.Name).Val() == 1
}

func (user *user) register(env *envState) bool {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Pass), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Crypt password error: %v; user: %v\n", err, user.Name)
		return false
	}

	if !user.Exists(env) {
		env.redisDB.Set(env.ctx, user.Name, string(hash), 0)
		return true
	}
	return false
}

func (user *user) login(env *envState) bool {
	err := bcrypt.CompareHashAndPassword([]byte(env.redisDB.Get(env.ctx, user.Name).Val()), []byte(user.Pass))
	if err != nil {
		log.Printf("Decrypt password error: %v; user: %v\n", err, user.Name)
		return false
	}

	return true
}

func (user *user) String() string {
	return fmt.Sprintf("{name: %s; pass: %s}", user.Name, user.Pass)
}

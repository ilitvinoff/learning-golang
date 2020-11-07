package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var env *envState

func main() {
	env = initEnvState()
	env.router.HandleFunc("/login", loginHandler)
	env.router.HandleFunc("/register", registerHandler)
	env.router.HandleFunc("/usersession/getItems", getItemsHandler)
	env.router.HandleFunc("/usersession/deleteItem", deleteItemHandler)
	env.router.HandleFunc("/usersession/addItem", addItemHandler)
	env.router.HandleFunc("/usersession/changeItem", changeItemHandler)
	env.router.HandleFunc("/usersession/logout", logoutHandler)
	env.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	log.Println("Listening on :3000...")
	log.Fatal(env.server.ListenAndServe())
}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	logIfError(err)

	user, err := unmarshalUser(string(body))
	logIfError(err)

	if !user.login(env) {
		http.Error(w, "user login failed!!!", http.StatusUnsupportedMediaType)
		return
	}

	fmt.Fprintf(w, "{ \"ok\": true }")
	log.Printf("loginHandler for user: %v - success!\n", user)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	logIfError(err)

	user, err := unmarshalUser(string(body))
	logIfError(err)

	if !user.register(env) {
		http.Error(w, "user register failed!!!", http.StatusUnsupportedMediaType)
		return
	}

	log.Printf("registerHandler for user: %v - success!\n", user)
}

func getItemsHandler(w http.ResponseWriter, req *http.Request){

}

func deleteItemHandler(w http.ResponseWriter, req *http.Request){

}

func addItemHandler(w http.ResponseWriter, req *http.Request){

}

func changeItemHandler(w http.ResponseWriter, req *http.Request){

}

func logoutHandler(w http.ResponseWriter, req *http.Request){

}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}

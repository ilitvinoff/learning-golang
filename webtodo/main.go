package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var env *envState

func main() {
	env = initEnvState()
	env.router.HandleFunc("/login", loginHandler)
	env.router.HandleFunc("/register", registerHandler)
	env.router.HandleFunc("/getItems", getItemsHandler)
	env.router.HandleFunc("/deleteItem", deleteItemHandler)
	env.router.HandleFunc("/addItem", addItemHandler)
	env.router.HandleFunc("/changeItem", changeItemHandler)
	env.router.HandleFunc("/logout", logoutHandler)
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
		http.Error(w, "user login failed!!!", http.StatusUnauthorized)
		return
	}

	cookie := addCookie(w, "name", user.Name, 30*time.Minute)
	userSession := &sessionExpiration{user.Name, cookie.Expires}

	err = userSession.addSession(env)
	if err != nil {
		http.Error(w, "user login failed!!!", http.StatusUnauthorized)
		log.Printf("ERR: `addSession` failed. user: %v;\n", user.Name)
		return
	}

	fmt.Fprintf(w, "{ \"ok\": true }")
	log.Printf("loginHandler for user: %v - success!\n", user)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, err := readBody(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := unmarshalUser(string(jsonStr))
	logIfError(err)

	if !user.register(env) {
		http.Error(w, "user register failed!!!", http.StatusUnauthorized)
		log.Printf("ERR: register failed. user: %v;\n", user.Name)
		return
	}

	cookie := addCookie(w, "name", user.Name, 30*time.Minute)
	userSession := &sessionExpiration{user.Name, cookie.Expires}

	if checkIfSessionExpired(env, userSession, req) {
		cookie.Expires = (time.Now().Add(-30 * time.Minute))
		http.Error(w, "ERR: session expired.", http.StatusUnauthorized)
		log.Printf("ERR: session expired. user: %v; url: %v;\n", user.Name, req.URL.Path)
		return
	}

	res := "{ \"ok\" : true }"
	fmt.Fprint(w, res)
	log.Printf("registerHandler; user: %v; remote addr: %v; request: %v; response: %v;\n", user.Name, req.RemoteAddr, string(jsonStr), res)
}

func getItemsHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, err := readBody(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	cookie, err := getCookieForUserName(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := getUserFromCookie(w, cookie, 30*time.Minute)
	if err != nil {
		log.Println(err)
		return
	}

	userSession := &sessionExpiration{user.Name, cookie.Expires}

	if checkIfSessionExpired(env, userSession, req) {
		endSessionWithCookie(w, req, cookie, userSession)
		http.Error(w, "ERR: session expired.", http.StatusUnauthorized)
		log.Printf("ERR: session expired. user: %v; url: %v;\n", user.Name, req.URL.Path)
		return
	}

	items, err := user.getItems(env)
	if err != nil {
		errBody := fmt.Sprintf("ERR: can't `getItems`. Error: %v; User: %v;\n", err, user.Name)
		http.Error(w, errBody, http.StatusNotFound)
		log.Print(errBody)
		return
	}

	fmt.Fprint(w, items)

	log.Printf("getItemsHandler; user: %v; remote addr: %v; request: %v;\n", user.Name, req.RemoteAddr, string(jsonStr))
}

func deleteItemHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, err := readBody(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	cookie, err := getCookieForUserName(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := getUserFromCookie(w, cookie, 30*time.Minute)
	if err != nil {
		log.Println(err)
		return
	}

	userSession := &sessionExpiration{user.Name, cookie.Expires}

	if checkIfSessionExpired(env, userSession, req) {
		endSessionWithCookie(w, req, cookie, userSession)
		http.Error(w, "ERR: session expired.", http.StatusUnauthorized)
		log.Printf("ERR: session expired. user: %v; url: %v;\n", user.Name, req.URL.Path)
		return
	}

	res, err := user.deleteItem(env, jsonStr)
	if err != nil {
		log.Printf("ERR: can't `deleteItem`. Error: %v; User: %v;\n", err, user.Name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, res)
	log.Printf("deleteItemHandler; user: %v; remote addr: %v; request: %v; response: %v;\n", user.Name, req.RemoteAddr, string(jsonStr), res)
}

func addItemHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, err := readBody(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	cookie, err := getCookieForUserName(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := getUserFromCookie(w, cookie, 30*time.Minute)
	if err != nil {
		log.Println(err)
		return
	}

	userSession := &sessionExpiration{user.Name, cookie.Expires}

	if checkIfSessionExpired(env, userSession, req) {
		endSessionWithCookie(w, req, cookie, userSession)
		http.Error(w, "ERR: session expired.", http.StatusUnauthorized)
		log.Printf("ERR: session expired. user: %v; url: %v;\n", user.Name, req.URL.Path)
		return
	}

	res, err := user.addItem(env, jsonStr)
	if err != nil {
		log.Printf("ERR: can't `addItem`. Error: %v; User: %v;\n", err, user.Name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("add ID: " + res)
	fmt.Fprint(w, res)

	log.Printf("addItemHandler; user: %v; remote addr: %v; request: %v; response: %v;\n", user.Name, req.RemoteAddr, string(jsonStr), res)
}

func changeItemHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, err := readBody(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	cookie, err := getCookieForUserName(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := getUserFromCookie(w, cookie, 30*time.Minute)
	if err != nil {
		log.Println(err)
		return
	}

	userSession := &sessionExpiration{user.Name, cookie.Expires}

	if checkIfSessionExpired(env, userSession, req) {
		endSessionWithCookie(w, req, cookie, userSession)
		http.Error(w, "ERR: session expired.", http.StatusUnauthorized)
		log.Printf("ERR: session expired. user: %v; url: %v;\n", user.Name, req.URL.Path)
		return
	}

	res, err := user.changeItem(env, jsonStr)
	if err != nil {
		log.Printf("ERR: can't `changeItem`. Error: %v; User: %v;\n", err, user.Name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, res)

	log.Printf("changeItemHandler; user: %v; remote addr: %v; request: %v; response: %v;\n", user.Name, req.RemoteAddr, string(jsonStr), res)
}

func logoutHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, err := readBody(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	cookie, err := getCookieForUserName(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	userSession := &sessionExpiration{cookie.Value, cookie.Expires}

	endSessionWithCookie(w, req, cookie, userSession)
	res := "{ \"ok\" : true }"
	fmt.Fprint(w, res)

	log.Printf("logoutHandler; user: %v; remote addr: %v; request: %v; response: %v;\n", cookie.Value, req.RemoteAddr, string(jsonStr), res)
}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func checkIfSessionExpired(env *envState, se *sessionExpiration, req *http.Request) bool {

	expireTime, err := se.getSessionExpTime(env)
	if err != nil {
		log.Printf("ERR: can't `getSessionExpTime`. Err: %v; User: %v; URL: %v;\n", err, se.id, req.RemoteAddr)
		err = se.deleteSession(env)
		if err != nil {
			log.Printf("ERR: can't `deleteSession`. Err: %v; User: %v; URL: %v;\n", err, se.id, req.RemoteAddr)
		}
		return true
	}

	if expireTime.Before(se.date) {
		err = se.deleteSession(env)
		if err != nil {
			log.Printf("ERR: can't `deleteSession`. Err: %v; User: %v; URL: %v;\n", err, se.id, req.RemoteAddr)
		}
		return true
	}

	se.addSession(env)

	return false
}

// addCookie will apply a new cookie to the response of a http request
// with the key/value specified.
func addCookie(w http.ResponseWriter, name, value string, ttl time.Duration) *http.Cookie {
	expire := time.Now().Add(ttl)
	cookie := &http.Cookie{
		Name:    name,
		Value:   value,
		Expires: expire,
	}
	http.SetCookie(w, cookie)
	return cookie
}

func getUserFromCookie(w http.ResponseWriter, cookie *http.Cookie, ttl time.Duration) (*user, error) {
	cookie.Expires = time.Now().Add(ttl)
	http.SetCookie(w, cookie)

	return &user{Name: cookie.Value}, nil
}

func getCookieForUserName(w http.ResponseWriter, req *http.Request) (*http.Cookie, error) {
	cookie, err := req.Cookie("name")
	if err != nil {
		errBody := fmt.Sprintf("ERR: no cookie found for `%v`. Remote addr: %v;", req.URL.Path, req.RemoteAddr)
		http.Error(w, errBody, http.StatusNotFound)
		return &http.Cookie{}, fmt.Errorf(errBody)
	}

	return cookie, nil
}

func readBody(w http.ResponseWriter, req *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errBody := fmt.Sprintf("ERR: can't read request body for `%v`. Error: %v; Remote addr: %v;", req.URL.Path, err, req.RemoteAddr)
		http.Error(w, errBody, http.StatusBadRequest)
	}

	return body, nil
}

func endSessionWithCookie(w http.ResponseWriter, req *http.Request, cookie *http.Cookie, se *sessionExpiration) {
	cookie.Expires = time.Now().Add(-30 * time.Minute)
	http.SetCookie(w, cookie)

	err := se.deleteSession(env)
	if err != nil {
		log.Printf("ERR: can't `deleteSession`. Err: %v; User: %v; URL: %v;\n", err, se.id, req.RemoteAddr)
	}
}

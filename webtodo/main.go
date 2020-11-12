package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	cookieLiveTime = 30 * time.Minute
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
	jsonStr, err := readBody(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := unmarshal(string(jsonStr))
	logIfError(err)

	if !user.login(env) {
		http.Error(w, "user login failed!!!", http.StatusUnauthorized)
		return
	}

	cookie := addCookie(w, "name", user.Name, cookieLiveTime)
	userSession := &sessionExpiration{user.Name, cookie.Expires}

	err = userSession.addSession(env)
	if err != nil {
		http.Error(w, "user login failed!!!", http.StatusUnauthorized)
		log.Printf("ERR: `addSession` failed. user: %v;\n", user.Name)
		return
	}

	fmt.Fprintf(w, "{ \"ok\": true }")
	log.Printf("loginHandler for user: %v - success!\n", user.Name)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, err := readBody(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := unmarshal(string(jsonStr))
	logIfError(err)

	res, err := user.register(env)
	if err != nil {
		http.Error(w, "user register failed!!!", http.StatusUnauthorized)
		log.Printf("ERR: %v", err)
		return
	}

	fmt.Fprint(w, res)
	successLogger("registerHandler", &http.Cookie{}, req, []byte(user.Name), res)
}

func getItemsHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, cookie, user, userSession, err := getUserParams(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	expired, err := checkIfSessionExpired(env, userSession, req)
	logIfError(err)
	if expired {
		expirationHandler(w, req, cookie, user, userSession)
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

	successLogger("getItemsHandler", cookie, req, jsonStr, "")
}

func deleteItemHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, cookie, user, userSession, err := getUserParams(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	expired, err := checkIfSessionExpired(env, userSession, req)
	logIfError(err)
	if expired {
		expirationHandler(w, req, cookie, user, userSession)
		return
	}

	res, err := user.deleteItem(env, jsonStr)
	if err != nil {
		log.Printf("ERR: can't `deleteItem`. Error: %v; User: %v;\n", err, user.Name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, res)

	successLogger("deleteItemHandler", cookie, req, jsonStr, res)
}

func addItemHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, cookie, user, userSession, err := getUserParams(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	expired, err := checkIfSessionExpired(env, userSession, req)
	logIfError(err)
	if expired {
		expirationHandler(w, req, cookie, user, userSession)
		return
	}

	res, err := user.addItem(env, jsonStr)
	if err != nil {
		log.Printf("ERR: can't `addItem`. Error: %v; User: %v;\n", err, user.Name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, res)

	successLogger("addItemHandler", cookie, req, jsonStr, res)
}

func changeItemHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, cookie, user, userSession, err := getUserParams(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	expired, err := checkIfSessionExpired(env, userSession, req)
	logIfError(err)
	if expired {
		expirationHandler(w, req, cookie, user, userSession)
		return
	}

	res, err := user.changeItem(env, jsonStr)
	if err != nil {
		log.Printf("ERR: can't `changeItem`. Error: %v; User: %v;\n", err, user.Name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, res)

	successLogger("changeItemHandler", cookie, req, jsonStr, res)
}

func logoutHandler(w http.ResponseWriter, req *http.Request) {
	jsonStr, cookie, _, userSession, err := getUserParams(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	endSessionWithCookie(w, req, cookie, userSession)

	res, err := marshal(result{Ok: true})
	if err != nil {
		log.Println("ERR: can't marshal when logout.")
		return
	}

	fmt.Fprint(w, res)

	successLogger("logoutHandler", cookie, req, jsonStr, res)
}

func successLogger(operationName string, cookie *http.Cookie, req *http.Request, jsonStr []byte, operationResult string) {
	log.Printf("%v; user: %v; remote addr: %v; request: %v; response: %v;\n", operationName, cookie.Value, req.RemoteAddr, string(jsonStr), operationResult)
}

func expirationHandler(w http.ResponseWriter, req *http.Request, cookie *http.Cookie, user *user, userSession *sessionExpiration) {
	endSessionWithCookie(w, req, cookie, userSession)
	http.Error(w, "ERR: session expired.", http.StatusUnauthorized)
	log.Printf("ERR: session expired. user: %v; url: %v;\n", user.Name, req.URL.Path)
}

func getUserParams(w http.ResponseWriter, req *http.Request) ([]byte, *http.Cookie, *user, *sessionExpiration, error) {
	jsonStr, err := readBody(w, req)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, nil, err
	}

	cookie, err := getCookieForUserName(w, req)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	user, err := getUserFromCookie(w, cookie, cookieLiveTime)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, nil, err
	}

	userSession := &sessionExpiration{user.Name, cookie.Expires}

	return jsonStr, cookie, user, userSession, nil
}

func logIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func checkIfSessionExpired(env *envState, se *sessionExpiration, req *http.Request) (bool, error) {

	expireTime, err := se.getSessionExpTime(env)
	if err != nil {
		err = se.deleteSession(env)
		if err != nil {
			return true, fmt.Errorf("can't `deleteSession`. Err: %v; User: %v; URL: %v;", err, se.id, req.RemoteAddr)
		}

		return true, fmt.Errorf("can't `getSessionExpTime`. Err: %v; User: %v; URL: %v;", err, se.id, req.RemoteAddr)
	}

	if expireTime.Before(se.date) {
		err = se.deleteSession(env)
		if err != nil {
			return true, fmt.Errorf("can't `deleteSession`. Err: %v; User: %v; URL: %v;", err, se.id, req.RemoteAddr)
		}

		return true, nil
	}

	se.addSession(env)

	return false, nil
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

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
)

type envState struct {
	mut    *sync.Mutex
	router *mux.Router
	ctx    context.Context
	sqlDB  *sql.DB
	server *http.Server
}

func newSQLdb() *sql.DB {
	envs, err := getUserPassDB()
	if err != nil {
		log.Fatalf("ERR: DB user initiation failed. Err: %v;", err)
	}

	//envs{dbUser,dbPass,dbName}
	dataSourceName := fmt.Sprint(envs[0], ":", envs[1], "@/", envs[2], "?parseTime=true")

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal("Can't open database:\n" + err.Error())
	}
	return db
}

func newServer() *http.Server {
	return &http.Server{
		Addr: "localhost:3000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func initEnvState() *envState {
	env := &envState{&sync.Mutex{}, mux.NewRouter(), context.Background(), newSQLdb(), newServer()}
	env.server.Handler = env.router
	return env
}

func getUserPassDB() ([]string, error) {
	dbUser, exists := os.LookupEnv("dbuser")
	if !exists {
		return nil, fmt.Errorf("No such environment variable as `dbuser`")
	}

	dbPass, exists := os.LookupEnv("dbpass")
	if !exists {
		return nil, fmt.Errorf("No such environment variable as `dbpass`")
	}

	dbName, exists := os.LookupEnv("dbname")
	if !exists {
		return nil, fmt.Errorf("No such environment variable as `dbname`")
	}

	return []string{dbUser, dbPass, dbName}, nil
}

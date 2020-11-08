package main

import (
	"context"
	"log"
	"net/http"
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
	db, err := sql.Open("mysql", "dbuser:ololo@/counterDB?parseTime=true")
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

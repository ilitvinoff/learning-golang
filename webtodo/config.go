package main

import (
	"context"
	"net/http"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type envState struct {
	router *mux.Router
	ctx    context.Context
	sqlDB  *sql.DB
	server *http.Server
}

func newSQLdb() *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 1})
	return client
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
	env := &envState{mux.NewRouter(), context.Background(), newSQLdb(), newServer()}
	env.server.Handler = env.router
	return env
}

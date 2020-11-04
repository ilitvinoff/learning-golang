package main

import (
	"context"
	"database/sql"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
)

//Counter counts amount of visits to web server
type Counter struct {
	Mut   *sync.Mutex
	value int
}

type myConfig struct {
	ctx      context.Context
	db       *sql.DB
	rdb      *redis.Client
	iconPath string
	counter  *Counter
}

func newRDBclient() *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})
	client.Set(context.Background(), "total per session", 0, 0)
	return client
}

func newSQLdb() *sql.DB {
	db, err := sql.Open("mysql", "dbuser:ololo@/counterDB")
	if err != nil {
		log.Fatal("Can't open database:\n" + err.Error())
	}

	_, err = db.Exec("insert into counterDB.info (remoteAddr,visits) values(?, ?)", "total per session", "0")
	if err != nil {
		log.Fatal("Can't add 'total' row:\n" + err.Error())
	}
	return db
}

func newCounter() *Counter {
	return &Counter{Mut: &sync.Mutex{}, value: 0}
}

func newConfig() *myConfig {
	return &myConfig{context.Background(), newSQLdb(), newRDBclient(), "server/favicon_io.zip", newCounter()}
}

func (mc *myConfig) Close() {
	mc.db.Close()
	mc.rdb.Close()
}

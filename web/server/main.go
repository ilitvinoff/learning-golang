package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

const (
	iconPath = "server/favicon_io.zip"
)

func main() {
	mc := newConfig()
	defer mc.Close()

	countHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, counter(mc, req))
	}

	printBases := func(w http.ResponseWriter, req *http.Request) {
		result := "SQL:\n"
		result = fmt.Sprint(result, showSQLbase(mc), "Redis:\n", showRDB(mc))
		io.WriteString(w, result)
	}

	http.HandleFunc("/favicon.ico", favIconHandler)
	http.HandleFunc("/", countHandler)
	http.HandleFunc("/showbase", printBases)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func favIconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, iconPath)
	fmt.Println("Icon")
}

func counter(mc *myConfig, req *http.Request) string {
	mc.counter.Mut.Lock()
	defer mc.counter.Mut.Unlock()
	mc.counter.value++
	counterAsString := strconv.Itoa(mc.counter.value)

	mc.rdb.Set(mc.ctx, "total per session", mc.counter.value, 0)

	if mc.rdb.Exists(mc.ctx, req.RemoteAddr).String() == "1" {
		mc.rdb.Incr(mc.ctx, req.RemoteAddr)
	} else {
		mc.rdb.Set(mc.ctx, req.RemoteAddr, 1, 0)
	}

	if addressExits(req.RemoteAddr, mc.db) {
		visits := stringToInt(getVisitsFromSQL(req.RemoteAddr, mc.db))
		visits++
		updateSQLvisits(req.RemoteAddr, visits, mc.db)
		updateSQLvisits("total per session", mc.counter.value, mc.db)

		return counterAsString
	}

	result, err := mc.db.Exec("insert into counterDB.info (remoteAddr,visits) values(?, ?)", req.RemoteAddr, "1")
	updateSQLvisits("total per session", mc.counter.value, mc.db)
	if err != nil {
		s := fmt.Sprint("result: ", result, "; error: ", err.Error())
		return s
	}
	return counterAsString
}

func showSQLbase(mc *myConfig) string {
	rows, err := mc.db.Query("select * from counterDB.info")
	if err != nil {
		return "can't show database: " + err.Error()
	}

	var totalVisits string
	total := "total per session"
	result := ""

	for rows.Next() {
		var remoteAddr, visits string

		err = rows.Scan(&remoteAddr, &visits)
		if err != nil {
			return err.Error()
		}
		if remoteAddr == total {
			totalVisits = visits
			continue
		}
		result = fmt.Sprint(result, "remote address: ", remoteAddr, "; visits: ", visits, "\n")
	}
	result = fmt.Sprint(result, "remote address: ", total, "; visits: ", totalVisits, "\n")

	return result
}

func addressExits(addr string, db *sql.DB) bool {
	var indicator int

	row := db.QueryRow("select exists(select 1 from counterDB.info where remoteAddr = ?)", addr)

	err := row.Scan(&indicator)
	if err != nil {
		fmt.Println("err: " + err.Error())
		return false
	}

	return indicator == 1
}

func getVisitsFromSQL(addr string, db *sql.DB) string {
	var value string

	row := db.QueryRow("select visits from counterDB.info where remoteAddr = ?", addr)

	err := row.Scan(&value)
	if err != nil {
		fmt.Println("err: " + err.Error())
		return "0"
	}

	return value
}

func updateSQLvisits(addr string, visits int, db *sql.DB) {
	result, err := db.Exec("update counterDB.info set visits = ? where remoteAddr = ?", visits, addr)
	if err != nil {
		log.Fatal(fmt.Sprint("result: ", result, "; error: ", err.Error()))

	}
}

func stringToInt(value string) int {
	res, err := strconv.Atoi(value)
	if err != nil {
		log.Fatal("Can't convert value to number:\n" + err.Error())
	}
	return res
}

func showRDB(mc *myConfig) string {
	var result string
	total := "total per session"
	totalVisits := "0"

	keys := mc.rdb.Keys(mc.ctx, "*").Val()

	for _, key := range keys {
		visits := mc.rdb.Get(mc.ctx, key).Val()
		if key == total {
			totalVisits = visits
			continue
		}
		result = fmt.Sprint(result, "remote address: ", key, "; visits: ", visits, "\n")
	}
	result = fmt.Sprint(result, "remote address: ", total, "; visits: ", totalVisits, "\n")
	return result
}

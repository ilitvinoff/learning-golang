package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

const (
	iconPath = "server/favicon_io.zip"
)

//Counter counts amount of visits to web server
type Counter struct {
	Mut   *sync.Mutex
	value int
}

var db *sql.DB

func main() {
	user := printDataIn("enter username, please: ")
	password := printDataIn("enter the password, please: ")
	dbOpenStr := fmt.Sprint(user, ":", password, "@/counterDB")
	//dbOpenStr := fmt.Sprint("dbuser:ololo@/counterDB")

	db, err := sql.Open("mysql", dbOpenStr)
	if err != nil {
		log.Fatal("Can't open database:\n" + err.Error())
	}
	defer db.Close()

	c := &Counter{Mut: &sync.Mutex{}, value: 0}

	countHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, counter(c, req, db))
	}

	showBaseHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, showBase(db))
	}

	http.HandleFunc("/favicon.ico", favIconHandler)
	http.HandleFunc("/", countHandler)
	http.HandleFunc("/showbase", showBaseHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func favIconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, iconPath)
	fmt.Println("Icon")
}

func counter(c *Counter, req *http.Request, db *sql.DB) string {
	c.Mut.Lock()
	defer c.Mut.Unlock()
	c.value++
	counterAsString := strconv.Itoa(c.value)

	if addressExits(req.RemoteAddr, db) {
		query := fmt.Sprintf("update counterDB.info set visits = %v where remoteAddr = '%v'", counterAsString, req.RemoteAddr)
		result, err := db.Exec(query)
		if err != nil {
			s := fmt.Sprint("result: ", result, "; error: ", err.Error())
			return s
		}
		return counterAsString
	}

	result, err := db.Exec("insert into counterDB.info (remoteAddr,visits) values(?, ?)", req.RemoteAddr, counterAsString)
	if err != nil {
		s := fmt.Sprint("result: ", result, "; error: ", err.Error())
		return s
	}
	return counterAsString
}

func showBase(db *sql.DB) string {
	rows, err := db.Query("select * from counterDB.info")
	if err != nil {
		return "can't show database: " + err.Error()
	}

	result := ""

	for rows.Next() {
		buffer := struct {
			remoteAddr string
			visits     string
		}{}
		err = rows.Scan(&buffer.remoteAddr, &buffer.visits)
		if err != nil {
			return err.Error()
		}
		result = fmt.Sprint(result, "remote address: ", buffer.remoteAddr, "; visits: ", buffer.visits, "\n")
	}

	return result
}

func addressExits(addr string, db *sql.DB) bool {
	var indicator int
	query := fmt.Sprintf("select exists(select 1 from counterDB.info where remoteAddr = '%v')", addr)
	row := db.QueryRow(query)

	err := row.Scan(&indicator)
	if err != nil {
		fmt.Println("err: " + err.Error())
		return false
	}

	return indicator == 1
}

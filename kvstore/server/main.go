package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

//Program deploys the server with database(redis format) on it.
//Possible client's commands:
// set <key> <value> - add key:value pair to database
// get <key> - returns the value corresponding to the key
// getset <key> <value> - set value to key-element and returns it's previous value. If no previous value - returns error
// exist <key> - check if element correspondig to key - is exist. Return true - if it is, false - if not.
// del <key> <key> ...- delete all elements corresponded to pool of keys. Return amount of deleted values
// ex <key> <seconds> - set expiration date to key's-element.
// save <filepath> - save database as json to file(if file not exist - creats it).
// restore <filepath> - restore database from file.
// showall - return all information about database

const (
	defaultProtocol = "tcp"
	defaultPort     = ":16998"
)

type config struct {
	protocol string
	port     string
}

func main() {
	config := getConfig(os.Args)

	l, err := net.Listen(config.protocol, config.port)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("LOG: The server started listening;")

	rc := newKVCache()

	go rc.expirationWatcher()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("LOG: New client connected: %s;\n", conn.RemoteAddr())

		go handleConnection(rc, conn)
	}
}

func getConfig(args []string) *config {
	config := &config{defaultProtocol, defaultPort}

	if len(args) == 2 {
		config.port = fmt.Sprint(":", args[1])
		return config
	}

	if len(args) == 3 {
		config.port = fmt.Sprint(":", args[1])
		config.protocol = args[2]
		return config
	}

	return config
}

func ifErrFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func logErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

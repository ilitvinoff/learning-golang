package main

import (
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
// ex <key> <seconds> - set exparation date to key's-element.
// save <filepath> - save database as json to file(if file not exist - creats it).
// restore <filepath> - restore database from file.
// showall - return all information about database

const (
	protocol         = "tcp"
	defaultPort      = ":16998"
	startMessage     = "LOG: The server started listening;"
	newClientMessage = "LOG: New client connected: %s;\n"
)

func main() {
	port := defaultPort
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	l, err := net.Listen(protocol, port)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(startMessage)

	rc := newRCashe()

	go rc.exparationWatcher()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf(newClientMessage, conn.RemoteAddr())

		go handleConnection(rc, conn)
	}
}

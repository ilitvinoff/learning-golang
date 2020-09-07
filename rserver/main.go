package main

import (
	"log"
	"net"
	"os"
)

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

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	requestReadErrMessage    = "ERR: Request reading error. Request: %s. Client addres: %s. IO err: %s;"
	requestLengthErrMessage  = "ERR: Request length reading error. Request length: %s. IO err: %s;"
	clientAddrTitle          = "Client addres: %s;"
	unknownCommandErrMessage = "ERR:Unknown command: %s\n;"
	responseErrMessage       = "ERR:Unknown command: %s. Client addres: %s;"
	socketClosedLogMessage   = "LOG: the end of socket for client: %s;"
	requestLogMessage        = "LOG: client: %s, request: %s;"
)

func getRequestLength(reader *bufio.Reader) (int, error) {
	requestLength, err := reader.ReadString(':')
	if err != nil {
		return -1, fmt.Errorf(requestLengthErrMessage, requestLength, err)
	}

	requestLength = requestLength[0 : len(requestLength)-1]

	n, err := strconv.Atoi(requestLength)

	if err != nil {
		return -1, fmt.Errorf("Can't get request length: %s", err)
	}

	return n, nil
}

func readRequest(conn net.Conn) (string, error) {

	connReader := bufio.NewReader(conn)

	requestLength, err := getRequestLength(connReader)
	if err != nil {
		return "", fmt.Errorf(err.Error()+clientAddrTitle, conn.RemoteAddr())
	}

	request := make([]byte, requestLength)

	_, err = connReader.Read(request)

	if err != nil {
		return "", fmt.Errorf(requestReadErrMessage, string(request), conn.RemoteAddr(), err)
	}

	return string(request), nil
}

func parseRequest(request string) *command {
	expression := strings.Split(request, " ")

	return &command{expression[0], expression[1:]}
}

func getResponse(conn net.Conn, cmd *command, rc *RCashe) (string, error) {
	executor, ok := commands[cmd.name]
	if !ok {
		return "", fmt.Errorf(responseErrMessage, cmd.name, conn.RemoteAddr())
	}

	result, err := executor(rc, cmd)
	if err != nil {
		return "", err
	}

	return addLenPrefix(result), nil
}

func handleConnection(rc *RCashe, conn net.Conn) {
	defer conn.Close()
	defer log.Printf(socketClosedLogMessage, conn.RemoteAddr())

	for {
		request, err := readRequest(conn)

		if err != nil {
			log.Println(err)
			break
		}
		log.Printf(requestLogMessage, conn.RemoteAddr(), request)

		cmd := parseRequest(request)

		response, err := getResponse(conn, cmd, rc)
		if err != nil {
			log.Println(err)

			_, err = conn.Write([]byte(addLenPrefix(err.Error())))
			if err != nil {
				log.Printf("ERR: Request send error: %s", err)
			}

			continue
		}

		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Printf("ERR: Request send error: %s", err)
		}

		log.Printf("LOG: Response: %s, client adrres: %s", response, conn.RemoteAddr())
	}

}

func addLenPrefix(str string) string {
	return strconv.Itoa(len(str)) + ":" + str
}

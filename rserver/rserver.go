package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
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

//getRequestLength - based on the netstring protocol, returns the request length from the request.
//Returns error - if it was occured
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

//readRequest -  reads and returns client's reaquest.
//Returns error - if it was occured
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

//parseRequest - parses request to command and arguments.
//Return:
//*command,nil - if successful;
//nil,error - if not.
func parseRequest(request string) (*command, error) {
	var parseResult []string
	state := "start"
	current := ""
	quote := "\""
	escapeNext := true
	for i := 0; i < len(request); i++ {
		c := request[i]

		if state == "quotes" {
			if string(c) != quote {
				current += string(c)
			} else {
				parseResult = append(parseResult, current)
				current = ""
				state = "start"
			}
			continue
		}

		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		if c == '"' || c == '\'' {
			state = "quotes"
			quote = string(c)
			continue
		}

		if state == "arg" {
			if c == ' ' || c == '\t' {
				parseResult = append(parseResult, current)
				current = ""
				state = "start"
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = "arg"
			current += string(c)
		}
	}

	if state == "quotes" {
		return nil, fmt.Errorf("Unclosed quote in command line: %s", request)
	}

	if current != "" {
		parseResult = append(parseResult, current)
	}

	return &command{parseResult[0], parseResult[1:]}, nil
}

//getResponse - generates a response based on user's request.
//Return:
//response, nil - if successful;
//"", error - if not.
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

//handleConnection - when client connected handle the connection.
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

		cmd, err := parseRequest(request)
		if err != nil {
			_, err = conn.Write([]byte(addLenPrefix(err.Error())))
			if err != nil {
				log.Printf("ERR: <Parse err> send error: %s", err)
			}
			continue
		}

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

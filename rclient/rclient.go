package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

const (
	protocol                      = "tcp"
	defaultAddr                   = "127.0.0.1:16998"
	withoutResponseLimit          = 7
	responseLengthValueErrMessage = "ERR: Bad respoonse length Value: %s"
	responseLengthReadErrMessage  = "ERR: Can't read response lentgth. Have read: %s. Error: %s"
	sendTitle                     = ">>send: "
	stdinErrMessage               = "STDIN reading problem. Input: %s"
	writeRequestErrMessage        = "ERR: %d bytes were written. ERR: %s;"
	readResponseErrMessage        = "ERR: No response from server: %s;"
	responseTitle                 = "LOG: RESPONSE: %s"
)

func main() {
	addr := defaultAddr
	if len(os.Args) == 2 {
		addr = os.Args[1]
	}

	conn, err := net.Dial(protocol, addr)

	ifErrFatal(err)
	defer conn.Close()

	withoutResponseCounter := 0
	for {

		request, err := getRequest()
		if err != nil {
			log.Println(err)
			continue
		}

		n, err := conn.Write([]byte(request))
		if err != nil {
			log.Printf(writeRequestErrMessage, n, err.Error())
		}

		connReader := bufio.NewReader(conn)

		responseLength, err := getResponseLength(connReader)
		if err != nil {
			log.Println(err)
			continue
		}
		response := make([]byte, responseLength)

		_, err = connReader.Read(response)

		if err != nil {
			log.Printf(readResponseErrMessage, err.Error())

			withoutResponseCounter++
			if withoutResponseCounter >= withoutResponseLimit {
				break
			}
			continue
		}
		log.Printf(responseTitle, string(response))
	}
}

func getRequest() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(sendTitle)
	request, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf(stdinErrMessage, request)
	}

	request = request[0 : len(request)-1]
	request = strconv.Itoa(len(request)) + ":" + request

	return request, nil
}

func getResponseLength(connReader *bufio.Reader) (int, error) {
	responseLength, err := connReader.ReadString(':')
	if err != nil {
		return -1, fmt.Errorf(responseLengthReadErrMessage, responseLength, err)
	}

	responseLength = responseLength[0 : len(responseLength)-1]

	result, err := strconv.Atoi(responseLength)
	if err != nil {
		return -1, fmt.Errorf(responseLengthValueErrMessage, responseLength)
	}

	return result, nil
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

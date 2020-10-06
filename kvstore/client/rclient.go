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
	defaultProtocol      = "tcp"
	defaultAddr          = "127.0.0.1:16998"
	withoutResponseLimit = 7
)

type config struct {
	protocol string
	addr     string
}

func main() {
	config := getConfig(os.Args)

	conn, err := net.Dial(config.protocol, config.addr)

	ifErrFatal(err)
	handleConnection(conn)
	conn.Close()

}

func handleConnection(conn net.Conn) {
	withoutResponseCounter := 0
	for {

		request, err := getRequest()
		if err != nil {
			log.Println(err)
			continue
		}

		if request == "exit" {
			break
		}

		n, err := conn.Write([]byte(makeNetstring(request)))
		if err != nil {
			log.Printf("ERR: %d bytes were written. ERR: %s;", n, err.Error())
		}

		connReader := bufio.NewReader(conn)

		responseLength, err := getResponseLength(connReader)
		if err != nil {
			log.Println(err)

			withoutResponseCounter++
			if withoutResponseCounter >= withoutResponseLimit {
				break
			}
			continue
		}
		response := make([]byte, responseLength)

		_, err = connReader.Read(response)

		if err != nil {
			log.Printf("ERR: No response from server: %s;", err.Error())

			withoutResponseCounter++
			if withoutResponseCounter >= withoutResponseLimit {
				break
			}
			continue
		}
		log.Printf("LOG: RESPONSE: %s;", string(response))
	}
}

func getRequest() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(">>send: ")
	request, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("ERR: STDIN reading problem. Input: %s;", request)
	}
	request = request[0 : len(request)-1]
	return request, nil
}

func makeNetstring(str string) string {
	return strconv.Itoa(len(str)) + ":" + str
}

func getResponseLength(connReader *bufio.Reader) (int, error) {
	responseLength, err := connReader.ReadString(':')
	if err != nil {
		return -1, fmt.Errorf("ERR: Can't read response lentgth. Have read: %s. Error: %s;", responseLength, err)
	}

	responseLength = responseLength[0 : len(responseLength)-1]

	result, err := strconv.Atoi(responseLength)
	if err != nil {
		return -1, fmt.Errorf("ERR: Bad respoonse length Value: %s;", responseLength)
	}

	return result, nil
}

func getConfig(args []string) *config {
	config := &config{defaultProtocol, defaultAddr}

	if len(args) == 2 {
		config.addr = args[1]
		return config
	}

	if len(args) == 3 {
		config.addr = args[1]
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

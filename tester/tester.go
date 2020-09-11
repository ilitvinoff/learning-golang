package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

//Program to make stresstest to server with database(redis format).

const (
	protocol                      = "tcp"
	defaultAddr                   = "127.0.0.1:16998"
	limitOfRequests               = 100
	withoutResponseLimit          = 7
	responseLengthValueErrMessage = "ERR: Bad respoonse length Value: %s"
	responseLengthReadErrMessage  = "ERR: Can't read response lentgth. Have read: %s. Error: %s"
	sendTitle                     = ">>send: "
	stdinErrMessage               = "STDIN reading problem. Input: %s"
	writeRequestErrMessage        = "ERR: %d bytes were written. ERR: %s;"
	readResponseErrMessage        = "ERR: No response from server: %s;"
	responseTitle                 = "LOG: RESPONSE: %s"
)

//commands []string - List of commands to test with
var commands = []string{"set n 1", "get n", "getset n 2", "exist n", "ex n 20"}

//routinesLimit []int - list of the number of clients currently connected to the server
var routinesLimit = []int{10, 100, 1000}

//metrics - created to store metrics of testing
type metrics struct {
	command        string
	routinesAmount int
	execTimeChan   chan time.Duration
}

func main() {
	addr := defaultAddr
	if len(os.Args) == 2 {
		addr = os.Args[1]
	}

	//for each command from commands-list
	for _, command := range commands {
		//for each possible amount of concurrent clients(from routinesLimit = []int{10, 100, 1000})
		for _, n := range routinesLimit {
			result := &metrics{command, n, make(chan time.Duration, n*limitOfRequests)}

			//start n-amount of consurrent clients(strestesters)
			for i := 0; i < n; i++ {
				go stressTester(addr, result)
			}

			//printout result of testing
			fmt.Println(logResult(result))
			close(result.execTimeChan)
		}
	}
}

//logReslt calculate min,max,average time of response from server for requested command.
//Return max,min,average time, command name, amount of clients.
func logResult(result *metrics) string {
	var max, min, average time.Duration
	var counter int64
	var execTime time.Duration

	channelBufferSize := result.routinesAmount * limitOfRequests
	for i := 0; i < channelBufferSize; i++ {
		execTime = <-result.execTimeChan

		if execTime > max {
			max = execTime
		}

		if execTime < min {
			min = execTime
		}
		average += execTime
		counter++
	}

	average = time.Duration(int64(average.Nanoseconds()) / counter)

	return fmt.Sprintf("COMMAND: %s, ROUTINES AMOUNT: %d, DURATIONS:\nmin: %s\nmax: %s\naverage: %s", result.command, result.routinesAmount, min, max, average)
}

//stressTester - open single connection to server, "atack" it with request-command 100-times(limitOfRequests),
//measure response time from server and store it to metrics chan.
func stressTester(address string, metrics *metrics) {
	conn, err := net.Dial(protocol, address)

	ifErrFatal(err)
	defer conn.Close()

	for i := 0; i < limitOfRequests; i++ {

		n, err := conn.Write(getRequest(metrics.command))
		if err != nil {
			log.Printf(writeRequestErrMessage, n, err.Error())
		}

		start := time.Now()
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
		}

		metrics.execTimeChan <- time.Since(start)
	}
}

func getRequest(cmd string) []byte {
	cmd = cmd[0:len(cmd)]
	cmd = strconv.Itoa(len(cmd)) + ":" + cmd

	return []byte(cmd)
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

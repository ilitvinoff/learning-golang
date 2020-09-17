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
	defaultProtocol      = "tcp"
	defaultAddr          = "127.0.0.1:16998"
	limitOfRequests      = 100
	withoutResponseLimit = 7
)

type config struct {
	protocol string
	addr     string
}

//commands []string - List of commands to test with
var commands = []string{"set n 1", "get n", "getset n 2", "exist n", "ex n 20"}

//routinesLimit []int - list of the number of clients currently connected to the server
var routinesLimit = []int{10, 100, 1000}

type logStruct struct {
	err      error
	execTime time.Duration
}

//logReslt calculates min,max,average time of response from server for requested command.
//Counts the number of errors.
//Return max,min,average time, command name, amount of clients.
func getMetrics(command string, routinesCount int, logChan chan logStruct) string {
	var max, min, average time.Duration
	var counter, errCounter int64
	var execTime time.Duration
	errMap := make(map[error]int)

	for i := 0; i < cap(logChan); i++ {
		metricUnit := <-logChan
		if metricUnit.err != nil {
			errMap[metricUnit.err] = errMap[metricUnit.err] + 1
			errCounter++
			continue
		}

		execTime = metricUnit.execTime

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
	res := fmt.Sprintf("COMMAND: %s, ROUTINES AMOUNT: %d, DURATIONS:\nmin: %s\nmax: %s\naverage: %s\n", command, routinesCount, min, max, average)
	res = fmt.Sprintf(res+"ERRORS:\nTOTAL: %d\nError's values:\n", errCounter)
	for k, v := range errMap {
		res = fmt.Sprintf(res+"%s [%d]\n", k, v)
	}
	return res
}

func main() {
	config := getConfig(os.Args)

	//for each command from commands-list
	for _, command := range commands {
		//for each possible amount of concurrent clients(from routinesLimit = []int{10, 100, 1000})
		for _, n := range routinesLimit {
			logChan := make(chan logStruct, n*limitOfRequests)
			logUnit := logStruct{}

			//start n-amount of consurrent clients(strestesters)
			for i := 0; i < n; i++ {
				go stressTester(config, command, logUnit, logChan)
			}

			//printout result of testing
			fmt.Println(getMetrics(command, n, logChan))
			close(logChan)
		}
	}

}

//stressTester - open single connection to server, "atack" it with request-command 100-times(limitOfRequests),
//measure response time from server and store it to metrics chan.
func stressTester(config *config, command string, logUnit logStruct, logChan chan<- logStruct) {
	conn, err := net.Dial(config.protocol, config.addr)

	ifErrFatal(err)
	defer conn.Close()

	for i := 0; i < limitOfRequests; i++ {

		_, err := conn.Write(makeNetString(command))
		if err != nil {
			logUnit.err = err
			logChan <- logUnit
			continue
		}

		start := time.Now()
		connReader := bufio.NewReader(conn)

		responseLength, err := getResponseLength(connReader)
		if errHandler(err, logUnit, start, logChan) {
			continue
		}
		response := make([]byte, responseLength)

		_, err = connReader.Read(response)
		if errHandler(err, logUnit, start, logChan) {
			continue
		}

		if errHandler(checkIfResponseIsError(response), logUnit, start, logChan) {
			continue
		}

		logUnit.execTime = time.Since(start)
		logChan <- logUnit
	}
}

func errHandler(err error, logUnit logStruct, start time.Time, logChan chan<- logStruct) bool {
	if err != nil {
		logUnit.err = err
		logUnit.execTime = time.Since(start)
		logChan <- logUnit
		return true
	}

	return false
}

func checkIfResponseIsError(response []byte) error {
	if response[0] == 'E' {
		return fmt.Errorf(string(response))
	}
	return nil
}

func makeNetString(s string) []byte {
	s = s[0:len(s)]
	s = strconv.Itoa(len(s)) + ":" + s

	return []byte(s)
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
		config.addr = fmt.Sprint(":", args[1])
		return config
	}

	if len(args) == 3 {
		config.addr = fmt.Sprint(":", args[1])
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

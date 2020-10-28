package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
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

//Metrics to save time from start and response time
type Metrics struct {
	Mut                       *sync.Mutex
	responseTimeDurationSlice []time.Duration
	requestSecondSinceStart   time.Duration
}

//commands []string - List of commands to test with
var commands = []string{"set", "get", "getset", "ex"}

var stressTimeDuration = []time.Duration{time.Second * 10, time.Second * 30, time.Second * 60}

func main() {

	config := getConfig(os.Args)
	metric := &Metrics{responseTimeDurationSlice: make([]time.Duration, 0, 10000)}
	name := 0
	arg := 0
	start := time.Now()
	stopMarker := false

	go metricsPrintOut(metric, stopMarker)

	for _, stressTimeDuration := range stressTimeDuration {
		fmt.Println("ololo")
		for {

			if time.Since(start) > stressTimeDuration {

				fmt.Println("stop")
				stopMarker = true
				break
			}

			name++
			arg++

			cmdBody := "" + strconv.Itoa(name) + " " + strconv.Itoa(arg)

			for _, cmd := range commands {

				cmd = cmd + cmdBody

				go stressTester(config, cmd, metric, stopMarker)

			}
		}
	}
}

func metricsPrintOut(metric *Metrics, stopMarker bool) {
	secondSinceStart := metric.requestSecondSinceStart
	routinesCount := 0
	totalRoutinesCount := 0
	maxResponseTime := time.Duration(0)
	var secondWithMaxTimeResponse time.Duration

	for {

		if secondSinceStart != metric.requestSecondSinceStart {

			metric.Mut.Lock()
			second := metric.requestSecondSinceStart
			responseTimeSlice := metric.responseTimeDurationSlice
			routinesCount = len(responseTimeSlice)
			totalRoutinesCount += routinesCount
			metric.responseTimeDurationSlice = make([]time.Duration, 0, 10000)
			metric.Mut.Unlock()
			sum, maxResponseTimeInSecond := sum(responseTimeSlice)
			if maxResponseTime < maxResponseTimeInSecond {
				maxResponseTime = maxResponseTimeInSecond
				secondWithMaxTimeResponse = secondSinceStart
			}
			fmt.Printf("Second of stress: %v\nRoutines count: %v\nAverage response time: %v\n",
				secondSinceStart, routinesCount, time.Duration(sum/routinesCount))
			secondSinceStart = second
		}

		if stopMarker {
			break
		}
	}
	fmt.Printf("Total routines count: %v\nMax response time: %v\nSecond of stress with max response time: %v\n",
		totalRoutinesCount, maxResponseTime, secondWithMaxTimeResponse)

}

func sum(durationSlice []time.Duration) (int, time.Duration) {
	var sum, maxResponseTime time.Duration
	for _, el := range durationSlice {
		sum += el
		if el > maxResponseTime {
			maxResponseTime = el
		}
	}
	return int(sum), maxResponseTime
}

func stressTester(config *config, cmd string, metric *Metrics, stopMarker bool) {
	conn, err := net.Dial(config.protocol, config.addr)

	ifErrFatal(err)
	defer conn.Close()

	start := time.Now()

	for {

		if stopMarker {
			break
		}

		_, err := conn.Write(makeNetString(cmd))
		if err != nil {
			break
		}

		startRequest := time.Now()
		connReader := bufio.NewReader(conn)

		responseLength, err := getResponseLength(connReader)
		if err != nil {
			break
		}

		response := make([]byte, responseLength)
		_, err = connReader.Read(response)
		if err != nil {
			break
		}

		metric.Mut.Lock()
		metric.responseTimeDurationSlice = append(metric.responseTimeDurationSlice, time.Since(startRequest))
		metric.requestSecondSinceStart = time.Since(start).Truncate(time.Second)
		metric.Mut.Unlock()
	}

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

package main

import (
	"fmt"
	"sync"
	"time"
)

type Counter struct {
	i int
	l *sync.Mutex
}

func incCounter(c *Counter, n int, wg *sync.WaitGroup) {
	i := 0
	for i < n {
		i++
		c.i++
	}
}

func routine(c *Counter, n int) time.Duration {
	var wg sync.WaitGroup

	t := time.Now()
	for i := 0; i < 10; i++ {
		go incCounter(c, n, &wg)
	}

	return time.Since(t)
}

func average(t []time.Duration) time.Duration {
	var average time.Duration

	for i := range t {
		average += t[i]
	}

	return average / time.Duration(len(t))
}

func printout(t time.Duration, c Counter) {
	fmt.Print("execution time: ")
	fmt.Print(t)
	fmt.Printf("; counter value: %d\n", c.i)
}

func main() {
	var t [100]time.Duration

	c := Counter{i: 0, l: &sync.Mutex{}}

	for i := 0; i < 100; i++ {
		t[i] = routine(&c, 100000)
	}

	s := t[0:100]

	printout(average(s), c)
}

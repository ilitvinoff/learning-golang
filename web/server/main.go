package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

//Counter counts amount of visits to web server
type Counter struct {
	Mut   *sync.Mutex
	value int
}

func main() {
	// Hello world, the web server
	c := &Counter{Mut: &sync.Mutex{}, value: 0}

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, counter(c))
	}

	http.HandleFunc("/", helloHandler)
	log.Fatal(http.ListenAndServe(":80", nil))
}

func counter(c *Counter) string {
	c.Mut.Lock()
	defer c.Mut.Unlock()
	c.value++
	return strconv.Itoa(c.value)
}

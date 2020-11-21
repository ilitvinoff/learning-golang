package main

import (
	"sync"
)

func main() {
	tailConfig := &myTailConfig{getConfigFromFlags(), sync.WaitGroup{}}
	ParallelTail(tailConfig)

}

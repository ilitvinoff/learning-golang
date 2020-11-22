package main

import (
	"fmt"
	"sync"
)

func main() {
	tailConfig := &myTailConfig{getConfigFromFlags(), sync.WaitGroup{}}
	if len(tailConfig.configArr) < 1 {
		fmt.Println("Use flags to tweak the config you want. See help message. Enter: \n./tail -h")
	}
	ConcurentTail(tailConfig)

}

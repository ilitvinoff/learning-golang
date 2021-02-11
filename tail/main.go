package main

import (
	"fmt"
	"sync"
)

func main() {
	tailState := &TailState{getConfigFromFlags(), userWatchPollDellay, sync.WaitGroup{}}

	if len(tailState.configArr) < 1 {
		fmt.Println("Use flags to tweak the config you want. See help message. Enter: \n./tail -h")
	}
	fmt.Println("isDebug:", isDebug)
	ConcurentTail(tailState)

}

package main

import "os"

//channel buffer size
const (
	channelBufferLimit = 1000
)

func main() {
	var configuration argsConfig

	//Set config
	configuration.setArgsConfig(os.Args)
	//Create working directories
	createDestDirs(&configuration)

	//Create channels to comunicate between routines
	channels := channels{
		urls:            make(chan string, channelBufferLimit),
		downloadedFiles: make(chan downloadedFile, channelBufferLimit),
		resizedFiles:    make(chan bool, channelBufferLimit),
		urlsCount:       make(chan int),
	}

	//Read urls from file and push them to next steps
	go pusher(configuration, channels)

	//downloading routines start
	limit := int(configuration.downloadRoutinesAmount)
	for i := 0; i < limit; i++ {
		go downloader(&configuration, channels)
	}

	//resizing routines start
	limit = int(configuration.downloadRoutinesAmount)
	for i := 0; i < limit; i++ {
		go resizer(&configuration, channels)
	}

	//How many urls was read from file
	processedCount := <-channels.urlsCount

	//ensure that all routines have finished their tasks
	for i := 0; i < processedCount; i++ {
		<-channels.resizedFiles
	}

	channels.closeChannels()
}

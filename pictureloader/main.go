package main

//Uploads a pool of files by http(from urls) taken from a local text file.
//Saves them to the specified folder.
//Compresses the size of images proportionally to the specified by user value in pixels (width)
//and saves to the "avatars" folder.
//The folder "avatars" is created automatically in the folder specified by the user
//for storing uploaded files.
//
//To start the program, you need to specify 5 arguments:
// 1 - path to local file with urls
// 2 - path for storing downloaded files (if the folder does not exist,
//     it will be created at the specified path)
// 3 - the value of the width in pixels for the compressed images
// 4 - number of threads for parallel downloading of files (n> = 1)
// 5 - number of threads for parallel compression of several downloaded images (n> = 1)

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

package main

import (
	"bufio"
	"os"
)

//pusher - read file line by line (each line of th file must be url)
//and push each line to channels.urls to communicate with downloader's routines.
func pusher(configuration argsConfig, channels channels) {
	file, err := os.Open(configuration.urlsFilePath)
	logFatal(err)

	defer func() {
		err := file.Close()
		logFatal(err)
	}()

	scanner := bufio.NewScanner(file)
	urlsCounter := 0 // counter counts how many urls we read from file

	for scanner.Scan() {
		channels.urls <- scanner.Text()
		urlsCounter++
	}

	channels.urlsCount <- urlsCounter

	err = scanner.Err()
	logFatal(err)
}

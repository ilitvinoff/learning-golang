package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

//downloader - download images from url.
//args:
// configuration *argsConfig - program args entered by user
// channels channels - channels for communication between routines
func downloader(configuration *argsConfig, channels channels) {

	for url := range channels.urls {

		filePath, err := getFileNameFromPath(configuration, url, true)
		//if no valid name available, skip iteration
		if err != nil {
			log.Println(err)
			channels.downloadedFiles <- downloadedFile{path: "", successMarker: false}
			continue
		}

		response, err := http.Get(url)

		//If can't get response from url - create emty file for this url
		if err != nil {
			destFile, err := os.Create(filePath)

			if err != nil {
				log.Print(err)
				channels.downloadedFiles <- downloadedFile{path: "", successMarker: false}
				continue
			}

			err = destFile.Close()
			logErr(err)

			//send file for resizing
			channels.downloadedFiles <- downloadedFile{path: filePath, successMarker: true}
			continue
		}

		destFile, err := os.Create(filePath)
		if err != nil {
			log.Print(err)
			channels.downloadedFiles <- downloadedFile{path: "", successMarker: false}
			continue
		}

		_, err = io.Copy(destFile, response.Body)
		logErr(err)
		err = response.Body.Close()
		logErr(err)

		err = destFile.Close()
		logErr(err)

		//send file for resizing
		channels.downloadedFiles <- downloadedFile{path: filePath, successMarker: true}
	}
}

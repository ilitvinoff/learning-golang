package main

import (
	"log"
	"os"
)

const (
	avatarDirName = "avatars"
)

//argsConfig - struct to store program args
type argsConfig struct {
	urlsFilePath           string //path to file with urls list
	destDir                string //path yo destination directory
	destAvatarDir          string //path yo destination directory
	newPictureWidth        uint   //desired image width to resize for
	downloadRoutinesAmount uint   //amount of routines to download images
	resizeRoutinesAmount   uint   //amount of routines to resize image
}

//proccesedEntity - struct to store file path and success of operation
type downloadedFile struct {
	path          string //file path (url, downloaded, resized)
	successMarker bool   //if read/downloade/resized successfully - true; else - false
}

//channels - struct to store channels for routines communicating
type channels struct {
	urls            chan string         //urls readed from file
	downloadedFiles chan downloadedFile //downloaded file's names
	resizedFiles    chan bool           //indicate amount of resized files
	urlsCount       chan int            //count of urls pushed for processing
}

//setArgsConfig - validating program args and set it to config obj
func (configuration *argsConfig) setArgsConfig(args []string) {
	var err error

	configuration.urlsFilePath, err = pathValidator(args[1])
	logFatal(err)

	configuration.destDir, err = pathValidator(args[2])
	logFatal(err)

	configuration.destAvatarDir = configuration.destDir + "/" + avatarDirName

	configuration.newPictureWidth, err = uintValidator(args[3])
	logFatal(err)

	configuration.downloadRoutinesAmount, err = uintValidator(args[4])
	logFatal(err)

	configuration.resizeRoutinesAmount, err = uintValidator(args[5])
	logFatal(err)
}

func logErr(err error) {
	if err != nil {
		log.Print(err.Error())
	}
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func createDestDirs(configuration *argsConfig) {
	//Creat destination dir for download files, if it doesn't exist
	err := os.MkdirAll(configuration.destDir, 0744)
	logFatal(err)

	//Creat destination dir for resized images, if it doesn't exist
	err = os.MkdirAll(configuration.destAvatarDir, 0744)
	logFatal(err)
}

func (channels *channels) closeChannels() {
	close(channels.urls)
	close(channels.downloadedFiles)
	close(channels.resizedFiles)
	close(channels.urlsCount)
}

package main

import (
	"errors"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/nfnt/resize"
)

var mimeMap = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true}

const (
	jpegMime             = "image/jpeg"
	pngMime              = "image/png"
	gifMime              = "image/gif"
	jpegDecodeErrMessage = "Something went wrong with decoding jpeg file:\n"
	pngDecodeErrMessage  = "Something went wrong with decoding png file:\n"
	gifDecodeErrMessage  = "Something went wrong with decoding gif file:\n"
)

func resizer(configuration *argsConfig, channels channels) {

	for downloadedFile := range channels.downloadedFiles {

		//if download operatin was unsuccesfull, then skip resizing
		if !downloadedFile.successMarker {
			channels.resizedFiles <- false
			continue
		}

		mimeType := detectFilesMime(downloadedFile.path)

		if mimeMap[mimeType] {
			file, err := os.Open(downloadedFile.path)
			//if can't to open, then create empty avatar file
			if err != nil {
				log.Println(err)
				goto emptyFileCreate
			}

			switch mimeType {
			case jpegMime:
				jpegCreate(file, configuration)
			case pngMime:
				pngCreate(file, configuration)
			default:
				gifCreate(file, configuration)
			}

			channels.resizedFiles <- true
			continue
		}

	emptyFileCreate:
		filePathAvatar, err := getFileNameFromPath(configuration, downloadedFile.path, false)
		//if no valid name available, skip iteration
		if err != nil {
			log.Println(err)
			channels.resizedFiles <- false
			continue
		}

		dest, err := os.Create(filePathAvatar)
		logErr(err)
		dest.Close()

		channels.resizedFiles <- true

	}
}

func detectFilesMime(filePath string) string {
	buff, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Print(err)
	}
	return http.DetectContentType(buff)
}

func jpegCreate(file *os.File, configuration *argsConfig) error {

	// decode file into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		return errors.New(jpegDecodeErrMessage + file.Name())
	}

	// resize to desired width using Lanczos resampling
	// and preserve aspect ratio
	imgRes := resize.Resize(configuration.newPictureWidth, 0, img, resize.Lanczos3)

	//Choose file name for our resized image
	filePath, err := getFileNameFromPath(configuration, file.Name(), false)
	//if no valid name available, return error
	if err != nil {
		return err
	}

	dest, err := os.Create(filePath)
	if err != nil {
		return err
	}
	file.Close()
	defer dest.Close()

	// write new image to file
	jpeg.Encode(dest, imgRes, nil)
	fmt.Println("jpeg: ", dest.Name())

	return nil
}

func pngCreate(file *os.File, configuration *argsConfig) error {

	// decode file into image.Image
	img, err := png.Decode(file)
	if err != nil {
		return errors.New(pngDecodeErrMessage + file.Name())
	}

	// resize to desired width using Lanczos resampling
	// and preserve aspect ratio
	imgRes := resize.Resize(configuration.newPictureWidth, 0, img, resize.Lanczos3)

	//Choose file name for our resized image
	filePath, err := getFileNameFromPath(configuration, file.Name(), false)
	//if no valid name available, return error
	if err != nil {
		return err
	}

	dest, err := os.Create(filePath)
	if err != nil {
		return err
	}
	file.Close()
	defer dest.Close()

	// write new image to file
	png.Encode(dest, imgRes)
	fmt.Println("png: ", dest.Name())

	return nil
}

func gifCreate(file *os.File, configuration *argsConfig) error {

	// decode file into image.Image
	img, err := gif.Decode(file)
	if err != nil {
		return errors.New(gifDecodeErrMessage + file.Name())
	}

	// resize to desired width using Lanczos resampling
	// and preserve aspect ratio
	imgRes := resize.Resize(configuration.newPictureWidth, 0, img, resize.Lanczos3)

	//Choose file name for our resized image
	filePath, err := getFileNameFromPath(configuration, file.Name(), false)
	//if no valid name available, return error
	if err != nil {
		return err
	}

	dest, err := os.Create(filePath)
	if err != nil {
		return err
	}
	file.Close()
	defer dest.Close()

	// write new image to file
	gif.Encode(dest, imgRes, nil)
	fmt.Println("gif: ", dest.Name())

	return nil
}

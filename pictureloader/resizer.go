package main

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/nfnt/resize"
)

const (
	jpegMime         = "jpeg"
	pngMime          = "png"
	decodeErrMessage = "Something went wrong with decoding file: "
)

func resizer(configuration *argsConfig, channels channels) {

	for downloadedFile := range channels.downloadedFiles {

		//if download operatin was unsuccesfull, then skip resizing
		if !downloadedFile.successMarker {
			channels.resizedFiles <- false
			continue
		}

		file, err := os.Open(downloadedFile.path)
		if err != nil {
			logErr(err)
			err = creatEmptyAvatar(downloadedFile.path, configuration, channels)
			logErr(err)
			continue
		}

		img, format, err := image.Decode(file)
		if err != nil {
			logErr(errors.New(decodeErrMessage + downloadedFile.path))
			err = creatEmptyAvatar(downloadedFile.path, configuration, channels)
			logErr(err)
			continue
		}

		fileName := file.Name()
		err = file.Close()
		logErr(err)

		switch format {
		case jpegMime:
			err = jpegCreate(img, fileName, configuration)
		case pngMime:
			err = pngCreate(img, fileName, configuration)
		default:
			err = gifCreate(img, fileName, configuration)
		}

		if err != nil {
			logErr(err)
			channels.resizedFiles <- false
			continue
		}
		channels.resizedFiles <- true

	}
}

func creatEmptyAvatar(downloadedFilePath string, configuration *argsConfig, channels channels) error {
	filePathAvatar, err := getFileNameFromPath(configuration, downloadedFilePath, false)
	//if no valid name available, skip iteration
	if err != nil {
		channels.resizedFiles <- false
		return err
	}

	dest, err := os.Create(filePathAvatar)
	if err != nil {
		channels.resizedFiles <- false
		return err
	}
	dest.Close()

	channels.resizedFiles <- true
	return nil
}

func jpegCreate(img image.Image, fileName string, configuration *argsConfig) error {
	imgRes, destFile, err := resizeImage(img, fileName, configuration)
	if err != nil {
		return err
	}

	// write new image to file
	jpeg.Encode(destFile, imgRes, nil)
	fmt.Println("jpeg: ", destFile.Name())

	return nil
}

func pngCreate(img image.Image, fileName string, configuration *argsConfig) error {
	imgRes, destFile, err := resizeImage(img, fileName, configuration)
	if err != nil {
		return err
	}

	// write new image to file
	png.Encode(destFile, imgRes)
	fmt.Println("png: ", destFile.Name())

	return nil
}

func gifCreate(img image.Image, fileName string, configuration *argsConfig) error {
	imgRes, destFile, err := resizeImage(img, fileName, configuration)
	if err != nil {
		return err
	}
	// write new image to file
	gif.Encode(destFile, imgRes, nil)
	fmt.Println("gif: ", destFile.Name())

	return nil
}

func resizeImage(img image.Image, fileName string, configuration *argsConfig) (image.Image, *os.File, error) {
	// resize to desired width using Lanczos resampling
	// and preserve aspect ratio
	imgRes := resize.Resize(configuration.newPictureWidth, 0, img, resize.Lanczos3)

	//Choose file name for our resized image
	filePath, err := getFileNameFromPath(configuration, fileName, false)
	//if no valid name available, return error
	if err != nil {
		return nil, nil, err
	}

	dest, err := os.Create(filePath)
	if err != nil {
		return nil, nil, err
	}

	defer dest.Close()
	return imgRes, dest, nil
}

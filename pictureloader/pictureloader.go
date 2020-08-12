// pictureloader.go
package main

import (
	"bufio"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

var mimeMap = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true}

const (
	pathDirOrigImages = "originals/"
	nameDirOrigImages = "originals"
	jpegMime          = "image/jpeg"
	pngMime           = "image/png"
	gifMime           = "image/gif"
)

//Read file's content. Each line - is url. Return slice of urls.
func readFromFile(filePath string) (res []string) {

	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		fmt.Print("Opening file error: ")
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		res = append(res, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Print("Reading file error:")
		fmt.Println(err)
	}

	return
}

//Download files from different sources in parallel routines.
//urls - array with src adresses
//routineAmount - routines amount
func downloadFiles(urls []string, routinesAmount uint, wg *sync.WaitGroup) {

	tokens := make(chan struct{}, routinesAmount)
	for _, url := range urls {
		go downloadFile(url, tokens, wg)
	}

	wg.Wait()
	return
}

//Download singe file.
//url - src download from
//tokens - chan to control amount of goroutines
// wg - to sure that all routines has executed
func downloadFile(url string, tokens chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	tokens <- struct{}{}

	filePath := pathDirOrigImages + getFileNameFromPath(url)

	resp, err := http.Get(url)

	if err != nil {
		_, err := os.Create(filePath)

		panicErr(err)
		return
	}
	defer resp.Body.Close()

	destFile, err := os.Create(filePath)

	panicErr(err)
	defer destFile.Close()

	_, err = io.Copy(destFile, resp.Body)
	panicErr(err)
	<-tokens

}

//Return the file name from url(last part in url, after last "/")
func getFileNameFromPath(url string) string {

	res := strings.Split(url, "/")
	return (res[len(res)-1])
}

func creatFolder(name string) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		err1 := os.Mkdir(name, 0755)
		if err1 != nil {
			panic(err1)
		}
	}
}

func creatIcons(iconsDirPath string, width uint, routinesAmount uint, wg *sync.WaitGroup) {
	filesList, err := ioutil.ReadDir(nameDirOrigImages)
	panicErr(err)

	tokens := make(chan struct{}, routinesAmount)
	for _, fileInfo := range filesList {
		go creatIcon(fileInfo, iconsDirPath, width, tokens, wg)
	}
	wg.Wait()

}

func creatIcon(fileInfo os.FileInfo, iconsDirPath string, width uint, tokens chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	tokens <- struct{}{}

	filePath := pathDirOrigImages + fileInfo.Name()
	writeIconToFile(filePath, iconsDirPath, width)

	<-tokens
}

func writeIconToFile(filePath string, iconsDirPath string, width uint) {
	mimeType := detectFilesMime(filePath)

	if mimeMap[mimeType] {
		file := openFile(filePath)

		switch mimeType {
		case jpegMime:
			jpegCreate(file, width, iconsDirPath)
		case pngMime:
			pngCreate(file, width, iconsDirPath)
		default:
			gifCreate(file, width, iconsDirPath)
		}

	} else {
		dest, err := os.Create(iconsDirPath + getFileNameFromPath(filePath))
		panicErr(err)
		dest.Close()
	}
}

func jpegCreate(file *os.File, width uint, iconsDirPath string) {

	// decode file into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Println("file: " + file.Name())
		panic(err)
	}
	fileInfo, err := file.Stat()
	file.Close()

	// resize to desired width using Lanczos resampling
	// and preserve aspect ratio
	imgRes := resize.Resize(width, 0, img, resize.Lanczos3)

	dest, err := os.Create(iconsDirPath + fileInfo.Name())
	panicErr(err)
	defer dest.Close()

	// write new image to file
	jpeg.Encode(dest, imgRes, nil)
	fmt.Println(dest.Name())
}

func pngCreate(file *os.File, width uint, iconsDirPath string) {

	// decode file into image.Image
	img, err := png.Decode(file)
	if err != nil {
		fmt.Println("file: " + file.Name())
		panic(err)
	}
	fileInfo, err := file.Stat()
	file.Close()

	// resize to desired width using Lanczos resampling
	// and preserve aspect ratio
	imgRes := resize.Resize(width, 0, img, resize.Lanczos3)

	dest, err := os.Create(iconsDirPath + fileInfo.Name())
	panicErr(err)
	defer dest.Close()

	// write new image to file
	png.Encode(dest, imgRes)
	fmt.Println(dest.Name())
}

func gifCreate(file *os.File, width uint, iconsDirPath string) {

	// decode file into image.Image
	img, err := gif.Decode(file)
	if err != nil {
		fmt.Println("file: " + file.Name())
		panic(err)
	}
	fileInfo, err := file.Stat()
	file.Close()

	// resize to desired width using Lanczos resampling
	// and preserve aspect ratio
	imgRes := resize.Resize(width, 0, img, resize.Lanczos3)

	dest, err := os.Create(iconsDirPath + fileInfo.Name())
	panicErr(err)
	defer dest.Close()

	// write new image to file
	gif.Encode(dest, imgRes, nil)
	fmt.Println(dest.Name())
}

func openFile(filePath string) *os.File {
	file, err := os.Open(filePath)
	panicErr(err)
	return file
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func detectFilesMime(filePath string) string {
	buff, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Print(err)
	}
	return http.DetectContentType(buff)
}

func main() {
	var wg sync.WaitGroup
	var urls []string
	args := os.Args

	dwldRoutinesAmount, err := strconv.ParseUint(args[4], 10, 64)
	panicErr(err)

	width, err := strconv.ParseUint(args[3], 10, 64)
	panicErr(err)

	iconsRoutinesAmount, err := strconv.ParseUint(args[5], 10, 64)
	panicErr(err)

	iconsDirPath := args[2] + "/"

	creatFolder(nameDirOrigImages)
	creatFolder(args[2])

	urls = append(urls[0:], readFromFile(args[1])...)

	downloadFiles(urls, uint(dwldRoutinesAmount), &wg)
	creatIcons(iconsDirPath, uint(width), uint(iconsRoutinesAmount), &wg)

}

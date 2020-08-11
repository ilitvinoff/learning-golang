// pictureloader.go
package main

import (
	"bufio"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

const (
	originalImagesDirPath = "originals/"
	originalImagesDirName = "originals"
	iconsDirPath          = "icons/"
	iconsDirName          = "icons"
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
func downloadFiles(urls []string, routinesAmount int, wg *sync.WaitGroup) {

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

	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	filePath := originalImagesDirPath + getFileNameFromURL(url)

	destFile, err := os.Create(filePath)

	if err != nil {
		panic(err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		panic(err)
	}
	<-tokens

}

//Return the file name from url(last part in url, after last "/")
func getFileNameFromURL(url string) string {

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

func creatIcons(routinesAmount int, wg *sync.WaitGroup) {
	files, err := ioutil.ReadDir(originalImagesDirName)
	if err != nil {
		panic(err)
	}

	tokens := make(chan struct{}, routinesAmount)
	for _, file := range files {
		go creatIcon(file, tokens, wg)
	}
	wg.Wait()

}

func creatIcon(fileInfo os.FileInfo, tokens chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	tokens <- struct{}{}

	file, err := os.Open(originalImagesDirPath + fileInfo.Name())
	if err != nil {
		panic(err)
	}

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Printf("file: %s", file.Name())
		panic(err)
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(64, 0, img, resize.Lanczos3)

	dest, err := os.Create(iconsDirPath + fileInfo.Name())
	if err != nil {
		panic(err)
	}
	defer dest.Close()

	// write new image to file
	jpeg.Encode(dest, m, nil)
	fmt.Println(dest.Name())
	<-tokens
}

func main() {
	var wg sync.WaitGroup
	var urls []string

	creatFolder(originalImagesDirName)
	creatFolder(iconsDirName)
	urls = append(urls[0:], readFromFile("urls.txt")...)
	downloadFiles(urls, 15, &wg)
	creatIcons(4, &wg)

}

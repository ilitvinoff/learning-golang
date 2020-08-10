// pictureloader.go
package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	tempPath       = "temp/"
	tempFolderName = "temp"
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
func downloadFiles(urls []string, routineAmount int, wg *sync.WaitGroup) {

	tokens := make(chan struct{}, routineAmount)
	for _, url := range urls {
		go downloadFile(url, tokens, wg)
	}

	wg.Wait()
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

	destFile, err := os.Create(tempPath + getFileNameFromURL(url))

	if err != nil {
		panic(err)
	}
	defer destFile.Close()

	fmt.Println(getFileNameFromURL(url))

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

func main() {
	var wg sync.WaitGroup
	var urls []string
	creatFolder(tempFolderName)
	urls = append(urls[0:], readFromFile("urls.txt")...)
	fmt.Println(getFileNameFromURL(urls[2]))
	downloadFiles(urls, 20, &wg)
}

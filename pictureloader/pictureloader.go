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

//Read file's content. Each line - is url. Return slice of urls.
func readFromFile(filePath string) (res []string) {

	file, err := os.Open(filePath)

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

func downloadFiles(urls []string, routineAmount int) {

	var wg sync.WaitGroup
	tokens := make(chan struct{}, routineAmount)
	for _, url := range urls {
		go downloadFile(url, tokens, &wg)
	}
	wg.Wait()
}

func downloadFile(url string, tokens chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	tokens <- struct{}{}

	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		panic(err)
	}

	destFile, err := os.Create(getFileNameFromURL(url))
	defer destFile.Close()

	if err != nil {
		panic(err)
	}

	_, err = io.Copy(destFile, resp.Body)

	if err != nil {
		panic(err)
	}

	<-tokens

}

func getFileNameFromURL(url string) string {

	res := strings.Split(url, "/")
	return res[len(res)-1]
}

func main() {
	var urls []string
	urls = append(urls[0:], readFromFile("/home/legion/GoProj/pictureloader/urls.txt")...)
	downloadFiles(urls, 20)
}

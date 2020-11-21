package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/hpcloud/tail"
	"gopkg.in/djherbis/times.v1"
)

const (
	watchPollDelay = 100 * time.Millisecond
)

type myTailConfig struct {
	configArr []*config
	wg        sync.WaitGroup
}

//ParallelTail ...
func ParallelTail(t *myTailConfig) {
	for _, cfg := range t.configArr {
		t.wg.Add(1)
		go masterTail(cfg, t)
	}

	t.wg.Wait()
}

func masterTail(config *config, t *myTailConfig) {
	defer t.wg.Done()
	TailF(config)

}

//TailF ...
func TailF(config *config) {
	for {
		err := waitForPathExists(config, config.isFilepath)
		logFatalIfError(err)

		filepath := config.path

		if !config.isFilepath {
			filepath, err = getPathWithRegex(config)
			logFatalIfError(err)
		}

		setCursorPos(config, filepath)
		w := initiateWatcher(config, !config.isFilepath)
		t := getTailer(filepath, *config.tailConfig)

		go startWatcher(config, w, t)

		for line := range t.Lines {
			fmt.Println(config.prefix, line.Text)
		}
		t.Cleanup()
	}
}

func setCursorPos(c *config, filepath string) {
	if c.readFromBeginning {
		c.tailConfig.Location = &tail.SeekInfo{Offset: 0, Whence: 0}

		fmt.Println("file:", filepath, "n:", c.n, "seek", c.tailConfig.Location)
		return
	}

	file, err := os.Open(filepath)
	logFatalIfError(err)
	defer file.Close()

	stats, err := file.Stat()
	logFatalIfError(err)

	filesize := stats.Size()
	lineCounter := 0
	var cursor int64

	for cursor = 0; math.Abs(float64(cursor)) < float64(filesize) && lineCounter < c.n; {
		cursor--
		file.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		file.Read(char)

		if cursor != -1 && char[0] == '\n' { // stop if we find a line
			lineCounter++
		}

	}

	c.tailConfig.Location = &tail.SeekInfo{Offset: cursor + 1, Whence: 2}
	fmt.Println("file:", filepath, "n:", c.n, "seek", c.tailConfig.Location)

}

func getTailer(path string, tailConfig tail.Config) *tail.Tail {
	t, err := tail.TailFile(path, tailConfig)
	logFatalIfError(err)
	return t
}

func logFatalIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func waitForPathExists(config *config, isFilepath bool) error {
START:
	dir, err := os.Stat(config.path)
	if err != nil {
		if os.IsNotExist(err) {
			time.Sleep(watchPollDelay)
			goto START
		}
		return err
	}

	if !isFilepath && !dir.IsDir() {
		return fmt.Errorf("given path is not directory! path: %v", config.path)
	}

	return nil
}

func getPathWithRegex(config *config) (string, error) {
	for {
		candidates, err := getFilesFromDir(config.path)
		if err != nil {
			return "", err
		}

		matchedCandidates, err := getMatchedRegexFiles(candidates, config.regex)
		if err != nil {
			return "", err
		}

		if len(matchedCandidates) == 0 {
			continue
		}

		res := matchedCandidates[0]

		for _, file := range matchedCandidates {
			res = compareWhichYonger(res, file, config)
		}

		return filepath.Join(config.path, res.Name()), nil
	}
}

func getFilesFromDir(path string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(path)
	res := []os.FileInfo{}
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		res = append(res, file)
	}
	return res, nil
}

func getMatchedRegexFiles(files []os.FileInfo, regex string) ([]os.FileInfo, error) {
	matchedCandidates := []os.FileInfo{}
	for _, file := range files {
		matched, err := regexp.MatchString(regex, file.Name())
		if err != nil {
			return nil, err
		}

		if matched {
			matchedCandidates = append(matchedCandidates, file)
		}
	}
	return matchedCandidates, nil
}

func compareWhichYonger(file1 os.FileInfo, file2 os.FileInfo, c *config) os.FileInfo {
	filepath1 := filepath.Join(c.path, file1.Name())
	filepath2 := filepath.Join(c.path, file2.Name())

	t1, err := times.Stat(filepath1)
	logFatalIfError(err)

	t2, err := times.Stat(filepath2)
	logFatalIfError(err)

	if t1.ChangeTime().Before(t2.ChangeTime()) {
		return file2
	}

	return file1
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

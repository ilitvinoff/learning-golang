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
	"time"

	"github.com/hpcloud/tail"
	"gopkg.in/djherbis/times.v1"
)

const (
	fileReadBufferSize = 1023
)

//TailState ...
type TailState struct {
	configArr      []*config
	watchPollDelay time.Duration
	wg             sync.WaitGroup
}

//ConcurentTail ...
func ConcurentTail(tailState *TailState) {
	for _, cfg := range tailState.configArr {
		tailState.wg.Add(1)
		go masterTail(cfg, tailState.watchPollDelay, &tailState.wg)
	}

	tailState.wg.Wait()
}

func masterTail(config *config, watchPollDelay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	TailF(config, watchPollDelay)

}

//TailF ...
func TailF(config *config, watchPollDelay time.Duration) {
	for {
		err := waitForPathExists(config, watchPollDelay)
		logFatalIfError(err)

		filepath := config.path

		if !config.isFilepath {
			filepath, err = getYongestFilepathMatchedRegex(config, watchPollDelay)
			logFatalIfError(err)
		}

		setCursorPos(config, filepath)

		w := initWatcher(config)
		tail := getTailer(filepath, *config.hpcloudTailCfg)

		go startWatcher(config, w, tail, watchPollDelay)

		for line := range tail.Lines {
			fmt.Println(config.messagePrefix, line.Text)
		}
		tail.Cleanup()
	}
}

//setCursorPos - set cursor position in file to tail from
func setCursorPos(cfg *config, filepath string) {
	if cfg.readFromBeginning {
		cfg.hpcloudTailCfg.Location = &tail.SeekInfo{Offset: 0, Whence: 0}
		return
	}

	file, err := os.Open(filepath)
	logFatalIfError(err)
	defer file.Close()

	stats, err := file.Stat()
	logFatalIfError(err)

	filesize := stats.Size()
	lineCounter := 0

	var currentPosition, positionToReadFrom int64

	if filesize > fileReadBufferSize {
		positionToReadFrom = filesize - fileReadBufferSize
	}

	for currentPosition = -1; math.Abs(float64(currentPosition)) < float64(filesize)-1 && lineCounter < cfg.n; {
		_, err := file.Seek(positionToReadFrom, io.SeekStart)
		logFatalIfError(err)

		char := make([]byte, fileReadBufferSize)

		n, err := file.Read(char)
		if err != nil && err != io.EOF {
			log.Fatal(err.Error())
		}

		for i := n - 1; i > 0 && lineCounter < cfg.n; i-- {
			if currentPosition != -1 && char[i] == '\n' { // stop if we find a line
				lineCounter++
			}
			currentPosition--
		}

		positionToReadFrom = positionToReadFrom - filesize
		if positionToReadFrom < 0 {
			positionToReadFrom = 0
		}
	}

	cfg.hpcloudTailCfg.Location = &tail.SeekInfo{Offset: currentPosition, Whence: 2}
}

func getTailer(path string, hpcloudTailCfg tail.Config) *tail.Tail {
	t, err := tail.TailFile(path, hpcloudTailCfg)
	logFatalIfError(err)
	return t
}

func logFatalIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func waitForPathExists(config *config, watchPollDelay time.Duration) error {
	var dir os.FileInfo
	var err error

	for {
		dir, err = os.Stat(config.path)

		if err != nil {

			if os.IsNotExist(err) {
				time.Sleep(watchPollDelay)
				continue
			}

			return err
		}
		break
	}
	if !config.isFilepath && !dir.IsDir() {
		return fmt.Errorf("given path is not directory! path: %v", config.path)
	}

	config.messagePrefix = config.userPrefix

	return nil
}

func getYongestFilepathMatchedRegex(config *config, watchPollDelay time.Duration) (string, error) {
	for {
		candidates, err := getFileListFromDir(config.path)
		if err != nil {
			return "", err
		}

		matchedCandidates, err := getMatchedRegexFilesFromList(candidates, config.regex)
		if err != nil {
			return "", err
		}

		if len(matchedCandidates) == 0 {
			time.Sleep(watchPollDelay)
			continue
		}

		res := matchedCandidates[0]

		for _, file := range matchedCandidates {
			res = compareWhichYonger(res, file, config)
		}

		config.messagePrefix = config.userPrefix + " {filename: " + res.Name() + "}"

		return filepath.Join(config.path, res.Name()), nil
	}
}

func getFileListFromDir(path string) ([]os.FileInfo, error) {
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

func getMatchedRegexFilesFromList(files []os.FileInfo, regex string) ([]os.FileInfo, error) {
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

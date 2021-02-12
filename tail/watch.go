package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/hpcloud/tail"
	"github.com/radovskyb/watcher"
)

func initWatcher(config *config) *watcher.Watcher {
	w := watcher.New()

	// Only notify rename, move, create, remove events.
	w.FilterOps(watcher.Rename, watcher.Move, watcher.Create, watcher.Remove, watcher.Write)

	// Only files that match the regular expression during file listings
	// will be watched.
	if !config.isFilepath {
		r := regexp.MustCompile(config.regex)
		w.AddFilterHook(watcher.RegexFilterHook(r, false))
	}

	// Watch this file/folder for changes.
	err := w.Add(config.path)
	logFatalIfError(err)

	return w
}

<<<<<<< HEAD
func startWatcher(config *config, filepath string, w *watcher.Watcher, t *tail.Tail, watchPollDelay time.Duration) {

	// Start the watching process - it'll check for changes periodically (default 100ms).
	err := w.Start(watchPollDelay)
	logFatalIfError(err)
=======
func startWatcher(config *config, w *watcher.Watcher, t *tail.Tail, watchPollDelay time.Duration) {

	go eventsHandler(w, t, config)

	// Start the watching process - it'll check for changes periodically (default 100ms).
	if err := w.Start(watchPollDelay); err != nil {
		log.Fatalln(err)
	}
>>>>>>> a3fab59cdabc7435959b3fce756f4251b40be15b
}

func eventsHandler(filepath string, w *watcher.Watcher, tail *tail.Tail, cfg *config) {
	previousSize := getFileSize(filepath)
	var err error

	for {
		//time.Sleep(userWatchPollDellay)

		select {
<<<<<<< HEAD

		case e := <-w.Event:

			if e.Op == 1 {
				previousSize, err = fileSizeController(filepath, previousSize)

				if err != nil {
					ifDebugPrintMsg(fmt.Sprintln(" \nEVENT:", cfg.messagePrefix, "{", "file:", e.Path, "; event:", e.Op.String(), "}"))
					stopWatcher(tail, cfg, w)
				}

			} else {

				ifDebugPrintMsg(fmt.Sprintln(" \nEVENT:", cfg.messagePrefix, "{", "file:", e.Path, "; event:", e.Op.String(), "}"))
				stopWatcher(tail, cfg, w)
			}
=======
		case e := <-w.Event:
			if isDebug {
				log.Println(" Event:", c.messagePrefix, e)
			}

			err := t.Stop()
			logFatalIfError(err)
			c.readFromBeginning = true
			w.Close()
>>>>>>> a3fab59cdabc7435959b3fce756f4251b40be15b

		case _, _ = <-w.Closed:
			return

		case err := <-w.Error:

			ifDebugPrintMsg(fmt.Sprintln(" \nERR:", cfg.messagePrefix, "path:", cfg.path, "; error:", err.Error()))

			if err != watcher.ErrWatchedFileDeleted {
<<<<<<< HEAD
				log.Fatalln("ERR:", cfg.messagePrefix, err)
=======
				log.Fatalln("Err:", c.messagePrefix, err)
>>>>>>> a3fab59cdabc7435959b3fce756f4251b40be15b
			}

			stopWatcher(tail, cfg, w)
		}
	}
}

func getFileSize(filepath string) int64 {
	stats, err := os.Stat(filepath)
	logFatalIfError(err)
	return stats.Size()
}

func fileSizeController(path string, previousSize int64) (int64, error) {
	currentSize := getFileSize(path)

	if currentSize-previousSize < 0 {
		ifDebugPrintMsg(fmt.Sprintln("\nCurrent filesize less than previous filesize.", "{", "path:", path, "; previous size:", previousSize, "current size:", currentSize, "}"))
		return 0, fmt.Errorf("current filesize less than previous filesize")
	}

	return currentSize, nil
}

func stopWatcher(t *tail.Tail, c *config, w *watcher.Watcher) {
	t.Cleanup()
	err := t.Stop()
	logFatalIfError(err)
	//To start read from the beggining of the file - set cursor position to start.
	c.hpcloudTailCfg.Location = &tail.SeekInfo{Offset: 0, Whence: 0}
	w.Close()
}

func ifDebugPrintMsg(msg string) {
	if isDebug {
		fmt.Println("--------------------------")
		log.Println(msg)
		fmt.Println("--------------------------")
	}
}

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

	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	w.SetMaxEvents(1)

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

func startWatcher(config *config, filepath string, w *watcher.Watcher, t *tail.Tail, watchPollDelay time.Duration) {

	go eventsHandler(filepath, w, t, config)

	// Start the watching process - it'll check for changes periodically (default 100ms).
	if err := w.Start(watchPollDelay); err != nil {
		log.Fatalln(err)
	}
}

func eventsHandler(filepath string, w *watcher.Watcher, tail *tail.Tail, cfg *config) {
	previousSize := getFileSize(filepath)
	var err error

	for {
		time.Sleep(userWatchPollDellay / 2)

		select {

		case e := <-w.Event:

			if e.Op == 1 {
				previousSize, err = fileSizeController(filepath, previousSize)

				if err != nil {
					stopWatcher(tail, cfg, w)
				}

			} else {

				if isDebug {
					fmt.Println("--------------------------")
					log.Println(" \nEVENT:", cfg.messagePrefix, "{", "file:", e.Path, "; event:", e.Op.String(), "}")
					fmt.Println("--------------------------")
				}
				stopWatcher(tail, cfg, w)
			}

		case _, _ = <-w.Closed:
			return

		case err := <-w.Error:
			if isDebug {
				fmt.Println("--------------------------")
				log.Println(" ERR:", cfg.messagePrefix, "path:", cfg.path, "; error:", err.Error())
				fmt.Println("--------------------------")
			}

			if err != watcher.ErrWatchedFileDeleted {
				log.Fatalln("ERR:", cfg.messagePrefix, err)
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
		if isDebug {
			fmt.Println("--------------------------")
			log.Println("Current filesize less than previous filesize.", "{", "path:", path, "; previous size:", previousSize, "current size:", currentSize, "}")
			fmt.Println("--------------------------")
		}

		return 0, fmt.Errorf("current filesize less than previous filesize")
	}

	return currentSize, nil
}

func stopWatcher(t *tail.Tail, c *config, w *watcher.Watcher) {
	err := t.Stop()
	logFatalIfError(err)
	c.readFromBeginning = true
	w.Close()
}

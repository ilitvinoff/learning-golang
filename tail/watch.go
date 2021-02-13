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

func startWatcher(config *config, filepath string, w *watcher.Watcher, t *tail.Tail, watchPollDelay time.Duration) {
	ifDebugPrintMsg("New watcher created")

	// Start the watching process - it'll check for changes periodically (default 100ms).
	err := w.Start(watchPollDelay)
	logFatalIfError(err)
}

func eventsHandler(filepath string, w *watcher.Watcher, tail *tail.Tail, cfg *config) {
	previousSize := getFileSize(filepath)
	var err error

	for {
		//time.Sleep(userWatchPollDellay)

		select {

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

		case _, _ = <-w.Closed:
			return

		case err := <-w.Error:

			ifDebugPrintMsg(fmt.Sprintln(" \nERR:", cfg.messagePrefix, "path:", cfg.path, "; error:", err.Error()))

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

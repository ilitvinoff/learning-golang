package main

import (
	"log"
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
	w.FilterOps(watcher.Rename, watcher.Move, watcher.Create, watcher.Remove)

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

func startWatcher(config *config, w *watcher.Watcher, t *tail.Tail, watchPollDelay time.Duration) {

	go eventsHandler(w, t, config)

	// Start the watching process - it'll check for changes periodically (default 100ms).
	if err := w.Start(watchPollDelay); err != nil {
		log.Fatalln(err)
	}
}

func eventsHandler(w *watcher.Watcher, t *tail.Tail, c *config) {
	for {
		select {
		case <-w.Event:
			err := t.Stop()
			logFatalIfError(err)
			c.readFromBeginning = true
			w.Close()

		case _, _ = <-w.Closed:
			return
		case err := <-w.Error:
			if err != watcher.ErrWatchedFileDeleted {
				log.Fatalln(err)
			}
			err = t.Stop()
			logFatalIfError(err)
			c.readFromBeginning = true
			w.Close()
		}
	}
}

package fswatch

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Note that [Watcher] is not an interface
type Watcher struct {
	*fsnotify.Watcher
}

func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{watcher}, nil
}

// Same as [fsnotify.Watcher.Add] but adds subdirectories recursively.
func (w *Watcher) AddRecursive(root string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = w.Add(path)
			if err != nil {
				log.Fatal("error adding path:", path, err)
			} else {
				log.Println("watching path:", path)
			}
		}
		return nil
	})
	return err
}

func (w *Watcher) Watch(f func(fsnotify.Event), opts ...(func(*WatchOpts))) {
	// apply options
	options := &WatchOpts{}
	for _, o := range opts {
		o(options)
	}

	for {
		select {

		case event, ok := <-w.Events:
			if !ok {
				return
			}

			if options.filter != nil && !options.filter(event) {
				continue
			}

			f(event)

		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)

		case <-time.After(2 * time.Second):
			log.Println("no events..")
		}
	}
}

// WatchDedup is like [Watcher.Watch] but it waits for a given duration
// before calling the function. This is a simple way to deduplicate events.
func (w *Watcher) WatchDedup(d time.Duration, f func(fsnotify.Event), opts ...(func(*WatchOpts))) {
	// apply options
	options := &WatchOpts{}
	for _, o := range opts {
		o(options)
	}

	var mu sync.Mutex
	activeTimers := make(map[string]*time.Timer)

	for {
		select {

		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Printf("ERROR: %s", err)

		case e, ok := <-w.Events:
			if !ok {
				return
			}

			if options.filter != nil && !options.filter(e) {
				continue
			}

			// Get the timer associated to the event name (ie the file path)
			mu.Lock()
			t, ok := activeTimers[e.Name]
			mu.Unlock()

			// If this timer does not exist, create it with an arbitrary 1h expiration
			if !ok {
				t = time.AfterFunc(1*time.Hour, func() {
					f(e)

					// remove the timer after expiration
					mu.Lock()
					delete(activeTimers, e.Name)
					mu.Unlock()
				})

				mu.Lock()
				activeTimers[e.Name] = t
				mu.Unlock()
			}

			// Start/reset the timer for this event name
			t.Reset(d)
		}
	}
}

// Options for the [Watcher.Watch] and [Watcher.WatchDedup] methods.
type WatchOpts struct {
	filter func(fsnotify.Event) bool
}

func WithFilter(filter func(fsnotify.Event) bool) func(*WatchOpts) {
	return func(opts *WatchOpts) {
		opts.filter = filter
	}
}

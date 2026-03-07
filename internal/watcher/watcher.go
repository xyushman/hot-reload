package watcher

import (
	"io/fs"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsnotify *fsnotify.Watcher
	events   chan fsnotify.Event
	errors   chan error
}

func New(root string) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		fsnotify: fw,
		events:   make(chan fsnotify.Event),
		errors:   make(chan error),
	}
	// Walk and add watches
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return fw.Add(path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	go w.eventLoop()
	return w, nil
}

func (w *Watcher) eventLoop() {
	for {
		select {
		case event := <-w.fsnotify.Events:
			w.events <- event
		case err := <-w.fsnotify.Errors:
			w.errors <- err
		}
	}
}

func (w *Watcher) Events() <-chan fsnotify.Event { return w.events }
func (w *Watcher) Errors() <-chan error          { return w.errors }
func (w *Watcher) Close() error                  { return w.fsnotify.Close() }

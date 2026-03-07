package watcher

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Event represents a file change event.
type Event struct {
	Path string
	Op   fsnotify.Op
}

// Watcher wraps an fsnotify.Watcher to provide recursive directory watching.
type Watcher struct {
	watcher *fsnotify.Watcher
	root    string
	dirs    map[string]bool
	mu      sync.RWMutex
}

// NewWatcher creates a new Watcher for the given root directory.
func NewWatcher(root string) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher: fw,
		root:    root,
		dirs:    make(map[string]bool),
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if IsIgnored(path) {
				return filepath.SkipDir
			}
			w.addDir(path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) addDir(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.dirs[path] {
		return
	}
	err := w.watcher.Add(path)
	if err != nil {
		slog.Warn("Failed to watch directory", "path", path, "error", err)
		return
	}
	w.dirs[path] = true
	slog.Debug("Watching directory", "path", path)
}

func (w *Watcher) removeDir(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.dirs[path] {
		return
	}
	_ = w.watcher.Remove(path)
	delete(w.dirs, path)
	slog.Debug("Stopped watching directory", "path", path)
}

// Start begins listening for file events and sending them to the returned channel.
func (w *Watcher) Start(ctx context.Context) <-chan struct{} {
	out := make(chan struct{})

	go func() {
		defer close(out)
		defer w.watcher.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				slog.Error("Watcher error", "error", err)
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

				if IsIgnored(event.Name) {
					continue
				}

				slog.Debug("File event", "event", event.String())

				// Check if this is a directory create/remove
				stat, err := os.Stat(event.Name)
				if err == nil && stat.IsDir() {
					if event.Has(fsnotify.Create) {
						w.addDir(event.Name)
					}
				}

				// If directory was removed, attempt to remove it
				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					// We can't always stat a removed file to know if it was a dir.
					// But we can check our map.
					w.mu.RLock()
					isWatchedDir := w.dirs[event.Name]
					w.mu.RUnlock()
					if isWatchedDir {
						w.removeDir(event.Name)
					}
				}

				// Signal that a relevant change happened
				select {
				case out <- struct{}{}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

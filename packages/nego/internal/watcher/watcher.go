package watcher

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/kerollosmagdy/nego/internal/logger"
)

// Watcher watches a source directory for .ts file changes and calls
// the onChange callback (debounced).
type Watcher struct {
	w          *fsnotify.Watcher
	closeOnce  sync.Once
	sourceRoot string
	debounce   time.Duration
	onChange   func()
}

// New creates and starts a directory watcher on sourceRoot.
func New(ctx context.Context, sourceRoot string, debounce time.Duration, onChange func()) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	wt := &Watcher{
		w:          w,
		sourceRoot: sourceRoot,
		debounce:   debounce,
		onChange:   onChange,
	}

	if err := wt.addAll(sourceRoot); err != nil {
		w.Close()
		return nil, err
	}

	go wt.loop(ctx)
	return wt, nil
}

func (wt *Watcher) addAll(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Warn("WalkDir error at %q: %v", path, err)
			return nil
		}
		if d.IsDir() {
			if err := wt.w.Add(path); err != nil {
				logger.Warn("Failed to watch %q: %v", path, err)
			}
		}
		return nil
	})
}

func (wt *Watcher) loop(ctx context.Context) {
	var timer *time.Timer
	for {
		select {
		case <-ctx.Done():
			wt.Close()
			if timer != nil {
				timer.Stop()
			}
			return

		case event, ok := <-wt.w.Events:
			if !ok {
				return
			}

			// Watch new directories
			if event.Has(fsnotify.Create) {
				_ = filepath.WalkDir(event.Name, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						logger.Warn("WalkDir error at %q: %v", path, err)
						return nil
					}
					if d.IsDir() {
						if err := wt.w.Add(path); err != nil {
							logger.Warn("Failed to watch %q: %v", path, err)
						}
					}
					return nil
				})
			}

			if event.Op == fsnotify.Chmod {
				continue
			}

			if !isTypeScriptFile(event.Name) {
				continue
			}

			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(wt.debounce, wt.onChange)

		case _, ok := <-wt.w.Errors:
			if !ok {
				return
			}
		}
	}
}

func (wt *Watcher) Close() {
	wt.closeOnce.Do(func() { _ = wt.w.Close() })
}

func isTypeScriptFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".ts" || ext == ".tsx"
}

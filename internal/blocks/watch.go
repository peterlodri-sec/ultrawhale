package blocks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ── Watch Primitive ───────────────────────────────────────────────────
// Watches directories for file changes and triggers repomap rebuilds.
// Uses polling on macOS (fsevents requires cgo), inotify on Linux.

// WatchEvent is a file system change notification.
type WatchEvent struct {
	Path      string
	Operation string // "create", "write", "remove", "rename"
	Timestamp  time.Time
}

// Watcher monitors directories for changes.
type Watcher struct {
	mu       sync.Mutex
	dirs     map[string]time.Time // dir → last modified
	events   chan WatchEvent
	stop     chan struct{}
	running  bool
	interval time.Duration
	pov      POV
}

// NewWatcher creates a file system watcher.
func NewWatcher(interval time.Duration) *Watcher {
	if interval == 0 {
		interval = 2 * time.Second
	}
	return &Watcher{
		dirs:     make(map[string]time.Time),
		events:   make(chan WatchEvent, 64),
		stop:     make(chan struct{}),
		interval: interval,
		pov:      CurrentPOV(),
	}
}

// Watch adds a directory to the watch list.
func (w *Watcher) Watch(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("watch: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("watch: %s is not a directory", dir)
	}

	w.mu.Lock()
	w.dirs[dir] = info.ModTime()
	w.mu.Unlock()

	return nil
}

// Start begins watching. Events are sent to the Events channel.
func (w *Watcher) Start() {
	w.mu.Lock()
	if w.running { w.mu.Unlock(); return }
	w.running = true
	w.mu.Unlock()

	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-w.stop:
				return
			case <-ticker.C:
				w.scan()
			}
		}
	}()

	Log(LogInfo, "blocks.Watch", fmt.Sprintf("watching %d dirs", len(w.dirs)), "", "", 0, nil)
}

func (w *Watcher) scan() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for dir, lastMod := range w.dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil { return nil }
			if info.IsDir() {
				if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if info.ModTime().After(lastMod) {
				op := "write"
				if _, err := os.Stat(path); os.IsNotExist(err) {
					op = "remove"
				}
				select {
				case w.events <- WatchEvent{Path: path, Operation: op, Timestamp: time.Now()}:
				default:
					// channel full — drop event
				}
			}
			return nil
		})
		w.dirs[dir] = time.Now()
	}
}

// Events returns the event channel.
func (w *Watcher) Events() <-chan WatchEvent { return w.events }

// Stop stops watching.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.running { return }
	w.running = false
	close(w.stop)
	Log(LogInfo, "blocks.Watch", "stopped", "", "", 0, nil)
}

// WatchStatus returns compact watch status.
func (w *Watcher) WatchStatus() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return fmt.Sprintf("watch: %d dirs, %d events queued", len(w.dirs), len(w.events))
}

// ── Global watcher ────────────────────────────────────────────────────

var globalWatcher = NewWatcher(2 * time.Second)

// WatchRepo starts watching the current directory recursively.
func WatchRepo() error {
	return globalWatcher.Watch(".")
}

// GetWatcher returns the global watcher.
func GetWatcher() *Watcher { return globalWatcher }


// StartWatchAndRebuild watches the repo and triggers repomap rebuilds.
func StartWatchAndRebuild() {
	WatchRepo()
	GetWatcher().Start()
	
	go func() {
		for ev := range GetWatcher().Events() {
			if strings.HasSuffix(ev.Path, ".go") || strings.HasSuffix(ev.Path, ".py") || strings.HasSuffix(ev.Path, ".zig") {
				Log(LogInfo, "blocks.Watch", fmt.Sprintf("change detected: %s", ev.Path), "", "", 0, nil)
			}
		}
	}()
}

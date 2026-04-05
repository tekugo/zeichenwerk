package main

import (
	"path/filepath"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
)

// watchEntry holds the fsnotify watcher and glob pattern for one target.
type watchEntry struct {
	fw      *fsnotify.Watcher
	pattern string
}

// Watcher manages per-target filesystem watchers. Each watched target gets its
// own fsnotify.Watcher so they can be started and stopped independently.
type Watcher struct {
	mu      sync.Mutex
	entries map[string]*watchEntry // target name → entry
	runner  *Runner
	dir     string
}

// NewWatcher creates a Watcher that enqueues runs via runner when files change.
func NewWatcher(runner *Runner, dir string) *Watcher {
	return &Watcher{
		entries: make(map[string]*watchEntry),
		runner:  runner,
		dir:     dir,
	}
}

// Start begins watching files matching pattern for the given target.
// If a watcher is already active for that target it is stopped first.
func (w *Watcher) Start(target Target, pattern string) error {
	w.Stop(target.Name)

	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Register all directories containing files that match the pattern.
	absPattern := filepath.Join(w.dir, pattern)
	matches, _ := doublestar.FilepathGlob(absPattern)
	seen := make(map[string]bool)
	for _, path := range matches {
		dir := filepath.Dir(path)
		if !seen[dir] {
			seen[dir] = true
			fw.Add(dir)
		}
	}
	// Always watch the root dir itself so new files are detected.
	fw.Add(w.dir)

	w.mu.Lock()
	w.entries[target.Name] = &watchEntry{fw: fw, pattern: pattern}
	w.mu.Unlock()

	go w.watch(fw, target, pattern)
	return nil
}

// Stop terminates the watcher for the named target (no-op if not watching).
func (w *Watcher) Stop(name string) {
	w.mu.Lock()
	entry, ok := w.entries[name]
	if ok {
		delete(w.entries, name)
	}
	w.mu.Unlock()
	if ok {
		entry.fw.Close()
	}
}

// StopAll terminates all active watchers.
func (w *Watcher) StopAll() {
	w.mu.Lock()
	entries := make(map[string]*watchEntry, len(w.entries))
	for k, v := range w.entries {
		entries[k] = v
	}
	w.entries = make(map[string]*watchEntry)
	w.mu.Unlock()
	for _, e := range entries {
		e.fw.Close()
	}
}

// IsWatching reports whether a watcher is active for the named target.
func (w *Watcher) IsWatching(name string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, ok := w.entries[name]
	return ok
}

// Count returns the number of active watchers.
func (w *Watcher) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.entries)
}

// Pattern returns the glob pattern for the named target, or "".
func (w *Watcher) Pattern(name string) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if e, ok := w.entries[name]; ok {
		return e.pattern
	}
	return ""
}

// watch is the per-target event loop.
func (w *Watcher) watch(fw *fsnotify.Watcher, target Target, pattern string) {
	for {
		select {
		case event, ok := <-fw.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				rel, err := filepath.Rel(w.dir, event.Name)
				if err != nil {
					continue
				}
				matched, _ := doublestar.Match(pattern, rel)
				if matched {
					w.runner.Enqueue(target)
				}
			}
		case _, ok := <-fw.Errors:
			if !ok {
				return
			}
		}
	}
}

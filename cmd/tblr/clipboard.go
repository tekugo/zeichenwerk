package main

import (
	"crypto/sha256"
	"time"

	"github.com/atotto/clipboard"
	"github.com/tekugo/zeichenwerk/cmd/tblr/format"
)

// ClipboardWatcher polls the clipboard and calls onChange when new table content
// is detected. Deduplicates by SHA-256 of last processed content.
type ClipboardWatcher struct {
	interval time.Duration
	lastHash [32]byte
	stop     chan struct{}
	onChange func(t *format.MutableTable, fmt format.Format)
}

// NewClipboardWatcher creates a watcher with 200ms poll interval.
func NewClipboardWatcher(onChange func(*format.MutableTable, format.Format)) *ClipboardWatcher {
	return &ClipboardWatcher{
		interval: 200 * time.Millisecond,
		stop:     make(chan struct{}),
		onChange: onChange,
	}
}

// Start begins polling in a background goroutine.
func (w *ClipboardWatcher) Start() {
	go w.loop()
}

// Stop halts the watcher.
func (w *ClipboardWatcher) Stop() {
	select {
	case w.stop <- struct{}{}:
	default:
	}
}

func (w *ClipboardWatcher) loop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			w.check()
		}
	}
}

func (w *ClipboardWatcher) check() {
	text, err := clipboard.ReadAll()
	if err != nil || text == "" {
		return
	}
	data := []byte(text)
	h := sha256.Sum256(data)
	if h == w.lastHash {
		return
	}
	f := format.Detect(data)
	if f == nil {
		return
	}
	t, err := f.Parse(data, format.ParseOpts{})
	if err != nil || t == nil {
		return
	}
	w.lastHash = h
	if w.onChange != nil {
		w.onChange(t, f)
	}
}

// WriteToClipboard serialises the table in the given format and writes to the
// system clipboard.
func WriteToClipboard(t *format.MutableTable, f format.Format, pretty bool) error {
	data, err := f.Serialize(t, format.SerialOpts{Pretty: pretty})
	if err != nil {
		return err
	}
	return clipboard.WriteAll(string(data))
}

// ReadFromClipboard reads the clipboard, auto-detects format, and returns the
// parsed table and format. Returns nil if the clipboard is empty or not a table.
func ReadFromClipboard() (*format.MutableTable, format.Format, error) {
	text, err := clipboard.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if text == "" {
		return nil, nil, nil
	}
	data := []byte(text)
	f := format.Detect(data)
	if f == nil {
		return nil, nil, nil
	}
	t, err := f.Parse(data, format.ParseOpts{})
	return t, f, err
}

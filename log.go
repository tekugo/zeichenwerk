package zeichenwerk

import (
	"fmt"
	"time"
)

type LogEntry struct {
	Time    time.Time
	Level   string
	Source  string
	Message string
}

func (le *LogEntry) String() string {
	return fmt.Sprintf("[%s] %s (%s): %s", le.Time.Format(time.RFC3339), le.Level, le.Source, le.Message)
}

type Log struct {
	entries []LogEntry
	columns []TableColumn
	size    int
	start   int
	count   int
}

func NewLog(size int) *Log {
	return &Log{
		entries: make([]LogEntry, size),
		columns: []TableColumn{
			{Header: "Time", Width: 12, Sortable: true, Filterable: false},
			{Header: "Level", Width: 5, Sortable: false, Filterable: true},
			{Header: "Source", Width: 20, Sortable: false, Filterable: true},
			{Header: "Message", Width: 200, Sortable: false, Filterable: false},
		},
		size: size,
	}
}

func (l *Log) Add(source, level, message string, params ...any) {
	index := (l.start + l.count) % l.size
	l.entries[index] = LogEntry{
		Time:    time.Now(),
		Level:   level,
		Source:  source,
		Message: fmt.Sprintf(message, params...),
	}

	if l.count < l.size {
		l.count++
	} else {
		l.start = (l.start + 1) % l.size
	}
}

func (l *Log) Columns() []TableColumn {
	return l.columns
}

func (l *Log) Length() int {
	return l.count
}

func (l *Log) Str(row, column int) string {
	entry := l.entries[(l.start+l.count-row-1)%l.size]
	switch column {
	case 0:
		return entry.Time.Format(time.TimeOnly)
	case 1:
		return entry.Level
	case 2:
		return entry.Source
	default:
		return entry.Message
	}
}

func (l *Log) Iter() <-chan LogEntry {
	ch := make(chan LogEntry)

	go func() {
		defer close(ch)
		for i := range l.count {
			ch <- l.entries[(l.start+i)%l.size]
		}
	}()

	return ch
}

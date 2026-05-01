package zeichenwerk

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/tekugo/zeichenwerk/v2/widgets"
)

// TableLogItem represents a single entry in the TableLog circular buffer.
type TableLogItem struct {
	Time    time.Time
	Level   string
	Source  string
	Message string
}

// String returns a formatted representation of the log item using RFC3339 timestamp.
func (le *TableLogItem) String() string {
	return fmt.Sprintf("[%s] %s (%s): %s", le.Time.Format(time.RFC3339), le.Level, le.Source, le.Message)
}

// TableLog is a circular buffer that stores log entries and implements the
// TableProvider interface, allowing it to be used as a data source for a Table widget.
// The buffer has a fixed capacity; when full, oldest entries are overwritten.
type TableLog struct {
	items   []TableLogItem
	columns []widgets.TableColumn
	size    int
	start   int
	count   int
}

// NewTableLog creates a new TableLog with the given maximum number of entries.
// The columns are predefined as Time (12), Level (5), Source (20), and Message (200).
func NewTableLog(size int) *TableLog {
	return &TableLog{
		items: make([]TableLogItem, size),
		columns: []widgets.TableColumn{
			{Header: "Time", Width: 12, Sortable: true, Filterable: false},
			{Header: "Level", Width: 5, Sortable: false, Filterable: true},
			{Header: "Source", Width: 20, Sortable: false, Filterable: true},
			{Header: "Message", Width: 200, Sortable: false, Filterable: false},
		},
		size: size,
	}
}

// Add inserts a new log entry into the buffer. The entry consists of a source,
// a level string, a format message, and optional printf-style parameters.
// The buffer grows until it reaches the configured size, after which the oldest
// entry is overwritten.
func (t *TableLog) Add(source, level, message string, params ...any) {
	index := (t.start + t.count) % t.size
	t.items[index] = TableLogItem{
		Time:    time.Now(),
		Level:   level,
		Source:  source,
		Message: fmt.Sprintf(message, params...),
	}

	if t.count < t.size {
		t.count++
	} else {
		t.start = (t.start + 1) % t.size
	}
}

// Columns returns the column definitions for the TableProvider interface.
func (t *TableLog) Columns() []widgets.TableColumn {
	return t.columns
}

// Length returns the number of log entries currently stored (up to the buffer size).
func (t *TableLog) Length() int {
	return t.count
}

// Str returns the string value for the cell at the given row and column.
// Rows are indexed from 0 (most recent entry) to Length-1 (oldest entry).
// Column indices: 0=Time, 1=Level, 2=Source, 3=Message.
func (t *TableLog) Str(row, column int) string {
	entry := t.items[(t.start+t.count-row-1)%t.size]
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

// Iter returns a channel that streams all log entries from oldest to newest.
// The channel is closed after all entries have been sent.
func (t *TableLog) Iter() <-chan TableLogItem {
	ch := make(chan TableLogItem)

	go func() {
		defer close(ch)
		for i := range t.count {
			ch <- t.items[(t.start+i)%t.size]
		}
	}()

	return ch
}

// UILogHandler is a slog.Handler that routes structured log entries to both
// a TableLog (for tabular display) and a Text widget (for human-readable scrolling logs).
// It optionally also writes logs to stderr for development debugging.
type UILogHandler struct {
	tableLog *TableLog
	text     *widgets.Text
	level    slog.Level
	console  bool
	attrs    []slog.Attr // attributes from WithAttrs
}

// Enabled reports whether the handler handles records at the given level.
func (h *UILogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle formats the log record and writes it to the TableLog, the Text widget,
// and optionally to stderr. It merges any attributes from WithAttrs.
// Structured attributes other than "source" and "widgetType" are appended to
// the message as key=value pairs.
func (h *UILogHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}

	// Merge base attrs and record attrs
	allAttrs := make([]slog.Attr, 0, len(h.attrs)+2)
	allAttrs = append(allAttrs, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		allAttrs = append(allAttrs, a)
		return true
	})

	// Extract source and widgetType
	var source, widgetType string
	var other []slog.Attr
	for _, a := range allAttrs {
		switch a.Key {
		case "source":
			if v, ok := a.Value.Any().(string); ok {
				source = v
			}
		case "widgetType":
			if v, ok := a.Value.Any().(string); ok {
				widgetType = v
			}
		default:
			other = append(other, a)
		}
	}

	// Build message: original + other attrs
	msg := r.Message
	if len(other) > 0 {
		parts := make([]string, len(other))
		for i, a := range other {
			parts[i] = fmt.Sprintf("%s=%v", a.Key, a.Value.Any())
		}
		msg = msg + " " + strings.Join(parts, " ")
	}

	// Add to TableLog with source and level (use constant format)
	h.tableLog.Add(source, r.Level.String(), "%s", msg)

	// Format for console output
	timeStr := r.Time.Format("15:04:05.000")
	line := fmt.Sprintf("%s %-5s (%s) %s %s", timeStr, r.Level.String(), source, widgetType, msg)

	// Write to text widget
	if h.text != nil {
		h.text.Add(line)
	}

	// Also write to stderr if console enabled
	if h.console {
		fmt.Fprintln(os.Stderr, line)
	}

	return nil
}

// WithAttrs returns a new handler that has the given attributes added to its base attrs.
func (h *UILogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	merged = append(merged, h.attrs...)
	merged = append(merged, attrs...)
	return &UILogHandler{
		tableLog: h.tableLog,
		text:     h.text,
		level:    h.level,
		console:  h.console,
		attrs:    merged,
	}
}

// WithGroup returns a new handler with the given group name. Since this handler
// does not use structured groups, it simply returns the receiver.
func (h *UILogHandler) WithGroup(name string) slog.Handler {
	// Groups are not specially handled; return unchanged.
	return h
}

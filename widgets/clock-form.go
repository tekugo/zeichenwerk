package widgets

import (
	"fmt"
	"io"
	"time"

	"github.com/tekugo/zeichenwerk/core"
)

// ClockForm is the WidgetForm for *Clock. The tick interval is
// surfaced as a string (parsed via time.ParseDuration) because the
// form layer has no native Duration control and a string is the most
// natural editing surface — "1s", "500ms", "1m".
type ClockForm struct {
	ComponentForm

	Interval string `group:"general" label:"Interval"`
	Format   string `group:"general" label:"Format"`
	Prefix   string `group:"general" label:"Prefix"`
}

func (f *ClockForm) Name() string  { return "Clock" }
func (f *ClockForm) Group() string { return "leaf" }
func (f *ClockForm) Help() string  { return "Animated clock displaying the current time" }

func (f *ClockForm) Load(w core.Widget) {
	c := w.(*Clock)
	f.ComponentForm.Load(&c.Component)
	f.Interval = c.interval.String()
	f.Format = c.format
	f.Prefix = c.prefix
}

func (f *ClockForm) Store(w core.Widget) {
	c := w.(*Clock)
	f.ComponentForm.Store(&c.Component)
	if d, err := time.ParseDuration(f.Interval); err == nil && d > 0 {
		c.interval = d
	}
	if f.Format != "" {
		c.format = f.Format
	}
	c.prefix = f.Prefix
}

func (f *ClockForm) New() core.Widget {
	d, err := time.ParseDuration(f.Interval)
	if err != nil || d <= 0 {
		d = time.Second
	}
	format := f.Format
	if format == "" {
		format = "15:04"
	}
	c := NewClock("", "", d, format, f.Prefix)
	f.Store(c)
	return c
}

func (f *ClockForm) Validate(field string) error {
	if field == "Interval" && f.Interval != "" {
		if _, err := time.ParseDuration(f.Interval); err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
	}
	return nil
}

// Emit writes the Clock constructor. The interval is emitted as a
// time.ParseDuration call wrapped in a must-style helper so the
// generated code does not need to handle the error inline; if a more
// idiomatic literal is preferred, the user can replace it after
// generation.
func (f *ClockForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		d, err := time.ParseDuration(f.Interval)
		if err != nil || d <= 0 {
			d = time.Second
		}
		args := fmt.Sprintf("%q, %s", f.ID, durationLiteral(d))
		if f.Format != "" || f.Prefix != "" {
			args += fmt.Sprintf(", %q", f.Format)
		}
		if f.Prefix != "" {
			args += fmt.Sprintf(", %q", f.Prefix)
		}
		_, err = fmt.Fprintf(w, "Clock(%s).\n", args)
		return err
	})
}

// durationLiteral renders a time.Duration as a Go expression. Common
// short durations get nicely-formed literals like "time.Second" or
// "500*time.Millisecond"; anything else falls back to a
// ParseDuration-based expression rather than a giant ns count.
func durationLiteral(d time.Duration) string {
	switch d {
	case time.Second:
		return "time.Second"
	case time.Minute:
		return "time.Minute"
	case time.Millisecond:
		return "time.Millisecond"
	}
	return fmt.Sprintf("time.Duration(%d)", d.Nanoseconds())
}

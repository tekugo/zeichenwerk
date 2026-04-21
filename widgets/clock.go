package widgets

import (
	"time"
	"unicode/utf8"

	. "github.com/tekugo/zeichenwerk/core"
)

// referenceTime is the Go time-format reference instant. Formatting it with
// any layout string produces a representative, stable-width sample — useful
// for computing a reliable Hint width.
var referenceTime = time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC)

// Clock is an animated widget that displays the current time. It updates at
// the interval supplied to the constructor and is started and stopped
// explicitly by the caller.
//
// The time is formatted with a standard Go time-layout string (e.g.
// "15:04:05" or "15:04"). An optional prefix string (e.g. a Nerd Font icon
// such as " ") is prepended to the formatted time.
//
// Usage:
//
//	c := z.NewClock("clock", "", time.Second, "15:04:05", " ")
//	c.Start()
//	// … later …
//	c.Stop()
type Clock struct {
	Animation
	format   string
	prefix   string
	interval time.Duration
}

// NewClock creates a Clock widget. interval is stored and used when Start is
// called. params are optional positional strings: params[0] is the Go
// time-layout (default "15:04"), params[1] is a prefix prepended to the time
// string (default "").
func NewClock(id, class string, interval time.Duration, params ...string) *Clock {
	format := "15:04"
	prefix := ""
	if len(params) > 0 && params[0] != "" {
		format = params[0]
	}
	if len(params) > 1 {
		prefix = params[1]
	}
	c := &Clock{
		Animation: Animation{
			Component: Component{id: id, class: class},
			stop:      make(chan struct{}),
		},
		format:   format,
		prefix:   prefix,
		interval: interval,
	}
	c.fn = c.Tick
	return c
}

// ---- Widget interface -------------------------------------------------------

// Apply registers the "clock" theme style.
func (c *Clock) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("clock"))
}

// Hint returns the preferred size. Width is computed by formatting the Go
// reference time with the layout and prepending the prefix, giving a stable
// representative width. Height is always 1.
func (c *Clock) Hint() (int, int) {
	if c.hwidth != 0 || c.hheight != 0 {
		return c.hwidth, c.hheight
	}
	sample := c.prefix + referenceTime.Format(c.format)
	return utf8.RuneCountInString(sample), 1
}

// Refresh triggers a redraw of only this widget.
func (c *Clock) Refresh() {
	Redraw(c)
}

// Tick is called on every animation frame; it requests a redraw so Render
// picks up the new time.Now() value.
func (c *Clock) Tick() {
	c.Refresh()
}

// Start begins the clock using the interval set at construction time.
func (c *Clock) Start() {
	c.Animation.Start(c.interval)
}

// Render draws the current time to the screen.
func (c *Clock) Render(r *Renderer) {
	c.Component.Render(r)

	x, y, w, h := c.Content()
	if w < 1 || h < 1 {
		return
	}
	style := c.Style()
	r.Set(style.Foreground(), style.Background(), style.Font())
	r.Text(x, y, c.prefix+time.Now().Format(c.format), w)
}

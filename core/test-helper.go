package core

import "github.com/gdamore/tcell/v3"

// BuildKey constructs a synthetic tcell.EventKey for the given key
// constant, carrying no rune and no modifier bits. It is intended for
// tests that need to feed special keys (arrows, Escape, function keys,
// ...) into widgets as if they had been received from tcell's event loop.
func BuildKey(key tcell.Key) *tcell.EventKey {
	return tcell.NewEventKey(key, "", tcell.ModNone)
}

// BuildRune constructs a synthetic tcell.EventKey representing a printable
// character. Only the first rune of s is meaningful to tcell, but accepting
// a string keeps callsites readable when writing unit tests ("a", "A",
// "ß" instead of rune literals).
func BuildRune(s string) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, s, tcell.ModNone)
}

// TestScreen is an in-memory renderer.Screen used by unit tests to observe
// what a widget draws. Every Put is recorded in a cell map keyed by (x, y)
// along with the foreground colour that was active at the time of the
// call, making it straightforward to assert on both characters and
// styling. Flush, Clear, Clip, Translate, and SetUnderline are no-ops
// because tests do not need to model screen buffering or dirty regions.
type TestScreen struct {
	cells map[[2]int]string
	fg    string
	bg    string
	bgs   map[[2]int]string
	fgs   map[[2]int]string // foreground colour captured at Put time
}

// NewTestScreen returns an empty TestScreen ready for use. The maps that
// back it are initialised so the zero-allocation path is the common case
// for tests that only perform a few writes.
func NewTestScreen() *TestScreen {
	return &TestScreen{
		cells: make(map[[2]int]string),
		bgs:   make(map[[2]int]string),
		fgs:   make(map[[2]int]string),
	}
}

// Bg returns the background colour recorded at the given coordinate, or
// the empty string if that cell was never written.
func (c *TestScreen) Bg(x, y int) string { return c.bgs[[2]int{x, y}] }

// Clear is a no-op; tests inspect recorded cells directly.
func (c *TestScreen) Clear() {}

// Clip is a no-op; the test screen does not model clipping regions.
func (c *TestScreen) Clip(x, y, w, h int) {}

// Fg returns the foreground colour that was active when the cell at the
// given coordinate was last written, or the empty string if never written.
func (c *TestScreen) Fg(x, y int) string { return c.fgs[[2]int{x, y}] }

// Flush is a no-op; TestScreen is always synchronous.
func (c *TestScreen) Flush() {}

// Get returns the character string written to the given coordinate, or the
// empty string if that cell was never written.
func (c *TestScreen) Get(x, y int) string { return c.cells[[2]int{x, y}] }

// Put records that ch was written at (x, y) under the currently active
// foreground colour. The background colour is intentionally not captured
// per cell because most rendering tests only need to verify glyphs and
// foreground styling.
func (c *TestScreen) Put(x, y int, ch string) {
	c.cells[[2]int{x, y}] = ch
	c.fgs[[2]int{x, y}] = c.fg
}

// Set updates the current foreground and background colours used by
// subsequent Put calls. The font argument is ignored.
func (c *TestScreen) Set(fg, bg, font string) { c.fg = fg; c.bg = bg }

// SetUnderline is a no-op; underline styling is not modelled.
func (c *TestScreen) SetUnderline(style int, color string) {}

// Translate is a no-op; the test screen uses absolute coordinates.
func (c *TestScreen) Translate(x, y int) {}

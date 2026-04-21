package core

import "github.com/gdamore/tcell/v3"

// buildKey creates a tcell.EventKey for the given key constant.
func BuildKey(key tcell.Key) *tcell.EventKey {
	return tcell.NewEventKey(key, "", tcell.ModNone)
}

func BuildRune(s string) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, s, tcell.ModNone)
}

// TellScreen records every Put call so tests can inspect rendered characters.
type TestScreen struct {
	cells map[[2]int]string
	fg    string
	bg    string
	bgs   map[[2]int]string
	fgs   map[[2]int]string // fg colour at Put time
}

func NewTestScreen() *TestScreen {
	return &TestScreen{
		cells: make(map[[2]int]string),
		bgs:   make(map[[2]int]string),
		fgs:   make(map[[2]int]string),
	}
}

func (c *TestScreen) Bg(x, y int) string                   { return c.bgs[[2]int{x, y}] }
func (c *TestScreen) Clear()                               {}
func (c *TestScreen) Clip(x, y, w, h int)                  {}
func (c *TestScreen) Fg(x, y int) string                   { return c.fgs[[2]int{x, y}] }
func (c *TestScreen) Flush()                               {}
func (c *TestScreen) Get(x, y int) string                  { return c.cells[[2]int{x, y}] }
func (c *TestScreen) Put(x, y int, ch string)              { c.cells[[2]int{x, y}] = ch; c.fgs[[2]int{x, y}] = c.fg }
func (c *TestScreen) Set(fg, bg, font string)              { c.fg = fg; c.bg = bg }
func (c *TestScreen) SetUnderline(style int, color string) {}
func (c *TestScreen) Translate(x, y int)                   {}

package main

import (
	"fmt"

	"github.com/gdamore/tcell/v3"
	zw "github.com/tekugo/zeichenwerk"
	"github.com/tekugo/zeichenwerk/core"
)

// App is the top-level malwerk runtime — owns the UI, theme, document,
// editor widget, status bar, history, and yank register.
type App struct {
	ui       *zw.UI
	theme    *core.Theme
	editor   *Editor
	status   *StatusBar
	history  *History
	register *Register
}

// NewApp builds the UI tree and wires the editor. If doc is nil, a fresh
// document is created with the requested width/height (or terminal size
// when w/h are 0).
func NewApp(theme *core.Theme, doc *Document, w, h int) *App {
	app := &App{
		theme:    theme,
		history:  NewHistory(200),
		register: &Register{},
	}

	if doc == nil {
		tw, th := terminalSize()
		if w > 0 {
			tw = w
		}
		if h > 0 {
			th = h
		}
		doc = NewDocument(tw, th)
	}

	app.editor = NewEditor("editor", "", app, doc)
	app.status = NewStatusBar("statusbar")

	b := zw.NewBuilder(theme).
		VFlex("root", core.Stretch, 0)
	b.Add(app.editor).Hint(0, -1)
	b.Add(app.status)
	app.ui = b.End().Build()

	app.refreshStatus()
	app.registerCommands()
	app.ui.SetFocus("editor")
	return app
}

// Run starts the UI event loop.
func (a *App) Run() {
	a.ui.Run()
}

// terminalSize returns the current terminal size as (cols, rows). Falls
// back to 80×24 if the screen cannot be probed.
func terminalSize() (int, int) {
	scr, err := tcell.NewScreen()
	if err != nil {
		return 80, 24
	}
	if err := scr.Init(); err != nil {
		return 80, 24
	}
	defer scr.Fini()
	w, h := scr.Size()
	if w <= 0 || h <= 0 {
		return 80, 24
	}
	return w, h
}

// refreshStatus updates the status-bar text from the editor's state.
func (a *App) refreshStatus() {
	if a.status == nil || !a.status.Visible() {
		return
	}
	cx, cy := a.editor.CursorPos()
	ox, oy := a.editor.Offset()
	style := a.editor.CurrentStyle()
	ds := a.editor.Doc().StyleFor(style)
	fg := ds.Fg
	if fg == "" {
		fg = "?"
	}
	bg := ds.Bg
	if bg == "" {
		bg = "?"
	}
	dirty := ""
	if a.editor.Doc().Dirty {
		dirty = " *"
	}
	a.status.Set(fmt.Sprintf("[%s] %d×%d @ %d,%d (+%d,%d)  Style: %s  fg=%s bg=%s%s",
		a.editor.CurrentMode(),
		a.editor.Doc().Width, a.editor.Doc().Height,
		cx, cy, ox, oy,
		style, fg, bg, dirty))
}

// ToggleStatus shows or hides the status bar; the editor reclaims the
// row when the bar is hidden.
func (a *App) ToggleStatus() {
	a.status.SetVisible(!a.status.Visible())
	a.refreshStatus()
}


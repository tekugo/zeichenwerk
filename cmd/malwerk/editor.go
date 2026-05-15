package main

import (
	"github.com/gdamore/tcell/v3"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// Mode is the editor's input state.
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
	ModeExtended
)

// String returns the upper-case label shown in the status bar.
func (m Mode) String() string {
	switch m {
	case ModeInsert:
		return "INSERT"
	case ModeVisual:
		return "VISUAL"
	case ModeExtended:
		return "EXTENDED"
	default:
		return "NORMAL"
	}
}

// Editor is the malwerk drawing widget. It owns the Document, the cursor,
// the viewport offset, the current mode, the current style / border
// selection, and dispatches all key events through mode-specific handlers.
type Editor struct {
	widgets.Component
	app *App

	doc      *Document
	cursorX  int
	cursorY  int
	offsetX  int
	offsetY  int
	visualX  int
	visualY  int
	mode     Mode
	style    string // current palette key
	border   string // current border family ("thin", "double", …)
	lastRune string // last typed rune for repeat / extended fallback
}

// NewEditor builds an editor over the given document, with the app
// reference so key handlers can mutate global state (history, register,
// status bar, dialogs).
func NewEditor(id, class string, app *App, doc *Document) *Editor {
	e := &Editor{
		app:      app,
		doc:      doc,
		mode:     ModeNormal,
		style:    "default",
		border:   "thin",
		lastRune: "█",
	}
	e.Component = *widgets.NewComponent(id, class)
	e.SetFlag(core.FlagFocusable, true)
	widgets.OnKey(e, e.handleKey)
	return e
}

// ---- Widget Methods -------------------------------------------------------

// Apply registers the editor styles. The selectors are "editor",
// "editor/cursor", and "editor/selection".
func (e *Editor) Apply(theme *core.Theme) {
	theme.Apply(e, e.Selector("editor"))
	theme.Apply(e, e.Selector("editor/cursor"))
	theme.Apply(e, e.Selector("editor/selection"))
}

// Cursor returns the screen-space cursor position so the UI can place
// the terminal caret at the editor's logical cursor.
func (e *Editor) Cursor() (int, int, string) {
	if !e.Flag(core.FlagFocused) {
		return -1, -1, ""
	}
	x0, y0, _, _ := e.Content()
	return x0 + e.cursorX - e.offsetX, y0 + e.cursorY - e.offsetY, "block"
}

// Hint returns the editor's preferred size. An explicit SetHint — used
// by the Builder via .Hint(0, -1) to make the editor fractional in the
// flex parent — wins; otherwise we report the document's natural
// dimensions.
func (e *Editor) Hint() (int, int) {
	if w, h := e.Component.Hint(); w != 0 || h != 0 {
		return w, h
	}
	return e.doc.Width, e.doc.Height
}

// Render paints the visible portion of the document, plus the visual-mode
// selection highlight if active.
func (e *Editor) Render(r *core.Renderer) {
	e.Component.Render(r)

	x0, y0, w, h := e.Content()
	base := e.Style("")
	baseFg := base.Foreground()
	baseBg := base.Background()

	selX1, selY1, selX2, selY2 := e.selectionRect()

	for sy := range h {
		dy := sy + e.offsetY
		for sx := range w {
			dx := sx + e.offsetX
			if dx >= e.doc.Width || dy >= e.doc.Height {
				r.Set(baseFg, baseBg, base.Font())
				r.Put(x0+sx, y0+sy, " ")
				continue
			}
			cell := e.doc.At(dx, dy)
			ds := e.doc.StyleFor(cell.Style)
			fg := ds.Fg
			if fg == "" {
				fg = baseFg
			}
			bg := ds.Bg
			if bg == "" {
				bg = baseBg
			}
			if e.mode == ModeVisual && dx >= selX1 && dx <= selX2 && dy >= selY1 && dy <= selY2 {
				sel := e.Style("/selection")
				if selBg := sel.Background(); selBg != "" {
					bg = selBg
				}
				if selFg := sel.Foreground(); selFg != "" {
					fg = selFg
				}
			}
			r.Set(fg, bg, ds.Font)
			ch := cell.Ch
			if ch == "" {
				ch = " "
			}
			r.Text(x0+sx, y0+sy, ch, 1)
		}
	}
}

// ---- Document accessors ---------------------------------------------------

// Doc returns the editor's current document.
func (e *Editor) Doc() *Document { return e.doc }

// SetDoc swaps the document and resets cursor / viewport state. Used after
// New / Open / Resize.
func (e *Editor) SetDoc(doc *Document) {
	e.doc = doc
	e.cursorX, e.cursorY = 0, 0
	e.offsetX, e.offsetY = 0, 0
	e.visualX, e.visualY = 0, 0
	e.mode = ModeNormal
	if _, ok := doc.Palette[e.style]; !ok {
		e.style = "default"
	}
	widgets.Relayout(e)
	e.app.refreshStatus()
}

// CursorPos returns the document-relative cursor position.
func (e *Editor) CursorPos() (int, int) { return e.cursorX, e.cursorY }

// Offset returns the viewport offset.
func (e *Editor) Offset() (int, int) { return e.offsetX, e.offsetY }

// CurrentStyle returns the active style name.
func (e *Editor) CurrentStyle() string { return e.style }

// SetCurrentStyle changes the active style name. Unknown names fall back
// to "default".
func (e *Editor) SetCurrentStyle(name string) {
	if _, ok := e.doc.Palette[name]; !ok {
		name = "default"
	}
	e.style = name
	e.app.refreshStatus()
	e.Refresh()
}

// CurrentBorder returns the active border family name.
func (e *Editor) CurrentBorder() string { return e.border }

// SetCurrentBorder updates the active border family.
func (e *Editor) SetCurrentBorder(name string) {
	e.border = name
	e.app.refreshStatus()
}

// SetMode changes the editor's mode and refreshes the status bar.
func (e *Editor) SetMode(m Mode) {
	if m == ModeVisual && e.mode != ModeVisual {
		e.visualX, e.visualY = e.cursorX, e.cursorY
	}
	e.mode = m
	e.app.refreshStatus()
	e.Refresh()
}

// CurrentMode returns the current mode.
func (e *Editor) CurrentMode() Mode { return e.mode }

// ---- Cursor / viewport ----------------------------------------------------

// MoveTo places the cursor at the given absolute document position,
// clamped to bounds, and scrolls the viewport to keep the cursor visible.
func (e *Editor) MoveTo(x, y int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= e.doc.Width {
		x = e.doc.Width - 1
	}
	if y >= e.doc.Height {
		y = e.doc.Height - 1
	}
	e.cursorX, e.cursorY = x, y
	e.scrollIntoView()
	e.app.refreshStatus()
	e.Refresh()
}

// Move advances the cursor by (dx, dy).
func (e *Editor) Move(dx, dy int) {
	e.MoveTo(e.cursorX+dx, e.cursorY+dy)
}

// scrollIntoView nudges the viewport so the cursor is visible within the
// content area.
func (e *Editor) scrollIntoView() {
	_, _, w, h := e.Content()
	if w <= 0 || h <= 0 {
		return
	}
	if e.cursorX < e.offsetX {
		e.offsetX = e.cursorX
	} else if e.cursorX >= e.offsetX+w {
		e.offsetX = e.cursorX - w + 1
	}
	if e.cursorY < e.offsetY {
		e.offsetY = e.cursorY
	} else if e.cursorY >= e.offsetY+h {
		e.offsetY = e.cursorY - h + 1
	}
	maxX := max(0, e.doc.Width-w)
	maxY := max(0, e.doc.Height-h)
	if e.offsetX > maxX {
		e.offsetX = maxX
	}
	if e.offsetY > maxY {
		e.offsetY = maxY
	}
	if e.offsetX < 0 {
		e.offsetX = 0
	}
	if e.offsetY < 0 {
		e.offsetY = 0
	}
}

// PageScroll moves the viewport by `lines` rows (positive = down). The
// cursor is dragged along when it would leave the new viewport.
func (e *Editor) PageScroll(lines int) {
	_, _, _, h := e.Content()
	if h <= 0 {
		return
	}
	e.offsetY += lines
	maxY := max(0, e.doc.Height-h)
	if e.offsetY > maxY {
		e.offsetY = maxY
	}
	if e.offsetY < 0 {
		e.offsetY = 0
	}
	if e.cursorY < e.offsetY {
		e.cursorY = e.offsetY
	} else if e.cursorY >= e.offsetY+h {
		e.cursorY = e.offsetY + h - 1
	}
	e.app.refreshStatus()
	e.Refresh()
}

// ---- Visual-mode selection rect ------------------------------------------

// selectionRect returns the inclusive rectangle (x1, y1)-(x2, y2) of the
// current visual selection in document coordinates. When not in Visual
// mode, returns a degenerate rect outside the document.
func (e *Editor) selectionRect() (x1, y1, x2, y2 int) {
	if e.mode != ModeVisual {
		return -1, -1, -2, -2
	}
	x1, x2 = e.visualX, e.cursorX
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	y1, y2 = e.visualY, e.cursorY
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	return
}

// ---- Key dispatch ---------------------------------------------------------

// handleKey is the editor's top-level key router. It checks for global
// shortcuts (Ctrl-K, Ctrl-S, Ctrl-Q, Ctrl-Z, Ctrl-R, Esc) and otherwise
// dispatches by mode.
func (e *Editor) handleKey(evt *tcell.EventKey) bool {
	switch evt.Key() {
	case tcell.KeyEsc:
		if e.mode != ModeNormal {
			e.SetMode(ModeNormal)
		}
		return true
	case tcell.KeyCtrlK:
		e.app.ui.Commands().Open()
		return true
	case tcell.KeyCtrlS:
		e.app.cmdSave()
		return true
	case tcell.KeyCtrlQ:
		e.app.cmdQuit()
		return true
	case tcell.KeyCtrlZ:
		e.app.history.Undo(e.doc, e)
		return true
	case tcell.KeyCtrlR:
		e.app.history.Redo(e.doc, e)
		return true
	}
	switch e.mode {
	case ModeInsert:
		return e.handleInsert(evt)
	case ModeVisual:
		return e.handleVisual(evt)
	case ModeExtended:
		return e.handleExtended(evt)
	default:
		return e.handleNormal(evt)
	}
}

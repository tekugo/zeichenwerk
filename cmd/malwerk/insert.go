package main

import (
	"github.com/gdamore/tcell/v3"
)

// handleInsert processes key events while in Insert mode.
func (e *Editor) handleInsert(evt *tcell.EventKey) bool {
	switch evt.Key() {
	case tcell.KeyUp:
		e.Move(0, -1)
		return true
	case tcell.KeyDown:
		e.Move(0, 1)
		return true
	case tcell.KeyLeft:
		e.Move(-1, 0)
		return true
	case tcell.KeyRight:
		e.Move(1, 0)
		return true
	case tcell.KeyHome:
		e.MoveTo(0, e.cursorY)
		return true
	case tcell.KeyEnd:
		e.MoveTo(e.doc.Width-1, e.cursorY)
		return true
	case tcell.KeyEnter:
		e.MoveTo(0, e.cursorY+1)
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if e.cursorX > 0 {
			e.Move(-1, 0)
			e.clearAt(e.cursorX, e.cursorY)
		}
		return true
	case tcell.KeyRune:
		e.typeRune(evt.Str())
		return true
	}
	return false
}

// typeRune writes a rune at the cursor with the current style and
// advances the cursor right (wraps to next row at the right edge).
func (e *Editor) typeRune(ch string) {
	e.lastRune = ch
	e.app.history.Begin()
	before := e.doc.At(e.cursorX, e.cursorY)
	after := Cell{Ch: ch, Style: e.style}
	if before != after {
		e.app.history.Record(e.cursorX, e.cursorY, before, after)
		e.doc.Set(e.cursorX, e.cursorY, after)
	}
	e.app.history.Commit()
	if e.cursorX < e.doc.Width-1 {
		e.Move(1, 0)
	} else if e.cursorY < e.doc.Height-1 {
		e.MoveTo(0, e.cursorY+1)
	} else {
		e.Refresh()
	}
}

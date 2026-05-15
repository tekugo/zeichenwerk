package main

import (
	"github.com/gdamore/tcell/v3"
)

// handleNormal processes key events while in Normal mode.
func (e *Editor) handleNormal(evt *tcell.EventKey) bool {
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
	case tcell.KeyPgUp:
		_, _, _, h := e.Content()
		e.PageScroll(-h)
		return true
	case tcell.KeyPgDn:
		_, _, _, h := e.Content()
		e.PageScroll(h)
		return true
	case tcell.KeyRune:
		switch evt.Str() {
		case "a":
			e.SetMode(ModeInsert)
			return true
		case "e":
			e.SetMode(ModeExtended)
			return true
		case "v":
			e.SetMode(ModeVisual)
			return true
		case "i":
			e.app.openGlyphPicker()
			return true
		case "s":
			e.cycleStyle(1)
			return true
		case "S":
			e.app.openStyleEditor(false)
			return true
		case "B":
			e.app.openBorderPicker()
			return true
		case "x":
			e.clearAt(e.cursorX, e.cursorY)
			if e.cursorX < e.doc.Width-1 {
				e.Move(1, 0)
			} else {
				e.Refresh()
			}
			return true
		case "f":
			e.floodFill(e.cursorX, e.cursorY, -1, -1, -1, -1)
			return true
		case "u":
			e.app.history.Undo(e.doc, e)
			return true
		case "p":
			e.app.register.Put(e.doc, e, e.cursorX, e.cursorY)
			e.Refresh()
			return true
		case "[":
			_, _, _, h := e.Content()
			e.PageScroll(-h / 2)
			return true
		case "]":
			_, _, _, h := e.Content()
			e.PageScroll(h / 2)
			return true
		}
	}
	return false
}

// clearAt resets a cell to the empty cell and records the change.
func (e *Editor) clearAt(x, y int) {
	before := e.doc.At(x, y)
	if before == EmptyCell {
		return
	}
	e.app.history.Begin()
	e.app.history.Record(x, y, before, EmptyCell)
	e.doc.Set(x, y, EmptyCell)
	e.app.history.Commit()
}

// cycleStyle advances the current style by `step` entries through the
// palette in alphabetical order ("default" first).
func (e *Editor) cycleStyle(step int) {
	names := paletteNames(e.doc)
	if len(names) == 0 {
		return
	}
	idx := 0
	for i, n := range names {
		if n == e.style {
			idx = i
			break
		}
	}
	idx = (idx + step + len(names)) % len(names)
	e.SetCurrentStyle(names[idx])
}

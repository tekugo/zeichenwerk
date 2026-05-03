package main

import (
	"github.com/gdamore/tcell/v3"
)

// handleVisual processes key events while in Visual mode.
func (e *Editor) handleVisual(evt *tcell.EventKey) bool {
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
	case tcell.KeyRune:
		switch evt.Str() {
		case "d":
			e.deleteSelection()
			return true
		case "s":
			e.cycleStyle(1)
			e.styleSelection(e.style)
			return true
		case "S":
			e.app.openStyleEditor(true)
			return true
		case "b":
			e.borderSelection(e.border)
			return true
		case "B":
			e.app.openBorderPicker()
			return true
		case "h":
			x1, y1, x2, y2 := e.selectionRect()
			midY := (y1 + y2) / 2
			e.drawHLine(x1, x2, midY)
			return true
		case "v":
			x1, y1, x2, y2 := e.selectionRect()
			midX := (x1 + x2) / 2
			e.drawVLine(midX, y1, y2)
			return true
		case "y":
			x1, y1, x2, y2 := e.selectionRect()
			e.app.register.Yank(e.doc, x1, y1, x2, y2)
			return true
		case "F":
			x1, y1, x2, y2 := e.selectionRect()
			e.floodFill(e.cursorX, e.cursorY, x1, y1, x2, y2)
			return true
		}
	}
	return false
}

// deleteSelection fills the rect with empty cells (default style + space).
func (e *Editor) deleteSelection() {
	x1, y1, x2, y2 := e.selectionRect()
	hist := e.app.history
	hist.Begin()
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			before := e.doc.At(x, y)
			if before == EmptyCell {
				continue
			}
			hist.Record(x, y, before, EmptyCell)
			e.doc.Set(x, y, EmptyCell)
		}
	}
	hist.Commit()
	e.Refresh()
}

// styleSelection re-styles every cell in the selection rect, preserving
// each cell's character.
func (e *Editor) styleSelection(name string) {
	x1, y1, x2, y2 := e.selectionRect()
	hist := e.app.history
	hist.Begin()
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			before := e.doc.At(x, y)
			if before.Style == name {
				continue
			}
			after := before
			after.Style = name
			hist.Record(x, y, before, after)
			e.doc.Set(x, y, after)
		}
	}
	hist.Commit()
	e.Refresh()
}

package main

import (
	"github.com/gdamore/tcell/v3"
)

// handleKey is the top-level key handler installed on the editor widget.
// It intercepts the drawing-mode paint keys first (so an active mode owns
// its key set and consumes events the editor would otherwise act on), then
// the application shortcuts. Anything not consumed falls through to the
// editor's built-in handler.
//
// Ctrl+M (Return) and Ctrl+B (tmux prefix) are deliberately NOT bound:
// terminals send Ctrl+M as KeyEnter, and Ctrl+B is reserved for tmux. The
// markdown preview and the drawing-mode toggles are reachable only through
// the command palette (Ctrl+P).
func (a *App) handleKey(ev *tcell.EventKey) bool {
	// Esc ends any active drawing mode. We do this before the per-mode
	// handlers so Esc can never be misinterpreted as a paint key.
	if ev.Key() == tcell.KeyEsc && (a.boxDraw || a.lineDraw) {
		a.boxDraw = false
		a.lineDraw = false
		a.refreshChrome()
		return true
	}

	if a.boxDraw {
		if a.handleBoxKey(ev) {
			return true
		}
	}
	if a.lineDraw {
		if a.handleLineKey(ev) {
			return true
		}
	}

	shift := ev.Modifiers()&tcell.ModShift != 0

	switch ev.Key() {
	case tcell.KeyCtrlP:
		a.ui.Commands().Open()
		return true
	case tcell.KeyCtrlO:
		a.openWithChooser()
		return true
	case tcell.KeyCtrlS:
		if shift {
			a.saveAs()
		} else {
			a.save()
		}
		return true
	case tcell.KeyCtrlK:
		a.deleteLine()
		return true
	case tcell.KeyCtrlQ:
		a.quit()
		return true
	}
	return false
}

// deleteLine removes the line under the cursor. With no selection support
// available we manipulate the line collection directly via SetContent.
func (a *App) deleteLine() {
	lines := a.editor.Lines()
	row := a.editor.Line()
	if row < 0 || row >= len(lines) {
		return
	}
	if len(lines) == 1 {
		a.editor.SetContent([]string{""})
		a.editor.MoveTo(0, 0)
		return
	}
	out := make([]string, 0, len(lines)-1)
	out = append(out, lines[:row]...)
	out = append(out, lines[row+1:]...)
	a.editor.SetContent(out)
	if row >= len(out) {
		row = len(out) - 1
	}
	a.editor.MoveTo(row, 0)
}

// quit asks for confirmation when the buffer is dirty; otherwise it exits.
func (a *App) quit() {
	if a.dirty {
		a.ui.Confirm("Quit?",
			"Buffer has unsaved changes. Quit without saving?",
			func() { a.ui.Quit() }, nil,
		)
		return
	}
	a.ui.Quit()
}

// registerCommands populates the palette with every action that is also
// reachable through a keyboard shortcut, so users can discover them by
// browsing the palette. Mode toggles, the border picker, and the markdown
// preview have no shortcut (they would clash with Enter / tmux); they are
// reachable only through the palette.
func (a *App) registerCommands() {
	c := a.ui.Commands()
	c.Register("File", "Open File", "Ctrl+O", a.openWithChooser)
	c.Register("File", "Save File", "Ctrl+S", a.save)
	c.Register("File", "Save File As", "Ctrl+Shift+S", a.saveAs)
	c.Register("File", "Quit", "Ctrl+Q", a.quit)

	c.Register("View", "Markdown Preview", "", a.showPreview)

	c.Register("Edit", "Delete Line", "Ctrl+K", a.deleteLine)

	c.Register("Drawing", "Toggle Box-Drawing (numpad)", "", a.toggleBoxDraw)
	c.Register("Drawing", "Toggle Line-Drawing (arrows)", "", a.toggleLineDraw)
	c.Register("Drawing", "Pick Border Style", "", a.pickBorder)
}


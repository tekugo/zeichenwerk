package main

import (
	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	. "github.com/tekugo/zeichenwerk/widgets"
)

const previewID = "md-preview"

func (a *App) showPreview() {
	if existing := Find(a.ui, previewID); existing != nil {
		a.ui.Close()
		return
	}

	popup := NewBuilder(a.theme).
		Box(previewID, "Markdown Preview").Hint(-1, -1).
		Styled("styled", a.editor.Text()).Hint(-1, -1).
		End().
		Container()

	_, _, w, h := a.ui.Bounds()
	pw, ph := w*7/10, h*7/10
	if pw < 40 {
		pw = w
	}
	if ph < 10 {
		ph = h
	}
	a.ui.Popup(-1, -1, pw, ph, popup)

	// Close on Esc while the preview owns focus.
	OnKey(popup, func(ev *tcell.EventKey) bool {
		if ev.Key() == tcell.KeyEsc {
			a.ui.Close()
			return true
		}
		return false
	})

	// Route key events to the popup container regardless of focus details.
	a.ui.SetFocus(previewID)
}

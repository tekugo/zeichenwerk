package main

import (
	"github.com/gdamore/tcell/v3"
)

// handleExtended processes key events while in Extended mode. Numpad-like
// keys insert box-drawing pieces from the current border family.
func (e *Editor) handleExtended(evt *tcell.EventKey) bool {
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
	case tcell.KeyRune:
		mask, ok := keypadMask(evt.Str())
		if !ok {
			return false
		}
		ch := runeForMask(e.border, mask)
		if ch == "" {
			return false
		}
		e.placeBoxRune(e.cursorX, e.cursorY, ch, e.border)
		if e.cursorX < e.doc.Width-1 {
			e.Move(1, 0)
		} else {
			e.Refresh()
		}
		return true
	}
	return false
}

// keypadMask maps a numpad-position key to a box-drawing connection mask
// (top, right, bottom, left bits).
func keypadMask(key string) (Mask, bool) {
	switch key {
	case "7":
		return MaskRight | MaskBottom, true
	case "8":
		return MaskLeft | MaskRight | MaskBottom, true
	case "9":
		return MaskLeft | MaskBottom, true
	case "4":
		return MaskTop | MaskRight | MaskBottom, true
	case "5":
		return MaskTop | MaskRight | MaskBottom | MaskLeft, true
	case "6":
		return MaskTop | MaskLeft | MaskBottom, true
	case "1":
		return MaskTop | MaskRight, true
	case "2":
		return MaskTop | MaskLeft | MaskRight, true
	case "3":
		return MaskTop | MaskLeft, true
	case "0":
		return MaskLeft | MaskRight, true
	case ".", ",":
		return MaskTop | MaskBottom, true
	}
	return 0, false
}

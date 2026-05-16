package main

import (
	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
	. "github.com/tekugo/zeichenwerk/widgets"
)

// Stroke directions encoded as a 4-bit mask: U|D|L|R.
const (
	strokeR = 1 << 0
	strokeL = 1 << 1
	strokeD = 1 << 2
	strokeU = 1 << 3
)

// numpad maps the 1-9 / 0 / ,/. keys to a directional mask. The numbers form
// a numpad-style block where 1 is the lower-left corner and 9 the upper-right.
//
//	7 ┌  8 ┬  9 ┐
//	4 ├  5 ┼  6 ┤
//	1 └  2 ┴  3 ┘
//	0 ─        , . │
var numpadMask = map[rune]int{
	'1': strokeU | strokeR,                     // └
	'2': strokeU | strokeL | strokeR,           // ┴
	'3': strokeU | strokeL,                     // ┘
	'4': strokeU | strokeD | strokeR,           // ├
	'5': strokeU | strokeD | strokeL | strokeR, // ┼
	'6': strokeU | strokeD | strokeL,           // ┤
	'7': strokeD | strokeR,                     // ┌
	'8': strokeD | strokeL | strokeR,           // ┬
	'9': strokeD | strokeL,                     // ┐
	'0': strokeL | strokeR,                     // ─
	',': strokeU | strokeD,                     // │
	'.': strokeU | strokeD,                     // │
}

// glyphMask reverses well-known box characters back to their stroke mask so
// successive paint operations can merge cleanly with existing strokes.
var glyphMask = map[rune]int{
	'─': strokeL | strokeR, '━': strokeL | strokeR, '═': strokeL | strokeR,
	'│': strokeU | strokeD, '┃': strokeU | strokeD, '║': strokeU | strokeD,
	'┌': strokeD | strokeR, '┏': strokeD | strokeR, '╔': strokeD | strokeR, '╭': strokeD | strokeR,
	'┐': strokeD | strokeL, '┓': strokeD | strokeL, '╗': strokeD | strokeL, '╮': strokeD | strokeL,
	'└': strokeU | strokeR, '┗': strokeU | strokeR, '╚': strokeU | strokeR, '╰': strokeU | strokeR,
	'┘': strokeU | strokeL, '┛': strokeU | strokeL, '╝': strokeU | strokeL, '╯': strokeU | strokeL,
	'┬': strokeD | strokeL | strokeR, '┳': strokeD | strokeL | strokeR, '╦': strokeD | strokeL | strokeR,
	'┴': strokeU | strokeL | strokeR, '┻': strokeU | strokeL | strokeR, '╩': strokeU | strokeL | strokeR,
	'├': strokeU | strokeD | strokeR, '┣': strokeU | strokeD | strokeR, '╠': strokeU | strokeD | strokeR,
	'┤': strokeU | strokeD | strokeL, '┫': strokeU | strokeD | strokeL, '╣': strokeU | strokeD | strokeL,
	'┼': strokeU | strokeD | strokeL | strokeR, '╋': strokeU | strokeD | strokeL | strokeR, '╬': strokeU | strokeD | strokeL | strokeR,
}

// borderGlyph picks the rune from the active theme border that best represents
// the given stroke mask. Single-stroke masks fall back to the matching axis
// (horizontal or vertical) glyph so a half-stroke still renders meaningfully.
func borderGlyph(b *Border, mask int) rune {
	if b == nil || mask == 0 {
		return ' '
	}
	pick := func(s string, fallback rune) rune {
		if s == "" {
			return fallback
		}
		for _, r := range s {
			return r
		}
		return fallback
	}
	switch mask {
	case strokeL | strokeR, strokeL, strokeR:
		return pick(b.InnerH, '─')
	case strokeU | strokeD, strokeU, strokeD:
		return pick(b.InnerV, '│')
	case strokeD | strokeR:
		return pick(b.TopLeft, '┌')
	case strokeD | strokeL:
		return pick(b.TopRight, '┐')
	case strokeU | strokeR:
		return pick(b.BottomLeft, '└')
	case strokeU | strokeL:
		return pick(b.BottomRight, '┘')
	case strokeD | strokeL | strokeR:
		return pick(b.InnerTopT, '┬')
	case strokeU | strokeL | strokeR:
		return pick(b.InnerBottomT, '┴')
	case strokeU | strokeD | strokeR:
		return pick(b.InnerLeftT, '├')
	case strokeU | strokeD | strokeL:
		return pick(b.InnerRightT, '┤')
	}
	return pick(b.InnerX, '┼')
}

// activeBorder returns the currently selected theme border, falling back to
// the first available one if the configured name has been removed from the
// theme.
func (a *App) activeBorder() *Border {
	if b := a.theme.Border(a.border); b != nil {
		return b
	}
	if names := a.theme.BorderNames(); len(names) > 0 {
		a.border = names[0]
		return a.theme.Border(a.border)
	}
	return nil
}

// toggleBoxDraw flips box-drawing (numpad glyph) mode and refreshes the
// status line so the user always sees the active state.
func (a *App) toggleBoxDraw() {
	a.boxDraw = !a.boxDraw
	a.refreshChrome()
}

// toggleLineDraw flips line-drawing (arrow-key paint) mode and refreshes the
// status line.
func (a *App) toggleLineDraw() {
	a.lineDraw = !a.lineDraw
	a.refreshChrome()
}

// handleBoxKey runs while box-drawing mode is active. It maps the 1-9 / 0 /
// `,` / `.` keys (laid out as a numpad block) onto the corresponding box
// glyphs from the active theme border and overwrites the rune under the
// cursor. All other keys fall through to the editor.
func (a *App) handleBoxKey(ev *tcell.EventKey) bool {
	if ev.Key() != tcell.KeyRune {
		return false
	}
	border := a.activeBorder()
	if border == nil {
		return false
	}
	s := ev.Str()
	if s == "" {
		return false
	}
	r := []rune(s)[0]
	mask, ok := numpadMask[r]
	if !ok {
		return false
	}
	a.paintRune(border, mask)
	return true
}

// handleLineKey runs while line-drawing mode is active. Arrow keys paint
// the connecting glyph at the current cell, advance the cursor by one cell
// in that direction, and merge the entry stroke into the destination cell.
// The arrow keys are always consumed when this mode is on so the editor's
// own cursor movement does not run in parallel.
func (a *App) handleLineKey(ev *tcell.EventKey) bool {
	var dir int
	switch ev.Key() {
	case tcell.KeyRight:
		dir = strokeR
	case tcell.KeyLeft:
		dir = strokeL
	case tcell.KeyDown:
		dir = strokeD
	case tcell.KeyUp:
		dir = strokeU
	default:
		return false
	}
	if border := a.activeBorder(); border != nil {
		a.paintWithDir(border, dir)
	}
	// Consume the arrow event regardless: if no border is registered the
	// user still expects the cursor not to drift while the mode is on.
	return true
}

// paintRune overwrites the rune under the cursor with the glyph corresponding
// to the supplied mask. Strokes already painted at this cell during the
// session are merged in so the user can layer glyphs (e.g. numpad-5 on a
// previously painted `│` yields `┼`).
func (a *App) paintRune(b *Border, mask int) {
	row, col := a.editor.Line(), a.editor.Column()
	mask |= a.cellMask(row, col)
	a.paintMask[[2]int{row, col}] = mask
	overwrite(a.editor, borderGlyph(b, mask))
}

// paintWithDir handles arrow-key painting: add the outgoing stroke to the
// current cell, advance the cursor by exactly one cell in the painted
// direction, and merge the entry stroke into the destination cell.
//
// The function leaves the editor cursor on the destination cell. We never
// rely on Editor.Right/Left/Up/Down for the move because overwriteAt has
// already advanced the cursor by one column as a side effect of Insert.
// Computing the destination directly and calling MoveTo keeps the logic
// independent of side-effects in the underlying editing primitives.
//
// Stroke masks come from the paintMask sidecar — see the App field doc for
// why we can't rely on reverse-lookup from the rendered glyph alone.
func (a *App) paintWithDir(b *Border, dir int) {
	row, col := a.editor.Line(), a.editor.Column()

	// Outgoing stroke on the current cell.
	mask := dir | a.cellMask(row, col)
	a.paintMask[[2]int{row, col}] = mask
	overwriteAt(a.editor, row, col, borderGlyph(b, mask))

	// Destination is exactly one cell away, clamped at document edges.
	nRow, nCol := neighbor(a.editor.Lines(), row, col, dir)
	if nRow == row && nCol == col {
		a.editor.MoveTo(row, col)
		return
	}

	// Incoming stroke on the destination cell.
	entry := opposite(dir) | a.cellMask(nRow, nCol)
	a.paintMask[[2]int{nRow, nCol}] = entry
	overwriteAt(a.editor, nRow, nCol, borderGlyph(b, entry))
	a.editor.MoveTo(nRow, nCol)
}

// cellMask returns the stroke mask for a cell. Cells painted during this
// session live in the sidecar (paintMask) and return their exact mask.
// For untracked cells we fall back to glyphMask — useful when the user
// crosses a hand-drawn or previously loaded box character.
func (a *App) cellMask(row, col int) int {
	if m, ok := a.paintMask[[2]int{row, col}]; ok {
		return m
	}
	if m, ok := glyphMask[runeAt(a.editor.Lines(), row, col)]; ok {
		return m
	}
	return 0
}

// neighbor returns the cell coordinate one step from (row, col) in the given
// direction, clamped at the document boundaries. The right direction never
// clamps because overwriteAt pads lines with spaces on demand.
func neighbor(lines []string, row, col, dir int) (int, int) {
	switch dir {
	case strokeR:
		return row, col + 1
	case strokeL:
		if col > 0 {
			return row, col - 1
		}
	case strokeD:
		if row+1 < len(lines) {
			return row + 1, col
		}
	case strokeU:
		if row > 0 {
			return row - 1, col
		}
	}
	return row, col
}

func opposite(dir int) int {
	switch dir {
	case strokeR:
		return strokeL
	case strokeL:
		return strokeR
	case strokeD:
		return strokeU
	case strokeU:
		return strokeD
	}
	return 0
}

// runeAt safely reads the rune at (row, col) in `lines`. Out-of-range
// positions return a space so callers can treat them as empty cells.
func runeAt(lines []string, row, col int) rune {
	if row < 0 || row >= len(lines) {
		return ' '
	}
	rs := []rune(lines[row])
	if col < 0 || col >= len(rs) {
		return ' '
	}
	return rs[col]
}

// overwrite replaces the rune under the cursor with `r` and advances the
// cursor by one column (the natural result of insert).
func overwrite(e *Editor, r rune) {
	row, col := e.Line(), e.Column()
	overwriteAt(e, row, col, r)
}

// overwriteAt replaces the rune at (row, col) with `r`, advancing the cursor
// to (row, col+1). Beyond the end of the line we pad with spaces so the user
// can draw freely on an empty canvas.
func overwriteAt(e *Editor, row, col int, r rune) {
	e.MoveTo(row, col)
	line := ""
	if row < len(e.Lines()) {
		line = e.Lines()[row]
	}
	rs := []rune(line)
	for col > len(rs) {
		e.Insert(' ')
		rs = append(rs, ' ')
	}
	if col < len(rs) {
		// Cell already occupied — remove it before inserting the new rune.
		e.MoveTo(row, col)
		e.DeleteForward()
	}
	e.Insert(r)
}

// pickBorder opens a popup listing every border style registered in the
// current theme and switches the active style to the chosen entry.
func (a *App) pickBorder() {
	names := a.theme.BorderNames()
	if len(names) == 0 {
		a.notify("no border styles registered in this theme")
		return
	}

	const popupID = "border-picker"
	if existing := Find(a.ui, popupID); existing != nil {
		a.ui.Close()
		return
	}

	b := a.ui.NewBuilder()
	b.Box(popupID, "Box-drawing border").Hint(40, len(names)+4).
		List("border-list", names...).Hint(-1, -1).
		End()
	popup := b.Container()

	list := MustFind[*List](popup, "border-list")
	for i, n := range names {
		if n == a.border {
			list.Select(i)
			break
		}
	}
	OnKey(list, func(ev *tcell.EventKey) bool {
		switch ev.Key() {
		case tcell.KeyEsc:
			a.ui.Close()
			return true
		case tcell.KeyEnter:
			a.border = names[list.Selected()]
			a.ui.Close()
			a.refreshChrome()
			return true
		}
		return false
	})

	a.ui.Popup(-1, -1, 0, 0, popup)
	a.ui.SetFocus("border-list")
}

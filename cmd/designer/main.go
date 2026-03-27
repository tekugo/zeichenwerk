// designer is a full-screen TUI canvas editor with VIM-style modal editing.
// See spec/designer.md for the full specification.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v3"
	zw "github.com/tekugo/zeichenwerk"
)

// ============================================================
// Mode
// ============================================================

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeDraw
	ModeVisual
	ModeCommand
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeDraw:
		return "DRAW"
	case ModeVisual:
		return "VISUAL"
	case ModeCommand:
		return "COMMAND"
	}
	return "?"
}

// ============================================================
// Palettes
// ============================================================

type Palette struct {
	name  string
	chars []rune
}

// evRune extracts the first rune from a KeyRune event string.
func evRune(ev *tcell.EventKey) rune {
	s := ev.Str()
	if len(s) == 0 {
		return 0
	}
	return []rune(s)[0]
}

// numpadIndex maps a draw-mode key to a palette character index.
//
//	7 8 9  →  upper-left corner, upper T, upper-right corner
//	4 5 6  →  left T, inner cross, right T
//	1 2 3  →  lower-left corner, lower T, lower-right corner
//	-      →  horizontal line
//	*      →  vertical line
func numpadIndex(r rune) int {
	switch r {
	case '7':
		return 2
	case '8':
		return 8
	case '9':
		return 3
	case '4':
		return 6
	case '5':
		return 10
	case '6':
		return 7
	case '1':
		return 4
	case '2':
		return 9
	case '3':
		return 5
	case '-':
		return 0
	case '*':
		return 1
	}
	return -1
}

var builtinPalettes = []Palette{
	{name: "thin", chars: []rune("─│┌┐└┘├┤┬┴┼╴╶")},
	{name: "round", chars: []rune("─│╭╮╰╯├┤┬┴┼╴╶")},
	{name: "double", chars: []rune("═║╔╗╚╝╠╣╦╩╬╸╺")},
	{name: "block", chars: []rune("█▓▒░▄▀▌▐▇▆▅▃▂")},
}

// ============================================================
// Designer – application state
// ============================================================

type Designer struct {
	// Document
	page     *Page
	docName  string
	filename string
	modified bool

	// Editing state
	mode       Mode
	curX, curY int
	ancX, ancY int // VISUAL mode anchor

	// DRAW mode
	palettes   []Palette
	palIdx     int
	showNumPad bool

	// COMMAND mode
	cmdBuf string

	// Transient status message
	msg string

	// Clipboard (VISUAL yank)
	clipboard [][]Cell

	// Named style registry – zero-size Canvas used as a style map
	styleRegistry *zw.Canvas
	styleOrder    []string // ordered list of style names
	drawStyle     string   // active drawing style name

	// Widget references (set after UI construction)
	canvas *DesignerCanvas
	status *DesignerStatus
	ui     *zw.UI
}

func newDesigner(page *Page, docName string) *Designer {
	sc, order := newStyleCanvas()
	return &Designer{
		page:          page,
		docName:       docName,
		mode:          ModeNormal,
		palettes:      builtinPalettes,
		styleRegistry: sc,
		styleOrder:    order,
		drawStyle:     "default",
	}
}

// ============================================================
// Cell writes – keep Page (data model) and Canvas (display buffer) in sync
// ============================================================

// setCell writes a cell to both the page data model and the canvas display
// buffer, keeping them in sync. The style *pointer* stored in the canvas
// means future in-place style edits are reflected automatically.
func (d *Designer) setCell(x, y int, cell Cell) {
	d.page.set(x, y, cell)
	s := d.styleRegistry.Style(cell.style)
	ch := ""
	if !cell.isEmpty() {
		ch = string(cell.ch)
	}
	d.canvas.SetCell(x, y, ch, s)
	d.modified = true
}

// clearCell erases the cell at (x,y) in both the page and the canvas buffer.
func (d *Designer) clearCell(x, y int) {
	d.page.set(x, y, Cell{})
	d.canvas.SetCell(x, y, "", nil)
	d.modified = true
}

// syncPageToCanvas populates the canvas cell buffer from the page data model.
// Called once after loading a document.
func (d *Designer) syncPageToCanvas() {
	for y, row := range d.page.cells {
		for x, cell := range row {
			s := d.styleRegistry.Style(cell.style)
			ch := ""
			if !cell.isEmpty() {
				ch = string(cell.ch)
			}
			d.canvas.SetCell(x, y, ch, s)
		}
	}
}

// ============================================================
// DesignerCanvas – extends zw.Canvas with VIM-style modal editing
// ============================================================

// DesignerCanvas embeds *zw.Canvas to reuse its cell buffer, SetCell API,
// and Render pipeline. We replace the built-in key handler, extend Render
// to overlay the VISUAL selection and the DRAW mode numpad, and override
// Cursor to reflect the designer's five-mode cursor behaviour.
type DesignerCanvas struct {
	*zw.Canvas
	d *Designer
}

func newDesignerCanvas(d *Designer, width, height int) *DesignerCanvas {
	canvas := zw.NewCanvas("canvas", width, height)
	canvas.ClearKeyHandlers()           // remove the built-in NORMAL/INSERT handler
	canvas.SetHint(0, -1)               // fill all remaining vertical space
	canvas.SetStyle("", zw.NewStyle("").WithColors("$fg0", "$bg0"))

	dc := &DesignerCanvas{Canvas: canvas, d: d}
	zw.OnKey(dc, d.handleCanvasKey)
	return dc
}

// Cursor overrides zw.Canvas.Cursor to reflect the designer's full mode set.
func (dc *DesignerCanvas) Cursor() (int, int, string) {
	if !dc.Flag("focused") || dc.d.mode == ModeCommand {
		return -1, -1, ""
	}
	switch dc.d.mode {
	case ModeInsert:
		return dc.d.curX, dc.d.curY, "*|"
	default:
		return dc.d.curX, dc.d.curY, "#"
	}
}

// Render delegates cell painting to zw.Canvas.Render, then overlays the
// VISUAL selection highlight and the optional DRAW mode numpad.
func (dc *DesignerCanvas) Render(r *zw.Renderer) {
	// Paint all cells via the inherited canvas renderer.
	dc.Canvas.Render(r)

	x0, y0, cw, ch := dc.Content()

	// VISUAL selection – re-render the selected rectangle with inverted colours.
	if dc.d.mode == ModeVisual {
		bx0, by0, bx1, by1 := dc.d.selectionBounds()
		for row := by0; row <= by1 && row < ch; row++ {
			for col := bx0; col <= bx1 && col < cw; col++ {
				cell := dc.d.page.at(col, row)
				fg, bg, font := dc.d.resolveStyle(cell.style)
				fg, bg = dc.d.selectionSwap(fg, bg)
				r.Set(fg, bg, font)
				glyph := " "
				if !cell.isEmpty() {
					glyph = string(cell.ch)
				}
				r.Put(x0+col, y0+row, glyph)
			}
		}
	}

	// DRAW mode numpad overlay (top-right corner, no focus change).
	if dc.d.mode == ModeDraw && dc.d.showNumPad {
		dc.d.renderNumPad(r, x0, y0, cw)
	}
}

// ============================================================
// DesignerStatus – custom 1-row status bar
// ============================================================

type DesignerStatus struct {
	zw.Component
	d *Designer
}

func newDesignerStatus(d *Designer) *DesignerStatus {
	ds := &DesignerStatus{d: d}
	ds.Component = *zw.NewComponent("status")
	ds.SetFlag("focusable", true)
	ds.SetHint(0, 1)
	ds.SetStyle("", zw.NewStyle("").WithColors("$bg0", "$fg1"))
	zw.OnKey(ds, d.handleStatusKey)
	return ds
}

func (ds *DesignerStatus) Cursor() (int, int, string) {
	if ds.d.mode == ModeCommand && ds.Flag("focused") {
		return 1 + len([]rune(ds.d.cmdBuf)), 0, "|"
	}
	return -1, -1, ""
}

func (ds *DesignerStatus) Render(r *zw.Renderer) {
	ds.Component.Render(r)
	cx, cy, cw, _ := ds.Content()

	r.Set("$bg0", "$fg1", "")
	r.Fill(cx, cy, cw, 1, " ")

	var text string
	switch {
	case ds.d.mode == ModeCommand:
		text = ":" + ds.d.cmdBuf
	case ds.d.msg != "":
		text = " " + ds.d.msg
	default:
		pal := ds.d.palettes[ds.d.palIdx]
		sty := ds.d.drawStyle
		if sty == "" {
			sty = "default"
		}
		mod := ""
		if ds.d.modified {
			mod = " [+]"
		}
		text = fmt.Sprintf(" %s  palette:%s  style:%s  %d,%d / %dx%d%s",
			ds.d.mode.String(), pal.name, sty,
			ds.d.curX, ds.d.curY, ds.d.page.width, ds.d.page.height, mod)
	}

	r.Set("$bg0", "$fg1", "")
	r.Text(cx, cy, text, cw)
}

// ============================================================
// Style helpers
// ============================================================

func (d *Designer) resolveStyle(name string) (fg, bg, font string) {
	if name == "" {
		name = "default"
	}
	s := d.styleRegistry.Style(name)
	return s.Foreground(), s.Background(), s.Font()
}

// selectionSwap inverts fg/bg for the VISUAL selection highlight,
// falling back to the canvas theme colours for unstyled cells.
func (d *Designer) selectionSwap(fg, bg string) (string, string) {
	if fg == "" {
		fg = "$fg0"
	}
	if bg == "" {
		bg = "$bg0"
	}
	return bg, fg
}

func (d *Designer) styleParentName(name string) string {
	s := d.styleRegistry.Style(name)
	target := s.Parent()
	if target == nil {
		return ""
	}
	for _, n := range d.styleOrder {
		if d.styleRegistry.Style(n) == target {
			return n
		}
	}
	return ""
}

func (d *Designer) styleDepth(name string) int {
	s := d.styleRegistry.Style(name)
	depth := 0
	p := s.Parent()
	for p != nil && depth < 32 {
		depth++
		p = p.Parent()
	}
	return depth
}

func (d *Designer) insertIntoStyleOrder(name string) {
	d.styleOrder = append(d.styleOrder, name)
	d.sortStyleOrder()
}

func (d *Designer) removeFromStyleOrder(name string) {
	for i, n := range d.styleOrder {
		if n == name {
			d.styleOrder = append(d.styleOrder[:i], d.styleOrder[i+1:]...)
			return
		}
	}
}

func (d *Designer) sortStyleOrder() {
	sort.Slice(d.styleOrder, func(i, j int) bool {
		if d.styleOrder[i] == "default" {
			return true
		}
		if d.styleOrder[j] == "default" {
			return false
		}
		return d.styleOrder[i] < d.styleOrder[j]
	})
}

func (d *Designer) renameCells(oldName, newName string) {
	for y := range d.page.cells {
		for x := range d.page.cells[y] {
			if d.page.cells[y][x].style == oldName {
				d.page.cells[y][x].style = newName
				// Re-point the canvas buffer cell to the (same) style pointer
				// registered under the new name. The pointer itself is the same
				// object; this just fixes the lookup key in the page model.
			}
		}
	}
}

// styleListItems returns display strings for the style List popup.
func (d *Designer) styleListItems() []string {
	items := make([]string, len(d.styleOrder))
	for i, name := range d.styleOrder {
		depth := d.styleDepth(name)
		indent := strings.Repeat("  ", depth)
		parentStr := ""
		if pn := d.styleParentName(name); pn != "" {
			parentStr = " ↑" + pn
		}
		active := ""
		if name == d.drawStyle || (d.drawStyle == "" && name == "default") {
			active = " *"
		}
		items[i] = indent + name + parentStr + active
	}
	return items
}

// ============================================================
// NumPad overlay (rendered inline by DesignerCanvas.Render)
// ============================================================

func (d *Designer) renderNumPad(r *zw.Renderer, x0, y0, canvasW int) {
	pal := d.palettes[d.palIdx]

	type numKey struct {
		label string
		idx   int
	}
	keys := []numKey{
		{"7", 2}, {"8", 8}, {"9", 3},
		{"4", 6}, {"5", 10}, {"6", 7},
		{"1", 4}, {"2", 9}, {"3", 5},
		{"-", 0}, {"*", 1},
	}

	const colW = 5
	const numCols = 3
	overlayW := numCols*colW + 2
	ox := x0 + canvasW - overlayW
	if ox < x0 {
		ox = x0
	}

	r.Set("$bg0", "$yellow", "bold")
	r.Text(ox, y0, fmt.Sprintf(" %-*s", overlayW-1, pal.name), overlayW)

	r.Set("$bg0", "$yellow", "")
	for i, k := range keys {
		col := i % numCols
		row := i/numCols + 1
		x := ox + col*colW
		y := y0 + row
		ch := ' '
		if k.idx < len(pal.chars) {
			ch = pal.chars[k.idx]
		}
		r.Text(x, y, fmt.Sprintf("%s:%-2c", k.label, ch), colW)
	}
}

// ============================================================
// Key handlers – canvas (all modes except COMMAND)
// ============================================================

func (d *Designer) handleCanvasKey(w zw.Widget, ev *tcell.EventKey) bool {
	d.msg = ""
	switch d.mode {
	case ModeNormal:
		return d.normalKey(ev)
	case ModeInsert:
		return d.insertKey(ev)
	case ModeDraw:
		return d.drawKey(ev)
	case ModeVisual:
		return d.visualKey(ev)
	}
	return false
}

func (d *Designer) move(dx, dy int) {
	d.curX = max(0, min(d.page.width-1, d.curX+dx))
	d.curY = max(0, min(d.page.height-1, d.curY+dy))
	d.canvas.SetCursor(d.curX, d.curY)
	d.canvas.Refresh()
}

func (d *Designer) normalKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyLeft:
		d.move(-1, 0)
	case tcell.KeyRight:
		d.move(1, 0)
	case tcell.KeyUp:
		d.move(0, -1)
	case tcell.KeyDown:
		d.move(0, 1)
	case tcell.KeyHome:
		d.curX = 0
		d.canvas.SetCursor(d.curX, d.curY)
		d.canvas.Refresh()
	case tcell.KeyEnd:
		d.curX = d.page.width - 1
		d.canvas.SetCursor(d.curX, d.curY)
		d.canvas.Refresh()
	case tcell.KeyEsc:
		return false // let the UI layer system close a popup if present
	case tcell.KeyRune:
		switch evRune(ev) {
		case 'h':
			d.move(-1, 0)
		case 'j':
			d.move(0, 1)
		case 'k':
			d.move(0, -1)
		case 'l':
			d.move(1, 0)
		case '0':
			d.curX = 0
			d.canvas.SetCursor(d.curX, d.curY)
			d.canvas.Refresh()
		case '$':
			d.curX = d.page.width - 1
			d.canvas.SetCursor(d.curX, d.curY)
			d.canvas.Refresh()
		case 'g':
			d.curY = 0
			d.canvas.SetCursor(d.curX, d.curY)
			d.canvas.Refresh()
		case 'G':
			d.curY = d.page.height - 1
			d.canvas.SetCursor(d.curX, d.curY)
			d.canvas.Refresh()
		case 'i':
			d.mode = ModeInsert
			d.canvas.Refresh()
		case 'd':
			d.mode = ModeDraw
			d.canvas.Refresh()
		case 'v':
			d.mode = ModeVisual
			d.ancX, d.ancY = d.curX, d.curY
			d.canvas.Refresh()
		case ':':
			d.mode = ModeCommand
			d.cmdBuf = ""
			d.ui.Focus(d.status)
			d.status.Refresh()
		case 'b':
			d.openPalettePopup()
		case 's':
			d.openStylePopup()
		case 'p':
			d.paste()
		case 'x':
			d.clearCell(d.curX, d.curY)
			d.canvas.Refresh()
		case '?':
			d.openHelpPopup()
		default:
			return false
		}
	default:
		return false
	}
	return true
}

func (d *Designer) insertKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEsc:
		d.mode = ModeNormal
		d.canvas.Refresh()
	case tcell.KeyLeft:
		d.move(-1, 0)
	case tcell.KeyRight:
		d.move(1, 0)
	case tcell.KeyUp:
		d.move(0, -1)
	case tcell.KeyDown:
		d.move(0, 1)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if d.curX > 0 {
			d.curX--
			d.canvas.SetCursor(d.curX, d.curY)
			d.clearCell(d.curX, d.curY)
			d.canvas.Refresh()
		}
	case tcell.KeyRune:
		r := evRune(ev)
		if unicode.IsPrint(r) {
			d.setCell(d.curX, d.curY, Cell{ch: r, style: d.drawStyle})
			if d.curX < d.page.width-1 {
				d.curX++
			} else if d.curY < d.page.height-1 {
				d.curX = 0
				d.curY++
			}
			d.canvas.SetCursor(d.curX, d.curY)
			d.canvas.Refresh()
		}
	default:
		return false
	}
	return true
}

func (d *Designer) drawKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyRune {
		r := evRune(ev)
		if idx := numpadIndex(r); idx >= 0 {
			pal := d.palettes[d.palIdx]
			if idx < len(pal.chars) {
				d.setCell(d.curX, d.curY, Cell{ch: pal.chars[idx], style: d.drawStyle})
				if r == '*' {
					if d.curY < d.page.height-1 {
						d.curY++
					}
				} else {
					if d.curX < d.page.width-1 {
						d.curX++
					}
				}
				d.canvas.SetCursor(d.curX, d.curY)
				d.canvas.Refresh()
			}
			return true
		}
	}

	switch ev.Key() {
	case tcell.KeyEsc:
		d.mode = ModeNormal
		d.showNumPad = false
		d.canvas.Refresh()
	case tcell.KeyLeft:
		d.move(-1, 0)
	case tcell.KeyRight:
		d.move(1, 0)
	case tcell.KeyUp:
		d.move(0, -1)
	case tcell.KeyDown:
		d.move(0, 1)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if d.curX > 0 {
			d.curX--
			d.canvas.SetCursor(d.curX, d.curY)
			d.clearCell(d.curX, d.curY)
			d.canvas.Refresh()
		}
	case tcell.KeyRune:
		switch evRune(ev) {
		case 'h':
			d.move(-1, 0)
		case 'j':
			d.move(0, 1)
		case 'k':
			d.move(0, -1)
		case 'l':
			d.move(1, 0)
		case 's':
			d.showNumPad = !d.showNumPad
			d.canvas.Refresh()
		case 'b':
			d.openPalettePopup()
		case ' ':
			d.clearCell(d.curX, d.curY)
			if d.curX < d.page.width-1 {
				d.curX++
			}
			d.canvas.SetCursor(d.curX, d.curY)
			d.canvas.Refresh()
		default:
			return false
		}
	default:
		return false
	}
	return true
}

func (d *Designer) visualKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEsc:
		d.mode = ModeNormal
		d.canvas.Refresh()
	case tcell.KeyLeft:
		d.move(-1, 0)
	case tcell.KeyRight:
		d.move(1, 0)
	case tcell.KeyUp:
		d.move(0, -1)
	case tcell.KeyDown:
		d.move(0, 1)
	case tcell.KeyRune:
		switch evRune(ev) {
		case 'h':
			d.move(-1, 0)
		case 'j':
			d.move(0, 1)
		case 'k':
			d.move(0, -1)
		case 'l':
			d.move(1, 0)
		case 'y':
			d.yank()
		case 'd':
			d.deleteSelection()
		case 'b':
			d.drawBox()
		default:
			return false
		}
	default:
		return false
	}
	return true
}

// ============================================================
// Key handler – status bar (COMMAND mode)
// ============================================================

func (d *Designer) handleStatusKey(w zw.Widget, ev *tcell.EventKey) bool {
	if d.mode != ModeCommand {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEsc:
		d.mode = ModeNormal
		d.cmdBuf = ""
		d.ui.Focus(d.canvas)
		d.canvas.Refresh()
	case tcell.KeyEnter:
		cmd := d.cmdBuf
		d.cmdBuf = ""
		d.mode = ModeNormal
		d.ui.Focus(d.canvas)
		d.execCommand(cmd)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		runes := []rune(d.cmdBuf)
		if len(runes) > 0 {
			d.cmdBuf = string(runes[:len(runes)-1])
			d.status.Refresh()
		} else {
			d.mode = ModeNormal
			d.ui.Focus(d.canvas)
			d.canvas.Refresh()
		}
	case tcell.KeyRune:
		d.cmdBuf += string(evRune(ev))
		d.status.Refresh()
	default:
		return false
	}
	return true
}

// ============================================================
// COMMAND execution
// ============================================================

func (d *Designer) execCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	switch {
	case cmd == "q":
		if d.modified {
			d.msg = "Unsaved changes — use :q! to force quit or :wq to save and quit"
			d.status.Refresh()
			return
		}
		d.ui.Quit()

	case cmd == "q!":
		d.ui.Quit()

	case cmd == "w":
		if d.filename == "" {
			d.msg = "No filename — use :w <filename>"
			d.status.Refresh()
			return
		}
		if err := saveDoc(d.filename, d.docName, d.page, d.styleRegistry, d.styleOrder); err != nil {
			d.msg = fmt.Sprintf("Write error: %v", err)
		} else {
			d.modified = false
			d.msg = fmt.Sprintf("Written: %s", d.filename)
		}
		d.status.Refresh()

	case strings.HasPrefix(cmd, "w "):
		name := strings.TrimSpace(cmd[2:])
		if name == "" {
			d.msg = "Usage: :w <filename>"
			d.status.Refresh()
			return
		}
		d.filename = name
		if err := saveDoc(d.filename, d.docName, d.page, d.styleRegistry, d.styleOrder); err != nil {
			d.msg = fmt.Sprintf("Write error: %v", err)
		} else {
			d.modified = false
			d.msg = fmt.Sprintf("Written: %s", d.filename)
		}
		d.status.Refresh()

	case cmd == "wq":
		if d.filename == "" {
			d.msg = "No filename — use :w <filename> first"
			d.status.Refresh()
			return
		}
		if err := saveDoc(d.filename, d.docName, d.page, d.styleRegistry, d.styleOrder); err != nil {
			d.msg = fmt.Sprintf("Write error: %v", err)
			d.status.Refresh()
			return
		}
		d.ui.Quit()

	case strings.HasPrefix(cmd, "e "):
		filename := strings.TrimSpace(cmd[2:])
		if filename == "" {
			d.msg = "Usage: :e <filename>"
			d.status.Refresh()
			return
		}
		name, page, registry, order, err := loadDoc(filename)
		if err != nil {
			d.msg = fmt.Sprintf("Read error: %v", err)
			d.status.Refresh()
			return
		}
		d.filename = filename
		d.docName = name
		d.page = page
		d.styleRegistry = registry
		d.styleOrder = order
		d.drawStyle = "default"
		d.curX, d.curY = 0, 0
		d.canvas.SetCursor(0, 0)
		d.modified = false
		d.msg = fmt.Sprintf("Loaded: %s", filename)
		d.syncPageToCanvas()
		d.canvas.Refresh()
		d.status.Refresh()

	default:
		if cmd != "" {
			d.msg = fmt.Sprintf("Unknown command: %s", cmd)
			d.status.Refresh()
		}
	}
}

// ============================================================
// Clipboard operations
// ============================================================

func (d *Designer) selectionBounds() (x0, y0, x1, y1 int) {
	x0, x1 = d.ancX, d.curX
	y0, y1 = d.ancY, d.curY
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return
}

func (d *Designer) yank() {
	x0, y0, x1, y1 := d.selectionBounds()
	h := y1 - y0 + 1
	w := x1 - x0 + 1
	buf := make([][]Cell, h)
	for r := range buf {
		buf[r] = make([]Cell, w)
		for c := range buf[r] {
			buf[r][c] = d.page.at(x0+c, y0+r)
		}
	}
	d.clipboard = buf
	d.mode = ModeNormal
	d.msg = fmt.Sprintf("Yanked %dx%d", w, h)
	d.canvas.Refresh()
	d.status.Refresh()
}

func (d *Designer) deleteSelection() {
	x0, y0, x1, y1 := d.selectionBounds()
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			d.clearCell(x, y)
		}
	}
	d.mode = ModeNormal
	d.canvas.Refresh()
}

func (d *Designer) paste() {
	if d.clipboard == nil {
		d.msg = "Nothing to paste"
		d.status.Refresh()
		return
	}
	for r, row := range d.clipboard {
		for c, cell := range row {
			d.setCell(d.curX+c, d.curY+r, cell)
		}
	}
	d.canvas.Refresh()
}

// drawBox draws a border around the VISUAL selection using the active palette.
func (d *Designer) drawBox() {
	x0, y0, x1, y1 := d.selectionBounds()
	if x1-x0 < 1 || y1-y0 < 1 {
		d.msg = "Selection too small for a box (need at least 2×2)"
		d.status.Refresh()
		return
	}
	pal := d.palettes[d.palIdx]
	if len(pal.chars) < 6 {
		d.msg = "Palette has too few characters for box drawing"
		d.status.Refresh()
		return
	}
	horiz, vert := pal.chars[0], pal.chars[1]
	tlc, trc := pal.chars[2], pal.chars[3]
	blc, brc := pal.chars[4], pal.chars[5]
	st := d.drawStyle

	for x := x0 + 1; x < x1; x++ {
		d.setCell(x, y0, Cell{ch: horiz, style: st})
		d.setCell(x, y1, Cell{ch: horiz, style: st})
	}
	for y := y0 + 1; y < y1; y++ {
		d.setCell(x0, y, Cell{ch: vert, style: st})
		d.setCell(x1, y, Cell{ch: vert, style: st})
	}
	d.setCell(x0, y0, Cell{ch: tlc, style: st})
	d.setCell(x1, y0, Cell{ch: trc, style: st})
	d.setCell(x0, y1, Cell{ch: blc, style: st})
	d.setCell(x1, y1, Cell{ch: brc, style: st})

	d.mode = ModeNormal
	d.canvas.Refresh()
}

// ============================================================
// Style popup
// ============================================================

func (d *Designer) openStylePopup() {
	items := d.styleListItems()
	popup := d.ui.NewBuilder().
		Dialog("style-popup", "Styles").
		List("style-list", items...).
		Container()

	list := zw.Find(popup, "style-list").(*zw.List)
	list.SetHint(42, min(len(d.styleOrder)+2, 20))
	list.SetQuickSearch(false)

	for i, n := range d.styleOrder {
		if n == d.drawStyle {
			list.Select(i)
			break
		}
	}

	list.On("activate", func(w zw.Widget, event string, data ...any) bool {
		if idx, ok := data[0].(int); ok && idx < len(d.styleOrder) {
			d.drawStyle = d.styleOrder[idx]
		}
		d.ui.Close()
		d.canvas.Refresh()
		d.status.Refresh()
		return true
	})

	zw.OnKey(list, func(w zw.Widget, ev *tcell.EventKey) bool {
		if ev.Key() != tcell.KeyRune {
			return false
		}
		switch evRune(ev) {
		case 'n':
			d.ui.Close()
			d.openStyleEditor(true, "")
			return true
		case 'e':
			if list.Selected() < len(d.styleOrder) {
				name := d.styleOrder[list.Selected()]
				d.ui.Close()
				d.openStyleEditor(false, name)
			}
			return true
		case 'd':
			if list.Selected() < len(d.styleOrder) {
				name := d.styleOrder[list.Selected()]
				if name == "default" {
					return true
				}
				d.renameCells(name, "default")
				if d.drawStyle == name {
					d.drawStyle = "default"
				}
				d.styleRegistry.DeleteStyle(name)
				d.removeFromStyleOrder(name)
				d.ui.Close()
				d.openStylePopup() // reopen with updated list
			}
			return true
		}
		return false
	})

	d.ui.Popup(-2, 1, 0, 0, popup)
}

// ============================================================
// Style editor popup
// ============================================================

func (d *Designer) openStyleEditor(isNew bool, origName string) {
	var fields [5]string
	if !isNew {
		s := d.styleRegistry.Style(origName)
		fields = [5]string{
			origName,
			d.styleParentName(origName),
			s.OwnForeground(),
			s.OwnBackground(),
			s.OwnFont(),
		}
	}

	title := "Edit Style"
	if isNew {
		title = "New Style"
	}

	const inputW = 22
	popup := d.ui.NewBuilder().
		Dialog("se-dialog", title).
		Flex("se-content", false, "stretch", 0).
		Flex("se-r0", true, "stretch", 0).Static("se-l0", "name  ").Input("se-name", fields[0]).End().
		Flex("se-r1", true, "stretch", 0).Static("se-l1", "parent").Input("se-parent", fields[1]).End().
		Flex("se-r2", true, "stretch", 0).Static("se-l2", "fg    ").Input("se-fg", fields[2]).End().
		Flex("se-r3", true, "stretch", 0).Static("se-l3", "bg    ").Input("se-bg", fields[3]).End().
		Flex("se-r4", true, "stretch", 0).Static("se-l4", "font  ").Input("se-attr", fields[4]).End().
		End().
		Container()

	for _, id := range []string{"se-name", "se-parent", "se-fg", "se-bg", "se-attr"} {
		zw.Find(popup, id).(*zw.Input).SetHint(inputW, 1)
	}

	submit := func(w zw.Widget, event string, data ...any) bool {
		d.commitStyleEdit(popup, origName, isNew)
		return true
	}
	for _, id := range []string{"se-name", "se-parent", "se-fg", "se-bg", "se-attr"} {
		zw.Find(popup, id).On("enter", submit)
	}

	d.ui.Popup(-1, -1, 0, 0, popup)
}

func (d *Designer) commitStyleEdit(popup zw.Container, origName string, isNew bool) {
	get := func(id string) string {
		return strings.TrimSpace(zw.Find(popup, id).(*zw.Input).Text())
	}
	name := get("se-name")
	parentName := get("se-parent")
	fg := get("se-fg")
	bg := get("se-bg")
	font := get("se-attr")

	if name == "" {
		d.msg = "Style name cannot be empty"
		d.status.Refresh()
		return
	}
	if parentName != "" && parentName == name {
		d.msg = "A style cannot be its own parent"
		d.status.Refresh()
		return
	}

	var parentStyle *zw.Style
	if parentName != "" {
		found := false
		for _, n := range d.styleOrder {
			if n == parentName {
				found = true
				break
			}
		}
		if !found {
			d.msg = fmt.Sprintf("Parent style %q not found", parentName)
			d.status.Refresh()
			return
		}
		parentStyle = d.styleRegistry.Style(parentName)
	}

	if isNew {
		for _, n := range d.styleOrder {
			if n == name {
				d.msg = fmt.Sprintf("Style %q already exists", name)
				d.status.Refresh()
				return
			}
		}
		s := zw.NewStyle(name).
			WithForeground(fg).
			WithBackground(bg).
			WithFont(font).
			WithParent(parentStyle)
		d.styleRegistry.SetStyle(name, s)
		d.insertIntoStyleOrder(name)
	} else {
		if name != origName {
			for _, n := range d.styleOrder {
				if n == name {
					d.msg = fmt.Sprintf("Name %q already taken", name)
					d.status.Refresh()
					return
				}
			}
		}
		// Modify in place; canvas buffer cells hold the same pointer so they
		// update automatically without needing a full sync.
		existing := d.styleRegistry.Style(origName)
		existing.WithForeground(fg).WithBackground(bg).WithFont(font).WithParent(parentStyle)

		if name != origName {
			d.styleRegistry.SetStyle(name, existing)
			d.styleRegistry.SetStyle(origName, nil)
			d.renameCells(origName, name)
			if d.drawStyle == origName {
				d.drawStyle = name
			}
			for i, n := range d.styleOrder {
				if n == origName {
					d.styleOrder[i] = name
					break
				}
			}
			d.sortStyleOrder()
		}
	}

	d.modified = true
	d.ui.Close()
	d.canvas.Refresh()
	d.status.Refresh()
}

// ============================================================
// Palette popup
// ============================================================

func (d *Designer) openPalettePopup() {
	names := make([]string, len(d.palettes))
	for i, p := range d.palettes {
		names[i] = p.name
	}

	popup := d.ui.NewBuilder().
		Dialog("pal-popup", "Palette").
		List("pal-list", names...).
		Container()

	list := zw.Find(popup, "pal-list").(*zw.List)
	list.SetHint(16, len(d.palettes))
	list.SetQuickSearch(false)
	list.Select(d.palIdx)

	list.On("activate", func(w zw.Widget, event string, data ...any) bool {
		if idx, ok := data[0].(int); ok {
			d.palIdx = idx
		}
		d.ui.Close()
		d.canvas.Refresh()
		d.status.Refresh()
		return true
	})

	d.ui.Popup(-1, -1, 0, 0, popup)
}

// ============================================================
// Help popup
// ============================================================

var helpLines = []string{
	"  NORMAL MODE                       ",
	"  h j k l / arrows   move           ",
	"  0 $                row start/end  ",
	"  g G                first/last row ",
	"  i                  INSERT mode    ",
	"  d                  DRAW mode      ",
	"  v                  VISUAL mode    ",
	"  :                  COMMAND mode   ",
	"  b                  palette picker ",
	"  s                  style picker   ",
	"  p                  paste          ",
	"  x                  delete char    ",
	"  ?                  this help      ",
	"                                    ",
	"  INSERT MODE                       ",
	"  arrows             move           ",
	"  Backspace          delete char    ",
	"  Esc                back to NORMAL ",
	"                                    ",
	"  DRAW MODE                         ",
	"  7-9 / 4-6 / 1-3   box chars      ",
	"  - *                horiz / vert   ",
	"  s                  numpad overlay ",
	"  b                  palette picker ",
	"  Esc                back to NORMAL ",
	"                                    ",
	"  VISUAL MODE                       ",
	"  movement           extend sel     ",
	"  y                  yank           ",
	"  d                  delete         ",
	"  b                  draw box       ",
	"  Esc                cancel         ",
	"                                    ",
	"  COMMAND MODE                      ",
	"  :w [file]          save           ",
	"  :e file            load           ",
	"  :q / :q!           quit           ",
	"  :wq                save and quit  ",
	"                                    ",
	"        Esc to close                ",
}

func (d *Designer) openHelpPopup() {
	popup := d.ui.NewBuilder().
		Dialog("help-popup", "Key Bindings").
		Text("help-text", helpLines, false, 0).
		Container()

	maxW := 0
	for _, l := range helpLines {
		if n := len([]rune(l)); n > maxW {
			maxW = n
		}
	}
	zw.Find(popup, "help-text").(*zw.Text).SetHint(maxW, len(helpLines))

	d.ui.Popup(-1, -1, 0, 0, popup)
}

// ============================================================
// Main
// ============================================================

func main() {
	var filename string
	docName := "untitled"
	var page *Page
	var initRegistry *zw.Canvas
	var initOrder []string

	if len(os.Args) > 1 {
		filename = os.Args[1]
		name, p, registry, order, err := loadDoc(filename)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "designer: cannot open %q: %v\n", filename, err)
			os.Exit(1)
		}
		if p != nil {
			docName = name
			page = p
			initRegistry = registry
			initOrder = order
		}
	}

	if page == nil {
		// Large default; rendering is clamped to the allocated widget area.
		page = newPage("main", 220, 60)
	}

	d := newDesigner(page, docName)
	if initRegistry != nil {
		d.styleRegistry = initRegistry
		d.styleOrder = initOrder
	}
	if filename != "" {
		d.filename = filename
	}

	theme := zw.TokyoNightTheme()

	canvas := newDesignerCanvas(d, page.width, page.height)
	status := newDesignerStatus(d)
	d.canvas = canvas
	d.status = status

	// Sync loaded page content into the canvas cell buffer.
	if initRegistry != nil {
		d.syncPageToCanvas()
	}

	root := zw.NewFlex("root", false, "stretch", 0)
	root.Add(canvas)
	root.Add(status)

	ui, err := zw.NewUI(theme, root, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "designer: %v\n", err)
		os.Exit(1)
	}
	d.ui = ui

	if err := ui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "designer: %v\n", err)
		os.Exit(1)
	}
}

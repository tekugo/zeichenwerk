package zeichenwerk

import "fmt"

// ---- Color type ------------------------------------------------------------

// Color encodes a terminal colour in a single uint32.
//
//	0x00000000                     → ColorDefault (terminal's own default)
//	0x00000001 – 0x00000100        → xterm-256 palette index (value-1)
//	0x80000000 – 0x80FFFFFF        → true colour (bits 0-23 = 0xRRGGBB)
type Color uint32

const ColorDefault Color = 0

// PaletteColor returns a Color for xterm-256 palette index idx (0–255).
func PaletteColor(idx int) Color { return Color(idx + 1) }

// TrueColor returns a Color from 8-bit RGB components.
func TrueColor(r, g, b uint8) Color {
	return Color(0x80000000 | uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}

// colorToHex converts a Color to the hex string expected by Renderer.Set.
// ColorDefault → "", palette → "#RRGGBB" from xterm table, true colour → "#RRGGBB".
func colorToHex(c Color) string {
	if c == ColorDefault {
		return ""
	}
	if c&0x80000000 != 0 {
		// True colour
		return fmt.Sprintf("#%06X", uint32(c)&0x00FFFFFF)
	}
	// Palette
	idx := int(c) - 1
	if idx < 0 || idx > 255 {
		return ""
	}
	rgb := xterm256[idx]
	return fmt.Sprintf("#%02X%02X%02X", rgb[0], rgb[1], rgb[2])
}

// xterm256 is the canonical xterm 256-colour palette (R, G, B).
var xterm256 [256][3]uint8

func init() {
	// 0-7: standard colours
	standard := [8][3]uint8{
		{0, 0, 0},       // 0 black
		{128, 0, 0},     // 1 maroon
		{0, 128, 0},     // 2 green
		{128, 128, 0},   // 3 olive
		{0, 0, 128},     // 4 navy
		{128, 0, 128},   // 5 purple
		{0, 128, 128},   // 6 teal
		{192, 192, 192}, // 7 silver
	}
	// 8-15: high-intensity
	bright := [8][3]uint8{
		{128, 128, 128}, // 8 grey
		{255, 0, 0},     // 9 red
		{0, 255, 0},     // 10 lime
		{255, 255, 0},   // 11 yellow
		{0, 0, 255},     // 12 blue
		{255, 0, 255},   // 13 fuchsia
		{0, 255, 255},   // 14 aqua
		{255, 255, 255}, // 15 white
	}
	for i := 0; i < 8; i++ {
		xterm256[i] = standard[i]
		xterm256[8+i] = bright[i]
	}
	// 16-231: 6×6×6 colour cube
	cubeVal := func(v int) uint8 {
		if v == 0 {
			return 0
		}
		return uint8(55 + 40*v)
	}
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				xterm256[16+36*r+6*g+b] = [3]uint8{cubeVal(r), cubeVal(g), cubeVal(b)}
			}
		}
	}
	// 232-255: greyscale ramp
	for i := 0; i < 24; i++ {
		v := uint8(8 + 10*i)
		xterm256[232+i] = [3]uint8{v, v, v}
	}
}

// ---- char uint32 bit layout ------------------------------------------------
//
//  31  30  29  28  27  26  25  24  23  22  21  20 ........... 0
//   └──┬──┘   │   │   │   │   │   │   │   │   └──────────────┘
//  ul_style  inv stk rev blk itl dim bld wid      rune (21 bits)
//  (3 bits)

const (
	charRuneMask uint32 = 0x001FFFFF // bits 0-20: Unicode code point
	charWide     uint32 = 1 << 21   // double-width cell
	charBold     uint32 = 1 << 22   // bold
	charDim      uint32 = 1 << 23   // faint/dim
	charItalic   uint32 = 1 << 24   // italic
	charBlink    uint32 = 1 << 25   // blink
	charReverse  uint32 = 1 << 26   // reverse video
	charStrike   uint32 = 1 << 27   // strikethrough
	charInvis    uint32 = 1 << 28   // invisible
	charULMask   uint32 = 0xE0000000 // bits 29-31: underline style
	charULShift          = 29

	ULNone   = 0
	ULSingle = 1
	ULDouble = 2
	ULCurly  = 3
	ULDotted = 4
	ULDashed = 5
)

// PackAttrs builds the attribute portion of a char word (bits 21–31).
// attrs is a bitmask of charBold | charDim | … constants.
// ulStyle is ULNone..ULDashed.
func PackAttrs(attrs uint32, ulStyle int) uint32 {
	return (attrs & ^charRuneMask) | (uint32(ulStyle) << charULShift)
}

// attrsToFont converts the attr bits of a char word to a space-separated
// font string accepted by Renderer.Set (e.g. "bold italic").
// Underline is included when ul_style bits are non-zero.
func attrsToFont(char uint32) string {
	var parts []string
	if char&charBold != 0 {
		parts = append(parts, "bold")
	}
	if char&charItalic != 0 {
		parts = append(parts, "italic")
	}
	if char&charBlink != 0 {
		parts = append(parts, "blink")
	}
	if char&charStrike != 0 {
		parts = append(parts, "strikethrough")
	}
	if char&charULMask != 0 {
		parts = append(parts, "underline")
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

// ---- CellBuffer ------------------------------------------------------------

// CellBuffer is a compact, flat 2D grid of terminal cells stored in four
// parallel []uint32 slices (char, fg, bg, ul). Cell (x,y) is at index y*w+x.
type CellBuffer struct {
	char []uint32 // rune + attrs + ul-style packed as uint32
	fg   []uint32 // foreground Color
	bg   []uint32 // background Color
	ul   []uint32 // underline Color (Kitty extension)
	w, h int
}

// NewCellBuffer creates a new CellBuffer with dimensions w×h.
// All cells start as default (zero).
// Panics if w < 1 or h < 1.
func NewCellBuffer(w, h int) *CellBuffer {
	if w < 1 || h < 1 {
		panic("CellBuffer: dimensions must be at least 1×1")
	}
	n := w * h
	return &CellBuffer{
		char: make([]uint32, n),
		fg:   make([]uint32, n),
		bg:   make([]uint32, n),
		ul:   make([]uint32, n),
		w:    w,
		h:    h,
	}
}

func (b *CellBuffer) Width() int  { return b.w }
func (b *CellBuffer) Height() int { return b.h }

// Get returns the rune, colours, and raw char word for cell (x,y).
// Returns zero values if coordinates are out of bounds (no panic).
func (b *CellBuffer) Get(x, y int) (ch rune, fg, bg, ul Color, attrs uint32) {
	if x < 0 || x >= b.w || y < 0 || y >= b.h {
		return 0, 0, 0, 0, 0
	}
	i := y*b.w + x
	raw := b.char[i]
	ch = rune(raw & charRuneMask)
	attrs = raw &^ charRuneMask
	fg = Color(b.fg[i])
	bg = Color(b.bg[i])
	ul = Color(b.ul[i])
	return
}

// Set writes a cell. attrs is the packed char word (bits 21–31); use PackAttrs.
// Silently ignores out-of-bounds coordinates.
func (b *CellBuffer) Set(x, y int, ch rune, fg, bg, ul Color, attrs uint32) {
	if x < 0 || x >= b.w || y < 0 || y >= b.h {
		return
	}
	i := y*b.w + x
	b.char[i] = (uint32(ch) & charRuneMask) | (attrs &^ charRuneMask)
	b.fg[i] = uint32(fg)
	b.bg[i] = uint32(bg)
	b.ul[i] = uint32(ul)
}

// SetChar updates only the rune in a cell, leaving attrs and colour unchanged.
// Silently ignores out-of-bounds coordinates.
func (b *CellBuffer) SetChar(x, y int, ch rune) {
	if x < 0 || x >= b.w || y < 0 || y >= b.h {
		return
	}
	i := y*b.w + x
	b.char[i] = (b.char[i] &^ charRuneMask) | (uint32(ch) & charRuneMask)
}

// Clear resets all cells to the default state (zero).
func (b *CellBuffer) Clear() {
	clear(b.char)
	clear(b.fg)
	clear(b.bg)
	clear(b.ul)
}

// ClearLine clears cells x1..x2-1 on row y to default. Clamps to buffer bounds.
func (b *CellBuffer) ClearLine(y, x1, x2 int) {
	if y < 0 || y >= b.h {
		return
	}
	if x1 < 0 {
		x1 = 0
	}
	if x2 > b.w {
		x2 = b.w
	}
	for x := x1; x < x2; x++ {
		i := y*b.w + x
		b.char[i] = 0
		b.fg[i] = 0
		b.bg[i] = 0
		b.ul[i] = 0
	}
}

// ClearLineColor clears cells x1..x2-1 on row y, filling bg with bgColor.
func (b *CellBuffer) ClearLineColor(y, x1, x2 int, bgColor Color) {
	if y < 0 || y >= b.h {
		return
	}
	if x1 < 0 {
		x1 = 0
	}
	if x2 > b.w {
		x2 = b.w
	}
	for x := x1; x < x2; x++ {
		i := y*b.w + x
		b.char[i] = 0
		b.fg[i] = 0
		b.bg[i] = uint32(bgColor)
		b.ul[i] = 0
	}
}

// Resize resizes the buffer to w×h, preserving as much content as possible.
// New cells are zero-initialised. Panics if w < 1 or h < 1.
func (b *CellBuffer) Resize(w, h int) {
	if w < 1 || h < 1 {
		panic("CellBuffer: dimensions must be at least 1×1")
	}
	if w == b.w && h == b.h {
		return
	}
	n := w * h
	newChar := make([]uint32, n)
	newFg := make([]uint32, n)
	newBg := make([]uint32, n)
	newUl := make([]uint32, n)

	copyH := b.h
	if h < copyH {
		copyH = h
	}
	copyW := b.w
	if w < copyW {
		copyW = w
	}
	for y := 0; y < copyH; y++ {
		for x := 0; x < copyW; x++ {
			src := y*b.w + x
			dst := y*w + x
			newChar[dst] = b.char[src]
			newFg[dst] = b.fg[src]
			newBg[dst] = b.bg[src]
			newUl[dst] = b.ul[src]
		}
	}
	b.char = newChar
	b.fg = newFg
	b.bg = newBg
	b.ul = newUl
	b.w = w
	b.h = h
}

// copyRow copies a full row from src buffer row sy to this buffer row dy.
// Used for scroll operations. Width is the smaller of the two buffers.
func (b *CellBuffer) copyRow(dy, sy int) {
	if dy < 0 || dy >= b.h || sy < 0 || sy >= b.h {
		return
	}
	copy(b.char[dy*b.w:dy*b.w+b.w], b.char[sy*b.w:sy*b.w+b.w])
	copy(b.fg[dy*b.w:dy*b.w+b.w], b.fg[sy*b.w:sy*b.w+b.w])
	copy(b.bg[dy*b.w:dy*b.w+b.w], b.bg[sy*b.w:sy*b.w+b.w])
	copy(b.ul[dy*b.w:dy*b.w+b.w], b.ul[sy*b.w:sy*b.w+b.w])
}

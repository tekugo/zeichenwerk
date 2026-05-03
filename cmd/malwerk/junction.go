package main

// Mask is a 4-bit connection mask: top | right | bottom | left.
type Mask uint8

const (
	MaskTop    Mask = 1 << 0
	MaskRight  Mask = 1 << 1
	MaskBottom Mask = 1 << 2
	MaskLeft   Mask = 1 << 3
)

// boxFamilies maps family name → array indexed by Mask of the rune that
// represents that connection pattern. Mask 0 is the empty cell (space)
// for every family, but we use "" to mean "no rune for this combo so
// don't auto-render".
var boxFamilies = map[string][16]string{
	"thin": {
		"",  // 0 = none
		"╵", // top only
		"╶", // right only
		"└", // top + right
		"╷", // bottom only
		"│", // top + bottom
		"┌", // right + bottom
		"├", // top + right + bottom
		"╴", // left only
		"┘", // top + left
		"─", // right + left
		"┴", // top + right + left
		"┐", // bottom + left
		"┤", // top + bottom + left
		"┬", // right + bottom + left
		"┼", // all four
	},
	"heavy": {
		"", "╹", "╺", "┗", "╻", "┃", "┏", "┣",
		"╸", "┛", "━", "┻", "┓", "┫", "┳", "╋",
	},
	"double": {
		// Double has no half-stub characters; missing ones fall through
		// to the closest representable.
		"", "║", "═", "╚", "║", "║", "╔", "╠",
		"═", "╝", "═", "╩", "╗", "╣", "╦", "╬",
	},
	"round": {
		"",  // 0
		"╵", // top
		"╶", // right
		"╰", // top + right (rounded corner)
		"╷", // bottom
		"│", // top + bottom
		"╭", // right + bottom (rounded corner)
		"├", // top + right + bottom
		"╴", // left
		"╯", // top + left (rounded corner)
		"─", // right + left
		"┴", // top + right + left
		"╮", // bottom + left (rounded corner)
		"┤", // top + bottom + left
		"┬", // right + bottom + left
		"┼", // all four
	},
}

// runeForMask returns the box-drawing rune for the given family and
// mask. Returns "" when the combination is empty.
func runeForMask(family string, mask Mask) string {
	t, ok := boxFamilies[family]
	if !ok {
		t = boxFamilies["thin"]
	}
	return t[mask&0xF]
}

// maskInFamily reports whether ch is a box-drawing rune in the given
// family and, if so, returns its connection mask. Returns ok=false when
// the rune is not part of that family — used to decide whether to merge.
func maskInFamily(ch, family string) (Mask, bool) {
	t, ok := boxFamilies[family]
	if !ok {
		return 0, false
	}
	for i, c := range t {
		if i == 0 || c == "" {
			continue
		}
		if c == ch {
			return Mask(i), true
		}
	}
	return 0, false
}


// placeBoxRune writes a box-drawing rune at (x, y) using the supplied
// family. If the existing cell is from the same family, the new mask is
// OR-merged with the existing one. After placement the four neighbours
// are re-evaluated so adjacent box-drawing cells absorb the new
// connection.
func (e *Editor) placeBoxRune(x, y int, ch, family string) {
	newMask, ok := maskInFamily(ch, family)
	if !ok {
		return
	}
	hist := e.app.history
	hist.Begin()
	defer hist.Commit()

	before := e.doc.At(x, y)
	merged := newMask
	if existingMask, isSameFam := maskInFamily(before.Ch, family); isSameFam {
		merged |= existingMask
	}
	finalCh := runeForMask(family, merged)
	if finalCh == "" {
		return
	}
	after := Cell{Ch: finalCh, Style: e.style}
	if before != after {
		hist.Record(x, y, before, after)
		e.doc.Set(x, y, after)
	}
	e.reevalNeighbour(x-1, y, family, hist)
	e.reevalNeighbour(x+1, y, family, hist)
	e.reevalNeighbour(x, y-1, family, hist)
	e.reevalNeighbour(x, y+1, family, hist)
}

// reevalNeighbour augments a cell's mask with new connections from its
// four neighbours. Existing connections are preserved — this only ever
// adds bits, never removes them — so deliberate "free" line endpoints
// are not downgraded to half-stubs. Cells holding non-box-drawing
// content (typed text) are left alone.
func (e *Editor) reevalNeighbour(x, y int, family string, hist *History) {
	if x < 0 || y < 0 || x >= e.doc.Width || y >= e.doc.Height {
		return
	}
	cur := e.doc.At(x, y)
	curMask, ok := maskInFamily(cur.Ch, family)
	if !ok {
		return
	}
	merged := curMask | neighbourMask(e.doc, x, y, family)
	if merged == curMask {
		return
	}
	ch := runeForMask(family, merged)
	if ch == "" {
		return
	}
	after := cur
	after.Ch = ch
	if after == cur {
		return
	}
	hist.Record(x, y, cur, after)
	e.doc.Set(x, y, after)
}

// neighbourMask returns the connection mask formed by looking at the four
// neighbours of (x, y) — only same-family box-drawing neighbours count.
func neighbourMask(d *Document, x, y int, family string) Mask {
	var m Mask
	if y > 0 {
		if mask, ok := maskInFamily(d.At(x, y-1).Ch, family); ok && mask&MaskBottom != 0 {
			m |= MaskTop
		}
	}
	if x+1 < d.Width {
		if mask, ok := maskInFamily(d.At(x+1, y).Ch, family); ok && mask&MaskLeft != 0 {
			m |= MaskRight
		}
	}
	if y+1 < d.Height {
		if mask, ok := maskInFamily(d.At(x, y+1).Ch, family); ok && mask&MaskTop != 0 {
			m |= MaskBottom
		}
	}
	if x > 0 {
		if mask, ok := maskInFamily(d.At(x-1, y).Ch, family); ok && mask&MaskRight != 0 {
			m |= MaskLeft
		}
	}
	return m
}

// drawHLine draws a horizontal line on row y from x1..x2 (inclusive)
// with junction merge. Every cell along the line uses the full
// horizontal rune; merges with existing same-family cells produce
// proper T- and cross-junctions.
func (e *Editor) drawHLine(x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	ch := runeForMask(e.border, MaskLeft|MaskRight)
	if ch == "" {
		return
	}
	for x := x1; x <= x2; x++ {
		e.placeBoxRune(x, y, ch, e.border)
	}
	e.Refresh()
}

// drawVLine draws a vertical line on column x from y1..y2 (inclusive)
// with junction merge.
func (e *Editor) drawVLine(x, y1, y2 int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	ch := runeForMask(e.border, MaskTop|MaskBottom)
	if ch == "" {
		return
	}
	for y := y1; y <= y2; y++ {
		e.placeBoxRune(x, y, ch, e.border)
	}
	e.Refresh()
}

// borderSelection draws a frame using the given family around the
// current visual selection rect.
func (e *Editor) borderSelection(family string) {
	x1, y1, x2, y2 := e.selectionRect()
	if x1 > x2 || y1 > y2 {
		return
	}
	prev := e.border
	e.border = family
	defer func() { e.border = prev }()

	corners := []struct {
		x, y int
		mask Mask
	}{
		{x1, y1, MaskRight | MaskBottom}, // top-left
		{x2, y1, MaskLeft | MaskBottom},  // top-right
		{x1, y2, MaskRight | MaskTop},    // bottom-left
		{x2, y2, MaskLeft | MaskTop},     // bottom-right
	}
	for _, c := range corners {
		ch := runeForMask(family, c.mask)
		if ch != "" {
			e.placeBoxRune(c.x, c.y, ch, family)
		}
	}
	if x2 > x1+1 {
		topMid := runeForMask(family, MaskLeft|MaskRight)
		botMid := runeForMask(family, MaskLeft|MaskRight)
		for x := x1 + 1; x < x2; x++ {
			e.placeBoxRune(x, y1, topMid, family)
			e.placeBoxRune(x, y2, botMid, family)
		}
	}
	if y2 > y1+1 {
		leftMid := runeForMask(family, MaskTop|MaskBottom)
		rightMid := runeForMask(family, MaskTop|MaskBottom)
		for y := y1 + 1; y < y2; y++ {
			e.placeBoxRune(x1, y, leftMid, family)
			e.placeBoxRune(x2, y, rightMid, family)
		}
	}
	e.Refresh()
}

// floodFill replaces all cells reachable from (sx, sy) sharing the seed
// cell's (Ch, Style) with the editor's current rune + style. When clip
// dimensions are valid (cx1 ≤ cx2 and cy1 ≤ cy2), the fill is constrained
// to that rectangle; pass -1 for any clip value to use the whole document.
func (e *Editor) floodFill(sx, sy, cx1, cy1, cx2, cy2 int) {
	if sx < 0 || sy < 0 || sx >= e.doc.Width || sy >= e.doc.Height {
		return
	}
	clipped := cx1 >= 0 && cy1 >= 0 && cx2 >= 0 && cy2 >= 0
	if !clipped {
		cx1, cy1 = 0, 0
		cx2, cy2 = e.doc.Width-1, e.doc.Height-1
	}

	seed := e.doc.At(sx, sy)
	target := Cell{Ch: e.lastRune, Style: e.style}
	if seed == target {
		return
	}

	hist := e.app.history
	hist.Begin()
	defer hist.Commit()

	visited := map[[2]int]bool{}
	queue := [][2]int{{sx, sy}}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		x, y := p[0], p[1]
		if visited[p] {
			continue
		}
		visited[p] = true
		if x < cx1 || x > cx2 || y < cy1 || y > cy2 {
			continue
		}
		cur := e.doc.At(x, y)
		if cur != seed {
			continue
		}
		hist.Record(x, y, cur, target)
		e.doc.Set(x, y, target)
		queue = append(queue, [2]int{x - 1, y}, [2]int{x + 1, y}, [2]int{x, y - 1}, [2]int{x, y + 1})
	}
	e.Refresh()
}

// paletteNames returns palette keys with "default" first, the rest sorted
// alphabetically.
func paletteNames(d *Document) []string {
	names := make([]string, 0, len(d.Palette))
	names = append(names, "default")
	for k := range d.Palette {
		if k != "default" {
			names = append(names, k)
		}
	}
	// stable, simple sort for the tail
	tail := names[1:]
	for i := 1; i < len(tail); i++ {
		for j := i; j > 0 && tail[j] < tail[j-1]; j-- {
			tail[j], tail[j-1] = tail[j-1], tail[j]
		}
	}
	return names
}

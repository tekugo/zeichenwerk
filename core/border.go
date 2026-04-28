package core

import "unicode/utf8"

// Border is the complete set of glyphs used to draw a widget's frame and any
// inner grid lines. Each field holds a short string (typically a single rune)
// describing one role in the border artwork. Fields left empty mean "don't
// draw this element", which allows partial or one-sided borders without
// requiring a different type.
//
// Borders are usually registered in a Theme and looked up by name from the
// renderer rather than constructed ad hoc at render time.
type Border struct {
	// Outer border elements — the perimeter strokes of the widget.
	Top    string // horizontal segment along the top edge
	Right  string // vertical segment along the right edge
	Bottom string // horizontal segment along the bottom edge
	Left   string // vertical segment along the left edge

	// Corner elements — connect the perpendicular edges at the four corners.
	TopLeft     string // top-left corner glyph
	TopRight    string // top-right corner glyph
	BottomRight string // bottom-right corner glyph
	BottomLeft  string // bottom-left corner glyph

	// Outer T-connectors — used where an inner grid line meets an outer edge.
	TopT    string // inner vertical meeting the top edge
	RightT  string // inner horizontal meeting the right edge
	BottomT string // inner vertical meeting the bottom edge
	LeftT   string // inner horizontal meeting the left edge

	// Inner grid elements — subdivide the widget into cells (for tables etc.).
	InnerH string // horizontal grid line
	InnerV string // vertical grid line
	InnerX string // cross where two inner grid lines meet

	// Inner T-connectors — used where three inner grid lines meet.
	InnerTopT    string // inner T opening downwards
	InnerRightT  string // inner T opening leftwards
	InnerBottomT string // inner T opening upwards
	InnerLeftT   string // inner T opening rightwards
}

// Horizontal returns the number of terminal cells consumed by the left and
// right border sides combined. It counts runes (not bytes), so multibyte
// box-drawing characters contribute one cell each. An empty side contributes
// zero, so partial borders report exactly the space they take.
func (b *Border) Horizontal() int {
	return utf8.RuneCountInString(b.Left) + utf8.RuneCountInString(b.Right)
}

// Vertical returns the number of terminal rows consumed by the top and
// bottom border sides combined. Each side contributes at most one row — any
// configured glyph occupies a single row regardless of its string length —
// so the result is in {0, 1, 2} depending on which sides are present.
func (b *Border) Vertical() int {
	return min(len(b.Top), 1) + min(len(b.Bottom), 1)
}

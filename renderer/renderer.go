package renderer

// Renderer wraps a Screen and provides higher-level drawing primitives
// (text, fills, lines, scrollbars) on top of the raw cell operations. It
// deliberately knows nothing about themes or styles — the Set method
// forwards colour strings verbatim to the underlying Screen. Theme-aware
// rendering is implemented one layer up by core.Renderer, which embeds
// this type and resolves theme variables before delegating.
type Renderer struct {
	Screen Screen
}

// NewRenderer constructs a Renderer bound to the given Screen.
func NewRenderer(screen Screen) *Renderer {
	return &Renderer{
		Screen: screen,
	}
}

// ---- Primitive Rendering Operations from Screen ----

// Clear clears the entire screen.
func (r *Renderer) Clear() {
	r.Screen.Clear()
}

// Clip restricts subsequent drawing to the rectangle (x, y, width, height).
// Pass (0, 0, 0, 0) to remove the clipping region.
func (r *Renderer) Clip(x, y, width, height int) {
	r.Screen.Clip(x, y, width, height)
}

// Flush writes all pending cell changes to the terminal.
func (r *Renderer) Flush() {
	r.Screen.Flush()
}

// Get returns the character string currently at cell (x, y).
func (r *Renderer) Get(x, y int) string {
	return r.Screen.Get(x, y)
}

// Put writes the single character ch into the cell at (x, y) using the
// current style. For multi-character strings use Text.
func (r *Renderer) Put(x, y int, ch string) {
	r.Screen.Put(x, y, ch)
}

// Set installs the foreground colour, background colour, and font for
// subsequent drawing operations. The values are passed straight to the
// underlying Screen, so they must already be literal colours understood
// by the back-end — theme variables (prefixed with "$") are resolved at
// the core.Renderer layer, not here.
func (r *Renderer) Set(foreground, background, font string) {
	r.Screen.Set(foreground, background, font)
}

// SetUnderline sets the underline style and colour for subsequent Put calls.
// style: 0=none, 1=single, 2=double, 3=curly, 4=dotted, 5=dashed.
// color: empty string = terminal default.
func (r *Renderer) SetUnderline(style int, color string) {
	r.Screen.SetUnderline(style, color)
}

// Translate shifts the origin for subsequent drawing by (tx, ty). Pass (0, 0)
// to reset.
func (r *Renderer) Translate(tx, ty int) {
	r.Screen.Translate(tx, ty)
}

// ---- Extended Rendering ---------------------------------------------------

// Colorize re-applies the current style to every cell in a rectangular
// area without changing the characters already written there. It reads
// each cell with Get and immediately writes it back with Put, which
// causes the back-end to attach the style currently set via Set. This is
// useful for dimming, highlighting, or otherwise re-tinting an already
// rendered region.
//
// Parameters:
//   - x, y:          Top-left corner of the area (local coordinates).
//   - width, height: Dimensions of the area to process.
func (r *Renderer) Colorize(x, y, width, height int) {
	for i := range width {
		for j := range height {
			ch := r.Get(x+i, y+j)
			r.Put(x+i, y+j, ch)
		}
	}
}

// Fill fills a rectangular area with a characters using the current style.
// This method can be used to clear backgrounds and create solid color areas by
// overwriting existing content with spaces in the current color.
//
// Parameters:
//   - x, y: Top-left corner coordinates of the area to fill
//   - w, h: Width and height of the area to fill
//   - ch: Character to fill the area with
func (r *Renderer) Fill(x, y, w, h int, ch string) {
	cx, cy := x, y
	for range h {
		for range w {
			r.Put(cx, cy, ch)
			cx++
		}
		cx = x
		cy++
	}
}

// Line draws a straight line composed of three distinct parts: a start
// cap, a run of repeated middle characters, and an end cap. It is the
// primitive behind border top/bottom strokes, table separators, and
// decorative rules.
//
// The direction is expressed as a unit step vector (dx, dy):
//   - ( 1,  0): horizontal, left to right
//   - ( 0,  1): vertical, top to bottom
//   - (-1,  0): horizontal, right to left
//   - ( 0, -1): vertical, bottom to top
//
// Diagonal vectors work too but are rarely used in terminal UI.
//
// Parameters:
//   - x, y:               Starting cell for the start cap.
//   - dx, dy:             Unit step vector; each step of the line
//     advances by this amount.
//   - length:             Number of middle cells between start and end.
//     Zero produces a start cap immediately followed by the end cap.
//   - start, middle, end: Glyphs used for the three segments.
func (r *Renderer) Line(x, y, dx, dy, length int, start, middle, end string) {
	r.Put(x, y, start)
	cx := x + dx
	cy := y + dy
	for range length {
		r.Put(cx, cy, middle)
		cx += dx
		cy += dy
	}
	r.Put(cx, cy, end)
}

// Repeat draws a single character multiple times in a specified direction.
// This method is used for creating patterns, filling areas, and drawing
// simple graphical elements like progress indicators or decorative patterns.
//
// Parameters:
//   - x, y: Starting coordinates for the pattern
//   - dx, dy: Direction vector for pattern advancement
//   - length: Number of characters to draw
//   - ch: Unicode character to repeat
func (r *Renderer) Repeat(x, y, dx, dy, length int, ch string) {
	for range length {
		r.Put(x, y, ch)
		x += dx
		y += dy
	}
}

// ScrollbarV renders a vertical scrollbar indicating scroll position and
// content size. The thumb size is proportional to the ratio of visible to
// total content and its position reflects the current offset, so the
// scrollbar doubles as a progress-like indicator for how far the viewer
// has scrolled through the content.
//
// No scrollbar is drawn when height or total is non-positive.
//
// Parameters:
//   - x, y:   Top-left coordinate of the scrollbar track.
//   - height: Number of cells occupied by the scrollbar (track length).
//   - offset: Current scroll offset (items scrolled past the top edge).
//   - total:  Total number of items in the content, including those not
//     currently visible.
func (r *Renderer) ScrollbarV(x, y, height, offset, total int) {
	if height <= 0 || total <= 0 {
		return
	}

	// Calculate scrollbar thumb position and size
	thumb := min(max(height*height/total, 1), height)

	// Calculate thumb position, handling edge case where total <= height
	var pos int
	if total > height {
		pos = min(max(offset*(height-thumb)/(total-height), 0), height-thumb)
	} else {
		pos = 0 // Content fits within view, thumb starts at beginning
	}

	// Render scrollbar track
	for i := range height {
		var ch string
		if i >= pos && i < pos+thumb {
			ch = "█" // Solid block for thumb
		} else {
			ch = "░" // Light shade for track
		}
		r.Put(x, y+i, ch)
	}
}

// ScrollbarH renders a horizontal scrollbar indicating scroll position
// and content width. The layout logic mirrors ScrollbarV; when the
// content fits inside the visible width (total <= width) the thumb is
// parked at the start of the track to avoid flicker rather than
// suppressing the scrollbar entirely.
//
// No scrollbar is drawn when width or total is non-positive.
//
// Parameters:
//   - x, y:   Top-left coordinate of the scrollbar track.
//   - width:  Number of cells occupied by the scrollbar (track length).
//   - offset: Current horizontal scroll offset (cells scrolled past the
//     left edge).
//   - total:  Total width of the content, in cells.
func (r *Renderer) ScrollbarH(x, y, width, offset, total int) {
	if width <= 0 || total <= 0 {
		return
	}

	// Calculate scrollbar thumb position and size
	thumb := min(max(width*width/total, 1), width)

	// Calculate thumb position, handling edge case where total <= width
	var pos int
	if total > width {
		pos = min(max(offset*(width-thumb)/(total-width), 0), width-thumb)
	} else {
		pos = 0 // Content fits within view, thumb starts at beginning
	}

	// Render horizontal scrollbar track
	for i := range width {
		var ch string
		if i >= pos && i < pos+thumb {
			ch = "█" // Solid block for thumb
		} else {
			ch = "░" // Light shade for track
		}
		r.Put(x+i, y, ch)
	}
}

// Text renders a string at the specified coordinates using the current
// style. Each rune is written to its own cell, so the function handles
// multibyte characters correctly while treating the result as
// single-width in cell units (callers that need double-width glyph
// support must pre-measure).
//
// When max > 0 the text is clipped to that many cells; any remaining
// space is padded with spaces so the full max-cell band is overwritten.
// When max == 0 the string is rendered verbatim with no width constraint
// and no padding.
//
// Parameters:
//   - x, y: Starting cell for the first character.
//   - s:    The text to render.
//   - max:  Maximum number of cells the output may occupy, or 0 for no
//     limit.
func (r *Renderer) Text(x, y int, s string, max int) {
	i := 0
	for _, ch := range s {
		if max > 0 && i >= max {
			break
		}
		r.Put(x+i, y, string(ch))
		i++
	}
	if max > 0 && i < max {
		for ; i < max; i++ {
			r.Put(x+i, y, " ")
		}
	}
}

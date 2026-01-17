package next

type Renderer struct {
	screen Screen
	theme  *Theme
}

// NewRenderer creates a new Renderer instance.
func NewRenderer(screen Screen, theme *Theme) *Renderer {
	return &Renderer{
		screen: screen,
		theme:  theme,
	}
}

// ---- Primitive Rendering Operations from Screen ----

func (r *Renderer) Clear() {
	r.screen.Clear()
}

func (r *Renderer) Clip(x, y, width, height int) {
	r.screen.Clip(x, y, width, height)
}

func (r *Renderer) Flush() {
	r.screen.Flush()
}

func (r *Renderer) Get(x, y int) string {
	return r.screen.Get(x, y)
}

func (r *Renderer) Put(x, y int, ch string) {
	r.screen.Put(x, y, ch)
}

func (r *Renderer) Set(foreground, background, font string) {
	r.screen.Set(r.theme.Color(foreground), r.theme.Color(background), font)
}

// ---- Additional Rendering Operations ----

// Border draws a complete border around a rectangular area using the specified BorderStyle.
// This method renders the four sides and corners of a border, creating a frame around
// the given coordinates using the border characters.
//
// Parameters:
//   - x, y: Top-left corner coordinates of the border area
//   - w, h: Width and height of the area to border (inner dimensions)
//   - box: Border containing the characters for each border element
func (r *Renderer) Border(x, y, w, h int, border string) {
	b := r.theme.Border(border)
	r.Line(x, y, 1, 0, w-2, b.TopLeft, b.Top, b.TopRight)
	r.Line(x, y+h-1, 1, 0, w-2, b.BottomLeft, b.Bottom, b.BottomRight)
	for i := range h - 2 {
		r.Put(x, y+i+1, b.Left)
		r.Put(x+w-1, y+i+1, b.Right)
	}
}

// Colorize applies the current renderer style to all characters in a rectangular area.
// This method preserves the existing characters while changing their visual styling,
// effectively recoloring or reformatting text and graphics within the specified region.
//
// Parameters:
//   - x, y: Top-left corner of the area to colorize
//   - width, height: Dimensions of the area to process
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

// Line draws a line using specified start, middle, and end characters.
// This method is used for drawing borders, separators, and decorative elements
// with proper terminal characters for clean visual presentation.
//
// # Direction Control
//
// The dx and dy parameters control line direction:
//   - dx=1, dy=0: Horizontal line (left to right)
//   - dx=0, dy=1: Vertical line (top to bottom)
//   - dx=-1, dy=0: Horizontal line (right to left)
//   - dx=0, dy=-1: Vertical line (bottom to top)
//
// Parameters:
//   - x, y: Starting coordinates for the line
//   - dx, dy: Direction vector (1, 0, or -1 for each axis)
//   - length: Number of middle characters to draw
//   - start, middle, end: Unicode characters for line segments
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

// ScrollbarV renders a vertical scrollbar indicating scroll position and content size.
// This method creates a visual scrollbar with a thumb that represents the current
// scroll position and the proportion of visible content relative to total content.
//
// Parameters:
//   - x, y: Top-left coordinates for the scrollbar
//   - height: Height of the scrollbar area
//   - offset: Current scroll offset (how far scrolled from top)
//   - total: Total number of items/lines in the content
func (r *Renderer) ScrollbarV(x, y, height, offset, total int) {
	if height <= 0 || total <= 0 {
		return
	}

	// Calculate scrollbar thumb position and size
	thumb := min(max(height*height/total, 1), height)
	pos := min(max(offset*(height-thumb)/(total-height), 0), height-thumb)

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

// ScrollbarH renders a horizontal scrollbar indicating scroll position and content width.
// This method creates a visual scrollbar with a thumb that represents the current
// horizontal scroll position and the proportion of visible content relative to total content width.
//
// Parameters:
//   - x, y: Top-left coordinates for the scrollbar
//   - width: Width of the scrollbar area
//   - offset: Current horizontal scroll offset (how far scrolled from left)
//   - total: Total width of the content (in characters)
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

// Text renders a string at the specified coordinates with optional width limiting and padding.
// This is the fundamental text rendering method used by all widgets that display text content.
//
// Parameters:
//   - x, y: Starting coordinates for text rendering
//   - s: The string to render
//   - max: Maximum width for text (0 for no limit, >0 for width constraint)
//
// This method handles Unicode characters properly and provides the foundation
// for all text display in labels, buttons, inputs, and other text-based widgets.
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

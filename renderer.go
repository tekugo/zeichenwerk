package zeichenwerk

import (
	"github.com/gdamore/tcell/v2"
)

// ===========================================================================
// Attention! To keep the renderer.go file from getting to big, individual
// parts are put into separate render-*.go files especially for more
// complex widgets.
// ===========================================================================

// Screen represents an abstraction over terminal screen operations for rendering.
// This interface provides the essential methods needed for drawing content to
// the terminal, allowing for both direct screen access and viewport-based rendering.
//
// The interface abstracts the underlying tcell screen operations, enabling:
//   - Content reading and writing at specific coordinates
//   - Style application for visual formatting
//   - Viewport clipping for bounded rendering areas
//   - Testing and mocking of screen operations
//
// Implementations include the actual tcell.Screen for terminal output and
// Viewport for clipped rendering within specific rectangular areas. The interface
// was designed to mirror the tcell.Screen, so no adapter needed.
type Screen interface {
	// GetContent retrieves the character, combining characters, style, and width
	// at the specified screen coordinates. This is used for reading existing
	// content before modification or for implementing advanced rendering effects.
	//
	// Parameters:
	//   - x, y: Screen coordinates to read from
	//
	// Returns:
	//   - rune: Primary character at the position
	//   - []rune: Combining characters (for complex Unicode)
	//   - tcell.Style: Current style applied to the character
	//   - int: Character width (important for wide Unicode characters)
	GetContent(int, int) (rune, []rune, tcell.Style, int)

	// SetContent writes a character with optional combining characters and style
	// at the specified screen coordinates. This is the primary method for
	// drawing content to the terminal.
	//
	// Parameters:
	//   - x, y: Screen coordinates to write to
	//   - primary: Main character to display
	//   - combining: Additional combining characters for complex Unicode
	//   - style: Visual style (colors, attributes) to apply
	SetContent(int, int, rune, []rune, tcell.Style)
}

// Renderer handles the drawing and visual presentation of widgets to the terminal screen.
// It serves as the central rendering engine that converts widget hierarchies into
// terminal output, managing styles, clipping, and coordinate transformations.
//
// The Renderer provides:
//   - Widget-specific rendering methods for each widget type
//   - Style management and theme application
//   - Coordinate clipping for bounded rendering areas
//   - Primitive drawing operations (lines, borders, text)
//   - Visual effects (shadows, dimming, colorization)
//
// Rendering pipeline:
//  1. Widget bounds and content area calculation
//  2. Style resolution based on widget state and theme
//  3. Background and border rendering
//  4. Widget-specific content rendering
//  5. Child widget recursive rendering
//
// The renderer uses a viewport system for clipping, allowing widgets to render
// within their designated areas without affecting other screen regions.
type Renderer struct {
	theme  Theme       // Visual theme providing colors and styling rules
	screen Screen      // Current rendering target (may be clipped viewport)
	root   Screen      // Original unclipped screen for viewport management
	style  tcell.Style // Current active style for drawing operations
}

// SetStyle applies a visual style to the renderer for subsequent drawing operations.
// This method converts the high-level Style object into a tcell.Style by resolving
// theme colors and applying them to the renderer's current style state.
//
// Parameters:
//   - style: The Style object containing background, foreground, and other visual properties
func (r *Renderer) SetStyle(style *Style) {
	bg, _ := ParseColor(r.theme.Color(style.Background))
	fg, _ := ParseColor(r.theme.Color(style.Foreground))

	r.style = tcell.StyleDefault.Background(bg).Foreground(fg)
}

// ---- Internal Rendering Methods -------------------------------------------

// border draws a complete border around a rectangular area using the specified BorderStyle.
// This method renders the four sides and corners of a border, creating a frame around
// the given coordinates using Unicode box-drawing characters.
//
// Parameters:
//   - x, y: Top-left corner coordinates of the border area
//   - w, h: Width and height of the area to border (inner dimensions)
//   - box: BorderStyle containing the Unicode characters for each border element
//
// Border construction:
//  1. Draws top horizontal line with left corner, middle characters, and right corner
//  2. Draws bottom horizontal line with appropriate corner characters
//  3. Draws left and right vertical lines for the specified height
//
// The border is drawn outside the specified area, so the actual border occupies
// coordinates from (x,y) to (x+w+1, y+h+1), adding 1 character on each side.
func (r *Renderer) border(x, y, w, h int, box BorderStyle) {
	r.line(x, y, 1, 0, w, box.TopLeft, box.Top, box.TopRight)
	r.line(x, y+h+1, 1, 0, w, box.BottomLeft, box.Bottom, box.BottomRight)
	for i := range h {
		r.screen.SetContent(x, y+i+1, box.Left, nil, r.style)
		r.screen.SetContent(x+w+1, y+i+1, box.Right, nil, r.style)
	}
}

// clear fills a rectangular area with space characters using the current style.
// This method is used to clear backgrounds and create solid color areas by
// overwriting existing content with spaces in the current background color.
//
// Parameters:
//   - x, y: Top-left corner coordinates of the area to clear
//   - w, h: Width and height of the area to clear
//
// The clearing process:
//  1. Iterates through each position in the specified rectangle
//  2. Writes a space character with the current style at each position
//  3. Effectively creates a solid background color area
//
// This method is commonly used for:
//   - Widget background rendering
//   - Content area preparation before text rendering
//   - Creating solid color blocks for visual effects
func (r *Renderer) clear(x, y, w, h int) {
	cx, cy := x, y
	for range h {
		for range w {
			r.screen.SetContent(cx, cy, ' ', nil, r.style)
			cx++
		}
		cx = x
		cy++
	}
}

// clip establishes a rendering viewport that constrains all subsequent drawing operations
// to the content area of the specified widget. This implements clipping to prevent
// widgets from drawing outside their designated boundaries.
//
// Parameters:
//   - widget: The widget whose content area will define the clipping boundaries
//
// Clipping process:
//  1. Saves the current screen as the root screen for later restoration
//  2. Gets the widget's content area coordinates and dimensions
//  3. Creates a new Viewport that clips to the content area
//  4. Sets the clipped viewport as the active screen for rendering
//
// All subsequent drawing operations will be clipped to the widget's content area
// until unclip() is called. This ensures widgets cannot draw outside their bounds
// and enables safe rendering of child widgets within parent containers.
func (r *Renderer) clip(widget Widget) {
	r.root = r.screen
	x, y, w, h := widget.Content()
	r.screen = NewViewport(r.root, x, y, w, h)
}

// unclip restores the original unclipped screen for rendering operations.
// This method removes the current viewport clipping and returns to full-screen
// rendering mode, typically called after finishing widget-specific rendering.
//
// The restoration process:
//  1. Restores the original root screen as the active rendering target
//  2. Removes any viewport clipping constraints
//  3. Allows subsequent operations to draw anywhere on the screen
//
// This method should be called after clip() when the clipped rendering is complete,
// ensuring that subsequent widgets can render in their own coordinate spaces.
func (r *Renderer) unclip() {
	r.screen = r.root
}

func (r *Renderer) colorize(x, y, width, height int) {
	for i := range width {
		for j := range height {
			ch, _, _, _ := r.screen.GetContent(x+i, y+j)
			r.screen.SetContent(x+i, y+j, ch, nil, r.style)
		}
	}
}

func (r *Renderer) dim(x, y, width, height int) {
	for i := range width {
		for j := range height {
			ch, _, style, _ := r.screen.GetContent(x+i, y+j)
			r.screen.SetContent(x+i, y+j, ch, nil, style.Dim(true))
		}
	}
}

func (r *Renderer) line(x, y, dx, dy, length int, start, middle, end rune) {
	r.screen.SetContent(x, y, start, nil, r.style)
	cx := x + dx
	cy := y + dy
	for range length {
		r.screen.SetContent(cx, cy, middle, nil, r.style)
		cx += dx
		cy += dy
	}
	r.screen.SetContent(cx, cy, end, nil, r.style)
}

func (r *Renderer) repeat(x, y, dx, dy, length int, ch rune) {
	for range length {
		r.screen.SetContent(x, y, ch, nil, r.style)
		x += dx
		y += dy
	}
}

// text renders a string at the specified coordinates with optional width limiting and padding.
// This is the fundamental text rendering method used by all widgets that display text content.
//
// Parameters:
//   - x, y: Starting coordinates for text rendering
//   - s: The string to render
//   - max: Maximum width for text (0 for no limit, >0 for width constraint)
//
// Rendering behavior:
//  1. Iterates through each Unicode character in the string
//  2. Renders each character at the appropriate screen position
//  3. Stops rendering if the maximum width is exceeded
//  4. Pads remaining space with spaces if text is shorter than max width
//
// Width handling:
//   - max = 0: Renders the entire string without width constraints
//   - max > 0: Limits rendering to max characters and pads with spaces
//   - Truncation occurs if string exceeds max width
//   - Padding ensures consistent visual width for aligned layouts
//
// This method handles Unicode characters properly and provides the foundation
// for all text display in labels, buttons, inputs, and other text-based widgets.
func (r *Renderer) text(x, y int, s string, max int) {
	i := 0
	for _, ch := range s {
		if max > 0 && i >= max {
			break
		}
		r.screen.SetContent(x+i, y, ch, nil, r.style)
		i++
	}
	if max > 0 && i < max {
		for ; i < max; i++ {
			r.screen.SetContent(x+i, y, ' ', nil, r.style)
		}
	}
}

// ---- Widget Rendering -----------------------------------------------------

func (r *Renderer) render(widget Widget) {
	x, y, w, h := widget.Bounds()
	cx, cy, cw, ch := widget.Content()
	state := widget.State()
	style := widget.Style(":" + state)

	r.SetStyle(style)

	switch widget := widget.(type) {
	case *Box:
		r.renderBorder(x, y, w, h, style)
		r.SetStyle(widget.Style("title"))
		r.text(x+2, y, " "+widget.Title+" ", 0)
		r.render(widget.child)
	case *Button:
		r.renderBorder(x, y, w, h, style)
		r.text(cx, cy, widget.Text, cw)
	case *Checkbox:
		r.renderBorder(x, y, w, h, style)
		r.renderCheckbox(widget, cx, cy, cw, ch)
	case *Custom:
		r.renderBorder(x, y, w, h, style)
		widget.renderer(widget, r.screen)
	case *Editor:
		r.renderBorder(x, y, w, h, style)
		r.renderEditor(widget, cx, cy, cw, ch)
	case *Flex:
		r.renderBorder(x, y, w, h, style)
		if widget.ID() == "popup" {
			r.renderShadow(x, y, w, h, widget.Style("shadow"))
		}
		for _, child := range widget.Children(true) {
			r.render(child)
		}
	case *Grid:
		r.renderBorder(x, y, w, h, style)
		r.renderGrid(widget, r.theme.Border(widget.Style("").Border))
	case *Input:
		r.renderBorder(x, y, w, h, style)
		r.renderInput(widget, cx, cy, cw, ch)
	case *Label:
		r.renderBorder(x, y, w, h, style)
		r.text(cx, cy, widget.Text, 0)
	case *List:
		r.renderBorder(x, y, w, h, style)
		r.renderList(widget, cx, cy, cw, ch)
	case *ProgressBar:
		r.renderBorder(x, y, w, h, style)
		r.renderProgressBar(widget, x, y, w, h)
	case *Scroller:
		r.renderBorder(x, y, w, h, style)
		r.renderScroller(widget)
	case *Separator:
		if widget.Border != "" {
			box := r.theme.Border(widget.Border)
			if style.Height == 1 {
				r.line(cx, cy, 1, 0, cw-2, box.Top, box.Top, box.Top)
			} else {
				r.line(cx, cy, 0, 1, ch-2, box.Left, box.Left, box.Left)
			}
		}
	case *Switcher:
		r.render(widget.Panes[widget.Selected])
	case *Table:
		r.renderBorder(x, y, w, h, style)
		r.renderTable(widget)
	case *Tabs:
		r.renderTabs(widget, cx, cy, cw)
	case *Text:
		r.renderBorder(x, y, w, h, style)
		r.renderText(widget, cx, cy, cw, ch)
	case *ThemeSwitch:
		old := r.theme
		r.theme = widget.theme
		r.render(widget.child)
		r.theme = old
	}
}

func (r *Renderer) renderBorder(x, y, w, h int, style *Style) {
	if style.Background != "" {
		r.clear(x+style.Margin.Left, y+style.Margin.Top, w-style.Margin.Left-style.Margin.Right, h-style.Margin.Top-style.Margin.Bottom)
	}
	if style.Border != "" {
		box := r.theme.Border(style.Border)
		r.border(x+style.Margin.Left, y+style.Margin.Top, w-style.Margin.Left-style.Margin.Right-2, h-style.Margin.Top-style.Margin.Bottom-2, box)
	}
}

func (r *Renderer) renderShadow(x, y, w, h int, style *Style) {
	r.SetStyle(style)
	r.colorize(x+2, y+h, w, 1)
	r.colorize(x+w, y+1, 2, h)
}

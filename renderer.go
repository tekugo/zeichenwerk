// Package renderer.go implements the core rendering engine for zeichenwerk.
//
// This file contains the central Renderer type and Screen abstraction that
// handle the visual presentation of widgets to the terminal. The rendering
// system provides:
//   - Widget-specific rendering methods for all widget types
//   - Style management and theme integration
//   - Coordinate clipping and viewport management
//   - Primitive drawing operations (text, lines, borders, backgrounds)
//   - Visual effects (shadows, dimming, colorization)
//
// # Architecture
//
// The rendering system uses a two-level architecture:
//   - Screen interface: Abstraction over terminal operations
//   - Renderer struct: High-level widget rendering coordination
//
// # Modular Design
//
// Complex widget rendering is split across multiple files:
//   - renderer.go: Core renderer and simple widgets
//   - render-*.go: Specialized rendering for complex widgets
//   - This keeps file sizes manageable and organizes related functionality

package zeichenwerk

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

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
// # Core Capabilities
//
// The Renderer provides:
//   - Widget-specific rendering methods for each widget type
//   - Style management and theme application
//   - Coordinate clipping for bounded rendering areas
//   - Primitive drawing operations (lines, borders, text)
//   - Visual effects (shadows, dimming, colorization)
//   - Unicode support for complex characters and box drawing
//
// # Rendering Pipeline
//
// The standard rendering process follows these steps:
//  1. Widget bounds and content area calculation
//  2. Style resolution based on widget state and theme
//  3. Background and border rendering (via renderBorder)
//  4. Widget-specific content rendering (delegated to render-*.go files)
//  5. Child widget recursive rendering for containers
//
// # Viewport System
//
// The renderer uses a viewport system for clipping that enables:
//   - Widgets to render safely within their designated areas
//   - Child widgets to use relative coordinates within parent bounds
//   - Efficient rendering without manual coordinate transformation
//   - Prevention of visual artifacts from widgets drawing outside bounds
//
// # State Management
//
// The renderer maintains minimal state to support rendering operations:
//   - Current theme for style resolution and color mapping
//   - Active screen (either full screen or clipped viewport)
//   - Root screen reference for viewport management
//   - Current tcell.Style for immediate drawing operations
//
// # Thread Safety
//
// Renderer instances are not thread-safe and should only be used from
// the main UI thread. Multiple renderers can exist simultaneously for
// different UI instances.
type Renderer struct {
	theme  Theme       // Visual theme providing colors, border styles, and styling rules
	screen Screen      // Current rendering target (full screen or clipped viewport)
	root   Screen      // Original unclipped screen saved during clip operations
	style  tcell.Style // Current active tcell.Style for immediate drawing operations
}

// SetStyle applies a visual style to the renderer for subsequent drawing operations.
// This method converts the high-level Style object into a tcell.Style by resolving
// theme colors and font attributes, then stores it as the current rendering style.
//
// # Style Conversion Process
//
// The method performs the following conversions:
//  1. Resolves background color through theme.Color() for variable substitution
//  2. Resolves foreground color through theme.Color() for variable substitution
//  3. Parses font attribute string for text styling options
//  4. Applies all attributes to create a complete tcell.Style
//  5. Stores the result for use in subsequent drawing operations
//
// # Color Resolution
//
// Colors are processed through the theme's Color() method to support:
//   - Direct color names: "red", "blue", "#FF0000"
//   - Theme variables: "$primary", "$background", "$accent"
//   - Invalid colors are silently ignored, preserving default values
//
// # Font Attribute Processing
//
// The font string is parsed as a comma-separated list of attributes:
//   - "bold": Enables bold text rendering
//   - "italic": Enables italic text rendering
//   - "underline": Enables underlined text rendering
//   - "strikethrough": Enables strikethrough text rendering
//   - "blink": Enables blinking text (terminal dependent)
//   - "normal": Resets all attributes to default state
//
// # Usage Pattern
//
// This method is typically called before drawing operations:
//  1. Widget resolves its current style based on state and theme
//  2. SetStyle() converts the high-level style to tcell format
//  3. Subsequent drawing operations use the prepared tcell.Style
//
// Parameters:
//   - style: The Style object containing colors and font attributes to apply
func (r *Renderer) SetStyle(style *Style) {
	result := tcell.StyleDefault

	if style.Background != "" {
		if bg, err := ParseColor(r.theme.Color(style.Background)); err == nil {
			result = result.Background(bg)
		}
	}
	if style.Foreground != "" {
		if fg, err := ParseColor(r.theme.Color(style.Foreground)); err == nil {
			result = result.Foreground(fg)
		}
	}

	for part := range strings.SplitSeq(style.Font, ",") {
		option := strings.ToLower(strings.TrimSpace(part))
		switch option {
		case "blink":
			result = result.Blink(true)
		case "bold":
			result = result.Bold(true)
		case "normal":
			result = result.Blink(false).Bold(false).Italic(false).Underline(false).StrikeThrough(false)
		case "italic":
			result = result.Italic(true)
		case "strikethrough":
			result = result.StrikeThrough(true)
		case "underline":
			result = result.Underline(true)
		}
	}

	r.style = result
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

// colorize applies the current renderer style to all characters in a rectangular area.
// This method preserves the existing characters while changing their visual styling,
// effectively recoloring or reformatting text and graphics within the specified region.
//
// # Operation
//
// For each position in the rectangle:
//  1. Reads the existing character at that position
//  2. Rewrites the same character with the current renderer style
//  3. Preserves character content while updating visual appearance
//
// # Use Cases
//
// Common applications include:
//   - Highlighting text areas or selection regions
//   - Applying focus styling to widget areas
//   - Creating visual emphasis or attention effects
//   - Updating styling without changing content
//
// Parameters:
//   - x, y: Top-left corner of the area to colorize
//   - width, height: Dimensions of the area to process
func (r *Renderer) colorize(x, y, width, height int) {
	for i := range width {
		for j := range height {
			ch, _, _, _ := r.screen.GetContent(x+i, y+j)
			r.screen.SetContent(x+i, y+j, ch, nil, r.style)
		}
	}
}

// dim applies a dimming effect to all characters in a rectangular area.
// This method preserves existing characters and their styles while adding
// a dimming attribute to reduce visual prominence.
//
// # Operation
//
// For each position in the rectangle:
//  1. Reads the existing character and its current style
//  2. Applies the dim attribute to the existing style
//  3. Rewrites the character with the dimmed style
//
// # Visual Effect
//
// The dimming effect typically:
//   - Reduces color saturation and brightness
//   - Makes text appear faded or less prominent
//   - Provides visual hierarchy by de-emphasizing content
//   - Maintains readability while indicating secondary importance
//
// # Use Cases
//
// Common applications include:
//   - Disabled widget states
//   - Background or inactive content
//   - Shadow effects for overlays
//   - Visual depth and layering
//
// Parameters:
//   - x, y: Top-left corner of the area to dim
//   - width, height: Dimensions of the area to process
func (r *Renderer) dim(x, y, width, height int) {
	for i := range width {
		for j := range height {
			ch, _, style, _ := r.screen.GetContent(x+i, y+j)
			r.screen.SetContent(x+i, y+j, ch, nil, style.Dim(true))
		}
	}
}

// line draws a line using specified start, middle, and end characters.
// This method is used for drawing borders, separators, and decorative elements
// with proper terminal characters for clean visual presentation.
//
// # Drawing Process
//
// The line is constructed in three parts:
//  1. Start character at the initial position
//  2. Middle character repeated for the specified length
//  3. End character at the final position
//
// # Direction Control
//
// The dx and dy parameters control line direction:
//   - dx=1, dy=0: Horizontal line (left to right)
//   - dx=0, dy=1: Vertical line (top to bottom)
//   - dx=-1, dy=0: Horizontal line (right to left)
//   - dx=0, dy=-1: Vertical line (bottom to top)
//   - Other combinations: Diagonal lines
//
// # Use Cases
//
// Common applications include:
//   - Border drawing with corner and edge characters
//   - Separator lines in layouts
//   - Progress bar segments
//   - Decorative elements and dividers
//
// Parameters:
//   - x, y: Starting coordinates for the line
//   - dx, dy: Direction vector (1, 0, or -1 for each axis)
//   - length: Number of middle characters to draw
//   - start, middle, end: Unicode characters for line segments
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

// repeat draws a single character multiple times in a specified direction.
// This method is used for creating patterns, filling areas, and drawing
// simple graphical elements like progress indicators or decorative patterns.
//
// # Drawing Process
//
// The method places the specified character at consecutive positions:
//  1. Starts at the initial coordinates
//  2. Draws the character at each position
//  3. Advances by dx and dy for each iteration
//  4. Continues for the specified length
//
// # Direction Control
//
// The dx and dy parameters determine the pattern direction:
//   - dx=1, dy=0: Horizontal pattern (left to right)
//   - dx=0, dy=1: Vertical pattern (top to bottom)
//   - Other combinations: Diagonal or complex patterns
//
// # Use Cases
//
// Common applications include:
//   - Progress bar fill indicators
//   - Pattern backgrounds
//   - Simple border elements
//   - Separator dots or dashes
//   - Loading animations
//
// Parameters:
//   - x, y: Starting coordinates for the pattern
//   - dx, dy: Direction vector for pattern advancement
//   - length: Number of characters to draw
//   - ch: Unicode character to repeat
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
	for _, ch := range []rune(s) {
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

// render performs the complete rendering process for a widget and its children.
// This is the main entry point for widget rendering that handles the dispatch
// to widget-specific rendering methods and coordinates the overall rendering pipeline.
//
// # Rendering Process
//
// The method follows a standardized process for all widgets:
//  1. Extracts widget bounds and content area coordinates
//  2. Determines widget state (focus, hover, disabled, etc.)
//  3. Resolves widget style based on current state and theme
//  4. Sets the resolved style as the current renderer style
//  5. Dispatches to widget-specific rendering logic
//  6. Handles recursive rendering for container widgets
//
// # Widget Type Dispatch
//
// The method uses type switching to dispatch to appropriate rendering logic:
//   - Simple widgets: Rendered directly in this method
//   - Complex widgets: Delegated to specialized render methods
//   - Container widgets: Recursively render all visible children
//   - Custom widgets: Call user-provided rendering functions
//
// # State-Based Styling
//
// Widget styling is determined by current state:
//   - ":focus" - Widget has keyboard focus
//   - ":hover" - Widget is under mouse cursor
//   - ":disabled" - Widget is in disabled state
//   - Default state uses empty string selector
//
// # Border and Background Rendering
//
// Most widgets follow the pattern:
//  1. renderBorder() for background and border styling
//  2. Widget-specific content rendering within content area
//  3. Child widget rendering for containers
//
// # Clipping and Coordinates
//
// The renderer automatically handles:
//   - Coordinate transformation for widget positioning
//   - Clipping to prevent drawing outside widget bounds
//   - Content area calculation based on margins, padding, borders
//
// Parameters:
//   - widget: The widget to render (including all its children)
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
		r.text(x+style.Margin.Left+2, y+style.Margin.Top, " "+widget.Title+" ", 0)
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
	case *Dialog:
		r.renderBorder(x, y, w, h, style)
		r.SetStyle(widget.Style("title"))
		r.text(x+style.Margin.Left+2, y+style.Margin.Top, " "+widget.Title+" ", 0)
		r.render(widget.child)
	case *Digits:
		r.renderBorder(x, y, w, h, style)
		r.renderDigits(widget, cx, cy, cw, ch)
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
	case *Form:
		r.renderBorder(x, y, w, h, style)
		r.render(widget.child)
	case *FormGroup:
		r.renderBorder(x, y, w, h, style)
		r.renderFormGroup(widget)
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
		if widget.Border != "" && widget.Border != "none" {
			box := r.theme.Border(widget.Border)
			if style.Height == 1 {
				r.line(cx, cy, 1, 0, cw-2, box.Top, box.Top, box.Top)
			} else {
				r.line(cx, cy, 0, 1, ch-2, box.Left, box.Left, box.Left)
			}
		}
	case *Spinner:
		r.renderBorder(x, y, w, h, style)
		r.screen.SetContent(cx, cy, widget.Rune(), nil, r.style)
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

// renderBorder draws the background and border for a widget according to its style.
// This method handles the standard border and background rendering that most widgets
// use, respecting margin spacing and border style definitions from the theme.
//
// # Rendering Order
//
// The method renders in this order:
//  1. Background fill (if background color is specified)
//  2. Border drawing (if border style is specified)
//
// # Coordinate Calculation
//
// The method respects margin settings from the style:
//   - Background: Fills area inside margins but outside border
//   - Border: Drawn inside margins, reducing content area by border thickness
//   - Content area: Calculated as remaining space after margins, borders, and padding
//
// # Background Rendering
//
// If style.Background is specified:
//   - Fills the area with spaces using the background color
//   - Respects left/right/top/bottom margins
//   - Creates solid color background for widget content
//
// # Border Rendering
//
// If style.Border is specified:
//   - Retrieves BorderStyle from theme using the border name
//   - Draws complete border with corners and edges
//   - Reduces available content area by 2 characters (1 per side)
//
// # Style Requirements
//
// The style parameter should contain:
//   - Background: Color name or theme variable (optional)
//   - Border: Border style name defined in theme (optional)
//   - Margin: Spacing values for positioning (required)
//
// Parameters:
//   - x, y: Top-left coordinates of the widget's total area
//   - w, h: Total width and height of the widget area
//   - style: Style containing background, border, and margin specifications
func (r *Renderer) renderBorder(x, y, w, h int, style *Style) {
	if style.Border == "" || style.Border == "none" {
		return
	}
	if style.Background != "" {
		r.clear(x+style.Margin.Left, y+style.Margin.Top, w-style.Margin.Left-style.Margin.Right, h-style.Margin.Top-style.Margin.Bottom)
	}
	if style.Border != "" && style.Border != "none" {
		parts := strings.Split(style.Border, " ")
		if len(parts) > 1 {
			fg := parts[1]
			bg := style.Background
			if len(parts) > 2 {
				bg = parts[2]
			}
			r.SetStyle(NewStyle(fg, bg))
		}
		box := r.theme.Border(style.Border)
		r.border(x+style.Margin.Left, y+style.Margin.Top, w-style.Margin.Left-style.Margin.Right-2, h-style.Margin.Top-style.Margin.Bottom-2, box)
		r.SetStyle(style)
	}
}

// renderShadow creates a drop shadow effect for popup widgets and dialogs.
// This method draws a simple shadow by colorizing areas to the right and bottom
// of the widget, creating a visual depth effect that makes popups appear elevated.
//
// # Shadow Design
//
// The shadow consists of two rectangular areas:
//  1. Bottom shadow: Horizontal strip below the widget (width × 1 character)
//  2. Right shadow: Vertical strip to the right of the widget (2 × height characters)
//
// # Shadow Positioning
//
// The shadow is offset from the widget boundaries:
//   - Bottom shadow: Starts 2 characters to the right, extends full width
//   - Right shadow: Starts 1 character down, extends full height
//   - This creates an overlapping corner effect for visual continuity
//
// # Visual Effect
//
// The shadow effect:
//   - Uses the provided style (typically darker colors)
//   - Preserves existing characters while changing their appearance
//   - Creates illusion of widget floating above background
//   - Provides visual separation for modal dialogs and popups
//
// # Usage Context
//
// This method is typically called for:
//   - Popup dialogs and modal windows
//   - Context menus and dropdown overlays
//   - Floating panels and tool windows
//   - Any widget that should appear elevated
//
// Parameters:
//   - x, y: Top-left coordinates of the widget casting the shadow
//   - w, h: Width and height of the widget
//   - style: Style for the shadow appearance (colors and attributes)
func (r *Renderer) renderShadow(x, y, w, h int, style *Style) {
	r.SetStyle(style)
	r.colorize(x+2, y+h, w, 1)
	r.colorize(x+w, y+1, 2, h)
}

package zeichenwerk

import (
	"github.com/gdamore/tcell/v3"
)

// Canvas represents a virtual terminal screen buffer with individually styled
// character cells. It provides a low-level drawing surface for creating custom
// terminal UI components and art. The canvas maintains its own 2D buffer and
// supports cursor navigation with both arrow keys and vim-style keys in
// different modes.
type Canvas struct {
	Component
	mode    string     // current mode
	cells   [][][]Cell // 2D buffer of pages and their character cells
	page    int        // Current page
	cursorX int        // cursor column (0-based, relative to content)
	cursorY int        // cursor row (0-based, relative to content)
	visualX int        // visual mode starting column
	visualY int        // visual mode starting row
}

// Cell represents a single character cell in the canvas buffer.
type Cell struct {
	ch    string // character to display
	style *Style // style for rendering this cell
}

// Canvas modes
const (
	ModeNormal  = "NORMAL"
	ModeCommand = "COMMAND"
	ModeDraw    = "DRAW"
	ModeInsert  = "INSERT"
	ModePresent = "PRESENT"
	ModeVisual  = "VISUAL"
)

// NewCanvas creates a new Canvas widget with the specified ID and dimensions.
// The canvas is focusable by default and registers a key handler for cursor
// movement and character insertion. The content size is set to match the
// buffer dimensions so the container knows the canvas's preferred size.
//
// Parameters:
//   - id: Unique identifier for the canvas widget
//   - width: Width of the canvas in character cells
//   - height: Height of the canvas in character cells
//
// Returns:
//   - *Canvas: A new canvas instance ready for use
func NewCanvas(id, class string, pages, width, height int) *Canvas {
	c := &Canvas{
		Component: Component{id: id, class: class},
		page:      0,
		cursorX:   0,
		cursorY:   0,
		mode:      ModeNormal, // start in normal mode
	}

	// Initialize the cell buffer
	c.cells = make([][][]Cell, pages)
	for i := range pages {
		c.cells[i] = make([][]Cell, height)
		for y := range height {
			c.cells[i][y] = make([]Cell, width)
		}
	}

	// Set the content hint to match the canvas dimensions
	c.SetHint(width, height)

	// Make it focusable by default
	c.SetFlag(FlagFocusable, true)

	// Register key handler for cursor movement and editing
	OnKey(c, c.handleKey)

	return c
}

// Apply applies a theme style to the component.
func (c *Canvas) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("canvas"))
}

// CellAt returns a pointer to the cell at the specified position, or nil if
// the coordinates are out of bounds. The coordinates are relative to the
// canvas buffer (0,0 is top-left).
func (c *Canvas) Cell(x, y int) *Cell {
	if c.page >= len(c.cells) || y < 0 || y >= len(c.cells[c.page]) || x < 0 || x >= len(c.cells[c.page][y]) {
		return nil
	}
	return &c.cells[c.page][y][x]
}

// Clear removes all content from the canvas by resetting every cell to an
// empty string with a default style. The cursor position is reset to (0,0)
// and the widget is refreshed.
func (c *Canvas) Clear() {
	defaultStyle := NewStyle("")
	for y := range c.cells[c.page] {
		for x := range c.cells[c.page][y] {
			c.cells[c.page][y][x] = Cell{ch: "", style: defaultStyle}
		}
	}
	c.cursorX, c.cursorY = 0, 0
	c.Refresh()
}

// Cursor returns the current cursor position relative to the canvas content
// area, along with the cursor style string. When the canvas is not focused,
// it returns (-1, -1, "") to hide the cursor. The cursor style varies based
// on the current mode: block in normal mode, bar in insert mode. This can be
// overridden via the widget's style configuration.
func (c *Canvas) Cursor() (int, int, string) {
	if !c.Flag(FlagFocused) {
		return -1, -1, ""
	}
	cursor := c.Style().Cursor()
	if cursor == "" {
		// Default cursor style based on mode
		if c.mode == ModeInsert {
			cursor = "bar"
		} else if c.mode == ModePresent {
			cursor = "none"
		} else {
			cursor = "block"
		}
	}
	return c.cursorX, c.cursorY, cursor
}

// Fill sets every cell in the canvas to the specified character and style.
// If style is nil, the default style is used. This provides a fast way to
// clear or uniformly populate the canvas.
func (c *Canvas) Fill(ch string, style *Style) {
	if style == nil {
		style = NewStyle("")
	}
	for y := range c.cells[c.page] {
		for x := range c.cells[c.page][y] {
			c.cells[c.page][y][x] = Cell{ch: ch, style: style}
		}
	}
	c.Refresh()
}

// Mode returns the current mode of the canvas ("normal" or "insert").
func (c *Canvas) Mode() string {
	return c.mode
}

// Refresh redraws the canvas widget
func (c *Canvas) Refresh() {
	Redraw(c)
}

// Resize changes the pages and cell buffers to the new size.
// The cursor and current page are clamped to the new bounds.
func (c *Canvas) Resize(pages, rows, columns int) {
	// Grow or shrink the page slice.
	if len(c.cells) < pages {
		extra := make([][][]Cell, pages-len(c.cells))
		for i := range extra {
			extra[i] = make([][]Cell, rows)
			for y := range extra[i] {
				extra[i][y] = make([]Cell, columns)
			}
		}
		c.cells = append(c.cells, extra...)
	} else if len(c.cells) > pages {
		c.cells = c.cells[:pages]
	}

	// Grow or shrink each existing page's row count.
	for i, page := range c.cells {
		if len(page) < rows {
			extra := make([][]Cell, rows-len(page))
			for y := range extra {
				extra[y] = make([]Cell, columns)
			}
			c.cells[i] = append(c.cells[i], extra...)
		} else if len(page) > rows {
			c.cells[i] = c.cells[i][:rows]
		}
	}

	// Grow or shrink each row's column count.
	for i, page := range c.cells {
		for j, row := range page {
			if len(row) < columns {
				c.cells[i][j] = append(c.cells[i][j], make([]Cell, columns-len(row))...)
			} else if len(row) > columns {
				c.cells[i][j] = c.cells[i][j][:columns]
			}
		}
	}

	// Clamp page and cursor to the new bounds.
	if c.page >= pages {
		c.page = pages - 1
	}
	if rows > 0 && c.cursorY >= rows {
		c.cursorY = rows - 1
	}
	if columns > 0 && c.cursorX >= columns {
		c.cursorX = columns - 1
	}
}

// SetCell updates the character and style at the given position. If the
// coordinates are out of bounds, the operation is ignored. A nil style
// uses the default style. The widget is automatically refreshed.
func (c *Canvas) SetCell(x, y int, ch string, style *Style) {
	if c.page >= len(c.cells) || y < 0 || y >= len(c.cells[c.page]) || x < 0 || x >= len(c.cells[c.page][y]) {
		return
	}
	if style == nil {
		style = NewStyle("")
	}
	c.cells[c.page][y][x] = Cell{ch: ch, style: style}
	c.Dispatch(c, EvtChange)
	c.Refresh()
}

// SetCursor moves the cursor to the specified position without triggering
// a refresh. Useful for programmatic cursor positioning.
func (c *Canvas) SetCursor(x, y int) {
	c.cursorX = x
	c.cursorY = y
	c.Dispatch(c, EvtMove,x, y)
}

// SetMode sets the canvas mode. Valid modes are ModeNormal and ModeInsert.
// The widget is refreshed after the mode change to update the cursor style.
func (c *Canvas) SetMode(mode string) {
	c.mode = mode
	c.Refresh()
	c.Dispatch(c, EvtMode,c.mode)
}

// SetPage sets the current page
func (c *Canvas) SetPage(page int) {
	if page >= 0 && page < len(c.cells) {
		c.page = page
	}
}

// Size returns the logical dimensions of the canvas buffer in character cells.
// This differs from Content() which returns the positioned area including
// margins, padding, and borders. Size returns the actual buffer dimensions.
func (c *Canvas) Size() (width, height int) {
	if len(c.cells) == 0 {
		return 0, 0
	}
	return len(c.cells[0][0]), len(c.cells[0])
}

// handleKey processes keyboard events based on the current mode. In normal
// mode, keys are used for movement and mode switching. In insert mode,
// printable characters are inserted and ESC returns to normal mode.
// Returns true if the event was handled.
func (c *Canvas) handleKey(_ Widget, evt *tcell.EventKey) bool {
	switch c.mode {
	case ModeNormal:
		return c.handleNormalMode(evt)
	case ModeInsert:
		return c.handleInsertMode(evt)
	}
	return false
}

// handleNormalMode handles key events in normal mode. Supports arrow keys,
// vim keys (h,j,k,l), Home/End for navigation, and 'i' or 'a' to enter
// insert mode. ESC is ignored in normal mode. Returns true if handled.
func (c *Canvas) handleNormalMode(evt *tcell.EventKey) bool {
	switch evt.Key() {
	case tcell.KeyUp:
		c.move(0, -1)
		return true
	case tcell.KeyDown:
		c.move(0, 1)
		return true
	case tcell.KeyLeft:
		c.move(-1, 0)
		return true
	case tcell.KeyRight:
		c.move(1, 0)
		return true
	case tcell.KeyHome:
		c.cursorX = 0
		c.Refresh()
		return true
	case tcell.KeyEnd:
		c.cursorX = len(c.cells[c.page][0]) - 1
		c.Refresh()
		return true
	case tcell.KeyEsc:
		// Already in normal mode, ignore
		return true
	case tcell.KeyRune:
		ch := evt.Str()
		switch ch {
		case "h":
			c.move(-1, 0)
			return true
		case "j":
			c.move(0, 1)
			return true
		case "k":
			c.move(0, -1)
			return true
		case "l":
			c.move(1, 0)
			return true
		case "i", "a":
			c.mode = ModeInsert
			c.Refresh()
			return true
		}
		return false
	default:
		return false
	}
}

// handleInsertMode handles key events in insert mode. Printable characters
// are inserted at the cursor position. ESC switches back to normal mode.
// Arrow keys can also be used for navigation (staying in insert mode).
// Returns true if the event was handled.
func (c *Canvas) handleInsertMode(evt *tcell.EventKey) bool {
	switch evt.Key() {
	case tcell.KeyEsc:
		c.mode = ModeNormal
		c.Refresh()
		return true
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
		// Allow navigation in insert mode without changing mode
		c.handleNormalMode(evt)
		return true
	case tcell.KeyHome:
		c.cursorX = 0
		c.Refresh()
		return true
	case tcell.KeyEnd:
		c.cursorX = len(c.cells[c.page][0]) - 1
		c.Refresh()
		return true
	case tcell.KeyRune:
		ch := evt.Str()
		if len(ch) == 1 && ch[0] >= 32 && ch[0] <= 126 { // Basic printable ASCII
			c.insertCharacter(ch)
			return true
		}
		return false
	default:
		return false
	}
}

// move translates the cursor by the specified delta, clamping the result to
// the canvas bounds. The widget is refreshed after movement.
func (c *Canvas) move(dx, dy int) {
	newX := c.cursorX + dx
	newY := c.cursorY + dy

	if newX < 0 {
		newX = 0
	} else if newX >= len(c.cells[c.page][0]) {
		newX = len(c.cells[c.page][0]) - 1
	}

	if newY < 0 {
		newY = 0
	} else if newY >= len(c.cells[c.page]) {
		newY = len(c.cells[c.page]) - 1
	}

	c.cursorX, c.cursorY = newX, newY
	c.Dispatch(c, EvtMove,newX, newY)
	c.Refresh()
}

// insertCharacter places the character at the current cursor position,
// preserving the cell's existing style if present, otherwise using the
// widget's default style. After insertion, the cursor advances right
// when possible.
func (c *Canvas) insertCharacter(ch string) {
	cell := c.Cell(c.cursorX, c.cursorY)
	style := c.Style()
	if cell != nil && cell.style != nil {
		style = cell.style
	}
	c.SetCell(c.cursorX, c.cursorY, ch, style)
	if c.cursorX < len(c.cells[0])-1 {
		c.cursorX++
	}
}

// Render draws the canvas content to the screen. It iterates over the
// visible content area and renders each cell's character with its
// associated style. Empty cells are filled with spaces using the widget's
// background color.
func (c *Canvas) Render(r *Renderer) {
	x0, y0, w, h := c.Content()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := c.Cell(x, y)
			if cell != nil && cell.ch != "" {
				style := cell.style
				if style == nil {
					style = c.Style()
				}
				r.Set(style.Foreground(), style.Background(), style.Font())
				r.Text(x0+x, y0+y, cell.ch, 1)
			} else {
				style := c.Style()
				if style.Background() != "" {
					r.Set(style.Foreground(), style.Background(), style.Font())
					r.Put(x0+x, y0+y, " ")
				}
			}
		}
	}
}

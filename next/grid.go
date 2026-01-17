package next

// Cell represents a single cell within a grid container that holds a widget
// and defines its position and span within the grid layout.
//
// A cell specifies:
//   - Position: The starting row and column coordinates (x, y)
//   - Span: How many columns (w) and rows (h) the cell occupies
//   - Content: The widget contained within this cell
//
// Cells can span multiple rows and/or columns, allowing for complex
// grid layouts with widgets of varying sizes.
type Cell struct {
	x, y    int    // Starting column (x) and row (y) position in the grid (0-based)
	w, h    int    // Column span (w) and row span (h) - number of cells to occupy
	content Widget // The widget contained within this grid cell
}

// Grid line separator constants used for rendering grid lines and borders.
// These constants define which grid lines should be drawn around cells
// to create visual separation between grid elements.
const (
	GridH = 1 // Horizontal grid line - draws a line to the right of the cell
	GridV = 2 // Vertical grid line - draws a line below the cell
	GridB = 3 // Both horizontal and vertical grid lines (intersection)
)

// Grid is a container widget that arranges child widgets in a table-like layout
// with configurable rows and columns. It provides precise control over widget
// positioning and supports flexible sizing, spanning, and optional grid lines.
//
// Features:
//   - Fixed grid dimensions with configurable rows and columns
//   - Cell spanning: widgets can occupy multiple rows/columns
//   - Flexible sizing: fractional units for responsive layouts
//   - Optional grid lines for visual separation
//   - Precise positioning control for complex layouts
//
// Layout behavior:
//   - Fixed sizes: Positive values are treated as absolute sizes
//   - Flexible sizes: Negative values are fractional units of available space
//   - Grid lines: Optional visual separators between cells
//   - Cell spanning: Widgets can span multiple rows and/or columns
type Grid struct {
	Component
	cells           []*Cell // All cells containing widgets in the grid
	rows, columns   []int   // Size configuration for each row and column (-1 = flexible)
	widths, heights []int   // Calculated actual sizes for each column and row
	lines           bool    // Whether to draw grid lines between cells
	separators      [][]int // Grid line configuration for each cell intersection
}

// NewGrid creates a new grid container widget with the specified dimensions and configuration.
// The grid is initialized with flexible sizing for all rows and columns by default,
// meaning they will share available space equally unless explicitly configured otherwise.
//
// Parameters:
//   - id: Unique identifier for the grid widget
//   - rows: Number of rows in the grid
//   - columns: Number of columns in the grid
//   - lines: Whether to draw grid lines between cells for visual separation
//
// Returns:
//   - *Grid: A new grid container widget instance
//
// Default behavior:
//   - All rows and columns are initially set to flexible sizing (-1)
//   - Grid line separators are configured for all intersections
//   - No widgets are initially placed in the grid
//
// Example usage:
//
//	// Create a 3x3 grid with grid lines
//	grid := NewGrid("main-grid", 3, 3, true)
//
//	// Create a 2x4 grid without grid lines
//	layout := NewGrid("layout", 2, 4, false)
func NewGrid(id string, rows, columns int, lines bool) *Grid {
	grid := Grid{
		Component:  Component{id: id},
		cells:      make([]*Cell, 0, rows*columns),
		rows:       make([]int, rows),
		columns:    make([]int, columns),
		widths:     make([]int, columns),
		heights:    make([]int, rows),
		lines:      lines,
		separators: make([][]int, rows),
	}
	for i := range rows {
		grid.separators[i] = make([]int, columns)
		grid.rows[i] = -1 // Default to fractional rows (flexible sizing)
		for j := range columns {
			grid.separators[i][j] = GridB // Default to both grid lines
		}
	}
	for i := range columns {
		grid.columns[i] = -1 // Default to fractional columns (flexible sizing)
	}
	return &grid
}

// Children returns a slice of all child widgets contained in the grid cells.
// The widgets are returned in the order they were added to the grid, which
// may not correspond to their visual position within the grid layout.
//
// Returns:
//   - []Widget: A slice containing all child widgets in the grid
func (g *Grid) Children() []Widget {
	result := make([]Widget, len(g.cells))
	for i, cell := range g.cells {
		result[i] = cell.content
	}
	return result
}

// Add places a widget in the grid at the specified position with the given span.
// The widget will occupy a rectangular area from (x,y) spanning w columns and h rows.
// If the specified area extends beyond the grid boundaries, it will be clipped.
//
// Parameters:
//   - x: Starting column position (0-based)
//   - y: Starting row position (0-based)
//   - w: Number of columns to span (width)
//   - h: Number of rows to span (height)
//   - content: The widget to place in the grid
//
// The widget's parent is automatically set to this grid container.
//
// Example usage:
//
//	// Place a button in the top-left cell
//	grid.Add(0, 0, 1, 1, button)
//
//	// Place a text widget spanning 2 columns and 3 rows
//	grid.Add(1, 0, 2, 3, textWidget)
func (g *Grid) Add(x, y, w, h int, content Widget) {
	// If no width or height is specified, minimum is 1 cell
	if w == 0 {
		w = 1
	}
	if h == 0 {
		h = 1
	}

	// Check width boundaries
	if x >= len(g.columns) {
		x = len(g.columns) - 1
		w = 1
	}
	if x+w > len(g.columns) {
		w = len(g.columns) - x
	}

	// Check height boundaries
	if y >= len(g.rows) {
		y = len(g.rows) - 1
		h = 1
	}
	if y+h > len(g.rows) {
		h = len(g.rows) - y
	}

	g.cells = append(g.cells, &Cell{x: x, y: y, w: w, h: h, content: content})
	content.SetParent(g)
}

// Columns sets the column sizes for the grid. The number of columns must match the
// number of columns specified when the grid was created. If the number of columns
// does not match, an error is logged and the grid remains unchanged.
//
// Parameters:
//   - columns: A slice of integers representing the sizes of each column
func (g *Grid) Columns(columns ...int) {
	if len(columns) == len(g.columns) {
		g.columns = columns
	} else {
		g.Log(g, "error", "Cannot change grid size at runtime")
	}
}

func (g *Grid) Rows(rows ...int) {
	if len(rows) == len(g.rows) {
		g.rows = rows
	} else {
		g.Log(g, "error", "Cannot change grid size at runtime")
	}
}

// Layout arranges all child widgets within the grid according to the configured
// row and column sizes, cell positions, and spanning requirements. This method
// performs the complex calculations needed to position widgets in a grid layout.
//
// Layout process:
//  1. Calculate column and row sizes based on flexible/fixed sizing
//  2. Determine grid line positions and account for spacing
//  3. Position each cell's widget according to its span and alignment
//  4. Configure grid line separators for visual rendering
//  5. Recursively layout child containers
//
// Sizing modes:
//   - Fixed sizes: Positive values are treated as absolute sizes
//   - Flexible sizes: Negative values are fractional units of remaining space
//   - Grid lines: Accounted for in spacing calculations when enabled
//
// The layout handles:
//   - Cell spanning across multiple rows/columns
//   - Flexible sizing with fractional space distribution
//   - Grid line rendering and separator configuration
//   - Margin, padding, and border considerations
//   - Recursive layout of child containers
func (g *Grid) Layout() {
	style := g.Style()                // Grid style for margins, padding, borders
	_, _, iw, ih := g.Content()       // Available content size
	cf, rf := 0, 0                    // Total column and row fractions
	lc, lr := 0, 0                    // Last fractional column/row indices
	gx := make([]int, len(g.columns)) // Calculated x positions for each column
	gy := make([]int, len(g.rows))    // Calculated y positions for each row
	aw := make([]int, len(g.columns)) // Preferred width for auto columns
	ah := make([]int, len(g.rows))    // Preferred height for auto rows

	// Reset all separators to show both horizontal and vertical lines
	for i := range g.separators {
		for j := range g.separators[i] {
			g.separators[i][j] = GridB
		}
	}

	// Determine preferred width and height for all rows and columns
	// At this moment, we do not take row spans and column spans into account.
	for _, cell := range g.cells {
		pw, ph := cell.content.Hint()
		if cell.w == 1 && pw > aw[cell.x] {
			aw[cell.x] = pw
		}
		if cell.h == 1 && ph > ah[cell.y] {
			ah[cell.y] = ph
		}
	}

	// Adjust available space to account for grid lines between cells
	if g.lines {
		iw -= len(g.columns) - 1 // Subtract space for vertical grid lines
		ih -= len(g.rows) - 1    // Subtract space for horizontal grid lines
	}

	// ---- Calculate column sizes and positions ----
	// First pass: identify fractional columns and set fixed column widths
	for i := range g.columns {
		if g.columns[i] < 0 {
			cf -= g.columns[i] // Accumulate fractional units (negative values)
			lc = i             // Track last fractional column for remainder handling
		} else if g.columns[i] == 0 {
			g.widths[i] = aw[i]
		} else {
			g.widths[i] = g.columns[i] // Set fixed width
		}
	}

	// Second pass: calculate fractional column widths and positions
	rw := iw // Remaining width after fixed columns
	fc := 0  // Width per fractional unit (only when used)
	if cf > 0 {
		fc = iw / cf
	}
	for i := range g.columns {
		if g.columns[i] < 0 {
			if i == lc {
				g.widths[i] = rw // Last fractional column gets remaining space
				// TODO: Distribute remaining space evenly
			} else {
				g.widths[i] = -g.columns[i] * fc // Calculate fractional width
			}
		}

		// Calculate x position for this column
		if i > 0 {
			gx[i] = gx[i-1] + g.widths[i-1]
			if g.lines {
				gx[i]++ // Account for grid line space
			}
		} else {
			// First column starts after margins, padding, and border
			gx[i] = g.x + style.Margin().Left + style.Padding().Left
			border := style.Border()
			if border != "" && border != "none" {
				gx[i]++ // Account for border line
			}
		}
		rw -= g.widths[i]
	}

	// ---- Calculate row sizes and positions ----
	// First pass: identify fractional rows and set fixed row heights
	fh := ih // height remaining for fractional sizes
	for i := range g.rows {
		if g.rows[i] < 0 {
			rf -= g.rows[i] // Accumulate fractional units (negative values)
			lr = i          // Track last fractional row for remainder handling
		} else if g.rows[i] == 0 {
			g.heights[i] = ah[i]
		} else {
			fh -= g.heights[i]
			g.heights[i] = g.rows[i] // Set fixed height
		}
	}

	// Second pass: calculate fractional row heights and positions
	rh := ih // Remaining height after fixed rows
	fr := 0  // Height per fractional unit (only if used)
	if rf > 0 {
		fr = fh / rf
	}
	for i := range g.rows {
		if g.rows[i] < 0 {
			if i == lr {
				g.heights[i] = rh // Last fractional row gets remaining space
				// TODO: Distribute remaining space evenly
			} else {
				g.heights[i] = -g.rows[i] * fr // Calculate fractional height
			}
		}

		// Calculate y position for this row
		if i > 0 {
			gy[i] = gy[i-1] + g.heights[i-1]
			if g.lines {
				gy[i]++ // Account for grid line space
			}
		} else {
			// First row starts after margins, padding, and border
			gy[i] = g.y + style.Margin().Top + style.Padding().Top
			border := style.Border()
			if border != "" && border != "none" {
				gy[i]++ // Account for border line
			}
		}
		rh -= g.heights[i]
	}

	// ---- Position child widgets and configure grid line separators ----
	for _, cell := range g.cells {
		cw := 0                 // Total width for this cell
		ch := g.heights[cell.y] // Start with height of first row

		// Calculate total width by spanning across columns
		for i := 0; i < cell.w; i++ {
			if cell.x+i >= len(g.columns) {
				break // Don't exceed grid boundaries
			}
			cw += g.widths[cell.x+i]
			if i > 0 && g.lines {
				cw++ // Add space for grid lines between spanned columns
			}

			// Configure grid line separators for this cell's area
			for j := 0; j < cell.h; j++ {
				if cell.x+i < len(g.columns) && cell.y+j < len(g.rows) {
					if i < cell.w-1 && j < cell.h-1 {
						// Interior of spanned cell - no grid lines
						g.separators[cell.y+j][cell.x+i] = 0
					} else if i < cell.w-1 {
						// Right edge of spanned cell - only horizontal line
						g.separators[cell.y+j][cell.x+i] = GridH
					} else if j < cell.h-1 {
						// Bottom edge of spanned cell - only vertical line
						g.separators[cell.y+j][cell.x+i] = GridV
					}
				}
			}
		}

		// Calculate total height by spanning across rows
		for i := 1; i < cell.h; i++ {
			ch += g.heights[cell.y+i]
			if g.lines {
				ch++ // Add space for grid lines between spanned rows
			}
		}

		// Set the final bounds for the cell's widget
		cell.content.SetBounds(gx[cell.x], gy[cell.y], cw, ch)

		// Recursively layout child containers or refresh leaf widgets
		if inner, ok := cell.content.(Container); ok {
			inner.Layout()
		}
	}
}

func (g *Grid) Render(r *Renderer) {
	// Use the rendering of the component first
	g.Component.Render(r)

	// Render the children
	for _, cell := range g.cells {
		cell.content.Render(r)
	}

	// If no grid lines should be rendered, we are done
	if !g.lines {
		return
	}

	// Render the grid lines
	style := g.Style()
	border := style.Border()
	if border == "" || border == "none" {
		return
	}
	b := r.theme.Border(border)
	if b == nil {
		return
	}
	_, _, iw, ih := g.Content()

	// draw top and bottom Ts
	cx := g.x + style.Margin().Left + style.Padding().Left
	by := g.y + style.Margin().Top + style.Padding().Top + style.Padding().Bottom + 1 + ih
	for i := range len(g.columns) - 1 {
		cx += g.widths[i] + 1
		if g.separators[0][i]&GridV > 0 {
			r.screen.Put(cx, g.y+style.Margin().Top, b.TopT)
			r.Repeat(cx, g.y+style.Margin().Top+1, 0, 1, style.Padding().Top, b.InnerV)
		}
		if g.separators[len(g.rows)-1][i]&GridV > 0 {
			r.screen.Put(cx, by, b.BottomT)
			r.Repeat(cx, by-1, 0, -1, style.Padding().Bottom, b.InnerV)
		}
	}

	// draw left and right Ts
	cy := g.y + style.Margin().Top + style.Padding().Top
	rx := g.x + style.Margin().Left + style.Padding().Left + style.Padding().Right + 1 + iw
	for i := range len(g.rows) - 1 {
		cy += g.heights[i] + 1
		if g.separators[i][0]&GridH > 0 {
			r.screen.Put(g.x+style.Margin().Left, cy, b.LeftT)
			r.Repeat(g.x+style.Margin().Left+1, cy, 1, 0, style.Padding().Left, b.InnerH)
		}
		if g.separators[i][len(g.columns)-1]&GridH > 0 {
			r.screen.Put(rx, cy, b.RightT)
			r.Repeat(rx-1, cy, -1, 0, style.Padding().Right, b.InnerH)
		}
	}

	// draw inner grid lines
	cy = g.y + style.Margin().Top + style.Padding().Top
	for row := range len(g.rows) {
		cx = g.x + style.Margin().Left + style.Padding().Left + g.widths[0] + 1
		cy += g.heights[row] + 1
		for c := range len(g.columns) {
			connector := 0
			if row < len(g.rows)-1 && g.separators[row][c]&GridH > 0 {
				r.Repeat(cx-1, cy, -1, 0, g.widths[c], b.InnerH)
				connector |= 8
			}
			if c < len(g.columns)-1 && g.separators[row][c]&GridV > 0 {
				r.Repeat(cx, cy-1, 0, -1, g.heights[row], b.InnerV)
				connector |= 1
			}
			if row < len(g.rows)-1 && c < len(g.columns)-1 {
				if g.separators[row+1][c]&GridV > 0 {
					connector |= 4
				}
				if g.separators[row][c+1]&GridH > 0 {
					connector |= 2
				}
				switch connector {
				case 5:
					r.screen.Put(cx, cy, b.InnerV)
				case 7:
					r.screen.Put(cx, cy, b.InnerLeftT)
				case 10:
					r.screen.Put(cx, cy, b.InnerH)
				case 11:
					r.screen.Put(cx, cy, b.InnerBottomT)
				case 13:
					r.screen.Put(cx, cy, b.InnerRightT)
				case 14:
					r.screen.Put(cx, cy, b.InnerTopT)
				case 15:
					r.screen.Put(cx, cy, b.InnerX)
				}
			}
			if c < len(g.columns)-1 {
				cx += g.widths[c+1] + 1
			}
		}
	}
}

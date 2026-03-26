package zeichenwerk

import "github.com/gdamore/tcell/v3"

// Table represents a widget that displays tabular data with support for scrolling,
// keyboard navigation, and customizable styling. It uses a TableProvider interface
// to supply data, making it flexible for different data sources.
type Table struct {
	Component
	provider         TableProvider // Data source implementing TableProvider interface
	tableWidth       int           // Total table width across all columns (includes separators)
	row, column      int           // Current highlight position (row index, column unused)
	offsetX, offsetY int           // Scroll offsets for horizontal and vertical scrolling
	grid             *Border       // Border characters for grid lines and intersections
	inner, outer     bool          // Flags to control inner column/row separators and outer borders
}

// NewTable creates a new table widget with the specified ID and data provider.
// The table is initialized with default grid styling and both inner and outer
// borders enabled. The table is focusable by default to support keyboard navigation.
//
// Parameters:
//   - id: Unique identifier for the table widget
//   - provider: TableProvider implementation that supplies the table data
//
// Returns:
//   - *Table: Configured table widget ready for use
//
// Example:
//
//	headers := []string{"Name", "Age", "City"}
//	data := [][]string{{"John", "25", "NYC"}, {"Jane", "30", "LA"}}
//	provider := NewArrayTableProvider(headers, data)
//	table := NewTable("my-table", provider)
func NewTable(id string, provider TableProvider) *Table {
	table := &Table{
		Component: Component{id: id},
		grid:      &Border{InnerH: "-", InnerV: "|"},
		inner:     true,
		outer:     true,
	}
	table.SetFlag("focusable", true)
	table.Set(provider)
	OnKey(table, table.handleKey)
	return table
}

// Refresh updates the table.
func (t *Table) Refresh() {
	Redraw(t)
}

// Set updates the table's data provider and recalculates the total table width.
// This method can be used to dynamically change the table's data source.
// The table width is calculated as the sum of all column widths plus separators.
//
// Parameters:
//   - provider: New TableProvider implementation to use for data
//
// Note: This method resets the table width calculation, so it should be called
// whenever the column structure changes.
func (t *Table) Set(provider TableProvider) {
	t.provider = provider
	t.tableWidth = 0
	columns := provider.Columns()
	for _, column := range columns {
		t.tableWidth += column.Width
	}
	t.tableWidth += len(columns) - 1
}

// Hint returns the preferred size for the table widget.
// The width includes all column widths plus separators, and the height
// equals the number of data rows (excluding the header).
//
// Returns:
//   - int: Preferred width (sum of column widths + separators)
//   - int: Preferred height (number of data rows)
//
// Note: The actual rendered size may differ based on container constraints.
// The hint helps layout containers determine optimal sizing.
func (t *Table) Hint() (int, int) {
	if t.hwidth != 0 || t.hheight != 0 {
		return t.hwidth, t.hheight
	}

	w := 0
	h := t.provider.Length()
	for i, column := range t.provider.Columns() {
		if i > 0 {
			w++
		}
		w += column.Width
	}
	return w, h
}

// Handle processes keyboard events for table navigation.
// Supports comprehensive keyboard navigation including row selection, scrolling,
// and quick navigation shortcuts.
//
// Supported key bindings:
//   - Down Arrow: Move to next row
//   - Up Arrow: Move to previous row
//   - Left Arrow: Scroll left one character
//   - Right Arrow: Scroll right one character
//   - Home: Jump to first row
//   - End: Jump to last row
//   - Page Up: Move up by visible page height
//   - Page Down: Move down by visible page height
//   - Ctrl+Left: Scroll left by one full column width
//   - Ctrl+Right: Scroll right by one full column width
//   - Ctrl+Home: Jump to first row and reset horizontal scroll
//   - Ctrl+End: Jump to last row and reset horizontal scroll
//   - Ctrl+Up: Jump to first row (alternative)
//   - Ctrl+Down: Jump to last row (alternative)
//   - Tab: Move to next column (horizontal navigation)
//   - Shift+Tab: Move to previous column (horizontal navigation)
//   - Enter: Emit 'activate' event with current row data
//   - Space: Emit 'select' event with current row data
//
// Parameters:
//   - evt: tcell.Event to process
//
// Returns:
//   - bool: true if the event was handled, false otherwise
//
// The method automatically adjusts the viewport when navigating to ensure
// the selected row remains visible.
func (t *Table) handleKey(_ Widget, event *tcell.EventKey) bool {
	_, _, _, h := t.Content()
	pageSize := max(1, h-3) // Account for header and borders

	switch event.Key() {
	case tcell.KeyDown:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			// Ctrl+Down: Jump to last row
			t.row = max(0, t.provider.Length()-1)
			t.adjust()
		} else {
			// Regular down: Move to next row
			if t.row < t.provider.Length()-1 {
				t.row++
				t.adjust()
			}
		}
		return true
	case tcell.KeyUp:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			// Ctrl+Up: Jump to first row
			t.row = 0
			t.adjust()
		} else {
			// Regular up: Move to previous row
			if t.row > 0 {
				t.row--
				t.adjust()
			}
		}
		return true
	case tcell.KeyLeft:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			// Ctrl+Left: Scroll by full column width
			t.scrollByColumn(-1)
		} else if event.Modifiers()&tcell.ModShift != 0 {
			// Shift+Left: Move to previous column
			t.moveToColumn(-1)
		} else {
			// Regular left: Scroll by single character
			if t.offsetX > 0 {
				t.offsetX--
				t.Refresh()
			}
		}
		return true
	case tcell.KeyRight:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			// Ctrl+Right: Scroll by full column width
			t.scrollByColumn(1)
		} else if event.Modifiers()&tcell.ModShift != 0 {
			// Shift+Right: Move to next column
			t.moveToColumn(1)
		} else {
			// Regular right: Scroll by single character
			if t.offsetX+t.width < t.tableWidth {
				t.offsetX++
				t.Refresh()
			}
		}
		return true
	case tcell.KeyHome:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			// Ctrl+Home: First row and reset horizontal scroll
			t.row = 0
			t.offsetX = 0
			t.adjust()
		} else {
			// Home: Jump to first row
			t.row = 0
			t.adjust()
		}
		return true
	case tcell.KeyEnd:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			// Ctrl+End: Last row and reset horizontal scroll
			t.row = max(0, t.provider.Length()-1)
			t.offsetX = 0
			t.adjust()
		} else {
			// End: Jump to last row
			t.row = max(0, t.provider.Length()-1)
			t.adjust()
		}
		return true
	case tcell.KeyPgUp:
		// Page Up: Move up by page size
		t.row = max(0, t.row-pageSize)
		t.adjust()
		return true
	case tcell.KeyPgDn:
		// Page Down: Move down by page size
		t.row = min(t.provider.Length()-1, t.row+pageSize)
		t.adjust()
		return true
	case tcell.KeyEnter:
		// Enter: Emit 'activate' event with current row data
		if t.provider.Length() > 0 && t.row >= 0 && t.row < t.provider.Length() {
			rowData := t.getCurrentRowData()
			t.Dispatch(t, "activate", t.row, rowData)
		}
		return true
	case tcell.KeyRune:
		if event.Str() == " " {
			// Space: Emit 'select' event with current row data
			if t.provider.Length() > 0 && t.row >= 0 && t.row < t.provider.Length() {
				rowData := t.getCurrentRowData()
				t.Dispatch(t, "select", t.row, rowData)
			}
			return true
		}
		return false
	default:
		return false
	}
}

// adjust ensures the currently selected row remains visible by updating
// the vertical scroll offset when necessary. This method is called automatically
// during navigation to maintain proper viewport positioning.
//
// The adjustment logic:
//   - If all rows fit in the visible area, no adjustment is needed
//   - If the selected row is above the visible area, scroll up to show it
//   - If the selected row is below the visible area, scroll down to show it
//
// The method accounts for the header row (which takes 2 lines with separator)
// when calculating visible content area.
func (t *Table) adjust() {
	// Get actual content height
	_, _, _, h := t.Content()

	// We do not need to adjust anything, if all rows fit
	if t.provider.Length() < h-2 {
		return
	}
	if t.row < t.offsetY {
		t.offsetY = t.row
	}
	if t.row > t.offsetY+h-3 {
		t.offsetY = t.row - h + 3
	}
	t.Refresh()
}

// scrollByColumn scrolls horizontally by full column widths rather than single characters.
// This provides more efficient navigation when dealing with wide tables.
//
// Parameters:
//   - direction: -1 to scroll left, +1 to scroll right
func (t *Table) scrollByColumn(direction int) {
	columns := t.provider.Columns()
	if len(columns) == 0 {
		return
	}

	if direction < 0 {
		// Scroll left by one column
		if t.offsetX > 0 {
			// Find the column boundary to scroll to
			currentPos := 0
			for _, column := range columns {
				if currentPos+column.Width+1 > t.offsetX {
					t.offsetX = max(0, currentPos)
					break
				}
				currentPos += column.Width + 1
			}
			t.Refresh()
		}
	} else {
		// Scroll right by one column
		if t.offsetX+t.width < t.tableWidth {
			// Find the next column boundary
			currentPos := 0
			for _, column := range columns {
				if currentPos > t.offsetX {
					t.offsetX = min(t.tableWidth-t.width, currentPos)
					break
				}
				currentPos += column.Width + 1
			}
			t.Refresh()
		}
	}
}

// moveToColumn implements horizontal navigation by moving the view to show specific columns.
// This is used for Shift-Arrow horizontal navigation to jump between column boundaries.
//
// Parameters:
//   - direction: -1 to move to previous column, +1 to move to next column
func (t *Table) moveToColumn(direction int) {
	columns := t.provider.Columns()
	if len(columns) == 0 {
		return
	}

	// Calculate current visible column range
	visibleStart := t.offsetX
	visibleEnd := t.offsetX + t.width

	// Find current column index based on offset
	currentPos := 0
	currentColumn := 0
	for i, column := range columns {
		if currentPos <= visibleStart && currentPos+column.Width > visibleStart {
			currentColumn = i
			break
		}
		currentPos += column.Width + 1
	}

	// Calculate target column
	targetColumn := currentColumn + direction
	if targetColumn < 0 {
		targetColumn = 0
	} else if targetColumn >= len(columns) {
		targetColumn = len(columns) - 1
	}

	// Calculate position for target column
	targetPos := 0
	for i := 0; i < targetColumn; i++ {
		targetPos += columns[i].Width + 1
	}

	// Adjust offset to show the target column
	if targetPos < visibleStart {
		// Target column is to the left, scroll left
		t.offsetX = targetPos
	} else if targetPos+columns[targetColumn].Width > visibleEnd {
		// Target column is to the right, scroll right
		t.offsetX = min(t.tableWidth-t.width, targetPos+columns[targetColumn].Width-t.width)
	}

	t.Refresh()
}

// getCurrentRowData returns all column values for the currently selected row
// as a slice of strings. This is useful for event handlers that need access
// to the complete row data.
//
// Returns:
//   - []string: Slice containing all column values for the current row
func (t *Table) getCurrentRowData() []string {
	if t.provider == nil || t.row < 0 || t.row >= t.provider.Length() {
		return nil
	}

	columns := t.provider.Columns()
	rowData := make([]string, len(columns))
	for i := range columns {
		rowData[i] = t.provider.Str(t.row, i)
	}
	return rowData
}

// GetSelectedRow returns the index of the currently selected row.
//
// Returns:
//   - int: Zero-based index of the selected row, or -1 if no valid selection
func (t *Table) GetSelectedRow() int {
	if t.provider == nil || t.row < 0 || t.row >= t.provider.Length() {
		return -1
	}
	return t.row
}

// SetSelectedRow programmatically sets the selected row and adjusts the viewport.
// This is useful for external navigation or search functionality.
//
// Parameters:
//   - row: Zero-based row index to select
//
// Returns:
//   - bool: true if the row was successfully selected, false if invalid
func (t *Table) SetSelectedRow(row int) bool {
	if t.provider == nil || row < 0 || row >= t.provider.Length() {
		return false
	}
	t.row = row
	t.adjust()
	return true
}

// GetScrollOffset returns the current horizontal and vertical scroll offsets.
//
// Returns:
//   - int: Horizontal scroll offset in characters
//   - int: Vertical scroll offset in rows
func (t *Table) GetScrollOffset() (int, int) {
	return t.offsetX, t.offsetY
}

// SetScrollOffset programmatically sets the scroll offsets.
// Useful for implementing custom scrolling behavior or restoring scroll position.
//
// Parameters:
//   - offsetX: Horizontal scroll offset in characters
//   - offsetY: Vertical scroll offset in rows
func (t *Table) SetScrollOffset(offsetX, offsetY int) {
	t.offsetX = max(0, min(offsetX, max(0, t.tableWidth-t.width)))
	t.offsetY = max(0, min(offsetY, max(0, t.provider.Length()-1)))
	t.Refresh()
}

func (t *Table) Render(r *Renderer) {
	t.Component.Render(r)

	x, y, w, h := t.Content()
	state := t.State()
	if state != "" {
		state = ":" + state
	}
	headerStyle := t.Style("header" + state)
	gridStyle := t.Style("grid" + state)
	border := gridStyle.Border()
	if border != "" && border != "none" {
		t.grid = r.theme.Border(border)
	}
	t.renderTableHeader(r, x, y, w, h, headerStyle, gridStyle)
	t.renderTableContent(r, x, y+2, w, h-2, gridStyle)
}

func (t *Table) renderTableHeader(r *Renderer, x, y, w, h int, headerStyle, gridStyle *Style) {
	rx := 0 // current render position
	rw := w // remaining width
	columns := t.provider.Columns()
	for i, column := range columns {
		// If there is no room remaing, we stop rendering the header
		if rw <= 0 {
			break
		}
		// Check, if the column is visible on the left
		if rx+column.Width < t.offsetX {
			rx = rx + column.Width + 1
			continue
		}
		// Do we need to render it partially?
		r.Set(headerStyle.Foreground(), headerStyle.Background(), headerStyle.Font())

		// cell/column x-position
		cx := x - t.offsetX + rx
		if rx < t.offsetX {
			start := t.offsetX - rx
			rcw := min(column.Width-start, rw)
			runes := []rune(column.Header)
			if start < len(runes) {
				r.Text(cx+start, y, string(runes[start:]), rcw)
			}
			r.Set(gridStyle.Foreground(), gridStyle.Background(), gridStyle.Font())
			r.Repeat(cx+start, y+1, 1, 0, rcw, t.grid.InnerH)
			rw += start
		} else {
			// We render either to complete column or use the remaining width
			cw := min(rw, column.Width)
			r.Text(cx, y, column.Header, cw)
			r.Set(gridStyle.Foreground(), gridStyle.Background(), gridStyle.Font())
			r.Repeat(cx, y+1, 1, 0, cw, t.grid.InnerH)
		}
		// Do we need to render the column separator?
		if i < len(columns)-1 && rw > column.Width {
			r.screen.Put(cx+column.Width, y+1, t.grid.InnerTopT)
			// Bottom T is rendered with the header
			if t.outer {
				r.screen.Put(cx+column.Width, y+h, t.grid.BottomT)
			}
		}
		rx = rx + column.Width + 1
		rw = rw - column.Width - 1
	}

	// Draw the remaining header line
	r.Set(gridStyle.Foreground(), gridStyle.Background(), gridStyle.Font())
	if rx-t.offsetX < w {
		r.Repeat(x-t.offsetX+rx-1, y+1, 1, 0, w+t.offsetX-rx+1, t.grid.InnerH)
	}
	// Draw the left and right T connectors
	if t.outer {
		r.screen.Put(x-1, y+1, t.grid.LeftT)
		r.screen.Put(x+w, y+1, t.grid.RightT)
	}
}

func (t *Table) renderTableContent(r *Renderer, x, y, w, h int, gridStyle *Style) {
	row := t.offsetY
	columns := t.provider.Columns()
	// Traverse all visible rows
	for row < t.provider.Length() && row-t.offsetY < h {
		// Determine row style
		var style *Style
		if row == t.row && t.Flag("focused") {
			style = t.Style("highlight:focused")
		} else if row == t.row {
			style = t.Style("highlight")
		} else {
			style = t.Style()
		}

		rx := 0 // render x position
		rw := w // remaining width
		for i, column := range columns {
			if rw <= 0 {
				break
			}
			if rx+column.Width < t.offsetX {
				rx = rx + column.Width + 1
				continue
			}
			r.Set(style.Foreground(), style.Background(), style.Font())
			cx := x - t.offsetX + rx
			cy := y - t.offsetY + row
			if rx < t.offsetX {
				start := t.offsetX - rx
				rcw := min(column.Width-start, rw)
				runes := []rune(t.provider.Str(row, i))
				if start < len(runes) {
					r.Text(cx+start, cy, string(runes[start:]), rcw)
				} else {
					r.Repeat(cx+start, cy, 1, 0, rcw, " ")
				}
				rw += start
			} else {
				cw := min(rw, column.Width)
				r.Text(cx, cy, t.provider.Str(row, i), cw)
			}
			if i < len(columns)-1 && t.inner && rw > column.Width {
				if row != t.row {
					r.Set(gridStyle.Foreground(), gridStyle.Background(), gridStyle.Font())
				}
				r.screen.Put(cx+column.Width, cy, t.grid.InnerV)
			}
			rx = rx + column.Width + 1
			rw = rw - column.Width - 1
		}
		row++
	}
}

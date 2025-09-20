package zeichenwerk

import "github.com/gdamore/tcell/v2"

// Table represents a widget that displays tabular data with support for scrolling,
// keyboard navigation, and customizable styling. It uses a TableProvider interface
// to supply data, making it flexible for different data sources.
//
// The table widget supports:
//   - Keyboard navigation (arrow keys for movement)
//   - Horizontal and vertical scrolling for large datasets
//   - Row highlighting with focus states
//   - Configurable grid borders (inner and outer)
//   - Dynamic column sizing based on content
//
// Table styling can be customized using the following style classes:
//   - "" (default): Normal row styling
//   - "highlight": Row highlighting when not focused
//   - "highlight:focus": Row highlighting when table has focus
//   - "header": Header row styling
//   - "grid": Grid border and separator styling
//
// Usage Example:
//
//	// Create sample data
//	headers := []string{"Name", "Age", "City", "Country"}
//	data := [][]string{
//		{"John Doe", "25", "New York", "USA"},
//		{"Jane Smith", "30", "Los Angeles", "USA"},
//		{"Bob Johnson", "35", "Chicago", "USA"},
//		{"Alice Brown", "28", "Houston", "USA"},
//	}
//
//	// Create table provider and widget
//	provider := NewArrayTableProvider(headers, data)
//	table := NewTable("employee-table", provider)
//
//	// Use with builder pattern
//	ui := NewBuilder(DefaultTheme()).
//		Box("table-container", "Employee List").
//		Add(table).
//		Build()
//
//	// Or update data dynamically
//	newData := [][]string{{"Sam Wilson", "32", "Seattle", "USA"}}
//	newProvider := NewArrayTableProvider(headers, newData)
//	table.Set(newProvider)
//
//	// Handle table events
//	table.On("activate", func(widget Widget, event string, data ...any) bool {
//		rowIndex := data[0].(int)
//		rowData := data[1].([]string)
//		fmt.Printf("Activated row %d: %v\n", rowIndex, rowData)
//		return true
//	})
//
//	table.On("select", func(widget Widget, event string, data ...any) bool {
//		rowIndex := data[0].(int)
//		rowData := data[1].([]string)
//		fmt.Printf("Selected row %d: %v\n", rowIndex, rowData)
//		return true
//	})
//
// For large datasets or dynamic data sources, implement the TableProvider
// interface directly instead of using ArrayTableProvider:
//
//	type DatabaseTableProvider struct {
//		// your database connection, query, etc.
//	}
//
//	func (d *DatabaseTableProvider) Columns() []TableColumn { /* ... */ }
//	func (d *DatabaseTableProvider) Length() int { /* ... */ }
//	func (d *DatabaseTableProvider) Str(row, col int) string { /* ... */ }
type Table struct {
	BaseWidget
	provider         TableProvider // Data source implementing TableProvider interface
	tableWidth       int           // Total table width across all columns (includes separators)
	row, column      int           // Current highlight position (row index, column unused)
	offsetX, offsetY int           // Scroll offsets for horizontal and vertical scrolling
	grid             BorderStyle   // Border characters for grid lines and intersections
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
//   headers := []string{"Name", "Age", "City"}
//   data := [][]string{{"John", "25", "NYC"}, {"Jane", "30", "LA"}}
//   provider := NewArrayTableProvider(headers, data)
//   table := NewTable("my-table", provider)
func NewTable(id string, provider TableProvider) *Table {
	table := &Table{
		BaseWidget: BaseWidget{id: id, focusable: true},
		grid:       BorderStyle{InnerH: ' ', InnerV: '-'},
		inner:      true,
		outer:      true,
	}
	table.Set(provider)
	return table
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
	w := 0
	h := t.provider.Length()
	for i, column := range t.provider.Columns() {
		if i > 1 {
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
func (t *Table) Handle(evt tcell.Event) bool {
	event, ok := evt.(*tcell.EventKey)
	if !ok {
		return false
	}

	_, h := t.Size()
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
	case tcell.KeyTab:
		if event.Modifiers()&tcell.ModShift != 0 {
			// Shift+Tab: Move to previous column
			t.moveToColumn(-1)
		} else {
			// Tab: Move to next column
			t.moveToColumn(1)
		}
		return true
	case tcell.KeyEnter:
		// Enter: Emit 'activate' event with current row data
		if t.provider.Length() > 0 && t.row >= 0 && t.row < t.provider.Length() {
			rowData := t.getCurrentRowData()
			t.Emit("activate", t.row, rowData)
		}
		return true
	case tcell.KeyRune:
		if event.Rune() == ' ' {
			// Space: Emit 'select' event with current row data
			if t.provider.Length() > 0 && t.row >= 0 && t.row < t.provider.Length() {
				rowData := t.getCurrentRowData()
				t.Emit("select", t.row, rowData)
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
	_, h := t.Size()

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
// This is used for Tab/Shift+Tab navigation to jump between column boundaries.
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

package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestTableNavigation(t *testing.T) {
	// Create test data
	headers := []string{"Name", "Age", "City", "Country"}
	data := [][]string{
		{"John Doe", "25", "New York", "USA"},
		{"Jane Smith", "30", "Los Angeles", "USA"},
		{"Bob Johnson", "35", "Chicago", "USA"},
		{"Alice Brown", "28", "Houston", "USA"},
		{"Charlie Wilson", "45", "Seattle", "USA"},
		{"Diana Davis", "32", "Miami", "USA"},
		{"Edward Miller", "29", "Denver", "USA"},
		{"Fiona Garcia", "38", "Boston", "USA"},
		{"George Martinez", "41", "Phoenix", "USA"},
		{"Helen Rodriguez", "27", "Portland", "USA"},
	}

	provider := NewArrayTableProvider(headers, data)
	table := NewTable("test-table", provider)

	// Set a reasonable size for testing
	table.SetBounds(0, 0, 20, 20)

	// Test initial state
	if table.GetSelectedRow() != 0 {
		t.Errorf("Expected initial selected row to be 0, got %d", table.GetSelectedRow())
	}

	// Test basic navigation
	testKeyNavigation(t, table, tcell.KeyDown, 1, "Down arrow navigation")
	testKeyNavigation(t, table, tcell.KeyUp, 0, "Up arrow navigation")

	// Test Home/End keys
	testKeyNavigation(t, table, tcell.KeyEnd, len(data)-1, "End key navigation")
	testKeyNavigation(t, table, tcell.KeyHome, 0, "Home key navigation")

	// Test Page Down/Up
	testKeyNavigation(t, table, tcell.KeyPgDn, min(len(data)-1, 17), "Page Down navigation") // 20-3 = 17 visible rows
	testKeyNavigation(t, table, tcell.KeyPgUp, max(0, table.GetSelectedRow()-17), "Page Up navigation")

	// Test Ctrl+Down/Up
	testKeyNavigationWithMod(t, table, tcell.KeyDown, tcell.ModCtrl, len(data)-1, "Ctrl+Down navigation")
	testKeyNavigationWithMod(t, table, tcell.KeyUp, tcell.ModCtrl, 0, "Ctrl+Up navigation")

	// Test programmatic navigation
	if !table.SetSelectedRow(5) {
		t.Error("SetSelectedRow should succeed for valid row")
	}
	if table.GetSelectedRow() != 5 {
		t.Errorf("Expected selected row 5, got %d", table.GetSelectedRow())
	}

	// Test invalid row selection
	if table.SetSelectedRow(-1) {
		t.Error("SetSelectedRow should fail for negative row")
	}
	if table.SetSelectedRow(len(data)) {
		t.Error("SetSelectedRow should fail for row beyond data length")
	}

	// Test scroll offset methods
	table.SetScrollOffset(10, 2)
	offsetX, offsetY := table.GetScrollOffset()
	if offsetX != 10 || offsetY != 2 {
		t.Errorf("Expected scroll offset (10, 2), got (%d, %d)", offsetX, offsetY)
	}
}

func TestTableEvents(t *testing.T) {
	headers := []string{"Name", "Age"}
	data := [][]string{
		{"John", "25"},
		{"Jane", "30"},
	}

	provider := NewArrayTableProvider(headers, data)
	table := NewTable("test-table", provider)
	table.SetBounds(0, 0, 40, 10)

	// Test event emission
	var lastEvent string
	var lastRowIndex int
	var lastRowData []string

	table.On("activate", func(widget Widget, event string, data ...any) bool {
		lastEvent = event
		lastRowIndex = data[0].(int)
		lastRowData = data[1].([]string)
		return true
	})

	table.On("select", func(widget Widget, event string, data ...any) bool {
		lastEvent = event
		lastRowIndex = data[0].(int)
		lastRowData = data[1].([]string)
		return true
	})

	// Test Enter key (activate event)
	enterEvent := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	table.Handle(enterEvent)

	if lastEvent != "activate" {
		t.Errorf("Expected 'activate' event, got '%s'", lastEvent)
	}
	if lastRowIndex != 0 {
		t.Errorf("Expected row index 0, got %d", lastRowIndex)
	}
	if len(lastRowData) != 2 || lastRowData[0] != "John" || lastRowData[1] != "25" {
		t.Errorf("Expected row data ['John', '25'], got %v", lastRowData)
	}

	// Test Space key (select event)
	spaceEvent := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
	table.Handle(spaceEvent)

	if lastEvent != "select" {
		t.Errorf("Expected 'select' event, got '%s'", lastEvent)
	}
}

func TestTableProvider(t *testing.T) {
	headers := []string{"Col1", "Col2", "Col3"}
	data := [][]string{
		{"A", "B", "C"},
		{"Very Long Text", "Short", "Medium Text"},
		{"X", "Y", "Z"},
	}

	provider := NewArrayTableProvider(headers, data)

	// Test column width calculation
	columns := provider.Columns()
	if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}

	// First column should be width of "Very Long Text" (14 characters)
	if columns[0].Width != 14 {
		t.Errorf("Expected column 0 width 14, got %d", columns[0].Width)
	}

	// Test data access
	if provider.Str(1, 0) != "Very Long Text" {
		t.Errorf("Expected 'Very Long Text', got '%s'", provider.Str(1, 0))
	}

	if provider.Length() != 3 {
		t.Errorf("Expected length 3, got %d", provider.Length())
	}
}

func TestTableHorizontalNavigation(t *testing.T) {
	headers := []string{"Col1", "Col2", "Col3", "Col4", "Col5"}
	data := [][]string{
		{"A", "B", "C", "D", "E"},
		{"F", "G", "H", "I", "J"},
	}

	provider := NewArrayTableProvider(headers, data)
	table := NewTable("test-table", provider)
	table.SetBounds(0, 0, 20, 10) // Narrow width to test horizontal scrolling

	// Test horizontal scrolling
	rightEvent := tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)

	// Initial offset should be 0
	offsetX, _ := table.GetScrollOffset()
	if offsetX != 0 {
		t.Errorf("Expected initial horizontal offset 0, got %d", offsetX)
	}

	// Test right scrolling
	table.Handle(rightEvent)
	offsetX, _ = table.GetScrollOffset()
	if offsetX == 0 {
		t.Error("Expected horizontal scroll to increase after right arrow")
	}

	// Test Tab navigation
	tabEvent := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	table.Handle(tabEvent)

	// Test Shift+Tab navigation
	shiftTabEvent := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModShift)
	table.Handle(shiftTabEvent)

	// Test Ctrl+Left/Right for column scrolling
	ctrlLeftEvent := tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl)
	ctrlRightEvent := tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl)

	table.Handle(ctrlRightEvent)
	table.Handle(ctrlLeftEvent)
}

// Helper function for testing key navigation
func testKeyNavigation(t *testing.T, table *Table, key tcell.Key, expectedRow int, description string) {
	event := tcell.NewEventKey(key, 0, tcell.ModNone)
	table.Handle(event)

	if table.GetSelectedRow() != expectedRow {
		t.Errorf("%s: Expected row %d, got %d", description, expectedRow, table.GetSelectedRow())
	}
}

// Helper function for testing key navigation with modifiers
func testKeyNavigationWithMod(t *testing.T, table *Table, key tcell.Key, mod tcell.ModMask, expectedRow int, description string) {
	event := tcell.NewEventKey(key, 0, mod)
	table.Handle(event)

	if table.GetSelectedRow() != expectedRow {
		t.Errorf("%s: Expected row %d, got %d", description, expectedRow, table.GetSelectedRow())
	}
}


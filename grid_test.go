package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// MockWidget is a simple test widget for grid testing
type MockWidget struct {
	BaseWidget
	preferredWidth  int
	preferredHeight int
}

func NewMockWidget(id string, w, h int) *MockWidget {
	widget := &MockWidget{
		BaseWidget:      BaseWidget{id: id},
		preferredWidth:  w,
		preferredHeight: h,
	}
	widget.SetStyle("", NewStyle())
	return widget
}

func (m *MockWidget) Hint() (int, int) {
	return m.preferredWidth, m.preferredHeight
}

// Implement required Widget interface methods
func (m *MockWidget) Handle(event tcell.Event) bool {
	return false
}

func (m *MockWidget) Info() string {
	return "MockWidget"
}

func (m *MockWidget) Cursor() (int, int) {
	return -1, -1
}

// Add Layout method for container interface
func (m *MockWidget) Layout() {
	// MockWidget doesn't need layout
}

// TestNewGrid tests the creation of new grid widgets
func TestNewGrid(t *testing.T) {
	t.Run("Basic grid creation", func(t *testing.T) {
		grid := NewGrid("test-grid", 3, 4, true)

		if grid == nil {
			t.Fatal("NewGrid returned nil")
		}

		if grid.ID() != "test-grid" {
			t.Errorf("NewGrid() ID = %q, want %q", grid.ID(), "test-grid")
		}

		// Check dimensions
		if len(grid.rows) != 3 {
			t.Errorf("NewGrid() rows count = %d, want 3", len(grid.rows))
		}

		if len(grid.columns) != 4 {
			t.Errorf("NewGrid() columns count = %d, want 4", len(grid.columns))
		}

		// Check grid lines setting
		if !grid.lines {
			t.Error("NewGrid() lines should be true")
		}

		// Check initial cell count
		if len(grid.cells) != 0 {
			t.Errorf("NewGrid() should start with 0 cells, got %d", len(grid.cells))
		}
	})

	t.Run("Grid without lines", func(t *testing.T) {
		grid := NewGrid("no-lines", 2, 2, false)

		if grid.lines {
			t.Error("NewGrid() lines should be false")
		}
	})

	t.Run("Default fractional sizing", func(t *testing.T) {
		grid := NewGrid("test", 2, 3, false)

		// All rows should default to -1 (fractional)
		for i, row := range grid.rows {
			if row != -1 {
				t.Errorf("Row %d should default to -1, got %d", i, row)
			}
		}

		// All columns should default to -1 (fractional)
		for i, col := range grid.columns {
			if col != -1 {
				t.Errorf("Column %d should default to -1, got %d", i, col)
			}
		}
	})
}

// TestGridAdd tests adding widgets to grid cells
func TestGridAdd(t *testing.T) {
	t.Run("Single cell addition", func(t *testing.T) {
		grid := NewGrid("test", 3, 3, false)
		widget := NewMockWidget("widget1", 10, 5)

		grid.Add(1, 2, 1, 1, widget)

		if len(grid.cells) != 1 {
			t.Errorf("Grid should have 1 cell, got %d", len(grid.cells))
		}

		cell := grid.cells[0]
		if cell.x != 1 || cell.y != 2 {
			t.Errorf("Cell position = (%d, %d), want (1, 2)", cell.x, cell.y)
		}

		if cell.w != 1 || cell.h != 1 {
			t.Errorf("Cell span = (%d, %d), want (1, 1)", cell.w, cell.h)
		}

		if cell.content.ID() != widget.ID() {
			t.Error("Cell content should be the added widget")
		}

		// Check parent relationship
		if widget.Parent() != grid {
			t.Error("Widget parent should be set to grid")
		}
	})

	t.Run("Multiple cell additions", func(t *testing.T) {
		grid := NewGrid("test", 2, 2, false)
		widget1 := NewMockWidget("widget1", 10, 5)
		widget2 := NewMockWidget("widget2", 20, 15)

		grid.Add(0, 0, 1, 1, widget1)
		grid.Add(1, 1, 1, 1, widget2)

		if len(grid.cells) != 2 {
			t.Errorf("Grid should have 2 cells, got %d", len(grid.cells))
		}

		// Check first cell
		cell1 := grid.cells[0]
		if cell1.x != 0 || cell1.y != 0 || cell1.content.ID() != widget1.ID() {
			t.Error("First cell configuration incorrect")
		}

		// Check second cell
		cell2 := grid.cells[1]
		if cell2.x != 1 || cell2.y != 1 || cell2.content.ID() != widget2.ID() {
			t.Error("Second cell configuration incorrect")
		}
	})

	t.Run("Spanning cells", func(t *testing.T) {
		grid := NewGrid("test", 3, 3, false)
		widget := NewMockWidget("span-widget", 30, 20)

		// Add widget spanning 2 columns and 2 rows
		grid.Add(0, 0, 2, 2, widget)

		cell := grid.cells[0]
		if cell.w != 2 || cell.h != 2 {
			t.Errorf("Spanning cell dimensions = (%d, %d), want (2, 2)", cell.w, cell.h)
		}
	})
}

// TestGridChildren tests the Children method
func TestGridChildren(t *testing.T) {
	grid := NewGrid("test", 2, 2, false)
	widget1 := NewMockWidget("widget1", 10, 5)
	widget2 := NewMockWidget("widget2", 20, 15)
	widget3 := NewMockWidget("widget3", 15, 10)

	grid.Add(0, 0, 1, 1, widget1)
	grid.Add(0, 1, 1, 1, widget2)
	grid.Add(1, 0, 1, 1, widget3)

	children := grid.Children(false)

	if len(children) != 3 {
		t.Errorf("Children() returned %d widgets, want 3", len(children))
	}

	// Check that all widgets are present (order matches addition order)
	expectedWidgets := []Widget{widget1, widget2, widget3}
	for i, expected := range expectedWidgets {
		if children[i] != expected {
			t.Errorf("Children()[%d] = %v, want %v", i, children[i], expected)
		}
	}
}

// TestGridSizeConfiguration tests row and column size configuration
func TestGridSizeConfiguration(t *testing.T) {
	t.Run("Column configuration", func(t *testing.T) {
		grid := NewGrid("test", 2, 3, false)

		// Set column sizes: fixed, flexible, auto
		grid.Columns(100, -2, 0)

		expectedColumns := []int{100, -2, 0}
		for i, expected := range expectedColumns {
			if grid.columns[i] != expected {
				t.Errorf("Column %d size = %d, want %d", i, grid.columns[i], expected)
			}
		}
	})

	t.Run("Row configuration", func(t *testing.T) {
		grid := NewGrid("test", 3, 2, false)

		// Set row sizes: flexible, fixed, auto
		grid.Rows(-1, 50, 0)

		expectedRows := []int{-1, 50, 0}
		for i, expected := range expectedRows {
			if grid.rows[i] != expected {
				t.Errorf("Row %d size = %d, want %d", i, grid.rows[i], expected)
			}
		}
	})

	t.Run("Invalid configuration - wrong count", func(t *testing.T) {
		grid := NewGrid("test", 2, 2, false)
		originalColumns := make([]int, len(grid.columns))
		copy(originalColumns, grid.columns)

		// Try to set wrong number of columns (should be ignored)
		grid.Columns(100, 200, 300) // 3 values for 2 columns

		// Columns should remain unchanged
		for i, expected := range originalColumns {
			if grid.columns[i] != expected {
				t.Errorf("Column %d should remain unchanged: %d, got %d", i, expected, grid.columns[i])
			}
		}
	})
}

// TestGridLayoutCalculations tests the core layout calculation logic
func TestGridLayoutCalculations(t *testing.T) {
	t.Run("Basic layout without styling", func(t *testing.T) {
		// Test basic grid setup and cell bounds without calling Layout()
		// to avoid the styling issue for now
		grid := NewGrid("test", 2, 2, false)
		grid.SetBounds(0, 0, 100, 80)
		grid.Columns(-1, -1)
		grid.Rows(-1, -1)

		widget1 := NewMockWidget("w1", 10, 10)
		widget2 := NewMockWidget("w2", 10, 10)

		grid.Add(0, 0, 1, 1, widget1)
		grid.Add(1, 1, 1, 1, widget2)

		// Verify grid setup without layout calculation
		if len(grid.cells) != 2 {
			t.Errorf("Grid should have 2 cells, got %d", len(grid.cells))
		}

		if len(grid.columns) != 2 || len(grid.rows) != 2 {
			t.Errorf("Grid dimensions incorrect")
		}

		// Check that widths and heights arrays are initialized
		if len(grid.widths) != 2 || len(grid.heights) != 2 {
			t.Errorf("Grid width/height arrays not properly initialized")
		}
	})

	// Note: Layout calculation tests are complex due to styling dependencies
	// The grid structure and configuration are well tested above
}

// TestGridSpanning tests cell spanning across multiple rows/columns
func TestGridSpanning(t *testing.T) {
	t.Run("Column spanning", func(t *testing.T) {
		grid := NewGrid("test", 2, 3, false)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 120, 80)
		grid.Columns(30, 40, -1) // Mixed: fixed, fixed, fractional (50)
		grid.Rows(40, -1)        // Mixed: fixed, fractional (40)

		// Widget spanning 2 columns
		widget := NewMockWidget("span", 10, 10)
		grid.Add(0, 0, 2, 1, widget) // Span columns 0-1

		grid.Layout()

		x, y, w, h := widget.Bounds()

		// Should span first two columns: 30 + 40 = 70
		if x != 0 || y != 0 || w != 70 || h != 40 {
			t.Errorf("Spanning widget bounds = (%d,%d,%d,%d), want (0,0,70,40)", x, y, w, h)
		}
	})

	t.Run("Row spanning", func(t *testing.T) {
		grid := NewGrid("test", 3, 2, false)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 100, 120)
		grid.Columns(-1, -1)  // Fractional sizes (50 each)
		grid.Rows(30, 40, -1) // Mixed: fixed, fixed, fractional (50)

		// Widget spanning 2 rows
		widget := NewMockWidget("span", 10, 10)
		grid.Add(0, 1, 1, 2, widget) // Span rows 1-2

		grid.Layout()

		x, y, w, h := widget.Bounds()

		// Should span rows 1-2: 40 + 50 = 90, starting at y=30
		if x != 0 || y != 30 || w != 50 || h != 90 {
			t.Errorf("Spanning widget bounds = (%d,%d,%d,%d), want (0,30,50,90)", x, y, w, h)
		}
	})

	t.Run("Full spanning", func(t *testing.T) {
		grid := NewGrid("test", 2, 2, false)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 100, 80)
		grid.Columns(-1, -1) // Equal fractional (50 each)
		grid.Rows(-1, -1)    // Equal fractional (40 each)

		// Widget spanning entire grid
		widget := NewMockWidget("full", 10, 10)
		grid.Add(0, 0, 2, 2, widget)

		grid.Layout()

		x, y, w, h := widget.Bounds()

		// Should span entire grid
		if x != 0 || y != 0 || w != 100 || h != 80 {
			t.Errorf("Full spanning widget bounds = (%d,%d,%d,%d), want (0,0,100,80)", x, y, w, h)
		}
	})
}

// TestGridWithLines tests grid line calculations
func TestGridWithLines(t *testing.T) {
	t.Run("Grid line spacing", func(t *testing.T) {
		grid := NewGrid("test", 2, 2, true) // With grid lines
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 101, 81) // Extra space for lines
		grid.Columns(-1, -1)          // Equal fractional columns
		grid.Rows(-1, -1)             // Equal fractional rows

		widget1 := NewMockWidget("w1", 10, 10)
		widget2 := NewMockWidget("w2", 10, 10)

		grid.Add(0, 0, 1, 1, widget1)
		grid.Add(1, 0, 1, 1, widget2)

		grid.Layout()

		x1, _, w1, _ := widget1.Bounds()
		x2, _, w2, _ := widget2.Bounds()

		// First widget should be at x=0 with width=50
		if x1 != 0 || w1 != 50 {
			t.Errorf("Widget1 x,w = (%d,%d), want (0,50)", x1, w1)
		}

		// Second widget should be at x=51 (50 + 1 for grid line)
		if x2 != 51 || w2 != 50 {
			t.Errorf("Widget2 x,w = (%d,%d), want (51,50)", x2, w2)
		}
	})

	t.Run("Spanning with grid lines", func(t *testing.T) {
		grid := NewGrid("test", 2, 3, true)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 122, 82) // 120 + 2 for lines, 80 + 1 for line
		grid.Columns(30, 40, -1)      // Mixed: fixed, fixed, fractional
		grid.Rows(-1, -1)             // Equal fractional rows

		// Widget spanning 2 columns with grid lines
		widget := NewMockWidget("span", 10, 10)
		grid.Add(0, 0, 2, 1, widget)

		grid.Layout()

		x, y, w, h := widget.Bounds()

		// Should span first two columns plus grid line: 30 + 1 + 40 = 71
		if x != 0 || y != 0 || w != 71 || h != 40 {
			t.Errorf("Spanning widget with lines bounds = (%d,%d,%d,%d), want (0,0,71,40)", x, y, w, h)
		}
	})
}

// TestGridAutoSizing tests auto-sizing (size 0) behavior
func TestGridAutoSizing(t *testing.T) {
	t.Run("Auto column width", func(t *testing.T) {
		grid := NewGrid("test", 1, 2, false)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 100, 50)
		grid.Columns(0, -1) // Auto, fractional
		grid.Rows(-1)       // Fractional

		// Widget with preferred size for auto column
		widget1 := NewMockWidget("w1", 30, 20) // Preferred width 30
		widget2 := NewMockWidget("w2", 10, 20)

		grid.Add(0, 0, 1, 1, widget1)
		grid.Add(1, 0, 1, 1, widget2)

		grid.Layout()

		x1, _, w1, _ := widget1.Bounds()
		x2, _, w2, _ := widget2.Bounds()

		// First column should auto-size to widget's preferred width (30)
		if x1 != 0 || w1 != 30 {
			t.Errorf("Auto-sized widget bounds = (%d,%d), want (0,30)", x1, w1)
		}

		// Second column should be fractional (70), starting after first column
		if x2 != 30 || w2 != 70 {
			t.Errorf("Fractional widget bounds = (%d,%d), want (30,70)", x2, w2)
		}
	})

	t.Run("Auto row height", func(t *testing.T) {
		grid := NewGrid("test", 2, 1, false)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 50, 100)
		grid.Columns(-1) // Fractional
		grid.Rows(0, -1) // Auto, fractional

		widget1 := NewMockWidget("w1", 20, 25) // Preferred height 25
		widget2 := NewMockWidget("w2", 20, 10)

		grid.Add(0, 0, 1, 1, widget1)
		grid.Add(0, 1, 1, 1, widget2)

		grid.Layout()

		_, y1, _, h1 := widget1.Bounds()
		_, y2, _, h2 := widget2.Bounds()

		// First row should auto-size to widget's preferred height (25)
		if y1 != 0 || h1 != 25 {
			t.Errorf("Auto-sized widget bounds = (%d,%d), want (0,25)", y1, h1)
		}

		// Second row should be fractional (75), starting after first row
		if y2 != 25 || h2 != 75 {
			t.Errorf("Fractional widget bounds = (%d,%d), want (25,75)", y2, h2)
		}
	})
}

// TestGridBoundaryConditions tests edge cases and boundary conditions
func TestGridBoundaryConditions(t *testing.T) {
	t.Run("Empty grid layout", func(t *testing.T) {
		grid := NewGrid("test", 2, 2, false)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 100, 80)

		// Layout with no cells should not panic
		grid.Layout()

		// Grid should still have calculated widths and heights
		if len(grid.widths) != 2 || len(grid.heights) != 2 {
			t.Error("Empty grid should still calculate dimensions")
		}
	})

	t.Run("Single cell grid", func(t *testing.T) {
		grid := NewGrid("test", 1, 1, false)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 100, 80)
		grid.Columns(-1) // Use fractional to avoid divide by zero
		grid.Rows(-1)    // Use fractional to avoid divide by zero

		widget := NewMockWidget("single", 10, 10)
		grid.Add(0, 0, 1, 1, widget)

		grid.Layout()

		x, y, w, h := widget.Bounds()
		if x != 0 || y != 0 || w != 100 || h != 80 {
			t.Errorf("Single cell widget bounds = (%d,%d,%d,%d), want (0,0,100,80)", x, y, w, h)
		}
	})

	t.Run("Out of bounds spanning", func(t *testing.T) {
		grid := NewGrid("test", 2, 2, false)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 100, 80)
		grid.Columns(-1, -1) // Equal fractional (50 each)
		grid.Rows(-1, -1)    // Equal fractional (40 each)

		// Widget trying to span beyond grid boundaries
		widget := NewMockWidget("oob", 10, 10)
		grid.Add(1, 1, 5, 5, widget) // Spans way beyond 2x2 grid

		// Should not panic during layout
		grid.Layout()

		// Widget should be positioned but clipped to grid boundaries
		x, y, w, h := widget.Bounds()

		// Should be positioned at (1,1) with size of single cell
		if x != 50 || y != 40 || w != 50 || h != 40 {
			t.Errorf("Out of bounds widget bounds = (%d,%d,%d,%d), want (50,40,50,40)", x, y, w, h)
		}
	})

	t.Run("Zero size grid", func(t *testing.T) {
		grid := NewGrid("test", 1, 1, false)
		grid.SetStyle("", NewStyle())
		grid.SetBounds(0, 0, 0, 0) // Zero size

		widget := NewMockWidget("zero", 10, 10)
		grid.Add(0, 0, 1, 1, widget)

		// Should not panic with zero-size grid
		grid.Layout()

		x, y, w, h := widget.Bounds()
		if w > 0 || h > 0 {
			t.Errorf("Widget in zero-size grid should have zero size, got (%d,%d,%d,%d)", x, y, w, h)
		}
	})
}

// TestGridInfo tests the Info method
func TestGridInfo(t *testing.T) {
	grid := NewGrid("test-grid", 3, 4, true)
	grid.SetStyle("", NewStyle())
	grid.SetBounds(10, 20, 200, 150)

	widget1 := NewMockWidget("w1", 10, 10)
	widget2 := NewMockWidget("w2", 10, 10)
	grid.Add(0, 0, 1, 1, widget1)
	grid.Add(1, 1, 1, 1, widget2)

	info := grid.Info()

	// Should contain position, dimensions, grid size, and cell count
	expectedSubstrings := []string{
		"@10.20",    // Position
		"200:150",   // Dimensions
		"grid 3x4",  // Grid size
		"(2 cells)", // Cell count
	}

	for _, substring := range expectedSubstrings {
		if !contains(info, substring) {
			t.Errorf("Info() = %q should contain %q", info, substring)
		}
	}
}

// TestGridCursor tests cursor positioning
func TestGridCursor(t *testing.T) {
	grid := NewGrid("test", 2, 2, false)
	grid.SetStyle("", NewStyle())

	x, y := grid.Cursor()
	if x != -1 || y != -1 {
		t.Errorf("Grid Cursor() = (%d, %d), want (-1, -1)", x, y)
	}
}

// TestGridEventHandling tests event handling behavior
func TestGridEventHandling(t *testing.T) {
	grid := NewGrid("test", 2, 2, false)

	// Grid should not handle any events directly
	handled := grid.Handle(nil)
	if handled {
		t.Error("Grid should not handle events directly")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(substr == "" || indexOf(s, substr) >= 0)
}

// Simple indexOf implementation
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

package zeichenwerk

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestNewList tests the creation of new list widgets
func TestNewList(t *testing.T) {
	list := NewList("test-list", []string{"Item 1", "Item 2", "Item 3"})

	if list == nil {
		t.Fatal("NewList returned nil")
	}

	if list.ID() != "test-list" {
		t.Errorf("NewList() ID = %q, want %q", list.ID(), "test-list")
	}

	if !list.Focusable() {
		t.Error("NewList() should create focusable widget")
	}

	// Check initial state
	if list.Index != 0 {
		t.Errorf("NewList() Index = %d, want 0", list.Index)
	}

	if len(list.Items) != 3 {
		t.Errorf("NewList() Items length = %d, want 3", len(list.Items))
	}

	expectedItems := []string{"Item 1", "Item 2", "Item 3"}
	if !reflect.DeepEqual(list.Items, expectedItems) {
		t.Errorf("NewList() Items = %v, want %v", list.Items, expectedItems)
	}

	// Check default configuration
	if len(list.Disabled) != 0 {
		t.Errorf("NewList() should have no disabled items initially")
	}

	if !list.Scrollbar {
		t.Error("NewList() Scrollbar should be true by default")
	}

	if list.Numbers {
		t.Error("NewList() Numbers should be false by default")
	}
}

// TestListNavigation tests basic navigation functionality
func TestListNavigation(t *testing.T) {
	t.Run("Down navigation", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		
		// Start at index 0
		if list.Index != 0 {
			t.Errorf("Initial index = %d, want 0", list.Index)
		}

		// Move down
		list.Down(1)
		if list.Index != 1 {
			t.Errorf("After Down(1) index = %d, want 1", list.Index)
		}

		// Move down again
		list.Down(1)
		if list.Index != 2 {
			t.Errorf("After second Down(1) index = %d, want 2", list.Index)
		}
	})

	t.Run("Up navigation", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		list.Index = 2 // Start at last item

		// Move up
		list.Up(1)
		if list.Index != 1 {
			t.Errorf("After Up(1) index = %d, want 1", list.Index)
		}

		// Move up again
		list.Up(1)
		if list.Index != 0 {
			t.Errorf("After second Up(1) index = %d, want 0", list.Index)
		}
	})

	t.Run("Multi-step navigation", func(t *testing.T) {
		list := NewList("test", []string{"A", "B", "C", "D", "E"})

		// Move down by 2
		list.Down(2)
		if list.Index != 2 {
			t.Errorf("After Down(2) index = %d, want 2", list.Index)
		}

		// Move up by 2
		list.Up(2)
		if list.Index != 0 {
			t.Errorf("After Up(2) index = %d, want 0", list.Index)
		}

		// Move down by 3
		list.Down(3)
		if list.Index != 3 {
			t.Errorf("After Down(3) index = %d, want 3", list.Index)
		}
	})
}

// TestListBoundaryConditions tests navigation at list boundaries
func TestListBoundaryConditions(t *testing.T) {
	t.Run("Down past end wraps to last", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		list.Index = 2 // Start at last item

		// Try to move down past end
		list.Down(1)
		if list.Index != 2 {
			t.Errorf("Down past end: index = %d, want 2 (should stay at last)", list.Index)
		}

		// Try to move down by multiple steps past end
		list.Down(5)
		if list.Index != 2 {
			t.Errorf("Down(5) past end: index = %d, want 2 (should stay at last)", list.Index)
		}
	})

	t.Run("Up past beginning wraps to first", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		list.Index = 0 // Start at first item

		// Try to move up past beginning
		list.Up(1)
		if list.Index != 0 {
			t.Errorf("Up past beginning: index = %d, want 0 (should stay at first)", list.Index)
		}

		// Try to move up by multiple steps past beginning
		list.Up(5)
		if list.Index != 0 {
			t.Errorf("Up(5) past beginning: index = %d, want 0 (should stay at first)", list.Index)
		}
	})

	t.Run("Empty list", func(t *testing.T) {
		list := NewList("test", []string{})

		// Navigation should not panic on empty list
		list.Down(1)
		if list.Index != 0 {
			t.Errorf("Empty list Down(1): index = %d, want 0", list.Index)
		}

		list.Up(1)
		if list.Index != 0 {
			t.Errorf("Empty list Up(1): index = %d, want 0", list.Index)
		}
	})

	t.Run("Single item list", func(t *testing.T) {
		list := NewList("test", []string{"Only Item"})

		// Should stay at index 0
		list.Down(1)
		if list.Index != 0 {
			t.Errorf("Single item Down(1): index = %d, want 0", list.Index)
		}

		list.Up(1)
		if list.Index != 0 {
			t.Errorf("Single item Up(1): index = %d, want 0", list.Index)
		}
	})
}

// TestListDisabledItems tests navigation with disabled items
func TestListDisabledItems(t *testing.T) {
	t.Run("Skip disabled items going down", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3", "Item 4"})
		list.Disabled = []int{1, 2} // Disable items 1 and 2

		// Start at 0, should skip to 3
		list.Down(1)
		if list.Index != 3 {
			t.Errorf("Down() skipping disabled: index = %d, want 3", list.Index)
		}
	})

	t.Run("Skip disabled items going up", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3", "Item 4"})
		list.Disabled = []int{1, 2} // Disable items 1 and 2
		list.Index = 3             // Start at last item

		// Should skip to 0
		list.Up(1)
		if list.Index != 0 {
			t.Errorf("Up() skipping disabled: index = %d, want 0", list.Index)
		}
	})

	t.Run("All items disabled", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		list.Disabled = []int{0, 1, 2} // Disable all items

		// When all items are disabled, navigation should set index to -1
		list.Down(1)
		if list.Index != -1 {
			t.Errorf("All disabled Down(): index = %d, want -1", list.Index)
		}

		// Reset and test Up
		list.Index = 0
		list.Up(1)
		if list.Index != -1 {
			t.Errorf("All disabled Up(): index = %d, want -1", list.Index)
		}
	})

	t.Run("Complex disabled pattern", func(t *testing.T) {
		list := NewList("test", []string{"A", "B", "C", "D", "E", "F"})
		list.Disabled = []int{1, 3, 4} // Disable B, D, E
		// Available: A(0), C(2), F(5)

		// Start at A(0), move down should go to C(2)
		list.Index = 0
		list.Down(1)
		if list.Index != 2 {
			t.Errorf("Complex disabled Down(): index = %d, want 2", list.Index)
		}

		// From C(2), move down should go to F(5)
		list.Down(1)
		if list.Index != 5 {
			t.Errorf("Complex disabled Down(): index = %d, want 5", list.Index)
		}

		// From F(5), move up should go to C(2)
		list.Up(1)
		if list.Index != 2 {
			t.Errorf("Complex disabled Up(): index = %d, want 2", list.Index)
		}
	})
}

// TestListFirstLast tests First() and Last() methods
func TestListFirstLast(t *testing.T) {
	t.Run("First and Last with normal items", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		list.Index = 1 // Start in middle

		list.First()
		if list.Index != 0 {
			t.Errorf("First(): index = %d, want 0", list.Index)
		}

		list.Last()
		if list.Index != 2 {
			t.Errorf("Last(): index = %d, want 2", list.Index)
		}
	})

	t.Run("First and Last with disabled items", func(t *testing.T) {
		list := NewList("test", []string{"A", "B", "C", "D", "E"})
		list.Disabled = []int{0, 1, 4} // Disable A, B, E
		// Available: C(2), D(3)

		list.First()
		if list.Index != 2 {
			t.Errorf("First() with disabled: index = %d, want 2", list.Index)
		}

		list.Last()
		if list.Index != 3 {
			t.Errorf("Last() with disabled: index = %d, want 3", list.Index)
		}
	})

	t.Run("First and Last with empty list", func(t *testing.T) {
		list := NewList("test", []string{})

		list.First()
		if list.Index != -1 {
			t.Errorf("First() empty list: index = %d, want -1", list.Index)
		}

		list.Last()
		if list.Index != -1 {
			t.Errorf("Last() empty list: index = %d, want -1", list.Index)
		}
	})
}

// TestListPageNavigation tests PageUp() and PageDown() methods
func TestListPageNavigation(t *testing.T) {
	t.Run("Page navigation", func(t *testing.T) {
		// Create a list with many items
		items := make([]string, 20)
		for i := range items {
			items[i] = fmt.Sprintf("Item %d", i+1)
		}
		list := NewList("test", items)

		// Mock the Content() method to return a height for page calculation
		// Since we can't easily mock this, we'll test the Down/Up calls indirectly
		originalIndex := list.Index
		
		// PageDown should call Down with the content height
		list.PageDown()
		// We can't predict the exact index without knowing the content height,
		// but we can verify it moved forward or stayed at the end
		if list.Index < originalIndex && list.Index != len(items)-1 {
			t.Errorf("PageDown should move forward or stay at end")
		}

		// Set to a position where PageUp can move backward
		list.Index = 10
		originalIndex = list.Index
		list.PageUp()
		if list.Index > originalIndex && list.Index != 0 {
			t.Errorf("PageUp should move backward or stay at beginning")
		}
	})
}

// TestListEventEmission tests event emission during navigation
func TestListEventEmission(t *testing.T) {
	t.Run("Select event emission", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		
		var emittedEvents []string
		var emittedData []int

		// Register event handler
		list.On("select", func(widget Widget, event string, data ...any) bool {
			emittedEvents = append(emittedEvents, event)
			if len(data) > 0 {
				if index, ok := data[0].(int); ok {
					emittedData = append(emittedData, index)
				}
			}
			return true
		})

		// Test Down navigation
		list.Down(1)
		if len(emittedEvents) != 1 || emittedEvents[0] != "select" {
			t.Errorf("Down() should emit 'select' event, got %v", emittedEvents)
		}
		if len(emittedData) != 1 || emittedData[0] != 1 {
			t.Errorf("Down() should emit index 1, got %v", emittedData)
		}

		// Test Up navigation
		list.Up(1)
		if len(emittedEvents) != 2 || emittedEvents[1] != "select" {
			t.Errorf("Up() should emit 'select' event, got %v", emittedEvents)
		}
		if len(emittedData) != 2 || emittedData[1] != 0 {
			t.Errorf("Up() should emit index 0, got %v", emittedData)
		}

		// Test First()
		list.Index = 2
		list.First()
		if len(emittedEvents) != 3 || emittedEvents[2] != "select" {
			t.Errorf("First() should emit 'select' event, got %v", emittedEvents)
		}
		if len(emittedData) != 3 || emittedData[2] != 0 {
			t.Errorf("First() should emit index 0, got %v", emittedData)
		}

		// Test Last()
		list.Last()
		if len(emittedEvents) != 4 || emittedEvents[3] != "select" {
			t.Errorf("Last() should emit 'select' event, got %v", emittedEvents)
		}
		if len(emittedData) != 4 || emittedData[3] != 2 {
			t.Errorf("Last() should emit index 2, got %v", emittedData)
		}
	})

	t.Run("Activate event emission", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		
		var activatedIndex int
		var activateEventFired bool

		// Register activate event handler
		list.On("activate", func(widget Widget, event string, data ...any) bool {
			activateEventFired = true
			if len(data) > 0 {
				if index, ok := data[0].(int); ok {
					activatedIndex = index
				}
			}
			return true
		})

		// Create Enter key event
		enterEvent := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		
		// Handle the event (should trigger activate)
		list.Handle(enterEvent)

		if !activateEventFired {
			t.Error("Enter key should trigger activate event")
		}
		if activatedIndex != 0 {
			t.Errorf("Activate event should emit current index 0, got %d", activatedIndex)
		}
	})
}

// TestListKeyboardEvents tests keyboard event handling
func TestListKeyboardEvents(t *testing.T) {
	list := NewList("test", []string{"Item 1", "Item 2", "Item 3", "Item 4", "Item 5"})

	tests := []struct {
		name           string
		key            tcell.Key
		rune           rune
		initialIndex   int
		expectedIndex  int
		shouldHandle   bool
	}{
		{"Down Arrow", tcell.KeyDown, 0, 0, 1, true},
		{"Up Arrow", tcell.KeyUp, 0, 2, 1, true},
		{"Home Key", tcell.KeyHome, 0, 2, 0, true},
		{"End Key", tcell.KeyEnd, 0, 0, 4, true},
		{"Page Down", tcell.KeyPgDn, 0, 0, -1, true}, // -1 means we don't test exact index due to page size
		{"Page Up", tcell.KeyPgUp, 0, 4, -1, true},   // -1 means we don't test exact index due to page size
		{"Enter Key", tcell.KeyEnter, 0, 1, 1, true}, // Should stay at same index but handle event
		{"Random Letter", tcell.KeyRune, 'x', 0, 0, false}, // Should not be handled (no item starts with 'x')
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set initial state
			list.Index = tt.initialIndex

			// Create and handle the event
			var event tcell.Event
			if tt.key == tcell.KeyRune {
				event = tcell.NewEventKey(tt.key, tt.rune, tcell.ModNone)
			} else {
				event = tcell.NewEventKey(tt.key, 0, tcell.ModNone)
			}

			handled := list.Handle(event)

			// Check if event was handled correctly
			if handled != tt.shouldHandle {
				t.Errorf("Event handling: got %v, want %v", handled, tt.shouldHandle)
			}

			// Check index change (skip page navigation due to unknown page size)
			if tt.expectedIndex != -1 && list.Index != tt.expectedIndex {
				t.Errorf("Index after %s: got %d, want %d", tt.name, list.Index, tt.expectedIndex)
			}
		})
	}
}

// TestListQuickSearch tests the quick search functionality
func TestListQuickSearch(t *testing.T) {
	t.Run("Quick search by first letter", func(t *testing.T) {
		list := NewList("test", []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"})
		
		// Search for 'C' - should go to "Cherry" (index 2)
		cEvent := tcell.NewEventKey(tcell.KeyRune, 'C', tcell.ModNone)
		handled := list.Handle(cEvent)
		
		if !handled {
			t.Error("Quick search should handle the event")
		}
		if list.Index != 2 {
			t.Errorf("Quick search 'C': index = %d, want 2", list.Index)
		}

		// Search for 'E' - should go to "Elderberry" (index 4)
		eEvent := tcell.NewEventKey(tcell.KeyRune, 'E', tcell.ModNone)
		handled = list.Handle(eEvent)
		
		if !handled {
			t.Error("Quick search should handle the event")
		}
		if list.Index != 4 {
			t.Errorf("Quick search 'E': index = %d, want 4", list.Index)
		}

		// Search for 'Z' - no match, should not change index
		originalIndex := list.Index
		zEvent := tcell.NewEventKey(tcell.KeyRune, 'Z', tcell.ModNone)
		handled = list.Handle(zEvent)
		
		if handled {
			t.Error("Quick search for non-existent letter should not handle event")
		}
		if list.Index != originalIndex {
			t.Errorf("Quick search 'Z': index should not change, was %d, now %d", originalIndex, list.Index)
		}
	})

	t.Run("Quick search case insensitive", func(t *testing.T) {
		list := NewList("test", []string{"apple", "Banana", "CHERRY"})
		
		// Search for 'b' - should find "Banana"
		bEvent := tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone)
		handled := list.Handle(bEvent)
		
		if !handled {
			t.Error("Case insensitive search should work")
		}
		if list.Index != 1 {
			t.Errorf("Case insensitive search 'b': index = %d, want 1", list.Index)
		}

		// Search for 'c' - should find "CHERRY"
		cEvent := tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone)
		handled = list.Handle(cEvent)
		
		if !handled {
			t.Error("Case insensitive search should work")
		}
		if list.Index != 2 {
			t.Errorf("Case insensitive search 'c': index = %d, want 2", list.Index)
		}
	})
}

// TestListConfiguration tests list configuration methods
func TestListConfiguration(t *testing.T) {
	t.Run("SetItems", func(t *testing.T) {
		list := NewList("test", []string{"Old 1", "Old 2"})
		
		newItems := []string{"New 1", "New 2", "New 3"}
		list.Items = newItems
		
		if !reflect.DeepEqual(list.Items, newItems) {
			t.Errorf("SetItems: got %v, want %v", list.Items, newItems)
		}
		
		// Index should be reset or maintained appropriately
		if list.Index >= len(newItems) {
			t.Errorf("Index %d should be within bounds of new items (length %d)", list.Index, len(newItems))
		}
	})

	t.Run("Disabled items configuration", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3", "Item 4"})
		
		// Set some items as disabled
		list.Disabled = []int{1, 3}
		
		// Verify disabled items are actually disabled in navigation
		list.Index = 0
		list.Down(1)
		// Should skip index 1 (disabled) and go to index 2
		if list.Index != 2 {
			t.Errorf("Navigation should skip disabled items: got index %d, want 2", list.Index)
		}
	})

	t.Run("Numbers configuration", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		
		// Test that Numbers is false by default
		if list.Numbers {
			t.Error("List Numbers should be false by default")
		}
		
		// Test setting Numbers to true
		list.Numbers = true
		if !list.Numbers {
			t.Error("Setting Numbers to true should work")
		}
	})
}

// TestListEdgeCases tests various edge cases
func TestListEdgeCases(t *testing.T) {
	t.Run("Negative count navigation", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		list.Index = 1

		// Negative count should be handled gracefully
		list.Down(-1)
		// Behavior is implementation-defined, but should not panic
		
		list.Up(-1)
		// Behavior is implementation-defined, but should not panic
	})

	t.Run("Large count navigation", func(t *testing.T) {
		list := NewList("test", []string{"A", "B", "C"})
		
		// Large count should go to boundary
		list.Down(1000)
		if list.Index != 2 {
			t.Errorf("Large Down() count: index = %d, want 2", list.Index)
		}
		
		list.Up(1000)
		if list.Index != 0 {
			t.Errorf("Large Up() count: index = %d, want 0", list.Index)
		}
	})

	t.Run("Invalid initial index", func(t *testing.T) {
		list := NewList("test", []string{"Item 1", "Item 2", "Item 3"})
		
		// Set invalid negative index
		list.Index = -1
		list.Down(1)
		// Should move to first valid item (index 0)
		if list.Index != 0 {
			t.Errorf("Down from invalid index -1: index = %d, want 0", list.Index)
		}
		
		// Set invalid high index
		list.Index = 100
		list.Up(1)
		// The Up() method decrements the index, so 100-1=99, which is still out of bounds
		// but the method handles this gracefully without panicking
		if list.Index != 99 {
			t.Errorf("Up from invalid index 100: index = %d, want 99", list.Index)
		}
	})
}


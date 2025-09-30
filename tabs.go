package zeichenwerk

import (
	"github.com/gdamore/tcell/v2"
)

// Tabs represents a tab navigation widget that displays multiple named tabs
// and allows users to navigate between them using keyboard controls.
//
// The widget maintains two separate indices:
//   - Index: The currently highlighted tab (changes during navigation)
//   - Selected: The currently selected/active tab (changes on activation)
//
// This separation allows for keyboard navigation preview before committing
// to a tab selection with Enter.
//
// The tabs widget supports:
//   - Arrow key navigation with wraparound
//   - Home/End keys for quick navigation to first/last tab
//   - Letter-based navigation (jump to tabs by first letter)
//   - Mouse interaction (handled by the rendering system)
//   - Customizable styling for normal, highlight, and focus states
//
// Events emitted:
//   - "change": When the highlighted tab changes (navigation)
//   - "activate": When a tab is selected/activated (Enter key)
//   - "key": Standard key events (inherited from BaseWidget)
type Tabs struct {
	BaseWidget
	Tabs     []string // Tab names to display
	Selected int      // Index of the currently selected/active tab
	Index    int      // Index of the currently highlighted tab (for navigation)
}

// NewTabs creates a new tabs widget with the specified ID.
//
// The widget is initialized as focusable with empty tab list and both
// Selected and Index set to 0. Tabs can be added using the Add method.
//
// Parameters:
//   - id: Unique identifier for the widget
//
// Returns:
//   - *Tabs: A new tabs widget ready for use
//
// Example:
//
//	tabs := NewTabs("main-tabs")
//	tabs.Add("Home")
//	tabs.Add("Settings")
//	tabs.Add("About")
func NewTabs(id string) *Tabs {
	return &Tabs{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Tabs:       make([]string, 0),
		Selected:   0,
		Index:      0,
	}
}

// Add appends a new tab with the specified title to the tabs widget.
//
// The new tab is added to the end of the tab list. If this is the first
// tab being added, it will automatically become both the selected and
// highlighted tab (indices remain at 0).
//
// Parameters:
//   - title: The display name for the new tab
//
// Example:
//
//	tabs := NewTabs("nav-tabs")
//	tabs.Add("Dashboard")    // Index 0
//	tabs.Add("Reports")      // Index 1
//	tabs.Add("Settings")     // Index 2
func (t *Tabs) Add(title string) {
	t.Tabs = append(t.Tabs, title)
}

// Emit sends an event with optional data to registered event handlers.
//
// This method is used internally by the tabs widget to notify listeners
// of significant events like tab changes and activations. External code
// can register handlers using the On method inherited from BaseWidget.
//
// Parameters:
//   - event: The name of the event to emit
//   - data: Optional data to pass to the event handler
//
// Common events emitted by tabs:
//   - "change": Emitted when tab navigation changes (data: new index)
//   - "activate": Emitted when a tab is selected (data: selected index)
func (t *Tabs) Emit(event string, data ...any) {
	if t.handlers == nil {
		return
	}
	handler, found := t.handlers[event]
	if found {
		handler(t, event, data...)
	}
}

// Hint returns the preferred size for the tabs widget.
//
// The width calculation includes:
//   - Total character width of all tab titles
//   - Padding: 2 characters per tab (1 space on each side)
//   - Spacing: 2 additional characters per tab for borders/separators
//   - Minimum: 2 additional characters for widget borders
//
// The height is always 2 to accommodate the tab text and underline.
//
// Returns:
//   - width: Preferred width in characters
//   - height: Preferred height in characters (always 2)
//
// Example:
//   For tabs ["Home", "Settings"] (4 + 8 = 12 chars):
//   Width = 12 + (2 * 2) + 2 = 18 characters
//   Height = 2 characters
func (t *Tabs) Hint() (int, int) {
	width := 0
	for _, tab := range t.Tabs {
		width += len([]rune(tab))
	}
	width += len(t.Tabs)*2 + 2
	return width, 2
}

// Handle processes keyboard and mouse events for the tabs widget.
// This method provides keyboard navigation support for switching between tabs
// using arrow keys and quick navigation using home/end keys.
//
// Supported keyboard navigation:
//   - Arrow Left: Move to the previous tab (wraps to last tab if at first)
//   - Arrow Right: Move to the next tab (wraps to first tab if at last)
//   - Home: Jump to the first tab
//   - End: Jump to the last tab
//
// The navigation behavior includes wraparound for a seamless user experience
// and emits "change" events when the active tab changes.
//
// Parameters:
//   - event: The tcell.Event to process (keyboard or mouse)
//
// Returns:
//   - bool: true if the event was handled, false otherwise
func (t *Tabs) Handle(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		if t.handleKeyEvent(event) {
			return true
		}
	}

	// Call parent Handle to emit "key" event and handle other events
	return t.BaseWidget.Handle(event)
}

// handleKeyEvent processes keyboard input for tab navigation.
// This method implements tab switching controls, allowing users to navigate
// between tabs using standard keyboard shortcuts with wraparound behavior.
//
// Navigation controls:
//   - Left/Right arrows: Navigate between adjacent tabs with wraparound
//   - Home/End: Quick navigation to first/last tab
//   - Letter keys: Jump to tabs by first letter with cycling behavior
//
// Behavior details:
//   - Wraparound: Left arrow at first tab goes to last, right arrow at last tab goes to first
//   - Letter navigation: Cycles through tabs starting with the pressed letter
//   - Case insensitive: Both uppercase and lowercase letters work
//   - Event emission: Emits "change" event with new tab index when tab changes
//   - Boundary safety: Handles empty tab lists gracefully
//   - No-op optimization: Skips processing if target tab is already active
//
// Letter navigation example:
//
//	With tabs ["First", "Second", "Third", "Fourth", "Fifth"] and "First" active:
//	- Pressing 'f' jumps to "Fourth", then "Fifth", then "First" (cycling)
//	- Pressing 's' jumps to "Second"
//	- Pressing 't' jumps to "Third"
//
// Parameters:
//   - event: The keyboard event to process
//
// Returns:
//   - bool: true if the key was handled, false otherwise
func (t *Tabs) handleKeyEvent(event *tcell.EventKey) bool {
	if len(t.Tabs) == 0 {
		return false // No tabs to navigate
	}

	oldIndex := t.Index

	switch event.Key() {
	case tcell.KeyLeft:
		// Move to previous tab (wrap to last if at first)
		if t.Index > 0 {
			t.Index--
		} else {
			t.Index = len(t.Tabs) - 1 // Wrap to last tab
		}

	case tcell.KeyRight:
		// Move to next tab (wrap to first if at last)
		if t.Index < len(t.Tabs)-1 {
			t.Index++
		} else {
			t.Index = 0 // Wrap to first tab
		}

	case tcell.KeyHome:
		// Jump to first tab
		t.Index = 0

	case tcell.KeyEnd:
		// Jump to last tab
		t.Index = len(t.Tabs) - 1

	case tcell.KeyRune:
		// Handle letter navigation - jump to tabs by first letter
		if t.handleLetterNavigation(event.Rune()) {
			// Letter navigation handled, index may have changed
		} else {
			return false // Letter not found in any tab
		}

	case tcell.KeyEnter:
		t.Selected = t.Index
		t.Emit("activate", t.Selected)

	default:
		return false // Key not handled
	}

	// Only refresh and emit event if the tab actually changed
	if t.Index != oldIndex {
		t.Refresh()
		t.Emit("change", t.Index)
		return true
	}

	return false
}

// handleLetterNavigation implements first-letter tab navigation with cycling behavior.
// This method searches for tabs that start with the specified letter and cycles through
// them in order, wrapping around when reaching the end of matching tabs.
//
// Navigation algorithm:
//  1. Convert the input letter to lowercase for case-insensitive matching
//  2. Start searching from the tab after the current one
//  3. Find the next tab that starts with the specified letter
//  4. If no match is found after the current tab, wrap around and search from the beginning
//  5. If still no match, return false (letter not found in any tab)
//
// Cycling behavior:
//   - Always moves to the NEXT tab with the matching letter
//   - If currently on a tab starting with the letter, moves to the next matching tab
//   - Wraps around to the first matching tab after reaching the last one
//   - Provides consistent cycling behavior regardless of current position
//
// Case handling:
//   - Input letter is converted to lowercase for comparison
//   - Tab names are converted to lowercase for comparison
//   - Supports both uppercase and lowercase input
//
// Parameters:
//   - letter: The rune (character) to search for as the first letter of tab names
//
// Returns:
//   - bool: true if a matching tab was found and selected, false otherwise
//
// Example with tabs ["First", "Second", "Third", "Fourth", "Fifth"]:
//   - Current: "First", press 'f' → jumps to "Fourth"
//   - Current: "Fourth", press 'f' → jumps to "Fifth"
//   - Current: "Fifth", press 'f' → jumps to "First" (wraps around)
func (t *Tabs) handleLetterNavigation(letter rune) bool {
	if len(t.Tabs) == 0 {
		return false
	}

	// Convert to lowercase for case-insensitive comparison
	targetLetter := rune(letter)
	if targetLetter >= 'A' && targetLetter <= 'Z' {
		targetLetter = targetLetter - 'A' + 'a'
	}

	// Start searching from the next tab after current
	startIndex := (t.Index + 1) % len(t.Tabs)

	// First pass: search from current+1 to end
	for i := startIndex; i < len(t.Tabs); i++ {
		if len(t.Tabs[i]) > 0 {
			firstChar := rune(t.Tabs[i][0])
			if firstChar >= 'A' && firstChar <= 'Z' {
				firstChar = firstChar - 'A' + 'a'
			}
			if firstChar == targetLetter {
				t.Index = i
				return true
			}
		}
	}

	// Second pass: search from beginning to current (wraparound)
	for i := range startIndex {
		if len(t.Tabs[i]) > 0 {
			firstChar := rune(t.Tabs[i][0])
			if firstChar >= 'A' && firstChar <= 'Z' {
				firstChar = firstChar - 'A' + 'a'
			}
			if firstChar == targetLetter {
				t.Index = i
				return true
			}
		}
	}

	// No matching tab found
	return false
}

// SetSelected sets the selected tab by index and updates the highlighted index to match.
//
// This method provides programmatic control over tab selection, equivalent to
// the user navigating to a tab and pressing Enter. It performs bounds checking
// and emits appropriate events.
//
// Parameters:
//   - index: The zero-based index of the tab to select
//
// Returns:
//   - bool: true if the index was valid and selection changed, false otherwise
//
// Example:
//
//	tabs := NewTabs("nav")
//	tabs.Add("Home")
//	tabs.Add("Settings")
//	success := tabs.SetSelected(1) // Select "Settings" tab
func (t *Tabs) SetSelected(index int) bool {
	if index < 0 || index >= len(t.Tabs) {
		return false
	}
	
	oldSelected := t.Selected
	oldIndex := t.Index
	
	t.Selected = index
	t.Index = index
	
	// Emit events if anything changed
	if t.Index != oldIndex {
		t.Emit("change", t.Index)
	}
	if t.Selected != oldSelected {
		t.Emit("activate", t.Selected)
	}
	
	t.Refresh()
	return true
}

// GetSelected returns the index of the currently selected tab.
//
// Returns:
//   - int: The zero-based index of the selected tab, or -1 if no tabs exist
func (t *Tabs) GetSelected() int {
	if len(t.Tabs) == 0 {
		return -1
	}
	return t.Selected
}

// GetSelectedTitle returns the title of the currently selected tab.
//
// Returns:
//   - string: The title of the selected tab, or empty string if no tabs exist
//   - bool: true if a valid tab is selected, false otherwise
func (t *Tabs) GetSelectedTitle() (string, bool) {
	if t.Selected < 0 || t.Selected >= len(t.Tabs) {
		return "", false
	}
	return t.Tabs[t.Selected], true
}

// GetHighlighted returns the index of the currently highlighted tab.
//
// The highlighted tab is the one that would be selected if the user pressed Enter.
// This may differ from the selected tab during keyboard navigation.
//
// Returns:
//   - int: The zero-based index of the highlighted tab, or -1 if no tabs exist
func (t *Tabs) GetHighlighted() int {
	if len(t.Tabs) == 0 {
		return -1
	}
	return t.Index
}

// Count returns the total number of tabs in the widget.
//
// Returns:
//   - int: The number of tabs
func (t *Tabs) Count() int {
	return len(t.Tabs)
}

// Clear removes all tabs from the widget and resets indices to 0.
//
// This method will emit "change" and "activate" events if there were
// previously any tabs, since the indices change from their previous
// values to 0.
func (t *Tabs) Clear() {
	hadTabs := len(t.Tabs) > 0
	oldSelected := t.Selected
	oldIndex := t.Index
	
	t.Tabs = t.Tabs[:0] // Clear slice but keep capacity
	t.Selected = 0
	t.Index = 0
	
	if hadTabs {
		if t.Index != oldIndex {
			t.Emit("change", t.Index)
		}
		if t.Selected != oldSelected {
			t.Emit("activate", t.Selected)
		}
		t.Refresh()
	}
}

// Remove removes the tab at the specified index.
//
// If the removed tab was selected or highlighted, the indices are adjusted
// to maintain valid state. The selection/highlight will move to the previous
// tab if possible, or to index 0 if removing the first tab.
//
// Parameters:
//   - index: The zero-based index of the tab to remove
//
// Returns:
//   - bool: true if the tab was successfully removed, false if index was invalid
func (t *Tabs) Remove(index int) bool {
	if index < 0 || index >= len(t.Tabs) {
		return false
	}
	
	oldSelected := t.Selected
	oldIndex := t.Index
	
	// Remove the tab
	t.Tabs = append(t.Tabs[:index], t.Tabs[index+1:]...)
	
	// Adjust indices if necessary
	if len(t.Tabs) == 0 {
		t.Selected = 0
		t.Index = 0
	} else {
		// Adjust selected index
		if t.Selected > index {
			t.Selected--
		} else if t.Selected == index {
			// If we removed the selected tab, select the previous one
			// or the first one if we removed index 0
			if index > 0 {
				t.Selected = index - 1
			} else {
				t.Selected = 0
			}
		}
		
		// Adjust highlighted index
		if t.Index > index {
			t.Index--
		} else if t.Index == index {
			// If we removed the highlighted tab, highlight the previous one
			// or the first one if we removed index 0
			if index > 0 {
				t.Index = index - 1
			} else {
				t.Index = 0
			}
		}
		
		// Ensure indices are within bounds
		if t.Selected >= len(t.Tabs) {
			t.Selected = len(t.Tabs) - 1
		}
		if t.Index >= len(t.Tabs) {
			t.Index = len(t.Tabs) - 1
		}
	}
	
	// Emit events if indices changed
	if t.Index != oldIndex {
		t.Emit("change", t.Index)
	}
	if t.Selected != oldSelected {
		t.Emit("activate", t.Selected)
	}
	
	t.Refresh()
	return true
}

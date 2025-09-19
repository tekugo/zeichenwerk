package zeichenwerk

import (
	"github.com/gdamore/tcell/v2"
)

type Tabs struct {
	BaseWidget
	Tabs     []string // Tab names
	Selected int      // Currently selected
	Index    int      // Currently highlighted tab during focus
}

func NewTabs(id string) *Tabs {
	return &Tabs{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Tabs:       make([]string, 0),
		Selected:   0,
		Index:      0,
	}
}

func (t *Tabs) Add(title string) {
	t.Tabs = append(t.Tabs, title)
}

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
	for i := 0; i < startIndex; i++ {
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

func (t *Tabs) Emit(event string, data ...any) bool {
	if t.handlers == nil {
		return false
	}
	handler, found := t.handlers[event]
	if found {
		return handler(t, event, data...)
	}
	return false
}

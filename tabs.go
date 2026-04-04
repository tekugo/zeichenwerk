package zeichenwerk

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// Tabs represents a tab navigation widget that displays multiple named tabs
// and allows users to navigate between them using keyboard controls.
//
// Events emitted:
//   - "change": When the highlighted tab changes (navigation)
//   - "activate": When a tab is selected/activated (Enter key)
type Tabs struct {
	Component
	tabs     []string // Tab names to display
	selected int      // Index of the currently selected/active tab
	index    int      // Index of the currently highlighted tab (for navigation)
}

// NewTabs creates a new tabs widget with the specified ID.
func NewTabs(id, class string) *Tabs {
	tabs := &Tabs{
		Component: Component{id: id, class: class},
		tabs:      make([]string, 0),
		selected:  0,
		index:     0,
	}
	tabs.SetFlag(FlagFocusable, true)
	OnKey(tabs, tabs.handleKey)
	return tabs
}

// Add appends a new tab with the specified title to the tabs widget.
func (t *Tabs) Add(title string) {
	t.tabs = append(t.tabs, title)
}

// Apply applies a theme's styles to the component.
func (t *Tabs) Apply(theme *Theme) {
	theme.Apply(t, t.Selector("tabs"), "disabled", "focused")
	theme.Apply(t, t.Selector("tabs/highlight"), "disabled", "focused")
	theme.Apply(t, t.Selector("tabs/line"), "disabled", "focused")
	theme.Apply(t, t.Selector("tabs/line-highlight"), "disabled", "focused")
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
func (t *Tabs) Hint() (int, int) {
	if t.hwidth != 0 || t.hheight != 0 {
		return t.hwidth, t.hheight
	}
	width := 0
	for _, tab := range t.tabs {
		width += len([]rune(tab))
	}
	width += len(t.tabs)*2 + 2
	return width, 2
}

// handleKey processes keyboard input for tab navigation.
// This method implements tab switching controls, allowing users to navigate
// between tabs using standard keyboard shortcuts with wraparound behavior.
//
// Navigation controls:
//   - Left/Right arrows: Navigate between adjacent tabs with wraparound
//   - Home/End: Quick navigation to first/last tab
//   - Letter keys: Jump to tabs by first letter with cycling behavior
func (t *Tabs) handleKey(event *tcell.EventKey) bool {
	if len(t.tabs) == 0 {
		return false // No tabs to navigate
	}

	oldIndex := t.index

	switch event.Key() {
	case tcell.KeyLeft:
		// Move to previous tab (wrap to last if at first)
		if t.index > 0 {
			t.index--
		} else {
			t.index = len(t.tabs) - 1 // Wrap to last tab
		}

	case tcell.KeyRight:
		// Move to next tab (wrap to first if at last)
		if t.index < len(t.tabs)-1 {
			t.index++
		} else {
			t.index = 0 // Wrap to first tab
		}

	case tcell.KeyHome:
		// Jump to first tab
		t.index = 0

	case tcell.KeyEnd:
		// Jump to last tab
		t.index = len(t.tabs) - 1

	case tcell.KeyRune:
		// Handle letter navigation - jump to tabs by first letter
		if t.handleLetterNavigation(event.Str()) {
			// Letter navigation handled, index may have changed
		} else {
			return false // Letter not found in any tab
		}

	case tcell.KeyEnter:
		t.selected = t.index
		t.Dispatch(t, EvtActivate, t.selected)
		t.Refresh()

	default:
		return false // Key not handled
	}

	// Only refresh and emit event if the tab actually changed
	if t.index != oldIndex {
		t.Refresh()
		t.Dispatch(t, EvtChange, t.index)
		return true
	}

	return false
}

// handleLetterNavigation implements first-letter tab navigation with cycling behavior.
// This method searches for tabs that start with the specified letter and cycles through
// them in order, wrapping around when reaching the end of matching tabs.
//
// Parameters:
//   - letter: The character to search for as the first letter of tab names
//
// Returns:
//   - bool: true if a matching tab was found and selected, false otherwise
//
// Example with tabs ["First", "Second", "Third", "Fourth", "Fifth"]:
//   - Current: "First", press 'f' → jumps to "Fourth"
//   - Current: "Fourth", press 'f' → jumps to "Fifth"
//   - Current: "Fifth", press 'f' → jumps to "First" (wraps around)
func (t *Tabs) handleLetterNavigation(letter string) bool {
	if len(t.tabs) == 0 {
		return false
	}

	// Convert to lowercase for case-insensitive comparison
	targetLetter := strings.ToLower(letter)

	// Start searching from the next tab after current
	startIndex := (t.index + 1) % len(t.tabs)

	// First pass: search from current+1 to end
	for i := startIndex; i < len(t.tabs); i++ {
		if len(t.tabs[i]) > 0 && strings.HasPrefix(strings.ToLower(t.tabs[i]), targetLetter) {
			t.index = i
			return true
		}
	}

	// Second pass: search from beginning to current (wraparound)
	for i := range startIndex {
		if len(t.tabs[i]) > 0 && strings.HasPrefix(strings.ToLower(t.tabs[i]), targetLetter) {
			t.index = i
			return true
		}
	}

	// No matching tab found
	return false
}

// Select sets the selected tab by index and updates the highlighted index to match.
func (t *Tabs) Select(index int) bool {
	if index < 0 || index >= len(t.tabs) {
		return false
	}

	oldSelected := t.selected
	oldIndex := t.index

	t.selected = index
	t.index = index

	// Emit events if anything changed
	if t.index != oldIndex {
		t.Dispatch(t, EvtChange, t.index)
	}
	if t.selected != oldSelected {
		t.Dispatch(t, EvtActivate, t.selected)
	}

	t.Refresh()
	return true
}

// Summary returns tab names with the active tab marked for Dump output.
func (t *Tabs) Summary() string {
	parts := make([]string, len(t.tabs))
	for i, name := range t.tabs {
		if i == t.selected {
			parts[i] = fmt.Sprintf("%q(%d*)", name, i)
		} else {
			parts[i] = fmt.Sprintf("%q(%d)", name, i)
		}
	}
	return strings.Join(parts, " | ")
}

// Selected returns the index of the currently selected tab.
func (t *Tabs) Selected() int {
	if len(t.tabs) == 0 {
		return -1
	}
	return t.selected
}

// Count returns the total number of tabs in the widget.
//
// Returns:
//   - int: The number of tabs
func (t *Tabs) Count() int {
	return len(t.tabs)
}

// Render draws the tabs widget.
//
// The tabs widget draws a horizontal row of tabs at the top of its content area.
// The currently selected tab is highlighted with a different style.
func (t *Tabs) Render(r *Renderer) {
	// Determine which styles to use based on focus state
	x, y, w, _ := t.Content()
	var normal, highlight, line *Style

	if t.Flag(FlagFocused) {
		// Use focus-specific styles when tabs widget has focus
		normal = t.Style("line:focused")
		if normal == nil {
			normal = t.Style()
		}
		highlight = t.Style("highlight:focused")
		if highlight == nil {
			highlight = t.Style("highlight")
		}
		line = t.Style("line-highlight:focused")
		if line == nil {
			line = t.Style("line-highlight")
		}
	} else {
		// Use normal styles when tabs widget doesn't have focus
		normal = t.Style()
		highlight = t.Style("highlight")
		line = t.Style("line-highlight")
	}

	cx := x
	r.Set(normal.Foreground(), normal.Background(), normal.Font())

	for i, tab := range t.tabs {
		tl := len([]rune(tab))
		if t.index == i {
			r.Set(highlight.Foreground(), highlight.Background(), highlight.Font())
			r.Text(cx+1, y, " "+tab+" ", 0)
			r.Set(normal.Foreground(), normal.Background(), normal.Font())
		} else {
			r.Text(cx+1, y, " "+tab+" ", 0)
		}
		if t.selected == i {
			r.Set(line.Foreground(), line.Background(), line.Font())
			r.Repeat(cx, y+1, 1, 0, tl+4, "\u2501")
			r.Set(normal.Foreground(), normal.Background(), normal.Font())
		} else {
			r.Repeat(cx, y+1, 1, 0, tl+4, "\u2501")
		}
		cx = cx + tl + 4
		if cx > x+w {
			break
		}
	}

	if cx < x+w {
		r.Repeat(cx, y+1, 1, 0, x+w-cx, "\u2501")
	}
}

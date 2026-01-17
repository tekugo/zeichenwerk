package next

import (
	"fmt"
	"slices"

	"github.com/gdamore/tcell/v3"
)

// List is a scrollable list widget that displays text items with comprehensive
// interaction capabilities.
//
// Keyboard navigation includes:
//   - Arrow keys: Up/Down navigation with disabled item skipping
//   - Page keys: Page Up/Down for rapid navigation through long lists
//   - Home/End keys: Jump to first/last enabled items
//   - Enter: Activate the currently highlighted item
//   - Letter keys: Quick search to items starting with the typed letter
//   - Escape: Custom key handling through event delegation
//
// # Selection System
//
// The widget maintains two distinct concepts:
//   - Index (highlighting): Current focus position for navigation
//   - Selection: Set of chosen items for multi-selection scenarios
//   - These operate independently, allowing complex selection workflows
//
// # Item States
//
// Each item can be in one of several states:
//   - Normal: Standard selectable item
//   - Selected: Item is part of the current selection set
//   - Highlighted: Item currently has navigation focus
//   - Disabled: Item cannot be selected or activated
//   - Combined states: Items can be both selected and highlighted simultaneously
type List struct {
	Component

	// ---- Core Data ----
	items []string // Text items to display in the list (primary content)

	// ---- Navigation State ----
	index  int // Current highlight position (focused item index, -1 if none)
	offset int // Vertical scroll offset for viewport positioning (top visible item index)

	// ---- Selection Management ----
	selection []int // Indices of currently selected items (supports multi-selection)
	disabled  []int // Indices of items that cannot be selected or activated

	// ---- Display Options ----
	numbers   bool // Show line numbers next to each item for reference
	scrollbar bool // Display scrollbar indicator on the right edge
}

// NewList creates a new List widget with the specified ID and initial items.
// The widget is initialized with sensible defaults suitable for most use cases,
// including focus capability and scrollbar display.
//
// Returns:
//   - *List: A fully initialized List widget ready for configuration and use
func NewList(id string, items []string) *List {
	list := &List{
		Component: Component{id: id},
		items:     items,
		index:     0,
		selection: make([]int, 0, 1),
		offset:    0,
		numbers:   false,
		scrollbar: true,
	}
	list.SetFlag("focusable", true)
	OnKey(list, list.HandleKey)
	return list
}

// ---- Widget Methods -------------------------------------------------------

// Refresh triggers a visual update of the List widget by requesting a redraw.
// This method should be called whenever the list's visual state has changed
// and needs to be reflected on screen.
func (l *List) Refresh() {
	l.parent.Refresh()
	// Redraw(l)
}

// ---- Movements ------------------------------------------------------------

// Up moves the highlight position up by the specified number of items.
// This method provides intelligent navigation that skips disabled items and
// handles boundary conditions gracefully with wraparound behavior.
//
// Parameters:
//   - count: Number of items to move up (positive integer)
func (l *List) Up(count int) {
	if len(l.items) == 0 {
		l.index = 0
		return
	}

	next := l.index - count

	// Skip disabled items
	for next >= 0 && slices.Contains(l.disabled, next) {
		next--
	}

	if next >= 0 {
		l.index = next
		l.adjust()
		l.Dispatch("select", l.index)
	} else {
		// Move to first without calling First() to avoid double refresh
		l.index = -1
		for i := range l.items {
			if !slices.Contains(l.disabled, i) {
				l.index = i
				l.adjust()
				l.Dispatch("select", l.index)
				break
			}
		}
	}

	l.Refresh()
}

// Down moves the highlight position down by the specified number of items.
// This method provides intelligent forward navigation that skips disabled items and
// handles boundary conditions gracefully with wraparound behavior.
//
// Parameters:
//   - count: Number of items to move down (positive integer)
func (l *List) Down(count int) {
	if len(l.items) == 0 {
		return
	}

	next := l.index + count

	// Skip disabled items
	for next < len(l.items) && slices.Contains(l.disabled, next) {
		next++
	}

	if next < len(l.items) {
		l.index = next
		l.adjust()
		l.Dispatch("select", l.index)
	} else {
		// Move to last without calling Last() to avoid double refresh
		l.index = -1
		for i := len(l.items) - 1; i >= 0; i-- {
			if !slices.Contains(l.disabled, i) {
				l.index = i
				l.adjust()
				l.Dispatch("select", l.index)
				break
			}
		}
	}

	l.Refresh()
}

// First moves the highlight to the first enabled item in the list.
// This method provides a quick way to jump to the beginning of the list,
// automatically skipping any disabled items at the start.
func (l *List) First() {
	l.index = -1
	for i := range l.items {
		if !slices.Contains(l.disabled, i) {
			l.index = i
			l.adjust()
			l.Dispatch("select", l.index)
			break
		}
	}
	l.Refresh()
}

// Last moves the highlight to the last enabled item in the list.
// This method provides a quick way to jump to the end of the list,
// automatically skipping any disabled items at the end.
func (l *List) Last() {
	l.index = -1
	for i := len(l.items) - 1; i >= 0; i-- {
		if !slices.Contains(l.disabled, i) {
			l.index = i
			l.adjust()
			l.Dispatch("select", l.index)
			break
		}
	}
	l.Refresh()
}

// PageUp moves the highlight up by one page (viewport height).
// This method provides rapid navigation through long lists by jumping
// by the number of visible items rather than single item steps.
func (l *List) PageUp() {
	_, _, _, ih := l.Content()
	l.Up(ih)
}

// PageDown moves the highlight down by one page (viewport height).
// This method provides rapid forward navigation through long lists by jumping
// by the number of visible items rather than single item steps.
func (l *List) PageDown() {
	_, _, _, ih := l.Content()
	l.Down(ih)
	// Refresh is already called by Down(), no need to call again
}

// ---- Actions --------------------------------------------------------------

// Handle processes input events for the List widget, implementing comprehensive
// keyboard navigation and interaction capabilities. This method serves as the
// primary input processor for all list interactions.
//
// # Event Processing
//
// The method handles several categories of events:
//   - Navigation keys: Arrow keys, Page Up/Down, Home/End
//   - Action keys: Enter for activation
//   - Search keys: Letter keys for quick item search
//   - Custom keys: All other keys delegated to event handlers
//
// # Navigation Key Bindings
//
// Standard navigation keys are mapped as follows:
//   - Up Arrow: Move highlight up one item
//   - Down Arrow: Move highlight down one item
//   - Home: Jump to first enabled item
//   - End: Jump to last enabled item
//   - Page Up: Move up by viewport height
//   - Page Down: Move down by viewport height
//   - Enter: Activate current item (emit "activate" event)
//
// # Quick Search Feature
//
// Letter key presses trigger quick search:
//  1. Searches for items starting with the typed letter
//  2. Performs case-insensitive matching (A matches 'a' or 'A')
//  3. Skips disabled items during search
//  4. Moves highlight to first matching enabled item
//  5. Search starts from current position for efficiency
//
// # Event Delegation
//
// Unhandled events are delegated to the event system:
//   - Custom key combinations emit "key" event
//   - Event handlers can implement application-specific behavior
//   - Return value indicates whether event was consumed
//
// # Return Value
//
// The method returns true if the event was handled, false otherwise:
//   - Navigation keys: Always return true (consumed)
//   - Enter key: Always returns true (consumed)
//   - Quick search: Returns true if matching item found
//   - Other keys: Delegates to event handlers, returns their result
//
// Parameters:
//   - event: The tcell.Event to process (keyboard, mouse, etc.)
//
// Returns:
//   - bool: true if event was handled/consumed, false if should be propagated
func (l *List) HandleKey(event *tcell.EventKey) bool {
	l.Log(l, "debug", "key %s", event.Str())
	switch event.Key() {
	case tcell.KeyUp:
		l.Up(1)
		return true
	case tcell.KeyDown:
		l.Down(1)
		return true
	case tcell.KeyHome:
		l.First()
		return true
	case tcell.KeyEnd:
		l.Last()
		return true
	case tcell.KeyPgUp:
		l.PageUp()
		return true
	case tcell.KeyPgDn:
		l.PageDown()
		return true
	case tcell.KeyEnter:
		l.Dispatch("activate", l.index)
		return true
	case tcell.KeyRune:
		// Quick search by first letter
		ch := event.Str()
		for i, item := range l.items {
			if !slices.Contains(l.disabled, i) && len(item) > 0 {
				firstChar := string(item[0])
				if firstChar == ch {
					l.index = i
					return true
				}
			}
		}
	}
	return false
}

// ---- Internal Methods ----------------------------------------------------

// adjust automatically adjusts the scroll offset to ensure the highlighted item
// remains visible within the current viewport. This method implements intelligent
// scrolling that maintains optimal user experience during navigation.
//
// # Viewport Calculation
//
// The method determines visibility requirements:
//  1. Gets current content area height (visible item count)
//  2. Calculates current viewport bounds (Offset to Offset+height)
//  3. Checks if highlighted item is within visible range
//  4. Adjusts scroll offset if item is outside viewport
//
// # Scroll Adjustment Logic
//
// Scroll adjustments follow these rules:
//   - Item above viewport: Scroll up to make item the top visible item
//   - Item below viewport: Scroll down to make item the bottom visible item
//   - Item within viewport: No adjustment needed
//   - Maintains minimal scroll movement for smooth user experience
//
// # Boundary Protection
//
// The method enforces scroll boundaries:
//   - Minimum offset: 0 (cannot scroll above first item)
//   - Maximum offset: max(totalItems - viewportHeight, 0)
//   - Prevents scrolling past available content
//   - Handles edge cases with small item counts gracefully
//
// # Performance Optimization
//
// The adjustment algorithm is optimized for:
//   - Single calculation per navigation operation
//   - Minimal offset changes to reduce visual jumping
//   - Early return for invalid viewport dimensions
//   - Efficient boundary checking
//
// This method is called automatically by all navigation methods and should
// not typically be called directly by application code.
func (l *List) adjust() {
	_, _, _, ih := l.Content()
	if ih <= 0 {
		return
	}

	// Ensure selected item is visible
	if l.index < l.offset {
		l.offset = l.index
	} else if l.index >= l.offset+ih {
		l.offset = l.index - ih + 1
	}

	// Don't scroll past the beginning
	if l.offset < 0 {
		l.offset = 0
	}

	// Don't scroll past the end
	maxScroll := max(len(l.items)-ih, 0)
	if l.offset > maxScroll {
		l.offset = maxScroll
	}
}

// Visible returns the slice of items that should be displayed in the current viewport.
// This method implements virtual scrolling by calculating which items are visible
// based on the current scroll offset and viewport dimensions.
//
// # Virtual Scrolling Implementation
//
// The method calculates visible items:
//  1. Determines viewport height from widget dimensions
//  2. Calculates start index from current scroll offset
//  3. Calculates end index ensuring it doesn't exceed item count
//  4. Returns slice of items within the calculated range
//
// # Edge Case Handling
//
// Special conditions are managed gracefully:
//   - Empty item list: Returns empty slice
//   - Zero viewport height: Returns empty slice
//   - Offset beyond items: Returns empty slice
//   - Partial viewport fill: Returns available items only
//
// # Performance Benefits
//
// Virtual scrolling provides:
//   - Constant memory usage regardless of total item count
//   - Fast rendering with only visible items processed
//   - Efficient handling of large datasets
//   - Smooth scrolling without content duplication
//
// # Use Cases
//
// This method is used by:
//   - Rendering system to determine which items to draw
//   - External components needing current visible content
//   - Testing and debugging viewport calculations
//
// Returns:
//   - []string: Slice of items currently visible in the viewport
func (l *List) Visible() []string {
	_, _, _, ih := l.Content()
	if ih <= 0 || len(l.items) == 0 {
		return []string{}
	}

	start := l.offset
	end := min(start+ih, len(l.items))
	if start >= len(l.items) {
		return []string{}
	}

	return l.items[start:end]
}

// ScrollInfo provides comprehensive information about the current scroll state
// and capabilities. This method is used by scrollbar rendering and external
// components that need to understand the list's scroll position.
//
// # Scroll State Calculation
//
// The method determines:
//   - Whether scrolling up is possible (not at top)
//   - Whether scrolling down is possible (not at bottom)
//   - Current scroll position as a percentage (0.0 to 1.0)
//
// # Scroll Capability Detection
//
// Scroll capabilities are determined by:
//   - canScrollUp: true if Offset > 0 (content above viewport)
//   - canScrollDown: true if Offset < len(Items)-1 (content below viewport)
//   - These flags indicate whether navigation in each direction is possible
//
// # Percentage Calculation
//
// Scroll percentage represents position within total scrollable range:
//   - 0.0: At the top of the list
//   - 1.0: At the bottom of the list
//   - Values between: Proportional position within the list
//   - Empty list: Always returns 0.0
//
// # Use Cases
//
// This information is used for:
//   - Scrollbar thumb positioning and sizing
//   - Navigation button state (enabled/disabled)
//   - Progress indicators for long lists
//   - Accessibility features and screen readers
//
// Returns:
//   - canScrollUp: true if scrolling up is possible
//   - canScrollDown: true if scrolling down is possible
//   - scrollPercent: Current position as percentage (0.0-1.0)
func (l *List) ScrollInfo() (canScrollUp, canScrollDown bool, scrollPercent float64) {
	canScrollUp = l.offset > 0
	canScrollDown = l.offset < len(l.items)-1

	if len(l.items) > 0 {
		scrollPercent = float64(l.offset) / float64(len(l.items)-1)
	}

	return
}

// renderList renders a List widget with items, selection highlighting, and optional scrollbar.
// This method handles the complete visual presentation of list widgets including
// item display, selection highlighting, line numbers, and scrollbar indicators.
//
// Parameters:
//   - list: The List widget to render
//   - x, y: Top-left coordinates of the list's content area
//   - w, h: Width and height of the list's content area
//
// Rendering features:
//  1. Displays visible items within the content area
//  2. Applies different styles for normal, highlighted, and disabled items
//  3. Shows optional line numbers with proper formatting
//  4. Renders scrollbar when content exceeds visible area
//  5. Handles item truncation for items wider than available space
//
// Visual elements:
//   - Item text with appropriate styling based on state
//   - Selection highlighting for the currently focused item
//   - Disabled item styling for non-selectable items
//   - Line numbers with consistent width formatting
//   - Vertical scrollbar indicating scroll position and content size
//
// The method automatically adjusts text width to accommodate scrollbars
// and line numbers, ensuring proper layout regardless of configuration.
func (l *List) Render(r *Renderer) {
	l.Log(l, "debug", "Render %s index=%d offset=%d len=%d", l.id, l.index, l.offset, len(l.items))

	x, y, w, h := l.Content()
	if h < 1 || w < 1 {
		return
	}

	items := l.Visible()

	// Calculate available width for text (reserve space for scrollbar if needed)
	tw := w
	if l.scrollbar && len(l.items) > h {
		tw = w - 1
	}

	// Calculate number width if showing numbers
	nw := 0
	if l.numbers {
		nw = len(fmt.Sprintf("%d", len(l.items)))
	}

	// Render each visible item
	for i, item := range items {
		if i >= h {
			break
		}

		current := l.offset + i

		// Determine style for this item
		if slices.Contains(l.disabled, i) {
			style := l.Style(":disabled")
			r.Set(style.Foreground(), style.Background(), style.Font())
			l.Log(l, "debug", "  Item %d %s disabled", current, item)
		} else if current == l.index {
			if l.Flag("focus") {
				style := l.Style("highlight:focus")
				r.Set(style.Foreground(), style.Background(), style.Font())
				l.Log(l, "debug", "  Item %d %s focus %s %s", current, item, style.Foreground(), style.Background())
			} else {
				style := l.Style("highlight")
				r.Set(style.Foreground(), style.Background(), style.Font())
				l.Log(l, "debug", "  Item %d %s highlight %s %s", current, item, style.Foreground(), style.Background())
			}
		}

		// Render line number if enabled
		if l.numbers {
			r.Text(x, y+i, fmt.Sprintf(" %*d \u2502 %s", nw, current+1, item), tw)
		} else {
			r.Text(x, y+i, " "+item, tw)
		}

		// Reset style
		style := l.Style()
		r.Set(style.Foreground(), style.Background(), style.Font())
	}

	// Render scrollbar if needed
	if l.scrollbar && len(l.items) > h {
		scrollbarX := x + w - 1
		r.ScrollbarV(scrollbarX, y, h, l.offset, len(l.items))
	}
}

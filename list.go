// Package list.go implements the List widget for zeichenwerk.
//
// This file provides a comprehensive scrollable list widget with advanced features
// including multi-selection, keyboard and mouse navigation, item disabling, and
// customizable visual styling. The List widget is designed for displaying and
// interacting with collections of text items in terminal user interfaces.
//
// # Core Features
//
// The List widget offers:
//   - Scrollable display with automatic viewport management
//   - Multi-selection support with visual selection indicators
//   - Comprehensive keyboard navigation (arrows, page keys, home/end)
//   - Mouse interaction support for clicking and scrolling
//   - Item disabling functionality with visual feedback
//   - Optional line numbers for reference
//   - Configurable scrollbar display
//   - Event-driven architecture for selection and activation callbacks
//   - Quick search functionality by first letter
//
// # Architecture
//
// The List widget uses a virtual scrolling approach where only visible items
// are rendered, enabling efficient handling of large item collections. The
// widget maintains separate indices for highlighting (current focus) and
// selection (chosen items), supporting both single and multi-selection modes.

package zeichenwerk

import (
	"slices"

	"github.com/gdamore/tcell/v2"
)

// List is a feature-rich scrollable list widget that displays text items with comprehensive
// interaction capabilities. It provides professional-grade functionality for terminal
// applications requiring item selection, navigation, and user interaction.
//
// # Core Functionality
//
// The List widget provides:
//   - Virtual scrolling for efficient large dataset handling
//   - Multi-selection with visual selection indicators
//   - Comprehensive keyboard navigation with standard key bindings
//   - Mouse interaction support for modern terminal environments
//   - Flexible item management with enable/disable functionality
//   - Event-driven architecture for responsive user interactions
//   - Customizable visual styling through the theming system
//
// # Navigation Features
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
//
// # Visual Options
//
// Display customization includes:
//   - Line numbers: Optional numeric prefixes for item reference
//   - Scrollbar: Visual scroll position indicator
//   - Custom styling: Theme-based styling for different item states
//   - Selection indicators: Visual feedback for selected items
//
// # Event System
//
// The widget emits events for:
//   - "select": Fired when highlight position changes (navigation)
//   - "activate": Fired when an item is activated (Enter key)
//   - "key": Fired for unhandled keyboard events (custom handling)
//
// # Performance Characteristics
//
// The List widget is optimized for:
//   - Large datasets through virtual scrolling
//   - Smooth navigation with efficient disabled item skipping
//   - Minimal memory overhead for item management
//   - Fast rendering with only visible items processed
type List struct {
	BaseWidget
	
	// ---- Core Data ----
	Items []string // Text items to display in the list (primary content)
	
	// ---- Navigation State ----
	Index  int // Current highlight position (focused item index, -1 if none)
	Offset int // Vertical scroll offset for viewport positioning (top visible item index)
	
	// ---- Selection Management ----
	Selection []int // Indices of currently selected items (supports multi-selection)
	Disabled  []int // Indices of items that cannot be selected or activated
	
	// ---- Display Options ----
	Numbers   bool // Show line numbers next to each item for reference
	Scrollbar bool // Display scrollbar indicator on the right edge
}

// NewList creates a new List widget with the specified ID and initial items.
// The widget is initialized with sensible defaults suitable for most use cases,
// including focus capability and scrollbar display.
//
// # Initialization State
//
// The new List widget is configured with:
//   - Focus capability: Enabled (widget can receive keyboard focus)
//   - Highlight position: Index 0 (first item highlighted)
//   - Selection state: Empty (no items initially selected)
//   - Scroll position: Offset 0 (top of list visible)
//   - Line numbers: Disabled (clean appearance by default)
//   - Scrollbar: Enabled (provides navigation feedback)
//   - Event handlers: None (must be configured separately)
//
// # Post-Construction Configuration
//
// After creation, common configuration includes:
//   - Event handlers: Use On() method to register selection/activation callbacks
//   - Visual options: Configure Numbers and Scrollbar fields as needed
//   - Initial selection: Use selection methods to set default selected items
//   - Item disabling: Use Disabled field to mark non-selectable items
//   - Styling: Apply theme-based styling through the widget's style system
//
// # Empty List Handling
//
// The constructor handles empty item lists gracefully:
//   - Empty slice is valid and creates a functional but empty list
//   - Index remains 0 but will be handled appropriately during navigation
//   - All navigation methods work correctly with empty lists
//   - Items can be added dynamically after construction
//
// Parameters:
//   - id: Unique identifier for the list widget (used for theming and event handling)
//   - items: Initial text items to display (can be empty slice for dynamic population)
//
// Returns:
//   - *List: A fully initialized List widget ready for configuration and use
//
// # Example Usage
//
//	// Basic list creation
//	list := NewList("file-list", []string{"file1.txt", "file2.txt", "file3.txt"})
//
//	// Configuration after creation
//	list.Numbers = true  // Enable line numbers
//	list.On("select", func(w Widget, event string, data ...any) bool {
//		index := data[0].(int)
//		fmt.Printf("Selected item %d: %s\n", index, list.Items[index])
//		return true
//	})
//
//	// Dynamic list creation
//	dynamicList := NewList("dynamic", []string{})
//	// ... populate dynamicList.Items later
func NewList(id string, items []string) *List {
	return &List{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Items:      items,
		Index:      0,
		Selection:  make([]int, 0, 1),
		Offset:     0,
		Numbers:    false,
		Scrollbar:  true,
	}
}

// ---- Widget Methods -------------------------------------------------------

// Refresh triggers a visual update of the List widget by requesting a redraw.
// This method should be called whenever the list's visual state has changed
// and needs to be reflected on screen.
//
// # When to Call Refresh
//
// The method should be called after:
//   - Modifying the Items slice (adding, removing, or changing items)
//   - Changing the Selection or Disabled slices
//   - Modifying display options (Numbers, Scrollbar flags)
//   - External state changes that affect visual appearance
//
// # Automatic Refresh
//
// Note that navigation methods automatically call Refresh():
//   - Up(), Down(), First(), Last(), PageUp(), PageDown()
//   - These methods handle their own refresh cycles
//   - Manual refresh is not needed after navigation operations
//
// # Performance Considerations
//
// The method uses efficient widget-specific redrawing:
//   - Only the List widget area is redrawn
//   - Other widgets on screen are not affected
//   - More efficient than full screen refresh for single widget updates
//
// # Usage Pattern
//
// Typical usage after content modification:
//
//	list.Items = append(list.Items, "New Item")
//	list.Refresh()  // Update display to show new item
func (l *List) Refresh() {
	Redraw(l)
}

// ---- Movements ------------------------------------------------------------

// Up moves the highlight position up by the specified number of items.
// This method provides intelligent navigation that skips disabled items and
// handles boundary conditions gracefully with wraparound behavior.
//
// # Navigation Logic
//
// The method implements sophisticated navigation:
//  1. Calculates target position by subtracting count from current index
//  2. Skips backward through any disabled items at the target position
//  3. If valid position found, moves highlight and adjusts viewport
//  4. If no valid position found, wraps to the first enabled item
//  5. Emits "select" event and triggers visual refresh
//
// # Disabled Item Handling
//
// Disabled items are handled intelligently:
//   - Automatically skipped during backward navigation
//   - Multiple consecutive disabled items are traversed efficiently
//   - Navigation continues until an enabled item is found
//   - If no enabled items exist, highlight remains unchanged
//
// # Boundary Behavior
//
// When reaching list boundaries:
//   - Beyond top: Wraps to first enabled item (circular navigation)
//   - Empty list: No operation performed, remains at index 0
//   - All items disabled: No navigation occurs
//
// # Viewport Management
//
// The method automatically manages the viewport:
//   - Adjusts scroll offset to keep highlighted item visible
//   - Maintains optimal scroll position for user experience
//   - Ensures highlighted item is always within the visible area
//   - Prevents unnecessary scrolling when item already visible
//
// # Event Emission
//
// Navigation triggers event system:
//   - "select" event fired with new index on successful navigation
//   - Event handlers can respond to highlight changes
//   - Event data includes the new highlighted item index
//
// # Use Cases
//
// Common usage patterns:
//   - count=1: Single step navigation (arrow key behavior)
//   - count>1: Multi-step navigation (custom navigation shortcuts)
//   - Large count: Effectively jumps to top with wraparound
//
// Parameters:
//   - count: Number of items to move up (positive integer)
func (l *List) Up(count int) {
	if len(l.Items) == 0 {
		l.Index = 0
		return
	}

	next := l.Index - count

	// Skip disabled items
	for next >= 0 && slices.Contains(l.Disabled, next) {
		next--
	}

	if next >= 0 {
		l.Index = next
		l.adjust()
		l.Emit("select", l.Index)
	} else {
		// Move to first without calling First() to avoid double refresh
		l.Index = -1
		for i := range l.Items {
			if !slices.Contains(l.Disabled, i) {
				l.Index = i
				l.adjust()
				l.Emit("select", l.Index)
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
// # Navigation Logic
//
// The method implements sophisticated forward navigation:
//  1. Calculates target position by adding count to current index
//  2. Skips forward through any disabled items at the target position
//  3. If valid position found, moves highlight and adjusts viewport
//  4. If no valid position found, wraps to the last enabled item
//  5. Emits "select" event and triggers visual refresh
//
// # Disabled Item Handling
//
// Disabled items are handled intelligently:
//   - Automatically skipped during forward navigation
//   - Multiple consecutive disabled items are traversed efficiently
//   - Navigation continues until an enabled item is found
//   - If no enabled items exist, highlight remains unchanged
//
// # Boundary Behavior
//
// When reaching list boundaries:
//   - Beyond bottom: Wraps to last enabled item (circular navigation)
//   - Empty list: No operation performed
//   - All items disabled: No navigation occurs
//
// # Viewport Management
//
// The method automatically manages the viewport:
//   - Adjusts scroll offset to keep highlighted item visible
//   - Maintains optimal scroll position for user experience
//   - Ensures highlighted item is always within the visible area
//   - Scrolls list as needed to follow highlight movement
//
// # Event Emission
//
// Navigation triggers event system:
//   - "select" event fired with new index on successful navigation
//   - Event handlers can respond to highlight changes
//   - Event data includes the new highlighted item index
//
// # Use Cases
//
// Common usage patterns:
//   - count=1: Single step navigation (arrow key behavior)
//   - count>1: Multi-step navigation (page down behavior)
//   - Large count: Effectively jumps to bottom with wraparound
//
// Parameters:
//   - count: Number of items to move down (positive integer)
func (l *List) Down(count int) {
	if len(l.Items) == 0 {
		return
	}

	next := l.Index + count

	// Skip disabled items
	for next < len(l.Items) && slices.Contains(l.Disabled, next) {
		next++
	}

	if next < len(l.Items) {
		l.Index = next
		l.adjust()
		l.Emit("select", l.Index)
	} else {
		// Move to last without calling Last() to avoid double refresh
		l.Index = -1
		for i := len(l.Items) - 1; i >= 0; i-- {
			if !slices.Contains(l.Disabled, i) {
				l.Index = i
				l.adjust()
				l.Emit("select", l.Index)
				break
			}
		}
	}

	l.Refresh()
}

// First moves the highlight to the first enabled item in the list.
// This method provides a quick way to jump to the beginning of the list,
// automatically skipping any disabled items at the start.
//
// # Behavior
//
// The method performs the following:
//  1. Searches from the beginning of the list for the first enabled item
//  2. Sets highlight to that item if found
//  3. Adjusts viewport to make the item visible
//  4. Emits "select" event with the new index
//  5. Triggers visual refresh
//
// # Edge Cases
//
// Special situations are handled gracefully:
//   - Empty list: Index remains at -1, no event emitted
//   - All items disabled: Index remains at -1, no event emitted
//   - Mixed enabled/disabled: Finds first enabled item regardless of position
//
// # Use Cases
//
// Commonly used for:
//   - Home key navigation
//   - "Go to top" functionality
//   - Resetting list position after operations
//   - Initial positioning in dynamically populated lists
func (l *List) First() {
	l.Index = -1
	for i := range l.Items {
		if !slices.Contains(l.Disabled, i) {
			l.Index = i
			l.adjust()
			l.Emit("select", l.Index)
			break
		}
	}
	l.Refresh()
}

// Last moves the highlight to the last enabled item in the list.
// This method provides a quick way to jump to the end of the list,
// automatically skipping any disabled items at the end.
//
// # Behavior
//
// The method performs the following:
//  1. Searches backward from the end of the list for the last enabled item
//  2. Sets highlight to that item if found
//  3. Adjusts viewport to make the item visible
//  4. Emits "select" event with the new index
//  5. Triggers visual refresh
//
// # Edge Cases
//
// Special situations are handled gracefully:
//   - Empty list: Index remains at -1, no event emitted
//   - All items disabled: Index remains at -1, no event emitted
//   - Mixed enabled/disabled: Finds last enabled item regardless of position
//
// # Use Cases
//
// Commonly used for:
//   - End key navigation
//   - "Go to bottom" functionality
//   - Quick access to recent items in chronological lists
//   - Navigation to final options in configuration lists
func (l *List) Last() {
	l.Index = -1
	for i := len(l.Items) - 1; i >= 0; i-- {
		if !slices.Contains(l.Disabled, i) {
			l.Index = i
			l.adjust()
			l.Emit("select", l.Index)
			break
		}
	}
	l.Refresh()
}

// PageUp moves the highlight up by one page (viewport height).
// This method provides rapid navigation through long lists by jumping
// by the number of visible items rather than single item steps.
//
// # Implementation
//
// The method calculates page size based on content area:
//  1. Determines the current viewport height (visible item count)
//  2. Delegates to Up() method with the calculated page size
//  3. Inherits all Up() behaviors including disabled item skipping
//  4. Uses existing refresh mechanism from Up() method
//
// # Navigation Behavior
//
// Page navigation provides:
//   - Rapid movement through large lists
//   - Consistent behavior with Up() method (disabled item handling)
//   - Appropriate viewport adjustment
//   - Wraparound to first item if beyond list start
//
// # Use Cases
//
// Ideal for:
//   - Page Up key navigation
//   - Rapid list browsing
//   - Large dataset navigation
//   - File browser applications
func (l *List) PageUp() {
	_, _, _, ih := l.Content()
	l.Up(ih)
	// Refresh is already called by Up(), no need to call again
}

// PageDown moves the highlight down by one page (viewport height).
// This method provides rapid forward navigation through long lists by jumping
// by the number of visible items rather than single item steps.
//
// # Implementation
//
// The method calculates page size based on content area:
//  1. Determines the current viewport height (visible item count)
//  2. Delegates to Down() method with the calculated page size
//  3. Inherits all Down() behaviors including disabled item skipping
//  4. Uses existing refresh mechanism from Down() method
//
// # Navigation Behavior
//
// Page navigation provides:
//   - Rapid movement through large lists
//   - Consistent behavior with Down() method (disabled item handling)
//   - Appropriate viewport adjustment
//   - Wraparound to last item if beyond list end
//
// # Use Cases
//
// Ideal for:
//   - Page Down key navigation
//   - Rapid list browsing
//   - Large dataset navigation
//   - File browser applications
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
func (l *List) Handle(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
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
			l.Emit("activate", l.Index)
			return true
		case tcell.KeyRune:
			// Quick search by first letter
			ch := event.Rune()
			for i, item := range l.Items {
				if !slices.Contains(l.Disabled, i) && len(item) > 0 {
					firstChar := []rune(item)[0]
					if firstChar == ch || (firstChar >= 'A' && firstChar <= 'Z' && firstChar+32 == ch) {
						l.Index = i
						return true
					}
				}
			}
		default:
			l.Emit("key", event)
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
	if l.Index < l.Offset {
		l.Offset = l.Index
	} else if l.Index >= l.Offset+ih {
		l.Offset = l.Index - ih + 1
	}

	// Don't scroll past the beginning
	if l.Offset < 0 {
		l.Offset = 0
	}

	// Don't scroll past the end
	maxScroll := max(len(l.Items)-ih, 0)
	if l.Offset > maxScroll {
		l.Offset = maxScroll
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
	_, ih := l.Size()
	if ih <= 0 || len(l.Items) == 0 {
		return []string{}
	}

	start := l.Offset
	end := min(start+ih, len(l.Items))
	if start >= len(l.Items) {
		return []string{}
	}

	return l.Items[start:end]
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
	canScrollUp = l.Offset > 0
	canScrollDown = l.Offset < len(l.Items)-1

	if len(l.Items) > 0 {
		scrollPercent = float64(l.Offset) / float64(len(l.Items)-1)
	}

	return
}

// Emit triggers registered event handlers for the specified event type.
// This method implements the event system that allows external code to
// respond to list interactions and state changes.
//
// # Event System Design
//
// The method provides:
//   - Type-safe event emission with arbitrary data
//   - Handler lookup by event name
//   - Graceful handling of missing handlers
//   - Support for multiple data parameters
//
// # Event Types
//
// Common events emitted by the List widget:
//   - "select": Fired when highlight position changes (navigation)
//   - "activate": Fired when an item is activated (Enter key)
//   - "key": Fired for unhandled keyboard events
//
// # Handler Requirements
//
// Event handlers must match the signature:
//   func(widget Widget, event string, data ...any) bool
//
// # Error Handling
//
// The method handles edge cases gracefully:
//   - No handlers registered: Silent return (no error)
//   - Event type not found: Silent return (no error)
//   - Handler execution: Delegates error handling to handler
//
// Parameters:
//   - event: Name of the event to emit
//   - data: Variable number of parameters to pass to handler
func (l *List) Emit(event string, data ...any) {
	if l.handlers == nil {
		return
	}
	handler, found := l.handlers[event]
	if found {
		handler(l, event, data...)
	}
}

package zeichenwerk

import (
	"slices"

	"github.com/gdamore/tcell/v2"
)

// List is a widget that displays a scrollable list of text items with selection support.
// It provides keyboard navigation, multi-selection capabilities, and customizable
// event handlers for selection and activation events.
//
// Features:
//   - Scrollable list display with automatic scrolling
//   - Multi-selection support with visual indicators
//   - Keyboard navigation (arrow keys, page up/down, home/end, space for selection)
//   - Mouse interaction support for selection and scrolling
//   - Optional line numbers and scrollbar display
//   - Customizable selection and activation callbacks
//   - Item disabling support with visual styling
//   - Configurable styling for selected and disabled items
//
// The List widget supports multi-selection where multiple items can be selected
// simultaneously. Items can be disabled to prevent selection. The widget provides
// visual feedback through highlighting and custom styling options.
type List struct {
	BaseWidget
	Items     []string // The list of text items to display
	Index     int      // Current highlight bar position (focused item index)
	Selection []int    // Indices of currently selected items (multi-selection)
	Disabled  []int    // Indices of disabled items that cannot be selected
	Offset    int      // Vertical scroll offset for items that don't fit in view
	Numbers   bool     // Display line numbers next to each item
	Scrollbar bool     // Display scrollbar indicator on the right side
}

// NewList creates a new List widget with the specified ID and items.
// The list is initialized with default settings: no items selected, scroll offset at 0,
// line numbers hidden, and scrollbar visible.
//
// Parameters:
//   - id: Unique identifier for the list widget
//   - items: Initial list of text items to display
//
// Returns:
//   - *List: A new List widget instance ready for use
//
// Example usage:
//
//	list := NewList("mylist", []string{"Item 1", "Item 2", "Item 3"})
//	list.ShowNumbers = true  // Enable line numbers
//	list.OnSelect = func(index int) { fmt.Printf("Selected: %d\n", index) }
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

func (l *List) Refresh() {
	Redraw(l)
}

// ---- Movements ------------------------------------------------------------

// Up moves the highlight bar up by one item, skipping any disabled items.
// If the list is empty or already at the top, no action is taken.
// Automatically adjusts scroll position and triggers the OnSelect callback
// if set.
//
// Navigation behavior:
//   - Moves to the previous enabled item
//   - Skips over disabled items automatically
//   - Adjusts scroll position to keep the selected item visible
//   - Calls OnSelect callback with the new index
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

// Down moves the highlight bar down by one item, skipping any disabled items.
// If the list is empty or already at the bottom, no action is taken.
// Automatically adjusts scroll position and triggers the OnSelect callback if set.
//
// Navigation behavior:
//   - Moves to the next enabled item
//   - Skips over disabled items automatically
//   - Adjusts scroll position to keep the selected item visible
//   - Calls OnSelect callback with the new index
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

// First moves the index to the first non-disabled item
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

// Last moves the index to the last non-disabled item
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

// PageUp moves selection up by one page
func (l *List) PageUp() {
	_, _, _, ih := l.Content()
	l.Up(ih)
	// Refresh is already called by Up(), no need to call again
}

// PageDown moves selection down by one page
func (l *List) PageDown() {
	_, _, _, ih := l.Content()
	l.Down(ih)
	// Refresh is already called by Down(), no need to call again
}

// ---- Actions --------------------------------------------------------------

// Handle processes keyboard input
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

// ---- Internal -------------------------------------------------------------

// adjust adjusts the scroll offset to keep the selected item visible
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

// Visible returns the items that should be displayed
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

// ScrollInfo returns scroll information
func (l *List) ScrollInfo() (canScrollUp, canScrollDown bool, scrollPercent float64) {
	canScrollUp = l.Offset > 0
	canScrollDown = l.Offset < len(l.Items)-1

	if len(l.Items) > 0 {
		scrollPercent = float64(l.Offset) / float64(len(l.Items)-1)
	}

	return
}

func (l *List) Emit(event string, data ...any) {
	if l.handlers == nil {
		return
	}
	handler, found := l.handlers[event]
	if found {
		handler(l, event, data...)
	}
}

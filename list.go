package zeichenwerk

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"
)

// flagSearch is an internal flag used to track whether quick-search is enabled.
const flagSearch Flag = "search"

// doubleClickThreshold is the maximum time between two clicks on the same item
// for them to be treated as a double-click.
const doubleClickThreshold = 300 * time.Millisecond

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
	items    []string // Text items to display in the list (primary content)
	original []string // Unfiltered items saved while a filter is active (nil = no active filter)

	// ---- Navigation State ----
	index  int // Current highlight position (focused item index, -1 if none)
	offset int // Vertical scroll offset for viewport positioning (top visible item index)

	// ---- Selection Management ----
	selection []int // Indices of currently selected items (supports multi-selection)
	disabled  []int // Indices of items that cannot be selected or activated

	// ---- Display Options ----
	numbers     bool // Show line numbers next to each item for reference
	scrollbar   bool // Display scrollbar indicator on the right edge
	quickSearch bool // Enable quick-search by typing the first letter of an item

	// ---- Mouse State ----
	lastClickIndex int       // absolute list index of the last accepted click (-1 = none)
	lastClickTime  time.Time // timestamp of the last accepted click
}

// NewList creates a new List widget with the specified ID and initial items.
// The widget is initialized with sensible defaults suitable for most use cases,
// including focus capability and scrollbar display.
//
// Returns:
//   - *List: A fully initialized List widget ready for configuration and use
func NewList(id, class string, items []string) *List {
	list := &List{
		Component:      Component{id: id, class: class},
		items:          items,
		index:          0,
		selection:      make([]int, 0, 1),
		offset:         0,
		numbers:        false,
		scrollbar:      true,
		quickSearch:    true,
		lastClickIndex: -1,
	}
	list.SetFlag(FlagFocusable, true)
	list.SetFlag(flagSearch, true)
	OnKey(list, list.handleKey)
	OnMouse(list, list.handleMouse)
	return list
}

// Apply applies a theme's styles to the component.
func (l *List) Apply(theme *Theme) {
	theme.Apply(l, l.Selector("list"), "disabled", "focused", "hovered")
	theme.Apply(l, l.Selector("list/highlight"), "disabled", "focused", "hovered")
}

// Items returns the list of items displayed in the list.
func (l *List) Items() []string {
	return l.items
}

// Set replaces all items in the list and resets the selection to the first item.
func (l *List) Set(value []string) {
	l.items = value
	l.index = 0
	l.offset = 0
	l.Refresh()
}

// Filter applies a case-insensitive substring match and updates the list's
// visible content. An empty string clears the filter and restores the original
// unfiltered items.
func (l *List) Filter(filter string) {
	if filter == "" {
		if l.original != nil {
			l.Set(l.original)
			l.original = nil
		}
		return
	}
	if l.original == nil {
		l.original = l.items
	}
	lower := strings.ToLower(filter)
	var filtered []string
	for _, item := range l.original {
		if strings.Contains(strings.ToLower(item), lower) {
			filtered = append(filtered, item)
		}
	}
	l.Set(filtered)
}

// Suggest returns list items whose text has query as a case-insensitive prefix.
// It searches the unfiltered items when a filter is active, so candidates are
// not limited to the currently visible subset. Returns nil when nothing matches
// or query is empty.
func (l *List) Suggest(query string) []string {
	if query == "" {
		return nil
	}
	source := l.items
	if l.original != nil {
		source = l.original
	}
	lower := strings.ToLower(query)
	var results []string
	for _, item := range source {
		if strings.HasPrefix(strings.ToLower(item), lower) {
			results = append(results, item)
		}
	}
	if len(results) == 0 {
		return nil
	}
	return results
}

// Select selects the item at the specified index.
func (l *List) Select(index int) {
	l.index = index
	l.adjust()
	l.Refresh()
}

// Selected returns the selected index
func (l *List) Selected() int {
	return l.index
}

// ---- Widget Methods -------------------------------------------------------

// Refresh triggers a visual update of the List widget by requesting a redraw.
// This method should be called whenever the list's visual state has changed
// and needs to be reflected on screen.
func (l *List) Refresh() {
	Redraw(l)
}

// ---- Movements ------------------------------------------------------------

// Skip skips all disabled items in the given direction.
func (l *List) skip(index, direction int) int {
	next := index
	for next >= 0 && next < len(l.items) && slices.Contains(l.disabled, next) {
		next += direction
	}

	// Did we go past the first item, then look for the first enabled one
	if next < 0 {
		next = -1
		for i := range l.items {
			if !slices.Contains(l.disabled, i) {
				next = i
				break
			}
		}
		// Otherwise did we go past the end, look for the last enabled one
	} else if next >= len(l.items) {
		next = -1
		for i := len(l.items) - 1; i >= 0; i-- {
			if !slices.Contains(l.disabled, i) {
				next = i
				break
			}
		}
	}
	return next
}

// Move moves the highlight position up or down the specified number of steps.
// This method provides intelligent navigation that skips disabled items and
// handles boundary conditions gracefully with wraparound behavior.
//
// Parameters:
//   - count: Number of items to move up (positive or negative)
func (l *List) Move(count int) {
	if len(l.items) == 0 || count == 0 {
		return
	}

	// Determine direction and magnitude
	steps := count
	if steps < 0 {
		steps = -steps
	}

	// Calculate new index by moving step by step, skipping disabled items
	newIndex := l.index
	for i := 0; i < steps; i++ {
		if count > 0 {
			// Move down: find next enabled item after current
			newIndex = l.skip(newIndex+1, 1)
		} else {
			// Move up: find previous enabled item before current
			newIndex = l.skip(newIndex-1, -1)
		}
	}

	if newIndex != l.index {
		l.index = newIndex
		l.adjust()
		l.Dispatch(l, EvtSelect, l.index)
		l.Refresh()
	}
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
			l.Dispatch(l, EvtSelect, l.index)
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
			l.Dispatch(l, EvtSelect, l.index)
			break
		}
	}
	l.Refresh()
}

// moveTo moves the highlight directly to the given absolute list index without
// the step-by-step disabled-item skipping that Move uses. Callers must verify
// the index is not disabled before calling. No-op when already at that index.
func (l *List) moveTo(index int) {
	if index == l.index {
		return
	}
	l.index = index
	l.adjust()
	l.Dispatch(l, EvtSelect, l.index)
	l.Refresh()
}

// PageUp moves the highlight up by one page (viewport height).
// This method provides rapid navigation through long lists by jumping
// by the number of visible items rather than single item steps.
func (l *List) PageUp() {
	_, _, _, ih := l.Content()
	l.Move(-ih)
}

// PageDown moves the highlight down by one page (viewport height).
// This method provides rapid forward navigation through long lists by jumping
// by the number of visible items rather than single item steps.
func (l *List) PageDown() {
	_, _, _, ih := l.Content()
	l.Move(ih)
}

// ---- Actions --------------------------------------------------------------

// handleKey processes input events for the List widget, implementing
// comprehensive keyboard navigation and interaction capabilities. This method
// serves as the primary input processor for all list interactions.
//
// # Event Processing
//
// The method handles several categories of events:
//   - Navigation keys: Arrow keys, Page Up/Down, Home/End
//   - Action keys: Enter for activation
//   - Search keys: Letter keys for quick item search
//   - Custom keys: All other keys delegated to event handlers
//
// Parameters:
//   - event: The tcell.Event to process (keyboard, mouse, etc.)
//
// Returns:
//   - bool: true if event was handled/consumed, false if should be propagated
func (l *List) handleKey(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyUp:
		l.Move(-1)
		return true
	case tcell.KeyDown:
		l.Move(1)
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
		l.Dispatch(l, EvtActivate, l.index)
		return true
	case tcell.KeyRune:
		if l.Flag(flagSearch) {
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
	}
	return false
}

func (l *List) handleMouse(event *tcell.EventMouse) bool {
	if event.Buttons() != tcell.Button1 {
		return false
	}
	mx, my := event.Position()
	cx, cy, cw, ch := l.Content()
	if mx < cx || mx >= cx+cw || my < cy || my >= cy+ch {
		return false
	}
	// Ignore the scrollbar column when it is rendered
	if l.scrollbar && len(l.items) > ch && mx == cx+cw-1 {
		return false
	}
	index := l.offset + (my - cy)
	if index < 0 || index >= len(l.items) {
		return false
	}
	if slices.Contains(l.disabled, index) {
		return false
	}

	now := event.When()
	isDoubleClick := index == l.lastClickIndex && now.Sub(l.lastClickTime) <= doubleClickThreshold
	l.lastClickIndex = index
	l.lastClickTime = now

	l.moveTo(index)

	if isDoubleClick {
		l.Dispatch(l, EvtActivate, index)
		l.lastClickIndex = -1 // reset so a third click is a fresh single click
	}
	return true
}

// ---- Internal Methods ----------------------------------------------------

// adjust automatically adjusts the scroll offset to ensure the highlighted item
// remains visible within the current viewport. This method implements intelligent
// scrolling that maintains optimal user experience during navigation.
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
// Returns:
//   - []string: Slice of items currently visible in the viewport
func (l *List) visible() []string {
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

// renderList renders a List widget with items, selection highlighting, and optional scrollbar.
// This method handles the complete visual presentation of list widgets including
// item display, selection highlighting, line numbers, and scrollbar indicators.
//
// The method automatically adjusts text width to accommodate scrollbars
// and line numbers, ensuring proper layout regardless of configuration.
func (l *List) Render(r *Renderer) {
	x, y, w, h := l.Content()
	if h < 1 || w < 1 {
		return
	}

	items := l.visible()

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
		if slices.Contains(l.disabled, current) {
			style := l.Style(":disabled")
			r.Set(style.Foreground(), style.Background(), style.Font())
		} else if current == l.index {
			if l.Flag(FlagFocused) {
				style := l.Style("highlight:focused")
				r.Set(style.Foreground(), style.Background(), style.Font())
			} else {
				style := l.Style("highlight")
				r.Set(style.Foreground(), style.Background(), style.Font())
			}
		} else {
			style := l.Style()
			r.Set(style.Foreground(), style.Background(), style.Font())
		}

		// Render line number if enabled
		if l.numbers {
			r.Text(x, y+i, fmt.Sprintf(" %*d \u2502 %s", nw, current+1, item), tw)
		} else {
			r.Text(x, y+i, " "+item, tw)
		}
	}

	// Clear rows below the last rendered item.
	if len(items) < h {
		style := l.Style()
		r.Set(style.Foreground(), style.Background(), style.Font())
		r.Fill(x, y+len(items), w, h-len(items), " ")
	}

	// Render scrollbar if needed
	if l.scrollbar && len(l.items) > h {
		style := l.Style()
		r.Set(style.Foreground(), style.Background(), style.Font())
		scrollbarX := x + w - 1
		r.ScrollbarV(scrollbarX, y, h, l.offset, len(l.items))
	}
}

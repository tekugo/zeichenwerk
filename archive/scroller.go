package zeichenwerk

import (
	"github.com/gdamore/tcell/v2"
)

// Scroller represents a scrollable viewport widget that can contain a single child widget.
// It provides a windowed view into content that may be larger than the available display area,
// with support for both horizontal and vertical scrolling through keyboard navigation.
//
// Features:
//   - Single child widget container with viewport clipping
//   - Keyboard navigation with arrow keys for smooth scrolling
//   - Home/End key support for quick navigation to boundaries
//   - Automatic scroll boundary management
//   - Optional title display in the border area
//   - Integration with the styling system for visual customization
//
// The Scroller maintains internal scroll offsets (tx, ty) that determine which portion
// of the child widget is visible within the scroller's content area. When the child
// widget is larger than the scroller's content area, users can navigate through the
// content using keyboard controls.
//
// Keyboard Navigation:
//   - Arrow Up/Down: Scroll vertically by one line
//   - Arrow Left/Right: Scroll horizontally by one character
//   - Home: Reset to top-left corner (0, 0)
//   - End: Move to bottom-right corner
//
// The scroller is ideal for displaying large content such as:
//   - Long text documents or logs
//   - Large forms or data tables
//   - Image or diagram viewers
//   - Any content that exceeds the available screen space
type Scroller struct {
	BaseWidget
	Title  string // Optional title text to display in the scroller border
	child  Widget // The single child widget contained within this scroller
	tx, ty int    // Current horizontal and vertical scroll offsets
}

// NewScroller creates a new scrollable viewport widget with the specified ID and title.
// The scroller is initialized as focusable to enable keyboard navigation and is ready
// to contain a single child widget. The scroll offsets are initialized to (0, 0),
// showing the top-left corner of the child widget by default.
//
// Parameters:
//   - id: Unique identifier for the scroller widget
//   - title: Optional title text to display in the scroller border (can be empty)
//
// Returns:
//   - *Scroller: A new scroller widget instance ready to contain a child widget
//
// Example usage:
//
//	// Create a scroller with a title
//	scroller := NewScroller("document-viewer", "Document")
//
//	// Create a scroller without a title
//	scroller := NewScroller("content-area", "")
func NewScroller(id, title string) *Scroller {
	return &Scroller{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Title:      title,
	}
}

// Add sets the child widget for this scroller.
// Since Scroller is designed to contain exactly one child widget, calling Add
// will replace any previously set child widget. The child widget can be larger
// than the scroller's content area, enabling scrolling functionality.
//
// If you need to display multiple widgets within the scroller, consider wrapping
// them in a layout container (such as Flex, Grid, or Stack) as the child widget.
//
// Parameters:
//   - widget: The widget to be contained within this scroller
//
// Example usage:
//
//	// Add a large text widget that requires scrolling
//	largeText := NewText("content", textLines, false, 0)
//	scroller.Add(largeText)
//
//	// Add multiple widgets using a container
//	container := NewFlex("container", "vertical", "start", 0)
//	container.Add(widget1).Add(widget2).Add(widget3)
//	scroller.Add(container)
func (s *Scroller) Add(widget Widget) {
	s.child = widget
}

// Children returns a slice containing the single child widget of this scroller.
// This method implements the Container interface requirement. The returned
// slice will contain exactly one element if a child has been added, or will
// be empty if no child widget has been set.
//
// Parameters:
//   - visible: Whether to only return visible children (ignored for scroller)
//
// Returns:
//   - []Widget: A slice containing the child widget, or empty slice if no child is set
//
// Note: The returned slice should not be modified directly. Use the Add method
// to set or change the child widget.
func (s *Scroller) Children(_ bool) []Widget {
	if s.child == nil {
		return []Widget{}
	}
	return []Widget{s.child}
}

// Find searches for a widget with the specified ID within this scroller and its child widget.
// The search is performed recursively, first checking if this scroller matches the ID,
// then delegating to the generic Find function which will search the child widget
// and any of its descendants if the child is also a container.
//
// Parameters:
//   - id: The unique identifier of the widget to find
//   - visible: Whether to only search visible widgets
//
// Returns:
//   - Widget: The widget with the matching ID, or nil if not found
//
// The search order is:
//  1. Check if this scroller's ID matches
//  2. Recursively search within the child widget (if it exists)
//  3. If child is a container, search its descendants
func (s *Scroller) Find(id string, visible bool) Widget {
	return Find(s, id, visible)
}

// FindAt searches for the widget located at the specified screen coordinates.
// This method is used for mouse interaction to determine which widget is
// positioned at a given point. The search includes this scroller and its child widget,
// taking into account the current scroll offsets.
//
// Parameters:
//   - x: The x-coordinate to search at (absolute screen coordinates)
//   - y: The y-coordinate to search at (absolute screen coordinates)
//
// Returns:
//   - Widget: The widget at the specified coordinates, or nil if none found
//
// The search process:
//  1. Check if coordinates are within this scroller's bounds
//  2. If within bounds, recursively search the child widget (accounting for scroll offsets)
//  3. Return the most specific widget found at the coordinates
func (s *Scroller) FindAt(x, y int) Widget {
	return FindAt(s, x, y)
}

func (s *Scroller) Info() string {
	return "scroller [" + s.BaseWidget.Info() + "]"
}

// Layout positions and sizes the child widget within the scroller's content area.
// The child widget is always given its preferred width and height, allowing it to
// be larger than the scroller's visible area. The child is positioned based on the
// current scroll offsets (tx, ty), which determine what portion is visible.
//
// Layout process:
//  1. Get the content area coordinates and child's preferred size
//  2. Position the child widget offset by the current scroll position
//  3. Apply the child's preferred dimensions regardless of scroller size
//  4. Delegate to the generic Layout function for final processing
//
// The scroll offsets effectively translate the child widget's position:
//   - Positive tx scrolls content left (shows right portion)
//   - Positive ty scrolls content up (shows bottom portion)
//   - The renderer's clipping ensures only the visible portion is drawn
func (s *Scroller) Layout() {
	if s.child != nil {
		cx, cy, _, _ := s.Content()
		pw, ph := s.child.Hint()
		s.child.SetBounds(cx-s.tx, cy-s.ty, pw, ph)
	}
	Layout(s)
}

// Handle processes keyboard events for the scroller widget.
// This method provides keyboard navigation support for scrolling through
// the child widget content using arrow keys and home/end keys.
//
// Supported keyboard navigation:
//   - Arrow Up/Down: Scroll vertically by one line
//   - Arrow Left/Right: Scroll horizontally by one character
//   - Home: Reset offset to top-left corner (0, 0)
//   - End: Move to bottom-right corner based on child widget size
//
// The scrolling behavior respects content boundaries and will not scroll
// beyond the available content or into negative positions.
//
// Parameters:
//   - event: The tcell.Event to process (keyboard or mouse)
//
// Returns:
//   - bool: true if the event was handled, false otherwise
func (s *Scroller) Handle(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		return s.handleKeyEvent(event)
	}
	return s.BaseWidget.Handle(event)
}

// handleKeyEvent processes keyboard input for scroller navigation.
// This method implements scrolling controls for the scroller widget,
// allowing users to move through child content using standard keyboard shortcuts.
//
// Navigation controls:
//   - Vertical scrolling: Up/Down arrows
//   - Horizontal scrolling: Left/Right arrows
//   - Quick navigation: Home (top-left) and End (bottom-right)
//
// Scroll boundaries:
//   - Offsets are limited to non-negative values (no negative scrolling)
//   - Maximum scroll is determined by child widget size vs content area size
//
// Parameters:
//   - event: The keyboard event to process
//
// Returns:
//   - bool: true if the key was handled, false otherwise
func (s *Scroller) handleKeyEvent(event *tcell.EventKey) bool {
	if s.child == nil {
		return false
	}

	cw, ch := s.Size()       // Content area size
	pw, ph := s.child.Hint() // Child widget preferred size

	// Calculate maximum scroll offsets
	maxTx := max(pw-cw, 0)
	maxTy := max(ph-ch, 0)

	switch event.Key() {
	case tcell.KeyUp:
		// Scroll up by one line
		if s.ty > 0 {
			s.ty--
			s.Layout()
			s.Refresh()
			return true
		}

	case tcell.KeyDown:
		// Scroll down by one line
		if s.ty < maxTy {
			s.ty++
			s.Layout()
			s.Refresh()
			return true
		}

	case tcell.KeyLeft:
		// Scroll left by one character
		if s.tx > 0 {
			s.tx--
			s.Layout()
			s.Refresh()
			return true
		}

	case tcell.KeyRight:
		// Scroll right by one character
		if s.tx < maxTx {
			s.tx++
			s.Layout()
			s.Refresh()
			return true
		}

	case tcell.KeyHome:
		// Reset to top-left corner
		if s.tx > 0 || s.ty > 0 {
			s.tx = 0
			s.ty = 0
			s.Layout()
			s.Refresh()
			return true
		}

	case tcell.KeyEnd:
		// Move to bottom-right corner
		if s.tx < maxTx || s.ty < maxTy {
			s.tx = maxTx
			s.ty = maxTy
			s.Layout()
			s.Refresh()
			return true
		}
	}

	return false
}

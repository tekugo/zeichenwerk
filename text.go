package zeichenwerk

import "github.com/gdamore/tcell/v3"

// Text represents a multi-line text display widget with scrolling and
// automatic content management. It provides a scrollable text area that can
// display multiple lines of text with support for automatic scrolling,
// content limiting, and both manual and automatic scroll positioning.
//
// The Text widget is ideal for displaying logs, documentation, code, or any
// multi-line text content that may require scrolling or content management.
type Text struct {
	Component
	content []string // The lines of text content to display
	longest int      // Length of the longest line for vertical scrolling
	follow  bool     // Auto-follow mode: automatically scroll to show newest content
	max     int      // Maximum number of lines to retain (0 = unlimited)
	offsetX int      // Horizontal scroll offset for wide content
	offsetY int      // Vertical scroll offset for tall content
}

// NewText creates a new Text widget with the specified configuration.
// The widget is initialized as focusable to support keyboard navigation
// and scrolling when the user interacts with it.
//
// Parameters:
//   - id: Unique identifier for the text widget
//   - class: Style class
//   - content: Initial lines of text to display
//   - follow: Enable auto-follow mode to automatically scroll to newest content
//   - max: Maximum number of lines to retain (0 for unlimited, >0 for rotation)
//
// Returns:
//   - *Text: A new Text widget instance ready for display
func NewText(id, class string, content []string, follow bool, max int) *Text {
	text := &Text{
		Component: Component{id: id, class: class},
		content:   content,
		follow:    follow,
		max:       max,
		offsetX:   0,
		offsetY:   0,
	}
	text.SetFlag("focusable", true)
	OnKey(text, text.handleKey)
	return text
}

// Apply applies a theme's styles to the component.
func (t *Text) Apply(theme *Theme) {
	theme.Apply(t, t.Selector("text"))
}

// Refresh refreshes the widget.
func (t *Text) Refresh() {
	Redraw(t)
}

// Add appends one or more lines to the text content.
// This method supports adding multiple lines in a single operation and
// automatically manages content rotation when the maximum line limit is exceeded.
//
// Parameters:
//   - lines: One or more text lines to append to the content
func (t *Text) Add(lines ...string) {
	t.content = append(t.content, lines...)
	if t.max > 0 && len(t.content) > t.max {
		t.content = t.content[len(t.content)-t.max:]
	}
	t.adjust()
}

// Clear removes all content from the text widget.
// This method immediately empties the content array, resetting the widget
// to an empty state. The scroll offsets are not automatically reset.
func (t *Text) Clear() {
	t.content = []string{}
}

// Set replaces all content in the text widget with the provided lines.
// This method performs a complete content replacement, discarding any
// existing content and replacing it with the new content.
//
// Parameters:
//   - content: New lines of text to replace all existing content
func (t *Text) Set(content []string) {
	t.content = content
	t.adjust()
}

// adjust updates the scroll position and triggers a display refresh.
// This method recalculates the optimal scroll position based on the current
// content and widget configuration, then propagates the refresh to the parent.
func (t *Text) adjust() {
	// Determine longest line in regard to runes, not bytes
	t.longest = 0
	for _, line := range t.content {
		length := len([]rune(line))
		if length > t.longest {
			t.longest = length
		}
	}

	// Check, if we follow and need to update the offsets
	if t.follow && !t.Flag("focused") {
		t.offsetX = 0
		_, _, _, h := t.Content()
		t.offsetY = max(len(t.content)-h, 0)
	}

	t.Refresh()
}

// handleKeyEvent processes keyboard input for text navigation.
// This method implements scrolling and navigation controls for the text widget,
// allowing users to move through content using standard keyboard shortcuts.
//
// Navigation controls:
//   - Vertical scrolling: Up/Down arrows and Page Up/Down
//   - Horizontal scrolling: Left/Right arrows
//   - Quick navigation: Home (top-left) and End (bottom)
//
// Scroll boundaries:
//   - Vertical: Limited by content length and widget height
//   - Horizontal: Limited to non-negative values (no left overflow)
//   - Content shorter than widget: No scrolling allowed
//
// Parameters:
//   - event: The keyboard event to process
//
// Returns:
//   - bool: true if the key was handled, false otherwise
func (t *Text) handleKey(_ Widget, event *tcell.EventKey) bool {
	_, _, w, h := t.Content()
	maxOffsetY := max(len(t.content)-h, 0)

	switch event.Key() {
	case tcell.KeyUp:
		// Scroll up by one line
		if t.offsetY > 0 {
			t.offsetY--
			t.Refresh()
			return true
		}

	case tcell.KeyDown:
		// Scroll down by one line
		if t.offsetY < maxOffsetY {
			t.offsetY++
			t.Refresh()
			return true
		}

	case tcell.KeyLeft:
		// Scroll left by one character
		if t.offsetX > 0 {
			t.offsetX--
			t.Refresh()
			return true
		}

	case tcell.KeyRight:
		// Scroll right by one character
		if w+t.offsetX < t.longest {
			t.offsetX++
			t.Refresh()
			return true
		}

	case tcell.KeyPgUp:
		// Scroll up by one page (widget height)
		if t.offsetY > 0 {
			t.offsetY = max(t.offsetY-h, 0)
			t.Refresh()
			return true
		}

	case tcell.KeyPgDn:
		// Scroll down by one page (widget height)
		if t.offsetY < maxOffsetY {
			t.offsetY = min(t.offsetY+h, maxOffsetY)
			t.Refresh()
			return true
		}

	case tcell.KeyHome:
		// Jump to top-left (beginning of content)
		if t.offsetX > 0 || t.offsetY > 0 {
			t.offsetX = 0
			t.offsetY = 0
			t.Refresh()
			return true
		}

	case tcell.KeyEnd:
		// Jump to bottom of content, reset horizontal scroll
		if t.offsetY < maxOffsetY || t.offsetX > 0 {
			t.offsetX = 0
			t.offsetY = maxOffsetY
			t.Refresh()
			return true
		}
	}

	return false
}

// Render renders a Text widget's content with scrolling support.
// This method displays the visible portion of multi-line text content,
// handling both horizontal and vertical scrolling offsets.
func (t *Text) Render(r *Renderer) {
	x, y, w, h := t.Content()

	// Set style
	style := t.Style()
	r.Set(style.Foreground(), style.Background(), style.Font())

	// Check, if we need to render a vertical scroll bar
	iw := w
	if len(t.content) > h {
		iw--
	}

	// Check, if we need to render a horizontal scroll bar
	ih := h
	if t.longest > iw {
		ih--
	}

	if iw < w {
		r.ScrollbarV(x+w-1, y, ih, t.offsetY, len(t.content))
	}

	if ih < h {
		r.ScrollbarH(x, y+h-1, iw, t.offsetX, t.longest)
	}

	// Render visible text content
	for i := range len(t.content) - t.offsetY {
		if i >= ih {
			break
		}
		line := t.content[t.offsetY+i]
		if t.offsetX < len(line) {
			r.Text(x, y+i, line[t.offsetX:], iw)
		}
	}
}

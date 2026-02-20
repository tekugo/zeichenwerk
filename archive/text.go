package zeichenwerk

import (
	"github.com/gdamore/tcell/v2"
)

// Text represents a multi-line text display widget with scrolling and
// automatic content management. It provides a scrollable text area that can
// display multiple lines of text with support for automatic scrolling,
// content limiting, and both manual and automatic scroll positioning.
//
// Features:
//   - Multi-line text display with line-by-line rendering
//   - Automatic content limiting with configurable maximum line count
//   - Auto-follow mode for automatic scrolling to newest content (like log viewers)
//   - Manual scrolling support with horizontal and vertical offsets
//   - Dynamic content management (add, clear, replace operations)
//   - Keyboard navigation support when focused
//   - Efficient memory management through content rotation
//
// Scrolling behavior:
//   - Follow mode: Automatically scrolls to show the most recent content
//   - Manual mode: Maintains current scroll position when content is added
//   - Horizontal scrolling: Supports viewing wide content that exceeds widget width
//   - Vertical scrolling: Handles content that exceeds widget height
//
// Content management:
//   - Automatic line limiting prevents unlimited memory growth
//   - Content rotation removes oldest lines when maximum is exceeded
//   - Efficient append operations for real-time content like logs
//   - Bulk content replacement for static text display
//
// The Text widget is ideal for displaying logs, documentation, code, or any
// multi-line text content that may require scrolling or content management.
type Text struct {
	BaseWidget
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
//   - content: Initial lines of text to display
//   - follow: Enable auto-follow mode to automatically scroll to newest content
//   - max: Maximum number of lines to retain (0 for unlimited, >0 for rotation)
//
// Returns:
//   - *Text: A new Text widget instance ready for display
//
// Auto-follow behavior:
//   - When true: Automatically scrolls to show the most recent content when new lines are added
//   - When false: Maintains current scroll position, allowing manual navigation
//
// Content limiting:
//   - max = 0: No limit, all content is retained (use with caution for dynamic content)
//   - max > 0: Retains only the most recent 'max' lines, automatically removing older content
//
// Example usage:
//
//	// Log viewer with auto-scroll and 1000 line limit
//	logViewer := NewText("log", []string{}, true, 1000)
//
//	// Static document viewer without auto-scroll
//	docViewer := NewText("doc", documentLines, false, 0)
//
//	// Real-time status display with 50 line history
//	statusDisplay := NewText("status", []string{"Ready"}, true, 50)
func NewText(id string, content []string, follow bool, max int) *Text {
	return &Text{
		BaseWidget: BaseWidget{id: id, focusable: true},
		content:    content,
		follow:     follow,
		max:        max,
		offsetX:    0,
		offsetY:    0,
	}
}

// Add appends one or more lines to the text content.
// This method supports adding multiple lines in a single operation and
// automatically manages content rotation when the maximum line limit is exceeded.
//
// Parameters:
//   - lines: One or more text lines to append to the content
//
// Content management:
//  1. Appends all provided lines to the existing content
//  2. If content exceeds maximum line limit, removes oldest lines to maintain limit
//  3. Triggers refresh to update scroll position and display
//
// Auto-follow behavior:
//   - In follow mode: Automatically scrolls to show the newly added content
//   - In manual mode: Maintains current scroll position
//
// Memory management:
//   - When max > 0 and content exceeds limit, oldest lines are automatically removed
//   - This prevents unlimited memory growth for long-running applications
//   - Content rotation preserves the most recent lines
//
// Example usage:
//
//	text.Add("New log entry")
//	text.Add("Line 1", "Line 2", "Line 3")  // Add multiple lines
//	text.Add(fmt.Sprintf("Timestamp: %s", time.Now()))
func (t *Text) Add(lines ...string) {
	t.content = append(t.content, lines...)
	if t.max > 0 && len(t.content) > t.max {
		t.content = t.content[len(t.content)-t.max:]
	}
	t.Refresh()
}

// Clear removes all content from the text widget.
// This method immediately empties the content array, resetting the widget
// to an empty state. The scroll offsets are not automatically reset.
//
// Use cases:
//   - Clearing log displays
//   - Resetting content before loading new text
//   - Implementing "clear screen" functionality
//   - Preparing widget for new content
//
// Note: This method does not trigger a refresh automatically. Call Refresh()
// or add new content to update the display after clearing.
//
// Example usage:
//
//	text.Clear()
//	text.Add("Fresh content after clear")
//
//	// Or clear and refresh manually
//	text.Clear()
//	text.Refresh()
func (t *Text) Clear() {
	t.content = []string{}
}

// Refresh updates the scroll position and triggers a display refresh.
// This method recalculates the optimal scroll position based on the current
// content and widget configuration, then propagates the refresh to the parent.
//
// Auto-follow behavior:
//   - In follow mode: Resets horizontal scroll and positions vertical scroll to show newest content
//   - In manual mode: Maintains current scroll position
//
// Scroll calculation:
//  1. If follow mode is enabled, reset horizontal offset to 0
//  2. Calculate vertical offset to show the bottom of content within widget height
//  3. Ensure scroll position doesn't go negative (content shorter than widget)
//  4. Propagate refresh signal to parent widget for display update
//
// This method is called automatically by Add() and Set() operations,
// but can be called manually when scroll behavior needs to be recalculated.
//
// Manual refresh scenarios:
//   - After changing follow mode
//   - After widget resize
//   - After manual scroll position changes
//   - When content is modified externally
func (t *Text) Refresh() {
	// Determine longest line in regard to runes, not bytes
	t.longest = 0
	for _, line := range t.content {
		length := len([]rune(line))
		if length > t.longest {
			t.longest = length
		}
	}

	// Check, if we follow and need to update the offsets
	if t.follow && !t.focused {
		t.offsetX = 0
		_, h := t.Size()
		t.offsetY = max(len(t.content)-h, 0)
	}

	// Trigger parent refresh
	if t.parent != nil {
		t.parent.Refresh()
	}
}

// Set replaces all content in the text widget with the provided lines.
// This method performs a complete content replacement, discarding any
// existing content and replacing it with the new content.
//
// Parameters:
//   - content: New lines of text to replace all existing content
//
// Behavior:
//  1. Completely replaces existing content with provided lines
//  2. Does not apply maximum line limiting (unlike Add method)
//  3. Triggers refresh to update scroll position and display
//  4. Respects auto-follow mode for scroll positioning
//
// Use cases:
//   - Loading new documents or files
//   - Replacing content entirely (not appending)
//   - Updating display with processed or filtered content
//   - Implementing content reload functionality
//
// Note: Unlike Add(), this method does not enforce the maximum line limit.
// If you need to enforce limits after Set(), consider using Add() with Clear()
// or manually trimming the content.
//
// Example usage:
//
//	// Load new document
//	text.Set(documentLines)
//
//	// Replace with filtered content
//	filteredLines := filterContent(originalLines)
//	text.Set(filteredLines)
//
//	// Update with processed data
//	text.Set(processLogData(rawData))
func (t *Text) Set(content []string) {
	t.content = content
	t.Refresh()
}

// Handle processes keyboard and mouse events for the text widget.
// This method provides keyboard navigation support for scrolling through
// text content using arrow keys, page navigation, and home/end keys.
//
// Supported keyboard navigation:
//   - Arrow Up/Down: Scroll vertically by one line
//   - Arrow Left/Right: Scroll horizontally by one character
//   - Page Up/Down: Scroll vertically by one page (widget height)
//   - Home: Jump to the beginning of content (top-left)
//   - End: Jump to the end of content (bottom, reset horizontal scroll)
//
// The scrolling behavior respects content boundaries and will not scroll
// beyond the available content or into negative positions.
//
// Parameters:
//   - event: The tcell.Event to process (keyboard or mouse)
//
// Returns:
//   - bool: true if the event was handled, false otherwise
func (t *Text) Handle(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		return t.handleKeyEvent(event)
	}

	return false
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
func (t *Text) handleKeyEvent(event *tcell.EventKey) bool {
	w, h := t.Size()
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

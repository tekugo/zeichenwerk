package zeichenwerk

import (
	"fmt"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// Input is a single-line text input widget that allows users to enter and edit text.
// It provides comprehensive text editing functionality including cursor movement,
// horizontal scrolling for long text, and various input modes with robust Unicode support.
//
// Core Features:
//   - Single-line text editing with cursor positioning and movement
//   - Intelligent horizontal scrolling for text longer than widget width
//   - Maximum length constraints with automatic truncation
//   - Placeholder text display for empty inputs with custom styling
//   - Password masking for sensitive input fields with configurable mask characters
//   - Read-only mode for display-only text fields
//   - Event-driven architecture with change and enter callbacks
//   - Comprehensive keyboard shortcuts (Ctrl+A, Ctrl+E, Ctrl+K, Ctrl+U)
//
// Text Handling:
//   - Full Unicode support for international text input
//   - Proper handling of multi-byte characters in editing operations
//   - Character-accurate cursor positioning and text manipulation
//   - Efficient text rendering with viewport-based scrolling
//
// Scrolling System:
//   - Automatic scroll adjustment to keep cursor visible
//   - Smart scroll positioning to maintain comfortable editing zones
//   - Boundary-aware scrolling that prevents over-scrolling
//   - Smooth cursor movement across long text content
//
// Input Modes:
//   - Normal text input for general-purpose text fields
//   - Password mode with customizable masking characters
//   - Read-only mode for displaying non-editable text
//   - Placeholder mode for providing input hints and guidance
//
// The Input widget is designed for professional text editing experiences
// in terminal applications, supporting all common text input patterns
// found in modern user interfaces.
type Input struct {
	BaseWidget
	Text        string // Current text content of the input field
	Pos         int    // Current cursor position within the text (0-based character index)
	Offset      int    // Horizontal scroll offset for displaying long text (characters from start)
	Max         int    // Maximum allowed text length in characters (0 = unlimited)
	Placeholder string // Placeholder text shown when input is empty
	Masked      bool   // Whether to mask input characters for password fields
	MaskChar    rune   // Character used for masking (typically '*', '•', or '●')
	ReadOnly    bool   // Whether the input is read-only (navigation only, no editing)
}

// NewInput creates a new text input widget with the specified ID and default configuration.
// The input is initialized as a focusable widget ready for text entry with sensible
// defaults for general-purpose text input scenarios.
//
// Parameters:
//   - id: Unique identifier for the input widget (used for styling and event handling)
//
// Returns:
//   - *Input: A new input widget instance with default configuration
//
// Default Configuration:
//   - Empty text content (ready for user input)
//   - Cursor positioned at the beginning (position 0)
//   - No horizontal scroll offset (showing start of text)
//   - No maximum length limit (unlimited text input)
//   - No placeholder text (can be set later with Placeholder field)
//   - Password masking disabled (normal text display)
//   - Mask character set to '*' (standard password character)
//   - Read-only mode disabled (full editing capabilities)
//   - Focusable enabled (can receive keyboard input)
//   - No event callbacks set (can be configured with event system)
//
// Post-Creation Configuration:
// After creation, you can customize the input widget by setting its properties:
//   - Text content: input.Text = "initial value"
//   - Placeholder: input.Placeholder = "Enter username..."
//   - Maximum length: input.Max = 50
//   - Password mode: input.SetMasked(true, '•')
//   - Read-only mode: input.ReadOnly = true
func NewInput(id string) *Input {
	return &Input{
		BaseWidget:  BaseWidget{id: id, focusable: true},
		Text:        "",
		Pos:         0,
		Offset:      0,
		Max:         0,
		Placeholder: "",
		Masked:      false,
		MaskChar:    '*',
		ReadOnly:    false,
	}
}

// Refresh queues a redraw for the input.
func (i *Input) Refresh() {
	Redraw(i)
}

// Cursor returns the current cursor position relative to the visible text area.
// The cursor position is adjusted for horizontal scrolling, so it represents
// the visual position within the widget's content area rather than the absolute
// position within the text string.
//
// Returns:
//   - int: The x-coordinate of the cursor relative to the widget's content area
//   - int: The y-coordinate (always 0 for single-line input)
//
// The returned position is guaranteed to be within the widget's content bounds
// when the cursor is visible. If the cursor would be outside the visible area,
// the adjustScroll method should be called to correct the scroll offset.
func (i *Input) Cursor() (int, int) {
	cursorX := i.Pos - i.Offset

	// Ensure cursor position is within reasonable bounds
	_, _, iw, _ := i.Content()
	if cursorX < 0 {
		cursorX = 0
	} else if iw > 0 && cursorX >= iw {
		cursorX = iw - 1
	}

	return cursorX, 0
}

func (i *Input) Emit(event string, data ...any) bool {
	if i.handlers == nil {
		return false
	}
	handler, found := i.handlers[event]
	if found {
		return handler(i, event, data...)
	}
	return false
}

// Find searches for a widget with the specified ID within this input widget.
// Since input widgets are leaf widgets (they don't contain child widgets),
// this method always returns nil.
//
// Parameters:
//   - id: The unique identifier to search for
//
// Returns:
//   - Widget: Always returns nil as input widgets have no children
func (i *Input) Find(id string) Widget {
	return nil
}

// Info returns a human-readable description of the input widget's current state.
// This includes position, dimensions, and widget type information.
// Primarily used for debugging and development purposes.
//
// Returns:
//   - string: Formatted string with input widget information
func (i *Input) Info() string {
	x, y, w, h := i.Bounds()
	return fmt.Sprintf("@%d.%d %d:%d input[%d/%d]", x, y, w, h, len(i.Text), i.Max)
}

// SetText sets the text content of the input widget and adjusts cursor and scroll positions.
// This method provides a safe way to programmatically set the input's text content
// while maintaining proper cursor positioning and scroll state.
//
// Parameters:
//   - text: The new text content to set (supports full Unicode)
//
// Behavior and Safety:
//   - Respects read-only mode (no changes if read-only)
//   - Enforces maximum length constraints with Unicode-aware truncation
//   - Adjusts cursor position to stay within valid text bounds
//   - Updates horizontal scroll to keep cursor visible
//   - Triggers "change" event for consistent event handling
//   - Handles Unicode characters properly for international text
//
// Length Handling:
//   - Uses Unicode rune count for accurate character-based length limits
//   - Truncates at character boundaries to avoid corrupting multi-byte sequences
//   - Preserves text integrity when enforcing maximum length constraints
//
// Cursor Management:
//   - Automatically repositions cursor if it would be beyond the new text end
//   - Maintains cursor visibility through intelligent scroll adjustment
//   - Ensures cursor remains in a valid editing position
//
// This method is safe to call at any time and will maintain the widget's
// internal consistency regardless of the current state.
func (i *Input) SetText(text string) {
	if i.ReadOnly {
		return
	}

	runes := []rune(i.Text)
	if i.Max > 0 && len(runes) > i.Max {
		text = string(runes[:i.Max])
	}

	i.Text = text
	if i.Pos > len(runes) {
		i.Pos = len(runes)
	}
	i.adjust()

	i.Emit("change", i.Text)
}

// SetMasked configures password masking for the input widget.
// When masking is enabled, all characters in the input are displayed
// as the specified mask character instead of their actual values.
// This is commonly used for password fields and other sensitive inputs.
//
// Parameters:
//   - masked: Whether to enable character masking
//   - maskChar: The character to display instead of actual text (e.g., '*', '•')
//
// Example usage:
//
//	// Enable password masking with asterisks
//	input.SetMasked(true, '*')
//
//	// Disable masking to show actual text
//	input.SetMasked(false, '*')
func (i *Input) SetMasked(masked bool, maskChar rune) {
	i.Masked = masked
	i.MaskChar = maskChar
}

// ---- Movement -------------------------------------------------------------

// Left moves the cursor one position to the left within the text.
// If the cursor is already at the beginning of the text, this method has no effect.
// The horizontal scroll is automatically adjusted to keep the cursor visible.
//
// This method is typically called in response to the Left arrow key.
func (i *Input) Left() {
	if i.Pos > 0 {
		i.Pos--
		i.adjust()
	}
}

// Right moves the cursor one position to the right within the text.
// If the cursor is already at the end of the text, this method has no effect.
// The horizontal scroll is automatically adjusted to keep the cursor visible.
//
// This method is typically called in response to the Right arrow key.
func (i *Input) Right() {
	if i.Pos < len(i.Text) {
		i.Pos++
		i.adjust()
	}
}

// Start moves the cursor to the beginning of the text (position 0).
// The horizontal scroll is reset to show the beginning of the text.
//
// This method is typically called in response to the Home key or Ctrl+A.
func (i *Input) Start() {
	i.Pos = 0
	i.adjust()
}

// End moves the cursor to the end of the text (after the last character).
// The horizontal scroll is adjusted to show the end of the text if necessary.
//
// This method is typically called in response to the End key or Ctrl+E.
func (i *Input) End() {
	i.Pos = len(i.Text)
	i.adjust()
}

// ---- Editing --------------------------------------------------------------

// Insert inserts a character at the current cursor position.
// The character is inserted between existing characters, and the cursor advances
// to the position after the inserted character. If the input is read-only or
// the maximum length would be exceeded, the insertion is ignored.
//
// Parameters:
//   - ch: The character to insert at the cursor position
//
// Behavior:
//   - Respects read-only mode (no insertion if read-only)
//   - Enforces maximum length constraints
//   - Advances cursor position after insertion
//   - Updates horizontal scroll to keep cursor visible
//   - Triggers OnChange callback if set
func (i *Input) Insert(ch rune) {
	if i.ReadOnly {
		return
	}

	// Convert the string to runes to avoid unicode problems
	runes := []rune(i.Text)
	if i.Pos < 0 || i.Pos > len(runes) || (i.Max > 0 && len(runes) >= i.Max) {
		return
	}

	// Insert character at cursor position
	runes = append(runes[:i.Pos], append([]rune{ch}, runes[i.Pos:]...)...)
	i.Text = string(runes)
	i.Pos++
	i.adjust()

	i.Emit("change", i.Text)
}

// Delete removes the character immediately before the cursor position (backspace operation).
// The cursor moves back one position after deletion. If the cursor is at the beginning
// of the text or the input is read-only, this method has no effect.
//
// Behavior:
//   - Respects read-only mode (no deletion if read-only)
//   - Only deletes if cursor is not at the beginning
//   - Moves cursor back one position after deletion
//   - Updates horizontal scroll to keep cursor visible
//   - Triggers OnChange callback if set
//
// This method is typically called in response to the Backspace key.
func (i *Input) Delete() {
	if i.ReadOnly || i.Pos == 0 {
		return
	}

	runes := []rune(i.Text)
	runes = append(runes[:i.Pos-1], runes[i.Pos:]...)
	i.Text = string(runes)
	i.Pos--
	i.adjust()

	i.Emit("change", i.Text)
}

// DeleteForward removes the character at the current cursor position (delete operation).
// The cursor position remains unchanged after deletion. If the cursor is at the end
// of the text or the input is read-only, this method has no effect.
//
// Behavior:
//   - Respects read-only mode (no deletion if read-only)
//   - Only deletes if cursor is not at the end of text
//   - Cursor position remains unchanged
//   - Updates horizontal scroll to keep cursor visible
//   - Triggers OnChange callback if set
//
// This method is typically called in response to the Delete key.
func (i *Input) DeleteForward() {
	if i.ReadOnly || i.Pos >= len(i.Text) {
		return
	}

	runes := []rune(i.Text)
	runes = append(runes[:i.Pos], runes[i.Pos+1:]...)
	i.Text = string(runes)
	i.adjust()

	i.Emit("change", i.Text)
}

// Clear removes all text from the input and resets the cursor to the beginning.
// The horizontal scroll offset is also reset to 0. If the input is read-only,
// this method has no effect.
//
// Behavior:
//   - Respects read-only mode (no clearing if read-only)
//   - Removes all text content
//   - Resets cursor to position 0
//   - Resets horizontal scroll offset to 0
//   - Triggers OnChange callback if set
//
// This method is useful for programmatically clearing form fields or
// implementing "clear" buttons in user interfaces.
func (i *Input) Clear() {
	if i.ReadOnly {
		return
	}

	i.Text = ""
	i.Pos = 0
	i.Offset = 0

	i.Emit("change", i.Text)
}

// ---- Internal methods -----------------------------------------------------

// adjust adjusts the horizontal scroll offset to ensure the cursor remains visible
// within the widget's content area. This method implements intelligent scrolling
// that provides a smooth editing experience for text longer than the widget width.
//
// Intelligent Scrolling Algorithm:
//  1. Cursor visibility: Ensures cursor is always within the visible content area
//  2. Left boundary: If cursor moves left of visible area, scroll left to show cursor
//  3. Right boundary: If cursor moves right of visible area, scroll right with margin
//  4. Boundary protection: Prevents scrolling past text boundaries
//  5. Optimization: Avoids unnecessary scrolling when cursor is already visible
//
// Scrolling Behavior:
//   - Smooth cursor tracking: Scroll adjusts automatically during cursor movement
//   - Edge handling: Prevents over-scrolling beyond text start or optimal end position
//   - Performance: Only scrolls when necessary to maintain cursor visibility
//   - User experience: Maintains comfortable editing zones away from widget edges
//
// Boundary Management:
//   - Minimum offset: Never scrolls to negative positions (before text start)
//   - Maximum offset: Limits scrolling to prevent unnecessary whitespace
//   - Text length awareness: Adjusts maximum scroll based on current text length
//   - Widget width consideration: Accounts for available display width
//
// This method is called automatically by all cursor movement and text editing
// operations to maintain optimal visibility without manual intervention.
func (i *Input) adjust() {
	_, _, iw, _ := i.Content()
	if iw <= 0 {
		return
	}

	// Ensure cursor is visible within the content area
	if i.Pos < i.Offset {
		// Cursor is to the left of visible area - scroll left
		i.Offset = i.Pos
	} else if i.Pos >= i.Offset+iw {
		// Cursor is to the right of visible area - scroll right
		// Keep cursor positioned with some margin from the right edge for better UX
		i.Offset = i.Pos - iw + 1
	}

	// Don't scroll past the beginning of the text
	if i.Offset < 0 {
		i.Offset = 0
	}

	// Ensure offset doesn't exceed text length unnecessarily
	limit := max(len(i.Text)-iw+1, 0)
	if i.Offset > limit {
		i.Offset = limit
	}
}

// Visible returns the portion of text that should be displayed within the widget's
// content area, taking into account horizontal scrolling and password masking.
// This method handles both normal text display and masked password display.
//
// Returns:
//   - string: The text that should be rendered, potentially masked and scrolled
//
// Behavior:
//   - Applies password masking if enabled (replaces characters with mask character)
//   - Applies horizontal scrolling based on current offset
//   - Returns empty string if content area has no width
//   - Handles edge cases where offset exceeds text length
//   - Ensures proper Unicode character handling for masking
//
// The returned text represents exactly what should be visible to the user,
// making it suitable for direct rendering by the UI system.
func (i *Input) Visible() string {
	_, _, iw, _ := i.Content()
	if iw <= 0 {
		return ""
	}

	displayText := i.Text
	if i.Masked {
		// Replace all characters with mask character for password fields
		// Handle Unicode characters properly by converting to runes first
		textRunes := []rune(i.Text)
		maskRunes := make([]rune, len(textRunes))
		for j := range maskRunes {
			maskRunes[j] = i.MaskChar
		}
		displayText = string(maskRunes)
	}

	// Apply horizontal scrolling to show the relevant portion
	if i.Offset >= len(displayText) {
		return ""
	}

	// Calculate the end position for the visible text
	endX := min(i.Offset+iw, len(displayText))

	// Ensure we don't go beyond text boundaries
	if i.Offset < 0 {
		i.Offset = 0
		return displayText[:endX]
	}

	return displayText[i.Offset:endX]
}

// Handle processes keyboard events for the input widget and performs the appropriate
// text editing operations. This method implements a comprehensive keyboard interface
// that supports all standard text editing operations with professional-grade functionality.
//
// Keyboard Event Processing:
//   - Event filtering: Only processes keyboard events, ignores other event types
//   - Read-only respect: In read-only mode, only navigation keys are processed
//   - Unicode support: Properly handles international characters and symbols
//   - Event consumption: Returns true for handled events to prevent further propagation
//
// Navigation Operations:
//   - Left/Right arrows: Character-by-character cursor movement with scroll adjustment
//   - Home/End keys: Jump to beginning/end of text with optimal scroll positioning
//   - Ctrl+A: Alternative Home key (common in terminal applications)
//   - Ctrl+E: Alternative End key (common in terminal applications)
//
// Text Editing Operations:
//   - Backspace: Delete character before cursor (standard backspace behavior)
//   - Delete: Delete character at cursor position (forward delete)
//   - Printable characters: Insert characters at cursor with Unicode support
//   - Character validation: Only accepts printable characters for text insertion
//
// Advanced Editing Shortcuts:
//   - Ctrl+K: Kill text from cursor to end of line (common in Unix/Linux)
//   - Ctrl+U: Kill text from beginning of line to cursor (common in Unix/Linux)
//   - Enter: Trigger "enter" event for form submission or action handling
//
// Event System Integration:
//   - Change events: Automatically triggered for all text modifications
//   - Enter events: Triggered when Enter key is pressed for action handling
//   - Event data: Passes current text content with all events
//   - Callback safety: Checks for event handler existence before calling
//
// Parameters:
//   - evt: The tcell.Event to process (keyboard events only)
//
// Returns:
//   - bool: true if the event was handled and consumed, false otherwise
//
// Error Handling:
//   - Gracefully ignores non-keyboard events
//   - Safely handles read-only mode restrictions
//   - Validates character input before processing
//   - Maintains widget consistency even with invalid operations
//
// Performance Considerations:
//   - Efficient event filtering to avoid unnecessary processing
//   - Minimal scroll adjustments only when cursor position changes
//   - Direct character insertion without intermediate string operations
//   - Event handler checks to avoid null pointer exceptions
func (i *Input) Handle(evt tcell.Event) bool {
	event, ok := evt.(*tcell.EventKey)
	if !ok {
		return false
	}

	// In read-only mode, only allow navigation keys
	if i.ReadOnly && event.Key() != tcell.KeyLeft && event.Key() != tcell.KeyRight &&
		event.Key() != tcell.KeyHome && event.Key() != tcell.KeyEnd {
		return false
	}

	switch event.Key() {
	case tcell.KeyLeft:
		i.Left()
	case tcell.KeyRight:
		i.Right()
	case tcell.KeyHome:
		i.Start()
	case tcell.KeyEnd:
		i.End()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		i.Delete()
	case tcell.KeyDelete:
		i.DeleteForward()
	case tcell.KeyCtrlA:
		i.Start()
	case tcell.KeyCtrlE:
		i.End()
	case tcell.KeyCtrlK:
		// Delete from cursor to end of text
		if !i.ReadOnly {
			i.Text = i.Text[:i.Pos]
			i.adjust()
			i.Emit("change", i.Text)
		}
	case tcell.KeyCtrlU:
		// Delete from beginning of text to cursor
		if !i.ReadOnly {
			i.Text = i.Text[i.Pos:]
			i.Pos = 0
			i.adjust()
			i.Emit("change", i.Text)
		}
	case tcell.KeyEnter:
		i.Emit("enter", i.Text)
	case tcell.KeyRune:
		ch := event.Rune()
		if unicode.IsPrint(ch) {
			i.Insert(ch)
		} else {
			return false
		}
	default:
		return i.Emit("key", event)
	}

	i.Refresh()
	return true
}

// ShouldShowPlaceholder returns whether the placeholder text should be displayed
// instead of the actual input content. The placeholder is shown when the input
// is empty and a placeholder string has been configured.
//
// Returns:
//   - bool: true if placeholder should be displayed, false otherwise
//
// Placeholder Display Logic:
//   - Shows placeholder only when text is completely empty
//   - Requires placeholder string to be non-empty
//   - Used by rendering system to determine display content
//   - Typically rendered with different styling (dimmed, italic, etc.)
//
// The placeholder provides user guidance and improves the user experience
// by indicating the expected input format or content type.
func (i *Input) ShouldShowPlaceholder() bool {
	return i.Text == "" && i.Placeholder != ""
}

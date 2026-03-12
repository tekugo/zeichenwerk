package zeichenwerk

import (
	"github.com/gdamore/tcell/v3"
)

// Input is a single-line text input widget that allows users to enter and edit text.
// It provides comprehensive text editing functionality including cursor movement,
// horizontal scrolling for long text, and various input modes with robust Unicode support.
type Input struct {
	Component
	text        string // Current text content of the input field
	pos         int    // Current cursor position within the text (0-based character index)
	offset      int    // Horizontal scroll offset for displaying long text (characters from start)
	max         int    // Maximum allowed text length in characters (0 = unlimited)
	placeholder string // Placeholder text shown when input is empty
	mask        string // Character used for masking (typically '*', '•', or '●')
}

// NewInput creates a new text input widget with the specified ID and default configuration.
// The input is initialized as a focusable widget ready for text entry with sensible
// defaults for general-purpose text input scenarios.
func NewInput(id string, params ...string) *Input {
	var text, placeholder, mask string
	if len(params) > 0 {
		text = params[0]
	}
	if len(params) > 1 {
		placeholder = params[1]
	}
	if len(params) > 2 {
		mask = params[2]
	} else {
		mask = "*"
	}

	input := &Input{
		Component:   Component{id: id, hheight: 1},
		text:        text,
		pos:         0,
		offset:      0,
		max:         0,
		placeholder: placeholder,
		mask:        mask,
	}
	input.SetFlag("focusable", true)
	input.SetFlag("masked", false)
	input.SetFlag("readonly", false)
	OnKey(input, input.handleKey)
	return input
}

// ---- Widget Methods -------------------------------------------------------

// Cursor returns the current cursor position relative to the visible text
// area. The cursor position is adjusted for horizontal scrolling, so it
// represents the visual position within the widget's content area rather than
// the absolute position within the text string.
//
// Returns:
//   - int: The x-coordinate of the cursor relative to the widget's content area
//   - int: The y-coordinate (always 0 for single-line input)
//
// The returned position is guaranteed to be within the widget's content bounds
// when the cursor is visible. If the cursor would be outside the visible area,
// the adjustScroll method should be called to correct the scroll offset.
func (i *Input) Cursor() (int, int, string) {
	cursorX := i.pos - i.offset

	// Ensure cursor position is within reasonable bounds
	_, _, iw, _ := i.Content()
	if cursorX < 0 {
		cursorX = 0
	} else if iw > 0 && cursorX >= iw {
		cursorX = iw - 1
	}

	return cursorX, 0, "|"
}

// Refresh queues a redraw for the input.
func (i *Input) Refresh() {
	Redraw(i)
}

// ---- Input Methods --------------------------------------------------------

// SetMasked configures password masking for the input widget.
// When masking is enabled, all characters in the input are displayed
// as the specified mask character instead of their actual values.
// This is commonly used for password fields and other sensitive inputs.
// If the mask is the empty string, masking is disabled.
//
// Parameters:
//   - mask: The character to display instead of actual text (e.g., "*", "•")
func (i *Input) SetMask(mask string) {
	i.SetFlag("masked", mask != "")
	i.mask = mask
}

// SetText sets the text content of the input widget and adjusts cursor and
// scroll positions. This method provides a safe way to programmatically set
// the input's text content while maintaining proper cursor positioning and
// scroll state.
//
// Parameters:
//   - text: The new text content to set (supports full Unicode)
//
// This method is safe to call at any time and will maintain the widget's
// internal consistency regardless of the current state.
func (i *Input) SetText(text string) {
	if i.Flag("readonly") {
		return
	}

	runes := []rune(text)
	if i.max > 0 && len(runes) > i.max {
		text = string(runes[:i.max])
	}

	i.text = text
	if i.pos > len(runes) {
		i.pos = len(runes)
	}
	i.adjust()

	i.Dispatch(i, "change", i.text)
}

// Text returns the current text content.
func (i *Input) Text() string {
	return i.text
}

// ---- Movement -------------------------------------------------------------

// Left moves the cursor one position to the left within the text.
// If the cursor is already at the beginning of the text, this method has no effect.
// The horizontal scroll is automatically adjusted to keep the cursor visible.
//
// This method is typically called in response to the Left arrow key.
func (i *Input) Left() {
	if i.pos > 0 {
		i.pos--
		i.adjust()
	}
	i.Refresh()
}

// Right moves the cursor one position to the right within the text.
// If the cursor is already at the end of the text, this method has no effect.
// The horizontal scroll is automatically adjusted to keep the cursor visible.
//
// This method is typically called in response to the Right arrow key.
func (i *Input) Right() {
	if i.pos < len(i.text) {
		i.pos++
		i.adjust()
	}
	i.Refresh()
}

// Start moves the cursor to the beginning of the text (position 0).
// The horizontal scroll is reset to show the beginning of the text.
//
// This method is typically called in response to the Home key or Ctrl+A.
func (i *Input) Start() {
	i.pos = 0
	i.adjust()
	i.Refresh()
}

// End moves the cursor to the end of the text (after the last character).
// The horizontal scroll is adjusted to show the end of the text if necessary.
//
// This method is typically called in response to the End key or Ctrl+E.
func (i *Input) End() {
	i.pos = len(i.text)
	i.adjust()
	i.Refresh()
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
func (i *Input) Insert(ch string) {
	if i.Flag("readonly") {
		return
	}

	// Convert the string to runes to avoid unicode problems
	runes := []rune(i.text)
	if i.pos < 0 || i.pos > len(runes) || (i.max > 0 && len(runes) >= i.max) {
		return
	}

	// Insert character at cursor position
	runes = append(runes[:i.pos], append([]rune{[]rune(ch)[0]}, runes[i.pos:]...)...)
	i.text = string(runes)
	i.pos++
	i.adjust()
	i.Refresh()

	i.Dispatch(i, "change", i.text)
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
	if i.Flag("readonly") || i.pos == 0 {
		return
	}

	runes := []rune(i.text)
	runes = append(runes[:i.pos-1], runes[i.pos:]...)
	i.text = string(runes)
	i.pos--
	i.adjust()
	i.Refresh()

	i.Dispatch(i, "change", i.text)
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
	if i.Flag("readonly") || i.pos >= len(i.text) {
		return
	}

	runes := []rune(i.text)
	runes = append(runes[:i.pos], runes[i.pos+1:]...)
	i.text = string(runes)
	i.adjust()
	i.Refresh()

	i.Dispatch(i, "change", i.text)
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
	if i.Flag("readonly") {
		return
	}

	i.text = ""
	i.pos = 0
	i.offset = 0
	i.Refresh()

	i.Dispatch(i, "change", i.text)
}

// ---- Internal methods -----------------------------------------------------

// adjust adjusts the horizontal scroll offset to ensure the cursor remains visible
// within the widget's content area. This method implements intelligent scrolling
// that provides a smooth editing experience for text longer than the widget width.
func (i *Input) adjust() {
	_, _, iw, _ := i.Content()
	if iw <= 0 {
		return
	}

	// Ensure cursor is visible within the content area
	if i.pos < i.offset {
		// Cursor is to the left of visible area - scroll left
		i.offset = i.pos
	} else if i.pos >= i.offset+iw {
		// Cursor is to the right of visible area - scroll right
		// Keep cursor positioned with some margin from the right edge for better UX
		i.offset = i.pos - iw + 1
	}

	// Don't scroll past the beginning of the text
	if i.offset < 0 {
		i.offset = 0
	}

	// Ensure offset doesn't exceed text length unnecessarily
	limit := max(len(i.text)-iw+1, 0)
	if i.offset > limit {
		i.offset = limit
	}
}

// Visible returns the portion of text that should be displayed within the widget's
// content area, taking into account horizontal scrolling and password masking.
// This method handles both normal text display and masked password display.
//
// Returns:
//   - string: The text that should be rendered, potentially masked and scrolled
func (i *Input) visible() string {
	_, _, iw, _ := i.Content()
	if iw <= 0 {
		return ""
	}

	display := []rune(i.text)
	if i.Flag("masked") {
		// Replace all characters with mask character for password fields
		// Handle Unicode characters properly by converting to runes first
		maskRunes := make([]rune, len(display))
		for j := range maskRunes {
			maskRunes[j] = []rune(i.mask)[0]
		}
		display = maskRunes
	}

	// Apply horizontal scrolling to show the relevant portion
	if i.offset >= len(display) {
		return ""
	}

	// Calculate the end position for the visible text
	endX := min(i.offset+iw, len(display))

	// Ensure we don't go beyond text boundaries
	if i.offset < 0 {
		i.offset = 0
		return string(display[:endX])
	}

	return string(display[i.offset:endX])
}

// Handle processes keyboard events for the input widget and performs the appropriate
// text editing operations. This method implements a comprehensive keyboard interface
// that supports all standard text editing operations with professional-grade functionality.
func (i *Input) handleKey(_ Widget, evt *tcell.EventKey) bool {
	// In read-only mode, only allow navigation keys
	if i.Flag("readonly") && evt.Key() != tcell.KeyLeft && evt.Key() != tcell.KeyRight &&
		evt.Key() != tcell.KeyHome && evt.Key() != tcell.KeyEnd {
		return false
	}

	switch evt.Key() {
	case tcell.KeyLeft:
		i.Left()
		return true
	case tcell.KeyRight:
		i.Right()
		return true
	case tcell.KeyHome:
		i.Start()
		return true
	case tcell.KeyEnd:
		i.End()
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		i.Delete()
		return true
	case tcell.KeyDelete:
		i.DeleteForward()
		return true
	case tcell.KeyCtrlA:
		i.Start()
		return true
	case tcell.KeyCtrlE:
		i.End()
		return true
	case tcell.KeyCtrlK:
		// Delete from cursor to end of text
		if !i.Flag("readonly") {
			i.text = i.text[:i.pos]
			i.adjust()
			i.Dispatch(i, "change", i.text)
		}
		i.Refresh()
		return true
	case tcell.KeyCtrlU:
		// Delete from beginning of text to cursor
		if !i.Flag("readonly") {
			i.text = i.text[i.pos:]
			i.pos = 0
			i.adjust()
			i.Refresh()
			i.Dispatch(i, "change", i.text)
			return true
		}
	case tcell.KeyEnter:
		i.Dispatch(i, "enter", i.text)
		return true
	case tcell.KeyRune:
		ch := evt.Str()
		i.Insert(ch)
		i.Refresh()
		return true
	default:
		return false
	}
	return false
}

// Render renders an Input widget's text content with placeholder support.
// This method handles the display of user input text or placeholder text
// depending on the input's current state and content.
func (i *Input) Render(r *Renderer) {
	x, y, w, h := i.Content()
	if h < 1 || w < 1 {
		return
	}

	// Determine what text to display
	if i.text == "" && i.placeholder != "" {
		// Use a dimmed style for placeholder
		style := i.Style("placeholder")
		r.Set(style.Foreground(), style.Background(), style.Font())
		r.Text(x, y, i.placeholder, w)
	} else {
		state := i.State()
		if state != "" {
			state = ":" + state
		}
		style := i.Style(state)
		r.Set(style.Foreground(), style.Background(), style.Font())
		r.Text(x, y, i.visible(), w)
	}
}

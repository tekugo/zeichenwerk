package tui

// renderInput renders an Input widget's text content with placeholder support.
// This method handles the display of user input text or placeholder text
// depending on the input's current state and content.
//
// Parameters:
//   - input: The Input widget to render
//   - x, y: Top-left coordinates of the input's content area
//   - w, h: Width and height of the input's content area
//
// Rendering logic:
//  1. Validates that the content area has positive dimensions
//  2. Determines whether to show placeholder or actual input text
//  3. Applies appropriate styling (placeholder style vs normal style)
//  4. Renders the text within the specified width constraints
//
// Text display behavior:
//   - Shows placeholder text when input is empty or should show placeholder
//   - Uses dimmed/placeholder styling for placeholder text
//   - Shows actual input text with normal styling when content exists
//   - Automatically truncates text that exceeds the available width
//
// The method ensures proper visual feedback by using different styles
// for placeholder vs actual content, helping users understand the input state.
func (r *Renderer) renderInput(input *Input, x, y, w, h int) {
	if h < 1 || w < 1 {
		return
	}

	// Determine what text to display
	if input.ShouldShowPlaceholder() {
		// Use a dimmed style for placeholder
		r.SetStyle(input.Style("placeholder"))
		r.text(x, y, input.Placeholder, w)
	} else {
		r.text(x, y, input.Visible(), w)
	}
}

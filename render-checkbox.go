package zeichenwerk

// renderCheckbox renders a Checkbox widget with its checkbox indicator and label text.
// This method handles the display of the checkbox state and associated label.
//
// Parameters:
//   - checkbox: The Checkbox widget to render
//   - x, y: Top-left coordinates of the checkbox's content area
//   - w, h: Width and height of the checkbox's content area
//
// Rendering logic:
//  1. Validates that the content area has positive dimensions
//  2. Renders the checkbox indicator based on checked state
//  3. Renders the label text next to the checkbox indicator
//  4. Handles text truncation if label exceeds available space
//
// Checkbox indicator:
//   - "[x]" for checked state
//   - "[ ]" for unchecked state
//   - Uses widget's current style for proper theming
//
// Layout:
//   - Checkbox indicator takes 3 characters: "[x]" or "[ ]"
//   - Label text starts after a space: "[x] Label text"
//   - Total format: "[x] Label text" or "[ ] Label text"
func (r *Renderer) renderCheckbox(checkbox *Checkbox, x, y, w, h int) {
	if h < 1 || w < 1 {
		return
	}

	// Determine checkbox indicator based on state
	var indicator string
	if checkbox.Checked {
		indicator = "[x]"
	} else {
		indicator = "[ ]"
	}

	// Render checkbox indicator (takes 3 characters)
	if w >= 3 {
		r.text(x, y, indicator, 3)
	}

	// Render label text after the checkbox and a space
	if w > 4 && checkbox.Text != "" {
		labelX := x + 4 // Position after "[x] "
		labelWidth := w - 4
		r.text(labelX, y, checkbox.Text, labelWidth)
	}
}
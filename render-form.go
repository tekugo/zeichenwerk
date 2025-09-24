package zeichenwerk

// renderFormGroup renders a FormGroup widget by drawing field labels and delegating
// control rendering to their respective render methods.
//
// The rendering process:
//   1. Renders all field labels at their calculated positions
//   2. Renders all form control widgets (inputs, checkboxes, etc.)
//
// Label positioning is determined by the FormGroup's layout calculations,
// with labels only rendered if they have valid coordinates (x >= 0, y >= 0).
// Labels with negative coordinates are skipped (typically for multi-field lines
// in horizontal layout where only the first field gets a label).
//
// Parameters:
//   - fg: The FormGroup widget to render
func (r *Renderer) renderFormGroup(fg *FormGroup) {
	// Render labels at their calculated positions
	for _, line := range fg.fields {
		for _, field := range line {
			if field.x >= 0 && field.y >= 0 {
				r.text(field.x, field.y, field.Label, 0)
			}
		}
	}

	// Render form control widgets (inputs, checkboxes, etc.)
	for _, child := range fg.Children(true) {
		r.render(child)
	}
}

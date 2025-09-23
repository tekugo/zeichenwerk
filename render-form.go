package zeichenwerk

func (r *Renderer) renderFormGroup(fg *FormGroup) {
	// Render labels
	for _, line := range fg.fields {
		for _, field := range line {
			fg.Log(fg, "debug2", "Render label %d.%d %s", field.x, field.y, field.Label)
			if field.x >= 0 && field.y >= 0 {
				r.text(field.x, field.y, field.Label, 0)
			}
		}
	}

	// Render children
	for _, child := range fg.Children(true) {
		fg.Log(fg, "debug", "Render child %s", child.Info())
		r.render(child)
	}
}

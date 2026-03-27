package zeichenwerk

type field struct {
	label  string
	widget Widget
	x, y   int // label position; negative means don't render label
}

type FormGroup struct {
	Component
	title      string
	horizontal bool
	spacing    int // vertical spacing between lines
	lines      [][]*field
}

func NewFormGroup(id, class, title string, horizontal bool, spacing int) *FormGroup {
	fg := &FormGroup{
		Component:  Component{id: id, class: class},
		title:      title,
		horizontal: horizontal,
		spacing:    spacing,
	}
	return fg
}

func (fg *FormGroup) Add(line int, label string, widget Widget) {
	if widget == nil {
		return
	}
	for len(fg.lines) <= line {
		fg.lines = append(fg.lines, nil)
	}
	f := &field{label: label, widget: widget}
	fg.lines[line] = append(fg.lines[line], f)
	widget.SetParent(fg)
}

// Apply applies a theme's styles to the component.
func (fg *FormGroup) Apply(theme *Theme) {
	theme.Apply(fg, fg.Selector("form-group"))
}

func (fg *FormGroup) Children() []Widget {
	children := make([]Widget, 0, len(fg.lines))
	for _, line := range fg.lines {
		for _, f := range line {
			children = append(children, f.widget)
		}
	}
	return children
}

func (fg *FormGroup) Hint() (int, int) {
	w, h := 0, 0 // preferred width and height total
	mlw := 0     // maximum label width in horizontal layout
	for i, line := range fg.lines {
		lw, lh := 0, 0 // line width and height
		for j, field := range line {
			// Check the maximum label width
			if j == 0 && fg.horizontal {
				if len(field.label)+1 > mlw {
					mlw = len(field.label) + 1
				}
			}

			// Determine field width and height including styling
			fw, fh := field.widget.Hint() // field width and height
			style := field.widget.Style()
			fw += style.Horizontal()
			fh += style.Vertical()

			lw += fw
			if fh > lh {
				lh = fh
			}
			// Add 1 to the line width for every extra control as a spacing
			if j > 0 {
				lw++
			}
		}
		if !fg.horizontal {
			lh++
		}
		if i > 0 {
			lh++
		}
		if lw > w {
			w = lw
		}
		h += lh
	}
	return mlw + w, h
}

func (fg *FormGroup) Layout() {
	// Determine maximum label width (only for horizontal label placement)
	mlw := -1
	if fg.horizontal {
		for _, line := range fg.lines {
			if len(line[0].label) > mlw {
				mlw = len(line[0].label)
			}
		}
	}

	// Content area
	cx, cy, cw, _ := fg.Content()

	// Current line y position
	ly := cy
	for _, line := range fg.lines {
		// Current line x position
		lx := cx + mlw + 1
		lh := 0 // line height
		if !fg.horizontal {
			ly++
		}
		for j, field := range line {
			// Label position
			if fg.horizontal {
				if j == 0 {
					field.x, field.y = cx, ly
				} else {
					field.x, field.y = -1, -1
				}
			} else {
				field.x, field.y = lx, ly-1
			}

			// Control position
			fw, fh := field.widget.Hint()
			style := field.widget.Style()
			fw += style.Horizontal()
			fh += style.Vertical()
			if j < len(line)-1 {
				field.widget.SetBounds(lx, ly, fw, fh)
				lx += fw + 1
			} else {
				field.widget.SetBounds(lx, ly, cw+cx-lx, fh)
			}

			// Update line height
			if fh > lh {
				lh = fh
			}
		}
		ly = ly + lh + fg.spacing
	}
}

func (fg *FormGroup) Render(r *Renderer) {
	// Render common styling
	fg.Component.Render(r)

	// Render labels at their calculated positions
	for _, line := range fg.lines {
		for _, field := range line {
			if field.x >= 0 && field.y >= 0 {
				r.Text(field.x, field.y, field.label, 0)
			}
		}
	}

	// Render form control widgets (inputs, checkboxes, etc.)
	for _, child := range fg.Children() {
		if !child.Flag("hidden") {
			child.Render(r)
		}
	}
}

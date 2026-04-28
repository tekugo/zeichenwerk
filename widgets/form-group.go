package widgets

import (
	. "github.com/tekugo/zeichenwerk/core"
)

type field struct {
	label  string
	widget Widget
	x, y   int // label position; negative means don't render label
}

// FormGroup is a layout container for form controls. Each control occupies a
// named line; a label string may be placed to the left of the first control on
// each line. When horizontal is true labels appear to the left of the control
// row; when false labels appear above. spacing adds extra blank rows between
// lines.
type FormGroup struct {
	Component
	title      string
	horizontal bool
	spacing    int // vertical spacing between lines
	lines      [][]*field
}

// NewFormGroup creates a FormGroup with the given id, CSS class, title,
// orientation, and line spacing.
func NewFormGroup(id, class, title string, horizontal bool, spacing int) *FormGroup {
	fg := &FormGroup{
		Component:  Component{id: id, class: class},
		title:      title,
		horizontal: horizontal,
		spacing:    spacing,
	}
	return fg
}

// Add appends widget to the group. params[0] (int) selects the line index;
// params[1] (string) sets the label for that control. Returns ErrChildIsNil
// if widget is nil.
func (fg *FormGroup) Add(widget Widget, params ...any) error {
	if widget == nil {
		return ErrChildIsNil
	}
	line := 0
	label := ""
	if len(params) > 0 {
		if l, ok := params[0].(int); ok {
			line = l
		}
	}
	if len(params) > 1 {
		if l, ok := params[1].(string); ok {
			label = l
		}
	}
	for len(fg.lines) <= line {
		fg.lines = append(fg.lines, nil)
	}
	f := &field{label: label, widget: widget}
	fg.lines[line] = append(fg.lines[line], f)
	widget.SetParent(fg)
	return nil
}

// Apply applies a theme's styles to the component.
func (fg *FormGroup) Apply(theme *Theme) {
	theme.Apply(fg, fg.Selector("form-group"))
}

// Children returns all child widgets across all lines.
func (fg *FormGroup) Children() []Widget {
	children := make([]Widget, 0, len(fg.lines))
	for _, line := range fg.lines {
		for _, f := range line {
			children = append(children, f.widget)
		}
	}
	return children
}

// Hint returns the preferred content size needed to display all lines and
// their labels at their natural sizes.
func (fg *FormGroup) Hint() (int, int) {
	if fg.hwidth != 0 || fg.hheight != 0 {
		return fg.hwidth, fg.hheight
	}
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

// Layout calculates and applies positions for all labels and controls.
func (fg *FormGroup) Layout() error {
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
	return nil
}

// Render draws the group background, labels, and all child controls.
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
		if !child.Flag(FlagHidden) {
			child.Render(r)
		}
	}
}

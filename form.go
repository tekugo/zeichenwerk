package zeichenwerk

import "reflect"

type Form struct {
	BaseWidget
	Title string // Form title
	Data  any    // Form data as struct
	child Widget // Child container
}

func (f *Form) Add(widget Widget) {
	f.child = widget
}

func (f *Form) Children(_ bool) []Widget {
	if f.child != nil {
		return []Widget{f.child}
	} else {
		return []Widget{}
	}
}

func (f *Form) Find(id string, visible bool) Widget {
	return Find(f, id, visible)
}

func (f *Form) FindAt(x, y int) Widget {
	return FindAt(f, x, y)
}

func (f *Form) Hint() (int, int) {
	if f.child != nil {
		w, h := f.child.Hint()
		w += f.child.Style("").Horizontal()
		h += f.child.Style("").Vertical()
		style := f.Style("")
		if style != nil {
			style.Width = w
			style.Height = h
		}
		return w, h
	} else {
		return 0, 0
	}
}

func (f *Form) Layout() {
	if f.child != nil {
		cx, cy, cw, ch := f.Content()
		f.child.SetBounds(cx, cy, cw, ch)
	}
	Layout(f)
}

func (f *Form) Update(value reflect.Value) func(Widget, string, ...any) bool {
	return func(widget Widget, event string, values ...any) bool {
		f.Log(f, "debug", "Update %s %T", widget.ID(), widget)
		switch widget := widget.(type) {
		case *Input:
			value.SetString(widget.Text)
			f.Log(f, "debug", "Update %s %v %v", widget.ID(), values, f.Data)
		}
		return false
	}
}

func NewForm(id, title string, data any) *Form {
	return &Form{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Title:      title,
		Data:       data,
	}
}

type FormField struct {
	Name    string // Field name (same as struct field name)
	Control Widget // Form input control
	Label   string // Label
	x, y    int    // Label position
}

type FormGroup struct {
	BaseWidget
	Title     string         // Group title
	Placement string         // Label placement
	Spacing   int            // Line spacing
	fields    [][]*FormField // Lines of form fields
}

func NewFormGroup(id, title, placement string) *FormGroup {
	return &FormGroup{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Title:      title,
		Placement:  placement,
		fields:     [][]*FormField{},
	}
}

func (fg *FormGroup) Add(line int, label string, widget Widget) {
	if widget == nil {
		panic("Cannot add nil widget")
	}
	for line >= len(fg.fields) {
		fg.fields = append(fg.fields, make([]*FormField, 0, 1))
	}
	fg.fields[line] = append(fg.fields[line], &FormField{Label: label, Control: widget})
}

func (fg *FormGroup) Children(_ bool) []Widget {
	children := make([]Widget, 0, len(fg.fields))
	for _, line := range fg.fields {
		for _, field := range line {
			children = append(children, field.Control)
		}
	}
	return children
}

// Find searches for a widget with the specified ID within this box and its child widget.
// The search is performed recursively, first checking if this box matches the ID,
// then delegating to the generic Find function which will search the child widget
// and any of its descendants if the child is also a container.
//
// Parameters:
//   - id: The unique identifier of the widget to find
//   - visible: Only look for visible children
//
// Returns:
//   - Widget: The widget with the matching ID, or nil if not found
//
// The search order is:
//  1. Check if this box's ID matches
//  2. Recursively search within the child widget (if it exists)
//  3. If child is a container, search its descendants
func (fg *FormGroup) Find(id string, visible bool) Widget {
	return Find(fg, id, visible)
}

// FindAt searches for the widget located at the specified screen coordinates.
// This method is used for mouse interaction to determine which widget is
// positioned at a given point. The search includes this box and its child widget.
//
// Parameters:
//   - x: The x-coordinate to search at
//   - y: The y-coordinate to search at
//
// Returns:
//   - Widget: The widget at the specified coordinates, or nil if none found
//
// The search process:
//  1. Check if coordinates are within this box's bounds
//  2. If within bounds, recursively search the child widget
//  3. Return the most specific widget found at the coordinates
func (fg *FormGroup) FindAt(x, y int) Widget {
	return FindAt(fg, x, y)
}

func (fg *FormGroup) Hint() (int, int) {
	w, h := 0, 0 // preferred width and height total
	mlw := 0     // maximum label width in horizontal layout
	for i, line := range fg.fields {
		lw, lh := 0, 0 // line width and height
		for j, field := range line {
			// Check the maximum label width
			if j == 0 && fg.Placement == "horizontal" {
				if len(field.Label)+1 > mlw {
					mlw = len(field.Label) + 1
				}
			}

			// Determine field width and height including styling
			fw, fh := field.Control.Hint() // field width and height
			fw += field.Control.Style("").Horizontal()
			fh += field.Control.Style("").Vertical()
			fg.Log(fg, "debug", "  FormGroup %s %d.%d", field.Control.ID(), fw, fh)

			lw += fw
			if fh > lh {
				lh = fh
			}
			// Add 1 to the line width for every extra control as a spacing
			if j > 0 {
				lw++
			}
		}
		if fg.Placement == "vertical" {
			lh++
		}
		if i > 0 {
			lh++
		}
		if lw > w {
			w = lw
		}
		fg.Log(fg, "debug", "FormGroup line %d %d", i, w)
		h += lh
	}
	fg.Log(fg, "debug", "FormGroup Hint %d, %d", mlw+w, h)
	return mlw + w, h
}

func (fg *FormGroup) Layout() {
	fg.Log(fg, "debug", "FormGroup Layout %s", fg.Placement)
	// Determine maximum label width (only for horizontal label placement)
	mlw := -1
	if fg.Placement == "horizontal" {
		for _, line := range fg.fields {
			if len(line[0].Label) > mlw {
				mlw = len(line[0].Label)
			}
		}
	}
	fg.Log(fg, "debug", "Maximum label width %d", mlw)

	// Content area
	cx, cy, cw, _ := fg.Content()

	// Current line y position
	ly := cy
	for _, line := range fg.fields {
		// Current line x position
		lx := cx + mlw + 1
		lh := 0 // line height
		if fg.Placement == "vertical" {
			ly++
		}
		for j, field := range line {
			// Label position
			if fg.Placement == "horizontal" {
				if j == 0 {
					field.x, field.y = cx, ly
				} else {
					field.x, field.y = -1, -1
				}
			} else {
				field.x, field.y = lx, ly-1
			}
			fg.Log(fg, "debug", " -> %d.%d '%s'", field.x, field.y, field.Label)

			// Control position
			fw, fh := field.Control.Hint()
			fw += field.Control.Style("").Horizontal()
			fh += field.Control.Style("").Vertical()
			fg.Log(fg, "debug", " -> %s, lx, ly = %d, %d / fw, fh = %d, %d / cw, cx = %d, %d", field.Label, lx, ly, fw, fh, cw, cx)
			if j < len(line)-1 {
				field.Control.SetBounds(lx, ly, fw, fh)
				lx += fw + 1
			} else {
				field.Control.SetBounds(lx, ly, cw+cx-lx, fh)
			}

			// Update line height
			if fh > lh {
				lh = fh
			}
		}
		ly = ly + lh + fg.Spacing
	}
}

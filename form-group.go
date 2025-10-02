package zeichenwerk

// FormField represents a single field within a form group, consisting of a label and a control widget.
// It manages the layout and positioning of form elements.
type FormField struct {
	Name    string // Field name (same as struct field name)
	Control Widget // Form input control (Input, Checkbox, etc.)
	Label   string // Display label text
	x, y    int    // Label position coordinates for rendering
}

// FormGroup organizes related form fields into a cohesive layout with consistent labeling.
// It supports both horizontal and vertical label placement and manages the positioning
// of labels and controls within the group.
//
// FormGroup provides automatic layout management for form fields, handling:
//   - Label positioning (horizontal or vertical relative to controls)
//   - Field spacing and alignment
//   - Multi-line field arrangements
//   - Automatic sizing based on content
//
// Example usage:
//
//	group := NewFormGroup("contact-group", "Contact Information", "horizontal")
//	group.Add(0, "Name", nameInput)
//	group.Add(0, "Email", emailInput)  // Same line as name
//	group.Add(1, "Phone", phoneInput)  // New line
type FormGroup struct {
	BaseWidget
	Title     string         // Group title displayed in border or header
	Placement string         // Label placement: "horizontal" or "vertical"
	Spacing   int            // Vertical spacing between field lines
	fields    [][]*FormField // Fields organized by lines for layout
}

// NewFormGroup creates a new form group widget for organizing related form fields.
//
// Parameters:
//   - id: Unique identifier for the form group
//   - title: Display title for the group (can be empty)
//   - placement: Label placement relative to controls ("horizontal" or "vertical")
//   - "horizontal": Labels appear to the left of controls, aligned in columns
//   - "vertical": Labels appear above controls
//
// Returns:
//   - *FormGroup: New form group widget instance
//
// Example:
//
//	// Horizontal layout - labels left of controls
//	group := NewFormGroup("user-info", "User Information", "horizontal")
//
//	// Vertical layout - labels above controls
//	group := NewFormGroup("settings", "Settings", "vertical")
func NewFormGroup(id, title, placement string) *FormGroup {
	return &FormGroup{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Title:      title,
		Placement:  placement,
		fields:     [][]*FormField{},
	}
}

// Add adds a form field to the specified line within the form group.
// Multiple fields can be added to the same line for horizontal grouping.
//
// Parameters:
//   - line: Line number for field placement (0-based, auto-extends as needed)
//   - label: Display label for the field (can be empty)
//   - widget: Form control widget (Input, Checkbox, etc.)
//
// The method automatically extends the internal field structure as needed
// to accommodate the specified line number.
//
// Example:
//
//	group.Add(0, "First Name", firstNameInput)
//	group.Add(0, "Last Name", lastNameInput)   // Same line
//	group.Add(1, "Email", emailInput)          // New line
//	group.Add(1, "Phone", phoneInput)          // Same line as email
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
			style := field.Control.Style()
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
		if fg.Placement == "vertical" {
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
	if fg.Placement == "horizontal" {
		for _, line := range fg.fields {
			if len(line[0].Label) > mlw {
				mlw = len(line[0].Label)
			}
		}
	}

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

			// Control position
			fw, fh := field.Control.Hint()
			style := field.Control.Style()
			fw += style.Horizontal()
			fh += style.Vertical()
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

package zeichenwerk

import "strings"

// Dialog is a single-child container designed for overlay use. It renders a
// title bar, an optional border, and its body widget. Dialogs are typically
// shown as popups via UI.Popup rather than added to the main layout tree.
type Dialog struct {
	Component
	title string // The title text displayed in the dialog header (optional)
	child Widget // Dialog content
}

// NewDialog creates a Dialog with the given id, CSS class, and title.
// An empty title string suppresses the title bar.
func NewDialog(id, class, title string) *Dialog {
	return &Dialog{
		Component: Component{id: id, class: class},
		title:     title,
	}
}

// Add sets the single body widget, replacing any previous child.
func (d *Dialog) Add(widget Widget, params ...any) error {
	if d.child != nil {
		d.child.SetParent(nil) // clear old parent reference
	}
	if widget != nil {
		widget.SetParent(d)
	}
	d.child = widget
	return nil
}

// Apply applies a theme style to the component.
func (d *Dialog) Apply(theme *Theme) {
	theme.Apply(d, d.Selector("dialog"))
}

// Children returns the body widget slice (empty if no child has been set).
func (d *Dialog) Children() []Widget {
	if d.child == nil {
		return []Widget{}
	}
	return []Widget{d.child}
}

// Hint returns the preferred size of the dialog. If an explicit hint has been
// set it is returned directly; otherwise the size is derived from the child's
// hint plus the dialog's own style overhead (border, padding) and title bar.
func (d *Dialog) Hint() (int, int) {
	if d.hwidth != 0 && d.hheight != 0 {
		d.Log(d, Debug, "Dialog Fixed Hint", "w", d.hwidth, "h", d.hheight)
		return d.hwidth, d.hheight
	} else if d.child != nil {
		w, h := d.child.Hint()
		d.Log(d, Debug, "Dialog dynamic Hint 1", "w", w, "h", h)
		style := d.child.Style()
		w += style.Horizontal()
		h += style.Vertical()
		d.Log(d, Debug, "Dialog dynamic Hint 2", "w", w, "h", h)

		if d.title != "" {
			titleStyle := d.Style("title")
			h = h + titleStyle.Vertical() + 1
			d.Log(d, Debug, "Dialog title Vertical", "h", titleStyle.Vertical())
		}

		d.Log(d, Debug, "Dialog dynamic Hint 3", "w", w, "h", h)
		return w, h
	} else {
		return 0, 0
	}
}

// Layout positions the body widget below the title bar within the content area.
func (d *Dialog) Layout() error {
	if d.child != nil {
		cx, cy, cw, ch := d.Content()
		if d.title != "" {
			style := d.Style("title")
			cy = cy + style.Vertical() + 1
		}
		d.child.SetBounds(cx, cy, cw, ch)
	}
	return Layout(d)
}

// Render draws the dialog: title bar, background fill, border, and body widget.
func (d *Dialog) Render(r *Renderer) {
	// Check if the widget is visible
	if d.Flag(FlagHidden) {
		return
	}

	// Determine the style to use based on the widget state
	state := d.State()
	if state != "" {
		state = ":" + state
	}
	style := d.Style(state)
	r.Set(style.Foreground(), style.Background(), style.Font())

	// Render the title
	oy := 0
	if d.title != "" {
		titleStyle := d.Style("title" + state)
		r.Set(titleStyle.Foreground(), titleStyle.Background(), titleStyle.Font())

		// Use dialog style margin for positioning
		r.Fill(d.x+style.Margin().Left, d.y+style.Margin().Top, d.width-style.Margin().Horizontal(), 1, " ")
		r.Text(d.x+style.Margin().Left+titleStyle.Padding().Left, d.y+style.Margin().Top+titleStyle.Padding().Top, d.title, 0)
		oy = titleStyle.Padding().Vertical() + 1
	}

	// Clear the content area
	r.Set(style.Foreground(), style.Background(), style.Font())
	r.Fill(d.x+style.Margin().Left, d.y+style.Margin().Top+oy, d.width-style.Margin().Horizontal(), d.height+style.Margin().Vertical()-oy, " ")

	// Draw the dialog border
	border := style.Border()
	if border != "" && border != "none" {
		parts := strings.Split(border, " ")
		if len(parts) > 1 {
			border = parts[0]
			fg := parts[1]
			bg := style.Background()
			if len(parts) > 2 {
				bg = parts[2]
			}
			r.Set(fg, bg, "")
		} else {
			r.Set(style.Foreground(), style.Background(), "")
		}
		margin := style.Margin()
		r.Border(d.x+margin.Left, d.y+margin.Top+oy, d.width-margin.Horizontal(), d.height-margin.Vertical()-oy, border)
		r.Set(style.Foreground(), style.Background(), style.Font())
	}

	// Render the child
	if d.child != nil {
		d.child.Render(r)
	}
}

package zeichenwerk

import "github.com/gdamore/tcell/v3"

// Collapsible is a single-child container with a clickable header that expands
// and collapses the body. When collapsed, only the header row is visible and
// the widget's height hint shrinks accordingly, causing the parent layout to
// reclaim the freed space.
type Collapsible struct {
	Component
	title    string // Header label
	child    Widget // Body content (nil until Add is called)
	expanded bool   // Whether the body is currently visible
}

// NewCollapsible creates a new collapsible container.
//
// Parameters:
//   - id: Unique identifier for the widget
//   - class: CSS-like class name for styling
//   - title: Label displayed in the header row
//   - expanded: Initial expanded state
func NewCollapsible(id, class, title string, expanded bool) *Collapsible {
	c := &Collapsible{
		Component: Component{id: id, class: class},
		title:     title,
		expanded:  expanded,
	}
	c.SetFlag(FlagFocusable, true)
	OnKey(c, c.handleKey)
	OnMouse(c, c.handleMouse)
	return c
}

// Add sets the single child widget. Calling Add again replaces any existing
// child. The child is hidden immediately when the collapsible is collapsed.
func (c *Collapsible) Add(widget Widget, params ...any) error {
	if widget == nil {
		return ErrChildIsNil
	}
	if c.child != nil {
		c.child.SetParent(nil)
	}
	c.child = widget
	c.child.SetParent(c)
	c.child.SetFlag(FlagHidden, !c.expanded)
	return nil
}

// Apply applies the collapsible and header styles from the theme.
func (c *Collapsible) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("collapsible"), "focused", "hovered")
	theme.Apply(c, c.Selector("collapsible/header"), "focused", "hovered")
}

// Children returns a slice containing the child widget, or an empty slice.
func (c *Collapsible) Children() []Widget {
	if c.child == nil {
		return []Widget{}
	}
	return []Widget{c.child}
}

// Hint returns the preferred content size.
//
// Collapsed: (childW, 1) — header row only.
// Expanded:  (childW, 1 + childH) — header plus child.
//
// Style overhead (border, padding, margin) of the collapsible itself is added
// by the parent layout engine, consistent with Box.Hint().
func (c *Collapsible) Hint() (int, int) {
	if c.hwidth != 0 || c.hheight != 0 {
		return c.hwidth, c.hheight
	}
	childW, childH := 0, 0
	if c.child != nil {
		childW, childH = c.child.Hint()
	}
	if c.expanded {
		if childH == 0 {
			return childW, -1 // fractional: take remaining space when child has no fixed hint
		}
		return childW, 1 + childH
	}
	return childW, 1
}

// Layout positions the child within the body area (below the header row).
func (c *Collapsible) Layout() error {
	cx, cy, cw, ch := c.Content()
	if c.expanded && c.child != nil {
		c.child.SetBounds(cx, cy+1, cw, ch-1)
	}
	return Layout(c)
}

// Render draws the collapsible: background/border, header row, and (if
// expanded) the child widget.
func (c *Collapsible) Render(r *Renderer) {
	c.Component.Render(r)

	state := c.State()
	headerSelector := "header"
	if state != "" {
		headerSelector = "header:" + state
	}
	headerStyle := c.Style(headerSelector)

	cx, cy, cw, _ := c.Content()
	r.Set(headerStyle.Foreground(), headerStyle.Background(), headerStyle.Font())

	indicator := r.theme.String("collapsible.collapsed")
	if c.expanded {
		indicator = r.theme.String("collapsible.expanded")
	}
	r.Text(cx, cy, indicator+c.title, cw)

	if c.expanded && c.child != nil {
		c.child.Render(r)
	}
}

// Expand shows the body. No-op if already expanded.
func (c *Collapsible) Expand() {
	c.expanded = true
	if c.child != nil {
		c.child.SetFlag(FlagHidden, false)
	}
	Relayout(c)
	c.Dispatch(c, EvtChange, c.expanded)
}

// Collapse hides the body. No-op if already collapsed.
// If focus lives inside the child subtree it is moved to the Collapsible so
// that the hidden child no longer receives key events or draws a focus style.
func (c *Collapsible) Collapse() {
	c.expanded = false
	if c.child != nil {
		if focusedIn(c.child) {
			if ui := FindUI(c); ui != nil {
				ui.Focus(c)
			}
		}
		c.child.SetFlag(FlagHidden, true)
	}
	Relayout(c)
	c.Dispatch(c, EvtChange, c.expanded)
}

// focusedIn reports whether widget or any of its descendants carries FlagFocused.
func focusedIn(widget Widget) bool {
	if widget.Flag(FlagFocused) {
		return true
	}
	if container, ok := widget.(Container); ok {
		for _, child := range container.Children() {
			if focusedIn(child) {
				return true
			}
		}
	}
	return false
}

// Toggle switches between expanded and collapsed states.
func (c *Collapsible) Toggle() {
	if c.expanded {
		c.Collapse()
	} else {
		c.Expand()
	}
}

// Expanded reports whether the body is currently visible.
func (c *Collapsible) Expanded() bool {
	return c.expanded
}

func (c *Collapsible) handleKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEnter:
		c.Toggle()
		return true
	case tcell.KeyRune:
		if ev.Str() == " " {
			c.Toggle()
			return true
		}
	case tcell.KeyRight:
		if !c.expanded {
			c.Expand()
		}
		return true
	case tcell.KeyLeft:
		if c.expanded {
			c.Collapse()
		}
		return true
	}
	return false
}

func (c *Collapsible) handleMouse(ev *tcell.EventMouse) bool {
	_, my := ev.Position()
	style := c.Style()
	headerY := c.y + style.Margin().Top
	if my == headerY {
		switch ev.Buttons() {
		case tcell.Button1:
			c.Toggle()
			return true
		}
	}
	return false
}

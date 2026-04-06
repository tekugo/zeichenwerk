package zeichenwerk

import "github.com/gdamore/tcell/v3"

// Viewport is a scrollable single-child container. The child widget is given
// its full preferred size (from Hint) and the viewport shows a windowed view
// into it with horizontal and vertical scrollbars when needed.
//
// Set FlagVertical to restrict scrolling to the vertical axis (child fills
// viewport width). Set FlagHorizontal for horizontal-only scrolling (child
// fills viewport height). Without either flag both axes scroll freely.
type Viewport struct {
	Component
	Title  string // Optional title text to display in the border
	child  Widget // The single child widget contained within this viewport
	tx, ty int    // Current horizontal and vertical scroll offsets
}

// NewViewport creates a Viewport with the given id, CSS class, and optional
// border title. The widget is focusable and handles arrow keys for scrolling.
func NewViewport(id, class, title string) *Viewport {
	viewport := &Viewport{
		Component: Component{id: id, class: class},
		Title:     title,
	}
	viewport.SetFlag(FlagFocusable, true)
	OnKey(viewport, viewport.handleKey)
	return viewport
}

// Add sets the single child widget, replacing any previous child.
func (v *Viewport) Add(widget Widget, params ...any) error {
	if widget == nil {
		return ErrChildIsNil
	}
	if v.child != nil {
		v.child.SetParent(nil)
	}
	v.child = widget
	v.child.SetParent(v)
	return nil
}

// Apply applies a theme's styles to the component.
func (v *Viewport) Apply(theme *Theme) {
	theme.Apply(v, v.Selector("viewport"))
}

// Children returns the child widget slice (empty if no child has been set).
func (v *Viewport) Children() []Widget {
	if v.child == nil {
		return []Widget{}
	}
	return []Widget{v.child}
}

// Layout positions the child at its full preferred size offset by the current
// scroll position. For FlagVertical/FlagHorizontal the child is constrained
// to fill the perpendicular axis so it never needs to scroll that direction.
func (v *Viewport) Layout() error {
	if v.child != nil {
		cx, cy, vw, vh := v.Content()
		pw, ph := v.child.Hint()
		switch {
		case v.Flag(FlagVertical):
			// vertical only: child fills viewport width
			iw := vw
			if ph > vh {
				iw-- // reserve right column for V-scrollbar
			}
			v.child.SetBounds(cx, cy-v.ty, iw, ph)
		case v.Flag(FlagHorizontal):
			// horizontal only: child fills viewport height
			ih := vh
			if pw > vw {
				ih-- // reserve bottom row for H-scrollbar
			}
			v.child.SetBounds(cx-v.tx, cy, pw, ih)
		default:
			v.child.SetBounds(cx-v.tx, cy-v.ty, pw, ph)
		}
	}
	return Layout(v)
}

func (v *Viewport) handleKey(event *tcell.EventKey) bool {
	if v.child == nil {
		return false
	}

	cw, ch, _, _ := v.Content() // Content area size
	pw, ph := v.child.Hint()    // Child widget preferred size

	// Calculate maximum scroll offsets
	maxTx := max(pw-cw, 0)
	maxTy := max(ph-ch, 0)

	switch event.Key() {
	case tcell.KeyUp:
		if v.Flag(FlagHorizontal) {
			return false
		}
		if v.ty > 0 {
			v.ty--
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyDown:
		if v.Flag(FlagHorizontal) {
			return false
		}
		if v.ty < maxTy {
			v.ty++
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyLeft:
		if v.Flag(FlagVertical) {
			return false
		}
		if v.tx > 0 {
			v.tx--
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyRight:
		if v.Flag(FlagVertical) {
			return false
		}
		if v.tx < maxTx {
			v.tx++
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyHome:
		newTx, newTy := 0, 0
		if v.Flag(FlagVertical) {
			newTx = v.tx // don't touch horizontal
		}
		if v.Flag(FlagHorizontal) {
			newTy = v.ty // don't touch vertical
		}
		if v.tx != newTx || v.ty != newTy {
			v.tx = newTx
			v.ty = newTy
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyEnd:
		newTx, newTy := maxTx, maxTy
		if v.Flag(FlagVertical) {
			newTx = v.tx // don't touch horizontal
		}
		if v.Flag(FlagHorizontal) {
			newTy = v.ty // don't touch vertical
		}
		if v.tx != newTx || v.ty != newTy {
			v.tx = newTx
			v.ty = newTy
			v.Layout()
			v.Refresh()
			return true
		}
	}

	return false
}

// Refresh triggers a redraw of the viewport.
func (v *Viewport) Refresh() {
	Redraw(v)
}

// Render draws the viewport background and border, then clips to the content
// area and renders the child at its scrolled position. Scrollbars are drawn
// when the child is larger than the visible area.
func (v *Viewport) Render(r *Renderer) {
	// Render styling and border
	v.Component.Render(r)

	// Get the viewport's content area coordinates and dimensions
	x, y, w, h := v.Content()

	// Get the child widget's total bounds to determine content size
	_, _, cw, ch := v.child.Bounds()

	// ---- Scrollbar Necessity Calculation ----

	var iw, ih int
	switch {
	case v.Flag(FlagVertical):
		// Only vertical scrollbar; child fills full width
		iw = w
		if ch > h {
			iw-- // reserve right column for V-scrollbar
		}
		ih = h
	case v.Flag(FlagHorizontal):
		// Only horizontal scrollbar; child fills full height
		iw = w
		ih = h
		if cw > w {
			ih-- // reserve bottom row for H-scrollbar
		}
	default:
		// Both axes: calculate mutual scrollbar dependencies
		iw = w
		if ch > h {
			iw--
		}
		ih = h
		if cw > iw {
			ih--
		}
	}

	// Render vertical scrollbar when content is taller than visible area
	if !v.Flag(FlagHorizontal) && ch > h {
		r.ScrollbarV(x+w-1, y, ih, v.ty, ch)
	}

	// Render horizontal scrollbar when content is wider than visible area
	if !v.Flag(FlagVertical) && cw > iw {
		r.ScrollbarH(x, y+h-1, iw, v.tx, cw)
	}

	r.Clip(x, y, iw, ih)
	r.Translate(-v.tx, -v.ty)
	v.child.SetBounds(-v.tx, -v.ty, cw, ch)
	v.child.Render(r)
	r.Clip(0, 0, 0, 0)
	r.Translate(0, 0)
}

package zeichenwerk

import "github.com/gdamore/tcell/v3"

// Viewport is a scrollable single-child container. The child widget is given
// its full preferred size (from Hint) and the viewport shows a windowed view
// into it with horizontal and vertical scrollbars when needed.
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
// scroll position.
func (v *Viewport) Layout() error {
	if v.child != nil {
		cx, cy, _, _ := v.Content()
		pw, ph := v.child.Hint()
		v.child.SetBounds(cx-v.tx, cy-v.ty, pw, ph)
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
		// Scroll up by one line
		if v.ty > 0 {
			v.ty--
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyDown:
		// Scroll down by one line
		if v.ty < maxTy {
			v.ty++
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyLeft:
		// Scroll left by one character
		if v.tx > 0 {
			v.tx--
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyRight:
		// Scroll right by one character
		if v.tx < maxTx {
			v.tx++
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyHome:
		// Reset to top-left corner
		if v.tx > 0 || v.ty > 0 {
			v.tx = 0
			v.ty = 0
			v.Layout()
			v.Refresh()
			return true
		}

	case tcell.KeyEnd:
		// Move to bottom-right corner
		if v.tx < maxTx || v.ty < maxTy {
			v.tx = maxTx
			v.ty = maxTy
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

	// Calculate available width (iw) considering vertical scrollbar space
	// Start with full width, reduce by 1 if vertical scrollbar is needed
	iw := w
	if ch > h {
		iw--
	}

	// Calculate available height (ih) considering horizontal scrollbar space
	// Use adjusted width (iw) to account for vertical scrollbar space
	ih := h
	if cw > iw {
		ih--
	}

	// Render vertical scrollbar if width was reduced (indicates necessity)
	if iw < w {
		r.ScrollbarV(x+w-1, y, ih, v.ty, ch)
	}

	// Render horizontal scrollbar if height was reduced (indicates necessity)
	if ih < h {
		r.ScrollbarH(x, y+h-1, iw, v.tx, cw)
	}

	r.Clip(x, y, iw, ih)
	r.Translate(-v.tx, -v.ty)
	v.child.SetBounds(-v.tx, -v.ty, cw, ch)
	v.child.Render(r)
	r.Clip(0, 0, 0, 0)
	r.Translate(0, 0)
}

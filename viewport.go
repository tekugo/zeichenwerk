package next

import "github.com/gdamore/tcell/v3"

type Viewport struct {
	Component
	Title  string // Optional title text to display in the border
	child  Widget // The single child widget contained within this viewport
	tx, ty int    // Current horizontal and vertical scroll offsets
}

func NewViewport(id, title string) *Viewport {
	viewport := &Viewport{
		Component: Component{id: id},
		Title:     title,
	}
	viewport.SetFlag("focusable", true)
	OnKey(viewport, viewport.handleKey)
	return viewport
}

func (v *Viewport) Add(widget Widget) {
	if v.child != nil {
		v.child.SetParent(nil)
	}
	v.child = widget
	widget.SetParent(v)
}

func (v *Viewport) Children() []Widget {
	if v.child == nil {
		return []Widget{}
	}
	return []Widget{v.child}
}

func (v *Viewport) Layout() {
	if v.child != nil {
		cx, cy, _, _ := v.Content()
		pw, ph := v.child.Hint()
		v.child.SetBounds(cx-v.tx, cy-v.ty, pw, ph)
	}
	Layout(v)
}

func (v *Viewport) handleKey(_ Widget, event *tcell.EventKey) bool {
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

func (v *Viewport) Refresh() {
	Redraw(v)
}

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

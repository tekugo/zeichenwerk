package zeichenwerk

func (r *Renderer) renderScroller(scroller *Scroller) {
	x, y, w, h := scroller.Content()
	_, _, cw, ch := scroller.child.Bounds()

	// Check, if we need to render a vertical scroll bar
	iw := w
	if ch > h {
		iw--
	}

	// Check, if we need to render a horizontal scroll bar
	ih := h
	if cw > iw {
		ih--
	}

	if iw < w {
		r.renderScrollbarV(x+w-1, y, ih, scroller.ty, ch)
	}

	if ih < h {
		r.renderScrollbarH(x, y+h-1, iw, scroller.tx, cw)
	}

	// Render visible text content
	r.clip(scroller)
	if viewport, ok := r.screen.(*Viewport); ok {
		viewport.tx = x - scroller.tx
		viewport.ty = y - scroller.ty
		viewport.width = iw
		viewport.height = ih
	} else {
		scroller.Log(scroller, "error", "Cannot translate viewport %T", r.screen)
	}
	scroller.Log(scroller, "debug", "renderScroller x=%d, y=%d, tx=%d, ty=%d", x, y, scroller.tx, scroller.ty)
	r.render(scroller.child)
	r.unclip()
}

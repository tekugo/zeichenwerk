package zeichenwerk

func (r *Renderer) renderTabs(tabs *Tabs, x, y, w int) {
	// Determine which styles to use based on focus state
	var normal, highlight, line *Style

	if tabs.Focused() {
		// Use focus-specific styles when tabs widget has focus
		normal = tabs.Style("line:focus")
		if normal == nil {
			normal = tabs.Style("")
		}
		highlight = tabs.Style("highlight:focus")
		if highlight == nil {
			highlight = tabs.Style("highlight")
		}
		line = tabs.Style("highlight-line:focus")
		if line == nil {
			line = tabs.Style("highlight-line")
		}
	} else {
		// Use normal styles when tabs widget doesn't have focus
		normal = tabs.Style("")
		highlight = tabs.Style("highlight")
		line = tabs.Style("highlight-line")
	}

	cx := x
	r.SetStyle(normal)
	r.repeat(x, y+1, 1, 0, 1, '\u2501')

	for i, tab := range tabs.Tabs {
		tl := len([]rune(tab))
		if tabs.Index == i {
			r.SetStyle(highlight)
			r.text(cx+1, y, " "+tab+" ", 0)
			r.SetStyle(line)
			r.repeat(cx, y+1, 1, 0, min(tl+4, x+cx), '\u2501')
			r.SetStyle(normal)
		} else {
			r.text(cx+1, y, " "+tab+" ", w-cx)
			r.repeat(cx, y+1, 1, 0, min(tl+4, x+cx), '\u2501')
		}
		cx = cx + tl + 4
		if cx > x+w {
			break
		}
	}

	if cx < x+w {
		r.repeat(cx, y+1, 1, 0, x+w-cx, '\u2501')
	}
}

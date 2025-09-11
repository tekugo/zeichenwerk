package zeichenwerk

func (r *Renderer) renderTabs(tabs *Tabs, x, y, w int) {
	normal := tabs.Style("")
	highlight := tabs.Style("highlight")
	line := tabs.Style("highlight-line")

	cx := x
	r.SetStyle(normal)
	r.repeat(x, y+1, 1, 0, 1, '\u2501')

	for i, tab := range tabs.Tabs {
		tabs.Log("Rendering %d %s %d %d", i, tab, tabs.Index, cx)
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

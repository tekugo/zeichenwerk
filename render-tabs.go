package zeichenwerk

func (r *Renderer) renderTabs(tabs *Tabs, x, y, w int) {
	normal := tabs.Style("")
	highlight := tabs.Style("highlight")
	line := tabs.Style("highlight-line")

	cx := x
	r.SetStyle(normal)
	r.repeat(x, y+1, 1, 0, 1, '\u2501')

	for i, tab := range tabs.Tabs {
		tl := len([]rune(tab))
		if tabs.Index == i {
			r.SetStyle(highlight)
			r.text(x+cx, y, " "+tab+" ", w-cx)
			r.SetStyle(line)
			r.repeat(x+cx, y+1, 1, 0, min(tl+2, x+cx), '\u2501')
			r.SetStyle(normal)
		} else {
			r.text(x+cx, y, " "+tab+" ", w-cx)
			r.repeat(x+cx, y+1, 1, 0, min(tl+2, x+cx), '\u2501')
		}
		cx = cx + tl + 2
		if cx > w {
			break
		}
	}
}

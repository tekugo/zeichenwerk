package zeichenwerk

func (r *Renderer) renderTableHeader(table *Table, x, y, w, h int, headerStyle, gridStyle *Style) {
	rx := 0 // current render position
	rw := w // remaining width
	columns := table.provider.Columns()
	for i, column := range columns {
		// If there is no room remaing, we stop rendering the header
		if rw <= 0 {
			break
		}
		// Check, if the column is visible on the left
		if rx+column.Width < table.offsetX {
			rx = rx + column.Width + 1
			continue
		}
		// Do we need to render it partially?
		r.SetStyle(headerStyle)
		// cell/column x-position
		cx := x - table.offsetX + rx
		if rx < table.offsetX {
			start := table.offsetX - rx
			rcw := column.Width - start
			runes := []rune(column.Header)
			if start < len(runes) {
				r.text(cx+start, y, string(runes[start:]), rcw)
			}
			r.SetStyle(gridStyle)
			r.repeat(cx+start, y+1, 1, 0, rcw, table.grid.InnerH)
			rw += start
		} else {
			// We render either to complete column or use the remaining width
			cw := min(rw, column.Width)
			r.text(cx, y, column.Header, cw)
			r.SetStyle(gridStyle)
			r.repeat(cx, y+1, 1, 0, cw, table.grid.InnerH)
		}
		// Do we need to render the column separator?
		if i < len(columns)-1 && rw > column.Width {
			r.screen.SetContent(cx+column.Width, y+1, table.grid.InnerTopT, nil, r.style)
			// Bottom T is rendered with the header
			if table.outer {
				r.screen.SetContent(cx+column.Width, y+h, table.grid.BottomT, nil, r.style)
			}
		}
		rx = rx + column.Width + 1
		rw = rw - column.Width - 1
	}

	// Draw the remaining header line
	r.SetStyle(gridStyle)
	if rx-table.offsetX < w {
		r.repeat(x-table.offsetX+rx-1, y+1, 1, 0, w+table.offsetX-rx+1, table.grid.InnerH)
	}
	// Draw the left and right T connectors
	if table.outer {
		r.screen.SetContent(x-1, y+1, table.grid.LeftT, nil, r.style)
		r.screen.SetContent(x+w, y+1, table.grid.RightT, nil, r.style)
	}
}

func (r *Renderer) renderTableContent(table *Table, x, y, w, h int, gridStyle *Style) {
	row := table.offsetY
	columns := table.provider.Columns()
	// Traverse all visible rows
	for row < table.provider.Length() && row-table.offsetY < h {
		// Determine row style
		var style *Style
		if row == table.row && table.focused {
			style = table.Style("highlight:focus")
		} else if row == table.row {
			style = table.Style("highlight")
		} else {
			style = table.Style("")
		}

		rx := 0 // render x position
		rw := w // remaining width
		for i, column := range columns {
			if rw <= 0 {
				break
			}
			if rx+column.Width < table.offsetX {
				rx = rx + column.Width + 1
				continue
			}
			r.SetStyle(style)
			cx := x - table.offsetX + rx
			cy := y - table.offsetY + row
			if rx < table.offsetX {
				start := table.offsetX - rx
				rcw := column.Width - start
				runes := []rune(table.provider.Str(row, i))
				if start < len(runes) {
					r.text(cx+start, cy, string(runes[start:]), rcw)
				} else {
					r.repeat(cx+start, cy, 1, 0, rcw, ' ')
				}
				rw += start
			} else {
				cw := min(rw, column.Width)
				r.text(cx, cy, table.provider.Str(row, i), cw)
			}
			if i < len(columns)-1 && table.inner && rw > column.Width {
				if row != table.row {
					r.SetStyle(gridStyle)
				}
				r.screen.SetContent(cx+column.Width, cy, table.grid.InnerV, nil, r.style)
			}
			rx = rx + column.Width + 1
			rw = rw - column.Width - 1
		}
		row++
	}
}

func (r *Renderer) renderTable(table *Table) {
	x, y, w, h := table.Content()
	headerStyle := table.Style("header")
	gridStyle := table.Style("grid")
	if gridStyle.Border != "" {
		table.Log(table, "debug", "Grid border is %s", gridStyle.Border)
		table.grid = r.theme.Border(gridStyle.Border)
		r.renderTableHeader(table, x, y, w, h, headerStyle, gridStyle)
		r.renderTableContent(table, x, y+2, w, h-2, gridStyle)
	} else {
		table.Log(table, "debug", "Grid border is not set")
		r.renderTableHeader(table, x, y, w, h, headerStyle, gridStyle)
		r.renderTableContent(table, x, y+2, w, h-2, gridStyle)
	}
}

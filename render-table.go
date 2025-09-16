package zeichenwerk

func (r *Renderer) renderTableHeader(table *Table, x, y, w, h int) {
	rx := 0 // current render position
	rw := w // remaining width
	for _, column := range table.provider.Columns() {
		table.Log(table, "debug", "Render header %s %d %d", column.Header, rx, rw)
		// Check, if the column is visible
		if rx+column.Width < table.offsetX {
			rx = rx + column.Width + 1
			continue
		}
		// Do we need to render it partially?
		if rx < table.offsetX {
			start := table.offsetX - rx
			rcw := column.Width - start
			r.text(x-table.offsetX+rx, y, column.Header[start:], rcw)
			rw = rw - rcw - 1
		} else {
			r.text(x-table.offsetX+rx, y, column.Header, column.Width)
			rw = rw - column.Width - 1
		}
		rx = rx + column.Width + 1
	}
}

func (r *Renderer) renderTableContent(table *Table, x, y, w, h int, separator rune) {
	row := table.offsetY
	for row < table.provider.Length() && row-table.offsetY < h {
		rx := 0
		if row == table.row && table.focused {
			r.SetStyle(table.Style("highlight:focus"))
		} else if row == table.row {
			r.SetStyle(table.Style("highlight"))
		} else {
			r.SetStyle(table.Style(""))
		}
		for i, column := range table.provider.Columns() {
			if rx+column.Width < table.offsetX {
				rx = rx + column.Width + 1
				continue
			}
			if rx < table.offsetX {
			} else {
				r.text(x-table.offsetX+rx, y-table.offsetY+row, table.provider.Str(row, i), column.Width)
				r.screen.SetContent(x-table.offsetX+rx+column.Width, y-table.offsetY+row, separator, nil, r.style)
			}
			rx = rx + column.Width + 1
		}
		row++
	}
}

func (r *Renderer) renderTable(table *Table) {
	x, y, w, h := table.Content()
	style := table.Style("")
	if style.Border != "" {
		border := r.theme.Border(style.Border)
		r.renderTableHeader(table, x, y, w, 1)
		r.line(x-1, y+1, 1, 0, w, border.LeftT, border.InnerH, border.RightT)
		r.renderTableContent(table, x, y+2, w, h-2, border.InnerV)
	} else {
		r.renderTableHeader(table, x, y, w, 1)
		r.line(x, y+1, 1, 0, w-2, '-', '-', '-')
		r.renderTableContent(table, x, y+2, w, h-2, ' ')
	}
}

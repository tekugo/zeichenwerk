package zeichenwerk

var emptyBorder = BorderStyle{InnerH: ' ', InnerV: '-'}

func (r *Renderer) renderTableHeader(table *Table, x, y, w, h int, grid bool, border BorderStyle) {
	rx := 0 // current render position
	rw := w // remaining width
	columns := table.provider.Columns()
	for i, column := range columns {
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
			r.repeat(x-table.offsetX+rx, y+1, 1, 0, column.Width, border.InnerH)
			if i < len(columns)-1 && grid {
				r.screen.SetContent(x-table.offsetX+rx+column.Width, y+1, border.InnerTopT, nil, r.style)
				r.screen.SetContent(x-table.offsetX+rx+column.Width, y+h, border.BottomT, nil, r.style)
			}
			rw = rw - column.Width - 1
		}
		rx = rx + column.Width + 1
	}

	// Draw the remaining header line
	if rx-table.offsetX < w {
		r.repeat(x-table.offsetX+rx-1, y+1, 1, 0, w+table.offsetX-rx+1, border.InnerH)
	}
	if grid {
		r.screen.SetContent(x-1, y+1, border.LeftT, nil, r.style)
		r.screen.SetContent(x+w, y+1, border.RightT, nil, r.style)
	}
}

func (r *Renderer) renderTableContent(table *Table, x, y, w, h int, separator rune) {
	row := table.offsetY
	columns := table.provider.Columns()
	for row < table.provider.Length() && row-table.offsetY < h {
		rx := 0
		if row == table.row && table.focused {
			r.SetStyle(table.Style("highlight:focus"))
		} else if row == table.row {
			r.SetStyle(table.Style("highlight"))
		} else {
			r.SetStyle(table.Style(""))
		}
		for i, column := range columns {
			if rx+column.Width < table.offsetX {
				rx = rx + column.Width + 1
				continue
			}
			if rx < table.offsetX {
			} else {
				r.text(x-table.offsetX+rx, y-table.offsetY+row, table.provider.Str(row, i), column.Width)
				if i < len(columns)-1 {
					r.screen.SetContent(x-table.offsetX+rx+column.Width, y-table.offsetY+row, separator, nil, r.style)
				}
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
		r.renderTableHeader(table, x, y, w, h, true, border)
		r.renderTableContent(table, x, y+2, w, h-2, border.InnerV)
	} else {
		r.renderTableHeader(table, x, y, w, h, false, emptyBorder)
		r.renderTableContent(table, x, y+2, w, h-2, ' ')
	}
}

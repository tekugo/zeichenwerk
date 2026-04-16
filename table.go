package zeichenwerk

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

// Table displays tabular data with scrolling, keyboard navigation, and theming.
// It uses a TableProvider to supply data. Supports row mode and cell navigation mode.
type Table struct {
	Component
	provider         TableProvider
	tableWidth       int
	row, column      int
	offsetX, offsetY int
	grid             *Border
	inner, outer     bool
	cellNav          bool // false = row navigation mode, true = cell navigation mode
	cellStyler       func(row, col int, highlight bool) *Style
}

// NewTable creates a new table widget with the given ID, class, and data provider.
func NewTable(id, class string, provider TableProvider, cellNav bool) *Table {
	table := &Table{
		Component: Component{id: id, class: class},
		grid:      &Border{InnerH: "-", InnerV: "|"},
		inner:     true,
		outer:     true,
		cellNav:   cellNav,
	}
	table.SetFlag(FlagFocusable, true)
	table.Set(provider)
	OnKey(table, table.handleKey)
	return table
}

// Apply applies theme styles to the table and its sub-selectors.
func (t *Table) Apply(theme *Theme) {
	theme.Apply(t, t.Selector("table"), "disabled", "focused")
	theme.Apply(t, t.Selector("table/grid"), "disabled", "focused")
	theme.Apply(t, t.Selector("table/header"), "disabled", "focused")
	theme.Apply(t, t.Selector("table/highlight"), "disabled", "focused")
	theme.Apply(t, t.Selector("table/cell"), "disabled", "focused")
}

// Refresh triggers a redraw of the table.
func (t *Table) Refresh() {
	Redraw(t)
}

// Set updates the data provider and recalculates the total table width.
func (t *Table) Set(value TableProvider) {
	t.provider = value
	t.tableWidth = 0
	columns := value.Columns()
	for _, column := range columns {
		t.tableWidth += column.Width
	}
	t.tableWidth += len(columns) - 1
}

// Hint returns the preferred size of the table.
func (t *Table) Hint() (int, int) {
	if t.hwidth != 0 || t.hheight != 0 {
		return t.hwidth, t.hheight
	}

	w := 0
	h := t.provider.Length()
	for i, column := range t.provider.Columns() {
		if i > 0 {
			w++
		}
		w += column.Width
	}
	return w, h
}

// Selected returns the current (row, col) selection.
// In row mode col is always -1; in cell mode col is the active column index.
func (t *Table) Selected() (int, int) {
	if t.provider == nil || t.row < 0 || t.row >= t.provider.Length() {
		return -1, -1
	}
	if t.cellNav {
		return t.row, t.column
	}
	return t.row, -1
}

// SetSelected programmatically sets the selected row (and column in cell mode).
// In row mode the col argument is ignored. Returns false if the row is out of range.
func (t *Table) SetSelected(row, col int) bool {
	if t.provider == nil || row < 0 || row >= t.provider.Length() {
		return false
	}
	t.row = row
	if t.cellNav {
		cols := t.provider.Columns()
		t.column = max(0, min(col, len(cols)-1))
		t.adjust()
		t.adjustCol()
	} else {
		t.adjust()
	}
	return true
}

// Offset returns the current horizontal and vertical scroll offsets.
func (t *Table) Offset() (int, int) {
	return t.offsetX, t.offsetY
}

// SetOffset sets the scroll offsets, clamping to valid ranges.
func (t *Table) SetOffset(offsetX, offsetY int) {
	t.offsetX = max(0, min(offsetX, max(0, t.tableWidth-t.width)))
	t.offsetY = max(0, min(offsetY, max(0, t.provider.Length()-1)))
	t.Refresh()
}

func (t *Table) handleKey(event *tcell.EventKey) bool {
	_, _, _, h := t.Content()
	pageSize := max(1, h-3)
	columns := t.provider.Columns()
	lastCol := max(0, len(columns)-1)

	switch event.Key() {
	case tcell.KeyDown:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			t.row = max(0, t.provider.Length()-1)
			t.adjust()
		} else {
			if t.row < t.provider.Length()-1 {
				t.row++
				t.adjust()
			}
		}
		return true
	case tcell.KeyUp:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			t.row = 0
			t.adjust()
		} else {
			if t.row > 0 {
				t.row--
				t.adjust()
			}
		}
		return true
	case tcell.KeyLeft:
		if t.cellNav {
			if event.Modifiers()&tcell.ModCtrl != 0 {
				t.column = 0
				t.adjustCol()
			} else {
				if t.column > 0 {
					t.column--
				}
				t.adjustCol()
			}
		} else {
			if event.Modifiers()&tcell.ModCtrl != 0 {
				t.scrollByColumn(-1)
			} else {
				if t.offsetX > 0 {
					t.offsetX--
					t.Refresh()
				}
			}
		}
		return true
	case tcell.KeyRight:
		if t.cellNav {
			if event.Modifiers()&tcell.ModCtrl != 0 {
				t.column = lastCol
				t.adjustCol()
			} else {
				if t.column < lastCol {
					t.column++
				}
				t.adjustCol()
			}
		} else {
			if event.Modifiers()&tcell.ModCtrl != 0 {
				t.scrollByColumn(1)
			} else {
				if t.offsetX+t.width < t.tableWidth {
					t.offsetX++
					t.Refresh()
				}
			}
		}
		return true
	case tcell.KeyHome:
		if t.cellNav {
			if event.Modifiers()&tcell.ModCtrl != 0 {
				t.row = 0
				t.column = 0
				t.offsetX = 0
				t.adjust()
			} else {
				t.column = 0
				t.adjustCol()
			}
		} else {
			if event.Modifiers()&tcell.ModCtrl != 0 {
				t.row = 0
				t.offsetX = 0
				t.adjust()
			} else {
				t.row = 0
				t.adjust()
			}
		}
		return true
	case tcell.KeyEnd:
		if t.cellNav {
			if event.Modifiers()&tcell.ModCtrl != 0 {
				t.row = max(0, t.provider.Length()-1)
				t.column = lastCol
				t.offsetX = 0
				t.adjust()
			} else {
				t.column = lastCol
				t.adjustCol()
			}
		} else {
			if event.Modifiers()&tcell.ModCtrl != 0 {
				t.row = max(0, t.provider.Length()-1)
				t.offsetX = 0
				t.adjust()
			} else {
				t.row = max(0, t.provider.Length()-1)
				t.adjust()
			}
		}
		return true
	case tcell.KeyPgUp:
		t.row = max(0, t.row-pageSize)
		t.adjust()
		return true
	case tcell.KeyPgDn:
		t.row = min(t.provider.Length()-1, t.row+pageSize)
		t.adjust()
		return true
	case tcell.KeyEnter:
		if t.provider.Length() > 0 && t.row >= 0 && t.row < t.provider.Length() {
			rowData := t.getCurrentRowData()
			t.Dispatch(t, EvtActivate, t.row, rowData)
		}
		return true
	case tcell.KeyRune:
		if event.Str() == " " {
			if t.provider.Length() > 0 && t.row >= 0 && t.row < t.provider.Length() {
				col := -1
				if t.cellNav {
					col = t.column
				}
				t.Dispatch(t, EvtSelect, t.row, col)
			}
			return true
		}
		return false
	default:
		return false
	}
}

// adjust ensures the selected row is visible, scrolling vertically as needed.
func (t *Table) adjust() {
	_, _, _, h := t.Content()

	if t.provider.Length() < h-2 {
		t.Refresh()
		return
	}
	if t.row < t.offsetY {
		t.offsetY = t.row
	}
	if t.row > t.offsetY+h-3 {
		t.offsetY = t.row - h + 3
	}
	t.Refresh()
}

// adjustCol ensures the selected column is visible, scrolling horizontally as needed.
func (t *Table) adjustCol() {
	columns := t.provider.Columns()
	if t.column < 0 || t.column >= len(columns) {
		t.Refresh()
		return
	}
	colX := 0
	for i := 0; i < t.column; i++ {
		colX += columns[i].Width + 1
	}
	colW := columns[t.column].Width
	_, _, w, _ := t.Content()
	if colX < t.offsetX {
		t.offsetX = colX
	} else if colX+colW > t.offsetX+w {
		t.offsetX = colX + colW - w
	}
	t.Refresh()
}

// scrollByColumn scrolls horizontally by one full column width.
func (t *Table) scrollByColumn(direction int) {
	columns := t.provider.Columns()
	if len(columns) == 0 {
		return
	}

	if direction < 0 {
		if t.offsetX > 0 {
			// Walk column starts; keep the last one strictly before offsetX.
			target := 0
			currentPos := 0
			for _, column := range columns {
				if currentPos >= t.offsetX {
					break
				}
				target = currentPos
				currentPos += column.Width + 1
			}
			t.offsetX = target
			t.Refresh()
		}
	} else {
		if t.offsetX+t.width < t.tableWidth {
			currentPos := 0
			for _, column := range columns {
				if currentPos > t.offsetX {
					t.offsetX = min(t.tableWidth-t.width, currentPos)
					break
				}
				currentPos += column.Width + 1
			}
			t.Refresh()
		}
	}
}

// getCurrentRowData returns all column values for the currently selected row.
func (t *Table) getCurrentRowData() []string {
	if t.provider == nil || t.row < 0 || t.row >= t.provider.Length() {
		return nil
	}

	columns := t.provider.Columns()
	rowData := make([]string, len(columns))
	for i := range columns {
		rowData[i] = t.provider.Str(t.row, i)
	}
	return rowData
}

// SetCellStyler sets a per-cell style callback. fn receives the row, column, and
// whether the cell is the focused/highlighted cell. Return nil to use the default
// style. Set fn to nil to clear.
func (t *Table) SetCellStyler(fn func(row, col int, highlight bool) *Style) {
	t.cellStyler = fn
}

// CellBounds returns the screen-space position and width of cell (row, col).
// ok is false when the cell is outside the visible viewport.
func (t *Table) CellBounds(row, col int) (x, y, w int, ok bool) {
	columns := t.provider.Columns()
	if row < 0 || row >= t.provider.Length() || col < 0 || col >= len(columns) {
		return 0, 0, 0, false
	}
	cx, cy, cw, ch := t.Content()
	visRow := row - t.offsetY
	if visRow < 0 || visRow >= ch-2 {
		return 0, 0, 0, false
	}
	colX := 0
	for i := 0; i < col; i++ {
		colX += columns[i].Width + 1
	}
	colW := columns[col].Width
	screenX := cx - t.offsetX + colX
	if screenX+colW <= cx || screenX >= cx+cw {
		return 0, 0, 0, false
	}
	visX := max(screenX, cx)
	visW := min(screenX+colW, cx+cw) - visX
	if visW <= 0 {
		return 0, 0, 0, false
	}
	return visX, cy + 2 + visRow, visW, true
}

// cellText returns text padded to width runes according to alignment.
func cellText(text string, width, alignment int) string {
	runes := []rune(text)
	if len(runes) >= width {
		return string(runes[:width])
	}
	pad := width - len(runes)
	switch alignment {
	case AlignRight:
		return strings.Repeat(" ", pad) + text
	case AlignCenter:
		left := pad / 2
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", pad-left)
	default:
		return text
	}
}

func (t *Table) Render(r *Renderer) {
	t.Component.Render(r)

	x, y, w, h := t.Content()
	state := t.State()
	if state != "" {
		state = ":" + state
	}
	headerStyle := t.Style("header" + state)
	gridStyle := t.Style("grid" + state)
	border := gridStyle.Border()
	if border != "" && border != "none" {
		t.grid = r.theme.Border(border)
	}
	t.renderTableHeader(r, x, y, w, h, headerStyle, gridStyle)
	t.renderTableContent(r, x, y+2, w, h-2, gridStyle)
}

func (t *Table) renderTableHeader(r *Renderer, x, y, w, h int, headerStyle, gridStyle *Style) {
	rx := 0
	rw := w
	columns := t.provider.Columns()
	for i, column := range columns {
		if rw <= 0 {
			break
		}
		if rx+column.Width < t.offsetX {
			rx = rx + column.Width + 1
			continue
		}
		r.Set(headerStyle.Foreground(), headerStyle.Background(), headerStyle.Font())

		cx := x - t.offsetX + rx
		if rx < t.offsetX {
			start := t.offsetX - rx
			rcw := min(column.Width-start, rw)
			runes := []rune(column.Header)
			if start < len(runes) {
				r.Text(cx+start, y, string(runes[start:]), rcw)
			}
			r.Set(gridStyle.Foreground(), gridStyle.Background(), gridStyle.Font())
			r.Repeat(cx+start, y+1, 1, 0, rcw, t.grid.InnerH)
			rw += start
		} else {
			cw := min(rw, column.Width)
			r.Text(cx, y, column.Header, cw)
			r.Set(gridStyle.Foreground(), gridStyle.Background(), gridStyle.Font())
			r.Repeat(cx, y+1, 1, 0, cw, t.grid.InnerH)
		}
		if i < len(columns)-1 && rw > column.Width {
			r.screen.Put(cx+column.Width, y+1, t.grid.InnerTopT)
			if t.outer {
				r.screen.Put(cx+column.Width, y+h, t.grid.BottomT)
			}
		}
		rx = rx + column.Width + 1
		rw = rw - column.Width - 1
	}

	r.Set(gridStyle.Foreground(), gridStyle.Background(), gridStyle.Font())
	if rx-t.offsetX < w {
		r.Repeat(x-t.offsetX+rx-1, y+1, 1, 0, w+t.offsetX-rx+1, t.grid.InnerH)
	}
	if t.outer {
		r.screen.Put(x-1, y+1, t.grid.LeftT)
		r.screen.Put(x+w, y+1, t.grid.RightT)
	}
}

func (t *Table) renderTableContent(r *Renderer, x, y, w, h int, gridStyle *Style) {
	row := t.offsetY
	columns := t.provider.Columns()
	focused := t.Flag(FlagFocused)

	for row < t.provider.Length() && row-t.offsetY < h {
		rx := 0
		rw := w
		for i, column := range columns {
			if rw <= 0 {
				break
			}
			if rx+column.Width < t.offsetX {
				rx = rx + column.Width + 1
				continue
			}

			// Per-cell style: check styler first, then defaults.
			highlight := row == t.row && (!t.cellNav || i == t.column)
			var style *Style
			if t.cellStyler != nil {
				style = t.cellStyler(row, i, highlight)
			}
			if style == nil {
				if t.cellNav && row == t.row && i == t.column {
					if focused {
						style = t.Style("cell:focused")
					} else {
						style = t.Style("cell")
					}
				} else if row == t.row {
					if focused {
						style = t.Style("highlight:focused")
					} else {
						style = t.Style("highlight")
					}
				} else {
					style = t.Style()
				}
			}

			r.Set(style.Foreground(), style.Background(), style.Font())
			cx := x - t.offsetX + rx
			cy := y - t.offsetY + row
			if rx < t.offsetX {
				start := t.offsetX - rx
				rcw := min(column.Width-start, rw)
				runes := []rune(cellText(t.provider.Str(row, i), column.Width, column.Alignment))
				if start < len(runes) {
					r.Text(cx+start, cy, string(runes[start:]), rcw)
				} else {
					r.Repeat(cx+start, cy, 1, 0, rcw, " ")
				}
				rw += start
			} else {
				cw := min(rw, column.Width)
				r.Text(cx, cy, cellText(t.provider.Str(row, i), column.Width, column.Alignment), cw)
			}
			if i < len(columns)-1 && t.inner && rw > column.Width {
				// Always render separator in grid style, never in cell/highlight colour
				r.Set(gridStyle.Foreground(), gridStyle.Background(), gridStyle.Font())
				r.screen.Put(cx+column.Width, cy, t.grid.InnerV)
			}
			rx = rx + column.Width + 1
			rw = rw - column.Width - 1
		}
		row++
	}

	// Fill remaining rows with empty grid lines
	for row-t.offsetY < h {
		rx := 0
		rw := w
		cy := y - t.offsetY + row
		for i, column := range columns {
			if rw <= 0 {
				break
			}
			if rx+column.Width < t.offsetX {
				rx = rx + column.Width + 1
				continue
			}
			cx := x - t.offsetX + rx
			r.Set(gridStyle.Foreground(), gridStyle.Background(), gridStyle.Font())
			if rx < t.offsetX {
				start := t.offsetX - rx
				cw := min(column.Width-start, rw)
				if cw > 0 {
					r.Repeat(cx+start, cy, 1, 0, cw, " ")
				}
				rw += start
			} else {
				cw := min(rw, column.Width)
				r.Repeat(cx, cy, 1, 0, cw, " ")
			}
			if i < len(columns)-1 && t.inner && rw > column.Width {
				r.screen.Put(cx+column.Width, cy, t.grid.InnerV)
			}
			rx = rx + column.Width + 1
			rw = rw - column.Width - 1
		}
		row++
	}
}

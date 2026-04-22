package widgets

import (
	"fmt"

	. "github.com/tekugo/zeichenwerk/core"
)

// ==== AI ===================================================================

// Heatmap renders a rows×cols grid of float64 values as a colour-graded matrix,
// similar to a GitHub contribution graph. Cell colour interpolates from the
// "heatmap/zero" background (no activity) to the "heatmap/max" background
// (peak activity) using lerpColor.
type Heatmap struct {
	Component
	rows      int
	cols      int
	values    [][]float64 // [row][col], all non-negative
	rowLabels []string    // optional; len == rows or nil
	colLabels []string    // optional; len == cols or nil
	cellWidth int         // character width per cell (default 1)
}

// NewHeatmap creates a Heatmap with a zeroed rows×cols value grid and cellWidth 1.
func NewHeatmap(id, class string, rows, cols int) *Heatmap {
	values := make([][]float64, rows)
	for i := range values {
		values[i] = make([]float64, cols)
	}
	return &Heatmap{
		Component: Component{id: id, class: class},
		rows:      rows,
		cols:      cols,
		values:    values,
		cellWidth: 1,
	}
}

// ---- Widget interface -------------------------------------------------------

// Apply registers all heatmap style keys from the active theme.
func (h *Heatmap) Apply(theme *Theme) {
	theme.Apply(h, h.Selector("heatmap"))
	theme.Apply(h, h.Selector("heatmap/header"))
	theme.Apply(h, h.Selector("heatmap/zero"))
	theme.Apply(h, h.Selector("heatmap/mid"))
	theme.Apply(h, h.Selector("heatmap/max"))
}

// Hint returns the preferred size. Width accounts for label column, cell
// widths, and single-character gaps between cells. Height accounts for the
// optional column-label row plus one row per data row.
func (h *Heatmap) Hint() (int, int) {
	if h.hwidth != 0 || h.hheight != 0 {
		return h.hwidth, h.hheight
	}
	lcw := h.labelColWidth()
	lrh := 0
	if len(h.colLabels) > 0 {
		lrh = 1
	}
	w := lcw
	if h.cols > 0 {
		w += h.cols*(h.cellWidth+1) - 1
	}
	return w, lrh + h.rows
}

// Render draws the heatmap. Each cell is rendered as cellWidth spaces whose
// background colour is linearly interpolated between the "heatmap/zero" and
// "heatmap/max" background colours based on the cell's fraction of the maximum
// value in the grid.
func (h *Heatmap) Render(r *Renderer) {
	h.Component.Render(r)

	cx, cy, cw, ch := h.Content()
	if cw <= 0 || ch <= 0 {
		return
	}

	maxVal := h.maxVal()
	lcw := h.labelColWidth()
	lrh := 0
	if len(h.colLabels) > 0 {
		lrh = 1
	}

	zeroStyle := h.Style("zero")
	maxStyle := h.Style("max")
	headerStyle := h.Style("header")
	midFont := h.Style("mid").Font()

	theme := r.Theme
	zeroFg := theme.Color(zeroStyle.Foreground())
	zeroBg := theme.Color(zeroStyle.Background())
	maxFg := theme.Color(maxStyle.Foreground())
	maxBg := theme.Color(maxStyle.Background())

	if lrh > 0 {
		h.renderColHeaders(r, cx, cy, cw, lcw, headerStyle)
	}

	for row := 0; row < h.rows; row++ {
		rowY := cy + lrh + row
		if rowY >= cy+ch {
			break
		}
		h.renderRow(r, row, rowY, cx, cw, lcw, headerStyle, midFont,
			zeroFg, zeroBg, maxFg, maxBg, maxVal)
	}
}

// renderColHeaders draws column labels centred over their cell columns.
func (h *Heatmap) renderColHeaders(r *Renderer, cx, cy, cw, lcw int, headerStyle *Style) {
	r.Set(headerStyle.Foreground(), headerStyle.Background(), headerStyle.Font())
	for c := 0; c < h.cols; c++ {
		cellX := cx + lcw + c*(h.cellWidth+1)
		if cellX >= cx+cw {
			break
		}
		label := ""
		if c < len(h.colLabels) {
			label = h.colLabels[c]
		}
		writeLabel(r, cellX, cy, h.cellWidth, cw-(cellX-cx), label)
	}
}

// renderRow draws one data row: its optional label and all cells.
func (h *Heatmap) renderRow(r *Renderer, row, rowY, cx, cw, lcw int,
	headerStyle *Style, midFont string,
	zeroFg, zeroBg, maxFg, maxBg string, maxVal float64,
) {
	if lcw > 0 {
		r.Set(headerStyle.Foreground(), headerStyle.Background(), headerStyle.Font())
		label := ""
		if row < len(h.rowLabels) {
			label = h.rowLabels[row]
		}
		// Left-aligned in (lcw-1) chars, then a trailing space separator.
		writeLabel(r, cx, rowY, lcw-1, cw, label)
		if cx+lcw-1 < cx+cw {
			r.Put(cx+lcw-1, rowY, " ")
		}
	}

	for col := 0; col < h.cols; col++ {
		cellX := cx + lcw + col*(h.cellWidth+1)
		if cellX >= cx+cw {
			break
		}
		v := h.values[row][col]
		fg, bg := h.cellColors(v, maxVal, zeroFg, zeroBg, maxFg, maxBg)
		r.Set(fg, bg, midFont)
		limit := h.cellWidth
		if cellX+limit > cx+cw {
			limit = cx + cw - cellX
		}
		for i := 0; i < limit; i++ {
			r.Put(cellX+i, rowY, " ")
		}
	}
}

// cellColors returns the (fg, bg) pair for a cell value.
func (h *Heatmap) cellColors(v, maxVal float64,
	zeroFg, zeroBg, maxFg, maxBg string,
) (string, string) {
	if maxVal == 0 || v == 0 {
		return zeroFg, zeroBg
	}
	frac := v / maxVal
	if frac >= 1 {
		return maxFg, maxBg
	}
	return LerpColor(zeroFg, maxFg, frac), LerpColor(zeroBg, maxBg, frac)
}

// writeLabel writes up to width runes of label starting at (x, y), left-aligned,
// space-padded to width. maxW is the remaining clipping boundary.
func writeLabel(r *Renderer, x, y, width, maxW int, label string) {
	runes := []rune(label)
	for i := 0; i < width; i++ {
		if i >= maxW {
			break
		}
		ch := " "
		if i < len(runes) {
			ch = string(runes[i])
		}
		r.Put(x+i, y, ch)
	}
}

// ---- Data methods ----------------------------------------------------------

// SetValue sets a single cell value (non-negative) and calls Refresh.
func (h *Heatmap) SetValue(row, col int, v float64) {
	if row < 0 || row >= h.rows || col < 0 || col >= h.cols {
		return
	}
	h.values[row][col] = v
	h.Refresh()
}

// SetRow replaces an entire row and calls Refresh.
func (h *Heatmap) SetRow(row int, vs []float64) {
	if row < 0 || row >= h.rows {
		return
	}
	n := len(vs)
	if n > h.cols {
		n = h.cols
	}
	copy(h.values[row], vs[:n])
	h.Refresh()
}

// SetAll replaces the entire grid and calls Refresh. Panics if dimensions do
// not match the values the heatmap was constructed with.
func (h *Heatmap) SetAll(vs [][]float64) {
	if len(vs) != h.rows {
		panic(fmt.Sprintf("heatmap: SetAll: row count %d != %d", len(vs), h.rows))
	}
	for i, row := range vs {
		if len(row) != h.cols {
			panic(fmt.Sprintf("heatmap: SetAll: col count %d != %d at row %d", len(row), h.cols, i))
		}
		copy(h.values[i], row)
	}
	h.Refresh()
}

// Value returns the value of one cell.
func (h *Heatmap) Value(row, col int) float64 {
	if row < 0 || row >= h.rows || col < 0 || col >= h.cols {
		return 0
	}
	return h.values[row][col]
}

// SetRowLabels sets the row header strings. Pass nil to clear them.
func (h *Heatmap) SetRowLabels(labels []string) {
	h.rowLabels = labels
	h.Refresh()
}

// SetColLabels sets the column header strings. Pass nil to clear them.
func (h *Heatmap) SetColLabels(labels []string) {
	h.colLabels = labels
	h.Refresh()
}

// SetCellWidth sets the character width per cell (minimum 1) and calls Refresh.
func (h *Heatmap) SetCellWidth(w int) {
	if w < 1 {
		w = 1
	}
	h.cellWidth = w
	h.Refresh()
}

// ---- internal helpers -------------------------------------------------------

func (h *Heatmap) maxVal() float64 {
	var m float64
	for _, row := range h.values {
		for _, v := range row {
			if v > m {
				m = v
			}
		}
	}
	return m
}

func (h *Heatmap) labelColWidth() int {
	if len(h.rowLabels) == 0 {
		return 0
	}
	max := 0
	for _, l := range h.rowLabels {
		if n := len([]rune(l)); n > max {
			max = n
		}
	}
	return max + 1 // +1 for trailing space separator
}

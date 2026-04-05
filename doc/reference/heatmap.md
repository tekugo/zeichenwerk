# Heatmap

Colour-graded activity grid similar to a GitHub contribution graph. Each cell
represents one bucket in a two-dimensional key space. Cell background colour
interpolates from the `"heatmap/zero"` style (no activity) to the
`"heatmap/max"` style (peak activity) based on the cell's fraction of the
maximum value in the grid.

**Constructor:** `NewHeatmap(id, class string, rows, cols int) *Heatmap`

## Methods

- `SetValue(row, col int, v float64)` — set one cell; calls `Refresh`
- `SetRow(row int, vs []float64)` — replace an entire row; calls `Refresh`
- `SetAll(vs [][]float64)` — replace the entire grid; panics on dimension mismatch; calls `Refresh`
- `Value(row, col int) float64` — return the value of one cell
- `SetRowLabels(labels []string)` — set row header strings (nil clears them)
- `SetColLabels(labels []string)` — set column header strings (nil clears them)
- `SetCellWidth(w int)` — character width per cell (minimum 1, default 1); width 2 gives square-ish cells on most terminals

## Rendering

Each cell is rendered as `cellWidth` spaces. Background colour is linearly
interpolated between `"heatmap/zero"` background and `"heatmap/max"` background
using `lerpColor`; the foreground follows the same interpolation.

Cells with value 0 always use the `"heatmap/zero"` style. The cell(s) equal to
the maximum value use the `"heatmap/max"` style. If all values are 0, every cell
uses `"heatmap/zero"`.

Adjacent cells are separated by a single blank character drawn with the base
`"heatmap"` background.

## Hint

Width: `labelColW + cols×(cellWidth+1) − 1`  
Height: `labelRowH + rows`

where `labelColW` = widest row label + 1 (or 0 if no labels), and
`labelRowH` = 1 if column labels are set, else 0.

## Style keys

| Selector | Applied to |
|----------|-----------|
| `"heatmap"` | Background and border |
| `"heatmap/header"` | Row and column label text |
| `"heatmap/zero"` | Cells with value 0 |
| `"heatmap/mid"` | Font for intermediate-value cells |
| `"heatmap/max"` | Cells at the maximum value |

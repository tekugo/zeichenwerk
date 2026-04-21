# Heatmap

A grid widget that visualises activity density as a colour-graded matrix of
cells — similar to a GitHub contribution graph. Each cell represents one bucket
in a two-dimensional key space (e.g. hour-of-day × weekday). Cell colour
transitions from the theme's dim background towards a saturated accent colour
as the cell's value increases relative to the maximum in the dataset.

## Visual layout

```
     Mon  Tue  Wed  Thu  Fri  Sat  Sun
 0h   ░    ░    ░    ░    ░    ░    ░
 6h   ▒    ▒    ▒    ░    ░    ░    ░
12h   ▓    ▓    ▒    ▒    ░    ░    ░
18h   █    ▓    ▓    ▒    ▒    ░    ░
```

Row and column headers are optional. Each cell is rendered as a single
character. An empty (zero-value) cell uses the `"heatmap/zero"` style; the
maximum-value cell uses `"heatmap/max"`. Intermediate values are rendered by
blending between those two styles using the theme's colour interpolation.

## Structure

```go
type Heatmap struct {
    Component
    rows      int
    cols      int
    values    [][]float64   // [row][col], all non-negative
    rowLabels []string      // optional, len == rows or nil
    colLabels []string      // optional, len == cols or nil
    cellWidth int           // character width of each cell (default 1)
}
```

`values` is always `rows × cols`. `SetValue` and `SetAll` keep it
in-bounds. The zero value of a `float64` cell is considered "no activity".

## Constructor

```go
func NewHeatmap(id, class string, rows, cols int) *Heatmap
```

- Allocates a zeroed `rows × cols` value grid.
- `cellWidth = 1`.
- Not focusable (display-only widget).

## Methods

| Method | Description |
|--------|-------------|
| `SetValue(row, col int, v float64)` | Sets a single cell value (non-negative); calls `Refresh()` |
| `SetRow(row int, vs []float64)` | Replaces an entire row; calls `Refresh()` |
| `SetAll(vs [][]float64)` | Replaces the entire grid; panics if dimensions mismatch; calls `Refresh()` |
| `Value(row, col int) float64` | Returns the value of one cell |
| `SetRowLabels(labels []string)` | Sets the row header strings (nil clears them) |
| `SetColLabels(labels []string)` | Sets the column header strings (nil clears them) |
| `SetCellWidth(w int)` | Sets the character width per cell (minimum 1) |

## Rendering

```go
func (h *Heatmap) Render(r *Renderer)
```

1. `h.Component.Render(r)` — draws background and border.
2. Compute content area `(cx, cy, cw, ch)`.
3. Determine `maxVal = max of all values`. If `maxVal == 0`, every cell
   renders with `"heatmap/zero"`.
4. Determine layout offsets:
   - `labelColW`: width of the widest row label + 1 space, or 0 if no labels.
   - `labelRowH`: 1 if column labels are present, else 0.
5. Render column headers at `cy` if present: each label centred over its cell
   column, using `"heatmap/header"`.
6. For each row `r` in `0 … rows-1`:
   - Render the row label at `(cx, cy + labelRowH + r)` left-aligned in
     `labelColW` characters if labels are set, using `"heatmap/header"`.
   - For each column `c` in `0 … cols-1`:
     - `fraction = values[r][c] / maxVal` (0 when `maxVal == 0`).
     - Select the cell character and style:
       - `fraction == 0` → `"heatmap/zero"` style.
       - `fraction == 1` → `"heatmap/max"` style.
       - Otherwise → `"heatmap/mid"` style; the theme is responsible for
         rendering intermediate intensity (e.g. via background colour
         interpolation keyed on a `"heatmap.intensity"` theme attribute).
     - Write `cellWidth` characters at the cell's column position.

The widget does not perform colour interpolation itself — it exposes the
intensity fraction through the `"heatmap.intensity"` theme variable set before
each intermediate cell is drawn, and the theme definition maps that to a colour.
This keeps colour policy in the theme layer.

### Intensity variable

Before rendering each cell, the renderer's theme variable `"heatmap.intensity"`
is set to the cell's fraction as a string formatted to two decimal places
(e.g. `"0.72"`). Theme authors use this to drive colour blends.

## Hint

```go
func (h *Heatmap) Hint() (int, int)
```

- Width: `labelColW + cols * (cellWidth + 1) - 1` (gaps between cells but
  not after the last one), plus `labelColW`.
- Height: `labelRowH + rows`.

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"heatmap"` | Background and border |
| `"heatmap/header"` | Row and column label text |
| `"heatmap/zero"` | Cells with value 0 |
| `"heatmap/mid"` | Cells with value between 0 and max (exclusive) |
| `"heatmap/max"` | Cells with the maximum value |

## Theme variable keys

| Key | Format | Description |
|-----|--------|-------------|
| `heatmap.intensity` | `"0.00"` – `"1.00"` | Fraction used by theme to select intermediate colour; set per cell before drawing `"heatmap/mid"` cells |

## Implementation plan

1. **`heatmap.go`** — new file
   - Define `Heatmap` struct and `NewHeatmap`.
   - Implement `SetValue`, `SetRow`, `SetAll`, `Value`, `SetRowLabels`,
     `SetColLabels`, `SetCellWidth`.
   - Implement `maxVal` helper, `Apply`, `Hint`, `Render`.

2. **`builder.go`** — add `Heatmap` method
   ```go
   func (b *Builder) Heatmap(id string, rows, cols int) *Builder
   ```

3. **Theme** — add `"heatmap"`, `"heatmap/header"`, `"heatmap/zero"`,
   `"heatmap/mid"`, and `"heatmap/max"` entries to built-in themes, with
   example intensity-based colour stepping.

4. **Tests** — `heatmap_test.go`
   - All-zero grid renders every cell with `"heatmap/zero"`.
   - Single non-zero cell is both zero-relative max and renders with
     `"heatmap/max"`.
   - `SetAll` with mismatched dimensions panics.
   - Row and column labels are placed at correct offsets.
   - Hint dimensions are correct with and without labels.
   - `heatmap.intensity` is set to `"1.00"` for the max-value cell and
     `"0.00"` for a zero cell.

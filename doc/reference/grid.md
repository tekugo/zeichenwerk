# Grid

Table-based layout with cell spanning.

**Constructor:** `NewGrid(id, class string, rows, columns int, lines bool) *Grid`

`lines = true` draws thin separators between cells. After construction, all rows and columns are initialised to fractional weight `-1` (one fraction each); call `Rows(...)` and `Columns(...)` to override.

## Methods

- `Add(content Widget, params ...any) error` — place a widget in a cell. `params` is `x, y, w, h` (column, row, column-span, row-span). All ints. Without params, defaults to `(0, 0, 1, 1)`.
- `Children() []Widget` — all placed widgets
- `Columns(columns ...int)` — column sizes (one int per column)
- `Rows(rows ...int)` — row sizes (one int per row)
- `Lines() bool` — whether grid lines are drawn
- `GridCells() []GridCell` — internal cell descriptors (for inspection / custom rendering)
- `Layout() error` — compute cell positions and apply bounds to children

## Notes

**Track sizing** (same rule as `Hint`):

- `>0` — fixed cells
- `<0` — fractional weight (magnitude is the share)
- `0`  — auto (use the widget's preferred size)

For the grid to absorb all available space, at least one row and one column must use a fractional size.

The Builder method `Cell(x, y, w, h)` sets the position for the **next** widget added; you don't pass cell coordinates to the widget method itself when building through the fluent API.

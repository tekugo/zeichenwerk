# Grid

Table-based layout container with cell spanning.

**Constructor:** `NewGrid(id, class string, rows, columns int, lines bool) *Grid`

## Methods

- `Add(x, y, w, h int, widget Widget)` — places widget in cell with span
- `Children() []Widget` — returns all cell contents
- `Columns(columns ...int)` — sets column sizes
- `Layout()` — calculates cell positions and sizes
- `Rows(rows ...int)` — sets row sizes

## Notes

**Sizing:** positive = fixed chars, negative = fractional unit, zero = auto (preferred size)

**Separator constants:** `GridH` (horizontal), `GridV` (vertical), `GridB` (both)

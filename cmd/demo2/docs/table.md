# Table

Tabular data display with header, vertical/horizontal scrolling, and optional cell navigation. Implements `values.Setter[TableProvider]` so it works with `values.Update`.

**Constructor:** `NewTable(id, class string, provider TableProvider, cellNav bool) *Table`

`cellNav` enables column-by-column navigation in addition to row navigation.

## Methods

- `Set(provider TableProvider)` — replace the data source. **Does not redraw**; call `Refresh()` afterwards (or `core.Find(ui, id).Refresh()`)
- `Refresh()` — queue a redraw
- `Selected() (row, col int)` — currently highlighted row and column (column is meaningful only when `cellNav` is true)
- `SetSelected(row, col int) bool` — set the highlighted cell; returns true if the position is valid
- `Offset() (offsetX, offsetY int)` — current scroll offset
- `SetOffset(offsetX, offsetY int)` — set scroll position
- `SetCellStyler(fn func(row, col int, highlight bool) *Style)` — per-cell style override
- `CellBounds(row, col int) (x, y, w int, ok bool)` — screen coordinates of a cell

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"activate"` | `int, []string` | Row activated via Enter; second arg is the full row data |
| `"select"` | `int, int` | Selection changed (Space): row index and column index (`-1` when not in cell-nav mode) |

## Notes

Flags: `"focusable"`, `"grid"` (toggle inner grid lines).

`TableProvider`:

```go
type TableProvider interface {
    Columns() []TableColumn  // Header text + column width
    Length() int              // Row count
    Str(row, col int) string  // Cell content
}
```

Built-in: `NewArrayTableProvider(headers []string, data [][]string) *ArrayTableProvider`. Column widths are computed from the longest cell. For dynamic data sources (live cursors, paged APIs), implement `TableProvider` directly.

> **Important:** unlike `List.Set`, `Table.Set` updates the provider but does **not** auto-refresh. Always follow `Set` (or `values.Update`) with a `Refresh()` call on the table.

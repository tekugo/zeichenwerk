# Table

Tabular data display with scrolling.

**Constructor:** `NewTable(id, class string, provider TableProvider) *Table`

## Methods

- `GetScrollOffset() (int, int)` — returns horizontal and vertical scroll offsets
- `GetSelectedRow() int` — returns highlighted row index
- `Set(provider TableProvider)` — replaces data source
- `SetScrollOffset(offsetX, offsetY int)` — sets scroll position
- `SetSelectedRow(row int) bool` — selects row by index

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"activate"` | `int, []string` | Row activated via Enter |
| `"select"` | `int, []string` | Row selected via Space |

## Notes

Flags: `"focusable"`

**TableProvider interface:**
```go
type TableProvider interface {
    Columns() []TableColumn
    Length() int
    Str(row, col int) string
}
```

**Built-in:** `NewArrayTableProvider(headers []string, data [][]string)`

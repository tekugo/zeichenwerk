# List

Scrollable selectable list of strings. Implements `values.Setter[[]string]` and `Filterable`, so it works with `values.Update` and `Filter` widgets.

**Constructor:** `NewList(id, class string, items []string) *List`

## Methods

- `Items() []string` — current items
- `Set(value []string)` — replace all items, reset selection to 0, redraw
- `Selected() int` — highlighted index
- `Select(index int)` — set highlighted index
- `Move(count int)` — move highlight by `count` (skips disabled items)
- `First()` — jump to first enabled item
- `Last()` — jump to last enabled item
- `PageUp()` / `PageDown()` — move by viewport height
- `Filter(filter string)` — case-insensitive substring filter (empty string clears the filter)
- `Suggest(query string) []string` — items beginning with `query` (used by Typeahead/Combo for ghost-text)
- `Refresh()` — queue a redraw

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"select"` | `int` | Highlighted item changed (arrow keys, click) |
| `"activate"` | `int` | Item activated via Enter or double-click |

## Notes

Flags: `"focusable"`. Optional `"search"` flag enables incremental search-as-you-type.

Keyboard: ↑/↓ move; PgUp/PgDn page; Home/End jump to first/last; Enter activates the highlighted item.

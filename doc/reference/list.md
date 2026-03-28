# List

Scrollable selectable list.

**Constructor:** `NewList(id, class string, items []string) *List`

## Methods

- `First()` — jumps to first enabled item
- `Items() []string` — returns all items
- `Last()` — jumps to last enabled item
- `Move(count int)` — moves highlight by count (skips disabled items)
- `PageDown()` — moves down by viewport height
- `PageUp()` — moves up by viewport height
- `Select(index int)` — sets highlighted item by index
- `Selected() int` — returns highlighted index
- `SetItems(items []string)` — replaces all items

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"activate"` | `int` | Item activated via Enter |
| `"select"` | `int` | Highlighted item changed |

## Notes

Flags: `"focusable"`

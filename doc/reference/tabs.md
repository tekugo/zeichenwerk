# Tabs

Tab navigation widget.

**Constructor:** `NewTabs(id, class string) *Tabs`

## Methods

- `Add(title string)` — appends a new tab
- `Count() int` — returns number of tabs
- `Select(index int) bool` — sets active tab
- `Selected() int` — returns active tab index

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"activate"` | `int` | Tab selected via Enter |
| `"change"` | `int` | Highlighted tab changed |

## Notes

Flags: `"focusable"`

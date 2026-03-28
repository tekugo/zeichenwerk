# Switcher

Content switcher showing one child pane at a time.

**Constructor:** `NewSwitcher(id, class string) *Switcher`

## Methods

- `Add(widget Widget)` — appends a content pane
- `Children() []Widget` — returns all panes
- `Select(index any)` — shows pane by `int` index or `string` ID
- `Selected() int` — returns current pane index

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"hide"` | — | Pane became hidden |
| `"show"` | — | Pane became visible |

## Notes

Can auto-connect to a Tabs widget via the Builder's `connect` parameter.

# Tabs

Tab strip — a horizontal row of clickable labels. Pair with a `Switcher` to swap content panes.

**Constructor:** `NewTabs(id, class string) *Tabs`

## Methods

- `Add(title string)` — append a new tab
- `Count() int` — number of tabs
- `Get() int` — active tab index
- `Set(index int) bool` — set active tab; returns true if valid

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `int` | Highlighted tab changed (←/→ navigation) |
| `"activate"` | `int` | Tab activated via Enter, click, or letter shortcut |

## Notes

Flags: `"focusable"`.

Keyboard: ←/→ move highlight (fires `EvtChange`); Enter activates the highlighted tab (fires `EvtActivate`); typing a letter jumps to the first tab whose label starts with that letter and activates it.

Mouse: single click highlights and activates in one step.

# Collapsible

Single-child container with a clickable header that expands and collapses the body. When collapsed only the header row is visible; the parent layout reclaims the freed space automatically.

**Constructor:** `NewCollapsible(id, class, title string, expanded bool) *Collapsible`

## Methods

- `Add(widget Widget)` — sets the body widget; replaces any previous child
- `Children() []Widget` — returns the body widget (empty slice when none)
- `Collapse()` — hides the body; moves focus to the collapsible if the body had it
- `Expand()` — shows the body
- `Expanded() bool` — reports whether the body is currently visible
- `Hint() (int, int)` — collapsed: `(childW, 1)`; expanded: `(childW, 1+childH)`; expanded with no child height hint: `(childW, -1)` (fractional)
- `Layout()` — positions the child below the header row
- `Toggle()` — switches between expanded and collapsed

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `bool` | Expansion state changed; `true` = expanded |

## Notes

Flags: `"focusable"`

Keyboard: `Enter` / `Space` toggle; `→` expands; `←` collapses.

Mouse: clicking the header row toggles.

Style selectors: `"collapsible"`, `"collapsible/header"` — with `:focused` and `:hovered` states.

Theme strings: `"collapsible.expanded"` and `"collapsible.collapsed"` set the indicator character (defaults `▼ ` / `▶ `).

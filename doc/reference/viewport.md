# Viewport

Scrollable container for oversized content.

**Constructor:** `NewViewport(id, class, title string) *Viewport`

## Methods

- `Add(widget Widget)` — sets single scrollable child
- `Children() []Widget` — returns child widget
- `Layout()` — positions child with current scroll offsets

## Notes

Scroll via arrow keys, `PgUp`, `PgDn`, `Home`, `End`.

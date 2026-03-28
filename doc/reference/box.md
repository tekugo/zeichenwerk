# Box

Bordered box container with optional title, holding a single child.

**Constructor:** `NewBox(id, class, title string) *Box`

## Methods

- `Add(widget Widget)` — sets child widget
- `Children() []Widget` — returns child widget
- `Hint() (w, h int)` — calculates preferred size from child
- `Layout()` — positions child within box

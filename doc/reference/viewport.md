# Viewport

Scrollable container for oversized content.

**Constructor:** `NewViewport(id, class, title string) *Viewport`

## Methods

- `Add(widget Widget)` — sets single scrollable child
- `Children() []Widget` — returns child widget
- `Layout()` — positions child with current scroll offsets

## Scroll axes

By default the viewport scrolls both axes. Set a flag to restrict to one axis;
the child then fills the viewport on the disabled axis instead of using its own
preferred size.

| Flag | Behaviour |
|------|-----------|
| `FlagVertical` | Vertical scrolling only — child fills viewport width, no horizontal scrollbar |
| `FlagHorizontal` | Horizontal scrolling only — child fills viewport height, no vertical scrollbar |

```go
// Vertical-only — natural for a long VFlex
vp := NewViewport("vp", "", "")
vp.SetFlag(FlagVertical, true)
vp.Add(myLongFlex)
```

Builder and Compose APIs use their existing `SetFlag` support.

## Notes

Scroll via arrow keys and `Home`/`End`. Arrow keys for the disabled axis are
ignored and passed to the next handler.

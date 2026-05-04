# Breadcrumb

Path-style indicator showing a sequence of segments separated by a glyph (default `›`). Supports overflow truncation and click-to-navigate.

**Constructor:** `NewBreadcrumb(id, class string) *Breadcrumb`

## Methods

- `Get() []string` — current segments
- `Set(segs []string)` — replace all segments
- `Push(seg string)` — append a segment
- `Pop() string` — remove and return the last segment (empty if none)
- `Truncate(index int)` — keep only the first `index` segments
- `Segments() []string` — alias for `Get()`
- `Select(index int)` — focus a specific segment; fires `EvtSelect`
- `Selected() int` — index of currently focused segment
- `SetSeparator(sep string)` — change the separator glyph (default `›`)
- `SetOverflow(marker string)` — change the overflow marker (default `…`)

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"select"` | `int` | Focused segment changed |
| `"activate"` | `int` | Enter pressed or segment clicked twice |

## Notes

Flags: `"focusable"`.

Keyboard: ←/→ move selection; Home/End jump to ends; Enter activates. When the segments don't fit in the widget's width, leading segments are replaced with the overflow marker; the focused segment is always kept visible.

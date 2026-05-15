# Flex

Linear layout container. Stacks children horizontally by default; set `FlagVertical` for a vertical stack.

**Constructor:** `NewFlex(id, class string, alignment Alignment, spacing int) *Flex`

`alignment` is a `core.Alignment` constant — `Start`, `Left`, `Center`, `Right`, `End`, `Stretch`. It controls the cross-axis (perpendicular to the layout direction). `spacing` is the gap in cells between consecutive children.

The Builder offers `HFlex(...)` and `VFlex(...)` shortcuts that wrap this constructor and pre-set the orientation flag.

## Methods

- `Add(widget Widget) error` — append a child
- `Children() []Widget` — all direct children
- `Alignment() Alignment` — current cross-axis alignment
- `Spacing() int` — current spacing
- `Layout() error` — arrange children along the main axis
- `Hint() (w, h int)` — preferred size summed from children

## Notes

Flags: `"vertical"` selects vertical orientation (children top-to-bottom instead of left-to-right).

Children's `Hint()` controls space distribution: positive values are fixed cells, negative values share the remaining space by weight, zero means "auto-size" (the widget asks for its natural size). At least one child should have a non-positive width hint on the main axis if you want the Flex to absorb all available space.

# Flex

Linear layout container (horizontal or vertical).

**Constructor:** `NewFlex(id, class string, horizontal bool, alignment string, spacing int) *Flex`

## Methods

- `Add(widget Widget)` — appends child widget
- `Children() []Widget` — returns all children
- `Hint() (w, h int)` — calculates preferred size from children
- `Layout()` — arranges children along the main axis

## Notes

**Alignment values:** `"start"`, `"center"`, `"end"`, `"stretch"`

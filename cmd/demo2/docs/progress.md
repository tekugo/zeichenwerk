# Progress

Visual progress indicator (determinate or indeterminate).

**Constructor:** `NewProgress(id, class string, horizontal bool) *Progress`

## Methods

- `Increment(amount int)` — increases current value by amount
- `Percentage() float64` — returns completion as 0–100
- `SetTotal(total int)` — sets total work units; `0` = indeterminate
- `SetValue(value int)` — sets current value (clamped to 0..total)

## Notes

When `total=0`, the bar displays a spinning indeterminate animation.

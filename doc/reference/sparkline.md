# Sparkline

Compact time-series display that renders a sequence of `float64` values as
Unicode block characters (`▁▂▃▄▅▆▇█`). The widget adapts to any content
height: with `h` rows each column has `h×8` discrete levels.

**Constructor:** `NewSparkline(id, class string) *Sparkline`

## Methods

- `Append(v float64)` — add a data point; drops the oldest if at capacity
- `SetValues(vs []float64)` — replace the entire series (copies the slice)
- `Values() []float64` — return the current series
- `SetMode(m ScaleMode)` — `Relative` or `Absolute` (default: `Relative`)
- `SetMin(v float64)` — lower bound for `Absolute` mode
- `SetMax(v float64)` — upper bound for `Absolute` mode
- `SetThreshold(v float64)` — values ≥ threshold use `"sparkline/high"` style; pass `0` to disable
- `SetGradient(v bool)` — when `true`, foreground colour blends smoothly from `"sparkline"` to `"sparkline/high"` across `[threshold, max]` instead of a hard cutoff; has no effect when threshold is 0
- `SetCapacity(n int)` — ring-buffer size; `0` = unlimited

## Notes

**Scale modes:**

| Mode | Behaviour |
|------|-----------|
| `Relative` | Tallest visible bar = `█`; rescaled on every render. Good for shape comparisons. |
| `Absolute` | Fixed `[Min, Max]` bracket. Good for showing absolute magnitude. |

**Multi-row rendering:** the content height `h` is taken from the widget's
layout bounds (`SetHint(w, h)` or whatever the parent allocates). With `h`
rows, each column uses `h×8` levels — a single `▁` at the bottom for
`level=0` through all rows filled with `█` for `level=1`.

**Left-padding:** when the series is shorter than the widget width, the
leftmost columns are rendered as blank spaces. The rightmost columns always
show the most recent data.

**Style keys:**

| Selector | When |
|----------|------|
| `"sparkline"` | Default bar colour and background |
| `"sparkline/high"` | Bars whose raw value is ≥ the threshold |

# Sparkline

A compact time-series widget that renders a sequence of numeric values using
Unicode block characters (`▁▂▃▄▅▆▇█`). Works at any height: a 1-row sparkline
gives 8 discrete levels per column; each additional row multiplies that by 8,
so a 2-row sparkline has 16 levels, a 4-row sparkline has 32, etc.
Typical use is 1 row for navigation deck items and 2–4 rows for detail panels.

## Visual layout

**Height 1** (8 levels):
```
▁▂▃▄▅▆▃▂▁▂▃▄▅▇█▆
```

**Height 2** (16 levels, same data at higher resolution):
```
      █ ██ ██
▁▂▃▄▅██▃▂▁▂▃▄▅▇██
```

**Height 4** (32 levels):
```
         █
      █ ███
   ▄▅██▃▄▅▇
▁▂▃▄▅██▃▂▁▂
```

Each character column maps to one data point. When the data series is longer
than the available width, the most recent values are shown (right-aligned).
When shorter, the leftmost columns are left empty.

## Scale modes

```go
type ScaleMode int

const (
    Relative ScaleMode = iota // highest value in the visible window = █
    Absolute                  // explicit Min/Max bracket the whole range
)
```

**Relative** mode re-scales every render pass so the tallest visible bar always
reaches the top. Good for comparing the shape of activity across multiple
sparklines side by side (e.g. per-session token rates in a navigation deck).

**Absolute** mode maps value linearly between `Min` and `Max`. Values at or
below `Min` render as `▁`; values at or above `Max` render as `█`. Good for
showing absolute magnitude over time (e.g. cost per minute against a daily
budget).

## Structure

```go
type Sparkline struct {
    Component
    values    []float64  // data points, oldest first
    mode      ScaleMode
    min       float64    // lower bound for Absolute mode
    max       float64    // upper bound for Absolute mode
    threshold float64    // optional split point for dual-colour rendering (0 = disabled)
    gradient  bool       // smooth colour interpolation across the threshold range
    capacity  int        // maximum number of stored data points (0 = unlimited)
}
```

`values` is a ring buffer when `capacity > 0`: appending a new point beyond
capacity drops the oldest entry.

## Constructor

```go
func NewSparkline(id, class string) *Sparkline
```

- `mode = Relative`, `min = 0`, `max = 1`, `threshold = 0`, `capacity = 0`.
- Not focusable (display-only widget).

## Methods

| Method | Description |
|--------|-------------|
| `Append(v float64)` | Adds a data point; trims oldest when at capacity; calls `Refresh()` |
| `SetValues(vs []float64)` | Replaces the entire series; calls `Refresh()` |
| `Values() []float64` | Returns the current series slice |
| `SetMode(m ScaleMode)` | Sets the scale mode; calls `Refresh()` |
| `SetMin(v float64)` | Sets the lower bound for Absolute mode |
| `SetMax(v float64)` | Sets the upper bound for Absolute mode |
| `SetThreshold(v float64)` | Sets the split value for dual-colour rendering (0 disables) |
| `SetGradient(v bool)` | Enables smooth colour interpolation from base to high across `[threshold, max]`; has no effect when threshold is 0 |
| `SetCapacity(n int)` | Sets the maximum retained data points (trims excess immediately) |

## Rendering

```go
func (s *Sparkline) Render(r *Renderer)
```

1. `s.Component.Render(r)` — draws background and border.
2. Compute content area `(cx, cy, cw, ch)`. Use `h = ch` as the bar height.
3. Select the rightmost `cw` values from `s.values` (or fewer if the series is
   shorter).
4. Determine the effective range:
   - **Relative**: `lo = min(visible)`, `hi = max(visible)`. If `lo == hi`,
     treat all values as mid-range (`level = 0.5`).
   - **Absolute**: `lo = s.min`, `hi = s.max`.
5. For each column `i` in `0 … cw-1`:
   - If no corresponding data point exists, write spaces in all `h` rows.
   - Otherwise:
     ```
     level      = clamp((value - lo) / (hi - lo), 0, 1)
     totalSteps = h * 8
     step       = int(level * float64(totalSteps-1) + 0.5)  // 0 … h*8-1
     fullRows   = step / 8
     partial    = step % 8
     ```
   - For each row `r` in `0 … h-1` (0 = top):
     - `rowFromBottom = h - 1 - r`
     - If `rowFromBottom < fullRows`: character is `█`.
     - If `rowFromBottom == fullRows`: character is `blocks[partial]`
       where `blocks = "▁▂▃▄▅▆▇█"`.
     - If `rowFromBottom > fullRows`: character is space.
   - Choose the foreground colour for this column:
     - If `s.threshold > 0` and `s.gradient`:
       ```
       threshLevel = (s.threshold - lo) / (hi - lo)   // threshold position in [0,1]
       t = clamp((level - threshLevel) / (1 - threshLevel), 0, 1)
       fg = lerpColor(baseFg, highFg, t)
       ```
       where `baseFg` and `highFg` are the resolved foreground colours of
       `"sparkline"` and `"sparkline/high"` respectively. `t = 0` below the
       threshold (base colour); `t = 1` at the maximum value (full high colour).
     - Else if `s.threshold > 0` and `value >= s.threshold`: use `"sparkline/high"`.
     - Otherwise: use `"sparkline"`.
   - Apply the selected foreground with `baseStyle`'s background and font.

This maps `level = 0` to a single `▁` in the bottom row (minimum visible bar)
and `level = 1` to all rows filled with `█`. At height 1 the behaviour is
identical to the previous 8-level single-row rendering.

## Hint

```go
func (s *Sparkline) Hint() (int, int)
```

- Width: manually set hint, or `len(s.values)` if no hint is set (never less
  than 1).
- Height: manually set hint, or `1` if unset.

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"sparkline"` | Background, border, and default bar colour |
| `"sparkline/high"` | Bars whose value is at or above the threshold |

## Implementation plan

1. **`sparkline.go`** — new file
   - Define `ScaleMode` constants and `Sparkline` struct.
   - Implement `NewSparkline`, `Append`, `SetValues`, `Values`, `SetMode`,
     `SetMin`, `SetMax`, `SetThreshold`, `SetCapacity`.
   - Implement `Apply`, `Hint`, `Render` with the block-character mapping
     and dual-colour threshold logic.

2. **`builder.go`** — add `Sparkline` method
   ```go
   func (b *Builder) Sparkline(id string) *Builder
   ```

3. **Theme** — add `"sparkline"` and `"sparkline/high"` style entries to
   built-in themes.

4. **Tests** — `sparkline_test.go`
   - Height 1, Relative: maximum value → `█`; minimum → `▁`.
   - Height 1, Relative, all equal values → all `▄` (mid-range fallback).
   - Height 1, Absolute: value at `Min` → `▁`; at `Max` → `█`; below `Min`
     clamps to `▁`; above `Max` clamps to `█`.
   - Height 2: `level = 0` → bottom row `▁`, top row space.
   - Height 2: `level = 1` → both rows `█`.
   - Height 2: `level = 0.5` → bottom row `█`, top row space (step=7).
   - Height 4: `level = 1` → all four rows `█`.
   - Height 4: `level = 0.25` → bottom row `█`, rows above empty (step=7 with h=4 gives step=int(0.25*31+0.5)=8 → fullRows=1, partial=0 → bottom row █, rest empty).
   - At height 1, rendering matches the old 8-level formula exactly.
   - Series longer than width shows only the rightmost `w` values.
   - Series shorter than width pads left with spaces (all rows).
   - `Append` with `capacity = 5` drops the oldest value on the sixth append.
   - Threshold (hard cutoff): values below threshold use default style; values
     at or above use `"sparkline/high"` across all rows of that column.
   - Gradient disabled: value at max above threshold renders exact high colour.
   - Gradient enabled: value below threshold renders base colour (`t=0`); value
     at max renders full high colour (`t=1`); value at midpoint of
     `[threshold, max]` renders the linearly interpolated midpoint colour.
   - `SetGradient(false)` preserves the hard-cutoff behaviour.

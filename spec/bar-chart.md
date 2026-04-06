# BarChart

A stacked bar chart widget that renders multiple data series as vertically
stacked columns (or horizontally stacked rows). Each bar is divided into
labelled segments, one per series, drawn with Unicode block characters for
sub-character precision at segment boundaries. Supports a y-axis with tick
labels, a category axis, an optional legend, and keyboard navigation to select
a category.

---

## Visual layout

**Vertical (default), 3 series, 4 categories:**

```
 30 │               ▄▄▄
    │         ▄▄▄   ░░░
 20 │   ▄▄▄   ░░░   ░░░
    │   ░░░   ░░░   ▒▒▒   ▄▄▄
 10 │   ░░░   ▒▒▒   ▒▒▒   ░░░
    │   ▒▒▒   ███   ███   ▒▒▒
  0 └───┬─────┬─────┬─────┬───
       Jan   Feb   Mar   Apr

    ███ Revenue   ▒▒▒ Cost   ░░░ Profit   ▄▄▄ Tax
```

- **Y-axis** — tick labels on the left, `│` vertical rule, horizontal grid
  lines at each tick (drawn with `─` in the `"bar-chart/grid"` style).
- **Bars** — each `barWidth` columns wide, separated by `barGap` empty columns.
  Segments stack bottom-to-top in series order.
- **X-axis** — `└` corner, `─` baseline, `┬` tick under each bar centre.
- **Category labels** — centred below each bar's tick.
- **Legend** — one swatch per series, drawn below the category labels when
  `legend = true`.

**Horizontal mode:**

```
Jan  ████████▓▓▓▓░░░░░░▄▄▄ 30
Feb  ██████████▓▓▓▓▓░░░░░  27
Mar  ████████████▓▓▓░░░░░░ 28
Apr  ██████▓▓▓▓▓░░░░▄▄▄▄▄  26
```

Bars grow left to right; categories are the rows. The total value label is
drawn to the right of each bar when `showValues = true`.

---

## Data model

```go
type BarSeries struct {
    Label  string    // series name shown in the legend
    Values []float64 // one value per category, all >= 0
}
```

A `BarChart` holds an ordered slice of series and a parallel slice of category
labels. Bar `b` stacks `series[0].Values[b]` at the bottom, `series[1].Values[b]`
above it, and so on.

---

## Structure

```go
type BarChart struct {
    Component
    series     []BarSeries
    categories []string
    mode       ScaleMode  // Relative or Absolute (reused from Sparkline)
    max        float64    // explicit maximum for Absolute mode
    horizontal bool
    showAxis   bool       // draw y-axis labels and rule (default true)
    showGrid   bool       // draw horizontal grid lines at y-axis ticks (default true)
    showValues bool       // draw total-value label above/beside each bar
    legend     bool       // draw legend below chart (default true)
    barWidth   int        // column width per bar in vertical mode (default 3)
    barGap     int        // empty columns between bars (default 1)
    selected   int        // index of the focused category (-1 = none)
    ticks      int        // approximate number of y-axis ticks (default 5)
}
```

---

## Constructor

```go
func NewBarChart(id, class string) *BarChart
```

- `showAxis = true`, `showGrid = true`, `legend = true`, `barWidth = 3`,
  `barGap = 1`, `ticks = 5`, `mode = Relative`, `selected = -1`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

---

## Methods

### Data

| Method | Description |
|--------|-------------|
| `SetSeries(s []BarSeries)` | Replaces all series; calls `Refresh()` |
| `AddSeries(s BarSeries)` | Appends a series; calls `Refresh()` |
| `SetCategories(labels []string)` | Replaces category labels; calls `Refresh()` |
| `Series() []BarSeries` | Returns the current series slice |
| `Categories() []string` | Returns the current category labels |

### Display

| Method | Description |
|--------|-------------|
| `SetMode(m ScaleMode)` | `Relative` or `Absolute`; calls `Refresh()` |
| `SetMax(v float64)` | Explicit maximum for `Absolute` mode |
| `SetHorizontal(v bool)` | Switches bar orientation; calls `Refresh()` |
| `SetShowAxis(v bool)` | Show or hide the y-axis labels and rule |
| `SetShowGrid(v bool)` | Show or hide horizontal grid lines |
| `SetShowValues(v bool)` | Show or hide the per-bar total label |
| `SetLegend(v bool)` | Show or hide the series legend |
| `SetBarWidth(w int)` | Column width per bar in vertical mode (minimum 1) |
| `SetBarGap(g int)` | Empty columns between bars (minimum 0) |
| `SetTicks(n int)` | Approximate number of y-axis ticks (minimum 2) |

### Navigation

| Method | Description |
|--------|-------------|
| `Select(index int)` | Focuses a category; clamps to valid range; dispatches `EvtSelect` |
| `Selected() int` | Returns the focused category index (-1 if none) |

---

## Scale modes

Reuses the `ScaleMode` type from `Sparkline`.

**Relative** — the maximum visible stack total equals the full chart height.
The y-axis ticks are computed from the actual data maximum. Suitable for
comparing relative magnitudes across categories.

**Absolute** — the y-axis range is `[0, max]`. Stack values above `max` are
clamped. Suitable for tracking progress against a known ceiling (e.g. a budget).

---

## Axis layout

### Y-axis (vertical mode)

Ticks are computed as:

```
step  = niceCeil(effectiveMax / ticks)   // rounded to 1, 2, 5, 10, 20 …
labels = [0, step, 2*step, …]            // up to and including effectiveMax
yAxisW = max(len(label) for all labels) + 2  // +2 for "│ " separator
```

`niceCeil` rounds up to the nearest value of the form `k × 10^n` where
`k ∈ {1, 2, 5}`.

Tick labels are drawn right-aligned in `yAxisW - 2` columns; the `│` rule
occupies column `cx + yAxisW - 1`.

### X-axis (vertical mode)

The baseline is drawn at `cy + chartH`:

```
cx + yAxisW - 1  →  └
cx + yAxisW      …  ─ repeated for chartW columns
under each bar centre  →  ┬
```

Category labels are centred below each `┬` tick. Labels longer than
`barWidth + barGap - 1` are truncated.

### Axis characters (from theme strings)

| Key | Default | Description |
|-----|---------|-------------|
| `bar-chart.corner`   | `└` | Bottom-left axis corner |
| `bar-chart.hline`    | `─` | Horizontal axis line |
| `bar-chart.vline`    | `│` | Vertical axis rule |
| `bar-chart.tick-x`   | `┬` | X-axis tick under each bar |
| `bar-chart.tick-y`   | `┤` | Y-axis tick at each grid row |
| `bar-chart.grid`     | `─` | Grid line character (repeated across chartW) |
| `bar-chart.swatch`   | `█` | Legend colour swatch character |

---

## Segment rendering (vertical mode)

For each bar `b` and each series `i` (bottom = 0):

```
cumBelow  = sum(series[j].Values[b] for j < i)
cumTop    = cumBelow + series[i].Values[b]
rowBelow  = cumBelow / effectiveMax * chartH   // floating point rows from bottom
rowTop    = cumTop   / effectiveMax * chartH
```

Converting to screen rows (0 = chart top, chartH-1 = chart bottom):

```
screenBottom = cy + chartH - 1 - floor(rowBelow)
screenTop    = cy + chartH - 1 - floor(rowTop)
```

For each screen row `r` in `screenTop … screenBottom`:

- **Full rows** (`r` is entirely within the segment):
  draw `barWidth` copies of `█` with `"bar-chart/s<i>"` style.

- **Top partial row of the segment** (`r == screenTop` and `rowTop` has a
  fractional part `f = rowTop - floor(rowTop) > 0`):
  draw `blocks[int(f * 8)]` with:
  - `fg` = colour of series `i`
  - `bg` = colour of series `i+1` (or chart background if `i` is the top segment)

  This gives sub-character precision at segment boundaries.

- **Bottom partial row** (where this segment meets the one below):
  the lower segment already painted that row; this segment's contribution
  begins at the fractional position and is covered by the top-partial rule above
  applied to series `i-1`.

The selected category (when focused) is highlighted by drawing a `▲` indicator
one row above the bar using the `"bar-chart/selection"` style, and drawing the
category label in the same style.

---

## Legend

Rendered as a single row below the category labels when `legend = true`:

```
    █ Series A   ▒ Series B   ░ Series C
```

Each entry is `swatch + " " + series.Label`. Entries are space-separated and
left-aligned from `cx`. The legend is omitted if all series labels are empty.

Legend height is always 1 row when shown.

---

## Hint

```go
func (c *BarChart) Hint() (int, int)
```

**Vertical mode:**

- Width: manually set hint, or
  `yAxisW + len(categories) * (barWidth + barGap) - barGap + style.Horizontal()`
- Height: manually set hint, or 0 (fills parent).

**Horizontal mode:**

- Width: manually set hint, or 0 (fills parent).
- Height: manually set hint, or
  `len(categories) * (barHeight + barGap) - barGap + xAxisH + legendH + style.Vertical()`
  where `barHeight = 1` and `xAxisH = 1` (value axis) and `legendH = 0 or 1`.

---

## Keyboard interaction

| Key | Behaviour |
|-----|-----------|
| `←` / `→` | `Select(selected ± 1)` — vertical mode |
| `↑` / `↓` | `Select(selected ± 1)` — horizontal mode |
| `Home` | `Select(0)` |
| `End` | `Select(len(categories) - 1)` |
| `Enter` | Dispatch `EvtActivate` with `selected` |

---

## Mouse interaction

A click at `(mouseX, mouseY)` within the chart area maps to a category:

**Vertical mode:**
```
b = (mouseX - cx - yAxisW) / (barWidth + barGap)
```

If `b` is within range, call `Select(b)`. A second click on the already-selected
category dispatches `EvtActivate`.

**Horizontal mode:** analogous using `mouseY` and row height.

---

## Events

| Event | Data | Description |
|-------|------|-------------|
| `EvtSelect` | `int` | Focused category index changed |
| `EvtActivate` | `int` | Enter pressed or category clicked twice |

---

## Styling selectors

Series colours use indexed sub-part selectors `s0`–`s7`. Charts with more than
8 series wrap around (series 8 reuses `s0`, etc.).

| Selector | Applied to |
|----------|-----------|
| `"bar-chart"` | Background and border |
| `"bar-chart/s0"` … `"bar-chart/s7"` | Fill colour for each series |
| `"bar-chart/axis"` | Y-axis labels, baseline, and tick characters |
| `"bar-chart/grid"` | Horizontal grid lines |
| `"bar-chart/label"` | Category labels on the x-axis |
| `"bar-chart/label:focused"` | Category label for the selected bar |
| `"bar-chart/selection"` | `▲` indicator above the focused bar |
| `"bar-chart/value"` | Total-value label above/beside each bar |
| `"bar-chart/legend"` | Legend row text |

Example theme entries (Tokyo Night):

```go
NewStyle("bar-chart").WithBorder("none").WithPadding(0, 1),
NewStyle("bar-chart/s0").WithForeground("$blue"),
NewStyle("bar-chart/s1").WithForeground("$green"),
NewStyle("bar-chart/s2").WithForeground("$yellow"),
NewStyle("bar-chart/s3").WithForeground("$red"),
NewStyle("bar-chart/s4").WithForeground("$cyan"),
NewStyle("bar-chart/s5").WithForeground("$purple"),
NewStyle("bar-chart/s6").WithForeground("$orange"),
NewStyle("bar-chart/s7").WithForeground("$fg1"),
NewStyle("bar-chart/axis").WithForeground("$fg1"),
NewStyle("bar-chart/grid").WithForeground("$bg3"),
NewStyle("bar-chart/label").WithForeground("$fg1"),
NewStyle("bar-chart/label:focused").WithColors("$bg0", "$blue").WithFont("bold"),
NewStyle("bar-chart/selection").WithForeground("$blue"),
NewStyle("bar-chart/value").WithForeground("$fg0").WithFont("bold"),
NewStyle("bar-chart/legend").WithForeground("$fg1"),
```

---

## Implementation plan

1. **`bar-chart.go`** — new file
   - Define `BarSeries` struct.
   - Define `BarChart` struct and `NewBarChart`.
   - Implement setters: `SetSeries`, `AddSeries`, `SetCategories`, `SetMode`,
     `SetMax`, `SetHorizontal`, `SetShowAxis`, `SetShowGrid`, `SetShowValues`,
     `SetLegend`, `SetBarWidth`, `SetBarGap`, `SetTicks`, `Select`, `Selected`.
   - Implement helpers: `effectiveMax()`, `niceCeil(v float64) float64`,
     `yAxisLayout()` (returns tick values, labels, yAxisW), `seriesColor(i)`.
   - Implement `Apply`, `Hint`, `Render` (vertical), `renderHorizontal`.
   - Implement `handleKey`, `handleMouse`.

2. **`builder.go`** — add `BarChart` method
   ```go
   func (b *Builder) BarChart(id string) *Builder
   ```

3. **Theme** — add `"bar-chart"` family and `bar-chart.*` string keys to all
   built-in themes with the 8-colour series palette.

4. **Tests** — `bar-chart_test.go`
   - `effectiveMax` returns the correct stack total across categories in
     Relative mode; returns `max` in Absolute mode.
   - `niceCeil` rounds to the correct 1/2/5 multiplier for various inputs.
   - `yAxisLayout` produces correct tick labels and `yAxisW` for typical data.
   - Single series renders identically to a non-stacked bar chart.
   - Two series: segment boundary at the correct row; sub-character partial
     row uses correct block character with correct fg/bg colours.
   - Category with all-zero values renders an empty (baseline-only) bar.
   - `Select` clamps to valid range; dispatches `EvtSelect`.
   - `SetSeries` with no categories falls back gracefully (no render panic).
   - Horizontal mode: bar heights correspond to category count; bar widths
     are proportional to stack totals.
   - Legend is omitted when all series labels are empty.
   - `showValues` draws the total-value label at the correct position.

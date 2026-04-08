# LineChart

The natural companion to `BarChart` and `Sparkline`. Renders one or more data
series as smooth lines over a shared x-axis. Sub-character vertical precision
is achieved with Braille dots (the same technique as `Gauge`), giving smooth
curves at any chart height. Supports an optional area fill below each line and
a ring-buffer streaming mode that scrolls the x-axis left as new values arrive.

Shares `ScaleMode`, `niceCeil`, and y-axis layout logic with `BarChart`.

---

## Visual layout

**Single series, 4 categories:**

```
 30 │         · · ·
    │     · ·       · ·
 20 │ · ·               · ·
    │
 10 │
  0 └──┬──────┬──────┬──────┬──
      Jan    Feb    Mar    Apr
```

**Two series:**

```
 30 │         · · ·
    │     · ·       · · · ╌ ╌
 20 │ · ·               ╌     ╌ ╌
    │                 ╌
 10 │               ╌
  0 └──┬──────┬──────┬──────┬──
      Jan    Feb    Mar    Apr

     · Series A    ╌ Series B
```

**Single series with area fill:**

```
 30 │         ▒ ▒ ▒
    │     ▒ ▒ ▒ ▒ ▒ ▒ ▒
 20 │ ▒ ▒ ▒ ▒ ▒ ▒ ▒ ▒ ▒ ▒ ▒
    │
 10 │
  0 └──┬──────┬──────┬──────┬──
      Jan    Feb    Mar    Apr
```

The y-axis rule and tick labels are drawn on the left. The baseline and
category labels are drawn along the bottom. In streaming mode the x-axis
carries no category labels and the lines scroll left as new values arrive.

---

## Data model

```go
type LinePoint struct {
    Value float64 // math.NaN() signals a gap in the line
}

type LineSeries struct {
    Label  string    // series name shown in the legend
    Values []float64 // data points, oldest first; NaN creates a gap
}
```

Multiple series share the same x-axis. Series `i` uses values at index `j`
as the y position for x-category `j`. When series have different lengths the
chart renders up to the longest; shorter series leave the remaining x positions
empty.

---

## Structure

```go
type LineChart struct {
    Component
    series     []LineSeries
    categories []string   // x-axis labels; may be empty (streaming or unlabelled)
    mode       ScaleMode  // Relative or Absolute (reused from Sparkline/BarChart)
    min        float64    // explicit lower bound for Absolute mode
    max        float64    // explicit upper bound for Absolute mode
    showAxis   bool       // draw y-axis labels and rule (default true)
    showGrid   bool       // draw horizontal grid lines at y-axis ticks (default true)
    showFill   bool       // draw area fill below each line (default false)
    legend     bool       // draw legend below chart (default true)
    capacity   int        // ring-buffer size per series; 0 = unlimited
    ticks      int        // approximate number of y-axis ticks (default 5)
    selected   int        // focused x-index (-1 = none)
}
```

---

## Constructor

```go
func NewLineChart(id, class string) *LineChart
```

- `showAxis = true`, `showGrid = true`, `showFill = false`, `legend = true`.
- `ticks = 5`, `mode = Relative`, `selected = -1`, `capacity = 0`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

---

## Methods

### Data

| Method | Description |
|--------|-------------|
| `SetSeries(s []LineSeries)` | Replaces all series; calls `Refresh()` |
| `AddSeries(s LineSeries)` | Appends a series; calls `Refresh()` |
| `Append(seriesIndex int, v float64)` | Appends one value to the given series (streaming mode); trims the oldest value when at capacity; calls `Refresh()` |
| `SetCategories(labels []string)` | Replaces x-axis category labels; calls `Refresh()` |
| `Series() []LineSeries` | Returns the current series slice |
| `Categories() []string` | Returns the current category labels |

### Display

| Method | Description |
|--------|-------------|
| `SetMode(m ScaleMode)` | `Relative` or `Absolute`; calls `Refresh()` |
| `SetMin(v float64)` | Explicit lower bound for `Absolute` mode |
| `SetMax(v float64)` | Explicit upper bound for `Absolute` mode |
| `SetShowAxis(v bool)` | Show or hide y-axis labels and rule |
| `SetShowGrid(v bool)` | Show or hide horizontal grid lines |
| `SetShowFill(v bool)` | Show or hide the area fill below each line |
| `SetLegend(v bool)` | Show or hide the series legend |
| `SetCapacity(n int)` | Ring-buffer depth per series; trims oldest values immediately if already over capacity; `0` = unlimited |
| `SetTicks(n int)` | Approximate number of y-axis ticks (minimum 2) |

### Navigation

| Method | Description |
|--------|-------------|
| `Select(index int)` | Focuses an x-index; clamps to the valid range; dispatches `EvtSelect` |
| `Selected() int` | Returns the focused x-index (-1 if none) |

---

## Scale modes

Reuses the `ScaleMode` type from `Sparkline` and `BarChart`.

**Relative** — the effective y range is derived from the minimum and maximum
values visible across all series. When all visible values are equal, mid-range
is used. Suitable for comparing the shape of series relative to each other.

**Absolute** — the y range is `[min, max]`. Values outside this range are
clamped to the boundary. Suitable for tracking values against a known scale.

---

## Axis layout

### Y-axis

Reuses the same algorithm as `BarChart`:

```
step   = niceCeil(effectiveRange / ticks)   // rounded to 1, 2, 5, 10 …
labels = [lo, lo+step, lo+2·step, …]        // up to and including hi
yAxisW = max(len(label) for all labels) + 2  // +2 for "│ " separator
```

`niceCeil` rounds up to the nearest `k × 10^n` where `k ∈ {1, 2, 5}`.

Tick labels are right-aligned in `yAxisW − 2` columns; the `│` rule occupies
column `cx + yAxisW − 1`. In `Absolute` mode `lo = min`, `hi = max`; in
`Relative` mode `lo` and `hi` are computed from the data.

### X-axis

When `categories` is non-empty:

```
baseline row: cy + chartH           →  └ ─ … ─
under each data point centre        →  ┬ (tick)
category label centred below tick
```

Labels wider than the available column pitch are truncated. When `categories`
is empty (streaming or unlabelled), only the baseline `└─` is drawn with no
ticks or labels.

### Axis theme strings

| Key | Default | Description |
|-----|---------|-------------|
| `line-chart.corner`  | `└` | Bottom-left axis corner |
| `line-chart.hline`   | `─` | Horizontal axis line |
| `line-chart.vline`   | `│` | Vertical axis rule |
| `line-chart.tick-x`  | `┬` | X-axis tick under each data point |
| `line-chart.tick-y`  | `┤` | Y-axis tick at each grid row |
| `line-chart.grid`    | `─` | Grid line character (repeated across chartW) |
| `line-chart.swatch`  | `█` | Legend colour swatch character |
| `line-chart.cursor`  | `│` | Vertical cursor at the selected x-index |

---

## Rendering

### Braille dot grid

The chart area (after reserving space for the y-axis, x-axis, legend, and
cursor row) is `(cx + yAxisW, cy, chartW, chartH)`. The Braille dot grid
covering this area has dimensions:

```
dotW = chartW * 2   // two dot columns per character column
dotH = chartH * 4   // four dot rows    per character row
```

Each dot `(dx, dy)` maps to character cell `(cx + yAxisW + dx/2, cy + dy/4)`
and Braille bit index `(dx % 2) + (dy % 4) * 2` (0-indexed, row-major within
the 2×4 cell). The character is assembled by OR-ing dot bits from all series
into a single Braille codepoint per cell.

### Mapping a value to dot coordinates

For a data point with `fraction = clamp((value - lo) / (hi - lo), 0, 1)`:

```
dy = int((1.0 - fraction) * float64(dotH - 1) + 0.5)
```

A data point at x-index `xi` maps to dot column:

```
dx = xi * dotW / dataLen   // scales xi into [0, dotW)
```

where `dataLen` is the number of data points in the longest series.

### Line drawing

For each consecutive pair of data points `(xi, valueA)` and `(xi+1, valueB)`
in a series, compute `(dxA, dyA)` and `(dxB, dyB)` and draw a Bresenham line
between those two dot positions. Each dot along the path is OR-ed into the
appropriate Braille cell for that series.

NaN values break the line: no segment is drawn from or to a NaN data point.

### Area fill

When `showFill` is true, for every dot column `dx` where a data point exists,
set all dots from `dy + 1` through `dotH − 1` (the chart baseline). Fill dots
are merged into the same Braille cells as the line dots, using the same series
colour but rendered with the `"line-chart/fill"` style.

### Multi-series overlay

Each series is drawn into the same Braille dot grid independently. The dots
for different series are OR-ed together into shared Braille cells, so series
lines may visually merge where they overlap. The per-cell foreground colour is
that of the highest-indexed series that contributed a set dot in that cell.

### Selection cursor

When `selected ≥ 0`, a vertical cursor is drawn at the character column
corresponding to x-index `selected`, spanning the full `chartH` rows, using
the `"line-chart/axis"` style and the `line-chart.cursor` theme string. The
cursor is drawn after the Braille pass so it is always visible. Category labels
and y-values for all series at that index may optionally be displayed in a
single-row label above the chart area using the `"line-chart/cursor"` style.

### Render steps

```go
func (c *LineChart) Render(r *Renderer)
```

1. `c.Component.Render(r)` — draws background and border.
2. Compute content area; derive `chartW`, `chartH`, `yAxisW`, `legendH`,
   `xAxisH`, `dotW`, `dotH`.
3. Compute effective `lo` / `hi` from scale mode and data.
4. Draw y-axis tick labels and `│` rule if `showAxis`.
5. Draw horizontal grid lines at tick rows if `showGrid`.
6. For each series `i` (in ascending index order):
   a. Build line dot positions using Bresenham between consecutive points.
   b. If `showFill`, accumulate fill dots below each line point.
   c. OR dot bits into per-cell Braille accumulators keyed by series index.
7. Render each Braille cell: assemble the codepoint from OR-ed dot bits;
   apply `"line-chart/sN"` foreground for the highest contributing series `N`;
   apply `"line-chart/fill"` style for cells that contain only fill dots.
8. Draw x-axis baseline, ticks, and category labels if `categories` is set.
9. Draw the selection cursor and label if `selected ≥ 0`.
10. Draw the legend if `legend` is true.

---

## Streaming mode

When `capacity > 0`, each series is a ring buffer of at most `capacity` values.
`Append(i, v)` pushes `v` onto series `i` and drops the oldest entry when the
buffer is full, then calls `Refresh()`. The chart always displays the most
recent `capacity` points right-aligned in the chart area. `categories` is
typically empty in streaming mode; the x-axis renders only the baseline.

---

## Legend

Rendered as a single row below the x-axis labels when `legend` is true:

```
     █ Series A    ▒ Series B    ░ Series C
```

Each entry is `swatch + " " + series.Label`. Entries are space-separated,
left-aligned from `cx`. The legend is omitted when all series labels are empty.
Legend height is always 1 when shown.

---

## Hint

```go
func (c *LineChart) Hint() (int, int)
```

- Width: manually set hint, or `0` (fills parent).
- Height: manually set hint, or `0` (fills parent).

---

## Keyboard interaction

| Key | Behaviour |
|-----|-----------|
| `←` | `Select(selected − 1)` |
| `→` | `Select(selected + 1)` |
| `Home` | `Select(0)` |
| `End` | `Select(dataLen − 1)` |
| `Enter` | Dispatch `EvtActivate` with `selected` |

When `selected` is `−1` and the user presses `→`, selection starts at `0`.
When `selected` is `−1` and the user presses `←`, selection starts at
`dataLen − 1`.

---

## Mouse interaction

A click at `mouseX` within the chart area maps to the nearest x-index:

```
xi = clamp(int(float64(mouseX - cx - yAxisW) / float64(chartW) * float64(dataLen) + 0.5),
           0, dataLen - 1)
```

Call `Select(xi)`. A second click on the already-selected index dispatches
`EvtActivate`.

---

## Events

| Event | Data | Description |
|-------|------|-------------|
| `EvtSelect` | `int` | Focused x-index changed |
| `EvtActivate` | `int` | Enter pressed or x-index clicked twice |

---

## Styling selectors

Series colours use indexed sub-part selectors `s0`–`s7`. Charts with more than
8 series wrap (series 8 reuses `s0`, etc.).

| Selector | Applied to |
|----------|-----------|
| `"line-chart"` | Background and border |
| `"line-chart/s0"` … `"line-chart/s7"` | Line colour for each series |
| `"line-chart/axis"` | Y-axis labels, baseline, ticks, and cursor |
| `"line-chart/grid"` | Horizontal grid lines |
| `"line-chart/fill"` | Area fill below the line (all series share this style; the series foreground colour is applied on top) |
| `"line-chart/label"` | X-axis category labels |
| `"line-chart/label:focused"` | Category label at the selected x-index |
| `"line-chart/cursor"` | Value readout label drawn above the cursor |
| `"line-chart/legend"` | Legend row text |

Example theme entries (Tokyo Night):

```go
NewStyle("line-chart").WithBorder("none").WithPadding(0, 1),
NewStyle("line-chart/s0").WithForeground("$blue"),
NewStyle("line-chart/s1").WithForeground("$green"),
NewStyle("line-chart/s2").WithForeground("$yellow"),
NewStyle("line-chart/s3").WithForeground("$red"),
NewStyle("line-chart/s4").WithForeground("$cyan"),
NewStyle("line-chart/s5").WithForeground("$purple"),
NewStyle("line-chart/s6").WithForeground("$orange"),
NewStyle("line-chart/s7").WithForeground("$fg1"),
NewStyle("line-chart/axis").WithForeground("$fg1"),
NewStyle("line-chart/grid").WithForeground("$bg3"),
NewStyle("line-chart/fill").WithForeground("$bg3"),
NewStyle("line-chart/label").WithForeground("$fg1"),
NewStyle("line-chart/label:focused").WithColors("$bg0", "$blue").WithFont("bold"),
NewStyle("line-chart/cursor").WithForeground("$blue").WithFont("bold"),
NewStyle("line-chart/legend").WithForeground("$fg1"),
```

---

## Implementation plan

1. **`line-chart.go`** — new file
   - Define `LineSeries` struct.
   - Define `LineChart` struct and `NewLineChart`.
   - Implement setters: `SetSeries`, `AddSeries`, `Append`, `SetCategories`,
     `SetMode`, `SetMin`, `SetMax`, `SetShowAxis`, `SetShowGrid`, `SetShowFill`,
     `SetLegend`, `SetCapacity`, `SetTicks`, `Select`, `Selected`.
   - Implement helpers shared with `BarChart`: `effectiveBounds()` (returns `lo`,
     `hi`), `niceCeil` (import or reuse from `bar-chart.go`), `yAxisLayout()`.
   - Implement Braille helpers: `dotGrid` type (sparse map or `[][]uint8`),
     `setDot(dx, dy int)`, `bresenham(x0, y0, x1, y1 int)`, `assembleCells()`.
   - Implement `Hint`, `Render`, `handleKey`, `handleMouse`.

2. **`builder.go`** — add `LineChart` method
   ```go
   func (b *Builder) LineChart(id string) *Builder
   ```

3. **Theme** — add `"line-chart"` family and `line-chart.*` string keys to all
   built-in themes with the 8-colour series palette.

4. **Tests** — `line-chart_test.go`
   - `effectiveBounds` in Relative mode returns the global min/max across all
     series; in Absolute mode returns `min` / `max`.
   - A single data point renders a Braille dot at the correct vertical position
     for `fraction = 0`, `0.5`, and `1`.
   - Two consecutive equal values produce a horizontal line (all dots in the
     same dot row between the two x positions).
   - A NaN value breaks the line: no Bresenham segment is drawn to or from it.
   - `showFill = true`: dot cells below the line position are set; dot cells
     above it are clear.
   - Two series: OR-ing produces a combined Braille character where both series'
     dots are present.
   - `Append` with `capacity = 10` drops the oldest value on the eleventh call.
   - Streaming mode with `capacity` set and no categories renders only the
     baseline (`└─`) with no ticks or labels.
   - `Select` clamps to `[0, dataLen-1]`; dispatches `EvtSelect`.
   - Mouse click maps to the nearest x-index.
   - Legend is omitted when all series labels are empty.
   - `yAxisLayout` produces correct tick labels and `yAxisW` consistent with
     `BarChart` for the same data range and tick count.

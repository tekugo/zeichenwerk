# BarChart

Multi-series stacked bar chart with optional y-axis, grid, value labels, and legend. Bars render vertically (default) or horizontally.

**Constructor:** `NewBarChart(id, class string) *BarChart`

Configure data, axes, and orientation via the setter methods after construction.

## Data types

```go
type BarSeries struct {
    Label  string    // legend label; may be empty
    Values []float64 // one value per category, all >= 0
}
```

## Methods

- `SetSeries(s []BarSeries)` — replace all series
- `AddSeries(s BarSeries)` — append a series
- `Series() []BarSeries` — current series
- `SetCategories(labels []string)` — category labels (one per index of `Values`)
- `Categories() []string` — current category labels
- `SetAbsolute(v bool)` — when true, values are drawn as absolute heights instead of stacked
- `SetMax(v float64)` — explicit y-axis max (zero means auto)
- `SetHorizontal(v bool)` — flip to horizontal orientation
- `SetShowAxis(v bool)` — toggle y-axis labels
- `SetShowGrid(v bool)` — toggle background grid
- `SetShowValues(v bool)` — toggle per-bar value labels
- `SetLegend(v bool)` — toggle legend
- `SetBarWidth(w int)` — width of each bar in cells (≥ 1)
- `SetBarGap(g int)` — gap between bars (≥ 0)
- `SetTicks(n int)` — number of y-axis ticks (≥ 2)
- `Select(index int)` — set focused category; fires `EvtSelect`
- `Selected() int` — current focused category

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"select"` | `int` | Focused category index changed |
| `"activate"` | `int` | Enter pressed or category clicked twice |

## Notes

Flags: `"focusable"`.

Keyboard: ←/→ (or ↑/↓ when horizontal) move the selection by one category; Home/End jump to the first/last; Enter activates. Single mouse click selects, double click activates.

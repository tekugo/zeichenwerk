package zeichenwerk

import (
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
)

// ==== AI ===================================================================

// barChartBlocks maps a [0, 8] integer to the Unicode block-fill character
// that fills that many eighths of a cell from the bottom up.
// 0 = space (empty), 8 = full block (█).
var barChartBlocks = []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// BarSeries is a single data series in a [BarChart].
// Each series contributes one stacked segment per category.
type BarSeries struct {
	Label  string    // series name shown in the legend; may be empty
	Values []float64 // one value per category, all >= 0
}

// BarChart renders multiple data series as a stacked bar chart with optional
// y-axis, grid, category labels, value labels, and a legend.
// Bars can be oriented vertically (default) or horizontally.
//
// Events:
//   - [EvtSelect]   – int: focused category index changed
//   - [EvtActivate] – int: Enter pressed or category clicked twice
type BarChart struct {
	Component
	series     []BarSeries
	categories []string
	mode       ScaleMode
	max        float64
	horizontal bool
	showAxis   bool
	showGrid   bool
	showValues bool
	legend     bool
	barWidth   int
	barGap     int
	selected   int
	ticks      int
	// Characters read from theme strings in Apply; defaults set in constructor.
	chCorner string
	chHLine  string
	chVLine  string
	chTickX  string
	chTickY  string
	chGrid   string
	chSwatch string
}

// NewBarChart creates a new BarChart with sensible defaults: vertical
// orientation, y-axis visible, grid visible, legend visible, barWidth = 3,
// barGap = 1, 5 y-axis ticks, Relative scale, no category selected.
func NewBarChart(id, class string) *BarChart {
	c := &BarChart{
		Component: Component{id: id, class: class},
		showAxis:  true,
		showGrid:  true,
		legend:    true,
		barWidth:  3,
		barGap:    1,
		ticks:     5,
		selected:  -1,
		mode:      Relative,
		chCorner:  "└",
		chHLine:   "─",
		chVLine:   "│",
		chTickX:   "┬",
		chTickY:   "┤",
		chGrid:    "─",
		chSwatch:  "█",
	}
	c.SetFlag(FlagFocusable, true)
	OnKey(c, c.handleKey)
	OnMouse(c, c.handleMouse)
	return c
}

// ── Data setters ──────────────────────────────────────────────────────────────

// SetSeries replaces all series and redraws.
func (c *BarChart) SetSeries(s []BarSeries) {
	c.series = s
	c.Refresh()
}

// AddSeries appends a series and redraws.
func (c *BarChart) AddSeries(s BarSeries) {
	c.series = append(c.series, s)
	c.Refresh()
}

// SetCategories replaces the category labels and redraws.
func (c *BarChart) SetCategories(labels []string) {
	c.categories = labels
	if c.selected >= len(labels) {
		c.selected = -1
	}
	c.Refresh()
}

// Series returns the current series slice.
func (c *BarChart) Series() []BarSeries { return c.series }

// Categories returns the current category labels.
func (c *BarChart) Categories() []string { return c.categories }

// ── Display setters ───────────────────────────────────────────────────────────

// SetMode sets the scale mode (Relative or Absolute) and redraws.
func (c *BarChart) SetMode(m ScaleMode) { c.mode = m; c.Refresh() }

// SetMax sets the explicit maximum for Absolute mode.
func (c *BarChart) SetMax(v float64) { c.max = v }

// SetHorizontal switches bar orientation and redraws.
func (c *BarChart) SetHorizontal(v bool) { c.horizontal = v; c.Refresh() }

// SetShowAxis shows or hides the y-axis labels and rule.
func (c *BarChart) SetShowAxis(v bool) { c.showAxis = v; c.Refresh() }

// SetShowGrid shows or hides horizontal grid lines at y-axis ticks.
func (c *BarChart) SetShowGrid(v bool) { c.showGrid = v; c.Refresh() }

// SetShowValues shows or hides the total-value label above/beside each bar.
func (c *BarChart) SetShowValues(v bool) { c.showValues = v; c.Refresh() }

// SetLegend shows or hides the series legend row.
func (c *BarChart) SetLegend(v bool) { c.legend = v; c.Refresh() }

// SetBarWidth sets the column width per bar in vertical mode (minimum 1).
func (c *BarChart) SetBarWidth(w int) {
	if w < 1 {
		w = 1
	}
	c.barWidth = w
	c.Refresh()
}

// SetBarGap sets the empty columns between bars (minimum 0).
func (c *BarChart) SetBarGap(g int) {
	if g < 0 {
		g = 0
	}
	c.barGap = g
	c.Refresh()
}

// SetTicks sets the approximate number of y-axis ticks (minimum 2).
func (c *BarChart) SetTicks(n int) {
	if n < 2 {
		n = 2
	}
	c.ticks = n
	c.Refresh()
}

// ── Navigation ────────────────────────────────────────────────────────────────

// Select focuses a category by index, clamping to the valid range, and
// dispatches [EvtSelect]. No-op when already selected.
func (c *BarChart) Select(index int) {
	n := len(c.categories)
	if n == 0 {
		c.selected = -1
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= n {
		index = n - 1
	}
	if index == c.selected {
		return
	}
	c.selected = index
	c.Dispatch(c, EvtSelect, index)
	Redraw(c)
}

// Selected returns the focused category index, or -1 if none.
func (c *BarChart) Selected() int { return c.selected }

// ── Theme / Apply ─────────────────────────────────────────────────────────────

// Apply registers all bar-chart style selectors and reads theme strings.
func (c *BarChart) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("bar-chart"), "focused", "hovered", "disabled")
	theme.Apply(c, c.Selector("bar-chart/axis"))
	theme.Apply(c, c.Selector("bar-chart/grid"))
	theme.Apply(c, c.Selector("bar-chart/label"), "focused")
	theme.Apply(c, c.Selector("bar-chart/selection"))
	theme.Apply(c, c.Selector("bar-chart/value"))
	theme.Apply(c, c.Selector("bar-chart/legend"))
	for i := range 8 {
		theme.Apply(c, c.Selector(fmt.Sprintf("bar-chart/s%d", i)))
	}
	str := func(key, def string) string {
		if s := theme.String(key); s != "" {
			return s
		}
		return def
	}
	c.chCorner = str("bar-chart.corner", "└")
	c.chHLine = str("bar-chart.hline", "─")
	c.chVLine = str("bar-chart.vline", "│")
	c.chTickX = str("bar-chart.tick-x", "┬")
	c.chTickY = str("bar-chart.tick-y", "┤")
	c.chGrid = str("bar-chart.grid", "─")
	c.chSwatch = str("bar-chart.swatch", "█")
}

// ── Hint ──────────────────────────────────────────────────────────────────────

// Hint returns the preferred size.
// Vertical: natural width (axis + bars), height = 0 (fills parent).
// Horizontal: width = 0 (fills parent), natural height (one row per category).
func (c *BarChart) Hint() (int, int) {
	if c.hwidth != 0 || c.hheight != 0 {
		return c.hwidth, c.hheight
	}
	if c.horizontal {
		h := len(c.categories)
		if c.legend && c.hasLegendLabels() {
			h++
		}
		return 0, h
	}
	_, _, yAxisW := c.yAxisLayout()
	if !c.showAxis {
		yAxisW = 0
	}
	n := len(c.categories)
	if n == 0 {
		return yAxisW, 0
	}
	w := yAxisW + n*(c.barWidth+c.barGap) - c.barGap
	return w, 0
}

// ── Render ────────────────────────────────────────────────────────────────────

// Render draws the bar chart.
func (c *BarChart) Render(r *Renderer) {
	if c.Flag(FlagHidden) {
		return
	}
	c.Component.Render(r)
	if c.horizontal {
		c.renderHorizontal(r)
	} else {
		c.renderVertical(r)
	}
}

func (c *BarChart) renderVertical(r *Renderer) {
	cx, cy, cw, ch := c.Content()
	if cw < 2 || ch < 3 {
		return
	}

	legendH := 0
	if c.legend && c.hasLegendLabels() {
		legendH = 1
	}
	valueH := 0
	if c.showValues {
		valueH = 1
	}
	// Layout: [valueH] chart [1 baseline] [1 labels] [legendH]
	chartH := ch - valueH - 2 - legendH
	if chartH < 1 {
		return
	}
	chartY := cy + valueH
	baselineY := chartY + chartH
	labelY := baselineY + 1
	legendY := labelY + 1

	tickVals, tickLabels, yAxisW := c.yAxisLayout()
	if !c.showAxis {
		yAxisW = 0
	}
	chartW := cw - yAxisW
	if chartW < 1 {
		return
	}

	eMax := c.effectiveMax()
	chartBg := c.Style().Background()

	// ── Y-axis ────────────────────────────────────────────────────────────────
	if c.showAxis {
		axisS := c.Style("axis")
		gridS := c.Style("grid")

		// Draw the vertical rule (overwritten with ┤ at tick rows).
		r.Set(axisS.Foreground(), axisS.Background(), axisS.Font())
		for row := chartY; row < baselineY; row++ {
			r.Put(cx+yAxisW-1, row, c.chVLine)
		}

		for i, tv := range tickVals {
			var screenY int
			if tv <= 0 {
				screenY = baselineY // "0" sits on the baseline
			} else {
				n := int(math.Ceil(tv / eMax * float64(chartH)))
				screenY = chartY + chartH - n
				if screenY < chartY {
					screenY = chartY
				}
			}

			// Right-aligned tick label.
			labelW := yAxisW - 2
			if labelW > 0 {
				formatted := fmt.Sprintf("%*s", labelW, tickLabels[i])
				r.Set(axisS.Foreground(), axisS.Background(), axisS.Font())
				r.Text(cx, screenY, formatted, labelW)
				// Space between label and rule.
				r.Put(cx+yAxisW-2, screenY, " ")
			}

			if tv <= 0 {
				// Corner is drawn in the baseline section.
				continue
			}
			if screenY >= chartY && screenY < baselineY {
				r.Set(axisS.Foreground(), axisS.Background(), axisS.Font())
				r.Put(cx+yAxisW-1, screenY, c.chTickY)
			}

			// Horizontal grid line.
			if c.showGrid && screenY >= chartY && screenY < baselineY {
				r.Set(gridS.Foreground(), gridS.Background(), gridS.Font())
				for col := 0; col < chartW; col++ {
					r.Put(cx+yAxisW+col, screenY, c.chGrid)
				}
			}
		}
	}

	// ── Bars ──────────────────────────────────────────────────────────────────
	// totalSteps = chartH rows × 8 sub-pixel steps per row.
	totalSteps := float64(chartH * 8)

	for b := range len(c.categories) {
		barX := cx + yAxisW + b*(c.barWidth+c.barGap)
		if barX >= cx+cw {
			break
		}
		bw := c.barWidth
		if barX+bw > cx+cw {
			bw = cx + cw - barX
		}

		// Cumulative step boundaries for each series.
		boundaries := make([]float64, len(c.series)+1)
		for i, s := range c.series {
			val := 0.0
			if b < len(s.Values) {
				val = s.Values[b]
			}
			boundaries[i+1] = boundaries[i] + val/eMax*totalSteps
			if boundaries[i+1] > totalSteps {
				boundaries[i+1] = totalSteps
			}
		}

		// Render from top (row 0) to bottom (row chartH-1).
		for row := 0; row < chartH; row++ {
			screenRow := chartY + row
			rowFromBottom := chartH - 1 - row
			stepBot := float64(rowFromBottom * 8)
			stepTop := stepBot + 8.0

			// Find the highest series boundary within (stepBot, stepTop).
			boundaryIdx := -1
			boundaryStep := 0.0
			for i := range len(c.series) {
				bs := boundaries[i+1]
				if bs > stepBot && bs < stepTop {
					if boundaryIdx < 0 || bs > boundaryStep {
						boundaryIdx = i
						boundaryStep = bs
					}
				}
			}

			if boundaryIdx >= 0 {
				// Partial row: boundary between series boundaryIdx and boundaryIdx+1.
				stepsFromCellBottom := boundaryStep - stepBot
				idx := int(stepsFromCellBottom)
				if idx < 1 {
					idx = 1
				}
				sFg := c.Style(fmt.Sprintf("s%d", boundaryIdx%8)).Foreground()
				var bgColor string
				if boundaryIdx+1 < len(c.series) {
					bgColor = c.Style(fmt.Sprintf("s%d", (boundaryIdx+1)%8)).Foreground()
				} else {
					bgColor = chartBg
				}
				r.Set(sFg, bgColor, "")
				for col := 0; col < bw; col++ {
					r.Put(barX+col, screenRow, string(barChartBlocks[idx]))
				}
			} else {
				// Find the series that contains this cell (lowest i where boundaries[i+1] > stepBot).
				topSeries := -1
				for i := range len(c.series) {
					if boundaries[i+1] > stepBot {
						topSeries = i
						break
					}
				}
				if topSeries < 0 {
					continue // empty row
				}
				sFg := c.Style(fmt.Sprintf("s%d", topSeries%8)).Foreground()
				r.Set(sFg, chartBg, "")
				for col := 0; col < bw; col++ {
					r.Put(barX+col, screenRow, "█")
				}
			}
		}

		// ── Selection indicator ───────────────────────────────────────────────
		if c.Flag(FlagFocused) && b == c.selected {
			// Find the top row of this bar.
			total := boundaries[len(boundaries)-1]
			topSteps := total // steps from bottom
			barTopRow := chartY + chartH - int(math.Ceil(topSteps/8))
			if barTopRow < chartY {
				barTopRow = chartY
			}
			indicatorRow := barTopRow - 1
			if indicatorRow >= chartY {
				selS := c.Style("selection")
				r.Set(selS.Foreground(), selS.Background(), selS.Font())
				center := barX + bw/2
				r.Put(center, indicatorRow, "▲")
			}
		}

		// ── Value label ───────────────────────────────────────────────────────
		if c.showValues && valueH > 0 {
			total := 0.0
			for _, s := range c.series {
				if b < len(s.Values) {
					total += s.Values[b]
				}
			}
			label := fmt.Sprintf("%g", total)
			valS := c.Style("value")
			r.Set(valS.Foreground(), valS.Background(), valS.Font())
			r.Text(barX, cy, label, bw)
		}
	}

	// ── X-axis baseline ───────────────────────────────────────────────────────
	axisS := c.Style("axis")
	r.Set(axisS.Foreground(), axisS.Background(), axisS.Font())
	if c.showAxis {
		r.Put(cx+yAxisW-1, baselineY, c.chCorner)
	}
	for col := 0; col < chartW; col++ {
		r.Put(cx+yAxisW+col, baselineY, c.chHLine)
	}
	// Tick marks under bar centres.
	for b := range len(c.categories) {
		center := cx + yAxisW + b*(c.barWidth+c.barGap) + c.barWidth/2
		if center >= cx+cw {
			break
		}
		r.Put(center, baselineY, c.chTickX)
	}

	// ── Category labels ───────────────────────────────────────────────────────
	labelS := c.Style("label")
	selLabelS := c.Style("label:focused")
	for b := range len(c.categories) {
		center := cx + yAxisW + b*(c.barWidth+c.barGap) + c.barWidth/2
		if center >= cx+cw {
			break
		}
		label := ""
		if b < len(c.categories) {
			label = c.categories[b]
		}
		maxW := c.barWidth + c.barGap - 1
		if maxW < 1 {
			maxW = 1
		}
		runes := []rune(label)
		if len(runes) > maxW {
			label = string(runes[:maxW])
		}
		lw := utf8.RuneCountInString(label)
		labelX := center - lw/2

		focused := c.Flag(FlagFocused) && b == c.selected
		if focused {
			r.Set(selLabelS.Foreground(), selLabelS.Background(), selLabelS.Font())
		} else {
			r.Set(labelS.Foreground(), labelS.Background(), labelS.Font())
		}
		r.Text(labelX, labelY, label, lw)
	}

	// ── Legend ────────────────────────────────────────────────────────────────
	if c.legend && c.hasLegendLabels() && legendY < cy+ch {
		c.renderLegend(r, cx, legendY, cw)
	}
}

func (c *BarChart) renderHorizontal(r *Renderer) {
	cx, cy, cw, ch := c.Content()
	if cw < 4 || ch < 1 {
		return
	}

	// Label width: widest category + 1 space padding.
	labelW := 0
	for _, cat := range c.categories {
		if w := utf8.RuneCountInString(cat); w > labelW {
			labelW = w
		}
	}
	labelW++

	// Value width: width of the largest total string + 1 space.
	valueW := 0
	if c.showValues {
		eMax := c.effectiveMax()
		vl := len(fmt.Sprintf("%g", eMax))
		valueW = vl + 1
	}

	chartW := cw - labelW - valueW
	if chartW < 1 {
		return
	}

	eMax := c.effectiveMax()
	chartBg := c.Style().Background()
	labelS := c.Style("label")
	selLabelS := c.Style("label:focused")
	valS := c.Style("value")

	for b := range len(c.categories) {
		y := cy + b
		if y >= cy+ch {
			break
		}

		focused := c.Flag(FlagFocused) && b == c.selected

		// Category label.
		cat := ""
		if b < len(c.categories) {
			cat = c.categories[b]
		}
		if focused {
			r.Set(selLabelS.Foreground(), selLabelS.Background(), selLabelS.Font())
		} else {
			r.Set(labelS.Foreground(), labelS.Background(), labelS.Font())
		}
		r.Text(cx, y, cat, labelW)

		// Bar segments (left to right = series 0 first).
		x := cx + labelW
		cum := 0.0
		for i, s := range c.series {
			val := 0.0
			if b < len(s.Values) {
				val = s.Values[b]
			}
			prevCum := cum
			cum += val / eMax * float64(chartW)
			if cum > float64(chartW) {
				cum = float64(chartW)
			}
			w := int(math.Round(cum)) - int(math.Round(prevCum))
			if w <= 0 {
				continue
			}
			sS := c.Style(fmt.Sprintf("s%d", i%8))
			r.Set(sS.Foreground(), chartBg, "")
			r.Fill(x+int(math.Round(prevCum)), y, w, 1, "█")
		}
		// Clear remaining bar area.
		barEnd := x + int(math.Round(cum))
		if barEnd < x+chartW {
			r.Set("", chartBg, "")
			r.Fill(barEnd, y, x+chartW-barEnd, 1, " ")
		}

		// Value label.
		if c.showValues {
			total := 0.0
			for _, s := range c.series {
				if b < len(s.Values) {
					total += s.Values[b]
				}
			}
			r.Set(valS.Foreground(), valS.Background(), valS.Font())
			r.Text(cx+labelW+chartW, y, " "+fmt.Sprintf("%g", total), valueW)
		}
	}

	// Legend.
	legendY := cy + len(c.categories)
	if c.legend && c.hasLegendLabels() && legendY < cy+ch {
		c.renderLegend(r, cx, legendY, cw)
	}
}

func (c *BarChart) renderLegend(r *Renderer, cx, legendY, cw int) {
	legS := c.Style("legend")
	x := cx
	for i, s := range c.series {
		if s.Label == "" {
			continue
		}
		if x >= cx+cw {
			break
		}
		sS := c.Style(fmt.Sprintf("s%d", i%8))
		r.Set(sS.Foreground(), legS.Background(), "")
		r.Put(x, legendY, c.chSwatch)
		r.Set(legS.Foreground(), legS.Background(), legS.Font())
		entry := " " + s.Label + "   "
		ew := utf8.RuneCountInString(entry)
		r.Text(x+1, legendY, entry, min(ew, cx+cw-x-1))
		x += 1 + ew
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (c *BarChart) effectiveMax() float64 {
	if c.mode == Absolute && c.max > 0 {
		return c.max
	}
	maxTotal := 0.0
	for b := range len(c.categories) {
		total := 0.0
		for _, s := range c.series {
			if b < len(s.Values) {
				total += s.Values[b]
			}
		}
		if total > maxTotal {
			maxTotal = total
		}
	}
	if maxTotal == 0 {
		return 1
	}
	return maxTotal
}

// niceCeil rounds v up to the nearest value of the form k × 10^n,
// where k ∈ {1, 2, 5}.
func niceCeil(v float64) float64 {
	if v <= 0 {
		return 1
	}
	exp := math.Floor(math.Log10(v))
	base := math.Pow(10, exp)
	for _, k := range []float64{1, 2, 5, 10} {
		if k*base >= v {
			return k * base
		}
	}
	return base * 10
}

// yAxisLayout computes tick values, their string labels, and the y-axis width.
func (c *BarChart) yAxisLayout() (tickVals []float64, tickLabels []string, yAxisW int) {
	eMax := c.effectiveMax()
	step := niceCeil(eMax / float64(c.ticks))
	for v := 0.0; v <= eMax+step*0.5; v += step {
		tickVals = append(tickVals, v)
		tickLabels = append(tickLabels, fmt.Sprintf("%g", v))
	}
	maxLW := 0
	for _, l := range tickLabels {
		if w := utf8.RuneCountInString(l); w > maxLW {
			maxLW = w
		}
	}
	yAxisW = maxLW + 2 // label + space + │
	return
}

func (c *BarChart) hasLegendLabels() bool {
	for _, s := range c.series {
		if s.Label != "" {
			return true
		}
	}
	return false
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func (c *BarChart) handleKey(evt *tcell.EventKey) bool {
	n := len(c.categories)
	if n == 0 {
		return false
	}
	switch evt.Key() {
	case tcell.KeyLeft:
		if !c.horizontal {
			c.move(-1)
			return true
		}
	case tcell.KeyRight:
		if !c.horizontal {
			c.move(+1)
			return true
		}
	case tcell.KeyUp:
		if c.horizontal {
			c.move(-1)
			return true
		}
	case tcell.KeyDown:
		if c.horizontal {
			c.move(+1)
			return true
		}
	case tcell.KeyHome:
		c.Select(0)
		return true
	case tcell.KeyEnd:
		c.Select(n - 1)
		return true
	case tcell.KeyEnter:
		if c.selected >= 0 {
			c.Dispatch(c, EvtActivate, c.selected)
		}
		return true
	}
	return false
}

func (c *BarChart) move(delta int) {
	n := len(c.categories)
	if n == 0 {
		return
	}
	next := c.selected + delta
	if c.selected < 0 {
		if delta > 0 {
			next = 0
		} else {
			next = n - 1
		}
	}
	c.Select(next)
}

// ── Mouse ─────────────────────────────────────────────────────────────────────

func (c *BarChart) handleMouse(evt *tcell.EventMouse) bool {
	if evt.Buttons() != tcell.Button1 {
		return false
	}
	mx, my := evt.Position()
	cx, cy, cw, ch := c.Content()
	if mx < cx || mx >= cx+cw || my < cy || my >= cy+ch {
		return false
	}

	var b int
	if c.horizontal {
		b = my - cy
	} else {
		_, _, yAxisW := c.yAxisLayout()
		if !c.showAxis {
			yAxisW = 0
		}
		if mx < cx+yAxisW {
			return false
		}
		b = (mx - cx - yAxisW) / (c.barWidth + c.barGap)
	}

	if b < 0 || b >= len(c.categories) {
		return false
	}
	if b == c.selected {
		c.Dispatch(c, EvtActivate, b)
	} else {
		c.Select(b)
	}
	return true
}

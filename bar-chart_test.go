package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestBarChart_Defaults(t *testing.T) {
	bc := NewBarChart("bc", "")
	if bc.barWidth != 3 {
		t.Errorf("barWidth = %d; want 3", bc.barWidth)
	}
	if bc.barGap != 1 {
		t.Errorf("barGap = %d; want 1", bc.barGap)
	}
	if bc.ticks != 5 {
		t.Errorf("ticks = %d; want 5", bc.ticks)
	}
	if bc.selected != -1 {
		t.Errorf("selected = %d; want -1", bc.selected)
	}
	if bc.absolute {
		t.Error("absolute should be false by default")
	}
	if !bc.showAxis {
		t.Error("showAxis should be true by default")
	}
	if !bc.showGrid {
		t.Error("showGrid should be true by default")
	}
	if !bc.legend {
		t.Error("legend should be true by default")
	}
	if !bc.Flag(FlagFocusable) {
		t.Error("expected FlagFocusable to be set")
	}
}

// ── Minimum-clamp setters ─────────────────────────────────────────────────────

func TestBarChart_SetBarWidth_Clamps(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetBarWidth(0)
	if bc.barWidth != 1 {
		t.Errorf("barWidth = %d after SetBarWidth(0); want 1", bc.barWidth)
	}
	bc.SetBarWidth(-5)
	if bc.barWidth != 1 {
		t.Errorf("barWidth = %d after SetBarWidth(-5); want 1", bc.barWidth)
	}
	bc.SetBarWidth(4)
	if bc.barWidth != 4 {
		t.Errorf("barWidth = %d after SetBarWidth(4); want 4", bc.barWidth)
	}
}

func TestBarChart_SetBarGap_Clamps(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetBarGap(-1)
	if bc.barGap != 0 {
		t.Errorf("barGap = %d after SetBarGap(-1); want 0", bc.barGap)
	}
	bc.SetBarGap(2)
	if bc.barGap != 2 {
		t.Errorf("barGap = %d after SetBarGap(2); want 2", bc.barGap)
	}
}

func TestBarChart_SetTicks_Clamps(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetTicks(1)
	if bc.ticks != 2 {
		t.Errorf("ticks = %d after SetTicks(1); want 2", bc.ticks)
	}
	bc.SetTicks(0)
	if bc.ticks != 2 {
		t.Errorf("ticks = %d after SetTicks(0); want 2", bc.ticks)
	}
	bc.SetTicks(10)
	if bc.ticks != 10 {
		t.Errorf("ticks = %d after SetTicks(10); want 10", bc.ticks)
	}
}

// ── Data methods ──────────────────────────────────────────────────────────────

func TestBarChart_SetCategories_GetCategories(t *testing.T) {
	bc := NewBarChart("bc", "")
	cats := []string{"A", "B", "C"}
	bc.SetCategories(cats)
	got := bc.Categories()
	if len(got) != 3 || got[0] != "A" || got[2] != "C" {
		t.Errorf("Categories() = %v; want %v", got, cats)
	}
}

func TestBarChart_SetCategories_ClampsSelected(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A", "B", "C"})
	bc.selected = 2
	bc.SetCategories([]string{"X"})
	if bc.selected != -1 {
		t.Errorf("selected = %d after shrink; want -1 (clamped out of range)", bc.selected)
	}
}

func TestBarChart_SetSeries_GetSeries(t *testing.T) {
	bc := NewBarChart("bc", "")
	s := []BarSeries{
		{Label: "Rev", Values: []float64{10, 20}},
		{Label: "Cost", Values: []float64{5, 8}},
	}
	bc.SetSeries(s)
	got := bc.Series()
	if len(got) != 2 || got[0].Label != "Rev" || got[1].Label != "Cost" {
		t.Errorf("Series() = %v; want %v", got, s)
	}
}

func TestBarChart_AddSeries(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.AddSeries(BarSeries{Label: "A", Values: []float64{1}})
	bc.AddSeries(BarSeries{Label: "B", Values: []float64{2}})
	if len(bc.Series()) != 2 {
		t.Errorf("len(Series()) = %d; want 2", len(bc.Series()))
	}
}

// ── niceCeil ──────────────────────────────────────────────────────────────────

func TestNiceCeil(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{0, 1},
		{-1, 1},
		{0.5, 0.5}, // 5×10^-1 is already "nice"
		{1, 1},
		{1.5, 2},
		{2, 2},
		{2.5, 5},
		{5, 5},
		{5.5, 10},
		{10, 10},
		{11, 20},
		{50, 50},
		{51, 100},
		{99, 100},
		{100, 100},
		{101, 200},
		{1000, 1000},
	}
	for _, tc := range cases {
		got := niceCeil(tc.in)
		if got != tc.want {
			t.Errorf("niceCeil(%v) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

// ── effectiveMax ──────────────────────────────────────────────────────────────

func TestBarChart_EffectiveMax_Relative(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A", "B", "C"})
	bc.SetSeries([]BarSeries{
		{Values: []float64{10, 30, 20}},
		{Values: []float64{5, 15, 10}},
	})
	// Totals: 15, 45, 30 → max = 45
	got := bc.effectiveMax()
	if got != 45 {
		t.Errorf("effectiveMax() = %v; want 45", got)
	}
}

func TestBarChart_EffectiveMax_Absolute(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetAbsolute(true)
	bc.SetMax(200)
	bc.SetCategories([]string{"A"})
	bc.SetSeries([]BarSeries{{Values: []float64{50}}})
	got := bc.effectiveMax()
	if got != 200 {
		t.Errorf("effectiveMax() = %v; want 200 (explicit max)", got)
	}
}

func TestBarChart_EffectiveMax_AllZero(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A"})
	bc.SetSeries([]BarSeries{{Values: []float64{0}}})
	got := bc.effectiveMax()
	if got != 1 {
		t.Errorf("effectiveMax() = %v; want 1 (fallback for all-zero)", got)
	}
}

// ── yAxisLayout ───────────────────────────────────────────────────────────────

func TestBarChart_YAxisLayout_FirstTickZero(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A"})
	bc.SetSeries([]BarSeries{{Values: []float64{100}}})
	tickVals, tickLabels, _ := bc.yAxisLayout()
	if len(tickVals) < 2 {
		t.Fatalf("expected at least 2 ticks, got %d", len(tickVals))
	}
	if tickVals[0] != 0 {
		t.Errorf("tickVals[0] = %v; want 0", tickVals[0])
	}
	if len(tickVals) != len(tickLabels) {
		t.Errorf("tickVals len %d != tickLabels len %d", len(tickVals), len(tickLabels))
	}
}

func TestBarChart_YAxisLayout_Width(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A"})
	bc.SetSeries([]BarSeries{{Values: []float64{1000}}})
	// Labels will be like "0", "200", ..., "1000" → max label width 4 → yAxisW = 4+2 = 6
	_, _, yAxisW := bc.yAxisLayout()
	if yAxisW < 3 {
		t.Errorf("yAxisW = %d; want at least 3 (label + space + │)", yAxisW)
	}
}

// ── Select / Selected ─────────────────────────────────────────────────────────

func TestBarChart_Select_Clamps(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A", "B", "C"})
	bc.Select(10)
	if bc.Selected() != 2 {
		t.Errorf("Selected() = %d after Select(10); want 2", bc.Selected())
	}
	bc.Select(-5)
	if bc.Selected() != 0 {
		t.Errorf("Selected() = %d after Select(-5); want 0", bc.Selected())
	}
}

func TestBarChart_Select_NoOpOnEmpty(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.Select(0)
	if bc.Selected() != -1 {
		t.Errorf("Selected() = %d on empty chart; want -1", bc.Selected())
	}
}

func TestBarChart_Select_DispatchesEvtSelect(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A", "B", "C"})
	fired := -1
	bc.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(int)
		return true
	})
	bc.Select(2)
	if fired != 2 {
		t.Errorf("EvtSelect data = %d; want 2", fired)
	}
}

func TestBarChart_Select_NoDispatchIfSame(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A", "B", "C"})
	bc.Select(1)
	count := 0
	bc.On(EvtSelect, func(_ Widget, _ Event, _ ...any) bool {
		count++
		return true
	})
	bc.Select(1) // same index
	if count != 0 {
		t.Errorf("EvtSelect fired %d times on no-op Select; want 0", count)
	}
}

// ── Hint ──────────────────────────────────────────────────────────────────────

func TestBarChart_Hint_Vertical_NoCategories(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetShowAxis(false)
	w, h := bc.Hint()
	if w != 0 {
		t.Errorf("Hint() width = %d for empty chart; want 0", w)
	}
	_ = h
}

func TestBarChart_Hint_Vertical_WithCategories(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetShowAxis(false)
	bc.SetBarWidth(3)
	bc.SetBarGap(1)
	bc.SetCategories([]string{"A", "B"})
	// w = 0 + 2*(3+1) - 1 = 7
	w, _ := bc.Hint()
	if w != 7 {
		t.Errorf("Hint() width = %d; want 7", w)
	}
}

func TestBarChart_Hint_Horizontal(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetHorizontal(true)
	bc.SetCategories([]string{"A", "B", "C"})
	// No legend labels → h = 3
	w, h := bc.Hint()
	if w != 0 {
		t.Errorf("Hint() width = %d for horizontal; want 0 (fill parent)", w)
	}
	if h != 3 {
		t.Errorf("Hint() height = %d; want 3 (one row per category)", h)
	}
}

func TestBarChart_Hint_Horizontal_WithLegend(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetHorizontal(true)
	bc.SetCategories([]string{"A", "B"})
	bc.SetSeries([]BarSeries{{Label: "Rev", Values: []float64{1, 2}}})
	// legend=true and has labels → h = 2 + 1 = 3
	_, h := bc.Hint()
	if h != 3 {
		t.Errorf("Hint() height = %d with legend; want 3", h)
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func TestBarChart_Keyboard_Empty_Ignored(t *testing.T) {
	bc := NewBarChart("bc", "")
	handled := bc.handleKey(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone))
	if handled {
		t.Error("handleKey on empty chart should return false")
	}
}

func TestBarChart_Keyboard_Enter_DispatchesActivate(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A", "B"})
	bc.Select(1)
	fired := -1
	bc.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(int)
		return true
	})
	bc.handleKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	if fired != 1 {
		t.Errorf("EvtActivate data = %d after Enter; want 1", fired)
	}
}

func TestBarChart_Keyboard_Enter_NoSelectionNoFire(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A"})
	// selected = -1 (default)
	fired := false
	bc.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	bc.handleKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	if fired {
		t.Error("EvtActivate should not fire when selected == -1")
	}
}

func TestBarChart_Keyboard_LeftRight_Navigate(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A", "B", "C"})
	// From -1, Right should move to 0
	bc.handleKey(tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone))
	if bc.Selected() != 0 {
		t.Errorf("Selected() = %d after Right from -1; want 0", bc.Selected())
	}
	bc.handleKey(tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone))
	if bc.Selected() != 1 {
		t.Errorf("Selected() = %d after second Right; want 1", bc.Selected())
	}
	bc.handleKey(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone))
	if bc.Selected() != 0 {
		t.Errorf("Selected() = %d after Left; want 0", bc.Selected())
	}
}

func TestBarChart_Keyboard_HomeEnd(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetCategories([]string{"A", "B", "C"})
	bc.handleKey(tcell.NewEventKey(tcell.KeyEnd, "", tcell.ModNone))
	if bc.Selected() != 2 {
		t.Errorf("Selected() = %d after End; want 2", bc.Selected())
	}
	bc.handleKey(tcell.NewEventKey(tcell.KeyHome, "", tcell.ModNone))
	if bc.Selected() != 0 {
		t.Errorf("Selected() = %d after Home; want 0", bc.Selected())
	}
}

func TestBarChart_Keyboard_Horizontal_UpDown(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetHorizontal(true)
	bc.SetCategories([]string{"A", "B", "C"})
	bc.handleKey(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone))
	if bc.Selected() != 0 {
		t.Errorf("Selected() = %d after Down; want 0", bc.Selected())
	}
	bc.handleKey(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone))
	if bc.Selected() != 1 {
		t.Errorf("Selected() = %d after second Down; want 1", bc.Selected())
	}
}

// ── Mouse ──────────────────────────────────────────────────────────────────────

func TestBarChart_Mouse_VerticalSelect(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetShowAxis(false)
	bc.SetBarWidth(3)
	bc.SetBarGap(1)
	bc.SetCategories([]string{"A", "B", "C"})
	bc.SetBounds(0, 0, 20, 10)
	// Category 0 occupies cols 0-3, category 1 cols 4-7, category 2 cols 8-11
	// Click at col 0 → b = 0/(3+1) = 0
	ev := tcell.NewEventMouse(0, 5, tcell.Button1, tcell.ModNone)
	bc.handleMouse(ev)
	if bc.Selected() != 0 {
		t.Errorf("Selected() = %d after click at col 0; want 0", bc.Selected())
	}
	// Click at col 4 → b = 4/(3+1) = 1
	ev = tcell.NewEventMouse(4, 5, tcell.Button1, tcell.ModNone)
	bc.handleMouse(ev)
	if bc.Selected() != 1 {
		t.Errorf("Selected() = %d after click at col 4; want 1", bc.Selected())
	}
}

func TestBarChart_Mouse_ActivateOnSameCategory(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetShowAxis(false)
	bc.SetBarWidth(3)
	bc.SetBarGap(1)
	bc.SetCategories([]string{"A", "B"})
	bc.SetBounds(0, 0, 20, 10)
	fired := -1
	bc.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(int)
		return true
	})
	ev := tcell.NewEventMouse(0, 5, tcell.Button1, tcell.ModNone)
	bc.handleMouse(ev) // select category 0
	bc.handleMouse(ev) // click again → activate
	if fired != 0 {
		t.Errorf("EvtActivate data = %d after double-click; want 0", fired)
	}
}

func TestBarChart_Mouse_NonButton1Ignored(t *testing.T) {
	bc := NewBarChart("bc", "")
	bc.SetShowAxis(false)
	bc.SetCategories([]string{"A"})
	bc.SetBounds(0, 0, 20, 10)
	handled := bc.handleMouse(tcell.NewEventMouse(0, 5, tcell.Button2, tcell.ModNone))
	if handled {
		t.Error("handleMouse with Button2 should return false")
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func TestBarChart_Render_Vertical_DrawsBlocks(t *testing.T) {
	theme := NewTheme()
	theme.SetStyles(
		NewStyle("bar-chart/s0").WithColors("s0fg", ""),
	)
	bc := NewBarChart("bc", "")
	bc.Apply(theme)
	bc.SetShowAxis(false)
	bc.SetLegend(false)
	bc.SetCategories([]string{"A"})
	bc.SetSeries([]BarSeries{{Values: []float64{100}}})
	// h=5: chartH = 5-0-2 = 3, chartY=0, baselineY=3, labelY=4
	cs2 := newCellScreen()
	r2 := NewRenderer(cs2, NewTheme())
	bc.SetBounds(0, 0, 5, 5)
	bc.Render(r2)

	// Expect "█" at (0,0), (0,1), (0,2) (all chart rows filled for 100% bar)
	hasBlock := false
	for row := 0; row < 3; row++ {
		if cs2.Get(0, row) == "█" {
			hasBlock = true
			break
		}
	}
	if !hasBlock {
		t.Error("expected at least one '█' block in chart area")
	}
}

func newBarChartRenderer() (*cellScreen, *Renderer) {
	cs := newCellScreen()
	return cs, NewRenderer(cs, NewTheme())
}

func TestBarChart_Render_Vertical_StackingOrder(t *testing.T) {
	theme := NewTheme()
	theme.SetStyles(
		NewStyle("bar-chart/s0").WithColors("s0fg", ""),
		NewStyle("bar-chart/s1").WithColors("s1fg", ""),
	)
	bc := NewBarChart("bc", "")
	bc.Apply(theme)
	bc.SetShowAxis(false)
	bc.SetLegend(false)
	bc.SetCategories([]string{"A"})
	// Equal halves: s0=50, s1=50 → bottom half s0, top half s1
	bc.SetSeries([]BarSeries{
		{Values: []float64{50}},
		{Values: []float64{50}},
	})
	// h=5: chartH=3, rows 0-2
	// With totalSteps=24, boundaries=[0, 12, 24]
	// row=2 (stepBot=0): topSeries=0 (s0)
	// row=0 (stepBot=16): topSeries=1 (s1)
	cs, r := newBarChartRenderer()
	bc.SetBounds(0, 0, 5, 5)
	bc.Render(r)

	// Bottom row (row=2) should be s0 color
	if cs.fgs[[2]int{0, 2}] != "s0fg" {
		t.Errorf("bottom row fg = %q; want s0fg", cs.fgs[[2]int{0, 2}])
	}
	// Top row (row=0) should be s1 color
	if cs.fgs[[2]int{0, 0}] != "s1fg" {
		t.Errorf("top row fg = %q; want s1fg", cs.fgs[[2]int{0, 0}])
	}
}

func TestBarChart_Render_Horizontal_DrawsBlocks(t *testing.T) {
	theme := NewTheme()
	theme.SetStyles(
		NewStyle("bar-chart/s0").WithColors("s0fg", ""),
	)
	bc := NewBarChart("bc", "")
	bc.Apply(theme)
	bc.SetHorizontal(true)
	bc.SetLegend(false)
	bc.SetCategories([]string{"A", "B"})
	bc.SetSeries([]BarSeries{{Values: []float64{100, 50}}})

	cs, r := newBarChartRenderer()
	bc.SetBounds(0, 0, 20, 3)
	bc.Render(r)

	hasBlock := false
	for col := 1; col < 20; col++ {
		if cs.Get(col, 0) == "█" {
			hasBlock = true
			break
		}
	}
	if !hasBlock {
		t.Error("expected at least one '█' block in horizontal bar chart")
	}
}

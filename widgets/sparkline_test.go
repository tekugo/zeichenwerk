package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ---- test helpers ----------------------------------------------------------

func newSparklineRenderer() (*TestScreen, *Renderer) {
	cs := NewTestScreen()
	return cs, NewRenderer(cs, NewTheme())
}

// renderSparkline sets bounds and renders, returning the recording screen.
func renderSparkline(s *Sparkline, w, h int) *TestScreen {
	cs, r := newSparklineRenderer()
	s.SetBounds(0, 0, w, h)
	s.Render(r)
	return cs
}

// colChars returns the characters rendered in a single column from top to
// bottom (row 0 = top).
func colChars(cs *TestScreen, col, h int) []string {
	out := make([]string, h)
	for row := 0; row < h; row++ {
		ch := cs.Get(col, row)
		if ch == "" {
			ch = " "
		}
		out[row] = ch
	}
	return out
}

// ---- constructor -----------------------------------------------------------

func TestSparkline_Defaults(t *testing.T) {
	s := NewSparkline("s", "")
	if s.absolute {
		t.Errorf("absolute = true; want false (relative by default)")
	}
	if s.threshold != 0 {
		t.Errorf("threshold = %v; want 0", s.threshold)
	}
	if s.provider != nil {
		t.Errorf("provider = %v; want nil", s.provider)
	}
}

// ---- Hint ------------------------------------------------------------------

func TestSparkline_Hint_Default(t *testing.T) {
	s := NewSparkline("s", "")
	w, h := s.Hint()
	if w != 1 || h != 1 {
		t.Errorf("Hint() = (%d, %d); want (1, 1) for empty sparkline", w, h)
	}
}

func TestSparkline_Hint_FollowsProvider(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetProvider(FloatSlice{1, 2, 3, 4, 5})
	w, h := s.Hint()
	if w != 5 || h != 1 {
		t.Errorf("Hint() = (%d, %d); want (5, 1)", w, h)
	}
}

func TestSparkline_Hint_ExplicitOverrides(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetHint(20, 3)
	w, h := s.Hint()
	if w != 20 || h != 3 {
		t.Errorf("Hint() = (%d, %d); want (20, 3)", w, h)
	}
}

// ---- Relative scale, height 1 ----------------------------------------------

func TestSparkline_Relative_H1_MaxIsFullBlock(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetProvider(FloatSlice{0, 5, 10})
	cs := renderSparkline(s, 3, 1)
	ch := cs.Get(2, 0) // rightmost = newest = max
	if ch != "█" {
		t.Errorf("max value cell = %q; want █", ch)
	}
}

func TestSparkline_Relative_H1_MinIsLowestBlock(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetProvider(FloatSlice{0, 5, 10})
	cs := renderSparkline(s, 3, 1)
	ch := cs.Get(0, 0) // leftmost = oldest = min
	if ch != "▁" {
		t.Errorf("min value cell = %q; want ▁", ch)
	}
}

func TestSparkline_Relative_H1_AllEqual_MidRange(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetProvider(FloatSlice{5, 5, 5})
	cs := renderSparkline(s, 3, 1)
	// level=0.5, h=1, totalSteps=8, step=int(0.5*7+0.5)=4 → blocks[4]='▅'
	for col := 0; col < 3; col++ {
		ch := cs.Get(col, 0)
		if ch != "▅" {
			t.Errorf("col %d = %q; want ▅ (all-equal mid-range)", col, ch)
		}
	}
}

// ---- Absolute scale, height 1 ----------------------------------------------

func TestSparkline_Absolute_H1_AtMin_IsLowestBlock(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(10)
	s.SetProvider(FloatSlice{0})
	cs := renderSparkline(s, 1, 1)
	ch := cs.Get(0, 0)
	if ch != "▁" {
		t.Errorf("value at Min = %q; want ▁", ch)
	}
}

func TestSparkline_Absolute_H1_AtMax_IsFullBlock(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(10)
	s.SetProvider(FloatSlice{10})
	cs := renderSparkline(s, 1, 1)
	ch := cs.Get(0, 0)
	if ch != "█" {
		t.Errorf("value at Max = %q; want █", ch)
	}
}

func TestSparkline_Absolute_H1_BelowMin_Clamps(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(5)
	s.SetMax(10)
	s.SetProvider(FloatSlice{0}) // below min
	cs := renderSparkline(s, 1, 1)
	ch := cs.Get(0, 0)
	if ch != "▁" {
		t.Errorf("value below Min = %q; want ▁ (clamped)", ch)
	}
}

func TestSparkline_Absolute_H1_AboveMax_Clamps(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(10)
	s.SetProvider(FloatSlice{20}) // above max
	cs := renderSparkline(s, 1, 1)
	ch := cs.Get(0, 0)
	if ch != "█" {
		t.Errorf("value above Max = %q; want █ (clamped)", ch)
	}
}

// ---- Height 1 matches original 8-level formula -----------------------------

func TestSparkline_H1_MatchesOriginalFormula(t *testing.T) {
	// For h=1, step = int(level*7+0.5) which is exactly int(level*(1*8-1)+0.5).
	// Verify by checking a mid-value.
	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	cases := []struct {
		value float64
		want  rune
	}{
		{0.0 / 7, '▁'}, // step=0 → blocks[0]
		{1.0 / 7, '▂'}, // step=1 → blocks[1]
		{3.0 / 7, '▄'}, // step=3 → blocks[3]
		{7.0 / 7, '█'}, // step=7 → blocks[7]
	}
	for _, tc := range cases {
		s.SetProvider(FloatSlice{tc.value})
		cs := renderSparkline(s, 1, 1)
		got := []rune(cs.Get(0, 0))[0]
		if got != tc.want {
			t.Errorf("value=%.4f: got %q; want %q", tc.value, got, tc.want)
		}
	}
}

// ---- Multi-row rendering ---------------------------------------------------

func TestSparkline_H2_LevelZero_BottomLowest_TopSpace(t *testing.T) {
	// level=0 → step=0, fullRows=0, partial=0
	// bottom (rowFromBottom=0): blocks[0]='▁'
	// top    (rowFromBottom=1): space
	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	s.SetProvider(FloatSlice{0})
	cs := renderSparkline(s, 1, 2)
	top := cs.Get(0,  0)
	bot := cs.Get(0,  1)
	if top != " " {
		t.Errorf("top row = %q; want space", top)
	}
	if bot != "▁" {
		t.Errorf("bottom row = %q; want ▁", bot)
	}
}

func TestSparkline_H2_LevelOne_BothFullBlock(t *testing.T) {
	// level=1 → step=15, fullRows=1, partial=7
	// bottom (rowFromBottom=0): < fullRows(1) → '█'
	// top    (rowFromBottom=1): == fullRows(1) → blocks[7]='█'
	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	s.SetProvider(FloatSlice{1})
	cs := renderSparkline(s, 1, 2)
	top := cs.Get(0,  0)
	bot := cs.Get(0,  1)
	if top != "█" {
		t.Errorf("top row = %q; want █", top)
	}
	if bot != "█" {
		t.Errorf("bottom row = %q; want █", bot)
	}
}

func TestSparkline_H4_LevelOne_AllFullBlock(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	s.SetProvider(FloatSlice{1})
	cs := renderSparkline(s, 1, 4)
	for row := 0; row < 4; row++ {
		ch := cs.Get(0,  row)
		if ch != "█" {
			t.Errorf("row %d = %q; want █", row, ch)
		}
	}
}

// ---- Series width handling -------------------------------------------------

func TestSparkline_SeriesLongerThanWidth_ShowsRightmost(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	// 5 values, render width 3: only last 3 (newest) should appear.
	// FloatSlice{0,0,0,0,1}: oldest=0 at index 0, newest=1 at index 4.
	s.SetProvider(FloatSlice{0, 0, 0, 0, 1})
	cs := renderSparkline(s, 3, 1)
	// col 2 (rightmost) → newest value 1 → '█'
	if cs.Get(2,  0) != "█" {
		t.Errorf("rightmost col = %q; want █ (most recent value)", cs.Get(2,  0))
	}
	// col 0 → value 0 → '▁'
	if cs.Get(0,  0) != "▁" {
		t.Errorf("col 0 = %q; want ▁", cs.Get(0,  0))
	}
}

func TestSparkline_SeriesShorterThanWidth_PadsLeft(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetProvider(FloatSlice{5})
	cs := renderSparkline(s, 4, 1)
	// cols 0,1,2 have no data → blank; col 3 has data
	for col := 0; col < 3; col++ {
		ch := cs.Get(col,  0)
		if ch != " " {
			t.Errorf("blank col %d = %q; want space", col, ch)
		}
	}
}

func TestSparkline_SeriesShorterThanWidth_PadsAllRows(t *testing.T) {
	s := NewSparkline("s", "")
	s.SetProvider(FloatSlice{5})
	cs := renderSparkline(s, 3, 2) // 2 rows
	// cols 0,1 have no data → all rows blank
	for col := 0; col < 2; col++ {
		for row := 0; row < 2; row++ {
			ch := cs.Get(col,  row)
			if ch != " " {
				t.Errorf("blank col %d row %d = %q; want space", col, row, ch)
			}
		}
	}
}

// ---- Threshold / dual-colour -----------------------------------------------

func TestSparkline_Threshold_HighStyleApplied(t *testing.T) {
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())

	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(10)
	s.SetThreshold(7)

	// Manually set style colours so we can distinguish them.
	s.SetStyle("", NewStyle("").WithColors("low-fg", "low-bg"))
	s.SetStyle("high", NewStyle("high").WithColors("high-fg", "high-bg"))

	// FloatSlice{3,9}: oldest=3 (col 0), newest=9 (col 1).
	// 3 < threshold, 9 >= threshold.
	s.SetProvider(FloatSlice{3, 9})
	s.SetBounds(0, 0, 2, 1)
	s.Render(r)

	lowFG := cs.Fg(0,  0)
	highFG := cs.Fg(1,  0)
	if lowFG != "low-fg" {
		t.Errorf("below-threshold fg = %q; want low-fg", lowFG)
	}
	if highFG != "high-fg" {
		t.Errorf("above-threshold fg = %q; want high-fg", highFG)
	}
}

func TestSparkline_Threshold_AllRowsGetColumnStyle(t *testing.T) {
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())

	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(10)
	s.SetThreshold(5)
	s.SetStyle("", NewStyle("").WithColors("low-fg", "low-bg"))
	s.SetStyle("high", NewStyle("high").WithColors("high-fg", "high-bg"))

	s.SetProvider(FloatSlice{8}) // above threshold
	s.SetBounds(0, 0, 1, 3)
	s.Render(r)

	for row := 0; row < 3; row++ {
		fg := cs.Fg(0,  row)
		if fg != "high-fg" {
			t.Errorf("row %d fg = %q; want high-fg (all rows share column style)", row, fg)
		}
	}
}

func TestSparkline_Gradient_BelowThreshold_UsesBaseColor(t *testing.T) {
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())

	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	s.SetThreshold(0.5)
	s.SetGradient(true)
	s.SetStyle("", NewStyle("").WithColors("#000000", "bg"))
	s.SetStyle("high", NewStyle("high").WithColors("#ffffff", "bg"))

	// value = 0 → level = 0 → below threshold → t = 0 → base color
	s.SetProvider(FloatSlice{0})
	s.SetBounds(0, 0, 1, 1)
	s.Render(r)

	fg := cs.Fg(0,  0)
	if fg != "#000000" {
		t.Errorf("value below threshold: fg = %q; want #000000 (base color)", fg)
	}
}

func TestSparkline_Gradient_AtMax_UsesHighColor(t *testing.T) {
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())

	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	s.SetThreshold(0.5)
	s.SetGradient(true)
	s.SetStyle("", NewStyle("").WithColors("#000000", "bg"))
	s.SetStyle("high", NewStyle("high").WithColors("#ffffff", "bg"))

	// value = 1 → level = 1 → t = 1 → high color
	s.SetProvider(FloatSlice{1})
	s.SetBounds(0, 0, 1, 1)
	s.Render(r)

	fg := cs.Fg(0,  0)
	if fg != "#ffffff" {
		t.Errorf("value at max: fg = %q; want #ffffff (high color)", fg)
	}
}

func TestSparkline_Gradient_Midpoint_IsInterpolated(t *testing.T) {
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())

	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	// threshold=0.5, value=0.75 → midpoint of the above-threshold range [0.5, 1].
	// threshLevel = 0.5, level = 0.75, t = (0.75-0.5)/(1-0.5) = 0.5 exactly.
	// lerp(#000000, #ffffff, 0.5) = uint8(0 + 0.5*255 + 0.5) = uint8(128) = #808080.
	s.SetThreshold(0.5)
	s.SetGradient(true)
	s.SetStyle("", NewStyle("").WithColors("#000000", "bg"))
	s.SetStyle("high", NewStyle("high").WithColors("#ffffff", "bg"))

	s.SetProvider(FloatSlice{0.75})
	s.SetBounds(0, 0, 1, 1)
	s.Render(r)

	fg := cs.Fg(0,  0)
	if fg != "#808080" {
		t.Errorf("midpoint gradient: fg = %q; want #808080", fg)
	}
}

func TestSparkline_Gradient_False_HardCutoff(t *testing.T) {
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())

	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	s.SetThreshold(0.5)
	s.SetGradient(false) // explicit: no gradient
	s.SetStyle("", NewStyle("").WithColors("#000000", "bg"))
	s.SetStyle("high", NewStyle("high").WithColors("#ffffff", "bg"))

	// value = 0.9 → above threshold, no gradient → must be exact high color
	s.SetProvider(FloatSlice{0.9})
	s.SetBounds(0, 0, 1, 1)
	s.Render(r)

	fg := cs.Fg(0,  0)
	if fg != "#ffffff" {
		t.Errorf("hard cutoff above threshold: fg = %q; want #ffffff", fg)
	}
}

func TestSparkline_Gradient_AllRowsSameColor(t *testing.T) {
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())

	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(1)
	s.SetThreshold(0.5)
	s.SetGradient(true)
	s.SetStyle("", NewStyle("").WithColors("#000000", "bg"))
	s.SetStyle("high", NewStyle("high").WithColors("#ffffff", "bg"))

	s.SetProvider(FloatSlice{1})
	s.SetBounds(0, 0, 1, 3) // 3 rows
	s.Render(r)

	fg0 := cs.Fg(0,  0)
	fg1 := cs.Fg(0,  1)
	fg2 := cs.Fg(0,  2)
	if fg0 != fg1 || fg1 != fg2 {
		t.Errorf("all rows should share the same gradient color, got %q %q %q", fg0, fg1, fg2)
	}
}

func TestSparkline_Threshold_Zero_Disabled(t *testing.T) {
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())

	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(0)
	s.SetMax(10)
	s.SetThreshold(0) // disabled
	s.SetStyle("", NewStyle("").WithColors("base-fg", "base-bg"))
	s.SetStyle("high", NewStyle("high").WithColors("high-fg", "high-bg"))

	s.SetProvider(FloatSlice{9})
	s.SetBounds(0, 0, 1, 1)
	s.Render(r)

	fg := cs.Fg(0,  0)
	if fg != "base-fg" {
		t.Errorf("threshold=0: fg = %q; want base-fg (high style should not apply)", fg)
	}
}

// ---- RingBuffer as DataProvider --------------------------------------------

func TestSparkline_RingBufferProvider(t *testing.T) {
	// Verify RingBuffer[float64] satisfies DataProvider and renders correctly.
	rb := NewRingBuffer[float64](3)
	rb.Add(1.0) // oldest
	rb.Add(2.0)
	rb.Add(3.0) // newest

	s := NewSparkline("s", "")
	s.SetAbsolute(true)
	s.SetMin(1)
	s.SetMax(3)
	s.SetProvider(rb)
	cs := renderSparkline(s, 3, 1)

	// col 2 (rightmost) = newest = 3 = max → '█'
	if cs.Get(2,  0) != "█" {
		t.Errorf("rightmost col = %q; want █", cs.Get(2,  0))
	}
	// col 0 (leftmost) = oldest = 1 = min → '▁'
	if cs.Get(0,  0) != "▁" {
		t.Errorf("leftmost col = %q; want ▁", cs.Get(0,  0))
	}
}

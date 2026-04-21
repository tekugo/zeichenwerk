package widgets

import (
	"strings"
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

// newTestShimmer creates a Shimmer with a bare theme applied so tests can
// call Tick and Render without a live screen.
func newTestShimmer(text string, bandWidth, edgeWidth int) *Shimmer {
	sh := NewShimmer("sh", "")
	sh.SetBandWidth(bandWidth)
	sh.SetEdgeWidth(edgeWidth)
	sh.SetText(text)
	return sh
}

// renderShimmer renders sh at the given width, returns each line as a string.
func renderShimmer(t *testing.T, sh *Shimmer, w, h int) []string {
	t.Helper()
	theme := NewTheme()
	theme.SetColors(map[string]string{
		"$fg":   "#888888",
		"$band": "#ffffff",
		"$bg":   "#000000",
	})
	theme.AddStyles(
		NewStyle("").WithColors("$fg", "$bg").WithMargin(0).WithPadding(0),
		NewStyle("shimmer").WithColors("$fg", "$bg"),
		NewStyle("shimmer/band").WithForeground("$band"),
	)
	sh.Apply(theme)

	cs := NewTestScreen()
	ren := NewRenderer(cs, theme)
	sh.SetBounds(0, 0, w, h)
	sh.Render(ren)

	lines := make([]string, h)
	for y := 0; y < h; y++ {
		var sb strings.Builder
		for x := 0; x < w; x++ {
			ch := cs.Get(x, y)
			if ch == "" {
				ch = " "
			}
			sb.WriteString(ch)
		}
		lines[y] = sb.String()
	}
	return lines
}

// ---- Constructor defaults --------------------------------------------------

func TestShimmerDefaults(t *testing.T) {
	sh := NewShimmer("sh", "")
	if sh.bandWidth != 6 {
		t.Errorf("bandWidth = %d, want 6", sh.bandWidth)
	}
	if sh.edgeWidth != 3 {
		t.Errorf("edgeWidth = %d, want 3", sh.edgeWidth)
	}
	if sh.bandPos != 0 {
		t.Errorf("bandPos = %d, want 0", sh.bandPos)
	}
	if sh.maxWidth != 0 {
		t.Errorf("maxWidth = %d, want 0", sh.maxWidth)
	}
	if sh.Flag(FlagFocusable) {
		t.Error("shimmer must not be focusable by default")
	}
}

// ---- SetText ---------------------------------------------------------------

func TestShimmerSetTextSingleLine(t *testing.T) {
	sh := NewShimmer("sh", "")
	sh.SetText("Hello")
	if sh.maxWidth != 5 {
		t.Errorf("maxWidth = %d, want 5", sh.maxWidth)
	}
	if len(sh.lines) != 1 {
		t.Errorf("lines count = %d, want 1", len(sh.lines))
	}
}

func TestShimmerSetTextMultiLine(t *testing.T) {
	sh := NewShimmer("sh", "")
	sh.SetText("Hello\nWorld!\nHi")
	// longest is "World!" = 6 cols
	if sh.maxWidth != 6 {
		t.Errorf("maxWidth = %d, want 6", sh.maxWidth)
	}
	if len(sh.lines) != 3 {
		t.Errorf("lines count = %d, want 3", len(sh.lines))
	}
}

func TestShimmerSetTextResetsBandPos(t *testing.T) {
	sh := NewShimmer("sh", "")
	sh.SetText("AAAA")
	sh.bandPos = 3
	sh.SetText("BBBB")
	if sh.bandPos != 0 {
		t.Errorf("bandPos = %d after SetText, want 0", sh.bandPos)
	}
}

func TestShimmerSetTextWideRune(t *testing.T) {
	sh := NewShimmer("sh", "")
	// "日本" — two wide runes, each 2 cols → maxWidth = 4
	sh.SetText("日本")
	if sh.maxWidth != 4 {
		t.Errorf("maxWidth = %d for wide rune text, want 4", sh.maxWidth)
	}
}

// ---- Clamps ----------------------------------------------------------------

func TestShimmerSetBandWidthClamp(t *testing.T) {
	sh := NewShimmer("sh", "")
	sh.SetBandWidth(0)
	if sh.bandWidth != 1 {
		t.Errorf("bandWidth = %d after clamping 0, want 1", sh.bandWidth)
	}
	sh.SetBandWidth(-5)
	if sh.bandWidth != 1 {
		t.Errorf("bandWidth = %d after clamping -5, want 1", sh.bandWidth)
	}
}

func TestShimmerSetEdgeWidthClamp(t *testing.T) {
	sh := NewShimmer("sh", "")
	sh.SetEdgeWidth(-1)
	if sh.edgeWidth != 0 {
		t.Errorf("edgeWidth = %d after clamping -1, want 0", sh.edgeWidth)
	}
}

// ---- Tick ------------------------------------------------------------------

func TestShimmerTickAdvancesBandPos(t *testing.T) {
	sh := newTestShimmer("Hello", 2, 1)
	sh.Tick()
	if sh.bandPos != 1 {
		t.Errorf("bandPos = %d after one Tick, want 1", sh.bandPos)
	}
}

func TestShimmerTickWrapsAtMaxWidth(t *testing.T) {
	sh := newTestShimmer("Hello", 2, 1) // maxWidth = 5
	sh.bandPos = 4
	sh.Tick() // 4+1 = 5 % 5 = 0
	if sh.bandPos != 0 {
		t.Errorf("bandPos = %d after wrap, want 0", sh.bandPos)
	}
}

func TestShimmerTickNoopWhenEmpty(t *testing.T) {
	sh := NewShimmer("sh", "")
	sh.Tick() // maxWidth == 0, should not panic or advance
	if sh.bandPos != 0 {
		t.Errorf("bandPos = %d for empty shimmer, want 0", sh.bandPos)
	}
}

// ---- intensity -------------------------------------------------------------

func TestShimmerIntensityCoreIs1(t *testing.T) {
	// bandWidth=6 → halfWidth=3, edgeWidth=2
	// bandPos=0, maxWidth=20
	// dist(0) = 0 → edge region → intensity < 1
	// dist(1) = 1 → edge region → intensity < 1
	// dist(2) = 2 → edge == edgeWidth → inside core (dist <= ew+hw = 2+3 = 5)
	sh := newTestShimmer(strings.Repeat("x", 20), 6, 2)
	sh.maxWidth = 20
	sh.bandPos = 0

	// Column 0: dist = 0, in leading edge → intensity depends on gradient formula
	i0 := sh.intensity(0)
	if i0 <= 0.0 || i0 > 1.0 {
		t.Errorf("intensity at col 0 (dist=0) = %f, want (0,1]", i0)
	}

	// Column 3: dist = 3 = edgeWidth+bandWidth/2 = 2+3 = 5? No wait:
	// dist(3) from bandPos=0: d1=(3-0+20)%20=3, d2=(0-3+20)%20=17 → dist=3
	// edgeWidth=2, bandWidth=6, halfWidth=3
	// dist(3) > edgeWidth(2) and dist(3) <= edgeWidth+halfWidth(5) → intensity=1.0
	i3 := sh.intensity(3)
	if i3 != 1.0 {
		t.Errorf("intensity at col 3 (inside core) = %f, want 1.0", i3)
	}
}

func TestShimmerIntensityOutsideIs0(t *testing.T) {
	// bandPos=0, bandWidth=4, edgeWidth=1 → halfWidth=2
	// total span = edgeWidth + halfWidth = 1+2 = 3
	// col 4: dist=4 > 3 → intensity=0
	sh := newTestShimmer(strings.Repeat("x", 20), 4, 1)
	sh.maxWidth = 20
	sh.bandPos = 0

	i := sh.intensity(4)
	if i != 0.0 {
		t.Errorf("intensity at col 4 (outside band) = %f, want 0.0", i)
	}
}

func TestShimmerIntensityHardEdge(t *testing.T) {
	// edgeWidth=0 means: dist <= 0 is impossible (dist >= 0, but only 0 when c==bandPos).
	// col == bandPos: dist=0, NOT <= edgeWidth(0) in the edge branch since 0 <= 0 is true
	// but edgeWidth=0 so the gradient formula gives 1.0 - 0/(0+1) = 1.0
	// col != bandPos but dist <= halfWidth: intensity=1
	// col outside: intensity=0
	sh := newTestShimmer(strings.Repeat("x", 20), 6, 0)
	sh.maxWidth = 20
	sh.bandPos = 5

	// At bandPos itself: dist=0, edgeWidth=0 branch: intensity = 1.0 - 0/1 = 1.0
	iBand := sh.intensity(5)
	if iBand != 1.0 {
		t.Errorf("intensity at bandPos (edgeWidth=0) = %f, want 1.0", iBand)
	}

	// Well past band: dist > halfWidth → 0
	iOut := sh.intensity(12)
	if iOut != 0.0 {
		t.Errorf("intensity far from band (edgeWidth=0) = %f, want 0.0", iOut)
	}
}

// ---- Render ----------------------------------------------------------------

func TestShimmerRenderSingleLine(t *testing.T) {
	sh := newTestShimmer("Hello", 2, 1)
	lines := renderShimmer(t, sh, 5, 1)
	if lines[0] != "Hello" {
		t.Errorf("rendered = %q, want %q", lines[0], "Hello")
	}
}

func TestShimmerRenderPadsShortLines(t *testing.T) {
	sh := newTestShimmer("Hi\nHello", 2, 0)
	lines := renderShimmer(t, sh, 5, 2)
	if lines[0] != "Hi   " {
		t.Errorf("line 0 = %q, want %q", lines[0], "Hi   ")
	}
	if lines[1] != "Hello" {
		t.Errorf("line 1 = %q, want %q", lines[1], "Hello")
	}
}

func TestShimmerRenderMultiLinesSameColumnBand(t *testing.T) {
	// All lines should have the same column positions rendered (just verify
	// content equality independent of colour). bandPos=0, edgeWidth=0 gives
	// distinct but deterministic colours.
	sh := newTestShimmer("AB\nCD", 2, 0)
	sh.bandPos = 0
	lines := renderShimmer(t, sh, 2, 2)
	if lines[0] != "AB" {
		t.Errorf("line 0 = %q, want %q", lines[0], "AB")
	}
	if lines[1] != "CD" {
		t.Errorf("line 1 = %q, want %q", lines[1], "CD")
	}
}

// ---- Hint ------------------------------------------------------------------

func TestShimmerHint(t *testing.T) {
	sh := NewShimmer("sh", "")
	theme := NewTheme()
	theme.SetColors(map[string]string{"$fg": "#ffffff", "$bg": "#000000"})
	theme.AddStyles(
		NewStyle("").WithColors("$fg", "$bg").WithMargin(0).WithPadding(0),
		NewStyle("shimmer").WithColors("$fg", "$bg"),
		NewStyle("shimmer/band").WithForeground("$fg"),
	)
	sh.Apply(theme)
	sh.SetText("Hello\nWorld!")
	w, h := sh.Hint()
	if w != 6 {
		t.Errorf("Hint width = %d, want 6", w)
	}
	if h != 2 {
		t.Errorf("Hint height = %d, want 2", h)
	}
}

func TestShimmerHintManualOverride(t *testing.T) {
	sh := NewShimmer("sh", "")
	sh.SetText("Hello")
	sh.SetHint(-1, 1)
	w, h := sh.Hint()
	if w != -1 {
		t.Errorf("Hint width = %d, want -1", w)
	}
	if h != 1 {
		t.Errorf("Hint height = %d, want 1", h)
	}
}

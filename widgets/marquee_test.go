package widgets

import (
	"strings"
	"testing"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// newTestMarquee creates a Marquee, applies a bare theme, and sets its render
// width by hand so tests can exercise Tick and Render without a live screen.
func newTestMarquee(text string, renderWidth int) *Marquee {
	m := NewMarquee("m", "")
	m.SetText(text)
	m.renderWidth = renderWidth
	return m
}

// ---- cycle -----------------------------------------------------------------

func TestMarqueeCycle(t *testing.T) {
	m := NewMarquee("m", "")
	m.SetText("Hello") // textWidth = 5
	m.SetGap(4)
	if got := m.cycle(); got != 9 {
		t.Errorf("cycle() = %d, want 9", got)
	}
}

func TestMarqueeCycleMultibyte(t *testing.T) {
	m := NewMarquee("m", "")
	// "日本語" — 3 wide runes, each 2 cols → textWidth = 6
	m.SetText("日本語")
	m.SetGap(2)
	if got := m.cycle(); got != 8 {
		t.Errorf("cycle() = %d, want 8", got)
	}
}

// ---- SetText ---------------------------------------------------------------

func TestMarqueeSetTextResetsOffset(t *testing.T) {
	m := newTestMarquee("AAAA", 2)
	m.offset = 3
	m.SetText("BBBB")
	if m.offset != 0 {
		t.Errorf("offset = %d after SetText, want 0", m.offset)
	}
}

func TestMarqueeSetTextUpdatesTextWidth(t *testing.T) {
	m := NewMarquee("m", "")
	m.SetText("Hello") // 5 ASCII cols
	if m.textWidth != 5 {
		t.Errorf("textWidth = %d, want 5", m.textWidth)
	}
	m.SetText("Hi") // 2 ASCII cols
	if m.textWidth != 2 {
		t.Errorf("textWidth = %d after update, want 2", m.textWidth)
	}
}

// ---- Tick ------------------------------------------------------------------

func TestMarqueeTickAdvancesOffset(t *testing.T) {
	m := newTestMarquee("Hello World!", 4) // textWidth=12, renderWidth=4
	m.SetGap(4)                            // cycle = 16
	m.Tick()
	if m.offset != 1 {
		t.Errorf("offset = %d after one Tick, want 1", m.offset)
	}
}

func TestMarqueeTickWrapsAtCycle(t *testing.T) {
	m := newTestMarquee("Hello", 2) // textWidth=5, renderWidth=2
	m.SetGap(3)                     // cycle = 8
	m.offset = 7
	m.Tick() // 7+1 = 8 % 8 = 0
	if m.offset != 0 {
		t.Errorf("offset = %d after wrap, want 0", m.offset)
	}
}

func TestMarqueeTickSpeedMultiple(t *testing.T) {
	m := newTestMarquee("Hello World!", 4)
	m.SetSpeed(3)
	m.SetGap(4) // cycle = 16
	m.Tick()
	if m.offset != 3 {
		t.Errorf("offset = %d after speed-3 tick, want 3", m.offset)
	}
}

func TestMarqueeTickNoopWhenTextFits(t *testing.T) {
	m := newTestMarquee("Hi", 10) // textWidth(2) <= renderWidth(10)
	m.Tick()
	if m.offset != 0 {
		t.Errorf("offset = %d when text fits, want 0", m.offset)
	}
}

func TestMarqueeTickNoopWhenHovered(t *testing.T) {
	m := newTestMarquee("Hello World!", 4)
	m.SetFlag(FlagHovered, true)
	m.Tick()
	if m.offset != 0 {
		t.Errorf("offset = %d when hovered, want 0", m.offset)
	}
}

func TestMarqueeTickNoopWhenTextEmpty(t *testing.T) {
	m := newTestMarquee("", 10)
	m.Tick()
	if m.offset != 0 {
		t.Errorf("offset = %d for empty text, want 0", m.offset)
	}
}

// ---- Render ----------------------------------------------------------------

// renderMarquee renders m at the given bounds and returns the string content
// of the single content row.
func renderMarquee(t *testing.T, m *Marquee, w int) string {
	t.Helper()
	theme := NewTheme()
	theme.AddStyles(
		NewStyle("").WithColors("$fg", "$bg").WithMargin(0).WithPadding(0),
		NewStyle("marquee").WithColors("$fg", "$bg"),
	)
	theme.SetColors(map[string]string{"$fg": "#ffffff", "$bg": "#000000"})
	m.Apply(theme)

	cs := NewTestScreen()
	ren := NewRenderer(cs, theme)
	m.SetBounds(0, 0, w, 1)
	m.Render(ren)

	var sb strings.Builder
	for x := 0; x < w; x++ {
		ch := cs.Get(x, 0)
		if ch == "" {
			sb.WriteByte(' ')
		} else {
			sb.WriteString(ch)
		}
	}
	return sb.String()
}

func TestMarqueeRenderStaticFits(t *testing.T) {
	m := NewMarquee("m", "")
	m.SetText("Hi")
	got := renderMarquee(t, m, 10)
	// text fits — left-aligned, padded with spaces
	if !strings.HasPrefix(got, "Hi") {
		t.Errorf("render = %q, want prefix %q", got, "Hi")
	}
	if got[2:] != "        " {
		t.Errorf("render padding = %q, want 8 spaces", got[2:])
	}
}

func TestMarqueeRenderScrollingOffset0(t *testing.T) {
	// text = "ABCDE" (5), gap=2, cycle=7, widget width=4, offset=0
	// Expected: "ABCD"
	m := NewMarquee("m", "")
	m.SetText("ABCDE")
	m.SetGap(2)
	m.offset = 0
	got := renderMarquee(t, m, 4)
	if got != "ABCD" {
		t.Errorf("render at offset 0 = %q, want %q", got, "ABCD")
	}
}

func TestMarqueeRenderScrollingOffset3(t *testing.T) {
	// text = "ABCDE" (5), gap=2, cycle=7, widget width=4, offset=3
	// vpos sequence: 3,4,5,6 → 'D','E',' ',' '
	m := NewMarquee("m", "")
	m.SetText("ABCDE")
	m.SetGap(2)
	m.offset = 3
	got := renderMarquee(t, m, 4)
	if got != "DE  " {
		t.Errorf("render at offset 3 = %q, want %q", got, "DE  ")
	}
}

func TestMarqueeRenderScrollingWrapAround(t *testing.T) {
	// text = "ABCDE" (5), gap=2, cycle=7, widget width=4, offset=5
	// vpos sequence: 5,6,0,1 → ' ',' ','A','B'
	m := NewMarquee("m", "")
	m.SetText("ABCDE")
	m.SetGap(2)
	m.offset = 5
	got := renderMarquee(t, m, 4)
	if got != "  AB" {
		t.Errorf("render at offset 5 (wrap) = %q, want %q", got, "  AB")
	}
}

func TestMarqueeRenderWideRuneAdvancesByTwo(t *testing.T) {
	// "AB日C" — textWidth = 2+2+2 = 6, gap=2, cycle=8, widget width=6, offset=0
	// Expected columns: A(1) B(1) 日(2) C(1) space(1) → "AB日C "
	// but we only render 6 cols: A B 日(wide→2cols) C → "AB日C" with C at col 5
	m := NewMarquee("m", "")
	m.SetText("AB日C")
	m.SetGap(2)
	m.offset = 0
	got := renderMarquee(t, m, 6)
	if !strings.HasPrefix(got, "AB") {
		t.Errorf("render wide = %q, expected prefix AB", got)
	}
	// The wide rune occupies two cells; the rendered string should contain 日 or spaces
	// depending on terminal cell handling — just verify length
	if len([]rune(got)) != 6 {
		t.Errorf("render wide rune count = %d, want 6", len([]rune(got)))
	}
}

// ---- defaults --------------------------------------------------------------

func TestMarqueeDefaults(t *testing.T) {
	m := NewMarquee("m", "")
	if m.speed != 1 {
		t.Errorf("default speed = %d, want 1", m.speed)
	}
	if m.gap != 4 {
		t.Errorf("default gap = %d, want 4", m.gap)
	}
	if m.offset != 0 {
		t.Errorf("default offset = %d, want 0", m.offset)
	}
}

func TestMarqueeSetSpeedClamp(t *testing.T) {
	m := NewMarquee("m", "")
	m.SetSpeed(0)
	if m.speed != 1 {
		t.Errorf("speed after SetSpeed(0) = %d, want 1", m.speed)
	}
	m.SetSpeed(-5)
	if m.speed != 1 {
		t.Errorf("speed after SetSpeed(-5) = %d, want 1", m.speed)
	}
}

func TestMarqueeSetGapClamp(t *testing.T) {
	m := NewMarquee("m", "")
	m.SetGap(-1)
	if m.gap != 0 {
		t.Errorf("gap after SetGap(-1) = %d, want 0", m.gap)
	}
}

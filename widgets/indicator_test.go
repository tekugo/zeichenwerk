package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestIndicator_Defaults(t *testing.T) {
	i := NewIndicator("ind", "", Info, "Online")
	if i.Level() != Info {
		t.Errorf("Level() = %q; want %q", i.Level(), Info)
	}
	if i.Label() != "Online" {
		t.Errorf("Label() = %q; want %q", i.Label(), "Online")
	}
	if i.dot != "●" {
		t.Errorf("default glyph = %q; want %q", i.dot, "●")
	}
}

// ── Hint ──────────────────────────────────────────────────────────────────────

func TestIndicator_Hint_RuneCount(t *testing.T) {
	i := NewIndicator("ind", "", Info, "Online")
	w, h := i.Hint()
	if w != 2+len("Online") {
		t.Errorf("Hint width = %d; want %d", w, 2+len("Online"))
	}
	if h != 1 {
		t.Errorf("Hint height = %d; want 1", h)
	}
}

func TestIndicator_Hint_MultiByte(t *testing.T) {
	i := NewIndicator("ind", "", Info, "über") // 4 runes
	w, _ := i.Hint()
	if w != 2+4 {
		t.Errorf("Hint width = %d; want 6 (2 + rune count)", w)
	}
}

// ── State ─────────────────────────────────────────────────────────────────────

func TestIndicator_State_ReturnsLevel(t *testing.T) {
	cases := []struct {
		level Level
		want  string
	}{
		{Debug, "debug"},
		{Info, "info"},
		{Success, "success"},
		{Warning, "warning"},
		{Error, "error"},
		{Fatal, "fatal"},
	}
	for _, c := range cases {
		i := NewIndicator("ind", "", c.level, "x")
		if got := i.State(); got != c.want {
			t.Errorf("level=%q: State() = %q; want %q", c.level, got, c.want)
		}
	}
}

func TestIndicator_State_EmptyLevelFallsBackToInfo(t *testing.T) {
	i := NewIndicator("ind", "", "", "x")
	if got := i.State(); got != "info" {
		t.Errorf("empty level: State() = %q; want %q", got, "info")
	}
}

// ── Setters ───────────────────────────────────────────────────────────────────

func TestIndicator_SetLevel_UpdatesField(t *testing.T) {
	i := NewIndicator("ind", "", Info, "x")
	i.SetLevel(Error)
	if i.Level() != Error {
		t.Errorf("Level() after SetLevel(Error) = %q; want %q", i.Level(), Error)
	}
	if i.State() != "error" {
		t.Errorf("State() after SetLevel(Error) = %q; want %q", i.State(), "error")
	}
}

func TestIndicator_SetLabel_UpdatesField(t *testing.T) {
	i := NewIndicator("ind", "", Info, "old")
	i.SetLabel("a much longer label")
	if i.Label() != "a much longer label" {
		t.Errorf("Label() = %q; want %q", i.Label(), "a much longer label")
	}
	w, _ := i.Hint()
	if w != 2+len("a much longer label") {
		t.Errorf("Hint width after SetLabel = %d; want %d", w, 2+len("a much longer label"))
	}
}

// ── Apply / Theme strings ─────────────────────────────────────────────────────

func TestIndicator_Apply_ReadsDotString(t *testing.T) {
	theme := NewTheme()
	theme.SetStrings(map[string]string{"indicator.dot": "■"})
	i := NewIndicator("ind", "", Info, "x")
	i.Apply(theme)
	if i.dot != "■" {
		t.Errorf("dot after Apply with theme string = %q; want %q", i.dot, "■")
	}
}

func TestIndicator_Apply_KeepsDefaultDotWhenStringMissing(t *testing.T) {
	theme := NewTheme()
	i := NewIndicator("ind", "", Info, "x")
	i.Apply(theme)
	if i.dot != "●" {
		t.Errorf("dot after Apply with empty theme = %q; want %q", i.dot, "●")
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func newIndicatorRenderer() (*TestScreen, *Renderer) {
	cs := NewTestScreen()
	return cs, NewRenderer(cs, NewTheme())
}

func renderIndicator(i *Indicator, w, h int) *TestScreen {
	cs, r := newIndicatorRenderer()
	i.SetBounds(0, 0, w, h)
	i.Render(r)
	return cs
}

func TestIndicator_Render_GlyphAndLabel(t *testing.T) {
	i := NewIndicator("ind", "", Info, "Online")
	cs := renderIndicator(i, 10, 1)

	if got := cs.Get(0, 0); got != "●" {
		t.Errorf("col 0 = %q; want %q (glyph)", got, "●")
	}
	want := "Online"
	for k, ch := range want {
		if got := cs.Get(2+k, 0); got != string(ch) {
			t.Errorf("col %d = %q; want %q", 2+k, got, string(ch))
		}
	}
}

func TestIndicator_Render_GlyphUsesLevelStyle(t *testing.T) {
	theme := NewTheme()
	theme.AddStyles(
		NewStyle("indicator").WithColors("base-fg", "base-bg"),
		NewStyle("indicator:error").WithColors("error-fg", "base-bg"),
	)
	i := NewIndicator("ind", "", Error, "boom")
	i.Apply(theme)

	cs, r := newIndicatorRenderer()
	i.SetBounds(0, 0, 10, 1)
	i.Render(r)

	if got := cs.Fg(0, 0); got != "error-fg" {
		t.Errorf("glyph fg = %q; want %q (indicator:error)", got, "error-fg")
	}
}

func TestIndicator_Render_LabelUsesBaseStyle(t *testing.T) {
	theme := NewTheme()
	theme.AddStyles(
		NewStyle("indicator").WithColors("base-fg", "base-bg"),
		NewStyle("indicator:error").WithColors("error-fg", "base-bg"),
	)
	i := NewIndicator("ind", "", Error, "boom")
	i.Apply(theme)

	cs, r := newIndicatorRenderer()
	i.SetBounds(0, 0, 10, 1)
	i.Render(r)

	for k := 0; k < len("boom"); k++ {
		if got := cs.Fg(2+k, 0); got != "base-fg" {
			t.Errorf("label col %d fg = %q; want %q (base, never the level variant)", 2+k, got, "base-fg")
		}
	}
}

func TestIndicator_Render_EmptyLevelUsesInfoStyle(t *testing.T) {
	theme := NewTheme()
	theme.AddStyles(
		NewStyle("indicator").WithColors("base-fg", ""),
		NewStyle("indicator:info").WithColors("info-fg", ""),
	)
	i := NewIndicator("ind", "", "", "x")
	i.Apply(theme)

	cs, r := newIndicatorRenderer()
	i.SetBounds(0, 0, 10, 1)
	i.Render(r)

	if got := cs.Fg(0, 0); got != "info-fg" {
		t.Errorf("empty-level glyph fg = %q; want %q (indicator:info fallback)", got, "info-fg")
	}
}

func TestIndicator_Render_LabelClippedToContentWidth(t *testing.T) {
	i := NewIndicator("ind", "", Info, "Connection refused")
	cs := renderIndicator(i, 6, 1)

	// With content width 6: glyph at 0, space at 1, label fills cols 2..5 (4 chars).
	want := "Conn"
	for k, ch := range want {
		if got := cs.Get(2+k, 0); got != string(ch) {
			t.Errorf("col %d = %q; want %q", 2+k, got, string(ch))
		}
	}
	// Col 6 must be untouched.
	if got := cs.Get(6, 0); got != "" {
		t.Errorf("col 6 = %q; want empty (past content width)", got)
	}
}

package zeichenwerk

import "testing"

// ── Put ───────────────────────────────────────────────────────────────────────

func TestRenderer_Put_WritesCell(t *testing.T) {
	cs := newCellScreen()
	r := NewRenderer(cs, NewTheme())
	r.Put(3, 2, "X")
	if cs.Get(3, 2) != "X" {
		t.Errorf("Get(3,2) = %q after Put; want %q", cs.Get(3, 2), "X")
	}
}

func TestRenderer_Get_ReflectsPut(t *testing.T) {
	cs := newCellScreen()
	r := NewRenderer(cs, NewTheme())
	r.Put(0, 0, "Z")
	if r.Get(0, 0) != "Z" {
		t.Errorf("Get(0,0) = %q; want %q", r.Get(0, 0), "Z")
	}
}

// ── Fill ─────────────────────────────────────────────────────────────────────

func TestRenderer_Fill_FillsRectangle(t *testing.T) {
	cs := newCellScreen()
	r := NewRenderer(cs, NewTheme())
	r.Fill(1, 1, 3, 2, "#")
	for x := 1; x <= 3; x++ {
		for y := 1; y <= 2; y++ {
			if cs.Get(x, y) != "#" {
				t.Errorf("Get(%d,%d) = %q; want %q", x, y, cs.Get(x, y), "#")
			}
		}
	}
}

func TestRenderer_Fill_DoesNotWriteOutside(t *testing.T) {
	cs := newCellScreen()
	r := NewRenderer(cs, NewTheme())
	r.Fill(1, 1, 2, 2, "#")
	if cs.Get(0, 0) != "" {
		t.Errorf("Get(0,0) = %q; should be empty (outside fill rect)", cs.Get(0, 0))
	}
	if cs.Get(3, 3) != "" {
		t.Errorf("Get(3,3) = %q; should be empty (outside fill rect)", cs.Get(3, 3))
	}
}

// ── Text ─────────────────────────────────────────────────────────────────────

func TestRenderer_Text_WritesString(t *testing.T) {
	cs := newCellScreen()
	r := NewRenderer(cs, NewTheme())
	r.Text(0, 0, "Hello", 0)
	got := cs.Get(0, 0) + cs.Get(1, 0) + cs.Get(2, 0) + cs.Get(3, 0) + cs.Get(4, 0)
	if got != "Hello" {
		t.Errorf("Text wrote %q; want %q", got, "Hello")
	}
}

// ── Repeat ────────────────────────────────────────────────────────────────────

func TestRenderer_Repeat_Horizontal(t *testing.T) {
	cs := newCellScreen()
	r := NewRenderer(cs, NewTheme())
	r.Repeat(0, 0, 1, 0, 4, "-")
	for x := 0; x < 4; x++ {
		if cs.Get(x, 0) != "-" {
			t.Errorf("Get(%d,0) = %q; want %q", x, cs.Get(x, 0), "-")
		}
	}
}

func TestRenderer_Repeat_Vertical(t *testing.T) {
	cs := newCellScreen()
	r := NewRenderer(cs, NewTheme())
	r.Repeat(0, 0, 0, 1, 3, "|")
	for y := 0; y < 3; y++ {
		if cs.Get(0, y) != "|" {
			t.Errorf("Get(0,%d) = %q; want %q", y, cs.Get(0, y), "|")
		}
	}
}

// ── Line ─────────────────────────────────────────────────────────────────────

func TestRenderer_Line_WithEndpoints(t *testing.T) {
	cs := newCellScreen()
	r := NewRenderer(cs, NewTheme())
	// Horizontal line: length=2 middles → cells 0,1,2,3 = S,M,M,E
	r.Line(0, 0, 1, 0, 2, "S", "M", "E")
	if cs.Get(0, 0) != "S" {
		t.Errorf("Get(0,0) = %q; want start %q", cs.Get(0, 0), "S")
	}
	if cs.Get(1, 0) != "M" || cs.Get(2, 0) != "M" {
		t.Errorf("middle cells: %q %q; want M M", cs.Get(1, 0), cs.Get(2, 0))
	}
	if cs.Get(3, 0) != "E" {
		t.Errorf("Get(3,0) = %q; want end %q", cs.Get(3, 0), "E")
	}
}

// ── Set resolves theme variables ──────────────────────────────────────────────

func TestRenderer_Set_ResolvesThemeVariable(t *testing.T) {
	theme := NewTheme()
	theme.SetColors(map[string]string{"$fg": "#ffffff"})
	cs := newCellScreen()
	r := NewRenderer(cs, theme)
	// Just verify Set doesn't panic with a variable — the actual color is
	// written to the screen's internal state, not readable from cellScreen.
	r.Set("$fg", "", "")
}

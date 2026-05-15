package main

import "testing"

func TestJunction_RuneForMaskInFamilyRoundTrip(t *testing.T) {
	for fam := range boxFamilies {
		for mask := Mask(1); mask < 16; mask++ {
			ch := runeForMask(fam, mask)
			if ch == "" {
				continue
			}
			gotMask, ok := maskInFamily(ch, fam)
			if !ok {
				t.Errorf("family=%s mask=%04b: rune %q not in family table", fam, mask, ch)
				continue
			}
			// Some families (e.g. double) reuse the same rune for
			// multiple masks because they lack half-stub characters.
			// All we require is that the rune round-trips back to
			// SOME mask that produces the same rune.
			if runeForMask(fam, gotMask) != ch {
				t.Errorf("family=%s mask=%04b: rune %q → mask %04b → %q",
					fam, mask, ch, gotMask, runeForMask(fam, gotMask))
			}
		}
	}
}

func TestJunction_HorizontalCrossesVertical(t *testing.T) {
	d := NewDocument(5, 3)
	app := newTestApp(d)
	e := app.editor
	e.border = "thin"
	e.style = "default"

	// Vertical line down the centre column.
	e.drawVLine(2, 0, 2)
	// Horizontal line across the middle row.
	e.drawHLine(0, 4, 1)

	if got := d.Cells[1][2].Ch; got != "┼" {
		t.Errorf("intersection cell = %q; want ┼", got)
	}
	if got := d.Cells[1][0].Ch; got != "─" {
		t.Errorf("leftmost cell = %q; want ─", got)
	}
	if got := d.Cells[1][4].Ch; got != "─" {
		t.Errorf("rightmost cell = %q; want ─", got)
	}
	if got := d.Cells[0][2].Ch; got != "│" {
		t.Errorf("top of vertical = %q; want │", got)
	}
}

func TestJunction_BorderProducesCorners(t *testing.T) {
	d := NewDocument(5, 4)
	app := newTestApp(d)
	e := app.editor
	e.visualX, e.visualY = 0, 0
	e.cursorX, e.cursorY = 4, 3
	e.mode = ModeVisual
	e.borderSelection("thin")

	want := map[[2]int]string{
		{0, 0}: "┌", {4, 0}: "┐",
		{0, 3}: "└", {4, 3}: "┘",
		{1, 0}: "─", {2, 0}: "─", {3, 0}: "─",
		{1, 3}: "─", {2, 3}: "─", {3, 3}: "─",
		{0, 1}: "│", {0, 2}: "│",
		{4, 1}: "│", {4, 2}: "│",
	}
	for pos, expect := range want {
		got := d.Cells[pos[1]][pos[0]].Ch
		if got != expect {
			t.Errorf("cell %v = %q; want %q", pos, got, expect)
		}
	}
}

// newTestApp builds a minimal App + Editor pair without spinning up the
// UI runtime. Sufficient for tests that exercise drawing logic directly.
func newTestApp(d *Document) *App {
	app := &App{
		history:  NewHistory(50),
		register: &Register{},
	}
	app.editor = &Editor{
		app:      app,
		doc:      d,
		mode:     ModeNormal,
		style:    "default",
		border:   "thin",
		lastRune: "█",
	}
	return app
}

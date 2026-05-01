package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestProgress_Defaults_Horizontal(t *testing.T) {
	p := NewProgress("p", "", true)
	if p.Percentage() != 0 {
		t.Errorf("Percentage() = %f; want 0 for new progress", p.Percentage())
	}
	// total=0 means indeterminate
	w, h := p.Hint()
	if h != 1 {
		t.Errorf("Hint height = %d for horizontal; want 1", h)
	}
	_ = w
}

func TestProgress_Defaults_Vertical(t *testing.T) {
	p := NewProgress("p", "", false)
	w, _ := p.Hint()
	if w != 1 {
		t.Errorf("Hint width = %d for vertical; want 1", w)
	}
}

// ── SetTotal ─────────────────────────────────────────────────────────────────

func TestProgress_SetTotal_SwitchesToDeterminate(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(100)
	p.Set(50)
	if p.Percentage() != 50.0 {
		t.Errorf("Percentage() = %f; want 50.0", p.Percentage())
	}
}

func TestProgress_SetTotal_Negative_ClampedToZero(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(-10)
	// total=0 → indeterminate → Percentage() = 0
	if p.Percentage() != 0 {
		t.Errorf("Percentage() = %f after SetTotal(-10); want 0 (clamped to indeterminate)", p.Percentage())
	}
}

func TestProgress_SetTotal_ReclampValue(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(100)
	p.Set(80)
	p.SetTotal(50) // new total is 50, value should clamp to 50
	if p.Percentage() != 100.0 {
		t.Errorf("Percentage() = %f after reclamp; want 100.0", p.Percentage())
	}
}

// ── Set / Percentage ──────────────────────────────────────────────────────────

func TestProgress_Set_ClampedToZero(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(100)
	p.Set(-50)
	if p.Percentage() != 0 {
		t.Errorf("Percentage() = %f after Set(-50); want 0 (clamped)", p.Percentage())
	}
}

func TestProgress_Set_ClampedToTotal(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(100)
	p.Set(200)
	if p.Percentage() != 100.0 {
		t.Errorf("Percentage() = %f after Set(200); want 100.0 (clamped)", p.Percentage())
	}
}

func TestProgress_Set_Indeterminate_Unclamped(t *testing.T) {
	p := NewProgress("p", "", true)
	// total=0 → indeterminate mode, value stored unclamped
	p.Set(999)
	if p.Percentage() != 0 {
		// percentage is still 0 in indeterminate mode
		t.Errorf("Percentage() = %f in indeterminate mode; want 0", p.Percentage())
	}
}

func TestProgress_Percentage_AtHalf(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(200)
	p.Set(100)
	if p.Percentage() != 50.0 {
		t.Errorf("Percentage() = %f; want 50.0", p.Percentage())
	}
}

func TestProgress_Percentage_AtFull(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(10)
	p.Set(10)
	if p.Percentage() != 100.0 {
		t.Errorf("Percentage() = %f; want 100.0", p.Percentage())
	}
}

// ── Increment ────────────────────────────────────────────────────────────────

func TestProgress_Increment(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(100)
	p.Set(30)
	p.Increment(20)
	if p.Percentage() != 50.0 {
		t.Errorf("Percentage() = %f after Increment(20); want 50.0", p.Percentage())
	}
}

func TestProgress_Increment_ClampsAtTotal(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(100)
	p.Set(90)
	p.Increment(50)
	if p.Percentage() != 100.0 {
		t.Errorf("Percentage() = %f after Increment beyond total; want 100.0", p.Percentage())
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func TestProgress_Render_Horizontal_ProducesOutput(t *testing.T) {
	p := NewProgress("p", "", true)
	p.SetTotal(10)
	p.Set(5) // 50%
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	p.SetBounds(0, 0, 10, 1)
	p.Render(r)

	// At least one non-empty cell expected
	found := false
	for x := 0; x < 10; x++ {
		if cs.Get(x, 0) != "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("horizontal render should produce at least one non-empty cell")
	}
}

func TestProgress_Render_Vertical_ProducesOutput(t *testing.T) {
	p := NewProgress("p", "", false)
	p.SetTotal(10)
	p.Set(5) // 50%
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	p.SetBounds(0, 0, 1, 10)
	p.Render(r)

	found := false
	for y := 0; y < 10; y++ {
		if cs.Get(0, y) != "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("vertical render should produce at least one non-empty cell")
	}
}

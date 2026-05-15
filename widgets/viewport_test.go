package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestViewport_Defaults(t *testing.T) {
	vp := NewViewport("v", "", "My Viewport")
	if !vp.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
	if vp.Title != "My Viewport" {
		t.Errorf("Title = %q; want %q", vp.Title, "My Viewport")
	}
	if len(vp.Children()) != 0 {
		t.Errorf("Children() len = %d; want 0 before Add", len(vp.Children()))
	}
}

// ── Add ───────────────────────────────────────────────────────────────────────

func TestViewport_Add_SetsChild(t *testing.T) {
	vp := NewViewport("v", "", "")
	child := NewStatic("s", "", "hello")
	if err := vp.Add(child); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if len(vp.Children()) != 1 {
		t.Errorf("Children() len = %d; want 1 after Add", len(vp.Children()))
	}
}

func TestViewport_Add_NilReturnsError(t *testing.T) {
	vp := NewViewport("v", "", "")
	if err := vp.Add(nil); err == nil {
		t.Error("Add(nil) should return ErrChildIsNil")
	}
}

func TestViewport_Add_ReplacesExistingChild(t *testing.T) {
	vp := NewViewport("v", "", "")
	c1 := NewStatic("c1", "", "A")
	c2 := NewStatic("c2", "", "B")
	vp.Add(c1)
	vp.Add(c2)
	children := vp.Children()
	if len(children) != 1 {
		t.Errorf("Children() len = %d; want 1 (replaced)", len(children))
	}
	if children[0].ID() != "c2" {
		t.Errorf("remaining child ID = %q; want %q", children[0].ID(), "c2")
	}
}

func TestViewport_Add_SetsParent(t *testing.T) {
	vp := NewViewport("v", "", "")
	child := NewStatic("s", "", "")
	vp.Add(child)
	if child.Parent() != vp {
		t.Error("child's parent should be the viewport")
	}
}

// ── Keyboard ─────────────────────────────────────────────────────────────────

// makeScrollableViewport creates a viewport at (0,0) with a child larger than the viewport.
func makeScrollableViewport() (*Viewport, *Static) {
	vp := NewViewport("v", "", "")
	child := NewStatic("s", "", "")
	child.SetHint(100, 50) // 100w × 50h child
	vp.Add(child)
	vp.SetBounds(0, 0, 20, 10) // 20w × 10h viewport
	vp.Layout()
	return vp, child
}

// Scroll is observed via the viewport's offset fields (tx, ty), not via
// the child's bounds. Layout pins the child at (0, 0, w, h) regardless of
// scroll position; the offset is applied through Translate at Render time.

func TestViewport_Keyboard_Down_ScrollsDown(t *testing.T) {
	vp, _ := makeScrollableViewport()
	tyBefore := vp.ty
	handled := vp.handleKey(BuildKey(tcell.KeyDown))
	if !handled {
		t.Error("Down key should be handled when child is taller than viewport")
	}
	if vp.ty <= tyBefore {
		t.Errorf("vp.ty = %d after Down; want > %d", vp.ty, tyBefore)
	}
}

func TestViewport_Keyboard_Up_ScrollsUp(t *testing.T) {
	vp, _ := makeScrollableViewport()
	vp.handleKey(BuildKey(tcell.KeyDown)) // scroll down first
	tyAfterDown := vp.ty
	vp.handleKey(BuildKey(tcell.KeyUp))
	if vp.ty >= tyAfterDown {
		t.Errorf("vp.ty = %d after Up; want < %d", vp.ty, tyAfterDown)
	}
}

func TestViewport_Keyboard_Up_AtTop_NoOp(t *testing.T) {
	vp, _ := makeScrollableViewport()
	tyBefore := vp.ty
	handled := vp.handleKey(BuildKey(tcell.KeyUp))
	if handled {
		t.Error("Up at top should return false")
	}
	if vp.ty != tyBefore {
		t.Errorf("vp.ty = %d after Up at top; want %d (unchanged)", vp.ty, tyBefore)
	}
}

func TestViewport_Keyboard_Right_ScrollsRight(t *testing.T) {
	vp, _ := makeScrollableViewport()
	txBefore := vp.tx
	handled := vp.handleKey(BuildKey(tcell.KeyRight))
	if !handled {
		t.Error("Right key should be handled when child is wider than viewport")
	}
	if vp.tx <= txBefore {
		t.Errorf("vp.tx = %d after Right; want > %d", vp.tx, txBefore)
	}
}

func TestViewport_Keyboard_Left_ScrollsLeft(t *testing.T) {
	vp, _ := makeScrollableViewport()
	vp.handleKey(BuildKey(tcell.KeyRight))
	txAfterRight := vp.tx
	vp.handleKey(BuildKey(tcell.KeyLeft))
	if vp.tx >= txAfterRight {
		t.Errorf("vp.tx = %d after Left; want < %d", vp.tx, txAfterRight)
	}
}

func TestViewport_Keyboard_Home_ResetsScroll(t *testing.T) {
	vp, _ := makeScrollableViewport()
	vp.handleKey(BuildKey(tcell.KeyDown))
	vp.handleKey(BuildKey(tcell.KeyRight))
	vp.handleKey(BuildKey(tcell.KeyHome))
	if vp.tx != 0 || vp.ty != 0 {
		t.Errorf("vp offsets = (%d,%d) after Home; want (0,0)", vp.tx, vp.ty)
	}
}

func TestViewport_Keyboard_End_ScrollsToMax(t *testing.T) {
	vp, _ := makeScrollableViewport()
	tyBefore := vp.ty
	vp.handleKey(BuildKey(tcell.KeyEnd))
	if vp.ty <= tyBefore {
		t.Errorf("vp.ty = %d after End; want > %d (at maximum scroll)", vp.ty, tyBefore)
	}
}

func TestViewport_Keyboard_NoChild_ReturnsFalse(t *testing.T) {
	vp := NewViewport("v", "", "")
	handled := vp.handleKey(BuildKey(tcell.KeyDown))
	if handled {
		t.Error("key should not be handled when no child is set")
	}
}

// ── FlagVertical restricts horizontal scrolling ───────────────────────────────

func TestViewport_FlagVertical_BlocksHorizontalScroll(t *testing.T) {
	vp, _ := makeScrollableViewport()
	vp.SetFlag(FlagVertical, true)
	handled := vp.handleKey(BuildKey(tcell.KeyRight))
	if handled {
		t.Error("Right should not be handled when FlagVertical is set")
	}
}

// ── FlagHorizontal restricts vertical scrolling ───────────────────────────────

func TestViewport_FlagHorizontal_BlocksVerticalScroll(t *testing.T) {
	vp, _ := makeScrollableViewport()
	vp.SetFlag(FlagHorizontal, true)
	handled := vp.handleKey(BuildKey(tcell.KeyDown))
	if handled {
		t.Error("Down should not be handled when FlagHorizontal is set")
	}
}

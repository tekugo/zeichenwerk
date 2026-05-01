package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
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

func TestViewport_Keyboard_Down_ScrollsDown(t *testing.T) {
	vp, child := makeScrollableViewport()
	_, yBefore, _, _ := child.Bounds()
	handled := vp.handleKey(BuildKey(tcell.KeyDown))
	if !handled {
		t.Error("Down key should be handled when child is taller than viewport")
	}
	_, yAfter, _, _ := child.Bounds()
	if yAfter >= yBefore {
		t.Errorf("child y = %d after Down; want < %d (scrolled up)", yAfter, yBefore)
	}
}

func TestViewport_Keyboard_Up_ScrollsUp(t *testing.T) {
	vp, child := makeScrollableViewport()
	vp.handleKey(BuildKey(tcell.KeyDown)) // scroll down first
	_, yAfterDown, _, _ := child.Bounds()
	vp.handleKey(BuildKey(tcell.KeyUp))
	_, yAfterUp, _, _ := child.Bounds()
	if yAfterUp <= yAfterDown {
		t.Errorf("child y = %d after Up; want > %d (scrolled back)", yAfterUp, yAfterDown)
	}
}

func TestViewport_Keyboard_Up_AtTop_NoOp(t *testing.T) {
	vp, child := makeScrollableViewport()
	_, yBefore, _, _ := child.Bounds()
	handled := vp.handleKey(BuildKey(tcell.KeyUp))
	_, yAfter, _, _ := child.Bounds()
	if handled {
		t.Error("Up at top should return false")
	}
	if yAfter != yBefore {
		t.Error("child position should not change when Up at top")
	}
}

func TestViewport_Keyboard_Right_ScrollsRight(t *testing.T) {
	vp, child := makeScrollableViewport()
	xBefore, _, _, _ := child.Bounds()
	handled := vp.handleKey(BuildKey(tcell.KeyRight))
	if !handled {
		t.Error("Right key should be handled when child is wider than viewport")
	}
	xAfter, _, _, _ := child.Bounds()
	if xAfter >= xBefore {
		t.Errorf("child x = %d after Right; want < %d (scrolled left)", xAfter, xBefore)
	}
}

func TestViewport_Keyboard_Left_ScrollsLeft(t *testing.T) {
	vp, child := makeScrollableViewport()
	vp.handleKey(BuildKey(tcell.KeyRight))
	xAfterRight, _, _, _ := child.Bounds()
	vp.handleKey(BuildKey(tcell.KeyLeft))
	xAfterLeft, _, _, _ := child.Bounds()
	if xAfterLeft <= xAfterRight {
		t.Errorf("child x = %d after Left; want > %d (scrolled back)", xAfterLeft, xAfterRight)
	}
}

func TestViewport_Keyboard_Home_ResetsScroll(t *testing.T) {
	vp, child := makeScrollableViewport()
	vp.handleKey(BuildKey(tcell.KeyDown))
	vp.handleKey(BuildKey(tcell.KeyRight))
	vp.handleKey(BuildKey(tcell.KeyHome))
	x, y, _, _ := child.Bounds()
	if x != 0 || y != 0 {
		t.Errorf("child position = (%d,%d) after Home; want (0,0)", x, y)
	}
}

func TestViewport_Keyboard_End_ScrollsToMax(t *testing.T) {
	vp, child := makeScrollableViewport()
	_, yBefore, _, _ := child.Bounds()
	vp.handleKey(BuildKey(tcell.KeyEnd))
	_, yAfter, _, _ := child.Bounds()
	if yAfter >= yBefore {
		t.Errorf("child y = %d after End; want < %d (at maximum scroll)", yAfter, yBefore)
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

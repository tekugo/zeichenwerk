package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestText_Defaults(t *testing.T) {
	txt := NewText("t", "", nil, false, 0)
	if !txt.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
}

func TestText_InitialContent(t *testing.T) {
	txt := NewText("t", "", []string{"a", "b", "c"}, false, 0)
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.SetBounds(0, 0, 10, 5)
	txt.Render(r)
	if cs.Get(0, 0) != "a" {
		t.Errorf("first rendered char = %q; want %q", cs.Get(0, 0), "a")
	}
}

// ── Add ───────────────────────────────────────────────────────────────────────

func TestText_Add_ContentAppearsInRender(t *testing.T) {
	txt := NewText("t", "", nil, false, 0)
	txt.Add("Hello")
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.SetBounds(0, 0, 20, 5)
	txt.Render(r)
	if cs.Get(0, 0) != "H" {
		t.Errorf("rendered col 0 = %q; want %q", cs.Get(0, 0), "H")
	}
}

func TestText_Add_MultipleLines(t *testing.T) {
	txt := NewText("t", "", nil, false, 0)
	txt.Add("A", "B", "C")
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.SetBounds(0, 0, 10, 5)
	txt.Render(r)
	if cs.Get(0, 1) != "B" {
		t.Errorf("row 1 = %q; want %q", cs.Get(0, 1), "B")
	}
}

func TestText_Add_MaxRotatesOldLines(t *testing.T) {
	txt := NewText("t", "", nil, false, 3)
	txt.Add("L1", "L2", "L3", "L4", "L5")
	// Only last 3 should be retained
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.SetBounds(0, 0, 10, 5)
	txt.Render(r)
	// Row 0 should be "L3", not "L1"
	if cs.Get(0, 0) != "L" || cs.Get(1, 0) != "3" {
		got := cs.Get(0, 0) + cs.Get(1, 0)
		t.Errorf("row 0 = %q; want %q (rotation kept last 3)", got, "L3")
	}
}

// ── Clear ─────────────────────────────────────────────────────────────────────

func TestText_Clear_RemovesContent(t *testing.T) {
	txt := NewText("t", "", []string{"A", "B"}, false, 0)
	txt.Clear()
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.SetBounds(0, 0, 10, 5)
	txt.Render(r)
	// All cells should be empty after clear
	if cs.Get(0, 0) != "" {
		t.Errorf("after Clear, row 0 col 0 = %q; want empty", cs.Get(0, 0))
	}
}

// ── Set ───────────────────────────────────────────────────────────────────────

func TestText_Set_ReplacesContent(t *testing.T) {
	txt := NewText("t", "", []string{"old"}, false, 0)
	txt.Set([]string{"new1", "new2"})
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.SetBounds(0, 0, 10, 5)
	txt.Render(r)
	if cs.Get(0, 0) != "n" {
		t.Errorf("row 0 col 0 = %q after Set; want %q", cs.Get(0, 0), "n")
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

// makeScrollableText returns a text widget with more lines than its height.
func makeScrollableText() *Text {
	txt := NewText("t", "", []string{"L0", "L1", "L2", "L3", "L4"}, false, 0)
	txt.SetBounds(0, 0, 10, 3) // 5 lines, height 3 → can scroll
	return txt
}

func TestText_Keyboard_Down_Scrolls(t *testing.T) {
	txt := makeScrollableText()
	handled := txt.handleKey(BuildKey(tcell.KeyDown))
	if !handled {
		t.Error("Down should be handled when content exceeds height")
	}
	// Verify via render: row 0 should now show "L1" instead of "L0"
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.Render(r)
	if cs.Get(0, 0) != "L" || cs.Get(1, 0) != "1" {
		got := cs.Get(0, 0) + cs.Get(1, 0)
		t.Errorf("after Down, row 0 = %q; want %q", got, "L1")
	}
}

func TestText_Keyboard_Down_AtBottom_NoOp(t *testing.T) {
	txt := makeScrollableText()
	txt.handleKey(BuildKey(tcell.KeyDown))            // 0→1
	txt.handleKey(BuildKey(tcell.KeyDown))            // 1→2 (maxOffsetY=2)
	handled := txt.handleKey(BuildKey(tcell.KeyDown)) // already at bottom
	if handled {
		t.Error("Down at bottom should return false")
	}
}

func TestText_Keyboard_Up_AtTop_NoOp(t *testing.T) {
	txt := makeScrollableText()
	handled := txt.handleKey(BuildKey(tcell.KeyUp))
	if handled {
		t.Error("Up at top should return false")
	}
}

func TestText_Keyboard_Up_Scrolls(t *testing.T) {
	txt := makeScrollableText()
	txt.handleKey(BuildKey(tcell.KeyDown))          // scroll to 1
	handled := txt.handleKey(BuildKey(tcell.KeyUp)) // back to 0
	if !handled {
		t.Error("Up should be handled when scrolled down")
	}
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.Render(r)
	if cs.Get(0, 0) != "L" || cs.Get(1, 0) != "0" {
		got := cs.Get(0, 0) + cs.Get(1, 0)
		t.Errorf("after Down+Up, row 0 = %q; want %q", got, "L0")
	}
}

func TestText_Keyboard_Home_ResetsScroll(t *testing.T) {
	txt := makeScrollableText()
	txt.handleKey(BuildKey(tcell.KeyDown))
	txt.handleKey(BuildKey(tcell.KeyDown))
	handled := txt.handleKey(BuildKey(tcell.KeyHome))
	if !handled {
		t.Error("Home should be handled when scrolled")
	}
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.Render(r)
	if cs.Get(0, 0) != "L" || cs.Get(1, 0) != "0" {
		got := cs.Get(0, 0) + cs.Get(1, 0)
		t.Errorf("after Home, row 0 = %q; want %q", got, "L0")
	}
}

func TestText_Keyboard_End_ScrollsToBottom(t *testing.T) {
	txt := makeScrollableText()
	handled := txt.handleKey(BuildKey(tcell.KeyEnd))
	if !handled {
		t.Error("End should be handled when not already at bottom")
	}
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.Render(r)
	// At bottom (offsetY=2), row 0 shows "L2"
	if cs.Get(0, 0) != "L" || cs.Get(1, 0) != "2" {
		got := cs.Get(0, 0) + cs.Get(1, 0)
		t.Errorf("after End, row 0 = %q; want %q", got, "L2")
	}
}

func TestText_Keyboard_PageDown_Jumps(t *testing.T) {
	txt := NewText("t", "", []string{"L0", "L1", "L2", "L3", "L4", "L5", "L6"}, false, 0)
	txt.SetBounds(0, 0, 10, 3) // height=3
	handled := txt.handleKey(BuildKey(tcell.KeyPgDn))
	if !handled {
		t.Error("PageDown should be handled")
	}
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.Render(r)
	// Jumped by 3: offsetY=3, row 0 shows "L3"
	if cs.Get(0, 0) != "L" || cs.Get(1, 0) != "3" {
		got := cs.Get(0, 0) + cs.Get(1, 0)
		t.Errorf("after PageDown, row 0 = %q; want %q", got, "L3")
	}
}

func TestText_Keyboard_Right_ScrollsHorizontal(t *testing.T) {
	txt := NewText("t", "", nil, false, 0)
	txt.SetBounds(0, 0, 5, 3)
	txt.Add("abcdefghij") // Add triggers adjust() which computes longest=10
	handled := txt.handleKey(BuildKey(tcell.KeyRight))
	if !handled {
		t.Error("Right should be handled when content wider than view")
	}
}

func TestText_Keyboard_Left_AtStart_NoOp(t *testing.T) {
	txt := makeScrollableText()
	handled := txt.handleKey(BuildKey(tcell.KeyLeft))
	if handled {
		t.Error("Left at offsetX=0 should return false")
	}
}

// ── Follow mode ───────────────────────────────────────────────────────────────

func TestText_Follow_ScrollsToBottomOnAdd(t *testing.T) {
	txt := NewText("t", "", nil, true, 0)
	txt.SetBounds(0, 0, 10, 3)
	txt.Add("L0", "L1", "L2", "L3", "L4")
	// follow=true, not focused → offsetY = max(5-3, 0) = 2
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	txt.Render(r)
	// Row 0 should show "L2" (offset=2)
	if cs.Get(0, 0) != "L" || cs.Get(1, 0) != "2" {
		got := cs.Get(0, 0) + cs.Get(1, 0)
		t.Errorf("follow mode row 0 = %q; want %q (scrolled to bottom)", got, "L2")
	}
}

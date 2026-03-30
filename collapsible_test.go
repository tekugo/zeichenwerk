package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

// Ensure Collapsible implements the Container interface.
var _ Container = (*Collapsible)(nil)

func TestNewCollapsible(t *testing.T) {
	c := NewCollapsible("col1", "", "Details", false)

	if c.ID() != "col1" {
		t.Errorf("ID() = %q; want %q", c.ID(), "col1")
	}
	if c.title != "Details" {
		t.Errorf("title = %q; want %q", c.title, "Details")
	}
	if c.expanded {
		t.Error("expected expanded = false")
	}
	if !c.Flag(FlagFocusable) {
		t.Error("expected FlagFocusable to be set")
	}
}

func TestCollapsible_Hint_CollapsedNoChild(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	w, h := c.Hint()
	if w != 0 {
		t.Errorf("Hint() width = %d; want 0", w)
	}
	if h != 1 {
		t.Errorf("Hint() height = %d; want 1 (header only)", h)
	}
}

func TestCollapsible_Hint_CollapsedWithChild(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	child := NewComponent("child", "")
	child.SetHint(30, 10)
	c.Add(child)

	w, h := c.Hint()
	if w != 30 {
		t.Errorf("Hint() width = %d; want 30", w)
	}
	if h != 1 {
		t.Errorf("Hint() height = %d; want 1 (collapsed)", h)
	}
}

func TestCollapsible_Hint_ExpandedWithChild(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	child := NewComponent("child", "")
	child.SetHint(30, 10)
	c.Add(child)

	w, h := c.Hint()
	if w != 30 {
		t.Errorf("Hint() width = %d; want 30", w)
	}
	if h != 11 {
		t.Errorf("Hint() height = %d; want 11 (1 header + 10 child)", h)
	}
}

func TestCollapsible_Hint_ExpandedNoChildHint(t *testing.T) {
	// When expanded and child has no height hint (e.g. List), return -1 so the
	// parent Flex allocates remaining space rather than giving the collapsible
	// zero body height.
	c := NewCollapsible("c", "", "Title", true)
	child := NewComponent("child", "") // hint (0,0)
	c.Add(child)

	_, h := c.Hint()
	if h != -1 {
		t.Errorf("Hint() height = %d; want -1 (fractional) when child has no height hint", h)
	}
}

func TestCollapsible_Add_SetsParent(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	child := NewComponent("child", "")
	c.Add(child)
	if child.Parent() != c {
		t.Error("Add() should set child parent to the collapsible")
	}
}

func TestCollapsible_Add_HidesChildWhenCollapsed(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	child := NewComponent("child", "")
	c.Add(child)
	if !child.Flag(FlagHidden) {
		t.Error("child should be hidden when collapsible is collapsed")
	}
}

func TestCollapsible_Add_ShowsChildWhenExpanded(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	child := NewComponent("child", "")
	c.Add(child)
	if child.Flag(FlagHidden) {
		t.Error("child should be visible when collapsible is expanded")
	}
}

func TestCollapsible_Add_ReplacesPreviousChild(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	child1 := NewComponent("c1", "")
	child2 := NewComponent("c2", "")
	c.Add(child1)
	c.Add(child2)

	if c.child != child2 {
		t.Error("second Add() should replace the first child")
	}
	if child1.Parent() != nil {
		t.Error("replaced child's parent should be cleared")
	}
}

func TestCollapsible_Children_Empty(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	if len(c.Children()) != 0 {
		t.Errorf("Children() = %d; want 0 for no child", len(c.Children()))
	}
}

func TestCollapsible_Children_WithChild(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	child := NewComponent("child", "")
	c.Add(child)
	children := c.Children()
	if len(children) != 1 {
		t.Errorf("Children() = %d; want 1", len(children))
	}
	if children[0] != child {
		t.Error("Children()[0] should be the added child")
	}
}

func TestCollapsible_Toggle_FlagHidden(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	child := NewComponent("child", "")
	c.Add(child)

	if !child.Flag(FlagHidden) {
		t.Error("child should start hidden (collapsed)")
	}
	c.Toggle() // expand
	if child.Flag(FlagHidden) {
		t.Error("child should be visible after Toggle() (expanded)")
	}
	c.Toggle() // collapse
	if !child.Flag(FlagHidden) {
		t.Error("child should be hidden after second Toggle() (collapsed)")
	}
}

func TestCollapsible_EvtChange_Expand(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	var received bool
	var gotData bool
	c.On(EvtChange, func(w Widget, e Event, data ...any) bool {
		if len(data) > 0 {
			if v, ok := data[0].(bool); ok {
				gotData = true
				received = v
			}
		}
		return true
	})
	c.Toggle() // Expand
	if !gotData {
		t.Error("EvtChange should carry bool data")
	}
	if !received {
		t.Error("EvtChange data should be true after expanding")
	}
}

func TestCollapsible_EvtChange_Collapse(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	var received bool
	c.On(EvtChange, func(w Widget, e Event, data ...any) bool {
		if len(data) > 0 {
			if v, ok := data[0].(bool); ok {
				received = v
			}
		}
		return true
	})
	c.Toggle() // Collapse
	if received {
		t.Error("EvtChange data should be false after collapsing")
	}
}

func TestCollapsible_KeyEnter_Toggles(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	ev := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
	c.handleKey(c, ev)
	if !c.Expanded() {
		t.Error("Enter key should expand a collapsed collapsible")
	}
}

func TestCollapsible_KeySpace_Toggles(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	ev := tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone)
	c.handleKey(c, ev)
	if !c.Expanded() {
		t.Error("Space key should expand a collapsed collapsible")
	}
}

func TestCollapsible_KeyRight_Expands(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	ev := tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone)
	c.handleKey(c, ev)
	if !c.Expanded() {
		t.Error("→ should expand a collapsed collapsible")
	}
}

func TestCollapsible_KeyLeft_Collapses(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	ev := tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone)
	c.handleKey(c, ev)
	if c.Expanded() {
		t.Error("← should collapse an expanded collapsible")
	}
}

func TestCollapsible_KeyRight_NoopIfExpanded(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	ev := tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone)
	handled := c.handleKey(c, ev)
	if !handled {
		t.Error("→ should return true (consumed) even when already expanded")
	}
	if !c.Expanded() {
		t.Error("→ should not change state when already expanded")
	}
}

func TestCollapsible_KeyLeft_NoopIfCollapsed(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	ev := tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone)
	handled := c.handleKey(c, ev)
	if !handled {
		t.Error("← should return true (consumed) even when already collapsed")
	}
	if c.Expanded() {
		t.Error("← should not change state when already collapsed")
	}
}

func TestCollapsible_Layout_ChildPosition(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	c.SetBounds(0, 0, 40, 20)
	child := NewComponent("child", "")
	child.SetHint(40, 10)
	c.Add(child)
	c.Layout()

	_, cy, cw, ch := child.Bounds()
	// Content area (no style) starts at (0,0,40,20); header takes row 0,
	// child starts at row 1 with height 19.
	if cy != 1 {
		t.Errorf("child y = %d; want 1 (below header row)", cy)
	}
	if ch != 19 {
		t.Errorf("child height = %d; want 19", ch)
	}
	if cw != 40 {
		t.Errorf("child width = %d; want 40", cw)
	}
}

func TestCollapsible_Layout_CollapsedDoesNotMoveChild(t *testing.T) {
	c := NewCollapsible("c", "", "Title", false)
	c.SetBounds(0, 0, 40, 20)
	child := NewComponent("child", "")
	child.SetHint(40, 10)
	c.Add(child)
	c.Layout()

	// When collapsed, bounds are unchanged from initial zero values.
	_, cy, _, _ := child.Bounds()
	if cy != 0 {
		t.Errorf("collapsed child y = %d; want 0 (bounds not updated)", cy)
	}
}

func TestCollapsible_StateRestored(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	child := NewComponent("child", "")
	c.Add(child)

	c.Collapse()
	if c.Expanded() {
		t.Error("expected collapsed after Collapse()")
	}
	if !child.Flag(FlagHidden) {
		t.Error("child should be hidden after Collapse()")
	}

	c.Expand()
	if !c.Expanded() {
		t.Error("expected expanded after Expand()")
	}
	if child.Flag(FlagHidden) {
		t.Error("child should be visible after re-Expand()")
	}
}

func TestFocusedIn_DirectFocus(t *testing.T) {
	w := NewComponent("w", "")
	w.SetFlag(FlagFocused, true)
	if !focusedIn(w) {
		t.Error("focusedIn should return true when widget itself is focused")
	}
}

func TestFocusedIn_NoFocus(t *testing.T) {
	w := NewComponent("w", "")
	if focusedIn(w) {
		t.Error("focusedIn should return false when nothing is focused")
	}
}

func TestFocusedIn_NestedFocus(t *testing.T) {
	inner := NewFlex("inner", "", false, "stretch", 0)
	nested := NewComponent("nested", "")
	nested.SetFlag(FlagFocused, true)
	inner.Add(nested)
	if !focusedIn(inner) {
		t.Error("focusedIn should return true when a descendant is focused")
	}
}

func TestCollapsible_Collapse_HidesChildWhenFocused(t *testing.T) {
	// Without a real UI, ui.Focus cannot clear flags, but Collapse must not
	// panic and must still hide the child.
	c := NewCollapsible("c", "", "Title", true)
	child := NewComponent("child", "")
	child.SetFlag(FlagFocusable, true)
	child.SetFlag(FlagFocused, true)
	c.Add(child)

	c.Collapse() // must not panic

	if !child.Flag(FlagHidden) {
		t.Error("child should be hidden after Collapse()")
	}
}

func TestCollapsible_Expanded(t *testing.T) {
	c := NewCollapsible("c", "", "Title", true)
	if !c.Expanded() {
		t.Error("Expanded() should return true for initially-expanded collapsible")
	}
	c.Collapse()
	if c.Expanded() {
		t.Error("Expanded() should return false after Collapse()")
	}
}

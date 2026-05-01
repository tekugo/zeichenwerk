package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ---- helpers ---------------------------------------------------------------

func newTree(nodes ...*TreeNode) *Tree {
	t := NewTree("t", "")
	for _, n := range nodes {
		t.root.Add(n)
	}
	t.rebuild()
	return t
}

// ---- TreeNode --------------------------------------------------------------

func TestTreeNode_Defaults(t *testing.T) {
	n := NewTreeNode("hello")
	if n.Text() != "hello" {
		t.Fatalf("got %q want %q", n.Text(), "hello")
	}
	if n.Data() != nil {
		t.Fatal("expected nil data")
	}
	if !n.Leaf() {
		t.Fatal("expected leaf")
	}
	if n.Expanded() {
		t.Fatal("should start collapsed")
	}
	if n.Disabled() {
		t.Fatal("should start enabled")
	}
}

func TestTreeNode_Data(t *testing.T) {
	n := NewTreeNode("x", 42)
	if n.Data() != 42 {
		t.Fatalf("got %v want 42", n.Data())
	}
}

func TestTreeNode_AddChain(t *testing.T) {
	root := NewTreeNode("root").
		Add(NewTreeNode("a")).
		Add(NewTreeNode("b"))
	if len(root.Children()) != 2 {
		t.Fatalf("got %d children", len(root.Children()))
	}
	if root.Leaf() {
		t.Fatal("should not be a leaf")
	}
}

func TestTreeNode_Toggle(t *testing.T) {
	n := NewTreeNode("x")
	n.Toggle()
	if !n.Expanded() {
		t.Fatal("should be expanded after toggle")
	}
	n.Toggle()
	if n.Expanded() {
		t.Fatal("should be collapsed after second toggle")
	}
}

// ---- rebuild / flatten -----------------------------------------------------

func TestRebuild_SingleLevel(t *testing.T) {
	tr := newTree(
		NewTreeNode("a"),
		NewTreeNode("b"),
		NewTreeNode("c"),
	)
	if len(tr.flat) != 3 {
		t.Fatalf("got %d flat items", len(tr.flat))
	}
	for i, item := range tr.flat {
		if item.depth != 0 {
			t.Errorf("item %d: expected depth 0, got %d", i, item.depth)
		}
	}
	if !tr.flat[2].isLast {
		t.Error("last top-level item should have isLast=true")
	}
	if tr.flat[0].isLast {
		t.Error("first item should not have isLast")
	}
}

func TestRebuild_CollapsedChildrenHidden(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child1"))
	parent.Add(NewTreeNode("child2"))
	tr := newTree(parent)
	if len(tr.flat) != 1 {
		t.Fatalf("collapsed parent should show only itself, got %d", len(tr.flat))
	}
}

func TestRebuild_ExpandedChildrenVisible(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child1"))
	parent.Add(NewTreeNode("child2"))
	parent.Expand()
	tr := newTree(parent)
	if len(tr.flat) != 3 {
		t.Fatalf("expected 3 flat items, got %d", len(tr.flat))
	}
	if tr.flat[1].depth != 1 || tr.flat[2].depth != 1 {
		t.Error("children should have depth 1")
	}
}

func TestRebuild_Trunk(t *testing.T) {
	// a (not last)
	//   a1 (last)
	// b (last)
	//   b1 (last)
	a := NewTreeNode("a")
	a.Add(NewTreeNode("a1"))
	a.Expand()
	b := NewTreeNode("b")
	b.Add(NewTreeNode("b1"))
	b.Expand()
	tr := newTree(a, b)

	// flat order: a, a1, b, b1
	if len(tr.flat) != 4 {
		t.Fatalf("expected 4 flat items, got %d", len(tr.flat))
	}

	// a: depth=0, isLast=false
	if tr.flat[0].isLast {
		t.Error("a should not be isLast")
	}
	// a1: depth=1, isLast=true, trunk[0]=true (a is not last)
	if !tr.flat[1].isLast {
		t.Error("a1 should be isLast")
	}
	if len(tr.flat[1].trunk) < 1 || !tr.flat[1].trunk[0] {
		t.Error("a1 trunk[0] should be true because a has more siblings")
	}
	// b: depth=0, isLast=true
	if !tr.flat[2].isLast {
		t.Error("b should be isLast")
	}
	// b1: depth=1, isLast=true, trunk[0]=false (b is last)
	if len(tr.flat[3].trunk) < 1 || tr.flat[3].trunk[0] {
		t.Error("b1 trunk[0] should be false because b is last")
	}
}

// ---- navigation ------------------------------------------------------------

func TestTree_MoveDownUp(t *testing.T) {
	tr := newTree(NewTreeNode("a"), NewTreeNode("b"), NewTreeNode("c"))
	tr.index = 0
	tr.Move(1)
	if tr.index != 1 {
		t.Fatalf("got index %d want 1", tr.index)
	}
	tr.Move(-1)
	if tr.index != 0 {
		t.Fatalf("got index %d want 0", tr.index)
	}
}

func TestTree_FirstLast(t *testing.T) {
	tr := newTree(NewTreeNode("a"), NewTreeNode("b"), NewTreeNode("c"))
	tr.Last()
	if tr.index != 2 {
		t.Fatalf("Last(): got %d", tr.index)
	}
	tr.First()
	if tr.index != 0 {
		t.Fatalf("First(): got %d", tr.index)
	}
}

func TestTree_Move_SkipsDisabled(t *testing.T) {
	a := NewTreeNode("a")
	b := NewTreeNode("b")
	b.SetDisabled(true)
	c := NewTreeNode("c")
	tr := newTree(a, b, c)
	tr.index = 0
	tr.Move(1)
	if tr.index != 2 {
		t.Fatalf("expected skip disabled b, got index %d", tr.index)
	}
}

func TestTree_Select(t *testing.T) {
	b := NewTreeNode("b")
	tr := newTree(NewTreeNode("a"), b)
	tr.Select(b)
	if tr.index != 1 {
		t.Fatalf("Select: got index %d", tr.index)
	}
}

func TestTree_Selected_Nil(t *testing.T) {
	tr := newTree()
	if tr.Selected() != nil {
		t.Fatal("empty tree Selected() should be nil")
	}
}

// ---- keyboard: right / left ------------------------------------------------

func TestTree_RightExpands(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child"))
	tr := newTree(parent)
	tr.index = 0

	tr.handleKey(BuildKey(tcell.KeyRight))

	if !parent.Expanded() {
		t.Fatal("right on collapsed node should expand it")
	}
}

func TestTree_RightMovesToFirstChild(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child"))
	parent.Expand()
	tr := newTree(parent)
	tr.index = 0

	tr.handleKey(BuildKey(tcell.KeyRight))

	if tr.index != 1 {
		t.Fatalf("right on expanded node should move to first child, got %d", tr.index)
	}
}

func TestTree_LeftCollapses(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child"))
	parent.Expand()
	tr := newTree(parent)
	tr.index = 0

	var changed *TreeNode
	tr.On(EvtChange, func(_ Widget, _ Event, params ...any) bool {
		if len(params) > 0 {
			changed, _ = params[0].(*TreeNode)
		}
		return true
	})

	tr.handleKey(BuildKey(tcell.KeyLeft))

	if parent.Expanded() {
		t.Fatal("left on expanded node should collapse it")
	}
	if changed != parent {
		t.Fatal("EvtChange should carry toggled node")
	}
}

func TestTree_LeftMovesToParent(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child"))
	parent.Expand()
	tr := newTree(parent)
	tr.index = 1 // child

	tr.handleKey(BuildKey(tcell.KeyLeft))

	if tr.index != 0 {
		t.Fatalf("left on non-expanded child should move to parent, got %d", tr.index)
	}
}

// ---- keyboard: enter / space -----------------------------------------------

func TestTree_EnterTogglesAndActivates(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child"))
	tr := newTree(parent)
	tr.index = 0

	var activated *TreeNode
	tr.On(EvtActivate, func(_ Widget, _ Event, params ...any) bool {
		if len(params) > 0 {
			activated, _ = params[0].(*TreeNode)
		}
		return true
	})

	tr.handleKey(BuildKey(tcell.KeyEnter))

	if !parent.Expanded() {
		t.Fatal("enter should expand a collapsed node")
	}
	if activated != parent {
		t.Fatal("enter should dispatch EvtActivate with the node")
	}
}

func TestTree_SpaceToggles(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child"))
	tr := newTree(parent)
	tr.index = 0

	tr.handleKey(BuildRune(" "))

	if !parent.Expanded() {
		t.Fatal("space should expand a collapsed node")
	}
}

// ---- EvtChange dispatch ----------------------------------------------------

func TestTree_ExpandDispatchesChange(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child"))
	tr := newTree(parent)

	var changed *TreeNode
	tr.On(EvtChange, func(_ Widget, _ Event, params ...any) bool {
		if len(params) > 0 {
			changed, _ = params[0].(*TreeNode)
		}
		return true
	})

	tr.Expand(parent)

	if changed != parent {
		t.Fatal("Expand should dispatch EvtChange with the expanded node")
	}
}

func TestTree_CollapseDispatchesChange(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("child"))
	parent.Expand()
	tr := newTree(parent)

	var changed *TreeNode
	tr.On(EvtChange, func(_ Widget, _ Event, params ...any) bool {
		if len(params) > 0 {
			changed, _ = params[0].(*TreeNode)
		}
		return true
	})

	tr.Collapse(parent)

	if changed != parent {
		t.Fatal("Collapse should dispatch EvtChange with the collapsed node")
	}
}

// ---- ExpandAll / CollapseAll -----------------------------------------------

func TestTree_CollapseAllExpandAll(t *testing.T) {
	a := NewTreeNode("a")
	a.Add(NewTreeNode("a1"))
	a.Add(NewTreeNode("a2"))
	b := NewTreeNode("b")
	b.Add(NewTreeNode("b1"))
	a.Expand()
	b.Expand()
	tr := newTree(a, b)

	initialLen := len(tr.flat)
	tr.CollapseAll()
	if len(tr.flat) != 2 {
		t.Fatalf("CollapseAll: expected 2 top-level items, got %d", len(tr.flat))
	}
	tr.ExpandAll()
	if len(tr.flat) != initialLen {
		t.Fatalf("ExpandAll: expected %d items, got %d", initialLen, len(tr.flat))
	}
}

// ---- EvtSelect dispatch ----------------------------------------------------

func TestTree_MoveDispatchesSelect(t *testing.T) {
	tr := newTree(NewTreeNode("a"), NewTreeNode("b"))
	tr.index = 0

	var selected *TreeNode
	tr.On(EvtSelect, func(_ Widget, _ Event, params ...any) bool {
		if len(params) > 0 {
			selected, _ = params[0].(*TreeNode)
		}
		return true
	})

	tr.Move(1)

	if selected == nil || selected.Text() != "b" {
		t.Fatal("Move should dispatch EvtSelect with the newly highlighted node")
	}
}

// ---- Lazy loading ----------------------------------------------------------

func TestLazyTreeNode_IsNotLeaf(t *testing.T) {
	n := NewLazyTreeNode("dir", func(_ *TreeNode) {})
	if n.Leaf() {
		t.Fatal("node with loader should not be a leaf")
	}
}

func TestLazyTreeNode_LoadedOnExpand(t *testing.T) {
	calls := 0
	loader := func(n *TreeNode) {
		calls++
		n.Add(NewTreeNode("child"))
	}
	n := NewLazyTreeNode("dir", loader)
	tr := newTree(n)
	tr.index = 0

	tr.Expand(n)

	if calls != 1 {
		t.Fatalf("loader should be called once, got %d", calls)
	}
	if len(n.Children()) != 1 {
		t.Fatal("child should have been added by loader")
	}
}

func TestLazyTreeNode_LoaderNotCalledAgain(t *testing.T) {
	calls := 0
	loader := func(n *TreeNode) {
		calls++
		n.Add(NewTreeNode("child"))
	}
	n := NewLazyTreeNode("dir", loader)
	tr := newTree(n)

	tr.Expand(n)
	tr.Collapse(n)
	tr.Expand(n)

	if calls != 1 {
		t.Fatalf("loader should only be called once, got %d", calls)
	}
}

func TestLazyTreeNode_BecomesLeafWhenLoaderAddsNoChildren(t *testing.T) {
	n := NewLazyTreeNode("empty", func(_ *TreeNode) {}) // loader adds nothing
	tr := newTree(n)
	tr.index = 0

	tr.Expand(n)

	if !n.Leaf() {
		t.Fatal("after loading with no children, node should be a leaf")
	}
}

func TestLazyTreeNode_SetLoaderResetsState(t *testing.T) {
	calls := 0
	n := NewLazyTreeNode("dir", func(node *TreeNode) {
		calls++
		node.Add(NewTreeNode("child"))
	})
	tr := newTree(n)

	tr.Expand(n) // first load
	tr.Collapse(n)

	// Attach a new loader — should fire again on next expand
	n.SetLoader(func(node *TreeNode) {
		calls++
		node.Add(NewTreeNode("extra"))
	})
	tr.Expand(n)

	if calls != 2 {
		t.Fatalf("SetLoader should re-enable loading; got %d calls", calls)
	}
}

// ---- Render (smoke test) ---------------------------------------------------

func TestTree_RenderNoPanel(t *testing.T) {
	a := NewTreeNode("alpha")
	a.Add(NewTreeNode("beta"))
	a.Expand()
	tr := newTree(a)
	tr.index = 0
	tr.SetBounds(0, 0, 40, 10)
	tr.Apply(NewTheme())
	// Must not panic
	tr.Render(newTestRenderer())
}

package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

// makeWidgetTree returns a small hierarchy used by the tree-widgets tests:
//
//	Flex(#root) [horizontal]
//	├── List(#left)
//	└── Flex(#right) [vertical]
//	    ├── Static(#title)
//	    └── Button (no id)
func makeWidgetTree() *Flex {
	root := NewFlex("root", "", Stretch, 0)
	root.Add(NewList("left", "", []string{"a", "b"}))
	right := NewFlex("right", "", Stretch, 0)
	right.SetFlag(FlagVertical, true)
	right.Add(NewStatic("title", "", "hello"))
	right.Add(NewButton("", "", "ok"))
	root.Add(right)
	return root
}

func TestTreeWidgets_RootLabel(t *testing.T) {
	tw := NewTreeWidgets("widget-tree", "", makeWidgetTree())
	if got := len(tw.flat); got != 1 {
		t.Fatalf("expected 1 visible top-level node, got %d", got)
	}
	rootNode := tw.flat[0].node
	if rootNode.Text() != "Flex (#root)" {
		t.Fatalf("root label: got %q want %q", rootNode.Text(), "Flex (#root)")
	}
	if rootNode.Leaf() {
		t.Fatal("root node should not be a leaf — it has children")
	}
}

func TestTreeWidgets_DataIsWidget(t *testing.T) {
	root := makeWidgetTree()
	tw := NewTreeWidgets("wt", "", root)
	got, ok := tw.flat[0].node.Data().(Widget)
	if !ok {
		t.Fatalf("node data is not a Widget: %T", tw.flat[0].node.Data())
	}
	if got != root {
		t.Fatalf("node data: got %p want %p", got, root)
	}
}

func TestTreeWidgets_ChildrenMirrorContainer(t *testing.T) {
	tw := NewTreeWidgets("wt", "", makeWidgetTree())
	rootNode := tw.flat[0].node
	children := rootNode.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0].Text() != "List (#left)" {
		t.Errorf("child[0] label: got %q want %q", children[0].Text(), "List (#left)")
	}
	if children[1].Text() != "Flex (#right)" {
		t.Errorf("child[1] label: got %q want %q", children[1].Text(), "Flex (#right)")
	}
	// Drill into "right" Flex and check its grandchildren.
	rightChildren := children[1].Children()
	if len(rightChildren) != 2 {
		t.Fatalf("expected right Flex to have 2 children, got %d", len(rightChildren))
	}
	if rightChildren[0].Text() != "Static (#title)" {
		t.Errorf("grandchild[0] label: got %q want %q", rightChildren[0].Text(), "Static (#title)")
	}
	// Button without an id should render as just "Button".
	if rightChildren[1].Text() != "Button" {
		t.Errorf("grandchild[1] label: got %q want %q", rightChildren[1].Text(), "Button")
	}
}

func TestTreeWidgets_LeafWidgetIsLeafNode(t *testing.T) {
	leaf := NewStatic("only", "", "hi")
	tw := NewTreeWidgets("wt", "", leaf)
	if len(tw.flat) != 1 {
		t.Fatalf("expected 1 top-level node, got %d", len(tw.flat))
	}
	node := tw.flat[0].node
	if node.Text() != "Static (#only)" {
		t.Errorf("label: got %q want %q", node.Text(), "Static (#only)")
	}
	if !node.Leaf() {
		t.Error("expected leaf widget to produce a leaf node")
	}
}

func TestTreeWidgets_Refresh(t *testing.T) {
	root := NewFlex("root", "", Stretch, 0)
	root.Add(NewStatic("a", "", "a"))
	tw := NewTreeWidgets("wt", "", root)
	if got := len(tw.flat[0].node.Children()); got != 1 {
		t.Fatalf("initial children: got %d want 1", got)
	}
	root.Add(NewStatic("b", "", "b"))
	tw.Refresh()
	if got := len(tw.flat[0].node.Children()); got != 2 {
		t.Fatalf("after Refresh: got %d want 2", got)
	}
}

func TestTreeWidgets_NilRoot(t *testing.T) {
	tw := NewTreeWidgets("wt", "", nil)
	if got := len(tw.flat); got != 0 {
		t.Fatalf("nil root: expected 0 nodes, got %d", got)
	}
}

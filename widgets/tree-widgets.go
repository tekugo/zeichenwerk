package widgets

import (
	"fmt"
	"reflect"

	. "github.com/tekugo/zeichenwerk/core"
)

// TreeWidgets is a Tree pre-wired to display a widget hierarchy. Each node
// shows the widget's Go type and id, e.g. "Flex (#main)"; widgets without
// an id render with the type alone, e.g. "List".
//
// The opaque data stored on every TreeNode is the underlying Widget, so
// event handlers can recover the widget directly:
//
//	tw.Tree.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
//	    node := data[0].(*TreeNode)
//	    w := node.Data().(Widget)
//	    ...
//	})
//
// Intended consumers are the Inspector and a future designer application.
type TreeWidgets struct {
	*Tree
	root Widget
}

// NewTreeWidgets creates a TreeWidgets that mirrors the subtree rooted at
// root. The root widget appears as the single top-level node; its children
// (when root is a Container) are added recursively.
func NewTreeWidgets(id, class string, root Widget) *TreeWidgets {
	tw := &TreeWidgets{
		Tree: NewTree(id, class),
		root: root,
	}
	if root != nil {
		tw.Tree.Add(buildTreeWidgetsNode(root))
	}
	return tw
}

// Apply applies theme styles. TreeWidgets uses the "tree-widgets" selector
// family so themes can style it independently of a plain Tree, falling back
// to "tree" through the theme cascade.
func (tw *TreeWidgets) Apply(theme *Theme) {
	theme.Apply(tw.Tree, tw.Tree.Selector("tree-widgets"), "disabled", "focused", "hovered")
	theme.Apply(tw.Tree, tw.Tree.Selector("tree-widgets/highlight"), "focused")
	theme.Apply(tw.Tree, tw.Tree.Selector("tree-widgets/indent"))
}

// Root returns the widget the tree currently mirrors.
func (tw *TreeWidgets) Root() Widget { return tw.root }

// SetRoot replaces the mirrored widget and rebuilds the tree.
func (tw *TreeWidgets) SetRoot(root Widget) {
	tw.root = root
	parent := NewTreeNode("")
	if root != nil {
		parent.Add(buildTreeWidgetsNode(root))
	}
	tw.Tree.SetRoot(parent)
}

// Refresh rebuilds the tree from the current widget hierarchy. Call this
// when the underlying layout has changed (children added or removed).
func (tw *TreeWidgets) Refresh() {
	tw.SetRoot(tw.root)
}

// buildTreeWidgetsNode mirrors the subtree rooted at w as a TreeNode with
// the widget attached as opaque data.
func buildTreeWidgetsNode(w Widget) *TreeNode {
	node := NewTreeNode(formatWidgetLabel(w), w)
	if container, ok := w.(Container); ok {
		for _, child := range container.Children() {
			node.Add(buildTreeWidgetsNode(child))
		}
	}
	return node
}

// formatWidgetLabel returns a node label like "Flex (#main)" when the
// widget has an id, or just "Flex" when it does not.
func formatWidgetLabel(w Widget) string {
	name := widgetTypeName(w)
	if id := w.ID(); id != "" {
		return fmt.Sprintf("%s (#%s)", name, id)
	}
	return name
}

// widgetTypeName returns the unqualified Go type name of a widget, stripping
// pointer and package qualifiers (e.g. "*widgets.Flex" → "Flex").
func widgetTypeName(w Widget) string {
	t := reflect.TypeOf(w)
	if t == nil {
		return "Widget"
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if name := t.Name(); name != "" {
		return name
	}
	return "Widget"
}

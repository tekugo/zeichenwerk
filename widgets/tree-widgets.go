package widgets

import (
	"fmt"

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

// ---- Widget Methods -------------------------------------------------------

// Apply applies theme styles by delegating to the embedded Tree's
// Apply, so TreeWidgets uses the same "tree" selector family that
// every theme registers. A previous version applied a "tree-widgets"
// selector family on top, intending to let themes style TreeWidgets
// independently — but no theme ships those entries today, and the
// theme cascade resolves any unregistered selector to the empty
// default, which silently overwrote the baseline tree styles with
// black-on-black. Until themes opt in with explicit tree-widgets
// entries, TreeWidgets just looks like a Tree, which is the right
// fallback.
func (tw *TreeWidgets) Apply(theme *Theme) {
	tw.Tree.Apply(theme)
}

// Refresh rebuilds the tree from the current widget hierarchy. Call this
// when the underlying layout has changed (children added or removed).
func (tw *TreeWidgets) Refresh() {
	tw.Set(tw.root)
}

// ---- Setter interface -----------------------------------------------------

// Get returns the widget the tree currently mirrors.
func (tw *TreeWidgets) Get() Widget { return tw.root }

// Set replaces the mirrored widget and rebuilds the tree.
func (tw *TreeWidgets) Set(root Widget) {
	tw.root = root
	parent := NewTreeNode("")
	if root != nil {
		parent.Add(buildTreeWidgetsNode(root))
	}
	tw.Tree.SetRoot(parent)
}

// ---- Helper Methods -------------------------------------------------------

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
	name := WidgetType(w)
	if id := w.ID(); id != "" {
		return fmt.Sprintf("%s (#%s)", name, id)
	}
	return name
}

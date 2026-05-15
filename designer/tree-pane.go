package designer

import (
	"fmt"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// refreshTree rebuilds the tree pane from the live target. Called
// after any structural mutation (Add, Remove, Move) so the picker
// and selection reflect the new shape. Selection is snapped back to
// the survivor via selectAfterMutation when applicable.
func (s *session) refreshTree() {
	root := widgets.NewTreeNode("")
	root.Add(buildWidgetTreeNode(s.target))
	s.tree.SetRoot(root)
	widgets.Redraw(s.tree)
}

// onTreeSelect picks up the freshly-selected node and asks the
// details pane to rebuild around it. Bound from popup.wireActions.
func (s *session) onTreeSelect() {
	node := s.tree.Selected()
	if node == nil {
		return
	}
	w, ok := node.Data().(core.Widget)
	if !ok || w == nil {
		return
	}
	s.currentNode = node
	s.rebuildPane(w)
	s.setStatus(fmt.Sprintf("selected %s%s", widgetKind(w), idSuffix(w)))
}

// resolveAddParent chooses the container a freshly-added widget
// lands in. Preference order: the current selection if it is itself
// a container, then the current selection's parent, then the root
// target. This matches the user's intuition that "Add" attaches the
// new widget next to wherever they're focused.
func (s *session) resolveAddParent() core.Container {
	if s.currentWidget != nil {
		if c, ok := s.currentWidget.(core.Container); ok {
			return c
		}
		if p := s.currentWidget.Parent(); p != nil {
			return p
		}
	}
	return s.target
}

// add opens the kind-picker dialog. Selecting a kind appends a
// fresh widget under the resolved parent, applies the theme,
// relayouts, refreshes the tree, and surfaces the result on the
// status line.
func (s *session) add() {
	parent := s.resolveAddParent()
	s.openAddChildDialog(parent, func(child core.Widget) {
		child.Apply(s.theme)
		widgets.Relayout(child)
		s.refreshTree()
		s.setDirty(true)
		s.setStatus(fmt.Sprintf("added %s under %s%s",
			widgetKind(child), widgetKind(parent), idSuffix(parent)))
	})
}

// delete removes the current selection from its parent. The target
// root is non-removable — there is nothing for the selection to
// snap back to and the codegen walker would have nothing to emit.
// After a successful remove the selection snaps up to the parent.
func (s *session) delete() {
	if s.currentWidget == nil {
		s.setStatus("Delete: no widget selected")
		return
	}
	if s.currentWidget == s.target {
		s.setStatus("Delete: cannot remove the root")
		return
	}
	victim := s.currentWidget
	parent := victim.Parent()
	if err := s.d.Remove(victim); err != nil {
		s.setStatus("delete failed: " + err.Error())
		return
	}
	widgets.Relayout(parent)
	s.refreshTree()
	s.setDirty(true)
	s.setStatus(fmt.Sprintf("removed %s%s", widgetKind(victim), idSuffix(victim)))
	s.selectAfterMutation(parent)
}

// moveSibling reorders the current selection within its parent by
// delta (-1 = up, +1 = down). Returns true when the framework
// should treat the event as consumed; relevant when wired to a
// keystroke that should not bubble. Boundary attempts surface as
// a status message rather than silently no-oping.
func (s *session) moveSibling(delta int) bool {
	if s.currentWidget == nil {
		s.setStatus("Move: no widget selected")
		return false
	}
	parent := s.currentWidget.Parent()
	if parent == nil {
		s.setStatus("Move: no parent")
		return false
	}
	siblings := parent.Children()
	idx := -1
	for i, sib := range siblings {
		if sib == s.currentWidget {
			idx = i
			break
		}
	}
	if idx < 0 {
		s.setStatus("Move: child not in parent")
		return false
	}
	newIdx := idx + delta
	if newIdx < 0 || newIdx >= len(siblings) {
		s.setStatus("Move: already at boundary")
		return false
	}
	if err := s.d.Move(s.currentWidget, parent, newIdx); err != nil {
		s.setStatus("move failed: " + err.Error())
		return false
	}
	widgets.Relayout(parent)
	s.refreshTree()
	s.setDirty(true)
	direction := "↑"
	if delta > 0 {
		direction = "↓"
	}
	s.setStatus(fmt.Sprintf("moved %s%s %s",
		widgetKind(s.currentWidget), idSuffix(s.currentWidget), direction))
	return false
}

// selectAfterMutation re-points current* at the given survivor
// widget (typically the parent of a deleted child) and rebuilds the
// detail panes around it. Passing nil collapses the panes back to
// the empty-state placeholder. After refreshTree the survivor is
// located in the rebuilt tree and currentNode is updated so a
// subsequent Apply can re-label the node.
func (s *session) selectAfterMutation(w core.Widget) {
	if w == nil {
		s.currentWidget = nil
		s.currentForm = nil
		s.currentLayout = nil
		s.currentParent = nil
		s.currentNode = nil
		s.clearTabs()
		return
	}
	s.currentWidget = w
	s.currentNode = findTreeNodeFor(s.tree, w)
	s.rebuildPane(w)
}

// findTreeNodeFor walks the tree depth-first looking for a node
// whose opaque Data() pointer matches w. Returns nil when w isn't
// in the tree — for example because refreshTree hasn't run yet, or
// because w lives outside the target subtree the inspector mirrors.
func findTreeNodeFor(tree *widgets.Tree, w core.Widget) *widgets.TreeNode {
	root := tree.Root()
	if root == nil {
		return nil
	}
	var walk func(*widgets.TreeNode) *widgets.TreeNode
	walk = func(n *widgets.TreeNode) *widgets.TreeNode {
		if got, ok := n.Data().(core.Widget); ok && got == w {
			return n
		}
		for _, child := range n.Children() {
			if hit := walk(child); hit != nil {
				return hit
			}
		}
		return nil
	}
	return walk(root)
}

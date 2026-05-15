package inspector

import (
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// refreshTree rebuilds the tree from session.root. Bound to F5;
// the inspector doesn't observe app-side mutations automatically,
// so the user re-syncs manually when they know something has
// changed (e.g. after using the designer to add a widget).
func (s *session) refreshTree() {
	root := widgets.NewTreeNode("")
	root.Add(buildWidgetTreeNode(s.root))
	s.tree.SetRoot(root)
	widgets.Redraw(s.tree)
}

// onTreeSelect handles a tree selection event: pulls the widget
// out of the node's opaque data, stashes it on the session, and
// asks the details pane to rebuild around it.
func (s *session) onTreeSelect() {
	node := s.tree.Selected()
	if node == nil {
		return
	}
	w, ok := node.Data().(core.Widget)
	if !ok || w == nil {
		return
	}
	s.current = w
	s.rebuildDetails(w)
}

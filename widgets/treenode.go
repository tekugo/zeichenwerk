package widgets

// ==== AI ===================================================================

// NodeLoader is called the first time a node is expanded. It receives the node
// and is expected to call node.Add for each child. It is never called again
// after the first expansion — set a new loader via SetLoader to re-trigger.
type NodeLoader func(node *TreeNode)

// TreeNode is a node in a tree widget. Nodes are user-created; the widget
// treats the data field as opaque and passes it back in events.
type TreeNode struct {
	text     string
	data     any
	children []*TreeNode
	expanded bool
	disabled bool
	loader   NodeLoader // called once, on first expand; nil after invocation
}

// NewTreeNode creates a new tree node with the given display text and optional
// opaque data value. The node starts collapsed and enabled.
func NewTreeNode(text string, data ...any) *TreeNode {
	n := &TreeNode{text: text}
	if len(data) > 0 {
		n.data = data[0]
	}
	return n
}

// NewLazyTreeNode creates a tree node whose children are loaded on first
// expand. The loader is called once; subsequent expand/collapse cycles use the
// already-loaded children. Pass data as the third argument to attach an opaque
// value (e.g. a directory path) that the loader can retrieve via node.Data().
func NewLazyTreeNode(text string, loader NodeLoader, data ...any) *TreeNode {
	n := NewTreeNode(text, data...)
	n.loader = loader
	return n
}

// SetLoader attaches a lazy-load function that is called once the next time
// this node is expanded. Calling SetLoader again resets the load state so the
// new loader will be invoked on the following expand.
func (n *TreeNode) SetLoader(loader NodeLoader) *TreeNode {
	n.loader = loader
	return n
}

// Add appends child as a direct child of n and returns n for chaining.
func (n *TreeNode) Add(child *TreeNode) *TreeNode {
	n.children = append(n.children, child)
	return n
}

// Text returns the node's display text.
func (n *TreeNode) Text() string { return n.text }

// Data returns the opaque user data attached to the node.
func (n *TreeNode) Data() any { return n.data }

// Children returns the node's direct children.
func (n *TreeNode) Children() []*TreeNode { return n.children }

// Expanded reports whether the node is currently expanded.
func (n *TreeNode) Expanded() bool { return n.expanded }

// Leaf reports whether the node has no children. A node with a pending loader
// is never considered a leaf — it may yield children on first expand.
func (n *TreeNode) Leaf() bool { return n.loader == nil && len(n.children) == 0 }

// Expand expands the node.
func (n *TreeNode) Expand() { n.expanded = true }

// Collapse collapses the node.
func (n *TreeNode) Collapse() { n.expanded = false }

// Toggle toggles the node's expanded state.
func (n *TreeNode) Toggle() { n.expanded = !n.expanded }

// SetDisabled marks the node as non-selectable when d is true.
func (n *TreeNode) SetDisabled(d bool) { n.disabled = d }

// Disabled reports whether the node is non-selectable.
func (n *TreeNode) Disabled() bool { return n.disabled }

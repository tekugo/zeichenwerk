package zeichenwerk

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

// flatItem is an entry in the flattened visible-node list produced by rebuild.
type flatItem struct {
	node   *TreeNode
	depth  int
	isLast bool   // last child of its parent — selects └─ vs ├─
	trunk  []bool // trunk[d] = true → draw │ at indent column d
}

// Tree is a scrollable, hierarchical list widget. Nodes can be expanded and
// collapsed. Navigation uses a flattened visible-node list so scrolling and
// highlighting behave identically to List.
type Tree struct {
	Component
	root        *TreeNode  // invisible root; its children are top-level items
	flat        []flatItem // flattened ordered list of currently visible nodes
	index       int        // highlighted position in flat (-1 if empty)
	offset      int        // scroll offset (first visible row index in flat)
	scrollbar   bool       // whether to draw a vertical scrollbar
	filterQuery string     // active filter query ("" = no filter)
}

// NewTree creates a new Tree widget with the given id and CSS class.
func NewTree(id, class string) *Tree {
	t := &Tree{
		Component: Component{id: id, class: class},
		root:      NewTreeNode(""),
		index:     -1,
		scrollbar: true,
	}
	t.SetFlag(FlagFocusable, true)
	OnKey(t, t.handleKey)
	OnMouse(t, t.handleMouse)
	return t
}

// ---- Data ----------------------------------------------------------------

// Add appends a top-level node and rebuilds the flat list.
func (t *Tree) Add(node *TreeNode) {
	t.root.Add(node)
	t.rebuild()
}

// SetRoot replaces the invisible root and rebuilds the flat list. The direct
// children of root become the top-level items.
func (t *Tree) SetRoot(root *TreeNode) {
	t.root = root
	t.rebuild()
}

// Root returns the invisible root node. Its direct children are top-level items.
func (t *Tree) Root() *TreeNode { return t.root }

// ---- Navigation ----------------------------------------------------------

// Selected returns the currently highlighted node, or nil if nothing is highlighted.
func (t *Tree) Selected() *TreeNode {
	if t.index < 0 || t.index >= len(t.flat) {
		return nil
	}
	return t.flat[t.index].node
}

// Select highlights the given node if it is visible and adjusts the scroll.
func (t *Tree) Select(node *TreeNode) {
	for i, item := range t.flat {
		if item.node == node {
			t.moveTo(i)
			return
		}
	}
}

// Move moves the highlight by count, skipping disabled nodes. Clamps at bounds.
func (t *Tree) Move(count int) {
	if len(t.flat) == 0 || count == 0 {
		return
	}
	direction := 1
	if count < 0 {
		direction = -1
		count = -count
	}
	newIndex := t.index
	for i := 0; i < count; i++ {
		next := t.skip(newIndex+direction, direction)
		if next < 0 || next >= len(t.flat) {
			break
		}
		newIndex = next
	}
	if newIndex != t.index {
		t.moveTo(newIndex)
	}
}

// First highlights the first enabled node.
func (t *Tree) First() {
	for i, item := range t.flat {
		if !item.node.disabled {
			t.moveTo(i)
			return
		}
	}
}

// Last highlights the last enabled node.
func (t *Tree) Last() {
	for i := len(t.flat) - 1; i >= 0; i-- {
		if !t.flat[i].node.disabled {
			t.moveTo(i)
			return
		}
	}
}

// PageUp moves the highlight up by the viewport height.
func (t *Tree) PageUp() {
	_, _, _, h := t.Content()
	t.Move(-h)
}

// PageDown moves the highlight down by the viewport height.
func (t *Tree) PageDown() {
	_, _, _, h := t.Content()
	t.Move(h)
}

// ---- Expand / Collapse ---------------------------------------------------

// Expand expands node and rebuilds the flat list. If the node has a pending
// loader it is called first to populate the children, then cleared so it is
// not invoked again.
func (t *Tree) Expand(node *TreeNode) {
	if node.loader != nil {
		node.loader(node)
		node.loader = nil
	}
	node.Expand()
	t.rebuildKeep(node)
	t.Dispatch(t, EvtChange, node)
}

// Collapse collapses node and rebuilds the flat list.
func (t *Tree) Collapse(node *TreeNode) {
	node.Collapse()
	t.rebuildKeep(node)
	t.Dispatch(t, EvtChange, node)
}

// ExpandAll expands every node recursively and rebuilds.
func (t *Tree) ExpandAll() {
	expandAllNodes(t.root)
	t.rebuild()
}

// CollapseAll collapses every node recursively and rebuilds.
func (t *Tree) CollapseAll() {
	collapseAllNodes(t.root)
	t.rebuild()
}

func expandAllNodes(n *TreeNode) {
	n.Expand()
	for _, child := range n.children {
		expandAllNodes(child)
	}
}

func collapseAllNodes(n *TreeNode) {
	n.Collapse()
	for _, child := range n.children {
		collapseAllNodes(child)
	}
}

// ---- Apply / Hint / Render -----------------------------------------------

// Apply applies theme styles for the tree widget.
func (t *Tree) Apply(theme *Theme) {
	theme.Apply(t, t.Selector("tree"), "disabled", "focused", "hovered")
	theme.Apply(t, t.Selector("tree/highlight"), "focused")
	theme.Apply(t, t.Selector("tree/indent"))
}

// Hint returns (maxRowWidth, len(flat)) where maxRowWidth is the widest
// rendered row across all currently visible flat items. If either hwidth or
// hheight has been set explicitly, both are returned as-is.
func (t *Tree) Hint() (int, int) {
	if t.hwidth != 0 || t.hheight != 0 {
		return t.hwidth, t.hheight
	}
	maxW := 0
	for _, item := range t.flat {
		w := item.depth*2 + 2 + 2 + len([]rune(item.node.text))
		if item.depth == 0 {
			w = 2 + len([]rune(item.node.text)) // no connector for top-level
		}
		if w > maxW {
			maxW = w
		}
	}
	return maxW, len(t.flat)
}

// Render draws the visible rows of the tree.
func (t *Tree) Render(r *Renderer) {
	t.Component.Render(r)

	cx, cy, cw, ch := t.Content()
	if ch < 1 || cw < 1 {
		return
	}

	tw := cw
	if t.scrollbar && len(t.flat) > ch {
		tw = cw - 1
	}

	baseStyle := t.Style()
	indentStyle := t.Style("indent")
	highlightStyle := t.Style("highlight")
	highlightFocusedStyle := t.Style("highlight:focused")
	disabledStyle := t.Style(":disabled")
	_ = disabledStyle

	// Theme strings
	strExpanded := r.theme.String("tree.expanded")
	strCollapsed := r.theme.String("tree.collapsed")
	strBranch := r.theme.String("tree.branch")
	strLast := r.theme.String("tree.last")
	strTrunk := r.theme.String("tree.trunk")

	end := t.offset + ch
	if end > len(t.flat) {
		end = len(t.flat)
	}

	for i, item := range t.flat[t.offset:end] {
		rowIndex := t.offset + i
		rowY := cy + i

		// Determine style
		var fg, bg, font string
		if item.node.disabled {
			s := disabledStyle
			fg, bg, font = s.Foreground(), s.Background(), s.Font()
		} else if rowIndex == t.index {
			if t.Flag(FlagFocused) {
				s := highlightFocusedStyle
				fg, bg, font = s.Foreground(), s.Background(), s.Font()
			} else {
				s := highlightStyle
				fg, bg, font = s.Foreground(), s.Background(), s.Font()
			}
		} else {
			s := baseStyle
			fg, bg, font = s.Foreground(), s.Background(), s.Font()
		}

		// Draw indent columns (one per ancestor level, 3 chars each)
		col := cx
		r.Set(indentStyle.Foreground(), bg, "")
		for d := 1; d < item.depth; d++ {
			if d < len(item.trunk) && item.trunk[d] {
				r.Text(col, rowY, strTrunk, 3)
			} else {
				r.Text(col, rowY, "   ", 3)
			}
			col += 3
		}

		// Draw connector (3 chars; empty for top-level)
		if item.depth > 0 {
			if item.isLast {
				r.Text(col, rowY, strLast, 3)
			} else {
				r.Text(col, rowY, strBranch, 3)
			}
			col += 3
		}

		// Draw indicator (3 chars)
		var indicator string
		if item.node.Leaf() {
			indicator = "  "
		} else if item.node.expanded {
			indicator = strExpanded
		} else {
			indicator = strCollapsed
		}
		r.Set(fg, bg, font)
		r.Text(col, rowY, indicator, len(indicator))
		col += len(indicator) - 1

		// Draw text, truncated to remaining width
		remaining := cx + tw - col
		if remaining > 0 {
			r.Text(col, rowY, item.node.text, remaining)
		}
		// Fill rest of row with background
		textEnd := col + remaining
		if textEnd < cx+tw {
			r.Fill(textEnd, rowY, cx+tw-textEnd, 1, " ")
		}
	}

	// Fill empty rows below items
	filledRows := end - t.offset
	if filledRows < ch {
		r.Set(baseStyle.Foreground(), baseStyle.Background(), "")
		r.Fill(cx, cy+filledRows, tw, ch-filledRows, " ")
	}

	// Scrollbar
	if t.scrollbar && len(t.flat) > ch {
		r.Set(baseStyle.Foreground(), baseStyle.Background(), "")
		r.ScrollbarV(cx+cw-1, cy, ch, t.offset, len(t.flat))
	}
}

// ---- Keyboard / Mouse ----------------------------------------------------

func (t *Tree) handleKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyUp:
		t.Move(-1)
		return true
	case tcell.KeyDown:
		t.Move(1)
		return true
	case tcell.KeyHome:
		t.First()
		return true
	case tcell.KeyEnd:
		t.Last()
		return true
	case tcell.KeyPgUp:
		t.PageUp()
		return true
	case tcell.KeyPgDn:
		t.PageDown()
		return true
	case tcell.KeyRight:
		t.handleRight()
		return true
	case tcell.KeyLeft:
		t.handleLeft()
		return true
	case tcell.KeyEnter:
		t.handleActivate()
		return true
	case tcell.KeyRune:
		if ev.Str() == " " {
			t.handleActivate()
			return true
		}
	}
	return false
}

func (t *Tree) handleRight() {
	if t.index < 0 || t.index >= len(t.flat) {
		return
	}
	node := t.flat[t.index].node
	if node.Leaf() {
		return
	}
	if !node.expanded {
		t.Expand(node)
	} else {
		// Move to first child
		if t.index+1 < len(t.flat) {
			t.moveTo(t.index + 1)
		}
	}
}

func (t *Tree) handleLeft() {
	if t.index < 0 || t.index >= len(t.flat) {
		return
	}
	node := t.flat[t.index].node
	if node.expanded && !node.Leaf() {
		t.Collapse(node)
		return
	}
	// Move to parent: scan backward for first item with smaller depth
	currentDepth := t.flat[t.index].depth
	for i := t.index - 1; i >= 0; i-- {
		if t.flat[i].depth < currentDepth {
			t.moveTo(i)
			return
		}
	}
}

func (t *Tree) handleActivate() {
	if t.index < 0 || t.index >= len(t.flat) {
		return
	}
	node := t.flat[t.index].node
	if !node.Leaf() {
		if node.expanded {
			t.Collapse(node)
		} else {
			t.Expand(node)
		}
	}
	t.Dispatch(t, EvtActivate, node)
}

func (t *Tree) handleMouse(ev *tcell.EventMouse) bool {
	if ev.Buttons() != tcell.Button1 {
		return false
	}
	mx, my := ev.Position()
	cx, cy, cw, ch := t.Content()
	if mx < cx || mx >= cx+cw || my < cy || my >= cy+ch {
		return false
	}
	if t.scrollbar && len(t.flat) > ch && mx == cx+cw-1 {
		return false
	}
	index := t.offset + (my - cy)
	if index < 0 || index >= len(t.flat) {
		return false
	}
	if t.flat[index].node.disabled {
		return false
	}
	t.moveTo(index)
	return true
}

// ---- Internal ------------------------------------------------------------

// moveTo sets the highlight to the given flat index, adjusts scroll, and
// dispatches EvtSelect. No-op if already at that index.
func (t *Tree) moveTo(index int) {
	if index == t.index {
		return
	}
	t.index = index
	t.adjust()
	t.Dispatch(t, EvtSelect, t.flat[index].node)
	Redraw(t)
}

// rebuild re-flattens the visible tree and clamps index/offset.
func (t *Tree) rebuild() {
	selected := t.Selected()
	t.flat = t.flat[:0]
	top := t.visibleTopLevel()
	for i, child := range top {
		t.flattenNode(child, 0, nil, i == len(top)-1)
	}
	t.clamp(selected)
	Redraw(t)
}

// rebuildKeep rebuilds and tries to keep highlight on the given node.
func (t *Tree) rebuildKeep(node *TreeNode) {
	t.flat = t.flat[:0]
	top := t.visibleTopLevel()
	for i, child := range top {
		t.flattenNode(child, 0, nil, i == len(top)-1)
	}
	// Try to restore selection to node; fall back to previous index
	t.index = -1
	for i, item := range t.flat {
		if item.node == node {
			t.index = i
			break
		}
	}
	if t.index == -1 && len(t.flat) > 0 {
		t.index = 0
	}
	t.adjust()
	Redraw(t)
}

// clamp restores the highlight to selected (if still visible) or clamps.
func (t *Tree) clamp(selected *TreeNode) {
	if selected != nil {
		for i, item := range t.flat {
			if item.node == selected {
				t.index = i
				t.adjust()
				return
			}
		}
	}
	// Selected node no longer visible; clamp index
	if t.index >= len(t.flat) {
		t.index = len(t.flat) - 1
	}
	if t.index < 0 && len(t.flat) > 0 {
		t.index = 0
	}
	t.adjust()
}

// flattenNode performs a depth-first traversal, adding visible nodes to flat.
// parentTrunk is the trunk slice passed down from the parent level.
// When a filterQuery is active, nodes whose subtrees contain no match are
// skipped; parent nodes of matching descendants are auto-expanded for this
// render without mutating their expanded state.
func (t *Tree) flattenNode(node *TreeNode, depth int, parentTrunk []bool, isLast bool) {
	if t.filterQuery != "" && !nodeMatchesSubtree(node, t.filterQuery) {
		return
	}

	// Build trunk for this node's row
	trunk := make([]bool, len(parentTrunk))
	copy(trunk, parentTrunk)

	t.flat = append(t.flat, flatItem{
		node:   node,
		depth:  depth,
		isLast: isLast,
		trunk:  trunk,
	})

	// During filtering, auto-expand every node that has matching descendants so
	// the path to matches stays visible. The node's expanded state is not mutated.
	shouldExpand := node.expanded || (t.filterQuery != "" && !node.Leaf())
	if shouldExpand && len(node.children) > 0 {
		// Trunk for children: inherit parent trunk + whether this node has more siblings
		childTrunk := make([]bool, len(parentTrunk)+1)
		copy(childTrunk, parentTrunk)
		childTrunk[len(parentTrunk)] = !isLast // true = draw │ at this depth

		// When filtering, compute the visible child set first so isLast is correct.
		children := node.children
		if t.filterQuery != "" {
			var visible []*TreeNode
			for _, c := range node.children {
				if nodeMatchesSubtree(c, t.filterQuery) {
					visible = append(visible, c)
				}
			}
			children = visible
		}
		for i, child := range children {
			t.flattenNode(child, depth+1, childTrunk, i == len(children)-1)
		}
	}
}

// visibleTopLevel returns the root's direct children that should be included in
// the current flat list. When no filter is active the full children slice is
// returned as-is. When a filter is active only children whose subtrees contain
// at least one matching node are returned.
func (t *Tree) visibleTopLevel() []*TreeNode {
	if t.filterQuery == "" {
		return t.root.children
	}
	var out []*TreeNode
	for _, c := range t.root.children {
		if nodeMatchesSubtree(c, t.filterQuery) {
			out = append(out, c)
		}
	}
	return out
}

// Filter applies a case-insensitive substring match to the tree, showing only
// nodes (and their ancestors) whose text contains filter. An empty string clears
// the filter and restores the full tree view.
func (t *Tree) Filter(filter string) {
	t.filterQuery = filter
	t.rebuild()
}

// Suggest returns node labels that have query as a case-insensitive prefix.
// All nodes in the tree are searched depth-first regardless of their expanded
// or filtered state. Returns nil when nothing matches or query is empty.
func (t *Tree) Suggest(query string) []string {
	if query == "" {
		return nil
	}
	lower := strings.ToLower(query)
	var results []string
	var walk func(*TreeNode)
	walk = func(n *TreeNode) {
		for _, child := range n.children {
			if strings.HasPrefix(strings.ToLower(child.text), lower) {
				results = append(results, child.text)
			}
			walk(child)
		}
	}
	walk(t.root)
	if len(results) == 0 {
		return nil
	}
	return results
}

// nodeMatchesSubtree reports whether node or any of its descendants have text
// that contains query as a case-insensitive substring.
func nodeMatchesSubtree(node *TreeNode, query string) bool {
	if strings.Contains(strings.ToLower(node.text), strings.ToLower(query)) {
		return true
	}
	for _, child := range node.children {
		if nodeMatchesSubtree(child, query) {
			return true
		}
	}
	return false
}

// skip returns the next enabled index in direction (1 or -1), or -1/len(flat)
// if none exists in that direction.
func (t *Tree) skip(index, direction int) int {
	for index >= 0 && index < len(t.flat) && t.flat[index].node.disabled {
		index += direction
	}
	return index
}

// adjust clamps offset so the highlighted index is visible.
func (t *Tree) adjust() {
	_, _, _, h := t.Content()
	if h <= 0 {
		return
	}
	if t.index < t.offset {
		t.offset = t.index
	} else if t.index >= t.offset+h {
		t.offset = t.index - h + 1
	}
	if t.offset < 0 {
		t.offset = 0
	}
	maxScroll := max(len(t.flat)-h, 0)
	if t.offset > maxScroll {
		t.offset = maxScroll
	}
}

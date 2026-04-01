# Tree

Scrollable, hierarchical list widget for tree-structured data. Nodes can be expanded and collapsed. Navigation uses a flattened visible-node list so scrolling and highlighting behave identically to `List`.

## TreeNode

`TreeNode` is user-created and holds display text, optional opaque data, and child nodes.

**Constructor:** `NewTreeNode(text string, data ...any) *TreeNode`

| Method | Description |
|--------|-------------|
| `Add(child *TreeNode) *TreeNode` | Appends a child; returns the receiver for chaining |
| `Children() []*TreeNode` | Returns direct children |
| `Collapse()` | Collapses the node |
| `Data() any` | Returns the opaque user data |
| `Disabled() bool` | Reports whether the node is non-selectable |
| `Expand()` | Expands the node |
| `Expanded() bool` | Reports whether the node is expanded |
| `Leaf() bool` | True when the node has no children and no pending loader |
| `SetDisabled(bool)` | Marks the node as non-selectable |
| `SetLoader(fn NodeLoader) *TreeNode` | Attaches a lazy-load function (see Lazy loading) |
| `Text() string` | Returns the display text |
| `Toggle()` | Toggles expanded state |

## Lazy loading

A node whose children are loaded on first expand is created with `NewLazyTreeNode`:

```go
func NewLazyTreeNode(text string, loader NodeLoader, data ...any) *TreeNode
```

`NodeLoader` is `func(node *TreeNode)`. It is called once when the node is first expanded and is expected to call `node.Add` for each child. The loader is cleared after the first call so subsequent expand/collapse cycles use the already-loaded children.

`SetLoader` re-arms a node so the next expand calls the new loader again — useful for refresh.

## Tree widget

**Constructor:** `NewTree(id, class string) *Tree`

### Data methods

| Method | Description |
|--------|-------------|
| `Add(node *TreeNode)` | Appends a top-level node |
| `Root() *TreeNode` | Returns the invisible root (its children are top-level) |
| `SetRoot(root *TreeNode)` | Replaces the invisible root |

### Navigation methods

| Method | Description |
|--------|-------------|
| `First()` | Highlights first enabled node |
| `Last()` | Highlights last enabled node |
| `Move(count int)` | Moves highlight by count, skipping disabled nodes |
| `PageDown()` | Moves down by viewport height |
| `PageUp()` | Moves up by viewport height |
| `Select(node *TreeNode)` | Highlights the given node if visible |
| `Selected() *TreeNode` | Returns the highlighted node, or nil |

### Expand / collapse methods

| Method | Description |
|--------|-------------|
| `Collapse(node *TreeNode)` | Collapses node; dispatches `"change"` |
| `CollapseAll()` | Collapses every node recursively |
| `Expand(node *TreeNode)` | Expands node (runs loader if pending); dispatches `"change"` |
| `ExpandAll()` | Expands every node recursively |

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"activate"` | `*TreeNode` | Enter or Space pressed on a node |
| `"change"` | `*TreeNode` | Node was expanded or collapsed |
| `"select"` | `*TreeNode` | Highlighted node changed |

## Notes

Flags: `"focusable"`

Keyboard:

| Key | Behaviour |
|-----|-----------|
| `↑` / `↓` | Move highlight; skips disabled nodes |
| `→` | Expand collapsed node with children; move to first child if already expanded |
| `←` | Collapse expanded node; move to parent if collapsed |
| `Enter` / `Space` | Toggle expand/collapse; dispatch `"activate"` |
| `Home` / `End` | Jump to first / last enabled node |
| `PgUp` / `PgDn` | Move by viewport height |

Mouse: single click moves highlight.

Style selectors:

| Selector | Applied to |
|----------|------------|
| `"tree"` | Base background and foreground |
| `"tree/highlight"` | Highlighted row when unfocused |
| `"tree/highlight:focused"` | Highlighted row when focused |
| `"tree/indent"` | Indent lines and connector characters |
| `":disabled"` | Disabled row text |

Theme strings:

| Key | Default | Description |
|-----|---------|-------------|
| `tree.expanded` | `▼ ` | Indicator for an expanded non-leaf node |
| `tree.collapsed` | `▶ ` | Indicator for a collapsed non-leaf node |
| `tree.branch` | `├─` | Connector for a non-last child |
| `tree.last` | `└─` | Connector for the last child |
| `tree.trunk` | `│ ` | Vertical continuation line at an ancestor column |

## TreeFS

`TreeFS` is a `Tree` pre-wired for filesystem navigation. It wraps `*Tree` and loads directory contents lazily on first expand. Directories appear before files within each directory, both sorted alphabetically. Every node's `Data()` is the absolute path string of the entry.

**Constructor:** `NewTreeFS(id, class, root string, dirsOnly bool) *TreeFS`

`dirsOnly` hides files when true.

| Method | Description |
|--------|-------------|
| `DirsOnly() bool` | Reports whether files are hidden |
| `RootPath() string` | Returns the absolute path of the current root directory |
| `SetDirsOnly(bool)` | Toggles file visibility (takes effect on next unexpanded directory) |
| `SetRoot(path string)` | Replaces the root directory and resets the tree |

`TreeFS` uses the `"tree-fs"` selector family, falling back to `"tree"` styles through the theme cascade.

Read errors (e.g. permission denied) are shown as a single `(error message)` child node.

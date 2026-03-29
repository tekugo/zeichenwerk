# Tree

A scrollable, hierarchical list widget for displaying tree-structured data.
Nodes can be expanded and collapsed. Navigation uses a flattened visible-node
list so scrolling and highlighting behave identically to `List`.

## Data model — `TreeNode`

```go
type TreeNode struct {
    text     string
    data     any
    children []*TreeNode
    expanded bool
    disabled bool
}
```

`TreeNode` is user-created; its constructor and mutators are part of the public
API. User-supplied `data` is opaque to the widget; it is passed back in events.

### `TreeNode` constructor and methods

```go
func NewTreeNode(text string, data ...any) *TreeNode
```

| Method | Description |
|--------|-------------|
| `Add(child *TreeNode) *TreeNode` | Appends a child; returns the receiver for chaining |
| `Text() string` | Returns display text |
| `Data() any` | Returns user data |
| `Children() []*TreeNode` | Returns direct children |
| `Expanded() bool` | Returns expand state |
| `Leaf() bool` | True if no children |
| `Expand()` | Expands the node |
| `Collapse()` | Collapses the node |
| `Toggle()` | Toggles expand state |
| `SetDisabled(bool)` | Marks node as non-selectable |
| `Disabled() bool` | Returns disabled state |

## Tree widget

```go
type Tree struct {
    Component
    root      *TreeNode   // Invisible root; its children are the top-level items
    flat      []flatItem  // Flattened ordered list of currently visible nodes
    index     int         // Highlighted position in flat (-1 if empty)
    offset    int         // Scroll offset (first visible row index in flat)
    scrollbar bool        // Whether to draw a vertical scrollbar
}
```

`root` is never rendered. It acts solely as the container for top-level nodes,
keeping the flattening algorithm uniform at all depths.

### `flatItem` (unexported)

```go
type flatItem struct {
    node   *TreeNode
    depth  int
    isLast bool    // Last child of its parent — selects └─ vs ├─
    trunk  []bool  // trunk[d] = true → draw │ at indent column d
}
```

`trunk` is computed once during flattening and drives the vertical-line
characters that show nesting continuity.

## Constructor

```go
func NewTree(id, class string) *Tree
```

- Creates an invisible `root` node.
- Initialises `index = -1`, `scrollbar = true`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

## Methods

### Data

| Method | Description |
|--------|-------------|
| `Add(node *TreeNode)` | Appends a top-level node; calls `rebuild()` |
| `SetRoot(root *TreeNode)` | Replaces the invisible root; calls `rebuild()` |
| `Root() *TreeNode` | Returns the invisible root (direct children are top-level) |

### Navigation

| Method | Description |
|--------|-------------|
| `Select(node *TreeNode)` | Highlights the given node if visible; adjusts scroll |
| `Selected() *TreeNode` | Returns the currently highlighted node, or nil |
| `Move(count int)` | Moves highlight by count, skipping disabled nodes |
| `First()` | Highlights first enabled node |
| `Last()` | Highlights last enabled node |
| `PageUp()` / `PageDown()` | Moves by viewport height |

### Expand / Collapse

| Method | Description |
|--------|-------------|
| `Expand(node *TreeNode)` | Expands node; calls `rebuild()` |
| `Collapse(node *TreeNode)` | Collapses node; calls `rebuild()` |
| `ExpandAll()` | Expands every node recursively; calls `rebuild()` |
| `CollapseAll()` | Collapses every node recursively; calls `rebuild()` |

`rebuild()` re-flattens the visible tree and clamps `index` and `offset` to
remain valid.

## Keyboard interaction

| Key | Behaviour |
|-----|-----------|
| `↑` / `↓` | `Move(±1)` — skips disabled nodes |
| `→` | If highlighted node is collapsed with children: `Expand`; if expanded: move to first child |
| `←` | If highlighted node is expanded: `Collapse`; else: move to parent |
| `Enter` | Toggle expand/collapse on nodes with children; dispatch `EvtActivate` |
| `Space` | Same as Enter |
| `Home` / `End` | `First()` / `Last()` |
| `PgUp` / `PgDn` | `PageUp()` / `PageDown()` |

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"select"` | `*TreeNode` | Highlighted node changed |
| `"activate"` | `*TreeNode` | Enter/Space pressed on a node |
| `"change"` | `*TreeNode` | Node expanded or collapsed; data is the toggled node |

## Rendering

### Flattening (`rebuild`)

Depth-first traversal of `root.children`; only expanded nodes yield their
children. Each `flatItem` captures `depth`, `isLast`, and `trunk` — a boolean
slice where `trunk[d]` is `true` when the ancestor at depth `d` is **not** the
last child of its own parent (i.e., more siblings follow, so a `│` must be
drawn at that indent column).

### Row layout

Each visible row is rendered as:

```
<indent columns><connector><indicator> <text>
```

**Indent columns** (one per ancestor level, 2 chars each):

| `trunk[d]` | Rendered as |
|------------|-------------|
| `true` | `│ ` |
| `false` | `  ` |

**Connector** (2 chars, based on `isLast` and `depth`):

| Condition | Rendered as |
|-----------|-------------|
| `depth == 0` | `` (empty — no connector for top-level) |
| `isLast` | `└─` |
| else | `├─` |

**Indicator** (2 chars):

| Condition | Rendered as |
|-----------|-------------|
| Leaf | `  ` |
| Collapsed with children | `▶ ` |
| Expanded with children | `▼ ` |

**Text**: the node's `text`, truncated to the remaining width.

The row prefix characters (`▶`, `▼`, `├─`, `└─`, `│`) are fetched from the
theme via `theme.String("tree.*")` keys, matching the pattern used by `Select`
for its dropdown character. This allows themes to substitute ASCII alternatives.

### Highlight

Row styles follow the same pattern as `List`:

- `"tree/highlight"` — highlighted row when not focused
- `"tree/highlight:focused"` — highlighted row when focused
- `"tree/indent"` — colour for indent and connector characters (separate from text)
- Disabled rows use `":disabled"` style

### Scrollbar

Drawn on the rightmost column if `scrollbar` is true and `len(flat) > viewport
height`, using `r.ScrollbarV` — identical to `List`.

### `Hint()`

Returns `(maxRowWidth, len(flat))` where `maxRowWidth` is the widest rendered
row (indent + prefix + text) across all currently visible flat items.

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"tree"` | Base background and foreground |
| `"tree/highlight"` | Highlighted row (unfocused) |
| `"tree/highlight:focused"` | Highlighted row (focused) |
| `"tree/indent"` | Indent lines and connector characters |
| `":disabled"` | Disabled row text |

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `tree.expanded` | `▼ ` | Indicator for expanded non-leaf node |
| `tree.collapsed` | `▶ ` | Indicator for collapsed non-leaf node |
| `tree.branch` | `├─` | Connector for non-last child |
| `tree.last` | `└─` | Connector for last child |
| `tree.trunk` | `│ ` | Vertical continuation line |

## Implementation Plan

1. **`treenode.go`** — new file
   - Define `TreeNode` with all methods listed above.
   - `Add` returns the receiver (`*TreeNode`) to allow chaining:
     `root.Add(NewTreeNode("a").Add(NewTreeNode("b")))`.

2. **`tree.go`** — new file
   - Define `flatItem` (unexported).
   - Define `Tree` struct and `NewTree`.
   - Implement `rebuild` (flatten), `adjust` (scroll clamping), `skip`
     (disabled-aware neighbour search — mirrors `List.skip`).
   - Implement all navigation methods, `handleKey`, `Apply`, `Hint`, `Render`.
   - `Render` iterates `flat[offset : offset+viewportH]` and draws each row.

3. **`builder.go`** — add `Tree` method
   ```go
   func (b *Builder) Tree(id string) *Builder
   ```

4. **Theme** — add `"tree.*"` string keys and `"tree/indent"`,
   `"tree/highlight"` style entries to built-in themes.

5. **Tests** — `tree_test.go`
   - `rebuild` produces correct `depth`, `isLast`, and `trunk` for a
     multi-level tree.
   - `←` on an expanded node collapses it; `←` on a collapsed node moves to
     its parent.
   - `→` on a collapsed node expands it; `→` on an expanded node moves to
     its first child.
   - `EvtChange` is dispatched with the toggled node on expand/collapse.
   - `EvtActivate` is dispatched on Enter.
   - Disabled nodes are skipped by `Move`.
   - `CollapseAll` followed by `ExpandAll` produces the same flat list as the
     initial tree.

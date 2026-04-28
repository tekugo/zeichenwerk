# TreeFS

A `Tree` pre-wired for filesystem navigation. Loads directory contents lazily — children are read from disk the first time a node is expanded.

**Constructor:** `NewTreeFS(id, class, root string, dirsOnly bool) *TreeFS`

`root` is the path to display; if `dirsOnly` is true, files are hidden.

## Methods

- `RootPath() string` — absolute path of the current root
- `SetRoot(path string)` — replace the root and reset the tree
- `DirsOnly() bool` — whether files are hidden
- `SetDirsOnly(dirsOnly bool)` — toggle file visibility (does not reload already-expanded nodes; call `SetRoot` to start fresh)

Plus everything inherited from `Tree`: `Selected()`, `Select(node)`, `Expand`, `Collapse`, `Filter(query)`, navigation events, etc.

## Notes

The opaque `data` on every `TreeNode` is the **absolute path** (`string`) of the entry it represents. Activate handlers can recover it directly:

```go
tfs := zw.NewTreeFS("files", "", ".", false)
widgets.OnActivate(tfs.Tree, func(idx int) bool {
    node := tfs.Tree.Selected()
    if node != nil {
        path := node.Data().(string)
        // open the file, change directory, etc.
    }
    return true
})
```

Style selectors: `tree-fs`, `tree-fs/highlight`, `tree-fs/indent`. They cascade to `tree`, `tree/highlight`, `tree/indent` when not overridden, so a theme that styles `tree` automatically styles `TreeFS` too.

Errors during directory listing surface as a single child node with the error message.

package zeichenwerk

import (
	"os"
	"path/filepath"
	"sort"
)

// ==== AI ===================================================================

// TreeFS is a Tree widget pre-wired for filesystem navigation. It loads
// directory contents lazily — children are read from disk the first time a
// node is expanded. Directories are always shown; files are shown only when
// dirsOnly is false.
//
// The opaque data stored on every TreeNode is the absolute path (string) of
// the entry it represents.
type TreeFS struct {
	*Tree
	dirsOnly bool
	rootPath string // absolute path of the current root
}

// NewTreeFS creates a TreeFS rooted at root. If dirsOnly is true only
// directories appear; files are hidden. The root directory is represented as a
// single top-level node that loads its contents on first expand.
func NewTreeFS(id, class, root string, dirsOnly bool) *TreeFS {
	tfs := &TreeFS{
		Tree:     NewTree(id, class),
		dirsOnly: dirsOnly,
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	tfs.rootPath = abs
	tfs.Tree.Add(NewLazyTreeNode(filepath.Base(abs), tfs.loadDir, abs))
	return tfs
}

// RootPath returns the absolute path of the current root directory.
func (tfs *TreeFS) RootPath() string { return tfs.rootPath }

// Apply applies theme styles. TreeFS uses the "tree-fs" selector family so
// themes can style it independently of a plain Tree, while falling back to
// "tree" styles automatically through the theme's cascade.
func (tfs *TreeFS) Apply(theme *Theme) {
	theme.Apply(tfs.Tree, tfs.Tree.Selector("tree-fs"), "disabled", "focused", "hovered")
	theme.Apply(tfs.Tree, tfs.Tree.Selector("tree-fs/highlight"), "focused")
	theme.Apply(tfs.Tree, tfs.Tree.Selector("tree-fs/indent"))
}

// SetDirsOnly controls whether files are shown. Changing this setting does not
// reload already-expanded nodes; call SetRoot to start fresh.
func (tfs *TreeFS) SetDirsOnly(dirsOnly bool) {
	tfs.dirsOnly = dirsOnly
}

// DirsOnly reports whether files are hidden.
func (tfs *TreeFS) DirsOnly() bool {
	return tfs.dirsOnly
}

// SetRoot replaces the current root with path and resets the tree.
func (tfs *TreeFS) SetRoot(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	tfs.rootPath = abs
	// Reset the underlying tree's root
	tfs.Tree.root = NewTreeNode("")
	tfs.Tree.flat = tfs.Tree.flat[:0]
	tfs.Tree.index = -1
	tfs.Tree.offset = 0
	tfs.Tree.Add(NewLazyTreeNode(filepath.Base(abs), tfs.loadDir, abs))
}

// loadDir is the NodeLoader used for every directory node.
func (tfs *TreeFS) loadDir(node *TreeNode) {
	path := node.Data().(string)
	entries, err := os.ReadDir(path)
	if err != nil {
		node.Add(NewTreeNode("(" + err.Error() + ")"))
		return
	}

	dirs := make([]os.DirEntry, 0, len(entries))
	files := make([]os.DirEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		} else if !tfs.dirsOnly {
			files = append(files, e)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	for _, e := range dirs {
		child := NewLazyTreeNode(e.Name(), tfs.loadDir, filepath.Join(path, e.Name()))
		node.Add(child)
	}
	for _, e := range files {
		child := NewTreeNode(e.Name(), filepath.Join(path, e.Name()))
		node.Add(child)
	}
}

package widgets

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// makeFSTree builds a temporary directory structure for testing:
//
//	root/
//	  a/
//	    file-a1.txt
//	    file-a2.txt
//	    sub/
//	      deep.txt
//	  b/
//	  file-root.txt
func makeFSTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dirs := []string{"a", "a/sub", "b"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := []string{"a/file-a1.txt", "a/file-a2.txt", "a/sub/deep.txt", "file-root.txt"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(root, f), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// ---- constructor -----------------------------------------------------------

func TestTreeFS_RootNode(t *testing.T) {
	root := makeFSTree(t)
	tfs := NewTreeFS("fs", "", root, false)
	if len(tfs.flat) != 1 {
		t.Fatalf("expected 1 top-level node (root dir), got %d", len(tfs.flat))
	}
	rootNode := tfs.flat[0].node
	if rootNode.Data().(string) != root {
		t.Fatalf("root node data: got %q want %q", rootNode.Data(), root)
	}
	if rootNode.Text() != filepath.Base(root) {
		t.Fatalf("root node text: got %q want %q", rootNode.Text(), filepath.Base(root))
	}
	if rootNode.Leaf() {
		t.Fatal("root directory node should not be a leaf before loading")
	}
}

func TestTreeFS_DirsOnly_Default(t *testing.T) {
	tfs := NewTreeFS("fs", "", t.TempDir(), false)
	if tfs.DirsOnly() {
		t.Fatal("dirsOnly should be false by default")
	}
}

// ---- lazy loading ----------------------------------------------------------

func TestTreeFS_ExpandLoadsChildren(t *testing.T) {
	root := makeFSTree(t)
	tfs := NewTreeFS("fs", "", root, false)
	rootNode := tfs.flat[0].node
	tfs.Tree.index = 0

	tfs.Tree.Expand(rootNode)

	// Expect: a/, b/, file-root.txt — dirs first, then files
	if len(rootNode.Children()) != 3 {
		t.Fatalf("expected 3 children, got %d", len(rootNode.Children()))
	}
	if rootNode.Children()[0].Text() != "a" {
		t.Errorf("first child: got %q want %q", rootNode.Children()[0].Text(), "a")
	}
	if rootNode.Children()[1].Text() != "b" {
		t.Errorf("second child: got %q want %q", rootNode.Children()[1].Text(), "b")
	}
	if rootNode.Children()[2].Text() != "file-root.txt" {
		t.Errorf("third child: got %q want %q", rootNode.Children()[2].Text(), "file-root.txt")
	}
}

func TestTreeFS_DirsOnlyHidesFiles(t *testing.T) {
	root := makeFSTree(t)
	tfs := NewTreeFS("fs", "", root, true)
	rootNode := tfs.flat[0].node

	tfs.Tree.Expand(rootNode)

	for _, child := range rootNode.Children() {
		// no file should appear
		if child.Text() == "file-root.txt" {
			t.Fatal("dirsOnly=true should hide files")
		}
	}
	if len(rootNode.Children()) != 2 { // only a/ and b/
		t.Fatalf("expected 2 dir children, got %d", len(rootNode.Children()))
	}
}

func TestTreeFS_SubdirIsLazy(t *testing.T) {
	root := makeFSTree(t)
	tfs := NewTreeFS("fs", "", root, false)
	tfs.Tree.Expand(tfs.flat[0].node) // load root

	// "a" is the first child
	aNode := tfs.flat[0].node.Children()[0]
	if aNode.Leaf() {
		t.Fatal("subdirectory node should not be a leaf (has loader)")
	}
	if len(aNode.Children()) != 0 {
		t.Fatal("subdirectory children should not be loaded yet")
	}
}

func TestTreeFS_RecursiveExpand(t *testing.T) {
	root := makeFSTree(t)
	tfs := NewTreeFS("fs", "", root, false)
	rootNode := tfs.flat[0].node

	tfs.Tree.Expand(rootNode)
	aNode := rootNode.Children()[0] // "a"
	tfs.Tree.Expand(aNode)

	// a/ contains: file-a1.txt, file-a2.txt, sub/
	// dirs first: sub/, then files
	names := make([]string, len(aNode.Children()))
	for i, c := range aNode.Children() {
		names[i] = c.Text()
	}
	want := []string{"sub", "file-a1.txt", "file-a2.txt"}
	for i, w := range want {
		if names[i] != w {
			t.Errorf("a child %d: got %q want %q", i, names[i], w)
		}
	}
}

func TestTreeFS_DataIsAbsolutePath(t *testing.T) {
	root := makeFSTree(t)
	tfs := NewTreeFS("fs", "", root, false)
	tfs.Tree.Expand(tfs.flat[0].node)

	for _, child := range tfs.flat[0].node.Children() {
		path := child.Data().(string)
		if !filepath.IsAbs(path) {
			t.Errorf("child data should be absolute path, got %q", path)
		}
		if path != filepath.Join(root, child.Text()) {
			t.Errorf("child path: got %q want %q", path, filepath.Join(root, child.Text()))
		}
	}
}

func TestTreeFS_PermissionError(t *testing.T) {
	root := t.TempDir()
	restricted := filepath.Join(root, "locked")
	if err := os.Mkdir(restricted, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(restricted, 0o755) }) //nolint

	tfs := NewTreeFS("fs", "", root, false)
	tfs.Tree.Expand(tfs.flat[0].node)
	lockedNode := tfs.flat[0].node.Children()[0]
	tfs.Tree.Expand(lockedNode)

	// Should add an error child instead of panicking
	if len(lockedNode.Children()) != 1 {
		t.Fatalf("expected 1 error-child node, got %d", len(lockedNode.Children()))
	}
}

// ---- SetRoot / SetDirsOnly -------------------------------------------------

func TestTreeFS_SetRoot(t *testing.T) {
	root := makeFSTree(t)
	tfs := NewTreeFS("fs", "", root, false)
	tfs.Tree.Expand(tfs.flat[0].node) // load first root

	other := t.TempDir()
	tfs.SetRoot(other)

	if len(tfs.flat) != 1 {
		t.Fatalf("after SetRoot expected 1 node, got %d", len(tfs.flat))
	}
	if tfs.flat[0].node.Data().(string) != other {
		t.Fatalf("SetRoot: root data should be %q, got %q", other, tfs.flat[0].node.Data())
	}
}

func TestTreeFS_SetDirsOnly(t *testing.T) {
	tfs := NewTreeFS("fs", "", t.TempDir(), false)
	tfs.SetDirsOnly(true)
	if !tfs.DirsOnly() {
		t.Fatal("SetDirsOnly(true) did not take effect")
	}
	tfs.SetDirsOnly(false)
	if tfs.DirsOnly() {
		t.Fatal("SetDirsOnly(false) did not take effect")
	}
}

// ---- Render smoke test -----------------------------------------------------

func TestTreeFS_RenderNoPanel(t *testing.T) {
	tfs := NewTreeFS("fs", "", makeFSTree(t), false)
	tfs.Tree.SetBounds(0, 0, 40, 10)
	tfs.Apply(NewTheme())
	tfs.Tree.Render(newTestRenderer()) // must not panic
}

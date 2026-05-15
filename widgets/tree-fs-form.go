package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// TreeFSForm is the WidgetForm for *TreeFS. Even though TreeFS embeds
// *Tree, the registered widget instance is the inner *Tree (Builder
// adds tfs.Tree to the parent and discards the TreeFS wrapper). The
// form Load and Store therefore operate on a *Tree — Path and
// DirsOnly are surfaced as fields with no live wrapper to read from,
// so they are populated only on first Load when the TreeFS still
// exists and otherwise default to empty / false.
//
// In practice the form is most useful at codegen time where the
// constructor parameters travel via the form rather than the live
// widget.
type TreeFSForm struct {
	ComponentForm

	Path     string `group:"general" label:"Root Path"`
	DirsOnly bool   `group:"general" label:"Directories Only"`
}

func (f *TreeFSForm) Name() string  { return "TreeFS" }
func (f *TreeFSForm) Group() string { return "leaf" }
func (f *TreeFSForm) Help() string  { return "Filesystem-backed Tree with lazy directory loading" }

func (f *TreeFSForm) Load(w core.Widget) {
	switch x := w.(type) {
	case *TreeFS:
		f.ComponentForm.Load(&x.Tree.Component)
		f.Path = x.rootPath
		f.DirsOnly = x.dirsOnly
	case *Tree:
		// Builder-added TreeFS surfaces only its inner Tree; the
		// path / dirs-only state is not recoverable.
		f.ComponentForm.Load(&x.Component)
	}
}

func (f *TreeFSForm) Store(w core.Widget) {
	switch x := w.(type) {
	case *TreeFS:
		f.ComponentForm.Store(&x.Tree.Component)
		if f.Path != "" {
			x.SetRoot(f.Path)
		}
		x.SetDirsOnly(f.DirsOnly)
	case *Tree:
		f.ComponentForm.Store(&x.Component)
	}
}

func (f *TreeFSForm) New() core.Widget {
	path := f.Path
	if path == "" {
		path = "."
	}
	return NewTreeFS("", "", path, f.DirsOnly)
}

func (f *TreeFSForm) Validate(field string) error { return nil }

func (f *TreeFSForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		path := f.Path
		if path == "" {
			path = "."
		}
		_, err := fmt.Fprintf(w, "TreeFS(%q, %q, %t).\n", f.ID, path, f.DirsOnly)
		return err
	})
}

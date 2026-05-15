package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// TreeWidgetsForm is the WidgetForm for *TreeWidgets. The mirrored
// root widget is supplied at runtime by the host application
// (Inspector, designer); the static editing surface here only
// inherits the standard ComponentForm fields. New constructs a
// TreeWidgets with no root — callers patch the root in via
// TreeWidgets.Set after the widget is in the tree.
//
// Like TreeFSForm, the Builder publishes the inner *Tree rather than
// the wrapper, so Load and Store accept either to remain useful for
// runtime editing.
type TreeWidgetsForm struct {
	ComponentForm
}

func (f *TreeWidgetsForm) Name() string  { return "TreeWidgets" }
func (f *TreeWidgetsForm) Group() string { return "leaf" }
func (f *TreeWidgetsForm) Help() string  { return "Tree mirroring a widget hierarchy" }

func (f *TreeWidgetsForm) Load(w core.Widget) {
	switch x := w.(type) {
	case *TreeWidgets:
		f.ComponentForm.Load(&x.Tree.Component)
	case *Tree:
		f.ComponentForm.Load(&x.Component)
	}
}

func (f *TreeWidgetsForm) Store(w core.Widget) {
	switch x := w.(type) {
	case *TreeWidgets:
		f.ComponentForm.Store(&x.Tree.Component)
	case *Tree:
		f.ComponentForm.Store(&x.Component)
	}
}

func (f *TreeWidgetsForm) New() core.Widget {
	return NewTreeWidgets("", "", nil)
}

func (f *TreeWidgetsForm) Validate(field string) error { return nil }

// Emit writes the TreeWidgets constructor with a placeholder root
// identifier. The root widget is application-level state that
// codegen cannot synthesise; the user is expected to replace
// "rootWidget" with the actual variable.
func (f *TreeWidgetsForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "TreeWidgets(%q, rootWidget /* TODO */).\n", f.ID)
		return err
	})
}

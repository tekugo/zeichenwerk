package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// TreeForm is the WidgetForm for *Tree. The node hierarchy is
// runtime data (TreeNode trees are not naturally serialised through
// a struct form), so the static editing surface only covers the
// scrollbar flag.
type TreeForm struct {
	ComponentForm

	Scrollbar bool `group:"display" label:"Show Scrollbar"`
}

func (f *TreeForm) Name() string  { return "Tree" }
func (f *TreeForm) Group() string { return "leaf" }
func (f *TreeForm) Help() string  { return "Scrollable hierarchical list" }

func (f *TreeForm) Load(w core.Widget) {
	t := w.(*Tree)
	f.ComponentForm.Load(&t.Component)
	f.Scrollbar = t.scrollbar
}

func (f *TreeForm) Store(w core.Widget) {
	t := w.(*Tree)
	f.ComponentForm.Store(&t.Component)
	t.scrollbar = f.Scrollbar
}

func (f *TreeForm) New() core.Widget {
	t := NewTree("", "")
	f.Store(t)
	return t
}

func (f *TreeForm) Validate(field string) error { return nil }

func (f *TreeForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Tree(%q).\n", f.ID)
		return err
	})
}

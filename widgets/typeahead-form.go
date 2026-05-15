package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// TypeaheadForm is the WidgetForm for *Typeahead. Typeahead embeds
// Input; the suggestion callback is runtime-only and is not part of
// the static editing surface.
type TypeaheadForm struct {
	ComponentForm

	Text        string `group:"value" label:"Text"`
	Placeholder string `group:"value" label:"Placeholder"`
	Mask        string `group:"value" label:"Mask"`
	Masked      bool   `group:"flags" label:"Masked"`
	Readonly    bool   `group:"flags" label:"Read-only"`
}

func (f *TypeaheadForm) Name() string  { return "Typeahead" }
func (f *TypeaheadForm) Group() string { return "input" }
func (f *TypeaheadForm) Help() string  { return "Single-line text input with inline suggestion" }

func (f *TypeaheadForm) Load(w core.Widget) {
	t := w.(*Typeahead)
	f.ComponentForm.Load(&t.Component)
	f.Text = t.buf.String()
	f.Placeholder = t.placeholder
	f.Mask = t.mask
	f.Masked = t.Flag(core.FlagMasked)
	f.Readonly = t.Flag(core.FlagReadonly)
}

func (f *TypeaheadForm) Store(w core.Widget) {
	t := w.(*Typeahead)
	f.ComponentForm.Store(&t.Component)
	t.buf = core.NewGapBufferFromString(f.Text, 16)
	t.placeholder = f.Placeholder
	if f.Mask != "" {
		t.mask = f.Mask
	}
	t.SetFlag(core.FlagMasked, f.Masked)
	t.SetFlag(core.FlagReadonly, f.Readonly)
}

func (f *TypeaheadForm) New() core.Widget {
	t := NewTypeahead("", "", f.Text, f.Placeholder, f.Mask)
	f.Store(t)
	return t
}

func (f *TypeaheadForm) Validate(field string) error { return nil }

func (f *TypeaheadForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Typeahead(%q%s).\n", f.ID, inputParamArgs(f.Text, f.Placeholder, f.Mask))
		return err
	}); err != nil {
		return err
	}
	if f.Masked {
		fmt.Fprintf(w, "Flag(FlagMasked, true).\n")
	}
	if f.Readonly {
		fmt.Fprintf(w, "Flag(FlagReadonly, true).\n")
	}
	return nil
}

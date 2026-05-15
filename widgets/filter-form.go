package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// FilterForm is the WidgetForm for *Filter. Filter embeds Typeahead;
// the editable surface is the placeholder, mask, and read-only flag.
// The bound widget reference (Filter.bound) is runtime state and not
// part of the static form.
type FilterForm struct {
	ComponentForm

	Placeholder string `group:"value" label:"Placeholder"`
	Mask        string `group:"value" label:"Mask"`
	Masked      bool   `group:"flags" label:"Masked"`
	Readonly    bool   `group:"flags" label:"Read-only"`
}

func (f *FilterForm) Name() string  { return "Filter" }
func (f *FilterForm) Group() string { return "input" }
func (f *FilterForm) Help() string  { return "Filter input that progressively filters a bound List or Tree" }

func (f *FilterForm) Load(w core.Widget) {
	x := w.(*Filter)
	f.ComponentForm.Load(&x.Component)
	f.Placeholder = x.placeholder
	f.Mask = x.mask
	f.Masked = x.Flag(core.FlagMasked)
	f.Readonly = x.Flag(core.FlagReadonly)
}

func (f *FilterForm) Store(w core.Widget) {
	x := w.(*Filter)
	f.ComponentForm.Store(&x.Component)
	x.placeholder = f.Placeholder
	if f.Mask != "" {
		x.mask = f.Mask
	}
	x.SetFlag(core.FlagMasked, f.Masked)
	x.SetFlag(core.FlagReadonly, f.Readonly)
}

func (f *FilterForm) New() core.Widget {
	x := NewFilter("", "")
	f.Store(x)
	return x
}

func (f *FilterForm) Validate(field string) error { return nil }

func (f *FilterForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Filter(%q).\n", f.ID)
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
	if f.Placeholder != "" {
		fmt.Fprintf(w, "// TODO: SetPlaceholder(%q) — no Builder setter\n", f.Placeholder)
	}
	return nil
}

package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// InputForm is the WidgetForm for *Input. Leaf widget — no children,
// no per-child layout params.
//
// Edited fields:
//   - Text: initial text content
//   - Placeholder: text shown when empty
//   - Mask: character used for password masking (default "*")
//   - Max: max length in characters (0 = unlimited)
//   - Masked / Readonly: as flags
type InputForm struct {
	ComponentForm

	Text        string `group:"value" label:"Text"`
	Placeholder string `group:"value" label:"Placeholder"`
	Mask        string `group:"value" label:"Mask"`
	Max         int    `group:"value" label:"Max Length"`
	Masked      bool   `group:"flags" label:"Masked"`
	Readonly    bool   `group:"flags" label:"Read-only"`
}

func (f *InputForm) Name() string  { return "Input" }
func (f *InputForm) Group() string { return "leaf" }
func (f *InputForm) Help() string  { return "Single-line text input" }

func (f *InputForm) Load(w core.Widget) {
	i := w.(*Input)
	f.ComponentForm.Load(&i.Component)
	f.Text = i.buf.String()
	f.Placeholder = i.placeholder
	f.Mask = i.mask
	f.Max = i.max
	f.Masked = i.Flag(core.FlagMasked)
	f.Readonly = i.Flag(core.FlagReadonly)
}

func (f *InputForm) Store(w core.Widget) {
	i := w.(*Input)
	f.ComponentForm.Store(&i.Component)
	i.buf = core.NewGapBufferFromString(f.Text, 16)
	i.pos = 0
	i.offset = 0
	i.placeholder = f.Placeholder
	i.mask = f.Mask
	i.max = f.Max
	i.SetFlag(core.FlagMasked, f.Masked)
	i.SetFlag(core.FlagReadonly, f.Readonly)
}

func (f *InputForm) New() core.Widget {
	i := NewInput("", "", f.Text, f.Placeholder, f.Mask)
	f.Store(i)
	return i
}

func (f *InputForm) Validate(field string) error { return nil }

// Emit writes Input's call shape onto an in-progress chain. The
// Builder's Input signature is Input(id, params...) with
// text/placeholder/mask as trailing variadic strings, so we only
// emit as many as are needed to disambiguate later positional
// arguments.
//
// Masked / Readonly emit as Flag chain entries; Max emits a TODO
// because the Builder has no chained setter for it.
func (f *InputForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Input(%q%s).\n", f.ID, inputParamArgs(f.Text, f.Placeholder, f.Mask))
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
	if f.Max > 0 {
		fmt.Fprintf(w, "// TODO: SetMax(%d) — no Builder setter\n", f.Max)
	}
	return nil
}

// inputParamArgs renders the trailing variadic arguments to Input(id,
// ...). Mask defaults to "*" inside NewInput, so when only the mask
// would differ we still emit text+placeholder to keep parameter
// positions correct. Returns "" when all three are empty / default.
func inputParamArgs(text, placeholder, mask string) string {
	const defaultMask = "*"
	emitMask := mask != "" && mask != defaultMask
	emitPlaceholder := emitMask || placeholder != ""
	emitText := emitPlaceholder || text != ""
	switch {
	case emitMask:
		return fmt.Sprintf(", %q, %q, %q", text, placeholder, mask)
	case emitPlaceholder:
		return fmt.Sprintf(", %q, %q", text, placeholder)
	case emitText:
		return fmt.Sprintf(", %q", text)
	default:
		return ""
	}
}

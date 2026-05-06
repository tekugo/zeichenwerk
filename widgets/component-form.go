package widgets

import (
	"fmt"
	"io"

	. "github.com/tekugo/zeichenwerk/core"
)

// ComponentForm carries the editor + codegen surface for the fields
// owned by Component. The struct-tagged fields (ID, Class, Hint*,
// flags) are the editing surface rendered by the property panel; the
// unexported style field is a per-load snapshot of the widget's
// default style, used purely by codegen so EmitStyle can emit the
// widget-specific styling chain alongside the constructor. The Style
// editor surface is separate (the dedicated Styles tab edits the
// live *Style directly through its own StyleForm).
//
// Each widget's WidgetForm embeds ComponentForm; the embedding gives
// a uniform editing experience for id/class/hint/flags across every
// kind without per-widget repetition. Per-widget forms call
// EmitFrame to wrap their kind-specific constructor with the
// standard prefix (Class) and trailing chain (Hint, flags, style).
type ComponentForm struct {
	ID    string `group:"general" label:"ID" validate:"id-unique"`
	Class string `group:"general" label:"Class"`

	HintW int `group:"layout" label:"Hint W"`
	HintH int `group:"layout" label:"Hint H"`

	Skip     bool `group:"flags" label:"Skip Focus"`
	Hidden   bool `group:"flags" label:"Hidden"`
	Disabled bool `group:"flags" label:"Disabled"`

	// Codegen-only snapshot of the widget's default *Style. No
	// struct tag → invisible to BuildFormGroup. Populated by Load
	// and written back by Store (through Style.Modifiable, so the
	// widget's existing fixed/non-fixed cascade is preserved).
	// EmitStyle delegates here.
	style StyleForm
}

// Load copies state from c into the form fields, including a
// snapshot of c's default style so codegen can emit it.
func (f *ComponentForm) Load(c *Component) {
	f.ID = c.id
	f.Class = c.class
	f.HintW = c.hwidth
	f.HintH = c.hheight

	f.Skip = c.Flag(FlagSkip)
	f.Hidden = c.Flag(FlagHidden)
	f.Disabled = c.Flag(FlagDisabled)

	f.style.Load(c.Style())
}

// Store writes the form fields back into c. The style snapshot is
// applied through Style.Modifiable, so editing a previously-fixed
// (themed) style produces a new non-fixed child rather than
// mutating the theme.
func (f *ComponentForm) Store(c *Component) {
	c.id = f.ID
	c.class = f.Class
	c.hwidth = f.HintW
	c.hheight = f.HintH

	c.SetFlag(FlagSkip, f.Skip)
	c.SetFlag(FlagHidden, f.Hidden)
	c.SetFlag(FlagDisabled, f.Disabled)

	c.SetStyle("", f.style.Store(c.Style()))
}

// Style returns a pointer to the form's style snapshot. The
// inspector's Properties panel renders this through a separate
// Form widget so the user can edit foreground / background /
// border / padding / margin / font alongside the kind-specific
// fields. Because the returned pointer is the same struct Load
// reads into and Store writes out of, edits made through it land
// in the next Store call without explicit syncing.
func (f *ComponentForm) Style() *StyleForm {
	return &f.style
}

// CheckBuilderMode validates that mode is a supported codegen mode.
// Per-form Emit implementations either call this directly or via
// EmitFrame, which calls it before invoking the constructor body.
// Today only ModeBuilder is supported; ModeCompose is reserved and
// returns an error.
func (f *ComponentForm) CheckBuilderMode(mode string) error {
	switch mode {
	case "builder":
		return nil
	case "compose":
		return fmt.Errorf("compose mode not implemented")
	}
	return fmt.Errorf("unknown mode %q", mode)
}

// EmitClassPrefix writes a chained "Class(class).\n" call to set the
// builder's class register before the constructor that follows. The
// trailing ".\n" is part of the chain-element separator convention:
// every emitted chain call ends with ".\n" so the next call
// continues the chain without needing a leading separator. The very
// last chain element in the whole expression has its trailing "."
// stripped by Designer before formatting.
func (f *ComponentForm) EmitClassPrefix(w io.Writer) {
	if f.Class != "" {
		fmt.Fprintf(w, "Class(%q).\n", f.Class)
	}
}

// EmitChain writes the chained Builder methods that follow a
// widget's constructor for everything Component owns *except* style.
// Each call ends with ".\n" so subsequent chain elements concatenate
// directly. Final indentation is left to gofmt.
func (f *ComponentForm) EmitChain(w io.Writer) {
	if f.HintW != 0 || f.HintH != 0 {
		fmt.Fprintf(w, "Hint(%d, %d).\n", f.HintW, f.HintH)
	}
	if f.Skip {
		fmt.Fprintf(w, "Flag(FlagSkip, true).\n")
	}
	if f.Hidden {
		fmt.Fprintf(w, "Flag(FlagHidden, true).\n")
	}
	if f.Disabled {
		fmt.Fprintf(w, "Flag(FlagDisabled, true).\n")
	}
}

// EmitStyle writes the widget-specific styling chain
// (Background/Foreground/Border/Padding/Margin/...). Themed (fixed)
// styles produce no output — the widget inherits them from the
// active theme and emitting them in source would override theme
// changes at the call site.
func (f *ComponentForm) EmitStyle(w io.Writer) {
	f.style.EmitBuilderChain(w)
}

// EmitFrame is the template method per-widget forms call to wrap
// their constructor with the standard prefix-and-chain machinery.
// body writes the kind-specific constructor — typically a single
// ".\nKind(args...)" formatted line — and EmitFrame surrounds it
// with EmitClassPrefix on entry and EmitChain + EmitStyle on exit.
//
// Forms that need a kind-specific tail (Grid emits .Rows / .Columns
// after the standard chain) call EmitFrame for the standard part
// and append the tail manually after EmitFrame returns. Forms that
// need to override the standard prefix shape (Flex picks HFlex vs
// VFlex from a flag) call the primitives directly instead of using
// EmitFrame.
func (f *ComponentForm) EmitFrame(w io.Writer, mode string, body func() error) error {
	if err := f.CheckBuilderMode(mode); err != nil {
		return err
	}
	f.EmitClassPrefix(w)
	if err := body(); err != nil {
		return err
	}
	f.EmitChain(w)
	f.EmitStyle(w)
	return nil
}

package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// IndicatorForm is the WidgetForm for *Indicator. The level uses a
// select control whose options match the styled severity variants
// registered by Apply ("debug", "info", "success", "warning",
// "error", "fatal"). "success" has no constant in core; the codegen
// path falls back to a Level("success") cast in that case.
type IndicatorForm struct {
	ComponentForm

	Level string `group:"general" label:"Level" control:"select" options:"debug,info,success,warning,error,fatal"`
	Label string `group:"general" label:"Label"`
}

func (f *IndicatorForm) Name() string  { return "Indicator" }
func (f *IndicatorForm) Group() string { return "leaf" }
func (f *IndicatorForm) Help() string  { return "Coloured status glyph paired with a label" }

func (f *IndicatorForm) Load(w core.Widget) {
	i := w.(*Indicator)
	f.ComponentForm.Load(&i.Component)
	f.Level = string(i.level)
	f.Label = i.label
}

func (f *IndicatorForm) Store(w core.Widget) {
	i := w.(*Indicator)
	f.ComponentForm.Store(&i.Component)
	i.level = core.Level(f.Level)
	i.label = f.Label
}

func (f *IndicatorForm) New() core.Widget {
	i := NewIndicator("", "", core.Level(f.Level), f.Label)
	f.Store(i)
	return i
}

func (f *IndicatorForm) Validate(field string) error { return nil }

func (f *IndicatorForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Indicator(%q, %s, %q).\n", f.ID, levelConst(f.Level), f.Label)
		return err
	})
}

// levelConst maps a level string to its exported core constant name,
// or falls back to a Level("…") cast for non-constant values such as
// "success" that the theme registers but core does not name.
func levelConst(s string) string {
	switch s {
	case "debug":
		return "Debug"
	case "info":
		return "Info"
	case "warning":
		return "Warning"
	case "error":
		return "Error"
	case "fatal":
		return "Fatal"
	default:
		return fmt.Sprintf("Level(%q)", s)
	}
}

package widgets

import (
	"fmt"
	"io"
	"time"

	"github.com/tekugo/zeichenwerk/core"
)

// TypewriterForm is the WidgetForm for *Typewriter. Dwell and
// Interval are surfaced as duration strings parsed via
// time.ParseDuration. Reveal state (shown / dwellTick / cursorOn) is
// runtime-only.
type TypewriterForm struct {
	ComponentForm

	Text       string `group:"general" label:"Text"`
	Rate       int    `group:"general" label:"Rate"`
	ShowCursor bool   `group:"general" label:"Show Cursor"`
	Dwell      string `group:"general" label:"Dwell"`
	Repeat     bool   `group:"general" label:"Repeat"`
	Interval   string `group:"general" label:"Interval"`
}

func (f *TypewriterForm) Name() string  { return "Typewriter" }
func (f *TypewriterForm) Group() string { return "leaf" }
func (f *TypewriterForm) Help() string  { return "Animated text reveal effect" }

func (f *TypewriterForm) Load(w core.Widget) {
	tw := w.(*Typewriter)
	f.ComponentForm.Load(&tw.Component)
	f.Text = tw.text
	f.Rate = tw.rate
	f.ShowCursor = tw.showCursor
	f.Dwell = tw.dwell.String()
	f.Repeat = tw.repeat
	if tw.interval > 0 {
		f.Interval = tw.interval.String()
	}
}

func (f *TypewriterForm) Store(w core.Widget) {
	tw := w.(*Typewriter)
	f.ComponentForm.Store(&tw.Component)
	tw.SetText(f.Text)
	if f.Rate > 0 {
		tw.SetRate(f.Rate)
	}
	tw.SetCursor(f.ShowCursor)
	if d, err := time.ParseDuration(f.Dwell); err == nil && d >= 0 {
		tw.SetDwell(d)
	}
	tw.SetRepeat(f.Repeat)
	if d, err := time.ParseDuration(f.Interval); err == nil && d > 0 {
		tw.interval = d
	}
}

func (f *TypewriterForm) New() core.Widget {
	tw := NewTypewriter("", "")
	f.Store(tw)
	return tw
}

func (f *TypewriterForm) Validate(field string) error {
	switch field {
	case "Dwell":
		if f.Dwell != "" {
			if _, err := time.ParseDuration(f.Dwell); err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}
		}
	case "Interval":
		if f.Interval != "" {
			if _, err := time.ParseDuration(f.Interval); err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}
		}
	}
	return nil
}

// Emit writes the Typewriter constructor. The configuration setters
// (SetText, SetRate, SetCursor, SetDwell, SetRepeat) are not chained
// on the Builder, so changes from defaults emit as TODO comments.
func (f *TypewriterForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Typewriter(%q).\n", f.ID)
		return err
	}); err != nil {
		return err
	}
	if f.Text != "" {
		fmt.Fprintf(w, "// TODO: SetText(%q) — no Builder setter\n", f.Text)
	}
	if f.Rate > 1 {
		fmt.Fprintf(w, "// TODO: SetRate(%d) — no Builder setter\n", f.Rate)
	}
	if !f.ShowCursor {
		fmt.Fprintf(w, "// TODO: SetCursor(false) — no Builder setter\n")
	}
	if f.Repeat {
		fmt.Fprintf(w, "// TODO: SetRepeat(true) — no Builder setter\n")
	}
	if f.Dwell != "" && f.Dwell != "500ms" {
		fmt.Fprintf(w, "// TODO: SetDwell(%q) — no Builder setter\n", f.Dwell)
	}
	return nil
}

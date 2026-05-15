package widgets

import (
	"fmt"
	"io"
	"time"

	"github.com/tekugo/zeichenwerk/core"
)

// marqueeDefaultInterval is the tick interval the form uses when
// the Running checkbox is toggled on. Marquee itself does not
// persist an interval, so the form picks a reasonable default for
// interactive testing.
const marqueeDefaultInterval = 100 * time.Millisecond

// MarqueeForm is the WidgetForm for *Marquee. Speed and gap have
// SetSpeed / SetGap clamps; the form clamps Speed to >= 1 and Gap to
// >= 0 in Store so out-of-range edits round-trip cleanly.
//
// Running drives the embedded Animation: toggling it through Apply
// calls Start with marqueeDefaultInterval (or Stop) so the designer
// can test the animation interactively. Marquee does not persist an
// interval — this is a runtime-only convenience and is not
// reflected in codegen.
type MarqueeForm struct {
	ComponentForm

	Text    string `group:"general" label:"Text"`
	Speed   int    `group:"general" label:"Speed"`
	Gap     int    `group:"general" label:"Gap"`
	Running bool   `group:"general" label:"Running"`
}

func (f *MarqueeForm) Name() string  { return "Marquee" }
func (f *MarqueeForm) Group() string { return "leaf" }
func (f *MarqueeForm) Help() string  { return "Single-row scrolling text ticker" }

func (f *MarqueeForm) Load(w core.Widget) {
	m := w.(*Marquee)
	f.ComponentForm.Load(&m.Component)
	f.Text = m.text
	f.Speed = m.speed
	f.Gap = m.gap
	f.Running = m.Running()
}

func (f *MarqueeForm) Store(w core.Widget) {
	m := w.(*Marquee)
	f.ComponentForm.Store(&m.Component)
	m.SetText(f.Text)
	speed := f.Speed
	if speed < 1 {
		speed = 1
	}
	m.SetSpeed(speed)
	gap := f.Gap
	if gap < 0 {
		gap = 0
	}
	m.SetGap(gap)

	// Toggle the embedded Animation only when the requested state
	// differs from the live state — Animation.Start logs an error
	// when called twice, and Stop is a no-op when nothing is
	// running, but skipping the call avoids the noise.
	switch {
	case f.Running && !m.Running():
		m.Start(marqueeDefaultInterval)
	case !f.Running && m.Running():
		m.Stop()
	}
}

func (f *MarqueeForm) New() core.Widget {
	m := NewMarquee("", "")
	f.Store(m)
	return m
}

func (f *MarqueeForm) Validate(field string) error { return nil }

func (f *MarqueeForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Marquee(%q).\n", f.ID)
		return err
	}); err != nil {
		return err
	}
	if f.Text != "" {
		fmt.Fprintf(w, "// TODO: SetText(%q) — no Builder setter\n", f.Text)
	}
	if f.Speed > 1 {
		fmt.Fprintf(w, "// TODO: SetSpeed(%d) — no Builder setter\n", f.Speed)
	}
	if f.Gap != 4 {
		fmt.Fprintf(w, "// TODO: SetGap(%d) — no Builder setter\n", f.Gap)
	}
	return nil
}

// Tutorial chapter 8 — Writing a custom widget.
//
// `Status` is a tiny widget that draws a coloured dot followed by a label,
// e.g. "● Online" or "● Connecting". It demonstrates the four things every
// custom widget needs: a struct embedding *Component, Render, Apply, and
// Hint.
//
// Run with:  go run ./doc/tutorial/examples/08-custom-widget
package main

import (
	"unicode/utf8"

	zw "github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
	"github.com/tekugo/zeichenwerk/v2/widgets"
)

// Status displays "● Label" with the dot coloured by Stat.
//
// Stat is a free-form string that maps to a theme colour:
//
//	"ok"   → $green
//	"warn" → $yellow
//	"err"  → $red
//	other  → $gray
type Status struct {
	*widgets.Component         // embed pointer; gives us all Widget methods
	Stat               string  // current state — not "State" because Component has a State() method
	Label              string
}

// NewStatus follows the convention: id, class, then widget-specific args.
func NewStatus(id, class, stat, label string) *Status {
	return &Status{
		Component: widgets.NewComponent(id, class),
		Stat:      stat,
		Label:     label,
	}
}

// Apply hooks the widget into the theme. The selector "status" is what the
// theme uses to look up default colours/font. Always include the type name
// so themes can target your widget.
func (s *Status) Apply(theme *core.Theme) {
	theme.Apply(s, s.Selector("status"))
}

// Hint reports the natural size: 2 cells for "● " plus the label width.
func (s *Status) Hint() (int, int) {
	return 2 + utf8.RuneCountInString(s.Label), 1
}

// Render is where the widget draws itself.
//
//   - Always call Component.Render(r) first — that draws margin, border,
//     and background defined by the theme.
//   - The drawable area is reported by Content(); never draw past it.
//   - r.Set(fg, bg, font) is sticky — call it before each batch of draws.
func (s *Status) Render(r *core.Renderer) {
	s.Component.Render(r)

	x, y, w, _ := s.Content()
	if w <= 0 {
		return
	}

	style := s.Style()
	bg := style.Background()

	// 1) coloured dot
	dot := "$gray"
	switch s.Stat {
	case "ok":
		dot = "$green"
	case "warn":
		dot = "$yellow"
	case "err":
		dot = "$red"
	}
	r.Set(dot, bg, "")
	r.Text(x, y, "●", 1)

	// 2) label in the default foreground
	r.Set(style.Foreground(), bg, style.Font())
	r.Text(x+2, y, s.Label, w-2)
}

// Set updates the state and queues a redraw. Following the framework
// convention, Set never fires EvtChange.
func (s *Status) Set(stat, label string) {
	s.Stat = stat
	s.Label = label
	widgets.Redraw(s)
}

// Custom widgets land in the tree via Builder.Add — there is no built-in
// Builder method for them. (Adding one is the only thing that distinguishes
// "framework" widgets from "your" widgets.)
func main() {
	statuses := []*Status{
		NewStatus("api", "", "ok", "API server"),
		NewStatus("db", "", "warn", "Database (slow)"),
		NewStatus("worker", "", "err", "Worker offline"),
		NewStatus("cache", "", "unknown", "Cache (probing)"),
	}

	b := zw.NewBuilder(themes.TokyoNight()).
		VFlex("root", core.Stretch, 1).Padding(1, 2).
		Static("title", "Custom widget demo").
		Font("bold").Foreground("$cyan")

	for _, s := range statuses {
		b.Add(s) // <-- this is how custom widgets land in the tree
	}

	b.HRule("thin").
		Static("hint", "(custom Status widget — see source)").
		Foreground("$gray").
		End().
		Run()
}

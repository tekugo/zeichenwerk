package widgets

import (
	"fmt"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// Radio is a vertical group of mutually-exclusive options. It mirrors the
// public surface of [Select] (constructor, Select/Value/Text/Summary), but
// renders every option on its own row instead of hiding the choices behind a
// dropdown. The currently chosen option is the only state — there is no
// separate cursor, so any navigation key or mouse click on a row immediately
// changes the selection and dispatches [EvtChange].
//
// The on/off glyphs are read from the active theme's string registry under
// the keys "radio.on" and "radio.off". Both keys accept arbitrary rune
// widths (e.g. "(•)"/"( )" at 3 cells, or "◉"/"○" at 1 cell) — the widget
// pads the narrower glyph so labels stay aligned regardless of the theme.
type Radio struct {
	Component
	index   int
	options []option
}

// NewRadio creates a Radio widget with the given options. args is a flat
// list of alternating value/text pairs, exactly like [NewSelect]:
//
//	NewRadio("size", "", "s", "Small", "m", "Medium", "l", "Large")
//
// The first option is selected by default.
func NewRadio(id, class string, args ...string) *Radio {
	r := &Radio{
		Component: Component{id: id, class: class},
		options:   make([]option, 0, len(args)/2),
	}
	r.SetFlag(FlagFocusable, true)
	for i := 0; i+1 < len(args); i += 2 {
		r.options = append(r.options, option{value: args[i], text: args[i+1]})
	}
	OnKey(r, r.handleKey)
	OnMouse(r, r.handleMouse)
	return r
}

// Apply installs the theme styles. Beyond the usual widget states the radio
// also installs a "radio/selected" part (and its :focused variant) used for
// the currently chosen row.
func (r *Radio) Apply(theme *Theme) {
	theme.Apply(r, r.Selector("radio"), "disabled", "focused", "hovered")
	theme.Apply(r, r.Selector("radio/selected"), "focused")
}

// Hint returns the preferred size. Width is the longest option label plus a
// four-cell prefix budget (worst case for "(•) "/"[x] "); height is the
// number of options.
func (r *Radio) Hint() (int, int) {
	if r.hwidth != 0 || r.hheight != 0 {
		return r.hwidth, r.hheight
	}
	mw := 0
	for _, o := range r.options {
		mw = max(mw, utf8.RuneCountInString(o.text))
	}
	return mw + 4, len(r.options)
}

// Render draws every option on its own row. The selected row uses the
// "selected" part style (focus-aware); every other row uses the widget's
// base style. The on/off glyphs are looked up in the theme on each render so
// theme switches take effect without rebuilding the widget.
func (r *Radio) Render(rd *Renderer) {
	if r.Flag(FlagHidden) {
		return
	}

	// Base widget paint (background + border).
	r.Component.Render(rd)

	on := rd.Theme.String("radio.on")
	if on == "" {
		on = "(•)"
	}
	off := rd.Theme.String("radio.off")
	if off == "" {
		off = "( )"
	}
	onW := utf8.RuneCountInString(on)
	offW := utf8.RuneCountInString(off)
	gw := max(onW, offW)

	x, y, w, h := r.Content()
	if w <= 0 || h <= 0 {
		return
	}
	focused := r.Flag(FlagFocused)

	for i, opt := range r.options {
		if i >= h {
			break
		}

		// Resolve row style.
		var style *Style
		if i == r.index {
			if focused {
				style = r.Style("selected:focused")
			} else {
				style = r.Style("selected")
			}
		} else {
			style = r.Style()
		}
		rd.Set(style.Foreground(), style.Background(), style.Font())

		// Choose glyph and pad to gw so labels stay column-aligned.
		glyph := off
		glyphW := offW
		if i == r.index {
			glyph = on
			glyphW = onW
		}
		rd.Text(x, y+i, glyph, glyphW)
		if pad := gw - glyphW; pad > 0 {
			rd.Fill(x+glyphW, y+i, pad, 1, " ")
		}

		// Render label after one space separator; pad the rest of the row so
		// the row-wide style (selected highlight) covers the full width.
		labelX := x + gw + 1
		labelW := w - gw - 1
		if labelW > 0 {
			rd.Text(labelX, y+i, opt.text, labelW)
		}
		// Separator space between glyph and label.
		if labelX-1 >= x {
			rd.Put(labelX-1, y+i, " ")
		}
	}
}

// Select sets the chosen option by value. An unknown value resets the
// selection to the first option, matching [Select.Select]. This setter does
// not dispatch [EvtChange] (mirroring the Select behaviour); programmatic
// updates are silent so callers can initialise without triggering handlers.
func (r *Radio) Select(value string) {
	r.index = 0
	for i, o := range r.options {
		if o.value == value {
			r.index = i
			return
		}
	}
}

// Text returns the display text of the currently selected option.
func (r *Radio) Text() string {
	if len(r.options) == 0 {
		return ""
	}
	return r.options[r.index].text
}

// Value returns the value of the currently selected option.
func (r *Radio) Value() string {
	if len(r.options) == 0 {
		return ""
	}
	return r.options[r.index].value
}

// Summary returns the currently selected value for Dump output.
func (r *Radio) Summary() string {
	if len(r.options) == 0 {
		return ""
	}
	return fmt.Sprintf("selected=%q", r.options[r.index].value)
}

// setIndex moves the selection to i and dispatches [EvtChange] when the
// value actually changes. Disabled or read-only widgets reject the update.
func (r *Radio) setIndex(i int) {
	if r.Flag(FlagReadonly) || r.Flag(FlagDisabled) {
		return
	}
	if i < 0 || i >= len(r.options) || i == r.index {
		return
	}
	r.index = i
	r.Dispatch(r, EvtChange, r.Value())
	Redraw(r)
}

// handleKey processes keyboard navigation. Because the radio has no
// independent cursor, every navigation key changes the selection.
//
// Supported keys:
//   - Up / k:       previous option
//   - Down / j:     next option
//   - Home:         first option
//   - End:          last option
func (r *Radio) handleKey(evt *tcell.EventKey) bool {
	if r.Flag(FlagReadonly) || r.Flag(FlagDisabled) || len(r.options) == 0 {
		return false
	}
	switch evt.Key() {
	case tcell.KeyUp:
		if r.index > 0 {
			r.setIndex(r.index - 1)
		}
		return true
	case tcell.KeyDown:
		if r.index < len(r.options)-1 {
			r.setIndex(r.index + 1)
		}
		return true
	case tcell.KeyHome:
		r.setIndex(0)
		return true
	case tcell.KeyEnd:
		r.setIndex(len(r.options) - 1)
		return true
	case tcell.KeyRune:
		switch evt.Str() {
		case "k":
			if r.index > 0 {
				r.setIndex(r.index - 1)
			}
			return true
		case "j":
			if r.index < len(r.options)-1 {
				r.setIndex(r.index + 1)
			}
			return true
		}
	}
	return false
}

// handleMouse selects the row under the cursor on left-button click.
func (r *Radio) handleMouse(evt *tcell.EventMouse) bool {
	if r.Flag(FlagReadonly) || r.Flag(FlagDisabled) || len(r.options) == 0 {
		return false
	}
	if evt.Buttons() != tcell.Button1 {
		return false
	}
	mx, my := evt.Position()
	cx, cy, cw, _ := r.Content()
	if mx < cx || mx >= cx+cw {
		return false
	}
	row := my - cy
	if row < 0 || row >= len(r.options) {
		return false
	}
	r.setIndex(row)
	return true
}

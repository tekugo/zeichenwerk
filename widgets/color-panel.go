package widgets

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/tekugo/zeichenwerk/core"
)

// RGB is an 8-bit-per-channel sRGB colour. It is the source of truth used by
// the colour picker components; HSL and Hex representations are derived
// from RGB on demand and rounded for display, so repeated round-trips do
// not drift.
type RGB struct{ R, G, B uint8 }

// Hex returns the colour as a lowercase "#RRGGBB" string.
func (c RGB) Hex() string {
	return FormatHex(c.R, c.G, c.B)
}

// ColorPanel is one colour editor: a 3-row preview swatch followed by RGB,
// HSL, and Hex inputs. Editing any input updates the other representations
// and the swatch via the cross-channel handlers. The panel emits EvtChange
// whenever the colour changes (whether by user edit or programmatic
// SetColor) with itself as the payload.
//
// The panel reserves an outer 24×9 area (inner 22×7) under default sizing
// and is the building block of the ColorPicker composite widget.
type ColorPanel struct {
	*Box

	color RGB

	swatch *Static
	inR    *Input
	inG    *Input
	inB    *Input
	inH    *Input
	inS    *Input
	inL    *Input
	inHex  *Input
}

// NewColorPanel creates a colour editor panel. title is shown in the box
// title bar and may be empty.
func NewColorPanel(id, class, title string) *ColorPanel {
	cp := &ColorPanel{
		Box: NewBox(id, class, title),
	}
	// Content size 22×7; the box border adds 2 cols and 2 rows for outer 24×9.
	cp.Box.SetHint(22, 7)
	cp.build()
	cp.refresh()
	return cp
}

// build constructs the inner widget tree. All sub-widgets are owned by
// cp.Box's child Flex; cp keeps direct pointers to the editable parts so
// the change handlers can read and write them without extra lookups.
func (cp *ColorPanel) build() {
	id := cp.Box.ID()

	// Vertical inner flex
	body := NewFlex(id+"-body", "", Stretch, 0)
	body.SetFlag(FlagVertical, true)
	cp.Box.Add(body)

	// Swatch — content height 1; the swatch style adds padding 1,0 so the
	// swatch takes up 3 rows.
	cp.swatch = NewStatic(id+"-swatch", "", "")
	cp.swatch.SetHint(-1, 1)
	body.Add(cp.swatch)

	// Rule between swatch and numeric block
	rule := NewHRule("", "thin")
	rule.SetHint(-1, 1)
	body.Add(rule)

	// RGB row
	cp.inR = newColorInput(id+"-r", 3)
	cp.inG = newColorInput(id+"-g", 3)
	cp.inB = newColorInput(id+"-b", 3)
	body.Add(numericRow(id+"-rgb", "R", "G", "B", cp.inR, cp.inG, cp.inB))

	// HSL row
	cp.inH = newColorInput(id+"-h", 3)
	cp.inS = newColorInput(id+"-s", 3)
	cp.inL = newColorInput(id+"-l", 3)
	body.Add(numericRow(id+"-hsl", "H", "S", "L", cp.inH, cp.inS, cp.inL))

	// Hex row
	cp.inHex = NewInput(id+"-hex", "", "#000000")
	cp.inHex.SetHint(7, 1)
	hexRow := NewFlex(id+"-hex-row", "", Stretch, 0)
	hexRow.SetHint(-1, 1)
	hexRow.Add(staticCell("Hex ", 4))
	hexRow.Add(cp.inHex)
	hexRow.Add(staticCell("", 11)) // tail spacer to fill remaining 11 cols
	body.Add(hexRow)

	cp.inR.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { cp.applyRGB(); return false })
	cp.inG.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { cp.applyRGB(); return false })
	cp.inB.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { cp.applyRGB(); return false })
	cp.inH.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { cp.applyHSL(); return false })
	cp.inS.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { cp.applyHSL(); return false })
	cp.inL.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { cp.applyHSL(); return false })
	cp.inHex.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { cp.applyHex(); return false })
}

// newColorInput constructs an Input pre-sized for a numeric colour channel.
func newColorInput(id string, _ int) *Input {
	in := NewInput(id, "", "0")
	in.SetHint(4, 1)
	return in
}

// staticCell returns a fixed-width Static used as a label or spacer cell.
func staticCell(text string, width int) *Static {
	s := NewStatic("", "", text)
	s.SetHint(width, 1)
	return s
}

// numericRow lays out a single row of three label/input groups.
//   - Each group: 1-col label + 1-col separator + 4-col input.
//   - Between groups: 2 cols of blank space.
//
// Total: 3*(1+1+4) + 2*2 = 22 cols.
func numericRow(id, l1, l2, l3 string, in1, in2, in3 *Input) *Flex {
	row := NewFlex(id, "", Stretch, 0)
	row.SetHint(-1, 1)

	row.Add(staticCell(l1+" ", 2))
	row.Add(in1)
	row.Add(staticCell("  ", 2))
	row.Add(staticCell(l2+" ", 2))
	row.Add(in2)
	row.Add(staticCell("  ", 2))
	row.Add(staticCell(l3+" ", 2))
	row.Add(in3)
	return row
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies the colorpanel theme styles to the panel and recursively
// applies theme styles to every inner widget.
func (cp *ColorPanel) Apply(theme *Theme) {
	theme.Apply(cp.Box, cp.Box.Selector("colorpanel"))
	theme.Apply(cp.Box, cp.Box.Selector("colorpanel/title"))
	Traverse(cp.Box, func(w Widget) bool {
		w.Apply(theme)
		return true
	})
	cp.styleSwatch()
}

// ---- Public API -----------------------------------------------------------

// SetColor parses a "#RGB" or "#RRGGBB" string and updates the panel. On a
// parse error nothing changes. Dispatches EvtChange when the colour was
// updated.
func (cp *ColorPanel) SetColor(hex string) {
	r, g, b, ok := ParseHexColor(strings.TrimSpace(hex))
	if !ok {
		return
	}
	cp.setRGB(RGB{r, g, b}, true)
}

// SetRGB updates the panel to the given RGB colour and dispatches
// EvtChange.
func (cp *ColorPanel) SetRGB(c RGB) {
	cp.setRGB(c, true)
}

// Color returns the current colour as a "#RRGGBB" string.
func (cp *ColorPanel) Color() string {
	return cp.color.Hex()
}

// RGB returns the current colour as RGB.
func (cp *ColorPanel) RGB() RGB {
	return cp.color
}

// SetTitle updates the title shown in the box header.
func (cp *ColorPanel) SetTitle(title string) {
	cp.Box.Title = title
	cp.Box.Refresh()
}

// ---- Internal -------------------------------------------------------------

// setRGB updates the stored colour, refreshes every input/swatch from it,
// and dispatches EvtChange when emit is true.
func (cp *ColorPanel) setRGB(c RGB, emit bool) {
	if c == cp.color {
		// Still refresh in case inputs have been edited to malformed values
		cp.refresh()
		return
	}
	cp.color = c
	cp.refresh()
	if emit {
		cp.Box.Dispatch(cp, EvtChange, cp)
	}
}

// applyRGB reads R/G/B inputs, updates the panel colour, and refreshes the
// remaining inputs and swatch. Out-of-range values clamp to [0, 255].
func (cp *ColorPanel) applyRGB() {
	r, okR := parseChannel(cp.inR.Get(), 255)
	g, okG := parseChannel(cp.inG.Get(), 255)
	b, okB := parseChannel(cp.inB.Get(), 255)
	cp.markError(cp.inR, !okR)
	cp.markError(cp.inG, !okG)
	cp.markError(cp.inB, !okB)
	if !okR || !okG || !okB {
		return
	}
	cp.setRGB(RGB{uint8(r), uint8(g), uint8(b)}, true)
}

// applyHSL reads H/S/L inputs, converts to RGB, and refreshes the panel.
func (cp *ColorPanel) applyHSL() {
	h, okH := parseChannel(cp.inH.Get(), 360)
	s, okS := parseChannel(cp.inS.Get(), 100)
	l, okL := parseChannel(cp.inL.Get(), 100)
	cp.markError(cp.inH, !okH)
	cp.markError(cp.inS, !okS)
	cp.markError(cp.inL, !okL)
	if !okH || !okS || !okL {
		return
	}
	r, g, b := HSLToRGB(float64(h), float64(s), float64(l))
	cp.setRGB(RGB{r, g, b}, true)
}

// applyHex reads the Hex input and propagates a change only when the value
// has 3 or 6 hex digits (the leading "#" is optional). Intermediate states
// such as "#f" or "#ffff" are ignored without flagging an error so the
// other inputs do not flicker while the user is typing.
func (cp *ColorPanel) applyHex() {
	digits := strings.TrimPrefix(strings.TrimSpace(cp.inHex.Get()), "#")
	if len(digits) != 3 && len(digits) != 6 {
		// Partial input: leave error styling untouched and do nothing.
		cp.markError(cp.inHex, false)
		return
	}
	r, g, b, ok := ParseHexColor("#" + digits)
	cp.markError(cp.inHex, !ok)
	if !ok {
		return
	}
	cp.setRGB(RGB{r, g, b}, true)
}

// refresh rewrites every input from cp.color and updates the swatch. The
// input that currently has keyboard focus is left untouched so that the
// user can keep typing without the field jumping under their cursor;
// callers that want to overwrite the focused field anyway can call
// in.Set(...) directly.
func (cp *ColorPanel) refresh() {
	c := cp.color

	// RGB
	setIfNotFocused(cp.inR, formatChannel(int(c.R)))
	setIfNotFocused(cp.inG, formatChannel(int(c.G)))
	setIfNotFocused(cp.inB, formatChannel(int(c.B)))

	// HSL
	h, s, l := RGBToHSL(c.R, c.G, c.B)
	setIfNotFocused(cp.inH, formatChannel(int(h+0.5)))
	setIfNotFocused(cp.inS, formatChannel(int(s+0.5)))
	setIfNotFocused(cp.inL, formatChannel(int(l+0.5)))

	// Hex
	setIfNotFocused(cp.inHex, c.Hex())

	// Clear any error styling
	cp.markError(cp.inR, false)
	cp.markError(cp.inG, false)
	cp.markError(cp.inB, false)
	cp.markError(cp.inH, false)
	cp.markError(cp.inS, false)
	cp.markError(cp.inL, false)
	cp.markError(cp.inHex, false)

	// Swatch
	cp.styleSwatch()
}

// setIfNotFocused updates an Input's text only if it does not currently
// hold keyboard focus. This avoids overwriting whatever the user is in the
// middle of typing while still keeping the other fields synchronised.
func setIfNotFocused(in *Input, text string) {
	if in == nil || in.Flag(FlagFocused) {
		return
	}
	in.Set(text)
}

// styleSwatch applies the current colour as the swatch background.
func (cp *ColorPanel) styleSwatch() {
	if cp.swatch == nil {
		return
	}
	style := NewStyle("colorpanel/swatch").
		WithBackground(cp.color.Hex()).
		WithForeground(cp.color.Hex()).
		WithPadding(1, 0)
	cp.swatch.SetStyle("", style)
	Redraw(cp.swatch)
}

// markError toggles the error style on an input.
func (cp *ColorPanel) markError(in *Input, on bool) {
	if in == nil {
		return
	}
	if on {
		theme := findTheme(cp.Box)
		if theme != nil {
			theme.Apply(in, in.Selector("input.error"))
		} else {
			in.SetStyle("", NewStyle().WithColors("#ffffff", "#ff5555"))
		}
		Redraw(in)
	} else {
		theme := findTheme(cp.Box)
		if theme != nil {
			theme.Apply(in, in.Selector("input"), "focused", "hovered", "disabled")
		}
		Redraw(in)
	}
}

// findTheme walks up the widget tree to find the root UI's theme.
func findTheme(w Widget) *Theme {
	root := FindRoot(w)
	if root == nil {
		return nil
	}
	return root.Theme()
}

// parseChannel parses a channel value as an integer in [0, max]. Empty
// strings parse as 0; out-of-range values are reported as errors.
func parseChannel(s string, max int) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, true
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	if v < 0 || v > max {
		return v, false
	}
	return v, true
}

// formatChannel formats an integer as a right-padded 3-character string.
func formatChannel(v int) string {
	return fmt.Sprintf("%d", v)
}

package widgets

import (
	"fmt"

	. "github.com/tekugo/zeichenwerk/core"
)

// ColorPreview renders a 3-row foreground-on-background swatch with the
// word "Sample" centred, followed by the WCAG 2.1 contrast ratio between
// the two colours. It is a pure display component — no events are emitted
// — and is driven by the parent ColorPicker via SetForeground,
// SetBackground, or SetColors.
//
// Default outer dimensions are 15×9 (inner 13×7), matching the height of
// a ColorPanel so the two can sit side by side without padding.
type ColorPreview struct {
	*Box

	fg RGB
	bg RGB

	swatch        *Static
	contrastLabel *Static
}

// NewColorPreview creates a preview panel. The default colours are black
// foreground on white background, giving a contrast ratio of 21.0.
func NewColorPreview(id, class string) *ColorPreview {
	pp := &ColorPreview{
		Box: NewBox(id, class, "Preview"),
		fg:  RGB{0, 0, 0},
		bg:  RGB{255, 255, 255},
	}
	// Content size 13×7; the box border adds 2 cols and 2 rows for outer 15×9.
	pp.Box.SetHint(13, 7)
	pp.build()
	pp.Refresh()
	return pp
}

// build constructs the inner widget tree.
func (cp *ColorPreview) build() {
	id := cp.Box.ID()

	body := NewFlex(id+"-body", "", Stretch, 0)
	body.SetFlag(FlagVertical, true)
	cp.Box.Add(body)

	cp.swatch = NewStatic(id+"-swatch", "", "Sample")
	cp.swatch.SetAlignment("center")
	// content height 1; the swatch style adds padding 1,0 to fill 3 rows
	cp.swatch.SetHint(-1, 1)
	body.Add(cp.swatch)

	rule := NewHRule("", "thin")
	rule.SetHint(-1, 1)
	body.Add(rule)

	// Empty row above the contrast label
	body.Add(staticCell("", -1))

	cp.contrastLabel = NewStatic(id+"-contrast", "", "")
	cp.contrastLabel.SetHint(-1, 1)
	body.Add(cp.contrastLabel)

	// Empty row below
	body.Add(staticCell("", -1))
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies the previewpanel theme styles to the panel and recursively
// applies theme styles to every inner widget.
func (cp *ColorPreview) Apply(theme *Theme) {
	theme.Apply(cp.Box, cp.Box.Selector("previewpanel"))
	theme.Apply(cp.Box, cp.Box.Selector("previewpanel/title"))
	Traverse(cp.Box, func(w Widget) bool {
		w.Apply(theme)
		return true
	})
	cp.styleSwatch()
	cp.styleContrast()
}

// ---- Public API -----------------------------------------------------------

// SetColors updates both colours and refreshes the swatch and the contrast
// label.
func (cp *ColorPreview) SetColors(fg, bg RGB) {
	cp.fg = fg
	cp.bg = bg
	cp.Refresh()
}

// SetForeground updates the foreground colour only.
func (cp *ColorPreview) SetForeground(fg RGB) {
	cp.fg = fg
	cp.Refresh()
}

// SetBackground updates the background colour only.
func (cp *ColorPreview) SetBackground(bg RGB) {
	cp.bg = bg
	cp.Refresh()
}

// Foreground returns the current foreground colour.
func (cp *ColorPreview) Foreground() RGB { return cp.fg }

// Background returns the current background colour.
func (cp *ColorPreview) Background() RGB { return cp.bg }

// Contrast returns the WCAG 2.1 contrast ratio between fg and bg.
func (cp *ColorPreview) Contrast() float64 {
	return ContrastRatio(cp.fg.R, cp.fg.G, cp.fg.B, cp.bg.R, cp.bg.G, cp.bg.B)
}

// ---- Internal -------------------------------------------------------------

// Refresh redraws the swatch and the contrast label.
func (cp *ColorPreview) Refresh() {
	cp.styleSwatch()
	cp.styleContrast()
}

// styleSwatch repaints the preview swatch with the current fg/bg colours.
func (cp *ColorPreview) styleSwatch() {
	if cp.swatch == nil {
		return
	}
	cp.swatch.SetStyle("", NewStyle("previewpanel/swatch").
		WithColors(cp.fg.Hex(), cp.bg.Hex()).
		WithFont("bold").
		WithPadding(1, 0))
	Redraw(cp.swatch)
}

// styleContrast updates the contrast label text and the ok/warn style.
func (cp *ColorPreview) styleContrast() {
	if cp.contrastLabel == nil {
		return
	}
	ratio := cp.Contrast()
	cp.contrastLabel.Set(fmt.Sprintf("Contrast %4.1f", ratio))

	theme := findTheme(cp.Box)
	selector := "colorpicker/contrast.warn"
	if ratio >= 4.5 {
		selector = "colorpicker/contrast.ok"
	}
	if theme != nil {
		theme.Apply(cp.contrastLabel, selector)
	}
	Redraw(cp.contrastLabel)
}

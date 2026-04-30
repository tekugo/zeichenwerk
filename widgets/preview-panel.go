package widgets

import (
	"fmt"

	. "github.com/tekugo/zeichenwerk/core"
)

// PreviewPanel renders a 3-row foreground-on-background swatch with the
// word "Sample" centred, followed by the WCAG 2.1 contrast ratio between
// the two colours. It is a pure display component — no events are emitted
// — and is driven by the parent ColorPicker via SetForeground,
// SetBackground, or SetColors.
//
// Default outer dimensions are 15×9 (inner 13×7), matching the height of
// a ColorPanel so the two can sit side by side without padding.
type PreviewPanel struct {
	*Box

	fg RGB
	bg RGB

	swatch        *Static
	contrastLabel *Static
}

// NewPreviewPanel creates a preview panel. The default colours are black
// foreground on white background, giving a contrast ratio of 21.0.
func NewPreviewPanel(id, class string) *PreviewPanel {
	pp := &PreviewPanel{
		Box: NewBox(id, class, "Preview"),
		fg:  RGB{0, 0, 0},
		bg:  RGB{255, 255, 255},
	}
	// Content size 13×7; the box border adds 2 cols and 2 rows for outer 15×9.
	pp.Box.SetHint(13, 7)
	pp.build()
	pp.refresh()
	return pp
}

// build constructs the inner widget tree.
func (pp *PreviewPanel) build() {
	id := pp.Box.ID()

	body := NewFlex(id+"-body", "", Stretch, 0)
	body.SetFlag(FlagVertical, true)
	pp.Box.Add(body)

	pp.swatch = NewStatic(id+"-swatch", "", "Sample")
	pp.swatch.SetAlignment("center")
	// content height 1; the swatch style adds padding 1,0 to fill 3 rows
	pp.swatch.SetHint(-1, 1)
	body.Add(pp.swatch)

	rule := NewHRule("", "thin")
	rule.SetHint(-1, 1)
	body.Add(rule)

	// Empty row above the contrast label
	body.Add(staticCell("", -1))

	pp.contrastLabel = NewStatic(id+"-contrast", "", "")
	pp.contrastLabel.SetHint(-1, 1)
	body.Add(pp.contrastLabel)

	// Empty row below
	body.Add(staticCell("", -1))
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies the previewpanel theme styles to the panel and recursively
// applies theme styles to every inner widget.
func (pp *PreviewPanel) Apply(theme *Theme) {
	theme.Apply(pp.Box, pp.Box.Selector("previewpanel"))
	theme.Apply(pp.Box, pp.Box.Selector("previewpanel/title"))
	Traverse(pp.Box, func(w Widget) bool {
		w.Apply(theme)
		return true
	})
	pp.styleSwatch()
	pp.styleContrast()
}

// ---- Public API -----------------------------------------------------------

// SetColors updates both colours and refreshes the swatch and the contrast
// label.
func (pp *PreviewPanel) SetColors(fg, bg RGB) {
	pp.fg = fg
	pp.bg = bg
	pp.refresh()
}

// SetForeground updates the foreground colour only.
func (pp *PreviewPanel) SetForeground(fg RGB) {
	pp.fg = fg
	pp.refresh()
}

// SetBackground updates the background colour only.
func (pp *PreviewPanel) SetBackground(bg RGB) {
	pp.bg = bg
	pp.refresh()
}

// Foreground returns the current foreground colour.
func (pp *PreviewPanel) Foreground() RGB { return pp.fg }

// Background returns the current background colour.
func (pp *PreviewPanel) Background() RGB { return pp.bg }

// Contrast returns the WCAG 2.1 contrast ratio between fg and bg.
func (pp *PreviewPanel) Contrast() float64 {
	return ContrastRatio(pp.fg.R, pp.fg.G, pp.fg.B, pp.bg.R, pp.bg.G, pp.bg.B)
}

// ---- Internal -------------------------------------------------------------

// refresh redraws the swatch and the contrast label.
func (pp *PreviewPanel) refresh() {
	pp.styleSwatch()
	pp.styleContrast()
}

// styleSwatch repaints the preview swatch with the current fg/bg colours.
func (pp *PreviewPanel) styleSwatch() {
	if pp.swatch == nil {
		return
	}
	pp.swatch.SetStyle("", NewStyle("previewpanel/swatch").
		WithColors(pp.fg.Hex(), pp.bg.Hex()).
		WithFont("bold").
		WithPadding(1, 0))
	Redraw(pp.swatch)
}

// styleContrast updates the contrast label text and the ok/warn style.
func (pp *PreviewPanel) styleContrast() {
	if pp.contrastLabel == nil {
		return
	}
	ratio := pp.Contrast()
	pp.contrastLabel.Set(fmt.Sprintf("Contrast %4.1f", ratio))

	theme := findTheme(pp.Box)
	selector := "colorpicker/contrast.warn"
	if ratio >= 4.5 {
		selector = "colorpicker/contrast.ok"
	}
	if theme != nil {
		theme.Apply(pp.contrastLabel, selector)
	}
	Redraw(pp.contrastLabel)
}

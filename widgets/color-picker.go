package widgets

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// ColorPickerMode selects what the picker can edit.
type ColorPickerMode int

const (
	// ColorSingle edits a single colour and shows only one ColorPanel.
	ColorSingle ColorPickerMode = iota
	// ColorFgBg edits a foreground and background pair and shows two
	// ColorPanels followed by a PreviewPanel that displays the contrast
	// ratio between them.
	ColorFgBg
)

// ColorPicker is the outer composite widget. It arranges either one or
// two ColorPanels (and an optional PreviewPanel) horizontally with a
// one-column gap. Editing any panel re-emits a single EvtChange on the
// picker with the picker as payload — consumers call Foreground,
// Background, and Contrast to read the new values.
//
// The picker has no buttons; confirm and cancel belong to the host
// container (a Dialog, side panel, …).
type ColorPicker struct {
	*Flex

	mode    ColorPickerMode
	fg      *ColorPanel
	bg      *ColorPanel   // nil in ColorSingle
	preview *PreviewPanel // nil in ColorSingle
}

// NewColorPicker creates a colour picker in the given mode.
func NewColorPicker(id, class string, mode ColorPickerMode) *ColorPicker {
	flex := NewFlex(id, class, Stretch, 1)
	cp := &ColorPicker{
		Flex: flex,
		mode: mode,
	}
	cp.build()
	cp.applyMode()
	return cp
}

// build wires the inner panels for the current mode.
func (cp *ColorPicker) build() {
	id := cp.Flex.ID()

	cp.fg = NewColorPanel(id+"-fg", "", "Foreground")
	cp.Flex.Add(cp.fg)
	cp.fg.On(EvtChange, cp.onPanelChange)

	if cp.mode == ColorFgBg {
		cp.bg = NewColorPanel(id+"-bg", "", "Background")
		// Sensible default contrast: black on white.
		cp.bg.SetRGB(RGB{255, 255, 255})
		cp.Flex.Add(cp.bg)
		cp.bg.On(EvtChange, cp.onPanelChange)

		cp.preview = NewPreviewPanel(id+"-preview", "")
		cp.Flex.Add(cp.preview)
	}
}

// applyMode sets the picker's hint based on its mode. Outer dimensions:
//   - ColorSingle: 24×9 (one ColorPanel)
//   - ColorFgBg:   65×9 (two ColorPanels + PreviewPanel + 2 gaps)
func (cp *ColorPicker) applyMode() {
	if cp.mode == ColorFgBg {
		cp.Flex.SetHint(65, 9)
	} else {
		cp.Flex.SetHint(24, 9)
	}
}

// onPanelChange propagates fg/bg changes to the preview swatch and
// re-emits a single EvtChange on the picker.
func (cp *ColorPicker) onPanelChange(_ Widget, _ Event, _ ...any) bool {
	if cp.preview != nil && cp.bg != nil {
		cp.preview.SetColors(cp.fg.RGB(), cp.bg.RGB())
	}
	cp.Flex.Dispatch(cp, EvtChange, cp)
	return false
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies the colorpicker theme styles to the picker and recursively
// applies theme styles to every inner widget.
func (cp *ColorPicker) Apply(theme *Theme) {
	theme.Apply(cp.Flex, cp.Flex.Selector("colorpicker"))
	Traverse(cp.Flex, func(w Widget) bool {
		w.Apply(theme)
		return true
	})
}

// ---- Public API -----------------------------------------------------------

// SetForeground updates the foreground colour from a "#RGB" or "#RRGGBB"
// string. No-op on parse error.
func (cp *ColorPicker) SetForeground(hex string) {
	cp.fg.SetColor(hex)
}

// SetBackground updates the background colour from a hex string. No-op in
// ColorSingle mode or on parse error.
func (cp *ColorPicker) SetBackground(hex string) {
	if cp.bg != nil {
		cp.bg.SetColor(hex)
	}
}

// Foreground returns the current foreground colour as "#RRGGBB".
func (cp *ColorPicker) Foreground() string {
	return cp.fg.Color()
}

// Background returns the current background colour as "#RRGGBB", or the
// empty string in ColorSingle mode.
func (cp *ColorPicker) Background() string {
	if cp.bg == nil {
		return ""
	}
	return cp.bg.Color()
}

// Mode returns the picker's current mode.
func (cp *ColorPicker) Mode() ColorPickerMode {
	return cp.mode
}

// Contrast returns the WCAG 2.1 contrast ratio between fg and bg, or 1.0
// in ColorSingle mode.
func (cp *ColorPicker) Contrast() float64 {
	if cp.preview == nil {
		return 1.0
	}
	return cp.preview.Contrast()
}

// ForegroundPanel returns the foreground ColorPanel, useful for tests and
// for embedding the panel into custom layouts.
func (cp *ColorPicker) ForegroundPanel() *ColorPanel { return cp.fg }

// BackgroundPanel returns the background ColorPanel, or nil in
// ColorSingle mode.
func (cp *ColorPicker) BackgroundPanel() *ColorPanel { return cp.bg }

// Preview returns the PreviewPanel, or nil in ColorSingle mode.
func (cp *ColorPicker) Preview() *PreviewPanel { return cp.preview }

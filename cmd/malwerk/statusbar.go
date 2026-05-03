package main

import (
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// StatusBar is malwerk's bottom-row status indicator. Unlike a plain
// Static, it can be collapsed to zero rows so the editor reclaims the
// row when the user toggles the bar off.
type StatusBar struct {
	widgets.Component
	text    string
	visible bool
}

// NewStatusBar constructs a status bar with the given id.
func NewStatusBar(id string) *StatusBar {
	sb := &StatusBar{visible: true}
	sb.Component = *widgets.NewComponent(id, "")
	return sb
}

// Apply installs the status bar's style. We always install our local
// default ($fg0 on $bg2) — using theme.Apply would resolve the
// "statusbar" selector via the cascade up to the registered "" root
// style, which has the same fg/bg as the editor and would render the
// bar invisible. Themes that want to override can register a "statusbar"
// entry; we apply it after the local default so it wins.
//
// Setting the style here (rather than in NewStatusBar) is important —
// the Builder calls Apply automatically when the widget is added, and
// any earlier SetStyle would be overwritten.
func (s *StatusBar) Apply(theme *core.Theme) {
	s.SetStyle("", core.NewStyle("").WithColors("$fg0", "$bg2"))

	// If a theme explicitly registers "statusbar", let it win. We use
	// theme.styles directly is not exposed, so we detect a real entry
	// by comparing the cascade's resolved style against the root "":
	// if they differ, "statusbar" must have been registered along the
	// cascade path.
	sel := s.Selector("statusbar")
	if theme.Get(sel) != theme.Get("") {
		theme.Apply(s, sel)
	}
}

// Hint reports a single row when visible, zero rows when hidden — the
// flex parent allocates space accordingly so the editor expands into
// the freed row.
func (s *StatusBar) Hint() (int, int) {
	if !s.visible {
		return 0, 0
	}
	return 0, 1
}

// Render paints the background across the whole bar (so the row reads
// as a unit even when the text is short) and writes the status text on
// top, left-aligned, clipped to the bar's width.
func (s *StatusBar) Render(r *core.Renderer) {
	if !s.visible {
		return
	}
	x, y, w, h := s.Content()
	if w <= 0 || h <= 0 {
		return
	}
	style := s.Style()
	r.Set(style.Foreground(), style.Background(), style.Font())
	r.Fill(x, y, w, h, " ")
	r.Text(x, y, s.text, w)
}

// Set updates the text and queues a redraw.
func (s *StatusBar) Set(text string) {
	s.text = text
	s.Refresh()
}

// SetVisible shows or hides the bar. Hiding triggers a relayout so the
// editor expands into the freed row.
func (s *StatusBar) SetVisible(v bool) {
	if s.visible == v {
		return
	}
	s.visible = v
	widgets.Relayout(s)
}

// Visible reports the current visibility.
func (s *StatusBar) Visible() bool {
	return s.visible
}

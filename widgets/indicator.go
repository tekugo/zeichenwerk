package widgets

import (
	"unicode/utf8"

	. "github.com/tekugo/zeichenwerk/core"
)

// Indicator is a compact display widget that pairs a coloured status glyph
// with a static label — for example "● Online" or "● Failed". The glyph
// colour is driven by a core.Level value via the :debug, :info, :warning,
// :error, and :fatal style variants. The label always renders in the base
// "indicator" style and is unaffected by the level.
//
// Indicator is display-only: it is not focusable and does not respond to
// keyboard or mouse input.
type Indicator struct {
	Component
	level Level
	label string
	dot   string // glyph character; resolved from theme string "indicator.dot"
}

// NewIndicator creates a new indicator with the given level and label. An
// empty level is treated as core.Info at render time.
func NewIndicator(id, class string, level Level, label string) *Indicator {
	return &Indicator{
		Component: Component{id: id, class: class},
		level:     level,
		label:     label,
		dot:       "●",
	}
}

// ---- Widget Methods -------------------------------------------------------

// Apply registers the base "indicator" style and one variant per severity
// level, then resolves the "indicator.dot" theme string for the glyph.
func (i *Indicator) Apply(theme *Theme) {
	theme.Apply(i, i.Selector("indicator"), "debug", "info", "success", "warning", "error", "fatal")
	if s := theme.String("indicator.dot"); s != "" {
		i.dot = s
	}
}

// State returns the current level as its string form so the renderer's
// state-selector machinery picks up "indicator:<level>". An empty level
// falls back to "info" so a styled state always exists.
func (i *Indicator) State() string {
	if i.level == "" {
		return string(Info)
	}
	return string(i.level)
}

// Hint returns the natural width (glyph + space + label runes) and a
// height of one row.
func (i *Indicator) Hint() (int, int) {
	if i.hwidth != 0 || i.hheight != 0 {
		return i.hwidth, i.hheight
	}
	return 2 + utf8.RuneCountInString(i.label), 1
}

// Render draws the glyph using the level-specific foreground and the label
// using the base "indicator" style — the level colour tints only the glyph,
// never the label. The base style is also used for background and border
// (Component.Render does not see the level state because Go's embedded
// methods bind to *Component, not the outer type).
func (i *Indicator) Render(r *Renderer) {
	i.Component.Render(r)

	x, y, w, _ := i.Content()
	if w <= 0 {
		return
	}

	base := i.Style("")
	glyph := i.Style(":" + i.State())

	r.Set(glyph.Foreground(), base.Background(), base.Font())
	r.Text(x, y, i.dot, 1)

	if w < 2 {
		return
	}

	r.Set(base.Foreground(), base.Background(), base.Font())
	r.Text(x+2, y, i.label, w-2)
}

// ---- Getters and Setters --------------------------------------------------

// Level returns the current severity level.
func (i *Indicator) Level() Level {
	return i.level
}

// SetLevel updates the level and queues a redraw. EvtChange is not
// dispatched — level changes are not user-driven.
func (i *Indicator) SetLevel(l Level) {
	i.level = l
	i.Refresh()
}

// Label returns the current label.
func (i *Indicator) Label() string {
	return i.label
}

// SetLabel updates the label and triggers a relayout because the natural
// width depends on the label.
func (i *Indicator) SetLabel(s string) {
	i.label = s
	Relayout(i)
}

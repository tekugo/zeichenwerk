package widgets

import (
	"unicode/utf8"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ==== AI ===================================================================

// shortcutPair holds a single key/label hint pair.
type shortcutPair struct {
	key   string
	label string
}

// Shortcuts renders a row of keyboard hint pairs, each consisting of a
// highlighted key and a dimmed label. Visual decoration (prefix, pair
// separator, suffix) is supplied by the theme via named string tokens.
//
// Style keys used:
//
//	"shortcuts"        base style — background, optional border, prefix/suffix colour
//	"shortcuts/key"    style for the key portion of each pair
//	"shortcuts/label"  style for the label portion of each pair
//
// Theme string tokens:
//
//	"shortcuts.prefix"     rendered before the first pair (default: "")
//	"shortcuts.separator"  rendered between consecutive pairs (default: "   ")
//	"shortcuts.suffix"     rendered after the last pair (default: "")
type Shortcuts struct {
	Component
	pairs     []shortcutPair
	prefix    string
	separator string
	suffix    string
}

// NewShortcuts creates a Shortcuts widget from alternating key/label strings,
// following the same convention as NewSelect:
//
//	NewShortcuts("id", "", "r", "run", "w", "watch", "q", "quit")
func NewShortcuts(id, class string, pairs ...string) *Shortcuts {
	s := &Shortcuts{
		Component: Component{id: id, class: class},
		pairs:     make([]shortcutPair, 0, len(pairs)/2),
	}
	for i := 0; i+1 < len(pairs); i += 2 {
		s.pairs = append(s.pairs, shortcutPair{key: pairs[i], label: pairs[i+1]})
	}
	return s
}

// SetPairs replaces the key/label pairs and triggers a redraw. Pairs are
// supplied as alternating key, label strings (same convention as NewShortcuts).
func (s *Shortcuts) SetPairs(pairs ...string) {
	s.pairs = make([]shortcutPair, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		s.pairs = append(s.pairs, shortcutPair{key: pairs[i], label: pairs[i+1]})
	}
	Redraw(s)
}

// ---- Widget Methods ---------------------------------------------------------

// Apply applies theme styles and caches the decoration strings.
func (s *Shortcuts) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("shortcuts"))
	theme.Apply(s, s.Selector("shortcuts/key"))
	theme.Apply(s, s.Selector("shortcuts/label"))
	s.prefix = theme.String("shortcuts.prefix")
	s.separator = theme.String("shortcuts.separator")
	s.suffix = theme.String("shortcuts.suffix")
}

// Hint returns the natural width (sum of all rendered runes) and height 1.
func (s *Shortcuts) Hint() (int, int) {
	if s.hwidth != 0 || s.hheight != 0 {
		return s.hwidth, s.hheight
	}
	w := utf8.RuneCountInString(s.prefix) + utf8.RuneCountInString(s.suffix)
	for i, p := range s.pairs {
		w += utf8.RuneCountInString(p.key) + 1 + utf8.RuneCountInString(p.label)
		if i < len(s.pairs)-1 {
			w += utf8.RuneCountInString(s.separator)
		}
	}
	return w, 1
}

// Render draws the shortcuts bar.
func (s *Shortcuts) Render(r *Renderer) {
	s.Component.Render(r)
	x0, y, w, _ := s.Content()
	if w <= 0 {
		return
	}

	base := s.Style("")
	key := s.Style("key")
	label := s.Style("label")

	bg := base.Background()
	x := x0

	// Render prefix.
	if s.prefix != "" {
		r.Set(base.Foreground(), bg, base.Font())
		r.Text(x, y, s.prefix, w-(x-x0))
		x += utf8.RuneCountInString(s.prefix)
	}

	for i, p := range s.pairs {
		// Key.
		r.Set(key.Foreground(), bg, key.Font())
		r.Text(x, y, p.key, w-(x-x0))
		x += utf8.RuneCountInString(p.key)

		// Space + label.
		r.Set(label.Foreground(), bg, label.Font())
		r.Text(x, y, " "+p.label, w-(x-x0))
		x += 1 + utf8.RuneCountInString(p.label)

		// Separator between pairs.
		if i < len(s.pairs)-1 && s.separator != "" {
			r.Set(base.Foreground(), bg, base.Font())
			r.Text(x, y, s.separator, w-(x-x0))
			x += utf8.RuneCountInString(s.separator)
		}
	}

	// Render suffix.
	if s.suffix != "" {
		r.Set(base.Foreground(), bg, base.Font())
		r.Text(x, y, s.suffix, w-(x-x0))
	}
}

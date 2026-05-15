package widgets

import (
	"fmt"
	"io"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// ShortcutsForm is the WidgetForm for *Shortcuts. Pairs are edited as
// a comma-separated list of "key:label" entries; the same convention
// is used in SelectForm, with the colon serving as the value/label
// separator.
type ShortcutsForm struct {
	ComponentForm

	PairsRaw string `group:"value" label:"Pairs (key:label, comma-separated)"`
}

func (f *ShortcutsForm) Name() string  { return "Shortcuts" }
func (f *ShortcutsForm) Group() string { return "leaf" }
func (f *ShortcutsForm) Help() string  { return "Row of keyboard hint pairs" }

func (f *ShortcutsForm) Load(w core.Widget) {
	s := w.(*Shortcuts)
	f.ComponentForm.Load(&s.Component)
	parts := make([]string, len(s.pairs))
	for i, p := range s.pairs {
		parts[i] = p.key + ":" + p.label
	}
	f.PairsRaw = strings.Join(parts, ", ")
}

func (f *ShortcutsForm) Store(w core.Widget) {
	s := w.(*Shortcuts)
	f.ComponentForm.Store(&s.Component)
	s.pairs = parseShortcutPairs(f.PairsRaw)
}

func (f *ShortcutsForm) New() core.Widget {
	args := flattenShortcutPairs(parseShortcutPairs(f.PairsRaw))
	s := NewShortcuts("", "", args...)
	f.Store(s)
	return s
}

func (f *ShortcutsForm) Validate(field string) error { return nil }

func (f *ShortcutsForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		pairs := parseShortcutPairs(f.PairsRaw)
		args := fmt.Sprintf("%q", f.ID)
		for _, p := range pairs {
			args += fmt.Sprintf(", %q, %q", p.key, p.label)
		}
		_, err := fmt.Fprintf(w, "Shortcuts(%s).\n", args)
		return err
	})
}

// parseShortcutPairs splits "key:label, key:label" into shortcutPair
// records. Entries without ':' are treated as a key with an empty
// label.
func parseShortcutPairs(raw string) []shortcutPair {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]shortcutPair, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		if i := strings.IndexByte(t, ':'); i >= 0 {
			out = append(out, shortcutPair{key: strings.TrimSpace(t[:i]), label: strings.TrimSpace(t[i+1:])})
		} else {
			out = append(out, shortcutPair{key: t})
		}
	}
	return out
}

// flattenShortcutPairs returns the alternating key/label slice
// NewShortcuts expects.
func flattenShortcutPairs(pairs []shortcutPair) []string {
	out := make([]string, 0, len(pairs)*2)
	for _, p := range pairs {
		out = append(out, p.key, p.label)
	}
	return out
}

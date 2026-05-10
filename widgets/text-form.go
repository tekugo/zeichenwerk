package widgets

import (
	"fmt"
	"io"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// TextForm is the WidgetForm for *Text. Content is a slice of lines
// edited as a single newline-joined string; the form converts to and
// from []string on Load/Store. Scroll offsets are runtime-only.
type TextForm struct {
	ComponentForm

	Content string `group:"value" label:"Content"`
	Follow  bool   `group:"general" label:"Auto-follow"`
	Max     int    `group:"general" label:"Max Lines"`
}

func (f *TextForm) Name() string  { return "Text" }
func (f *TextForm) Group() string { return "leaf" }
func (f *TextForm) Help() string  { return "Multi-line text display with scrolling" }

func (f *TextForm) Load(w core.Widget) {
	t := w.(*Text)
	f.ComponentForm.Load(&t.Component)
	f.Content = strings.Join(t.content, "\n")
	f.Follow = t.follow
	f.Max = t.max
}

func (f *TextForm) Store(w core.Widget) {
	t := w.(*Text)
	f.ComponentForm.Store(&t.Component)
	t.Set(splitLines(f.Content))
	t.follow = f.Follow
	t.max = f.Max
}

func (f *TextForm) New() core.Widget {
	t := NewText("", "", splitLines(f.Content), f.Follow, f.Max)
	f.Store(t)
	return t
}

func (f *TextForm) Validate(field string) error { return nil }

// Emit writes the Text constructor. Initial content is rarely a good
// fit for a generated source literal (very long block, lots of
// quoting), so the form emits a placeholder slice and flags it for
// the user to fill in.
func (f *TextForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		lines := splitLines(f.Content)
		var contentExpr string
		if len(lines) == 0 {
			contentExpr = "nil"
		} else {
			quoted := make([]string, len(lines))
			for i, l := range lines {
				quoted[i] = fmt.Sprintf("%q", l)
			}
			contentExpr = "[]string{" + strings.Join(quoted, ", ") + "}"
		}
		_, err := fmt.Fprintf(w, "Text(%q, %s, %t, %d).\n", f.ID, contentExpr, f.Follow, f.Max)
		return err
	})
}

// splitLines splits an editor-friendly newline-separated string into
// lines, preserving empty trailing lines is not desirable for the
// Text widget so the helper drops a single trailing empty entry that
// strings.Split otherwise produces for inputs ending in '\n'.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	out := strings.Split(s, "\n")
	if len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}

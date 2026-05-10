package widgets

import (
	"fmt"
	"io"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// BreadcrumbForm is the WidgetForm for *Breadcrumb. Segments are
// edited as a comma-separated string in the same convention used by
// ListForm; separator and overflow marker are exposed as plain
// strings so themes that override them are still inspectable here.
type BreadcrumbForm struct {
	ComponentForm

	SegmentsRaw string `group:"value" label:"Segments (comma-separated)"`
	Separator   string `group:"display" label:"Separator"`
	Overflow    string `group:"display" label:"Overflow"`
}

func (f *BreadcrumbForm) Name() string  { return "Breadcrumb" }
func (f *BreadcrumbForm) Group() string { return "leaf" }
func (f *BreadcrumbForm) Help() string  { return "Path indicator with focusable segments" }

func (f *BreadcrumbForm) Load(w core.Widget) {
	bc := w.(*Breadcrumb)
	f.ComponentForm.Load(&bc.Component)
	f.SegmentsRaw = strings.Join(bc.segments, ", ")
	f.Separator = bc.separator
	f.Overflow = bc.overflow
}

func (f *BreadcrumbForm) Store(w core.Widget) {
	bc := w.(*Breadcrumb)
	f.ComponentForm.Store(&bc.Component)
	bc.segments = parseItems(f.SegmentsRaw)
	if f.Separator != "" {
		bc.separator = f.Separator
	}
	if f.Overflow != "" {
		bc.overflow = f.Overflow
	}
	if bc.selected >= len(bc.segments) {
		bc.selected = len(bc.segments) - 1
	}
}

func (f *BreadcrumbForm) New() core.Widget {
	bc := NewBreadcrumb("", "")
	f.Store(bc)
	return bc
}

func (f *BreadcrumbForm) Validate(field string) error { return nil }

// Emit writes the Breadcrumb constructor onto an in-progress chain.
// The Builder has no chained setter for segments, separator, or
// overflow; those land as TODO comments after the standard frame.
func (f *BreadcrumbForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Breadcrumb(%q).\n", f.ID)
		return err
	}); err != nil {
		return err
	}
	if items := parseItems(f.SegmentsRaw); len(items) > 0 {
		quoted := make([]string, len(items))
		for i, it := range items {
			quoted[i] = fmt.Sprintf("%q", it)
		}
		fmt.Fprintf(w, "// TODO: Set([]string{%s}) — no Builder setter\n", strings.Join(quoted, ", "))
	}
	if f.Separator != "" {
		fmt.Fprintf(w, "// TODO: SetSeparator(%q) — no Builder setter\n", f.Separator)
	}
	if f.Overflow != "" {
		fmt.Fprintf(w, "// TODO: SetOverflow(%q) — no Builder setter\n", f.Overflow)
	}
	return nil
}

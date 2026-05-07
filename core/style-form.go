package core

import (
	"fmt"
	"io"
)

// StyleForm is the editor + codegen surface for a *Style. It mirrors the
// pattern of a per-widget *Form struct: a plain Go struct with
// reflection-friendly tags that the inspector's Form widget can render
// directly. Lives in the core package so Load and Store can read and
// write the unexported fields on Style without a public accessor surface.
//
// One StyleForm represents one selector. The Selector field is read-only
// because changing it would require re-keying the owning theme or widget
// style map, which is the caller's responsibility, not the form's.
type StyleForm struct {
	Selector string `group:"selector" label:"Selector" readonly:""`

	Foreground string `group:"colors" label:"Fg" control:"color" width:"16"`
	Background string `group:"colors" label:"Bg" control:"color" width:"16"`

	Border string `group:"box" label:"Border" control:"border"`
	// Padding and Margin are stored as comma- (or whitespace-)
	// separated decimal strings using CSS-shorthand: "" maps to
	// NoInsets, "1" applies to all four sides, "1 2" splits into
	// top/bottom and left/right, "1 2 3" into top, left/right,
	// and bottom, "1 2 3 4" into top right bottom left. Load
	// renders the canonical short form on read; Store re-parses
	// on Apply.
	Padding string `group:"box" label:"Padding" control:"insets" width:"16"`
	Margin  string `group:"box" label:"Margin"  control:"insets" width:"16"`

	Font   string `group:"text" label:"Font" control:"font"`
	Cursor string `group:"text" label:"Cursor"`

	Shadow string `group:"effects" label:"Shadow"`

	// fixed snapshots whether the loaded style was themed
	// (immutable). Themed styles are inherited from the active
	// theme; emitting them as widget-specific overrides in
	// generated source would clobber theme changes, so
	// EmitBuilderChain short-circuits when fixed is true. Not a
	// form field — no struct tag, lowercase, invisible to the
	// property panel.
	fixed bool
}

// Load copies state from s into the form. nil s is treated as an
// all-empty style. The fixed snapshot is captured so subsequent
// EmitBuilderChain calls can suppress emission for themed styles.
func (f *StyleForm) Load(s *Style) {
	if s == nil {
		*f = StyleForm{}
		return
	}
	f.Selector = s.selector
	f.Foreground = s.foreground
	f.Background = s.background
	f.Border = s.border
	if s.padding == nil {
		f.Padding = ""
	} else {
		f.Padding = s.padding.String()
	}
	if s.margin == nil {
		f.Margin = ""
	} else {
		f.Margin = s.margin.String()
	}
	f.Font = s.font
	f.Cursor = s.cursor
	f.Shadow = s.shadow
	f.fixed = s.fixed
}

// Store writes the form fields back into a modifiable copy of s and
// returns the resulting style. If s was unfixed the receiver is mutated
// in place and the returned pointer equals s; if s was fixed a fresh
// child style is created (via Modifiable) and returned. Callers that
// hold the original pointer must compare the result and reinstall when
// it differs.
//
// Padding and Margin parse the CSV form on the way back; if a value
// fails to parse, the existing inset is left unchanged so a half-
// typed entry mid-edit doesn't corrupt the live style.
func (f *StyleForm) Store(s *Style) *Style {
	if s == nil {
		s = NewStyle(f.Selector)
	}
	s = s.Modifiable()
	if s.padding == nil {
		s.padding = NewInsets()
	}
	if s.margin == nil {
		s.margin = NewInsets()
	}
	s.foreground = f.Foreground
	s.background = f.Background
	s.border = f.Border
	s.padding.Parse(f.Padding)
	s.margin.Parse(f.Margin)
	s.font = f.Font
	s.cursor = f.Cursor
	s.shadow = f.Shadow
	return s
}

// Validate runs per-field validation. Currently a stub; per-field
// validators land here once the inspector's editor surface is wired up.
func (f *StyleForm) Validate(field string) error { return nil }

// EmitBuilderChain writes the chained Builder method calls that set
// the non-zero style fields. Each call ends with ".\n" so subsequent
// chain elements concatenate directly. Final indentation is left to
// gofmt.
//
// When the loaded style was themed (fixed = true), no chain is
// emitted: the widget inherits the style from the active theme and
// emitting widget-specific overrides would override theme changes
// at the call site. Only widget-specific (non-fixed) styles produce
// output.
//
// Cursor and Shadow are not emitted because the Builder API has no
// chained setters for them today; they would need to land on the
// style directly via SetStyle or theme registration. They surface
// as TODO comments instead.
func (f *StyleForm) EmitBuilderChain(w io.Writer) {
	if f.fixed {
		return
	}
	if pad, ok := parseInsets(f.Padding); ok && !pad.IsZero() {
		fmt.Fprintf(w, "Padding(%s).\n", pad.String(", "))
	}
	if mar, ok := parseInsets(f.Margin); ok && !mar.IsZero() {
		fmt.Fprintf(w, "Margin(%s).\n", mar.String(", "))
	}
	if f.Border != "" {
		fmt.Fprintf(w, "Border(%q).\n", f.Border)
	}
	if f.Foreground != "" {
		fmt.Fprintf(w, "Foreground(%q).\n", f.Foreground)
	}
	if f.Background != "" {
		fmt.Fprintf(w, "Background(%q).\n", f.Background)
	}
	if f.Font != "" {
		fmt.Fprintf(w, "Font(%q).\n", f.Font)
	}
	if f.Cursor != "" {
		fmt.Fprintf(w, "// TODO: cursor=%q (no Builder setter)\n", f.Cursor)
	}
	if f.Shadow != "" {
		fmt.Fprintf(w, "// TODO: shadow=%q (no Builder setter)\n", f.Shadow)
	}
}

// ---- helpers ----

func parseInsets(s string) (*Insets, bool) {
	i := NewInsets()
	if !i.Parse(s) {
		return i, false
	} else {
		return i, true
	}
}

package core

import (
	"fmt"
	"io"
	"strconv"
	"strings"
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
	f.Padding = formatInsets(s.padding)
	f.Margin = formatInsets(s.margin)
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
	s.foreground = f.Foreground
	s.background = f.Background
	s.border = f.Border
	if pad, ok := parseInsetsCSV(f.Padding); ok {
		s.padding = insetsFromFormArray(pad)
	}
	if mar, ok := parseInsetsCSV(f.Margin); ok {
		s.margin = insetsFromFormArray(mar)
	}
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
	if pad, ok := parseInsetsCSV(f.Padding); ok && !zeroFormInsets(pad) {
		fmt.Fprintf(w, "Padding(%s).\n", formatFormInsets(pad))
	}
	if mar, ok := parseInsetsCSV(f.Margin); ok && !zeroFormInsets(mar) {
		fmt.Fprintf(w, "Margin(%s).\n", formatFormInsets(mar))
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

// formatInsets renders an *Insets as the canonical CSV-shorthand
// string used by StyleForm.Padding / Margin. nil and all-zero
// Insets both return "" (the form's representation of NoInsets);
// non-zero values use CSS-shorthand collapsing via formatFormInsets.
func formatInsets(in *Insets) string {
	a := insetsToFormArray(in)
	if zeroFormInsets(a) {
		return ""
	}
	return formatFormInsets(a)
}

// parseInsetsCSV parses a CSS-shorthand inset string into a [4]int
// keyed Top, Right, Bottom, Left. Whitespace and commas are both
// accepted as separators, so "1 2", "1, 2", "1,2", and "  1 ,  2"
// all yield the same result.
//
// Element-count semantics:
//
//	0 ("")           -> [0, 0, 0, 0]               (NoInsets)
//	1 ("a")          -> [a, a, a, a]
//	2 ("a b")        -> [a, b, a, b]               (top/bottom, left/right)
//	3 ("a b c")      -> [a, b, c, b]               (top, left/right, bottom)
//	4 ("a b c d")    -> [a, b, c, d]               (top, right, bottom, left)
//
// Returns ok=false on parse failure or unsupported element counts;
// callers (Store, Emit) keep the existing inset rather than
// applying a half-typed mid-edit value.
func parseInsetsCSV(s string) ([4]int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return [4]int{}, true
	}
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	nums := make([]int, len(fields))
	for i, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil {
			return [4]int{}, false
		}
		nums[i] = n
	}
	switch len(nums) {
	case 1:
		v := nums[0]
		return [4]int{v, v, v, v}, true
	case 2:
		t, h := nums[0], nums[1]
		return [4]int{t, h, t, h}, true
	case 3:
		t, h, b := nums[0], nums[1], nums[2]
		return [4]int{t, h, b, h}, true
	case 4:
		return [4]int{nums[0], nums[1], nums[2], nums[3]}, true
	}
	return [4]int{}, false
}

func zeroFormInsets(a [4]int) bool { return a == [4]int{} }

func insetsToFormArray(in *Insets) [4]int {
	if in == nil {
		return [4]int{}
	}
	return [4]int{in.Top, in.Right, in.Bottom, in.Left}
}

func insetsFromFormArray(a [4]int) *Insets {
	if zeroFormInsets(a) {
		return nil
	}
	return &Insets{Top: a[0], Right: a[1], Bottom: a[2], Left: a[3]}
}

// formatFormInsets reverses the CSS-shorthand expansion so the shortest
// equivalent variadic argument list is emitted (matches WithPadding /
// WithMargin's input expectations).
func formatFormInsets(a [4]int) string {
	t, r, b, l := a[0], a[1], a[2], a[3]
	switch {
	case t == r && r == b && b == l:
		return fmt.Sprintf("%d", t)
	case t == b && r == l:
		return fmt.Sprintf("%d, %d", t, r)
	case r == l:
		return fmt.Sprintf("%d, %d, %d", t, r, b)
	default:
		return fmt.Sprintf("%d, %d, %d, %d", t, r, b, l)
	}
}

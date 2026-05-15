package inspector

import (
	"reflect"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// buildPropertiesPane renders w's form fields as read-only
// "label: value" Static lines. Two paths:
//
//   - typed form: designer.FormFor returns a *XxxForm already
//     loaded from w. The walker recurses through any embedded
//     ComponentForm so inherited fields appear above declared
//     ones.
//   - fallback: no kind registered. We extract w's embedded
//     widgets.Component via reflection, instantiate a fresh
//     ComponentForm{}, load it, and walk that. Widgets that
//     don't embed Component (none today) get a single-line
//     placeholder.
//
// Each emitted line takes the form "  Label: value". The walker
// skips unexported fields (so the StyleForm pointer on
// ComponentForm doesn't leak) and nil pointers.
func (s *session) buildPropertiesPane(w core.Widget) core.Widget {
	stack := widgets.NewFlex("props-stack", "", core.Stretch, 0)
	stack.SetFlag(core.FlagVertical, true)

	hdr := widgets.NewStatic("props-header", "section", " Properties ")
	hdr.Apply(s.theme)
	_ = stack.Add(hdr)

	form := s.d.FormFor(w)
	if form != nil {
		s.walkFormFields(stack, reflect.ValueOf(form).Elem())
		return stack
	}

	if cmp, ok := extractComponent(w); ok {
		var cf widgets.ComponentForm
		cf.Load(cmp)
		s.walkFormFields(stack, reflect.ValueOf(&cf).Elem())
		return stack
	}

	note := widgets.NewStatic("props-fallback", "muted",
		"  (no form available; widget does not embed Component)")
	note.Apply(s.theme)
	_ = stack.Add(note)
	return stack
}

// walkFormFields appends one Static line per visible field of v
// to stack. Anonymous embedded structs recurse before declared
// fields so inherited values appear in source-order.
//
// Unexported fields are skipped (PkgPath != ""), nil pointers
// emit a placeholder, and struct/slice/map values fall through
// to formatValue's "%v" path — none of those appear in current
// widget forms.
func (s *session) walkFormFields(stack *widgets.Flex, v reflect.Value) {
	t := v.Type()
	for i := range v.NumField() {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue // unexported
		}
		fv := v.Field(i)
		if sf.Anonymous && fv.Kind() == reflect.Struct {
			s.walkFormFields(stack, fv)
			continue
		}
		s.addPropertyLine(stack, fieldLabel(sf), fv)
	}
}

// addPropertyLine renders one field as a Static line. The label
// is left-padded to 10 chars so columns align across siblings
// without a tab character (which would render unevenly in
// monospace terminals at smaller widths).
func (s *session) addPropertyLine(stack *widgets.Flex, label string, v reflect.Value) {
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return
	}
	line := widgets.NewStatic("prop-"+label, "", "  "+padRight(label, 10)+" "+formatValue(v))
	line.Apply(s.theme)
	_ = stack.Add(line)
}

// padRight pads s with spaces on the right to width n, returning
// s unchanged when it already exceeds n. Used by addPropertyLine
// to align the label column.
func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	buf := make([]byte, n)
	copy(buf, s)
	for i := len(s); i < n; i++ {
		buf[i] = ' '
	}
	return string(buf)
}

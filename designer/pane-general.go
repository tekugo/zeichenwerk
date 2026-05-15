package designer

import (
	"reflect"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// buildGeneralPane renders form's editable fields under a header
// per embedded struct level — ComponentForm / *Form / etc. — so a
// reader can see which inherited surface a field comes from.
// Direct fields of form's outermost type appear last under a
// header derived from form's own type name.
//
// Returns a fresh Widget; callers add it to the General pane Box
// and Relayout. No session state is read or written here.
func (s *session) buildGeneralPane(_ core.Widget, form WidgetForm) core.Widget {
	stack := widgets.NewFlex("general-stack", "", core.Stretch, 0)
	stack.SetFlag(core.FlagVertical, true)

	f := widgets.NewForm("form-general", "", "", form)
	f.Apply(s.theme)

	content := widgets.NewFlex("general-content", "", core.Stretch, 0)
	content.SetFlag(core.FlagVertical, true)

	v := reflect.ValueOf(form).Elem()
	t := v.Type()

	isFirst := true
	addSection := func(title string, fv reflect.Value) {
		if !isFirst {
			addSeparator(content, s.theme)
		}
		isFirst = false
		addFormSection(f, content, s.theme, title, fv)
	}
	for i := range v.NumField() {
		sf := t.Field(i)
		fv := v.Field(i)
		if !sf.Anonymous || fv.Kind() != reflect.Struct {
			continue
		}
		addSection(sectionTitle(sf.Type.Name()), fv)
	}
	addSection(sectionTitle(t.Name()), v)
	_ = f.Add(content)
	_ = stack.Add(f)
	return stack
}

package designer

import (
	"reflect"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// buildStylePane renders the form's StyleForm under a single
// "Style" section, or a muted placeholder when the form doesn't
// expose one. Style edits flow back through Store on the parent
// WidgetForm — the StyleForm pointer is owned by the form, not
// by us, so this builder doesn't return any side-effect state.
func (s *session) buildStylePane(_ core.Widget, form WidgetForm) core.Widget {
	stack := widgets.NewFlex("style-stack", "", core.Stretch, 0)
	stack.SetFlag(core.FlagVertical, true)

	styleForm := form.Style()
	if styleForm == nil {
		note := widgets.NewStatic("style-note", "muted", "  no style form available")
		note.Apply(s.theme)
		_ = stack.Add(note)
		return stack
	}

	f := widgets.NewForm("form-style", "", "", styleForm)
	f.Apply(s.theme)
	content := widgets.NewFlex("style-content", "", core.Stretch, 0)
	content.SetFlag(core.FlagVertical, true)
	addFormSection(f, content, s.theme, "Style", reflect.ValueOf(styleForm).Elem())
	_ = f.Add(content)
	_ = stack.Add(f)
	return stack
}

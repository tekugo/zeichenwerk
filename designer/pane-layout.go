package designer

import (
	"fmt"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// buildLayoutPane renders the per-child layout form (when the
// parent is a ContainerForm) above a read-only Computed block
// showing the widget's current bounds + hint.
//
// Returns (widget, layoutForm, parent). The layoutForm + parent
// are nil when the parent has no per-child params; rebuildPane
// stores them on the session so apply can flush them back. The
// extra return values keep the side effect explicit — the
// original PoC wrote them deep inside the build, which surprised
// readers.
func (s *session) buildLayoutPane(w core.Widget, _ WidgetForm) (core.Widget, core.LayoutForm, core.Container) {
	stack := widgets.NewFlex("layout-stack", "", core.Stretch, 0)
	stack.SetFlag(core.FlagVertical, true)

	var (
		layoutForm core.LayoutForm
		parent     core.Container
	)

	if p := w.Parent(); p != nil {
		if pf := s.d.FormFor(p); pf != nil {
			if cf, ok := pf.(ContainerForm); ok {
				if lf := cf.LayoutForm(p, w); lf != nil {
					layoutForm = lf
					parent = p

					hdr := widgets.NewStatic("layout-header", "section",
						fmt.Sprintf(" Layout in %s%s ", widgetKind(p), idSuffix(p)))
					hdr.Apply(s.theme)
					_ = stack.Add(hdr)

					form := widgets.NewForm("layout-form", "", "", lf)
					group := widgets.NewFormGroup("layout-fg", "", "", true, 0)
					form.Apply(s.theme)
					group.Apply(s.theme)
					widgets.BuildFormGroup(form, group, "", s.theme)
					_ = form.Add(group)
					_ = stack.Add(form)

					addSeparator(stack, s.theme)
				}
			}
		}
	}

	if layoutForm == nil {
		note := widgets.NewStatic("layout-note", "muted",
			"  parent has no per-child layout parameters")
		note.Apply(s.theme)
		_ = stack.Add(note)
		addSeparator(stack, s.theme)
	}

	comp := widgets.NewStatic("layout-computed-header", "section", " Computed ")
	comp.Apply(s.theme)
	_ = stack.Add(comp)

	x, y, ww, wh := w.Bounds()
	hw, hh := w.Hint()
	bounds := widgets.NewStatic("layout-bounds", "",
		fmt.Sprintf("  bounds   x=%d  y=%d  w=%d  h=%d", x, y, ww, wh))
	bounds.Apply(s.theme)
	_ = stack.Add(bounds)
	hint := widgets.NewStatic("layout-hint", "",
		fmt.Sprintf("  hint     w=%d  h=%d", hw, hh))
	hint.Apply(s.theme)
	_ = stack.Add(hint)

	return stack, layoutForm, parent
}

package designer

import (
	"fmt"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// buildInfoPane renders a read-only summary of w: type, id, parent,
// child count, persistent flags. Useful for confirming what's
// selected when the General tab is dominated by editable controls.
func (s *session) buildInfoPane(w core.Widget, _ WidgetForm) core.Widget {
	stack := widgets.NewFlex("info-stack", "", core.Stretch, 0)
	stack.SetFlag(core.FlagVertical, true)

	hdr := widgets.NewStatic("info-header", "section", " Widget ")
	hdr.Apply(s.theme)
	_ = stack.Add(hdr)

	row := func(id, label, value string) {
		line := widgets.NewStatic(id, "", "  "+label+" "+value)
		line.Apply(s.theme)
		_ = stack.Add(line)
	}

	row("info-type", "type    ", widgetKind(w))

	idStr := w.ID()
	if idStr == "" {
		idStr = "—"
	}
	row("info-id", "id      ", idStr)

	parentDesc := "—"
	if p := w.Parent(); p != nil {
		parentDesc = fmt.Sprintf("%s%s", widgetKind(p), idSuffix(p))
	}
	row("info-parent", "parent  ", parentDesc)

	childCount := "—"
	if c, ok := w.(core.Container); ok {
		childCount = fmt.Sprintf("%d", len(c.Children()))
	}
	row("info-children", "children", childCount)

	row("info-flags", "flags   ", flagSummary(w))

	return stack
}

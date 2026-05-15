package inspector

import (
	"fmt"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// classer is implemented by anything embedding widgets.Component —
// effectively every shipped widget. We assert against it from the
// inspector to surface the widget's class without forcing Class()
// onto the Widget interface itself.
type classer interface {
	Class() string
}

// buildLayoutPane renders w's runtime layout state as read-only
// "label: value" Static lines: type, id, bounds, content, hint,
// state, flags, class, parent, children. No side effects.
func (s *session) buildLayoutPane(w core.Widget) core.Widget {
	stack := widgets.NewFlex("layout-stack", "", core.Stretch, 0)
	stack.SetFlag(core.FlagVertical, true)

	hdr := widgets.NewStatic("layout-header", "section", " Layout ")
	hdr.Apply(s.theme)
	_ = stack.Add(hdr)

	row := func(id, label, value string) {
		line := widgets.NewStatic(id, "", "  "+label+" "+value)
		line.Apply(s.theme)
		_ = stack.Add(line)
	}

	row("layout-type", "type    ", widgetKind(w))

	idStr := w.ID()
	if idStr == "" {
		idStr = "—"
	}
	row("layout-id", "id      ", idStr)

	x, y, ww, wh := w.Bounds()
	row("layout-bounds", "bounds  ", fmt.Sprintf("x=%d y=%d w=%d h=%d", x, y, ww, wh))

	cx, cy, cw, ch := w.Content()
	row("layout-content", "content ", fmt.Sprintf("x=%d y=%d w=%d h=%d", cx, cy, cw, ch))

	hw, hh := w.Hint()
	row("layout-hint", "hint    ", fmt.Sprintf("w=%d h=%d", hw, hh))

	state := w.State()
	if state == "" {
		state = "—"
	}
	row("layout-state", "state   ", state)

	row("layout-flags", "flags   ", flagSummary(w))

	class := "—"
	if cl, ok := w.(classer); ok {
		if c := cl.Class(); c != "" {
			class = c
		}
	}
	row("layout-class", "class   ", class)

	parent := "—"
	if p := w.Parent(); p != nil {
		parent = widgetKind(p) + idSuffix(p)
	}
	row("layout-parent", "parent  ", parent)

	children := "—"
	if c, ok := w.(core.Container); ok {
		children = fmt.Sprintf("%d", len(c.Children()))
	}
	row("layout-children", "children", children)

	return stack
}

// flagSummary collapses the persistent runtime flags into a short
// space-separated label. Focusable/focused/hovered are dynamic;
// skip/hidden/disabled are persistent. We show all six because
// the inspector is a debugging surface — flicker isn't a concern
// when the user is paused on a selection.
func flagSummary(w core.Widget) string {
	parts := make([]string, 0, 6)
	if w.Flag(core.FlagFocusable) {
		parts = append(parts, "focusable")
	}
	if w.Flag(core.FlagFocused) {
		parts = append(parts, "focused")
	}
	if w.Flag(core.FlagHovered) {
		parts = append(parts, "hovered")
	}
	if w.Flag(core.FlagSkip) {
		parts = append(parts, "skip")
	}
	if w.Flag(core.FlagHidden) {
		parts = append(parts, "hidden")
	}
	if w.Flag(core.FlagDisabled) {
		parts = append(parts, "disabled")
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, " ")
}

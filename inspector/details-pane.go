package inspector

import (
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// rebuildDetails fills the right pane with the Properties +
// Layout sections for w. Builders are stateless renderers of w
// — they don't read s.current — so the contract is simple and
// the ordering of writes here is the only place state moves.
func (s *session) rebuildDetails(w core.Widget) {
	stack := widgets.NewFlex("details-stack", "", core.Stretch, 0)
	stack.SetFlag(core.FlagVertical, true)

	_ = stack.Add(s.buildPropertiesPane(w))
	addSeparator(stack, s.theme)
	_ = stack.Add(s.buildLayoutPane(w))

	_ = s.paneDetails.Add(stack)
	widgets.Relayout(s.paneDetails)
}

// clearDetails replaces the right pane with a muted placeholder.
// Called at startup and any time the selection collapses.
func (s *session) clearDetails() {
	hint := widgets.NewStatic("details-empty", "muted", "  (no widget selected)")
	hint.Apply(s.theme)
	_ = s.paneDetails.Add(hint)
	widgets.Relayout(s.paneDetails)
}

// addSeparator drops a thin horizontal rule into stack so the
// Properties and Layout sections read as distinct blocks.
func addSeparator(stack core.Container, theme *core.Theme) {
	rule := widgets.NewHRule("", "thin")
	rule.Apply(theme)
	_ = stack.Add(rule)
}

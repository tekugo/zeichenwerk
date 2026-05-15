package designer

import (
	"fmt"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// rebuildPane refills the four detail tabs for w. The only place
// that writes currentForm / currentLayout / currentParent — apply
// depends on those being in sync with the panes the user is
// looking at.
//
// Pane builders are stateless renderers of (widget, form). Layout
// is the exception: it also returns the discovered LayoutForm and
// parent so the side effect is explicit at the call site instead
// of hidden deep inside the build.
func (s *session) rebuildPane(w core.Widget) {
	form := s.d.FormFor(w)
	if form == nil {
		s.fillPanesWithPlaceholder("(no form registered for this widget)")
		s.currentWidget = w
		s.currentForm = nil
		s.currentLayout = nil
		s.currentParent = nil
		return
	}

	gen := s.buildGeneralPane(w, form)
	lay, lf, lp := s.buildLayoutPane(w, form)
	sty := s.buildStylePane(w, form)
	inf := s.buildInfoPane(w, form)

	s.currentWidget = w
	s.currentForm = form
	s.currentLayout = lf
	s.currentParent = lp

	_ = s.paneGeneral.Add(gen)
	_ = s.paneLayout.Add(lay)
	_ = s.paneStyle.Add(sty)
	_ = s.paneInfo.Add(inf)

	widgets.Relayout(s.paneGeneral)
	widgets.Relayout(s.paneLayout)
	widgets.Relayout(s.paneStyle)
	widgets.Relayout(s.paneInfo)
}

// clearTabs fills every pane with an empty-state placeholder. Used
// at startup, after Delete when no survivor exists, and any other
// time the selection collapses to nothing.
func (s *session) clearTabs() {
	s.fillPanesWithPlaceholder("(no widget selected)")
}

// fillPanesWithPlaceholder is the shared body of clearTabs and
// rebuildPane's no-form branch. msg is the muted text shown
// centred in each tab.
func (s *session) fillPanesWithPlaceholder(msg string) {
	stat := func(suffix string) core.Widget {
		w := widgets.NewStatic("placeholder-"+suffix, "muted", msg)
		w.Apply(s.theme)
		return w
	}
	_ = s.paneGeneral.Add(stat("g"))
	_ = s.paneLayout.Add(stat("l"))
	_ = s.paneStyle.Add(stat("s"))
	_ = s.paneInfo.Add(stat("i"))
	widgets.Relayout(s.paneGeneral)
	widgets.Relayout(s.paneLayout)
	widgets.Relayout(s.paneStyle)
	widgets.Relayout(s.paneInfo)
}

// apply flushes the current form back to the live widget, then
// (if a per-child layout form was loaded) flushes that too. The
// tree node label is re-derived so widget renames are reflected
// without rebuilding the tree, and the details pane is rebuilt so
// derived values like Computed bounds reflect the new state.
//
// Surfaces drift defensively: if a widget is selected but no form
// is current, that's a bug in the rebuild path; we say so on the
// status line instead of silently no-oping.
func (s *session) apply() {
	if s.currentWidget == nil {
		s.setStatus("Apply: no widget selected")
		return
	}
	if s.currentForm == nil {
		s.setStatus("Apply: form drift (widget set but no form loaded)")
		return
	}
	s.currentForm.Store(s.currentWidget)
	if s.currentLayout != nil && s.currentParent != nil {
		s.currentLayout.Store(s.currentParent, s.currentWidget)
	}
	widgets.Relayout(s.currentWidget)
	if s.currentNode != nil {
		s.currentNode.SetText(treeLabel(s.currentWidget))
		widgets.Redraw(s.tree)
	}
	w := s.currentWidget
	s.rebuildPane(w)
	s.setDirty(true)
	s.setStatus(fmt.Sprintf("applied → %s%s", widgetKind(w), idSuffix(w)))
}

// reset rebuilds the detail panes from the current widget's state,
// discarding any unflushed edits the user made through the forms.
// No mutation of the live widget; the side effect is only visual.
func (s *session) reset() {
	if s.currentWidget == nil {
		return
	}
	s.rebuildPane(s.currentWidget)
	s.setStatus("reset from widget state")
}

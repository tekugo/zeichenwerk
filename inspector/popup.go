// Package inspector hosts the runtime widget inspector — a
// read-only counterpart to the heavy design-time editor in
// designer/. Open(ui) attaches a popup toggled by Ctrl+D that
// shows the live widget tree, per-widget form fields read-only,
// runtime layout state, and the application log.
//
// The package is organised as:
//
//   - popup.go (this file) — public Open entrypoint, session
//     state, popup chrome layout, Ctrl+D toggle binding.
//   - tree-pane.go — left pane: widget tree + manual refresh
//     (F5). No add/delete/move toolbar; the inspector is
//     read-only by design.
//   - details-pane.go — right pane orchestrator: rebuildPane,
//     clearTabs.
//   - pane-properties.go — Properties section: reflective
//     form-field walker (typed form via designer.FormFor, or
//     a ComponentForm fallback).
//   - pane-layout.go — Layout section: bounds, content, hint,
//     state, flags, class, parent, children.
//   - pane-log.go — second top-level tab: Table mounted on
//     *UI.Logs() (the framework's circular log buffer).
//   - helpers.go — pure helpers (widgetKind, idSuffix,
//     treeLabel, buildWidgetTreeNode, extractComponent,
//     fieldLabel, formatValue).
package inspector

import (
	"github.com/gdamore/tcell/v3"

	zw "github.com/tekugo/zeichenwerk"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/designer"
	"github.com/tekugo/zeichenwerk/widgets"
)

// session owns all state for one inspector popup attached to one
// UI. Open closes over it; nothing outside the package needs the
// type, so it stays unexported.
type session struct {
	ui    *zw.UI
	theme *core.Theme
	root  core.Container     // captured at Open: the user's base layer
	d     *designer.Designer // typed-form lookup only; codegen unused

	popup       core.Container
	tree        *widgets.Tree
	tabs        *widgets.Tabs
	paneDetails *widgets.Viewport

	current core.Widget // last-selected widget; nil before first select
}

// Open attaches a read-only inspector popup to ui. Ctrl+D toggles
// the popup; ESC closes it. The popup shows ui's base layer
// (ui.Children()[0]) as a widget tree; selecting a widget renders
// its form fields and runtime layout state. A second top-level
// tab shows the application log.
//
// Open uses ui.Theme() for the popup chrome and builds an
// internal designer.Designer with the default kind table for
// typed-form lookup; the codegen path is never exercised.
//
// Call Open at most once per *UI — a second call would install a
// second Ctrl+D handler. Pairs naturally with designer.Open;
// both can be attached to the same ui (different keystrokes).
func Open(ui *zw.UI) {
	root := firstBaseChild(ui)
	if root == nil {
		// Nothing to inspect; bail rather than installing a popup
		// that opens onto an empty tree. The caller should attach
		// the inspector after building the UI.
		return
	}
	d := designer.NewDesigner(root)
	designer.RegisterDefaults(d)

	s := &session{
		ui:    ui,
		theme: ui.Theme(),
		root:  root,
		d:     d,
	}
	s.buildPopup()
	s.wireActions()
	s.installKeyBindings()
}

// firstBaseChild returns the first widget on ui's base layer
// (typically the user's root container), or nil if ui has no
// children yet. Captured once so subsequent popups don't pick up
// added layers — keeps the tree stable and excludes the inspector
// itself.
func firstBaseChild(ui *zw.UI) core.Container {
	children := ui.Children()
	if len(children) == 0 {
		return nil
	}
	c, ok := children[0].(core.Container)
	if !ok {
		return nil
	}
	return c
}

// buildPopup constructs the popup chrome:
//
//	[Inspector] [Log]                <- top-level Tabs
//	+-- Inspector ----------------+
//	| Tree pane   |  Details pane |
//	+-----------------------------+
//	| Log pane (Table)            |
//	+-----------------------------+
//
// Hint(96, 36) keeps the popup a focused modal rather than letting
// it grow to fill the terminal.
func (s *session) buildPopup() {
	b := zw.NewBuilder(s.theme).
		Box("inspector-box", "").Hint(96, 36).Border("round").
		Class("inspector").
		VFlex("inspector-root", core.Stretch, 0).
		Tabs("inspector-tabs", "Inspector", "Log").Hint(0, 2).
		Switcher("inspector-switcher", true).Hint(0, -1).
		// ===== Inspector tab =====
		HFlex("inspector-main", core.Stretch, 0).
		VFlex("inspector-tree-pane", core.Stretch, 0).Hint(34, -1).
		TreeWidgets("tree", s.root).Hint(0, -1).
		Static("inspector-help", "  F5 refresh   ESC close").Hint(0, 1).
		End(). // closes tree-pane
		Viewport("inspector-details", "").Flag(core.FlagVertical).Flag(core.FlagHorizontal).Border("none").Hint(62, 36).
		End(). // closes details viewport
		End(). // closes inspector-main
		// ===== Log tab =====
		Box("inspector-log-box", "").Border("none").
		End(). // closes log box
		End(). // closes switcher
		End(). // closes inspector-root
		End()  // closes inspector-box

	s.popup = b.Container()
	s.tree = core.MustFind[*widgets.Tree](s.popup, "tree")
	s.tabs = core.MustFind[*widgets.Tabs](s.popup, "inspector-tabs")
	s.paneDetails = core.MustFind[*widgets.Viewport](s.popup, "inspector-details")

	s.mountLogPane()
}

// wireActions binds the popup's interactive surface. There are
// no buttons — only select on the tree and F5 to refresh.
func (s *session) wireActions() {
	s.tree.On(widgets.EvtSelect, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.onTreeSelect()
		return false
	})

	// F5 on the popup refreshes the tree from the live root —
	// the inspector doesn't observe app-side structural mutations
	// automatically.
	s.popup.On(widgets.EvtKey, func(_ core.Widget, _ core.Event, data ...any) bool {
		if len(data) == 0 {
			return false
		}
		ev, ok := data[0].(*tcell.EventKey)
		if !ok {
			return false
		}
		if ev.Key() == tcell.KeyF5 {
			s.refreshTree()
			return true
		}
		return false
	})

	s.clearDetails()
}

// installKeyBindings registers Ctrl+D on the root UI to toggle
// the inspector popup. ESC closes via the framework's existing
// layer logic; we don't intercept it.
func (s *session) installKeyBindings() {
	s.ui.On(widgets.EvtKey, func(_ core.Widget, _ core.Event, data ...any) bool {
		if len(data) == 0 {
			return false
		}
		ev, ok := data[0].(*tcell.EventKey)
		if !ok {
			return false
		}
		if ev.Key() == tcell.KeyCtrlD {
			s.ui.Popup(-1, -1, 0, 0, s.popup)
			return true
		}
		return false
	})
}

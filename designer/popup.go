// Package designer hosts the design-time UI editor.
//
// Open(ui, target, theme, proj) attaches a designer popup to ui that
// edits target. Ctrl+Space toggles the popup; ESC closes it. The
// popup is built eagerly and reused across opens — the same
// selection state survives close/reopen cycles within one process.
//
// The package is organised as:
//
//   - designer.go, kind.go, widget-form.go — headless engine + the
//     WidgetForm interface widget forms satisfy structurally.
//   - popup.go (this file) — public Open entrypoint, session state,
//     popup chrome layout, Ctrl+Space key binding.
//   - tree-pane.go — left pane: widget tree + add/remove/move
//     toolbar actions.
//   - details-pane.go, pane-*.go — right pane: General / Layout /
//     Style / Info tabs; rebuildPane orchestrator.
//   - dialogs.go — add-child picker and project-settings dialogs;
//     save and header/dirty/status mutators.
//   - project.go, defaults.go — Project codegen-output settings and
//     the 37-entry default kind table.
//   - helpers.go — pure helpers (widgetKind, idSuffix, treeLabel, …)
//     that don't touch session state.
package designer

import (
	"github.com/gdamore/tcell/v3"

	zw "github.com/tekugo/zeichenwerk"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// Nerd-font glyphs for the tree-action toolbar. Pre-resolved
// codepoints (FontAwesome subset). Terminals without a Nerd Font
// render them as a missing-glyph indicator, still preferable to
// three-letter abbreviations next to four icons in a row.
const (
	iconAdd    = "" // nf-fa-plus
	iconDelete = "" // nf-fa-trash
	iconUp     = "" // nf-fa-arrow_up
	iconDown   = "" // nf-fa-arrow_down
)

// session owns all state for one designer popup attached to one UI.
// It is package-private; Open closes over it and never returns it.
// Tests that need to poke at internals construct a session through
// newSession.
type session struct {
	// Inputs (set once in newSession, never reassigned).
	ui     *zw.UI
	target core.Container
	theme  *core.Theme
	d      *Designer
	proj   *Project

	// Popup chrome (resolved by buildPopup, never reassigned).
	popup       core.Container
	tree        *widgets.Tree
	tabs        *widgets.Tabs
	dirtyDot    *widgets.Static
	fileLabel   *widgets.Static
	themeLabel  *widgets.Static
	status      *widgets.Static
	paneGeneral *widgets.Box
	paneLayout  *widgets.Box
	paneStyle   *widgets.Box
	paneInfo    *widgets.Box

	// Mutable selection state. Owned by the tree-select and the
	// details-pane rebuild path; only the action methods (apply,
	// delete, …) read these without writing.
	currentWidget core.Widget
	currentNode   *widgets.TreeNode
	currentForm   WidgetForm
	currentLayout core.LayoutForm
	currentParent core.Container
}

// Open attaches a designer popup to ui that edits target.
// Ctrl+Space toggles the popup; ESC closes it. The popup is built
// eagerly so the first Ctrl+Space is instantaneous.
//
// theme is applied to freshly-built form widgets inside the popup.
// proj seeds the codegen output settings the Settings dialog mutates
// in place; pass nil for DefaultProject().
//
// Open registers form factories for every widget kind shipping in
// widgets via RegisterDefaults. Drivers needing custom kinds should
// build their own *Designer + Open variant for now (a WithKinds
// option can be added later if a use case materialises).
//
// Call Open at most once per *UI — a second call would install a
// second Ctrl+Space handler.
func Open(ui *zw.UI, target core.Container, theme *core.Theme, proj *Project) {
	if proj == nil {
		proj = DefaultProject()
	}
	d := NewDesigner(target)
	RegisterDefaults(d)

	s := newSession(ui, target, theme, d, proj)
	s.buildPopup()
	s.wireActions()
	s.installKeyBindings()
}

// newSession builds an empty session with its inputs set. The popup
// chrome and mutable selection state stay zero until buildPopup runs.
func newSession(ui *zw.UI, target core.Container, theme *core.Theme, d *Designer, proj *Project) *session {
	return &session{
		ui:     ui,
		target: target,
		theme:  theme,
		d:      d,
		proj:   proj,
	}
}

// buildPopup constructs the 3×2 Grid popup and caches every chrome
// widget Open's other methods need to reach. Layout:
//
//	row 0 (h=1)        : header band (title + file/dirty/theme + actions)
//	row 1 (h=-1, fills): tree pane | details pane
//	row 2 (h=1)        : status line
//
//	col 0 (w=34): tree pane — TreeWidgets above an icon toolbar
//	col 1 (w=-1): details pane — Tabs / Switcher / Apply-Reset toolbar
//
// Hint(96, 36) keeps the popup a focused modal rather than letting
// it grow to fill the terminal; the rendered grid lines do the
// visual framing.
func (s *session) buildPopup() {
	b := zw.NewBuilder(s.theme).
		Grid("designer-grid", 3, 2, true).Hint(96, 36).
		Columns(34, -1).Rows(1, -1, 1).
		// ===== Row 0: header band, spanning both columns =====
		Cell(0, 0, 2, 1).
		HFlex("header-band", core.Center, 1).
		Static("designer-title", " Designer ").Class("title").
		Static("header-file", s.proj.Name).
		Static("header-dirty", "●").Hint(1, 1).
		Static("header-spacer-1", "  ").
		Static("header-theme", s.proj.Theme).
		Spacer().Hint(-1, 0).
		Button("save-btn", "Save").
		Button("generate-btn", "Generate").
		Button("run-btn", "Run").
		Button("settings-btn", "Settings").
		End(). // closes header-band
		// ===== Row 1, col 0: tree pane =====
		Cell(0, 1, 1, 1).
		VFlex("tree-pane", core.Stretch, 0).
		TreeWidgets("tree", s.target).Hint(0, -1).
		HFlex("tree-toolbar", core.Center, 2).Hint(0, 1).
		Button("add-btn", iconAdd).
		Button("del-btn", iconDelete).
		Button("up-btn", iconUp).
		Button("down-btn", iconDown).
		End(). // closes tree-toolbar
		End(). // closes tree-pane
		// ===== Row 1, col 1: details pane =====
		Cell(1, 1, 1, 1).
		VFlex("details-pane", core.Stretch, 0).
		Tabs("details-tabs", "General", "Layout", "Style", "Info").Hint(0, 2).
		Switcher("details-switcher", true).Hint(0, -1).
		Box("tab-general", "").Border("none").
		End().
		Box("tab-layout", "").Border("none").
		End().
		Box("tab-style", "").Border("none").
		End().
		Box("tab-info", "").Border("none").
		End().
		End(). // closes details-switcher
		HFlex("details-toolbar", core.End, 2).Hint(0, 1).
		Spacer().Hint(-1, 0).
		Button("apply-btn", "Apply").
		Button("reset-btn", "Reset").
		End(). // closes details-toolbar
		End(). // closes details-pane
		// ===== Row 2: status line, spanning both columns =====
		Cell(0, 2, 2, 1).
		Static("status", " ").
		End() // closes designer-grid

	s.popup = b.Container()

	s.tree = core.MustFind[*widgets.Tree](s.popup, "tree")
	s.tabs = core.MustFind[*widgets.Tabs](s.popup, "details-tabs")
	s.dirtyDot = core.MustFind[*widgets.Static](s.popup, "header-dirty")
	s.fileLabel = core.MustFind[*widgets.Static](s.popup, "header-file")
	s.themeLabel = core.MustFind[*widgets.Static](s.popup, "header-theme")
	s.status = core.MustFind[*widgets.Static](s.popup, "status")
	s.paneGeneral = core.MustFind[*widgets.Box](s.popup, "tab-general")
	s.paneLayout = core.MustFind[*widgets.Box](s.popup, "tab-layout")
	s.paneStyle = core.MustFind[*widgets.Box](s.popup, "tab-style")
	s.paneInfo = core.MustFind[*widgets.Box](s.popup, "tab-info")
}

// wireActions binds every interactive control inside the popup. Each
// branch hands off to a method on session that lives in the file
// dealing with that subject (tree pane, details pane, dialogs). The
// table is centralised here so a reader sees every button → method
// edge at one glance.
func (s *session) wireActions() {
	core.MustFind[*widgets.Button](s.popup, "apply-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.apply()
		return false
	})
	core.MustFind[*widgets.Button](s.popup, "reset-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.reset()
		return false
	})
	core.MustFind[*widgets.Button](s.popup, "save-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.save()
		return false
	})
	core.MustFind[*widgets.Button](s.popup, "generate-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.save()
		return false
	})
	core.MustFind[*widgets.Button](s.popup, "run-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.run()
		return false
	})
	core.MustFind[*widgets.Button](s.popup, "settings-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.openSettingsDialog()
		return false
	})
	core.MustFind[*widgets.Button](s.popup, "add-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.add()
		return false
	})
	core.MustFind[*widgets.Button](s.popup, "del-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.delete()
		return false
	})
	core.MustFind[*widgets.Button](s.popup, "up-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		return s.moveSibling(-1)
	})
	core.MustFind[*widgets.Button](s.popup, "down-btn").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		return s.moveSibling(+1)
	})

	s.tree.On(widgets.EvtSelect, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.onTreeSelect()
		return false
	})

	// Alt+1..4 jumps to the matching detail tab. Bound to the popup
	// so it works anywhere inside; the handler returns true on a
	// match so the framework doesn't treat the digit as input to a
	// focused field.
	s.popup.On(widgets.EvtKey, func(_ core.Widget, _ core.Event, data ...any) bool {
		if len(data) == 0 {
			return false
		}
		ev, ok := data[0].(*tcell.EventKey)
		if !ok || ev.Key() != tcell.KeyRune {
			return false
		}
		if ev.Modifiers()&tcell.ModAlt == 0 {
			return false
		}
		switch ev.Str() {
		case "1":
			s.tabs.Set(0)
			return true
		case "2":
			s.tabs.Set(1)
			return true
		case "3":
			s.tabs.Set(2)
			return true
		case "4":
			s.tabs.Set(3)
			return true
		}
		return false
	})

	// Empty-state placeholders so the popup is well-formed before
	// any selection. refreshHeader picks up the initial project
	// values; setDirty(false) hides the dirty dot.
	s.clearTabs()
	s.refreshHeader()
	s.setDirty(false)
}

// installKeyBindings registers Ctrl+Space on the root UI to toggle
// the designer popup. The handler runs before UI's hard-coded
// global-key block so it can intercept Ctrl+Space without modifying
// the framework. ESC closes via the framework's existing layer
// logic; we don't intercept it.
func (s *session) installKeyBindings() {
	s.ui.On(widgets.EvtKey, func(_ core.Widget, _ core.Event, data ...any) bool {
		if len(data) == 0 {
			return false
		}
		ev, ok := data[0].(*tcell.EventKey)
		if !ok {
			return false
		}
		if ev.Key() == tcell.KeyNUL ||
			(ev.Key() == tcell.KeyRune && ev.Str() == " " && ev.Modifiers()&tcell.ModCtrl != 0) {
			s.ui.Popup(-1, -1, 0, 0, s.popup)
			return true
		}
		return false
	})
}

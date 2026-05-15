package designer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// setStatus updates the status line with a timestamped message and
// forces a redraw. Status messages are transient by nature; a real
// inspector would dispatch a toast. Time suffix gives the user a
// rough sense of "did this just happen, or was it from a minute
// ago?".
func (s *session) setStatus(msg string) {
	s.status.Set(msg + "    " + time.Now().Format("15:04:05"))
	widgets.Redraw(s.status)
}

// setDirty toggles the visibility of the dirty dot in the header
// band. The dot IS the dirty state — no separate flag needed.
func (s *session) setDirty(v bool) {
	s.dirtyDot.SetFlag(core.FlagHidden, !v)
	widgets.Redraw(s.dirtyDot)
}

// refreshHeader copies the current Project labels into the header
// chrome. Called after Settings closes and once at startup.
func (s *session) refreshHeader() {
	s.fileLabel.Set(s.proj.Name)
	s.themeLabel.Set(s.proj.Theme)
	widgets.Redraw(s.fileLabel)
	widgets.Redraw(s.themeLabel)
}

// save writes the current target subtree as a complete Go source
// file to proj.OutPath. Both the "Save" and "Generate" buttons
// land here — they did the same thing in the PoC, and splitting
// them would add a settings toggle without a clear use case.
func (s *session) save() {
	var buf bytes.Buffer
	if err := s.d.GenerateFile(ModeBuilder, &buf, s.proj.Package, s.proj.FuncName); err != nil {
		s.setStatus("generate failed: " + err.Error())
		return
	}
	if err := os.WriteFile(s.proj.OutPath, buf.Bytes(), 0o644); err != nil {
		s.setStatus("write failed: " + err.Error())
		return
	}
	s.setDirty(false)
	s.setStatus("wrote " + s.proj.OutPath)
}

// run is a placeholder until the designer can launch the generated
// program in a child terminal. The PoC stubs it; we preserve the
// stub so the button doesn't disappear from the chrome.
func (s *session) run() {
	s.setStatus("Run: not implemented yet")
}

// openSettingsDialog shows the project-settings modal. Edits are
// applied to s.proj in place on commit; refreshHeader is called so
// the chrome reflects the new values.
func (s *session) openSettingsDialog() {
	dialog := s.ui.NewBuilder().
		Dialog("settings-dialog", "Project Settings").
		Class("dialog").
		VFlex("settings-body", core.Stretch, 1).Padding(1, 2).
		Static("settings-prompt", "Edit codegen and chrome settings.").Hint(0, 1).
		HFlex("settings-name-row", core.Start, 1).Hint(0, 1).
		Static("settings-name-label", "Name      ").Hint(12, 1).
		Input("settings-name", s.proj.Name).Hint(40, 1).
		End().
		HFlex("settings-out-row", core.Start, 1).Hint(0, 1).
		Static("settings-out-label", "Output    ").Hint(12, 1).
		Input("settings-out", s.proj.OutPath).Hint(40, 1).
		End().
		HFlex("settings-pkg-row", core.Start, 1).Hint(0, 1).
		Static("settings-pkg-label", "Package   ").Hint(12, 1).
		Input("settings-pkg", s.proj.Package).Hint(40, 1).
		End().
		HFlex("settings-fn-row", core.Start, 1).Hint(0, 1).
		Static("settings-fn-label", "Func name ").Hint(12, 1).
		Input("settings-fn", s.proj.FuncName).Hint(40, 1).
		End().
		HFlex("settings-theme-row", core.Start, 1).Hint(0, 1).
		Static("settings-theme-label", "Theme     ").Hint(12, 1).
		Input("settings-theme", s.proj.Theme).Hint(40, 1).
		End().
		HFlex("settings-buttons", core.End, 2).Hint(0, 1).
		Spacer().Hint(-1, 0).
		Button("settings-ok", "Save").
		Button("settings-cancel", "Cancel").
		End().
		End().
		Class("").
		Container()

	commit := func() {
		nameIn := core.MustFind[*widgets.Input](dialog, "settings-name")
		outIn := core.MustFind[*widgets.Input](dialog, "settings-out")
		pkgIn := core.MustFind[*widgets.Input](dialog, "settings-pkg")
		fnIn := core.MustFind[*widgets.Input](dialog, "settings-fn")
		themeIn := core.MustFind[*widgets.Input](dialog, "settings-theme")

		newName := strings.TrimSpace(nameIn.Get())
		if newName == "" {
			newName = filepath.Base(outIn.Get())
		}
		s.proj.Name = newName
		s.proj.OutPath = strings.TrimSpace(outIn.Get())
		s.proj.Package = strings.TrimSpace(pkgIn.Get())
		s.proj.FuncName = strings.TrimSpace(fnIn.Get())
		s.proj.Theme = strings.TrimSpace(themeIn.Get())
		s.ui.Close()
		s.refreshHeader()
		s.setStatus("settings updated")
	}

	core.MustFind[*widgets.Button](dialog, "settings-ok").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		commit()
		return true
	})
	core.MustFind[*widgets.Button](dialog, "settings-cancel").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.ui.Close()
		return true
	})

	s.ui.Popup(-1, -1, 0, 0, dialog)
}

// openAddChildDialog shows the kind-picker modal. Selecting a kind
// (Enter on the list or the Add button) calls Designer.Add on
// parent and invokes onAdded with the new widget; the caller does
// the theme apply / relayout / tree refresh.
func (s *session) openAddChildDialog(parent core.Container, onAdded func(core.Widget)) {
	kinds := s.d.KindNames()
	if len(kinds) == 0 {
		s.setStatus("no kinds registered")
		return
	}

	dialog := s.ui.NewBuilder().
		Dialog("add-child-dialog", "Add Child").
		Class("dialog").
		VFlex("add-child-body", core.Stretch, 1).
		Static("add-child-prompt",
			fmt.Sprintf("Add child to %s%s", widgetKind(parent), idSuffix(parent))).
		List("add-child-list", kinds...).Hint(28, 8).
		HFlex("add-child-buttons", core.End, 2).
		Button("add-child-ok", "Add").
		Button("add-child-cancel", "Cancel").
		End().
		End().
		Class("").
		Container()

	list := core.MustFind[*widgets.List](dialog, "add-child-list")
	commit := func() {
		idx := list.Selected()
		if idx < 0 || idx >= len(kinds) {
			s.ui.Close()
			return
		}
		kindName := kinds[idx]
		child, err := s.d.Add(parent, kindName)
		if err != nil {
			s.setStatus("add failed: " + err.Error())
			s.ui.Close()
			return
		}
		s.ui.Close()
		if onAdded != nil {
			onAdded(child)
		}
	}

	core.MustFind[*widgets.Button](dialog, "add-child-ok").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		commit()
		return true
	})
	core.MustFind[*widgets.Button](dialog, "add-child-cancel").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		s.ui.Close()
		return true
	})
	list.On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		commit()
		return true
	})

	s.ui.Popup(-1, -1, 0, 0, dialog)
}

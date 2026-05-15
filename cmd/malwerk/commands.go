package main

import (
	"path/filepath"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// registerCommands installs the command palette items per the spec.
func (a *App) registerCommands() {
	cmds := a.ui.Commands()

	// File group
	cmds.Register("File", "New", "", a.cmdNew)
	cmds.Register("File", "Open", "", a.cmdOpen)
	cmds.Register("File", "Save", "Ctrl-S", a.cmdSave)
	cmds.Register("File", "Save As", "", a.cmdSaveAs)
	cmds.Register("File", "Export ANSI", "", a.cmdExportANSI)
	cmds.Register("File", "Quit", "Ctrl-Q", a.cmdQuit)

	// Document group
	cmds.Register("Document", "Resize", "", a.cmdResize)
	cmds.Register("Document", "Toggle Status Bar", "", a.ToggleStatus)
	cmds.Register("Document", "Clear", "", a.cmdClear)

	// Style group
	cmds.Register("Style", "Pick Style", "", func() { a.openStyleEditor(false) })
	cmds.Register("Style", "Edit Styles", "S", func() { a.openStyleEditor(false) })
	cmds.Register("Style", "Pick Border", "B", a.openBorderPicker)
}

// ---- File commands -------------------------------------------------------

func (a *App) cmdNew() {
	a.confirmIfDirty(func() {
		a.openSizeDialog("New Document", a.editor.doc.Width, a.editor.doc.Height, func(w, h int) {
			a.editor.SetDoc(NewDocument(w, h))
			a.editor.Refresh()
		})
	})
}

func (a *App) cmdOpen() {
	a.confirmIfDirty(func() {
		dlg := a.ui.FileChooser("Open", "Open", "file", "", false)
		dlg.On(widgets.EvtAccept, func(_ core.Widget, _ core.Event, data ...any) bool {
			path, _ := data[0].(string)
			doc, err := LoadDocument(path)
			if err != nil {
				a.ui.Confirm("Open failed", err.Error(), nil, nil)
				return true
			}
			a.editor.SetDoc(doc)
			a.editor.Refresh()
			return true
		})
	})
}

func (a *App) cmdSave() {
	if a.editor.doc.Path == "" {
		a.cmdSaveAs()
		return
	}
	if err := a.editor.doc.Save(); err != nil {
		a.ui.Confirm("Save failed", err.Error(), nil, nil)
		return
	}
	a.refreshStatus()
}

func (a *App) cmdSaveAs() {
	dlg := a.ui.FileChooser("Save As", "Save", "save", a.editor.doc.Path, false)
	dlg.On(widgets.EvtAccept, func(_ core.Widget, _ core.Event, data ...any) bool {
		path, _ := data[0].(string)
		if !strings.HasSuffix(path, ".malwerk.json") && filepath.Ext(path) == "" {
			path += ".malwerk.json"
		}
		if err := a.editor.doc.SaveAs(path); err != nil {
			a.ui.Confirm("Save failed", err.Error(), nil, nil)
			return true
		}
		a.refreshStatus()
		return true
	})
}

func (a *App) cmdExportANSI() {
	dlg := a.ui.FileChooser("Export ANSI", "Export", "save", ansiDefaultPath(a.editor.doc.Path), false)
	dlg.On(widgets.EvtAccept, func(_ core.Widget, _ core.Event, data ...any) bool {
		path, _ := data[0].(string)
		if filepath.Ext(path) == "" {
			path += ".ans"
		}
		if err := ExportANSI(a.editor.doc, a.theme, path); err != nil {
			a.ui.Confirm("Export failed", err.Error(), nil, nil)
		}
		return true
	})
}

func (a *App) cmdQuit() {
	a.confirmIfDirty(func() {
		a.ui.Quit()
	})
}

// ---- Document commands ---------------------------------------------------

func (a *App) cmdResize() {
	a.openSizeDialog("Resize", a.editor.doc.Width, a.editor.doc.Height, func(w, h int) {
		a.editor.doc.Resize(w, h)
		a.editor.Refresh()
		a.refreshStatus()
	})
}

func (a *App) cmdClear() {
	hist := a.history
	hist.Begin()
	for y, row := range a.editor.doc.Cells {
		for x, c := range row {
			if c == EmptyCell {
				continue
			}
			hist.Record(x, y, c, EmptyCell)
			a.editor.doc.Cells[y][x] = EmptyCell
		}
	}
	hist.Commit()
	a.editor.doc.Dirty = true
	a.editor.Refresh()
	a.refreshStatus()
}

// confirmIfDirty runs onYes immediately when the document is clean,
// otherwise asks the user via a Confirm dialog first.
func (a *App) confirmIfDirty(onYes func()) {
	if !a.editor.doc.Dirty {
		onYes()
		return
	}
	a.ui.Confirm("Discard changes?",
		"The current document has unsaved changes. Continue?",
		onYes, nil)
}

// ansiDefaultPath proposes a sibling .ans path next to the current
// document, or "untitled.ans" when no path is set.
func ansiDefaultPath(path string) string {
	if path == "" {
		return "untitled.ans"
	}
	base := strings.TrimSuffix(path, ".malwerk.json")
	if base == path {
		base = strings.TrimSuffix(path, filepath.Ext(path))
	}
	return base + ".ans"
}


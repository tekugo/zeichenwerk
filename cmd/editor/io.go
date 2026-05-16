package main

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/tekugo/zeichenwerk/core"
	. "github.com/tekugo/zeichenwerk/widgets"
)

func (a *App) openPath(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	a.editor.Load(string(data))
	a.path = path
	a.dirty = false
	a.paintMask = map[[2]int]int{}
	a.refreshChrome()
	return nil
}

func (a *App) save() {
	if a.path == "" {
		a.saveAs()
		return
	}
	if err := a.writeCurrent(); err != nil {
		a.notify(fmt.Sprintf("save failed: %v", err))
	}
}

func (a *App) writeCurrent() error {
	if err := os.WriteFile(a.path, []byte(a.editor.Text()), 0o644); err != nil {
		return err
	}
	a.dirty = false
	a.refreshChrome()
	a.notify(fmt.Sprintf("saved %s", filepath.Base(a.path)))
	return nil
}

func (a *App) saveAs() {
	initial := a.path
	if initial == "" {
		initial, _ = os.Getwd()
	}
	d := a.ui.FileChooser("Save As", "Save", "save", initial, false)
	d.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 0 {
			return true
		}
		path, _ := data[0].(string)
		if path == "" {
			return true
		}
		commit := func() {
			a.path = path
			if err := a.writeCurrent(); err != nil {
				a.notify(fmt.Sprintf("save failed: %v", err))
			}
		}
		if _, err := os.Stat(path); err == nil {
			a.ui.Confirm("Overwrite?",
				filepath.Base(path)+" already exists. Overwrite?",
				commit, nil,
			)
		} else {
			commit()
		}
		return true
	})
}

func (a *App) openWithChooser() {
	initial := a.path
	if initial == "" {
		initial, _ = os.Getwd()
	}
	d := a.ui.FileChooser("Open File", "Open", "file", initial, false)
	d.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 0 {
			return true
		}
		path, _ := data[0].(string)
		if path == "" {
			return true
		}
		load := func() {
			if err := a.openPath(path); err != nil {
				a.notify(fmt.Sprintf("open failed: %v", err))
			}
		}
		if a.dirty {
			a.ui.Confirm("Discard unsaved changes?",
				"Current buffer has unsaved changes. Open new file anyway?",
				load, nil,
			)
		} else {
			load()
		}
		return true
	})
}

func (a *App) notify(msg string) {
	if a.status == nil {
		return
	}
	a.status.Set(" " + msg)
}

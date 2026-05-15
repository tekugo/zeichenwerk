package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// openSizeDialog presents a width × height entry popup. Calls onAccept
// with the chosen dimensions when the user confirms.
func (a *App) openSizeDialog(title string, w, h int, onAccept func(width, height int)) {
	b := a.ui.NewBuilder()
	dialog := b.
		Dialog("size-dialog", title).Class("dialog").
		VFlex("size-body", core.Stretch, 1).
		Static("size-w-label", "Width:").
		Input("size-w", strconv.Itoa(w)).Hint(20, 1).
		Static("size-h-label", "Height:").
		Input("size-h", strconv.Itoa(h)).Hint(20, 1).
		HFlex("size-buttons", core.End, 2).
		Button("size-ok", "OK").
		Button("size-cancel", "Cancel").
		End().
		End().
		Class("").
		Container()

	wInput := core.Find(dialog, "size-w").(*widgets.Input)
	hInput := core.Find(dialog, "size-h").(*widgets.Input)

	core.Find(dialog, "size-ok").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		nw, errW := strconv.Atoi(wInput.Get())
		nh, errH := strconv.Atoi(hInput.Get())
		if errW != nil || errH != nil || nw < 1 || nh < 1 {
			a.ui.Confirm("Invalid size", "Width and height must be positive integers.", nil, nil)
			return true
		}
		a.ui.Close()
		onAccept(nw, nh)
		return true
	})
	core.Find(dialog, "size-cancel").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		a.ui.Close()
		return true
	})

	a.ui.Popup(-1, -1, 0, 0, dialog)
}

// openBorderPicker shows a list of border family names; the chosen name
// becomes the editor's current border family.
func (a *App) openBorderPicker() {
	families := []string{"thin", "heavy", "double", "round"}
	b := a.ui.NewBuilder()
	dialog := b.
		Dialog("border-dialog", "Pick Border").Class("dialog").
		VFlex("border-body", core.Stretch, 1).
		List("border-list", families...).Hint(20, len(families)).
		HFlex("border-buttons", core.End, 2).
		Button("border-cancel", "Cancel").
		End().
		End().
		Class("").
		Container()

	list := core.Find(dialog, "border-list").(*widgets.List)
	list.On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, data ...any) bool {
		idx, _ := data[0].(int)
		if idx >= 0 && idx < len(families) {
			a.editor.SetCurrentBorder(families[idx])
		}
		a.ui.Close()
		return true
	})
	core.Find(dialog, "border-cancel").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		a.ui.Close()
		return true
	})

	a.ui.Popup(-1, -1, 0, 0, dialog)
}

// openStyleEditor shows the style palette editor. When applyToSelection
// is true and the editor is in Visual mode, the chosen style is applied
// to the selection on Pick.
func (a *App) openStyleEditor(applyToSelection bool) {
	doc := a.editor.doc
	names := paletteNames(doc)

	theme := a.ui.Theme()

	// Right-pane controls: a FormGroup for Name / Border / Font, with a
	// ColorPicker (fg + bg + preview) sitting above it. Every manually
	// constructed widget carries class "dialog" so the theme's
	// dialog-scoped variants (matching $bg2 background) apply.
	form := widgets.NewFormGroup("style-form", "dialog", "", false, 1)
	form.Apply(theme)

	nameIn := widgets.NewInput("style-name", "dialog")
	nameIn.SetHint(20, 1)
	nameIn.Apply(theme)
	form.Add(nameIn, 0, "Name:")

	borderSel := widgets.NewSelect("style-border", "dialog",
		"", "(default)",
		"thin", "thin",
		"heavy", "heavy",
		"double", "double",
		"round", "round",
	)
	borderSel.SetHint(15, 1)
	borderSel.Apply(theme)
	form.Add(borderSel, 1, "Border:")

	fontRow := widgets.NewFlex("style-font", "dialog", core.Start, 2)
	fontRow.SetHint(36, 1)
	fontRow.Apply(theme)
	boldBox := widgets.NewCheckbox("style-font-bold", "dialog", "Bold", false)
	boldBox.Apply(theme)
	italicBox := widgets.NewCheckbox("style-font-italic", "dialog", "Italic", false)
	italicBox.Apply(theme)
	underlineBox := widgets.NewCheckbox("style-font-underline", "dialog", "Underline", false)
	underlineBox.Apply(theme)
	fontRow.Add(boldBox)
	fontRow.Add(italicBox)
	fontRow.Add(underlineBox)
	form.Add(fontRow, 2, "Font:")

	b := a.ui.NewBuilder()
	dialog := b.
		Dialog("style-dialog", "Styles").Class("dialog").
		HFlex("style-body", core.Stretch, 1).
		VFlex("style-list-pane", core.Stretch, 0).
		List("style-list", names...).Hint(20, 8).
		HFlex("style-list-buttons", core.Start, 1).
		Button("style-new", "New").
		Button("style-pick", "Pick").
		Button("style-close", "Close").
		End().
		End().
		VFlex("style-edit-pane", core.Stretch, 1).
		ColorPicker("style-color", widgets.ColorFgBg).
		Add(form).End().
		End().
		End().
		Class("").
		Container()

	list := core.Find(dialog, "style-list").(*widgets.List)
	picker := core.Find(dialog, "style-color").(*widgets.ColorPicker)

	parseFont := func(s string) (bold, italic, underline bool) {
		for _, attr := range strings.Fields(s) {
			switch attr {
			case "bold":
				bold = true
			case "italic":
				italic = true
			case "underline":
				underline = true
			}
		}
		return
	}
	formatFont := func(bold, italic, underline bool) string {
		var parts []string
		if bold {
			parts = append(parts, "bold")
		}
		if italic {
			parts = append(parts, "italic")
		}
		if underline {
			parts = append(parts, "underline")
		}
		return strings.Join(parts, " ")
	}

	loadStyle := func(name string) {
		ds := doc.Palette[name]
		if ds == nil {
			return
		}
		nameIn.Set(name)
		if ds.Fg != "" {
			picker.SetForeground(ds.Fg)
		}
		if ds.Bg != "" {
			picker.SetBackground(ds.Bg)
		}
		borderSel.Select(ds.Border)
		bold, italic, underline := parseFont(ds.Font)
		boldBox.Set(bold)
		italicBox.Set(italic)
		underlineBox.Set(underline)
	}
	if len(names) > 0 {
		loadStyle(names[0])
	}

	list.On(widgets.EvtSelect, func(_ core.Widget, _ core.Event, data ...any) bool {
		idx, _ := data[0].(int)
		if idx >= 0 && idx < len(names) {
			loadStyle(names[idx])
		}
		return true
	})

	saveCurrent := func() string {
		oldName := nameIn.Get()
		// the list's selected name takes priority; we always edit that entry.
		idx := list.Selected()
		if idx < 0 || idx >= len(names) {
			return ""
		}
		current := names[idx]
		ds, ok := doc.Palette[current]
		if !ok || ds == nil {
			ds = &DocStyle{}
			doc.Palette[current] = ds
		}
		ds.Fg = picker.Foreground()
		ds.Bg = picker.Background()
		ds.Font = formatFont(
			boldBox.Flag(core.FlagChecked),
			italicBox.Flag(core.FlagChecked),
			underlineBox.Flag(core.FlagChecked),
		)
		ds.Border = borderSel.Value()
		if oldName != current && oldName != "" {
			_ = doc.RenameStyle(current, oldName)
			current = oldName
		}
		doc.Dirty = true
		return current
	}

	core.Find(dialog, "style-new").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		a.ui.Prompt("New Style", "Style name:", func(name string) {
			name = strings.TrimSpace(name)
			if name == "" {
				return
			}
			if _, exists := doc.Palette[name]; exists {
				a.ui.Confirm("Already exists", fmt.Sprintf("A style named %q already exists.", name), nil, nil)
				return
			}
			doc.Palette[name] = &DocStyle{}
			doc.Dirty = true
			names = paletteNames(doc)
			list.Set(names)
			for i, n := range names {
				if n == name {
					list.Select(i)
					loadStyle(name)
					break
				}
			}
		}, nil)
		return true
	})
	core.Find(dialog, "style-pick").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		picked := saveCurrent()
		if picked != "" {
			a.editor.SetCurrentStyle(picked)
			if applyToSelection && a.editor.mode == ModeVisual {
				a.editor.styleSelection(picked)
			}
		}
		a.ui.Close()
		a.editor.Refresh()
		return true
	})
	core.Find(dialog, "style-close").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		a.ui.Close()
		return true
	})

	a.ui.Popup(-1, -1, 0, 0, dialog)
}

// openGlyphPicker shows a typeahead over the embedded glyph index. The
// chosen glyph is inserted at the cursor and the cursor advances right.
func (a *App) openGlyphPicker() {
	entries := GlyphIndex()

	b := a.ui.NewBuilder()
	dialog := b.
		Dialog("glyph-dialog", "Glyph").Class("dialog").
		VFlex("glyph-body", core.Stretch, 1).
		Input("glyph-input", "").Hint(40, 1).
		List("glyph-list").Hint(40, 12).
		End().
		Class("").
		Container()

	input := core.Find(dialog, "glyph-input").(*widgets.Input)
	list := core.Find(dialog, "glyph-list").(*widgets.List)

	current := entries
	refreshList := func(query string) {
		current = filterGlyphs(entries, query)
		labels := make([]string, len(current))
		for i, g := range current {
			labels[i] = fmt.Sprintf("%s  %s", g.Char, g.Name)
		}
		list.Set(labels)
	}
	refreshList("")

	input.On(widgets.EvtChange, func(_ core.Widget, _ core.Event, data ...any) bool {
		q, _ := data[0].(string)
		refreshList(q)
		return true
	})
	list.On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, data ...any) bool {
		idx, _ := data[0].(int)
		if idx >= 0 && idx < len(current) {
			a.editor.typeRune(current[idx].Char)
		}
		a.ui.Close()
		return true
	})

	a.ui.Popup(-1, -1, 0, 0, dialog)
}


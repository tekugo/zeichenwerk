// Command designer-poc is the minimum interactive driver for the
// inspector / form architecture.
//
// The visible main screen wraps the live target tree (a VFlex with a
// header HFlex and a Grid) in a widgets.Preview, plus a thin status
// bar at the bottom. Preview's Children() returns empty so the
// framework's focus / hit-testing / event walks skip the target
// subtree entirely; Layout and Render still drive it normally so it
// renders and updates as a non-interactive preview.
//
// Ctrl+Space opens the designer popup. The popup has three areas:
// a top header band (project filename + dirty dot + theme + Save /
// Generate / Run / Settings actions); a left tree pane stacking the
// widget tree above its structural-action toolbar; and a right
// detail pane that switches between General / Layout / Style / Info
// tabs. Apply mutates the live target so the preview reflects edits
// immediately. ESC closes the popup; Ctrl+Q quits.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/inspector"
	"github.com/tekugo/zeichenwerk/themes"
	. "github.com/tekugo/zeichenwerk/widgets"
)

// project holds the codegen output settings the Settings dialog
// edits. The defaults match the previous PoC behaviour so existing
// muscle memory (open the popup, hit Generate, find the file in /tmp)
// keeps working.
type project struct {
	Name     string // shown in the header band
	OutPath  string // file written by Save / Generate
	Package  string // emitted package declaration
	FuncName string // emitted func wrapper name
	Theme    string // theme label shown in the header
}

func defaultProject() *project {
	return &project{
		Name:     "untitled.go",
		OutPath:  "/tmp/designer-poc-out.go",
		Package:  "main",
		FuncName: "BuildUI",
		Theme:    "TokyoNight",
	}
}

func main() {
	theme := themes.TokyoNight()

	// Build the main UI: target tree on top, status bar at the bottom.
	// Inlining the target-tree construction in the builder chain means
	// the same widgets that the Designer edits are the ones being
	// rendered — Apply mutates the live tree and the preview updates.
	ui := NewBuilder(theme).
		VFlex("ui-root", Stretch, 0).
		Preview("preview").Hint(0, -1).
		VFlex("target-root", Stretch, 1).
		HFlex("header", Center, 2).
		// Padding on the title forces a non-fixed style at fixture
		// build-time, so codegen exercises the StyleForm emission
		// path on first Generate without the user having to edit
		// anything.
		Static("title", "Inspector PoC").Padding(0, 1).
		Input("search", "", "", "type to filter…").
		End(). // closes header
		Grid("g1", 2, 2, false).
		Cell(0, 0, 1, 1).Static("s1", "Hello").
		Cell(1, 0, 1, 1).Class("highlight").Static("s2", "World").
		End(). // closes Grid
		End(). // closes target-root
		End(). // closes preview
		Static("main-status", " Ctrl+Space → designer    Ctrl+Q → quit ").Hint(0, 1).
		End(). // closes ui-root
		Build()

	// Designer reads & mutates the same widgets the Preview is
	// rendering. Find can't reach inside a Preview (its Children()
	// is empty), so we grab the Preview by id and ask for its
	// target. Once we have the target, downstream code uses normal
	// Container traversal — Preview only hides the subtree from the
	// framework, not from the inspector tooling that calls
	// Target() explicitly.
	preview := MustFind[*Preview](ui, "preview")
	target := preview.Target().(*Flex)
	d := inspector.NewDesigner(target)
	registerKinds(d)

	proj := defaultProject()

	// Build the designer popup as a free-standing Container.
	popup := buildDesignerPopup(theme, target, proj)

	// State shared between popup widgets. The dirty dot's visibility
	// IS the dirty state — no separate flag needed.
	var (
		currentWidget Widget
		currentNode   *TreeNode
		currentForm   inspector.WidgetForm
		currentLayout LayoutForm
		currentParent Container
	)

	tree := MustFind[*Tree](popup, "tree")
	tabs := MustFind[*Tabs](popup, "details-tabs")
	dirtyDot := MustFind[*Static](popup, "header-dirty")
	fileLabel := MustFind[*Static](popup, "header-file")
	themeLabel := MustFind[*Static](popup, "header-theme")
	status := MustFind[*Static](popup, "status")

	paneGeneral := MustFind[*Box](popup, "tab-general")
	paneLayout := MustFind[*Box](popup, "tab-layout")
	paneStyle := MustFind[*Box](popup, "tab-style")
	paneInfo := MustFind[*Box](popup, "tab-info")

	setDirty := func(v bool) {
		dirtyDot.SetFlag(FlagHidden, !v)
		Redraw(dirtyDot)
	}

	refreshHeader := func() {
		fileLabel.Set(proj.Name)
		themeLabel.Set(proj.Theme)
		Redraw(fileLabel)
		Redraw(themeLabel)
	}

	clearTabs := func() {
		empty := func(name string) Widget {
			s := NewStatic("empty-"+name, "muted", "(no widget selected)")
			s.Apply(theme)
			return s
		}
		_ = paneGeneral.Add(empty("g"))
		_ = paneLayout.Add(empty("l"))
		_ = paneStyle.Add(empty("s"))
		_ = paneInfo.Add(empty("i"))
		Relayout(paneGeneral)
	}

	rebuildPane := func(w Widget) {
		form := d.FormFor(w)
		if form == nil {
			lbl := func(name string) Widget {
				s := NewStatic("noform-"+name, "muted", "(no form registered for this widget)")
				s.Apply(theme)
				return s
			}
			_ = paneGeneral.Add(lbl("g"))
			_ = paneLayout.Add(lbl("l"))
			_ = paneStyle.Add(lbl("s"))
			_ = paneInfo.Add(lbl("i"))
			currentWidget, currentForm, currentLayout, currentParent = w, nil, nil, nil
			Relayout(paneGeneral)
			return
		}
		currentWidget = w
		currentForm = form
		currentLayout, currentParent = nil, nil

		// ---- General tab: ComponentForm + per-widget fields. ----
		general := NewFlex("general-stack", "", Stretch, 0)
		general.SetFlag(FlagVertical, true)

		f := NewForm("form-general", "", "", form)
		f.Apply(theme)

		generalContent := NewFlex("general-content", "", Stretch, 0)
		generalContent.SetFlag(FlagVertical, true)

		v := reflect.ValueOf(form).Elem()
		t := v.Type()

		isFirst := true
		addSection := func(title string, fv reflect.Value) {
			if !isFirst {
				addSeparator(generalContent, theme)
			}
			isFirst = false
			addFormSection(f, generalContent, theme, title, fv)
		}
		for i := 0; i < v.NumField(); i++ {
			sf := t.Field(i)
			fv := v.Field(i)
			if !sf.Anonymous || fv.Kind() != reflect.Struct {
				continue
			}
			addSection(sectionTitle(sf.Type.Name()), fv)
		}
		addSection(sectionTitle(t.Name()), v)
		_ = f.Add(generalContent)
		_ = general.Add(f)
		_ = paneGeneral.Add(general)

		// ---- Layout tab: parent's LayoutForm (if any) + Computed. ----
		layoutStack := NewFlex("layout-stack", "", Stretch, 0)
		layoutStack.SetFlag(FlagVertical, true)

		hasLayoutForm := false
		if parent := w.Parent(); parent != nil {
			if pf := d.FormFor(parent); pf != nil {
				if cf, ok := pf.(inspector.ContainerForm); ok {
					if lf := cf.LayoutForm(parent, w); lf != nil {
						currentLayout = lf
						currentParent = parent
						hasLayoutForm = true

						hdr := NewStatic("layout-header", "section",
							fmt.Sprintf(" Layout in %s%s ", widgetKind(parent), idSuffix(parent)))
						hdr.Apply(theme)
						_ = layoutStack.Add(hdr)

						lfForm := NewForm("layout-form", "", "", lf)
						lfGroup := NewFormGroup("layout-fg", "", "", true, 0)
						lfForm.Apply(theme)
						lfGroup.Apply(theme)
						BuildFormGroup(lfForm, lfGroup, "", theme)
						_ = lfForm.Add(lfGroup)
						_ = layoutStack.Add(lfForm)

						addSeparator(layoutStack, theme)
					}
				}
			}
		}
		if !hasLayoutForm {
			note := NewStatic("layout-note", "muted",
				"  parent has no per-child layout parameters")
			note.Apply(theme)
			_ = layoutStack.Add(note)
			addSeparator(layoutStack, theme)
		}
		// Always-visible Computed block.
		comp := NewStatic("layout-computed-header", "section", " Computed ")
		comp.Apply(theme)
		_ = layoutStack.Add(comp)

		x, y, ww, wh := w.Bounds()
		hw, hh := w.Hint()
		bounds := NewStatic("layout-bounds", "",
			fmt.Sprintf("  bounds   x=%d  y=%d  w=%d  h=%d", x, y, ww, wh))
		bounds.Apply(theme)
		_ = layoutStack.Add(bounds)
		hint := NewStatic("layout-hint", "",
			fmt.Sprintf("  hint     w=%d  h=%d", hw, hh))
		hint.Apply(theme)
		_ = layoutStack.Add(hint)
		_ = paneLayout.Add(layoutStack)

		// ---- Style tab: StyleForm if available. ----
		styleStack := NewFlex("style-stack", "", Stretch, 0)
		styleStack.SetFlag(FlagVertical, true)
		if styleForm := form.Style(); styleForm != nil {
			styleW := NewForm("form-style", "", "", styleForm)
			styleW.Apply(theme)
			styleContent := NewFlex("style-content", "", Stretch, 0)
			styleContent.SetFlag(FlagVertical, true)
			addFormSection(styleW, styleContent, theme, "Style",
				reflect.ValueOf(styleForm).Elem())
			_ = styleW.Add(styleContent)
			_ = styleStack.Add(styleW)
		} else {
			note := NewStatic("style-note", "muted",
				"  no style form available")
			note.Apply(theme)
			_ = styleStack.Add(note)
		}
		_ = paneStyle.Add(styleStack)

		// ---- Info tab: read-only kind summary. ----
		infoStack := NewFlex("info-stack", "", Stretch, 0)
		infoStack.SetFlag(FlagVertical, true)

		hdr := NewStatic("info-header", "section", " Widget ")
		hdr.Apply(theme)
		_ = infoStack.Add(hdr)

		typ := NewStatic("info-type", "",
			fmt.Sprintf("  type     %s", widgetKind(w)))
		typ.Apply(theme)
		_ = infoStack.Add(typ)

		idStr := w.ID()
		if idStr == "" {
			idStr = "—"
		}
		idLine := NewStatic("info-id", "", "  id       "+idStr)
		idLine.Apply(theme)
		_ = infoStack.Add(idLine)

		parentDesc := "—"
		if p := w.Parent(); p != nil {
			parentDesc = fmt.Sprintf("%s%s", widgetKind(p), idSuffix(p))
		}
		parentLine := NewStatic("info-parent", "", "  parent   "+parentDesc)
		parentLine.Apply(theme)
		_ = infoStack.Add(parentLine)

		childCount := "—"
		if c, ok := w.(Container); ok {
			childCount = fmt.Sprintf("%d", len(c.Children()))
		}
		childLine := NewStatic("info-children", "", "  children "+childCount)
		childLine.Apply(theme)
		_ = infoStack.Add(childLine)

		flagsLine := NewStatic("info-flags", "", "  flags    "+flagSummary(w))
		flagsLine.Apply(theme)
		_ = infoStack.Add(flagsLine)
		_ = paneInfo.Add(infoStack)

		Relayout(paneGeneral)
		Relayout(paneLayout)
		Relayout(paneStyle)
		Relayout(paneInfo)
	}

	tree.On(EvtSelect, func(_ Widget, _ Event, _ ...any) bool {
		node := tree.Selected()
		if node == nil {
			return false
		}
		w, ok := node.Data().(Widget)
		if !ok || w == nil {
			return false
		}
		currentNode = node
		rebuildPane(w)
		setStatus(status, fmt.Sprintf("selected %s%s", widgetKind(w), idSuffix(w)))
		return false
	})

	// Tab keys 1–4 jump to the matching detail tab. We attach to the
	// popup so the binding is active anywhere inside it; the handler
	// short-circuits before letting the framework treat the digit as
	// input to a focused field.
	popup.On(EvtKey, func(_ Widget, _ Event, data ...any) bool {
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
			tabs.Set(0)
			return true
		case "2":
			tabs.Set(1)
			return true
		case "3":
			tabs.Set(2)
			return true
		case "4":
			tabs.Set(3)
			return true
		}
		return false
	})

	apply := func() {
		if currentForm == nil || currentWidget == nil {
			setStatus(status, "no widget selected")
			return
		}
		currentForm.Store(currentWidget)
		if currentLayout != nil && currentParent != nil {
			currentLayout.Store(currentParent, currentWidget)
		}
		Relayout(currentWidget)
		if currentNode != nil {
			currentNode.SetText(treeLabel(currentWidget))
			Redraw(tree)
		}
		w := currentWidget
		rebuildPane(w)
		setDirty(true)
		setStatus(status, fmt.Sprintf("applied → %s%s", widgetKind(w), idSuffix(w)))
	}

	MustFind[*Button](popup, "apply-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		apply()
		return false
	})

	MustFind[*Button](popup, "reset-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if currentWidget == nil {
			return false
		}
		rebuildPane(currentWidget)
		setStatus(status, "reset from widget state")
		return false
	})

	writeFile := func() {
		var buf bytes.Buffer
		if err := d.GenerateFile(inspector.ModeBuilder, &buf, proj.Package, proj.FuncName); err != nil {
			setStatus(status, "generate failed: "+err.Error())
			return
		}
		if err := os.WriteFile(proj.OutPath, buf.Bytes(), 0o644); err != nil {
			setStatus(status, "write failed: "+err.Error())
			return
		}
		setDirty(false)
		setStatus(status, "wrote "+proj.OutPath)
	}

	MustFind[*Button](popup, "save-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		writeFile()
		return false
	})
	MustFind[*Button](popup, "generate-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		writeFile()
		return false
	})
	MustFind[*Button](popup, "run-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		setStatus(status, "Run: not implemented yet")
		return false
	})
	MustFind[*Button](popup, "settings-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		openSettingsDialog(ui, proj, refreshHeader, func(msg string) { setStatus(status, msg) })
		return false
	})

	// refreshTree rebuilds the tree from the live target. Called
	// after any structural mutation (Add, Remove, Move) so the
	// picker and selection reflect the new shape.
	refreshTree := func() {
		root := NewTreeNode("")
		root.Add(buildWidgetTreeNode(target))
		tree.SetRoot(root)
		Redraw(tree)
	}

	resolveAddParent := func() Container {
		if currentWidget != nil {
			if c, ok := currentWidget.(Container); ok {
				return c
			}
			if p := currentWidget.Parent(); p != nil {
				return p
			}
		}
		return target
	}

	MustFind[*Button](popup, "add-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		parent := resolveAddParent()
		openAddChildDialog(ui, theme, status, parent, d, func(child Widget) {
			child.Apply(theme)
			Relayout(child)
			refreshTree()
			setDirty(true)
			setStatus(status, fmt.Sprintf("added %s under %s%s",
				widgetKind(child), widgetKind(parent), idSuffix(parent)))
		})
		return false
	})

	// selectAfterMutation re-points currentWidget at the given widget
	// (typically the parent of a deleted child) and rebuilds the
	// detail panes around it. clearWhenNil collapses the panes back to
	// the empty placeholder when nothing sensible remains selected.
	selectAfterMutation := func(w Widget) {
		if w == nil {
			currentWidget, currentForm, currentLayout, currentParent, currentNode = nil, nil, nil, nil, nil
			clearTabs()
			return
		}
		currentWidget = w
		rebuildPane(w)
	}

	MustFind[*Button](popup, "del-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if currentWidget == nil {
			setStatus(status, "Delete: no widget selected")
			return false
		}
		if currentWidget == target {
			setStatus(status, "Delete: cannot remove the root")
			return false
		}
		victim := currentWidget
		parent := victim.Parent()
		if err := d.Remove(victim); err != nil {
			setStatus(status, "delete failed: "+err.Error())
			return false
		}
		Relayout(parent)
		refreshTree()
		setDirty(true)
		setStatus(status, fmt.Sprintf("removed %s%s", widgetKind(victim), idSuffix(victim)))
		// Snap selection up to the parent so the property panel does
		// not keep showing a widget that is no longer in the tree.
		selectAfterMutation(parent)
		return false
	})

	moveSibling := func(delta int) bool {
		if currentWidget == nil {
			setStatus(status, "Move: no widget selected")
			return false
		}
		parent := currentWidget.Parent()
		if parent == nil {
			setStatus(status, "Move: no parent")
			return false
		}
		siblings := parent.Children()
		idx := -1
		for i, s := range siblings {
			if s == currentWidget {
				idx = i
				break
			}
		}
		if idx < 0 {
			setStatus(status, "Move: child not in parent")
			return false
		}
		newIdx := idx + delta
		if newIdx < 0 || newIdx >= len(siblings) {
			setStatus(status, "Move: already at boundary")
			return false
		}
		if err := d.Move(currentWidget, parent, newIdx); err != nil {
			setStatus(status, "move failed: "+err.Error())
			return false
		}
		Relayout(parent)
		refreshTree()
		setDirty(true)
		direction := "↑"
		if delta > 0 {
			direction = "↓"
		}
		setStatus(status, fmt.Sprintf("moved %s%s %s",
			widgetKind(currentWidget), idSuffix(currentWidget), direction))
		return false
	}

	MustFind[*Button](popup, "up-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		return moveSibling(-1)
	})
	MustFind[*Button](popup, "down-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		return moveSibling(+1)
	})

	// Initialise empty-state placeholders so the popup is well-formed
	// before any selection.
	clearTabs()
	refreshHeader()
	setDirty(false)

	// Toggle the designer popup with Ctrl+Space. The handler runs
	// before UI's hardcoded global-key block, so it can intercept
	// Ctrl+Space without modifying the framework. We never close on
	// Ctrl+Space — ESC handles that, via UI's existing layer logic.
	ui.On(EvtKey, func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 0 {
			return false
		}
		ev, ok := data[0].(*tcell.EventKey)
		if !ok {
			return false
		}
		if ev.Key() == tcell.KeyNUL ||
			(ev.Key() == tcell.KeyRune && ev.Str() == " " && ev.Modifiers()&tcell.ModCtrl != 0) {
			ui.Popup(-1, -1, 0, 0, popup)
			return true
		}
		return false
	})

	ui.Run()
}

// Nerd-font glyphs used on the tree-action toolbar. Pre-resolved
// codepoints (FontAwesome subset, identical to the ones already used
// in cmd/malwerk/glyphs.go) so terminals with a configured Nerd Font
// render them as proper icons; non-Nerd terminals fall back to a
// generic missing-glyph indicator, which is still better than a row
// of three-letter abbreviations.
const (
	iconAdd    = "" // nf-fa-plus
	iconDelete = "" // nf-fa-trash
	iconUp     = "" // nf-fa-arrow_up
	iconDown   = "" // nf-fa-arrow_down
)

// buildDesignerPopup constructs the popup as a 3-row × 2-column
// Grid with rendered grid lines. The Grid replaces the previous
// Box+VFlex combination so the whole frame, the title, the
// header-band, the tree/details split and the status line are all
// expressed as cells with shared dividers.
//
// Layout:
//
//	row 0 (h=1)        : title + project chrome (spans both cols)
//	row 1 (h=-1, fills): tree pane | details pane
//	row 2 (h=1)        : status line (spans both cols)
//
//	col 0 (w=34): tree pane — TreeWidgets above an icon toolbar
//	col 1 (w=-1): details pane — Tabs / Switcher / Apply-Reset toolbar
//
// The 96×36 outer hint keeps the popup a focused modal rather than
// letting it grow to fill the terminal; grid lines do the visual
// framing the Box border used to do.
func buildDesignerPopup(theme *Theme, target Widget, proj *project) Container {
	b := NewBuilder(theme).
		Grid("designer-grid", 3, 2, true).Hint(96, 36).
		Columns(34, -1).Rows(1, -1, 1).
		// ===== Row 0: header band, spanning both columns =====
		Cell(0, 0, 2, 1).
		HFlex("header-band", Center, 1).
		Static("designer-title", " Designer ").Class("title").
		Static("header-file", proj.Name).
		Static("header-dirty", "●").Hint(1, 1).
		Static("header-spacer-1", "  ").
		Static("header-theme", proj.Theme).
		Spacer().Hint(-1, 0).
		Button("save-btn", "Save").
		Button("generate-btn", "Generate").
		Button("run-btn", "Run").
		Button("settings-btn", "Settings").
		End(). // closes header-band
		// ===== Row 1, col 0: tree pane =====
		Cell(0, 1, 1, 1).
		VFlex("tree-pane", Stretch, 0).
		TreeWidgets("tree", target).Hint(0, -1).
		HFlex("tree-toolbar", Center, 2).Hint(0, 1).
		Button("add-btn", iconAdd).
		Button("del-btn", iconDelete).
		Button("up-btn", iconUp).
		Button("down-btn", iconDown).
		End(). // closes tree-toolbar
		End(). // closes tree-pane
		// ===== Row 1, col 1: details pane =====
		Cell(1, 1, 1, 1).
		VFlex("details-pane", Stretch, 0).
		Tabs("details-tabs", "General", "Layout", "Style", "Info").Hint(0, 2).
		Switcher("details-switcher", true).Hint(0, -1).
		// Tab panes: borderless boxes so they read as plain panels;
		// the Tabs strip and the surrounding grid lines already
		// supply the framing, an extra box border just adds noise.
		Box("tab-general", "").Border("none").
		End(). // closes tab-general Box (no-op)
		Box("tab-layout", "").Border("none").
		End(). // closes tab-layout Box (no-op)
		Box("tab-style", "").Border("none").
		End(). // closes tab-style Box (no-op)
		Box("tab-info", "").Border("none").
		End(). // closes tab-info Box (no-op)
		End(). // closes details-switcher
		HFlex("details-toolbar", End, 2).Hint(0, 1).
		Spacer().Hint(-1, 0).
		Button("apply-btn", "Apply").
		Button("reset-btn", "Reset").
		End(). // closes details-toolbar
		End(). // closes details-pane
		// ===== Row 2: status line, spanning both columns =====
		Cell(0, 2, 2, 1).
		Static("status", " ").
		End() // closes designer-grid

	return b.Container()
}

// openSettingsDialog edits the project struct in place. Fields
// covered: filename label (header chrome), output path, package
// name, function name, theme label. Theme is currently a label only —
// switching themes at runtime would require re-applying every widget,
// which is out of scope for the PoC; the field is here so the chrome
// reflects the theme the user thinks they're designing in.
func openSettingsDialog(ui *UI, proj *project, onChange func(), notify func(string)) {
	dialog := ui.NewBuilder().
		Dialog("settings-dialog", "Project Settings").
		Class("dialog").
		VFlex("settings-body", Stretch, 1).Padding(1, 2).
		Static("settings-prompt", "Edit codegen and chrome settings.").Hint(0, 1).
		HFlex("settings-name-row", Start, 1).Hint(0, 1).
		Static("settings-name-label", "Name      ").Hint(12, 1).
		Input("settings-name", proj.Name).Hint(40, 1).
		End().
		HFlex("settings-out-row", Start, 1).Hint(0, 1).
		Static("settings-out-label", "Output    ").Hint(12, 1).
		Input("settings-out", proj.OutPath).Hint(40, 1).
		End().
		HFlex("settings-pkg-row", Start, 1).Hint(0, 1).
		Static("settings-pkg-label", "Package   ").Hint(12, 1).
		Input("settings-pkg", proj.Package).Hint(40, 1).
		End().
		HFlex("settings-fn-row", Start, 1).Hint(0, 1).
		Static("settings-fn-label", "Func name ").Hint(12, 1).
		Input("settings-fn", proj.FuncName).Hint(40, 1).
		End().
		HFlex("settings-theme-row", Start, 1).Hint(0, 1).
		Static("settings-theme-label", "Theme     ").Hint(12, 1).
		Input("settings-theme", proj.Theme).Hint(40, 1).
		End().
		HFlex("settings-buttons", End, 2).Hint(0, 1).
		Spacer().Hint(-1, 0).
		Button("settings-ok", "Save").
		Button("settings-cancel", "Cancel").
		End().
		End().
		Class("").
		Container()

	commit := func() {
		nameIn := MustFind[*Input](dialog, "settings-name")
		outIn := MustFind[*Input](dialog, "settings-out")
		pkgIn := MustFind[*Input](dialog, "settings-pkg")
		fnIn := MustFind[*Input](dialog, "settings-fn")
		themeIn := MustFind[*Input](dialog, "settings-theme")

		newName := strings.TrimSpace(nameIn.Get())
		if newName == "" {
			newName = filepath.Base(outIn.Get())
		}
		proj.Name = newName
		proj.OutPath = strings.TrimSpace(outIn.Get())
		proj.Package = strings.TrimSpace(pkgIn.Get())
		proj.FuncName = strings.TrimSpace(fnIn.Get())
		proj.Theme = strings.TrimSpace(themeIn.Get())
		ui.Close()
		if onChange != nil {
			onChange()
		}
		if notify != nil {
			notify("settings updated")
		}
	}

	MustFind[*Button](dialog, "settings-ok").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		commit()
		return true
	})
	MustFind[*Button](dialog, "settings-cancel").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		ui.Close()
		return true
	})

	ui.Popup(-1, -1, 0, 0, dialog)
}

// buildWidgetTreeNode mirrors w as a TreeNode, recursing into
// containers. Each node carries the widget itself as opaque data,
// matching the convention used by widgets.NewTreeWidgets — but built
// fresh from the live target rather than a snapshot, so callers can
// rebuild the tree after structural mutations.
func buildWidgetTreeNode(w Widget) *TreeNode {
	node := NewTreeNode(treeLabel(w), w)
	node.Expand()
	if c, ok := w.(Container); ok {
		for _, child := range c.Children() {
			node.Add(buildWidgetTreeNode(child))
		}
	}
	return node
}

// openAddChildDialog shows a modal picker listing every kind the
// Designer knows about. Selecting a kind (Enter on the list, or the
// [Add] button) calls Designer.Add on parent, runs onAdded with the
// new widget, and closes the dialog. ESC or [Cancel] dismisses
// without changes.
func openAddChildDialog(ui *UI, _ *Theme, status *Static, parent Container, d *inspector.Designer, onAdded func(Widget)) {
	kinds := d.KindNames()
	if len(kinds) == 0 {
		setStatus(status, "no kinds registered")
		return
	}

	dialog := ui.NewBuilder().
		Dialog("add-child-dialog", "Add Child").
		Class("dialog").
		VFlex("add-child-body", Stretch, 1).
		Static("add-child-prompt",
			fmt.Sprintf("Add child to %s%s", widgetKind(parent), idSuffix(parent))).
		List("add-child-list", kinds...).Hint(28, 8).
		HFlex("add-child-buttons", End, 2).
		Button("add-child-ok", "Add").
		Button("add-child-cancel", "Cancel").
		End().
		End().
		Class("").
		Container()

	list := MustFind[*List](dialog, "add-child-list")
	commit := func() {
		idx := list.Selected()
		if idx < 0 || idx >= len(kinds) {
			ui.Close()
			return
		}
		kindName := kinds[idx]
		child, err := d.Add(parent, kindName)
		if err != nil {
			setStatus(status, "add failed: "+err.Error())
			ui.Close()
			return
		}
		ui.Close()
		if onAdded != nil {
			onAdded(child)
		}
	}

	MustFind[*Button](dialog, "add-child-ok").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		commit()
		return true
	})
	MustFind[*Button](dialog, "add-child-cancel").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		ui.Close()
		return true
	})
	list.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		commit()
		return true
	})

	ui.Popup(-1, -1, 0, 0, dialog)
}

// registerKinds is the same registration table inspector-poc uses.
// Register validates each Kind at registration time; any mismatch
// panics here rather than later in Load.
func registerKinds(d *inspector.Designer) {
	register := func(t reflect.Type, mk func() inspector.WidgetForm) {
		if err := d.Register(inspector.Kind{Type: t, Make: mk}); err != nil {
			panic(err)
		}
	}
	register(reflect.TypeOf((*Static)(nil)),
		func() inspector.WidgetForm { return &StaticForm{} })
	register(reflect.TypeOf((*Grid)(nil)),
		func() inspector.WidgetForm { return &GridForm{} })
	register(reflect.TypeOf((*Flex)(nil)),
		func() inspector.WidgetForm { return &FlexForm{} })
	register(reflect.TypeOf((*Input)(nil)),
		func() inspector.WidgetForm { return &InputForm{} })
	register(reflect.TypeOf((*Button)(nil)),
		func() inspector.WidgetForm { return &ButtonForm{} })
	register(reflect.TypeOf((*Box)(nil)),
		func() inspector.WidgetForm { return &BoxForm{} })
	register(reflect.TypeOf((*Card)(nil)),
		func() inspector.WidgetForm { return &CardForm{} })
	register(reflect.TypeOf((*Checkbox)(nil)),
		func() inspector.WidgetForm { return &CheckboxForm{} })
	register(reflect.TypeOf((*List)(nil)),
		func() inspector.WidgetForm { return &ListForm{} })

	// Forms added with the per-widget *-form.go files.
	register(reflect.TypeOf((*Breadcrumb)(nil)),
		func() inspector.WidgetForm { return &BreadcrumbForm{} })
	register(reflect.TypeOf((*Clock)(nil)),
		func() inspector.WidgetForm { return &ClockForm{} })
	register(reflect.TypeOf((*Collapsible)(nil)),
		func() inspector.WidgetForm { return &CollapsibleForm{} })
	register(reflect.TypeOf((*Combo)(nil)),
		func() inspector.WidgetForm { return &ComboForm{} })
	register(reflect.TypeOf((*Deck)(nil)),
		func() inspector.WidgetForm { return &DeckForm{} })
	register(reflect.TypeOf((*Dialog)(nil)),
		func() inspector.WidgetForm { return &DialogForm{} })
	register(reflect.TypeOf((*Digits)(nil)),
		func() inspector.WidgetForm { return &DigitsForm{} })
	register(reflect.TypeOf((*Editor)(nil)),
		func() inspector.WidgetForm { return &EditorForm{} })
	register(reflect.TypeOf((*Filter)(nil)),
		func() inspector.WidgetForm { return &FilterForm{} })
	register(reflect.TypeOf((*Indicator)(nil)),
		func() inspector.WidgetForm { return &IndicatorForm{} })
	register(reflect.TypeOf((*Marquee)(nil)),
		func() inspector.WidgetForm { return &MarqueeForm{} })
	register(reflect.TypeOf((*Progress)(nil)),
		func() inspector.WidgetForm { return &ProgressForm{} })
	register(reflect.TypeOf((*Rule)(nil)),
		func() inspector.WidgetForm { return &RuleForm{} })
	register(reflect.TypeOf((*Scanner)(nil)),
		func() inspector.WidgetForm { return &ScannerForm{} })
	register(reflect.TypeOf((*Select)(nil)),
		func() inspector.WidgetForm { return &SelectForm{} })
	register(reflect.TypeOf((*Shortcuts)(nil)),
		func() inspector.WidgetForm { return &ShortcutsForm{} })
	register(reflect.TypeOf((*Spinner)(nil)),
		func() inspector.WidgetForm { return &SpinnerForm{} })
	register(reflect.TypeOf((*Styled)(nil)),
		func() inspector.WidgetForm { return &StyledForm{} })
	register(reflect.TypeOf((*Switcher)(nil)),
		func() inspector.WidgetForm { return &SwitcherForm{} })
	register(reflect.TypeOf((*Table)(nil)),
		func() inspector.WidgetForm { return &TableForm{} })
	register(reflect.TypeOf((*Tabs)(nil)),
		func() inspector.WidgetForm { return &TabsForm{} })
	register(reflect.TypeOf((*Terminal)(nil)),
		func() inspector.WidgetForm { return &TerminalForm{} })
	register(reflect.TypeOf((*Text)(nil)),
		func() inspector.WidgetForm { return &TextForm{} })
	register(reflect.TypeOf((*Tiles)(nil)),
		func() inspector.WidgetForm { return &TilesForm{} })
	// TreeFS and TreeWidgets publish their inner *Tree to the
	// builder, so a single *Tree registration must serve all three
	// roles. TreeForm covers the common case; specialised TreeFS /
	// TreeWidgets editing happens through dedicated dialogs.
	register(reflect.TypeOf((*Tree)(nil)),
		func() inspector.WidgetForm { return &TreeForm{} })
	register(reflect.TypeOf((*Typeahead)(nil)),
		func() inspector.WidgetForm { return &TypeaheadForm{} })
	register(reflect.TypeOf((*Typewriter)(nil)),
		func() inspector.WidgetForm { return &TypewriterForm{} })
	register(reflect.TypeOf((*Viewport)(nil)),
		func() inspector.WidgetForm { return &ViewportForm{} })
}

// widgetKind returns the Go type name without the "*widgets." prefix.
func widgetKind(w Widget) string {
	t := reflect.TypeOf(w)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}

// idSuffix returns "#id" when w has an id, "" otherwise. Used to
// build human-readable labels like "Static#title" in the breadcrumb,
// status line, and Layout-tab header — never appearing for unnamed
// widgets that would otherwise read as "Static#".
func idSuffix(w Widget) string {
	if id := w.ID(); id != "" {
		return "#" + id
	}
	return ""
}

// flagSummary collapses the most relevant runtime flags into a short
// space-separated label for the Info tab. Focused / hovered are not
// included because they're transient and would flicker; Skip / Hidden
// / Disabled are persistent enough to be worth showing.
func flagSummary(w Widget) string {
	parts := make([]string, 0, 4)
	if w.Flag(FlagFocused) {
		parts = append(parts, "focused")
	}
	if w.Flag(FlagSkip) {
		parts = append(parts, "skip")
	}
	if w.Flag(FlagHidden) {
		parts = append(parts, "hidden")
	}
	if w.Flag(FlagDisabled) {
		parts = append(parts, "disabled")
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, " ")
}

// treeLabel formats a widget as it appears in the tree, matching
// widgets.NewTreeWidgets' internal convention.
func treeLabel(w Widget) string {
	if id := w.ID(); id != "" {
		return fmt.Sprintf("%s (#%s)", widgetKind(w), id)
	}
	return widgetKind(w)
}

// addSeparator appends a thin horizontal rule to container. Used as
// a visual divider between adjacent form sections (Component /
// Static / Layout / Info) so they read as distinct blocks rather
// than running together.
func addSeparator(container Container, theme *Theme) {
	rule := NewHRule("", "thin")
	rule.Apply(theme)
	_ = container.Add(rule)
}

// addFormSection appends a section header (Static) and a FormGroup
// rendering only the directly-declared fields of v (no recursion
// into embedded structs) to container. The FormGroup's controls
// write back through f.Update — addressable reflect.Values along v
// reach the underlying struct fields, so each section can target a
// different embedded level while sharing one *Form data root.
func addFormSection(f *Form, container Container, theme *Theme, title string, v reflect.Value) {
	hdr := NewStatic("section-"+title, "section", " "+title+" ")
	hdr.Apply(theme)
	_ = container.Add(hdr)

	fg := NewFormGroup("fg-"+title, "", "", true, 0)
	fg.Apply(theme)
	BuildFormGroupAt(f, fg, v, "", theme)
	_ = container.Add(fg)
}

// sectionTitle turns a Go type name into a section header. Strips
// the "Form" suffix so "ComponentForm" → "Component", "StaticForm"
// → "Static". Types that don't end in "Form" pass through.
func sectionTitle(typeName string) string {
	if strings.HasSuffix(typeName, "Form") {
		return strings.TrimSuffix(typeName, "Form")
	}
	return typeName
}

// setStatus updates the status line and refreshes. Status messages are
// short-lived; a real inspector would dispatch a transient toast.
func setStatus(s *Static, msg string) {
	s.Set(msg + "    " + time.Now().Format("15:04:05"))
	Redraw(s)
}

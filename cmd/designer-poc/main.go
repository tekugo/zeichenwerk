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
// Ctrl+Space opens the designer in a popup. The popup contains a Tree
// of the target widgets, a Properties pane that renders the selected
// widget's WidgetForm via BuildFormGroup, and an Apply / Reset /
// Generate toolbar. Apply mutates the live target so the preview
// reflects edits immediately. ESC closes the popup; Ctrl+Q quits.
package main

import (
	"bytes"
	"fmt"
	"os"
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

const outputPath = "/tmp/designer-poc-out.go"

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

	// Build the designer popup as a free-standing Container. Same
	// layout as the previous standalone version: tree | properties on
	// top, toolbar on bottom.
	popup := buildDesignerPopup(theme, target)

	// State shared between popup widgets.
	var (
		currentWidget Widget
		currentNode   *TreeNode
		currentForm   inspector.WidgetForm
		currentLayout LayoutForm
		currentParent Container
	)

	tree := MustFind[*Tree](popup, "tree")
	pane := MustFind[*Box](popup, "form-pane")
	status := MustFind[*Static](popup, "status")

	rebuildPane := func(w Widget) {
		form := d.FormFor(w)
		if form == nil {
			lbl := NewStatic("no-form", "", "(no form registered for this widget)")
			lbl.Apply(theme)
			pane.Add(lbl)
			currentWidget, currentForm, currentLayout, currentParent = nil, nil, nil, nil
			Relayout(pane)
			return
		}
		currentWidget = w
		currentForm = form
		currentLayout, currentParent = nil, nil

		stack := NewFlex("form-stack", "", Stretch, 0)
		stack.SetFlag(FlagVertical, true)

		// One Form widget pointing at the underlying *WidgetForm
		// struct. The Update handlers BuildFormGroupAt installs all
		// hold reflect.Values rooted at this same form.Data, so
		// rendering each struct level into its own FormGroup still
		// writes back to the right field.
		f := NewForm("form", "", "", form)
		f.Apply(theme)

		formContent := NewFlex("form-content", "", Stretch, 0)
		formContent.SetFlag(FlagVertical, true)

		v := reflect.ValueOf(form).Elem()
		t := v.Type()

		// addSection wraps addFormSection with a thin horizontal
		// separator before each section except the first, so the
		// Component / Static / Input / etc groups read as distinct
		// blocks rather than running together.
		isFirst := true
		addSection := func(title string, fv reflect.Value) {
			if !isFirst {
				addSeparator(formContent, theme)
			}
			isFirst = false
			addFormSection(f, formContent, theme, title, fv)
		}

		for i := 0; i < v.NumField(); i++ {
			sf := t.Field(i)
			fv := v.Field(i)
			if !sf.Anonymous || fv.Kind() != reflect.Struct {
				continue
			}
			addSection(sectionTitle(sf.Type.Name()), fv)
		}
		// Outer struct's own (non-embedded) fields.
		addSection(sectionTitle(t.Name()), v)

		_ = f.Add(formContent)
		_ = stack.Add(f)

		// Style section. Backed by a separate Form widget pointing
		// at the same *core.StyleForm snapshot ComponentForm holds
		// internally, so edits made here flow back through
		// ComponentForm.Store on Apply (Modifiable creates a non-
		// fixed child for previously-themed styles, so the next
		// generation stops being suppressed).
		if styleForm := form.Style(); styleForm != nil {
			addSeparator(stack, theme)

			styleW := NewForm("style-form", "", "", styleForm)
			styleW.Apply(theme)

			styleContent := NewFlex("style-content", "", Stretch, 0)
			styleContent.SetFlag(FlagVertical, true)

			addFormSection(styleW, styleContent, theme, "Style", reflect.ValueOf(styleForm).Elem())

			_ = styleW.Add(styleContent)
			_ = stack.Add(styleW)
		}

		if parent := w.Parent(); parent != nil {
			if pf := d.FormFor(parent); pf != nil {
				if cf, ok := pf.(inspector.ContainerForm); ok {
					if lf := cf.LayoutForm(parent, w); lf != nil {
						currentLayout = lf
						currentParent = parent

						addSeparator(stack, theme)

						hdr := NewStatic("layout-header", "",
							fmt.Sprintf(" Layout in %s#%s ", widgetKind(parent), parent.ID()))
						hdr.Apply(theme)
						_ = stack.Add(hdr)

						lfForm := NewForm("layout-form", "", "", lf)
						lfGroup := NewFormGroup("layout-fg", "", "", true, 0)
						lfForm.Apply(theme)
						lfGroup.Apply(theme)
						BuildFormGroup(lfForm, lfGroup, "", theme)
						_ = lfForm.Add(lfGroup)
						_ = stack.Add(lfForm)
					}
				}
			}
		}

		// Read-only Info section: bounds + hint as computed by the
		// most recent Layout pass. Not part of the form because these
		// are runtime-derived rather than user-editable; they refresh
		// whenever the user re-selects the widget.
		addSeparator(stack, theme)
		infoHdr := NewStatic("info-header", "", " Info ")
		infoHdr.Apply(theme)
		_ = stack.Add(infoHdr)

		x, y, ww, wh := w.Bounds()
		hw, hh := w.Hint()
		info := NewStatic("info", "",
			fmt.Sprintf("bounds: x=%d y=%d w=%d h=%d   hint: w=%d h=%d", x, y, ww, wh, hw, hh))
		info.Apply(theme)
		_ = stack.Add(info)

		_ = pane.Add(stack)
		Relayout(pane)
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
		setStatus(status, fmt.Sprintf("selected %s#%s", widgetKind(w), w.ID()))
		return false
	})

	MustFind[*Button](popup, "apply-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if currentForm == nil || currentWidget == nil {
			setStatus(status, "no widget selected")
			return false
		}
		currentForm.Store(currentWidget)
		if currentLayout != nil && currentParent != nil {
			currentLayout.Store(currentParent, currentWidget)
		}
		// Target is part of the live UI now — Relayout walks up to
		// the visible root so the preview reflects the change.
		Relayout(currentWidget)

		if currentNode != nil {
			currentNode.SetText(treeLabel(currentWidget))
			Redraw(tree)
		}

		// Rebuild the pane so the Info section picks up the new
		// post-layout bounds. This also re-Loads the form from the
		// just-stored state, which is a no-op for fields the user
		// edited and a refresh for any value the widget may have
		// massaged during Store (e.g. clamped, normalized).
		w := currentWidget
		rebuildPane(w)

		setStatus(status, fmt.Sprintf("applied → %s#%s", widgetKind(w), w.ID()))
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

	MustFind[*Button](popup, "generate-btn").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		var buf bytes.Buffer
		// GenerateFile produces a complete, gofmt-canonical Go
		// file (package + imports + func) that can be saved
		// directly and re-run. GenerateFragment would produce
		// just the chained expression — useful for splicing into
		// existing code, but less convenient as a save-to-disk
		// target.
		if err := d.GenerateFile(inspector.ModeBuilder, &buf, "main", "BuildUI"); err != nil {
			setStatus(status, "generate failed: "+err.Error())
			return false
		}
		if err := os.WriteFile(outputPath, buf.Bytes(), 0o644); err != nil {
			setStatus(status, "write failed: "+err.Error())
			return false
		}
		setStatus(status, "wrote "+outputPath)
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

	// resolveAddParent picks the container that a new child should
	// land in: the selected widget if it's a container, otherwise
	// its nearest container ancestor, falling back to the target.
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
			setStatus(status, fmt.Sprintf("added %s under %s#%s",
				widgetKind(child), widgetKind(parent), parent.ID()))
		})
		return false
	})

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
		// Most terminals send Ctrl+Space as the NUL byte (KeyNUL).
		// Some send Rune ' ' with the Ctrl modifier set; cover both.
		if ev.Key() == tcell.KeyNUL ||
			(ev.Key() == tcell.KeyRune && ev.Str() == " " && ev.Modifiers()&tcell.ModCtrl != 0) {
			ui.Popup(-1, -1, 0, 0, popup)
			return true
		}
		return false
	})

	ui.Run()
}

// buildDesignerPopup constructs the popup as a standalone Box-wrapped
// VFlex. Each pane has its own toolbar: the tree column carries the
// structural actions ([+ Child], [Generate]) and the properties
// column carries the form actions ([Apply], [Reset]). The status
// line spans the full width at the bottom. Returns the root Box so
// the caller can pass it directly to ui.Popup.
func buildDesignerPopup(theme *Theme, target Widget) Container {
	b := NewBuilder(theme).
		Box("designer-frame", " Designer ").Hint(96, 36).
		VFlex("designer-main", Stretch, 0).
		HFlex("designer-upper", Stretch, 1).Hint(0, -1).
		// Left column: tree + tree-related toolbar.
		VFlex("tree-pane", Stretch, 0).Hint(36, 0).
		TreeWidgets("tree", target).Hint(0, -1).
		HFlex("tree-toolbar", Start, 2).Hint(0, 3).
		Button("add-btn", "+ Child").
		Button("generate-btn", "Generate").
		End(). // closes tree-toolbar HFlex
		End(). // closes tree-pane VFlex
		// Right column: properties pane + form-related toolbar.
		// width=-1 → take the rest of designer-upper's horizontal
		// space after tree-pane's fixed 36 cols. height is the
		// cross-axis here (HFlex with Stretch alignment) so it's
		// filled regardless of what we put.
		VFlex("props-pane", Stretch, 0).Hint(-1, 0).
		Box("form-pane", " Properties ").Hint(0, -1).
		End(). // closes form-pane Box
		HFlex("props-toolbar", Start, 2).Hint(0, 3).
		Button("apply-btn", "Apply").
		Button("reset-btn", "Reset").
		End(). // closes props-toolbar HFlex
		End(). // closes props-pane VFlex
		End(). // closes designer-upper HFlex
		// Status line spans the full popup width at the bottom.
		Static("status", "                                                                ").Hint(0, 1).
		End(). // closes designer-main VFlex
		End()  // closes designer-frame Box (no-op at root)

	return b.Container()
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
			fmt.Sprintf("Add child to %s#%s", widgetKind(parent), parent.ID())).
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
}

// widgetKind returns the Go type name without the "*widgets." prefix.
func widgetKind(w Widget) string {
	t := reflect.TypeOf(w)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
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
	hdr := NewStatic("section-"+title, "", " "+title+" ")
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

package designer

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// widgetKind returns the Go type name of w without the package
// prefix — e.g. "Static" for a *widgets.Static. Used everywhere the
// inspector wants to label a widget for the user.
func widgetKind(w core.Widget) string {
	t := reflect.TypeOf(w)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}

// idSuffix returns "#id" when w has an id, "" otherwise. Lets code
// build human-readable labels like "Static#title" that read cleanly
// for unnamed widgets too ("Static").
func idSuffix(w core.Widget) string {
	if id := w.ID(); id != "" {
		return "#" + id
	}
	return ""
}

// flagSummary collapses the persistent runtime flags into a short
// space-separated label. Focus / hover are intentionally omitted —
// they're transient and would flicker.
func flagSummary(w core.Widget) string {
	parts := make([]string, 0, 4)
	if w.Flag(core.FlagFocused) {
		parts = append(parts, "focused")
	}
	if w.Flag(core.FlagSkip) {
		parts = append(parts, "skip")
	}
	if w.Flag(core.FlagHidden) {
		parts = append(parts, "hidden")
	}
	if w.Flag(core.FlagDisabled) {
		parts = append(parts, "disabled")
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, " ")
}

// treeLabel formats a widget as it appears in the tree, matching
// widgets.NewTreeWidgets' internal convention so a tree rebuilt by
// hand reads identically to one built by the helper.
func treeLabel(w core.Widget) string {
	if id := w.ID(); id != "" {
		return fmt.Sprintf("%s (#%s)", widgetKind(w), id)
	}
	return widgetKind(w)
}

// buildWidgetTreeNode mirrors w as a TreeNode, recursing into
// containers. Each node carries the widget as opaque data so the
// inspector's select handler can recover the widget from the
// selected node.
func buildWidgetTreeNode(w core.Widget) *widgets.TreeNode {
	node := widgets.NewTreeNode(treeLabel(w), w)
	node.Expand()
	if c, ok := w.(core.Container); ok {
		for _, child := range c.Children() {
			node.Add(buildWidgetTreeNode(child))
		}
	}
	return node
}

// sectionTitle turns a Go type name into a section header by
// trimming the "Form" suffix: "ComponentForm" → "Component",
// "StaticForm" → "Static". Types without that suffix pass through.
func sectionTitle(typeName string) string {
	if strings.HasSuffix(typeName, "Form") {
		return strings.TrimSuffix(typeName, "Form")
	}
	return typeName
}

// addSeparator appends a thin horizontal rule to container so
// adjacent form sections read as distinct blocks.
func addSeparator(container core.Container, theme *core.Theme) {
	rule := widgets.NewHRule("", "thin")
	rule.Apply(theme)
	_ = container.Add(rule)
}

// addFormSection appends a section header plus a FormGroup rendering
// only the directly-declared fields of v (no recursion into embedded
// structs) to container. Sharing one *Form across multiple sections
// lets each section target a different embedded level while writes
// reach the underlying struct fields through addressable
// reflect.Values.
func addFormSection(f *widgets.Form, container core.Container, theme *core.Theme, title string, v reflect.Value) {
	hdr := widgets.NewStatic("section-"+title, "section", " "+title+" ")
	hdr.Apply(theme)
	_ = container.Add(hdr)

	fg := widgets.NewFormGroup("fg-"+title, "", "", true, 0)
	fg.Apply(theme)
	widgets.BuildFormGroupAt(f, fg, v, "", theme)
	_ = container.Add(fg)
}

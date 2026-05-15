package inspector

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// widgetKind returns the Go type name of w without the package
// prefix — e.g. "Static" for a *widgets.Static. Used everywhere
// the inspector wants to label a widget for the user.
func widgetKind(w core.Widget) string {
	t := reflect.TypeOf(w)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}

// idSuffix returns "#id" when w has an id, "" otherwise.
func idSuffix(w core.Widget) string {
	if id := w.ID(); id != "" {
		return "#" + id
	}
	return ""
}

// treeLabel formats a widget as it appears in the tree, matching
// widgets.NewTreeWidgets' convention so the manually rebuilt tree
// reads identically to one built by the helper.
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

// extractComponent pulls the embedded widgets.Component out of w
// via reflection. Every shipped widget embeds Component as an
// anonymous field by value, so the address of the embedded struct
// is the *Component a ComponentForm.Load needs.
//
// Returns (nil, false) when w doesn't embed Component or the
// reflection traversal fails — the caller falls back to rendering
// the Widget interface directly.
func extractComponent(w core.Widget) (*widgets.Component, bool) {
	v := reflect.ValueOf(w)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return nil, false
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	f := v.FieldByName("Component")
	if !f.IsValid() || f.Kind() != reflect.Struct || !f.CanAddr() {
		return nil, false
	}
	c, ok := f.Addr().Interface().(*widgets.Component)
	return c, ok
}

// fieldLabel returns the user-facing label for a struct field —
// the `label:"..."` tag when present, the field name otherwise.
// Trims surrounding whitespace so a tag like `"  Hint width  "`
// renders cleanly.
func fieldLabel(sf reflect.StructField) string {
	if tag := sf.Tag.Get("label"); tag != "" {
		return strings.TrimSpace(tag)
	}
	return sf.Name
}

// formatValue renders v as a one-line string suitable for a
// "Label: value" line. Strings are quoted to disambiguate empty
// strings ("") from missing fields; numbers and bools render
// naturally; everything else falls back to fmt.Sprintf("%v").
func formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return fmt.Sprintf("%q", v.String())
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

package zeichenwerk

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
	. "github.com/tekugo/zeichenwerk/v2/values"
	. "github.com/tekugo/zeichenwerk/v2/widgets"
)

// Inspector provides a debugging interface for exploring widget hierarchies.
type Inspector struct {
	ui        Container
	container Container
	current   Widget
	widgets   *List
	styles    *List
}

// NewInspector creates a new Inspector for the given root container.
func NewInspector(root Container) *Inspector {
	inspector := &Inspector{
		container: root,
		current:   root,
	}
	inspector.BuildUI()
	return inspector
}

// BuildUI constructs the Inspector interface.
func (i *Inspector) BuildUI() {
	ui := FindRoot(i.container).(*UI)
	i.ui = ui.NewBuilder().
		Class("inspector").
		Box("inspector-box", "Inspector").Border("double").
		VFlex("inspector", Stretch, 0).
		Tabs("inspector-tabs").
		Switcher("inspector-switcher", true).
		Tab("Widgets").
		HFlex("inspector-widgets", Stretch, 0).Flag(FlagVertical).
		Static("breadcrumbs", "Breadcrumbs").
		HFlex("inspector-content", Stretch, 0).
		VFlex("inspector-lists", Stretch, 0).
		Box("widget-box", "Widgets").Border("round").
		List("children").Hint(30, 15).
		End().
		Box("styles-box", "Styles").Border("round").
		List("styles").Hint(30, 10).
		End().
		End().
		VFlex("info-boxes", Stretch, 0).
		Box("widget-info-box", "Information").Border("round").
		Text("widget-info", []string{}, false, 0).Hint(50, 15).
		End().
		Box("style-info-box", "Information").Border("round").
		Text("style-info", []string{}, false, 0).Hint(50, 10).
		End().
		End().
		End().
		End().
		Tab("Debug Log").
		Table("inspector-log-table", ui.tableLog, false).
		Border("none").
		Border("grid", "double-thin").Foreground("grid", "$fg0").
		Border(":focused", "none").
		Border("grid:focused", "double-thin").
		Hint(50, 25).
		End().
		End().
		Class("").
		Container()

	i.widgets = MustFind[*List](i.ui, "children")
	i.styles = MustFind[*List](i.ui, "styles")

	OnActivate(i.widgets, i.Activate)
	OnSelect(i.widgets, i.SelectWidget)
	OnSelect(i.styles, i.SelectStyle)

	OnKey(i.widgets, func(event *tcell.EventKey) bool {
		switch event.Key() {
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if i.container.Parent() != nil {
				i.container = i.container.Parent()
				i.Refresh()
			}
			return true
		}
		return false
	})

	i.Refresh()
}

// SelectWidget handles selection of a widget from the children list.
func (i *Inspector) SelectWidget(index int) bool {
	items := i.widgets.Items()
	if index < 0 || index >= len(items) {
		return true
	}
	id := items[index]
	i.current = Find(i.container, id)
	if i.current != nil {
		// Get styles via interface
		type stylesProvider interface {
			Styles() []string
		}
		var styles []string
		if sp, ok := i.current.(stylesProvider); ok {
			styles = sp.Styles()
			for i, s := range styles {
				if s == "" {
					styles[i] = "(default)"
				}
			}
			slices.Sort(styles)
		} else {
			styles = []string{}
		}
		Update(i.ui, "styles", styles)
		Update(i.ui, "widget-info", strings.Split(widgetDetails(i.current), "\n"))
	} else {
		Update(i.ui, "styles", []string{})
	}
	return true
}

// Activate navigates into a container widget.
func (i *Inspector) Activate(_ int) bool {
	if i.current != nil {
		if container, ok := i.current.(Container); ok {
			i.container = container
			i.current = nil
			i.Refresh()
		}
	}
	return true
}

// SelectStyle displays details of the selected style.
func (i *Inspector) SelectStyle(index int) bool {
	items := i.styles.Items()
	if index < 0 || index >= len(items) {
		return true
	}
	name := items[index]
	style := i.current.Style(name)
	if style != nil {
		Update(i.ui, "style-info", strings.Split(style.Info(), "\n"))
	} else {
		i.styles.Log(i.styles, Error, "Style not found in widget", "name", name, "ID", i.styles.ID())
	}
	i.ui.Refresh()
	return true
}

// Refresh updates the inspector UI to reflect current state.
func (i *Inspector) Refresh() {
	if i.container == nil {
		i.ui.Log(i.ui, Error, "No current container!")
		return
	}
	i.ui.Log(i.ui, Debug, "Refresh inspector", "ID", i.container.ID())
	children := i.container.Children()
	items := make([]string, len(children))
	for j, child := range children {
		if i.current == nil {
			i.current = child
		}
		items[j] = child.ID()
	}
	Update(i.ui, "children", items)

	path := i.container.ID()
	current := i.container.Parent()
	for current != nil {
		path = current.ID() + " > " + path
		current = current.Parent()
	}
	Update(i.ui, "breadcrumbs", path)
	i.ui.Refresh()
}

// Hint returns the preferred size hint for the inspector.
func (i *Inspector) Hint() (int, int) {
	return i.ui.Hint()
}

// UI returns the inspector's UI container.
func (i *Inspector) UI() Container {
	return i.ui
}

// widgetDetails generates a detailed multi-line string for a widget.
func widgetDetails(w Widget) string {
	result := widgetType(w) + "\n"
	result += "ID        : '" + w.ID() + "'\n"
	parent := "<nil>"
	if w.Parent() != nil {
		parent = "'" + w.Parent().ID() + "'"
	}
	result += "Parent-ID : " + parent + "\n"
	x, y, ww, h := w.Bounds()
	result += fmt.Sprintf("Bounds    : x=%d, y=%d, w=%d, h=%d\n", x, y, ww, h)
	x, y, ww, h = w.Content()
	result += fmt.Sprintf("Content   : x=%d, y=%d, w=%d, h=%d\n", x, y, ww, h)
	prefW, prefH := w.Hint()
	result += fmt.Sprintf("Hint      : w=%d, h=%d\n", prefW, prefH)
	result += "State     : " + w.State() + "\n"
	flags := make([]string, 0)
	if w.Flag(FlagFocusable) {
		flags = append(flags, "focusable")
	}
	if w.Flag(FlagFocused) {
		flags = append(flags, "focused")
	}
	if w.Flag(FlagHovered) {
		flags = append(flags, "hovered")
	}
	result += "Flags     : " + strings.Join(flags, ", ") + "\n"
	// Grid-specific information
	if summary, ok := w.(Summarizer); ok {
		result += fmt.Sprintf("Summary   : %s\n", summary.Summary())
	}
	return result
}

// widgetType returns the widget type name without package prefix.
func widgetType(w Widget) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", w), "*zeichenwerk.")
}

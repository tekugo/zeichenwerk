package zeichenwerk

import (
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Inspector struct {
	ui        Container
	container Container
	current   Widget
}

func NewInspector(root Container) *Inspector {
	inspector := &Inspector{
		container: root,
		current:   root,
	}
	inspector.BuildUI()
	return inspector
}

func (i *Inspector) BuildUI() {
	i.ui = NewBuilder(TokyoNightTheme()).
		Class("inspector").
		Box("inspector-box", "Inspector").Border("", "double").
		Flex("inspector", "vertical", "stretch", 0).Background("", "$comments").
		Label("breadcrumbs", "Breadcrumbs", 0).
		Flex("inspector-content", "horizontal", "stretch", 0).
		Flex("inspector-lists", "vertical", "stretch", 0).
		Box("widget-box", "Widgets").Border("", "round").
		List("children", []string{}).Border("", "").Border("focus", "").Hint(30, 15).
		End().
		Box("styles-box", "Styles").Border("", "round").
		List("styles", []string{}).Border("", "").Border("focus", "").Hint(30, 10).
		End().
		End().
		Flex("info-boxes", "vertical", "stretch", 0).
		Box("widget-info-box", "Information").Border("", "round").
		Text("widget-info", []string{}, false, 0).Hint(50, 15).
		End().
		Box("style-info-box", "Information").Border("", "round").
		Text("style-info", []string{}, false, 0).Hint(50, 10).
		End().
		End().
		End().
		End().
		End().
		Class("").
		Container()

	HandleListEvent(i.ui, "children", "select", i.SelectWidget)
	HandleListEvent(i.ui, "children", "activate", i.Activate)
	HandleListEvent(i.ui, "styles", "select", i.SelectStyle)
	HandleKeyEvent(i.ui, "children", func(widget Widget, event *tcell.EventKey) bool {
		switch event.Key() {
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if i.container.Parent() != nil {
				container, ok := i.container.Parent().(Container)
				if !ok {
					widget.Log(container, "error", "Parent is no container! %T", i.container.Parent())
				}
				i.container = container
				widget.Log(i.ui, "debug", "Going back to %s", i.container.ID())
				i.Refresh()
			}
			return true
		}
		return false
	})

	i.Refresh()
}

func (i *Inspector) SelectWidget(list *List, event string, index int) bool {
	id := list.Items[list.Index]
	i.current = i.container.Find(id, false)
	if i.current != nil {
		styles := i.current.Styles()
		for i, str := range styles {
			if str == "" {
				styles[i] = "(default)"
			}
		}
		slices.Sort(styles)
		Update(i.ui, "styles", styles)
		Update(i.ui, "widget-info", strings.Split(WidgetDetails(i.current), "\n"))
	} else {
		Update(i.ui, "styles", []string{})
	}

	return true
}

func (i *Inspector) Activate(_ *List, _ string, _ int) bool {
	if i.current != nil {
		container, ok := i.current.(Container)
		if ok {
			i.container = container
			i.current = nil
			i.Refresh()
		}
	}
	return true
}

func (i *Inspector) SelectStyle(list *List, _ string, _ int) bool {
	name := list.Items[list.Index]
	style := i.current.Style(name)
	list.Log(list, "debug", "Style name if %s is %s", i.current.ID(), name)
	if style != nil {
		Update(i.ui, "style-info", strings.Split(style.Info(), "\n"))
	} else {
		list.Log(list, "error", "Style %s not found in widget %s", name, list.ID())
	}

	i.ui.Refresh()
	return true
}

func (i *Inspector) Refresh() {
	if i.container == nil {
		i.ui.Log(i.ui, "error", "No current container!")
		return
	}
	i.ui.Log(i.ui, "debug", "Refresh inspector %s", i.container.ID())
	children := i.container.Children(false)
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

func (i *Inspector) Hint() (int, int) {
	return i.ui.Hint()
}

func (i *Inspector) UI() Container {
	return i.ui
}

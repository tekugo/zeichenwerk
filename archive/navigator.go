package zeichenwerk

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

type NavNode struct {
	Name     string
	Icon     string
	Code     rune
	Level    int
	Expanded bool
	Shortcut string
	Children []NavNode
}

func (node *NavNode) Add(name, icon, shortcut string) *NavNode {
	child := NavNode{Name: name, Icon: icon, Expanded: false, Shortcut: shortcut}
	if node.Children == nil {
		node.Children = make([]NavNode, 0, 5)
		node.Children = append(node.Children, child)
	}
	child.Code = rune(int('0') + len(node.Children))
	return &child
}

type Navigator struct {
	BaseWidget
	Input  *Input
	Root   NavNode
	Result []*NavNode
	Flat   bool
	Index  int
	Offset int
}

func NewNavigator(id, name string) *Navigator {
	nav := &Navigator{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Root:       NavNode{Name: name, Children: make([]NavNode, 0, 10)},
		Input:      NewInput(id + "-search"),
	}
	nav.Input.On("change", func(w Widget, ev string, data ...any) bool {
		nav.Search()
		nav.Refresh()
		return true
	})
	nav.Input.On("enter", func(w Widget, ev string, data ...any) bool {
		// Forward enter from input to navigator
		items := nav.Items()
		if len(items) > 0 && nav.Index >= 0 && nav.Index < len(items) {
			nav.Emit("activate", nav.Index, items[nav.Index])
		}
		return true
	})
	return nav
}

func (nav *Navigator) Handle(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyUp:
			nav.Up()
			return true
		case tcell.KeyDown:
			nav.Down()
			return true
		case tcell.KeyRight:
			items := nav.Items()
			if nav.Index >= 0 && nav.Index < len(items) {
				node := items[nav.Index]
				if len(node.Children) > 0 {
					node.Expanded = true
					nav.Refresh()
				}
			}
			return true
		case tcell.KeyLeft:
			items := nav.Items()
			if nav.Index >= 0 && nav.Index < len(items) {
				node := items[nav.Index]
				if node.Expanded {
					node.Expanded = false
					nav.Refresh()
				}
			}
			return true
		case tcell.KeyEnter:
			items := nav.Items()
			if nav.Index >= 0 && nav.Index < len(items) {
				nav.Emit("activate", nav.Index, items[nav.Index])
			}
			return true
		default:
			return nav.Input.Handle(event)
		}
	}
	return false
}

func (nav *Navigator) Items() []*NavNode {
	if nav.Result != nil {
		return nav.Result
	}
	// For tree view, collect all visible nodes
	items := make([]*NavNode, 0)
	for i := range nav.Root.Children {
		nav.collectVisible(&nav.Root.Children[i], &items, 0)
	}
	return items
}

func (nav *Navigator) collectVisible(node *NavNode, items *[]*NavNode, level int) {
	node.Level = level
	*items = append(*items, node)
	if node.Expanded {
		for i := range node.Children {
			nav.collectVisible(&node.Children[i], items, level+1)
		}
	}
}

func (nav *Navigator) Search() {
	term := nav.Input.Text
	if term == "" {
		nav.Result = nil
		nav.Flat = false
		nav.Index = 0
		nav.Offset = 0
		return
	}
	nav.Flat = true
	nav.Result = make([]*NavNode, 0)
	nav.flatten(&nav.Root, term)

	// Reset selection and offset when search results change
	nav.Index = 0
	nav.Offset = 0
}

func (nav *Navigator) flatten(node *NavNode, term string) {
	if strings.Contains(strings.ToLower(node.Name), strings.ToLower(term)) {
		// Create a copy with level 0 for flat list
		flatNode := *node
		flatNode.Level = 0
		flatNode.Expanded = false // Flat list doesn't show expansion state usually
		nav.Result = append(nav.Result, &flatNode)
	}
	for i := range node.Children {
		nav.flatten(&node.Children[i], term)
	}
}

func (nav *Navigator) Refresh() {
	Redraw(nav)
}

func (nav *Navigator) Up() {
	if nav.Index > 0 {
		nav.Index--
		nav.adjust()
		nav.Emit("select", nav.Index)
	}
}

func (nav *Navigator) Down() {
	items := nav.Items()
	if nav.Index < len(items)-1 {
		nav.Index++
		nav.adjust()
		nav.Emit("select", nav.Index)
	}
}

func (nav *Navigator) adjust() {
	_, _, _, h := nav.Content()
	if nav.Input != nil {
		h--
	}
	if h <= 0 {
		return
	}

	// Ensure selected item is visible
	if nav.Index < nav.Offset {
		nav.Offset = nav.Index
	} else if nav.Index >= nav.Offset+h {
		nav.Offset = nav.Index - h + 1
	}

	// Don't scroll past the beginning
	if nav.Offset < 0 {
		nav.Offset = 0
	}

	items := nav.Items()
	// Don't scroll past the end
	maxScroll := max(len(items)-h, 0)
	if nav.Offset > maxScroll {
		nav.Offset = maxScroll
	}
}

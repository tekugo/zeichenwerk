package zeichenwerk

import (
	"maps"
	"slices"
)

type Switcher struct {
	BaseWidget
	Selected string
	Panes    map[string]Widget
}

func NewSwitcher(id string) *Switcher {
	return &Switcher{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Panes:      make(map[string]Widget),
	}
}

func (s *Switcher) Children(visible bool) []Widget {
	if visible {
		if s.Panes[s.Selected] != nil {
			return []Widget{s.Panes[s.Selected]}
		} else {
			return []Widget{}
		}
	} else {
		return slices.Collect(maps.Values(s.Panes))
	}
}

func (s *Switcher) Find(id string, visible bool) Widget {
	return Find(s, id, visible)
}

func (s *Switcher) FindAt(x, y int) Widget {
	return FindAt(s, x, y)
}

func (s *Switcher) Select(name string) {
	s.Selected = name
	s.Refresh()
}

func (s *Switcher) Set(name string, widget Widget) {
	s.Panes[name] = widget
	if s.Selected == "" {
		s.Selected = name
	}
	x, y, w, h := s.Content()
	widget.SetBounds(x, y, w, h)
}

func (s *Switcher) Layout() {
	x, y, w, h := s.Content()
	for _, widget := range s.Panes {
		widget.SetBounds(x, y, w, h)
	}
	Layout(s)
}

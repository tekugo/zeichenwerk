package tui

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

func (s *Switcher) Add(name string, widget Widget) {
	s.Panes[name] = widget
	if s.Selected == "" {
		s.Selected = name
	}
}

func (s *Switcher) Select(name string) {
	s.Selected = name
	s.Refresh()
}

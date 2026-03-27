package zeichenwerk

// Switcher represents a container widget that displays one pane at a time
// from a collection of named panes. It acts like a tabbed interface without
// visible tabs, where only the currently selected pane is visible.
//
// The switcher automatically selects the first pane added if no selection
// has been made. All panes are laid out to fill the switcher's content area
// but only the selected pane is visible and interactive.
type Switcher struct {
	Component
	selected int      // Index of the currently selected/visible pane
	panes    []Widget // Array the corresponding pane widgets
}

// NewSwitcher creates a new switcher container with the specified identifier.
// The switcher is initialized with an empty pane collection and no selected
// pane. It is non-focusable itself, as focus is managed by the contained
// panes.
//
// Parameters:
//   - id: Unique identifier for the switcher widget
//
// Returns:
//   - *Switcher: A new switcher widget instance
func NewSwitcher(id, class string) *Switcher {
	return &Switcher{
		Component: Component{id: id, class: class},
		panes:     make([]Widget, 0, 3),
	}
}

// Add adds a new pane to the switcher.
//
// Parameters:
//   - label: Name of the pane
//   - widget: Widget to be added as the pane content
func (s *Switcher) Add(widget Widget) {
	if widget == nil {
		return
	}
	widget.SetParent(s)
	s.panes = append(s.panes, widget)
	x, y, w, h := s.Content()
	widget.SetBounds(x, y, w, h)

	// Check if the pane is hidden
	widget.SetFlag("hidden", s.selected != len(s.panes)-1)
}

// Apply applies a theme's styles to the component.
func (s *Switcher) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("switcher"))
}

// Children returns the child widgets of the switcher based on visibility preference.
// This method supports two modes: visible-only and all children.
//
// Returns:
//   - []Widget: Slice of all child widgets
func (s *Switcher) Children() []Widget {
	return s.panes
}

// Hint determines the preferred size of the switcher.
// The preferred size is the maximum width and height of all children.
func (s *Switcher) Hint() (int, int) {
	// If a hint is set manually, we return it instead
	if s.hwidth != 0 && s.hheight != 0 {
		return s.hwidth, s.hheight
	}

	width := 0
	height := 0

	for _, child := range s.panes {
		cw, ch := child.Hint()
		width = max(cw, width)
		height = max(ch, height)
	}

	return width, height
}

// Refresh redraws the switcher.
func (s *Switcher) Refresh() {
	Redraw(s)
}

// Select sets the currently selected pane by index or id.
func (s *Switcher) Select(index any) {
	switch pane := index.(type) {
	case int:
		s.Log(s, "debug", "Hiding", "index", s.selected, "ID", s.panes[s.selected].ID())
		s.panes[s.selected].SetFlag("hidden", true)
		s.panes[s.selected].Dispatch(s.panes[s.selected], "hide")
		s.Log(s, "debug", "Showing", "index", pane, "ID", s.panes[pane].ID())
		s.panes[pane].SetFlag("hidden", false)
		s.panes[pane].Dispatch(s.panes[pane], "show")
		s.selected = pane
	case string:
		index := -1
		// Search for the pane with the given id
		for i, p := range s.panes {
			if p.ID() == pane {
				index = i
				break
			}
		}
		// If the pane was found, select it
		if index >= 0 {
			s.panes[s.selected].SetFlag("hidden", true)
			s.panes[s.selected].Dispatch(s.panes[s.selected], "hide")
			s.panes[index].SetFlag("hidden", false)
			s.panes[index].Dispatch(s.panes[index], "show")
			s.selected = index
		}
	}
	s.Refresh()
}

// Layout calculates and applies layout positioning for all panes in the
// switcher. All panes are positioned to fill the switcher's content area,
// ensuring they are ready for display when selected, though only the selected
// pane is visible.
//
// This method is called automatically by the UI system when the switcher's
// size changes or when the layout needs to be recalculated. The uniform
// sizing ensures smooth transitions when switching between panes.
func (s *Switcher) Layout() {
	x, y, w, h := s.Content()
	for _, widget := range s.panes {
		widget.SetBounds(x, y, w, h)
	}
	Layout(s)
}

// Render draws the switcher, which is the selected pane.
func (s *Switcher) Render(r *Renderer) {
	s.Component.Render(r)
	if s.selected >= 0 && s.selected < len(s.panes) {
		s.panes[s.selected].Render(r)
	}
}

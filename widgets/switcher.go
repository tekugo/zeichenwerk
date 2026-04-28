package widgets

import (
	"fmt"

	. "github.com/tekugo/zeichenwerk/core"
)

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

// ---- Widget Methods -------------------------------------------------------

// Apply applies a theme's styles to the component.
func (s *Switcher) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("switcher"))
}

// Hint determines the preferred size of the switcher.
// The preferred size is the maximum width and height of all children.
func (s *Switcher) Hint() (int, int) {
	// If a hint is set manually, we return it instead
	if s.hwidth != 0 || s.hheight != 0 {
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

// Render draws the switcher, which is the selected pane.
func (s *Switcher) Render(r *Renderer) {
	s.Component.Render(r)
	if s.selected >= 0 && s.selected < len(s.panes) {
		s.panes[s.selected].Render(r)
	}
}

// ---- Container Methods ----------------------------------------------------

// Add appends widget as a new pane. The pane is hidden immediately unless it
// is the first pane added. Returns ErrChildIsNil if widget is nil.
func (s *Switcher) Add(widget Widget, params ...any) error {
	if widget == nil {
		return ErrChildIsNil
	}
	widget.SetParent(s)
	s.panes = append(s.panes, widget)
	x, y, w, h := s.Content()
	widget.SetBounds(x, y, w, h)

	// Check if the pane is hidden
	widget.SetFlag(FlagHidden, s.selected != len(s.panes)-1)
	return nil
}

// Children returns all pane widgets regardless of visibility.
func (s *Switcher) Children() []Widget {
	return s.panes
}

// Layout calculates and applies layout positioning for all panes in the
// switcher. All panes are positioned to fill the switcher's content area,
// ensuring they are ready for display when selected, though only the selected
// pane is visible.
//
// This method is called automatically by the UI system when the switcher's
// size changes or when the layout needs to be recalculated. The uniform
// sizing ensures smooth transitions when switching between panes.
func (s *Switcher) Layout() error {
	x, y, w, h := s.Content()
	for _, widget := range s.panes {
		widget.SetBounds(x, y, w, h)
	}
	return Layout(s)
}

// ---- Selection ------------------------------------------------------------

// Select sets the visible pane. Pass an int index or a string widget ID.
// The previously visible pane receives EvtHide; the new one receives EvtShow.
// Panics if an int index is out of range, does nothing if an invalid type is
// passed.
func (s *Switcher) Select(index any) {
	switch pane := index.(type) {
	case int:
		s.Log(s, Debug, "Hiding", "index", s.selected, "ID", s.panes[s.selected].ID())
		s.panes[s.selected].SetFlag(FlagHidden, true)
		s.panes[s.selected].Dispatch(s.panes[s.selected], EvtHide)
		s.Log(s, Debug, "Showing", "index", pane, "ID", s.panes[pane].ID())
		s.panes[pane].SetFlag(FlagHidden, false)
		s.panes[pane].Dispatch(s.panes[pane], EvtShow)
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
			s.panes[s.selected].SetFlag(FlagHidden, true)
			s.panes[s.selected].Dispatch(s.panes[s.selected], EvtHide)
			s.panes[index].SetFlag(FlagHidden, false)
			s.panes[index].Dispatch(s.panes[index], EvtShow)
			s.selected = index
		}
	}
	s.Refresh()
}

// ---- Getter and Setter ----------------------------------------------------

// Get returns the index of the currently visible pane.
func (s *Switcher) Get() int {
	return s.selected
}

// Get sets the index of the visible pane.
// It is less powerful than Select, but uses int to mirror the Get() type.
func (s *Switcher) Set(index int) {
	s.Select(index)
}

// ---- Summary --------------------------------------------------------------

// Summary returns the active pane index for Dump output.
func (s *Switcher) Summary() string {
	return fmt.Sprintf("showing=%d", s.selected)
}

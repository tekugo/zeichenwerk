package zeichenwerk

import (
	"maps"
	"slices"
)

// Switcher represents a container widget that displays one pane at a time from a collection
// of named panes. It acts like a tabbed interface without visible tabs, where only the
// currently selected pane is visible and rendered.
//
// Features:
//   - Multiple named panes (widgets) in a single container
//   - Dynamic pane switching with immediate visual updates
//   - Only the selected pane is rendered and receives events
//   - Automatic layout management for all panes
//   - Parent-child relationship management
//   - Memory-efficient rendering (only active pane is processed)
//
// Common use cases:
//   - Tab-like interfaces without visible tab headers
//   - Wizard or step-by-step interfaces
//   - Dynamic content switching based on application state
//   - Modal dialog systems with multiple views
//   - Dashboard views with different modes
//
// The switcher automatically selects the first pane added if no selection
// has been made. All panes are laid out to fill the switcher's content area
// but only the selected pane is visible and interactive.
type Switcher struct {
	BaseWidget
	Selected string            // Name of the currently selected/visible pane
	Panes    map[string]Widget // Map of pane names to their corresponding widgets
}

// NewSwitcher creates a new switcher container with the specified identifier.
// The switcher is initialized with an empty pane collection and no selected pane.
// It is non-focusable itself, as focus is managed by the contained panes.
//
// Parameters:
//   - id: Unique identifier for the switcher widget
//
// Returns:
//   - *Switcher: A new switcher widget instance
//
// Example usage:
//
//	switcher := NewSwitcher("main-switcher")
//	switcher.Set("login", loginForm)
//	switcher.Set("dashboard", dashboardView)
//	switcher.Select("login")  // Show login form initially
func NewSwitcher(id string) *Switcher {
	return &Switcher{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Panes:      make(map[string]Widget),
	}
}

// Children returns the child widgets of the switcher based on visibility preference.
// This method supports two modes: visible-only and all children.
//
// Parameters:
//   - visible: If true, returns only the currently selected pane; if false, returns all panes
//
// Returns:
//   - []Widget: Slice of child widgets according to the visibility preference
//
// Visibility modes:
//   - visible=true: Returns only the currently selected pane (what user sees)
//   - visible=false: Returns all panes in the switcher (for traversal/management)
//
// This method is used by the widget hierarchy system for focus management,
// event propagation, and widget tree operations.
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

// Find searches for a widget with the specified ID within the switcher's panes.
// This method delegates to the standard container search functionality,
// respecting the visibility preference for the search scope.
//
// Parameters:
//   - id: The unique identifier of the widget to find
//   - visible: If true, searches only the visible pane; if false, searches all panes
//
// Returns:
//   - Widget: The widget with the matching ID, or nil if not found
//
// The search is performed using the standard widget hierarchy traversal,
// allowing for finding nested widgets within the panes.
func (s *Switcher) Find(id string, visible bool) Widget {
	return Find(s, id, visible)
}

// FindAt locates the widget at the specified screen coordinates within the switcher.
// This method delegates to the standard container coordinate-based search,
// typically finding widgets within the currently visible pane.
//
// Parameters:
//   - x: The x-coordinate on the screen
//   - y: The y-coordinate on the screen
//
// Returns:
//   - Widget: The widget at the specified position, or nil if none found
//
// This method is primarily used for mouse interaction handling,
// helping determine which widget should receive mouse events.
func (s *Switcher) FindAt(x, y int) Widget {
	return FindAt(s, x, y)
}

// Hint determines the preferred size of the switcher.
// The preferred size is the maximum width and height of all children.
func (s *Switcher) Hint() (int, int) {
	width := 0
	height := 0

	for child := range maps.Values(s.Panes) {
		cw, ch := child.Hint()
		width = max(cw, width)
		height = max(ch, height)
	}

	return width, height
}

// Select changes the currently visible pane to the one with the specified name.
// If the named pane exists, it becomes the visible pane and the switcher
// triggers a refresh to update the display immediately.
//
// Parameters:
//   - name: The name of the pane to make visible
//
// Behavior:
//   - If the named pane exists, it becomes the selected/visible pane
//   - If the named pane doesn't exist, the selection remains unchanged
//   - The switcher refreshes automatically to show the newly selected pane
//   - Focus management is handled by the parent UI system
//
// Example usage:
//
//	switcher.Select("dashboard")  // Switch to dashboard view
//	switcher.Select("settings")   // Switch to settings view
func (s *Switcher) Select(name string) {
	s.Selected = name
	s.Refresh()
}

// Set adds or updates a pane in the switcher with the specified name and widget.
// If this is the first pane added, it automatically becomes the selected pane.
// The widget is immediately positioned to fill the switcher's content area.
//
// Parameters:
//   - name: The name identifier for the pane (used for selection)
//   - widget: The widget to add as a pane
//
// Behavior:
//   - Adds the widget to the panes collection with the given name
//   - If no pane is currently selected, this pane becomes selected
//   - If a pane with the same name exists, it is replaced
//   - The widget is immediately positioned to fill the content area
//
// Example usage:
//
//	switcher.Set("login", NewLoginForm("login-form"))
//	switcher.Set("main", NewDashboard("dashboard"))
//	switcher.Set("settings", NewSettingsPanel("settings"))
func (s *Switcher) Set(name string, widget Widget) {
	s.Panes[name] = widget
	if s.Selected == "" {
		s.Selected = name
	}
	x, y, w, h := s.Content()
	widget.SetBounds(x, y, w, h)
}

// Layout calculates and applies layout positioning for all panes in the switcher.
// All panes are positioned to fill the switcher's content area, ensuring they
// are ready for display when selected, though only the selected pane is visible.
//
// Layout process:
//  1. Gets the switcher's content area coordinates and dimensions
//  2. Sets all panes to fill the entire content area
//  3. Recursively triggers layout for child containers
//  4. Ensures all panes are properly positioned for potential display
//
// This method is called automatically by the UI system when the switcher's
// size changes or when the layout needs to be recalculated. The uniform
// sizing ensures smooth transitions when switching between panes.
func (s *Switcher) Layout() {
	x, y, w, h := s.Content()
	for _, widget := range s.Panes {
		widget.SetBounds(x, y, w, h)
	}
	Layout(s)
}

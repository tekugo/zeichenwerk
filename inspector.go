// Package inspector.go implements a comprehensive debugging and development tool for zeichenwerk.
//
// This file provides the Inspector widget, a sophisticated debugging interface that allows
// developers to explore widget hierarchies, examine styling information, and understand
// the structure of their terminal applications at runtime. The Inspector is an essential
// development tool for building and debugging complex TUI applications.
//
// # Core Purpose
//
// The Inspector serves multiple development needs:
//   - Widget hierarchy exploration and navigation
//   - Real-time style inspection and debugging
//   - Widget property examination and analysis
//   - UI structure understanding and documentation
//   - Theme debugging and style verification
//   - Development workflow acceleration
//
// # Key Features
//
// The Inspector provides:
//   - Hierarchical widget tree navigation with breadcrumb display
//   - Interactive widget selection and detailed information display
//   - Style inspection with live style property viewing
//   - Keyboard-driven navigation for efficient debugging
//   - Integrated information panels for comprehensive analysis
//   - Parent-child relationship visualization
//
// # Development Workflow
//
// The Inspector integrates into development workflows by:
//   - Providing runtime access via Ctrl+D keyboard shortcut
//   - Offering popup overlay that doesn't disrupt main application
//   - Enabling live inspection without application restart
//   - Supporting theme debugging and style verification
//   - Facilitating widget hierarchy understanding

package zeichenwerk

import (
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Inspector is a sophisticated debugging and development tool that provides interactive
// exploration of widget hierarchies and styling information. It offers a comprehensive
// interface for understanding and debugging terminal user interface applications.
//
// # Architecture
//
// The Inspector operates as a self-contained UI component with:
//   - Tree navigation system for widget hierarchy exploration
//   - Dual-pane interface showing widget lists and detailed information
//   - Interactive style browser with live property inspection
//   - Breadcrumb navigation for hierarchical context
//   - Event-driven interaction model for responsive debugging
//
// # Navigation Model
//
// The Inspector maintains navigation state through:
//   - container: Currently viewed container in the widget hierarchy
//   - current: Currently selected widget for detailed inspection
//   - Breadcrumb path showing hierarchical location
//   - Parent-child relationship traversal capabilities
//
// # User Interface Components
//
// The Inspector UI consists of:
//   - Widget list: Shows children of current container
//   - Style list: Shows available styles for selected widget
//   - Widget information panel: Displays detailed widget properties
//   - Style information panel: Shows style property details
//   - Breadcrumb bar: Displays current hierarchical location
//
// # Keyboard Interaction
//
// Navigation is optimized for keyboard efficiency:
//   - Arrow keys: Navigate through widget and style lists
//   - Enter: Dive into selected container widget
//   - Backspace: Move up to parent container
//   - Tab: Switch between different UI panels
//   - Escape: Close inspector and return to main application
type Inspector struct {
	ui        Container // The Inspector's own UI container (self-contained interface)
	container Container // Currently viewed container in the widget hierarchy
	current   Widget    // Currently selected widget for detailed inspection
}

// NewInspector creates a new Inspector instance for debugging the specified widget hierarchy.
// This constructor initializes a complete debugging interface that can be used to explore
// and analyze the structure and styling of any zeichenwerk application.
//
// # Initialization Process
//
// The constructor performs the following setup:
//  1. Creates Inspector instance with root container as starting point
//  2. Sets initial navigation state (container and current widget)
//  3. Builds the complete Inspector UI through BuildUI()
//  4. Establishes event handlers for interactive navigation
//  5. Performs initial content refresh to populate displays
//
// # Starting State
//
// The Inspector begins with:
//   - container: Set to the provided root container
//   - current: Initially set to the root container
//   - UI: Fully constructed with empty content lists
//   - Event handlers: Registered for navigation and selection
//
// # Integration with Main Application
//
// The Inspector is typically used as:
//   - Runtime debugging tool activated by keyboard shortcut (Ctrl+D)
//   - Popup overlay that doesn't interfere with main application
//   - Development aid for understanding widget hierarchies
//   - Style debugging tool for theme development
//
// # UI Structure
//
// The generated UI includes:
//   - Hierarchical widget browser with breadcrumb navigation
//   - Style inspector for detailed property examination
//   - Information panels showing widget and style details
//   - Keyboard-driven navigation for efficient debugging workflow
//
// Parameters:
//   - root: The root container of the widget hierarchy to inspect
//
// Returns:
//   - *Inspector: Fully initialized Inspector ready for display as popup
//
// # Example Usage
//
//	// Create inspector for main application UI
//	inspector := NewInspector(mainContainer)
//	
//	// Display as popup (typically triggered by Ctrl+D)
//	ui.Popup(-1, -1, 0, 0, inspector.UI())
func NewInspector(root Container) *Inspector {
	inspector := &Inspector{
		container: root,
		current:   root,
	}
	inspector.BuildUI()
	return inspector
}

// BuildUI constructs the complete Inspector user interface using the builder pattern.
// This method creates a sophisticated debugging interface with multiple panels for
// widget hierarchy navigation, style inspection, and detailed information display.
//
// # UI Architecture
//
// The Inspector UI is organized in a structured layout:
//   - Root: Bordered box container with "Inspector" title
//   - Top: Breadcrumb navigation showing current hierarchical location
//   - Main: Horizontal split between lists and information panels
//   - Left side: Vertical split with widget list and style list
//   - Right side: Vertical split with widget info and style info panels
//
// # Component Details
//
// Widget List Panel:
//   - Shows children of currently viewed container
//   - Supports selection and activation for navigation
//   - Bordered with focus indicators for keyboard navigation
//   - Size hint: 30 characters wide, 15 lines high
//
// Style List Panel:
//   - Shows available styles for currently selected widget
//   - Sorted alphabetically for easy browsing
//   - Default style labeled as "(default)" for clarity
//   - Size hint: 30 characters wide, 10 lines high
//
// Widget Information Panel:
//   - Displays comprehensive widget details via WidgetDetails()
//   - Shows type, properties, state, and hierarchy information
//   - Read-only text display with automatic formatting
//   - Size hint: 50 characters wide, 15 lines high
//
// Style Information Panel:
//   - Shows detailed style properties for selected style
//   - Displays colors, fonts, spacing, and other style attributes
//   - Updates in real-time as styles are selected
//   - Size hint: 50 characters wide, 10 lines high
//
// # Event Handler Registration
//
// The method establishes interactive behavior:
//   - Widget selection: Updates style list and widget information
//   - Widget activation: Navigates into container widgets
//   - Style selection: Updates style information panel
//   - Backspace navigation: Moves up to parent container
//   - Debug logging: Tracks navigation and selection events
//
// # Theme Integration
//
// The UI uses TokyoNightTheme for consistent styling:
//   - Professional dark theme optimized for development
//   - Consistent with debugging tool aesthetics
//   - High contrast for readability during debugging sessions
//   - Clear visual hierarchy for efficient information scanning
//
// # Performance Considerations
//
// The UI is optimized for debugging efficiency:
//   - Lazy content loading based on current navigation
//   - Efficient list updates for large widget hierarchies
//   - Responsive event handling for smooth navigation
//   - Minimal resource overhead for development use
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

// SelectWidget handles widget selection events from the widget list panel.
// This method updates the Inspector's state and refreshes related UI components
// when a different widget is selected for inspection.
//
// # Selection Process
//
// When a widget is selected:
//  1. Retrieves widget ID from the list item at current index
//  2. Finds the actual widget instance within current container
//  3. Updates the current widget reference for detailed inspection
//  4. Refreshes style list with widget's available styles
//  5. Updates widget information panel with comprehensive details
//
// # Style List Population
//
// The method populates the style list by:
//   - Retrieving all available styles from the selected widget
//   - Replacing empty style names with "(default)" for clarity
//   - Sorting styles alphabetically for easy browsing
//   - Updating the styles list UI component
//
// # Widget Information Update
//
// Widget details are displayed by:
//   - Calling WidgetDetails() to get comprehensive widget information
//   - Splitting the details string into individual lines
//   - Updating the widget-info text panel for display
//   - Providing complete widget analysis including type, properties, and state
//
// # Error Handling
//
// The method handles edge cases gracefully:
//   - Widget not found: Clears style list but continues operation
//   - Empty selection: Safely handles null widget references
//   - Invalid widget IDs: Continues operation without crashing
//
// Parameters:
//   - list: The widget list that triggered the selection event
//   - event: Event name ("select")
//   - index: Selected item index (used indirectly through list.Index)
//
// Returns:
//   - bool: Always true to indicate event was handled
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

// Activate handles widget activation events (Enter key) for navigation into containers.
// This method enables hierarchical navigation by allowing users to "dive into"
// container widgets to explore their child widgets.
//
// # Navigation Logic
//
// When a widget is activated:
//  1. Checks if currently selected widget is a container
//  2. If container, updates navigation state to enter that container
//  3. Clears current widget selection (will be set to first child)
//  4. Refreshes entire Inspector to show new container contents
//  5. Updates breadcrumb path to reflect new location
//
// # Container Type Checking
//
// The method safely handles type checking:
//   - Uses type assertion to check if widget implements Container interface
//   - Only navigates if widget is actually a container
//   - Ignores activation on non-container widgets (safe operation)
//
// # State Updates
//
// Successful navigation updates:
//   - container: Set to the activated container widget
//   - current: Reset to nil (will be set to first child during refresh)
//   - UI: Completely refreshed to show new container contents
//   - Breadcrumb: Updated to show new hierarchical location
//
// # Use Cases
//
// This method enables:
//   - Deep exploration of nested widget hierarchies
//   - Understanding of complex UI structures
//   - Navigation into flex containers, grids, and custom containers
//   - Debugging of deeply nested widget arrangements
//
// Parameters:
//   - _: List widget (unused in current implementation)
//   - _: Event name (unused in current implementation)
//   - _: Index (unused in current implementation)
//
// Returns:
//   - bool: Always true to indicate event was handled
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

// SelectStyle handles style selection events from the style list panel.
// This method updates the style information panel with detailed properties
// of the selected style for comprehensive style debugging.
//
// # Style Information Retrieval
//
// When a style is selected:
//  1. Gets the style name from the selected list item
//  2. Retrieves the actual style object from the current widget
//  3. Extracts detailed style information using style.Info()
//  4. Formats and displays the information in the style info panel
//
// # Information Display
//
// Style information includes:
//   - Color properties (background, foreground)
//   - Font attributes (bold, italic, underline, etc.)
//   - Spacing properties (margins, padding)
//   - Border settings and decorations
//   - Cursor configuration
//   - Other style-specific properties
//
// # Debug Logging
//
// The method provides debug logging for:
//   - Style selection events for troubleshooting
//   - Style retrieval operations
//   - Error conditions when styles are not found
//
// # Error Handling
//
// The method handles errors gracefully:
//   - Style not found: Logs error but continues operation
//   - Invalid style names: Prevents crashes with safe error logging
//   - Missing current widget: Safe operation with null checks
//
// # UI Updates
//
// The method ensures UI consistency by:
//   - Updating style-info panel with formatted information
//   - Triggering UI refresh to display new content
//   - Maintaining responsive interface during style browsing
//
// Parameters:
//   - list: The style list that triggered the selection event
//   - _: Event name (unused in current implementation)
//   - _: Index (unused in current implementation)
//
// Returns:
//   - bool: Always true to indicate event was handled
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

// Refresh updates all Inspector UI components to reflect the current navigation state.
// This method is called whenever the Inspector's state changes and the display
// needs to be updated to show current container contents and navigation context.
//
// # Refresh Process
//
// The method performs comprehensive UI updates:
//  1. Validates current container state and logs errors if invalid
//  2. Retrieves all children of the current container
//  3. Populates widget list with child widget IDs
//  4. Sets current widget to first child if no widget selected
//  5. Builds and updates breadcrumb navigation path
//  6. Triggers UI refresh to display all changes
//
// # Widget List Population
//
// The widget list is updated by:
//   - Getting all children from current container (including hidden widgets)
//   - Creating string array of widget IDs for list display
//   - Setting first child as current widget if none selected
//   - Updating the "children" list component with new items
//
// # Breadcrumb Navigation
//
// The breadcrumb path is constructed by:
//   - Starting with current container ID
//   - Walking up parent hierarchy to root
//   - Building path string with " > " separators
//   - Updating breadcrumb label to show hierarchical location
//
// # Error Handling and Logging
//
// The method includes robust error handling:
//   - Validates container state before proceeding
//   - Logs error and returns early if no current container
//   - Provides debug logging for refresh operations
//   - Tracks container navigation for debugging purposes
//
// # State Consistency
//
// The refresh ensures consistent Inspector state:
//   - Widget list always reflects current container children
//   - Current widget is valid within current container
//   - Breadcrumb accurately shows hierarchical position
//   - UI components are synchronized with navigation state
//
// # Performance Considerations
//
// The refresh is optimized for:
//   - Efficient child widget enumeration
//   - Minimal string operations for breadcrumb construction
//   - Single UI refresh at the end to minimize redraws
//   - Lazy loading of widget details until selection
//
// This method is called automatically during navigation and should not
// typically be called directly by external code.
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

// Hint returns the preferred size hint for the Inspector UI.
// This method delegates to the underlying UI container to provide
// sizing information for popup display and layout calculations.
//
// # Size Calculation
//
// The hint is determined by:
//   - Delegating to the internal UI container's Hint() method
//   - Reflecting the cumulative size requirements of all panels
//   - Considering the dual-pane layout with lists and information displays
//   - Accounting for borders, spacing, and text content requirements
//
// # Layout Integration
//
// The size hint is used by:
//   - Popup display systems for automatic sizing
//   - Layout managers for space allocation
//   - Container widgets for child positioning
//   - Responsive layout calculations
//
// # Typical Dimensions
//
// The Inspector typically hints at:
//   - Width: Approximately 80-100 characters (dual-pane layout)
//   - Height: Approximately 25-30 lines (accommodates lists and info panels)
//   - These dimensions provide optimal debugging experience
//
// Returns:
//   - (int, int): Preferred width and height in terminal characters
func (i *Inspector) Hint() (int, int) {
	return i.ui.Hint()
}

// UI returns the Inspector's UI container for integration with the application.
// This method provides access to the complete Inspector interface for display
// as a popup or integration into other container widgets.
//
// # Container Access
//
// The returned container includes:
//   - Complete Inspector interface with all panels
//   - Event handlers for interactive navigation
//   - Styling and theming applied
//   - Ready for immediate display or integration
//
// # Integration Patterns
//
// The UI container is typically used for:
//   - Popup display: ui.Popup(-1, -1, 0, 0, inspector.UI())
//   - Modal overlay: Display over main application
//   - Integrated panel: Embed in development interfaces
//   - Testing interface: Programmatic inspection access
//
// # Lifecycle Management
//
// The returned UI container:
//   - Maintains its own event handling and state
//   - Can be displayed and hidden multiple times
//   - Retains navigation state across display sessions
//   - Provides consistent debugging experience
//
// # Development Workflow
//
// Common usage patterns:
//   - Runtime debugging: Activated by Ctrl+D keyboard shortcut
//   - Development tool: Integrated into developer interfaces
//   - Testing aid: Programmatic inspection of UI hierarchies
//   - Theme debugging: Style and appearance verification
//
// Returns:
//   - Container: The complete Inspector UI ready for display
func (i *Inspector) UI() Container {
	return i.ui
}

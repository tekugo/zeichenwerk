package zeichenwerk

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func FindUI(widget Widget) *UI {
	current := widget
	for current != nil {
		if ui, ok := current.(*UI); ok {
			return ui
		}
		current = current.Parent()
	}
	return nil
}

// HandleInputEvent registers an event handler for Input widget events within a container.
// This function simplifies the process of handling input-specific events by automatically
// finding the widget, performing type assertions, and validating event data.
//
// The function performs the following operations:
//   - Locates the widget by ID within the container hierarchy
//   - Registers an event handler that validates the widget type as Input
//   - Extracts and validates the text data from the event
//   - Calls the provided handler function with properly typed parameters
//
// Common events for Input widgets include:
//   - "change": Triggered when the input text is modified
//   - "enter": Triggered when the user presses Enter
//   - "focus": Triggered when the input gains focus
//   - "blur": Triggered when the input loses focus
//
// Parameters:
//   - container: The container widget to search for the target widget
//   - id: The unique identifier of the Input widget to handle events for
//   - event: The name of the event to listen for (e.g., "change", "enter")
//   - fn: The handler function that will be called when the event occurs.
//     It receives the Input widget, event name, and text data as parameters.
//     Should return true if the event was handled, false otherwise.
//
// Example usage:
//
//	HandleInputEvent(container, "username", "change", func(input *Input, event string, text string) bool {
//		fmt.Printf("Username changed to: %s\n", text)
//		return true
//	})
//
//	HandleInputEvent(container, "password", "enter", func(input *Input, event string, text string) bool {
//		// Handle form submission when Enter is pressed
//		submitForm(text)
//		return true
//	})
func HandleInputEvent(container Container, id string, event string, fn func(*Input, string, string) bool) {
	widget := container.Find(id, false)
	if widget == nil {
		container.Log(container, "error", "Widget %s not found", id)
		return
	}
	widget.On(event, func(widget Widget, event string, data ...any) bool {
		input, ok := widget.(*Input)
		if !ok {
			widget.Log(widget, "error", "Error converting %s (%T) to Input", widget.ID(), widget)
			return false
		}
		if len(data) != 1 {
			widget.Log(widget, "error", "Error handling input event, wrong parameter count %d", len(data))
			return false
		}
		text, ok := data[0].(string)
		if !ok {
			widget.Log(widget, "error", "Error converting parameter to string, is %T", data[0])
			return false
		}
		return fn(input, event, text)
	})
}

// HandleKeyEvent registers a keyboard event handler for any widget within a container.
// This function provides a convenient way to handle raw keyboard input events,
// giving access to the full tcell.EventKey structure for detailed key processing.
//
// The function performs the following operations:
//   - Locates the widget by ID within the container hierarchy
//   - Registers a "key" event handler with proper type validation
//   - Extracts and validates the tcell.EventKey data from the event
//   - Calls the provided handler function with the widget and key event
//
// This is useful for handling:
//   - Custom keyboard shortcuts and hotkeys
//   - Special key combinations (Ctrl+C, Alt+F4, etc.)
//   - Raw key processing that requires access to modifiers
//   - Global key handlers that work across different widget types
//
// Parameters:
//   - container: The container widget to search for the target widget
//   - id: The unique identifier of the widget to handle key events for
//   - fn: The handler function that will be called when a key event occurs.
//     It receives the widget and the tcell.EventKey event as parameters.
//     Should return true if the event was handled, false to allow propagation.
//
// Example usage:
//
//	HandleKeyEvent(container, "mainwindow", func(widget Widget, key *tcell.EventKey) bool {
//		switch key.Key() {
//		case tcell.KeyCtrlC:
//			// Handle Ctrl+C
//			return true
//		case tcell.KeyF1:
//			// Show help
//			showHelp()
//			return true
//		}
//		return false // Let other handlers process the key
//	})
func HandleKeyEvent(container Container, id string, fn func(Widget, *tcell.EventKey) bool) {
	widget := container.Find(id, false)
	if widget == nil {
		container.Log(container, "error", "Widget %s not found", id)
		return
	}
	widget.On("key", func(widget Widget, event string, data ...any) bool {
		if event != "key" {
			widget.Log(widget, "error", "Event is no key event")
			return false
		}
		if len(data) != 1 {
			widget.Log(widget, "error", "Error handling key event, wrong parameter count %d", len(data))
			return false
		}
		key, ok := data[0].(*tcell.EventKey)
		if !ok {
			widget.Log(widget, "error", "Error converting parameter to key event, is %T", data[0])
			return false
		}
		return fn(widget, key)
	})
}

// HandleListEvent registers an event handler for List widget events within a container.
// This function simplifies the process of handling list-specific events by automatically
// finding the widget, performing type assertions, and validating event data.
//
// The function performs the following operations:
//   - Locates the widget by ID within the container hierarchy
//   - Registers an event handler that validates the widget type as List
//   - Extracts and validates the index data from the event
//   - Calls the provided handler function with properly typed parameters
//
// Common events for List widgets include:
//   - "select": Triggered when an item is highlighted/focused (navigation)
//   - "activate": Triggered when an item is activated (Enter key or double-click)
//   - "toggle": Triggered when an item's selection state is toggled (Space key)
//   - "change": Triggered when the selection state changes
//
// Parameters:
//   - container: The container widget to search for the target widget
//   - id: The unique identifier of the List widget to handle events for
//   - event: The name of the event to listen for (e.g., "select", "activate")
//   - fn: The handler function that will be called when the event occurs.
//     It receives the List widget, event name, and item index as parameters.
//     Should return true if the event was handled, false otherwise.
//
// Example usage:
//
//	HandleListEvent(container, "menu", "select", func(list *List, event string, index int) bool {
//		fmt.Printf("Item %d selected: %s\n", index, list.Items[index])
//		return true
//	})
//
//	HandleListEvent(container, "files", "activate", func(list *List, event string, index int) bool {
//		// Open the selected file
//		openFile(list.Items[index])
//		return true
//	})
func HandleListEvent(container Container, id, event string, fn func(*List, string, int) bool) {
	widget := container.Find(id, false)
	if widget == nil {
		container.Log(container, "error", "Widget %s not found", id)
		return
	}
	widget.On(event, func(widget Widget, event string, data ...any) bool {
		list, ok := widget.(*List)
		if !ok {
			widget.Log(widget, "error", "Error converting %s (%T) to List", widget.ID(), widget)
			return false
		}
		if len(data) != 1 {
			widget.Log(widget, "error", "Error handling list event, wrong parameter count %d", len(data))
			return false
		}
		index, ok := data[0].(int)
		if !ok {
			widget.Log(widget, "error", "Error converting parameter to int, is %T", data[0])
			return false
		}
		return fn(list, event, index)
	})
}

// Update provides a convenient way to update widget content based on the widget type.
// This function automatically determines the widget type and applies the appropriate
// update method, making it easier to update widgets dynamically without explicit
// type checking in application code.
//
// The function performs the following operations:
//   - Locates the widget by ID within the container hierarchy
//   - Determines the widget type using type assertion
//   - Applies the appropriate update based on widget type and value type
//   - Triggers necessary events and refreshes as needed
//
// Supported widget types and their expected value types:
//   - Label: Accepts any type, converts to string using fmt.Sprintf
//   - List: Accepts []string, updates items and resets selection to first item
//   - ProgressBar: Accepts int, updates the progress value
//   - Text: Accepts []string, updates the text content lines
//
// Parameters:
//   - container: The container widget to search for the target widget
//   - id: The unique identifier of the widget to update
//   - value: The new value to set. Type should match the widget's expected type.
//
// Example usage:
//
//	// Update a label with a string
//	Update(container, "status", "Connected")
//
//	// Update a label with a number (converted to string)
//	Update(container, "counter", 42)
//
//	// Update a list with new items
//	Update(container, "menu", []string{"File", "Edit", "View", "Help"})
//
//	// Update a progress bar
//	Update(container, "progress", 75)
//
//	// Update a text widget with multiple lines
//	Update(container, "log", []string{"Line 1", "Line 2", "Line 3"})
//
// Note: If the widget is not found or the value type doesn't match the widget's
// expected type, the update operation will be silently ignored.
func Update(container Container, id string, value any) {
	widget := container.Find(id, false)
	switch widget := widget.(type) {
	case *Label:
		widget.Text = fmt.Sprintf("%v", value)
	case *List:
		a, ok := value.([]string)
		if ok {
			widget.Items = a
			widget.Index = 0
			if len(a) > 0 {
				widget.Emit("select", 0)
			}
		}
	case *ProgressBar:
		v, ok := value.(int)
		if ok {
			widget.Value = v
		}
	case *Switcher:
		pane, ok := value.(string)
		if ok {
			widget.Select(pane)
		}
	case *Text:
		t, ok := value.([]string)
		if ok {
			widget.Set(t)
		}
	}
}

// WidgetType returns a clean, human-readable string representation of the widget's type.
// This function extracts the type name from the widget's Go type, removing the
// package prefix and pointer notation to provide a simple type identifier.
//
// The function performs the following operations:
//   - Gets the full Go type string using fmt.Sprintf("%T", widget)
//   - Removes the "*tui." prefix to get just the widget type name
//   - Returns the clean type name (e.g., "Label", "Input", "List")
//
// This is useful for:
//   - Debugging and logging widget information
//   - Dynamic widget type checking in generic code
//   - User interface introspection and development tools
//   - Error messages and diagnostic output
//
// Parameters:
//   - widget: The widget to get the type name for
//
// Returns:
//   - string: The clean widget type name without package prefix
//
// Example usage:
//
//	widget := NewLabel("test", "Hello")
//	fmt.Println(WidgetType(widget)) // Output: "Label"
//
//	input := NewInput("username")
//	fmt.Println(WidgetType(input))  // Output: "Input"
func WidgetType(widget Widget) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", widget), "*tui.")
}

// WidgetDetails returns a comprehensive, formatted string containing detailed
// information about a widget's current state, properties, and layout information.
// This function is primarily intended for debugging, development, and diagnostic purposes.
//
// The function provides the following information:
//   - Widget type (clean type name without package prefix)
//   - Widget ID (unique identifier)
//   - Parent widget ID (or "<nil>" if no parent)
//   - Bounds: outer position and dimensions (x, y, width, height)
//   - Content: inner content area position and dimensions
//   - State: current widget state as reported by the widget
//   - Flags: boolean properties (focusable, focussed, hovered)
//
// The output format is designed to be human-readable and suitable for:
//   - Debug logging and console output
//   - Development tools and widget inspectors
//   - Troubleshooting layout and positioning issues
//   - Understanding widget hierarchy and relationships
//
// Parameters:
//   - widget: The widget to generate detailed information for
//
// Returns:
//   - string: Multi-line formatted string with comprehensive widget details
//
// Example output:
//
//	Label
//	ID        : 'status'
//	Parent-ID : 'main-container'
//	Bounds    : x=10, y=5, w=200, h=30
//	Content   : x=12, y=7, w=196, h=26
//	State     : normal
//	Flags     : focusable, focussed
//
// Example usage:
//
//	widget := container.Find("username")
//	fmt.Println(WidgetDetails(widget))
//
//	// For logging during development
//	log.Printf("Widget details:\n%s", WidgetDetails(problematicWidget))
func WidgetDetails(widget Widget) string {
	result := fmt.Sprintf("%s", WidgetType(widget))
	result += fmt.Sprintf("\nID        : '%s'", widget.ID())
	parent := "<nil>"
	if widget.Parent() != nil {
		parent = "'" + widget.Parent().ID() + "'"
	}
	result += fmt.Sprintf("\nParent-ID : %s", parent)
	x, y, w, h := widget.Bounds()
	result += fmt.Sprintf("\nBounds    : x=%d, y=%d, w=%d, h=%d", x, y, w, h)
	x, y, w, h = widget.Content()
	result += fmt.Sprintf("\nContent   : x=%d, y=%d, w=%d, h=%d", x, y, w, h)
	result += fmt.Sprintf("\nState     : %s", widget.State())

	flags := make([]string, 0)
	if widget.Focusable() {
		flags = append(flags, "focusable")
	}
	if widget.Focused() {
		flags = append(flags, "focussed")
	}
	if widget.Hovered() {
		flags = append(flags, "hovered")
	}
	result += fmt.Sprintf("\nFlags     : %s", strings.Join(flags, ", "))

	return result
}

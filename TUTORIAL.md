# TUI Package Tutorial

A comprehensive guide to building terminal user interfaces with the `pkg/tui` package.

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Core Concepts](#core-concepts)
4. [Building Your First UI](#building-your-first-ui)
5. [Layout Containers](#layout-containers)
6. [Widgets](#widgets)
7. [Styling and Themes](#styling-and-themes)
8. [Event Handling](#event-handling)
9. [Advanced Features](#advanced-features)
10. [Complete Example](#complete-example)

## Introduction

The `pkg/tui` package provides a powerful framework for creating terminal user interfaces in Go. It offers a fluent builder API, comprehensive widget library, flexible layout system, and rich styling capabilities using Unicode box-drawing characters.

### Key Features

- **Fluent Builder API**: Chain method calls to construct complex UIs
- **Rich Widget Library**: Buttons, inputs, lists, progress bars, and more
- **Flexible Layouts**: Grid, flex, and container-based positioning
- **Comprehensive Styling**: Borders, colors, padding, margins, and themes
- **Event System**: Keyboard and mouse interaction support
- **Focus Management**: Tab navigation and widget focus handling

## Getting Started

### Basic Import

```go
import tui "github.com/tekugo/zeichenwerk"
```

### Minimal Application

```go
package main

import "github.com/tekugo/zeichenwerk"

func main() {
    ui := tui.NewBuilder().
        Label("hello", "Hello, World!", 0).
        Build()
    
    ui.Run()
}
```

## Core Concepts

### Builder Pattern

The TUI package uses a fluent builder pattern that allows you to chain method calls to construct your interface:

```go
ui := tui.NewBuilder().
    Flex("main", "vertical", "stretch", 0).
        Label("title", "My App", 0).
        Button("quit", "Quit").
    End().
    Build()
```

### Widget Hierarchy

Widgets are organized in a tree structure:

- **Containers**: Can hold child widgets (Flex, Grid, Box)
- **Widgets**: Individual UI components (Label, Button, Input, etc.)
- **Root**: The UI itself acts as the root container

### Layout System

The package provides several layout containers:

- **Flex**: Arranges widgets in rows or columns
- **Grid**: Positions widgets in a grid with cells
- **Box**: Wraps a single widget with optional title and border

## Building Your First UI

Let's build a simple application with a header, content area, and footer:

```go
func main() {
    ui := tui.NewBuilder().
        // Main vertical layout
        Flex("main", "vertical", "stretch", 0).
            // Header
            Flex("header", "horizontal", "start", 0).Hint(0, 1).
                Label("title", "My Application", 20).
                Label("time", "12:00", 6).
            End().
            
            // Content area
            Label("content", "Welcome to my app!", 0).Hint(0, -1).
            
            // Footer
            Label("footer", "Press Ctrl+C to quit", 0).Hint(0, 1).
        End().
        Build()
    
    ui.Run()
}
```

### Key Methods Explained

- `Flex(id, direction, alignment, grow)`: Creates a flex container
  - `direction`: "horizontal" or "vertical"
  - `alignment`: "start", "center", "end", "stretch"
  - `grow`: How much the container should expand (0 = fixed, >0 = flexible)

- `Label(id, text, width)`: Creates a text label
  - `width`: 0 = auto-size, >0 = fixed width

- `Hint(width, height)`: Sets preferred size (-1 = fill available space)

- `End()`: Closes the current container and returns to parent

## Layout Containers

### Flex Layout

Flex containers arrange widgets in a single direction (horizontal or vertical):

```go
// Horizontal layout
Flex("toolbar", "horizontal", "start", 0).
    Button("new", "New").
    Button("open", "Open").
    Button("save", "Save").
End()

// Vertical layout with spacing
Flex("sidebar", "vertical", "stretch", 0).
    Label("title", "Navigation", 0).
    List("menu", menuItems).Hint(0, -1).
    Button("settings", "Settings").
End()
```

### Grid Layout

Grid containers position widgets in a 2D grid:

```go
Grid("layout", 3, 3, false). // 3x3 grid, no uniform sizing
    Cell(0, 0, 1, 3).Label("header", "Header", 0).     // Row 0, spans 3 columns
    Cell(1, 0, 2, 1).List("sidebar", items).           // Rows 1-2, column 0
    Cell(1, 1, 1, 2).Text("content", [], true, 100).   // Row 1, columns 1-2
    Cell(2, 1, 1, 2).Label("status", "Ready", 0).      // Row 2, columns 1-2
End()
```

### Box Container

Box containers wrap a single widget with optional title and border:

```go
Box("settings", "Settings").Border("", "thin").
    Flex("form", "vertical", "start", 1).
        Input("name", "Enter name", 30).
        Input("email", "Enter email", 30).
        Button("submit", "Submit").
    End().
End()
```

## Widgets

### Labels

Display static text:

```go
Label("title", "Application Title", 0).
Label("fixed", "Fixed Width", 20).
Label("status", "Status: Ready", 0).Hint(-1, 1) // Fill width
```

### Buttons

Interactive clickable elements:

```go
Button("ok", "OK").
Button("cancel", "Cancel").
Button("submit", " Submit Form ") // Padding with spaces
```

### Input Fields

Text input widgets:

```go
Input("username", "Enter username", 30).
Input("password", "", 30).       // Empty placeholder
Input("search", "Search...", 40)
```

### Lists

Scrollable item lists:

```go
items := []string{"Item 1", "Item 2", "Item 3"}
List("menu", items)
```

### Progress Bars

Visual progress indicators:

```go
ProgressBar("download", 50, 0, 100).Hint(30, 1)  // 50% progress
ProgressBar("upload", 0, 0, 100).Hint(30, 1)     // 0% progress
ProgressBar("complete", 100, 0, 100).Hint(30, 1) // 100% progress
```

### Text Areas

Multi-line text display with scrolling:

```go
Text("log", []string{}, true, 1000) // Scrollable, max 1000 lines
```

## Styling and Themes

### Basic Styling

Apply styling to widgets using method chaining:

```go
Label("title", "Styled Label", 0).
    Background("", "$blue").        // Background color
    Foreground("", "$white").       // Text color
    Border("", "thick").            // Border style
    Padding(1, 2).                  // Vertical, horizontal padding
    Margin(1)                       // All-around margin
```

### Border Styles

Available border styles:

- `"thin"`: Standard single-line borders (most common)
- `"thick"`: Bold single-line borders
- `"double"`: Double-line borders
- `"round"`: Rounded corners
- `"thick-thin"`: Mixed weight borders
- `"lines"`: Minimal horizontal lines only

```go
Box("dialog", "Dialog").Border("", "double").
    Label("message", "Important message", 0).
End()
```

### CSS-like Classes

Apply styling using CSS-like classes:

```go
ui := NewBuilder().
    Class("header").                    // Apply header class
    Label("title", "My App", 0).
    Class("").                          // Reset to default class
    Label("content", "Content", 0).
    Build()
```

### Color Variables

Use theme color variables:

- `$bg`: Background color
- `$fg`: Foreground color
- `$blue`, `$red`, `$green`, `$yellow`: Theme colors
- `$aqua`, `$purple`, `$orange`: Additional colors
- `$comments`: Muted text color

## Event Handling

### Widget Events

Attach event handlers to widgets:

```go
ui.FindOn("submit-btn", func(action string, widget tui.Widget) {
    if action == "click" {
        // Handle button click
        fmt.Println("Button clicked!")
    }
})

ui.FindOn("input-field", func(action string, widget tui.Widget) {
    if action == "change" {
        input := widget.(*tui.Input)
        fmt.Printf("Input changed: %s\n", input.Text)
    }
})
```

### List Events

Handle list selection changes:

```go
ui.FindOn("menu-list", func(action string, widget tui.Widget) {
    list, ok := widget.(*tui.List)
    if !ok {
        return
    }
    
    switch action {
    case "select":
        selected := list.Selected()
        fmt.Printf("Selected: %s\n", selected)
    case "change":
        // Handle selection change
    }
})
```

## Advanced Features

### Popups and Dialogs

Create modal dialogs:

```go
// Create main UI
ui := NewBuilder().
    Label("main", "Main Content", 0).
    Build()

// Create popup dialog
popup := NewBuilder().
    Class("popup").
    Flex("dialog", "vertical", "stretch", 0).
        Label("title", "Dialog Title", 0).Background("", "$blue").
        Label("message", "Dialog message", 0).Padding(1).
        Flex("buttons", "horizontal", "end", 0).
            Button("ok", "OK").
            Button("cancel", "Cancel").
        End().
    End().
    Container()

popup.SetParent(ui)

// Show popup at specific position
pw, ph := popup.Hint()
ui.Popup(-2, 2, pw, ph, popup)
```

### Logging and Debugging

Use the built-in logging system:

```go
// Enable debug mode
ui.SetDebug(true)

// Log messages from widgets
widget.Log("Debug message: %s", value)

// Access log widget
logWidget := ui.Find("debug-log").(*tui.Text)
logWidget.Append("Custom log message")
```

### Dynamic Content Updates

Update widget content at runtime:

```go
// Update label text
label := ui.Find("status").(*tui.Label)
label.Text = "Updated status"

// Update progress bar
progress := ui.Find("progress").(*tui.ProgressBar)
progress.SetValue(75)

// Add items to list
list := ui.Find("items").(*tui.List)
list.Add("New item")
```

## Complete Example

Here's the complete demo application that showcases most features:

```go
package main

import (
    "maps"
    "slices"
    "github.com/thomas-rustemeyer/fuego/pkg/tui"
)

func main() {
    // Sample data
    countries := slices.Collect(maps.Values(map[string]string{
        "US": "United States",
        "CA": "Canada", 
        "UK": "United Kingdom",
        "DE": "Germany",
        "FR": "France",
    }))
    
    names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve"}

    // Build main UI
    ui := tui.NewBuilder().
        // Main vertical layout
        Flex("main", "vertical", "stretch", 0).
            // Header
            Flex("header", "horizontal", "start", 0).Hint(0, 1).
                Class("header").
                Label("title", "TUI Demo Application", 20).
                Label("spacer", "", 0).Hint(-1, 1).
                Label("time", "12:00", 6).
            End().
            Class("").
            
            // Main content grid
            Grid("content", 3, 3, false).Hint(0, -1).
                // Debug log (left column, full height)
                Cell(0, 0, 1, 3).Text("debug-log", []string{}, true, 1000).
                
                // Form section
                Cell(1, 0, 2, 1).
                Box("form-box", "Controls").Border("", "thin").Margin(1).
                    Flex("form", "vertical", "start", 1).Padding(1).
                        ProgressBar("progress1", 25, 0, 100).Hint(30, 1).
                        ProgressBar("progress2", 50, 0, 100).Hint(30, 1).
                        ProgressBar("progress3", 75, 0, 100).Hint(30, 1).
                        Input("input", "Enter text here", 30).
                    End().
                End().
                
                // Countries list
                Cell(1, 1, 1, 2).List("countries", countries).
                
                // Names list
                Cell(2, 1, 1, 1).List("names", names).
                
                // Buttons
                Cell(2, 2, 1, 1).
                Box("button-box", "Actions").Border("", "thin").
                    Flex("buttons", "vertical", "center", 1).Padding(1).
                        Button("action1", " Action 1 ").Hint(15, 1).
                        Button("action2", " Action 2 ").Hint(15, 1).
                        Button("quit", " Quit ").Hint(15, 1).
                    End().
                End().
            End().
            
            // Footer
            Flex("footer", "horizontal", "start", 0).Hint(0, 1).
                Class("footer").
                Label("status", "Ready", 0).Hint(-1, 1).
                Label("help", "Press Tab to navigate, Enter to select", 0).
            End().
        End().
        Build()

    // Create a popup dialog
    popup := tui.NewBuilder().
        Class("popup").
        Flex("dialog", "vertical", "stretch", 0).
            Label("dialog-title", "Welcome!", 0).
                Background("", "$blue").Foreground("", "$bg").Padding(1, 2).
            Flex("dialog-content", "vertical", "stretch", 0).Hint(0, -1).Padding(1, 2).
                Label("message", "Welcome to the TUI Demo!", 0).Padding(0, 0, 1, 0).
                Label("instructions", "Use Tab/Shift+Tab to navigate", 0).
                Input("dialog-input", "Enter your name", 30).
            End().
            Flex("dialog-buttons", "horizontal", "end", 0).Padding(1, 2).
                Button("dialog-ok", "OK").
                Button("dialog-cancel", "Cancel").
            End().
        End().
        Container()
    
    popup.SetParent(ui)
    
    // Position popup in center
    pw, ph := popup.Hint()
    ui.Popup(-2, 2, pw, ph, popup)

    // Add event handlers
    ui.FindOn("countries", func(action string, widget tui.Widget) {
        list, ok := widget.(*tui.List)
        if !ok {
            return
        }
        if action == "select" {
            list.Log("Selected country: %s", list.Selected())
        }
    })

    ui.FindOn("quit", func(action string, widget tui.Widget) {
        if action == "click" {
            ui.Quit()
        }
    })

    ui.FindOn("dialog-ok", func(action string, widget tui.Widget) {
        if action == "click" {
            input := ui.Find("dialog-input").(*tui.Input)
            ui.Log("Hello, %s!", input.Text)
            ui.ClosePopup()
        }
    })

    ui.FindOn("dialog-cancel", func(action string, widget tui.Widget) {
        if action == "click" {
            ui.ClosePopup()
        }
    })

    // Run the application
    ui.Run()
}
```

## Best Practices

### 1. Structure Your Code

Organize complex UIs into functions:

```go
func createHeader() *tui.Builder {
    return tui.NewBuilder().
        Flex("header", "horizontal", "start", 0).
            Label("title", "My App", 0).
            Label("status", "Ready", 0).
        End()
}

func createMainContent() *tui.Builder {
    return tui.NewBuilder().
        Grid("content", 2, 2, false).
            Cell(0, 0, 1, 1).Label("section1", "Section 1", 0).
            Cell(0, 1, 1, 1).Label("section2", "Section 2", 0).
        End()
}
```

### 2. Use Meaningful IDs

Give widgets descriptive IDs for easy reference:

```go
Button("save-button", "Save").
Input("username-input", "Username", 20).
List("file-list", files)
```

### 3. Handle Events Gracefully

Always check widget types in event handlers:

```go
ui.FindOn("my-widget", func(action string, widget tui.Widget) {
    button, ok := widget.(*tui.Button)
    if !ok {
        return
    }
    // Safe to use button methods
})
```

### 4. Use Appropriate Layout Containers

- **Flex**: For simple linear layouts
- **Grid**: For complex 2D layouts
- **Box**: For single widgets with borders/titles

### 5. Leverage Styling

Use consistent styling throughout your application:

```go
// Define common styles
ui := tui.NewBuilder().
    Class("primary-button").
    Button("save", "Save").
    Class("secondary-button").
    Button("cancel", "Cancel").
    Class("").
    Build()
```

## Troubleshooting

### Common Issues

1. **Widgets not showing**: Check that containers are properly closed with `End()`
2. **Layout issues**: Verify grid cell coordinates and spans
3. **Event handlers not firing**: Ensure widget IDs match exactly
4. **Styling not applied**: Check class names and theme variables

### Debug Mode

Enable debug mode to see widget hierarchy and events:

```go
ui.SetDebug(true)
```

This will show:

- Widget tree structure
- Event firing information
- Layout calculations
- Focus and hover states

## Conclusion

The `pkg/tui` package provides a powerful and flexible framework for building terminal user interfaces. With its fluent builder API, comprehensive widget library, and rich styling system, you can create professional-looking terminal applications with ease.

Start with simple layouts and gradually add complexity as you become more familiar with the API. The demo application serves as an excellent reference for implementing common UI patterns and features.

For more advanced usage, explore the source code and experiment with different combinations of widgets, layouts, and styling options.


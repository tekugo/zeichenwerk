# Zeichenwerk Tutorial

Welcome to zeichenwerk! This tutorial will guide you through creating terminal user interfaces using this modern Go TUI framework.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Concepts](#basic-concepts)
3. [Your First Application](#your-first-application)
4. [Working with Widgets](#working-with-widgets)
5. [Layout Containers](#layout-containers)
6. [Event Handling](#event-handling)
7. [Theming and Styling](#theming-and-styling)
8. [Advanced Features](#advanced-features)

## Getting Started

### Installation

Add zeichenwerk to your Go project:

```bash
go mod init your-app
go get github.com/tekugo/zeichenwerk
```

### Import Convention

It's recommended to use a short alias when importing zeichenwerk:

```go
import (
    . "github.com/tekugo/zeichenwerk"  // Use dot import for cleaner code
    // or
    tui "github.com/tekugo/zeichenwerk"  // Use 'tui' alias
)
```

## Basic Concepts

### The Builder Pattern

Zeichenwerk uses a fluent builder pattern for constructing UIs. This makes the code readable and allows for easy nesting of components:

```go
ui := NewBuilder(TokyoNightTheme()).
    Flex("main", "vertical", "stretch", 0).
    Label("title", "Hello World", 0).
    Button("ok", "OK").
    End().
    Build()
```

### Widget Hierarchy

All UI components implement the `Widget` interface. Containers extend this with the ability to hold child widgets:

- **Widget**: Base interface for all UI components
- **Container**: Widgets that can contain other widgets
- **UI**: Top-level container that manages the entire application

## Your First Application

Let's create a simple "Hello World" application:

```go
package main

import (
    . "github.com/tekugo/zeichenwerk"
)

func main() {
    ui := NewBuilder(TokyoNightTheme()).
        Flex("main", "vertical", "center", 2).
        Label("title", "Hello, Zeichenwerk!", 0).
        Button("quit", "Quit").
        End().
        Build()
    
    // Handle the quit button
    if button := ui.Find("quit", false); button != nil {
        button.On("click", func(w Widget, event string, data ...any) bool {
            ui.Stop()
            return true
        })
    }
    
    ui.Run()
}
```

This creates a simple vertical layout with a centered label and button.

## Working with Widgets

### Labels

Labels display static text and support various styling options:

```go
builder.Label("simple", "Simple text", 0).
Label("styled", "Styled text", 0).
    Background("", "$blue").
    Foreground("", "$white").
    Padding(1, 2)
```

### Buttons

Buttons are interactive elements that respond to clicks:

```go
builder.Button("save", "Save").
Button("cancel", "Cancel")

// Add click handlers
saveBtn.On("click", func(w Widget, event string, data ...any) bool {
    // Handle save action
    return true
})
```

### Input Fields

Input widgets allow user text entry:

```go
builder.Input("username", "", 20).  // ID, default text, width
Input("password", "", 20)

// Access input value
if input := ui.Find("username", false).(*Input); input != nil {
    username := input.Text
}
```

### Lists

Lists display selectable items with keyboard navigation:

```go
items := []string{"Item 1", "Item 2", "Item 3"}
builder.List("mylist", items)

// Handle selection
HandleListEvent(container, "mylist", "activate", func(list *List, event string, index int) bool {
    selectedItem := items[index]
    // Handle selection
    return true
})
```

### Progress Bars

Progress bars show completion status:

```go
builder.ProgressBar("progress", 50, 0, 100)  // ID, current, min, max

// Update progress
if progress := ui.Find("progress", false).(*ProgressBar); progress != nil {
    progress.Current = 75
    progress.Refresh()
}
```

### Tabs

Tabs organize content into multiple views:

```go
builder.Tabs("tabs", "First", "Second", "Third")

// Handle tab changes
HandleEvent(container, "tabs", "change", func(w Widget, event string, data ...any) bool {
    if len(data) > 0 {
        if index, ok := data[0].(int); ok {
            // Handle tab change
        }
    }
    return true
})
```

## Layout Containers

### Flex Layouts

Flex containers arrange widgets in horizontal or vertical lines:

```go
// Horizontal layout
builder.Flex("header", "horizontal", "stretch", 0).
    Label("title", "App Title", 20).
    Spacer().  // Takes remaining space
    Label("time", "12:00", 6).
    End()

// Vertical layout with spacing
builder.Flex("sidebar", "vertical", "start", 2).  // 2 units spacing
    Button("home", "Home").
    Button("settings", "Settings").
    Button("about", "About").
    End()
```

#### Flex Alignment Options:
- **"start"**: Align items to the beginning
- **"center"**: Center items
- **"end"**: Align items to the end
- **"stretch"**: Stretch items to fill container

### Grid Layouts

Grid containers provide precise positioning with rows and columns:

```go
builder.Grid("form", 3, 2, false).  // columns, rows, auto-sizing
    Cell(0, 0, 1, 1).Label("", "Name:", 0).
    Cell(1, 0, 1, 1).Input("name", "", 20).
    Cell(0, 1, 1, 1).Label("", "Email:", 0).
    Cell(1, 1, 1, 1).Input("email", "", 20).
    Cell(2, 0, 1, 2).Button("submit", "Submit").  // Spans 2 rows
    End()

// Configure grid sizing
if grid := builder.Container().(*Grid); ok {
    grid.Columns(10, -1, 8)  // Fixed, flexible, fixed
    grid.Rows(1, 1, 1)       // Equal height rows
}
```

### Box Containers

Box containers wrap a single widget with optional borders and titles:

```go
builder.Box("info", "Information").
    Border("", "round").
    Padding(1).
    Label("content", "This is inside a box", 0).
    End()
```

### Switcher Containers

Switchers display one child widget at a time:

```go
builder.Switcher("views").
    With(homeView).    // Function that builds home view
    With(settingsView). // Function that builds settings view
    End()

// Switch views
Update(ui, "views", "home-view")  // Switch to specific child by ID
```

## Event Handling

### Basic Event Handling

Widgets emit events that you can handle:

```go
widget.On("click", func(w Widget, event string, data ...any) bool {
    // Handle the event
    return true  // Return true to consume the event
})
```

### Common Events:
- **"click"**: Button clicks, widget activation
- **"change"**: Value changes in inputs
- **"focus"**: Widget gains focus
- **"blur"**: Widget loses focus
- **"select"**: Item selection in lists
- **"activate"**: Item activation (Enter key, double-click)

### Helper Functions

Zeichenwerk provides helper functions for common event scenarios:

```go
// Input field events
HandleInputEvent(container, "username", "change", func(input *Input, event string) bool {
    username := input.Text
    // Validate input
    return true
})

// List events
HandleListEvent(container, "menu", "activate", func(list *List, event string, index int) bool {
    selectedItem := list.Items[index]
    // Handle selection
    return true
})

// Generic widget updates
Update(ui, "status", "Connected")  // Update widget content
```

### Keyboard Shortcuts

Handle global keyboard events:

```go
ui.On("key", func(w Widget, event string, data ...any) bool {
    if len(data) > 0 {
        if keyEvent, ok := data[0].(*tcell.EventKey); ok {
            switch keyEvent.Key() {
            case tcell.KeyCtrlQ:
                ui.Stop()
                return true
            case tcell.KeyF1:
                // Show help
                return true
            }
        }
    }
    return false
})
```

## Theming and Styling

### Built-in Themes

Zeichenwerk includes several built-in themes:

```go
// Available themes
ui := NewBuilder(DefaultTheme()).Build()
ui := NewBuilder(TokyoNightTheme()).Build()
ui := NewBuilder(MidnightNeonTheme()).Build()
```

### Custom Styling

Apply styles to individual widgets:

```go
builder.Label("error", "Error message", 0).
    Foreground("", "$red").
    Background("", "$bg0").
    Border("", "thick").
    Padding(1, 2)
```

### CSS-like Classes

Use classes for consistent styling:

```go
// Apply class
builder.Class("header").
    Label("title", "Application", 0).
    Class("")  // Reset to default

// Define class styles in theme
theme.Set(".header", NewStyle("$cyan", "$bg1").SetPadding(0, 1))
```

### Color Variables

Themes define color variables you can use:

```go
// Tokyo Night theme colors
"$bg0"     // Background
"$fg0"     // Foreground
"$blue"    // Accent blue
"$red"     // Error red
"$green"   // Success green
// ... and more
```

### Border Styles

Various border styles are available:

```go
.Border("", "thin")     // Thin lines
.Border("", "thick")    // Thick lines
.Border("", "double")   // Double lines
.Border("", "round")    // Rounded corners
.Border("", "dashed")   // Dashed lines
```

## Advanced Features

### Inspector

Zeichenwerk includes a built-in inspector for debugging:

```go
// Show inspector
inspector := NewInspector(ui)
ui.Popup(-1, -1, 0, 0, inspector.UI())
```

The inspector shows:
- Widget hierarchy
- Widget properties
- Style information
- Event debugging

### Custom Widgets

Create custom widgets by implementing the Widget interface:

```go
type CustomWidget struct {
    *BaseWidget
    // Custom fields
}

func NewCustomWidget(id string) *CustomWidget {
    return &CustomWidget{
        BaseWidget: NewBaseWidget(id),
    }
}

func (c *CustomWidget) Render(screen Screen) {
    // Custom rendering logic
    width, height := c.Size()
    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            screen.SetContent(x, y, '*', nil, tcell.StyleDefault)
        }
    }
}
```

### Popups and Modals

Create popup dialogs:

```go
popup := NewBuilder(theme).
    Class("popup").
    Flex("dialog", "vertical", "stretch", 0).
    Label("title", "Confirm Action", 0).
    Label("message", "Are you sure?", 0).
    Flex("buttons", "horizontal", "end", 1).
        Button("yes", "Yes").
        Button("no", "No").
        End().
    End().
    Container()

// Show popup centered
ui.Popup(-1, -1, 40, 10, popup)
```

### Dynamic Updates

Update widget content at runtime:

```go
// Update text
Update(ui, "status", "Loading...")

// Update progress
if progress := ui.Find("progress", false).(*ProgressBar); progress != nil {
    progress.Current = newValue
    progress.Refresh()
}

// Update list items
if list := ui.Find("items", false).(*List); list != nil {
    list.Items = newItems
    list.Refresh()
}
```

## Best Practices

### Code Organization

1. **Separate UI construction**: Use functions to build different parts of your UI
2. **Event handler isolation**: Keep event handlers focused and testable
3. **Theme consistency**: Use theme variables instead of hardcoded colors

```go
func buildHeader(builder *Builder) {
    builder.Class("header").
        Flex("header", "horizontal", "stretch", 0).
        Label("title", "My App", 0).
        Spacer().
        Label("status", "Ready", 0).
        End()
}

func buildMainContent(builder *Builder) {
    // Main content construction
}
```

### Performance Tips

1. **Minimize redraws**: Only call `Refresh()` when necessary
2. **Use appropriate containers**: Choose the right layout for your needs
3. **Batch updates**: Group related updates together

### Error Handling

1. **Check widget existence**: Always verify widgets exist before accessing them
2. **Validate input**: Check user input in event handlers
3. **Graceful degradation**: Handle missing resources gracefully

```go
if widget := ui.Find("input", false); widget != nil {
    if input, ok := widget.(*Input); ok {
        // Safe to use input
        text := input.Text
    }
}
```

## Complete Example

Here's a complete example demonstrating many features:

```go
package main

import (
    "fmt"
    . "github.com/tekugo/zeichenwerk"
)

func main() {
    ui := NewBuilder(TokyoNightTheme()).
        Flex("main", "vertical", "stretch", 0).
        With(buildHeader).
        With(buildContent).
        With(buildFooter).
        Build()
    
    setupEventHandlers(ui)
    ui.Run()
}

func buildHeader(builder *Builder) {
    builder.Class("header").
        Flex("header", "horizontal", "stretch", 0).
        Padding(0, 1).
        Label("title", "My Application", 0).
        Spacer().
        Label("time", "12:00", 6).
        End()
}

func buildContent(builder *Builder) {
    builder.Grid("content", 2, 1, false).
        Cell(0, 0, 1, 1).
        Flex("sidebar", "vertical", "start", 1).
            Button("home", "Home").
            Button("settings", "Settings").
            Button("about", "About").
            End().
        Cell(1, 0, 1, 1).
        Switcher("views").
            With(buildHomeView).
            With(buildSettingsView).
            With(buildAboutView).
            End().
        End()
}

func buildFooter(builder *Builder) {
    builder.Class("footer").
        Flex("footer", "horizontal", "start", 0).
        Padding(0, 1).
        Label("help", "F1: Help | Ctrl+Q: Quit", 0).
        End()
}

func buildHomeView(builder *Builder) {
    builder.Flex("home", "vertical", "stretch", 1).
        Label("welcome", "Welcome to the application!", 0).
        ProgressBar("progress", 50, 0, 100).
        End()
}

func buildSettingsView(builder *Builder) {
    builder.Grid("settings", 2, 3, false).
        Cell(0, 0, 1, 1).Label("", "Username:", 0).
        Cell(1, 0, 1, 1).Input("username", "", 20).
        Cell(0, 1, 1, 1).Label("", "Theme:", 0).
        Cell(1, 1, 1, 1).List("themes", []string{"Tokyo Night", "Default", "Midnight Neon"}).
        Cell(0, 2, 2, 1).Button("save", "Save Settings").
        End()
}

func buildAboutView(builder *Builder) {
    builder.Flex("about", "vertical", "center", 1).
        Label("app", "My Application v1.0", 0).
        Label("desc", "Built with zeichenwerk", 0).
        End()
}

func setupEventHandlers(ui *UI) {
    // Navigation buttons
    for _, view := range []string{"home", "settings", "about"} {
        if btn := ui.Find(view, false); btn != nil {
            viewName := view
            btn.On("click", func(w Widget, event string, data ...any) bool {
                Update(ui, "views", viewName)
                return true
            })
        }
    }
    
    // Global keyboard shortcuts
    ui.On("key", func(w Widget, event string, data ...any) bool {
        if len(data) > 0 {
            if keyEvent, ok := data[0].(*tcell.EventKey); ok {
                switch keyEvent.Key() {
                case tcell.KeyCtrlQ:
                    ui.Stop()
                    return true
                }
            }
        }
        return false
    })
}
```

This tutorial covers the main features of zeichenwerk. For more examples, check out the demo application included with the library. Happy coding!
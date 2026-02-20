# Events

## Events by Widget

| Widget | Event | Parameters | Description |
| --- | ---- | ---- | ---- |
| **BaseWidget** | `focus` | None | Widget has gained keyboard focus |
| | `blur` | None | Widget has lost keyboard focus |
| | `hover` | None | Mouse cursor is over the widget |
| | `key` | tcell.EventKey | Raw keyboard event (fallback for unhandled keys) |
| **Button** | `click` | None | Button was clicked or activated |
| **Checkbox** | `change` | bool (checked state) | Checkbox state was toggled |
| **Editor** | `change` | None | Text content was modified |
| | `key` | tcell.EventKey | Raw keyboard event (for unhandled keys) |
| **Input** | `change` | string (text content) | Text content was modified |
| | `enter` | string (text content) | Enter key was pressed |
| **List** | `select` | int (index) | Selection changed to different item |
| | `activate` | int (index) | Item was activated (Enter key or double-click) |
| | `key` | tcell.EventKey | Raw keyboard event (for unhandled keys) |
| **Table** | `activate` | int (row index), []string (row data) | Row was activated (Enter key) |
| | `select` | int (row index), []string (row data) | Row was selected (Space key) |
| **Tabs** | `change` | int (index) | Tab selection changed |
| | `activate` | int (index) | Tab was activated (Enter key or click) |

## Event System Overview

The zeichenwerk event system provides a flexible way to handle user
interactions and widget state changes. All widgets inherit basic event
capabilities from BaseWidget and can emit custom events specific to their
functionality.

### Basic Event Handling

All widgets support the `On()` method to register event handlers:

```go
widget.On("eventName", func(widget Widget, event string, data ...any) bool {
    // Handle the event
    // Return true if the event was handled, false to allow propagation
    return true
})
```

### Event Handler Helpers

For common event patterns, zeichenwerk provides helper functions that simplify
event handling:

#### HandleInputEvent

Simplified event handling for Input widgets:

```go
HandleInputEvent(container, "username", "change", func(input *Input, event string, text string) bool {
    fmt.Printf("Username changed to: %s\n", text)
    return true
})
```

#### HandleListEvent

Simplified event handling for List widgets:

```go
HandleListEvent(container, "menu", "activate", func(list *List, event string, index int) bool {
    fmt.Printf("Menu item %d activated\n", index)
    return true
})
```

#### HandleKeyEvent

Raw keyboard event handling for any widget:

```go
HandleKeyEvent(container, "editor", func(widget Widget, key *tcell.EventKey) bool {
    if key.Key() == tcell.KeyCtrlS {
        // Handle Ctrl+S
        return true
    }
    return false
})
```

## Event Flow

1. **Event Generation**: User interactions (keyboard, mouse) or programmatic actions trigger events
2. **Event Emission**: Widgets call `Emit()` to send events with optional data
3. **Handler Execution**: Registered event handlers are called in order of registration
4. **Event Consumption**: Handlers return `true` to consume the event or `false` to allow propagation

## Common Event Patterns

### Form Validation

```go
// Validate input on change
HandleInputEvent(form, "email", "change", func(input *Input, event string, text string) bool {
    if !isValidEmail(text) {
        input.SetStyle("error", errorStyle)
    } else {
        input.SetStyle("", normalStyle)
    }
    return true
})

// Submit form on enter
HandleInputEvent(form, "email", "enter", func(input *Input, event string, text string) bool {
    submitForm(form)
    return true
})
```

### Navigation

```go
// Handle tab navigation
HandleListEvent(sidebar, "menu", "activate", func(list *List, event string, index int) bool {
    contentArea.ShowPage(menuItems[index].page)
    return true
})

// Handle table row selection
table.On("select", func(widget Widget, event string, data ...any) bool {
    rowIndex := data[0].(int)
    rowData := data[1].([]string)
    showDetails(rowData)
    return true
})
```

### State Management

```go
// Sync checkbox state
checkbox.On("change", func(widget Widget, event string, data ...any) bool {
    checked := data[0].(bool)
    settings.EnableFeature = checked
    updateUI(settings)
    return true
})

// Handle focus changes
widget.On("focus", func(widget Widget, event string, data ...any) bool {
    statusBar.SetText("Editing: " + widget.ID())
    return true
})
```

## Event Data Types

Events can carry different types of data depending on the widget and event type:

- **None**: Events like `focus`, `blur`, `hover`, `click` carry no additional data
- **String**: Text-based events like `change` and `enter` from Input widgets
- **Integer**: Index-based events like `select` and `activate` from List and Tabs
- **Boolean**: State events like `change` from Checkbox widgets
- **Multiple**: Complex events like Table `activate` and `select` carry multiple values
- **tcell.Event**: Raw keyboard events provide access to the underlying tcell event

## Best Practices

1. **Return Values**: Always return `true` if you handle an event, `false` to allow propagation
2. **Type Assertions**: Use type assertions carefully when accessing event data
3. **Error Handling**: Wrap event handlers in error handling to prevent crashes
4. **Performance**: Keep event handlers lightweight to maintain UI responsiveness
5. **Cleanup**: Remove event handlers when widgets are destroyed to prevent memory leaks

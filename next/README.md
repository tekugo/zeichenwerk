# zeichenwerk/next

![Version](https://img.shields.io/badge/version-2.0-blue)
![Go](https://img.shields.io/github/go-mod/go-module/github.com/tekugo/zeichenwerk/next)
![License](https://img.shields.io/github/license/tekugo/zeichenwerk/next)

Zeichenwerk (German for "character works") is a modern, idiomatic Go library for building terminal user interfaces. This refactored version features a fluent builder API, improved architecture, and an enhanced widget system.

## Quick Example

```go
package main

import . "github.com/tekugo/zeichenwerk/next"

func main() {
    NewBuilder(TokyoNightTheme()).
        Flex("main", false, "stretch", 0).
            Flex("header", true, "center", 1).
                Static("title", "My App").
            End().
            Grid("content", 2, 2, false).Columns(20, -1).
                Cell(0, 0, 1, 1).List("menu", "Item 1", "Item 2", "Item 3").
                Cell(1, 0, 1, 1).Button("action", "Click Me").
            End().
        End().
        Run()
}
```

## Installation

```bash
go get github.com/tekugo/zeichenwerk/next
```

## Widgets

### Containers
- **Flex**: Linear layout (horizontal/vertical) with alignment and spacing
- **Grid**: Table-like layout with spanning and grid lines
- **Box**: Single-child container with borders and title
- **Switcher**: Stack of widgets with selection

### Input & Display
- **Button**: Clickable button with keyboard/mouse support
- **Checkbox**: Toggle with checked/unchecked states
- **Input**: Single-line text entry with editing
- **List**: Scrollable list with navigation and selection
- **Table**: Data tables with custom providers
- **Tabs**: Tabbed interface
- **Text**: Multi-line text with scrolling
- **Styled**: Rich text with markup (bold, italic, underline)
- **Static**: Simple labeled text display
- **Spinner**: Animated loading indicator

## Features

### Event System
```go
button.On("click", func(w Widget, event string, data ...any) bool {
    // Handle click
    return true
})
```

### Styling & Themes
```go
theme := NewMapTheme()
theme.Set("button.primary", NewStyle("blue", "white", ""))
theme.Set("button#submit", NewStyle("green", "black", "bold"))
```

Built-in themes:
- `TokyoNightTheme()` - Dark theme with purple/blue accents
- `UnicodeBordersTheme()` - Unicode border styles

### Focus Navigation
- Tab/Shift+Tab: Move focus between widgets
- Arrow keys: Navigate within widgets (lists, tables)
- Enter/Space: Activate buttons, toggle checkboxes

### Mouse Support
- Click to focus widgets
- Hover states with visual feedback
- Drag support for interactive widgets

## Architecture

```
UI (root)
├── Component (embedded)
│   ├── Bounds (x, y, width, height)
│   ├── Styles (CSS-like selectors)
│   ├── Events (handlers, bubbling)
│   └── Parent/Child hierarchy
├── Layers (popups/modals)
├── Event Loop (tcell integration)
├── Renderer (dirty-flag optimizations)
└── Focus Manager
```

## Demo

Explore the demo application for examples of all widgets:

```bash
go run ./cmd/demo
```

## Documentation

- Package docs: [doc.go](doc.go)
- Builder pattern: [builder.go](builder.go)
- Theme system: [theme.go](theme.go)
- Component base: [component.go](component.go)

## Development Status

**Stable** - This is a production-ready refactoring of the original zeichenwerk library. All core features are implemented and tested.

### Known Limitations
- Large text files should use the `Text` widget with a `Scroller` container for performance
- Widgets must have unique IDs within a container hierarchy
- Theme color variables must match terminal palette or use 256/truecolor codes

## License

MIT © Thomas Rustemeyer

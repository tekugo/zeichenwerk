# zeichenwerk

![Version](https://img.shields.io/badge/version-2.0-blue)
![Go](https://img.shields.io/github/go-mod/go-module/github.com/tekugo/zeichenwerk)
![License](https://img.shields.io/github/license/tekugo/zeichenwerk)

Zeichenwerk (German for "character works") is a modern, idiomatic Go library for
building terminal user interfaces. This refactored version features a fluent
builder API, improved architecture, and an enhanced widget system.

## How it looks

[Showcase](showcase-1.png)

## Quick Example

```go
package main

import . "github.com/tekugo/zeichenwerk"

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

## Why zeichenwerk

Zeichenwerk is designed for developers who want:

- A fluent, chainable builder API
- Higher-level widgets than tcell
- More composable layouts than tview
- A traditional retained widget hierarchy rather than an event/message
  architecture

Compare to other Go TUI libraries:

| Library     | Style                             |
| ----------- | --------------------------------- |
| tcell       | Low-level terminal primitives     |
| tview       | Traditional widget toolkit        |
| bubbletea   | Elm-style update loop             |
| zeichenwerk | Fluent builder + widget hierarchy |

## Installation

```bash
go get github.com/tekugo/zeichenwerk
```

## Widgets

| Category    | Widgets                               |
| ----------- | ------------------------------------- |
| Containers  | Box, Flex, Grid, Form, Switcher, Tabs |
| Interaction | Button, Checkbox, Input, Select       |
| Display     | Collapsible, Deck, List, Table        |
| Overlay     | Dialog + Containers                   |
| Animation   | Animation, Grow, Scanner, Spinner     |

Also includes:

- Multi-line text editor/area
- Inspector for introspection and debugging
- Included structured logging
- Canvas, which will be used for a TUI designer
- Forms created from and bound to Go structs
- Styled text display with inline markup and word wrapping

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
theme := NewTheme()
theme.Set("button.primary", NewStyle("blue", "white", ""))
theme.Set("button#submit", NewStyle("green", "black", "bold"))
```

Built-in themes:

- `TokyoNightTheme()` - Dark theme with purple/blue accents
- more to follow

### Focus Navigation

- Tab/Shift+Tab: Move focus between widgets
- Arrow keys: Navigate within widgets (lists, tables)
- Enter/Space: Activate buttons, toggle checkboxes

### Mouse Support

- Click to focus widgets
- Hover states with visual feedback

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

**Active development** - Core features are implemented. Widget interface is
stable. Widget-level test coverage and some layout edge cases are still being
worked on.

## License

MIT © Thomas Rustemeyer

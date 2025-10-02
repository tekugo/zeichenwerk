# Theme System Documentation

## Overview
The `theme.go` file implements a comprehensive CSS-like theming system for TUI widgets in the Zeichenwerk library. It provides hierarchical style resolution with specificity rules, enabling flexible and maintainable widget styling.

## Architecture

### Theme Interface
The `Theme` interface defines the contract for all theme implementations, providing methods for:
- Style management and resolution
- Color variable handling
- Border style definitions
- Unicode character registry
- Boolean configuration flags
- Widget style application

### MapTheme Implementation
`MapTheme` is the concrete implementation that stores theme resources in maps for efficient O(1) lookup:
- **styles**: CSS-like selector to style mappings
- **colors**: Color variable registry (e.g., "$primary" → "#007ACC")
- **borders**: Border style definitions
- **runes**: Special Unicode characters for UI elements
- **flags**: Boolean configuration options

### CSS-Like Selector System
The theme system uses a sophisticated selector format: `type/part.class#id/part:state`

#### Selector Components
- **type**: Widget type (button, input, list, etc.)
- **part**: Widget sub-component (placeholder, bar, item, etc.)
- **class**: CSS class for categorization (.primary, .large, etc.)
- **id**: Unique widget identifier (#submit-button, #main-menu, etc.)
- **state**: Widget state (:focus, :hover, :disabled, etc.)

#### Selector Examples
- `"button"` - Styles all button widgets
- `"button.primary"` - Styles buttons with "primary" class
- `"button#submit"` - Styles the button with ID "submit"
- `"button:focus"` - Styles buttons in focus state
- `"input/placeholder"` - Styles the placeholder part of input widgets
- `"input.large:focus"` - Styles large input widgets when focused
- `"list/item.selected"` - Styles selected items in lists

## Functions and Methods

### Constructor
- `NewMapTheme() *MapTheme` - Creates a new MapTheme with empty registries

### Theme Interface Methods

#### Style Management
- `Add(*Style)` - Adds a style to the theme with automatic parent resolution
- `Get(string) *Style` - Retrieves resolved style for CSS-like selector
- `SetStyles(...*Style)` - Bulk style configuration
- `Styles() []*Style` - Returns all defined styles
- `Apply(Widget, string, ...string)` - Applies styles to widgets with multiple states

#### Color Management
- `Color(string) string` - Resolves color variables (e.g., "$primary" → "#007ACC")
- `Colors() map[string]string` - Returns color variable registry
- `SetColors(map[string]string)` - Bulk color configuration

#### Resource Registries
- `Border(string) BorderStyle` - Retrieves border style by name
- `SetBorders(map[string]BorderStyle)` - Bulk border configuration
- `Rune(string) rune` - Retrieves special Unicode character by name
- `SetRunes(map[string]rune)` - Bulk Unicode character configuration
- `Flag(string) bool` - Retrieves boolean configuration flag
- `SetFlags(map[string]bool)` - Bulk flag configuration

### Internal Utilities
- `split(string) []string` - Parses CSS-like selectors into components using regex
- `selector(int, []string) (string, int)` - Implements CSS specificity resolution algorithm

## Key Features

### Hierarchical Style Resolution
The theme system implements CSS-like specificity rules:
1. **Base styles** (type only) - Lowest specificity
2. **Class styles** (type + class)
3. **State styles** (type + state)
4. **ID styles** - Higher specificity
5. **Combined selectors** (class + state, ID + state, etc.) - Highest specificity

More specific styles override less specific ones, allowing for hierarchical theming with sensible defaults and targeted overrides.

### Automatic Parent Resolution
When styles are added via `Add()`, the system automatically:
1. Parses the selector into components
2. Searches for appropriate parent styles based on specificity
3. Establishes inheritance relationships
4. Fixes the style to prevent further modifications

### Color Variable System
Support for CSS-like color variables:
- Variables prefixed with "$" (e.g., "$primary", "$secondary")
- Automatic resolution in `Color()` method
- Fallback to original value if variable not found
- Enables consistent color schemes across applications

### Widget Style Application
The `Apply()` method provides convenient multi-state styling:
- Applies base selector and additional state combinations
- Handles part-specific styling (e.g., placeholders, bars)
- Supports both prefix and suffix part positioning
- Efficient runtime style assignment to widgets

### Border and Rune Management
- **Borders**: Named border style definitions for consistent widget framing
- **Runes**: Special Unicode characters for arrows, bullets, checkboxes, etc.
- **Flags**: Boolean configuration for theme behavior control

## Regular Expression Parsing
The selector parsing uses a sophisticated regex pattern that captures:
- Group 1: Widget type
- Group 2: Widget part after type
- Group 3: CSS class name (without '.' prefix)
- Group 4: Widget ID (without '#' prefix)
- Group 5: Widget part after ID (alternative placement)
- Group 6: State (without ':' prefix)

The dual part placement (groups 2 and 5) allows flexible selector formats:
- `"button/text.primary"` - part after type
- `"button.primary#submit/text"` - part after ID

## Performance Characteristics
- **O(1) style lookup** for exact selector matches
- **Efficient memory usage** through shared style instances
- **Fast theme switching** by replacing map contents
- **Minimal cascading overhead** with priority-based resolution

## Usage Patterns

### Basic Theme Setup
```go
theme := NewMapTheme()
theme.SetColors(map[string]string{
    "$primary":   "#007ACC",
    "$secondary": "#FF6B35",
})
theme.Add(NewStyle("button").WithBackground("$primary"))
theme.Add(NewStyle("button:focus").WithBackground("$secondary"))
```

### Widget Style Application
```go
// Apply base style and focus state
theme.Apply(button, "button.primary", "focus", "hover")
// Results in: "button.primary", "button.primary:focus", "button.primary:hover"
```

### Inheritance Chain
```go
theme.Add(NewStyle("").WithForeground("black"))           // Base
theme.Add(NewStyle("button").WithBorder("thin"))          // Inherits from base
theme.Add(NewStyle("button.primary").WithBackground("blue")) // Inherits from button
```

## Test Coverage
Achieved 96.8%+ test coverage with comprehensive tests covering:
- Constructor and basic functionality
- Style hierarchy and inheritance
- CSS selector parsing and resolution
- Color variable resolution
- Border, rune, and flag management
- Widget style application
- Edge cases and error conditions
- Complex inheritance scenarios
- Selector specificity rules

## Thread Safety
MapTheme is **not thread-safe**. Applications modifying themes from multiple goroutines must provide external synchronization.
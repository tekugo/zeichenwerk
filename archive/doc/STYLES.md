# Styles

This document describes all the style selectors used by widgets in the
zeichenwerk framework. Styles are automatically applied by the builder
when widgets are created and can be customized through themes.

## Style System Overview

The zeichenwerk style system uses CSS-like selectors to apply different
styles to widgets based on their type, state, and component parts. Styles
are applied automatically by the `Apply()` method in the builder when widgets
are created.

### Selector Format

Selectors follow the pattern: `widget-type[.class][#id][/component][:state]`

- **widget-type**: The base widget type (e.g., "button", "input", "table")
- **.class**: Optional CSS-like class modifier
- **#id**: Optional widget ID modifier  
- **/component**: Sub-component of complex widgets
- **:state**: Widget state (focus, hover, disabled, etc.)

## Widget Styles by Type

### Box

| Selector | Description |
|----------|-------------|
| `box` | Default box container styling |
| `box/title` | Box title text styling |

**States**: None (containers don't have interactive states)

### Button

| Selector | Description |
|----------|-------------|
| `button` | Default button appearance |
| `button:disabled` | Button when disabled/non-interactive |
| `button:focus` | Button when focused (keyboard navigation) |
| `button:hover` | Button when mouse is hovering over it |
| `button:pressed` | Button when being pressed/activated |

### Checkbox

| Selector | Description |
|----------|-------------|
| `checkbox` | Default checkbox appearance (box and label) |
| `checkbox:disabled` | Checkbox when disabled |
| `checkbox:focus` | Checkbox when focused |
| `checkbox:hover` | Checkbox when mouse hovers over it |

### Custom

| Selector | Description |
|----------|-------------|
| `custom` | Custom widget styling |

**Note**: Custom widgets handle their own rendering, styles provide background/border

### Digits

| Selector | Description |
|----------|-------------|
| `digits` | Large digit display styling |

**Usage**: For displaying large ASCII art numbers/letters

### Editor

| Selector | Description |
|----------|-------------|
| `editor` | Multi-line text editor styling |
| `editor/current-line` | Current editing line |
| `editor/current-line-number` | Current line number |
| `editor/line-numbers` | Line numbers |
| `editor/separator` | Separator between line numbers and editor |

### Flex

| Selector | Description |
|----------|-------------|
| `flex` | Flex container styling |
| `flex/shadow` | Shadow effect for flex container |

**Usage**: Layout container - styling typically applies to background and borders

### Grid

| Selector | Description |
|----------|-------------|
| `grid` | Grid container styling |

**Usage**: Table-like layout container - styling for background, borders, and
grid lines

### Hidden

| Selector | Description |
|----------|-------------|
| `hidden` | Hidden widget (spacer) styling |

**Usage**: Invisible widgets used for spacing in layouts

### Input

| Selector | Description |
|----------|-------------|
| `input` | Default text input field styling |
| `input:focus` | Input field when focused (actively being edited) |

**Usage**: Single-line text input with cursor and placeholder support

### Label

| Selector | Description |
|----------|-------------|
| `label` | Static text label styling |

**Usage**: Non-interactive text display

### List

| Selector | Description |
|----------|-------------|
| `list` | List container styling |
| `list:disabled` | List when disabled (non-interactive) |
| `list:focus` | List when focused (keyboard navigation) |
| `list/highlight` | Selected item highlighting |
| `list/highlight:focus` | Selected item when list is focused |

**Usage**: Scrollable list of selectable items

### ProgressBar

| Selector | Description |
|----------|-------------|
| `progress-bar` | Progress bar container/background |
| `progress-bar/bar` | Progress bar fill/foreground |

**Usage**: Visual progress indicators with customizable fill styling

### Scroller

| Selector | Description |
|----------|-------------|
| `scroller` | Scroll container styling |
| `scroller:focus` | Scroller when focused |

**Usage**: Scrollable viewport container with scrollbars

### Separator

| Selector | Description |
|----------|-------------|
| `separator` | Horizontal/vertical line separator styling |

**Usage**: Visual dividers between content sections

### Spinner

| Selector | Description |
|----------|-------------|
| `spinner` | Animated loading indicator styling |

**Usage**: Animated characters for loading states

### Switcher

| Selector | Description |
|----------|-------------|
| `switcher` | Content switcher container styling |

**Usage**: Container that displays one of multiple child widgets

### Table

| Selector | Description |
|----------|-------------|
| `table` | Table container styling |
| `table:focus` | Table when focused (keyboard navigation) |
| `table/grid` | Table grid lines and borders |
| `table/grid:focus` | Grid styling when table is focused |
| `table/header` | Table header row styling |
| `table/header:focus` | Header styling when table is focused |
| `table/highlight` | Selected row highlighting |
| `table/highlight:focus` | Selected row when table is focused |

**Usage**: Data tables with headers, grid lines, and row selection

### Tabs

| Selector | Description |
|----------|-------------|
| `tabs` | Tab container styling |
| `tabs:focus` | Tabs when focused |
| `tabs/line` | Tab separator lines |
| `tabs/line:focus` | Separator lines when focused |
| `tabs/highlight` | Active tab highlighting |
| `tabs/highlight:focus` | Active tab when focused |
| `tabs/highlight-line` | Active tab underline/indicator |
| `tabs/highlight-line:focus` | Tab indicator when focused |

**Usage**: Tab-based navigation with visual indicators

### Text

| Selector | Description |
|----------|-------------|
| `text` | Multi-line text display styling |

**Usage**: Read-only text content with scrolling support

### ThemeSwitch

| Selector | Description |
|----------|-------------|
| `theme-switch` | Theme switcher container styling |

**Usage**: Container for temporarily switching themes

## Style Properties

Each style can define the following properties:

### Visual Properties

- **Background**: Background color (color name or hex)
- **Foreground**: Text/foreground color  
- **Border**: Border style ("", "thin", "thick", "round", "double", etc.)
- **Font**: Font styling ("bold", "italic", "underline", "strikethrough")

### Layout Properties  

- **Margin**: External spacing around widget
- **Padding**: Internal spacing within widget
- **Width/Height**: Size hints for layout system

### Example Style Definition

```go
// In a theme file
theme.SetStyle("button", NewStyle().
    WithBorder("round").
    WithPadding(0, 1).
    WithMargin(0, 1))

theme.SetStyle("button:focus", NewStyle().
    WithBorder("double"))
```

## Custom Styling

### Using Builder Methods

The builder provides methods to customize widget styling:

```go
builder.Button("submit", "Submit").
    Background("blue").            // Default state background
    Foreground(":focus", "white"). // Focus state foreground  
    Border("", "round").           // Default border
    Font("", "bold").              // Bold text
    Padding(0, 1).                 // Horizontal padding
    Margin(1)                      // All-around margin
```

### Using Classes

Apply CSS-like classes for reusable styling:

```go
// Set a class for subsequent widgets
builder.Class("primary").
    Button("submit", "Submit").
    Button("save", "Save").
    Class("") // Reset class

// This creates selectors: "button.primary#submit" and "button.primary#save"
```

### State-Based Styling

Different styles for different widget states:

```go
// Define styles for different states
theme.SetStyle("input", baseInputStyle)
theme.SetStyle("input:focus", focusedInputStyle)
theme.SetStyle("input:disabled", disabledInputStyle)
```

## Built-in Themes

Zeichenwerk includes several built-in themes that define complete styling:

- **DefaultTheme**: Basic terminal-friendly styling
- **TokyoNightTheme**: Dark theme with purple/blue accents
- **GruvboxDarkTheme**: Warm dark theme with earth tones
- **GruvboxLightTheme**: Light variant of Gruvbox
- **MidnightNeonTheme**: Dark theme with bright neon accents
- **NordTheme**: Cool blue-tinted theme

## Best Practices

1. **Consistent State Styling**: Always define focus states for interactive widgets
2. **Theme Compatibility**: Test custom styles across different themes
3. **Accessibility**: Ensure sufficient color contrast for readability
4. **Component Hierarchy**: Use sub-component selectors for complex widgets
5. **Class Organization**: Group related widgets with meaningful class names

## Theme Development

When creating custom themes:

1. **Define Base Styles**: Start with default states for all widget types
2. **Add State Variations**: Define focus, hover, and disabled states
3. **Component Details**: Style sub-components like table headers and highlights
4. **Color Harmony**: Use a consistent color palette across all widgets
5. **Border Consistency**: Use consistent border styles for visual coherence

Example theme structure:

```go
func MyCustomTheme() Theme {
    theme := NewTheme()
    
    // Base widget styles
    theme.SetStyle("button", NewStyle("white", "blue"))
    theme.SetStyle("button:focus", NewStyle("black", "cyan"))
    theme.SetStyle("button:hover", NewStyle("white", "brightblue"))
    
    // Complex widget components
    theme.SetStyle("table", NewStyle("white", "black"))
    theme.SetStyle("table/header", NewStyle("black", "gray"))
    theme.SetStyle("table/highlight", NewStyle("white", "blue"))
    theme.SetStyle("table/highlight:focus", NewStyle("black", "cyan"))
    
    return theme
}
```


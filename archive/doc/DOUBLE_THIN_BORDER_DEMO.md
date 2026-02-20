# Double-Thin Border Style Demo

## Overview
Added a new "double-thin" border style to `theme-unicode-borders.go` that combines double-line outer borders with thin inner grid lines.

## Visual Representation

### Border Style: "double-thin"
```
╔═══╤═══╗  <- Double outer border with mixed connectors
║   │   ║  <- Double sides, thin inner vertical
╟───┼───╢  <- Mixed double-thin connectors with thin cross
║   │   ║  <- Double sides, thin inner vertical  
╚═══╧═══╝  <- Double outer border with mixed connectors
```

### Character Breakdown
- **Outer borders**: Double-line characters (═ ║ ╔ ╗ ╝ ╚)
- **Inner grid**: Thin single-line characters (─ │ ┼ ┬ ┤ ┴ ├)
- **Mixed connectors**: Double-thin transition characters (╤ ╡ ╧ ╞)

## Use Cases
Perfect for:
- **Important data tables** in dialogs that need clear outer boundaries
- **Primary dashboard widgets** with internal subdivisions
- **Featured content sections** that require both emphasis and organization
- **Form containers** with grouped fields that need strong visual hierarchy
- **Modal dialogs** with tabular or grid-based content

## Unicode Characters Used

### Outer Border (Double-line)
- `═` (U+2550) - Double horizontal line
- `║` (U+2551) - Double vertical line
- `╔` (U+2554) - Double top-left corner
- `╗` (U+2557) - Double top-right corner
- `╝` (U+255D) - Double bottom-right corner
- `╚` (U+255A) - Double bottom-left corner

### Mixed Connectors (Double-Thin)
- `╤` (U+2564) - Double horizontal with thin vertical down
- `╡` (U+2561) - Double vertical with thin horizontal left
- `╧` (U+2567) - Double horizontal with thin vertical up  
- `╞` (U+255E) - Double vertical with thin horizontal right

### Inner Grid (Thin single-line)
- `─` (U+2500) - Thin horizontal line
- `│` (U+2502) - Thin vertical line
- `┼` (U+253C) - Thin cross intersection
- `┬` (U+252C) - Thin T-junction (top)
- `┤` (U+2524) - Thin T-junction (right)
- `┴` (U+2534) - Thin T-junction (bottom)
- `├` (U+251C) - Thin T-junction (left)

## Usage Example
```go
theme := NewMapTheme()
AddUnicodeBorders(theme)

// Apply the double-thin border to a widget
widget.SetStyle("", NewStyle("").WithBorder("double-thin"))

// Use in a table or grid that needs strong outer boundaries
// but subtle inner organization
table.SetStyle("", theme.Get("table").WithBorder("double-thin"))
```

## Comparison with Other Border Styles

| Style | Outer Border | Inner Grid | Best For |
|-------|-------------|------------|----------|
| `thin` | Single thin | Single thin | General purpose, lightweight |
| `double` | Double thick | Double thick | Maximum emphasis, heavy visual |
| `thick` | Single thick | Single thick | Strong emphasis, medium weight |
| `thick-thin` | Single thick | Single thin | Strong outer, subtle inner |
| **`double-thin`** | **Double thick** | **Single thin** | **Maximum outer emphasis, subtle inner** |

The new `double-thin` style provides the strongest possible outer visual emphasis (double lines) while maintaining subtle and non-intrusive inner organization (thin lines), making it ideal for the most important UI elements that also need internal structure.
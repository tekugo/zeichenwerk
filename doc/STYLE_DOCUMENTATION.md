# Style System Documentation

## Overview
The `style.go` file implements a comprehensive styling system for TUI widgets in the Zeichenwerk library. It provides CSS-like styling capabilities with support for inheritance, colors, spacing, borders, and fonts.

## Architecture

### Style Struct
The `Style` struct is the core of the styling system, containing:
- **selector**: Style identifier/name
- **parent**: Reference to parent style for inheritance
- **fixed**: Immutability flag to prevent modifications
- **background/foreground**: Color properties
- **font**: Font styling
- **border**: Border style identifier
- **cursor**: Cursor style for interactive widgets
- **margin/padding**: Spacing properties using Insets

### Box Model
The styling system follows the CSS box model:
1. **Content area** (actual widget content)
2. **Padding** (inner spacing with background color)
3. **Border** (decorative border around padding)
4. **Margin** (outer transparent spacing)

## Functions and Methods

### Constructor
- `NewStyle(params ...string) *Style` - Creates a new style with optional selector

### Property Accessors
- `Background() string` - Gets background color with inheritance
- `Foreground() string` - Gets foreground color with inheritance
- `Border() string` - Gets border style with inheritance
- `Cursor() string` - Gets cursor style with inheritance
- `Font() string` - Gets font style with inheritance
- `Fixed() bool` - Checks if style is immutable
- `Margin() *Insets` - Gets margin spacing with inheritance
- `Padding() *Insets` - Gets padding spacing with inheritance
- `Parent() *Style` - Gets parent style reference

### Property Setters (Fluent Interface)
- `WithBackground(background string) *Style` - Sets background color
- `WithForeground(foreground string) *Style` - Sets foreground color
- `WithColors(foreground, background string) *Style` - Sets both colors
- `WithBorder(border string) *Style` - Sets border style
- `WithCursor(cursor string) *Style` - Sets cursor style
- `WithFont(font string) *Style` - Sets font style
- `WithMargin(values ...int) *Style` - Sets margin using CSS-style shorthand
- `WithPadding(values ...int) *Style` - Sets padding using CSS-style shorthand
- `WithParent(parent *Style) *Style` - Sets parent for inheritance

### Layout Calculation Methods
- `Horizontal() int` - Calculates total horizontal spacing (margins + padding + border)
- `Vertical() int` - Calculates total vertical spacing (margins + padding + border)

### Utility Methods
- `Fix() *Style` - Makes style immutable
- `Modifiable() *Style` - Returns modifiable copy if fixed, or self if not
- `Info() string` - Returns formatted string representation of all style properties

## Key Features

### Inheritance System
Styles support parent-child relationships where child styles inherit properties from their parents. If a property is not set on a child style, it will use the parent's value.

### Immutability Support
Styles can be marked as "fixed" to prevent further modifications. When attempting to modify a fixed style, a new child style is created instead.

### CSS-Style Spacing
Margin and padding support CSS-style shorthand notation:
- 1 value: all sides
- 2 values: top/bottom, left/right
- 3 values: top, left/right, bottom
- 4 values: top, right, bottom, left

### Layout Calculations
The `Horizontal()` and `Vertical()` methods calculate the total space consumed by a widget's styling, which is essential for layout algorithms.

## Global Variables
- `DefaultStyle` - Predefined fixed style with black background, white foreground, no border
- `NoInsets` - Empty insets instance for default spacing

## Bug Fixes Applied
- Fixed `Font()` method which was incorrectly returning cursor values instead of font values

## Test Coverage
Achieved 100% test coverage with comprehensive tests covering:
- Constructor edge cases
- Property inheritance scenarios
- Spacing calculations
- Layout calculations
- Edge cases and error conditions
- Default style behavior
- Immutability features
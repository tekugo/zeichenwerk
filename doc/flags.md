# Widget Flags

Widget flags are boolean state variables used to control widget behavior and appearance. They are managed via the `Flag()` and `SetFlag()` methods defined in the `Widget` interface.

## Standard Flags

The following flags are used throughout the codebase:

| Flag | Description |
|------|-------------|
| `hidden` | Controls widget visibility. When true, the widget is not rendered. |
| `focusable` | Indicates whether the widget can receive keyboard focus. Set on interactive widgets like buttons, inputs, checkboxes, etc. |
| `focused` | Indicates the widget currently has keyboard focus. Managed by the UI focus system. |
| `disabled` | Disables user interaction with the widget. Disabled widgets do not respond to input events. |
| `pressed` | Tracks mouse button press state on buttons. Set to true while mouse button is held down on the widget. |
| `checked` | Represents the toggle state of a checkbox. True when checked, false when unchecked. |
| `masked` | Enables password masking on input fields. When true, characters are displayed using a mask character (e.g., `*`). |
| `readonly` | Prevents text editing in input fields. When true, the input cannot be modified but can still receive focus and allow text selection. |
| `hovered` | Indicates the mouse cursor is currently over the widget area. Used for hover effects and styling. |

## Usage

Flags are typically set internally by widgets in response to user input or system events. They can also be set programmatically:

```go
widget.SetFlag("focused", true)
if widget.Flag("hidden") {
    // ...
}
```

The widget's current state (for style resolution) is determined by the `State()` method, which returns the highest priority state among: `disabled`, `pressed`, `focused`, `hovered`.

## Theme Flags

The `Theme.Flag()` method provides a generic flag registry for theme-specific configuration flags. These are not widget state flags but rather theme toggles (e.g., `"debug"`, `"production"`, `"feature_x"`). They are set via `Theme.SetFlags()` and accessed via `Theme.Flag()`.

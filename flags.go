package zeichenwerk

// Flag is the type for widget state flags. Using the named type instead of
// plain strings prevents accidental use of arbitrary strings and makes flag
// parameters self-documenting at call sites.
type Flag string

const (
	// FlagChecked marks a widget (e.g. Checkbox) as checked.
	FlagChecked Flag = "checked"
	// FlagDisabled marks a widget as non-interactive. Disabled widgets do not
	// receive focus or input events and are rendered with the ":disabled" style.
	FlagDisabled Flag = "disabled"
	// FlagFocused indicates that the widget currently holds keyboard focus.
	FlagFocused Flag = "focused"
	// FlagFocusable marks a widget as eligible to receive keyboard focus.
	// Widgets without this flag are skipped during tab-order traversal.
	FlagFocusable Flag = "focusable"
	// FlagHidden makes a widget invisible and excluded from layout.
	FlagHidden Flag = "hidden"
	// FlagHovered indicates that the mouse cursor is over the widget.
	FlagHovered Flag = "hovered"
	// FlagMasked causes input text to be hidden (e.g. password fields).
	FlagMasked Flag = "masked"
	// FlagPressed indicates that the widget is currently being activated
	// (mouse button held down).
	FlagPressed Flag = "pressed"
	// FlagReadonly prevents the widget's value from being modified by the user.
	FlagReadonly Flag = "readonly"
	// FlagSkip excludes the widget from Tab/Shift-Tab focus traversal even when
	// FlagFocusable is set. Useful for read-only or decorative interactive widgets
	// that should not participate in keyboard navigation.
	FlagSkip Flag = "skip"
)

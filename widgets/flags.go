package widgets

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// keep flag constants in alphabetical order
const (
	// FlagChecked marks a widget (e.g. Checkbox) as checked.
	FlagChecked Flag = "checked"

	// FlagDisabled marks a widget as non-interactive. Disabled widgets do not
	// receive focus or input events and are rendered with the ":disabled" style.
	FlagDisabled Flag = "disabled"

	// FlagFocusable marks a widget as eligible to receive keyboard focus.
	// Widgets without this flag are skipped during tab-order traversal.
	FlagFocusable Flag = "focusable"

	// FlagFocused indicates that the widget currently holds keyboard focus.
	FlagFocused Flag = "focused"

	// FlagGrid is used by Table to signal that the inner grid should be rendered.
	FlagGrid Flag = "grid"

	// FlagHidden already defined in core package

	// FlagHorizontal restricts a Viewport to horizontal scrolling only. The
	// child fills the viewport height; no vertical scrollbar is shown.
	FlagHorizontal Flag = "horizontal"

	// FlagHovered indicates that the mouse cursor is over the widget.
	FlagHovered Flag = "hovered"

	// FlagMasked causes input text to be hidden (e.g. password fields).
	FlagMasked Flag = "masked"

	// FlagPressed indicates that the widget is currently being activated
	// (mouse button held down).
	FlagPressed Flag = "pressed"

	// FlagReadonly prevents the widget's value from being modified by the user.
	FlagReadonly Flag = "readonly"

	// FlagRight right-aligns content within a widget's content area.
	// Supported by [Digits].
	FlagRight Flag = "right"

	// FlagSearch if search (currently only in List) is enabled
	FlagSearch Flag = "search"

	// FlagSkip excludes the widget from Tab/Shift-Tab focus traversal even when
	FlagSkip Flag = "skip"

	// FlagVertical restricts a Viewport to vertical scrolling only. The child
	// fills the viewport width; no horizontal scrollbar is shown.
	FlagVertical Flag = "vertical"
)

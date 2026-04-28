package core

// Flag is the type for widget state flags. Flags represent simple boolean
// aspects of a widget's state (for example whether it is hidden, disabled,
// or focused) and are queried and mutated through Widget.Flag and
// Widget.SetFlag.
//
// Using a named string type rather than plain strings provides two benefits:
// it prevents accidental misuse of arbitrary values at call sites, and it
// makes flag parameters self-documenting when they appear in method
// signatures.
type Flag string

// Well-known widget state flags. New flag constants should be added here in
// alphabetical order so the list remains easy to scan.
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

	// FlagHidden marks a widget as hidden. Hidden widgets are skipped by
	// the renderer, by mouse hit-testing (FindAt), and by tab-order focus
	// traversal — from the user's perspective they are not there at all.
	// Their layout bounds, however, are preserved: the widget still
	// occupies its slot in the parent's geometry, so revealing it later is
	// cheap and does not reshuffle surrounding widgets. This makes
	// FlagHidden the idiomatic way to toggle a widget's visibility without
	// detaching it from its parent container.
	FlagHidden Flag = "hidden"

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

	// FlagSkip excludes the widget from Tab/Shift-Tab focus traversal even
	// when FlagFocusable is set.
	FlagSkip Flag = "skip"

	// FlagVertical restricts a Viewport to vertical scrolling only. The child
	// fills the viewport width; no horizontal scrollbar is shown.
	FlagVertical Flag = "vertical"
)

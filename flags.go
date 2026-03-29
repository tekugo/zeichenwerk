package zeichenwerk

// Flag is the type for widget state flags. Using the named type instead of
// plain strings prevents accidental use of arbitrary strings and makes flag
// parameters self-documenting at call sites.
type Flag string

const (
	FlagChecked   Flag = "checked"
	FlagDisabled  Flag = "disabled"
	FlagFocused   Flag = "focused"
	FlagFocusable Flag = "focusable"
	FlagHidden    Flag = "hidden"
	FlagHovered   Flag = "hovered"
	FlagMasked    Flag = "masked"
	FlagPressed   Flag = "pressed"
	FlagReadonly  Flag = "readonly"
)

package core

// Flag is the type for widget state flags. Using the named type instead of
// plain strings prevents accidental use of arbitrary strings and makes flag
// parameters self-documenting at call sites.
type Flag string

// keep flag constants in alphabetical order
const (
	// FlagHidden makes a widget invisible and excluded from layout.
	FlagHidden Flag = "hidden"
)

package core

// Sentinel error codes used throughout the core package.
//
// Each value is a *MessageCode carrying a stable machine-readable identifier
// and a human-readable default message. They are declared as package-level
// variables so that callers can compare against them with errors.Is — or by
// direct identity — to branch on specific failure modes without matching on
// message text.
var (
	// ErrChildIsNil is returned by container Add methods when the widget
	// argument is nil. Containers never accept a nil child because it would
	// corrupt layout traversal and event dispatch.
	ErrChildIsNil *MessageCode = NewErrorCode("child-is-nil", "Child is nil")

	// ErrNoContainer is returned by operations that require a Container but
	// were given a plain Widget. It typically signals a Builder or lookup
	// mismatch where a leaf widget was addressed as if it could hold
	// children.
	ErrNoContainer *MessageCode = NewErrorCode("no-container", "Widget must be a Container")

	// ErrScreenInit is returned when the terminal screen cannot be
	// initialised, for example because the underlying tcell screen failed to
	// open the TTY or could not enter raw mode. It usually indicates an
	// environment problem (no TTY, insufficient permissions, unsupported
	// terminal) rather than a programming error.
	ErrScreenInit *MessageCode = NewErrorCode("screen-init", "Failed to initialise terminal screen")
)

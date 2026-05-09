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

	// ErrNotFound is returned by Container.Remove (and helpers built on it)
	// when the supplied widget is not a direct child of the receiver.
	// Callers can branch on it to decide between a hard failure and a
	// silent no-op.
	ErrNotFound *MessageCode = NewErrorCode("not-found", "Child not found in container")

	// ErrFull is returned by Container.Insert when the receiver cannot accept
	// any more children — typically a fixed-arity container (Box,
	// Collapsible, Dialog, Card, Viewport) that already holds its single
	// child and was asked to insert another at a non-replacing index.
	ErrFull *MessageCode = NewErrorCode("container-full", "Container is full")
)

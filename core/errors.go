package core

// ErrChildIsNil is returned by container Add methods when the widget argument is nil.
var (
	ErrChildIsNil  *MessageCode = NewErrorCode("child-is-nil", "Child is nil")
	ErrNoContainer *MessageCode = NewErrorCode("no-container", "Widget must be a Container")
	ErrScreenInit  *MessageCode = NewErrorCode("screen-init", "Failed to initialise terminal screen")
)

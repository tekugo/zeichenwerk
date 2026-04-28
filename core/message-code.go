package core

import "fmt"

// MessageCode is an error value that bundles three pieces of information:
//
//   - a Level indicating severity,
//   - a short, stable, machine-readable code (kebab-case by convention),
//   - a human-readable default message.
//
// MessageCode implements the standard error interface and participates in
// errors.Is through the Is method, so sentinel values declared once (for
// example in errors.go) can be compared against regardless of whether the
// error has been wrapped with fmt.Errorf("%w", ...).
type MessageCode struct {
	code    string
	message string
	level   Level
}

// Error formats the message code as "[level:code] message", giving log
// output a predictable, greppable prefix while still including the
// human-readable description.
func (m *MessageCode) Error() string {
	return fmt.Sprintf("[%s:%s] %s", m.level, m.code, m.message)
}

// NewErrorCode constructs a MessageCode with Error severity. It is the
// conventional constructor for package-level sentinel errors declared as
// exported variables.
func NewErrorCode(code, message string) *MessageCode {
	return &MessageCode{code: code, message: message, level: Error}
}

// Is reports whether target is a MessageCode carrying the same code as the
// receiver. Only the code field is compared; level and message are ignored
// so that localised or re-worded messages still match their sentinel. This
// makes MessageCode values safe to use with errors.Is even after the error
// has been wrapped.
func (m *MessageCode) Is(target error) bool {
	if t, ok := target.(*MessageCode); ok {
		return m.code == t.code
	}
	return false
}

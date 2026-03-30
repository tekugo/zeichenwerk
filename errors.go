package zeichenwerk

import "fmt"

// Level represents the severity of a log message or error.
type Level string

const (
	Debug   Level = "debug"
	Error   Level = "error"
	Fatal   Level = "fatal"
	Info    Level = "info"
	Warning Level = "warning"
)

// ErrChildIsNil is returned by container Add methods when the widget argument is nil.
var ErrChildIsNil *MessageCode = NewErrorCode("child-is-nil", "Child is nil")

// MessageCode is an error value that carries a severity level, a short
// machine-readable code, and a human-readable message.
type MessageCode struct {
	code    string
	message string
	level   Level
}

// Error implements the error interface, returning a formatted string of the
// form "[level:code] message".
func (m *MessageCode) Error() string {
	return fmt.Sprintf("[%s:%s] %s", m.level, m.code, m.message)
}

// NewErrorCode creates a MessageCode with Error severity.
func NewErrorCode(code, message string) *MessageCode {
	return &MessageCode{code: code, message: message, level: Error}
}

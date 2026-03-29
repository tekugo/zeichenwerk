package zeichenwerk

import "fmt"

type Level string

const (
	Debug   Level = "debug"
	Error   Level = "error"
	Fatal   Level = "fatal"
	Info    Level = "info"
	Warning Level = "warning"
)

var ErrChildIsNil *MessageCode = NewErrorCode("child-is-nil", "Child is nil")

type MessageCode struct {
	code    string
	message string
	level   Level
}

func (m *MessageCode) Error() string {
	return fmt.Sprintf("[%s:%s] %s", m.level, m.code, m.message)
}

func NewErrorCode(code, message string) *MessageCode {
	return &MessageCode{code: code, message: message, level: Error}
}

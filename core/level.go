package core

// Level represents the severity of a log message or error.
type Level string

const (
	Debug   Level = "debug"
	Error   Level = "error"
	Fatal   Level = "fatal"
	Info    Level = "info"
	Warning Level = "warning"
)

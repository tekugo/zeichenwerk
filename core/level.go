package core

// Level identifies the severity of a log entry or of a MessageCode-based
// error. It is declared as a string type so that level values serialise
// directly to log output and can be compared with plain string literals
// when configuring filters.
type Level string

// Severity levels from lowest to highest. The numeric ordering is not
// significant — filters compare on the string value — but this list is
// arranged in the conventional order for readability.
const (
	Debug   Level = "debug"   // fine-grained diagnostic output for development
	Info    Level = "info"    // routine informational messages
	Warning Level = "warning" // unexpected but recoverable conditions
	Error   Level = "error"   // a failure that prevented an operation from completing
	Fatal   Level = "fatal"   // an unrecoverable failure; the program is about to exit
)

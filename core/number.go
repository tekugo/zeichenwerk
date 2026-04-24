package core

// Number is a type set of the built-in numeric types the core package
// supports as generic parameters. It covers every signed integer width and
// both floating-point types, but deliberately excludes unsigned integers
// (which are rare in TUI contexts and whose wrap-around behaviour would
// complicate arithmetic in TimeSeries) and complex numbers.
//
// It is used as the constraint for TimeSeries[T] and for helpers such as
// Humanize that need to accept any numeric type without reflection.
type Number interface {
	int | int8 | int16 | int32 | int64 | float32 | float64
}

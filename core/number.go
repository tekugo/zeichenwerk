package core

// Number is the type constraint for TimeSeries values.
type Number interface {
	int | int8 | int16 | int32 | int64 | float32 | float64
}

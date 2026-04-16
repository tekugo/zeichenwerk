package zeichenwerk

// DataProvider is the data source interface for Sparkline.
// Get(0) returns the most recent (rightmost) value; Get(Size()-1) returns
// the oldest. RingBuffer[float64] and TimeSeries[T] satisfy this interface.
type DataProvider interface {
	Size() int
	Get(index int) float64
}

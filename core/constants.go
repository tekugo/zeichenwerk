package core

import "time"

// DoubleClickThreshold is the maximum elapsed time between two consecutive
// clicks on the same target for the pair to be treated as a double-click.
// Clicks farther apart than this are reported as two independent single
// clicks. The chosen value is a common desktop default and is short enough
// to avoid false positives on slow navigation yet long enough to tolerate
// modest input lag.
const DoubleClickThreshold = 300 * time.Millisecond

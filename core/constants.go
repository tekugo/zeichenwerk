package core

import "time"

// doubleClickThreshold is the maximum time between two clicks on the same item
// for them to be treated as a double-click.
const DoubleClickThreshold = 300 * time.Millisecond

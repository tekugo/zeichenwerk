package zeichenwerk

// This file contains small interface definitions.

// Filterable is implemented by widgets that can progressively filter their
// content based on a query string. Passing an empty string resets the filter
// and restores the full, unfiltered content.
type Filterable interface {
	Filter(filter string)
}

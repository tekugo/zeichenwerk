package zeichenwerk

// Filterable is implemented by widgets that can progressively filter their
// content based on a query string. Passing an empty string resets the filter
// and restores the full, unfiltered content.
type Filterable interface {
	Filter(filter string)
}

// Suggester is an optional extension of [Filterable]. Widgets that implement
// Suggester can supply prefix-based completion candidates to a bound [Filter],
// enabling ghost-text typeahead.
type Suggester interface {
	Suggest(query string) []string
}

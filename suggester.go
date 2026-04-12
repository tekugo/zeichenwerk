package zeichenwerk

// Suggester is an optional extension of [Filterable]. Widgets that implement
// Suggester can supply prefix-based completion candidates to a bound [Filter],
// enabling ghost-text typeahead.
type Suggester interface {
	Suggest(query string) []string
}

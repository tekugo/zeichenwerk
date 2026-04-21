package widgets

import "strings"

// Suggester is an optional extension of [Filterable]. Widgets that implement
// Suggester can supply prefix-based completion candidates to a bound [Filter],
// enabling ghost-text typeahead.
type Suggester interface {
	Suggest(query string) []string
}

// Suggest returns a suggest function for use with Typeahead.SetSuggest.
// It performs case-insensitive prefix matching against the provided candidates.
func Suggest(candidates []string) func(string) []string {
	return func(text string) []string {
		if text == "" {
			return nil
		}
		lower := strings.ToLower(text)
		var matches []string
		for _, c := range candidates {
			if strings.HasPrefix(strings.ToLower(c), lower) {
				matches = append(matches, c)
			}
		}
		return matches
	}
}

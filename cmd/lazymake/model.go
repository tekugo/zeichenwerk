package main

// Target represents a parsed Makefile or Justfile recipe with its metadata.
type Target struct {
	Name        string // Recipe/target name
	Description string // Optional description from ## comment or just doc comment
	Runner      string // "make" or "just"; empty string for status messages
}

// toItems converts a Target slice to []any for Deck.SetItems.
// If the slice is empty an informational placeholder is returned instead.
func toItems(targets []Target, dir string) []any {
	if len(targets) == 0 {
		return []any{Target{
			Name:        "No targets found",
			Description: "No Makefile or Justfile in " + dir,
		}}
	}
	items := make([]any, len(targets))
	for i, t := range targets {
		items[i] = t
	}
	return items
}


package main

// DocStyle describes a named style entry in a document's palette. Empty
// fields cascade to the palette's "default" entry; an empty Border means
// "use the editor's current border family is irrelevant for this style".
type DocStyle struct {
	Fg     string `json:"fg,omitempty"`
	Bg     string `json:"bg,omitempty"`
	Font   string `json:"font,omitempty"`
	Border string `json:"border,omitempty"`
}

// Clone returns an independent copy.
func (s *DocStyle) Clone() *DocStyle {
	if s == nil {
		return nil
	}
	c := *s
	return &c
}

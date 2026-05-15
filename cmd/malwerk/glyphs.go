package main

import "strings"

// GlyphEntry pairs a glyph rune (as a UTF-8 string) with its lookup name.
type GlyphEntry struct {
	Char string
	Name string
}

// glyphs is the curated index used by the i picker. Compact on purpose —
// covers the runes most likely to appear in terminal art. Add to it as
// needed.
var glyphs = []GlyphEntry{
	// Box drawing — light
	{"─", "box light horizontal"},
	{"│", "box light vertical"},
	{"┌", "box light corner top-left"},
	{"┐", "box light corner top-right"},
	{"└", "box light corner bottom-left"},
	{"┘", "box light corner bottom-right"},
	{"├", "box light tee right"},
	{"┤", "box light tee left"},
	{"┬", "box light tee down"},
	{"┴", "box light tee up"},
	{"┼", "box light cross"},

	// Box drawing — heavy
	{"━", "box heavy horizontal"},
	{"┃", "box heavy vertical"},
	{"┏", "box heavy corner top-left"},
	{"┓", "box heavy corner top-right"},
	{"┗", "box heavy corner bottom-left"},
	{"┛", "box heavy corner bottom-right"},
	{"┣", "box heavy tee right"},
	{"┫", "box heavy tee left"},
	{"┳", "box heavy tee down"},
	{"┻", "box heavy tee up"},
	{"╋", "box heavy cross"},

	// Box drawing — double
	{"═", "box double horizontal"},
	{"║", "box double vertical"},
	{"╔", "box double corner top-left"},
	{"╗", "box double corner top-right"},
	{"╚", "box double corner bottom-left"},
	{"╝", "box double corner bottom-right"},
	{"╠", "box double tee right"},
	{"╣", "box double tee left"},
	{"╦", "box double tee down"},
	{"╩", "box double tee up"},
	{"╬", "box double cross"},

	// Box drawing — round corners
	{"╭", "box round corner top-left"},
	{"╮", "box round corner top-right"},
	{"╰", "box round corner bottom-left"},
	{"╯", "box round corner bottom-right"},

	// Block elements
	{"█", "block full"},
	{"▓", "block dark shade"},
	{"▒", "block medium shade"},
	{"░", "block light shade"},
	{"▌", "block left half"},
	{"▐", "block right half"},
	{"▀", "block upper half"},
	{"▄", "block lower half"},
	{" ", "block space"},

	// Geometric shapes
	{"■", "square filled"},
	{"□", "square empty"},
	{"●", "circle filled"},
	{"○", "circle empty"},
	{"◆", "diamond filled"},
	{"◇", "diamond empty"},
	{"▲", "triangle up filled"},
	{"△", "triangle up empty"},
	{"▼", "triangle down filled"},
	{"▽", "triangle down empty"},
	{"◀", "triangle left filled"},
	{"▶", "triangle right filled"},
	{"★", "star filled"},
	{"☆", "star empty"},

	// Arrows
	{"←", "arrow left"},
	{"→", "arrow right"},
	{"↑", "arrow up"},
	{"↓", "arrow down"},
	{"↔", "arrow horizontal"},
	{"↕", "arrow vertical"},
	{"⇐", "arrow double left"},
	{"⇒", "arrow double right"},
	{"⇑", "arrow double up"},
	{"⇓", "arrow double down"},

	// Math / symbols
	{"±", "plus minus"},
	{"×", "multiply"},
	{"÷", "divide"},
	{"≈", "approx"},
	{"≠", "not equal"},
	{"≤", "less or equal"},
	{"≥", "greater or equal"},
	{"∞", "infinity"},
	{"°", "degree"},
	{"§", "section"},
	{"¶", "paragraph"},
	{"†", "dagger"},
	{"‡", "double dagger"},
	{"•", "bullet"},
	{"·", "middle dot"},

	// Currency / typography
	{"€", "euro sign"},
	{"£", "pound sign"},
	{"¥", "yen sign"},
	{"©", "copyright sign"},
	{"®", "registered sign"},
	{"™", "trademark"},
	{"…", "ellipsis"},
	{"–", "en dash"},
	{"—", "em dash"},

	// Nerd font icons (common subset; pre-resolved private use codepoints)
	{"", "nerd folder"},
	{"", "nerd file"},
	{"", "nerd gear"},
	{"", "nerd search"},
	{"", "nerd warning"},
	{"", "nerd error"},
	{"", "nerd info"},
	{"", "nerd check"},
	{"", "nerd cross"},
	{"", "nerd heart"},
	{"", "nerd star"},
	{"", "nerd home"},
	{"", "nerd user"},
	{"", "nerd users"},
	{"", "nerd envelope"},
	{"", "nerd linux"},
	{"", "nerd apple"},
	{"", "nerd windows"},
	{"", "nerd git"},
	{"", "nerd github"},
	{"", "nerd terminal"},
	{"", "nerd vim"},
	{"", "nerd go"},
}

// GlyphIndex returns the embedded glyph list.
func GlyphIndex() []GlyphEntry { return glyphs }

// filterGlyphs returns entries whose name contains every word in the
// query (case-insensitive). An empty query returns all entries.
func filterGlyphs(entries []GlyphEntry, query string) []GlyphEntry {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		out := make([]GlyphEntry, len(entries))
		copy(out, entries)
		return out
	}
	words := strings.Fields(q)
	var out []GlyphEntry
	for _, e := range entries {
		name := strings.ToLower(e.Name)
		match := true
		for _, w := range words {
			if !strings.Contains(name, w) {
				match = false
				break
			}
		}
		if match {
			out = append(out, e)
		}
	}
	return out
}

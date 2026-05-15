package main

import (
	"embed"
	"sort"
	"strings"
	"unicode/utf8"

	. "github.com/tekugo/zeichenwerk"
)

// Entry describes a single widget demo page.
type Entry struct {
	Category string
	Name     string
	Summary  string
	DocFile  string
	Builder  string
	Compose  string
	DemoFn   func(*Builder)
}

//go:embed docs/*.md
var docFS embed.FS

// Doc returns the rendered markdown reference for this entry, or a placeholder
// when no doc file is associated.
func (e Entry) Doc() string {
	if e.DocFile == "" {
		return "# " + e.Name + "\n\n_No reference documentation._"
	}
	data, err := docFS.ReadFile("docs/" + e.DocFile)
	if err != nil {
		return "# " + e.Name + "\n\n_Reference documentation missing: " + e.DocFile + "_"
	}
	return string(data)
}

// allEntries is assembled at package init from the per-category slices.
var allEntries []Entry

// entryByName indexes entries by lowercase name for navigation lookup.
var entryByName = map[string]int{}

func registerEntries(slices ...[]Entry) {
	for _, s := range slices {
		allEntries = append(allEntries, s...)
	}
	sort.SliceStable(allEntries, func(i, j int) bool {
		if allEntries[i].Category != allEntries[j].Category {
			return categoryOrder(allEntries[i].Category) < categoryOrder(allEntries[j].Category)
		}
		return allEntries[i].Name < allEntries[j].Name
	})
	for i, e := range allEntries {
		entryByName[strings.ToLower(e.Name)] = i
	}
}

func categoryOrder(cat string) int {
	switch cat {
	case "Containers":
		return 0
	case "Input":
		return 1
	case "Display":
		return 2
	case "Animated":
		return 3
	case "Custom":
		return 4
	}
	return 99
}

// navItems returns the navigation list rows. Category headers are inserted as
// disabled section markers (rendered as a separator line in the list).
//
// Header label format: `─ <Name> ────────…` padded with U+2500 to navHeaderWidth
// so the title aligns with the entry rows below it (2-space indent) and the
// trailing rule extends to the far right of the list.
func navItems() ([]string, map[int]int) {
	const navHeaderWidth = 32
	items := []string{}
	indexMap := map[int]int{} // list-row → entry-index
	lastCat := ""
	for i, e := range allEntries {
		if e.Category != lastCat {
			label := "─ " + e.Category + " "
			for utf8.RuneCountInString(label) < navHeaderWidth {
				label += "─"
			}
			items = append(items, label)
			lastCat = e.Category
		}
		indexMap[len(items)] = i
		items = append(items, "  "+e.Name)
	}
	return items, indexMap
}

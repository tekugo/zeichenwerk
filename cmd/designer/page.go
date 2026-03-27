package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"unicode/utf8"

	"github.com/tekugo/zeichenwerk"
)

// Cell is one character in the canvas grid.
type Cell struct {
	ch    rune
	style string // name of a zeichenwerk style; "" == terminal default
}

func (c Cell) isEmpty() bool { return c.ch == 0 }

// Page is a fixed-size grid of Cells.
type Page struct {
	name   string
	width  int
	height int
	cells  [][]Cell
}

func newPage(name string, w, h int) *Page {
	cells := make([][]Cell, h)
	for y := range cells {
		cells[y] = make([]Cell, w)
	}
	return &Page{name: name, width: w, height: h, cells: cells}
}

func (p *Page) at(x, y int) Cell {
	if x < 0 || y < 0 || x >= p.width || y >= p.height {
		return Cell{}
	}
	return p.cells[y][x]
}

func (p *Page) set(x, y int, c Cell) {
	if x < 0 || y < 0 || x >= p.width || y >= p.height {
		return
	}
	p.cells[y][x] = c
}

// --- JSON representation ---

type styleJSON struct {
	Parent string `json:"parent,omitempty"`
	Fg     string `json:"fg,omitempty"`
	Bg     string `json:"bg,omitempty"`
	Font   string `json:"font,omitempty"`
	Attr   string `json:"attr,omitempty"` // legacy alias for Font
}

type cellJSON struct {
	R     int    `json:"r"`
	C     int    `json:"c"`
	Ch    string `json:"ch"`
	Style string `json:"style,omitempty"`
	// Legacy inline fields from older file format.
	Fg   string `json:"fg,omitempty"`
	Bg   string `json:"bg,omitempty"`
	Attr string `json:"attr,omitempty"`
}

type pageJSON struct {
	Name   string     `json:"name"`
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Cells  []cellJSON `json:"cells"`
}

type docJSON struct {
	Version int                  `json:"version"`
	Name    string               `json:"name"`
	Styles  map[string]styleJSON `json:"styles,omitempty"`
	Pages   []pageJSON           `json:"pages"`
}

// newStyleCanvas creates a fresh Canvas used only for its style registry.
// It pre-registers the built-in "default" style.
func newStyleCanvas() (*zeichenwerk.Canvas, []string) {
	c := zeichenwerk.NewCanvas("styles", 0, 0)
	c.SetStyle("default", zeichenwerk.NewStyle("default"))
	return c, []string{"default"}
}

func saveDoc(filename, docName string, page *Page, canvas *zeichenwerk.Canvas, styleOrder []string) error {
	// Build a pointer→name reverse map for parent lookup.
	ptrName := map[*zeichenwerk.Style]string{}
	for _, name := range styleOrder {
		ptrName[canvas.Style(name)] = name
	}

	var stylesMap map[string]styleJSON
	for _, name := range styleOrder {
		if name == "default" {
			continue // implicit; never serialised
		}
		s := canvas.Style(name)
		parentName := ""
		if p := s.Parent(); p != nil {
			parentName = ptrName[p]
		}
		if stylesMap == nil {
			stylesMap = map[string]styleJSON{}
		}
		stylesMap[name] = styleJSON{
			Parent: parentName,
			Fg:     s.OwnForeground(),
			Bg:     s.OwnBackground(),
			Font:   s.OwnFont(),
		}
	}

	var cells []cellJSON
	for y := 0; y < page.height; y++ {
		for x := 0; x < page.width; x++ {
			c := page.at(x, y)
			if c.isEmpty() {
				continue
			}
			cells = append(cells, cellJSON{
				R:     y,
				C:     x,
				Ch:    string(c.ch),
				Style: c.style,
			})
		}
	}

	doc := docJSON{
		Version: 1,
		Name:    docName,
		Styles:  stylesMap,
		Pages: []pageJSON{{
			Name:   page.name,
			Width:  page.width,
			Height: page.height,
			Cells:  cells,
		}},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func loadDoc(filename string) (docName string, page *Page, canvas *zeichenwerk.Canvas, styleOrder []string, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", nil, nil, nil, err
	}
	var doc docJSON
	if err := json.Unmarshal(data, &doc); err != nil {
		return "", nil, nil, nil, fmt.Errorf("parse error: %w", err)
	}

	canvas, styleOrder = newStyleCanvas()

	// Two-pass style loading: first create all style objects, then wire parents.
	nameToStyle := map[string]*zeichenwerk.Style{}
	nameToStyle["default"] = canvas.Style("default")

	// Pass 1: create styles without parents.
	for name, sj := range doc.Styles {
		font := sj.Font
		if font == "" {
			font = sj.Attr // legacy field
		}
		s := zeichenwerk.NewStyle(name).
			WithForeground(sj.Fg).
			WithBackground(sj.Bg).
			WithFont(font)
		nameToStyle[name] = s
	}

	// Pass 2: wire parents and register; collect ordered names.
	for name, sj := range doc.Styles {
		s := nameToStyle[name]
		if sj.Parent != "" {
			if parent, ok := nameToStyle[sj.Parent]; ok {
				s.WithParent(parent)
			}
		}
		canvas.SetStyle(name, s)
		styleOrder = append(styleOrder, name)
	}

	// Sort: "default" first, then alphabetical.
	sort.Slice(styleOrder, func(i, j int) bool {
		if styleOrder[i] == "default" {
			return true
		}
		if styleOrder[j] == "default" {
			return false
		}
		return styleOrder[i] < styleOrder[j]
	})

	if len(doc.Pages) == 0 {
		return doc.Name, newPage("main", 80, 23), canvas, styleOrder, nil
	}
	pd := doc.Pages[0]
	p := newPage(pd.Name, pd.Width, pd.Height)

	for _, cj := range pd.Cells {
		r, _ := utf8.DecodeRuneInString(cj.Ch)
		if r == utf8.RuneError {
			continue
		}
		styleName := cj.Style
		// Backward compat: old files stored fg/bg/attr inline on each cell.
		if styleName == "" && (cj.Fg != "" || cj.Bg != "" || cj.Attr != "") {
			styleName = fmt.Sprintf("_auto_%s|%s|%s", cj.Fg, cj.Bg, cj.Attr)
			if _, exists := nameToStyle[styleName]; !exists {
				s := zeichenwerk.NewStyle(styleName).
					WithForeground(cj.Fg).
					WithBackground(cj.Bg).
					WithFont(cj.Attr)
				nameToStyle[styleName] = s
				canvas.SetStyle(styleName, s)
				styleOrder = append(styleOrder, styleName)
			}
		}
		p.set(cj.C, cj.R, Cell{ch: r, style: styleName})
	}

	return doc.Name, p, canvas, styleOrder, nil
}

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
)

func main() {
	themeName := flag.String("t", "tokyo", "theme: tokyo, gruvbox-dark, gruvbox-light, nord, lipstick, midnight")
	width := flag.Int("w", 0, "initial canvas width (default: terminal width)")
	height := flag.Int("h", 0, "initial canvas height (default: terminal height - 1 for status bar)")
	flag.Parse()

	theme := resolveTheme(*themeName)

	var doc *Document
	if path := flag.Arg(0); path != "" {
		loaded, err := LoadDocument(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "malwerk: %v\n", err)
			os.Exit(1)
		}
		doc = loaded
	}

	app := NewApp(theme, doc, *width, *height)
	app.Run()
}

func resolveTheme(name string) *core.Theme {
	switch name {
	case "tokyo":
		return themes.TokyoNight()
	case "gruvbox-dark":
		return themes.GruvboxDark()
	case "gruvbox-light":
		return themes.GruvboxLight()
	case "nord":
		return themes.Nord()
	case "lipstick":
		return themes.Lipstick()
	case "midnight":
		return themes.MidnightNeon()
	default:
		return themes.TokyoNight()
	}
}


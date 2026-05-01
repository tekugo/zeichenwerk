package main

import (
	"flag"
	"os"

	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
)

func main() {
	theme, dir := parseFlags()
	ui := buildUI(theme, dir)
	ui.Run()
}

func parseFlags() (*core.Theme, string) {
	t := flag.String("t", "tokyo", "Theme: midnight, tokyo, nord, gruvbox-dark, gruvbox-light, lipstick")
	d := flag.String("d", "", "Working directory (defaults to current directory)")
	flag.Parse()

	var theme *core.Theme
	switch *t {
	case "midnight":
		theme = themes.MidnightNeon()
	case "nord":
		theme = themes.Nord()
	case "gruvbox-dark":
		theme = themes.GruvboxDark()
	case "gruvbox-light":
		theme = themes.GruvboxLight()
	case "lipstick":
		theme = themes.Lipstick()
	default:
		theme = themes.TokyoNight()
	}

	dir := *d
	if dir == "" {
		dir, _ = os.Getwd()
	}

	return theme, dir
}

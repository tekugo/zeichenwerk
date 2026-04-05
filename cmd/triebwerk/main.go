package main

import (
	"flag"
	"os"

	. "github.com/tekugo/zeichenwerk"
)

func main() {
	theme, dir := parseFlags()
	ui := buildUI(theme, dir)
	ui.Run()
}

func parseFlags() (*Theme, string) {
	t := flag.String("t", "tokyo", "Theme: midnight, tokyo, nord, gruvbox-dark, gruvbox-light, lipstick")
	d := flag.String("d", "", "Working directory (defaults to current directory)")
	flag.Parse()

	var theme *Theme
	switch *t {
	case "midnight":
		theme = MidnightNeonTheme()
	case "nord":
		theme = NordTheme()
	case "gruvbox-dark":
		theme = GruvboxDarkTheme()
	case "gruvbox-light":
		theme = GruvboxLightTheme()
	case "lipstick":
		theme = LipstickTheme()
	default:
		theme = TokyoNightTheme()
	}

	dir := *d
	if dir == "" {
		dir, _ = os.Getwd()
	}

	return theme, dir
}

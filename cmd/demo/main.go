package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
	. "github.com/tekugo/zeichenwerk/values"
	. "github.com/tekugo/zeichenwerk/widgets"
)

func parseFlags() (*Theme, bool, bool, bool) {
	t := flag.String("t", "tokyo", "Theme: midnight, tokyo, nord, gruvbox-dark, gruvbox-light, lipstick")
	dbg := flag.Bool("debug", false, "Start in debug mode")
	dmp := flag.Bool("dump", false, "Dump widget hierarchy to stdout and exit")
	dmpV := flag.Bool("dump-verbose", false, "Dump widget hierarchy with style details to stdout and exit")
	flag.Parse()
	var theme *Theme
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
	return theme, *dbg, *dmp, *dmpV
}

// main function
func main() {
	theme, dbg, dmp, dmpV := parseFlags()
	ui := createUI(theme)
	if dmp || dmpV {
		ui.SetBounds(0, 0, 120, 40)
		ui.Layout()
		ui.Dump(os.Stdout, DumpOptions{Style: dmpV})
		return
	}
	if dbg {
		ui.Debug()
	}
	ui.Run()
}

// Create the terminal UI.
func createUI(theme *Theme) *UI {
	ui := NewBuilder(theme).
		VFlex("main", Stretch, 0).
		HFlex("header", Stretch, 2).
		Static("title", "Zeichenwerk Demo").
		Static("subtitle", "A terminal UI framework").
		End().
		Grid("content", 2, 2, true).Hint(0, -1).Columns(32, -1).Rows(-1, 10).
		Cell(0, 0, 1, 2).
		List("navigation", "Bar Chart", "Box", "Breadcrumb", "Canvas", "Checkbox", "Collapsible", "Color Picker", "Combo", "Deck", "Digits", "Editor", "Filter", "Form", "Grid", "Heatmap", "Marquee", "Progress", "Scanner", "Sparkline", "Select", "Shimmer", "Spinner", "Styled", "Table", "Tabs", "Terminal", "Tiles", "Tree FS", "Typeahead", "Typewriter", "Value", "Viewport", "Commands", "Dialog", "Confirm", "Prompt", "File Chooser", "Dir Chooser", "Save As").
		Cell(1, 0, 1, 1).
		Switcher("switcher", false).
		With(barChartDemo).
		With(box).
		With(breadcrumbDemo).
		With(canvas).
		With(checkbox).
		With(collapsibleDemo).
		With(colorPickerDemo).
		With(comboDemo).
		With(func(b *Builder) { deckDemo(b, theme) }).
		With(digits).
		With(editor).
		With(filterDemo).
		With(form).
		With(grid).
		With(heatmapDemo).
		With(marqueeDemo).
		With(progress).
		With(scanner).
		With(sparklineDemo).
		With(dropdown).
		With(shimmerDemo).
		With(spinner).
		With(styled).
		With(table).
		With(tabs).
		With(terminalDemo).
		With(tilesDemo).
		With(treeFSDemo).
		With(typeaheadDemo).
		With(typewriterDemo).
		With(valueDemo).
		With(viewport).
		With(commandsDemo).
		End().
		Cell(1, 1, 1, 1).
		VFlex("debug-log-pane", Stretch, 0).Hint(0, 10).
		Static("debug-log-title", "Debug Log").Background("green").
		Text("debug-log", []string{"Hello, World!"}, true, 100).Hint(0, -1).
		End().
		End().
		HFlex("footer", Center, 0).
		Shortcuts("footer-shortcuts", "↑↓", "navigate", "Enter", "select", "i", "debug", "t", "theme", "q", "quit").
		Spacer().Hint(-1, 0).
		Static("theme-label", " Theme: ").
		Select("theme-select", "tokyo", "Tokyo Night", "gruvbox-dark", "Gruvbox Dark", "gruvbox-light", "Gruvbox Light", "nrrd", "Nord", "neon", "Midnight Neon").
		End().
		Build()

	themes := map[string]*Theme{
		"tokyo":         themes.TokyoNight(),
		"gruvbox-dark":  themes.GruvboxDark(),
		"gruvbox-light": themes.GruvboxLight(),
		"nrrd":          themes.Nord(),
		"neon":          themes.MidnightNeon(),
	}

	Find(ui, "theme-select").On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 1 {
			if key, ok := data[0].(string); ok {
				if theme, found := themes[key]; found {
					ui.SetTheme(theme)
				}
			}
		}
		return true
	})

	// Register demo commands and bind Ctrl+K globally.
	registerCommandsDemo(ui)
	OnKey(Find(ui, "main"), func(e *tcell.EventKey) bool {
		if e.Key() == tcell.KeyCtrlK {
			ui.Commands().Open()
			return true
		}
		return false
	})

	switcher := Find(ui, "switcher").(*Switcher)
	Find(ui, "navigation").On(EvtActivate, func(_ Widget, event Event, data ...any) bool {
		if len(data) == 1 {
			if selected, ok := data[0].(int); ok {
				if selected < len(switcher.Children()) {
					switcher.Select(selected)
				} else {
					switch selected {
					case 33:
						dialog := ui.NewBuilder().
							Dialog("dialog", "Test Dialog").
							Class("dialog").
							VFlex("dialog-content", Stretch, 1).
							Static("", "Do you really want to do this?").
							HFlex("dialog-buttons", End, 2).
							Button("btn-yes", "YES").
							Button("btn-no", "NO").
							End().
							End().
							Class("").
							Container()
						Find(dialog, "btn-yes").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
							ui.Close()
							return true
						})
						Find(dialog, "btn-no").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
							ui.Close()
							return true
						})
						ui.Popup(-1, -1, 0, 0, dialog)
					case 34:
						ui.Confirm("Confirm Action", "Do you really want to do this?",
							func() {
								if log, ok := Find(ui, "debug-log").(*Text); ok {
									log.Add("Confirm → OK")
								}
							},
							func() {
								if log, ok := Find(ui, "debug-log").(*Text); ok {
									log.Add("Confirm → Cancel")
								}
							},
						)
					case 35:
						ui.Prompt("Enter Value", "Please enter a value:",
							func(text string) {
								if log, ok := Find(ui, "debug-log").(*Text); ok {
									log.Add("Prompt → " + text)
								}
							},
							func() {
								if log, ok := Find(ui, "debug-log").(*Text); ok {
									log.Add("Prompt → Cancel")
								}
							},
						)
					case 36:
						d := ui.FileChooser("Open File", "Open", "file", "", false)
						d.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
							if log, ok := Find(ui, "debug-log").(*Text); ok {
								log.Add("File → " + data[0].(string))
							}
							return true
						})
					case 37:
						d := ui.FileChooser("Open Directory", "Select", "dir", "", false)
						d.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
							if log, ok := Find(ui, "debug-log").(*Text); ok {
								log.Add("Dir → " + data[0].(string))
							}
							return true
						})
					case 38:
						d := ui.FileChooser("Save As", "Save", "save", "", false)
						d.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
							path := data[0].(string)
							save := func() {
								if log, ok := Find(ui, "debug-log").(*Text); ok {
									log.Add("Save As → " + path)
								}
							}
							if _, err := os.Stat(path); err == nil {
								ui.Confirm("Overwrite?",
									filepath.Base(path)+" already exists. Overwrite?",
									save, nil,
								)
							} else {
								save()
							}
							return true
						})
					}
				}
			}
		}
		return true
	})

	return ui
}

// Box demo
func box(builder *Builder) {
	builder.VFlex("box-demo", Stretch, 1).Padding(1).
		Static("box-title", "Box Widget Demo").Padding(0, 0, 1, 0).
		HRule("thin").
		VFlex("box-examples", Stretch, 1).
		Box("simple-box", "Simple Box").Padding(1).
		Static("box-content1", "This is content inside a simple box widget.").
		End().
		Box("styled-box", "Styled Box").Padding(1).Border("", "double").
		Static("box-content2", "This box has a double border style.").
		End().
		Box("padded-box", "Padded Box").Padding(2).Border("", "round").
		Static("box-content3", "This box has extra padding and rounded borders.").
		End().
		End().
		HRule("double").
		Static("box-info", "Boxes are containers that can hold a single child widget with optional borders and titles.").Padding(1, 0, 0, 0).
		End()
}

// Canvas demo
func canvas(builder *Builder) {
	// Create a 40x20 canvas with a simple pattern
	c := NewCanvas("demo-canvas", "", 1, 40, 20)

	// Set styles for normal and insert modes
	normalStyle := NewStyle("").WithColors("white", "black").WithCursor("block")
	insertStyle := NewStyle("").WithColors("cyan", "black").WithCursor("bar")

	c.SetStyle("", normalStyle)
	c.SetStyle(":insert", insertStyle)

	// Fill with empty cells
	c.Fill("", normalStyle)

	// Draw a border on the edges
	borderStyle := NewStyle("").WithColors("yellow", "black")
	c.SetCell(0, 0, "+", borderStyle)
	c.SetCell(39, 0, "+", borderStyle)
	c.SetCell(0, 19, "+", borderStyle)
	c.SetCell(39, 19, "+", borderStyle)
	// Horizontal lines
	for x := 1; x < 39; x++ {
		c.SetCell(x, 0, "-", borderStyle)
		c.SetCell(x, 19, "-", borderStyle)
	}
	// Vertical lines
	for y := 1; y < 19; y++ {
		c.SetCell(0, y, "|", borderStyle)
		c.SetCell(39, y, "|", borderStyle)
	}

	// Add some instruction text
	titleStyle := NewStyle("").WithColors("green", "black")
	c.SetCell(2, 2, "Canvas Widget Demo", titleStyle)

	infoStyle := NewStyle("").WithColors("gray", "black")
	info := "NORMAL mode: hjkl/arrows move, 'i' or 'a' enters INSERT mode, ESC returns"
	for i, ch := range info {
		if i < 38 {
			c.SetCell(2+i, 4, string(ch), infoStyle)
		}
	}

	info2 := "INSERT mode: type to insert chars, arrows still move, ESC returns"
	for i, ch := range info2 {
		if i < 38 {
			c.SetCell(2+i, 5, string(ch), infoStyle)
		}
	}

	// Add it to the builder
	builder.VFlex("canvas-demo", Stretch, 1).Padding(1).
		Static("canvas-title", "Canvas Widget (press 'i' to start editing)").Padding(0, 0, 1, 0).
		Add(c).
		End()
}

// Checkbox demo
func checkbox(builder *Builder) {
	builder.VFlex("checkbox-demo", Stretch, 1).Padding(1, 2).
		Static("checkbox-title", "Checkbox Widget Demo").Padding(0, 0, 1, 0).
		Static("checkbox-info", "Checkboxes toggle between checked and unchecked states.").Padding(0, 0, 1, 0).
		Checkbox("cb1", "Enable notifications", false).
		Checkbox("cb2", "Remember login", true).
		Checkbox("cb3", "Auto-save documents", false).
		Checkbox("cb4", "Show hidden files", true).
		Checkbox("cb5", "I agree to the terms and conditions", false).
		Static("checkbox-status", "Toggle checkboxes with Space or Enter key!").Padding(1, 0, 0, 0).
		End()

	container := builder.Container()
	for i := 1; i <= 5; i++ {
		cbID := fmt.Sprintf("cb%d", i)
		if cb := Find(container, cbID); cb != nil {
			cb.On(EvtChange, func(_ Widget, event Event, data ...any) bool {
				checked := data[0].(bool)
				if statusLabel := Find(container, "checkbox-status"); statusLabel != nil {
					if label, ok := statusLabel.(*Static); ok {
						var name string
						switch cbID {
						case "cb1":
							name = "Notifications"
						case "cb2":
							name = "Remember login"
						case "cb3":
							name = "Auto-save"
						case "cb4":
							name = "Show hidden"
						case "cb5":
							name = "Terms agreed"
						}
						label.Set(fmt.Sprintf("%s: %v", name, checked))
					}
				}
				return true
			})
		}
	}
}

// Combo demo
func comboDemo(builder *Builder) {
	history := []string{
		"fix memory leak in renderer",
		"add dark mode support",
		"refactor event handling",
		"update dependencies",
		"fix race condition in UI loop",
		"implement file chooser",
		"add keyboard shortcuts",
		"improve scroll performance",
	}

	builder.VFlex("combo-demo", Start, 1).Padding(1, 2).
		Static("combo-title", "Combo Widget Demo").Padding(0, 0, 1, 0).
		Static("combo-desc", "Type freely or pick from the list. ↓↑ navigate, Tab/→ accepts ghost text, Enter confirms.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Static("", "Search:").
		Combo("demo-combo", history...).
		Static("combo-status", "").Padding(1, 0, 0, 0).
		End()

	container := builder.Container()
	combo := Find(container, "demo-combo").(*Combo)
	combo.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if s, ok := data[0].(string); ok {
			if label, ok := Find(container, "combo-status").(*Static); ok {
				label.Set("Submitted: " + s)
			}
		}
		return true
	})
}

// Collapsible demo
func collapsibleDemo(builder *Builder) {
	builder.VFlex("collapsible-demo", Stretch, 1).Padding(1, 2).
		Static("collapsible-title", "Collapsible Widget Demo").Padding(0, 0, 1, 0).
		Static("collapsible-info", "Click the header or press Enter/Space to toggle. → expands, ← collapses.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Collapsible("col-basic", "Basic section (starts expanded)", true).
		VFlex("col-basic-content", Stretch, 1).Padding(0, 1).
		Static("", "This is the body of the first collapsible.").
		Static("", "It can contain any widget — here a few statics.").
		Static("", "Collapse me with ← or by clicking the header.").
		End().
		End().
		Collapsible("col-list", "List section (starts collapsed)", false).
		List("col-list-items", "Alpha", "Beta", "Gamma", "Delta", "Epsilon").
		End().
		Collapsible("col-inputs", "Input section (starts collapsed)", false).
		VFlex("col-inputs-content", Stretch, 1).Padding(0, 1).
		Static("", "Name:").
		Input("col-name", "").
		Static("", "Email:").
		Input("col-email", "").
		End().
		End().
		Static("col-status", "").Padding(1, 0, 0, 0).
		End()

	container := builder.Container()
	for _, id := range []string{"col-basic", "col-list", "col-inputs"} {
		id := id
		if w := Find(container, id); w != nil {
			w.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
				if v, ok := data[0].(bool); ok {
					state := "collapsed"
					if v {
						state = "expanded"
					}
					if label, ok := Find(container, "col-status").(*Static); ok {
						label.Set(fmt.Sprintf("%s: %s", id, state))
					}
				}
				return true
			})
		}
	}
}

// Color Picker demo — shows both single-colour and fg/bg pickers.
func colorPickerDemo(builder *Builder) {
	builder.VFlex("color-picker-demo", Stretch, 1).Padding(1, 2).
		Static("cp-title", "Color Picker Demo").Padding(0, 0, 1, 0).
		Static("cp-info", "Edit any of R/G/B, H/S/L, or Hex — the other representations update automatically.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Static("cp-single-label", "Single colour:").Padding(0, 0, 0, 0).
		ColorPicker("cp-single", ColorSingle).Padding(1, 0).
		Static("cp-fgbg-label", "Foreground / background with contrast ratio:").Padding(1, 0, 0, 0).
		ColorPicker("cp-fgbg", ColorFgBg).Padding(1, 0).
		Static("cp-status", "").Padding(1, 0, 0, 0).
		End()

	container := builder.Container()

	if single, ok := Find(container, "cp-single").(*ColorPicker); ok {
		single.SetForeground("#ff8040")
		single.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
			if cp, ok := data[0].(*ColorPicker); ok {
				if label, ok := Find(container, "cp-status").(*Static); ok {
					label.Set(fmt.Sprintf("single: %s", cp.Foreground()))
				}
			}
			return true
		})
	}

	if fgbg, ok := Find(container, "cp-fgbg").(*ColorPicker); ok {
		fgbg.SetForeground("#ffffff")
		fgbg.SetBackground("#1a1b26")
		fgbg.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
			if cp, ok := data[0].(*ColorPicker); ok {
				if label, ok := Find(container, "cp-status").(*Static); ok {
					label.Set(fmt.Sprintf("fg=%s  bg=%s  contrast=%.1f",
						cp.Foreground(), cp.Background(), cp.Contrast()))
				}
			}
			return true
		})
	}
}

// Deck demo — displays all theme colors as rich multi-line cards.
func deckDemo(builder *Builder, theme *Theme) {
	type colorItem struct{ name, hex string }

	colors := theme.Colors()

	names := make([]string, 0, len(colors))
	for k := range colors {
		names = append(names, k)
	}
	sort.Strings(names)

	items := make([]any, len(names))
	for i, name := range names {
		items[i] = colorItem{name: name, hex: colors[name]}
	}

	const previewW = 10 // columns for color swatch
	const borderW = 1   // column for left-side selection indicator
	const padW = 1      // padding column between indicator and text

	// deck is declared first so the render closure can reference it.
	var deck *Deck

	renderFn := func(r *Renderer, x, y, w, h, _ int, data any, selected, focused bool) {
		item := data.(colorItem)
		textW := w - borderW - padW - previewW

		bg := theme.Color("$bg1")

		var fg, font, indicator, indicatorFg string
		if selected {
			font = "bold"
			if focused {
				fg = theme.Color("$cyan")
				indicatorFg = theme.Color("$cyan")
			} else {
				fg = theme.Color("$fg1")
				indicatorFg = theme.Color("$gray")
			}
			indicator = "▍"
		} else {
			fg = theme.Color("$fg0")
			font = ""
			indicator = " "
			indicatorFg = theme.Color("$gray")
		}

		// Left-side border indicator — spans all rows.
		r.Set(indicatorFg, bg, "")
		for row := 0; row < h; row++ {
			r.Put(x, y+row, indicator)
		}

		// Padding column.
		r.Set("", bg, "")
		r.Fill(x+borderW, y, padW, h, " ")

		// Text area: name row (bold/colored), hex row, empty padding row.
		textX := x + borderW + padW
		r.Set(fg, bg, font)
		r.Text(textX, y, item.name, textW)
		r.Set(fg, bg, "")
		r.Text(textX, y+1, item.hex, textW)
		r.Fill(textX, y+2, textW, 1, " ")

		// Color swatch — 2 rows tall, right-aligned.
		swatchX := x + borderW + padW + textW
		r.Set("", item.hex, "")
		r.Fill(swatchX, y, previewW, 2, " ")
		// Clear the third row of the swatch area.
		r.Set("", bg, "")
		r.Fill(swatchX, y+2, previewW, 1, " ")
	}

	deck = NewDeck("deck-demo", "", renderFn, 3)
	deck.Set(items)
	// Wrap in a non-focusable Flex so the left/right padding is stable and
	// unaffected by the deck's own focus state changing its style.
	builder.VFlex("deck-wrapper", Stretch, 0).Padding(0, 1).
		Static("deck-title", "Color constants from the Tokyo Night theme").Padding(0, 0, 1, 0).
		Add(deck).Hint(0, -1).
		End()
}

func custom() Widget {
	result := NewCustom("custom", "", func(widget Widget, r *Renderer) {
		_, _, width, height := widget.Content()
		for x := 10; x < width; x += 10 {
			for y := 10; y < height; y += 10 {
				r.Put(x, y, "*")
			}
		}
	})
	result.SetStyle("", NewStyle().WithColors("green", "black").WithMargin(0).WithPadding(0))
	result.SetHint(200, 100)
	return result
}

// Digits demo
func digits(builder *Builder) {
	builder.VFlex("digits-demo", Stretch, 1).Padding(1).
		Static("digits-title", "Digits Widget Demo").Padding(0, 0, 1, 0).
		HFlex("digits-content", Center, 1).
		Digits("digits", "12:34").
		End().
		Static("digits-info", "Large ASCII art-style digits using Unicode box-drawing characters.").Padding(1, 0, 0, 0).
		End()
}

func dropdown(builder *Builder) {
	builder.VFlex("select-demo", Start, 1).Padding(1, 2).
		Static("select-title", "Select Widget Demo").Padding(0, 0, 1, 0).
		Static("select-info", "Select is a dropdown selection widget.").Padding(0, 0, 1, 0).
		Select("s1", "f", "Female", "m", "Male", "d", "Diverse").
		Static("select-status", "You selected: ").Padding(1, 0, 0, 0).
		End()
}

// Editor demo
func editor(builder *Builder) {
	builder.Editor("editor-demo").Hint(0, -1).Padding(1)
	if ed := Find(builder.Container(), "editor-demo"); ed != nil {
		if editor, ok := ed.(*Editor); ok {
			editor.ShowLineNumbers(true)
			editor.Load("This is a sample text.\nYou can edit me!\n\nPress Tab to insert tabs,\nBackspace to delete,\nand arrow keys to navigate.")
		}
	}
}

func filterDemo(builder *Builder) {
	items := []string{
		"Go", "Rust", "TypeScript", "Python", "Kotlin", "Swift", "Zig",
		"C", "C++", "C#", "Java", "Scala", "Haskell", "Erlang", "Elixir",
		"Ruby", "PHP", "Dart", "Lua", "Julia", "R", "MATLAB", "Clojure",
	}
	builder.VFlex("filter-demo", Stretch, 1).Padding(1, 2).
		Static("filter-title", "Filter Widget Demo").Padding(0, 0, 1, 0).
		Static("filter-desc", "Type to filter the list below. Ghost text suggests the first prefix match.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Filter("demo-filter").
		List("filter-list", items...).Hint(0, -1).
		End()
	container := builder.Container()
	filter := Find(container, "demo-filter").(*Filter)
	list := Find(container, "filter-list").(*List)
	filter.Bind(list)
}

func form(builder *Builder) {
	data := struct {
		Database string `width:"40"`
		Username string `width:"20"`
		Password string `control:"password" width:"20" line:"1"`
	}{
		Database: "sqlite:mem",
		Username: "admin",
		Password: "secret",
	}

	user := struct {
		ID         string `readonly:"true"`
		Login      string `label:"Login Name:" width:"40"`
		Name       string `width:"40"`
		Sex        string `control:"select" options:",n/a,m,Male,f,Female,d,Diverse"`
		Department string `width:"40"`
		Email      string `label:"E-Mail-Address" width:"40"`
		Phone      string `label:"Phone Number" width:"40"`
		Mobile     string `label:"Mobile Phone" width:"40"`
		Password   string `control:"password" width:"40"`
		Temporary  bool   `label:"Temporary"`
		Pending    bool   `label:"Pending"`
		Active     bool   `label:"Active"`
		Fixed      bool   `label:"Fixed" readOnly:"true"`
	}{ID: "JD", Name: "John Doe", Sex: "m"}

	builder.VFlex("form-demo", Start, 1).Margin(2).Border("", "round").Padding(2).
		Form("form", "Connect", &data).
		Group("form-group", "", "", false, 1).Border("", "round").
		End().
		End().
		Form("form2", "User", &user).
		Group("form-group-2", "user", "", true, 1).Border("", "round").
		End().
		End().
		HFlex("form-buttons", Start, 1).Margin(1).
		Button("save-button", "Save").
		Static("info-label", "Info").
		End().
		End()

	builder.Find("save-button").On(EvtActivate, func(widget Widget, _ Event, _ ...any) bool {
		Update(FindRoot(widget), "info-label", "Activate "+time.Now().String())
		text, _ := json.Marshal(user)
		widget.Log(widget, Debug, string(text))
		return true
	})
}

// Grid demo
func grid(builder *Builder) {
	builder.Grid("grid-demo", 4, 4, true).Margin(1).Border("", "round").
		Cell(0, 0, 4, 1).Static("", "First row, spans 4 columns").
		Cell(0, 1, 1, 3).Static("", "Spans 3 rows").
		Cell(2, 2, 2, 2).Static("", "2 x 2").
		End()
}

// Progress demo
func marqueeDemo(builder *Builder) {
	builder.VFlex("marquee-demo", Stretch, 1).Padding(1, 2).
		Static("marquee-title", "Marquee Widget Demo").Padding(0, 0, 1, 0).
		Static("marquee-desc", "Text wider than the widget scrolls continuously. Hover to pause.").Padding(0, 0, 1, 0).
		Marquee("marquee-ticker").Hint(-1, 1).
		Spacer().Hint(-1, 0).
		HFlex("marquee-controls", Center, 4).Padding(1, 0, 0, 0).
		Checkbox("marquee-running", "Running", true).
		End().
		End()

	pane := builder.Find("marquee-demo").(Container)
	m := Find(pane, "marquee-ticker").(*Marquee)
	m.SetText("Status: All systems operational.  CPU 4%  MEM 1.2 GB  NET ↑ 0.8 MB/s ↓ 2.1 MB/s  DISK 42%  TEMP 38°C  UPTIME 14d 7h")

	Find(pane, "marquee-running").On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if v, ok := data[0].(bool); ok {
			if v {
				m.Start(80 * time.Millisecond)
			} else {
				m.Stop()
			}
		}
		return true
	})

	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		m.Start(80 * time.Millisecond)
		return true
	})

	pane.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		m.Stop()
		return true
	})
}

func shimmerDemo(builder *Builder) {
	builder.VFlex("shimmer-demo", Stretch, 1).Padding(1, 2).
		Static("shimmer-title", "Shimmer Widget Demo").Padding(0, 0, 1, 0).
		Static("shimmer-stepped-label", "Stepped edge:").Padding(0, 0, 0, 0).
		Shimmer("shimmer-stepped").Hint(-1, 1).
		Spacer().Size(0, 1).
		Static("shimmer-gradient-label", "Cosine gradient:").Padding(0, 0, 0, 0).
		Shimmer("shimmer-gradient").Hint(-1, 1).
		Spacer().Size(0, 1).
		Static("shimmer-multi-label", "Multi-line (gradient):").Padding(0, 0, 0, 0).
		Shimmer("shimmer-multi").Hint(-1, 3).
		Spacer().Hint(-1, 0).
		HFlex("shimmer-controls", Center, 4).Padding(1, 0, 0, 0).
		Checkbox("shimmer-running", "Running", true).
		End().
		End()

	pane := builder.Find("shimmer-demo").(Container)

	s1 := Find(pane, "shimmer-stepped").(*Shimmer)
	s1.SetText("Analysing codebase…  Status: all systems operational.")
	s1.SetBandWidth(10).SetEdgeWidth(5)

	s2 := Find(pane, "shimmer-gradient").(*Shimmer)
	s2.SetText("Analysing codebase…  Status: all systems operational.")
	s2.SetBandWidth(10).SetEdgeWidth(5).SetGradient(true)

	s3 := Find(pane, "shimmer-multi").(*Shimmer)
	s3.SetText("Searching for references…\nProcessing matched files…\nUpdating cross-references…")
	s3.SetBandWidth(10).SetEdgeWidth(5).SetGradient(true)

	start := func() {
		s1.Start(40 * time.Millisecond)
		s2.Start(40 * time.Millisecond)
		s3.Start(40 * time.Millisecond)
	}
	stop := func() {
		s1.Stop()
		s2.Stop()
		s3.Stop()
	}

	Find(pane, "shimmer-running").On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if v, ok := data[0].(bool); ok {
			if v {
				start()
			} else {
				stop()
			}
		}
		return true
	})

	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		start()
		return true
	})

	pane.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		stop()
		return true
	})
}

func progress(builder *Builder) {
	builder.VFlex("progress-demo", Stretch, 1).Padding(1).
		Static("progress-title", "Progress Widget Demo").Padding(0, 0, 1, 0).
		VFlex("progress-content", Stretch, 1)
	// Indeterminate progress
	pIndet := NewProgress("progress-indet", "", true)
	builder.Add(pIndet)
	builder.Spacer().Size(0, 1)
	// Determinate: 25%
	p25 := NewProgress("progress-25", "", true)
	p25.SetTotal(100)
	p25.Set(25)
	builder.Add(p25)
	builder.Spacer().Size(0, 1)
	// 50%
	p50 := NewProgress("progress-50", "", true)
	p50.SetTotal(100)
	p50.Set(50)
	builder.Add(p50)
	builder.Spacer().Size(0, 1)
	// 75%
	p75 := NewProgress("progress-75", "", true)
	p75.SetTotal(100)
	p75.Set(75)
	builder.Add(p75)
	builder.Spacer().Size(0, 1)
	// 100%
	p100 := NewProgress("progress-full", "", true)
	p100.SetTotal(100)
	p100.Set(100)
	builder.Add(p100)
	builder.End().
		Static("progress-info", "Progress bars support determinate (with total>0) and indeterminate (total=0) modes. Use SetTotal/SetValue to control.").Padding(1, 0, 0, 0).
		End()
}

// Scanner demo
func scanner(builder *Builder) {
	builder.VFlex("scanner-container", Stretch, 1).Padding(1).
		Static("scanner-title", "Scanner Widget Demo").Padding(0, 0, 1, 0).
		Static("scanner-info", "Back-and-forth scanning animation with fading trail.").Padding(0, 0, 1, 0).
		VFlex("scanner-flex", Center, 1).
		Scanner("scanner-blocks", 12, "blocks").
		Scanner("scanner-circles", 12, "circles").
		Scanner("scanner-diamonds", 12, "diamonds").
		End().
		Static("scanner-note", "Scanner uses a dimmed trail and cycles: forward → hold → backward → hold").Padding(1, 0, 0, 0).
		End()

	container := builder.Find("scanner-container").(Container)
	container.On(EvtShow, func(_ Widget, event Event, data ...any) bool {
		container.Log(container, Debug, "Scanner panel shown")
		for _, scanner := range FindAll[*Scanner](container) {
			scanner.Start(50 * time.Millisecond)
		}
		return true
	})

	container.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		container.Log(container, Debug, "Scanner panel hidden")
		for _, scanner := range FindAll[*Scanner](container) {
			scanner.Stop()
		}
		return true
	})
}

// Spinner demo
func spinner(builder *Builder) {
	builder.Box("spinner-demo", "Spinner").Border("", "round").Margin(1).Padding(1, 5).
		HFlex("spinner-flex", Start, 2).
		Spinner("spinner", Spinners["bar"]).
		Spinner("spinner", Spinners["dot"]).
		Spinner("spinner", Spinners["dots"]).
		Spinner("spinner", Spinners["arrow"]).
		Spinner("spinner", Spinners["circle"]).
		Spinner("spinner", Spinners["bounce"]).
		Spinner("spinner", Spinners["braille"]).
		End().
		End()

	container := builder.Find("spinner-demo").(Container)
	container.On(EvtShow, func(_ Widget, event Event, data ...any) bool {
		for _, spinner := range FindAll[*Spinner](container) {
			spinner.Start(100 * time.Millisecond)
		}
		return true
	})

	container.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		for _, spinner := range FindAll[*Spinner](container) {
			spinner.Stop()
		}
		return true
	})
}

const styledDemoText = `# Styled Widget

The **Styled** widget renders a subset of Markdown with word wrapping and inline styles. It supports all common block types shown below.

## Inline Styles

Plain text sits alongside *italic*, **bold**, __underlined__, ~~strikethrough~~ and ` + "`" + `inline code` + "`" + `. Styles can be **combined: *bold and italic* works** just fine.

## Paragraphs

Paragraphs are separated by blank lines and their text is word-wrapped to the available width. A very long paragraph like this one will wrap gracefully across as many rows as needed without truncating any content.

## Unordered Lists

- First item in the list
- Second item, which is intentionally a bit longer to demonstrate that continuation lines are indented to align with the text rather than the bullet
- Third item

## Ordered Lists

1. Download the archive
2. Extract and ` + "`" + `cd` + "`" + ` into the directory
3. Run ` + "`" + `go build ./...` + "`" + ` to compile
4. Start the binary with your preferred flags

## Code Block

` + "```" + `
func NewStyled(id, class, text string) *Styled {
    s := &Styled{Component: Component{id: id, class: class}}
    s.SetFlag(FlagFocusable, true)
    s.Set(text)
    OnKey(s, s.handleKey)
    return s
}
` + "```" + `

### Sub-heading (h3)

H3 headings use __underlined bold__ text. Use them for sections within an h2 group.

## Blockquotes

> This is a blockquote. It is rendered with a left border and muted colours. Multiple lines are wrapped and joined into a single block.

> A second blockquote, separated by a blank line.

## Horizontal Rules

Text above the rule.

---

Text below the rule.

## Nested Lists

- Top-level item
  - Nested item (depth 1)
  - Another nested item
    - Deeply nested (depth 2)
- Back at top level

1. First step
2. Second step
   - Sub-point A
   - Sub-point B
3. Third step

## Task Lists

- [x] Design the API
- [x] Write the parser
- [ ] Add tests
- [ ] Update documentation

## Scrolling

Use **↑ ↓** to scroll one line, **PgUp PgDn** for page scrolling, and **Home End** to jump to the top or bottom.`

// Styled text demo
func styled(builder *Builder) {
	builder.
		VFlex("styled-pane", Stretch, 0).Hint(0, -1).
		Styled("styled-demo", styledDemoText).Hint(0, -1).
		Shortcuts("styled-shortcuts", "↑↓", "scroll", "PgUp PgDn", "page", "Home End", "top/bottom").
		End()
}

// Table demo
func table(builder *Builder) {
	headers := []string{
		"First name", "Last name", "Street address", "ZIP", "City", "State", "Country", "Phone", "E-Mail", "Date of Birth", "Age", "Place of Birth", "Income", "SSN", "Sex",
	}
	data := people(100)
	builder.Table("table-demo", NewArrayTableProvider(headers, data), false).Hint(0, -1)
}

// Tabs demo
func tabs(builder *Builder) {
	builder.VFlex("tabs-demo", Stretch, 1).Padding(1, 2).
		Tabs("tabs", "First", "Second", "Third", "Fourth").
		End()
}

// Terminal demo — feeds representative ANSI/VT sequences into a Terminal widget.
func terminalDemo(builder *Builder) {
	term := NewTerminal("terminal-demo", "")
	term.SetHint(0, -1)

	builder.VFlex("terminal-pane", Stretch, 0).Hint(0, -1).Padding(0, 1).
		Static("terminal-title", "Terminal Widget Demo").Padding(0, 0, 1, 0).
		Add(term).Hint(0, -1).
		End()

	// Re-render on each visit so the content fits the current dimensions.
	pane := builder.Find("terminal-pane").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		term.Clear()
		writeTerminalDemo(term)
		return true
	})
}

// writeTerminalDemo pipes a sequence of ANSI escape sequences into t to
// showcase the terminal widget's rendering capabilities.
func writeTerminalDemo(t *Terminal) {
	w := func(s string) { t.Write([]byte(s)) }

	// ---- Title ----
	w("\033[1;37mANSI / VT Terminal Demo\033[0m\r\n")
	w("\033[2m─────────────────────────────────────────────────────\033[0m\r\n\r\n")

	// ---- Text attributes ----
	w("\033[1mBold\033[0m  ")
	w("\033[2mDim\033[0m  ")
	w("\033[3mItalic\033[0m  ")
	w("\033[9mStrikethrough\033[0m  ")
	w("\033[5mBlink\033[0m\r\n\r\n")

	// ---- Underline styles ----
	w("Underline styles: ")
	w("\033[4mSingle\033[0m  ")
	w("\033[21mDouble\033[0m  ")
	w("\033[4:3mCurly\033[0m  ")
	w("\033[4:4mDotted\033[0m  ")
	w("\033[4:5mDashed\033[0m\r\n\r\n")

	// ---- Standard 16 ANSI colours ----
	w("Standard colours:  ")
	for i := 0; i < 8; i++ {
		w(fmt.Sprintf("\033[%dm  \033[0m", 40+i))
	}
	w("\r\n                   ")
	for i := 0; i < 8; i++ {
		w(fmt.Sprintf("\033[%dm  \033[0m", 100+i))
	}
	w("\r\n\r\n")

	// ---- 256-colour palette strip ----
	w("256-colour palette:\r\n")
	for row := 0; row < 4; row++ {
		w("  ")
		for col := 0; col < 36; col++ {
			idx := 16 + row*36 + col
			if idx > 231 {
				break
			}
			w(fmt.Sprintf("\033[48;5;%dm  \033[0m", idx))
		}
		w("\r\n")
	}
	// Greyscale ramp
	w("  ")
	for i := 232; i <= 255; i++ {
		w(fmt.Sprintf("\033[48;5;%dm  \033[0m", i))
	}
	w("\r\n\r\n")

	// ---- True colour gradient ----
	w("True colour: ")
	steps := 48
	for i := 0; i < steps; i++ {
		r := 255 * i / (steps - 1)
		g := 128
		b := 255 - r
		w(fmt.Sprintf("\033[48;2;%d;%d;%dm \033[0m", r, g, b))
	}
	w("\r\n\r\n")

	// ---- Underline colour ----
	w("\033[4:3m\033[58;2;255;100;0mCurly underline in orange\033[0m\r\n\r\n")

	// ---- Box-drawing ----
	w("Box-drawing: ┌──────────┐\r\n")
	w("             │  \033[1;32mHello!\033[0m  │\r\n")
	w("             └──────────┘\r\n\r\n")

	// ---- Reverse video ----
	w("\033[7m Reverse video \033[0m\r\n\r\n")

	// ---- OSC title (not visible but exercises the parser) ----
	w("\033]0;Terminal Widget Demo\007")
}

func typeaheadDemo(builder *Builder) {
	languages := []string{
		"Ada", "Clojure", "C", "C++", "C#", "Crystal", "D", "Dart", "Elixir",
		"Elm", "Erlang", "F#", "Go", "Groovy", "Haskell", "Java", "JavaScript",
		"Julia", "Kotlin", "Lisp", "Lua", "Nim", "OCaml", "Pascal", "Perl",
		"PHP", "Python", "R", "Ruby", "Rust", "Scala", "Scheme", "Swift",
		"TypeScript", "Zig",
	}
	countries := []string{
		"Afghanistan", "Albania", "Algeria", "Argentina", "Australia", "Austria",
		"Belgium", "Bolivia", "Brazil", "Bulgaria", "Canada", "Chile", "China",
		"Colombia", "Croatia", "Czech Republic", "Denmark", "Egypt", "Finland",
		"France", "Germany", "Greece", "Hungary", "India", "Indonesia", "Iran",
		"Iraq", "Ireland", "Israel", "Italy", "Japan", "Jordan", "Kenya",
		"Malaysia", "Mexico", "Morocco", "Netherlands", "New Zealand", "Nigeria",
		"Norway", "Pakistan", "Peru", "Philippines", "Poland", "Portugal",
		"Romania", "Russia", "Saudi Arabia", "Serbia", "Singapore", "South Africa",
		"South Korea", "Spain", "Sweden", "Switzerland", "Thailand", "Turkey",
		"Ukraine", "United Kingdom", "United States", "Vietnam",
	}

	builder.VFlex("typeahead-demo", Stretch, 1).Padding(1, 2).
		Static("typeahead-title", "Typeahead Widget Demo").Padding(0, 0, 1, 0).
		Static("typeahead-desc", "Type to see inline ghost-text completions. Tab or → accepts.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Static("", "Programming language:").
		Typeahead("ta-lang", "", "e.g. Go, Rust, Python…").
		Static("", "Country:").
		Typeahead("ta-country", "", "e.g. Germany, Japan…").
		Static("ta-accepted", "").Padding(1, 0, 0, 0).
		End()

	container := builder.Container()

	langTA := Find(container, "ta-lang").(*Typeahead)
	langTA.SetSuggest(Suggest(languages))
	langTA.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if s, ok := data[0].(string); ok {
			if label, ok := Find(container, "ta-accepted").(*Static); ok {
				label.Set("Accepted: " + s)
			}
		}
		return true
	})

	countryTA := Find(container, "ta-country").(*Typeahead)
	countryTA.SetSuggest(Suggest(countries))
	countryTA.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if s, ok := data[0].(string); ok {
			if label, ok := Find(container, "ta-accepted").(*Static); ok {
				label.Set("Accepted: " + s)
			}
		}
		return true
	})
}

func tilesDemo(builder *Builder) {
	type card struct {
		name  string
		icon  string
		color string
	}
	cards := []card{
		{"Dashboard", "◈", "$blue"},
		{"Analytics", "▦", "$green"},
		{"Reports", "▤", "$yellow"},
		{"Settings", "⚙", "$fg2"},
		{"Users", "◉", "$blue"},
		{"Billing", "◎", "$orange"},
		{"Security", "◆", "$red"},
		{"Integrations", "⬡", "$cyan"},
		{"Logs", "≡", "$fg2"},
		{"API Keys", "◆", "$purple"},
		{"Webhooks", "◈", "$blue"},
		{"Audit Trail", "▤", "$yellow"},
	}
	items := make([]any, len(cards))
	for i, c := range cards {
		items[i] = c
	}

	// tileWidth=14, tileHeight=4: row 0 blank, row 1 icon, row 2 name, row 3 blank.
	renderCard := func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool) {
		c := data.(card)
		bg := "$bg2"
		fg := "$fg1"
		if selected && focused {
			bg = "$blue"
			fg = "$bg0"
		} else if selected {
			bg = "$bg3"
			fg = "$fg0"
		}
		r.Set(fg, bg, "")
		r.Fill(x, y, w, h, " ")
		// Icon on row 1, centred horizontally.
		iconX := x + max(0, (w-1)/2)
		r.Set(c.color, bg, "bold")
		r.Put(iconX, y+1, c.icon)
		// Name on row 2, centred horizontally.
		nameRunes := []rune(c.name)
		nameX := x + max(0, (w-len(nameRunes))/2)
		r.Set(fg, bg, "")
		r.Text(nameX, y+2, c.name, w-(nameX-x))
	}

	// No vertical padding on the outer Flex — every row counts.
	builder.VFlex("tiles-demo", Stretch, 0).Padding(0, 2).
		Static("tiles-title", "Tiles  ←→↑↓ navigate · Enter activate").Padding(0, 0, 0, 0).
		Tiles("tiles-grid", renderCard, 14, 4).Hint(-1, -1).
		End()

	pane := builder.Find("tiles-demo").(Container)
	grid := Find(pane, "tiles-grid").(*Tiles)
	grid.SetItems(items)
}

func treeFSDemo(builder *Builder) {
	var tfs *TreeFS

	tfs = NewTreeFS("tree-fs", "", ".", false)

	builder.VFlex("tree-fs-demo", Stretch, 0).
		// Toolbar: Up button + current root path
		HFlex("tree-fs-toolbar", Center, 1).Padding(0, 1).
		Button("tree-fs-up", "↑ Up").
		Static("tree-fs-path", tfs.RootPath()).Padding(0, 1).
		End().
		// The tree itself, takes all remaining height
		Add(tfs.Tree).Hint(0, -1).
		// Status bar showing the highlighted path
		Static("tree-fs-selected", "").Padding(0, 1).
		End()

	container := builder.Container()

	// Up button: navigate to the parent directory
	builder.Find("tree-fs-up").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		parent := filepath.Dir(tfs.RootPath())
		if parent == tfs.RootPath() {
			return true // already at filesystem root
		}
		tfs.SetRoot(parent)
		if label, ok := Find(container, "tree-fs-path").(*Static); ok {
			label.Set(tfs.RootPath())
		}
		if label, ok := Find(container, "tree-fs-selected").(*Static); ok {
			label.Set("")
		}
		return true
	})

	// Update the status bar whenever the highlighted node changes
	tfs.Tree.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 0 {
			return true
		}
		node, ok := data[0].(*TreeNode)
		if !ok {
			return true
		}
		if label, ok := Find(container, "tree-fs-selected").(*Static); ok {
			label.Set(node.Data().(string))
		}
		return true
	})
}

// Typewriter demo — animated character-by-character reveal.
func typewriterDemo(builder *Builder) {
	phrases := []string{
		"Initialising subsystems…",
		"Loading configuration…",
		"Connecting to services…",
		"All systems operational.",
	}
	idx := 0

	builder.VFlex("tw-demo", Stretch, 1).Padding(1, 2).
		Static("tw-title", "Typewriter Widget Demo").Padding(0, 0, 1, 0).
		Static("tw-desc", "Text is revealed character by character with a blinking cursor.").Padding(0, 0, 1, 0).
		Typewriter("tw").
		Spacer().Hint(-1, 0).
		HFlex("tw-controls", Center, 4).Padding(1, 0, 0, 0).
		Checkbox("tw-repeat", "Repeat", false).
		Checkbox("tw-cursor", "Show cursor", true).
		Button("tw-restart", "Restart").
		End().
		End()

	pane := builder.Find("tw-demo").(Container)

	tw := Find(pane, "tw").(*Typewriter)
	tw.SetText(phrases[idx])

	startNext := func() {
		idx = (idx + 1) % len(phrases)
		tw.SetText(phrases[idx])
		tw.Start(30 * time.Millisecond)
	}

	tw.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		startNext()
		return true
	})

	Find(pane, "tw-repeat").On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if v, ok := data[0].(bool); ok {
			tw.SetRepeat(v)
		}
		return true
	})

	Find(pane, "tw-cursor").On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if v, ok := data[0].(bool); ok {
			tw.SetCursor(v)
		}
		return true
	})

	Find(pane, "tw-restart").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		tw.Stop()
		tw.Reset()
		tw.Start(30 * time.Millisecond)
		return true
	})

	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		tw.Reset()
		tw.Start(30 * time.Millisecond)
		return true
	})

	pane.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		tw.Stop()
		return true
	})
}

// Value demo — two groups of widgets sharing a reactive Value.
func valueDemo(builder *Builder) {
	builder.VFlex("value-demo", Stretch, 1).Padding(1, 2).
		Static("value-title", "Value Demo").Padding(0, 0, 1, 0).
		Static("value-info", "Widgets sharing the same Value stay in sync automatically.").Padding(0, 0, 1, 0)

	// --- Group 1: two checkboxes sharing a Value[bool] ---
	builder.Static("value-bool-label", "Shared bool — toggle either checkbox:").Padding(0, 0, 0, 0)
	cb1 := NewCheckbox("val-cb1", "", "Checkbox A", false)
	cb2 := NewCheckbox("val-cb2", "", "Checkbox B", false)
	boolVal := NewValue(false)
	boolVal.Bind(cb1).Bind(cb2)
	boolVal.Observe(cb1)
	boolVal.Observe(cb2)
	builder.Add(cb1)
	builder.Add(cb2)

	builder.Spacer().Size(0, 1)

	// --- Group 2: Input and Digits sharing a Value[string] ---
	builder.Static("value-str-label", "Shared string — type in the input to update Digits:").Padding(0, 0, 0, 0)
	strInput := NewInput("val-str-input", "")
	strInput.Set("12:34")
	digits := NewDigits("val-digits", "", "12:34")
	strVal := NewValue("12:34")
	strVal.Bind(strInput).Bind(digits)
	strVal.Observe(strInput)
	builder.Add(strInput)
	builder.Spacer().Size(0, 1)
	builder.HFlex("val-digits-row", Center, 0).
		Add(digits).
		End()

	builder.Spacer().Size(0, 1)

	// --- Group 3: Input and Progress sharing a Value[int] ---
	builder.Static("value-int-label", "Shared int — type 0–100 in the input to move the progress bar:").Padding(0, 0, 0, 0)
	intInput := NewInput("val-int-input", "")
	intInput.Set("50")
	prog := NewProgress("val-progress", "", true)
	prog.SetTotal(100)
	intVal := NewValue(50)
	intVal.Bind(prog)
	intVal.Observe(intInput, func(v any) (int, bool) {
		s, ok := v.(string)
		if !ok {
			return 0, false
		}
		n := 0
		_, err := fmt.Sscanf(s, "%d", &n)
		return n, err == nil
	})
	builder.Add(intInput)
	builder.Spacer().Size(0, 1)
	builder.Add(prog)

	builder.End()
}

func viewport(builder *Builder) {
	builder.VFlex("viewport-demo", Stretch, 1).Padding(1, 2).
		Static("viewport-title", "Viewport Demo").Padding(0, 0, 1, 0).
		Static("viewport-info", "A scrollable viewport of the inside widget.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Viewport("viewport", "Viewport").Border("thin").Hint(-1, -1).
		Add(custom()).
		End().
		End()
}

// Sparkline demo — shows all scale modes and multi-row rendering with live data.
func sparklineDemo(builder *Builder) {
	// Pre-seed each sparkline with 40 points so they aren't empty on first show.
	seed := func(fn func(i int) float64) []float64 {
		vs := make([]float64, 40)
		for i := range vs {
			vs[i] = fn(i)
		}
		return vs
	}

	sineNoise := func(i int) float64 {
		return math.Sin(float64(i)*0.3) + rand.Float64()*0.3 - 0.15
	}
	sineAbs := func(i int) float64 {
		return (math.Sin(float64(i)*0.2) + 1.0) / 2.0
	}

	builder.VFlex("sparkline-demo", Stretch, 1).Padding(1, 2).
		Static("sp-title", "Sparkline Widget Demo").Padding(0, 0, 1, 0).
		Static("sp-desc", "Live data — each chart updates every 100 ms.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		// ---- Row 1: Relative scale h=1 ----
		Static("sp-lbl-rel", "Relative scale (h=1) — shape follows data:").Padding(1, 0, 0, 0).
		Sparkline("sp-rel").Hint(-1, 1).
		// ---- Row 2: Absolute scale h=1 ----
		Static("sp-lbl-abs", "Absolute scale (h=1, min=−1 max=+1):").Padding(1, 0, 0, 0).
		Sparkline("sp-abs").Hint(-1, 1).
		// ---- Row 3: Hard threshold h=2 ----
		Static("sp-lbl-thr", "Hard threshold at 0.65 (h=2):").Padding(1, 0, 0, 0).
		Sparkline("sp-thr").Hint(-1, 2).
		// ---- Row 4: Gradient threshold h=2 ----
		Static("sp-lbl-grad", "Gradient threshold at 0.65 (h=2):").Padding(1, 0, 0, 0).
		Sparkline("sp-grad").Hint(-1, 2).
		// ---- Row 5: Multi-row h=4 ----
		Static("sp-lbl-multi", "Multi-row (h=4, 32 levels per column):").Padding(1, 0, 0, 0).
		Sparkline("sp-multi").Hint(-1, 4).
		End()

	// Ring buffers hold the live data for each sparkline (capacity = 120).
	// Pre-seed with 40 points so the charts aren't empty on first show.
	newRB := func(fn func(i int) float64) *RingBuffer[float64] {
		rb := NewRingBuffer[float64](120)
		for _, v := range seed(fn) {
			rb.Add(v)
		}
		return rb
	}

	rbRel := newRB(sineNoise)
	rbAbs := newRB(sineNoise)
	rbThr := newRB(sineAbs)
	rbGrad := newRB(sineAbs)
	rbMulti := newRB(sineNoise)

	spRel := builder.Find("sp-rel").(*Sparkline)
	spRel.SetProvider(rbRel)

	spAbs := builder.Find("sp-abs").(*Sparkline)
	spAbs.SetAbsolute(true)
	spAbs.SetMin(-1.0)
	spAbs.SetMax(1.0)
	spAbs.SetProvider(rbAbs)

	spThr := builder.Find("sp-thr").(*Sparkline)
	spThr.SetAbsolute(true)
	spThr.SetMin(0.0)
	spThr.SetMax(1.0)
	spThr.SetThreshold(0.65)
	spThr.SetProvider(rbThr)

	spGrad := builder.Find("sp-grad").(*Sparkline)
	spGrad.SetAbsolute(true)
	spGrad.SetMin(0.0)
	spGrad.SetMax(1.0)
	spGrad.SetThreshold(0.65)
	spGrad.SetGradient(true)
	spGrad.SetProvider(rbGrad)

	spMulti := builder.Find("sp-multi").(*Sparkline)
	spMulti.SetAbsolute(true)
	spMulti.SetMin(-1.0)
	spMulti.SetMax(1.0)
	spMulti.SetProvider(rbMulti)

	container := builder.Find("sparkline-demo").(Container)

	var stop chan struct{}
	var phase float64

	container.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		stop = make(chan struct{})
		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					v := math.Sin(phase)*0.8 + rand.Float64()*0.35 - 0.175
					v01 := (math.Sin(phase*0.7) + 1.0) / 2.0
					phase += 0.25

					rbRel.Add(v)
					rbAbs.Add(v)
					rbThr.Add(v01)
					rbGrad.Add(v01)
					rbMulti.Add(v)

					spRel.Refresh()
					spAbs.Refresh()
					spThr.Refresh()
					spGrad.Refresh()
					spMulti.Refresh()
				}
			}
		}()
		return true
	})

	container.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		if stop != nil {
			close(stop)
			stop = nil
		}
		return true
	})
}

// Heatmap demo — shows a 24×7 activity grid (hour × weekday) with live random data.
func heatmapDemo(builder *Builder) {
	const rows, cols = 24, 7

	builder.VFlex("heatmap-demo", Stretch, 1).Padding(1, 2).
		Static("hm-title", "Heatmap Widget Demo").Padding(0, 0, 1, 0).
		Heatmap("hm", rows, cols).Hint(-1, -1).
		End()

	hm := builder.Find("hm").(*Heatmap)
	hm.SetCellWidth(2)
	hm.SetColLabels([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"})
	rowLabels := make([]string, rows)
	for i := range rowLabels {
		rowLabels[i] = fmt.Sprintf("%2dh", i)
	}
	hm.SetRowLabels(rowLabels)

	// Seed with random-looking activity data that peaks during business hours.
	rng := [4]uint64{0x243f6a88, 0x85a308d3, 0x13198a2e, 0x03707344}
	next := func() float64 {
		x := rng[0] ^ rng[0]<<13
		rng[0] = rng[1]
		rng[1] = rng[2]
		rng[2] = rng[3]
		rng[3] = x ^ x>>7 ^ rng[3] ^ rng[3]>>19
		return float64(rng[3]&0xffff) / 0xffff
	}
	data := make([][]float64, rows)
	for r := range data {
		data[r] = make([]float64, cols)
		for c := range data[r] {
			// Business hours (8–18) on weekdays (0–4) get higher baseline.
			base := 0.1
			if r >= 8 && r < 18 && c < 5 {
				base = 0.5
			}
			data[r][c] = base + next()*0.5
		}
	}
	hm.SetAll(data)

	var stop chan struct{}
	container := builder.Find("heatmap-demo").(Container)
	container.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		stop = make(chan struct{})
		go func() {
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					r := int(next()*float64(rows)) % rows
					c := int(next()*float64(cols)) % cols
					base := 0.1
					if r >= 8 && r < 18 && c < 5 {
						base = 0.5
					}
					hm.SetValue(r, c, base+next()*0.5)
				}
			}
		}()
		return true
	})
	container.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		if stop != nil {
			close(stop)
			stop = nil
		}
		return true
	})
}

// barChartDemo demonstrates the BarChart widget with stacked series, a
// category label row, y-axis, grid, and a legend. Two views are shown:
// a vertical stacked bar chart and a horizontal variant below it.
func barChartDemo(builder *Builder) {
	categories := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	series := []BarSeries{
		{Label: "Revenue", Values: []float64{42, 55, 61, 49, 70, 88, 95, 83, 74, 66, 52, 78}},
		{Label: "Costs", Values: []float64{30, 32, 35, 33, 38, 41, 44, 40, 37, 35, 30, 36}},
		{Label: "Profit", Values: []float64{12, 23, 26, 16, 32, 47, 51, 43, 37, 31, 22, 42}},
	}

	builder.VFlex("bar-chart-demo", Stretch, 1).Padding(1, 2).
		Static("bc-title", "Bar Chart Widget Demo").Padding(0, 0, 1, 0).
		Static("bc-desc", "Stacked bar chart with y-axis, grid, and legend. Use ←→ to navigate categories.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Static("bc-vert-label", "Vertical (stacked):").Padding(0, 0, 0, 0).
		BarChart("bc-vert").Hint(-1, 16).
		Static("bc-horiz-label", "Horizontal:").Padding(1, 0, 0, 0).
		BarChart("bc-horiz").Hint(-1, 12).
		End()

	bcVert := builder.Find("bc-vert").(*BarChart)
	bcVert.SetCategories(categories)
	bcVert.SetSeries(series)
	bcVert.SetShowValues(true)

	bcHoriz := builder.Find("bc-horiz").(*BarChart)
	bcHoriz.SetCategories(categories)
	bcHoriz.SetSeries(series)
	bcHoriz.SetHorizontal(true)
	bcHoriz.SetShowValues(false)
}

// breadcrumbDemo demonstrates the Breadcrumb widget with Push/Pop controls
// and an EvtActivate handler that truncates the path on segment click.
func breadcrumbDemo(builder *Builder) {
	path := []string{"Home", "Projects", "zeichenwerk", "cmd", "demo"}

	builder.VFlex("breadcrumb-demo", Stretch, 1).Padding(1, 2).
		Static("bc-title", "Breadcrumb Widget Demo").Padding(0, 0, 1, 0).
		Static("bc-desc", "Click a segment to navigate. ←→ to move, Enter to activate (truncates path).").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Static("bc-label", "Path:").Padding(0, 0, 0, 0).
		Breadcrumb("bc").Hint(-1, 1).
		HFlex("bc-controls", Center, 2).Padding(1, 0, 0, 0).
		Button("bc-push", "Push").
		Button("bc-pop", "Pop").
		Button("bc-reset", "Reset").
		End().
		Static("bc-status", "").Padding(1, 0, 0, 0).
		End()

	bc := builder.Find("bc").(*Breadcrumb)
	bc.Set(path)

	status := builder.Find("bc-status").(*Static)

	bc.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		idx := data[0].(int)
		segs := bc.Segments()
		if idx < len(segs) {
			status.Set(fmt.Sprintf("Selected: [%d] %s", idx, segs[idx]))
		}
		return true
	})

	bc.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		idx := data[0].(int)
		bc.Truncate(idx)
		segs := bc.Segments()
		if idx < len(segs) {
			status.Set(fmt.Sprintf("Activated: truncated to [%d] %s", idx, segs[idx]))
		}
		return true
	})

	pushBtn := builder.Find("bc-push").(*Button)
	pushBtn.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		n := len(bc.Segments())
		bc.Push(fmt.Sprintf("dir%d", n))
		status.Set(fmt.Sprintf("Pushed: %d segments", len(bc.Segments())))
		return true
	})

	popBtn := builder.Find("bc-pop").(*Button)
	popBtn.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		seg := bc.Pop()
		if seg != "" {
			status.Set(fmt.Sprintf("Popped: %q", seg))
		} else {
			status.Set("Nothing to pop")
		}
		return true
	})

	resetBtn := builder.Find("bc-reset").(*Button)
	resetBtn.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		bc.Set(path)
		status.Set("Reset")
		return true
	})
}

// Table demo data generation
func people(n int) [][]string {
	firstNames := []string{"John", "Jane", "Michael", "Emily", "David", "Sophia", "James", "Olivia", "Daniel", "Ava", "Liam", "Emma", "Noah", "Isabella", "Ethan", "Mia", "Lucas", "Charlotte", "Mason", "Amelia"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin"}
	streets := []string{"Maple St", "Oak Ave", "Pine Rd", "Birch Blvd", "Cedar Ln", "Spruce Ct", "Willow Way", "Elm Pl", "Aspen Dr", "Cypress St"}
	cities := []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose"}
	states := []string{"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA", "HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD", "MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ", "NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC", "SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY"}

	result := make([][]string, n)
	for i := range n {
		first := firstNames[rand.Intn(len(firstNames))]
		last := lastNames[rand.Intn(len(lastNames))]
		street := streets[rand.Intn(len(streets))]
		city := cities[rand.Intn(len(cities))]
		state := states[rand.Intn(len(states))]
		zip := fmt.Sprintf("%05d", 10000+rand.Intn(90000))
		phone := fmt.Sprintf("+1-%03d-%03d-%04d", rand.Intn(900)+100, rand.Intn(900)+100, rand.Intn(10000))
		email := strings.ToLower(fmt.Sprintf("%s.%s@example.com", first, last))
		birth := fmt.Sprintf("%04d-%02d-%02d", 1950+rand.Intn(50), rand.Intn(12)+1, rand.Intn(28)+1)
		age := rand.Intn(50) + 20
		pob := cities[rand.Intn(len(cities))]
		income := 20000 + rand.Intn(100000)
		ssn := fmt.Sprintf("%03d-%02d-%04d", rand.Intn(900)+100, rand.Intn(80)+10, rand.Intn(10000))
		sex := []string{"M", "F"}[rand.Intn(2)]

		result[i] = []string{first, last, street, zip, city, state, "USA", phone, email, birth, fmt.Sprintf("%d", age), pob, fmt.Sprintf("$%d", income), ssn, sex}
	}
	return result
}

// commandsDemo builds the Commands Palette demo pane.
// Commands are registered separately in registerCommandsDemo (called after Build).
// Bind Ctrl+K anywhere in the app to open the palette.
func commandsDemo(b *Builder) {
	b.VFlex("commands-demo", Stretch, 1).Padding(1).
		Static("commands-title", "Commands Palette").Padding(0, 0, 1, 0).
		HRule("thin").
		Static("commands-instructions",
			"Press Ctrl+K to open the commands palette.\n\n"+
				"Start typing to fuzzy-filter commands.\n"+
				"Use ↑↓ to navigate, Enter to execute, Esc to dismiss.\n\n"+
				"Registered groups: File · View · Navigation\n"+
				"Commands that modify visible state log output to the debug pane.").
		End()
}

// registerCommandsDemo registers representative commands on the UI's Commands
// singleton and is called once after Build() so the UI reference is available.
func registerCommandsDemo(ui *UI) {
	logMsg := func(msg string) {
		if t, ok := Find(ui, "debug-log").(*Text); ok {
			t.Add("Commands → " + msg)
		}
	}
	cmds := ui.Commands()
	cmds.Register("File", "New File", "Ctrl+N", func() { logMsg("New File") })
	cmds.Register("File", "Open File", "Ctrl+O", func() { logMsg("Open File") })
	cmds.Register("File", "Save File", "Ctrl+S", func() { logMsg("Save File") })
	cmds.Register("File", "Close File", "", func() { logMsg("Close File") })
	cmds.Register("View", "Toggle Theme", "", func() { logMsg("Toggle Theme") })
	cmds.Register("View", "Split Pane", "Ctrl+\\", func() { logMsg("Split Pane") })
	cmds.Register("View", "Toggle Sidebar", "", func() { logMsg("Toggle Sidebar") })
	cmds.Register("Navigation", "Go to Line", "Ctrl+G", func() { logMsg("Go to Line") })
	cmds.Register("Navigation", "Find in Files", "Ctrl+F", func() { logMsg("Find in Files") })
	cmds.Register("Navigation", "Go to Symbol", "Ctrl+R", func() { logMsg("Go to Symbol") })
}

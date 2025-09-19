package main

import (
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	. "github.com/tekugo/zeichenwerk"
)

func main() {
	ui := createUI()
	ui.Run()
}

func createUI() *UI {
	return NewBuilder(TokyoNightTheme()).
		Flex("main", "vertical", "stretch", 0).
		With(header).
		With(content).
		With(footer).
		Class("").
		Build()
}

func header(builder *Builder) {
	builder.Class("header").
		Flex("header", "horizontal", "start", 0).Padding(0, 1).Hint(0, 1).
		Label("title", "zeichenwerk", 20).Hint(20, 1).
		Label("", "Demo Application", 0).Hint(-1, 1).
		Label("time", "12:00", 0).Hint(6, 1).
		Class("").
		End()
}

func footer(builder *Builder) {
	builder.Class("footer").
		Flex("footer", "horizontal", "start", 0).Padding(0, 1).Hint(0, 1).
		Class("shortcut").Label("1", "Esc", 0).
		Class("footer").Label("2", "Close \u2502", 0).
		Class("shortcut").Label("3", "Ctrl-D", 0).
		Class("footer").Label("4", "Inspector \u2502", 0).
		Class("shortcut").Label("5", "Ctrl-Q", 0).
		Class("footer").Label("6", "Quit Application \u2502", 0).
		Class("").
		Spacer().
		End()
}

func content(builder *Builder) {
	demos := []string{"Overview", "Box", "Button", "Checkbox", "Digits", "Editor", "Flex", "Grid", "Input", "Inspector", "Label", "List", "Pop-up", "Progress Bar", "Scroller", "Spinner", "Table", "Tabs", "Theme Switch", "Gruvbox Dark Theme", "Gruvbox Light Theme", "Midnight Neon Theme", "Tokyo Night Theme", "Debug-Log"}
	builder.Grid("grid", 1, 2, true).Hint(0, -1).
		Cell(0, 0, 1, 1).
		List("demos", demos).
		Cell(1, 0, 1, 1).
		Switcher("demo").
		With(overview).
		With(box).
		With(button).
		With(checkbox).
		With(flex).
		With(label).
		With(progress).
		With(list).
		With(tabs).
		With(scroller).
		With(grid).
		With(input).
		With(table).
		With(theme).
		With(spinner).
		With(editor).
		With(digits).
		Text("debug-log", []string{}, true, 1000).
		End(). // Switcher
		End()  // Grid

	// TODO: Add functionality to set grid sizes
	grid := builder.Container().Find("grid", false)
	if grid, ok := grid.(*Grid); ok {
		grid.Columns(30, -1)

		HandleListEvent(grid, "demos", "activate", func(list *List, event string, index int) bool {
			ui := FindUI(list)

			// Stop all spinners
			for _, spinner := range FindAll[*Spinner](ui, false) {
				spinner.Stop()
			}

			switch demos[index] {
			case "Overview":
				Update(ui, "demo", "overview-demo")
			case "Box":
				Update(ui, "demo", "box-demo")
			case "Button":
				Update(ui, "demo", "button-demo")
			case "Checkbox":
				Update(ui, "demo", "checkbox-demo")
			case "Digits":
				Update(ui, "demo", "digits-demo")
			case "Editor":
				Update(ui, "demo", "editor")
			case "Flex":
				Update(ui, "demo", "flex-demo")
			case "Grid":
				Update(ui, "demo", "grid-demo")
			case "Input":
				Update(ui, "demo", "input-demo")
			case "Label":
				Update(ui, "demo", "label-demo")
			case "Progress Bar":
				Update(ui, "demo", "progress-demo")
			case "Debug-Log":
				Update(ui, "demo", "debug-log")
			case "Inspector":
				ui.Popup(-1, -1, 0, 0, NewInspector(ui).UI())
			case "List":
				Update(ui, "demo", "list-demo")
			case "Pop-up":
				ui.Popup(-1, -1, 0, 0, popup())
			case "Scroller":
				Update(ui, "demo", "scroller-demo")
			case "Spinner":
				Update(ui, "demo", "spinner-demo")
				for _, spinner := range FindAll[*Spinner](ui, false) {
					spinner.Start(100 * time.Millisecond)
				}
			case "Table":
				Update(ui, "demo", "table-demo")
			case "Tabs":
				Update(ui, "demo", "tabs-demo")
			case "Theme Switch":
				Update(ui, "demo", "theme-switch")
			case "Gruvbox Dark Theme":
				ui.SetTheme(GruvboxDarkTheme())
			case "Gruvbox Light Theme":
				ui.SetTheme(GruvboxLightTheme())
			case "Midnight Neon Theme":
				ui.SetTheme(MidnightNeonTheme())
			case "Tokyo Night Theme":
				ui.SetTheme(TokyoNightTheme())
			}
			return true
		})
	}
}

func list(builder *Builder) {
	countries := slices.Collect(maps.Keys(Countries))

	builder.Flex("list-demo", "horizontal", "stretch", 1).Padding(1).
		List("countries", countries).Border("", "round").Border(":focus", "double").Hint(-1, 0).
		List("names", names).Border("", "round").Border(":focus", "double").Hint(-1, 0).
		End()
}

func tabs(builder *Builder) {
	builder.Flex("tabs-demo", "vertical", "stretch", 1).Padding(1, 2).
		Tabs("tabs", "First", "Second", "Third", "Fourth").
		End()
}

func overview(builder *Builder) {
	builder.Flex("overview-demo", "vertical", "stretch", 0).Padding(1, 2).
		Label("welcome", "Welcome to zeichenwerk!", 0).Padding(0, 0, 1, 0).Font("", "bold").
		Label("description", "A comprehensive terminal user interface framework for Go", 0).Padding(0, 0, 1, 0).
		Separator("sep1", "thin", 0, 1).Padding(0, 0, 1, 0).
		Label("features-title", "Key Features:", 0).Font("", "underline").
		Label("feature1", "• Rich widget set (buttons, inputs, lists, tabs, etc.)", 0).Padding(1, 0, 0, 0).
		Label("feature2", "• Flexible layout system (flex, grid, box containers)", 0).
		Label("feature3", "• Event-driven architecture with keyboard and mouse support", 0).
		Label("feature4", "• Comprehensive theming and styling system", 0).
		Label("feature5", "• Built-in themes (Tokyo Night, Default)", 0).
		Label("feature6", "• Focus management and accessibility features", 0).
		Separator("sep2", "thin", 0, 1).Padding(1, 0, 1, 0).
		Label("instructions", "Use the list on the left to explore different widget demos.", 0).
		Label("navigation", "Navigation: Arrow keys, Tab, Enter, Esc", 0).Font("", "italic").
		Spacer().
		End()
}

func grid(builder *Builder) {
	builder.Grid("grid-demo", 4, 4, true).Margin(1).Border("", "round").
		Cell(0, 0, 4, 1).Label("", "First row, spans 4 columns", 0).
		Cell(0, 1, 1, 3).Label("", "Spans 3 rows", 0).
		Cell(2, 2, 2, 2).Label("", "2 x 2", 0).
		End()

	if grid, ok := builder.Container().(*Grid); ok {
		grid.Columns(0, -1, -1, -1)
	}
}

func input(builder *Builder) {
	builder.Grid("input-demo", 10, 2, false).Margin(1).
		Cell(0, 0, 1, 1).Label("", "First Name", 0).
		Cell(0, 1, 1, 1).Label("", "Last Name", 0).
		Cell(1, 0, 1, 1).Input("input-first-name", "", 40).
		Cell(1, 1, 1, 1).Input("input-last-name", "", 40).
		End()

	if grid, ok := builder.Container().Find("input-demo", false).(*Grid); ok {
		grid.Rows(1, 1, 1, 1, 1, 1, 1, 1, 1, -1)
		grid.Columns(0, -1)
	}
}

func box(builder *Builder) {
	builder.Flex("box-demo", "vertical", "stretch", 1).Padding(1).
		Label("box-title", "Box Widget Demo", 0).Padding(0, 0, 1, 0).
		Flex("box-examples", "horizontal", "stretch", 2).
		Box("simple-box", "Simple Box").Padding(1).
		Label("box-content1", "This is content inside a simple box widget.", 0).
		End().
		Box("styled-box", "Styled Box").Padding(1).Border("", "double").
		Label("box-content2", "This box has a double border style.", 0).
		End().
		Box("padded-box", "Padded Box").Padding(2).Border("", "round").
		Label("box-content3", "This box has extra padding and rounded borders.", 0).
		End().
		End().
		Label("box-info", "Boxes are containers that can hold a single child widget with optional borders and titles.", 0).Padding(1, 0, 0, 0).
		End()
}

func checkbox(builder *Builder) {
	builder.Flex("checkbox-demo", "vertical", "stretch", 1).Padding(1, 2).
		Label("checkbox-title", "Checkbox Widget Demo", 0).Padding(0, 0, 1, 0).
		Label("checkbox-info", "Checkboxes toggle between checked and unchecked states.", 0).Padding(0, 0, 1, 0).
		Separator("checkbox-sep", "thin", 0, 1).Padding(0, 0, 1, 0).
		Checkbox("cb1", "Enable notifications", false).
		Checkbox("cb2", "Remember login", true).
		Checkbox("cb3", "Auto-save documents", false).
		Checkbox("cb4", "Show hidden files", true).
		Checkbox("cb5", "I agree to the terms and conditions", false).
		Label("checkbox-status", "Toggle checkboxes with Space or Enter key!", 0).Padding(1, 0, 0, 0).
		End()

	// Add checkbox event handlers for interactivity
	container := builder.Container()
	for i := 1; i <= 5; i++ {
		cbId := fmt.Sprintf("cb%d", i)
		if cb := container.Find(cbId, false); cb != nil {
			cb.On("change", func(widget Widget, event string, data ...any) bool {
				checked := data[0].(bool)
				if statusLabel := container.Find("checkbox-status", false); statusLabel != nil {
					if label, ok := statusLabel.(*Label); ok {
						var checkboxName string
						switch widget.ID() {
						case "cb1":
							checkboxName = "Notifications"
						case "cb2":
							checkboxName = "Remember login"
						case "cb3":
							checkboxName = "Auto-save"
						case "cb4":
							checkboxName = "Show hidden"
						case "cb5":
							checkboxName = "Terms agreed"
						}
						label.SetText(fmt.Sprintf("%s: %v", checkboxName, checked))
					}
				}
				return true
			})
		}
	}
}

func button(builder *Builder) {
	builder.Flex("button-demo", "vertical", "stretch", 1).Padding(1, 2).
		Label("button-title", "Button Widget Demo", 0).Padding(0, 0, 1, 0).
		Label("button-info", "Buttons respond to Enter key, Space bar, and mouse clicks.", 0).Padding(0, 0, 1, 0).
		Separator("button-sep", "thin", 0, 1).Padding(0, 0, 1, 0).
		Flex("button-row1", "horizontal", "start", 2).
		Button("btn1", "Primary").
		Button("btn2", "Secondary").
		Button("btn3", "Action").
		End().
		Flex("button-row2", "horizontal", "start", 2).Padding(1, 0, 0, 0).
		Button("btn4", "Save").
		Button("btn5", "Cancel").
		Button("btn6", "Delete").
		End().
		Flex("button-row3", "horizontal", "start", 2).Padding(1, 0, 0, 0).
		Button("btn7", "Very Long Button Text").
		Button("btn8", "OK").
		End().
		Label("button-status", "Click any button to see it in action!", 0).Padding(1, 0, 0, 0).
		End()

	// Add button click handlers for interactivity
	container := builder.Container()
	for i := 1; i <= 8; i++ {
		btnId := fmt.Sprintf("btn%d", i)
		if btn := container.Find(btnId, false); btn != nil {
			btn.On("click", func(widget Widget, event string, data ...any) bool {
				if statusLabel := container.Find("button-status", false); statusLabel != nil {
					if label, ok := statusLabel.(*Label); ok {
						label.Text = fmt.Sprintf("Button '%s' was clicked!", widget.ID())
						label.Refresh()
					}
				}
				return true
			})
		}
	}
}

func flex(builder *Builder) {
	builder.Flex("flex-demo", "vertical", "stretch", 1).Padding(1).
		Label("flex-title", "Flex Layout Demo", 0).Padding(0, 0, 1, 0).
		Label("flex-info", "Flex containers arrange widgets horizontally or vertically with flexible sizing.", 0).Padding(0, 0, 1, 0).
		Separator("flex-sep1", "thin", 0, 1).Padding(0, 0, 1, 0).
		Label("horizontal-title", "Horizontal Flex (stretch alignment):", 0).
		Flex("horizontal-demo", "horizontal", "stretch", 1).Hint(0, 3).Border("", "thin").
		Label("h1", "Left", 0).Background("", "$blue").Foreground("", "$bg").Padding(1).
		Label("h2", "Center", 0).Background("", "$green").Foreground("", "$bg").Padding(1).
		Label("h3", "Right", 0).Background("", "$orange").Foreground("", "$bg").Padding(1).
		End().
		Label("vertical-title", "Vertical Flex (start alignment):", 0).Padding(1, 0, 0, 0).
		Flex("vertical-demo", "vertical", "start", 1).Hint(0, 6).Border("", "thin").
		Label("v1", "Top", 0).Background("", "$red").Foreground("", "$bg").Padding(0, 1).
		Label("v2", "Middle", 0).Background("", "$cyan").Foreground("", "$bg").Padding(0, 1).
		Label("v3", "Bottom", 0).Background("", "$magenta").Foreground("", "$bg").Padding(0, 1).
		End().
		End()
}

func label(builder *Builder) {
	builder.Flex("label-demo", "vertical", "stretch", 1).Padding(1, 2).
		Label("label-title", "Label Widget Demo", 0).Padding(0, 0, 1, 0).
		Label("label-info", "Labels display static text with various styling options.", 0).Padding(0, 0, 1, 0).
		Separator("label-sep", "thin", 0, 1).Padding(0, 0, 1, 0).
		Label("default-label", "Default Label", 0).
		Label("colored-label", "Colored Label", 0).Background("", "$blue").Foreground("", "$bg").Padding(0, 1).
		Label("padded-label", "Padded Label", 0).Padding(1, 2).Border("", "round").
		Label("long-label", "This is a very long label that demonstrates how text wrapping and display works in the zeichenwerk framework.", 0).Padding(1, 0, 0, 0).
		Label("unicode-label", "Unicode Support: ★ ♠ ♣ ♥ ♦ → ← ↑ ↓ ✓ ✗", 0).Padding(1, 0, 0, 0).
		Label("box-drawing", "Box Drawing: ┌─┬─┐ │ │ │ ├─┼─┤ │ │ │ └─┴─┘", 0).
		End()
}

func progress(builder *Builder) {
	builder.Flex("progress-demo", "vertical", "stretch", 1).Padding(1, 2).
		Label("progress-title", "Progress Bar Demo", 0).Padding(0, 0, 1, 0).
		Label("progress-info", "Progress bars show completion status with customizable ranges.", 0).Padding(0, 0, 1, 0).
		Separator("progress-sep", "thin", 0, 1).Padding(0, 0, 1, 0).
		Label("progress1-label", "25% Complete:", 0).
		ProgressBar("progress1", 25, 0, 100).Hint(30, 1).
		Label("progress2-label", "50% Complete:", 0).Padding(1, 0, 0, 0).
		ProgressBar("progress2", 50, 0, 100).Hint(30, 1).
		Label("progress3-label", "75% Complete:", 0).Padding(1, 0, 0, 0).
		ProgressBar("progress3", 75, 0, 100).Hint(30, 1).
		Label("progress4-label", "100% Complete:", 0).Padding(1, 0, 0, 0).
		ProgressBar("progress4", 100, 0, 100).Hint(30, 1).
		Label("progress5-label", "Custom Range (30/50):", 0).Padding(1, 0, 0, 0).
		ProgressBar("progress5", 30, 0, 50).Hint(30, 1).
		Label("progress-note", "Progress bars can have custom min/max values and styling.", 0).Padding(1, 0, 0, 0).
		End()
}

func popup() Container {
	return NewBuilder(TokyoNightTheme()).
		Class("popup").
		Flex("popup", "vertical", "stretch", 0).
		Label("title", "Dialog", 0).Padding(1, 2).Background("", "$aqua").Foreground("", "$bg0").
		Flex("content", "vertical", "stretch", 0).Hint(0, -1).Padding(1, 2).
		Label("test", "Hello World", 0).Padding(0, 0, 1, 0).
		Label("label", "Input", 0).
		Input("prompt", "", 20).
		End().
		Separator("button-separator", "thick", 0, 1).Foreground("", "$bg0").
		Flex("popup-buttons", "horizontal", "start", 1).Padding(0, 2, 1).
		Label("", "", 0).Hint(-1, 1).
		Button("ok", "OK").
		Button("cancel", "Cancel").
		End().
		Container()
}

func scroller(builder *Builder) {
	builder.Flex("scroller-demo", "vertical", "stretch", 1).Padding(1, 2).
		Label("scroller-title", "Scroller Demo", 0).Padding(0, 0, 1, 0).
		Label("scroller-info", "A scroller shows a viewport of the inside widget.", 0).Padding(0, 0, 1, 0).
		Separator("progress-sep", "thin", 0, 1).Padding(0, 0, 1, 0).
		Scroller("scroller", "Scroller").Border("", "thin").Hint(-1, -1).
		Add(custom()).
		End().
		End()
}

func custom() Widget {
	result := NewCustom("custom", false, func(widget Widget, screen Screen) {
		width, height := widget.Size()
		widget.Log(widget, "debug", "Custom render %d %d", width, height)
		for x := 10; x < width; x += 10 {
			for y := 10; y < height; y += 10 {
				screen.SetContent(x, y, '*', nil, tcell.StyleDefault)
			}
		}
	})
	result.SetStyle("", NewStyle("green", "black").SetMargin(0).SetPadding(0))
	result.SetHint(200, 100)
	return result
}

func theme(builder *Builder) {
	builder.ThemeSwitch("theme-switch", GruvboxLightTheme()).
		Box("theme-box", "Gruvbox Light").Border("", "thin").
		Flex("theme-flex", "vertical", "stretch", 1).Padding(1).
		Button("theme-button", "Just a button").
		Checkbox("theme-cb", "Checkbox", true).
		Input("theme-input", "Label", 20).
		Spacer().
		End().
		End().
		End()
}

func table(builder *Builder) {
	headers := []string{
		"First name",
		"Last name",
		"Street address",
		"ZIP",
		"City",
		"State",
		"Country",
		"Phone",
		"E-Mail",
		"Date of Birth",
		"Age",
		"Place of Birth",
		"Income",
		"SSN",
		"Sex",
	}
	data := people(100)
	builder.Table("table-demo", NewArrayTableProvider(headers, data))
}

func spinner(builder *Builder) {
	builder.Box("spinner-demo", "Spinner").Border("", "round").Margin(1).Padding(1, 5).
		Flex("spinner-flex", "horizontal", "start", 2).
		Spinner("spinner", []rune(Spinners["bar"])).
		Spinner("spinner", []rune(Spinners["dot"])).
		Spinner("spinner", []rune(Spinners["dots"])).
		Spinner("spinner", []rune(Spinners["arrow"])).
		Spinner("spinner", []rune(Spinners["circle"])).
		Spinner("spinner", []rune(Spinners["bounce"])).
		Spinner("spinner", []rune(Spinners["braille"])).
		End().
		End()
}

func editor(builder *Builder) {
	builder.Editor("editor")
}

func digits(builder *Builder) {
	builder.Flex("digits-demo", "vertical", "stretch", 1).
		Box("digits-input-box", "Digits").Border("", "thin").
		Input("digits-input", "Digits", 20).
		End().
		Box("digits-output-box", "Result").Border("", "thin").Padding(1).
		Digits("digits", "0123456789ABCDEF.,#:").
		End().
		End()

	builder.Find("digits-input").On("enter", func(widget Widget, event string, values ...any) bool {
		ui := FindUI(widget)
		Update(ui, "digits", fmt.Sprintf("%v", values[0]))
		Redraw(ui.Find("digits", true))
		return true
	})
}

var (
	firstNames = []string{"John", "Jane", "Michael", "Emily", "David", "Sophia", "James", "Olivia", "Daniel", "Ava", "Liam", "Emma", "Noah", "Isabella", "Ethan", "Mia", "Lucas", "Charlotte", "Mason", "Amelia"}
	lastNames  = []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin"}
	streets    = []string{"Maple St", "Oak Ave", "Pine Rd", "Birch Blvd", "Cedar Ln", "Spruce Ct", "Willow Way", "Elm Pl", "Aspen Dr", "Cypress St"}
	cities     = []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose"}
	states     = []string{"NY", "CA", "IL", "TX", "AZ", "PA", "OH", "MI", "GA", "NC"}
	countries  = []string{"USA"}
	sexes      = []string{"M", "F"}
)

func randomFrom(list []string) string {
	return list[rand.Intn(len(list))]
}

func randomPhone() string {
	return fmt.Sprintf("+1-%03d-%03d-%04d", rand.Intn(999), rand.Intn(999), rand.Intn(10000))
}

func randomEmail(first, last string) string {
	domains := []string{"example.com", "mail.com", "test.org", "demo.net"}
	return fmt.Sprintf("%s.%s@%s", first, last, randomFrom(domains))
}

func randomDOB() (string, int) {
	year := rand.Intn(60) + 1955 // between 1955–2015
	month := rand.Intn(12) + 1
	day := rand.Intn(28) + 1
	age := time.Now().Year() - year
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day), age
}

func randomSSN() string {
	return fmt.Sprintf("%03d-%02d-%04d", rand.Intn(900)+100, rand.Intn(90)+10, rand.Intn(10000))
}

func randomIncome() string {
	return strconv.Itoa((rand.Intn(90) + 30) * 1000) // 30k–120k
}

func generatePerson() []string {
	first := randomFrom(firstNames)
	last := randomFrom(lastNames)
	streetNumber := strconv.Itoa(rand.Intn(999) + 1)
	street := streetNumber + " " + randomFrom(streets)
	zip := fmt.Sprintf("%05d", rand.Intn(99999))
	city := randomFrom(cities)
	state := randomFrom(states)
	country := randomFrom(countries)
	phone := randomPhone()
	email := randomEmail(first, last)
	dob, age := randomDOB()
	placeOfBirth := randomFrom(cities)
	income := randomIncome()
	ssn := randomSSN()
	sex := randomFrom(sexes)

	return []string{
		first,
		last,
		street,
		zip,
		city,
		state,
		country,
		phone,
		email,
		dob,
		strconv.Itoa(age),
		placeOfBirth,
		income,
		ssn,
		sex,
	}
}

func people(count int) [][]string {
	// Generate 100 people
	people := make([][]string, count)
	for i := range count {
		people[i] = generatePerson()
	}
	return people
}

var Countries = map[string]string{
	"AC": "Ascension",
	"AD": "Andorra",
	"AE": "United Arab Emirates",
	"AF": "Afghanistan",
	"AG": "Antigua and Barbuda",
	"AI": "Anguilla",
	"AL": "Albania",
	"AM": "Armenia",
	"AO": "Angola",
	"AQ": "Antarctica",
	"AR": "Argentina",
	"AS": "American Samoa",
	"AT": "Austria",
	"AU": "Australia",
	"AW": "Aruba",
	"AX": "Aland Islands",
	"AZ": "Azerbaijan",
	"BA": "Bosnia and Herzegovina",
	"BB": "Barbados",
	"BD": "Bangladesh",
	"BE": "Belgium",
	"BF": "Burkina Faso",
	"BG": "Bulgaria",
	"BH": "Bahrain",
	"BI": "Burundi",
	"BJ": "Benin",
	"BL": "Saint Barthelemy",
	"BM": "Bermuda",
	"BN": "Brunei Darussalam",
	"BO": "Bolivia",
	"BQ": "Bonaire, Sint Eustatius and Saba",
	"BR": "Brazil",
	"BS": "Bahamas",
	"BT": "Bhutan",
	"BV": "Bouvet Island",
	"BW": "Botswana",
	"BY": "Belarus",
	"BZ": "Belize",
	"CA": "Canada",
	"CC": "Cocos (Keeling) Islands",
	"CD": "Congo, Democratic Republic of the",
	"CF": "Central African Republic",
	"CG": "Congo",
	"CH": "Switzerland",
	"CI": "Cote d'Ivoire",
	"CK": "Cook Islands",
	"CL": "Chile",
	"CM": "Cameroon",
	"CN": "China",
	"CO": "Colombia",
	"CR": "Costa Rica",
	"CU": "Cuba",
	"CV": "Cabo Verde",
	"CW": "Curacao",
	"CX": "Christmas Island",
	"CY": "Cyprus",
	"CZ": "Czechia",
	"DE": "Germany",
	"DJ": "Djibouti",
	"DK": "Denmark",
	"DM": "Dominica",
	"DO": "Dominican Republic",
	"DZ": "Algeria",
	"EC": "Ecuador",
	"EE": "Estonia",
	"EG": "Egypt",
	"EH": "Western Sahara",
	"ER": "Eritrea",
	"ES": "Spain",
	"ET": "Ethiopia",
	"FI": "Finland",
	"FJ": "Fiji",
	"FK": "Falkland Islands",
	"FM": "Micronesia",
	"FO": "Faroe Islands",
	"FR": "France",
	"GA": "Gabon",
	"GB": "United Kingdom",
	"GD": "Grenada",
	"GE": "Georgia",
	"GF": "French Guiana",
	"GG": "Guernsey",
	"GH": "Ghana",
	"GI": "Gibraltar",
	"GL": "Greenland",
	"GM": "Gambia",
	"GN": "Guinea",
	"GP": "Guadeloupe",
	"GQ": "Equatorial Guinea",
	"GR": "Greece",
	"GS": "South Georgia and the South Sandwich Islands",
	"GT": "Guatemala",
	"GU": "Guam",
	"GW": "Guinea-Bissau",
	"GY": "Guyana",
	"HK": "Hong Kong",
	"HM": "Heard Island and McDonald Islands",
	"HN": "Honduras",
	"HR": "Croatia",
	"HT": "Haiti",
	"HU": "Hungary",
	"ID": "Indonesia",
	"IE": "Ireland",
	"IL": "Israel",
	"IM": "Isle of Man",
	"IN": "India",
	"IO": "British Indian Ocean Territory",
	"IQ": "Iraq",
	"IR": "Iran",
	"IS": "Iceland",
	"IT": "Italy",
	"JE": "Jersey",
	"JM": "Jamaica",
	"JO": "Jordan",
	"JP": "Japan",
	"KE": "Kenya",
	"KG": "Kyrgyzstan",
	"KH": "Cambodia",
	"KI": "Kiribati",
	"KM": "Comoros",
	"KN": "Saint Kitts and Nevis",
	"KP": "Korea, Democratic People's Republic of",
	"KR": "Korea, Republic of",
	"KW": "Kuwait",
	"KY": "Cayman Islands",
	"KZ": "Kazakhstan",
	"LA": "Lao People's Democratic Republic",
	"LB": "Lebanon",
	"LC": "Saint Lucia",
	"LI": "Liechtenstein",
	"LK": "Sri Lanka",
	"LR": "Liberia",
	"LS": "Lesotho",
	"LT": "Lithuania",
	"LU": "Luxembourg",
	"LV": "Latvia",
	"LY": "Libya",
	"MA": "Morocco",
	"MC": "Monaco",
	"MD": "Moldova",
	"ME": "Montenegro",
	"MF": "Saint Martin (French part)",
	"MG": "Madagascar",
	"MH": "Marshall Islands",
	"MK": "North Macedonia",
	"ML": "Mali",
	"MM": "Myanmar",
	"MN": "Mongolia",
	"MO": "Macao",
	"MP": "Northern Mariana Islands",
	"MQ": "Martinique",
	"MR": "Mauritania",
	"MS": "Montserrat",
	"MT": "Malta",
	"MU": "Mauritius",
	"MV": "Maldives",
	"MW": "Malawi",
	"MX": "Mexico",
	"MY": "Malaysia",
	"MZ": "Mozambique",
	"NA": "Namibia",
	"NC": "New Caledonia",
	"NE": "Niger",
	"NF": "Norfolk Island",
	"NG": "Nigeria",
	"NI": "Nicaragua",
	"NL": "Netherlands",
	"NO": "Norway",
	"NP": "Nepal",
	"NR": "Nauru",
	"NU": "Niue",
	"NZ": "New Zealand",
	"OM": "Oman",
	"PA": "Panama",
	"PE": "Peru",
	"PF": "French Polynesia",
	"PG": "Papua New Guinea",
	"PH": "Philippines",
	"PK": "Pakistan",
	"PL": "Poland",
	"PM": "Saint Pierre and Miquelon",
	"PN": "Pitcairn",
	"PR": "Puerto Rico",
	"PS": "Palestine, State of",
	"PT": "Portugal",
	"PW": "Palau",
	"PY": "Paraguay",
	"QA": "Qatar",
	"RE": "Reunion",
	"RO": "Romania",
	"RS": "Serbia",
	"RU": "Russia",
	"RW": "Rwanda",
	"SA": "Saudi Arabia",
	"SB": "Solomon Islands",
	"SC": "Seychelles",
	"SD": "Sudan",
	"SE": "Sweden",
	"SG": "Singapore",
	"SH": "Saint Helena, Ascension and Tristan da Cunha",
	"SI": "Slovenia",
	"SJ": "Svalbard and Jan Mayen",
	"SK": "Slovakia",
	"SL": "Sierra Leone",
	"SM": "San Marino",
	"SN": "Senegal",
	"SO": "Somalia",
	"SR": "Suriname",
	"SS": "South Sudan",
	"ST": "Sao Tome and Principe",
	"SV": "El Salvador",
	"SX": "Sint Maarten (Dutch part)",
	"SY": "Syrian Arab Republic",
	"SZ": "Eswatini",
	"TC": "Turks and Caicos Islands",
	"TD": "Chad",
	"TF": "French Southern Territories",
	"TG": "Togo",
	"TH": "Thailand",
	"TJ": "Tajikistan",
	"TK": "Tokelau",
	"TL": "Timor-Leste",
	"TM": "Turkmenistan",
	"TN": "Tunisia",
	"TO": "Tonga",
	"TR": "Turkey",
	"TT": "Trinidad and Tobago",
	"TV": "Tuvalu",
	"TW": "Taiwan",
	"TZ": "Tanzania",
	"UA": "Ukraine",
	"UG": "Uganda",
	"UM": "United States Minor Outlying Islands",
	"US": "United States of America",
	"UY": "Uruguay",
	"UZ": "Uzbekistan",
	"VA": "Holy See",
	"VC": "Saint Vincent and the Grenadines",
	"VE": "Venezuela",
	"VG": "Virgin Islands, British",
	"VI": "Virgin Islands, U.S.",
	"VN": "Vietnam",
	"VU": "Vanuatu",
	"WF": "Wallis and Futuna",
	"WS": "Samoa",
	"YE": "Yemen",
	"YT": "Mayotte",
	"ZA": "South Africa",
	"ZM": "Zambia",
	"ZW": "Zimbabwe",
}

var names = []string{
	"Alice",
	"Andrew",
	"Benjamin",
	"Charlotte",
	"Daniel",
	"David",
	"Edward",
	"Elizabeth",
	"Emily",
	"Ethan",
	"George",
	"Grace",
	"Hannah",
	"Harry",
	"Isabella",
	"Jack",
	"James",
	"Jessica",
	"John",
	"Joseph",
	"Joshua",
	"Julia",
	"Katherine",
	"Laura",
	"Matthew",
	"Michael",
	"Nathan",
	"Olivia",
	"Sophia",
	"William",
}

var CountryData = [][]string{
	{"US", "United States", "North America", "Washington, D.C.", "331000000"},
	{"CA", "Canada", "North America", "Ottawa", "38000000"},
	{"MX", "Mexico", "North America", "Mexico City", "128900000"},
	{"BR", "Brazil", "South America", "Brasília", "213000000"},
	{"AR", "Argentina", "South America", "Buenos Aires", "45100000"},
	{"CL", "Chile", "South America", "Santiago", "19400000"},
	{"CO", "Colombia", "South America", "Bogotá", "50800000"},
	{"PE", "Peru", "South America", "Lima", "32900000"},
	{"VE", "Venezuela", "South America", "Caracas", "28400000"},
	{"UY", "Uruguay", "South America", "Montevideo", "3470000"},
	{"GB", "United Kingdom", "Europe", "London", "67800000"},
	{"FR", "France", "Europe", "Paris", "65200000"},
	{"DE", "Germany", "Europe", "Berlin", "83100000"},
	{"IT", "Italy", "Europe", "Rome", "60400000"},
	{"ES", "Spain", "Europe", "Madrid", "47300000"},
	{"PT", "Portugal", "Europe", "Lisbon", "10300000"},
	{"NL", "Netherlands", "Europe", "Amsterdam", "17400000"},
	{"BE", "Belgium", "Europe", "Brussels", "11500000"},
	{"CH", "Switzerland", "Europe", "Bern", "8700000"},
	{"AT", "Austria", "Europe", "Vienna", "9000000"},
	{"SE", "Sweden", "Europe", "Stockholm", "10400000"},
	{"NO", "Norway", "Europe", "Oslo", "5400000"},
	{"FI", "Finland", "Europe", "Helsinki", "5500000"},
	{"DK", "Denmark", "Europe", "Copenhagen", "5800000"},
	{"PL", "Poland", "Europe", "Warsaw", "38300000"},
	{"CZ", "Czech Republic", "Europe", "Prague", "10700000"},
	{"HU", "Hungary", "Europe", "Budapest", "9700000"},
	{"GR", "Greece", "Europe", "Athens", "10700000"},
	{"RU", "Russia", "Europe/Asia", "Moscow", "146000000"},
	{"TR", "Turkey", "Asia/Europe", "Ankara", "84300000"},
	{"CN", "China", "Asia", "Beijing", "1412000000"},
	{"IN", "India", "Asia", "New Delhi", "1393000000"},
	{"JP", "Japan", "Asia", "Tokyo", "126000000"},
	{"KR", "South Korea", "Asia", "Seoul", "51700000"},
	{"ID", "Indonesia", "Asia", "Jakarta", "273000000"},
	{"TH", "Thailand", "Asia", "Bangkok", "70000000"},
	{"VN", "Vietnam", "Asia", "Hanoi", "97300000"},
	{"MY", "Malaysia", "Asia", "Kuala Lumpur", "32700000"},
	{"PH", "Philippines", "Asia", "Manila", "109600000"},
	{"AU", "Australia", "Oceania", "Canberra", "25700000"},
	{"NZ", "New Zealand", "Oceania", "Wellington", "5000000"},
	{"EG", "Egypt", "Africa", "Cairo", "102000000"},
	{"NG", "Nigeria", "Africa", "Abuja", "206000000"},
	{"ZA", "South Africa", "Africa", "Pretoria", "59300000"},
	{"KE", "Kenya", "Africa", "Nairobi", "53700000"},
	{"ET", "Ethiopia", "Africa", "Addis Ababa", "115000000"},
	{"DZ", "Algeria", "Africa", "Algiers", "43800000"},
	{"MA", "Morocco", "Africa", "Rabat", "36900000"},
	{"GH", "Ghana", "Africa", "Accra", "31000000"},
	{"SN", "Senegal", "Africa", "Dakar", "16700000"},
}

package main

import (
	"maps"
	"slices"

	. "github.com/tekugo/zeichenwerk"
)

func main() {
	// Create a simple list
	countries := slices.Collect(maps.Values(Countries))

	ui := NewBuilder().
		Flex("main", "vertical", "stretch", 0).
		Add(header).
		Grid("grid", 3, 3, false).Hint(0, -1).
		Cell(0, 0, 1, 3).Text("debug-log", []string{}, true, 1000).
		Cell(1, 0, 2, 1).
		Flex("form", "vertical", "start", 1).Border("", "thin").Margin(1).Padding(1).
		Flex("name-flex", "horizontal", "center", 1).
		Label("name-label", "Name", 20).
		Input("name-input", "", 20).Hint(20, 1).
		End().
		Flex("date-flex", "horizontal", "center", 1).
		Label("date-label", "Date of Birth", 20).
		Input("input2", "", 20).Hint(20, 1).
		End().
		ProgressBar("bar1", 0, 0, 100).Hint(30, 1).
		ProgressBar("bar2", 50, 0, 100).Hint(30, 1).
		ProgressBar("bar3", 100, 0, 100).Hint(30, 1).
		End().
		Cell(1, 1, 1, 2).List("countries", countries).
		Cell(2, 1, 1, 1).List("names", names).
		Cell(2, 2, 1, 1).Box("button-box", "Buttons").Border("", "thin").
		Flex("buttons", "vertical", "center", 1).
		Padding(1, 2).
		Button("b1", " Button 1 ").Hint(20, 1).
		Button("b2", " Button 2 ").Hint(20, 1).
		End().
		End().
		End().
		Add(footer).
		Class("").
		Build()

	// Work around to adjust grid sizing from fractional to auto or fixed for
	// frist row, have to find a better solution
	grid := ui.Find("grid")
	if grid, ok := grid.(*Grid); ok {
		form := ui.Find("form")
		_, ph := form.Hint()
		ph += form.Style("").Vertical()
		grid.Log("Preferred height of form is %d", ph)
		grid.Rows(ph, -1, -1)
		grid.Layout()
	}

	popup := popup()
	popup.SetParent(ui)

	pw, ph := popup.Hint()
	ui.Popup(-1, -1, pw, ph, popup)

	HandleListEvent(ui, "countries", "select", func(list *List, event string, index int) bool {
		list.Log("List event %s", event)
		return false
	})

	ui.Run()
}

func header(builder *Builder) *Builder {
	builder.Class("header").
		Flex("header", "horizontal", "start", 0).Hint(0, 1).
		Label("title", "My Application", 20).Hint(20, 1).
		Label("", "Hello", 0).Hint(-1, 1).
		Label("time", "12:00", 0).Hint(6, 1).
		Class("").
		End()
	return builder
}

func footer(builder *Builder) *Builder {
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

	return builder
}

func popup() Container {
	return NewBuilder().
		Class("popup").
		Flex("popup", "vertical", "stretch", 0).
		Label("title", "Dialog", 0).Padding(1, 2).Background("", "$aqua").Foreground("", "$bg").
		Flex("content", "vertical", "stretch", 0).Hint(0, -1).Padding(1, 2).
		Label("test", "Hello World", 0).Padding(0, 0, 1, 0).
		Label("label", "Input", 0).
		Input("prompt", "", 20).
		End().
		Separator("button-separator", "thick", 0, 1).Background("", "$comments").Foreground("", "$bg").
		Flex("popup-buttons", "horizontal", "start", 1).Padding(0, 2, 1).
		Label("", "", 0).Hint(-1, 1).
		Button("ok", "OK").
		Button("cancel", "Cancel").
		End().
		Container()
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

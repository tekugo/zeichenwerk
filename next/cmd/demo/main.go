package main

import . "github.com/tekugo/zeichenwerk/next"

func main() {
	ui := createUI()
	ui.Run()
}

func createUI() *UI {
	return NewBuilder(TokyoNightTheme()).
		Flex("main", false, "stretch", 0).
		Flex("header", true, "stretch", 2).
		Static("title", "Zeichenwerk Demo").
		Static("subtitle", "A terminal UI framework").
		End().
		Grid("content", 1, 2, true).Hint(0, -1).Columns(20, -1).
		Cell(0, 0, 1, 1).
		List("navigation", []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen", "twenty", "twenty-one", "twenty-two", "twenty-three", "twenty-four", "twenty-five", "twenty-six", "twenty-seven", "twenty-eight", "twenty-nine", "thirty", "thirty-one", "thirty-two", "thirty-three", "thirty-four", "thirty-five", "thirty-six", "thirty-seven", "thirty-eight", "thirty-nine", "forty", "forty-one", "forty-two", "forty-three", "forty-four", "forty-five", "forty-six", "forty-seven", "forty-eight", "forty-nine", "fifty", "fifty-one", "fifty-two", "fifty-three", "fifty-four", "fifty-five", "fifty-six", "fifty-seven", "fifty-eight", "fifty-nine", "sixty", "sixty-one", "sixty-two", "sixty-three", "sixty-four", "sixty-five", "sixty-six", "sixty-seven", "sixty-eight", "sixty-nine", "seventy", "seventy-one", "seventy-two", "seventy-three", "seventy-four", "seventy-five", "seventy-six", "seventy-seven", "seventy-eight", "seventy-nine", "eighty", "eighty-one", "eighty-two", "eighty-three", "eighty-four", "eighty-five", "eighty-six", "eighty-seven", "eighty-eight", "eighty-nine", "ninety", "ninety-one", "ninety-two", "ninety-three", "ninety-four", "ninety-five", "ninety-six", "ninety-seven", "ninety-eight", "ninety-nine", "one hundred"}).
		Cell(1, 0, 1, 1).
		Box("display", "").
		Static("display-title", "Display").
		End().
		End().
		Flex("footer", true, "stretch", 0).
		Static("footer-text", "Footer").
		End().
		Build()
}

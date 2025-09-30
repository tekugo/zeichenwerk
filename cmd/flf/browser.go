package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mbndr/figlet4go"
	. "github.com/tekugo/zeichenwerk"
)

func main() {
	ui := createFontBrowserUI()
	ui.Run()
}

func createFontBrowserUI() *UI {
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
		Label("title", "FIGlet Font Browser", 30).Hint(30, 1).
		Label("", "Browse and preview figlet fonts", 0).Hint(-1, 1).
		Class("").
		End()
}

func footer(builder *Builder) {
	builder.Class("footer").
		Flex("footer", "horizontal", "start", 0).Padding(0, 1).Hint(0, 1).
		Class("shortcut").Label("1", "Esc", 0).
		Class("footer").Label("2", "Quit \u2502", 0).
		Class("shortcut").Label("3", "Enter", 0).
		Class("footer").Label("4", "Preview \u2502", 0).
		Class("shortcut").Label("5", "Ctrl-Q", 0).
		Class("footer").Label("6", "Exit Application", 0).
		Class("").
		Spacer().
		End()
}

func content(builder *Builder) {
	// Get list of all .flf files
	fontFiles := getFontFiles()

	builder.Grid("grid", 1, 2, true).Hint(0, -1).
		Cell(0, 0, 1, 1).
		List("fonts", fontFiles).Border("", "round").Border(":focus", "double").
		Cell(1, 0, 1, 1).
		Flex("preview-area", "vertical", "stretch", 0).
		Box("input-box", "Text Input").Border("", "round").Padding(1).
		Input("text-input", "Hello World", 40).
		End().
		Box("preview-box-custom", "Our Implementation").Border("", "round").Padding(1).
		Text("preview-custom", []string{"Enter text above and select a font to see preview"}, true, 100).Hint(-1, 10).
		End().
		Box("preview-box-figlet4go", "figlet4go Reference").Border("", "round").Padding(1).
		Text("preview-figlet4go", []string{"figlet4go library output will appear here"}, true, 100).Hint(-1, 10).
		End().
		End().
		End()

	// Configure grid layout
	grid := builder.Container().Find("grid", false)
	if grid, ok := grid.(*Grid); ok {
		grid.Columns(30, -1) // Font list takes 30 chars, preview takes rest
	}

	// Add event handlers
	setupEventHandlers(builder.Container())

	// Set initial preview content
	if len(getFontFiles()) > 0 {
		firstFont := getFontFiles()[0]
		defaultText := "Hello World"

		// Set custom implementation preview
		if previewWidget := builder.Container().Find("preview-custom", false); previewWidget != nil {
			if text, ok := previewWidget.(*Text); ok {
				customPreview := generateCustomFontPreview(firstFont, defaultText)
				text.Set(customPreview)
			}
		}

		// Set figlet4go reference preview
		if previewWidget := builder.Container().Find("preview-figlet4go", false); previewWidget != nil {
			if text, ok := previewWidget.(*Text); ok {
				figlet4goPreview := generateFiglet4goPreview(firstFont, defaultText)
				text.Set(figlet4goPreview)
			}
		}
	}
}

func getFontFiles() []string {
	var fontFiles []string

	// Get all .flf files in the cmd/flf directory
	files, err := filepath.Glob("*.flf")
	if err != nil {
		// If we're not in the flf directory, try the full path
		files, err = filepath.Glob("cmd/flf/*.flf")
		if err != nil {
			return []string{"No font files found"}
		}
	}

	for _, file := range files {
		// Extract just the filename without extension
		base := filepath.Base(file)
		name := strings.TrimSuffix(base, ".flf")
		fontFiles = append(fontFiles, name)
	}

	sort.Strings(fontFiles)
	return fontFiles
}

func setupEventHandlers(container Container) {
	// Find the list widget and set up event handlers directly
	if fontList := container.Find("fonts", false); fontList != nil {
		if list, ok := fontList.(*List); ok {
			list.On("activate", func(widget Widget, event string, data ...any) bool {
				ui := FindUI(widget)
				updatePreview(ui)
				return true
			})

			list.On("select", func(widget Widget, event string, data ...any) bool {
				ui := FindUI(widget)
				updatePreview(ui)
				return true
			})
		}
	}

	// Handle text input changes
	if textInput := container.Find("text-input", false); textInput != nil {
		textInput.On("enter", func(widget Widget, event string, data ...any) bool {
			ui := FindUI(widget)
			updatePreview(ui)
			return true
		})

		textInput.On("change", func(widget Widget, event string, data ...any) bool {
			ui := FindUI(widget)
			updatePreview(ui)
			return true
		})
	}
}

func updatePreview(ui *UI) {
	// Get selected font
	fontList := ui.Find("fonts", true)
	if fontList == nil {
		// Try with recursion disabled
		fontList = ui.Find("fonts", false)
		if fontList == nil {
			return
		}
	}

	list, ok := fontList.(*List)
	if !ok {
		return
	}

	selectedIndex := list.Index
	if selectedIndex < 0 || selectedIndex >= len(list.Items) {
		// If no selection, use first font
		if len(list.Items) > 0 {
			selectedIndex = 0
			list.Index = 0
		} else {
			return
		}
	}

	fontName := list.Items[selectedIndex]

	// Get input text
	textInput := ui.Find("text-input", true)
	if textInput == nil {
		textInput = ui.Find("text-input", false)
		if textInput == nil {
			return
		}
	}

	input, ok := textInput.(*Input)
	if !ok {
		return
	}

	inputText := input.Text
	if inputText == "" {
		inputText = "Hello World"
	}

	// Generate previews for both implementations
	customPreview := generateCustomFontPreview(fontName, inputText)
	figlet4goPreview := generateFiglet4goPreview(fontName, inputText)

	// Update custom implementation preview
	if previewWidget := ui.Find("preview-custom", true); previewWidget != nil {
		if text, ok := previewWidget.(*Text); ok {
			text.Set(customPreview)
			text.Refresh()
		}
	} else if previewWidget := ui.Find("preview-custom", false); previewWidget != nil {
		if text, ok := previewWidget.(*Text); ok {
			text.Set(customPreview)
			text.Refresh()
		}
	}

	// Update figlet4go reference preview
	if previewWidget := ui.Find("preview-figlet4go", true); previewWidget != nil {
		if text, ok := previewWidget.(*Text); ok {
			text.Set(figlet4goPreview)
			text.Refresh()
		}
	} else if previewWidget := ui.Find("preview-figlet4go", false); previewWidget != nil {
		if text, ok := previewWidget.(*Text); ok {
			text.Set(figlet4goPreview)
			text.Refresh()
		}
	}
}

func generateCustomFontPreview(fontName, text string) []string {
	// Try different possible paths for the font file
	fontPaths := []string{
		fontName + ".flf",
		"cmd/flf/" + fontName + ".flf",
		filepath.Join("cmd", "flf", fontName+".flf"),
	}

	var fontPath string
	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			fontPath = path
			break
		}
	}

	if fontPath == "" {
		return []string{
			"Error: Font file not found",
			"Font: " + fontName,
			"Tried paths:",
			"  " + strings.Join(fontPaths, "\n  "),
		}
	}

	// Use the figlet rendering function (copied from main.go)
	lines, err := renderFigletFromFile(fontPath, text)
	if err != nil {
		return []string{
			"Error rendering font: " + err.Error(),
			"Font: " + fontName,
			"Text: " + text,
		}
	}

	// Add header with font info
	result := []string{
		fmt.Sprintf("Font: %s | Text: %s", fontName, text),
		strings.Repeat("─", 50),
	}
	result = append(result, lines...)

	return result
}

func generateFiglet4goPreview(fontName, text string) []string {
	// Create a new figlet4go renderer for each call to ensure clean state
	ascii := figlet4go.NewAsciiRender()

	// Find the font file path
	fontPaths := []string{
		"./" + fontName + ".flf",
		"cmd/flf/" + fontName + ".flf",
		filepath.Join("cmd", "flf", fontName+".flf"),
	}

	var fontPath string
	for _, path := range fontPaths {
		if _, statErr := os.Stat(path); statErr == nil {
			fontPath = path
			break
		}
	}

	// Always try to load the specific font file (even for standard)
	var err error
	if fontPath != "" {
		err = ascii.LoadFont(fontPath)
	} else {
		return []string{
			"figlet4go Error: Font file not found",
			"Font: " + fontName,
			"Text: " + text,
			"Tried paths:",
			"  " + strings.Join(fontPaths, "\n  "),
		}
	}

	if err != nil {
		return []string{
			"figlet4go Error: " + err.Error(),
			"Font: " + fontName + " (path: " + fontPath + ")",
			"Text: " + text,
			"",
			"Note: figlet4go may not support this font format",
		}
	}

	// Render the text
	opts := figlet4go.NewRenderOptions()
	opts.FontName = fontName
	renderStr, err := ascii.RenderOpts(text, opts)
	if err != nil {
		return []string{
			"figlet4go Render Error: " + err.Error(),
			"Font: " + fontName,
			"Text: " + text,
		}
	}

	// Split into lines and add header
	lines := strings.Split(strings.TrimRight(renderStr, "\n"), "\n")
	result := []string{
		fmt.Sprintf("figlet4go | Font: %s | Text: %s", fontName, text),
		strings.Repeat("─", 50),
	}
	result = append(result, lines...)

	return result
}

// Copy the figlet rendering functions from main.go
type glyph struct {
	lines [][]rune
	width int
}

type LayoutMode int

const (
	ModeFullWidth LayoutMode = iota
	ModeFitting
	ModeSmushingControlled
)

type SmushRules struct {
	Equal        bool
	Underscore   bool
	Hierarchy    bool
	OppositePair bool
	BigX         bool
	HardBlank    bool
}

func renderFigletFromFile(filename string, text string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}
	return renderFiglet(string(data), text)
}

func renderFiglet(fontData string, text string) ([]string, error) {
	scanner := bufio.NewScanner(strings.NewReader(fontData))
	if !scanner.Scan() {
		return nil, errors.New("font has no header")
	}
	header := scanner.Text()
	parts := strings.Fields(header)
	if len(parts) < 6 || !strings.HasPrefix(parts[0], "flf2a") {
		return nil, fmt.Errorf("invalid FIGlet header: %q", header)
	}
	hardblank := rune(parts[0][5])
	height, _ := strconv.Atoi(parts[1])
	oldLayout, _ := strconv.Atoi(parts[4])
	commentLines, _ := strconv.Atoi(parts[5])

	// skip comments
	for i := 0; i < commentLines; i++ {
		if !scanner.Scan() {
			return nil, errors.New("unexpected EOF while skipping comments")
		}
	}

	// read glyphs 32..126
	const firstChar = 32
	const lastChar = 126
	numChars := lastChar - firstChar + 1
	glyphs := make([]glyph, numChars)

	var endmark rune = '@' // Standard figlet fonts use @ as endmark
	setEndmark := false
	cleanLine := func(raw string) string {
		raw = strings.TrimRight(raw, "\r")
		if !setEndmark {
			// Look for the endmark in the first non-empty line
			if raw != "" {
				r, _ := utf8.DecodeLastRuneInString(raw)
				endmark = r
				setEndmark = true
			}
		}
		// strip trailing endmarks
		for {
			if raw == "" {
				break
			}
			r, size := utf8.DecodeLastRuneInString(raw)
			if r == endmark {
				raw = raw[:len(raw)-size]
			} else {
				break
			}
		}
		return raw
	}

	for gi := 0; gi < numChars; gi++ {
		lines := make([][]rune, height)
		maxw := 0
		for h := 0; h < height; h++ {
			if !scanner.Scan() {
				return nil, fmt.Errorf("unexpected EOF reading glyph %d", gi+firstChar)
			}
			clean := cleanLine(scanner.Text())
			runes := []rune(clean)
			lines[h] = runes
			if len(runes) > maxw {
				maxw = len(runes)
			}
		}

		glyphs[gi] = glyph{lines: lines, width: maxw}
	}

	// Determine layout mode
	mode, rules := interpretLayout(oldLayout)

	// render
	canvas := make([][]rune, height)
	canvasWidth := 0
	ensureCanvasWidth := func() {
		for r := 0; r < height; r++ {
			if len(canvas[r]) < canvasWidth {
				pad := make([]rune, canvasWidth-len(canvas[r]))
				for i := range pad {
					pad[i] = ' '
				}
				canvas[r] = append(canvas[r], pad...)
			}
		}
	}

	for _, rr := range text {
		var g *glyph
		if int(rr) >= firstChar && int(rr) <= lastChar {
			g = &glyphs[int(rr)-firstChar]
		} else {
			g = &glyphs['?'-firstChar]
		}

		switch mode {
		case ModeFullWidth:
			// append without overlap
			ensureCanvasWidth()
			newW := canvasWidth + g.width
			for r := 0; r < height; r++ {
				if len(canvas[r]) < canvasWidth {
					pad := make([]rune, canvasWidth-len(canvas[r]))
					for i := range pad {
						pad[i] = ' '
					}
					canvas[r] = append(canvas[r], pad...)
				}
				for i := 0; i < g.width; i++ {
					if i < len(g.lines[r]) {
						canvas[r] = append(canvas[r], g.lines[r][i])
					} else {
						canvas[r] = append(canvas[r], ' ')
					}
				}
			}
			canvasWidth = newW

		case ModeFitting:
			// kerning only
			ensureCanvasWidth()
			overlap := computeSmushAmount(canvas, canvasWidth, g, SmushRules{}, hardblank)
			prefix := canvasWidth - overlap
			newWidth := prefix + g.width
			newCanvas := make([][]rune, height)
			for r := 0; r < height; r++ {
				if len(canvas[r]) < canvasWidth {
					pad := make([]rune, canvasWidth-len(canvas[r]))
					for i := range pad {
						pad[i] = ' '
					}
					canvas[r] = append(canvas[r], pad...)
				}
				row := make([]rune, 0, newWidth)
				if prefix > 0 {
					row = append(row, canvas[r][:prefix]...)
				}
				for i := 0; i < g.width; i++ {
					if i < len(g.lines[r]) {
						row = append(row, g.lines[r][i])
					} else {
						row = append(row, ' ')
					}
				}
				newCanvas[r] = row
			}
			canvas = newCanvas
			canvasWidth = newWidth

		case ModeSmushingControlled:
			ensureCanvasWidth()
			canvas, canvasWidth = appendWithSmush(canvas, canvasWidth, g, rules, hardblank)
		}
	}

	out := make([]string, height)
	for r := 0; r < height; r++ {
		rowRunes := make([]rune, len(canvas[r]))
		for i, ch := range canvas[r] {
			if ch == hardblank {
				rowRunes[i] = ' '
			} else {
				rowRunes[i] = ch
			}
		}
		out[r] = string(rowRunes)
	}
	return out, nil
}

func interpretLayout(oldLayout int) (LayoutMode, SmushRules) {
	var rules SmushRules
	mode := ModeFitting
	if oldLayout == -1 {
		mode = ModeFullWidth
	} else if oldLayout == 0 {
		mode = ModeFitting
	} else if oldLayout > 0 {
		mode = ModeSmushingControlled
		rules = SmushRules{
			Equal:        oldLayout&1 != 0,
			Underscore:   oldLayout&2 != 0,
			Hierarchy:    oldLayout&4 != 0,
			OppositePair: oldLayout&8 != 0,
			BigX:         oldLayout&16 != 0,
			HardBlank:    oldLayout&32 != 0,
		}
	}
	return mode, rules
}

func computeSmushAmount(canvas [][]rune, canvasWidth int, g *glyph, rules SmushRules, hardblank rune) int {
	if canvasWidth == 0 {
		return 0
	}

	height := len(canvas)
	maxSmush := g.width // Start with maximum possible smush

	for row := 0; row < height; row++ {
		// Find rightmost non-space character in existing canvas
		linebd := -1
		for col := canvasWidth - 1; col >= 0; col-- {
			var ch rune = ' '
			if col < len(canvas[row]) {
				ch = canvas[row][col]
			}
			if ch != ' ' && ch != 0 {
				linebd = col
				break
			}
		}

		// Find leftmost non-space character in new glyph
		charbd := g.width // Default to no character found
		for col := 0; col < g.width; col++ {
			var ch rune = ' '
			if row < len(g.lines) && col < len(g.lines[row]) {
				ch = g.lines[row][col]
			}
			if ch != ' ' && ch != 0 {
				charbd = col
				break
			}
		}

		// Calculate how much we can smush on this row
		var amt int
		if linebd < 0 {
			// No existing characters on this line
			amt = charbd + 1
		} else if charbd >= g.width {
			// No new characters on this line
			amt = g.width
		} else {
			// Both lines have characters - calculate possible overlap
			amt = charbd + canvasWidth - 1 - linebd

			// Now check if this overlap is actually valid
			if amt > 0 {
				// Get the characters that would be overlapping
				leftChar := canvas[row][linebd]
				rightChar := g.lines[row][charbd]

				// Apply the figlet overlap rules
				if leftChar == ' ' || leftChar == 0 {
					// Left is space, we can definitely overlap
					amt++
				} else if rightChar != ' ' && rightChar != 0 {
					// Both are non-space - check if they can smush
					if canSmush(leftChar, rightChar, rules, hardblank) {
						amt++
					} else {
						// Characters cannot smush - no overlap allowed
						amt = 0
					}
				}
				// If right is space and left is not, amt stays as calculated
			}
		}

		// Take the minimum across all rows
		if amt < maxSmush {
			maxSmush = amt
		}
	}

	// Clamp to reasonable bounds
	if maxSmush < 0 {
		maxSmush = 0
	}
	if maxSmush > g.width {
		maxSmush = g.width
	}

	return maxSmush
}

func canSmush(left, right rune, rules SmushRules, hardblank rune) bool {
	if left == ' ' || right == ' ' {
		return true
	}

	// Check hardblank rules
	if left == hardblank || right == hardblank {
		return rules.HardBlank && left == hardblank && right == hardblank
	}

	// Apply smushing rules
	if rules.Equal && left == right {
		return true
	}

	if rules.Underscore {
		if left == '_' && (right == '|' || right == '/' || right == '\\' ||
			right == '[' || right == ']' || right == '{' || right == '}' ||
			right == '(' || right == ')' || right == '<' || right == '>') {
			return true
		}
		if right == '_' && (left == '|' || left == '/' || left == '\\' ||
			left == '[' || left == ']' || left == '{' || left == '}' ||
			left == '(' || left == ')' || left == '<' || left == '>') {
			return true
		}
	}

	if rules.OppositePair {
		if (left == '[' && right == ']') || (left == ']' && right == '[') ||
			(left == '{' && right == '}') || (left == '}' && right == '{') ||
			(left == '(' && right == ')') || (left == ')' && right == '(') {
			return true
		}
	}

	if rules.BigX {
		if (left == '/' && right == '\\') || (left == '>' && right == '<') {
			return true
		}
	}

	return false
}

func appendWithSmush(canvas [][]rune, canvasWidth int, g *glyph, rules SmushRules, hardblank rune) ([][]rune, int) {
	height := len(canvas)

	// Calculate the correct smush amount
	smushAmount := computeSmushAmount(canvas, canvasWidth, g, rules, hardblank)

	// The new width is the original width plus glyph width minus smush overlap
	newWidth := canvasWidth + g.width - smushAmount

	// Build new canvas
	newCanvas := make([][]rune, height)
	for r := 0; r < height; r++ {
		// Ensure canvas row is properly sized
		row := canvas[r]
		if len(row) < canvasWidth {
			pad := make([]rune, canvasWidth-len(row))
			for i := range pad {
				pad[i] = ' '
			}
			row = append(row, pad...)
		}

		// Create new row
		newRow := make([]rune, newWidth)

		// Copy original canvas content
		for i := 0; i < canvasWidth; i++ {
			if i < len(row) {
				newRow[i] = row[i]
			} else {
				newRow[i] = ' '
			}
		}

		// Add glyph content with proper smushing
		for i := 0; i < g.width; i++ {
			pos := canvasWidth - smushAmount + i
			if pos >= 0 && pos < newWidth {
				var glyphChar rune = ' '
				if r < len(g.lines) && i < len(g.lines[r]) {
					glyphChar = g.lines[r][i]
				}

				if pos < canvasWidth {
					// This is an overlap position - need to smush
					canvasChar := newRow[pos]
					if canvasChar == ' ' || canvasChar == 0 {
						newRow[pos] = glyphChar
					} else if glyphChar == ' ' || glyphChar == 0 {
						// Keep canvas char
					} else {
						// Both are non-space - apply smushing rules
						if result, ok := controlledSmushResult(canvasChar, glyphChar, rules, hardblank); ok {
							newRow[pos] = result
						} else {
							// If can't smush, prefer the new character
							newRow[pos] = glyphChar
						}
					}
				} else {
					// Non-overlap position
					newRow[pos] = glyphChar
				}
			}
		}

		newCanvas[r] = newRow
	}

	return newCanvas, newWidth
}

func controlledSmushResult(left, right rune, rules SmushRules, hardblank rune) (rune, bool) {
	if left == hardblank || right == hardblank {
		if rules.HardBlank {
			if left == hardblank && right == hardblank {
				return hardblank, true
			}
			return 0, false
		}
		return 0, false
	}
	if rules.Equal && left == right {
		return left, true
	}
	if rules.Underscore {
		if left == '_' && strings.ContainsRune("|/\\[]{}()<>", right) {
			return right, true
		}
		if right == '_' && strings.ContainsRune("|/\\[]{}()<>", left) {
			return left, true
		}
	}
	if rules.OppositePair {
		if (left == '[' && right == ']') || (left == ']' && right == '[') ||
			(left == '{' && right == '}') || (left == '}' && right == '{') ||
			(left == '(' && right == ')') || (left == ')' && right == '(') {
			return '|', true
		}
	}
	if rules.BigX {
		if left == '/' && right == '\\' {
			return '|', true
		}
		if left == '\\' && right == '/' {
			return 'Y', true
		}
		if left == '>' && right == '<' {
			return 'X', true
		}
	}
	return 0, false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

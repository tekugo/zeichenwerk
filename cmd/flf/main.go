package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ---------------- FIGlet types ----------------
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

// ---------------- FIGlet renderer ----------------
func RenderFigletFromFile(filename string, text string) ([]string, error) {
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
			overlap := computeMaxKerningOverlap(canvas, canvasWidth, g, hardblank)
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

// ---------------- FIGlet helpers ----------------
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

func computeMaxKerningOverlap(canvas [][]rune, canvasWidth int, g *glyph, hardblank rune) int {
	if canvasWidth == 0 {
		return 0
	}
	height := len(canvas)
	maxOverlap := min(canvasWidth, g.width)
tryOverlap:
	for overlap := maxOverlap; overlap >= 0; overlap-- {
		prefix := canvasWidth - overlap
		for r := 0; r < height; r++ {
			row := canvas[r]
			for i := 0; i < overlap; i++ {
				leftIdx := prefix + i
				var left rune = ' '
				if leftIdx < len(row) {
					left = row[leftIdx]
				}
				var right rune = ' '
				if i < len(g.lines[r]) {
					right = g.lines[r][i]
				}
				if left != ' ' && left != hardblank || right != ' ' && right != hardblank {
					continue tryOverlap
				}
			}
		}
		return overlap
	}
	return 0
}

// computeSmushAmount implements the original figlet smushamt() algorithm
func computeSmushAmount(canvas [][]rune, canvasWidth int, g *glyph, rules SmushRules, hardblank rune) int {
	if canvasWidth == 0 {
		return 0
	}
	
	height := len(canvas)
	maxSmush := g.width
	
	for row := 0; row < height; row++ {
		// Find rightmost non-space char in current line (linebd)
		linebd := canvasWidth - 1
		for linebd >= 0 {
			if linebd < len(canvas[row]) {
				ch := canvas[row][linebd]
				if ch != ' ' && ch != 0 {
					break
				}
			}
			linebd--
		}
		
		// Find leftmost non-space char in new character (charbd)  
		charbd := 0
		for charbd < g.width {
			var ch rune = ' '
			if charbd < len(g.lines[row]) {
				ch = g.lines[row][charbd]
			}
			if ch != ' ' && ch != 0 {
				break
			}
			charbd++
		}
		if charbd == g.width {
			charbd = g.width // No characters found in this row
		}
		
		// Calculate potential overlap
		amt := charbd + canvasWidth - 1 - linebd
		
		// Get the characters that would overlap
		var ch1, ch2 rune = ' ', ' '
		if linebd >= 0 && linebd < len(canvas[row]) {
			ch1 = canvas[row][linebd]
		}
		if charbd < len(g.lines[row]) {
			ch2 = g.lines[row][charbd]
		}
		
		// Apply figlet's overlap rules
		if ch1 == ' ' || ch1 == 0 {
			amt++ // Can overlap one more if left char is space
		} else if ch2 != ' ' && ch2 != 0 {
			// Both are non-space - check if they can smush
			if canSmush(ch1, ch2, rules, hardblank) {
				amt++ // Can overlap one more if they can smush
			}
		}
		
		if amt < maxSmush {
			maxSmush = amt
		}
	}
	
	// Don't allow negative overlap
	if maxSmush < 0 {
		maxSmush = 0
	}
	
	return maxSmush
}

// canSmush implements the figlet smushem logic for checking if two characters can smush
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
	
	// Implement the original figlet smushamt() algorithm
	overlap := computeSmushAmount(canvas, canvasWidth, g, rules, hardblank)
	
	prefix := canvasWidth - overlap
	
	// Build new canvas
	newWidth := prefix + g.width
	newCanvas := make([][]rune, height)
		for r := 0; r < height; r++ {
			row := canvas[r]
			if len(row) < canvasWidth {
				pad := make([]rune, canvasWidth-len(row))
				for i := range pad {
					pad[i] = ' '
				}
				row = append(row, pad...)
			}
			newRow := make([]rune, 0, newWidth)
			newRow = append(newRow, row[:prefix]...)
			for i := 0; i < g.width; i++ {
				right := ' '
				if i < len(g.lines[r]) {
					right = g.lines[r][i]
				}
				if i < overlap {
					left := row[prefix+i]
					if left == ' ' {
						newRow = append(newRow, right)
					} else if right == ' ' {
						newRow = append(newRow, left)
					} else if left == hardblank && right == hardblank {
						newRow = append(newRow, hardblank)
					} else if left == hardblank {
						newRow = append(newRow, right)
					} else if right == hardblank {
						newRow = append(newRow, left)
					} else {
						// Both are real characters - apply smushing rules
						res, ok := controlledSmushResult(left, right, rules, hardblank)
						if ok {
							newRow = append(newRow, res)
						} else {
							// This shouldn't happen if our overlap detection worked correctly
							// but as fallback, prefer the right character
							newRow = append(newRow, right)
						}
					}
				} else {
					newRow = append(newRow, right)
				}
			}
			newCanvas[r] = newRow
		}
		return newCanvas, newWidth
}

func canControlledSmush(left, right rune, rules SmushRules) bool {
	_, ok := controlledSmushResult(left, right, rules, 0)
	return ok
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

// ---------------- Example main ----------------
func main() {
	tests := []string{"Hello, World!", "Somewhere", "Autobahn", "Just do it!"}
	
	for _, test := range tests {
		out, err := RenderFigletFromFile("standard.flf", test)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		for _, line := range out {
			fmt.Println(line)
		}
		fmt.Println()
	}
}

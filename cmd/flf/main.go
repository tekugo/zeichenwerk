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

// computeSmushAmount implements the figlet smushamt() algorithm correctly
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

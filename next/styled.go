package next

import (
	"iter"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	Italic        = 1
	Bold          = 2
	Underline     = 4
	Strikethrough = 8
	Code          = 16
)

type Block struct {
	Type    string
	Text    string
	Content []Span
	Height  int
}

type Span struct {
	Start  int
	End    int
	Length int
	Style  int
}

// Words returns an iterator over the words in the span.
// It yields each word and its length in runes.
func (s Span) Words(text string) iter.Seq2[string, int] {
	return func(yield func(string, int) bool) {
		if s.Start >= len(text) || s.End > len(text) {
			return
		}

		sub := text[s.Start:s.End]

		start := 0
		for start < len(sub) {
			// Consuming spaces
			i := start
			for i < len(sub) {
				r, size := utf8.DecodeRuneInString(sub[i:])
				if !unicode.IsSpace(r) {
					break
				}
				i += size
			}

			// Yield whitespace segment if present
			if i > start {
				segment := sub[start:i]
				if !yield(segment, utf8.RuneCountInString(segment)) {
					return
				}
				start = i
			}

			if i >= len(sub) {
				break
			}

			// Scan word
			for i < len(sub) {
				r, size := utf8.DecodeRuneInString(sub[i:])
				if unicode.IsSpace(r) {
					break
				}
				i += size
			}

			if i > start {
				segment := sub[start:i]
				if !yield(segment, utf8.RuneCountInString(segment)) {
					return
				}
				start = i
			}
		}
	}
}

func (b *Block) Parse() {
	b.Content = make([]Span, 0)
	n := len(b.Text)
	start := 0
	style := 0
	skip := false

	for i, r := range b.Text {
		if skip {
			skip = false
			continue
		}

		// Handle Code style special case (no other styles apply inside)
		if style&Code != 0 {
			if r == '`' {
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style &^= Code
				start = i + 1
			}
			continue
		}

		switch r {
		case '*':
			if i+1 < n && b.Text[i+1] == '*' {
				// ** Bold
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Bold
				start = i + 2
				skip = true
			} else {
				// * Italic
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Italic
				start = i + 1
			}
		case '_':
			if i+1 < n && b.Text[i+1] == '_' {
				// __ Underline
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Underline
				start = i + 2
				skip = true
			} else {
				// _ Italic
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Italic
				start = i + 1
			}
		case '~':
			if i+1 < n && b.Text[i+1] == '~' {
				// ~~ Strikethrough
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Strikethrough
				start = i + 2
				skip = true
			}
		case '`':
			// Code
			if i > start {
				b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
			}
			style |= Code
			start = i + 1
		}
	}

	// Flush remaining text
	if start < n {
		b.Content = append(b.Content, Span{Start: start, End: n, Length: n - start, Style: style})
	}
}

func (s *Styled) Parse() {
	s.blocks = make([]Block, 0)
	lines := strings.Split(s.text, "\n")
	var currentBlock *Block

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle Code Block
		if strings.HasPrefix(trimmed, "```") {
			if currentBlock != nil && currentBlock.Type == "code" {
				// End code block
				s.blocks = append(s.blocks, *currentBlock)
				currentBlock = nil
			} else {
				// Start code block
				if currentBlock != nil {
					s.blocks = append(s.blocks, *currentBlock)
				}
				currentBlock = &Block{Type: "code", Text: ""}
			}
			continue
		}

		// content inside code block
		if currentBlock != nil && currentBlock.Type == "code" {
			if currentBlock.Text != "" {
				currentBlock.Text += "\n"
			}
			currentBlock.Text += line
			continue
		}

		if strings.HasPrefix(line, "# ") {
			if currentBlock != nil {
				s.blocks = append(s.blocks, *currentBlock)
			}
			s.blocks = append(s.blocks, Block{Type: "h1", Text: strings.TrimPrefix(line, "# ")})
			currentBlock = nil
			continue
		}

		if strings.HasPrefix(line, "## ") {
			if currentBlock != nil {
				s.blocks = append(s.blocks, *currentBlock)
			}
			s.blocks = append(s.blocks, Block{Type: "h2", Text: strings.TrimPrefix(line, "## ")})
			currentBlock = nil
			continue
		}

		// Handle List Item
		if strings.HasPrefix(line, "- ") {
			if currentBlock != nil {
				s.blocks = append(s.blocks, *currentBlock)
			}
			s.blocks = append(s.blocks, Block{Type: "list", Text: strings.TrimPrefix(line, "- ")})
			currentBlock = nil
			continue
		}

		// Handle Paragraphs
		if trimmed == "" {
			if currentBlock != nil {
				s.blocks = append(s.blocks, *currentBlock)
				currentBlock = nil
			}
			continue
		}

		if currentBlock == nil {
			currentBlock = &Block{Type: "p", Text: line}
		} else {
			if currentBlock.Text != "" {
				currentBlock.Text += " "
			}
			currentBlock.Text += strings.TrimSpace(line)
		}
	}

	if currentBlock != nil {
		s.blocks = append(s.blocks, *currentBlock)
	}

	// Parse inline styles for all blocks except code
	for i := range s.blocks {
		if s.blocks[i].Type != "code" {
			s.blocks[i].Parse()
		}
	}
}

type Styled struct {
	Component
	text    string
	blocks  []Block
	offsetX int
	offsetY int
}

func NewStyled(id, text string) *Styled {
	styled := &Styled{
		Component: Component{id: id},
		text:      text,
	}
	styled.Parse()
	return styled
}

func (s *Styled) Refresh() {
	Redraw(s)
}

func (s *Styled) Render(r *Renderer) {
	x, y, w, h := s.Content()
	for _, block := range s.blocks {
		switch block.Type {
		case "p":
			s.renderP(r, block, x, y, w, h)
		case "list":
			s.renderList(r, block)
		case "code":
			s.renderCode(r, block)
		}
	}
}

func (s *Styled) renderP(r *Renderer, block Block, x, y, w, h int) int {
	cx := x
	cy := y
	for _, span := range block.Content {
		font := ""
		if span.Style&Bold != 0 {
			font += "bold "
		}
		if span.Style&Italic != 0 {
			font += "italic "
		}
		if span.Style&Underline != 0 {
			font += "underline "
		}
		if span.Style&Strikethrough != 0 {
			font += "strikethrough "
		}
		if span.Style&Code != 0 {
			font += "code"
		}
		r.Set("", "", font)
		s.Log(s, "debug", "span %d %d %s", span.Start, span.End, font)
		for word, ww := range span.Words(block.Text) {
			if cx-x+ww > w {
				cx = x
				cy++
			}
			r.Text(cx, cy, word, 0)
			cx = cx + ww
		}
	}
	return cy - y + 1
}

func (s *Styled) renderList(r *Renderer, block Block) {

}

func (s *Styled) renderCode(r *Renderer, block Block) {

}

package zeichenwerk

import (
	"reflect"
	"testing"
)

func TestBlock_Parse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Span
	}{
		{
			name:  "plain text",
			input: "hello world",
			expected: []Span{
				{Start: 0, End: 11, Length: 11, Style: 0},
			},
		},
		{
			name:  "bold",
			input: "hello **world**",
			expected: []Span{
				{Start: 0, End: 6, Length: 6, Style: 0},
				{Start: 8, End: 13, Length: 5, Style: Bold},
			},
		},
		{
			name:  "italic star",
			input: "*italic*",
			expected: []Span{
				{Start: 1, End: 7, Length: 6, Style: Italic},
			},
		},
		{
			name:  "italic underscore",
			input: "_italic_",
			expected: []Span{
				{Start: 1, End: 7, Length: 6, Style: Italic},
			},
		},
		{
			name:  "underline",
			input: "__underlined__",
			expected: []Span{
				{Start: 2, End: 12, Length: 10, Style: Underline},
			},
		},
		{
			name:  "strikethrough",
			input: "~~strike~~",
			expected: []Span{
				{Start: 2, End: 8, Length: 6, Style: Strikethrough},
			},
		},
		{
			name:  "code",
			input: "`code`",
			expected: []Span{
				{Start: 1, End: 5, Length: 4, Style: Code},
			},
		},
		{
			name:  "mixed styles",
			input: "**bold** and *italic*",
			expected: []Span{
				{Start: 2, End: 6, Length: 4, Style: Bold},
				{Start: 8, End: 13, Length: 5, Style: 0},
				{Start: 14, End: 20, Length: 6, Style: Italic},
			},
		},
		{
			name:     "styles without text",
			input:    "****",
			expected: []Span{},
		},
		{
			name:  "code ignores other styles",
			input: "`**not bold**`",
			expected: []Span{
				{Start: 1, End: 13, Length: 12, Style: Code},
			},
		},
		{
			name:  "nested styles",
			input: "**bold *and* italic**",
			expected: []Span{
				{Start: 2, End: 7, Length: 5, Style: Bold},
				{Start: 8, End: 11, Length: 3, Style: Bold | Italic},
				{Start: 12, End: 19, Length: 7, Style: Bold},
			},
		},
		{
			name:  "incomplete markers at end",
			input: "hello **world",
			expected: []Span{
				{Start: 0, End: 6, Length: 6, Style: 0},
				{Start: 8, End: 13, Length: 5, Style: Bold},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := &Block{Text: tt.input}
			block.Parse()

			// Initialize nil slice to empty slice for comparison if needed
			if block.Content == nil {
				block.Content = []Span{}
			}
			if tt.expected == nil {
				tt.expected = []Span{}
			}

			if !reflect.DeepEqual(block.Content, tt.expected) {
				t.Errorf("Block.Parse() = %v, want %v", block.Content, tt.expected)
			}
		})
	}
}

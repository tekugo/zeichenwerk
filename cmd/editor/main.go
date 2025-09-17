package main

import (
	"os"

	. "github.com/tekugo/zeichenwerk"
)

func main() {
	// Create the editor widget
	editor := NewEditor("main-editor")

	// Configure editor settings
	editor.ShowLineNumbers(true)
	editor.SetTabWidth(4)
	editor.UseSpaces(false)
	editor.SetAutoIndent(true)
	editor.SetStyle("", NewStyle("", "").SetCursor("*bar"))

	// Load some sample content
	sampleText := `// Welcome to the Zeichenwerk Editor!
// This is a demonstration of the multi-line text editor widget.

package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
    
    // You can edit this text
    numbers := []int{1, 2, 3, 4, 5}
    
    for i, num := range numbers {
        fmt.Printf("Item %d: %d\n", i, num)
    }
}

// Features demonstrated:
// - Multi-line editing with gap buffers
// - Line numbers
// - Auto-indentation
// - Tab handling
// - Cursor navigation
// - Scrolling for large documents

// Try these keyboard shortcuts:
// - Arrow keys: Navigate cursor
// - Home/End: Beginning/end of line
// - Ctrl+A/Ctrl+E: Beginning/end of document
// - Page Up/Down: Page navigation
// - Enter: New line with auto-indent
// - Tab: Insert tab or spaces
// - Backspace/Delete: Remove characters`

	editor.Load(sampleText)

	// Create a container for the editor
	container := NewBox("editor-container", "Text Editor Demo")
	container.Add(editor)
	editor.SetParent(container)

	// Create UI with theme
	theme := TokyoNightTheme()
	ui, err := NewUI(theme, container, true)
	if err != nil {
		panic(err)
	}

	// Set up event handlers
	editor.On("change", func(w Widget, event string, data ...any) bool {
		// Handle text changes (could implement auto-save, etc.)
		return false
	})

	// Run the application
	if err := ui.Run(); err != nil {
		os.Exit(1)
	}
}

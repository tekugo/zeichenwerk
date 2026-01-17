Based on my analysis of the zeichenwerk codebase, I can provide you with a comprehensive plan for adding selection functionality to the Editor widget. Here's what you need to implement:

## **Current State Analysis** 📋

The Editor widget currently has:
- ✅ Cursor positioning (`line`, `column`)
- ✅ Navigation (arrow keys, page up/down, home/end)
- ✅ Text editing operations
- ✅ Viewport management with scrolling
- ❌ **No selection support** (mentioned in docs but not implemented)

The List widget already has robust selection implementation that we can use as a reference.

## **What You Need to Add for Selection** 🎯

### **1. Selection State Structure**

Add these fields to the `Editor` struct:

```go
type Editor struct {
    BaseWidget
    // ... existing fields ...
    
    // ---- Selection Support ----
    hasSelection  bool  // Whether there's an active selection
    anchorLine    int   // Selection start line
    anchorColumn  int   // Selection start column
    selectLine    int   // Selection end line (cursor line)
    selectColumn  int   // Selection end column (cursor column)
}
```

### **2. Selection Management Methods**

```go
// Selection state management
func (e *Editor) StartSelection() {
    e.hasSelection = true
    e.anchorLine = e.line
    e.anchorColumn = e.column
    e.selectLine = e.line
    e.selectColumn = e.column
}

func (e *Editor) ClearSelection() {
    e.hasSelection = false
    e.Refresh()
}

func (e *Editor) HasSelection() bool {
    return e.hasSelection
}

// Get normalized selection bounds (start always before end)
func (e *Editor) GetSelectionBounds() (startLine, startCol, endLine, endCol int, hasSelection bool) {
    if !e.hasSelection {
        return 0, 0, 0, 0, false
    }
    
    // Normalize selection (ensure start comes before end)
    if e.anchorLine < e.line || (e.anchorLine == e.line && e.anchorColumn <= e.column) {
        return e.anchorLine, e.anchorColumn, e.line, e.column, true
    } else {
        return e.line, e.column, e.anchorLine, e.anchorColumn, true
    }
}

// Get selected text
func (e *Editor) GetSelectedText() string {
    startLine, startCol, endLine, endCol, hasSelection := e.GetSelectionBounds()
    if !hasSelection {
        return ""
    }
    
    if startLine == endLine {
        // Single line selection
        lineText := e.content[startLine].String()
        runes := []rune(lineText)
        if startCol >= len(runes) {
            return ""
        }
        endCol = min(endCol, len(runes))
        return string(runes[startCol:endCol])
    }
    
    // Multi-line selection
    var result strings.Builder
    
    // First line (from startCol to end)
    if startLine < len(e.content) {
        lineText := e.content[startLine].String()
        runes := []rune(lineText)
        if startCol < len(runes) {
            result.WriteString(string(runes[startCol:]))
        }
        result.WriteRune('\n')
    }
    
    // Middle lines (complete lines)
    for line := startLine + 1; line < endLine && line < len(e.content); line++ {
        result.WriteString(e.content[line].String())
        result.WriteRune('\n')
    }
    
    // Last line (from start to endCol)
    if endLine < len(e.content) {
        lineText := e.content[endLine].String()
        runes := []rune(lineText)
        endCol = min(endCol, len(runes))
        if endCol > 0 {
            result.WriteString(string(runes[:endCol]))
        }
    }
    
    return result.String()
}
```

### **3. Enhanced Movement Methods**

Modify existing movement methods to support shift+key selection:

```go
// Modified movement methods that accept extend parameter
func (e *Editor) Left(extend ...bool) {
    extending := len(extend) > 0 && extend[0]
    
    if extending && !e.hasSelection {
        e.StartSelection()
    } else if !extending && e.hasSelection {
        e.ClearSelection()
    }
    
    // Existing movement logic...
    if e.column > 0 {
        e.column--
        e.adjustViewport()
        e.RefreshCursor()
    } else if e.line > 0 {
        e.line--
        e.column = e.content[e.line].Length()
        e.adjustViewport()
        e.Refresh()
    }
    
    // Update selection end if extending
    if extending && e.hasSelection {
        e.selectLine = e.line
        e.selectColumn = e.column
        e.Refresh() // Need full refresh for selection highlight
    }
}

// Similar modifications for Right(), Up(), Down(), Home(), End(), etc.
```

### **4. Enhanced Event Handling**

Update the `Handle` method to detect Shift modifier:

```go
func (e *Editor) Handle(evt tcell.Event) bool {
    switch event := evt.(type) {
    case *tcell.EventKey:
        // Check for shift modifier
        extending := event.Modifiers()&tcell.ModShift != 0
        
        switch event.Key() {
        case tcell.KeyLeft:
            e.Left(extending)
            return true
        case tcell.KeyRight:
            e.Right(extending)
            return true
        case tcell.KeyUp:
            e.Up(extending)
            return true
        case tcell.KeyDown:
            e.Down(extending)
            return true
        case tcell.KeyHome:
            e.Home(extending)
            return true
        case tcell.KeyEnd:
            e.End(extending)
            return true
        case tcell.KeyPgUp:
            e.PageUp(extending)
            return true
        case tcell.KeyPgDn:
            e.PageDown(extending)
            return true
            
        // Selection shortcuts
        case tcell.KeyCtrlA:
            e.SelectAll()
            return true
        case tcell.KeyEscape:
            if e.hasSelection {
                e.ClearSelection()
                return true
            }
            
        // Text operations with selection
        case tcell.KeyBackspace, tcell.KeyBackspace2:
            if e.hasSelection {
                e.DeleteSelection()
            } else {
                e.Delete()
            }
            return true
        case tcell.KeyDelete:
            if e.hasSelection {
                e.DeleteSelection()
            } else {
                e.DeleteForward()
            }
            return true
            
        // ... rest of existing key handling
        }
    }
    return false
}
```

### **5. Text Operations with Selection**

```go
func (e *Editor) SelectAll() {
    if len(e.content) == 0 {
        return
    }
    
    e.hasSelection = true
    e.anchorLine = 0
    e.anchorColumn = 0
    e.selectLine = len(e.content) - 1
    e.selectColumn = e.content[e.selectLine].Length()
    e.Refresh()
}

func (e *Editor) DeleteSelection() {
    if !e.hasSelection {
        return
    }
    
    startLine, startCol, endLine, endCol, _ := e.GetSelectionBounds()
    
    if startLine == endLine {
        // Single line deletion
        buffer := e.content[startLine]
        buffer.Move(startCol)
        for i := startCol; i < endCol; i++ {
            buffer.Delete()
        }
    } else {
        // Multi-line deletion
        // 1. Truncate start line at startCol
        startBuffer := e.content[startLine]
        startBuffer.Move(startCol)
        for startBuffer.Length() > startCol {
            startBuffer.Delete()
        }
        
        // 2. Get remaining text from end line
        endLineText := e.content[endLine].String()
        endRunes := []rune(endLineText)
        if endCol < len(endRunes) {
            remainingText := string(endRunes[endCol:])
            for _, ch := range remainingText {
                startBuffer.Move(startBuffer.Length())
                startBuffer.Insert(ch)
            }
        }
        
        // 3. Remove lines between start and end (inclusive of end)
        e.content = append(e.content[:startLine+1], e.content[endLine+1:]...)
    }
    
    // Position cursor at selection start
    e.line = startLine
    e.column = startCol
    e.ClearSelection()
    e.updateLongestLine()
    e.adjustViewport()
    e.Emit("change")
    e.Refresh()
}

// Override Insert to replace selection
func (e *Editor) Insert(ch rune) {
    if e.disabled {
        return
    }
    
    // Delete selection if it exists
    if e.hasSelection {
        e.DeleteSelection()
    }
    
    // Existing insert logic...
    // (same as current implementation)
}
```

### **6. Rendering Updates**

Update `render-editor.go` to highlight selected text:

```go
// In renderEditor function, add selection highlighting
func (r *Renderer) renderEditor(editor *Editor) {
    // ... existing rendering logic ...
    
    // Get selection bounds for highlighting
    startLine, startCol, endLine, endCol, hasSelection := editor.GetSelectionBounds()
    
    // Render each visible line
    for row := 0; row < contentHeight; row++ {
        lineNum := editor.offsetY + row
        if lineNum >= len(editor.content) {
            break
        }
        
        lineText := editor.content[lineNum].String()
        // ... tab expansion logic ...
        
        // Apply selection highlighting if this line is selected
        if hasSelection && lineNum >= startLine && lineNum <= endLine {
            r.renderLineWithSelection(x, y+row, lineText, lineNum, startLine, startCol, endLine, endCol, editor.offsetX, contentWidth)
        } else {
            r.renderLine(x, y+row, lineText, editor.offsetX, contentWidth)
        }
    }
}

func (r *Renderer) renderLineWithSelection(x, y int, line string, lineNum, startLine, startCol, endLine, endCol, offsetX, width int) {
    runes := []rune(line)
    
    // Calculate selection bounds for this line
    selStart := 0
    selEnd := len(runes)
    
    if lineNum == startLine {
        selStart = startCol
    }
    if lineNum == endLine {
        selEnd = endCol
    }
    
    // Render character by character with selection highlighting
    for i, r := range runes {
        if i < offsetX || i >= offsetX+width {
            continue
        }
        
        screenX := x + i - offsetX
        
        if i >= selStart && i < selEnd {
            // Render with selection style (reverse video)
            r.screen.SetContent(screenX, y, r, nil, tcell.StyleDefault.Reverse(true))
        } else {
            // Render normally
            r.screen.SetContent(screenX, y, r, nil, tcell.StyleDefault)
        }
    }
}
```

### **7. Theme Support**

Add selection styling to theme files:

```go
// In theme files, add selection styles
"editor/selection":     {ForegroundColor: ColorWhite, BackgroundColor: ColorBlue},
"editor/selection:focus": {ForegroundColor: ColorWhite, BackgroundColor: ColorBlue, Bold: true},
```

## **Implementation Priority** 🚀

1. **Phase 1**: Basic selection state and bounds calculation
2. **Phase 2**: Enhanced movement methods with shift support
3. **Phase 3**: Event handling for shift+keys
4. **Phase 4**: Text operations (delete selection, select all)
5. **Phase 5**: Visual selection highlighting in renderer
6. **Phase 6**: Copy/paste integration (if clipboard support exists)

## **Key Design Decisions** 🎯

- **Anchor-based selection**: Similar to most text editors
- **Normalized bounds**: Always ensure start ≤ end for easier processing
- **Shift+movement extends**: Standard behavior users expect
- **Replace selection on insert**: Standard text editor behavior
- **Visual feedback**: Reverse video highlighting like most terminals

**Would you like me to help implement any specific part of this selection system, starting with the basic selection state management, or would you prefer to see a complete implementation of one of the phases?**

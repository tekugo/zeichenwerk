# StatusBar

Based on my analysis of the zeichenwerk TUI library architecture, I'll design a comprehensive StatusBar widget that follows the established patterns and conventions. Here's my suggestion:

## **StatusBar Widget Design** 📊

### **Core Structure**

```go
// StatusBar represents a horizontal status bar widget that displays multiple 
// segments of information along the bottom or top of an application window.
// It provides a flexible way to show application status, context information,
// shortcuts, and other relevant data to users.
//
// Features:
//   - Multiple configurable segments with individual content and styling
//   - Left, center, and right alignment zones
//   - Automatic text truncation and overflow handling
//   - Support for icons, separators, and progress indicators
//   - Dynamic content updates with automatic refresh
//   - Customizable styling per segment through the theme system
//
// Common use cases:
//   - Application status and mode indicators
//   - File information (line numbers, encoding, file size)
//   - Progress indicators and completion status
//   - Keyboard shortcuts and help text
//   - Connection status and system information
type StatusBar struct {
    BaseWidget
    Segments []StatusSegment // Individual segments displayed in the status bar
    Spacing  int            // Pixels/characters between segments
}

// StatusSegment represents an individual content area within the status bar.
// Each segment can contain text, have specific alignment, and maintain its
// own styling and behavior configuration.
type StatusSegment struct {
    ID        string  // Unique identifier for styling and updates
    Text      string  // Content to display in this segment
    Alignment string  // Positioning: "left", "center", "right"
    MinWidth  int     // Minimum width to reserve for this segment
    MaxWidth  int     // Maximum width before truncation (0 = unlimited)
    Weight    float64 // Flex weight for space distribution (0 = fixed size)
    Separator bool    // Whether to show a separator after this segment
    Visible   bool    // Whether this segment is currently visible
}
```

### **Constructor and Core Methods**

```go
// NewStatusBar creates a new status bar widget with default configuration.
// The status bar is initialized with no segments and standard spacing.
//
// Parameters:
//   - id: Unique identifier for the status bar widget
//
// Returns:
//   - *StatusBar: A new status bar widget instance
func NewStatusBar(id string) *StatusBar {
    return &StatusBar{
        BaseWidget: BaseWidget{id: id, focusable: false},
        Segments:   []StatusSegment{},
        Spacing:    1,
    }
}

// AddSegment appends a new segment to the status bar.
// The segment will be positioned according to its alignment setting.
//
// Parameters:
//   - segment: The segment configuration to add
func (sb *StatusBar) AddSegment(segment StatusSegment) {
    segment.Visible = true // Default to visible
    sb.Segments = append(sb.Segments, segment)
    sb.Refresh()
}

// AddText is a convenience method to add a simple text segment.
//
// Parameters:
//   - id: Segment identifier for styling and updates
//   - text: Content to display
//   - alignment: "left", "center", or "right"
func (sb *StatusBar) AddText(id, text, alignment string) {
    sb.AddSegment(StatusSegment{
        ID:        id,
        Text:      text,
        Alignment: alignment,
        Weight:    0, // Fixed size by default
        Visible:   true,
    })
}

// UpdateSegment updates the content of an existing segment by ID.
// This is the primary method for dynamic status updates.
//
// Parameters:
//   - id: The segment identifier to update
//   - text: New content for the segment
//
// Returns:
//   - bool: true if segment was found and updated, false otherwise
func (sb *StatusBar) UpdateSegment(id, text string) bool {
    for i := range sb.Segments {
        if sb.Segments[i].ID == id {
            sb.Segments[i].Text = text
            sb.Refresh()
            return true
        }
    }
    return false
}

// SetSegmentVisible controls the visibility of a segment by ID.
//
// Parameters:
//   - id: The segment identifier
//   - visible: Whether the segment should be visible
func (sb *StatusBar) SetSegmentVisible(id string, visible bool) bool {
    for i := range sb.Segments {
        if sb.Segments[i].ID == id {
            sb.Segments[i].Visible = visible
            sb.Refresh()
            return true
        }
    }
    return false
}

// RemoveSegment removes a segment by ID.
//
// Parameters:
//   - id: The segment identifier to remove
//
// Returns:
//   - bool: true if segment was found and removed, false otherwise
func (sb *StatusBar) RemoveSegment(id string) bool {
    for i, segment := range sb.Segments {
        if segment.ID == id {
            sb.Segments = append(sb.Segments[:i], sb.Segments[i+1:]...)
            sb.Refresh()
            return true
        }
    }
    return false
}

// Clear removes all segments from the status bar.
func (sb *StatusBar) Clear() {
    sb.Segments = []StatusSegment{}
    sb.Refresh()
}
```

### **Layout and Sizing Methods**

```go
// Hint returns the preferred size for the status bar.
// Height is typically 1 line, width expands to fit content.
//
// Returns:
//   - int: Preferred width (sum of all segment content)
//   - int: Preferred height (always 1 for horizontal status bar)
func (sb *StatusBar) Hint() (int, int) {
    totalWidth := 0
    visibleCount := 0
    
    for _, segment := range sb.Segments {
        if !segment.Visible {
            continue
        }
        
        segmentWidth := len(segment.Text)
        if segment.MinWidth > segmentWidth {
            segmentWidth = segment.MinWidth
        }
        
        totalWidth += segmentWidth
        visibleCount++
        
        // Add separator space
        if segment.Separator && visibleCount > 0 {
            totalWidth += 3 // " | "
        }
    }
    
    // Add spacing between segments
    if visibleCount > 1 {
        totalWidth += (visibleCount - 1) * sb.Spacing
    }
    
    return totalWidth, 1
}

// Info returns debugging information about the status bar.
func (sb *StatusBar) Info() string {
    return fmt.Sprintf("statusbar [%d segments]", len(sb.Segments))
}
```

### **Convenience Builder Methods**

```go
// WithLeftText adds a left-aligned text segment (builder pattern).
func (sb *StatusBar) WithLeftText(id, text string) *StatusBar {
    sb.AddText(id, text, "left")
    return sb
}

// WithCenterText adds a center-aligned text segment (builder pattern).
func (sb *StatusBar) WithCenterText(id, text string) *StatusBar {
    sb.AddText(id, text, "center")
    return sb
}

// WithRightText adds a right-aligned text segment (builder pattern).
func (sb *StatusBar) WithRightText(id, text string) *StatusBar {
    sb.AddText(id, text, "right")
    return sb
}

// WithSpacing sets the spacing between segments (builder pattern).
func (sb *StatusBar) WithSpacing(spacing int) *StatusBar {
    sb.Spacing = spacing
    return sb
}
```

### **Advanced Segment Methods**

```go
// SetSegmentWeight sets the flex weight for dynamic sizing.
// Segments with weight > 0 will expand to fill available space.
func (sb *StatusBar) SetSegmentWeight(id string, weight float64) bool {
    for i := range sb.Segments {
        if sb.Segments[i].ID == id {
            sb.Segments[i].Weight = weight
            sb.Refresh()
            return true
        }
    }
    return false
}

// SetSegmentSeparator controls whether a separator is shown after the segment.
func (sb *StatusBar) SetSegmentSeparator(id string, separator bool) bool {
    for i := range sb.Segments {
        if sb.Segments[i].ID == id {
            sb.Segments[i].Separator = separator
            sb.Refresh()
            return true
        }
    }
    return false
}

// GetSegmentText retrieves the current text of a segment.
func (sb *StatusBar) GetSegmentText(id string) (string, bool) {
    for _, segment := range sb.Segments {
        if segment.ID == id {
            return segment.Text, true
        }
    }
    return "", false
}

// GetVisibleSegments returns all currently visible segments.
func (sb *StatusBar) GetVisibleSegments() []StatusSegment {
    visible := make([]StatusSegment, 0, len(sb.Segments))
    for _, segment := range sb.Segments {
        if segment.Visible {
            visible = append(visible, segment)
        }
    }
    return visible
}
```

## **Styling Support** 🎨

The StatusBar would support these style selectors:

```css
/* Overall status bar styling */
statusbar                    /* Base status bar appearance */
statusbar:focus             /* When status bar has focus (rare) */

/* Individual segment styling */
statusbar/segment           /* Default segment appearance */
statusbar/segment.primary   /* Primary/important segments */
statusbar/segment.secondary /* Secondary/less important segments */
statusbar/segment.warning   /* Warning status segments */
statusbar/segment.error     /* Error status segments */
statusbar/segment.success   /* Success status segments */

/* Specific segment by ID */
statusbar/segment#status    /* Segment with ID "status" */
statusbar/segment#progress  /* Segment with ID "progress" */

/* Separators between segments */
statusbar/separator         /* Separator styling */
```

## **Usage Examples** 💡

```go
// Create a status bar with common segments
statusBar := NewStatusBar("main-status").
    WithLeftText("mode", "NORMAL", "left").
    WithLeftText("file", "main.go", "left").
    WithCenterText("status", "Ready", "center").
    WithRightText("position", "1:1", "right").
    WithRightText("encoding", "UTF-8", "right").
    WithSpacing(2)

// Dynamic updates
statusBar.UpdateSegment("mode", "INSERT")
statusBar.UpdateSegment("position", "10:25")
statusBar.UpdateSegment("status", "Saving...")

// Add progress indicator
statusBar.AddSegment(StatusSegment{
    ID:        "progress",
    Text:      "[====    ] 50%",
    Alignment: "right",
    MinWidth:  15,
    Separator: true,
})

// Toggle segment visibility
statusBar.SetSegmentVisible("progress", false)

// Add weighted segment that expands
statusBar.AddSegment(StatusSegment{
    ID:        "info",
    Text:      "Additional information",
    Alignment: "center",
    Weight:    1.0, // Expands to fill space
})
```

## **Key Design Benefits** ✅

1. **Flexible Layout**: Supports left/center/right alignment zones like VS Code
2. **Dynamic Content**: Easy updates without rebuilding the entire bar
3. **Consistent API**: Follows zeichenwerk patterns (BaseWidget, theming, etc.)
4. **Performance**: Efficient updates and minimal redraws
5. **Extensible**: Easy to add icons, progress bars, or custom content later
6. **Responsive**: Handles overflow and truncation gracefully

**Would you like me to implement this StatusBar widget, or would you prefer to see a different approach or additional features like progress indicators or icon support?**

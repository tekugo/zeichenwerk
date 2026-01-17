Based on my analysis of the zeichenwerk codebase, I'll design a comprehensive fluent API for the NavigationBar widget that follows the established patterns. Here's my suggestion:

## **NavigationBar Fluent API Design** 🚀

### **Core Builder Structure**

```go
// NavigationBuilder provides a fluent interface for constructing navigation bars.
// It follows the same pattern as the main Builder but is specialized for hierarchical
// navigation structures with search and code-based navigation capabilities.
//
// Key features:
//   - Method chaining for readable navigation definition
//   - Hierarchical nesting with automatic scope management
//   - Auto-generation of navigation codes
//   - Integration with the main Builder for seamless UI construction
//   - Event handler attachment with fluent syntax
type NavigationBuilder struct {
    nav           *NavigationBar      // The navigation bar being built
    nodeStack     []NavigationNode   // Stack for nested node construction
    currentNode   *NavigationNode    // Currently active node being configured
    pathStack     [][]int            // Path stack for maintaining hierarchy
    codeGenerator func([]int) string // Custom code generation function
}

// NewNavigationBuilder creates a new navigation builder for the specified navigation bar.
//
// Parameters:
//   - nav: The navigation bar to build structure for
//
// Returns:
//   - *NavigationBuilder: A new navigation builder instance
func NewNavigationBuilder(nav *NavigationBar) *NavigationBuilder {
    return &NavigationBuilder{
        nav:           nav,
        nodeStack:     []NavigationNode{},
        currentNode:   nil,
        pathStack:     [][]int{},
        codeGenerator: nav.generateNavigationCode,
    }
}

// Build finalizes the navigation structure and returns the configured navigation bar.
// This method should be called after all navigation items have been defined.
//
// Returns:
//   - *NavigationBar: The fully configured navigation bar
func (nb *NavigationBuilder) Build() *NavigationBar {
    nb.nav.rebuildFilteredItems()
    return nb.nav
}
```

### **Category and Group Management**

```go
// Category creates a new category section in the navigation.
// Categories are non-selectable grouping nodes that organize related items.
//
// Parameters:
//   - title: Display name for the category
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
//
// Example:
//   nav.Category("File Operations").
//       Item("New File", "Create a new file").Action(newFileAction).
//       Item("Open File", "Open existing file").Action(openFileAction).
//   End()
func (nb *NavigationBuilder) Category(title string) *NavigationBuilder {
    category := NavigationNode{
        ID:         nb.generateID(title),
        Title:      title,
        Selectable: false,
        Expanded:   true,
        Children:   []NavigationNode{},
    }
    
    nb.pushNode(category)
    return nb
}

// Group creates a collapsible group within the current context.
// Groups are similar to categories but can be collapsed/expanded by users.
//
// Parameters:
//   - title: Display name for the group
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) Group(title string) *NavigationBuilder {
    group := NavigationNode{
        ID:         nb.generateID(title),
        Title:      title,
        Selectable: false,
        Expanded:   false, // Groups start collapsed
        Children:   []NavigationNode{},
    }
    
    nb.pushNode(group)
    return nb
}

// Section creates a visual section separator with optional title.
// Sections are purely visual and don't affect navigation hierarchy.
//
// Parameters:
//   - title: Optional section title (empty for separator line only)
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) Section(title string) *NavigationBuilder {
    separator := NavigationNode{
        ID:         nb.generateID("section-" + title),
        Title:      title,
        Selectable: false,
        Expanded:   true,
        Children:   []NavigationNode{},
        Data:       map[string]interface{}{"type": "section"},
    }
    
    nb.addToCurrentContext(separator)
    return nb
}
```

### **Item Definition**

```go
// Item creates a selectable navigation item with title and optional description.
//
// Parameters:
//   - title: Display name for the item
//   - description: Optional description text (can be empty)
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
//
// Example:
//   nav.Item("Save File", "Save the current document").
//       Icon("💾").
//       Code("s").
//       Action(saveAction)
func (nb *NavigationBuilder) Item(title, description string) *NavigationBuilder {
    item := NavigationNode{
        ID:          nb.generateID(title),
        Title:       title,
        Description: description,
        Selectable:  true,
        Expanded:    false,
        Children:    []NavigationNode{},
    }
    
    nb.currentNode = &item
    return nb
}

// Submenu creates a navigation item that contains child items.
// This creates a selectable parent item with expandable children.
//
// Parameters:
//   - title: Display name for the submenu
//   - description: Optional description text
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
//
// Example:
//   nav.Submenu("Recent Files", "Recently opened documents").
//       Item("document1.txt", "").
//       Item("document2.txt", "").
//   End()
func (nb *NavigationBuilder) Submenu(title, description string) *NavigationBuilder {
    submenu := NavigationNode{
        ID:          nb.generateID(title),
        Title:       title,
        Description: description,
        Selectable:  true,
        Expanded:    false,
        Children:    []NavigationNode{},
    }
    
    nb.pushNode(submenu)
    return nb
}
```

### **Item Configuration**

```go
// Icon sets the icon for the current item.
//
// Parameters:
//   - icon: Icon character or string (emoji, Unicode, etc.)
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) Icon(icon string) *NavigationBuilder {
    if nb.currentNode != nil {
        nb.currentNode.Icon = icon
    }
    return nb
}

// Code sets a custom navigation code for the current item.
// If not specified, codes are auto-generated based on position.
//
// Parameters:
//   - code: Custom navigation code string
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) Code(code string) *NavigationBuilder {
    if nb.currentNode != nil {
        nb.currentNode.Code = code
    }
    return nb
}

// Data associates custom data with the current item.
// This data can be retrieved when the item is selected or activated.
//
// Parameters:
//   - data: Any custom data to associate with the item
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) Data(data interface{}) *NavigationBuilder {
    if nb.currentNode != nil {
        nb.currentNode.Data = data
    }
    return nb
}

// Action sets the action function to execute when the item is activated.
//
// Parameters:
//   - action: Function to call when item is selected/activated
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) Action(action func()) *NavigationBuilder {
    if nb.currentNode != nil {
        nb.currentNode.Action = action
        nb.addCurrentNodeToContext()
    }
    return nb
}

// Expanded sets whether a group/submenu starts in expanded state.
//
// Parameters:
//   - expanded: true to start expanded, false to start collapsed
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) Expanded(expanded bool) *NavigationBuilder {
    if nb.currentNode != nil {
        nb.currentNode.Expanded = expanded
    } else if len(nb.nodeStack) > 0 {
        nb.nodeStack[len(nb.nodeStack)-1].Expanded = expanded
    }
    return nb
}
```

### **Event Handling**

```go
// OnSelect attaches an event handler for item selection (highlighting).
//
// Parameters:
//   - handler: Function to call when item is selected
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) OnSelect(handler func(*NavigationItem)) *NavigationBuilder {
    nb.nav.On("select", func(w Widget, event string, data ...any) bool {
        if item, ok := data[0].(*NavigationItem); ok {
            handler(item)
        }
        return false
    })
    return nb
}

// OnActivate attaches an event handler for item activation (enter/click).
//
// Parameters:
//   - handler: Function to call when item is activated
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) OnActivate(handler func(*NavigationItem)) *NavigationBuilder {
    nb.nav.On("activate", func(w Widget, event string, data ...any) bool {
        if item, ok := data[0].(*NavigationItem); ok {
            handler(item)
        }
        return false
    })
    return nb
}

// OnNavigate attaches an event handler for navigation code usage.
//
// Parameters:
//   - handler: Function to call when navigation occurs via code
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) OnNavigate(handler func(*NavigationItem, string)) *NavigationBuilder {
    nb.nav.On("navigate", func(w Widget, event string, data ...any) bool {
        if len(data) >= 2 {
            if item, ok := data[0].(*NavigationItem); ok {
                if code, ok := data[1].(string); ok {
                    handler(item, code)
                }
            }
        }
        return false
    })
    return nb
}

// OnSearch attaches an event handler for search operations.
//
// Parameters:
//   - handler: Function to call when search text changes
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) OnSearch(handler func(string, []NavigationItem)) *NavigationBuilder {
    nb.nav.On("search", func(w Widget, event string, data ...any) bool {
        if len(data) >= 2 {
            if searchText, ok := data[0].(string); ok {
                if items, ok := data[1].([]NavigationItem); ok {
                    handler(searchText, items)
                }
            }
        }
        return false
    })
    return nb
}
```

### **Configuration Methods**

```go
// ShowCodes configures whether navigation codes are displayed.
//
// Parameters:
//   - show: true to show codes, false to hide them
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) ShowCodes(show bool) *NavigationBuilder {
    nb.nav.ShowCodes = show
    return nb
}

// ShowBreadcrumbs configures whether breadcrumb navigation is displayed.
//
// Parameters:
//   - show: true to show breadcrumbs, false to hide them
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) ShowBreadcrumbs(show bool) *NavigationBuilder {
    nb.nav.ShowBreadcrumbs = show
    return nb
}

// MaxDepth sets the maximum nesting depth for navigation items.
//
// Parameters:
//   - depth: Maximum depth (0 for unlimited)
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) MaxDepth(depth int) *NavigationBuilder {
    nb.nav.MaxDepth = depth
    return nb
}

// SearchPlaceholder sets the placeholder text for the search input.
//
// Parameters:
//   - placeholder: Placeholder text string
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) SearchPlaceholder(placeholder string) *NavigationBuilder {
    nb.nav.SearchPlaceholder = placeholder
    nb.nav.SearchInput.Placeholder = placeholder
    return nb
}

// CodeGenerator sets a custom function for generating navigation codes.
//
// Parameters:
//   - generator: Function that takes a path and returns a code string
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) CodeGenerator(generator func([]int) string) *NavigationBuilder {
    nb.codeGenerator = generator
    return nb
}
```

### **Hierarchy Management**

```go
// End closes the current context (category, group, or submenu).
// This is used to return to the parent level in the hierarchy.
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) End() *NavigationBuilder {
    nb.popNode()
    return nb
}

// With executes a configuration function within the current context.
// This is useful for complex nested configurations.
//
// Parameters:
//   - fn: Configuration function to execute
//
// Returns:
//   - *NavigationBuilder: Builder instance for method chaining
func (nb *NavigationBuilder) With(fn func(*NavigationBuilder)) *NavigationBuilder {
    fn(nb)
    return nb
}
```

### **Integration with Main Builder**

```go
// Add NavigationBar support to the main Builder
func (b *Builder) NavigationBar(id string) *NavigationBar {
    nav := NewNavigationBar(id)
    b.Add(nav)
    b.Apply(nav)
    return nav
}

// Navigation creates a navigation bar and returns a navigation builder.
// This integrates seamlessly with the main builder fluent API.
//
// Parameters:
//   - id: Unique identifier for the navigation bar
//
// Returns:
//   - *NavigationBuilder: Navigation builder for fluent configuration
func (b *Builder) Navigation(id string) *NavigationBuilder {
    nav := b.NavigationBar(id)
    return NewNavigationBuilder(nav)
}
```

## **Usage Examples** 💡

### **Simple Navigation Structure**

```go
nav := NewNavigationBuilder(NewNavigationBar("main-nav")).
    ShowCodes(true).
    ShowBreadcrumbs(true).
    SearchPlaceholder("Search commands or type code...").
    
    Category("File").
        Item("New File", "Create a new document").
            Icon("📄").Code("n").Action(newFileAction).
        Item("Open File", "Open existing document").
            Icon("📂").Code("o").Action(openFileAction).
        Item("Save", "Save current document").
            Icon("💾").Code("s").Action(saveAction).
    End().
    
    Category("Edit").
        Item("Undo", "Undo last action").
            Icon("↶").Code("u").Action(undoAction).
        Item("Redo", "Redo last undone action").
            Icon("↷").Code("r").Action(redoAction).
        Submenu("Find & Replace", "Search and replace operations").
            Item("Find", "Search in document").Icon("🔍").Action(findAction).
            Item("Replace", "Replace text").Icon("🔄").Action(replaceAction).
        End().
    End().
    
    Section(""). // Visual separator
    
    Category("View").
        Group("Appearance").Expanded(false).
            Item("Theme", "Change color theme").Action(themeAction).
            Item("Font Size", "Adjust font size").Action(fontAction).
        End().
    End().
    
    OnActivate(func(item *NavigationItem) {
        fmt.Printf("Activated: %s (Code: %s)\n", item.FullPath, item.Code)
    }).
    
    Build()
```

### **Complex Hierarchical Navigation**

```go
nav := builder.Navigation("app-nav").
    CodeGenerator(func(path []int) string {
        // Custom code generation - use letters for top level, numbers for sub-levels
        code := ""
        for i, index := range path {
            if i == 0 {
                code += string(rune('a' + index)) // a, b, c, ...
            } else {
                code += strconv.Itoa(index + 1) // 1, 2, 3, ...
            }
        }
        return code
    }).
    
    Category("Application").
        Submenu("Settings", "Application configuration").
            Submenu("Display", "Display preferences").
                Item("Resolution", "Screen resolution").Data("display.resolution").
                Item("Color Depth", "Color bit depth").Data("display.colors").
                Item("Refresh Rate", "Monitor refresh rate").Data("display.refresh").
            End().
            Submenu("Audio", "Audio preferences").
                Item("Volume", "Master volume").Data("audio.volume").
                Item("Output Device", "Audio output").Data("audio.output").
            End().
        End().
    End().
    
    Category("Development").
        Item("Debug Mode", "Toggle debug output").
            Data("dev.debug").
            Action(func() { toggleDebug() }).
        Item("Profiler", "Performance profiling").
            Data("dev.profiler").
            Action(func() { startProfiler() }).
    End().
    
    OnNavigate(func(item *NavigationItem, code string) {
        fmt.Printf("Navigated via code '%s' to: %s\n", code, item.FullPath)
    }).
    
    Build()

// Usage:
// Type "a11" to navigate to Application > Settings > Display > Resolution
// Type "b" to navigate to Development category
// Type "b2" to navigate to Development > Profiler
```

### **Integration with Builder**

```go
ui := NewBuilder(theme).
    Box("main", "Application").
        Flex("layout", "vertical", "start", 0).
            
            Navigation("nav").
                ShowCodes(true).
                Category("Commands").
                    Item("Exit", "Quit application").Code("q").Action(exitApp).
                End().
                Build().
            
            Spacer("spacer", 1).
            
            Editor("editor").
            
        Close().
    Close().
    Build()
```

## **Key Design Benefits** ✅

1. **Consistent API**: Follows zeichenwerk Builder patterns and conventions
2. **Hierarchical Clarity**: Clear nesting with `Category()...End()` structure
3. **Method Chaining**: Fluent interface for readable configuration
4. **Auto-Code Generation**: Smart navigation codes with custom override options
5. **Event Integration**: Seamless event handling attachment
6. **Builder Integration**: Works naturally with the main Builder API
7. **Flexible Configuration**: Rich customization options for appearance and behavior
8. **Type Safety**: Compile-time checking for method chaining

**Would you like me to implement specific parts of this fluent API, such as the core NavigationBuilder structure or the integration with the main Builder class?**

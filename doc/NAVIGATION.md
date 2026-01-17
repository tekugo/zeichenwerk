Based on my analysis of the zeichenwerk codebase, I'll design a comprehensive Navigation widget that combines search functionality with hierarchical numeric navigation, similar to modern web application navigation bars. Here's my design:

## **NavigationBar Widget Design** 🧭

### **Core Structure**

```go
// NavigationBar represents a hierarchical navigation widget that combines text search
// with numeric/letter-based navigation codes for quick access to nested content.
// 
// Features:
//   - Text search input field at the top for filtering items
//   - Hierarchical tree structure with collapsible/expandable nodes
//   - Numeric/letter codes for direct navigation (e.g., "423" = 4th→2nd→3rd)
//   - Visual breadcrumb trail showing current navigation path
//   - Support for icons, categories, and custom item rendering
//   - Keyboard shortcuts for common operations
//   - Event-driven architecture for selection and activation
//
// Navigation Modes:
//   - Search Mode: Type text to filter items by name/content
//   - Code Mode: Type numeric/letter codes for direct navigation
//   - Browse Mode: Use arrow keys to navigate the tree structure
//
// Common use cases:
//   - Application command palette (VS Code style)
//   - File/folder navigation with quick access codes
//   - Menu systems with hierarchical organization
//   - Documentation browsers with section codes
//   - Settings panels with category navigation
type NavigationBar struct {
    BaseWidget
    
    // ---- Core Components ----
    SearchInput    *Input              // Text input for search/navigation codes
    Tree           []NavigationNode    // Hierarchical navigation structure
    Breadcrumbs    []string            // Current navigation path display
    
    // ---- Navigation State ----
    CurrentPath    []int               // Current position in hierarchy [level1, level2, level3]
    SelectedIndex  int                 // Currently highlighted item index
    ExpandedNodes  map[string]bool     // Which nodes are currently expanded
    
    // ---- Display Configuration ----
    ShowBreadcrumbs bool               // Whether to show navigation breadcrumbs
    ShowCodes       bool               // Whether to display navigation codes
    MaxDepth        int                // Maximum nesting level (0 = unlimited)
    ItemHeight      int                // Height per navigation item
    SearchPlaceholder string           // Placeholder text for search input
    
    // ---- Filtering and Search ----
    FilteredItems  []NavigationItem    // Current filtered/visible items
    SearchMode     bool                // Whether currently in search mode
    LastSearch     string              // Last search query for optimization
}

// NavigationNode represents a node in the hierarchical navigation structure.
// Each node can contain child nodes and associated data/actions.
type NavigationNode struct {
    ID          string              // Unique identifier for this node
    Title       string              // Display name for the node
    Description string              // Optional description/subtitle
    Icon        string              // Optional icon character/string
    Code        string              // Navigation code (auto-generated if empty)
    Expanded    bool                // Whether child nodes are visible
    Selectable  bool                // Whether this node can be selected
    Children    []NavigationNode    // Child nodes for hierarchical structure
    Data        interface{}         // Custom data associated with this node
    Action      func()              // Optional action to execute on selection
}

// NavigationItem represents a flattened navigation item used for search results
// and current view rendering. This is generated from the hierarchical structure.
type NavigationItem struct {
    Node        *NavigationNode     // Reference to the original node
    Path        []int               // Path to this item in the hierarchy
    Level       int                 // Nesting level (0 = root)
    FullPath    string              // Human-readable path (e.g., "Settings > Display > Theme")
    SearchText  string              // Searchable text (title + description + path)
    Code        string              // Navigation code for direct access
}
```

### **Constructor and Core Methods**

```go
// NewNavigationBar creates a new navigation bar widget with default configuration.
//
// Parameters:
//   - id: Unique identifier for the navigation bar widget
//
// Returns:
//   - *NavigationBar: A new navigation bar widget instance
func NewNavigationBar(id string) *NavigationBar {
    searchInput := NewInput(id + "-search")
    searchInput.Placeholder = "Search or enter navigation code..."
    
    nav := &NavigationBar{
        BaseWidget:        BaseWidget{id: id, focusable: true},
        SearchInput:       searchInput,
        Tree:              []NavigationNode{},
        Breadcrumbs:       []string{},
        CurrentPath:       []int{},
        ExpandedNodes:     make(map[string]bool),
        ShowBreadcrumbs:   true,
        ShowCodes:         true,
        MaxDepth:          5,
        ItemHeight:        1,
        SearchPlaceholder: "Search or enter navigation code...",
        FilteredItems:     []NavigationItem{},
    }
    
    // Setup search input event handling
    searchInput.On("change", func(w Widget, event string, data ...any) bool {
        nav.handleSearchChange()
        return true
    })
    
    searchInput.On("enter", func(w Widget, event string, data ...any) bool {
        nav.handleSearchEnter()
        return true
    })
    
    return nav
}

// AddNode adds a new navigation node to the specified path in the hierarchy.
//
// Parameters:
//   - path: Path where to add the node (empty = root level)
//   - node: The navigation node to add
//
// Returns:
//   - bool: true if node was added successfully, false otherwise
func (nb *NavigationBar) AddNode(path []int, node NavigationNode) bool {
    // Auto-generate code if not provided
    if node.Code == "" {
        node.Code = nb.generateNavigationCode(path, &node)
    }
    
    if len(path) == 0 {
        // Add to root level
        nb.Tree = append(nb.Tree, node)
        nb.rebuildFilteredItems()
        return true
    }
    
    // Navigate to parent and add child
    parent := nb.getNodeAtPath(path[:len(path)-1])
    if parent != nil {
        parent.Children = append(parent.Children, node)
        nb.rebuildFilteredItems()
        return true
    }
    
    return false
}

// AddCategory is a convenience method to add a category node with children.
//
// Parameters:
//   - title: Category title
//   - icon: Optional icon for the category
//   - children: Child nodes to add to this category
func (nb *NavigationBar) AddCategory(title, icon string, children ...NavigationNode) {
    category := NavigationNode{
        ID:         fmt.Sprintf("category-%s", strings.ToLower(title)),
        Title:      title,
        Icon:       icon,
        Expanded:   true,
        Selectable: false,
        Children:   children,
    }
    nb.AddNode([]int{}, category)
}

// SetSearchText programmatically sets the search input text.
func (nb *NavigationBar) SetSearchText(text string) {
    nb.SearchInput.SetText(text)
    nb.handleSearchChange()
}

// GetCurrentItem returns the currently selected navigation item.
func (nb *NavigationBar) GetCurrentItem() *NavigationItem {
    if nb.SelectedIndex >= 0 && nb.SelectedIndex < len(nb.FilteredItems) {
        return &nb.FilteredItems[nb.SelectedIndex]
    }
    return nil
}
```

### **Navigation Code System**

```go
// generateNavigationCode creates a unique navigation code for a node.
// Codes use alphanumeric characters: 1-9, then a-z for positions 1-35.
func (nb *NavigationBar) generateNavigationCode(path []int, node *NavigationNode) string {
    code := ""
    
    // Build code from path
    for _, index := range path {
        if index < 9 {
            code += strconv.Itoa(index + 1) // 1-9
        } else if index < 35 {
            code += string(rune('a' + index - 9)) // a-z
        } else {
            code += fmt.Sprintf("(%d)", index + 1) // Fallback for large numbers
        }
    }
    
    return code
}

// NavigateByCode navigates directly to an item using its navigation code.
//
// Parameters:
//   - code: Navigation code (e.g., "423" for 4th→2nd→3rd)
//
// Returns:
//   - bool: true if navigation was successful, false otherwise
func (nb *NavigationBar) NavigateByCode(code string) bool {
    path, valid := nb.parseNavigationCode(code)
    if !valid {
        return false
    }
    
    // Validate path exists
    item := nb.getItemAtPath(path)
    if item == nil {
        return false
    }
    
    // Navigate to the item
    nb.CurrentPath = path
    nb.updateSelection()
    nb.updateBreadcrumbs()
    nb.Emit("navigate", item)
    nb.Refresh()
    return true
}

// parseNavigationCode converts a code string into a navigation path.
func (nb *NavigationBar) parseNavigationCode(code string) ([]int, bool) {
    path := []int{}
    
    for _, char := range code {
        var index int
        
        switch {
        case char >= '1' && char <= '9':
            index = int(char - '1') // 0-8 (representing positions 1-9)
        case char >= 'a' && char <= 'z':
            index = int(char - 'a') + 9 // 9-34 (representing positions 10-35)
        case char >= 'A' && char <= 'Z':
            index = int(char - 'A') + 9 // Also accept uppercase
        default:
            return nil, false // Invalid character
        }
        
        path = append(path, index)
    }
    
    return path, len(path) > 0
}
```

### **Search and Filtering**

```go
// handleSearchChange processes changes to the search input.
func (nb *NavigationBar) handleSearchChange() {
    searchText := strings.TrimSpace(nb.SearchInput.Text)
    
    if searchText == "" {
        // Clear search - show full tree
        nb.SearchMode = false
        nb.rebuildFilteredItems()
        nb.SelectedIndex = 0
    } else if nb.isNavigationCode(searchText) {
        // Handle navigation code input
        nb.SearchMode = false
        if nb.NavigateByCode(searchText) {
            nb.SearchInput.SetText("") // Clear after successful navigation
        }
    } else {
        // Handle text search
        nb.SearchMode = true
        nb.filterItemsBySearch(searchText)
        nb.SelectedIndex = 0
    }
    
    nb.LastSearch = searchText
    nb.Refresh()
}

// isNavigationCode determines if the input looks like a navigation code.
func (nb *NavigationBar) isNavigationCode(text string) bool {
    if len(text) == 0 {
        return false
    }
    
    // Check if all characters are valid navigation code characters
    for _, char := range text {
        valid := (char >= '1' && char <= '9') || 
                (char >= 'a' && char <= 'z') || 
                (char >= 'A' && char <= 'Z')
        if !valid {
            return false
        }
    }
    
    return true
}

// filterItemsBySearch filters navigation items based on search text.
func (nb *NavigationBar) filterItemsBySearch(searchText string) {
    searchLower := strings.ToLower(searchText)
    nb.FilteredItems = []NavigationItem{}
    
    allItems := nb.getAllItems()
    for _, item := range allItems {
        // Search in title, description, and full path
        searchableText := strings.ToLower(item.SearchText)
        if strings.Contains(searchableText, searchLower) {
            nb.FilteredItems = append(nb.FilteredItems, item)
        }
    }
}

// rebuildFilteredItems rebuilds the filtered items list from the current tree state.
func (nb *NavigationBar) rebuildFilteredItems() {
    nb.FilteredItems = []NavigationItem{}
    nb.buildItemsFromNodes(nb.Tree, []int{}, 0)
}

// buildItemsFromNodes recursively builds the flattened item list.
func (nb *NavigationBar) buildItemsFromNodes(nodes []NavigationNode, basePath []int, level int) {
    for i, node := range nodes {
        currentPath := append(basePath, i)
        
        // Create item for this node
        item := NavigationItem{
            Node:       &nodes[i],
            Path:       currentPath,
            Level:      level,
            FullPath:   nb.buildFullPath(currentPath),
            SearchText: fmt.Sprintf("%s %s %s", node.Title, node.Description, nb.buildFullPath(currentPath)),
            Code:       node.Code,
        }
        
        nb.FilteredItems = append(nb.FilteredItems, item)
        
        // Add children if expanded
        if node.Expanded && len(node.Children) > 0 {
            nb.buildItemsFromNodes(node.Children, currentPath, level+1)
        }
    }
}
```

### **Event Handling and Navigation**

```go
// Handle processes keyboard events for navigation.
func (nb *NavigationBar) Handle(evt tcell.Event) bool {
    // Always give search input first chance to handle events
    if nb.SearchInput.Handle(evt) {
        return true
    }
    
    switch event := evt.(type) {
    case *tcell.EventKey:
        switch event.Key() {
        case tcell.KeyUp:
            nb.navigateUp()
            return true
        case tcell.KeyDown:
            nb.navigateDown()
            return true
        case tcell.KeyLeft:
            nb.collapseCurrentNode()
            return true
        case tcell.KeyRight:
            nb.expandCurrentNode()
            return true
        case tcell.KeyEnter:
            nb.activateCurrentItem()
            return true
        case tcell.KeyEscape:
            nb.clearSearch()
            return true
        case tcell.KeyHome:
            nb.navigateToFirst()
            return true
        case tcell.KeyEnd:
            nb.navigateToLast()
            return true
        }
        
        // Handle alphanumeric keys for quick navigation
        if event.Rune() != 0 {
            nb.SearchInput.SetFocus(true)
            return nb.SearchInput.Handle(evt)
        }
    }
    
    return false
}

// navigateUp moves selection to the previous item.
func (nb *NavigationBar) navigateUp() {
    if len(nb.FilteredItems) == 0 {
        return
    }
    
    nb.SelectedIndex--
    if nb.SelectedIndex < 0 {
        nb.SelectedIndex = len(nb.FilteredItems) - 1 // Wrap around
    }
    
    nb.updateBreadcrumbs()
    nb.Emit("select", nb.GetCurrentItem())
    nb.Refresh()
}

// navigateDown moves selection to the next item.
func (nb *NavigationBar) navigateDown() {
    if len(nb.FilteredItems) == 0 {
        return
    }
    
    nb.SelectedIndex++
    if nb.SelectedIndex >= len(nb.FilteredItems) {
        nb.SelectedIndex = 0 // Wrap around
    }
    
    nb.updateBreadcrumbs()
    nb.Emit("select", nb.GetCurrentItem())
    nb.Refresh()
}

// activateCurrentItem executes the action for the currently selected item.
func (nb *NavigationBar) activateCurrentItem() {
    item := nb.GetCurrentItem()
    if item == nil {
        return
    }
    
    if item.Node.Action != nil {
        item.Node.Action()
    }
    
    nb.Emit("activate", item)
}

// expandCurrentNode expands the currently selected node if it has children.
func (nb *NavigationBar) expandCurrentNode() {
    item := nb.GetCurrentItem()
    if item == nil || len(item.Node.Children) == 0 {
        return
    }
    
    item.Node.Expanded = true
    nb.rebuildFilteredItems()
    nb.Refresh()
}

// collapseCurrentNode collapses the currently selected node.
func (nb *NavigationBar) collapseCurrentNode() {
    item := nb.GetCurrentItem()
    if item == nil {
        return
    }
    
    item.Node.Expanded = false
    nb.rebuildFilteredItems()
    nb.Refresh()
}
```

### **Breadcrumb and Path Management**

```go
// updateBreadcrumbs updates the breadcrumb trail based on current selection.
func (nb *NavigationBar) updateBreadcrumbs() {
    item := nb.GetCurrentItem()
    if item == nil {
        nb.Breadcrumbs = []string{}
        return
    }
    
    nb.Breadcrumbs = strings.Split(item.FullPath, " > ")
}

// buildFullPath creates a human-readable path string for an item.
func (nb *NavigationBar) buildFullPath(path []int) string {
    if len(path) == 0 {
        return ""
    }
    
    pathParts := []string{}
    currentNodes := nb.Tree
    
    for _, index := range path {
        if index >= 0 && index < len(currentNodes) {
            node := currentNodes[index]
            pathParts = append(pathParts, node.Title)
            currentNodes = node.Children
        }
    }
    
    return strings.Join(pathParts, " > ")
}
```

## **Usage Examples** 💡

```go
// Create navigation bar with hierarchical structure
nav := NewNavigationBar("main-nav")

// Add file operations category
nav.AddCategory("File", "📁",
    NavigationNode{
        ID: "file-new", Title: "New File", Icon: "📄",
        Description: "Create a new file", Selectable: true,
        Action: func() { fmt.Println("Creating new file...") },
    },
    NavigationNode{
        ID: "file-open", Title: "Open File", Icon: "📂",
        Description: "Open existing file", Selectable: true,
        Action: func() { fmt.Println("Opening file...") },
    },
)

// Add settings category with nested structure
nav.AddCategory("Settings", "⚙️",
    NavigationNode{
        ID: "settings-appearance", Title: "Appearance", Icon: "🎨",
        Selectable: false, Expanded: true,
        Children: []NavigationNode{
            {ID: "theme", Title: "Theme", Description: "Choose color theme", Selectable: true},
            {ID: "font", Title: "Font", Description: "Configure font settings", Selectable: true},
        },
    },
    NavigationNode{
        ID: "settings-behavior", Title: "Behavior", Icon: "🔧",
        Selectable: false, Expanded: false,
        Children: []NavigationNode{
            {ID: "auto-save", Title: "Auto Save", Description: "Automatic saving options", Selectable: true},
            {ID: "shortcuts", Title: "Shortcuts", Description: "Keyboard shortcuts", Selectable: true},
        },
    },
)

// Navigation examples:
// Type "21" to go to File > Open File
// Type "221" to go to Settings > Appearance > Font  
// Type "text" to search for items containing "text"

// Event handling
nav.On("activate", func(w Widget, event string, data ...any) bool {
    if item, ok := data[0].(*NavigationItem); ok {
        fmt.Printf("Activated: %s (Code: %s)\n", item.FullPath, item.Code)
    }
    return true
})
```

## **Key Design Benefits** ✅

1. **Dual Input Mode**: Seamlessly switches between search and navigation codes
2. **Hierarchical Structure**: Supports unlimited nesting with visual feedback
3. **Fast Navigation**: Direct access via numeric codes (e.g., "423")
4. **Smart Search**: Searches titles, descriptions, and full paths
5. **Visual Feedback**: Breadcrumbs, icons, and clear hierarchy display
6. **Extensible**: Easy to add custom actions and data to nodes
7. **Keyboard Friendly**: Full keyboard navigation with intuitive shortcuts
8. **Consistent API**: Follows zeichenwerk patterns and conventions

**Would you like me to implement specific parts of this NavigationBar widget, such as the rendering logic, or would you prefer to see additional features like fuzzy search or command palette integration?**

# Roadmap

Based on my comprehensive analysis of the zeichenwerk TUI library, I can provide informed suggestions for the next widgets to implement. The project already has an impressive collection of 20+ widgets, but there are some key gaps that would significantly enhance its capabilities.

## **Current Widget Inventory** 📊

**Implemented widgets:**
- **Input/Data**: Input, Editor, List, Table, Form, Checkbox
- **Display**: Label, Text, Digits, ProgressBar, Separator  
- **Interactive**: Button, Tabs, Switcher, Spinner
- **Layout**: Box, Flex, Grid, Scroller
- **Advanced**: Dialog, Inspector, Custom, Hidden

## **Recommended Next Widgets** 🚀

### **High Priority - Missing Essentials**

1. **TreeView Widget** 🌳
   - **Why**: Essential for file browsers, hierarchical data, navigation menus
   - **Use cases**: File explorers, configuration trees, nested data structures
   - **Features**: Expandable/collapsible nodes, icons, selection, keyboard navigation
   - **Complexity**: Medium - builds on List widget concepts

2. **Menu/MenuBar Widget** 📋
   - **Why**: Standard UI pattern missing from the toolkit
   - **Use cases**: Application menus, context menus, dropdown selections
   - **Features**: Hierarchical menus, keyboard shortcuts, popup positioning
   - **Complexity**: Medium - requires popup management

3. **StatusBar Widget** 📊
   - **Why**: Critical for professional applications
   - **Use cases**: Show app status, progress, shortcuts, current mode
   - **Features**: Multiple segments, icons, progress indicators, right-alignment
   - **Complexity**: Low - mostly layout and text formatting

### **Medium Priority - Enhanced UX**

4. **DatePicker/Calendar Widget** 📅
   - **Why**: Common input pattern for date selection
   - **Use cases**: Forms, scheduling, filtering by date ranges
   - **Features**: Month/year navigation, date selection, range selection
   - **Complexity**: Medium-High - date calculations and navigation

5. **ComboBox/Dropdown Widget** ⬇️
   - **Why**: Space-efficient selection from many options
   - **Use cases**: Settings, forms, filtering
   - **Features**: Searchable, keyboard navigation, custom rendering
   - **Complexity**: Medium - popup management + input field

6. **Slider/Range Widget** 🎚️
   - **Why**: Intuitive numeric input for ranges
   - **Use cases**: Volume controls, progress seeking, numeric ranges
   - **Features**: Horizontal/vertical, step values, labels
   - **Complexity**: Medium - mouse interaction + value calculation

### **Advanced Features**

7. **Chart Widgets** 📈
   - **Bar Chart**: Simple ASCII bar charts for data visualization
   - **Line Chart**: Trend visualization using Unicode characters
   - **Use cases**: Dashboards, monitoring, data analysis
   - **Complexity**: High - data processing + rendering algorithms

8. **FileChooser Widget** 📁
   - **Why**: Essential for file operations
   - **Use cases**: Open/save dialogs, configuration
   - **Features**: Directory navigation, file filtering, preview
   - **Complexity**: High - file system integration

## **Roadmap Priorities** 📋

Based on the README roadmap, I'd recommend this order:

### **Phase 1: Complete Existing Roadmap** ✅
1. **Animator Widget** (already planned) - for smooth transitions
2. **Connect Helper** (already planned) - for widget communication
3. Style system improvements (underline colors, bordercolor)

### **Phase 2: Essential Missing Widgets** 🎯
1. **StatusBar** - Quick win, high impact
2. **TreeView** - Major functionality gap
3. **Menu/MenuBar** - Standard UI pattern

### **Phase 3: Enhanced Input Widgets** 🎨
1. **ComboBox/Dropdown**
2. **DatePicker** 
3. **Slider/Range**

### **Phase 4: Advanced Visualization** 📊
1. **Simple Chart widgets**
2. **FileChooser**

## **Immediate Recommendations** 💡

Given the current state, I'd suggest starting with:

1. **StatusBar Widget** - Low complexity, high value, complements existing widgets perfectly
2. **Animator Widget** - Already on roadmap, would enhance the entire framework
3. **TreeView Widget** - Fills a major functional gap, commonly requested

These would provide the biggest impact while building on the solid foundation you already have.

**What type of widget would you like me to help implement first? Would you prefer to start with a high-impact essential like StatusBar or TreeView, or focus on completing the existing roadmap items like the Animator widget?**

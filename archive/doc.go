// Package zeichenwerk provides a comprehensive terminal user interface framework for building
// interactive console applications with rich widget support, event handling, and theming.
//
// Zeichenwerk (German for "character works") aims to be an easy-to-use, yet modern UI library
// for console applications in Go, inspired by classic terminal interfaces and modern TUI applications.
//
// # Overview
//
// The zeichenwerk package offers a complete toolkit for creating professional terminal applications
// with modern UI patterns. It provides a widget-based architecture with container layouts,
// event-driven programming, and extensive customization options through themes and styling.
//
// # Widgets
//
// The package includes a rich set of built-in widgets for common UI patterns:
//
//   - Input: Single-line text input with editing capabilities and validation
//   - Label: Static text display with alignment and styling options
//   - Button: Interactive buttons with click handlers and focus management
//   - List: Scrollable item lists with selection and keyboard navigation
//   - Text: Multi-line text display with scrolling capabilities
//   - ProgressBar: Visual progress indicators with customizable ranges and styling
//   - Spinner: Animated loading indicators with predefined and custom character sequences
//   - Table: Data table widget with scrolling, navigation, and flexible data providers
//   - Tabs: Tabbed interface for organizing content into multiple views
//   - Switcher: Container that displays one of multiple child widgets
//   - Scroller: Widget that provides scrollable viewport for content larger than display area
//   - Separator: Visual dividers with customizable styles (thin, thick, etc.)
//   - Checkbox: Interactive checkboxes for boolean input with customizable labels
//   - Dialog: Modal dialog containers with keyboard shortcuts and custom actions
//   - Digits: Large ASCII art-style digit display for clocks, counters, and indicators
//   - Editor: Multi-line text editor with gap buffer implementation and editing capabilities
//   - Form: Data binding form container that maps Go structs to UI controls automatically
//   - Hidden: Invisible spacer widget for layout management and spacing control
//   - Inspector: Development and debugging tool for widget introspection and analysis
//
// # Containers
//
// Layout containers organize widgets into structured interfaces:
//
//   - Box: Simple single-widget container with optional borders, titles, and padding
//   - Flex: Flexible layouts with horizontal/vertical orientation and dynamic sizing
//   - Grid: Precise grid-based layouts with cell positioning and spanning
//   - Switcher: Content switching container
//
// # Event System
//
// A robust event-driven architecture enables responsive user interactions:
//
//   - Widget-specific events (click, change, focus, select, activate)
//   - Raw keyboard and mouse event handling
//   - Event propagation and consumption model
//   - Custom event registration and emission
//
// # Styling and Theming
//
// Comprehensive visual customization through a sophisticated CSS-like theming system:
//
//   - Built-in themes: Default, Tokyo Night, Midnight Neon, Nord, Gruvbox (Dark/Light)
//   - CSS-like selector system: type/part.class#id/part:state
//   - Hierarchical style inheritance with CSS-style specificity rules
//   - Color variable system with $ prefix (e.g., "$primary", "$secondary")
//   - Border style registry with Unicode drawing characters
//   - Special Unicode character (rune) management for UI elements
//   - Boolean configuration flags for theme behavior control
//   - Box model styling with margins, padding, borders, and content areas
//   - Dynamic theme switching and custom theme creation
//   - Widget style application with multi-state support (hover, focus, disabled)
//   - Style composition and cascading for maintainable theme hierarchies
//
// # Widget Interface
//
// All UI components implement the Widget interface, providing:
//
//   - Bounds management for positioning and sizing
//   - Event handling for user interactions
//   - Focus management for keyboard navigation
//   - Parent-child relationships for hierarchical layouts
//   - State management and refresh capabilities
//
// # Container Interface
//
// Container widgets extend the Widget interface with:
//
//   - Child widget management and organization
//   - Widget lookup by ID for dynamic updates
//   - Layout algorithms for automatic positioning
//   - Event propagation to child widgets
//
// # Builder Pattern
//
// The Builder provides a fluent interface for constructing complex UIs:
//
//   - Method chaining for readable layout definitions
//   - Theme integration for consistent styling across components
//   - Container stack management for nested layouts and hierarchy
//   - Grid positioning and flexible layout options
//   - Built-in widget creation methods for all supported widget types
//   - Theme switching capabilities for dynamic appearance changes
//   - Automatic widget ID generation and management
//   - Layout container creation (Box, Flex, Grid, Stack) with configuration
//
// # Theme Functions
//
// Built-in theme creation functions provide ready-to-use color schemes:
//
//   - DefaultTheme(): Basic theme with minimal styling
//   - TokyoNightTheme(): Modern dark theme with vibrant colors
//   - MidnightNeonTheme(): High-contrast neon aesthetic for dark environments
//   - NordTheme(): Arctic-inspired theme with cool blues and subtle accents
//   - GruvboxDarkTheme(): Retro groove color scheme with warm, earthy colors
//   - GruvboxLightTheme(): Light variant of Gruvbox with inverted color relationships
//   - NewMapTheme(): Creates empty theme for custom configuration
//
// # Utility Functions
//
// The package provides utility functions for common operations:
//
//   - FindUI: Traverses widget hierarchy to find the root UI instance
//   - HandleInputEvent: Simplified Input widget event handling with type safety
//   - HandleKeyEvent: Raw keyboard event processing with tcell.EventKey access
//   - HandleListEvent: List widget event management for selection and activation
//   - Redraw: Queues individual widgets for redraw operations
//   - Update: Generic widget content updates with automatic type detection
//   - WidgetType: Runtime type introspection returning clean type names
//   - WidgetDetails: Comprehensive widget debugging and state information
//   - With: Type-safe widget operations with generic type constraints
//
// # Performance Considerations
//
// The zeichenwerk framework is designed for efficiency in terminal environments:
//
//   - Minimal screen updates through dirty region tracking
//   - Efficient event handling with consumption model
//   - Lazy rendering for off-screen widgets
//   - Memory-conscious widget lifecycle management
//
// # Terminal Compatibility
//
// The package supports a wide range of terminal environments:
//
//   - Modern terminal emulators with full color support
//   - Legacy terminals with limited color palettes
//   - Various terminal sizes and resize handling
//   - Cross-platform compatibility (Linux, macOS, Windows)
//
// # Dependencies
//
// The zeichenwerk package builds on the tcell library for low-level terminal operations:
//
//   - github.com/gdamore/tcell/v2: Terminal cell manipulation and event handling
//
// This provides robust terminal abstraction with wide compatibility and
// comprehensive input/output capabilities.
package zeichenwerk

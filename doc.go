// Package zeichenwerk provides a comprehensive terminal user interface framework for building
// interactive console applications with rich widget support, event handling, and theming.
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
//   - Input: Single-line text input with editing capabilities, password masking, and validation
//   - Label: Static text display with alignment options
//   - Button: Interactive buttons with click handlers and styling
//   - List: Scrollable item lists with multi-selection and keyboard navigation
//   - Text: Multi-line text display with scrolling and content management
//   - ProgressBar: Visual progress indicators with customizable styling
//   - Viewport: Scrollable content containers for large data sets
//
// # Containers
//
// Layout containers organize widgets into structured interfaces:
//
//   - Box: Simple single-widget container with borders and padding
//   - Flex: Flexible layouts with horizontal/vertical orientation and dynamic sizing
//   - Grid: Precise grid-based layouts with cell positioning and spanning
//   - Stack: Layered widget management for overlays and modal dialogs
//
// # Event System
//
// A robust event-driven architecture enables responsive user interactions:
//
//   - Widget-specific events (change, focus, select, activate)
//   - Raw keyboard and mouse event handling
//   - Event propagation and consumption model
//   - Custom event registration and emission
//
// # Styling and Theming
//
// Comprehensive visual customization through themes and styles:
//
//   - Built-in themes (Default, Tokyo Night) with consistent color schemes
//   - CSS-like styling system with classes and inheritance
//   - Border styles and decorative elements
//   - Color management with terminal compatibility
//   - Dynamic theme switching and custom theme creation
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
//   - Theme integration for consistent styling
//   - Container stack management for nested layouts
//   - Grid positioning and flexible layout options
//
// The package provides utility functions for common operations:
//
//   - HandleInputEvent: Simplified input event handling
//   - HandleKeyEvent: Raw keyboard event processing
//   - HandleListEvent: List-specific event management
//   - Update: Generic widget content updates
//   - WidgetType: Runtime type introspection
//   - WidgetDetails: Comprehensive widget debugging information
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

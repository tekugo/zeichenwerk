# Guidelines

## Normative Keywords

- **MUST**: mandatory requirement
- **SHOULD**: recommended but optional with justification
- **MAY**: optional

## Core Principles

- MUST follow idiomatic Go patterns
- MUST minimize dependencies
- MUST use explicit error handling
- SHOULD keep functions <= 50 lines
- SHOULD prefer composition over inheritance patterns

## File Formats

- MUST use Markdown for documentation files

## Libraries

- MUST use log/slog for logging

## Project Overview

- This is a TUI component library based on tcell/v3
- Widget is the interface for all widgets
- Component is the base class implementing Widget including style rendering
- Style is the main class for styling components
- Styles are hierarchical and CSS-like
- Containers are extended components, which can include widgets
- Containers are responsible for layout
- Rendering in encapsulated by the Renderer interface, no tcell in rendering
- Themes (colors, characters, styling) are supported by the renderer
- UI is the main class for running the UI, processing events and rendering
- UI is the root class for all containers and widgets
- UI takes up the whole screen
- Builder provides a fluent API to build UIs and style components
- All controls SHOULD be added to Builder
- New components get builder functions
- Containers must be added to Add to be supported
- All widgets must be added to Apply to get the styling from the theme

## Project Structure

```
zeichenwerk/
+- cmd/        # Command line tools
|  +- demo/    # Demo application
+- doc/        # Documentation
+- archive/    # Old version (separate Go project, ignore)
```

## Clean Code

- MUST not create getter/setter boilerplate
- MUST NOT use `Get` prefixes for property accessors

## Documentation

- MUST provide doc.go per package directory
- MUST document all exported symbols
- Inline comments MUST explain _Why_, not _What_
- Documentation should be short and concise, but describe parameters and return values
- Examples SHOULD only be part of doc.go

## Error Handling

- MUST wrap errors: `fmt.Errorf("context: %w", err)`
- MUST define sentinel errors for common error cases

## Naming

- Exported: CamelCase
- Unexported: camelCase
- Packages: short, lowercase, no underscores

## Logging

- MUST use structured logging
- MUST use log/slog

## Formatting

- MUST use `go fmt` on every go file

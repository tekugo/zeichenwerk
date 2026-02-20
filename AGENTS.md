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
- Inline comments MUST explain *Why*, not *What*

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

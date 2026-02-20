# Form System Overview

The zeichenwerk form system provides automatic data binding between Go structs
and form controls, enabling rapid development of data entry interfaces.

## Quick Reference

### Core Components

- **Form** - Container that manages data binding with Go structs
- **FormGroup** - Organizes related fields with consistent layout
- **FormField** - Individual field with label and control

### Builder Methods

```go
// Create form with data binding
builder.Form(id, title, structPointer)

// Create form group for field organization  
builder.Group(id, title, groupName, placement, spacing)
```

### Struct Tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `label` | Field label | `label:"Full Name"` |
| `width` | Control width | `width:"30"` |
| `control` | Control type | `control:"checkbox"` |
| `group` | Group assignment | `group:"personal"` |
| `line` | Line number | `line:"1"` |
| `readOnly` | Read-only field | `readOnly:"true"` |

## Basic Usage

```go
type User struct {
    Name  string `label:"Full Name" width:"30"`
    Email string `label:"Email Address" width:"40"`
    Admin bool   `label:"Administrator"`
}

user := &User{}
builder.Form("user-form", "User Registration", user).
    Group("info", "", "", "vertical", 1).
    End()
```

## Control Types

- **input** - Text input (default for strings)
- **checkbox** - Boolean checkbox (default for bools)  
- **password** - Hidden text input

## Layout Options

- **horizontal** - Labels left of controls
- **vertical** - Labels above controls

For detailed documentation and examples, see [FORMS.md](FORMS.md).

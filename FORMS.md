# Form Handling Tutorial

This tutorial covers how to create and manage forms in the zeichenwerk TUI framework. Forms provide automatic data binding between Go structs and form controls, making it easy to create data entry interfaces.

## Table of Contents

- [Basic Concepts](#basic-concepts)
- [Creating Simple Forms](#creating-simple-forms)
- [Struct Tags](#struct-tags)
- [Form Groups](#form-groups)
- [Layout Options](#layout-options)
- [Event Handling](#event-handling)
- [Advanced Examples](#advanced-examples)
- [Best Practices](#best-practices)

## Basic Concepts

### Form Components

The zeichenwerk form system consists of three main components:

1. **Form** (`*Form`) - The container that manages data binding and holds form groups
2. **FormGroup** (`*FormGroup`) - Organizes related fields with consistent labeling and layout
3. **FormField** (`*FormField`) - Individual field consisting of a label and control widget

### Data Binding

Forms work with Go structs using reflection and struct tags. The struct must be passed as a pointer to enable two-way data binding:

```go
type User struct {
    Name  string `label:"Full Name" width:"30"`
    Email string `label:"Email Address" width:"40"`
    Admin bool   `label:"Administrator"`
}

user := &User{} // Must be a pointer
form := NewForm("user-form", "User Registration", user)
```

## Creating Simple Forms

### Basic Form Creation

```go
// Define your data structure
type LoginData struct {
    Username string `label:"Username" width:"20"`
    Password string `label:"Password" width:"20" control:"password"`
    Remember bool   `label:"Remember me"`
}

// Create form using the builder pattern
func createLoginForm() *UI {
    data := &LoginData{}
    
    return NewBuilder(TokyoNightTheme()).
        Form("login-form", "Login", data).
            Group("credentials", "", "", "vertical", 1).
            End().
        End().
        Build()
}
```

### Manual Form Construction

You can also create forms manually for more control:

```go
func createManualForm() *Form {
    data := &LoginData{}
    form := NewForm("login", "Login", data)
    
    group := NewFormGroup("fields", "", "horizontal")
    group.Add(0, "Username", NewInput("username"))
    group.Add(1, "Password", NewInput("password"))
    
    form.Add(group)
    return form
}
```

## Struct Tags

Struct tags control how form fields are generated and displayed:

### Available Tags

| Tag | Description | Example | Default |
|-----|-------------|---------|---------|
| `label` | Display label for the field | `label:"Full Name"` | Field name |
| `width` | Control width in characters | `width:"30"` | 10 |
| `control` | Control type | `control:"checkbox"` | Auto-detected |
| `group` | Group name for organization | `group:"personal"` | "" (ungrouped) |
| `line` | Line number within group | `line:"1"` | Auto-increment |
| `readOnly` | Make field read-only | `readOnly:"true"` | false |

### Control Types

- `input` - Text input field (default for strings)
- `checkbox` - Checkbox control (default for booleans)
- `password` - Password input field (hidden text)

### Example with All Tags

```go
type UserProfile struct {
    // Basic information group
    FirstName string `group:"basic" label:"First Name" width:"20" line:"0"`
    LastName  string `group:"basic" label:"Last Name" width:"20" line:"0"`
    Email     string `group:"basic" label:"Email Address" width:"40" line:"1"`
    
    // Settings group
    IsAdmin   bool   `group:"settings" label:"Administrator"`
    IsActive  bool   `group:"settings" label:"Active Account"`
    
    // Hidden or special fields
    ID        string `label:"-"` // Hidden field
    CreatedAt string `readOnly:"true" label:"Created" width:"20"`
}
```

## Form Groups

Form groups organize related fields and control their layout:

### Group Creation

```go
builder.Form("user-form", "User", &user).
    Group("basic-info", "Basic Information", "basic", "horizontal", 1).
    End().
    Group("settings", "Account Settings", "settings", "vertical", 1).
    End()
```

### Parameters

- **id**: Unique identifier for the group
- **title**: Display title (shown in border if styled)
- **name**: Group name to match struct field `group` tags
- **placement**: "horizontal" or "vertical" label placement
- **spacing**: Vertical spacing between lines

## Layout Options

### Horizontal Layout

Labels appear to the left of controls, aligned in columns:

```go
group := NewFormGroup("info", "Information", "horizontal")
// Results in: [Label] [Control] [Label] [Control]
```

### Vertical Layout

Labels appear above controls:

```go
group := NewFormGroup("info", "Information", "vertical")
// Results in:
// [Label]
// [Control]
// [Label]
// [Control]
```

### Multi-field Lines

Multiple fields can share the same line:

```go
type Name struct {
    First string `line:"0" label:"First Name" width:"15"`
    Last  string `line:"0" label:"Last Name" width:"15"`
    Title string `line:"1" label:"Title" width:"10"`
}
```

## Event Handling

### Automatic Data Binding

Form controls automatically update the bound struct when values change:

```go
// Changes to form controls automatically update the user struct
user := &User{}
form := NewForm("user-form", "User", user)

// After user interaction, user.Name will contain the input value
fmt.Printf("User entered: %s\n", user.Name)
```

### Custom Event Handlers

You can add custom event handlers for form interactions:

```go
builder.Form("user-form", "User", &user).
    Group("info", "", "", "vertical", 1).
    End()

// Add custom save button handler
builder.Find("save-button").On("click", func(widget Widget, event string, data ...any) bool {
    // user struct is automatically updated
    fmt.Printf("Saving user: %+v\n", user)
    
    // Validate data
    if user.Name == "" {
        // Show error message
        return false
    }
    
    // Save to database, etc.
    return true
})
```

## Advanced Examples

### Complex Form with Multiple Groups

```go
type Employee struct {
    // Personal Information
    FirstName string `group:"personal" label:"First Name" width:"20" line:"0"`
    LastName  string `group:"personal" label:"Last Name" width:"20" line:"0"`
    Email     string `group:"personal" label:"Email" width:"40" line:"1"`
    Phone     string `group:"personal" label:"Phone" width:"20" line:"1"`
    
    // Employment Details
    Department string `group:"employment" label:"Department" width:"30"`
    Position   string `group:"employment" label:"Position" width:"30"`
    StartDate  string `group:"employment" label:"Start Date" width:"15" line:"1"`
    Salary     string `group:"employment" label:"Salary" width:"15" line:"1"`
    
    // Permissions
    CanEdit   bool `group:"permissions" label:"Can Edit"`
    CanDelete bool `group:"permissions" label:"Can Delete"`
    IsManager bool `group:"permissions" label:"Manager"`
}

func createEmployeeForm() *UI {
    employee := &Employee{}
    
    return NewBuilder(TokyoNightTheme()).
        Flex("main", "vertical", "stretch", 0).
            Form("employee-form", "Employee Information", employee).
                Group("personal-group", "Personal Information", "personal", "horizontal", 1).
                    Border("", "round").Padding(1).
                End().
                Group("employment-group", "Employment Details", "employment", "horizontal", 1).
                    Border("", "round").Padding(1).
                End().
                Group("permissions-group", "Permissions", "permissions", "vertical", 1).
                    Border("", "round").Padding(1).
                End().
            End().
            Flex("buttons", "horizontal", "end", 1).Padding(1).
                Button("save", "Save").
                Button("cancel", "Cancel").
            End().
        End().
        Build()
}
```

### Form with Validation

```go
type RegistrationForm struct {
    Username string `label:"Username" width:"20"`
    Email    string `label:"Email" width:"30"`
    Password string `label:"Password" width:"20" control:"password"`
    Confirm  string `label:"Confirm Password" width:"20" control:"password"`
    Terms    bool   `label:"I agree to the terms and conditions"`
}

func createRegistrationWithValidation() *UI {
    form := &RegistrationForm{}
    
    builder := NewBuilder(TokyoNightTheme()).
        Flex("main", "vertical", "stretch", 1).
            Form("registration", "Registration", form).
                Group("fields", "", "", "vertical", 1).
                End().
            End().
            Label("error-message", "", 0).Foreground("", "red").
            Flex("buttons", "horizontal", "end", 1).
                Button("register", "Register").
                Button("cancel", "Cancel").
            End().
        End()
    
    // Add validation to register button
    builder.Find("register").On("click", func(widget Widget, event string, data ...any) bool {
        ui := FindUI(widget)
        errorLabel := ui.Find("error-message", false).(*Label)
        
        // Validate form
        if form.Username == "" {
            errorLabel.SetText("Username is required")
            return false
        }
        if form.Password != form.Confirm {
            errorLabel.SetText("Passwords do not match")
            return false
        }
        if !form.Terms {
            errorLabel.SetText("You must accept the terms and conditions")
            return false
        }
        
        // Clear error and proceed
        errorLabel.SetText("")
        // Process registration...
        return true
    })
    
    return builder.Build()
}
```

## Best Practices

### 1. Use Descriptive Labels

```go
// Good
`label:"Email Address"`
`label:"Phone Number"`

// Avoid
`label:"Email"`
`label:"Phone"`
```

### 2. Set Appropriate Widths

```go
// Consider content length
`label:"First Name" width:"20"`          // Name fields
`label:"Email Address" width:"40"`       // Email addresses
`label:"ZIP Code" width:"10"`            // Short codes
```

### 3. Group Related Fields

```go
type User struct {
    // Group personal information
    Name  string `group:"personal" label:"Full Name"`
    Email string `group:"personal" label:"Email"`
    
    // Group preferences separately
    Theme     string `group:"preferences" label:"Theme"`
    Language  string `group:"preferences" label:"Language"`
}
```

### 4. Use Consistent Layouts

```go
// Be consistent within each group
builder.Form("user", "User", &user).
    Group("personal", "Personal", "personal", "horizontal", 1).    // Horizontal
    End().
    Group("address", "Address", "address", "horizontal", 1).       // Horizontal
    End().
    Group("preferences", "Preferences", "preferences", "vertical", 1). // Vertical for checkboxes
    End()
```

### 5. Handle Validation Gracefully

```go
// Provide clear error messages
func validateUser(user *User, errorLabel *Label) bool {
    if user.Email == "" {
        errorLabel.SetText("Email address is required")
        return false
    }
    if !strings.Contains(user.Email, "@") {
        errorLabel.SetText("Please enter a valid email address")
        return false
    }
    errorLabel.SetText("") // Clear errors
    return true
}
```

### 6. Use Meaningful Field Names

```go
// Good - matches common form patterns
type ContactForm struct {
    FirstName string `label:"First Name"`
    LastName  string `label:"Last Name"`
    Email     string `label:"Email Address"`
    Subject   string `label:"Subject"`
    Message   string `label:"Message"`
}
```

This completes the comprehensive forms tutorial for the zeichenwerk framework. The form system provides powerful automatic data binding while remaining flexible for custom scenarios.
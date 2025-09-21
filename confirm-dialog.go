package zeichenwerk

import "github.com/gdamore/tcell/v2"

// ConfirmDialog represents a modal confirmation dialog with customizable content
// and actions. It provides a standardized way to ask users for confirmation
// before performing destructive or important actions.
//
// Features:
//   - Customizable title, message, and button text
//   - Icon support for different dialog types (warning, question, info, error)
//   - Flexible button configuration (OK/Cancel, Yes/No, custom buttons)
//   - Keyboard shortcuts (Enter for confirm, Escape for cancel)
//   - Callback functions for user actions
//   - Auto-sizing based on content
//   - Theme-aware styling
//
// Usage Examples:
//
//	// Simple confirmation
//	ConfirmDialog(ui, "Delete File", "Are you sure you want to delete this file?", func(confirmed bool) {
//		if confirmed {
//			deleteFile()
//		}
//	})
//
//	// Custom confirmation with options
//	NewConfirmDialog("Save Changes").
//		SetMessage("You have unsaved changes. What would you like to do?").
//		SetIcon(IconQuestion).
//		SetButtons("Save", "Discard", "Cancel").
//		SetCallback(func(result ConfirmResult) {
//			switch result.Button {
//			case 0: saveChanges()
//			case 1: discardChanges()
//			case 2: // cancel - do nothing
//			}
//		}).
//		Show(ui)
type ConfirmDialog struct {
	title      string              // Dialog title text
	message    string              // Main message content
	icon       ConfirmIcon         // Icon type to display
	buttons    []string            // Button labels
	callback   ConfirmCallback     // Function to call with result
	defaultBtn int                 // Index of default button (focused initially)
	cancelBtn  int                 // Index of cancel button (Escape key)
	width      int                 // Fixed width (0 = auto-size)
	height     int                 // Fixed height (0 = auto-size)
	modal      bool                // Whether to block input to other widgets
	container  Container           // Built dialog container
}

// ConfirmIcon represents the type of icon to display in the dialog
type ConfirmIcon int

const (
	IconNone     ConfirmIcon = iota // No icon
	IconQuestion                    // Question mark (?)
	IconWarning                     // Warning/exclamation (!)
	IconError                       // Error/X symbol (✗)
	IconInfo                        // Information (i)
	IconSuccess                     // Success/checkmark (✓)
)

// ConfirmResult contains the result of a confirmation dialog
type ConfirmResult struct {
	Confirmed bool   // True if user confirmed (clicked default button or Enter)
	Button    int    // Index of the button that was clicked
	Label     string // Text of the button that was clicked
	Cancelled bool   // True if dialog was cancelled (Escape key)
}

// ConfirmCallback is the function signature for confirmation callbacks
type ConfirmCallback func(result ConfirmResult)

// NewConfirmDialog creates a new confirmation dialog with the specified title.
// The dialog is not shown until Show() is called.
//
// Parameters:
//   - title: The title text to display in the dialog header
//
// Returns:
//   - *ConfirmDialog: A new dialog instance ready for configuration
//
// Default Configuration:
//   - No icon
//   - Empty message (must be set with SetMessage)
//   - OK/Cancel buttons
//   - Default button: OK (index 0)
//   - Cancel button: Cancel (index 1)
//   - Auto-sizing enabled
//   - Modal behavior enabled
func NewConfirmDialog(title string) *ConfirmDialog {
	return &ConfirmDialog{
		title:      title,
		message:    "",
		icon:       IconNone,
		buttons:    []string{"OK", "Cancel"},
		defaultBtn: 0,
		cancelBtn:  1,
		width:      0,
		height:     0,
		modal:      true,
	}
}

// SetMessage sets the main message content of the dialog.
// The message can be multi-line and will be automatically wrapped.
//
// Parameters:
//   - message: The message text to display
//
// Returns:
//   - *ConfirmDialog: The dialog instance for method chaining
func (cd *ConfirmDialog) SetMessage(message string) *ConfirmDialog {
	cd.message = message
	return cd
}

// SetIcon sets the icon to display in the dialog.
// The icon appears next to the message and helps convey the dialog's purpose.
//
// Parameters:
//   - icon: The icon type to display
//
// Returns:
//   - *ConfirmDialog: The dialog instance for method chaining
func (cd *ConfirmDialog) SetIcon(icon ConfirmIcon) *ConfirmDialog {
	cd.icon = icon
	return cd
}

// SetButtons configures the buttons displayed in the dialog.
// The first button is considered the default (Enter key).
// The last button is considered the cancel button (Escape key).
//
// Parameters:
//   - buttons: Variable number of button label strings
//
// Returns:
//   - *ConfirmDialog: The dialog instance for method chaining
//
// Examples:
//   dialog.SetButtons("Yes", "No")
//   dialog.SetButtons("Save", "Don't Save", "Cancel")
//   dialog.SetButtons("Delete", "Cancel")
func (cd *ConfirmDialog) SetButtons(buttons ...string) *ConfirmDialog {
	cd.buttons = buttons
	if len(buttons) > 0 {
		cd.defaultBtn = 0
		cd.cancelBtn = len(buttons) - 1
	}
	return cd
}

// SetDefaultButton sets which button should be focused by default.
// This button will be activated when the user presses Enter.
//
// Parameters:
//   - index: Zero-based index of the default button
//
// Returns:
//   - *ConfirmDialog: The dialog instance for method chaining
func (cd *ConfirmDialog) SetDefaultButton(index int) *ConfirmDialog {
	if index >= 0 && index < len(cd.buttons) {
		cd.defaultBtn = index
	}
	return cd
}

// SetCancelButton sets which button should be activated when Escape is pressed.
//
// Parameters:
//   - index: Zero-based index of the cancel button
//
// Returns:
//   - *ConfirmDialog: The dialog instance for method chaining
func (cd *ConfirmDialog) SetCancelButton(index int) *ConfirmDialog {
	if index >= 0 && index < len(cd.buttons) {
		cd.cancelBtn = index
	}
	return cd
}

// SetSize sets a fixed size for the dialog.
// If width or height is 0, auto-sizing is used for that dimension.
//
// Parameters:
//   - width: Fixed width in characters (0 = auto-size)
//   - height: Fixed height in lines (0 = auto-size)
//
// Returns:
//   - *ConfirmDialog: The dialog instance for method chaining
func (cd *ConfirmDialog) SetSize(width, height int) *ConfirmDialog {
	cd.width = width
	cd.height = height
	return cd
}

// SetModal sets whether the dialog should be modal.
// Modal dialogs capture all input and prevent interaction with the main UI.
//
// Parameters:
//   - modal: True for modal behavior, false to allow main UI interaction
//
// Returns:
//   - *ConfirmDialog: The dialog instance for method chaining
func (cd *ConfirmDialog) SetModal(modal bool) *ConfirmDialog {
	cd.modal = modal
	return cd
}

// SetCallback sets the function to call when the dialog is closed.
// The callback receives a ConfirmResult with details about how the dialog was closed.
//
// Parameters:
//   - callback: Function to call with the dialog result
//
// Returns:
//   - *ConfirmDialog: The dialog instance for method chaining
func (cd *ConfirmDialog) SetCallback(callback ConfirmCallback) *ConfirmDialog {
	cd.callback = callback
	return cd
}

// Show displays the confirmation dialog in the specified UI.
// The dialog is shown as a popup layer and captures focus.
//
// Parameters:
//   - ui: The UI instance to show the dialog in
func (cd *ConfirmDialog) Show(ui *UI) {
	cd.buildDialog(ui.Theme())
	
	// Calculate size if auto-sizing
	width := cd.width
	height := cd.height
	if width == 0 || height == 0 {
		w, h := cd.container.Hint()
		if width == 0 {
			width = max(40, min(80, w)) // Reasonable bounds
		}
		if height == 0 {
			height = max(8, min(20, h)) // Reasonable bounds
		}
	}
	
	ui.Popup(-1, -1, width, height, cd.container)
}

// buildDialog constructs the dialog UI using the builder pattern
func (cd *ConfirmDialog) buildDialog(theme Theme) {
	builder := NewBuilder(theme)
	
	// Main dialog container
	builder.Class("popup").
		Flex("confirm-dialog", "vertical", "stretch", 0).
		
		// Title bar
		Label("title", cd.title, 0).Padding(1, 2).Background("", theme.Get("popup#title").Background).
		
		// Content area with icon and message
		Flex("content", "horizontal", "start", 1).Padding(1, 2).Hint(0, -1)
	
	// Add icon if specified
	if cd.icon != IconNone {
		iconText := cd.getIconText()
		builder.Label("icon", iconText, 3).Font("", "bold")
	}
	
	// Add message
	builder.Label("message", cd.message, 0).Hint(-1, 0).
		End(). // End content flex
		
		// Button area
		Flex("buttons", "horizontal", "end", 1).Padding(1, 2, 1, 1)
	
	// Add buttons
	for i, buttonText := range cd.buttons {
		buttonId := fmt.Sprintf("btn-%d", i)
		builder.Class("popup").Button(buttonId, buttonText)
		
		// Set up button click handler
		cd.setupButtonHandler(builder, buttonId, i)
	}
	
	builder.Class("").End(). // End buttons flex
		End() // End main dialog flex
	
	cd.container = builder.Container()
	
	// Set up keyboard handlers
	cd.setupKeyboardHandlers()
	
	// Focus the default button
	if cd.defaultBtn >= 0 && cd.defaultBtn < len(cd.buttons) {
		defaultBtnId := fmt.Sprintf("btn-%d", cd.defaultBtn)
		if btn := cd.container.Find(defaultBtnId, true); btn != nil {
			btn.SetFocused(true)
		}
	}
}

// setupButtonHandler sets up click handler for a button
func (cd *ConfirmDialog) setupButtonHandler(builder *Builder, buttonId string, buttonIndex int) {
	builder.Find(buttonId).On("click", func(widget Widget, event string, data ...any) bool {
		cd.handleButtonClick(buttonIndex)
		return true
	})
}

// setupKeyboardHandlers sets up keyboard shortcuts for the dialog
func (cd *ConfirmDialog) setupKeyboardHandlers() {
	// Handle Enter key (activate default button)
	cd.container.On("key", func(widget Widget, event string, data ...any) bool {
		if keyEvent, ok := data[0].(*tcell.EventKey); ok {
			switch keyEvent.Key() {
			case tcell.KeyEnter:
				cd.handleButtonClick(cd.defaultBtn)
				return true
			case tcell.KeyEscape:
				cd.handleCancel()
				return true
			}
		}
		return false
	})
}

// handleButtonClick processes a button click
func (cd *ConfirmDialog) handleButtonClick(buttonIndex int) {
	if buttonIndex < 0 || buttonIndex >= len(cd.buttons) {
		return
	}
	
	result := ConfirmResult{
		Confirmed: buttonIndex == cd.defaultBtn,
		Button:    buttonIndex,
		Label:     cd.buttons[buttonIndex],
		Cancelled: false,
	}
	
	cd.closeDialog(result)
}

// handleCancel processes dialog cancellation (Escape key)
func (cd *ConfirmDialog) handleCancel() {
	result := ConfirmResult{
		Confirmed: false,
		Button:    cd.cancelBtn,
		Label:     cd.buttons[cd.cancelBtn],
		Cancelled: true,
	}
	
	cd.closeDialog(result)
}

// closeDialog closes the dialog and calls the callback
func (cd *ConfirmDialog) closeDialog(result ConfirmResult) {
	// Find UI and close the popup
	if ui := FindUI(cd.container); ui != nil {
		ui.Close()
	}
	
	// Call callback if set
	if cd.callback != nil {
		cd.callback(result)
	}
}

// getIconText returns the text representation of the dialog icon
func (cd *ConfirmDialog) getIconText() string {
	switch cd.icon {
	case IconQuestion:
		return " ? "
	case IconWarning:
		return " ! "
	case IconError:
		return " ✗ "
	case IconInfo:
		return " i "
	case IconSuccess:
		return " ✓ "
	default:
		return ""
	}
}

// ---- Convenience Functions ------------------------------------------------

// ConfirmDialog shows a simple confirmation dialog with OK/Cancel buttons.
// This is a convenience function for the most common use case.
//
// Parameters:
//   - ui: The UI instance to show the dialog in
//   - title: Dialog title
//   - message: Dialog message
//   - callback: Function to call with true/false result
//
// Example:
//   ConfirmDialog(ui, "Delete File", "Are you sure?", func(confirmed bool) {
//       if confirmed { deleteFile() }
//   })
func ConfirmDialog(ui *UI, title, message string, callback func(bool)) {
	NewConfirmDialog(title).
		SetMessage(message).
		SetIcon(IconQuestion).
		SetCallback(func(result ConfirmResult) {
			callback(result.Confirmed)
		}).
		Show(ui)
}

// WarningDialog shows a warning dialog with custom buttons.
// 
// Parameters:
//   - ui: The UI instance to show the dialog in
//   - title: Dialog title
//   - message: Warning message
//   - callback: Function to call with the result
//   - buttons: Button labels (first is default, last is cancel)
func WarningDialog(ui *UI, title, message string, callback ConfirmCallback, buttons ...string) {
	if len(buttons) == 0 {
		buttons = []string{"OK", "Cancel"}
	}
	
	NewConfirmDialog(title).
		SetMessage(message).
		SetIcon(IconWarning).
		SetButtons(buttons...).
		SetCallback(callback).
		Show(ui)
}

// ErrorDialog shows an error dialog with an OK button.
//
// Parameters:
//   - ui: The UI instance to show the dialog in
//   - title: Dialog title
//   - message: Error message
//   - callback: Optional function to call when dismissed
func ErrorDialog(ui *UI, title, message string, callback func()) {
	NewConfirmDialog(title).
		SetMessage(message).
		SetIcon(IconError).
		SetButtons("OK").
		SetCallback(func(result ConfirmResult) {
			if callback != nil {
				callback()
			}
		}).
		Show(ui)
}

// InfoDialog shows an informational dialog with an OK button.
//
// Parameters:
//   - ui: The UI instance to show the dialog in
//   - title: Dialog title
//   - message: Information message
//   - callback: Optional function to call when dismissed
func InfoDialog(ui *UI, title, message string, callback func()) {
	NewConfirmDialog(title).
		SetMessage(message).
		SetIcon(IconInfo).
		SetButtons("OK").
		SetCallback(func(result ConfirmResult) {
			if callback != nil {
				callback()
			}
		}).
		Show(ui)
}

// YesNoDialog shows a confirmation dialog with Yes/No buttons.
//
// Parameters:
//   - ui: The UI instance to show the dialog in
//   - title: Dialog title
//   - message: Question message
//   - callback: Function to call with true (Yes) or false (No)
func YesNoDialog(ui *UI, title, message string, callback func(bool)) {
	NewConfirmDialog(title).
		SetMessage(message).
		SetIcon(IconQuestion).
		SetButtons("Yes", "No").
		SetCallback(func(result ConfirmResult) {
			callback(result.Button == 0) // Yes = true, No = false
		}).
		Show(ui)
}
# TUI Designer Specification

## Goal

- Full-screen interactive canvas occupying the entire terminal
- VIM-style modal editing
- Fast character insertion from unicode palettes using the NumPad
- Per-cell styling via Style (foreground, background, attribute)
- Named and hierarchical styles that can be globally edited
- Multi-page documents with named pages and quick page navigation
- Load and Save as JSON

## UI Layout

The design area covers the whole screen, only the bottom line is used
for mode information, current style, drawing characters and position
(selection). Everything else is shown via pop-ups on the right side
of the screen on demand.

## Modes

### NORMAL

VIM Style cursor movement and mode changes. ESC always returns to normal
mode.

d DRAW mode
i INSERT mode
v VISUAL mode
: COMMAND mode
b Character page selection for NumPad (0-9, comma, plus, minus)
p pastes the yanked area
s Style selection and editing popup

### DRAW (d)

In draw mode the NumPad characters are assigned special characters.

i inserts icons from Nerd Fonts, shows a searchable popup list of available icons
s Shows or hides a display of character assignments for the NumPad

### INSERT (i)

In insert mode, all typed characters are inserted using the currently
selected style.

### VISUAL (v)

Visual mode allows the selection of a rectangular area.

b shows a popup to select the border style and then draws a box
d Deletes the current area (space character and default style)
f fills the area with the selected style
y yanks/copies the selected area into the clipboard

## Drawing Palettes

- Thin-round borders
- Thin borders
- Double borders
- Double-outer, thin inner borders
- Block elements
- Geometric shapes

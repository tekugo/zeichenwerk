# Appendix B — Cheatsheet

One-page reference. Print and stick to your monitor.

## Imports

```go
import (
    . "github.com/tekugo/zeichenwerk/v2"     // NewBuilder, Builder, UI
    "github.com/tekugo/zeichenwerk/v2/core"  // alignment & flag constants, MustFind, Find, Layout, …
    "github.com/tekugo/zeichenwerk/v2/themes"// TokyoNight, Nord, …
    "github.com/tekugo/zeichenwerk/v2/widgets" // OnActivate, OnChange, NewArrayTableProvider, …
    "github.com/tekugo/zeichenwerk/v2/values"  // Update[T], NewValue, Bind
)
```

## Themes

```go
themes.TokyoNight()
themes.Nord()
themes.MidnightNeon()
themes.GruvboxDark()
themes.GruvboxLight()
themes.Lipstick()
```

## Theme variables

`$bg0 $bg1 $bg2 $bg3   $fg0 $fg1 $fg2   $gray   $blue $cyan $aqua $magenta $red $orange $yellow $green`

## Hint(w, h) values

| Value | Meaning |
|-------|---------|
| `>0`  | Fixed cells |
| `<0`  | Fractional weight (magnitude = relative weight) |
| `0`   | Auto — ask the widget |

## Alignment (core)

`core.Default  core.Start  core.Left  core.Center  core.Right  core.End  core.Stretch`

## Flags (core)

`FlagFocusable  FlagFocused  FlagDisabled  FlagHidden  FlagSkip  FlagChecked  FlagPressed  FlagHovered  FlagReadonly  FlagMasked  FlagVertical  FlagHorizontal  FlagGrid  FlagSearch  FlagRight`

## Events (widgets)

`EvtActivate  EvtSelect  EvtChange  EvtAccept  EvtEnter  EvtClick  EvtKey  EvtMouse  EvtFocus  EvtBlur  EvtShow  EvtHide  EvtClose  EvtHover  EvtMode  EvtMove  EvtPaste`

## Typed event helpers

```go
widgets.OnActivate(w, func(idx int) bool       { … })
widgets.OnSelect  (w, func(idx int) bool       { … })
widgets.OnChange  (w, func(value string) bool  { … })
widgets.OnEnter   (w, func(value string) bool  { … })
widgets.OnAccept  (w, func(value string) bool  { … })
widgets.OnKey     (w, func(*tcell.EventKey) bool { … })
widgets.OnMouse   (w, func(*tcell.EventMouse) bool { … })
widgets.OnHide    (w, func() bool { … })
widgets.OnShow    (w, func() bool { … })
```

Raw form (any event, any data type):

```go
w.On(core.EvtChange, func(src core.Widget, ev core.Event, data ...any) bool {
    // assert data[0] yourself
    return true
})
```

## Lookup

```go
core.Find(container, "id") core.Widget          // untyped, may be nil
core.MustFind[*widgets.List](ui, "id") *List    // typed, panics on miss
core.FindAll[*widgets.Button](ui)               // all of a type
core.FindAt(ui, x, y)                           // deepest at (x,y)
```

## Push data into widgets

```go
values.Update(ui, "tables", []string{...})            // List, Tree, …
values.Update(ui, "result", widgets.NewArrayTableProvider(cols, rows))
values.Update(ui, "name", "value")                    // Input, Static, Editor
```

## Builder — containers

```go
HFlex(id, alignment, spacing)        VFlex(id, alignment, spacing)
Grid(id, rows, cols, lines)
    .Columns(c1, c2, …).Rows(r1, …)
    .Cell(x, y, w, h)
Box(id, title)         Dialog(id, title)
Switcher(id, connect)  Tabs(id, names...)   .Tab(name)
Form(id, title, data) .Group(id, title, name, horizontal, spacing)
Viewport(id, title)    Collapsible(id, title, expanded)
```

## Builder — common widgets

```go
Static(id, text)   Styled(id, text)   Text(id, content, follow, max)
Input(id, params...)   Editor(id)
Button(id, text)   Checkbox(id, text, checked)
List(id, items...)   Table(id, provider, cellNav)   Tree(id)
Select(id, args...)   Combo(id, items...)   Typeahead(id, params...)
Spinner(id, sequence)   Progress(id, horizontal)
HRule(style)   VRule(style)   Spacer()
```

## Builder — styling chain

```go
.Foreground("$cyan")                   .Foreground(":focus", "$bg0")
.Background("$bg2")                    .Background(":focus", "$blue")
.Font("bold")          // bold | italic | underline | strikethrough
.Border("round")                       .Border(":focus", "double")
.Padding(1)            .Padding(1, 2)  .Padding(1, 2, 3, 4)
.Margin(1)             // same shape as Padding
.Class("header")       // also empty .Class("") to reset
.Hint(w, h)            .Bounds(x, y, w, h)
.Cell(x, y, w, h)      // before each grid child
.Flag(core.FlagHidden, true)
```

## UI — runtime control

```go
ui.Run()                          // start the event loop
ui.Quit()                         // end it
ui.Refresh()                      // full screen redraw
widgets.Redraw(w)                 // single-widget redraw
widgets.Relayout(w)               // re-layout starting at w
ui.Focus(widget)                  // jump focus to a widget
ui.SetFocus("first" | "last" | "next" | "previous")
ui.SetTheme(theme)                // hot-swap a theme

ui.Popup(x, y, w, h, dialog)      // overlay layer (-1 = center)
ui.Close()                        // close topmost popup
ui.Confirm(title, msg, onYes, onNo)
ui.Prompt(title, msg, onAccept, onCancel)

ui.Debug()                        // turn on debug bar (Ctrl-D opens inspector)
ui.Log(self, core.Info, "msg", "k", v)
ui.SetLogLevel(slog.LevelDebug)
```

## Built-in keyboard shortcuts

| Key | Action |
|-----|--------|
| `Tab`, `↓`, `→` | Next focusable widget |
| `Shift-Tab`, `↑`, `←` | Previous focusable widget |
| `Esc` | Close topmost popup |
| `q`, `Q`, `Ctrl-Q`, `Ctrl-C` | Quit |
| `Ctrl-D` | Open inspector (when `.Debug()`) |

## Custom widget skeleton

```go
type MyWidget struct {
    *widgets.Component
    // your fields
}

func NewMyWidget(id, class string, /* args */) *MyWidget {
    return &MyWidget{Component: widgets.NewComponent(id, class)}
}

func (m *MyWidget) Apply(t *core.Theme) {
    t.Apply(m, m.Selector("mywidget"))
}

func (m *MyWidget) Hint() (int, int) {
    return /* w */, /* h */
}

func (m *MyWidget) Render(r *core.Renderer) {
    m.Component.Render(r)
    x, y, w, _ := m.Content()
    style := m.Style()
    r.Set(style.Foreground(), style.Background(), style.Font())
    r.Text(x, y, "hello", w)
}

// In main:  builder.Add(NewMyWidget(...))
```

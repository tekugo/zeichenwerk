package zeichenwerk

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// fcNodeData is the opaque data attached to every tree node in the FileChooser.
type fcNodeData struct {
	path  string
	isDir bool
}

// FileChooser shows a modal file/directory chooser dialog and returns the
// dialog widget. Attach event handlers to the returned widget before yielding
// control back to the event loop.
//
// title is shown in the dialog title bar.
//
// label is the confirm button text — e.g. "Open", "Save", or "Select".
//
// mode controls what can be selected:
//   - "dir"  — directories only (files are visible but not selectable)
//   - "file" — files only (directories are navigable but not selectable)
//   - "any"  — both files and directories are selectable
//
// initial is the starting path. If empty, os.Getwd() is used.
//
// showHidden controls whether dotfiles and dot-directories are initially
// visible. The user can toggle this at runtime via a checkbox.
//
// Events fired on the returned widget:
//   - EvtAccept (payload: string path) — user confirmed a valid selection
//   - EvtClose — dialog is closing for any reason (confirm or cancel)
func (ui *UI) FileChooser(title, label, mode, initial string, showHidden bool) Widget {
	if title == "" {
		title = "Choose"
	}
	if label == "" {
		label = "Open"
	}
	if initial == "" {
		initial, _ = os.Getwd()
	}
	initial = filepath.Clean(initial)

	// hidden is mutated by the checkbox; the loader closure captures its address.
	hidden := showHidden

	// suppress input→tree feedback when tree selection drives the input text.
	ignoreInputChange := false

	b := ui.NewBuilder()
	dialog := b.
		Dialog("fc-dialog", title).
		Class("dialog").
		Flex("fc-body", false, "stretch", 1).
			Typeahead("fc-input", initial).Hint(0, 1).
			Tree("fc-tree").Hint(0, -1).
			Flex("fc-footer", true, "center", 0).Hint(0, 1).
				Checkbox("fc-hidden", "show hidden", hidden).
				Spacer().Hint(-1, 0).
				Button("fc-ok", label).
				Button("fc-cancel", "Cancel").
			End().
		End().
		Class("").
		Container()

	input := Find(dialog, "fc-input").(*Typeahead)

	suggestPath := func(text string) []string {
		if text == "" {
			return nil
		}
		if text == "~" || strings.HasPrefix(text, "~/") {
			home, _ := os.UserHomeDir()
			if text == "~" {
				text = home
			} else {
				text = filepath.Join(home, text[2:])
			}
		}
		var dir, prefix string
		if strings.HasSuffix(text, "/") {
			dir = text
			prefix = ""
		} else {
			dir = filepath.Dir(text)
			prefix = filepath.Base(text)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}
		var candidates []string
		for _, e := range entries {
			if !hidden && strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if mode == "dir" && !e.IsDir() {
				continue
			}
			if prefix != "" && !strings.HasPrefix(strings.ToLower(e.Name()), strings.ToLower(prefix)) {
				continue
			}
			full := filepath.Join(dir, e.Name())
			if e.IsDir() {
				full += "/"
			}
			candidates = append(candidates, full)
		}
		sort.Strings(candidates)
		return candidates
	}
	input.SetSuggest(suggestPath)
	tree := Find(dialog, "fc-tree").(*Tree)
	okBtn := Find(dialog, "fc-ok").(*Button)
	cancelBtn := Find(dialog, "fc-cancel").(*Button)
	hiddenCb := Find(dialog, "fc-hidden").(*Checkbox)

	// Capture the normal input style (set by Apply) for error state toggling.
	normalInputStyle := input.Style()
	theme := ui.Theme()
	errorInputStyle := NewStyle("").WithColors("$red", theme.Color(normalInputStyle.Background()))

	// ---- helpers ------------------------------------------------------------

	isSelectable := func(nd fcNodeData) bool {
		switch mode {
		case "dir":
			return nd.isDir
		case "file":
			return !nd.isDir
		default: // "any"
			return true
		}
	}

	setInputError := func(bad bool) {
		if bad {
			input.SetStyle("", errorInputStyle)
		} else {
			input.SetStyle("", normalInputStyle)
		}
		Redraw(input)
	}

	updateOK := func() {
		node := tree.Selected()
		var ok bool
		if node != nil {
			if nd, valid := node.Data().(fcNodeData); valid {
				ok = isSelectable(nd)
			}
		}
		okBtn.SetFlag(FlagDisabled, !ok)
		Redraw(okBtn)
	}

	confirm := func() {
		path := filepath.Clean(input.Text())
		dialog.Dispatch(dialog, EvtAccept, path)
		ui.Close()
	}

	// ---- tree population ----------------------------------------------------

	// fcLoader returns a NodeLoader that populates dir children on first expand.
	var fcLoader func(dirPath string) NodeLoader
	fcLoader = func(dirPath string) NodeLoader {
		return func(node *TreeNode) {
			entries, _ := os.ReadDir(dirPath)

			var dirs, files []os.DirEntry
			for _, e := range entries {
				if !hidden && strings.HasPrefix(e.Name(), ".") {
					continue
				}
				if e.IsDir() {
					dirs = append(dirs, e)
				} else {
					files = append(files, e)
				}
			}

			sort.Slice(dirs, func(i, j int) bool {
				return strings.ToLower(dirs[i].Name()) < strings.ToLower(dirs[j].Name())
			})
			sort.Slice(files, func(i, j int) bool {
				return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
			})

			for _, e := range dirs {
				childPath := filepath.Join(dirPath, e.Name())
				child := NewLazyTreeNode(e.Name(), fcLoader(childPath), fcNodeData{childPath, true})
				node.Add(child)
			}
			if mode != "dir" {
				for _, e := range files {
					childPath := filepath.Join(dirPath, e.Name())
					child := NewTreeNode(e.Name(), fcNodeData{childPath, false})
					node.Add(child)
				}
			}
		}
	}

	// buildTree (re)populates the tree from / with the current hidden setting.
	buildTree := func() {
		root := NewTreeNode("", nil)
		rootNode := NewLazyTreeNode("/", fcLoader("/"), fcNodeData{"/", true})
		root.Add(rootNode)
		tree.SetRoot(root)
	}

	// navigateTo expands and selects the given absolute path in the tree.
	// Returns the final node that was selected (may be a partial match).
	navigateTo := func(path string) {
		path = filepath.Clean(path)
		parts := strings.Split(path, string(filepath.Separator))
		// parts[0] is "" (left of leading /); parts[1:] are path components.

		roots := tree.Root().Children()
		if len(roots) == 0 {
			return
		}
		node := roots[0] // the "/" node
		if !node.Expanded() {
			tree.Expand(node)
		}

		for _, part := range parts[1:] {
			if part == "" {
				continue
			}
			found := false
			for _, child := range node.Children() {
				if child.Text() == part {
					node = child
					if !node.Leaf() && !node.Expanded() {
						tree.Expand(node)
					}
					found = true
					break
				}
			}
			if !found {
				break
			}
		}

		tree.Select(node)
	}

	// ---- event wiring -------------------------------------------------------

	// Tree selection → update path input and OK button.
	tree.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		node, ok := data[0].(*TreeNode)
		if !ok {
			return false
		}
		nd, ok := node.Data().(fcNodeData)
		if !ok {
			return false
		}
		ignoreInputChange = true
		input.Set(nd.path)
		ignoreInputChange = false
		setInputError(false)
		updateOK()
		return true
	})

	// Path input → navigate tree.
	input.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if ignoreInputChange {
			return true
		}
		typed := input.Text()
		if typed == "" {
			setInputError(true)
			okBtn.SetFlag(FlagDisabled, true)
			Redraw(okBtn)
			return true
		}
		path := filepath.Clean(typed)
		info, err := os.Stat(path)
		if err != nil {
			setInputError(true)
			okBtn.SetFlag(FlagDisabled, true)
			Redraw(okBtn)
			return true
		}
		nd := fcNodeData{path: path, isDir: info.IsDir()}
		setInputError(!isSelectable(nd))
		navigateTo(path)
		updateOK()
		return true
	})

	// Intercept Enter on the tree: confirm if node is selectable, else
	// let the tree's own handler expand/collapse (new handlers run first).
	OnKey(tree, func(e *tcell.EventKey) bool {
		switch e.Key() {
		case tcell.KeyEnter:
			node := tree.Selected()
			if node == nil {
				return false
			}
			nd, ok := node.Data().(fcNodeData)
			if ok && isSelectable(nd) {
				confirm()
				return true
			}
			// not selectable — fall through to tree's default toggle
			return false

		case tcell.KeyRune:
			switch e.Str() {
			case "~":
				home, err := os.UserHomeDir()
				if err != nil {
					return true
				}
				navigateTo(home)
				return true
			case "/":
				input.Set("/")
				ui.Focus(input)
				return true
			}
		case tcell.KeyEscape:
			ui.Close()
			return true
		}
		return false
	})

	// ↑/↓ from input: move focus to tree.
	OnKey(input, func(e *tcell.EventKey) bool {
		switch e.Key() {
		case tcell.KeyUp, tcell.KeyDown:
			ui.Focus(tree)
			return true
		case tcell.KeyEscape:
			ui.Close()
			return true
		}
		return false
	})

	// OK button.
	okBtn.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		confirm()
		return true
	})

	// Cancel button.
	cancelBtn.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		ui.Close()
		return true
	})

	// show-hidden checkbox.
	hiddenCb.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		hidden = data[0].(bool)
		currentPath := filepath.Clean(input.Text())
		buildTree()
		navigateTo(currentPath)
		return true
	})

	// ---- initial state ------------------------------------------------------

	buildTree()
	navigateTo(initial)
	updateOK()

	ui.Popup(-1, -1, 60, 20, dialog)
	return dialog
}

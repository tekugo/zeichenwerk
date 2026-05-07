// Command zw scaffolds zeichenwerk projects.
//
// Two subcommands:
//
//	zw init [--flat] [--theme=Name] [--name=binary]
//	    Scaffold the current directory. Requires an existing go.mod;
//	    reads the module path from it. Writes zw.json plus the UI/event/
//	    main files according to layout. Never overwrites a hand-written
//	    file (only ui_gen.go is regenerated).
//
//	zw new <name> [--module=path] [--flat] [--theme=Name]
//	    Convenience wrapper: mkdir, go mod init, then zw init.
//
// Layouts:
//
//	default  cmd/<name>/main.go + internal/ui/{ui_gen,events}.go
//	--flat   main.go + events.go + ui_gen.go in module root
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "init":
		if err := runInit(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "zw init:", err)
			os.Exit(1)
		}
	case "new":
		if err := runNew(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "zw new:", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		usage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "zw: unknown subcommand %q\n", os.Args[1])
		usage(os.Stderr)
		os.Exit(2)
	}
}

func usage(w *os.File) {
	fmt.Fprintln(w, "zw — zeichenwerk project scaffolder")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  zw init [--flat] [--theme=Name] [--name=binary] [--replace=path]")
	fmt.Fprintln(w, "  zw new <name> [--module=path] [--flat] [--theme=Name] [--replace=path]")
}

// ---- init -----------------------------------------------------------------

type initOpts struct {
	flat    bool
	theme   string
	name    string // binary name for cmd/<name>; default = last segment of module path
	replace string // local path to inject as a `replace` directive for zeichenwerk
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	opts := initOpts{theme: "TokyoNight"}
	fs.BoolVar(&opts.flat, "flat", false, "flat layout (everything in module root, package main)")
	fs.StringVar(&opts.theme, "theme", opts.theme, "theme constructor name from the themes package")
	fs.StringVar(&opts.name, "name", "", "binary directory name under cmd/ (default: last segment of module path)")
	fs.StringVar(&opts.replace, "replace", "", "local path to use as a replace directive for github.com/tekugo/zeichenwerk")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !validTheme(opts.theme) {
		return fmt.Errorf("unknown theme %q (known: %s)", opts.theme, strings.Join(knownThemes, ", "))
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	module, err := readModulePath(filepath.Join(cwd, "go.mod"))
	if err != nil {
		return fmt.Errorf("reading go.mod: %w", err)
	}

	binaryName := opts.name
	if binaryName == "" {
		binaryName = lastSegment(module)
	}

	cfg := projectConfig{Module: module, Theme: opts.theme}
	if opts.flat {
		cfg.UIFile = "ui_gen.go"
		cfg.UIPackage = "main"
	} else {
		cfg.UIFile = "internal/ui/ui_gen.go"
		cfg.UIPackage = "ui"
	}

	if err := writeConfig(cwd, cfg); err != nil {
		return err
	}
	if err := writeScaffold(cwd, cfg, binaryName, opts.flat); err != nil {
		return err
	}
	if opts.replace != "" {
		abs, err := filepath.Abs(opts.replace)
		if err != nil {
			return fmt.Errorf("resolving --replace path: %w", err)
		}
		if err := runIn(cwd, "go", "mod", "edit",
			"-replace=github.com/tekugo/zeichenwerk="+abs); err != nil {
			return fmt.Errorf("go mod edit -replace: %w", err)
		}
	}
	if err := goGet(cwd); err != nil {
		return fmt.Errorf("go get zeichenwerk: %w", err)
	}
	if err := goModTidy(cwd); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	fmt.Println("zw: initialized in", cwd)
	if opts.flat {
		fmt.Println("    main.go, events.go, ui_gen.go (package main)")
	} else {
		fmt.Printf("    cmd/%s/main.go, internal/ui/{ui_gen,events}.go\n", binaryName)
	}
	return nil
}

// ---- new ------------------------------------------------------------------

func runNew(args []string) error {
	// flag.Parse stops at the first positional, so callers would have
	// to write `zw new --module=x mygame`; extracting the bare name
	// ourselves accepts either order.
	name, rest := extractPositional(args)
	if name == "" {
		return fmt.Errorf("expected exactly one positional argument <name>")
	}
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	var (
		flat    bool
		theme   = "TokyoNight"
		module  string
		replace string
	)
	fs.BoolVar(&flat, "flat", false, "flat layout")
	fs.StringVar(&theme, "theme", theme, "theme constructor name")
	fs.StringVar(&module, "module", "", "module path for go mod init (default: <name>)")
	fs.StringVar(&replace, "replace", "", "local zeichenwerk path for a replace directive")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected extra positional arguments: %v", fs.Args())
	}
	if module == "" {
		module = name
	}
	if !validTheme(theme) {
		return fmt.Errorf("unknown theme %q (known: %s)", theme, strings.Join(knownThemes, ", "))
	}

	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory %q already exists", name)
	}
	if err := os.MkdirAll(name, 0o755); err != nil {
		return err
	}
	abs, err := filepath.Abs(name)
	if err != nil {
		return err
	}
	if err := runIn(abs, "go", "mod", "init", module); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}
	if err := os.Chdir(abs); err != nil {
		return err
	}
	initArgs := []string{"--theme=" + theme, "--name=" + name}
	if flat {
		initArgs = append(initArgs, "--flat")
	}
	if replace != "" {
		initArgs = append(initArgs, "--replace="+replace)
	}
	return runInit(initArgs)
}

// ---- file writers ---------------------------------------------------------

type projectConfig struct {
	Module    string `json:"module"`
	UIFile    string `json:"ui_file"`
	UIPackage string `json:"ui_package"`
	Theme     string `json:"theme"`
}

func writeConfig(root string, cfg projectConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(root, "zw.json"), data, 0o644)
}

func writeScaffold(root string, cfg projectConfig, binaryName string, flat bool) error {
	if flat {
		return firstErr(
			writeIfAbsent(filepath.Join(root, "main.go"), flatMainSrc()),
			writeIfAbsent(filepath.Join(root, "events.go"), flatEventsSrc()),
			writeAlways(filepath.Join(root, "ui_gen.go"), flatUIGenSrc(cfg.Theme)),
		)
	}
	mainPath := filepath.Join(root, "cmd", binaryName, "main.go")
	uiDir := filepath.Join(root, filepath.Dir(cfg.UIFile))
	uiPath := filepath.Join(root, cfg.UIFile)
	eventsPath := filepath.Join(uiDir, "events.go")
	if err := os.MkdirAll(filepath.Dir(mainPath), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(uiDir, 0o755); err != nil {
		return err
	}
	uiImport := cfg.Module + "/" + filepath.ToSlash(filepath.Dir(cfg.UIFile))
	return firstErr(
		writeIfAbsent(mainPath, structuredMainSrc(uiImport, cfg.UIPackage)),
		writeIfAbsent(eventsPath, structuredEventsSrc(cfg.UIPackage)),
		writeAlways(uiPath, structuredUIGenSrc(cfg.UIPackage, cfg.Theme)),
	)
}

func writeIfAbsent(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func writeAlways(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

// ---- templates ------------------------------------------------------------

func flatMainSrc() string {
	return "package main\n\nfunc main() {\n\tui := BuildUI()\n\tWireEvents(ui)\n\tif err := ui.Run(); err != nil {\n\t\tpanic(err)\n\t}\n}\n"
}

func flatEventsSrc() string {
	return `package main

import (
	. "github.com/tekugo/zeichenwerk"
)

// WireEvents attaches event handlers to widgets created in ui_gen.go.
// Look up widgets by id with Find or MustFind from this dot-imported
// zeichenwerk package, then attach handlers via On. Bring core or
// widgets into scope (also via dot-import) when you need their symbols:
//
//	import . "github.com/tekugo/zeichenwerk/core"
//	import . "github.com/tekugo/zeichenwerk/widgets"
//
// Example:
//
//	if b, ok := Find(ui, "save").(*Button); ok {
//	    b.On(EvtClick, func(Widget, Event, ...any) bool { /* ... */ return false })
//	}
func WireEvents(ui *UI) {
	_ = ui
}
`
}

func flatUIGenSrc(theme string) string {
	return uiGenBody("main", theme)
}

func structuredMainSrc(uiImport, uiPkg string) string {
	return fmt.Sprintf(`package main

import (
	%q
)

func main() {
	app := %s.BuildUI()
	%s.WireEvents(app)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
`, uiImport, uiPkg, uiPkg)
}

func structuredEventsSrc(uiPkg string) string {
	return fmt.Sprintf(`package %s

import (
	. "github.com/tekugo/zeichenwerk"
)

// WireEvents attaches event handlers to widgets created in ui_gen.go.
// Look up widgets by id with Find or MustFind from this dot-imported
// zeichenwerk package, then attach handlers via On. Bring core or
// widgets into scope (also via dot-import) when you need their symbols:
//
//	import . "github.com/tekugo/zeichenwerk/core"
//	import . "github.com/tekugo/zeichenwerk/widgets"
//
// Example:
//
//	if b, ok := Find(ui, "save").(*Button); ok {
//	    b.On(EvtClick, func(Widget, Event, ...any) bool { /* ... */ return false })
//	}
func WireEvents(ui *UI) {
	_ = ui
}
`, uiPkg)
}

func structuredUIGenSrc(uiPkg, theme string) string {
	return uiGenBody(uiPkg, theme)
}

func uiGenBody(pkg, theme string) string {
	return fmt.Sprintf(`// Code generated by zw. DO NOT EDIT.

package %s

import (
	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
)

// BuildUI constructs the widget hierarchy. The designer regenerates this
// file from the live tree on every save; hand edits are lost on the next
// regenerate. Attach event handlers in events.go instead.
func BuildUI() *UI {
	return NewBuilder(themes.%s()).
		VFlex("root", Stretch, 0).
			Static("title", "Hello, zeichenwerk").
		End(). // VFlex#root
		Build()
}
`, pkg, theme)
}

// ---- helpers --------------------------------------------------------------

var knownThemes = []string{
	"GruvboxDark", "GruvboxLight", "Lipstick", "MidnightNeon", "Nord", "TokyoNight",
}

func validTheme(t string) bool {
	return slices.Contains(knownThemes, t)
}

func readModulePath(goMod string) (string, error) {
	data, err := os.ReadFile(goMod)
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("no module directive in %s", goMod)
}

func lastSegment(modulePath string) string {
	if i := strings.LastIndex(modulePath, "/"); i >= 0 {
		return modulePath[i+1:]
	}
	return modulePath
}

func goGet(dir string) error {
	return runIn(dir, "go", "get", "github.com/tekugo/zeichenwerk")
}

func goModTidy(dir string) error {
	return runIn(dir, "go", "mod", "tidy")
}

// extractPositional returns the first non-flag token as the positional
// argument and the remaining tokens (still including any flags) for
// flag.Parse. Returns "" if no positional was found.
func extractPositional(args []string) (string, []string) {
	for i, a := range args {
		if a == "--" {
			if i+1 < len(args) {
				return args[i+1], slices.Concat(args[:i], args[i+2:])
			}
			return "", args[:i]
		}
		if !strings.HasPrefix(a, "-") {
			return a, slices.Concat(args[:i], args[i+1:])
		}
	}
	return "", args
}

func runIn(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

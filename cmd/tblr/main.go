package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tekugo/zeichenwerk/cmd/tblr/format"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
	"golang.org/x/term"
)

func main() {
	// flags
	fromFmt := flag.String("from", "", "Input format")
	toFmt := flag.String("to", "", "Output format")
	fmtFlag := flag.String("format", "", "Pretty-print (same in/out format)")
	writeBack := flag.Bool("w", false, "Write result back to source file(s)")
	delimiter := flag.String("delimiter", "", "CSV/TSV field separator")
	noHeader := flag.Bool("no-header", false, "Treat all rows as data")
	pretty := flag.Bool("pretty", false, "Force pretty-print padding")
	ansiFlag := flag.Bool("ansi", false, "Render table as ANSI-styled output")
	ansiBorder := flag.String("ansi-border", "thin", "Border style: thin, double, rounded, thick, none")
	ansiTheme := flag.String("ansi-theme", "auto", "Colour theme: auto, dark, light, 16")
	noZebra := flag.Bool("no-ansi-zebra", false, "Disable zebra striping")
	width := flag.Int("width", 0, "Output width override")
	themeName := flag.String("t", "tokyo", "Theme: midnight, tokyo, nord, gruvbox-dark, gruvbox-light, lipstick")
	dir := flag.String("d", "", "Working directory")
	flag.Parse()

	args := flag.Args()

	// check if we're in headless mode
	headless := !isTerminal(os.Stdin) || *toFmt != "" || *fromFmt != "" || *fmtFlag != "" || *ansiFlag

	if headless {
		if err := runHeadless(args, *fromFmt, *toFmt, *fmtFlag, *writeBack, *delimiter, *noHeader, *pretty, *ansiFlag, *ansiBorder, *ansiTheme, !*noZebra, *width); err != nil {
			fmt.Fprintln(os.Stderr, "tblr:", err)
			os.Exit(1)
		}
		return
	}

	// TUI mode
	theme := resolveTheme2(*themeName)
	workDir := *dir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	tbl := format.NewMutableTable()
	var filePath string
	var activeFormat format.Format

	if len(args) > 0 {
		filePath = args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "tblr:", err)
			os.Exit(1)
		}
		activeFormat = detectByExtension(filePath)
		if activeFormat == nil {
			activeFormat = format.Detect(data)
		}
		if activeFormat == nil {
			activeFormat = format.ByName("csv")
		}
		t2, err := activeFormat.Parse(data, format.ParseOpts{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "tblr: parse error:", err)
			os.Exit(1)
		}
		tbl.Load(t2.Headers(), t2.Data())
		tbl.LoadAlignments(t2.Alignments())
		tbl.RecalcWidths()
	} else {
		activeFormat = format.ByName("csv")
		// empty table with one column
		tbl.Load([]string{"Column1"}, nil)
		tbl.RecalcWidths()
	}

	ui := buildUI(theme, tbl, workDir, filePath, activeFormat)
	ui.Run()
}

func runHeadless(args []string, fromName, toName, fmtName string, writeBack bool, delimStr string, noHeader, pretty, ansiMode bool, ansiBorder, ansiTheme string, zebra bool, width int) error {
	if ansiMode && writeBack {
		return fmt.Errorf("--ansi and -w cannot be combined")
	}
	if writeBack && len(args) == 0 {
		return fmt.Errorf("-w requires a file argument")
	}

	var delim rune
	if delimStr != "" {
		runes := []rune(delimStr)
		if len(runes) > 0 {
			delim = runes[0]
		}
	}

	parseOpts := format.ParseOpts{Delimiter: delim}

	// determine input format
	var inFmt format.Format
	if fromName != "" {
		inFmt = format.ByName(fromName)
		if inFmt == nil {
			return fmt.Errorf("unknown format: %s", fromName)
		}
	}

	// determine output format
	var outFmt format.Format
	if toName != "" {
		outFmt = format.ByName(toName)
		if outFmt == nil {
			return fmt.Errorf("unknown format: %s", toName)
		}
	}
	if fmtName != "" {
		f := format.ByName(fmtName)
		if f == nil {
			return fmt.Errorf("unknown format: %s", fmtName)
		}
		inFmt = f
		outFmt = f
	}

	serialOpts := format.SerialOpts{Pretty: pretty, Delimiter: delim}
	if outFmt != nil {
		n := outFmt.Name()
		if n == "markdown" || n == "asciidoc" || n == "typst" {
			serialOpts.Pretty = true
		}
	}
	if pretty {
		serialOpts.Pretty = true
	}

	// no file args — read from stdin
	if len(args) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		inf := inFmt
		if inf == nil {
			inf = format.Detect(data)
		}
		if inf == nil {
			return fmt.Errorf("cannot detect format")
		}
		tbl, err := inf.Parse(data, parseOpts)
		if err != nil {
			return err
		}
		if noHeader {
			tbl.SetHasHeader(false)
		}
		tbl.RecalcWidths()

		of := outFmt
		if of == nil {
			of = inf
		}

		if ansiMode {
			opts := DefaultANSIOpts()
			opts.Border = ansiBorder
			opts.Theme = ansiTheme
			opts.Zebra = zebra
			opts.Width = width
			if width == 0 {
				opts.Width = termWidth()
			}
			return RenderANSI(os.Stdout, tbl, opts)
		}

		out, err := of.Serialize(tbl, serialOpts)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(out)
		return err
	}

	// process files
	var firstErr error
	for _, path := range args {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "tblr:", err)
			firstErr = err
			continue
		}

		inf := inFmt
		if inf == nil {
			inf = detectByExtension(path)
		}
		if inf == nil {
			inf = format.Detect(data)
		}
		if inf == nil {
			inf = format.ByName("csv")
		}

		tbl, err := inf.Parse(data, parseOpts)
		if err != nil {
			fmt.Fprintln(os.Stderr, "tblr:", path+":", err)
			firstErr = err
			continue
		}
		if noHeader {
			tbl.SetHasHeader(false)
		}
		tbl.RecalcWidths()

		of := outFmt
		if of == nil {
			of = inf
		}

		if ansiMode {
			opts := DefaultANSIOpts()
			opts.Border = ansiBorder
			opts.Theme = ansiTheme
			opts.Zebra = zebra
			opts.Width = width
			if width == 0 {
				opts.Width = termWidth()
			}
			if err := RenderANSI(os.Stdout, tbl, opts); err != nil {
				firstErr = err
			}
			continue
		}

		out, err := of.Serialize(tbl, serialOpts)
		if err != nil {
			fmt.Fprintln(os.Stderr, "tblr:", path+":", err)
			firstErr = err
			continue
		}

		if writeBack {
			if err := os.WriteFile(path, out, 0644); err != nil {
				fmt.Fprintln(os.Stderr, "tblr:", path+":", err)
				firstErr = err
			}
		} else {
			os.Stdout.Write(out)
		}
	}
	return firstErr
}

// resolveTheme2 maps a theme name to a Theme.
func resolveTheme2(name string) *core.Theme {
	switch name {
	case "midnight":
		return themes.MidnightNeon()
	case "nord":
		return themes.Nord()
	case "gruvbox-dark":
		return themes.GruvboxDark()
	case "gruvbox-light":
		return themes.GruvboxLight()
	case "lipstick":
		return themes.Lipstick()
	default:
		return themes.TokyoNight()
	}
}

// detectByExtension returns a Format based on the file extension.
func detectByExtension(path string) format.Format {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	for _, f := range format.All() {
		for _, e := range f.Extensions() {
			if e == ext {
				return f
			}
		}
	}
	return nil
}

// isTerminal reports whether fd is a terminal.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// termWidth returns the terminal width or 80 as fallback.
func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// writeFile writes data to path (used by ui.go doSave).
func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

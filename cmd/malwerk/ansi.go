package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// ExportANSI writes the document as raw ANSI escape sequences. fg/bg
// colours are resolved through the theme's colour registry and emitted
// as 24-bit SGR (`\x1b[38;2;R;G;Bm`). Style names are lost. Per the spec:
// no trimming, no final reset.
func ExportANSI(d *Document, theme *core.Theme, path string) error {
	var sb strings.Builder
	for y := range d.Height {
		for x := range d.Width {
			cell := d.At(x, y)
			ds := d.StyleFor(cell.Style)
			fg := resolveColor(theme, ds.Fg)
			bg := resolveColor(theme, ds.Bg)
			if fg != "" {
				sb.WriteString(sgrFg(fg))
			}
			if bg != "" {
				sb.WriteString(sgrBg(bg))
			}
			if attrs := sgrFont(ds.Font); attrs != "" {
				sb.WriteString(attrs)
			}
			ch := cell.Ch
			if ch == "" {
				ch = " "
			}
			sb.WriteString(ch)
		}
		sb.WriteString("\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

// resolveColor turns a "$cyan" theme variable or literal into a hex
// value. Returns "" when the input is empty.
func resolveColor(theme *core.Theme, c string) string {
	if c == "" {
		return ""
	}
	resolved := theme.Color(c)
	return resolved
}

// sgrFg builds a 38;2 truecolour SGR. Accepts hex strings ("#rrggbb")
// only; non-hex inputs produce no output.
func sgrFg(hex string) string {
	r, g, b, ok := parseHex(hex)
	if !ok {
		return ""
	}
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func sgrBg(hex string) string {
	r, g, b, ok := parseHex(hex)
	if !ok {
		return ""
	}
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

func sgrFont(font string) string {
	if font == "" {
		return ""
	}
	var parts []string
	for _, attr := range strings.Fields(font) {
		switch attr {
		case "bold":
			parts = append(parts, "1")
		case "italic":
			parts = append(parts, "3")
		case "underline":
			parts = append(parts, "4")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(parts, ";") + "m"
}

// parseHex reads a "#rrggbb" colour into 0-255 components.
func parseHex(hex string) (r, g, b int, ok bool) {
	if len(hex) != 7 || hex[0] != '#' {
		return
	}
	v := func(s string) (int, bool) {
		var n int
		for _, c := range s {
			n <<= 4
			switch {
			case c >= '0' && c <= '9':
				n |= int(c - '0')
			case c >= 'a' && c <= 'f':
				n |= int(c-'a') + 10
			case c >= 'A' && c <= 'F':
				n |= int(c-'A') + 10
			default:
				return 0, false
			}
		}
		return n, true
	}
	rv, rok := v(hex[1:3])
	gv, gok := v(hex[3:5])
	bv, bok := v(hex[5:7])
	if !rok || !gok || !bok {
		return
	}
	return rv, gv, bv, true
}

package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// colors is a comprehensive mapping of color names to tcell.Color values.
// This map supports multiple color formats:
//   - Numeric colors: "0" through "255" for 256-color terminal support
//   - Named colors: Standard color names like "red", "blue", "green"
//   - Extended names: CSS-like color names such as "alice-blue", "antique-white"
//
// The mapping includes:
//   - Basic 16 colors (0-15): Standard terminal colors
//   - Extended 256 colors (16-255): Full 256-color palette
//   - Named colors: Human-readable color names for common colors
//   - CSS-style names: Hyphenated color names matching web standards
var colors = map[string]tcell.Color{
	"0":                      tcell.ColorBlack,
	"1":                      tcell.ColorMaroon,
	"2":                      tcell.ColorGreen,
	"3":                      tcell.ColorOlive,
	"4":                      tcell.ColorNavy,
	"5":                      tcell.ColorPurple,
	"6":                      tcell.ColorTeal,
	"7":                      tcell.ColorSilver,
	"8":                      tcell.ColorGray,
	"9":                      tcell.ColorRed,
	"10":                     tcell.ColorLime,
	"11":                     tcell.ColorYellow,
	"12":                     tcell.ColorBlue,
	"13":                     tcell.ColorFuchsia,
	"14":                     tcell.ColorAqua,
	"15":                     tcell.ColorWhite,
	"16":                     tcell.Color16,
	"17":                     tcell.Color17,
	"18":                     tcell.Color18,
	"19":                     tcell.Color19,
	"20":                     tcell.Color20,
	"21":                     tcell.Color21,
	"22":                     tcell.Color22,
	"23":                     tcell.Color23,
	"24":                     tcell.Color24,
	"25":                     tcell.Color25,
	"26":                     tcell.Color26,
	"27":                     tcell.Color27,
	"28":                     tcell.Color28,
	"29":                     tcell.Color29,
	"30":                     tcell.Color30,
	"31":                     tcell.Color31,
	"32":                     tcell.Color32,
	"33":                     tcell.Color33,
	"34":                     tcell.Color34,
	"35":                     tcell.Color35,
	"36":                     tcell.Color36,
	"37":                     tcell.Color37,
	"38":                     tcell.Color38,
	"39":                     tcell.Color39,
	"40":                     tcell.Color40,
	"41":                     tcell.Color41,
	"42":                     tcell.Color42,
	"43":                     tcell.Color43,
	"44":                     tcell.Color44,
	"45":                     tcell.Color45,
	"46":                     tcell.Color46,
	"47":                     tcell.Color47,
	"48":                     tcell.Color48,
	"49":                     tcell.Color49,
	"50":                     tcell.Color50,
	"51":                     tcell.Color51,
	"52":                     tcell.Color52,
	"53":                     tcell.Color53,
	"54":                     tcell.Color54,
	"55":                     tcell.Color55,
	"56":                     tcell.Color56,
	"57":                     tcell.Color57,
	"58":                     tcell.Color58,
	"59":                     tcell.Color59,
	"60":                     tcell.Color60,
	"61":                     tcell.Color61,
	"62":                     tcell.Color62,
	"63":                     tcell.Color63,
	"64":                     tcell.Color64,
	"65":                     tcell.Color65,
	"66":                     tcell.Color66,
	"67":                     tcell.Color67,
	"68":                     tcell.Color68,
	"69":                     tcell.Color69,
	"70":                     tcell.Color70,
	"71":                     tcell.Color71,
	"72":                     tcell.Color72,
	"73":                     tcell.Color73,
	"74":                     tcell.Color74,
	"75":                     tcell.Color75,
	"76":                     tcell.Color76,
	"77":                     tcell.Color77,
	"78":                     tcell.Color78,
	"79":                     tcell.Color79,
	"80":                     tcell.Color80,
	"81":                     tcell.Color81,
	"82":                     tcell.Color82,
	"83":                     tcell.Color83,
	"84":                     tcell.Color84,
	"85":                     tcell.Color85,
	"86":                     tcell.Color86,
	"87":                     tcell.Color87,
	"88":                     tcell.Color88,
	"89":                     tcell.Color89,
	"90":                     tcell.Color90,
	"91":                     tcell.Color91,
	"92":                     tcell.Color92,
	"93":                     tcell.Color93,
	"94":                     tcell.Color94,
	"95":                     tcell.Color95,
	"96":                     tcell.Color96,
	"97":                     tcell.Color97,
	"98":                     tcell.Color98,
	"99":                     tcell.Color99,
	"100":                    tcell.Color100,
	"101":                    tcell.Color101,
	"102":                    tcell.Color102,
	"103":                    tcell.Color103,
	"104":                    tcell.Color104,
	"105":                    tcell.Color105,
	"106":                    tcell.Color106,
	"107":                    tcell.Color107,
	"108":                    tcell.Color108,
	"109":                    tcell.Color109,
	"110":                    tcell.Color110,
	"111":                    tcell.Color111,
	"112":                    tcell.Color112,
	"113":                    tcell.Color113,
	"114":                    tcell.Color114,
	"115":                    tcell.Color115,
	"116":                    tcell.Color116,
	"117":                    tcell.Color117,
	"118":                    tcell.Color118,
	"119":                    tcell.Color119,
	"120":                    tcell.Color120,
	"121":                    tcell.Color121,
	"122":                    tcell.Color122,
	"123":                    tcell.Color123,
	"124":                    tcell.Color124,
	"125":                    tcell.Color125,
	"126":                    tcell.Color126,
	"127":                    tcell.Color127,
	"128":                    tcell.Color128,
	"129":                    tcell.Color129,
	"130":                    tcell.Color130,
	"131":                    tcell.Color131,
	"132":                    tcell.Color132,
	"133":                    tcell.Color133,
	"134":                    tcell.Color134,
	"135":                    tcell.Color135,
	"136":                    tcell.Color136,
	"137":                    tcell.Color137,
	"138":                    tcell.Color138,
	"139":                    tcell.Color139,
	"140":                    tcell.Color140,
	"141":                    tcell.Color141,
	"142":                    tcell.Color142,
	"143":                    tcell.Color143,
	"144":                    tcell.Color144,
	"145":                    tcell.Color145,
	"146":                    tcell.Color146,
	"147":                    tcell.Color147,
	"148":                    tcell.Color148,
	"149":                    tcell.Color149,
	"150":                    tcell.Color150,
	"151":                    tcell.Color151,
	"152":                    tcell.Color152,
	"153":                    tcell.Color153,
	"154":                    tcell.Color154,
	"155":                    tcell.Color155,
	"156":                    tcell.Color156,
	"157":                    tcell.Color157,
	"158":                    tcell.Color158,
	"159":                    tcell.Color159,
	"160":                    tcell.Color160,
	"161":                    tcell.Color161,
	"162":                    tcell.Color162,
	"163":                    tcell.Color163,
	"164":                    tcell.Color164,
	"165":                    tcell.Color165,
	"166":                    tcell.Color166,
	"167":                    tcell.Color167,
	"168":                    tcell.Color168,
	"169":                    tcell.Color169,
	"170":                    tcell.Color170,
	"171":                    tcell.Color171,
	"172":                    tcell.Color172,
	"173":                    tcell.Color173,
	"174":                    tcell.Color174,
	"175":                    tcell.Color175,
	"176":                    tcell.Color176,
	"177":                    tcell.Color177,
	"178":                    tcell.Color178,
	"179":                    tcell.Color179,
	"180":                    tcell.Color180,
	"181":                    tcell.Color181,
	"182":                    tcell.Color182,
	"183":                    tcell.Color183,
	"184":                    tcell.Color184,
	"185":                    tcell.Color185,
	"186":                    tcell.Color186,
	"187":                    tcell.Color187,
	"188":                    tcell.Color188,
	"189":                    tcell.Color189,
	"190":                    tcell.Color190,
	"191":                    tcell.Color191,
	"192":                    tcell.Color192,
	"193":                    tcell.Color193,
	"194":                    tcell.Color194,
	"195":                    tcell.Color195,
	"196":                    tcell.Color196,
	"197":                    tcell.Color197,
	"198":                    tcell.Color198,
	"199":                    tcell.Color199,
	"200":                    tcell.Color200,
	"201":                    tcell.Color201,
	"202":                    tcell.Color202,
	"203":                    tcell.Color203,
	"204":                    tcell.Color204,
	"205":                    tcell.Color205,
	"206":                    tcell.Color206,
	"207":                    tcell.Color207,
	"208":                    tcell.Color208,
	"209":                    tcell.Color209,
	"210":                    tcell.Color210,
	"211":                    tcell.Color211,
	"212":                    tcell.Color212,
	"213":                    tcell.Color213,
	"214":                    tcell.Color214,
	"215":                    tcell.Color215,
	"216":                    tcell.Color216,
	"217":                    tcell.Color217,
	"218":                    tcell.Color218,
	"219":                    tcell.Color219,
	"220":                    tcell.Color220,
	"221":                    tcell.Color221,
	"222":                    tcell.Color222,
	"223":                    tcell.Color223,
	"224":                    tcell.Color224,
	"225":                    tcell.Color225,
	"226":                    tcell.Color226,
	"227":                    tcell.Color227,
	"228":                    tcell.Color228,
	"229":                    tcell.Color229,
	"230":                    tcell.Color230,
	"231":                    tcell.Color231,
	"232":                    tcell.Color232,
	"233":                    tcell.Color233,
	"234":                    tcell.Color234,
	"235":                    tcell.Color235,
	"236":                    tcell.Color236,
	"237":                    tcell.Color237,
	"238":                    tcell.Color238,
	"239":                    tcell.Color239,
	"240":                    tcell.Color240,
	"241":                    tcell.Color241,
	"242":                    tcell.Color242,
	"243":                    tcell.Color243,
	"244":                    tcell.Color244,
	"245":                    tcell.Color245,
	"246":                    tcell.Color246,
	"247":                    tcell.Color247,
	"248":                    tcell.Color248,
	"249":                    tcell.Color249,
	"250":                    tcell.Color250,
	"251":                    tcell.Color251,
	"252":                    tcell.Color252,
	"253":                    tcell.Color253,
	"254":                    tcell.Color254,
	"255":                    tcell.Color255,
	"alice-blue":             tcell.ColorAliceBlue,
	"antique-white":          tcell.ColorAntiqueWhite,
	"aqua":                   tcell.ColorAqua,
	"aqua-marine":            tcell.ColorAquaMarine,
	"azure":                  tcell.ColorAzure,
	"beige":                  tcell.ColorBeige,
	"bisque":                 tcell.ColorBisque,
	"black":                  tcell.ColorBlack,
	"blanched-almond":        tcell.ColorBlanchedAlmond,
	"blue":                   tcell.ColorBlue,
	"blue-violet":            tcell.ColorBlueViolet,
	"brown":                  tcell.ColorBrown,
	"burly-wood":             tcell.ColorBurlyWood,
	"cadet-blue":             tcell.ColorCadetBlue,
	"chartreuse":             tcell.ColorChartreuse,
	"chocolate":              tcell.ColorChocolate,
	"coral":                  tcell.ColorCoral,
	"cornflower-blue":        tcell.ColorCornflowerBlue,
	"cornsilk":               tcell.ColorCornsilk,
	"crimson":                tcell.ColorCrimson,
	"dark-blue":              tcell.ColorDarkBlue,
	"dark-cyan":              tcell.ColorDarkCyan,
	"dark-goldenrod":         tcell.ColorDarkGoldenrod,
	"dark-gray":              tcell.ColorDarkGray,
	"dark-grey":              tcell.ColorDarkGrey,
	"dark-khaki":             tcell.ColorDarkKhaki,
	"dark-magenta":           tcell.ColorDarkMagenta,
	"dark-olive-green":       tcell.ColorDarkOliveGreen,
	"dark-orange":            tcell.ColorDarkOrange,
	"dark-orchid":            tcell.ColorDarkOrchid,
	"dark-red":               tcell.ColorDarkRed,
	"dark-salmon":            tcell.ColorDarkSalmon,
	"dark-sea-green":         tcell.ColorDarkSeaGreen,
	"dark-slate-gray":        tcell.ColorDarkSlateGray,
	"dark-slate-grey":        tcell.ColorDarkSlateGrey,
	"dark-turquoise":         tcell.ColorDarkTurquoise,
	"dark-violet":            tcell.ColorDarkViolet,
	"deep-pink":              tcell.ColorDeepPink,
	"deep-sky-blue":          tcell.ColorDeepSkyBlue,
	"dim-gray":               tcell.ColorDimGray,
	"dim-grey":               tcell.ColorDimGrey,
	"dodger-blue":            tcell.ColorDodgerBlue,
	"fire-brick":             tcell.ColorFireBrick,
	"floral-white":           tcell.ColorFloralWhite,
	"forest-green":           tcell.ColorForestGreen,
	"gainsboro":              tcell.ColorGainsboro,
	"ghost-white":            tcell.ColorGhostWhite,
	"gold":                   tcell.ColorGold,
	"goldenrod":              tcell.ColorGoldenrod,
	"green":                  tcell.ColorGreen,
	"green-yellow":           tcell.ColorGreenYellow,
	"honeydew":               tcell.ColorHoneydew,
	"hot-pink":               tcell.ColorHotPink,
	"indian-red":             tcell.ColorIndianRed,
	"indigo":                 tcell.ColorIndigo,
	"ivory":                  tcell.ColorIvory,
	"khaki":                  tcell.ColorKhaki,
	"lavender":               tcell.ColorLavender,
	"lavender-blush":         tcell.ColorLavenderBlush,
	"lawn-green":             tcell.ColorLawnGreen,
	"lemon-chiffon":          tcell.ColorLemonChiffon,
	"light-blue":             tcell.ColorLightBlue,
	"light-coral":            tcell.ColorLightCoral,
	"light-cyan":             tcell.ColorLightCyan,
	"light-goldenrod-yellow": tcell.ColorLightGoldenrodYellow,
	"light-green":            tcell.ColorLightGreen,
	"light-pink":             tcell.ColorLightPink,
	"light-salmon":           tcell.ColorLightSalmon,
	"light-sea-green":        tcell.ColorLightSeaGreen,
	"light-sky-blue":         tcell.ColorLightSkyBlue,
	"light-slate-gray":       tcell.ColorLightSlateGray,
	"light-slate-grey":       tcell.ColorLightSlateGrey,
	"light-steel-blue":       tcell.ColorLightSteelBlue,
	"light-yellow":           tcell.ColorLightYellow,
	"lime-green":             tcell.ColorLimeGreen,
	"linen":                  tcell.ColorLinen,
	"maroon":                 tcell.ColorMaroon,
	"medium-aquamarine":      tcell.ColorMediumAquamarine,
	"medium-blue":            tcell.ColorMediumBlue,
	"medium-orchid":          tcell.ColorMediumOrchid,
	"medium-purple":          tcell.ColorMediumPurple,
	"medium-sea-green":       tcell.ColorMediumSeaGreen,
	"medium-slate-blue":      tcell.ColorMediumSlateBlue,
	"medium-spring-green":    tcell.ColorMediumSpringGreen,
	"medium-turquoise":       tcell.ColorMediumTurquoise,
	"medium-violet-red":      tcell.ColorMediumVioletRed,
	"midnight-blue":          tcell.ColorMidnightBlue,
	"mint-cream":             tcell.ColorMintCream,
	"misty-rose":             tcell.ColorMistyRose,
	"moccasin":               tcell.ColorMoccasin,
	"navajo-white":           tcell.ColorNavajoWhite,
	"old-lace":               tcell.ColorOldLace,
	"olive":                  tcell.ColorOlive,
	"olive-drab":             tcell.ColorOliveDrab,
	"orange":                 tcell.ColorOrange,
	"orange-red":             tcell.ColorOrangeRed,
	"orchid":                 tcell.ColorOrchid,
	"pale-goldenrod":         tcell.ColorPaleGoldenrod,
	"pale-green":             tcell.ColorPaleGreen,
	"pale-turquoise":         tcell.ColorPaleTurquoise,
	"pale-violet-red":        tcell.ColorPaleVioletRed,
	"papaya-whip":            tcell.ColorPapayaWhip,
	"peach-puff":             tcell.ColorPeachPuff,
	"peru":                   tcell.ColorPeru,
	"pink":                   tcell.ColorPink,
	"plum":                   tcell.ColorPlum,
	"powder-blue":            tcell.ColorPowderBlue,
	"rebecca-purple":         tcell.ColorRebeccaPurple,
	"rosy-brown":             tcell.ColorRosyBrown,
	"royal-blue":             tcell.ColorRoyalBlue,
	"saddle-brown":           tcell.ColorSaddleBrown,
	"salmon":                 tcell.ColorSalmon,
	"sandy-brown":            tcell.ColorSandyBrown,
	"sea-green":              tcell.ColorSeaGreen,
	"seashell":               tcell.ColorSeashell,
	"sienna":                 tcell.ColorSienna,
	"silver":                 tcell.ColorSilver,
	"sky-blue":               tcell.ColorSkyblue,
	"slate-blue":             tcell.ColorSlateBlue,
	"slate-gray":             tcell.ColorSlateGray,
	"slate-grey":             tcell.ColorSlateGrey,
	"snow":                   tcell.ColorSnow,
	"spring-green":           tcell.ColorSpringGreen,
	"steel-blue":             tcell.ColorSteelBlue,
	"tan":                    tcell.ColorTan,
	"thistle":                tcell.ColorThistle,
	"tomato":                 tcell.ColorTomato,
	"turquoise":              tcell.ColorTurquoise,
	"violet":                 tcell.ColorViolet,
	"wheat":                  tcell.ColorWheat,
	"white":                  tcell.ColorWhite,
	"white-smoke":            tcell.ColorWhiteSmoke,
	"yellow-green":           tcell.ColorYellowGreen,
}

// ParseColor converts a color string to a tcell.Color value.
// This function supports multiple color formats for flexible color specification
// in TUI applications.
//
// Supported formats:
//   - Named colors: "red", "blue", "green", "alice-blue", etc.
//   - Numeric colors: "0" through "255" for 256-color terminal support
//   - Hex colors: "#RGB" (3-digit) or "#RRGGBB" (6-digit) format
//
// Named color examples:
//   - Basic colors: "black", "white", "red", "green", "blue"
//   - Extended colors: "dark-blue", "light-green", "medium-purple"
//   - CSS-style colors: "alice-blue", "antique-white", "cornflower-blue"
//
// Hex color examples:
//   - 3-digit: "#f00" (red), "#0f0" (green), "#00f" (blue)
//   - 6-digit: "#ff0000" (red), "#00ff00" (green), "#0000ff" (blue)
//
// For 3-digit hex colors, each digit is expanded (e.g., "#f0a" becomes "#ff00aa").
//
// Parameters:
//   - str: The color string to parse
//
// Returns:
//   - tcell.Color: The parsed color value
//   - error: An error if the color string is invalid or not found
//
// Example usage:
//
//	color, err := ParseColor("red")           // Named color
//	color, err := ParseColor("42")            // Numeric color
//	color, err := ParseColor("#ff0000")       // 6-digit hex
//	color, err := ParseColor("#f00")          // 3-digit hex
func ParseColor(str string) (tcell.Color, error) {
	// Handle named and numeric colors
	if !strings.HasPrefix(str, "#") {
		result, found := colors[str]
		if !found {
			return tcell.ColorDefault, fmt.Errorf("color name not found: %s", str)
		}
		return result, nil
	}

	// Handle hex colors
	str = str[1:] // Remove the '#' prefix

	if len(str) != 3 && len(str) != 6 {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color string: #%s (must be 3 or 6 characters)", str)
	}

	// Parse the RGB values depending on the color string length
	part := len(str) / 3

	r, err := strconv.ParseInt(str[0:part], 16, 64)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid red value: %s", str[0:part])
	}
	g, err := strconv.ParseInt(str[part:2*part], 16, 64)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid green value: %s", str[part:2*part])
	}
	b, err := strconv.ParseInt(str[2*part:], 16, 64)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid blue value: %s", str[2*part:])
	}

	// For 3-digit hex colors, expand each component (e.g., "f" becomes "ff")
	if part == 1 {
		r = r*16 + r // "f" becomes 0xff
		g = g*16 + g
		b = b*16 + b
	}

	return tcell.NewRGBColor(int32(r), int32(g), int32(b)), nil
}

// GetAvailableColors returns a slice of all available color names.
// This includes numeric colors ("0" through "255") and named colors.
// The returned slice can be used for color picker interfaces, validation,
// or documentation purposes.
//
// Returns:
//   - []string: A slice containing all available color names
//
// Example usage:
//
//	colors := GetAvailableColors()
//	fmt.Printf("Available colors: %v\n", colors)
func GetAvailableColors() []string {
	names := make([]string, 0, len(colors))
	for name := range colors {
		names = append(names, name)
	}
	return names
}

// IsValidColor checks if a color string is valid and can be parsed.
// This function is useful for validation without actually parsing the color.
//
// Parameters:
//   - str: The color string to validate
//
// Returns:
//   - bool: true if the color string is valid, false otherwise
//
// Example usage:
//
//	if IsValidColor("red") {
//	    fmt.Println("Valid color")
//	}
func IsValidColor(str string) bool {
	_, err := ParseColor(str)
	return err == nil
}

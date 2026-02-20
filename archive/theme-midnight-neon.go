package zeichenwerk

func MidnightNeonTheme() Theme {
	t := NewMapTheme()

	AddUnicodeBorders(t)

	t.SetColors(map[string]string{
		"$bg0":     "#0f1117",
		"$bg1":     "#1a1c23",
		"$bg2":     "#242730",
		"$bg3":     "#2f323d",
		"$fg0":     "#5ee9f0",
		"$fg1":     "#c7ccd9",
		"$fg2":     "#a0a4b3",
		"$fg3":     "#6c7384",
		"$blue":    "#5aaaff",
		"$cyan":    "#40e000",
		"$green":   "#4cd964",
		"$yellow":  "#ffd866",
		"$orange":  "#ff9f43",
		"$magenta": "#c792ea",
	})

	return t
}

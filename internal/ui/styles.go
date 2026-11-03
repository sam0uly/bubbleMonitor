package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"samouly.fun/bubbleMonitor/internal/config"
)

// ThemePalette holds colors for the application. Fields are color.Color so each
// theme can store a direct color (e.g. tty uses lipgloss.ANSIColor) or an
// adaptive color (compat.AdaptiveColor picks Light/Dark at render time).
type ThemePalette struct {
	Primary    color.Color
	Secondary  color.Color
	Success    color.Color
	Warning    color.Color
	Alert      color.Color
	Text       color.Color
	Muted      color.Color
	Border     color.Color
	Background color.Color
	Name       string
}

// makeColor returns an adaptive color with distinct values for Light/Dark terminal backgrounds.
func makeColor(light, dark string) color.Color {
	return compat.AdaptiveColor{
		Light: lipgloss.Color(light),
		Dark:  lipgloss.Color(dark),
	}
}

// makeTTYColor returns a direct ANSI indexed color. tty is a fixed palette
// that does not adapt to terminal background; the ANSI 0-15 colors are
// designed to be readable on both.
func makeTTYColor(ansi int) color.Color {
	return lipgloss.ANSIColor(ansi)
}

var themes = map[string]ThemePalette{
	"charmtone": {
		Name:       "charmtone",
		Primary:    makeColor("#5A56E0", "#7571F9"),
		Secondary:  makeColor("#EE6FF8", "#EE6FF8"),
		Success:    makeColor("#04B575", "#04B575"),
		Warning:    makeColor("#04B575", "#ECFD65"),
		Alert:      makeColor("#FF4672", "#ED567A"),
		Text:       makeColor("#1A1A1A", "#dddddd"),
		Muted:      makeColor("#A49FA5", "#777777"),
		Border:     makeColor("#B2B2B2", "#4A4A4A"),
		Background: makeColor("#FFFDF5", "#1A1A1A"),
	},
	"nord": {
		Name:       "nord",
		Primary:    makeColor("#5E81AC", "#88C0D0"),
		Secondary:  makeColor("#8B5E96", "#B48EAD"),
		Success:    makeColor("#6B8E5A", "#A3BE8C"),
		Warning:    makeColor("#B58E5A", "#EBCB8B"),
		Alert:      makeColor("#8B3A42", "#BF616A"),
		Text:       makeColor("#2E3440", "#ECEFF4"),
		Muted:      makeColor("#4C566A", "#D8DEE9"),
		Border:     makeColor("#D8DEE9", "#4C566A"),
		Background: makeColor("#ECEFF4", "#2E3440"),
	},
	"dracula": {
		Name:       "dracula",
		Primary:    makeColor("#7C3AED", "#BD93F9"),
		Secondary:  makeColor("#DB2777", "#FF79C6"),
		Success:    makeColor("#16A34A", "#50FA7B"),
		Warning:    makeColor("#D97706", "#FFB86C"),
		Alert:      makeColor("#DC2626", "#FF5555"),
		Text:       makeColor("#282A36", "#F8F8F2"),
		Muted:      makeColor("#6272A4", "#6272A4"),
		Border:     makeColor("#E1E4E8", "#44475A"),
		Background: makeColor("#F8F8F2", "#282A36"),
	},
	"gruvbox": {
		Name:       "gruvbox",
		Primary:    makeColor("#076678", "#83A598"),
		Secondary:  makeColor("#8F3F71", "#D3869B"),
		Success:    makeColor("#79740E", "#B8BB26"),
		Warning:    makeColor("#B57614", "#FABD2F"),
		Alert:      makeColor("#9D0006", "#FB4934"),
		Text:       makeColor("#282828", "#EBDBB2"),
		Muted:      makeColor("#7C6F64", "#A89984"),
		Border:     makeColor("#D5C4A1", "#504945"),
		Background: makeColor("#FBF1C7", "#282828"),
	},
	"rosepine": {
		Name:       "rosepine",
		Primary:    makeColor("#907AA9", "#C4A7E7"),
		Secondary:  makeColor("#D7827E", "#EBBCBA"),
		Success:    makeColor("#56949F", "#9CCFD8"),
		Warning:    makeColor("#EA9D34", "#F6C177"),
		Alert:      makeColor("#B4637A", "#EB6F92"),
		Text:       makeColor("#191724", "#E0DEF4"),
		Muted:      makeColor("#797593", "#908CAA"),
		Border:     makeColor("#DFD9E2", "#403D52"),
		Background: makeColor("#FAF4ED", "#191724"),
	},
	"everforest": {
		Name:       "everforest",
		Primary:    makeColor("#3DA89C", "#7FBBB3"),
		Secondary:  makeColor("#B16286", "#D699B6"),
		Success:    makeColor("#6C9E4E", "#A7C080"),
		Warning:    makeColor("#C1841D", "#DBBC7F"),
		Alert:      makeColor("#E66865", "#E67E80"),
		Text:       makeColor("#2D353B", "#D3C6AA"),
		Muted:      makeColor("#5C6A72", "#859289"),
		Border:     makeColor("#A7C5A0", "#3D484D"),
		Background: makeColor("#F3EFDA", "#2D353B"),
	},
	"nightowl": {
		Name:       "nightowl",
		Primary:    makeColor("#3F68C9", "#82AAFF"),
		Secondary:  makeColor("#8230AB", "#C792EA"),
		Success:    makeColor("#769D27", "#C5E478"),
		Warning:    makeColor("#D27E46", "#FFD590"),
		Alert:      makeColor("#C44343", "#EF5350"),
		Text:       makeColor("#011627", "#D6DEEB"),
		Muted:      makeColor("#5F7E97", "#5F7E97"),
		Border:     makeColor("#BCC8D6", "#2A3F5F"),
		Background: makeColor("#F0F4F8", "#011627"),
	},
	"palenight": {
		Name:       "palenight",
		Primary:    makeColor("#3F68C9", "#82AAFF"),
		Secondary:  makeColor("#8230AB", "#C792EA"),
		Success:    makeColor("#769D27", "#C3E88D"),
		Warning:    makeColor("#D27E46", "#FFCB6B"),
		Alert:      makeColor("#C44343", "#F07178"),
		Text:       makeColor("#292D3E", "#BFCBDB"),
		Muted:      makeColor("#676E95", "#676E95"),
		Border:     makeColor("#C5CAE0", "#3B4252"),
		Background: makeColor("#F5F6FA", "#292D3E"),
	},
	"material": {
		Name:       "material",
		Primary:    makeColor("#0288D1", "#89DDFF"),
		Secondary:  makeColor("#E64A19", "#F78C6C"),
		Success:    makeColor("#7CB342", "#C3E88D"),
		Warning:    makeColor("#FFA726", "#FFCB6B"),
		Alert:      makeColor("#E53935", "#FF5370"),
		Text:       makeColor("#263238", "#EEFFFF"),
		Muted:      makeColor("#607D8B", "#546E7A"),
		Border:     makeColor("#CFD8DC", "#2E3C43"),
		Background: makeColor("#FAFAFA", "#263238"),
	},
	"synthwave": {
		Name:       "synthwave",
		Primary:    makeColor("#B83280", "#FF7EDB"),
		Secondary:  makeColor("#1A8E8C", "#36F9F6"),
		Success:    makeColor("#2EA37A", "#72F1B8"),
		Warning:    makeColor("#C5A300", "#FED800"),
		Alert:      makeColor("#B82F3D", "#FE4450"),
		Text:       makeColor("#262335", "#FFFFFF"),
		Muted:      makeColor("#6B6B6B", "#B6B1B1"),
		Border:     makeColor("#6F6FA8", "#495495"),
		Background: makeColor("#F5F0FA", "#262335"),
	},
	"cobalt2": {
		Name:       "cobalt2",
		Primary:    makeColor("#005FB8", "#0088FF"),
		Secondary:  makeColor("#B8336A", "#FF628C"),
		Success:    makeColor("#2A9100", "#3AD900"),
		Warning:    makeColor("#B88A00", "#FFC600"),
		Alert:      makeColor("#C40000", "#FF0000"),
		Text:       makeColor("#193549", "#FFFFFF"),
		Muted:      makeColor("#6B7B5C", "#8F9D6A"),
		Border:     makeColor("#A8D0E0", "#0D3A58"),
		Background: makeColor("#FFFFFF", "#193549"),
	},
	"horizon": {
		Name:       "horizon",
		Primary:    makeColor("#1A6F7A", "#25B0BC"),
		Secondary:  makeColor("#B8334F", "#E95678"),
		Success:    makeColor("#1B9165", "#29D398"),
		Warning:    makeColor("#C77F5A", "#FAB795"),
		Alert:      makeColor("#B53D58", "#EC6A88"),
		Text:       makeColor("#1C1E26", "#FDF0ED"),
		Muted:      makeColor("#4F516B", "#6C6F93"),
		Border:     makeColor("#D5C0BC", "#2E303E"),
		Background: makeColor("#FDF0ED", "#1C1E26"),
	},
	"oceanic": {
		Name:       "oceanic",
		Primary:    makeColor("#4C6B95", "#6699CC"),
		Secondary:  makeColor("#8B6B8B", "#C594C5"),
		Success:    makeColor("#6B8E6B", "#99C794"),
		Warning:    makeColor("#B58E45", "#FAC863"),
		Alert:      makeColor("#B0454C", "#EC5f67"),
		Text:       makeColor("#1B2B34", "#CDD3DE"),
		Muted:      makeColor("#4A5258", "#65737E"),
		Border:     makeColor("#BFCAD0", "#343D46"),
		Background: makeColor("#ECF0F2", "#1B2B34"),
	},
	"palefire": {
		Name:       "palefire",
		Primary:    makeColor("#5C8A93", "#95C4CE"),
		Secondary:  makeColor("#A574B5", "#E1A3EE"),
		Success:    makeColor("#5A9876", "#8DD4AA"),
		Warning:    makeColor("#A18C72", "#E4C9AF"),
		Alert:      makeColor("#B25D66", "#EA868F"),
		Text:       makeColor("#242321", "#E2E0D7"),
		Muted:      makeColor("#5C5A4A", "#7D7A68"),
		Border:     makeColor("#C9C5BA", "#4A4845"),
		Background: makeColor("#E8E5DC", "#242321"),
	},
	"github": {
		Name:       "github",
		Primary:    makeColor("#0969DA", "#58A6FF"),
		Secondary:  makeColor("#8250DF", "#BC8CFF"),
		Success:    makeColor("#1A7F37", "#3FB950"),
		Warning:    makeColor("#9A6700", "#D29922"),
		Alert:      makeColor("#D1242F", "#F85149"),
		Text:       makeColor("#1F2328", "#E6EDF3"),
		Muted:      makeColor("#656D76", "#8B949E"),
		Border:     makeColor("#D1D9E0", "#30363D"),
		Background: makeColor("#FFFFFF", "#0D1117"),
	},
	"moonlight": {
		Name:       "moonlight",
		Primary:    makeColor("#3F68C9", "#82AAFF"),
		Secondary:  makeColor("#6F5DB5", "#BAACFF"),
		Success:    makeColor("#6B9A2E", "#C3E88D"),
		Warning:    makeColor("#C0882E", "#FFCB8B"),
		Alert:      makeColor("#C43D5A", "#FF5572"),
		Text:       makeColor("#222436", "#C8D3F5"),
		Muted:      makeColor("#4F5C8F", "#7A88CF"),
		Border:     makeColor("#9DAEDA", "#3E68D7"),
		Background: makeColor("#F5F6FA", "#222436"),
	},
	"shades": {
		Name:       "shades",
		Primary:    makeColor("#007CA5", "#00B4D8"),
		Secondary:  makeColor("#4FA0B0", "#90E0EF"),
		Success:    makeColor("#058F6F", "#06D6A0"),
		Warning:    makeColor("#C29A00", "#FFD60A"),
		Alert:      makeColor("#B8364C", "#EF476F"),
		Text:       makeColor("#212529", "#F8F9FA"),
		Muted:      makeColor("#6C757D", "#ADB5BD"),
		Border:     makeColor("#CED4DA", "#495057"),
		Background: makeColor("#F8F9FA", "#212529"),
	},
	"midnight": {
		Name:       "midnight",
		Primary:    makeColor("#1F6DA8", "#4D9DE0"),
		Secondary:  makeColor("#B04140", "#E15554"),
		Success:    makeColor("#4F8A2A", "#7BC043"),
		Warning:    makeColor("#B58E0A", "#F1C40F"),
		Alert:      makeColor("#B53A2C", "#E74C3C"),
		Text:       makeColor("#0B132B", "#ECF0F1"),
		Muted:      makeColor("#5C6A70", "#95A5A6"),
		Border:     makeColor("#BFCAD7", "#34495E"),
		Background: makeColor("#F0F4F8", "#0B132B"),
	},
	"forest": {
		Name:       "forest",
		Primary:    makeColor("#2D6A4F", "#52B788"),
		Secondary:  makeColor("#5BA882", "#95D5B2"),
		Success:    makeColor("#1B4332", "#40916C"),
		Warning:    makeColor("#B5783A", "#F4A261"),
		Alert:      makeColor("#B04A30", "#E76F51"),
		Text:       makeColor("#1B4332", "#E9F5DB"),
		Muted:      makeColor("#6B9080", "#B7E4C7"),
		Border:     makeColor("#95D5B2", "#2D6A4F"),
		Background: makeColor("#F0F7F1", "#1B4332"),
	},
	"autumn": {
		Name:       "autumn",
		Primary:    makeColor("#E07A5F", "#C2604A"),
		Secondary:  makeColor("#F2CC8F", "#D4B47A"),
		Success:    makeColor("#81B29A", "#5A8A75"),
		Warning:    makeColor("#F4A261", "#C58450"),
		Alert:      makeColor("#CA6702", "#9C4F02"),
		Text:       makeColor("#3D405B", "#3D405B"),
		Muted:      makeColor("#81B29A", "#5A8A75"),
		Border:     makeColor("#E07A5F", "#C2604A"),
		Background: makeColor("#F4F1DE", "#2A1F1A"),
	},
	"cyberpunk": {
		Name:       "cyberpunk",
		Primary:    makeColor("#0099A8", "#00F0FF"),
		Secondary:  makeColor("#B300B3", "#FF00FF"),
		Success:    makeColor("#00992B", "#00FF41"),
		Warning:    makeColor("#B3A000", "#FFE400"),
		Alert:      makeColor("#B30038", "#FF0055"),
		Text:       makeColor("#0A0E27", "#FFFFFF"),
		Muted:      makeColor("#707070", "#A0A0A0"),
		Border:     makeColor("#4D00B3", "#6600FF"),
		Background: makeColor("#F5F0F8", "#0A0E27"),
	},
	"sunset": {
		Name:       "sunset",
		Primary:    makeColor("#C44848", "#FF6B6B"),
		Secondary:  makeColor("#C5A937", "#FFE66D"),
		Success:    makeColor("#2A8A82", "#4ECDC4"),
		Warning:    makeColor("#B0821B", "#F7B731"),
		Alert:      makeColor("#B03848", "#EE5A6F"),
		Text:       makeColor("#2C3E50", "#F8F9FA"),
		Muted:      makeColor("#8B939B", "#CED4DA"),
		Border:     makeColor("#CED4DA", "#495057"),
		Background: makeColor("#F8F9FA", "#2C3E50"),
	},
	"ocean": {
		Name:       "ocean",
		Primary:    makeColor("#0077B6", "#00A6E2"),
		Secondary:  makeColor("#00B4D8", "#0099B8"),
		Success:    makeColor("#48CAE4", "#2A9CB0"),
		Warning:    makeColor("#CAF0F8", "#88BFD0"),
		Alert:      makeColor("#E63946", "#B82F3A"),
		Text:       makeColor("#023E8A", "#90E0EF"),
		Muted:      makeColor("#90E0EF", "#5C7A8B"),
		Border:     makeColor("#0096C7", "#00628A"),
		Background: makeColor("#F8F9FA", "#0A1F2E"),
	},
	"coffee": {
		Name:       "coffee",
		Primary:    makeColor("#6B4530", "#A97155"),
		Secondary:  makeColor("#8B6A4A", "#D4A373"),
		Success:    makeColor("#506A48", "#8FA383"),
		Warning:    makeColor("#A0825A", "#E2C799"),
		Alert:      makeColor("#7A3025", "#B85042"),
		Text:       makeColor("#3E2723", "#E8DCC4"),
		Muted:      makeColor("#6B5A5C", "#9D8D8F"),
		Border:     makeColor("#C9B0A0", "#6F5F5A"),
		Background: makeColor("#F5E8DC", "#3E2723"),
	},
	"solarized": {
		Name:       "solarized",
		Primary:    makeColor("#1A6FB0", "#268BD2"),
		Secondary:  makeColor("#4F539B", "#6C71C4"),
		Success:    makeColor("#5F7000", "#859900"),
		Warning:    makeColor("#806000", "#B58900"),
		Alert:      makeColor("#B02320", "#DC322F"),
		Text:       makeColor("#002B36", "#839496"),
		Muted:      makeColor("#4A555B", "#657B83"),
		Border:     makeColor("#C9D1D5", "#073642"),
		Background: makeColor("#FDF6E3", "#002B36"),
	},
	"monokai": {
		Name:       "monokai",
		Primary:    makeColor("#1A8FA8", "#66D9EF"),
		Secondary:  makeColor("#6A3FB0", "#AE81FF"),
		Success:    makeColor("#5C8F1A", "#A6E22E"),
		Warning:    makeColor("#B36A0A", "#FD971F"),
		Alert:      makeColor("#B01A52", "#F92672"),
		Text:       makeColor("#272822", "#F8F8F2"),
		Muted:      makeColor("#525046", "#75715E"),
		Border:     makeColor("#C9C5B5", "#49483E"),
		Background: makeColor("#F8F8F2", "#272822"),
	},
	"catppuccin": {
		Name:       "catppuccin",
		Primary:    makeColor("#1E66F5", "#89B4FA"),
		Secondary:  makeColor("#EA76CB", "#F5C2E7"),
		Success:    makeColor("#40A02B", "#A6E3A1"),
		Warning:    makeColor("#DF8E1D", "#F9E2AF"),
		Alert:      makeColor("#D20F39", "#F38BA8"),
		Text:       makeColor("#4C4F69", "#CDD6F4"),
		Muted:      makeColor("#6C6F85", "#9399B2"),
		Border:     makeColor("#BCC0CC", "#45475A"),
		Background: makeColor("#EFF1F5", "#1E1E2E"),
	},
	"tokyonight": {
		Name:       "tokyonight",
		Primary:    makeColor("#2E7DE9", "#7AA2F7"),
		Secondary:  makeColor("#9854F1", "#BB9AF7"),
		Success:    makeColor("#587539", "#9ECE6A"),
		Warning:    makeColor("#8C6C0E", "#E0AF68"),
		Alert:      makeColor("#C64343", "#F7768E"),
		Text:       makeColor("#3760BF", "#C0CAF5"),
		Muted:      makeColor("#848CB5", "#565F89"),
		Border:     makeColor("#C8D2E0", "#3B4261"),
		Background: makeColor("#E1E2E7", "#1A1B26"),
	},
	"onedark": {
		Name:       "onedark",
		Primary:    makeColor("#4078F2", "#61AFEF"),
		Secondary:  makeColor("#A626A4", "#C678DD"),
		Success:    makeColor("#50A14F", "#98C379"),
		Warning:    makeColor("#C18401", "#E5C07B"),
		Alert:      makeColor("#E45649", "#E06C75"),
		Text:       makeColor("#383A42", "#ABB2BF"),
		Muted:      makeColor("#A0A1A7", "#5C6370"),
		Border:     makeColor("#D3D3D3", "#3E4451"),
		Background: makeColor("#FAFAFA", "#282C34"),
	},
	"ayu": {
		Name:       "ayu",
		Primary:    makeColor("#FF6B00", "#FF8F40"),
		Secondary:  makeColor("#399EE6", "#5CCFEE"),
		Success:    makeColor("#86B300", "#B8CC52"),
		Warning:    makeColor("#F2AE49", "#F2AE49"),
		Alert:      makeColor("#F07171", "#FF3333"),
		Text:       makeColor("#5C6166", "#CBCCC6"),
		Muted:      makeColor("#ABB0B6", "#5C6773"),
		Border:     makeColor("#D9D9D9", "#3B4254"),
		Background: makeColor("#FAFAFA", "#1F2430"),
	},
	"tty": {
		Name:       "tty",
		Primary:    makeTTYColor(14),
		Secondary:  makeTTYColor(13),
		Success:    makeTTYColor(10),
		Warning:    makeTTYColor(11),
		Alert:      makeTTYColor(9),
		Text:       makeTTYColor(15),
		Muted:      makeTTYColor(8),
		Border:     makeTTYColor(7),
		Background: lipgloss.NoColor{},
	},
}

// GetTheme returns the color palette for the given theme name
func GetTheme(name string) ThemePalette {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["charmtone"]
}

// GetThemeFromCustom creates a ThemePalette from CustomThemeConfig.
// Custom themes use the same hex value for both Light and Dark terminal backgrounds.
func GetThemeFromCustom(custom *config.CustomThemeConfig) ThemePalette {
	if custom == nil {
		return GetTheme("charmtone")
	}
	return ThemePalette{
		Name:       "custom",
		Primary:    makeColor(custom.Primary, custom.Primary),
		Secondary:  makeColor(custom.Secondary, custom.Secondary),
		Success:    makeColor(custom.Success, custom.Success),
		Warning:    makeColor(custom.Warning, custom.Warning),
		Alert:      makeColor(custom.Alert, custom.Alert),
		Text:       makeColor(custom.Text, custom.Text),
		Muted:      makeColor(custom.Muted, custom.Muted),
		Border:     makeColor(custom.Border, custom.Border),
		Background: makeColor(custom.Background, custom.Background),
	}
}

// GetAppTheme helps resolve the theme from config
func GetAppTheme(theme string, custom *config.CustomThemeConfig) ThemePalette {
	if theme == "custom" {
		return GetThemeFromCustom(custom)
	}
	return GetTheme(theme)
}

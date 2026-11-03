package overlays

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
)

func RenderHelp(s *data.AppState, b, p, bg color.Color) string {
	sec := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#1E40AF"), Dark: lipgloss.Color("#3B82F6")}).Bold(true).MarginTop(1)
	key := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#047857"), Dark: lipgloss.Color("#10B981")}).Bold(true)
	desc := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#6B7280"), Dark: lipgloss.Color("#9CA3AF")})
	spacer := lipgloss.NewStyle()
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	useTwoColumns := s.UI.Width >= 100
	isCompact := s.UI.Width < 60

	var boxWidth int
	if useTwoColumns {
		boxWidth = int(float64(s.UI.Width) * 0.7)
		if boxWidth > 100 {
			boxWidth = 100
		}
	} else {
		boxWidth = int(float64(s.UI.Width) * 0.8)
		if boxWidth > 70 {
			boxWidth = 70
		}
	}
	if boxWidth < 40 {
		boxWidth = 40
	}
	if boxWidth > s.UI.Width-4 {
		boxWidth = s.UI.Width - 4
	}

	sp := spacer.Render

	var helpContent string

	if useTwoColumns {
		colWidth := (boxWidth - 12) / 2
		if colWidth < 25 {
			colWidth = 25
		}

		leftCol := lipgloss.JoinVertical(lipgloss.Left,
			sec.Width(colWidth).Render("NAVIGATION"),
			spacer.Width(colWidth).Render(key.Render("Tab / → / l")+sp("  ")+desc.Render("Next tab")),
			spacer.Width(colWidth).Render(key.Render("Shift+Tab / ← / h")+sp("  ")+desc.Render("Previous tab")),
			spacer.Width(colWidth).Render(key.Render("1-9, 0")+sp("  ")+desc.Render("Jump to tab")),
			spacer.Width(colWidth).Render(key.Render("PgUp / PgDn")+sp("  ")+desc.Render("Scroll page")),
			spacer.Width(colWidth).Render(key.Render("g")+sp("  ")+desc.Render("Go to top")),
			spacer.Width(colWidth).Render(key.Render("G / End")+sp("  ")+desc.Render("Go to bottom")),
			spacer.Width(colWidth).Render(""),
			sec.Width(colWidth).Render("GLOBAL"),
			spacer.Width(colWidth).Render(key.Render("?")+sp("  ")+desc.Render("Toggle help")),
			spacer.Width(colWidth).Render(key.Render(".")+sp("  ")+desc.Render("Toggle settings")),
			spacer.Width(colWidth).Render(key.Render("p")+sp("  ")+desc.Render("Pause/Resume")),
			spacer.Width(colWidth).Render(key.Render("r")+sp("  ")+desc.Render("Refresh metrics")),
			spacer.Width(colWidth).Render(key.Render("e")+sp("  ")+desc.Render("Export snapshot")),
			spacer.Width(colWidth).Render(key.Render("H")+sp("  ")+desc.Render("Cycle history")),
			spacer.Width(colWidth).Render(key.Render("C")+sp("  ")+desc.Render("Cycle chart")),
			spacer.Width(colWidth).Render(key.Render("n")+sp("  ")+desc.Render("CPU normalization")),
			spacer.Width(colWidth).Render(key.Render("q / Ctrl+C")+sp("  ")+desc.Render("Quit")),
			spacer.Width(colWidth).Render(key.Render("esc")+sp("  ")+desc.Render("Close overlay")),
		)

		rightCol := lipgloss.JoinVertical(lipgloss.Left,
			sec.Width(colWidth).Render("PROCESSES TAB"),
			spacer.Width(colWidth).Render(key.Render("j / ↓")+sp(" ")+desc.Render("Move down")),
			spacer.Width(colWidth).Render(key.Render("k / ↑")+sp(" ")+desc.Render("Move up")),
			spacer.Width(colWidth).Render(key.Render("f / /")+sp("   ")+desc.Render("Filter")),
			spacer.Width(colWidth).Render(key.Render("c")+sp("     ")+desc.Render("Clear filter")),
			spacer.Width(colWidth).Render(key.Render("s")+sp("     ")+desc.Render("Sort")),
			spacer.Width(colWidth).Render(key.Render("S")+sp("     ")+desc.Render("Toggle Sort Direction")),
			spacer.Width(colWidth).Render(key.Render("K")+sp("     ")+desc.Render("Kill")),
			spacer.Width(colWidth).Render(key.Render("z")+sp("     ")+desc.Render("Suspend")),
			spacer.Width(colWidth).Render(key.Render("x")+sp("     ")+desc.Render("Resume")),
			spacer.Width(colWidth).Render(key.Render("o")+sp("     ")+desc.Render("Open files")),
			spacer.Width(colWidth).Render(key.Render("b")+sp("     ")+desc.Render("Bookmark")),
			spacer.Width(colWidth).Render(key.Render("T")+sp("     ")+desc.Render("Tree view")),
			spacer.Width(colWidth).Render(key.Render("Space")+sp(" ")+desc.Render("Collapse tree")),
			spacer.Width(colWidth).Render(key.Render("+ / -")+sp(" ")+desc.Render("Priority +/-")),
			spacer.Width(colWidth).Render(""),
			lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#6B7280"), Dark: lipgloss.Color("#9CA3AF")}).Italic(true).Width(colWidth).Render("Press ? or ESC to close"),
		)

		helpContent = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(colWidth).Render(leftCol),
			lipgloss.NewStyle().Width(4).Render(""),
			lipgloss.NewStyle().Width(colWidth).Render(rightCol),
		)
	} else if isCompact {
		contentWidth := boxWidth - 8
		if contentWidth < 20 {
			contentWidth = 20
		}
		helpContent = lipgloss.JoinVertical(lipgloss.Left,
			sec.Width(contentWidth).Render("NAVIGATION"),
			spacer.Width(contentWidth).Render(key.Render("Tab/→/l")+" "+desc.Render("Next")),
			spacer.Width(contentWidth).Render(key.Render("S-Tab/←/h")+" "+desc.Render("Prev")),
			spacer.Width(contentWidth).Render(key.Render("1-9,0")+" "+desc.Render("Jump tab")),
			spacer.Width(contentWidth).Render(key.Render("PgUp/Dn")+" "+desc.Render("Scroll")),
			spacer.Width(contentWidth).Render(key.Render("g/G")+" "+desc.Render("Top/Bottom")),
			spacer.Width(contentWidth).Render(""),
			sec.Width(contentWidth).Render("GLOBAL"),
			spacer.Width(contentWidth).Render(key.Render("?")+" "+desc.Render("Help")),
			spacer.Width(contentWidth).Render(key.Render(".")+" "+desc.Render("Settings")),
			spacer.Width(contentWidth).Render(key.Render("p")+" "+desc.Render("Pause")),
			spacer.Width(contentWidth).Render(key.Render("r")+" "+desc.Render("Refresh")),
			spacer.Width(contentWidth).Render(key.Render("q")+" "+desc.Render("Quit")),
			spacer.Width(contentWidth).Render(""),
			lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#6B7280"), Dark: lipgloss.Color("#9CA3AF")}).Italic(true).Width(contentWidth).Render("? or ESC to close"),
		)
	} else {
		contentWidth := boxWidth - 8
		if contentWidth < 20 {
			contentWidth = 20
		}
		helpContent = lipgloss.JoinVertical(lipgloss.Left,
			sec.Width(contentWidth).Render("NAVIGATION"),
			spacer.Width(contentWidth).Render(key.Render("Tab / → / l")+sp("       ")+desc.Render("Next tab")),
			spacer.Width(contentWidth).Render(key.Render("Shift+Tab / ← / h")+sp("  ")+desc.Render("Previous tab")),
			spacer.Width(contentWidth).Render(key.Render("1-9, 0")+sp("           ")+desc.Render("Jump to specific tab")),
			spacer.Width(contentWidth).Render(key.Render("PgUp / PgDn")+sp("         ")+desc.Render("Scroll page")),
			spacer.Width(contentWidth).Render(key.Render("g")+sp("               ")+desc.Render("Go to top")),
			spacer.Width(contentWidth).Render(key.Render("G / End")+sp("          ")+desc.Render("Go to bottom")),
			spacer.Width(contentWidth).Render(""),
			sec.Width(contentWidth).Render("GLOBAL"),
			spacer.Width(contentWidth).Render(key.Render("?")+sp("   ")+desc.Render("Toggle help")),
			spacer.Width(contentWidth).Render(key.Render(".")+sp("   ")+desc.Render("Toggle settings")),
			spacer.Width(contentWidth).Render(key.Render("p")+sp("   ")+desc.Render("Pause/Resume monitoring")),
			spacer.Width(contentWidth).Render(key.Render("r")+sp("   ")+desc.Render("Refresh metrics")),
			spacer.Width(contentWidth).Render(key.Render("e")+sp("   ")+desc.Render("Export snapshot")),
			spacer.Width(contentWidth).Render(key.Render("H")+sp("   ")+desc.Render("Cycle history length")),
			spacer.Width(contentWidth).Render(key.Render("C")+sp("   ")+desc.Render("Cycle chart type")),
			spacer.Width(contentWidth).Render(key.Render("n")+sp("   ")+desc.Render("Toggle CPU normalization")),
			spacer.Width(contentWidth).Render(key.Render("q / Ctrl+C")+sp("  ")+desc.Render("Quit application")),
			spacer.Width(contentWidth).Render(key.Render("esc")+sp("  ")+desc.Render("Close overlays/menus")),
			spacer.Width(contentWidth).Render(""),
			sec.Width(contentWidth).Render("PROCESSES TAB"),
			spacer.Width(contentWidth).Render(key.Render("j / ↓")+sp("   ")+desc.Render("Move down")),
			spacer.Width(contentWidth).Render(key.Render("k / ↑")+sp("   ")+desc.Render("Move up")),
			spacer.Width(contentWidth).Render(key.Render("f / /")+sp("     ")+desc.Render("Filter processes")),
			spacer.Width(contentWidth).Render(key.Render("c")+sp("       ")+desc.Render("Clear filter")),
			spacer.Width(contentWidth).Render(key.Render("s")+sp("       ")+desc.Render("Sort processes")),
			spacer.Width(contentWidth).Render(key.Render("S")+sp("       ")+desc.Render("Toggle Sort Direction")),
			spacer.Width(contentWidth).Render(key.Render("K")+sp("       ")+desc.Render("Kill selected process")),
			spacer.Width(contentWidth).Render(key.Render("z")+sp("       ")+desc.Render("Suspend process")),
			spacer.Width(contentWidth).Render(key.Render("x")+sp("       ")+desc.Render("Resume process")),
			spacer.Width(contentWidth).Render(key.Render("o")+sp("       ")+desc.Render("Open files")),
			spacer.Width(contentWidth).Render(key.Render("b")+sp("       ")+desc.Render("Bookmark process")),
			spacer.Width(contentWidth).Render(key.Render("T")+sp("       ")+desc.Render("Toggle tree view")),
			spacer.Width(contentWidth).Render(key.Render("Space")+sp("   ")+desc.Render("Collapse/Expand tree")),
			spacer.Width(contentWidth).Render(key.Render("+ / -")+sp("   ")+desc.Render("Increase/Decrease priority")),
			spacer.Width(contentWidth).Render(""),
			lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#6B7280"), Dark: lipgloss.Color("#9CA3AF")}).Italic(true).Width(contentWidth).Render("Press ? or ESC to close"),
		)
	}

	var boxHeight int
	if useTwoColumns {
		boxHeight = 22
	} else if isCompact {
		boxHeight = 16
	} else {
		boxHeight = 42
	}
	maxHeight := int(float64(s.UI.Height) * 0.8)
	if boxHeight > maxHeight {
		boxHeight = maxHeight
	}
	if boxHeight < 10 {
		boxHeight = 10
	}

	container := lipgloss.NewStyle().
		Border(border).
		BorderForeground(b).
		Padding(1, 2).
		Width(boxWidth - 6).
		Height(boxHeight).
		BorderTop(false)

	body := container.Render(helpContent)
	actualWidth := lipgloss.Width(body)

	topBorder := widgets.RenderTopBorderWithBg("HELP", actualWidth, border, b, p)

	helpBox := lipgloss.JoinVertical(lipgloss.Left, topBorder, body)

	return helpBox
}

package overlays

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
)

func RenderSamLab(s *data.AppState, p, b, bg color.Color) string {
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	overlayWidth := data.SamLabDefaultWidth
	if overlayWidth > s.UI.Width-4 {
		overlayWidth = s.UI.Width - 4
	}

	boxStyle := lipgloss.NewStyle().
		Border(border).
		BorderForeground(b).
		Width(overlayWidth-6).
		BorderTop(false).
		Padding(1, 2)

	if s.Config.BackgroundOpaque {
		boxStyle = boxStyle.Background(bg)
	}

	centeredWidth := overlayWidth - 12

	titleStyle := lipgloss.NewStyle().
		Foreground(p).
		Bold(true).
		Width(centeredWidth).
		Align(lipgloss.Center)

	linkStyle := lipgloss.NewStyle().
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#2563EB"), Dark: lipgloss.Color("#60A5FA")}).
		Width(centeredWidth).
		Align(lipgloss.Center)

	infoStyle := lipgloss.NewStyle().
		Foreground(b).
		Width(centeredWidth).
		Align(lipgloss.Center)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("Made with love by Alaa El-Samouly"),
		linkStyle.Render("https://github.com/sam0uly/bubbleMonitor"),
		infoStyle.Render(""),
		infoStyle.Render("License: GNU AGPLv3"),
		infoStyle.Render("First Release: 2025-12-15"),
	)

	body := boxStyle.Render(content)
	actualWidth := lipgloss.Width(body)

	topBorder := widgets.RenderTopBorderWithBg("SAMLAB", actualWidth, border, b, p)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
}

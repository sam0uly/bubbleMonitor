package overlays

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/input"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
)

// RenderKillDialog renders the kill confirmation dialog using AppState
func RenderKillDialog(s *data.AppState, b, p, danger, t, mu, bg color.Color, zoneManager input.ZoneManager) string {
	boxWidth := min(data.KillDialogDefaultWidth, s.UI.Width-4)

	isHoverYes := zoneManager.IsHovered("kill-yes")
	isHoverNo := zoneManager.IsHovered("kill-no")

	warningStyle := lipgloss.NewStyle().Foreground(danger).Bold(true)

	yesSelected := s.UI.KillDialogSel == 0
	noSelected := s.UI.KillDialogSel == 1

	var yesBtnStyle, noBtnStyle lipgloss.Style
	if yesSelected || isHoverYes {
		yesBtnStyle = lipgloss.NewStyle().Bold(true).Background(t).Foreground(danger).Padding(0, 1).Underline(true)
		noBtnStyle = lipgloss.NewStyle().Bold(true).Foreground(bg).Background(mu).Padding(0, 1)
	} else if noSelected || isHoverNo {
		yesBtnStyle = lipgloss.NewStyle().Bold(true).Foreground(bg).Background(danger).Padding(0, 1)
		noBtnStyle = lipgloss.NewStyle().Bold(true).Background(t).Foreground(mu).Padding(0, 1).Underline(true)
	} else {
		yesBtnStyle = lipgloss.NewStyle().Bold(true).Foreground(bg).Background(danger).Padding(0, 1)
		noBtnStyle = lipgloss.NewStyle().Bold(true).Foreground(bg).Background(mu).Padding(0, 1)
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		warningStyle.Render(fmt.Sprintf("⚠ KILL PROCESS %s?", s.Process.KillTargetName)),
		"",
		lipgloss.JoinHorizontal(lipgloss.Center,
			yesBtnStyle.Render("[Y]es"),
			"   ",
			noBtnStyle.Render("[N]o"),
		),
	)

	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	container := lipgloss.NewStyle().
		Border(border).
		BorderForeground(danger).
		Padding(1, 3).
		Width(boxWidth - 6).
		Align(lipgloss.Center).
		BorderTop(false)

	body := container.Render(content)
	actualWidth := lipgloss.Width(body)
	topBorder := widgets.RenderTopBorderWithBg("CONFIRM KILL", actualWidth, border, danger, p)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
}

package overlays

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
)

// RenderOpenFilesOverlay renders the open files list
func RenderOpenFilesOverlay(s *data.AppState, width, height int, b, p, t, mu, bg color.Color) string {
	boxWidth := data.OpenFilesDefaultWidth
	if boxWidth > width-4 {
		boxWidth = width - 4
	}
	boxHeight := data.OpenFilesDefaultHeight
	if boxHeight > height-4 {
		boxHeight = height - 4
	}

	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)
	title := fmt.Sprintf("OPEN FILES (PID %d)", s.Process.OpenFilesPid)

	vpWidth := boxWidth - 10
	vpHeight := boxHeight - 4
	if vpWidth < 10 {
		vpWidth = 10
	}
	if vpHeight < 5 {
		vpHeight = 5
	}

	s.Process.OpenFilesView.SetWidth(vpWidth)
	s.Process.OpenFilesView.SetHeight(vpHeight)

	container := lipgloss.NewStyle().
		Border(border).
		BorderForeground(b).
		Padding(1, 2).
		Width(boxWidth - 6).
		Height(boxHeight).
		BorderTop(false)

	hint := lipgloss.NewStyle().Foreground(mu).Italic(true).Render("↑↓/jk to scroll • O or ESC to close")

	// Render Viewport
	renderedView := s.Process.OpenFilesView.View()

	contentWithHint := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(t).Render(renderedView),
		"",
		hint,
	)

	body := container.Render(contentWithHint)
	actualWidth := lipgloss.Width(body)

	topBorder := widgets.RenderTopBorderWithBg(title, actualWidth, border, b, p)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
}

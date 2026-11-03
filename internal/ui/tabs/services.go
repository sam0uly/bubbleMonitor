package tabs

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
	"samouly.fun/bubbleMonitor/internal/util"
)

func RenderServices(s *data.AppState, container lipgloss.Style, su, w, a, t, mu, p, b color.Color, availHeight int) string {
	boxWidth := s.UI.Width
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	contentWidth := boxWidth - 4

	nameWidth := contentWidth * 50 / 100
	statusWidth := contentWidth * 15 / 100
	descWidth := contentWidth - nameWidth - statusWidth - 2

	if nameWidth < 20 {
		nameWidth = 20
	}
	if statusWidth < 10 {
		statusWidth = 10
	}
	if descWidth < 20 {
		descWidth = 20
	}

	sb := &s.UI.ContentBuilder
	sb.Reset()
	sb.Grow(4096)

	hdrStyle := lipgloss.NewStyle().Bold(true).Foreground(p)

	nameH := util.PadRight("UNIT", nameWidth)
	statusH := util.PadRight("STATUS", statusWidth)
	descH := util.PadRight("DESCRIPTION", descWidth)

	headerRow := hdrStyle.Render(nameH) + " " +
		hdrStyle.Render(statusH) + " " +
		hdrStyle.Render(descH)

	sb.WriteString(headerRow)
	sb.WriteString("\n\n")

	visibleRows := availHeight - 4
	if visibleRows < 1 {
		visibleRows = 1
	}

	cellStyle := lipgloss.NewStyle().Foreground(t)
	runningStyle := lipgloss.NewStyle().Foreground(su)
	stoppedStyle := lipgloss.NewStyle().Foreground(mu)
	failedStyle := lipgloss.NewStyle().Foreground(a)

	startIdx := s.Process.ServicesScrollOffset
	if startIdx > len(s.Process.Services) {
		startIdx = 0
	}
	endIdx := startIdx + visibleRows
	if endIdx > len(s.Process.Services) {
		endIdx = len(s.Process.Services)
	}

	for i := startIdx; i < endIdx; i++ {
		svc := s.Process.Services[i]

		stStyle := cellStyle
		if svc.Status == "running" || svc.Status == "active" {
			stStyle = runningStyle
		} else if svc.Status == "failed" {
			stStyle = failedStyle
		} else {
			stStyle = stoppedStyle
		}

		name := svc.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		status := svc.Status
		if len(status) > statusWidth {
			status = status[:statusWidth-3] + "..."
		}

		desc := svc.Description
		if len(desc) > descWidth {
			desc = desc[:descWidth-3] + "..."
		}

		nameCell := cellStyle.Render(util.PadRight(name, nameWidth))
		statusCell := stStyle.Render(util.PadRight(status, statusWidth))
		descCell := cellStyle.Render(util.PadRight(desc, descWidth))

		sb.WriteString(nameCell)
		sb.WriteString(" ")
		sb.WriteString(statusCell)
		sb.WriteString(" ")
		sb.WriteString(descCell)

		if i < endIdx-1 {
			sb.WriteString("\n")
		}
	}

	content := sb.String()

	scrollInfo := ""
	if len(s.Process.Services) > visibleRows {
		endIdx := startIdx + visibleRows
		if endIdx > len(s.Process.Services) {
			endIdx = len(s.Process.Services)
		}
		scrollInfo = " [" + util.FastInt(startIdx+1) + "-" + util.FastInt(endIdx) + " of " + util.FastInt(len(s.Process.Services)) + "]"
	}

	topBorder := widgets.RenderTopBorderWithBg("SYSTEM SERVICES"+scrollInfo, boxWidth, border, b, p)

	c := container.Width(boxWidth).BorderTop(false)
	body := c.Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
}

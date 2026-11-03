package tabs

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
	"samouly.fun/bubbleMonitor/internal/util"
)

func RenderConnections(s *data.AppState, container lipgloss.Style, su, w, a, t, mu, p, b color.Color, availHeight int) string {
	boxWidth := s.UI.Width
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	contentWidth := boxWidth - 4

	protoWidth := 6
	stateWidth := 12
	pidWidth := 8

	remWidth := max(contentWidth-protoWidth-stateWidth-pidWidth-4, 30)
	localWidth := remWidth / 2
	remoteWidth := remWidth - localWidth

	hdrStyle := lipgloss.NewStyle().Bold(true).Foreground(p)

	s.UI.ContentBuilder.Reset()

	hRow := util.PadRight("PROTO", protoWidth) + " " +
		util.PadRight("LOCAL ADDR", localWidth) + " " +
		util.PadRight("REMOTE ADDR", remoteWidth) + " " +
		util.PadRight("STATE", stateWidth) + " " +
		util.PadRight("PID", pidWidth)

	headerRow := hdrStyle.Render(hRow)
	s.UI.ContentBuilder.WriteString(lipgloss.NewStyle().Width(contentWidth).Render(headerRow))
	s.UI.ContentBuilder.WriteString("\n\n")

	visibleRows := max(availHeight-4, 1)

	cellStyle := lipgloss.NewStyle().Foreground(t)

	truncate := func(str string, max int) string {
		if len(str) > max {
			return str[:max-1] + "…"
		}
		return str
	}

	padTruncate := func(str string, max int) string {
		return util.PadRight(truncate(str, max), max)
	}

	startIdx := s.Process.ConnectionsScrollOffset
	if startIdx > len(s.Process.Connections) {
		startIdx = 0
	}
	endIdx := min(startIdx+visibleRows, len(s.Process.Connections))

	for i := startIdx; i < endIdx; i++ {
		conn := s.Process.Connections[i]

		stateStyle := cellStyle
		switch conn.State {
		case "ESTABLISHED":
			stateStyle = lipgloss.NewStyle().Foreground(su)
		case "LISTEN":
			stateStyle = lipgloss.NewStyle().Foreground(p)
		case "TIME_WAIT", "CLOSE_WAIT":
			stateStyle = lipgloss.NewStyle().Foreground(w)
		case "CLOSED", "SYN_SENT", "SYN_RECV":
			stateStyle = lipgloss.NewStyle().Foreground(mu)
		}

		rowStr := padTruncate(conn.Protocol, protoWidth) + " " +
			padTruncate(conn.LocalAddr, localWidth) + " " +
			padTruncate(conn.RemoteAddr, remoteWidth) + " " +
			stateStyle.Render(padTruncate(conn.State, stateWidth)) + " " +
			util.PadRight(util.FastInt64(int64(conn.Pid)), pidWidth)
		s.UI.ContentBuilder.WriteString(rowStr)
		if i < endIdx-1 {
			s.UI.ContentBuilder.WriteString("\n")
		}
	}

	content := s.UI.ContentBuilder.String()

	c := container.Width(boxWidth).BorderTop(false)
	body := c.Render(content)

	scrollInfo := ""
	if len(s.Process.Connections) > visibleRows {
		endIdx := min(startIdx+visibleRows, len(s.Process.Connections))
		scrollInfo = " [" + util.FastInt(startIdx+1) + "-" + util.FastInt(endIdx) + " of " + util.FastInt(len(s.Process.Connections)) + "]"
	}

	topBorder := widgets.RenderTopBorderWithBg("NETWORK CONNECTIONS"+scrollInfo, boxWidth, border, b, p)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
}

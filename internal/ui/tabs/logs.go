package tabs

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
	"samouly.fun/bubbleMonitor/internal/util"
)

const (
	logError = "error"
	logFail  = "fail"
	logWarn  = "warn"
)

func RenderLogs(s *data.AppState, container lipgloss.Style, su, w, a, t, mu, p, b color.Color, availHeight int) string {
	boxWidth := s.UI.Width
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	contentWidth := boxWidth - 4

	visibleRows := max(availHeight-3, 1)

	logStyle := lipgloss.NewStyle().Foreground(t)

	count := len(s.Process.SystemLogs)
	startIdx := s.Process.LogsScrollOffset
	if startIdx > count {
		startIdx = 0
	}
	endIdx := min(startIdx+visibleRows, count)

	if s.Process.LogsScrollOffset == 0 && count > visibleRows {
		startIdx = count - visibleRows
		endIdx = count
	}

	errorStyle := lipgloss.NewStyle().Foreground(a)
	warnStyle := lipgloss.NewStyle().Foreground(w)
	defaultStyle := logStyle

	sb := &s.UI.ContentBuilder
	sb.Reset()
	sb.Grow(visibleRows * (contentWidth + 1))

	for i := startIdx; i < endIdx; i++ {
		line := s.Process.SystemLogs[i]
		if len(line) > contentWidth {
			line = line[:contentWidth]
		}
		formattedLine := util.PadRight(line, contentWidth)

		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, logError) || strings.Contains(lowerLine, logFail) {
			sb.WriteString(errorStyle.Render(formattedLine))
		} else if strings.Contains(lowerLine, logWarn) {
			sb.WriteString(warnStyle.Render(formattedLine))
		} else {
			sb.WriteString(defaultStyle.Render(formattedLine))
		}
		if i < endIdx-1 {
			sb.WriteString("\n")
		}
	}

	content := sb.String()

	c := container.Width(boxWidth).Height(visibleRows).BorderTop(false)
	body := c.Render(content)

	startLine := startIdx + 1
	if startLine > count {
		startLine = 1
	}
	title := "SYSTEM LOGS [" + util.FastInt(startLine) + "-" + util.FastInt(endIdx) + " of " + util.FastInt(count) + "]"
	topBorder := widgets.RenderTopBorderWithBg(title, boxWidth, border, b, p)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
}

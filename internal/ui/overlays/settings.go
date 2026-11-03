package overlays

import (
	"fmt"
	"image/color"
	"slices"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
)

// RenderSettingsOverlay renders the settings configuration modal
func RenderSettingsOverlay(s *data.AppState, width, height int, b, p, t, mu, bg color.Color) string {
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	itemStyle := lipgloss.NewStyle().Foreground(t)
	selectedStyle := lipgloss.NewStyle().Foreground(p).Bold(true).Border(border, false, false, false, true).BorderForeground(p).PaddingLeft(1)
	headerStyle := lipgloss.NewStyle().Foreground(p).Bold(true).MarginBottom(1)

	// Calculate dynamic label width from the longest label
	labelWidth := 0
	allLabels := []string{
		"CPU Alert:", "Mem Alert:", "Disk Alert:", "Temp Alert:",
		"Chart Type:", "View Type:", "Sort By:", "History Length:", "Process CPU:", "Sort Direction:",
		"Theme:", "Refresh Rate:", "Border Type:", "Border Style:", "Background:",
	}
	for _, l := range allLabels {
		if len(l) > labelWidth {
			labelWidth = len(l)
		}
	}
	labelFmt := fmt.Sprintf("%%-%ds %%s", labelWidth)

	var col1 []string
	col1 = append(col1, headerStyle.Render("THRESHOLDS & DISPLAY"))

	thresholdItems := []struct {
		label  string
		metric config.MetricType
		idx    int
	}{
		{"CPU Alert:", config.MetricCPU, 0},
		{"Mem Alert:", config.MetricMem, 1},
		{"Disk Alert:", config.MetricDisk, 2},
		{"Temp Alert:", config.MetricTemp, 3},
	}

	for _, item := range thresholdItems {
		val := fmt.Sprintf("%.0f%%", s.Config.Config.Thresholds[item.metric])
		if item.metric == config.MetricTemp {
			val = fmt.Sprintf("%.0f°C", s.Config.Config.Thresholds[item.metric])
		}
		line := fmt.Sprintf(labelFmt, item.label, val)
		if s.UI.SettingsIdx == item.idx {
			col1 = append(col1, selectedStyle.Render(line))
		} else {
			col1 = append(col1, itemStyle.Render("  "+line))
		}
	}

	viewName := "normal"
	if s.Process.TreeView {
		viewName = "tree"
	}

	displayStart := data.ThresholdCount

	displayItems := []struct {
		label string
		value string
		idx   int
	}{
		{"Chart Type:", s.UI.ChartType, displayStart},
		{"View Type:", viewName, displayStart + 1},
		{"Sort By:", s.Process.SortBy, displayStart + 2},
		{"History Length:", fmt.Sprintf("%ds", s.UI.HistoryLength), displayStart + 3},
		{"Process CPU:", "Raw", displayStart + 4},
		{"Sort Direction:", s.Process.SortDirection, displayStart + 5},
	}
	if s.Config.ProcessCpuNormalized {
		displayItems[4].value = "Normalized"
	}

	for _, item := range displayItems {
		line := fmt.Sprintf(labelFmt, item.label, item.value)
		if s.UI.SettingsIdx == item.idx {
			col1 = append(col1, selectedStyle.Render(line))
		} else {
			col1 = append(col1, itemStyle.Render("  "+line))
		}
	}

	var col2 []string
	col2 = append(col2, headerStyle.Render("TABS"))

	currentTabIdxBase := displayStart + data.DisplayCount

	for i, tabName := range data.AllAvailableTabs {
		idx := currentTabIdxBase + i
		isEnabled := slices.Contains(s.UI.ActiveTabs, tabName)

		status := "[ ]"
		if isEnabled {
			status = "[x]"
		}

		line := fmt.Sprintf("%-15s %s", tabName, status)
		if s.UI.SettingsIdx == idx {
			col2 = append(col2, selectedStyle.Render(line))
		} else {
			col2 = append(col2, itemStyle.Render("  "+line))
		}
	}

	var col3 []string
	col3 = append(col3, headerStyle.Render("APPEARANCE"))

	appearanceIdxBase := currentTabIdxBase + len(data.AllAvailableTabs)

	bgLabel := "transparent"
	if s.Config.BackgroundOpaque {
		bgLabel = "opaque"
	}

	appItems := []struct {
		label string
		value string
		idx   int
	}{
		{"Theme:", s.Config.Theme, appearanceIdxBase},
		{"Refresh Rate:", fmt.Sprintf("%dms", s.Config.RefreshRate), appearanceIdxBase + 1},
		{"Border Type:", s.Config.BorderType, appearanceIdxBase + 2},
		{"Border Style:", s.Config.BorderStyle, appearanceIdxBase + 3},
		{"Background:", bgLabel, appearanceIdxBase + 4},
	}

	for _, item := range appItems {
		line := fmt.Sprintf(labelFmt, item.label, item.value)
		if s.UI.SettingsIdx == item.idx {
			col3 = append(col3, selectedStyle.Render(line))
		} else {
			col3 = append(col3, itemStyle.Render("  "+line))
		}
	}

	// Calculate box width from actual content, not a fixed constant
	col1Content := lipgloss.JoinVertical(lipgloss.Left, col1...)
	col2Content := lipgloss.JoinVertical(lipgloss.Left, col2...)
	col3Content := lipgloss.JoinVertical(lipgloss.Left, col3...)

	// Use MarginRight for gap between columns instead of forced Width
	contentBlock := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().MarginRight(2).Render(col1Content),
		lipgloss.NewStyle().MarginRight(2).Render(col2Content),
		col3Content,
	)

	// Measure actual rendered width instead of calculating
	contentWidth := lipgloss.Width(contentBlock)
	// Clamp to screen
	if contentWidth > width-8 {
		contentWidth = width - 8
	}

	// Count content lines to push hint to bottom
	contentLines := lipgloss.Height(contentBlock)
	hintText := lipgloss.NewStyle().Foreground(mu).Align(lipgloss.Center).Width(contentWidth).Render("↑/↓ select • ←/→ toggle • [ ] reorder tabs • . to close")
	hintLines := lipgloss.Height(hintText)
	// boxHeight = content padding(2) + contentLines + gap + hintLines + bottom padding(1)
	boxHeight := 2 + contentLines + 1 + hintLines + 1
	boxHeight = min(boxHeight, height-4)

	container := lipgloss.NewStyle().
		Border(border).
		BorderForeground(b).
		Padding(1, 2).
		MaxWidth(width - 6).
		Height(boxHeight).
		BorderTop(false)

	// Place hint at the bottom: contentBlock fills top, padding pushes hint down
	innerHeight := boxHeight - 2 - 2 // subtract border + padding
	paddingBelow := innerHeight - contentLines - hintLines
	if paddingBelow < 0 {
		paddingBelow = 0
	}

	paddedContent := lipgloss.NewStyle().Height(contentLines).Render(contentBlock)
	spacer := lipgloss.NewStyle().Height(paddingBelow).Render("")
	body := container.Render(lipgloss.JoinVertical(lipgloss.Left, paddedContent, spacer, hintText))
	actualWidth := lipgloss.Width(body)

	topBorder := widgets.RenderTopBorderWithBg("CONFIGURATION", actualWidth, border, b, p)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
}

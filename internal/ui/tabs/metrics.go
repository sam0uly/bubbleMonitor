package tabs

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
	"samouly.fun/bubbleMonitor/internal/util"
)

// SumAccessor sums two accessors at each index
type SumAccessor struct {
	A, B data.Accessor
}

func (s *SumAccessor) Len() int {
	if s.A == nil {
		return 0
	}
	return s.A.Len()
}

func (s *SumAccessor) Get(i int) float64 {
	val := 0.0
	if s.A != nil {
		val += s.A.Get(i)
	}
	if s.B != nil {
		val += s.B.Get(i)
	}
	return val
}

func (s *SumAccessor) Max() float64 {
	length := s.Len()
	if length == 0 {
		return 0
	}
	max := 0.0
	for i := range length {
		v := s.Get(i)
		if v > max {
			max = v
		}
	}
	return max
}

// RenderMetrics renders the metrics/charts tab
func RenderMetrics(app *data.AppState, container lipgloss.Style, su, w, a, s, t, mu, p, b color.Color, availHeight int) string {
	width := app.UI.Width

	coreCols := 1
	if width >= 60 {
		coreCols = 2
	}
	if width >= 90 {
		coreCols = 3
	}
	if width >= 120 {
		coreCols = 4
	}
	if width >= 150 {
		coreCols = 5
	}

	numCores := len(app.Metrics.CpuPerCore)
	if numCores == 0 {
		numCores = 1
	}

	numCoreRows := (numCores + coreCols - 1) / coreCols
	coreSectionHeight := numCoreRows + 2

	chartCols := 1
	if width >= 100 {
		chartCols = 2
	}

	availChartSpace := availHeight - coreSectionHeight
	minChartSpace := 6
	if availChartSpace < minChartSpace {
		availChartSpace = minChartSpace
	}

	numChartRows := (4 + chartCols - 1) / chartCols
	chartBlockHeight := max(availChartSpace/numChartRows, 5)

	textStyle := lipgloss.NewStyle().Foreground(t)
	chartsTitles := []string{
		"CPU HISTORY (Window: " + util.FastInt(app.UI.HistoryLength) + "s)",
		"MEMORY HISTORY (Window: " + util.FastInt(app.UI.HistoryLength) + "s)",
		"NETWORK ACTIVITY (Window: " + util.FastInt(app.UI.HistoryLength) + "s)",
		"DISK I/O HISTORY (Window: " + util.FastInt(app.UI.HistoryLength) + "s)",
	}
	var renderedCharts []string

	chartWidths := util.CalculateColumnWidths(width, chartCols)

	renderChart := func(data data.Accessor, chartW, chartH int, c1, c2 color.Color, fixedMax float64) string {
		switch app.UI.ChartType {
		case "line":
			return widgets.RenderLineChart(data, chartW, chartH, c1, c2, fixedMax, app.Config.Theme)
		case "bar":
			return widgets.RenderBarChart(data, chartW, chartH, c1, c2, fixedMax, app.Config.Theme)
		default:
			// Default to Braille (highest resolution)
			return widgets.RenderBrailleChart(data, chartW, chartH, c1, c2, fixedMax, app.Config.Theme)
		}
	}

	for i := range 4 {
		var boxW int
		if chartCols == 1 {
			boxW = width
		} else {
			boxW = chartWidths[i%chartCols]
		}

		sparklineH := max(chartBlockHeight-3, 1)

		contentW := boxW - 4
		chartW := max(contentW, 5)

		var innerBlock string

		switch i {
		case 0: // CPU
			ch := renderChart(app.Metrics.CpuHistory, chartW, sparklineH, p, w, 100.0)
			stats := "Cur: " + util.FastPercent1(app.Metrics.Cpu) + " Avg: " + util.FastPercent1(app.Metrics.CpuHistory.Avg()) + " Peak: " + util.FastPercent1(app.Metrics.CpuHistory.Max())
			innerBlock = lipgloss.JoinVertical(lipgloss.Left, ch, textStyle.Render(stats))
		case 1: // Mem
			ch := renderChart(app.Metrics.MemHistory, chartW, sparklineH, s, w, 100.0)
			stats := "Cur: " + util.FastPercent1(app.Metrics.Memory) + " Avg: " + util.FastPercent1(app.Metrics.MemHistory.Avg()) + " Peak: " + util.FastPercent1(app.Metrics.MemHistory.Max())
			innerBlock = lipgloss.JoinVertical(lipgloss.Left, ch, textStyle.Render(stats))
		case 2: // Net
			ch := renderChart(app.Metrics.NetHistory, chartW, sparklineH, su, w, 100.0)
			stats := "Peak: " + util.FastPercent1(app.Metrics.NetHistory.Max()) + " Recv: " + util.FastMbPerSec(app.Metrics.NetRecvRate) + " Sent: " + util.FastMbPerSec(app.Metrics.NetSentRate)
			innerBlock = lipgloss.JoinVertical(lipgloss.Left, ch, textStyle.Render(stats))
		case 3: // Disk I/O
			totalIO := &SumAccessor{A: app.Metrics.DiskHORead, B: app.Metrics.DiskHOWrite}
			ch := renderChart(totalIO, chartW, sparklineH, mu, w, 0.0)
			stats := "Read: " + util.FastMbPerSec(app.Metrics.DiskReadRate) + " Write: " + util.FastMbPerSec(app.Metrics.DiskWriteRate)
			innerBlock = lipgloss.JoinVertical(lipgloss.Left, ch, textStyle.Render(stats))
		}

		c := container.Width(boxW).Height(chartBlockHeight - 1).BorderTop(false)
		body := c.Render(innerBlock)

		topBorder := widgets.RenderTopBorderWithBg(chartsTitles[i], boxW, widgets.GetBorder(app.Config.BorderStyle, app.Config.BorderType), b, p)
		renderedCharts = append(renderedCharts, lipgloss.JoinVertical(lipgloss.Left, topBorder, body))
	}

	var rows []string
	for i := 0; i < len(renderedCharts); i += chartCols {
		end := min(i+chartCols, len(renderedCharts))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, renderedCharts[i:end]...))
	}
	topSection := lipgloss.JoinVertical(lipgloss.Left, rows...)

	coreColWidths := util.CalculateColumnWidths(width, coreCols)
	var coreBlocks []string
	textStyle = lipgloss.NewStyle().Foreground(t)

	for i, usage := range app.Metrics.CpuPerCore {
		cW := max(coreColWidths[i%coreCols]-4, 10)
		barW := max(cW-16, 5)

		bar := widgets.RenderProgressBar(usage, barW, su, w, a)
		coreLabel := "Core " + util.FastInt(i) + ": " + util.FastPercent1(usage) + " "
		line := lipgloss.JoinHorizontal(lipgloss.Left,
			textStyle.Width(16).Render(coreLabel),
			bar,
		)
		coreBlocks = append(coreBlocks, line)
	}

	var coreRows []string
	for i := 0; i < len(coreBlocks); i += coreCols {
		end := min(i+coreCols, len(coreBlocks))
		rowItems := coreBlocks[i:end]

		var rowStr string
		for j, item := range rowItems {
			w := coreColWidths[(i+j)%coreCols]
			rowStr = lipgloss.JoinHorizontal(lipgloss.Top, rowStr, lipgloss.NewStyle().Width(w).Render(item))
		}
		coreRows = append(coreRows, rowStr)
	}

	const maxCoreRows = 4
	totalCoreRows := len(coreRows)
	visibleCoreRows := coreRows

	if totalCoreRows > maxCoreRows {
		start := max(app.UI.CpuCoreScrollOffset, 0)
		end := start + maxCoreRows
		if end > totalCoreRows {
			end = totalCoreRows
			start = max(end-maxCoreRows, 0)
		}
		visibleCoreRows = coreRows[start:end]
	}

	coresC := strings.Join(visibleCoreRows, "\n")

	title := "CPU PER CORE"
	if totalCoreRows > maxCoreRows {
		title += " (PgUp/PgDn " + util.FastInt(app.UI.CpuCoreScrollOffset+1) + "/" + util.FastInt(totalCoreRows-maxCoreRows+1) + ")"
	}

	coresBoxWidth := width
	topBorder := widgets.RenderTopBorderWithBg(title, coresBoxWidth, widgets.GetBorder(app.Config.BorderStyle, app.Config.BorderType), b, p)

	c := container.Width(coresBoxWidth).Height(len(visibleCoreRows)).BorderTop(false)
	body := c.Render(coresC)

	bottomSection := lipgloss.JoinVertical(lipgloss.Left, topBorder, body)

	return lipgloss.JoinVertical(lipgloss.Left, topSection, bottomSection)
}

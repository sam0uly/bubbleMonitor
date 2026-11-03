package tabs

import (
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
	"samouly.fun/bubbleMonitor/internal/util"
)

// RenderOverview renders the overview tab
func RenderOverview(s *data.AppState, container, titleStyle, labelStyle, valueStyle lipgloss.Style, su, w, a, t, mu, bg, p, b color.Color, availHeight int) string {
	width := s.UI.Width
	cols := 1
	if width >= 80 {
		cols = 2
	}
	if width >= 120 {
		cols = 3
	}

	colWidths := util.CalculateColumnWidths(width, cols)
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	getContentWidth := func(i int) int {
		return colWidths[i%cols] - 4
	}

	sp := func(str string) string { return str }

	// CPU (Index 0)
	idx := 0
	cw := getContentWidth(idx)
	lAS := "N/A"
	if s.Metrics.LoadAvg != nil {
		lAS = util.FastFloat2(s.Metrics.LoadAvg.Load1) + ", " + util.FastFloat2(s.Metrics.LoadAvg.Load5) + ", " + util.FastFloat2(s.Metrics.LoadAvg.Load15)
	}
	cpuBar := widgets.RenderProgressBar(s.Metrics.Cpu, cw, su, w, a)
	// Pad cpuVal manually
	cpuStr := util.FastPercent1(s.Metrics.Cpu)
	cpuVal := valueStyle.Foreground(widgets.GetColorForValue(s.Metrics.Cpu, su, w, a)).Render(util.PadRight(cpuStr, cw))

	var cpuTemp string
	if s.Metrics.CpuTemp > 0 {
		cpuTemp = util.FastTemp(s.Metrics.CpuTemp)
	} else {
		cpuTemp = "N/A (Admin?)"
	}

	// Manual padding for composite lines
	loadLine := labelStyle.Render("Load:") + sp(" ") + labelStyle.Render(lAS)
	loadLine = util.PadRight(loadLine, cw)

	tempLine := labelStyle.Render("Temp:") + sp(" ") + labelStyle.Render(cpuTemp)
	tempLine = util.PadRight(tempLine, cw)

	cpuBlock := lipgloss.JoinVertical(lipgloss.Left,
		cpuVal,
		cpuBar,
		loadLine,
		tempLine,
	)

	// Memory (Index 1) - Use cached MemInfo
	idx = 1
	cw = getContentWidth(idx)
	memBar := widgets.RenderProgressBar(s.Metrics.Memory, cw, su, w, a)
	memStr := util.FastPercent1(s.Metrics.Memory)
	memVal := valueStyle.Foreground(widgets.GetColorForValue(s.Metrics.Memory, su, w, a)).Render(util.PadRight(memStr, cw))

	memUsed := uint64(0)
	memTotal := uint64(0)
	if s.Metrics.MemInfo != nil {
		memUsed = s.Metrics.MemInfo.Used
		memTotal = s.Metrics.MemInfo.Total
	}
	memLabel := labelStyle.Render(util.FormatBytes(memUsed)) + sp(" / ") + labelStyle.Render(util.FormatBytes(memTotal))
	memLine := util.PadRight(memLabel, cw)

	memBlock := lipgloss.JoinVertical(lipgloss.Left,
		memVal,
		memBar,
		memLine,
	)

	// Disk (Index 2) - Use cached disk data from partitions
	idx = 2
	cw = getContentWidth(idx)
	diskBar := widgets.RenderProgressBar(s.Metrics.Disk, cw, su, w, a)
	diskStr := util.FastPercent1(s.Metrics.Disk)
	diskVal := valueStyle.Foreground(widgets.GetColorForValue(s.Metrics.Disk, su, w, a)).Render(util.PadRight(diskStr, cw))

	diskUsed := uint64(0)
	diskTotal := uint64(0)
	if len(s.Metrics.DiskPartitions) > 0 {
		diskUsed = s.Metrics.DiskPartitions[0].Used
		diskTotal = s.Metrics.DiskPartitions[0].Total
	}
	diskLabel := labelStyle.Render(util.FormatBytes(diskUsed)) + sp(" / ") + labelStyle.Render(util.FormatBytes(diskTotal))
	diskLine := util.PadRight(diskLabel, cw)

	diskBlock := lipgloss.JoinVertical(lipgloss.Left,
		diskVal,
		diskBar,
		diskLine,
	)

	// Network (Index 3)
	idx = 3
	cw = getContentWidth(idx)
	nP := CalcNetPercent(s)
	netBar := widgets.RenderProgressBar(nP, cw, su, w, a)
	netStr := util.FastPercent1(nP)
	netVal := valueStyle.Foreground(widgets.GetColorForValue(nP, su, w, a)).Render(util.PadRight(netStr, cw))

	recvMb := s.Metrics.NetRecvRate * 8
	sentMb := s.Metrics.NetSentRate * 8

	recvStr := util.FastMbPerSec(recvMb)
	sentStr := util.FastMbPerSec(sentMb)

	netLabel := labelStyle.Render("↓") + sp(" ") + labelStyle.Render(recvStr) + sp(" ") + labelStyle.Render("↑") + sp(" ") + labelStyle.Render(sentStr)
	netLine := util.PadRight(netLabel, cw)

	netBlock := lipgloss.JoinVertical(lipgloss.Left,
		netVal,
		netBar,
		netLine,
	)

	// Quick Stats (Index 4)
	idx = 4
	cw = getContentWidth(idx)

	uptimeStr := "Loading..."
	if s.Metrics.HostInfo != nil {
		uptimeStr = util.FormatDuration(time.Duration(s.Metrics.HostInfo.Uptime) * time.Second)
	}

	loadStr := "Loading..."
	if s.Metrics.LoadAvg != nil {
		if s.Metrics.LoadAvg.Load1 == 0 && s.Metrics.LoadAvg.Load5 == 0 && s.Metrics.LoadAvg.Load15 == 0 {
			loadStr = "N/A (Windows)"
		} else {
			loadStr = util.FastFloat2(s.Metrics.LoadAvg.Load1) + ", " + util.FastFloat2(s.Metrics.LoadAvg.Load5) + ", " + util.FastFloat2(s.Metrics.LoadAvg.Load15)
		}
	}

	batteryStr := "N/A"
	if len(s.Metrics.Battery) > 0 {
		batt := s.Metrics.Battery[0]
		pct := 0.0
		if batt.Full > 0 {
			pct = batt.Current / batt.Full * 100
		}
		batteryStr = util.FastPercent(pct) + " (" + batt.State.String() + ")"
	}

	qsLine := func(label, value string, indent int) string {
		lbl := labelStyle.Render(label) + sp(strings.Repeat(" ", indent))
		val := valueStyle.Render(value)
		line := lbl + val
		return util.PadRight(line, cw)
	}

	quickStatsBlock := lipgloss.JoinVertical(lipgloss.Left,
		qsLine("Uptime:", uptimeStr, 6),
		qsLine("Processes:", util.FastInt(s.Process.ProcessCount), 3),
		qsLine("Load Avg:", loadStr, 4),
		qsLine("Battery:", batteryStr, 5),
		qsLine("Health:", util.FastPercent(float64(s.Metrics.HealthScore)), 6),
	)

	// GPU (Index 5) - Display ALL GPUs
	var gpuBlock string
	if len(s.Metrics.GpuInfo) > 0 {
		idx = 5
		cw = getContentWidth(idx)

		var gpuLines []string
		for _, gpu := range s.Metrics.GpuInfo {
			gpuName := valueStyle.Render(util.PadRight(gpu.Name, cw))

			var memStr string
			if gpu.MemoryTotal != "N/A" && gpu.MemoryTotal != "Shared" {
				if gpu.MemoryUsed != "N/A" {
					memStr = util.FastMbUsed(gpu.MemoryTotal, gpu.MemoryUsed)
				} else {
					memStr = gpu.MemoryTotal + " MB"
				}
			} else {
				memStr = gpu.MemoryTotal
			}

			memLine := labelStyle.Render("VRAM:") + sp(" ") + valueStyle.Render(memStr)
			memLine = util.PadRight(memLine, cw)

			driverLine := labelStyle.Render("Driver:") + sp(" ") + valueStyle.Render(gpu.Driver)
			driverLine = util.PadRight(driverLine, cw)

			var extraLines []string
			if gpu.Temperature != "" && gpu.Temperature != "N/A" {
				extraLines = append(extraLines, labelStyle.Render("Temp:")+sp(" ")+valueStyle.Render(gpu.Temperature))
			}
			if gpu.Utilization != "" && gpu.Utilization != "N/A" {
				extraLines = append(extraLines, labelStyle.Render("GPU:")+sp(" ")+valueStyle.Render(gpu.Utilization))
			}
			if gpu.PowerUsage != "" && gpu.PowerUsage != "N/A" {
				extraLines = append(extraLines, labelStyle.Render("Power:")+sp(" ")+valueStyle.Render(gpu.PowerUsage))
			}
			if gpu.ClockSpeed != "" && gpu.ClockSpeed != "N/A" {
				extraLines = append(extraLines, labelStyle.Render("Clock:")+sp(" ")+valueStyle.Render(gpu.ClockSpeed))
			}

			var gpuEntry []string
			gpuEntry = append(gpuEntry, gpuName, memLine, driverLine)
			for _, el := range extraLines {
				gpuEntry = append(gpuEntry, util.PadRight(el, cw))
			}
			gpuLines = append(gpuLines, gpuEntry...)
		}

		gpuBlock = lipgloss.JoinVertical(lipgloss.Left, gpuLines...)
	}

	// Swap (Index 6) - Use cached SwapInfo
	swapBlock := ""
	if gpuBlock != "" {
		idx = 6
	} else {
		idx = 5
	}
	if s.Metrics.Swap > 0 && s.Metrics.SwapInfo != nil {
		cw = getContentWidth(idx)
		swapBar := widgets.RenderProgressBar(s.Metrics.Swap, cw, su, w, a)
		swapStr := util.FastPercent1(s.Metrics.Swap)
		swapVal := valueStyle.Foreground(widgets.GetColorForValue(s.Metrics.Swap, su, w, a)).Render(util.PadRight(swapStr, cw))

		swapLabel := labelStyle.Render(util.FormatBytes(s.Metrics.SwapInfo.Used)) + sp(" / ") + labelStyle.Render(util.FormatBytes(s.Metrics.SwapInfo.Total))
		swapLine := util.PadRight(swapLabel, cw)

		swapBlock = lipgloss.JoinVertical(lipgloss.Left,
			swapVal,
			swapBar,
			swapLine,
		)
	}

	blocks := []string{cpuBlock, memBlock, diskBlock, netBlock, quickStatsBlock}
	if gpuBlock != "" {
		blocks = append(blocks, gpuBlock)
	}
	if swapBlock != "" {
		blocks = append(blocks, swapBlock)
	}

	// Top Consumers
	topStr := "N/A (Lazy Load)"

	topIdx := len(blocks)
	topCw := colWidths[topIdx%cols] - 4

	if len(s.Process.Processes) > 0 {
		var sb strings.Builder
		limit := 4
		for i := 0; i < len(s.Process.Processes) && i < limit; i++ {
			p := s.Process.Processes[i]
			name := p.Name
			if len(name) > 12 {
				name = name[:9] + "..."
			}
			line := name + " " + util.FastPercent1(p.Cpu)
			sb.WriteString(util.PadRight(line, topCw))
			sb.WriteString("\n")
		}
		topStr = sb.String()
	} else {
		topStr = util.PadRight(topStr, topCw)
	}
	blocks = append(blocks, topStr)

	numRows := (len(blocks) + cols - 1) / cols
	rowHeight := availHeight / numRows
	contentHeight := rowHeight - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	titles := []string{"CPU USAGE", "MEMORY", "DISK USAGE", "NETWORK", "QUICK STATS"}
	if gpuBlock != "" {
		titles = append(titles, "GPU")
	}
	if swapBlock != "" {
		titles = append(titles, "SWAP")
	}
	titles = append(titles, "TOP CONSUMERS")

	var renderedBlocks []string
	for i, block := range blocks {
		title := titles[i]
		bw := colWidths[i%cols]

		blockWithMargin := lipgloss.NewStyle().MarginTop(1).Render(block)
		c := container.Height(contentHeight).BorderTop(false).Padding(0, 1)
		body := c.Render(blockWithMargin)
		topBorder := widgets.RenderTopBorderWithBg(title, bw, border, b, p)

		renderedBlocks = append(renderedBlocks, lipgloss.JoinVertical(lipgloss.Left, topBorder, body))
	}

	for i := len(renderedBlocks); i < numRows*cols; i++ {
		bw := colWidths[i%cols]
		emptyBlock := lipgloss.Place(bw, rowHeight, lipgloss.Center, lipgloss.Center, "")
		renderedBlocks = append(renderedBlocks, emptyBlock)
	}

	var rows []string
	for i := 0; i < len(renderedBlocks); i += cols {
		end := i + cols
		if end > len(renderedBlocks) {
			end = len(renderedBlocks)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, renderedBlocks[i:end]...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

package tabs

import (
	"image/color"
	"runtime"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
	"samouly.fun/bubbleMonitor/internal/util"
)

func RenderSystem(s *data.AppState, container, titleStyle, labelStyle, valueStyle lipgloss.Style, t, mu, p, b, bg, su, w, a color.Color, availHeight int, activeBlock int) string {
	if s.Metrics.HostInfo == nil {
		return "Loading system information..."
	}

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
		return colWidths[i%cols] - 6
	}

	fwLine := func(str string, idx int) string {
		return util.FullWidthBg(str, getContentWidth(idx))
	}

	sp := func(str string) string { return str }

	// Host Info (Index 0)
	idx := 0
	hostInfo := lipgloss.JoinVertical(lipgloss.Left,
		fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Hostname:"+sp(" ")), valueStyle.Render(s.Metrics.HostInfo.Hostname)), idx),
		fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("OS:"+sp("       ")), valueStyle.Render(s.Metrics.HostInfo.OS)), idx),
		fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Platform:"+sp(" ")), valueStyle.Render(s.Metrics.HostInfo.Platform)), idx),
		fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Kernel:"+sp("   ")), valueStyle.Render(s.Metrics.HostInfo.KernelVersion)), idx),
	)

	// CPU Info (Index 1) - Model + Usage
	idx = 1
	cw := getContentWidth(idx)
	cpuModel := "N/A"
	if len(s.Metrics.CpuInfoStatic) > 0 {
		cpuModel = s.Metrics.CpuInfoStatic[0].ModelName
	}

	cpuBar := widgets.RenderProgressBar(s.Metrics.Cpu, cw, su, w, a)
	cpuStr := util.FastPercent1(s.Metrics.Cpu)
	cpuVal := valueStyle.Foreground(widgets.GetColorForValue(s.Metrics.Cpu, su, w, a)).Render(util.PadRight(cpuStr, cw))

	var cpuTemp string
	if s.Metrics.CpuTemp > 0 {
		cpuTemp = util.FastTemp(s.Metrics.CpuTemp)
	} else {
		cpuTemp = "N/A (Admin?)"
	}

	lAS := "N/A"
	if s.Metrics.LoadAvg != nil {
		lAS = util.FastFloat2(s.Metrics.LoadAvg.Load1) + ", " + util.FastFloat2(s.Metrics.LoadAvg.Load5) + ", " + util.FastFloat2(s.Metrics.LoadAvg.Load15)
	}

	loadLine := labelStyle.Render("Load:") + sp(" ") + labelStyle.Render(lAS)
	loadLine = util.PadRight(loadLine, cw)

	tempLine := labelStyle.Render("Temp:") + sp(" ") + labelStyle.Render(cpuTemp)
	tempLine = util.PadRight(tempLine, cw)

	cpuInfo := lipgloss.JoinVertical(lipgloss.Left,
		fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Model:")+sp("    "), valueStyle.Render(cpuModel)), idx),
		fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Cores:")+sp("    "), valueStyle.Render(util.FastInt(runtime.NumCPU())+" physical,")+sp(" ")+valueStyle.Render(util.FastInt(runtime.GOMAXPROCS(0))+" logical")), idx),
		cpuVal,
		cpuBar,
		loadLine,
		tempLine,
	)

	// Memory Info (Index 2)
	idx = 2
	memTotal := uint64(0)
	memAvailable := uint64(0)
	memUsed := uint64(0)
	memUsedPct := 0.0
	if s.Metrics.MemInfo != nil {
		memTotal = s.Metrics.MemInfo.Total
		memAvailable = s.Metrics.MemInfo.Available
		memUsed = s.Metrics.MemInfo.Used
		memUsedPct = s.Metrics.MemInfo.UsedPercent
	}

	memBar := widgets.RenderProgressBar(s.Metrics.Memory, getContentWidth(idx), su, w, a)
	memStr := util.FastPercent1(s.Metrics.Memory)
	memVal := valueStyle.Foreground(widgets.GetColorForValue(s.Metrics.Memory, su, w, a)).Render(util.PadRight(memStr, getContentWidth(idx)))

	memLabel := labelStyle.Render(util.FormatBytes(memUsed)) + sp(" / ") + labelStyle.Render(util.FormatBytes(memTotal))
	memLine := util.PadRight(memLabel, getContentWidth(idx))

	memInfo := lipgloss.JoinVertical(lipgloss.Left,
		fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Total:"+sp("     ")), valueStyle.Render(util.FormatBytes(memTotal))), idx),
		fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Available:"+sp(" ")), valueStyle.Render(util.FormatBytes(memAvailable))), idx),
		fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Used:"+sp("      ")), valueStyle.Render(util.FormatBytes(memUsed)+sp(" ")+"("+util.FastFloat2(memUsedPct)+"%)")), idx),
		memVal,
		memBar,
		memLine,
	)

	// GPU Info - Display ALL GPUs as separate blocks
	var gpuBlocks []string
	var gpuTitles []string
	for gpuIdx, gpu := range s.Metrics.GpuInfo {
		idx = 5 + gpuIdx

		var gpuLines []string

		gpuHeader := "GPU " + util.FastInt(gpuIdx+1)
		gpuName := gpu.Name
		if gpuName == "" {
			gpuName = "Unknown"
		}
		gpuLines = append(gpuLines,
			fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(gpuHeader+":")+sp(" "), valueStyle.Render(gpuName)), idx),
		)

		if gpu.Vendor != "" {
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Vendor:")+sp(" "), valueStyle.Render(gpu.Vendor)), idx),
			)
		}

		if gpu.Driver != "" {
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Driver:")+sp(" "), valueStyle.Render(gpu.Driver)), idx),
			)
		}

		if gpu.Type != "" {
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Type:")+sp(" "), valueStyle.Render(gpu.Type)), idx),
			)
		}

		if gpu.MemoryTotal != "" && gpu.MemoryTotal != "N/A" && gpu.MemoryTotal != "Shared" {
			memStr := gpu.MemoryTotal + " MB"
			if gpu.MemoryUsed != "" && gpu.MemoryUsed != "N/A" {
				memStr = gpu.MemoryUsed + " / " + memStr
			}
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("VRAM:")+sp(" "), valueStyle.Render(memStr)), idx),
			)
		}

		if gpu.Utilization != "" && gpu.Utilization != "N/A" {
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Use:")+sp(" "), valueStyle.Render(gpu.Utilization)), idx),
			)
		}

		if gpu.Temperature != "" && gpu.Temperature != "N/A" {
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Temp:")+sp(" "), valueStyle.Render(gpu.Temperature)), idx),
			)
		}

		if gpu.PowerUsage != "" && gpu.PowerUsage != "N/A" {
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Power:")+sp(" "), valueStyle.Render(gpu.PowerUsage)), idx),
			)
		}

		if gpu.ClockSpeed != "" && gpu.ClockSpeed != "N/A" {
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Clock:")+sp(" "), valueStyle.Render(gpu.ClockSpeed)), idx),
			)
		}

		if gpu.FanSpeed != "" && gpu.FanSpeed != "N/A" {
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("Fan:")+sp(" "), valueStyle.Render(gpu.FanSpeed)), idx),
			)
		}

		if gpu.Slot != "" {
			gpuLines = append(gpuLines,
				fwLine(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render("PCI:")+sp(" "), valueStyle.Render(gpu.Slot)), idx),
			)
		}

		if len(gpuLines) > 0 {
			gpuBlock := lipgloss.JoinVertical(lipgloss.Left, gpuLines...)
			gpuBlocks = append(gpuBlocks, gpuBlock)
			gpuTitles = append(gpuTitles, gpuHeader)
		}
	}

	// Disks (Index 4)
	idx = 4
	cw = getContentWidth(idx)

	diskUsageBlock := ""
	var unmountedParts []data.DiskPartition

	// The data layer (provider/system/disk.go) is responsible for putting
	// real usage data on every partition that has it — mounted filesystems
	// from gopsutil, swap enriched from OS-level stats, etc. Anything still
	// showing UsedPct < 0 has no usage data available, so we render it as
	// a compact row under "Unmounted".
	for _, dp := range s.Metrics.DiskPartitions {
		if dp.UsedPct < 0 {
			unmountedParts = append(unmountedParts, dp)
			continue
		}
		name := dp.Mountpoint
		if name == "" {
			name = dp.Device
		}
		nameLine := labelStyle.Render(name) + sp(" ") + labelStyle.Render("("+dp.Fstype+")")
		nameLine = util.PadRight(nameLine, cw)
		bar := widgets.RenderProgressBar(dp.UsedPct, cw, su, w, a)
		pctStr := util.FastPercent1(dp.UsedPct)
		val := valueStyle.Foreground(widgets.GetColorForValue(dp.UsedPct, su, w, a)).Render(util.PadRight(pctStr, cw))
		label := labelStyle.Render(util.FormatBytes(dp.Used)) + sp(" / ") + labelStyle.Render(util.FormatBytes(dp.Total))
		labelLine := util.PadRight(label, cw)
		entry := lipgloss.JoinVertical(lipgloss.Left, nameLine, labelLine, val, bar)

		if diskUsageBlock != "" {
			diskUsageBlock += "\n"
		}
		diskUsageBlock += entry
	}

	if len(unmountedParts) > 0 {
		if diskUsageBlock != "" {
			diskUsageBlock += "\n" + labelStyle.Render(strings.Repeat("─", cw))
		}
		diskUsageBlock += "\n" + labelStyle.Render("Unmounted")
		for _, dp := range unmountedParts {
			name := dp.Device
			if name == "" {
				name = dp.Mountpoint
			}
			diskUsageBlock += "\n" + labelStyle.Render("─ ") +
				labelStyle.Render(util.PadRight(name, cw-12)) +
				labelStyle.Render(util.FormatBytes(dp.Total))
		}
	}

	// If no disks, show placeholder
	if diskUsageBlock == "" {
		diskUsageBlock = labelStyle.Render("No disks")
	}

	// Network (Index 5)
	idx = 5
	cw = getContentWidth(idx)

	netName := "Network"
	netType := ""
	if len(s.Metrics.NetworkInterfaces) > 0 && s.Metrics.NetworkInterfaces[0].Name != "" {
		netName = s.Metrics.NetworkInterfaces[0].Name
		if strings.HasPrefix(netName, "eth") || strings.HasPrefix(netName, "en") {
			netType = "Ethernet"
		} else if strings.HasPrefix(netName, "wl") || strings.HasPrefix(netName, "wlan") {
			netType = "Wireless"
		} else if strings.HasPrefix(netName, "lo") {
			netType = "Loopback"
		} else if strings.HasPrefix(netName, "docker") || strings.HasPrefix(netName, "br-") {
			netType = "Virtual"
		} else if strings.HasPrefix(netName, "usb") {
			netType = "USB"
		}
	}

	netNameLine := labelStyle.Render(netName)
	if netType != "" {
		netNameLine = netNameLine + sp(" ") + labelStyle.Render("("+netType+")")
	}
	netNameLine = util.PadRight(netNameLine, cw)

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
		netNameLine,
		netVal,
		netBar,
		netLine,
	)

	// Quick Stats (Index 3)
	idx = 3
	cw = getContentWidth(idx)

	uptimeStr := "Loading..."
	if s.Metrics.HostInfo != nil {
		uptimeStr = util.FormatDuration(time.Duration(s.Metrics.HostInfo.Uptime) * time.Second)
	}

	bootTime := time.Unix(int64(s.Metrics.HostInfo.BootTime), 0)
	bootStr := bootTime.Format("2006-01-02")
	monitorStr := util.FormatDuration(time.Since(s.Config.StartTime))

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
		qsLine("Boot:", bootStr, 7),
		qsLine("Monitor:", monitorStr, 3),
		qsLine("Uptime:", uptimeStr, 6),
		qsLine("Processes:", util.FastInt(s.Process.ProcessCount), 3),
		qsLine("Battery:", batteryStr, 5),
		qsLine("Health:", util.FastPercent(float64(s.Metrics.HealthScore)), 6),
	)

	// Top Consumers
	topStr := "No processes"

	topIdx := 12
	topCw := colWidths[topIdx%cols] - 4

	if len(s.Process.Processes) > 0 {
		var sb strings.Builder
		limit := data.TopProcessesTrackCount
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
	}

	// Build all blocks
	blocks := []string{hostInfo, cpuInfo, memInfo}
	if len(gpuBlocks) > 0 {
		blocks = append(blocks, gpuBlocks...)
	}
	// Add remaining Overview content
	blocks = append(blocks, diskUsageBlock, netBlock, quickStatsBlock, topStr)

	s.UI.SystemBlockCount = len(blocks)

	numRows := (len(blocks) + cols - 1) / cols
	rowHeight := availHeight / numRows
	contentHeight := max(rowHeight-2, 1)

	titles := []string{"HOST", "CPU", "MEMORY"}
	if len(gpuBlocks) > 0 {
		titles = append(titles, gpuTitles...)
	}
	titles = append(titles, "DISKS", "NETWORK", "QUICK STATS", "TOP CONSUMERS")

	var renderedBlocks []string
	for i, block := range blocks {
		title := titles[i]
		bw := colWidths[i%cols]

		blockLines := strings.Split(block, "\n")
		totalLines := len(blockLines)

		scrollOffset := 0
		if s.UI.SystemBlockScrollOffsets != nil {
			scrollOffset = s.UI.SystemBlockScrollOffsets[i]
		}
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		maxScroll := max(totalLines-contentHeight, 0)
		if scrollOffset > maxScroll {
			scrollOffset = maxScroll
		}

		needsScrollbar := totalLines > contentHeight
		if s.UI.SystemBlockScrollable == nil {
			s.UI.SystemBlockScrollable = make(map[int]bool)
		}
		s.UI.SystemBlockScrollable[i] = needsScrollbar
		if s.UI.SystemBlockMaxScroll == nil {
			s.UI.SystemBlockMaxScroll = make(map[int]int)
		}
		s.UI.SystemBlockMaxScroll[i] = maxScroll

		var visibleLines []string
		if totalLines <= contentHeight {
			visibleLines = blockLines
			if len(visibleLines) < contentHeight {
				for j := len(visibleLines); j < contentHeight; j++ {
					visibleLines = append(visibleLines, strings.Repeat(" ", getContentWidth(i)))
				}
			}
		} else {
			endLine := scrollOffset + contentHeight
			if endLine > totalLines {
				endLine = totalLines
			}
			visibleLines = blockLines[scrollOffset:endLine]
			for j := len(visibleLines); j < contentHeight; j++ {
				visibleLines = append(visibleLines, strings.Repeat(" ", getContentWidth(i)))
			}
		}

		visibleContent := lipgloss.JoinVertical(lipgloss.Left, visibleLines...)

		c := container.Width(bw).Height(contentHeight).BorderTop(false)
		body := c.Render(visibleContent)

		scrollInfo := ""
		if needsScrollbar {
			scrollInfo = " [" + util.FastInt(scrollOffset+1) + "-" + util.FastInt(scrollOffset+contentHeight) + " of " + util.FastInt(totalLines) + "]"
		}

		borderColor := b
		if i == activeBlock {
			borderColor = p
		}

		topBorder := widgets.RenderTopBorderWithBg(title+scrollInfo, bw, border, borderColor, p)

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

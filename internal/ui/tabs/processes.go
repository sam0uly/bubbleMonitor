package tabs

import (
	"image/color"
	"runtime"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/input"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
	"samouly.fun/bubbleMonitor/internal/util"
)

// RenderProcesses renders the processes tab
func RenderProcesses(s *data.AppState, visibleProcs []data.ProcessInfo, treeIndents map[int32]int, container lipgloss.Style, su, w, a, t, mu, p, b color.Color, availHeight int, mouseY int, listStartY int, zoneManager input.ZoneManager) string {
	boxWidth := s.UI.Width
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	// Split available height: processes list (70%) + details panel (30%)
	detailsHeight := ProcessDetailsHeight
	listHeight := availHeight - detailsHeight - 1
	if listHeight < 10 {
		listHeight = 10
		detailsHeight = availHeight - listHeight - 1
	}

	contentWidth := boxWidth - 4

	pidWidth := PIDColumnWidth
	statusWidth := StatusColumnWidth
	cpuWidth := CPUColumnWidth
	memWidth := MemColumnWidth

	nameWidth := max(contentWidth-pidWidth-statusWidth-cpuWidth-memWidth-4, 20)

	sp := func(str string) string { return str }

	sI, mSI, pSI, nSI := "", "", "", ""
	dirInd := " ▼"
	if s.Process.SortDirection == "asc" {
		dirInd = " ▲"
	}
	switch s.Process.SortBy {
	case "cpu":
		sI = dirInd
	case "mem":
		mSI = dirInd
	case "pid":
		pSI = dirInd
	case "name":
		nSI = dirInd
	}

	hdrStyle := lipgloss.NewStyle().Bold(true).Foreground(p)
	pidH := util.PadRight("PID"+pSI, pidWidth)
	nameH := util.PadRight("NAME"+nSI, nameWidth)
	statusH := util.PadRight("STATUS", statusWidth)
	cpuH := util.PadLeft("CPU"+sI, cpuWidth)
	memH := util.PadLeft("MEM"+mSI, memWidth)

	headerRow := hdrStyle.Render(pidH) + sp(" ") +
		hdrStyle.Render(nameH) + sp(" ") +
		hdrStyle.Render(statusH) + sp(" ") +
		hdrStyle.Render(cpuH) + sp(" ") +
		hdrStyle.Render(memH)

	filtered := visibleProcs

	visibleRows := max(listHeight-4, 1)

	startIdx := s.Process.ProcessScrollOffset
	endIdx := min(startIdx+visibleRows, len(filtered))
	if startIdx >= len(filtered) {
		startIdx = 0
		endIdx = min(visibleRows, len(filtered))
	}

	selColor := compat.AdaptiveColor{Light: lipgloss.Color("#E0E7FF"), Dark: lipgloss.Color("#3730A3")}
	cellStyle := lipgloss.NewStyle()

	// Pre-allocate styles without widths to avoid ansi parser overhead in the loop
	// Standard cells - no background, apply background only to entire row
	pidStyle := cellStyle
	statusStyle := cellStyle
	nameStyle := cellStyle

	// Metric styles (Low/Med/High) - Include Align (Width handled by Sprintf)
	// CPU and Mem have same width (8)
	baseMetric := cellStyle.Align(lipgloss.Right)

	styleLow := baseMetric.Foreground(su)
	styleMed := baseMetric.Foreground(w)
	styleHigh := baseMetric.Foreground(a)

	getStyle := func(val float64) lipgloss.Style {
		if val < 50 {
			return styleLow
		}
		if val < 80 {
			return styleMed
		}
		return styleHigh
	}

	// Use shared builder to reduce allocations
	sb := &s.UI.ContentBuilder
	sb.Reset()
	// Pre-allocate estimate: 100 rows * 100 chars = 10KB
	sb.Grow(10240)

	// Write header
	sb.WriteString(headerRow)
	sb.WriteString("\n\n") // Double newline for header spacing

	var selectedProc *data.ProcessInfo

	// Build children map for tree view (from visible processes to determine last child)
	childrenMap := make(map[int32][]int32, len(filtered))
	for i := range filtered {
		childrenMap[filtered[i].Ppid] = append(childrenMap[filtered[i].Ppid], filtered[i].Pid)
	}

	// Build parent lookup map for displaying parent names - O(n) using pid index
	pidToName := make(map[int32]string, len(filtered))
	for i := range filtered {
		pidToName[filtered[i].Pid] = filtered[i].Name
	}

	parentMap := make(map[int32]string, len(filtered))
	for i := range filtered {
		if filtered[i].Ppid != 0 {
			if parentName, ok := pidToName[filtered[i].Ppid]; ok {
				parentMap[filtered[i].Pid] = parentName
			}
		}
	}

	// Determine last child for each parent
	isLastChild := make(map[int32]bool, len(filtered))
	for _, children := range childrenMap {
		if len(children) > 0 {
			isLastChild[children[len(children)-1]] = true
		}
	}

	// Pre-calculate hasMoreAtLevel - O(n) using a single pass from end
	hasMoreAtLevel := make(map[int]bool)
	if s.Process.TreeView && len(filtered) > 0 {
		// Track the highest level seen as we iterate backwards
		maxLevelSeen := 0
		for i := len(filtered) - 1; i >= 0; i-- {
			if level, ok := treeIndents[filtered[i].Pid]; ok && level > 0 {
				// If current level is less than max level seen, there are more at this level
				if level < maxLevelSeen {
					hasMoreAtLevel[level] = true
				}
				if level > maxLevelSeen {
					maxLevelSeen = level
				}
			}
		}
	}

	type procInfo struct {
		zoneID     string
		index      int
		rowY       int
		isSelected bool
		proc       data.ProcessInfo
	}
	var procInfos []procInfo

	for i := startIdx; i < endIdx; i++ {
		proc := filtered[i]
		rowY := listStartY + 3 + (i - startIdx)
		procInfos = append(procInfos, procInfo{
			index:      i,
			proc:       proc,
			isSelected: i == s.Process.SelectedProcess,
			rowY:       rowY,
			zoneID:     "process-row-" + util.FastInt64(int64(proc.Pid)),
		})

		if zoneManager != nil {
			procIndex := i
			procPid := proc.Pid
			zoneManager.Register(input.Zone{
				ID:     "process-row-" + util.FastInt64(int64(proc.Pid)),
				Type:   input.ZoneTypeListItem,
				X:      0,
				Y:      rowY,
				Width:  s.UI.Width,
				Height: 1,
				Metadata: map[string]any{
					"pid":   procPid,
					"index": procIndex,
				},
				OnClick: func() tea.Cmd {
					s.Process.SelectedProcess = procIndex
					return nil
				},
			})
		}
	}

	zoneManager.UpdateMousePos(s.UI.MouseX, s.UI.MouseY)

	for _, info := range procInfos {
		i := info.index
		proc := info.proc
		isSelected := info.isSelected
		zoneID := info.zoneID
		noOverlay := !s.Process.ShowOpenFiles && !s.Process.ShowProcessMenu && !s.UI.ShowSettings && !s.UI.ShowHelp && !s.UI.ShowKillDialog && !s.UI.ShowSamLab
		isHovered := noOverlay && !isSelected && zoneManager.IsHovered(zoneID)

		if isSelected {
			selectedProc = &proc
		}

		pStyle := pidStyle
		nStyle := nameStyle
		sStyle := statusStyle

		if i < 5 {
			var topColor color.Color
			switch i {
			case 0:
				topColor = a
			case 1:
				topColor = w
			case 2:
				topColor = su
			case 3:
				topColor = p
			case 4:
				topColor = t
			}
			pStyle = pStyle.Foreground(topColor)
			nStyle = nStyle.Foreground(topColor)
			sStyle = sStyle.Foreground(topColor)
		}

		cpuVal := proc.Cpu
		if s.Config.ProcessCpuNormalized {
			cpuVal = cpuVal / float64(runtime.NumCPU())
		}

		cpuStyle := getStyle(cpuVal)
		memStyle := getStyle(proc.Memory)

		name := proc.Name
		if s.IsBookmarked(proc.Pid) {
			name = "★ " + name
		}

		// Tree view prefix and suffixes
		var treePrefix strings.Builder
		var treeSuffix string
		if s.Process.TreeView {
			if level, ok := treeIndents[proc.Pid]; ok {
				// Build tree prefix with proper box-drawing characters
				if level > 0 {
					// For each ancestor level (1 to level-1), add continuation or space
					for l := 1; l < level; l++ {
						if hasMoreAtLevel[l] {
							treePrefix.WriteString("│   ")
						} else {
							treePrefix.WriteString("    ")
						}
					}
					// Add branch character: ├── for non-last child, └── for last child
					if isLastChild[proc.Pid] {
						treePrefix.WriteString("└──")
					} else {
						treePrefix.WriteString("├──")
					}
				}

				// Add collapsed/expanded indicator and children count/parent name
				children := childrenMap[proc.Ppid]
				if level > 0 && len(children) > 0 {
					// This is a child process - show parent name
					if parentName, ok := parentMap[proc.Pid]; ok && parentName != "" {
						treeSuffix = " (parent: " + parentName + ")"
					}
				} else if len(childrenMap[proc.Pid]) > 0 {
					// This is a parent process - show children count
					treeSuffix = " (" + strconv.Itoa(len(childrenMap[proc.Pid])) + " children)"
				}
			}
			// Apply tree prefix to name
			name = treePrefix.String() + name + treeSuffix
		}

		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		status := proc.Status
		if status == "" {
			status = "running"
		}

		if s.IsSuspended(proc.Pid) {
			status = "SUSPENDED"
		}

		if len(status) > statusWidth {
			status = status[:statusWidth-3] + "..."
		}

		// Use util.Pad* for proper visual alignment (handles unicode/emoji)
		pidStr := util.PadRight(strconv.Itoa(int(proc.Pid)), pidWidth)

		// Name is already truncated above, just pad it
		nameStr := util.PadRight(name, nameWidth)

		cpuStr := util.FastPercent1(cpuVal)
		memStr := util.FastPercent1(proc.Memory)

		// Right align metrics
		paddedCPU := util.PadLeft(cpuStr, cpuWidth)
		paddedMem := util.PadLeft(memStr, memWidth)

		// Apply background to each cell individually to avoid ANSI reset issues
		var rowBg color.Color
		hasBg := false
		if isSelected {
			rowBg = selColor
			hasBg = true
		} else if isHovered {
			rowBg = mu
			hasBg = true
		}

		// Apply background to each cell style
		if hasBg {
			pStyle = pStyle.Background(rowBg)
			nStyle = nStyle.Background(rowBg)
			sStyle = sStyle.Background(rowBg)
			cpuStyle = cpuStyle.Background(rowBg)
			memStyle = memStyle.Background(rowBg)
		}

		// Render status cell - must be after background is applied to sStyle
		var statusStr string
		paddedStatus := util.PadRight(status, statusWidth)
		if s.IsSuspended(proc.Pid) {
			// Suspended state styling (warning color) - apply foreground to sStyle (which may have background)
			statusStr = sStyle.Foreground(lipgloss.Color("#F59E0B")).Render(paddedStatus)
		} else {
			statusStr = sStyle.Render(paddedStatus)
		}

		// Render cells with background
		pidCell := pStyle.Render(pidStr)
		nameCell := nStyle.Render(nameStr)
		cpuCell := cpuStyle.Render(paddedCPU)
		memCell := memStyle.Render(paddedMem)

		// Render spaces with background
		space := " "
		if hasBg {
			space = lipgloss.NewStyle().Background(rowBg).Render(" ")
		}

		rowContent := pidCell + space + nameCell + space + statusStr + space + cpuCell + space + memCell

		sb.WriteString(rowContent)

		if i < endIdx-1 {
			sb.WriteString("\n")
		}
	}

	scrollInfo := ""
	if len(filtered) > visibleRows {
		scrollInfo = " [" + util.FastInt(startIdx+1) + "-" + util.FastInt(endIdx) + " of " + util.FastInt(len(filtered)) + "]"
	}

	content := sb.String()

	titleText := "PROCESSES (Sort: " + strings.ToUpper(s.Process.SortBy) + ")" + scrollInfo

	listContentHeight := max(listHeight-2, 0)

	c := container.Width(boxWidth).Height(listContentHeight).BorderTop(false)
	body := c.Render(content)

	topBorder := widgets.RenderTopBorderWithBg(titleText, boxWidth, border, b, p)

	listBlock := lipgloss.JoinVertical(lipgloss.Left, topBorder, body)

	detailsBlock := renderProcessDetails(s, selectedProc, container, boxWidth, detailsHeight-2, t, mu, p, b, su, w, a)

	var filterIndicator string
	if s.Process.FilterMode {
		filterIndicator = lipgloss.NewStyle().
			Foreground(p).
			Bold(true).
			MarginLeft(2).
			Render(" Filter: " + s.Process.ProcessFilter + "█")
	} else if s.Process.ProcessFilter != "" {
		filterIndicator = lipgloss.NewStyle().
			Foreground(mu).
			MarginLeft(2).
			Render(" Filter: " + s.Process.ProcessFilter + " (press 'c' to clear, 'f' to edit)")
	}

	result := lipgloss.JoinVertical(lipgloss.Left, listBlock, detailsBlock)
	if filterIndicator != "" {
		result = lipgloss.JoinVertical(lipgloss.Left, filterIndicator, result)
	}

	return result
}

// renderProcessDetails renders the details panel for the selected process
func renderProcessDetails(s *data.AppState, proc *data.ProcessInfo, container lipgloss.Style, boxWidth, contentHeight int, t, mu, p, b, su, w, a color.Color) string {
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)
	contentWidth := boxWidth - 4

	if proc == nil {
		c := container.Width(boxWidth).Height(contentHeight).BorderTop(false)
		body := c.Render(lipgloss.NewStyle().Foreground(mu).Render("No process selected - use j/k or ↑↓ to navigate"))
		topBorder := widgets.RenderTopBorderWithBg("PROCESS DETAILS", boxWidth, border, b, p)
		return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
	}

	col1Width := contentWidth / 3
	col2Width := contentWidth / 3
	col3Width := contentWidth - col1Width - col2Width

	labelStyle := lipgloss.NewStyle().Foreground(mu)
	valueStyle := lipgloss.NewStyle().Foreground(t).Bold(true)

	statusColor := su
	status := proc.Status
	if status == "" {
		status = "running"
	}

	if s.IsSuspended(proc.Pid) {
		status = "SUSPENDED"
		statusColor = compat.AdaptiveColor{Light: lipgloss.Color("#F59E0B"), Dark: lipgloss.Color("#F59E0B")}
	}

	memStr := util.FormatBytes(proc.MemoryBytes)

	username := proc.Username
	// Check lazy-loaded username cache first
	if username == "" {
		if cachedUsername, ok := s.Process.ProcessUsernames[proc.Pid]; ok {
			username = cachedUsername
		}
	}
	if len(username) > 15 {
		username = username[:12] + "..."
	}
	if username == "" {
		username = "N/A"
	}

	cpuColor := widgets.GetColorForValue(proc.Cpu, su, w, a)
	memColor := widgets.GetColorForValue(proc.Memory, su, w, a)

	leftCol := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("PID: ")+valueStyle.Render(util.FastInt64(int64(proc.Pid))),
		labelStyle.Render("Name: ")+valueStyle.Render(proc.Name),
		labelStyle.Render("Status: ")+lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(status),
	)

	midCol := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("User: ")+valueStyle.Render(username),
		labelStyle.Render("Nice: ")+valueStyle.Render(util.FastInt(int(proc.Nice))),
		labelStyle.Render("PPID: ")+valueStyle.Render(util.FastInt64(int64(proc.Ppid))),
	)

	rightCol := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("CPU: ")+lipgloss.NewStyle().Foreground(cpuColor).Bold(true).Render(util.FastPercent1(proc.Cpu)),
		labelStyle.Render("Memory: ")+lipgloss.NewStyle().Foreground(memColor).Bold(true).Render(util.FastPercent1(proc.Memory)+" ("+memStr+")"),
		labelStyle.Render("Started: ")+valueStyle.Render(time.Unix(proc.CreateTime/1000, 0).Format("15:04:05")),
	)

	details := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(col1Width).Render(leftCol),
		lipgloss.NewStyle().Width(col2Width).Render(midCol),
		lipgloss.NewStyle().Width(col3Width).Render(rightCol),
	)

	var chart string
	hist, ok := s.GetHistory(proc.Pid)
	if ok && hist.Len() > 2 {
		chartW := contentWidth - 14
		if chartW > 10 {
			chartVal := widgets.RenderSparkline(hist, chartW, 1, p, w, 100.0, s.Config.Theme)
			chart = lipgloss.JoinHorizontal(lipgloss.Left, labelStyle.Render("CPU History: "), chartVal)
		}
	}

	finalContent := details
	if chart != "" {
		finalContent = lipgloss.JoinVertical(lipgloss.Left, details, "", chart)
	}

	c := container.Width(boxWidth).Height(contentHeight).BorderTop(false)
	body := c.Render(finalContent)
	topBorder := widgets.RenderTopBorderWithBg("PROCESS DETAILS", boxWidth, border, b, p)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
}

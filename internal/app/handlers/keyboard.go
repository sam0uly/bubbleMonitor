package handlers

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
	messages "samouly.fun/bubbleMonitor/internal/msg"
	"samouly.fun/bubbleMonitor/internal/provider/process"
)

var TitleCaser = cases.Title(language.English)

func KillProcessCmd(pid int32) tea.Cmd {
	return func() tea.Msg {
		proc, err := os.FindProcess(int(pid))
		if err != nil {
			return messages.KillProcessMsg{Pid: pid, Success: false, Error: err.Error()}
		}
		err = proc.Kill()
		if err != nil {
			return messages.KillProcessMsg{Pid: pid, Success: false, Error: err.Error()}
		}
		return messages.KillProcessMsg{Pid: pid, Success: true}
	}
}

func HandleKey(m *data.AppState, msg tea.KeyMsg, getVisibleProcessesFunc func() ([]data.ProcessInfo, map[int32]int), invalidateCacheFunc func(), getFilteredProcessCountFunc func() int, getVisibleProcessRowsFunc func() int) tea.Cmd {
	currentTab := ""
	if msg.String() == "space" {
		currentTab = ""
		if m.UI.SelectedTab >= 0 && m.UI.SelectedTab < len(m.UI.ActiveTabs) {
			currentTab = m.UI.ActiveTabs[m.UI.SelectedTab]
		}
		if m.Process.TreeView && currentTab == "Processes" {
			procs, _ := getVisibleProcessesFunc()
			if m.Process.SelectedProcess >= 0 && m.Process.SelectedProcess < len(procs) {
				proc := procs[m.Process.SelectedProcess]
				m.ToggleCollapsed(proc.Pid)
				invalidateCacheFunc()
			}
		}
		return nil
	}

	if m.UI.ShowKillDialog {
		switch msg.String() {
		case "left", "h":
			m.UI.KillDialogSel = 0
		case "right", "l":
			m.UI.KillDialogSel = 1
		case "y":
			if m.UI.KillDialogSel == 0 {
				pid := m.Process.KillTargetPid
				m.UI.ShowKillDialog = false
				m.Process.KillTargetPid = 0
				m.Process.KillTargetName = ""
				m.UI.KillDialogSel = 0
				return KillProcessCmd(pid)
			}
		case "n", "esc":
			m.UI.ShowKillDialog = false
			m.Process.KillTargetPid = 0
			m.Process.KillTargetName = ""
			m.UI.KillDialogSel = 0
		case "enter":
			if m.UI.KillDialogSel == 0 {
				pid := m.Process.KillTargetPid
				m.UI.ShowKillDialog = false
				m.Process.KillTargetPid = 0
				m.Process.KillTargetName = ""
				m.UI.KillDialogSel = 0
				return KillProcessCmd(pid)
			}
			m.UI.ShowKillDialog = false
			m.Process.KillTargetPid = 0
			m.Process.KillTargetName = ""
			m.UI.KillDialogSel = 0
		}
		return nil
	}

	if m.UI.ShowHelp {
		if msg.String() == "?" || msg.String() == "esc" {
			m.UI.ShowHelp = false
		}
		return nil
	}

	if m.UI.ShowSamLab {
		if msg.String() == "esc" || msg.String() == "enter" || msg.String() == "q" {
			m.UI.ShowSamLab = false
		}
		return nil
	}

	if m.Process.ShowOpenFiles {
		switch msg.String() {
		case "o", "esc":
			m.Process.ShowOpenFiles = false
			m.Process.OpenFilesView.GotoTop()
		default:
			switch msg.String() {
			case "j", "down":
				m.Process.OpenFilesView.LineDown(1)
			case "k", "up":
				m.Process.OpenFilesView.LineUp(1)
			case "pgdown", "ctrl+d":
				m.Process.OpenFilesView.HalfViewDown()
			case "pgup", "ctrl+u":
				m.Process.OpenFilesView.HalfViewUp()
			case "home":
				m.Process.OpenFilesView.GotoTop()
			case "end":
				m.Process.OpenFilesView.GotoBottom()
			}
			return nil
		}
		return nil
	}

	if m.Process.ShowProcessMenu {
		if msg.String() == "esc" {
			m.Process.ShowProcessMenu = false
			return nil
		}
	}

	if m.Process.FilterMode {
		switch msg.String() {
		case "esc":
			m.Process.FilterMode = false
		case "backspace":
			if len(m.Process.ProcessFilter) > 0 {
				m.Process.ProcessFilter = m.Process.ProcessFilter[:len(m.Process.ProcessFilter)-1]
				m.Process.ProcessFilterLower = strings.ToLower(m.Process.ProcessFilter)
				invalidateCacheFunc()
			}
		case "enter":
			m.Process.FilterMode = false
		default:
			if len(msg.String()) == 1 {
				m.Process.ProcessFilter += msg.String()
				m.Process.ProcessFilterLower = strings.ToLower(m.Process.ProcessFilter)
				invalidateCacheFunc()
			}
		}
		m.Process.SelectedProcess = 0
		m.Process.ProcessScrollOffset = 0
		return nil
	}

	if m.UI.ShowSettings {
		totalSettings := data.TotalSettingsCount
		tabsStart := data.ThresholdCount + data.DisplayCount

		switch msg.String() {
		case "esc", ".":
			m.UI.ShowSettings = false
			if err := config.SaveConfig(m.Config.Config); err != nil {
				return AddToastCmd(fmt.Sprintf("Failed to save config: %v", err), data.ToastError)
			}
			return nil
		case "up", "k":
			m.UI.SettingsIdx = (m.UI.SettingsIdx - 1 + totalSettings) % totalSettings
			if m.UI.SettingsIdx < data.ThresholdCount {
				metrics := []config.MetricType{config.MetricCPU, config.MetricMem, config.MetricDisk, config.MetricTemp}
				m.UI.SettingsSel = metrics[m.UI.SettingsIdx]
			}
		case "down", "j":
			m.UI.SettingsIdx = (m.UI.SettingsIdx + 1) % totalSettings
			if m.UI.SettingsIdx < data.ThresholdCount {
				metrics := []config.MetricType{config.MetricCPU, config.MetricMem, config.MetricDisk, config.MetricTemp}
				m.UI.SettingsSel = metrics[m.UI.SettingsIdx]
			}
		case "+", "=", "right", "l":
			if m.UI.SettingsIdx < data.ThresholdCount {
				curr := m.Config.Config.Thresholds[m.UI.SettingsSel]
				if curr < 100 {
					m.Config.Config.Thresholds[m.UI.SettingsSel] = curr + 1
				}
			} else {
				handleSettingsChange(m, 1)
				return AddToastCmd("Setting Changed", data.ToastSuccess)
			}
		case "-", "_", "left", "h":
			if m.UI.SettingsIdx < data.ThresholdCount {
				curr := m.Config.Config.Thresholds[m.UI.SettingsSel]
				if curr > 0 {
					m.Config.Config.Thresholds[m.UI.SettingsSel] = curr - 1
				}
			} else {
				handleSettingsChange(m, -1)
				return AddToastCmd("Setting Changed", data.ToastSuccess)
			}
		case "[":
			// Move tab earlier in the active tabs list
			if m.UI.SettingsIdx >= tabsStart && m.UI.SettingsIdx < tabsStart+data.TabCount {
				tabName := data.AllAvailableTabs[m.UI.SettingsIdx-tabsStart]
				reorderTab(m, tabName, -1)
			}
		case "]":
			// Move tab later in the active tabs list
			if m.UI.SettingsIdx >= tabsStart && m.UI.SettingsIdx < tabsStart+data.TabCount {
				tabName := data.AllAvailableTabs[m.UI.SettingsIdx-tabsStart]
				reorderTab(m, tabName, 1)
			}
		}
		return nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		if m.UI.SelectedTab >= 0 && m.UI.SelectedTab < len(m.UI.ActiveTabs) {
			m.Config.Config.DefaultTab = m.UI.ActiveTabs[m.UI.SelectedTab]
		}
		if err := config.SaveConfig(m.Config.Config); err != nil {
			return AddToastCmd(fmt.Sprintf("Failed to save config: %v", err), data.ToastError)
		}
		return tea.Quit
	case "e":
		return func() tea.Msg { return messages.ExportSnapshotMsg{} }
	case ".":
		m.UI.ShowSettings = !m.UI.ShowSettings
		if !m.UI.ShowSettings {
			if err := config.SaveConfig(m.Config.Config); err != nil {
				return AddToastCmd(fmt.Sprintf("Failed to save config: %v", err), data.ToastError)
			}
		}
		return nil
	case "space":
		currentTab = ""
		if m.UI.SelectedTab >= 0 && m.UI.SelectedTab < len(m.UI.ActiveTabs) {
			currentTab = m.UI.ActiveTabs[m.UI.SelectedTab]
		}
		if m.Process.TreeView && currentTab == "Processes" {
			procs, _ := getVisibleProcessesFunc()
			if m.Process.SelectedProcess >= 0 && m.Process.SelectedProcess < len(procs) {
				proc := procs[m.Process.SelectedProcess]
				m.ToggleCollapsed(proc.Pid)
				invalidateCacheFunc()
			}
		}
		return nil
	case "n":
		m.Config.ProcessCpuNormalized = !m.Config.ProcessCpuNormalized
		m.Config.Config.ProcessCpuNormalized = m.Config.ProcessCpuNormalized
		if err := config.SaveConfig(m.Config.Config); err != nil {
			return AddToastCmd(fmt.Sprintf("Failed to save config: %v", err), data.ToastError)
		}
		status := "Raw (>100%)"
		if m.Config.ProcessCpuNormalized {
			status = "Normalized (0-100%)"
		}
		return AddToastCmd("CPU Mode: "+status, data.ToastInfo)
	}

	currentTab = ""
	if m.UI.SelectedTab < len(m.UI.ActiveTabs) && m.UI.SelectedTab >= 0 {
		currentTab = m.UI.ActiveTabs[m.UI.SelectedTab]
	}

	switch msg.String() {
	case "tab", "right", "l":
		if len(m.UI.ActiveTabs) > 0 {
			m.UI.SelectedTab = (m.UI.SelectedTab + 1) % len(m.UI.ActiveTabs)
		}
		newTab := ""
		if len(m.UI.ActiveTabs) > 0 && m.UI.SelectedTab >= 0 && m.UI.SelectedTab < len(m.UI.ActiveTabs) {
			newTab = m.UI.ActiveTabs[m.UI.SelectedTab]
		}
		if newTab == "Processes" && !m.Process.ProcessesLoaded {
			m.Process.ProcessesLoaded = true
			return process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection)
		}
		if newTab == "System" && !m.Process.ProcessesLoaded {
			m.Process.ProcessesLoaded = true
			return process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection)
		}
		if currentTab != "Processes" && newTab != "Processes" {
			m.Process.SelectedProcess = 0
			m.Process.ProcessScrollOffset = 0
		}
		if currentTab != "Services" && newTab != "Services" {
			m.Process.ServicesScrollOffset = 0
		}
		if currentTab != "Connections" && newTab != "Connections" {
			m.Process.ConnectionsScrollOffset = 0
		}
		if currentTab != "Logs" && newTab != "Logs" {
			m.Process.LogsScrollOffset = 0
		}
	case "shift+tab", "left", "h":
		if len(m.UI.ActiveTabs) > 0 {
			m.UI.SelectedTab = (m.UI.SelectedTab - 1 + len(m.UI.ActiveTabs)) % len(m.UI.ActiveTabs)
		}
		newTab := ""
		if len(m.UI.ActiveTabs) > 0 && m.UI.SelectedTab >= 0 && m.UI.SelectedTab < len(m.UI.ActiveTabs) {
			newTab = m.UI.ActiveTabs[m.UI.SelectedTab]
		}
		if newTab == "Processes" && !m.Process.ProcessesLoaded {
			m.Process.ProcessesLoaded = true
			return process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection)
		}
		if newTab == "System" && !m.Process.ProcessesLoaded {
			m.Process.ProcessesLoaded = true
			return process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection)
		}
		if currentTab != "Processes" && newTab != "Processes" {
			m.Process.SelectedProcess = 0
			m.Process.ProcessScrollOffset = 0
		}
		if currentTab != "Services" && newTab != "Services" {
			m.Process.ServicesScrollOffset = 0
		}
		if currentTab != "Connections" && newTab != "Connections" {
			m.Process.ConnectionsScrollOffset = 0
		}
		if currentTab != "Logs" && newTab != "Logs" {
			m.Process.LogsScrollOffset = 0
		}
	case "H":
		switch m.UI.HistoryLength {
		case 60:
			m.UI.HistoryLength = 300
		case 300:
			m.UI.HistoryLength = 900
		case 900:
			m.UI.HistoryLength = 3600
		default:
			m.UI.HistoryLength = 60
		}
		m.Config.HistoryLength = m.UI.HistoryLength
		m.Config.Config.HistoryLength = m.UI.HistoryLength
		m.Metrics.CpuHistory.Resize(m.UI.HistoryLength)
		m.Metrics.MemHistory.Resize(m.UI.HistoryLength)
		m.Metrics.NetHistory.Resize(m.UI.HistoryLength)
		m.Metrics.SwapHistory.Resize(m.UI.HistoryLength)
		m.Metrics.HistoryTemp.Resize(m.UI.HistoryLength)
		m.Metrics.DiskHORead.Resize(m.UI.HistoryLength)
		m.Metrics.DiskHOWrite.Resize(m.UI.HistoryLength)
	case "C":
		switch m.UI.ChartType {
		case "line":
			m.UI.ChartType = "bar"
		case "bar":
			m.UI.ChartType = "braille"
		default:
			m.UI.ChartType = "line"
		}
		m.Config.Config.ChartType = m.UI.ChartType
	case "]", "}":
		if currentTab == "System" && m.UI.SystemBlockCount > 0 {
			m.UI.ActiveScrollBlock = (m.UI.ActiveScrollBlock + 1) % m.UI.SystemBlockCount
		}
	case "[", "{":
		if currentTab == "System" && m.UI.SystemBlockCount > 0 {
			m.UI.ActiveScrollBlock = (m.UI.ActiveScrollBlock - 1 + m.UI.SystemBlockCount) % m.UI.SystemBlockCount
		}
	case "pgup":
		if currentTab == "Processes" {
			m.Process.SelectedProcess -= 10
			if m.Process.SelectedProcess < 0 {
				m.Process.SelectedProcess = 0
			}
			if m.Process.SelectedProcess < m.Process.ProcessScrollOffset {
				m.Process.ProcessScrollOffset = m.Process.SelectedProcess
			}
		} else if currentTab == "Metrics" {
			m.UI.CpuCoreScrollOffset -= 4
			if m.UI.CpuCoreScrollOffset < 0 {
				m.UI.CpuCoreScrollOffset = 0
			}
		} else if currentTab == "Services" && len(m.Process.Services) > 0 {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			m.Process.ServicesScrollOffset -= rows
			if m.Process.ServicesScrollOffset < 0 {
				m.Process.ServicesScrollOffset = 0
			}
		} else if currentTab == "Connections" && len(m.Process.Connections) > 0 {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			m.Process.ConnectionsScrollOffset -= rows
			if m.Process.ConnectionsScrollOffset < 0 {
				m.Process.ConnectionsScrollOffset = 0
			}
		} else if currentTab == "Logs" && len(m.Process.SystemLogs) > 0 {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			m.Process.LogsScrollOffset -= rows
			if m.Process.LogsScrollOffset < 0 {
				m.Process.LogsScrollOffset = 0
			}
		} else if currentTab == "System" && m.UI.ActiveScrollBlock >= 0 && m.UI.SystemBlockScrollable[m.UI.ActiveScrollBlock] {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock] -= rows
			if m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock] < 0 {
				m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock] = 0
			}
		}
	case "pgdown":
		if currentTab == "Processes" {
			visibleProcs, _ := getVisibleProcessesFunc()
			filteredLen := len(visibleProcs)
			if filteredLen > 0 {
				m.Process.SelectedProcess += 10
				if m.Process.SelectedProcess >= filteredLen {
					m.Process.SelectedProcess = filteredLen - 1
				}
				rows := getVisibleProcessRowsFunc()
				if m.Process.SelectedProcess >= m.Process.ProcessScrollOffset+rows {
					m.Process.ProcessScrollOffset = m.Process.SelectedProcess - rows + 1
				}
			}
		} else if currentTab == "Metrics" {
			m.UI.CpuCoreScrollOffset += 4
		} else if currentTab == "Services" && len(m.Process.Services) > 0 {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			maxScroll := len(m.Process.Services) - rows
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.Process.ServicesScrollOffset += rows
			if m.Process.ServicesScrollOffset > maxScroll {
				m.Process.ServicesScrollOffset = maxScroll
			}
		} else if currentTab == "Connections" && len(m.Process.Connections) > 0 {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			maxScroll := len(m.Process.Connections) - rows
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.Process.ConnectionsScrollOffset += rows
			if m.Process.ConnectionsScrollOffset > maxScroll {
				m.Process.ConnectionsScrollOffset = maxScroll
			}
		} else if currentTab == "Logs" && len(m.Process.SystemLogs) > 0 {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			maxScroll := len(m.Process.SystemLogs) - rows
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.Process.LogsScrollOffset += rows
			if m.Process.LogsScrollOffset > maxScroll {
				m.Process.LogsScrollOffset = maxScroll
			}
		} else if currentTab == "System" && m.UI.ActiveScrollBlock >= 0 && m.UI.SystemBlockScrollable[m.UI.ActiveScrollBlock] {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			maxScroll := 0
			if m.UI.SystemBlockMaxScroll != nil {
				maxScroll = m.UI.SystemBlockMaxScroll[m.UI.ActiveScrollBlock]
			}
			m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock] += rows
			if m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock] > maxScroll {
				m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock] = maxScroll
			}
		}
	case "1":
		m.UI.SelectedTab = 0
	case "2":
		m.UI.SelectedTab = 1
	case "3":
		m.UI.SelectedTab = 2
	case "4":
		m.UI.SelectedTab = 3
	case "5":
		m.UI.SelectedTab = 4
	case "6":
		m.UI.SelectedTab = 5
	case "7":
		m.UI.SelectedTab = 6
	case "8":
		m.UI.SelectedTab = 7
	case "9":
		m.UI.SelectedTab = 8
	case "s":
		switch m.Process.SortBy {
		case "cpu":
			m.Process.SortBy = "mem"
		case "mem":
			m.Process.SortBy = "pid"
		case "pid":
			m.Process.SortBy = "name"
		default:
			m.Process.SortBy = "cpu"
		}
		m.Config.Config.SortBy = m.Process.SortBy
		invalidateCacheFunc()

	case "S":
		if m.Process.SortDirection == "asc" {
			m.Process.SortDirection = "desc"
		} else {
			m.Process.SortDirection = "asc"
		}
		m.Config.Config.SortDirection = m.Process.SortDirection
		invalidateCacheFunc()

	case "b":
		if currentTab == "Processes" {
			procs, _ := getVisibleProcessesFunc()
			if m.Process.SelectedProcess >= 0 && m.Process.SelectedProcess < len(procs) {
				proc := procs[m.Process.SelectedProcess]
				m.ToggleBookmark(proc.Pid)
				invalidateCacheFunc()
			}
		}
	case "T":
		if currentTab == "Processes" {
			m.Process.TreeView = !m.Process.TreeView
			viewName := "normal"
			if m.Process.TreeView {
				viewName = "tree"
			}
			m.Config.Config.ViewType = viewName
			invalidateCacheFunc()
		}
	case "+", "=":
		if currentTab == "Processes" {
			filtered := m.GetFilteredProcesses()
			if m.Process.SelectedProcess < len(filtered) {
				proc := filtered[m.Process.SelectedProcess]
				return process.ReniceProcessCmdSafe(proc.Pid, -1)
			}
		}
	case "-", "_":
		if currentTab == "Processes" {
			filtered := m.GetFilteredProcesses()
			if m.Process.SelectedProcess < len(filtered) {
				proc := filtered[m.Process.SelectedProcess]
				return process.ReniceProcessCmdSafe(proc.Pid, 1)
			}
		}
	case "o":
		if currentTab == "Processes" {
			if m.Process.ShowOpenFiles {
				m.Process.ShowOpenFiles = false
			} else {
				filtered := m.GetFilteredProcesses()
				if m.Process.SelectedProcess < len(filtered) {
					proc := filtered[m.Process.SelectedProcess]
					m.Process.ShowOpenFiles = true
					m.Process.OpenFilesList = nil
					m.Process.OpenFilesPid = proc.Pid
					return process.FetchOpenFilesCmd(proc.Pid)
				}
			}
		}

	case "z":
		if currentTab == "Processes" {
			filtered := m.GetFilteredProcesses()
			if m.Process.SelectedProcess < len(filtered) {
				proc := filtered[m.Process.SelectedProcess]
				return process.SuspendProcessCmd(proc.Pid)
			}
		}
	case "x":
		if currentTab == "Processes" {
			filtered := m.GetFilteredProcesses()
			if m.Process.SelectedProcess < len(filtered) {
				proc := filtered[m.Process.SelectedProcess]
				return process.ResumeProcessCmd(proc.Pid)
			}
		}
	case "p":
		m.UI.Paused = !m.UI.Paused
	case "r":
		return func() tea.Msg { return messages.ForceRefreshMsg{} }
	case "?":
		m.UI.ShowHelp = true
		m.UI.LastError = ""

	case "j", "down":
		if currentTab == "Processes" {
			visibleProcs, _ := getVisibleProcessesFunc()
			filteredLen := len(visibleProcs)

			if filteredLen > 0 && m.Process.SelectedProcess < filteredLen-1 {
				m.Process.SelectedProcess++
				visibleRows := getVisibleProcessRowsFunc()
				if m.Process.SelectedProcess >= m.Process.ProcessScrollOffset+visibleRows {
					m.Process.ProcessScrollOffset = m.Process.SelectedProcess - visibleRows + 1
				}
			}
		} else if currentTab == "Services" && len(m.Process.Services) > 0 {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			maxScroll := len(m.Process.Services) - rows
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.Process.ServicesScrollOffset < maxScroll {
				m.Process.ServicesScrollOffset++
			}
		} else if currentTab == "Connections" && len(m.Process.Connections) > 0 {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			maxScroll := len(m.Process.Connections) - rows
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.Process.ConnectionsScrollOffset < maxScroll {
				m.Process.ConnectionsScrollOffset++
			}
		} else if currentTab == "Logs" && len(m.Process.SystemLogs) > 0 {
			rows := m.UI.Height - data.ReservedContentRows
			if rows < 1 {
				rows = 1
			}
			maxScroll := len(m.Process.SystemLogs) - rows
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.Process.LogsScrollOffset < maxScroll {
				m.Process.LogsScrollOffset++
			}
		}
		if currentTab == "System" && m.UI.SystemBlockCount > 0 && m.UI.ActiveScrollBlock >= 0 && m.UI.SystemBlockScrollable[m.UI.ActiveScrollBlock] {
			maxScroll := 0
			if m.UI.SystemBlockMaxScroll != nil {
				maxScroll = m.UI.SystemBlockMaxScroll[m.UI.ActiveScrollBlock]
			}
			m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock]++
			if m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock] > maxScroll {
				m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock] = maxScroll
			}
		}
	case "k", "up":
		if currentTab == "Processes" {
			if m.Process.SelectedProcess > 0 {
				m.Process.SelectedProcess--
				if m.Process.SelectedProcess < m.Process.ProcessScrollOffset {
					m.Process.ProcessScrollOffset = m.Process.SelectedProcess
				}
			}
		} else if currentTab == "Services" && m.Process.ServicesScrollOffset > 0 {
			m.Process.ServicesScrollOffset--
		} else if currentTab == "Connections" && m.Process.ConnectionsScrollOffset > 0 {
			m.Process.ConnectionsScrollOffset--
		} else if currentTab == "Logs" && m.Process.LogsScrollOffset > 0 {
			m.Process.LogsScrollOffset--
		} else if currentTab == "System" && m.UI.SystemBlockCount > 0 && m.UI.ActiveScrollBlock >= 0 && m.UI.SystemBlockScrollable[m.UI.ActiveScrollBlock] {
			if m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock] > 0 {
				m.UI.SystemBlockScrollOffsets[m.UI.ActiveScrollBlock]--
			}
		}
	case "K":
		if currentTab == "Processes" && len(m.Process.Processes) > 0 {
			filtered := m.GetFilteredProcesses()
			if m.Process.SelectedProcess < len(filtered) {
				proc := filtered[m.Process.SelectedProcess]
				m.UI.ShowKillDialog = true
				m.Process.KillTargetPid = proc.Pid
				m.Process.KillTargetName = proc.Name
				m.UI.KillDialogSel = 0
			}
		}
	case "f":
		if currentTab == "Processes" {
			m.Process.FilterMode = true
		}
	case "c":
		if currentTab == "Processes" {
			m.Process.ProcessFilter = ""
			m.Process.ProcessFilterLower = ""
			m.Process.SelectedProcess = 0
			m.Process.ProcessScrollOffset = 0
			invalidateCacheFunc()
		}
	case "g":
		if currentTab == "Processes" {
			m.Process.SelectedProcess = 0
			m.Process.ProcessScrollOffset = 0
		}
	case "G":
		if currentTab == "Processes" {
			filteredLen := getFilteredProcessCountFunc()
			if filteredLen > 0 {
				m.Process.SelectedProcess = filteredLen - 1
				visibleRows := getVisibleProcessRowsFunc()
				if m.Process.SelectedProcess >= visibleRows {
					m.Process.ProcessScrollOffset = m.Process.SelectedProcess - visibleRows + 1
				}
			}
		}
	case "home":
		if currentTab == "Services" {
			m.Process.ServicesScrollOffset = 0
		} else if currentTab == "Connections" {
			m.Process.ConnectionsScrollOffset = 0
		} else if currentTab == "Logs" {
			m.Process.LogsScrollOffset = 0
		} else if currentTab == "Processes" {
			m.Process.SelectedProcess = 0
			m.Process.ProcessScrollOffset = 0
		} else if currentTab == "System" && m.UI.SystemBlockCount > 0 {
			for i := 0; i < m.UI.SystemBlockCount; i++ {
				m.UI.SystemBlockScrollOffsets[i] = 0
			}
		}
	case "end":
		if currentTab == "Services" && len(m.Process.Services) > 0 {
			rows := max(m.UI.Height-19, 1)
			maxScroll := max(len(m.Process.Services)-rows, 0)
			m.Process.ServicesScrollOffset = maxScroll
		} else if currentTab == "Connections" && len(m.Process.Connections) > 0 {
			rows := max(m.UI.Height-19, 1)
			maxScroll := max(len(m.Process.Connections)-rows, 0)
			m.Process.ConnectionsScrollOffset = maxScroll
		} else if currentTab == "Logs" && len(m.Process.SystemLogs) > 0 {
			rows := max(m.UI.Height-19, 1)
			maxScroll := max(len(m.Process.SystemLogs)-rows, 0)
			m.Process.LogsScrollOffset = maxScroll
		} else if currentTab == "System" && m.UI.SystemBlockCount > 0 {
			for i := 0; i < m.UI.SystemBlockCount; i++ {
				m.UI.SystemBlockScrollOffsets[i] = 1000
			}
		} else if currentTab == "Processes" {
			filteredLen := getFilteredProcessCountFunc()
			if filteredLen > 0 {
				m.Process.SelectedProcess = filteredLen - 1
				visibleRows := getVisibleProcessRowsFunc()
				if m.Process.SelectedProcess >= visibleRows {
					m.Process.ProcessScrollOffset = m.Process.SelectedProcess - visibleRows + 1
				}
			}
		}
	}
	return nil
}

func handleSettingsChange(m *data.AppState, dir int) {
	displayStart := data.ThresholdCount
	tabsStart := displayStart + data.DisplayCount
	appearanceStart := tabsStart + data.TabCount

	switch m.UI.SettingsIdx {
	case displayStart:
		types := []string{"line", "bar", "braille"}
		for i, t := range types {
			if t == m.UI.ChartType {
				nextIdx := (i + dir + len(types)) % len(types)
				m.UI.ChartType = types[nextIdx]
				break
			}
		}
		m.Config.Config.ChartType = m.UI.ChartType

	case displayStart + 1:
		m.Process.TreeView = !m.Process.TreeView
		viewName := "normal"
		if m.Process.TreeView {
			viewName = "tree"
		}
		m.Config.Config.ViewType = viewName
		m.InvalidateProcessCache()

	case displayStart + 2:
		opts := []string{"cpu", "mem", "pid", "name"}
		for i, o := range opts {
			if o == m.Process.SortBy {
				nextIdx := (i + dir + len(opts)) % len(opts)
				m.Process.SortBy = opts[nextIdx]
				break
			}
		}
		m.Config.Config.SortBy = m.Process.SortBy
		m.InvalidateProcessCache()

	case displayStart + 3:
		lens := []int{60, 300, 900, 3600}
		for i, l := range lens {
			if l == m.UI.HistoryLength {
				nextIdx := (i + dir + len(lens)) % len(lens)
				m.UI.HistoryLength = lens[nextIdx]
				break
			}
		}
		m.Config.HistoryLength = m.UI.HistoryLength
		m.Config.Config.HistoryLength = m.UI.HistoryLength
		m.Metrics.CpuHistory.Resize(m.UI.HistoryLength)
		m.Metrics.MemHistory.Resize(m.UI.HistoryLength)
		m.Metrics.NetHistory.Resize(m.UI.HistoryLength)
		m.Metrics.SwapHistory.Resize(m.UI.HistoryLength)
		m.Metrics.HistoryTemp.Resize(m.UI.HistoryLength)
		m.Metrics.DiskHORead.Resize(m.UI.HistoryLength)
		m.Metrics.DiskHOWrite.Resize(m.UI.HistoryLength)

	case displayStart + 4:
		m.Config.ProcessCpuNormalized = !m.Config.ProcessCpuNormalized
		m.Config.Config.ProcessCpuNormalized = m.Config.ProcessCpuNormalized

	case displayStart + 5:
		opts := []string{"asc", "desc"}
		currIdx := 0
		for i, o := range opts {
			if m.Process.SortDirection == o {
				currIdx = i
				break
			}
		}
		newIdx := (currIdx + dir + len(opts)) % len(opts)
		m.Process.SortDirection = opts[newIdx]
		m.Config.Config.SortDirection = m.Process.SortDirection
		m.InvalidateProcessCache()

	default:
		// Tab toggle items
		if m.UI.SettingsIdx >= tabsStart && m.UI.SettingsIdx < tabsStart+data.TabCount {
			tabIdx := m.UI.SettingsIdx - tabsStart
			targetTab := data.AllAvailableTabs[tabIdx]

			idxInActive := -1
			for i, t := range m.UI.ActiveTabs {
				if t == targetTab {
					idxInActive = i
					break
				}
			}

			if idxInActive >= 0 {
				m.UI.ActiveTabs = append(m.UI.ActiveTabs[:idxInActive], m.UI.ActiveTabs[idxInActive+1:]...)
			} else {
				m.UI.ActiveTabs = append(m.UI.ActiveTabs, targetTab)
			}
			m.Config.Config.Tabs = m.UI.ActiveTabs
			return
		}

		// Appearance items
		appIdx := m.UI.SettingsIdx - appearanceStart
		switch appIdx {
		case 0:
			themes := config.GetThemeNames()
			for i, t := range themes {
				if t == m.Config.Theme {
					nextIdx := (i + dir + len(themes)) % len(themes)
					m.Config.Theme = themes[nextIdx]
					break
				}
			}
			m.Config.Config.Theme = m.Config.Theme
			if m.Config.Theme == "custom" && m.Config.Config.CustomTheme == nil {
				m.Config.Config.CustomTheme = config.DefaultCustomTheme()
			}

		case 1:
			rates := config.GetRefreshRates()
			for i, r := range rates {
				if r == m.Config.RefreshRate {
					nextIdx := (i + dir + len(rates)) % len(rates)
					m.Config.RefreshRate = rates[nextIdx]
					break
				}
			}
			m.Config.Config.RefreshRate = m.Config.RefreshRate

		case 2:
			types := config.GetBorderTypes()
			for i, t := range types {
				if t == m.Config.BorderType {
					nextIdx := (i + dir + len(types)) % len(types)
					m.Config.BorderType = types[nextIdx]
					break
				}
			}
			m.Config.Config.BorderType = m.Config.BorderType

		case 3:
			styles := config.GetBorderStyles()
			for i, s := range styles {
				if s == m.Config.BorderStyle {
					nextIdx := (i + dir + len(styles)) % len(styles)
					m.Config.BorderStyle = styles[nextIdx]
					break
				}
			}
			m.Config.Config.BorderStyle = m.Config.BorderStyle

		case 4:
			m.Config.BackgroundOpaque = !m.Config.BackgroundOpaque
			m.Config.Config.BackgroundOpaque = m.Config.BackgroundOpaque
		}
	}
}

// reorderTab moves a tab earlier (dir=-1) or later (dir=+1) in the active tabs list.
func reorderTab(m *data.AppState, tabName string, dir int) {
	idx := -1
	for i, t := range m.UI.ActiveTabs {
		if t == tabName {
			idx = i
			break
		}
	}
	if idx < 0 {
		// Tab is not currently active, nothing to reorder
		return
	}
	newIdx := idx + dir
	if newIdx < 0 || newIdx >= len(m.UI.ActiveTabs) {
		return
	}
	m.UI.ActiveTabs[idx], m.UI.ActiveTabs[newIdx] = m.UI.ActiveTabs[newIdx], m.UI.ActiveTabs[idx]
	m.Config.Config.Tabs = m.UI.ActiveTabs
}

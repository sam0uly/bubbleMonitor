package app

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/shirou/gopsutil/v3/net"

	"samouly.fun/bubbleMonitor/internal/app/handlers"
	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
	messages "samouly.fun/bubbleMonitor/internal/msg"
	"samouly.fun/bubbleMonitor/internal/provider/process"
)

// Update handles all messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.renderCache.Force = true
		m.UI.Width = msg.Width
		m.UI.Height = msg.Height

		if m.Process.ShowOpenFiles && m.Process.OpenFilesPid > 0 && m.Process.OpenFilesList == nil {
			return m, m.Provider.Process.FetchOpenFilesCmd(m.Process.OpenFilesPid)
		}

	case tea.MouseMsg:
		switch msg.(type) {
		case tea.MouseClickMsg, tea.MouseWheelMsg:
			m.renderCache.Force = true
		}
		return m, handlers.HandleMouse(&m.AppState, msg)

	case messages.KillProcessMsg:
		m.UI.ShowKillDialog = false
		m.Process.KillTargetPid = 0
		m.Process.KillTargetName = ""
		m.UI.KillDialogSel = 0
		return m, m.Provider.Process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection)

	case messages.PriorityChangeMsg:
		if msg.Err != nil {
			return m, handlers.AddToastCmd(fmt.Sprintf("Priority Error: %v", msg.Err), data.ToastError)
		}
		return m, tea.Batch(m.Provider.Process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection), handlers.AddToastCmd("Priority Changed", data.ToastSuccess))

	case messages.ProcessControlMsg:
		if msg.Err != nil {
			return m, handlers.AddToastCmd(fmt.Sprintf("%s Failed: %v", handlers.TitleCaser.String(msg.Action), msg.Err), data.ToastError)
		}
		switch msg.Action {
		case "suspend":
			m.SetSuspended(msg.Pid, true)
		case "resume":
			m.SetSuspended(msg.Pid, false)
		}
		return m, tea.Batch(m.Provider.Process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection), handlers.AddToastCmd(fmt.Sprintf("Process %sd", handlers.TitleCaser.String(msg.Action)), data.ToastSuccess))

	case messages.ToastMsg:
		return m, handlers.HandleToast(&m.AppState, msg)

	case messages.ToastTimeoutMsg:
		handlers.HandleToastTimeout(&m.AppState, msg)

	case messages.OpenFilesMsg:
		m.Process.OpenFilesPid = msg.Pid

		var lines []string
		if msg.Err != nil {
			lines = []string{fmt.Sprintf("Error: %v", msg.Err), "", "Try running as root for full access."}
		} else {
			m.Process.OpenFilesList = msg.Files
			for _, f := range msg.Files {
				lines = append(lines, f.Path)
			}
			if len(lines) == 0 {
				lines = []string{"No open files found or access denied.", "", "Try running as root for full access."}
			}
		}
		content := strings.Join(lines, "\n")
		m.Process.OpenFilesView.SetContent(content)
		m.Process.OpenFilesView.GotoTop()

		return m, nil

	case messages.ProcessCmdlineMsg:
		if msg.Cmdline != "" {
			m.Process.ProcessCmdlines[msg.Pid] = msg.Cmdline
		}

	case messages.ProcessUsernameMsg:
		if msg.Username != "" {
			m.Process.ProcessUsernames[msg.Pid] = msg.Username
		}

	case messages.OpenFilesRequestMsg:
		m.Process.OpenFilesPid = msg.Pid
		m.Process.OpenFilesList = nil
		m.Process.ShowOpenFiles = true
		return m, m.Provider.Process.FetchOpenFilesCmd(msg.Pid)

	case messages.ExportSnapshotMsg:
		return m, SaveSnapshotCmd(m.Metrics.Cpu, m.Metrics.Memory, m.Metrics.Disk, m.Process.ProcessCount)

	case messages.ForceRefreshMsg:
		cmds := []tea.Cmd{
			m.Provider.System.FastMetricsCmd(),
			m.Provider.System.SlowMetricsCmd(),
			m.Provider.System.HostInfoCmd(),
		}
		currentTab := ""
		if m.UI.SelectedTab < len(m.UI.ActiveTabs) && m.UI.SelectedTab >= 0 {
			currentTab = m.UI.ActiveTabs[m.UI.SelectedTab]
		}
		if currentTab == "Processes" || currentTab == "System" {
			cmds = append(cmds, m.Provider.Process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection))
		}
		return m, tea.Batch(cmds...)

	case messages.QuitMsg:
		if m.UI.SelectedTab >= 0 && m.UI.SelectedTab < len(m.UI.ActiveTabs) {
			m.Config.Config.DefaultTab = m.UI.ActiveTabs[m.UI.SelectedTab]
		}
		if err := config.SaveConfig(m.Config.Config); err != nil {
			return m, handlers.AddToastCmd(fmt.Sprintf("Failed to save config: %v", err), data.ToastError)
		}
		return m, tea.Quit

	case config.ConfigWatchTickMsg:
		return m, config.WatchConfig(m.Config.LastConfigModTime)

	case config.ConfigChangeMsg:
		m.Config.LastConfigModTime = msg.NewModTime
		newConfig, err := config.LoadConfig()
		if err == nil {
			if reflect.DeepEqual(m.Config, newConfig) {
				return m, config.WatchConfig(m.Config.LastConfigModTime)
			}

			m.Config.Config = newConfig
			if m.UI.HistoryLength != newConfig.HistoryLength {
				m.UI.HistoryLength = newConfig.HistoryLength
				m.Config.HistoryLength = newConfig.HistoryLength
				m.Metrics.CpuHistory.Resize(m.UI.HistoryLength)
				m.Metrics.MemHistory.Resize(m.UI.HistoryLength)
				m.Metrics.NetHistory.Resize(m.UI.HistoryLength)
				m.Metrics.SwapHistory.Resize(m.UI.HistoryLength)
				m.Metrics.HistoryTemp.Resize(m.UI.HistoryLength)
				m.Metrics.DiskHORead.Resize(m.UI.HistoryLength)
				m.Metrics.DiskHOWrite.Resize(m.UI.HistoryLength)
			}
			m.UI.ChartType = newConfig.ChartType
			m.Process.SortBy = newConfig.SortBy
			m.Process.SortDirection = newConfig.SortDirection
			m.Config.ProcessCpuNormalized = newConfig.ProcessCpuNormalized
			oldTreeView := m.Process.TreeView
			m.Process.TreeView = newConfig.ViewType == "tree"
			if oldTreeView != m.Process.TreeView {
				m.InvalidateProcessCache()
			}
			m.UI.ActiveTabs = newConfig.Tabs
			m.Config.Theme = newConfig.Theme
			m.Config.RefreshRate = newConfig.RefreshRate
			m.Config.BorderType = newConfig.BorderType
			m.Config.BorderStyle = newConfig.BorderStyle
			m.Config.BackgroundOpaque = newConfig.BackgroundOpaque
			return m, tea.Batch(config.WatchConfig(m.Config.LastConfigModTime), handlers.AddToastCmd("Config Reloaded", data.ToastSuccess))
		}
		return m, tea.Batch(config.WatchConfig(m.Config.LastConfigModTime), handlers.AddToastCmd(fmt.Sprintf("Config Error: %v", err), data.ToastError))

	case tea.KeyPressMsg:
		m.renderCache.Force = true
		return m, handlers.HandleKey(
			&m.AppState,
			msg,
			m.GetVisibleProcesses,
			m.InvalidateProcessCache,
			m.getFilteredProcessCount,
			m.getVisibleProcessRows,
		)
	case tea.KeyReleaseMsg:
		return m, nil

	case messages.TickMsg:
		if m.Metrics.AlertManager != nil {
			m.Metrics.AlertManager.CheckAlerts(&m.AppState)
		}

		m.UI.TickCount++
		cmds := []tea.Cmd{
			m.Provider.System.TickCmd(time.Duration(m.Config.RefreshRate) * time.Millisecond),
		}

		cmds = append(cmds, m.Provider.System.FastMetricsCmd())

		currentTab := ""
		if m.UI.SelectedTab < len(m.UI.ActiveTabs) && m.UI.SelectedTab >= 0 {
			currentTab = m.UI.ActiveTabs[m.UI.SelectedTab]
		}

		switch currentTab {
		case "Network":
			if m.Metrics.HasNetworkInterfaces {
				cmds = append(cmds, m.Provider.System.NetworkInterfacesCmd())
			}
		case "Disks":
			if m.Metrics.HasDiskIO {
				cmds = append(cmds, m.Provider.System.DiskIOCmd())
			}
		case "System":
			if m.Metrics.HasTempSensors {
				cmds = append(cmds, m.Provider.System.TempCmd())
			}
		case "Metrics":
			if m.Metrics.HasDiskIO {
				cmds = append(cmds, m.Provider.System.DiskIOCmd())
			}
			if m.Metrics.HasNetworkInterfaces {
				cmds = append(cmds, m.Provider.System.NetworkInterfacesCmd())
			}
			if m.Metrics.HasTempSensors {
				cmds = append(cmds, m.Provider.System.TempCmd())
			}
		default:
		}

		if (currentTab == "Processes" || currentTab == "System") && m.Process.ProcessesLoaded {
			cmds = append(cmds, m.Provider.Process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection))
		} else if currentTab != "Processes" && currentTab != "System" && m.UI.TickCount%2 == 0 {
			cmds = append(cmds, m.Provider.Process.ProcessCountCmd())
		}

		if currentTab == "Processes" && m.Process.SelectedProcess >= 0 && m.Process.SelectedProcess < len(m.Process.Processes) {
			selectedPid := m.Process.Processes[m.Process.SelectedProcess].Pid
			if _, ok := m.Process.ProcessCmdlines[selectedPid]; !ok {
				cmds = append(cmds, m.Provider.Process.FetchProcessCmdlineCmd(selectedPid))
			}
			if _, ok := m.Process.ProcessUsernames[selectedPid]; !ok {
				cmds = append(cmds, m.Provider.Process.FetchProcessUsernameCmd(selectedPid))
			}
		}

		if currentTab == "Connections" {
			cmds = append(cmds, m.Provider.System.ConnectionsCmd())
		}

		if m.UI.TickCount%2 == 0 {
			cmds = append(cmds, m.Provider.System.SlowMetricsCmd(), m.Provider.System.HostInfoCmd())

			if m.Metrics.HasBattery {
				cmds = append(cmds, m.Provider.System.BatteryCmd())
			}

			if currentTab == "Services" && m.Metrics.HasServices {
				cmds = append(cmds, m.Provider.System.ServicesCmd())
			}
			if currentTab == "Logs" {
				cmds = append(cmds, m.Provider.System.SystemLogsCmd())
			}
			if currentTab == "Remote" {
				for _, h := range m.Config.Config.RemoteHosts {
					cmds = append(cmds, m.Provider.Remote.CheckRemoteCmd(h))
				}
			}
		}

		return m, tea.Batch(cmds...)

	case messages.CpuMemMsg:
		if msg.Err != nil {
			return m, handlers.AddToastCmd("CPU/Memory metrics error: "+msg.Err.Error(), data.ToastError)
		}
		m.Metrics.Cpu = msg.Cpu
		m.Metrics.CpuPerCore = msg.CpuPerCore
		m.Metrics.Memory = msg.Memory
		m.Metrics.Swap = msg.Swap
		m.Metrics.LoadAvg = msg.LoadAvg
		m.Metrics.MemInfo = msg.MemInfo
		m.Metrics.SwapInfo = msg.SwapInfo

		m.Metrics.CpuHistory.Push(m.Metrics.Cpu)
		m.Metrics.MemHistory.Push(m.Metrics.Memory)
		m.Metrics.SwapHistory.Push(m.Metrics.Swap)
		UpdateHealthScore(&m.AppState)

	case messages.DiskNetMsg:
		if msg.Err != nil {
			return m, handlers.AddToastCmd("Disk/Network metrics error: "+msg.Err.Error(), data.ToastError)
		}
		if m.Metrics.LastNetSent > 0 && m.Metrics.LastNetRecv > 0 {
			m.Metrics.NetSentRate = float64(msg.NetSent-m.Metrics.LastNetSent) / 1024 / 1024
			m.Metrics.NetRecvRate = float64(msg.NetRecv-m.Metrics.LastNetRecv) / 1024 / 1024
		}
		if m.Metrics.NetSentRate < 0 {
			m.Metrics.NetSentRate = 0
		}
		if m.Metrics.NetRecvRate < 0 {
			m.Metrics.NetRecvRate = 0
		}

		m.Metrics.LastNetSent = msg.NetSent
		m.Metrics.LastNetRecv = msg.NetRecv
		m.Metrics.Disk = msg.Disk

		totalNetRate := m.Metrics.NetSentRate + m.Metrics.NetRecvRate
		netPercent := (totalNetRate / 10) * 100
		if netPercent > 100 {
			netPercent = 100
		}

		m.Metrics.NetHistory.Push(netPercent)

	case messages.ProcessCountMsg:
		m.Process.ProcessCount = int(msg)

	case messages.ProcessesMsg:
		allProcesses := msg
		m.Process.ProcessCount = len(allProcesses)
		m.Process.ProcessesLoaded = true

		currentPids := make(map[int32]bool, len(allProcesses))
		for _, p := range allProcesses {
			currentPids[p.Pid] = true
		}

		m.PruneDeadProcessMaps(currentPids)

		maxProcesses := len(allProcesses)
		if len(allProcesses) > maxProcesses {
			allProcesses = allProcesses[:maxProcesses]
		}

		if cap(m.Process.Processes) > 0 {
			oldSlice := m.Process.Processes
			process.PutProcSlice(&oldSlice)
		}

		m.Process.Processes = allProcesses
		m.SyncProcessesMap()
		m.InvalidateProcessCache()
		UpdateHealthScore(&m.AppState)
		UpdateProcessHistory(&m.AppState, currentPids)

		filteredLen := m.getFilteredProcessCount()
		if m.Process.SelectedProcess >= filteredLen {
			m.Process.SelectedProcess = filteredLen - 1
			if m.Process.SelectedProcess < 0 {
				m.Process.SelectedProcess = 0
			}
		}

	case messages.HostInfoMsg:
		if msg.Err != nil {
			return m, handlers.AddToastCmd("Host info error: "+msg.Err.Error(), data.ToastError)
		}
		m.Metrics.HostInfo = msg.Info
	case messages.DiskInfoMsg:
		if msg.Err != nil {
			return m, handlers.AddToastCmd("Disk partitions error: "+msg.Err.Error(), data.ToastError)
		}
		m.Metrics.DiskPartitions = msg.Partitions
	case messages.GpuInfoMsg:
		m.Metrics.GpuInfo = msg.Gpus
		if msg.Err != nil {
			return m, handlers.AddToastCmd("GPU: "+msg.Err.Error(), data.ToastError)
		}
	case messages.DiskIOMsg:
		if len(m.Metrics.LastDiskIO) > 0 {
			var totalRead, totalWrite uint64
			var lastTotalRead, lastTotalWrite uint64

			for k, v := range msg {
				totalRead += v.ReadBytes
				totalWrite += v.WriteBytes
				if last, ok := m.Metrics.LastDiskIO[k]; ok {
					lastTotalRead += last.ReadBytes
					lastTotalWrite += last.WriteBytes
				}
			}

			if totalRead >= lastTotalRead {
				m.Metrics.DiskReadRate = float64(totalRead-lastTotalRead) / 1024 / 1024
			}
			if totalWrite >= lastTotalWrite {
				m.Metrics.DiskWriteRate = float64(totalWrite-lastTotalWrite) / 1024 / 1024
			}

			m.Metrics.DiskHORead.Push(m.Metrics.DiskReadRate)
			m.Metrics.DiskHOWrite.Push(m.Metrics.DiskWriteRate)
		}
		m.Metrics.DiskIO = msg
		m.Metrics.LastDiskIO = msg
	case messages.TempMsg:
		m.Metrics.Sensors = msg
		var totalTemp float64
		var count int
		for i := range msg {
			t := &msg[i]
			key := strings.ToLower(t.SensorKey)
			if strings.Contains(key, "core") || strings.Contains(key, "cpu") || strings.Contains(key, "package") {
				if t.Temperature > 0 {
					totalTemp += t.Temperature
					count++
				}
			}
		}
		if count > 0 {
			m.Metrics.CpuTemp = totalTemp / float64(count)
		} else if len(msg) > 0 {
			m.Metrics.CpuTemp = msg[0].Temperature
		}

		m.Metrics.HistoryTemp.Push(m.Metrics.CpuTemp)

	case messages.NetworkInterfacesMsg:
		if msg.Err != nil {
			return m, handlers.AddToastCmd("Network interfaces error: "+msg.Err.Error(), data.ToastError)
		}
		m.Metrics.NetworkInterfaces = msg.Interfaces
		if m.Metrics.LastNetworkInterfaces == nil {
			m.Metrics.LastNetworkInterfaces = make(map[string]net.IOCountersStat)
		}

		for _, nic := range msg.Interfaces {
			m.Metrics.LastNetworkInterfaces[nic.Name] = nic
		}
		m.renderCache.Force = true

	case messages.BatteryMsg:
		m.Metrics.Battery = msg
	case messages.ServicesMsg:
		m.Process.Services = msg
		m.renderCache.Force = true
	case messages.ConnectionsMsg:
		m.Process.Connections = msg
		m.renderCache.Force = true
	case messages.SysLogMsg:
		m.Process.SystemLogs = msg
		m.renderCache.Force = true
	case messages.RemoteMsg:
		m.Remote.Metrics[msg.Host] = msg.Metrics
	}
	return m, nil
}

func (m *Model) getFilteredProcessCount() int {
	return len(m.GetFilteredProcesses())
}

func (m *Model) getVisibleProcessRows() int {
	rows := max(m.UI.Height-data.ReservedContentRows, 3)
	return rows
}

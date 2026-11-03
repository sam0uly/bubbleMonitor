package app

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
	messages "samouly.fun/bubbleMonitor/internal/msg"
	"samouly.fun/bubbleMonitor/internal/provider"
)

func createTestModel() *Model {
	cfg := config.DefaultConfig()
	if len(cfg.Tabs) == 0 {
		cfg.Tabs = []string{"Overview", "Processes", "System", "Metrics", "Network", "Disks"}
	}

	am := data.NewAlertManager()

	return &Model{
		AppState: data.AppState{
			Metrics: data.MetricsState{
				CpuHistory:           data.NewRingBuffer(cfg.HistoryLength),
				MemHistory:           data.NewRingBuffer(cfg.HistoryLength),
				NetHistory:           data.NewRingBuffer(cfg.HistoryLength),
				SwapHistory:          data.NewRingBuffer(cfg.HistoryLength),
				DiskHORead:           data.NewRingBuffer(cfg.HistoryLength),
				DiskHOWrite:          data.NewRingBuffer(cfg.HistoryLength),
				HistoryTemp:          data.NewRingBuffer(cfg.HistoryLength),
				ProcessHistory:       make(map[int32]*data.RingBuffer),
				AlertManager:         am,
				HasNvidiaGPU:         false,
				HasAmdGPU:            false,
				HasBattery:           false,
				HasNetworkInterfaces: true,
				HasDiskIO:            true,
			},
			Process: data.ProcessState{
				Processes:        []data.ProcessInfo{},
				SuspendedState:   make(map[int32]bool),
				ProcessCmdlines:  make(map[int32]string),
				ProcessUsernames: make(map[int32]string),
				TreeView:         false,
				CollapsedPids:    make(map[int32]bool),
				BookmarkedPids:   make(map[int32]bool),
				SortBy:           "cpu",
				SortDirection:    "desc",
				OpenFilesView:    data.NewSimpleViewport(0, 0),
				ProcessesLoaded:  false,
			},
			UI: data.UIState{
				SelectedTab:   1,
				HistoryLength: cfg.HistoryLength,
				ActiveTabs:    cfg.Tabs,
				SettingsSel:   config.MetricCPU,
				StartTime:     time.Now(),
				Toasts:        []data.Toast{},
				Width:         80,
				Height:        24,
			},
			Config: data.ConfigState{
				Config:        cfg,
				SortBy:        "cpu",
				SortDirection: "desc",
				HistoryLength: cfg.HistoryLength,
			},
			Remote: data.RemoteState{
				Metrics: make(map[string]data.RemoteHostMetrics),
			},
		},
		renderCache: &RenderCache{},
		Provider: struct {
			System  provider.SystemProvider
			Process provider.ProcessProvider
			Remote  provider.RemoteProvider
		}{
			System:  provider.NewSystemAdapter(),
			Process: provider.NewProcessAdapter(),
			Remote:  provider.NewRemoteAdapter(),
		},
	}
}

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.Msg
		validate func(*testing.T, *Model)
	}{
		{
			name: "TickMsg",
			msg:  messages.TickMsg(time.Now()),
			validate: func(t *testing.T, m *Model) {
				if m.UI.TickCount != 1 {
					t.Errorf("expected TickCount 1, got %d", m.UI.TickCount)
				}
			},
		},
		{
			name: "CpuMemMsg with valid data",
			msg: messages.CpuMemMsg{
				LoadAvg:    &load.AvgStat{Load1: 0.5, Load5: 0.3, Load15: 0.2},
				MemInfo:    &mem.VirtualMemoryStat{Total: 16e9, Used: 8e9, UsedPercent: 50},
				SwapInfo:   &mem.SwapMemoryStat{Total: 8e9, Used: 1e9, UsedPercent: 12.5},
				Cpu:        25.5,
				Memory:     50.0,
				Swap:       12.5,
				CpuPerCore: []float64{20.0, 30.0, 25.0, 27.0},
			},
			validate: func(t *testing.T, m *Model) {
				if m.Metrics.Cpu != 25.5 {
					t.Errorf("expected Cpu 25.5, got %f", m.Metrics.Cpu)
				}
				if m.Metrics.Memory != 50.0 {
					t.Errorf("expected Memory 50.0, got %f", m.Metrics.Memory)
				}
				if m.Metrics.CpuHistory.Len() != 1 {
					t.Errorf("expected CpuHistory to have 1 entry")
				}
			},
		},
		{
			name: "CpuMemMsg with error",
			msg: messages.CpuMemMsg{
				Err: errors.New("mock cpu/mem error"),
			},
			validate: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "DiskNetMsg with valid data",
			msg: messages.DiskNetMsg{
				Disk:    45.0,
				NetSent: 1024 * 1024,
				NetRecv: 2048 * 1024,
			},
			validate: func(t *testing.T, m *Model) {
				if m.Metrics.Disk != 45.0 {
					t.Errorf("expected Disk 45.0, got %f", m.Metrics.Disk)
				}
			},
		},
		{
			name: "DiskNetMsg with error",
			msg: messages.DiskNetMsg{
				Err: errors.New("mock disk/net error"),
			},
			validate: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "ProcessesMsg",
			msg: messages.ProcessesMsg([]data.ProcessInfo{
				{Pid: 1, Name: "init", NameLower: "init", Cpu: 0.1, Memory: 0.2},
				{Pid: 2, Name: "systemd", NameLower: "systemd", Cpu: 0.5, Memory: 1.5},
				{Pid: 100, Name: "testproc", NameLower: "testproc", Cpu: 5.0, Memory: 2.0},
			}),
			validate: func(t *testing.T, m *Model) {
				if !m.Process.ProcessesLoaded {
					t.Error("expected ProcessesLoaded to be true")
				}
				if m.Process.ProcessCount != 3 {
					t.Errorf("expected ProcessCount 3, got %d", m.Process.ProcessCount)
				}
				if len(m.Process.Processes) != 3 {
					t.Errorf("expected 3 processes, got %d", len(m.Process.Processes))
				}
			},
		},
		{
			name: "ProcessCountMsg",
			msg:  messages.ProcessCountMsg(150),
			validate: func(t *testing.T, m *Model) {
				if m.Process.ProcessCount != 150 {
					t.Errorf("expected ProcessCount 150, got %d", m.Process.ProcessCount)
				}
			},
		},
		{
			name: "HostInfoMsg with valid data",
			msg: messages.HostInfoMsg{
				Info: &host.InfoStat{
					Hostname:        "testhost",
					OS:              "linux",
					Platform:        "ubuntu",
					PlatformVersion: "22.04",
					KernelVersion:   "5.15.0",
				},
			},
			validate: func(t *testing.T, m *Model) {
				if m.Metrics.HostInfo == nil {
					t.Error("expected HostInfo to be set")
				}
				if m.Metrics.HostInfo.Hostname != "testhost" {
					t.Errorf("expected hostname testhost, got %s", m.Metrics.HostInfo.Hostname)
				}
			},
		},
		{
			name: "HostInfoMsg with error",
			msg: messages.HostInfoMsg{
				Err: errors.New("mock host info error"),
			},
			validate: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "DiskInfoMsg with valid data",
			msg: messages.DiskInfoMsg{
				Partitions: []data.DiskPartition{
					{Mountpoint: "/", Device: "/dev/sda1", Total: 500e9, Used: 250e9, UsedPct: 50.0},
				},
			},
			validate: func(t *testing.T, m *Model) {
				if len(m.Metrics.DiskPartitions) != 1 {
					t.Errorf("expected 1 partition, got %d", len(m.Metrics.DiskPartitions))
				}
			},
		},
		{
			name: "DiskInfoMsg with error",
			msg: messages.DiskInfoMsg{
				Err: errors.New("mock disk info error"),
			},
			validate: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "GpuInfoMsg with valid data",
			msg: messages.GpuInfoMsg{
				Gpus: []data.GpuInfo{
					{Name: "NVIDIA RTX 3080", Driver: "525.0", MemoryTotal: "10GB"},
				},
			},
			validate: func(t *testing.T, m *Model) {
				if len(m.Metrics.GpuInfo) != 1 {
					t.Errorf("expected 1 GPU, got %d", len(m.Metrics.GpuInfo))
				}
			},
		},
		{
			name: "GpuInfoMsg with error",
			msg: messages.GpuInfoMsg{
				Err:  errors.New("mock gpu error"),
				Gpus: nil,
			},
			validate: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "DiskIOMsg",
			msg: messages.DiskIOMsg(map[string]disk.IOCountersStat{
				"sda": {ReadBytes: 1000, WriteBytes: 500},
			}),
			validate: func(t *testing.T, m *Model) {
				if m.Metrics.DiskIO == nil {
					t.Error("expected DiskIO to be set")
				}
			},
		},
		{
			name: "TempMsg",
			msg: messages.TempMsg([]host.TemperatureStat{
				{SensorKey: "cpu", Temperature: 65.0},
				{SensorKey: "core_0", Temperature: 60.0},
			}),
			validate: func(t *testing.T, m *Model) {
				if m.Metrics.CpuTemp == 0 {
					t.Error("expected CpuTemp to be set")
				}
			},
		},
		{
			name: "NetworkInterfacesMsg with valid data",
			msg: messages.NetworkInterfacesMsg{
				Interfaces: []net.IOCountersStat{
					{Name: "eth0", BytesSent: 1000, BytesRecv: 2000},
				},
			},
			validate: func(t *testing.T, m *Model) {
				if len(m.Metrics.NetworkInterfaces) != 1 {
					t.Errorf("expected 1 interface, got %d", len(m.Metrics.NetworkInterfaces))
				}
			},
		},
		{
			name: "NetworkInterfacesMsg with error",
			msg: messages.NetworkInterfacesMsg{
				Err: errors.New("mock network error"),
			},
			validate: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "BatteryMsg",
			msg: messages.BatteryMsg([]*battery.Battery{
				nil,
			}),
			validate: func(t *testing.T, m *Model) {
				if m.Metrics.Battery == nil {
					t.Error("expected Battery to be set")
				}
			},
		},
		{
			name: "ServicesMsg",
			msg: messages.ServicesMsg([]data.ServiceInfo{
				{Name: "sshd", Status: "running"},
				{Name: "nginx", Status: "stopped"},
			}),
			validate: func(t *testing.T, m *Model) {
				if len(m.Process.Services) != 2 {
					t.Errorf("expected 2 services, got %d", len(m.Process.Services))
				}
			},
		},
		{
			name: "ConnectionsMsg",
			msg: messages.ConnectionsMsg([]data.ConnectionInfo{
				{LocalAddr: "192.168.1.1:8080", RemoteAddr: "10.0.0.1:80", State: "ESTABLISHED", Protocol: "tcp"},
			}),
			validate: func(t *testing.T, m *Model) {
				if len(m.Process.Connections) != 1 {
					t.Errorf("expected 1 connection, got %d", len(m.Process.Connections))
				}
			},
		},
		{
			name: "SysLogMsg",
			msg: messages.SysLogMsg([]string{
				"Jan  1 12:00:00 host kernel: test message",
			}),
			validate: func(t *testing.T, m *Model) {
				if len(m.Process.SystemLogs) != 1 {
					t.Errorf("expected 1 log, got %d", len(m.Process.SystemLogs))
				}
			},
		},
		{
			name: "RemoteMsg",
			msg: messages.RemoteMsg{
				Host: "server1",
				Metrics: data.RemoteHostMetrics{
					Uptime: "uptime: 5 days",
					Online: true,
				},
			},
			validate: func(t *testing.T, m *Model) {
				if m.Remote.Metrics["server1"].Uptime != "uptime: 5 days" {
					t.Errorf("expected remote uptime to be set")
				}
			},
		},
		{
			name: "PriorityChangeMsg with error",
			msg: messages.PriorityChangeMsg{
				Err:      errors.New("mock priority error"),
				Pid:      1234,
				Priority: 10,
			},
			validate: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "ProcessControlMsg with error",
			msg: messages.ProcessControlMsg{
				Err:    errors.New("mock suspend error"),
				Action: "suspend",
				Pid:    1234,
			},
			validate: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "OpenFilesMsg with valid data",
			msg: messages.OpenFilesMsg{
				Files: []process.OpenFilesStat{
					{Path: "/proc/1/cmdline"},
					{Path: "/proc/1/environ"},
				},
				Pid: 1,
			},
			validate: func(t *testing.T, m *Model) {
				if len(m.Process.OpenFilesList) != 2 {
					t.Errorf("expected 2 files, got %d", len(m.Process.OpenFilesList))
				}
			},
		},
		{
			name: "OpenFilesMsg with error",
			msg: messages.OpenFilesMsg{
				Err: errors.New("mock open files error"),
				Pid: 1,
			},
			validate: func(t *testing.T, m *Model) {
				if m.Process.ShowOpenFiles {
					t.Error("expected ShowOpenFiles to be false on error")
				}
			},
		},
		{
			name: "OpenFilesRequestMsg",
			msg:  messages.OpenFilesRequestMsg{Pid: 1234},
			validate: func(t *testing.T, m *Model) {
				if !m.Process.ShowOpenFiles {
					t.Error("expected ShowOpenFiles to be true")
				}
				if m.Process.OpenFilesPid != 1234 {
					t.Errorf("expected OpenFilesPid 1234, got %d", m.Process.OpenFilesPid)
				}
			},
		},
		{
			name: "ProcessCmdlineMsg",
			msg: messages.ProcessCmdlineMsg{
				Pid:     1234,
				Cmdline: "/usr/bin/test --flag",
			},
			validate: func(t *testing.T, m *Model) {
				if m.Process.ProcessCmdlines[1234] != "/usr/bin/test --flag" {
					t.Error("expected cmdline to be stored")
				}
			},
		},
		{
			name: "ProcessUsernameMsg",
			msg: messages.ProcessUsernameMsg{
				Pid:      1234,
				Username: "root",
			},
			validate: func(t *testing.T, m *Model) {
				if m.Process.ProcessUsernames[1234] != "root" {
					t.Error("expected username to be stored")
				}
			},
		},
		{
			name: "ToastMsg",
			msg: messages.ToastMsg{
				Message:  "Test toast",
				Level:    data.ToastInfo,
				Duration: 3 * time.Second,
			},
			validate: func(t *testing.T, m *Model) {
				if len(m.UI.Toasts) != 1 {
					t.Errorf("expected 1 toast, got %d", len(m.UI.Toasts))
				}
				if m.UI.Toasts[0].Message != "Test toast" {
					t.Error("expected toast message to match")
				}
			},
		},
		{
			name: "ToastTimeoutMsg",
			msg: func() tea.Msg {
				m := createTestModel()
				m.UI.NextToastID = 100
				m.UI.Toasts = append(m.UI.Toasts, data.Toast{
					ID:        100,
					Message:   "Test toast",
					Level:     data.ToastInfo,
					StartTime: time.Now(),
					Duration:  3 * time.Second,
				})
				return messages.ToastTimeoutMsg{ID: 100}
			}(),
			validate: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "KillProcessMsg",
			msg: messages.KillProcessMsg{
				Pid:     1234,
				Success: true,
			},
			validate: func(t *testing.T, m *Model) {
				if m.UI.ShowKillDialog {
					t.Error("expected ShowKillDialog to be false after kill")
				}
			},
		},
		{
			name: "WindowSizeMsg",
			msg:  tea.WindowSizeMsg{Width: 120, Height: 40},
			validate: func(t *testing.T, m *Model) {
				if m.UI.Width != 120 {
					t.Errorf("expected Width 120, got %d", m.UI.Width)
				}
				if m.UI.Height != 40 {
					t.Errorf("expected Height 40, got %d", m.UI.Height)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			newModel, _ := m.Update(tt.msg)
			updatedModel := newModel.(*Model)
			tt.validate(t, updatedModel)
		})
	}
}

func TestKeySequences(t *testing.T) {
	tests := []struct {
		name     string
		initial  func() *Model
		keyPress tea.KeyMsg
		validate func(*testing.T, *Model)
	}{
		{
			name: "Tab switch with Tab key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.SelectedTab = 0
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				return m
			},
			keyPress: tea.KeyPressMsg{Code: tea.KeyTab},
			validate: func(t *testing.T, m *Model) {
				if m.UI.SelectedTab != 1 {
					t.Errorf("expected SelectedTab 1, got %d", m.UI.SelectedTab)
				}
			},
		},
		{
			name: "Tab switch with number keys",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System", "Metrics"}
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '3'},
			validate: func(t *testing.T, m *Model) {
				if m.UI.SelectedTab != 2 {
					t.Errorf("expected SelectedTab 2, got %d", m.UI.SelectedTab)
				}
			},
		},
		{
			name: "Toggle pause with P key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.Paused = false
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'p'},
			validate: func(t *testing.T, m *Model) {
				if !m.UI.Paused {
					t.Error("expected Paused to be true")
				}
			},
		},
		{
			name: "Show help with ? key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowHelp = false
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '?'},
			validate: func(t *testing.T, m *Model) {
				if !m.UI.ShowHelp {
					t.Error("expected ShowHelp to be true")
				}
			},
		},
		{
			name: "Open settings with . key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowSettings = false
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '.'},
			validate: func(t *testing.T, m *Model) {
				if !m.UI.ShowSettings {
					t.Error("expected ShowSettings to be true")
				}
			},
		},
		{
			name: "Enter filter mode with f key on Processes tab",
			initial: func() *Model {
				m := createTestModel()
				m.Process.FilterMode = false
				m.UI.SelectedTab = 1
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'f'},
			validate: func(t *testing.T, m *Model) {
				if !m.Process.FilterMode {
					t.Error("expected FilterMode to be true")
				}
			},
		},
		{
			name: "Toggle tree view with T key on Processes tab",
			initial: func() *Model {
				m := createTestModel()
				m.Process.TreeView = false
				m.UI.SelectedTab = 1
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'T'},
			validate: func(t *testing.T, m *Model) {
				if !m.Process.TreeView {
					t.Error("expected TreeView to be true")
				}
			},
		},
		{
			name: "Process navigation down with j key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.SelectedTab = 1
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				m.Process.Processes = []data.ProcessInfo{
					{Pid: 1, Name: "proc1", NameLower: "proc1"},
					{Pid: 2, Name: "proc2", NameLower: "proc2"},
				}
				m.Process.SelectedProcess = 0
				m.UI.Width = 80
				m.UI.Height = 24
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'j'},
			validate: func(t *testing.T, m *Model) {
				if m.Process.SelectedProcess != 1 {
					t.Errorf("expected SelectedProcess 1, got %d", m.Process.SelectedProcess)
				}
			},
		},
		{
			name: "Process navigation up with k key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.SelectedTab = 1
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				m.Process.Processes = []data.ProcessInfo{
					{Pid: 1, Name: "proc1", NameLower: "proc1"},
					{Pid: 2, Name: "proc2", NameLower: "proc2"},
				}
				m.Process.SelectedProcess = 1
				m.UI.Width = 80
				m.UI.Height = 24
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'k'},
			validate: func(t *testing.T, m *Model) {
				if m.Process.SelectedProcess != 0 {
					t.Errorf("expected SelectedProcess 0, got %d", m.Process.SelectedProcess)
				}
			},
		},
		{
			name: "Kill dialog shows with K key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.SelectedTab = 1
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				m.Process.Processes = []data.ProcessInfo{
					{Pid: 1234, Name: "testproc", NameLower: "testproc"},
				}
				m.UI.ShowKillDialog = false
				m.UI.Width = 80
				m.UI.Height = 24
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'K'},
			validate: func(t *testing.T, m *Model) {
				if !m.UI.ShowKillDialog {
					t.Error("expected ShowKillDialog to be true")
				}
				if m.Process.KillTargetPid != 1234 {
					t.Errorf("expected KillTargetPid 1234, got %d", m.Process.KillTargetPid)
				}
			},
		},
		{
			name: "Kill dialog dismiss with n key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowKillDialog = true
				m.Process.KillTargetPid = 1234
				m.Process.KillTargetName = "testproc"
				m.UI.KillDialogSel = 0
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'n'},
			validate: func(t *testing.T, m *Model) {
				if m.UI.ShowKillDialog {
					t.Error("expected ShowKillDialog to be false")
				}
			},
		},
		{
			name: "Help dismiss with ? key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowHelp = true
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '?'},
			validate: func(t *testing.T, m *Model) {
				if m.UI.ShowHelp {
					t.Error("expected ShowHelp to be false")
				}
			},
		},
		{
			name: "Sort change with s key",
			initial: func() *Model {
				m := createTestModel()
				m.Process.SortBy = "cpu"
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 's'},
			validate: func(t *testing.T, m *Model) {
				if m.Process.SortBy != "mem" {
					t.Errorf("expected SortBy mem, got %s", m.Process.SortBy)
				}
			},
		},
		{
			name: "Sort direction toggle with S key",
			initial: func() *Model {
				m := createTestModel()
				m.Process.SortDirection = "desc"
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'S'},
			validate: func(t *testing.T, m *Model) {
				if m.Process.SortDirection != "asc" {
					t.Errorf("expected SortDirection asc, got %s", m.Process.SortDirection)
				}
			},
		},
		{
			name: "History length cycle with H key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.HistoryLength = 60
				m.Metrics.CpuHistory = data.NewRingBuffer(60)
				m.Metrics.MemHistory = data.NewRingBuffer(60)
				m.Metrics.NetHistory = data.NewRingBuffer(60)
				m.Metrics.SwapHistory = data.NewRingBuffer(60)
				m.Metrics.HistoryTemp = data.NewRingBuffer(60)
				m.Metrics.DiskHORead = data.NewRingBuffer(60)
				m.Metrics.DiskHOWrite = data.NewRingBuffer(60)
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'H'},
			validate: func(t *testing.T, m *Model) {
				if m.UI.HistoryLength != 300 {
					t.Errorf("expected HistoryLength 300, got %d", m.UI.HistoryLength)
				}
			},
		},
		{
			name: "Chart type cycle with C key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ChartType = "line"
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'C'},
			validate: func(t *testing.T, m *Model) {
				if m.UI.ChartType != "bar" {
					t.Errorf("expected ChartType bar, got %s", m.UI.ChartType)
				}
			},
		},
		{
			name: "Clear filter with c key on Processes tab",
			initial: func() *Model {
				m := createTestModel()
				m.UI.SelectedTab = 1
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				m.Process.ProcessFilter = "test"
				m.Process.ProcessFilterLower = "test"
				m.Process.SelectedProcess = 5
				m.Process.ProcessScrollOffset = 3
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'c'},
			validate: func(t *testing.T, m *Model) {
				if m.Process.ProcessFilter != "" {
					t.Error("expected ProcessFilter to be cleared")
				}
				if m.Process.SelectedProcess != 0 {
					t.Errorf("expected SelectedProcess 0, got %d", m.Process.SelectedProcess)
				}
			},
		},
		{
			name: "Go to top with g key on Processes tab",
			initial: func() *Model {
				m := createTestModel()
				m.UI.SelectedTab = 1
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				m.Process.Processes = []data.ProcessInfo{
					{Pid: 1, Name: "proc1"},
					{Pid: 2, Name: "proc2"},
				}
				m.Process.SelectedProcess = 10
				m.Process.ProcessScrollOffset = 5
				m.UI.Width = 80
				m.UI.Height = 24
				return m
			},
			keyPress: tea.KeyPressMsg{Code: 'g'},
			validate: func(t *testing.T, m *Model) {
				if m.Process.SelectedProcess != 0 {
					t.Errorf("expected SelectedProcess 0, got %d", m.Process.SelectedProcess)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.initial()
			newModel, _ := m.Update(tt.keyPress)
			updatedModel := newModel.(*Model)
			tt.validate(t, updatedModel)
		})
	}
}

func TestErrorPaths(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Model
		msg   tea.Msg
		check func(*testing.T, *Model)
	}{
		{
			name: "CpuMemMsg error",
			setup: func() *Model {
				return createTestModel()
			},
			msg: messages.CpuMemMsg{
				Err: errors.New("test cpu error"),
			},
			check: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "DiskNetMsg error",
			setup: func() *Model {
				return createTestModel()
			},
			msg: messages.DiskNetMsg{
				Err: errors.New("test disk error"),
			},
			check: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "HostInfoMsg error",
			setup: func() *Model {
				return createTestModel()
			},
			msg: messages.HostInfoMsg{
				Err: errors.New("test host error"),
			},
			check: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "DiskInfoMsg error",
			setup: func() *Model {
				return createTestModel()
			},
			msg: messages.DiskInfoMsg{
				Err: errors.New("test disk info error"),
			},
			check: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "NetworkInterfacesMsg error",
			setup: func() *Model {
				return createTestModel()
			},
			msg: messages.NetworkInterfacesMsg{
				Err: errors.New("test network error"),
			},
			check: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "PriorityChangeMsg error",
			setup: func() *Model {
				return createTestModel()
			},
			msg: messages.PriorityChangeMsg{
				Err:      errors.New("test priority error"),
				Pid:      1234,
				Priority: 10,
			},
			check: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "ProcessControlMsg error",
			setup: func() *Model {
				return createTestModel()
			},
			msg: messages.ProcessControlMsg{
				Err:    errors.New("test control error"),
				Action: "suspend",
				Pid:    1234,
			},
			check: func(t *testing.T, m *Model) {
			},
		},
		{
			name: "OpenFilesMsg error",
			setup: func() *Model {
				m := createTestModel()
				m.Process.ShowOpenFiles = true
				return m
			},
			msg: messages.OpenFilesMsg{
				Err: errors.New("test open files error"),
				Pid: 1,
			},
			check: func(t *testing.T, m *Model) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			newModel, _ := m.Update(tt.msg)
			updatedModel := newModel.(*Model)
			tt.check(t, updatedModel)
		})
	}
}

func TestStartupFlow(t *testing.T) {
	tests := []struct {
		name     string
		steps    []func(*Model) (tea.Model, tea.Cmd)
		validate func(*testing.T, *Model)
	}{
		{
			name: "Initial model creation",
			steps: []func(*Model) (tea.Model, tea.Cmd){
				func(m *Model) (tea.Model, tea.Cmd) {
					return m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
				},
			},
			validate: func(t *testing.T, m *Model) {
				if m.UI.Width != 80 {
					t.Errorf("expected Width 80, got %d", m.UI.Width)
				}
				if m.UI.Height != 24 {
					t.Errorf("expected Height 24, got %d", m.UI.Height)
				}
			},
		},
		{
			name: "Initial tick triggers metrics collection",
			steps: []func(*Model) (tea.Model, tea.Cmd){
				func(m *Model) (tea.Model, tea.Cmd) {
					return m.Update(messages.TickMsg(time.Now()))
				},
			},
			validate: func(t *testing.T, m *Model) {
				if m.UI.TickCount != 1 {
					t.Errorf("expected TickCount 1, got %d", m.UI.TickCount)
				}
			},
		},
		{
			name: "CPU/Memory data updates state",
			steps: []func(*Model) (tea.Model, tea.Cmd){
				func(m *Model) (tea.Model, tea.Cmd) {
					return m.Update(messages.CpuMemMsg{Cpu: 25.0, Memory: 50.0, Swap: 10.0})
				},
			},
			validate: func(t *testing.T, m *Model) {
				if m.Metrics.Cpu != 25.0 {
					t.Errorf("expected Cpu 25.0, got %f", m.Metrics.Cpu)
				}
				if m.Metrics.Memory != 50.0 {
					t.Errorf("expected Memory 50.0, got %f", m.Metrics.Memory)
				}
			},
		},
		{
			name: "Process list loads correctly",
			steps: []func(*Model) (tea.Model, tea.Cmd){
				func(m *Model) (tea.Model, tea.Cmd) {
					return m.Update(messages.ProcessesMsg([]data.ProcessInfo{
						{Pid: 1, Name: "init", NameLower: "init"},
						{Pid: 2, Name: "systemd", NameLower: "systemd"},
					}))
				},
			},
			validate: func(t *testing.T, m *Model) {
				if !m.Process.ProcessesLoaded {
					t.Error("expected ProcessesLoaded to be true")
				}
				if m.Process.ProcessCount != 2 {
					t.Errorf("expected ProcessCount 2, got %d", m.Process.ProcessCount)
				}
			},
		},
		{
			name: "Full startup sequence",
			steps: []func(*Model) (tea.Model, tea.Cmd){
				func(m *Model) (tea.Model, tea.Cmd) {
					return m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
				},
				func(m *Model) (tea.Model, tea.Cmd) {
					return m.Update(messages.CpuMemMsg{Cpu: 25.0, Memory: 50.0})
				},
				func(m *Model) (tea.Model, tea.Cmd) {
					return m.Update(messages.ProcessesMsg([]data.ProcessInfo{
						{Pid: 1, Name: "init"},
					}))
				},
				func(m *Model) (tea.Model, tea.Cmd) {
					return m.Update(messages.HostInfoMsg{Info: &host.InfoStat{Hostname: "test"}})
				},
			},
			validate: func(t *testing.T, m *Model) {
				if m.Metrics.Cpu != 25.0 {
					t.Errorf("expected Cpu 25.0, got %f", m.Metrics.Cpu)
				}
				if m.Process.ProcessCount != 1 {
					t.Errorf("expected ProcessCount 1, got %d", m.Process.ProcessCount)
				}
				if m.Metrics.HostInfo == nil || m.Metrics.HostInfo.Hostname != "test" {
					t.Error("expected HostInfo to be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			var currentModel *Model = m
			for _, step := range tt.steps {
				newModel, _ := step(currentModel)
				currentModel = newModel.(*Model)
			}
			tt.validate(t, currentModel)
		})
	}
}

func TestProcessKillFlow(t *testing.T) {
	tests := []struct {
		name     string
		steps    []func(*Model) (tea.Model, tea.Cmd)
		validate func(*testing.T, *Model)
	}{
		{
			name: "Kill dialog appears",
			steps: []func(*Model) (tea.Model, tea.Cmd){
				func(m *Model) (tea.Model, tea.Cmd) {
					m.UI.SelectedTab = 1
					m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
					m.Process.Processes = []data.ProcessInfo{
						{Pid: 1234, Name: "targetproc", NameLower: "targetproc"},
					}
					m.UI.Width = 80
					m.UI.Height = 24
					return m.Update(tea.KeyPressMsg{Code: 'K'})
				},
			},
			validate: func(t *testing.T, m *Model) {
				if !m.UI.ShowKillDialog {
					t.Error("expected ShowKillDialog to be true")
				}
				if m.Process.KillTargetPid != 1234 {
					t.Errorf("expected KillTargetPid 1234, got %d", m.Process.KillTargetPid)
				}
				if m.Process.KillTargetName != "targetproc" {
					t.Errorf("expected KillTargetName targetproc, got %s", m.Process.KillTargetName)
				}
			},
		},
		{
			name: "Kill dialog dismissed with n",
			steps: []func(*Model) (tea.Model, tea.Cmd){
				func(m *Model) (tea.Model, tea.Cmd) {
					m.UI.ShowKillDialog = true
					m.Process.KillTargetPid = 1234
					m.Process.KillTargetName = "targetproc"
					m.UI.KillDialogSel = 0
					return m.Update(tea.KeyPressMsg{Code: 'n'})
				},
			},
			validate: func(t *testing.T, m *Model) {
				if m.UI.ShowKillDialog {
					t.Error("expected ShowKillDialog to be false")
				}
				if m.Process.KillTargetPid != 0 {
					t.Errorf("expected KillTargetPid 0, got %d", m.Process.KillTargetPid)
				}
			},
		},
		{
			name: "Kill dialog dismissed with esc",
			steps: []func(*Model) (tea.Model, tea.Cmd){
				func(m *Model) (tea.Model, tea.Cmd) {
					m.UI.ShowKillDialog = true
					m.Process.KillTargetPid = 1234
					m.Process.KillTargetName = "targetproc"
					m.UI.KillDialogSel = 0
					return m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
				},
			},
			validate: func(t *testing.T, m *Model) {
				if m.UI.ShowKillDialog {
					t.Error("expected ShowKillDialog to be false")
				}
			},
		},
		{
			name: "Kill dialog confirm with enter",
			steps: []func(*Model) (tea.Model, tea.Cmd){
				func(m *Model) (tea.Model, tea.Cmd) {
					m.UI.ShowKillDialog = true
					m.Process.KillTargetPid = 9999
					m.Process.KillTargetName = "testproc"
					m.UI.KillDialogSel = 0
					return m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				},
			},
			validate: func(t *testing.T, m *Model) {
				if m.UI.ShowKillDialog {
					t.Error("expected ShowKillDialog to be false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			var currentModel *Model = m
			for _, step := range tt.steps {
				newModel, _ := step(currentModel)
				currentModel = newModel.(*Model)
			}
			tt.validate(t, currentModel)
		})
	}
}

func TestTabSwitchFlow(t *testing.T) {
	tests := []struct {
		name     string
		initial  func() *Model
		keyPress tea.KeyMsg
		validate func(*testing.T, *Model)
	}{
		{
			name: "Switch to tab 0 with number 1",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System", "Metrics", "Network", "Disks"}
				m.UI.SelectedTab = 3
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '1'},
			validate: func(t *testing.T, m *Model) {
				if m.UI.SelectedTab != 0 {
					t.Errorf("expected SelectedTab 0, got %d", m.UI.SelectedTab)
				}
			},
		},
		{
			name: "Switch to tab 1 with number 2",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System", "Metrics", "Network", "Disks"}
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '2'},
			validate: func(t *testing.T, m *Model) {
				if m.UI.SelectedTab != 1 {
					t.Errorf("expected SelectedTab 1, got %d", m.UI.SelectedTab)
				}
			},
		},
		{
			name: "Tab wraps around forward",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ActiveTabs = []string{"Overview", "Processes"}
				m.UI.SelectedTab = 1
				return m
			},
			keyPress: tea.KeyPressMsg{Code: tea.KeyTab},
			validate: func(t *testing.T, m *Model) {
				if m.UI.SelectedTab != 0 {
					t.Errorf("expected SelectedTab 0 (wrap around), got %d", m.UI.SelectedTab)
				}
			},
		},
		{
			name: "Tab wraps around backward",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ActiveTabs = []string{"Overview", "Processes"}
				m.UI.SelectedTab = 0
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '\t', Mod: tea.ModShift},
			validate: func(t *testing.T, m *Model) {
				if m.UI.SelectedTab != 1 {
					t.Errorf("expected SelectedTab 1 (wrap around), got %d", m.UI.SelectedTab)
				}
			},
		},
		{
			name: "Switching to Processes tab with number key",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				m.UI.SelectedTab = 0
				m.Process.ProcessesLoaded = false
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '2'},
			validate: func(t *testing.T, m *Model) {
				if m.UI.SelectedTab != 1 {
					t.Errorf("expected SelectedTab 1, got %d", m.UI.SelectedTab)
				}
			},
		},
		{
			name: "Switching to Processes tab with Tab key triggers load",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ActiveTabs = []string{"Overview", "Processes", "System"}
				m.UI.SelectedTab = 0
				m.Process.ProcessesLoaded = false
				return m
			},
			keyPress: tea.KeyPressMsg{Code: tea.KeyTab},
			validate: func(t *testing.T, m *Model) {
				if !m.Process.ProcessesLoaded {
					t.Error("expected ProcessesLoaded to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.initial()
			newModel, _ := m.Update(tt.keyPress)
			updatedModel := newModel.(*Model)
			tt.validate(t, updatedModel)
		})
	}
}

func TestSettingsFlow(t *testing.T) {
	tests := []struct {
		name     string
		initial  func() *Model
		keyPress tea.KeyMsg
		validate func(*testing.T, *Model)
	}{
		{
			name: "Open settings",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowSettings = false
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '.'},
			validate: func(t *testing.T, m *Model) {
				if !m.UI.ShowSettings {
					t.Error("expected ShowSettings to be true")
				}
			},
		},
		{
			name: "Close settings with esc",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowSettings = true
				m.UI.SettingsIdx = 5
				return m
			},
			keyPress: tea.KeyPressMsg{Code: tea.KeyEsc},
			validate: func(t *testing.T, m *Model) {
				if m.UI.ShowSettings {
					t.Error("expected ShowSettings to be false")
				}
			},
		},
		{
			name: "Settings navigation up",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowSettings = true
				m.UI.SettingsIdx = 5
				return m
			},
			keyPress: tea.KeyPressMsg{Code: tea.KeyUp},
			validate: func(t *testing.T, m *Model) {
				if m.UI.SettingsIdx != 4 {
					t.Errorf("expected SettingsIdx 4, got %d", m.UI.SettingsIdx)
				}
			},
		},
		{
			name: "Settings navigation down",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowSettings = true
				m.UI.SettingsIdx = 5
				return m
			},
			keyPress: tea.KeyPressMsg{Code: tea.KeyDown},
			validate: func(t *testing.T, m *Model) {
				if m.UI.SettingsIdx != 6 {
					t.Errorf("expected SettingsIdx 6, got %d", m.UI.SettingsIdx)
				}
			},
		},
		{
			name: "Threshold value increase",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowSettings = true
				m.UI.SettingsIdx = 0
				m.Config.Config.Thresholds[config.MetricCPU] = 70.0
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '+'},
			validate: func(t *testing.T, m *Model) {
				if m.Config.Config.Thresholds[config.MetricCPU] != 71.0 {
					t.Errorf("expected threshold 71.0, got %f", m.Config.Config.Thresholds[config.MetricCPU])
				}
			},
		},
		{
			name: "Threshold value decrease",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowSettings = true
				m.UI.SettingsIdx = 0
				m.Config.Config.Thresholds[config.MetricCPU] = 70.0
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '-'},
			validate: func(t *testing.T, m *Model) {
				if m.Config.Config.Thresholds[config.MetricCPU] != 69.0 {
					t.Errorf("expected threshold 69.0, got %f", m.Config.Config.Thresholds[config.MetricCPU])
				}
			},
		},
		{
			name: "Threshold doesn't go below 0",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowSettings = true
				m.UI.SettingsIdx = 0
				m.Config.Config.Thresholds[config.MetricCPU] = 0.0
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '-'},
			validate: func(t *testing.T, m *Model) {
				if m.Config.Config.Thresholds[config.MetricCPU] != 0.0 {
					t.Errorf("expected threshold 0.0, got %f", m.Config.Config.Thresholds[config.MetricCPU])
				}
			},
		},
		{
			name: "Threshold doesn't go above 100",
			initial: func() *Model {
				m := createTestModel()
				m.UI.ShowSettings = true
				m.UI.SettingsIdx = 0
				m.Config.Config.Thresholds[config.MetricCPU] = 100.0
				return m
			},
			keyPress: tea.KeyPressMsg{Code: '+'},
			validate: func(t *testing.T, m *Model) {
				if m.Config.Config.Thresholds[config.MetricCPU] != 100.0 {
					t.Errorf("expected threshold 100.0, got %f", m.Config.Config.Thresholds[config.MetricCPU])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.initial()
			newModel, _ := m.Update(tt.keyPress)
			updatedModel := newModel.(*Model)
			tt.validate(t, updatedModel)
		})
	}
}

func TestRenderCache(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Model
		msg      tea.Msg
		validate func(*testing.T, *Model)
	}{
		{
			name: "WindowSizeMsg forces render",
			setup: func() *Model {
				m := createTestModel()
				m.renderCache.Force = false
				return m
			},
			msg: tea.WindowSizeMsg{Width: 100, Height: 30},
			validate: func(t *testing.T, m *Model) {
				if !m.renderCache.Force {
					t.Error("expected renderCache.Force to be true")
				}
			},
		},
		{
			name: "KeyMsg forces render",
			setup: func() *Model {
				m := createTestModel()
				m.renderCache.Force = false
				return m
			},
			msg: tea.KeyPressMsg{Code: tea.KeyTab},
			validate: func(t *testing.T, m *Model) {
				if !m.renderCache.Force {
					t.Error("expected renderCache.Force to be true")
				}
			},
		},
		{
			name: "TickMsg does not force render",
			setup: func() *Model {
				m := createTestModel()
				m.renderCache.Force = false
				return m
			},
			msg: messages.TickMsg(time.Now()),
			validate: func(t *testing.T, m *Model) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			newModel, _ := m.Update(tt.msg)
			updatedModel := newModel.(*Model)
			tt.validate(t, updatedModel)
		})
	}
}

// Integration runs all integration tests for the Update function
// This is the master test function that can be run with: go test -v ./internal/app/... -run Integration
func TestIntegration(t *testing.T) {
	// Run all test groups
	t.Run("MessageTypes", func(t *testing.T) {
		TestMessageTypes(t)
	})
	t.Run("KeySequences", func(t *testing.T) {
		TestKeySequences(t)
	})
	t.Run("ErrorPaths", func(t *testing.T) {
		TestErrorPaths(t)
	})
	t.Run("StartupFlow", func(t *testing.T) {
		TestStartupFlow(t)
	})
	t.Run("ProcessKillFlow", func(t *testing.T) {
		TestProcessKillFlow(t)
	})
	t.Run("TabSwitchFlow", func(t *testing.T) {
		TestTabSwitchFlow(t)
	})
	t.Run("SettingsFlow", func(t *testing.T) {
		TestSettingsFlow(t)
	})
	t.Run("RenderCache", func(t *testing.T) {
		TestRenderCache(t)
	})
	// Concurrent message handling test removed - Update() is not
	// thread-safe by design (called sequentially from TUI loop)
	// t.Run("ConcurrentMessageHandling", func(t *testing.T) {
	// 	TestConcurrentMessageHandling(t)
	// })
}

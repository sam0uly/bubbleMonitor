package testutil

import (
	"time"

	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
)

type TestModel struct {
	data.AppState
	RenderCache *RenderCache
}

func NewTestModel() *TestModel {
	cfg := config.DefaultConfig()
	if len(cfg.Tabs) == 0 {
		cfg.Tabs = []string{"Overview", "Processes", "System", "Metrics", "Network", "Disks"}
	}

	am := data.NewAlertManager()

	return &TestModel{
		AppState: data.AppState{
			Metrics: data.MetricsState{
				CpuHistory:           data.NewRingBuffer(cfg.HistoryLength),
				MemHistory:           data.NewRingBuffer(cfg.HistoryLength),
				NetHistory:           data.NewRingBuffer(cfg.HistoryLength),
				SwapHistory:          data.NewRingBuffer(cfg.HistoryLength),
				DiskHORead:           data.NewRingBuffer(cfg.HistoryLength),
				DiskHOWrite:          data.NewRingBuffer(cfg.HistoryLength),
				HistoryTemp:          data.NewRingBuffer(cfg.HistoryLength),
				AlertManager:         am,
				HasNvidiaGPU:         false,
				HasAmdGPU:            false,
				HasBattery:           false,
				HasNetworkInterfaces: true,
				HasDiskIO:            true,
				HasServices:          false,
				HasTempSensors:       false,
				ProcessHistory:       make(map[int32]*data.RingBuffer),
				HealthScore:          100,
			},
			Process: data.ProcessState{
				Processes:               []data.ProcessInfo{},
				TreeView:                false,
				CollapsedPids:           make(map[int32]bool),
				BookmarkedPids:          make(map[int32]bool),
				SortBy:                  "cpu",
				SortDirection:           "desc",
				ProcessesLoaded:         false,
				SuspendedState:          make(map[int32]bool),
				ProcessCmdlines:         make(map[int32]string),
				ProcessUsernames:        make(map[int32]string),
				OpenFilesView:           data.NewSimpleViewport(0, 0),
				ProcessFilter:           "",
				ProcessFilterLower:      "",
				FilterMode:              false,
				SelectedProcess:         0,
				ProcessScrollOffset:     0,
				ShowProcessMenu:         false,
				ProcessMenuIdx:          0,
				ProcessMenuX:            0,
				ProcessMenuY:            0,
				Services:                []data.ServiceInfo{},
				Connections:             []data.ConnectionInfo{},
				ServicesScrollOffset:    0,
				ConnectionsScrollOffset: 0,
				SystemLogs:              []string{},
				LogsScrollOffset:        0,
				ProcessCacheDirty:       false,
				OpenFilesScrollOffset:   0,
			},
			UI: data.UIState{
				Width:                    80,
				Height:                   24,
				SelectedTab:              1,
				ActiveTabs:               cfg.Tabs,
				HistoryLength:            cfg.HistoryLength,
				Paused:                   false,
				ShowHelp:                 false,
				ShowSettings:             false,
				ShowSamLab:               false,
				SettingsEdit:             false,
				SettingsSel:              config.MetricCPU,
				SettingsIdx:              0,
				ShowKillDialog:           false,
				KillDialogSel:            0,
				Toasts:                   []data.Toast{},
				NextToastID:              1,
				LastError:                "",
				LastErrorTime:            time.Time{},
				TickCount:                0,
				ChartType:                cfg.ChartType,
				CpuCoreScrollOffset:      0,
				SystemBlockScrollOffsets: make(map[int]int),
				ActiveScrollBlock:        -1,
				SystemBlockCount:         0,
				SystemBlockScrollable:    make(map[int]bool),
				SystemBlockMaxScroll:     make(map[int]int),
				MouseX:                   0,
				MouseY:                   0,
				Zones:                    nil,
				ZoneManager:              nil,
				StartTime:                time.Now(),
				LastConfigModTime:        time.Now(),
			},
			Config: data.ConfigState{
				Theme:                cfg.Theme,
				RefreshRate:          cfg.RefreshRate,
				BorderType:           cfg.BorderType,
				BorderStyle:          cfg.BorderStyle,
				BackgroundOpaque:     cfg.BackgroundOpaque,
				ProcessCpuNormalized: cfg.ProcessCpuNormalized,
				SortBy:               cfg.SortBy,
				SortDirection:        cfg.SortDirection,
				HistoryLength:        cfg.HistoryLength,
				Config:               cfg,
				StartTime:            time.Now(),
				LastConfigModTime:    time.Now(),
			},
			Remote: data.RemoteState{
				Metrics: make(map[string]data.RemoteHostMetrics),
			},
		},
		RenderCache: &RenderCache{},
	}
}

func NewModelWithProcesses() *TestModel {
	m := NewTestModel()
	m.Process.Processes = []data.ProcessInfo{
		{
			Pid:         1,
			Name:        "init",
			NameLower:   "init",
			Status:      "running",
			Cpu:         0.1,
			Memory:      0.2,
			MemoryBytes: 1024 * 1024,
			Nice:        0,
			Ppid:        0,
			CreateTime:  time.Now().UnixMilli(),
		},
		{
			Pid:         2,
			Name:        "systemd",
			NameLower:   "systemd",
			Status:      "running",
			Cpu:         0.5,
			Memory:      1.5,
			MemoryBytes: 5 * 1024 * 1024,
			Nice:        0,
			Ppid:        1,
			CreateTime:  time.Now().UnixMilli(),
		},
		{
			Pid:         100,
			Name:        "testproc",
			NameLower:   "testproc",
			Status:      "running",
			Cpu:         5.0,
			Memory:      2.0,
			MemoryBytes: 8 * 1024 * 1024,
			Nice:        0,
			Ppid:        2,
			CreateTime:  time.Now().UnixMilli(),
		},
		{
			Pid:         101,
			Name:        "bash",
			NameLower:   "bash",
			Status:      "sleeping",
			Cpu:         0.0,
			Memory:      0.5,
			MemoryBytes: 2 * 1024 * 1024,
			Nice:        0,
			Ppid:        100,
			CreateTime:  time.Now().UnixMilli(),
		},
		{
			Pid:         102,
			Name:        "vim",
			NameLower:   "vim",
			Status:      "running",
			Cpu:         1.2,
			Memory:      1.0,
			MemoryBytes: 4 * 1024 * 1024,
			Nice:        0,
			Ppid:        101,
			CreateTime:  time.Now().UnixMilli(),
		},
	}
	m.Process.ProcessCount = len(m.Process.Processes)
	m.Process.ProcessesLoaded = true

	m.Process.ProcessesByPid = make(map[int32]data.ProcessInfo, len(m.Process.Processes))
	for _, p := range m.Process.Processes {
		m.Process.ProcessesByPid[p.Pid] = p
	}

	return m
}

func NewModelWithMetrics() *TestModel {
	m := NewTestModel()
	m.Metrics.Cpu = 25.5
	m.Metrics.Memory = 50.0
	m.Metrics.Swap = 12.5
	m.Metrics.Disk = 45.0
	m.Metrics.CpuTemp = 65.0

	m.Metrics.CpuPerCore = []float64{20.0, 30.0, 25.0, 27.0, 22.0, 28.0, 24.0, 26.0}

	for i := 0; i < 10; i++ {
		m.Metrics.CpuHistory.Push(float64(20 + i))
		m.Metrics.MemHistory.Push(float64(45 + i))
		m.Metrics.NetHistory.Push(float64(1000 + i*100))
		m.Metrics.SwapHistory.Push(float64(10 + i))
		m.Metrics.HistoryTemp.Push(float64(60 + i))
		m.Metrics.DiskHORead.Push(float64(500 + i*50))
		m.Metrics.DiskHOWrite.Push(float64(300 + i*30))
	}

	return m
}

type RenderCache struct {
	LastRenderTime time.Time
	Content        string
	Force          bool
}

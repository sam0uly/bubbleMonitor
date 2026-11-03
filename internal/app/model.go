package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	configpkg "samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/provider"
	"samouly.fun/bubbleMonitor/internal/provider/system"
	"samouly.fun/bubbleMonitor/internal/ui"
)

type RenderCache struct {
	LastRenderTime time.Time
	Content        tea.View
	Force          bool
}

type Model struct {
	data.AppState
	renderCache *RenderCache
	Provider    struct {
		System  provider.SystemProvider
		Process provider.ProcessProvider
		Remote  provider.RemoteProvider
	}
}

func (m *Model) Init() tea.Cmd {
	capabilities := system.DetectHardware()
	m.Metrics.HasNvidiaGPU = capabilities.HasNvidiaGPU
	m.Metrics.HasAmdGPU = capabilities.HasAmdGPU
	m.Metrics.HasIntelGPU = capabilities.HasIntelGPU
	m.Metrics.HasBattery = capabilities.HasBattery
	m.Metrics.HasNetworkInterfaces = capabilities.HasNetworkInterfaces
	m.Metrics.HasDiskIO = capabilities.HasDiskIO
	m.Metrics.HasServices = capabilities.HasServices
	m.Metrics.HasTempSensors = capabilities.HasTempSensors

	cmds := []tea.Cmd{
		m.Provider.System.TickCmd(time.Duration(m.Config.RefreshRate) * time.Millisecond),
		m.Provider.System.FastMetricsCmd(),
		m.Provider.System.SlowMetricsCmd(),
		m.Provider.System.HostInfoCmd(),
		m.Provider.System.DiskInfoCmd(),
	}

	if m.Metrics.HasBattery {
		cmds = append(cmds, m.Provider.System.BatteryCmd())
	}

	if m.Metrics.HasNvidiaGPU || m.Metrics.HasAmdGPU {
		cmds = append(cmds, m.Provider.System.GpuInfoCmd())
	}

	cmds = append(cmds, configpkg.WatchConfig(m.Config.LastConfigModTime))

	if m.UI.SelectedTab >= 0 && m.UI.SelectedTab < len(m.UI.ActiveTabs) {
		switch m.UI.ActiveTabs[m.UI.SelectedTab] {
		case "Processes":
			m.Process.ProcessesLoaded = true
			cmds = append(cmds, m.Provider.Process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection))
		case "System":
			m.Process.ProcessesLoaded = true
			cmds = append(cmds, m.Provider.Process.ProcessesCmd(m.Process.SortBy, m.Process.SortDirection))
			if m.Metrics.HasTempSensors {
				cmds = append(cmds, m.Provider.System.TempCmd())
			}
			cmds = append(cmds, m.Provider.System.HostInfoCmd())
			if m.Metrics.HasNvidiaGPU || m.Metrics.HasAmdGPU {
				cmds = append(cmds, m.Provider.System.GpuInfoCmd())
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
		case "Services":
			if m.Metrics.HasServices {
				cmds = append(cmds, m.Provider.System.ServicesCmd())
			}
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() tea.View {
	// Throttle rendering to ~30 FPS (33ms)
	if !m.renderCache.Force && time.Since(m.renderCache.LastRenderTime) < 33*time.Millisecond {
		return m.renderCache.Content
	}

	content := ui.RenderFromAppState(&m.AppState)
	m.renderCache.Content = content
	m.renderCache.LastRenderTime = time.Now()
	m.renderCache.Force = false
	return content
}

func (m *Model) GetBorder() lipgloss.Border {
	switch m.Config.BorderStyle {
	case "double":
		return lipgloss.DoubleBorder()
	case "dashed":
		return lipgloss.Border{
			Top:         "-",
			Bottom:      "-",
			Left:        "|",
			Right:       "|",
			TopLeft:     "+",
			TopRight:    "+",
			BottomLeft:  "+",
			BottomRight: "+",
		}
	default: // single
		if m.Config.BorderType == "rounded" {
			return lipgloss.RoundedBorder()
		}
		return lipgloss.NormalBorder()
	}
}

func InitialModelWithConfig(cfg configpkg.AppConfig) *Model {
	// Ensure tabs are populated immediately - fallback to defaults if empty
	if len(cfg.Tabs) == 0 {
		cfg.Tabs = configpkg.DefaultConfig().Tabs
	}

	am := data.NewAlertManager()

	cpuInfo, _ := cpu.Info()

	selectedTab := 1
	if cfg.DefaultTab != "" {
		for i, tab := range cfg.Tabs {
			if tab == cfg.DefaultTab {
				selectedTab = i
				break
			}
		}
	}

	return &Model{
		AppState: data.AppState{
			Metrics: data.MetricsState{
				CpuHistory:     data.NewRingBuffer(cfg.HistoryLength),
				MemHistory:     data.NewRingBuffer(cfg.HistoryLength),
				NetHistory:     data.NewRingBuffer(cfg.HistoryLength),
				SwapHistory:    data.NewRingBuffer(cfg.HistoryLength),
				DiskHORead:     data.NewRingBuffer(cfg.HistoryLength),
				DiskHOWrite:    data.NewRingBuffer(cfg.HistoryLength),
				HistoryTemp:    data.NewRingBuffer(cfg.HistoryLength),
				AlertManager:   am,
				CpuInfoStatic:  cpuInfo, // Static CPU info fetched once
				ProcessHistory: make(map[int32]*data.RingBuffer),
				HealthScore:    100,
			},
			Process: data.ProcessState{
				Processes:        []data.ProcessInfo{},
				SuspendedState:   make(map[int32]bool),
				ProcessCmdlines:  make(map[int32]string),
				ProcessUsernames: make(map[int32]string),
				TreeView:         cfg.ViewType == "tree",
				CollapsedPids:    make(map[int32]bool),
				BookmarkedPids:   make(map[int32]bool),
				SortBy:           cfg.SortBy,
				SortDirection:    cfg.SortDirection,
				OpenFilesView:    data.NewSimpleViewport(0, 0),
			},
			UI: data.UIState{
				SelectedTab:              selectedTab,
				HistoryLength:            cfg.HistoryLength,
				ChartType:                cfg.ChartType,
				ActiveTabs:               cfg.Tabs,
				StartTime:                time.Now(),
				Toasts:                   []data.Toast{},
				SettingsSel:              configpkg.MetricCPU,
				LastConfigModTime:        time.Now(),
				SystemBlockScrollOffsets: make(map[int]int),
				ActiveScrollBlock:        -1,
				SystemBlockCount:         0,
				SystemBlockScrollable:    make(map[int]bool),
				SystemBlockMaxScroll:     make(map[int]int),
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
				LastConfigModTime:    time.Now(),
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

func InitialModel() *Model {
	cfg, err := configpkg.LoadConfig()
	if err != nil {
		cfg = configpkg.DefaultConfig()
	}
	return InitialModelWithConfig(cfg)
}

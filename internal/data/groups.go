package data

import (
	"strings"
	"time"

	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/ui/input"
	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// MetricsState holds all system metrics and history data.
type MetricsState struct {
	// History buffers
	CpuHistory  *RingBuffer
	MemHistory  *RingBuffer
	NetHistory  *RingBuffer
	SwapHistory *RingBuffer
	HistoryTemp *RingBuffer

	// Current metrics
	HostInfo      *host.InfoStat
	LoadAvg       *load.AvgStat
	MemInfo       *mem.VirtualMemoryStat
	SwapInfo      *mem.SwapMemoryStat
	CpuInfoStatic []cpu.InfoStat
	Battery       []*battery.Battery

	// Disk I/O history and current
	DiskHORead  *RingBuffer
	DiskHOWrite *RingBuffer
	DiskIO      map[string]disk.IOCountersStat
	LastDiskIO  map[string]disk.IOCountersStat

	// Hardware detection flags
	HasNvidiaGPU         bool
	HasAmdGPU            bool
	HasIntelGPU          bool
	HasBattery           bool
	HasNetworkInterfaces bool
	HasDiskIO            bool
	HasServices          bool
	HasTempSensors       bool

	// Current values
	Cpu           float64
	Memory        float64
	Disk          float64
	Swap          float64
	DiskReadRate  float64
	DiskWriteRate float64
	CpuTemp       float64
	CpuPerCore    []float64

	// Disk and network info
	DiskPartitions        []DiskPartition
	Sensors               []host.TemperatureStat
	NetworkInterfaces     []net.IOCountersStat
	LastNetworkInterfaces map[string]net.IOCountersStat
	LastNetSent           uint64
	LastNetRecv           uint64
	NetSentRate           float64
	NetRecvRate           float64

	// GPU info
	GpuInfo []GpuInfo

	// Alerts
	AlertManager *AlertManager

	// Process history for graphs
	ProcessHistory map[int32]*RingBuffer

	// Health
	HealthScore int
}

// ProcessState holds all process management data.
type ProcessState struct {
	// Process lists
	Processes          []ProcessInfo
	ProcessesByPid     map[int32]ProcessInfo
	CachedVisibleProcs []ProcessInfo
	CachedTreeIndents  map[int32]int

	// Tree view state
	TreeView       bool
	CollapsedPids  map[int32]bool
	BookmarkedPids map[int32]bool

	// Filtering and sorting
	ProcessFilter      string
	ProcessFilterLower string
	SortBy             string
	SortDirection      string
	FilterMode         bool

	// Selection and scroll
	SelectedProcess     int
	ProcessCount        int
	ProcessScrollOffset int

	// Kill dialog
	KillTargetName string
	KillTargetPid  int32

	// Loading state
	ProcessesLoaded bool

	// Suspended processes
	SuspendedState map[int32]bool

	// Open files
	OpenFilesView SimpleViewport
	OpenFilesList []process.OpenFilesStat
	ShowOpenFiles bool
	OpenFilesPid  int32

	// Lazy-loaded data
	ProcessCmdlines  map[int32]string
	ProcessUsernames map[int32]string

	// Context menu
	ShowProcessMenu bool
	ProcessMenuIdx  int
	ProcessMenuX    int
	ProcessMenuY    int

	// Services and connections
	Services                []ServiceInfo
	Connections             []ConnectionInfo
	ServicesScrollOffset    int
	ConnectionsScrollOffset int
	SystemLogs              []string
	LogsScrollOffset        int

	// Dirty flag
	ProcessCacheDirty bool

	// Open files scroll
	OpenFilesScrollOffset int
}

// UIState holds all UI and display-related state.
type UIState struct {
	// Window dimensions
	Width  int
	Height int

	// Tab state
	SelectedTab   int
	ActiveTabs    []string
	HistoryLength int

	// Interactive flags
	Paused       bool
	ShowHelp     bool
	ShowSettings bool
	ShowSamLab   bool

	// Settings dialog
	SettingsEdit bool
	SettingsSel  config.MetricType
	SettingsIdx  int

	// Kill dialog
	ShowKillDialog bool
	KillDialogSel  int

	// Toasts
	Toasts      []Toast
	NextToastID int64

	// Error handling
	LastError string
	TickCount uint64

	// Chart display
	ChartType string

	// Scroll offsets for blocks
	CpuCoreScrollOffset      int
	SystemBlockScrollOffsets map[int]int
	ActiveScrollBlock        int
	SystemBlockCount         int
	SystemBlockScrollable    map[int]bool
	SystemBlockMaxScroll     map[int]int

	// Mouse position
	MouseX int
	MouseY int

	// Input zones
	Zones       []input.Zone
	ZoneManager any // *input.ZoneManager - stored as any to avoid circular import

	// Content rendering
	// ContentBuilder is single-owner: a tab borrows it via Reset()+WriteString
	// and immediately calls String(). The render path runs single-threaded on
	// the Bubble Tea loop, so the strings.Builder value semantics are safe.
	ContentBuilder strings.Builder

	// Config timestamps
	StartTime         time.Time
	LastConfigModTime time.Time
}

// ConfigState holds all configuration-related state.
type ConfigState struct {
	// Display configuration
	Theme                string
	RefreshRate          int
	BorderType           string
	BorderStyle          string
	BackgroundOpaque     bool
	ProcessCpuNormalized bool

	// Sorting
	SortBy        string
	SortDirection string

	// History
	HistoryLength int

	// Full config
	Config config.AppConfig

	// Timestamps
	StartTime         time.Time
	LastConfigModTime time.Time
}

// RemoteState holds all remote monitoring state.
type RemoteState struct {
	Metrics map[string]RemoteHostMetrics
}

type RemoteHostMetrics struct {
	Uptime      string
	Cpu         float64
	CpuCount    int
	MemoryTotal uint64
	MemoryUsed  uint64
	MemoryPct   float64
	SwapTotal   uint64
	SwapUsed    uint64
	SwapPct     float64
	DiskTotal   uint64
	DiskUsed    uint64
	DiskPct     float64
	NetSent     uint64
	NetRecv     uint64
	Processes   []RemoteProcessInfo
	LoadAvg1    float64
	LoadAvg5    float64
	LoadAvg15   float64
	Error       string
	Online      bool
}

type RemoteProcessInfo struct {
	Pid    int32
	Name   string
	Cpu    float64
	Memory float64
	Status string
}

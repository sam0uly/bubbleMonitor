package msg

import (
	"time"

	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	// Import shared data types
	"samouly.fun/bubbleMonitor/internal/data"
)

// Message types for Bubble Tea
type TickMsg time.Time

// CpuMemMsg contains fast-updating metrics (CPU, Memory, Swap)
type CpuMemMsg struct {
	LoadAvg    *load.AvgStat
	MemInfo    *mem.VirtualMemoryStat
	SwapInfo   *mem.SwapMemoryStat
	Err        error
	CpuPerCore []float64
	Cpu        float64
	Memory     float64
	Swap       float64
}

// DiskNetMsg contains slow-updating metrics (Disk, Network)
type DiskNetMsg struct {
	Err     error
	Disk    float64
	NetSent uint64
	NetRecv uint64
}

type MetricsMsg struct {
	LoadAvg    *load.AvgStat
	MemInfo    *mem.VirtualMemoryStat
	SwapInfo   *mem.SwapMemoryStat
	CpuPerCore []float64
	Cpu        float64
	Memory     float64
	Disk       float64
	Swap       float64
	NetSent    uint64
	NetRecv    uint64
}

type ProcessesMsg []data.ProcessInfo
type ProcessCountMsg int

type HostInfoMsg struct {
	Err  error
	Info *host.InfoStat
}

type DiskInfoMsg struct {
	Err        error
	Partitions []data.DiskPartition
}

type GpuInfoMsg struct {
	Err  error
	Gpus []data.GpuInfo
}
type DiskIOMsg map[string]disk.IOCountersStat
type TempMsg []host.TemperatureStat

type NetworkInterfacesMsg struct {
	Err        error
	Interfaces []net.IOCountersStat
}
type BatteryMsg []*battery.Battery
type ServicesMsg []data.ServiceInfo
type ConnectionsMsg []data.ConnectionInfo
type SysLogMsg []string
type RemoteMsg struct {
	Metrics data.RemoteHostMetrics
	Host    string
}

// Control Messages
type PriorityChangeMsg struct {
	Err      error
	Pid      int32
	Priority int32
}

type ProcessControlMsg struct {
	Err    error
	Action string // "suspend" or "resume"
	Pid    int32
}

type OpenFilesMsg struct {
	Err   error
	Files []process.OpenFilesStat
	Pid   int32
}

type OpenFilesRequestMsg struct {
	Pid int32
}

// ProcessCmdlineMsg is sent when cmdline is fetched lazily
type ProcessCmdlineMsg struct {
	Pid     int32
	Cmdline string
}

// ProcessUsernameMsg is sent when username is fetched lazily
type ProcessUsernameMsg struct {
	Pid      int32
	Username string
}

// Toast Messages
type ToastMsg struct {
	Message  string
	Level    string
	Duration time.Duration
}

type ToastTimeoutMsg struct {
	ID int64
}

type QuitMsg struct{}

// ExportSnapshotMsg triggers a metrics snapshot export to file.
type ExportSnapshotMsg struct{}

// ForceRefreshMsg triggers an immediate metrics refresh without rescheduling the tick timer.
type ForceRefreshMsg struct{}

// KillProcessMsg is sent when a process kill is requested
type KillProcessMsg struct {
	Error   string
	Pid     int32
	Success bool
}

// ContainerInfoMsg contains container metrics (Docker/Kubernetes)
type ContainerInfoMsg struct {
	Err           error
	Containers    []data.ContainerInfo
	Pods          []data.K8sPodInfo
	HasDocker     bool
	HasKubernetes bool
}

// VmInfoMsg contains VM/hypervisor metrics
type VmInfoMsg struct {
	Err    error
	VmInfo *data.VmInfo
	IsVM   bool
}

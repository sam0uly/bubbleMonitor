package provider

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"samouly.fun/bubbleMonitor/internal/config"
)

// SystemProvider defines the interface for system metrics providers.
// Each method returns a tea.Cmd that produces a message when executed.
type SystemProvider interface {
	// TickCmd returns a command that sends tick messages at the specified duration.
	TickCmd(d time.Duration) tea.Cmd

	// FastMetricsCmd returns a command that fetches CPU, memory, and load average metrics.
	FastMetricsCmd() tea.Cmd

	// SlowMetricsCmd returns a command that fetches disk and network I/O metrics.
	SlowMetricsCmd() tea.Cmd

	// HostInfoCmd returns a command that fetches host information (OS, platform, uptime, etc.).
	HostInfoCmd() tea.Cmd

	// GpuInfoCmd returns a command that fetches GPU information (NVIDIA, AMD, Intel).
	GpuInfoCmd() tea.Cmd

	// TempCmd returns a command that fetches temperature sensor data.
	TempCmd() tea.Cmd

	// BatteryCmd returns a command that fetches battery status.
	BatteryCmd() tea.Cmd

	// DiskInfoCmd returns a command that fetches disk partition information.
	DiskInfoCmd() tea.Cmd

	// DiskIOCmd returns a command that fetches disk I/O statistics.
	DiskIOCmd() tea.Cmd

	// NetworkInterfacesCmd returns a command that fetches network interface statistics.
	NetworkInterfacesCmd() tea.Cmd

	// ConnectionsCmd returns a command that fetches network connections.
	ConnectionsCmd() tea.Cmd

	// ServicesCmd returns a command that fetches system services.
	ServicesCmd() tea.Cmd

	// SystemLogsCmd returns a command that fetches system logs.
	SystemLogsCmd() tea.Cmd

	// ContainerCmd returns a command that fetches docker/k8s container metrics.
	ContainerCmd() tea.Cmd

	// VmCmd returns a command that fetches VM/hypervisor information.
	VmCmd() tea.Cmd

	// HasNvidiaGPU returns true if an NVIDIA GPU is detected.
	HasNvidiaGPU() bool

	// HasAmdGPU returns true if an AMD GPU is detected.
	HasAmdGPU() bool

	// HasIntelGPU returns true if an Intel GPU is detected.
	HasIntelGPU() bool

	// HasBattery returns true if a battery is detected.
	HasBattery() bool

	// HasNetworkInterfaces returns true if network interfaces are detected.
	HasNetworkInterfaces() bool

	// HasDiskIO returns true if disk I/O metrics are available.
	HasDiskIO() bool

	// HasServices returns true if system services are available.
	HasServices() bool

	// HasTempSensors returns true if temperature sensors are available.
	HasTempSensors() bool

	// HasDocker returns true if docker is available.
	HasDocker() bool

	// HasKubernetes returns true if kubernetes is available.
	HasKubernetes() bool

	// DetectHardware runs hardware capability detection.
	DetectHardware()
}

// ProcessProvider defines the interface for process management providers.
// Each method returns a tea.Cmd that produces a message when executed.
type ProcessProvider interface {
	// ProcessesCmd returns a command that fetches all running processes with sorting.
	ProcessesCmd(sortBy string, sortDirection string) tea.Cmd

	// ProcessCountCmd returns a command that fetches the number of running processes (lightweight).
	ProcessCountCmd() tea.Cmd

	// PidsOnlyCmd returns a command that fetches only PIDs for cache warming (lightweight).
	PidsOnlyCmd() tea.Cmd

	// ReniceProcessCmdSafe returns a command that changes process priority.
	// delta < 0 increases priority, delta > 0 decreases priority.
	ReniceProcessCmdSafe(pid int32, delta int) tea.Cmd

	// SuspendProcessCmd returns a command that suspends a process.
	SuspendProcessCmd(pid int32) tea.Cmd

	// ResumeProcessCmd returns a command that resumes a suspended process.
	ResumeProcessCmd(pid int32) tea.Cmd

	// FetchOpenFilesCmd returns a command that fetches open files for a process.
	FetchOpenFilesCmd(pid int32) tea.Cmd

	// FetchProcessCmdlineCmd returns a command that lazily fetches cmdline for a process.
	FetchProcessCmdlineCmd(pid int32) tea.Cmd

	// FetchProcessUsernameCmd returns a command that lazily fetches username for a process.
	FetchProcessUsernameCmd(pid int32) tea.Cmd
}

// RemoteProvider defines the interface for remote system access providers.
// Each method returns a tea.Cmd that produces a message when executed.
type RemoteProvider interface {
	// CheckRemoteCmd returns a command that checks remote host connectivity and fetches full metrics.
	CheckRemoteCmd(host config.RemoteHostConfig) tea.Cmd
}

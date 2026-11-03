package provider

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/provider/process"
	"samouly.fun/bubbleMonitor/internal/provider/remote"
	"samouly.fun/bubbleMonitor/internal/provider/system"
)

// SystemAdapter wraps system provider functions and implements SystemProvider interface.
type SystemAdapter struct{}

// NewSystemAdapter creates a new SystemAdapter instance.
func NewSystemAdapter() *SystemAdapter {
	return &SystemAdapter{}
}

// TickCmd returns a command that sends tick messages at the specified duration.
func (a *SystemAdapter) TickCmd(d time.Duration) tea.Cmd {
	return system.TickCmd(d)
}

// FastMetricsCmd returns a command that fetches CPU, memory, and load average metrics.
func (a *SystemAdapter) FastMetricsCmd() tea.Cmd {
	return system.FastMetricsCmd()
}

// SlowMetricsCmd returns a command that fetches disk and network I/O metrics.
func (a *SystemAdapter) SlowMetricsCmd() tea.Cmd {
	return system.SlowMetricsCmd()
}

// HostInfoCmd returns a command that fetches host information.
func (a *SystemAdapter) HostInfoCmd() tea.Cmd {
	return system.HostInfoCmd()
}

// GpuInfoCmd returns a command that fetches GPU information.
func (a *SystemAdapter) GpuInfoCmd() tea.Cmd {
	return system.GpuInfoCmd()
}

// TempCmd returns a command that fetches temperature sensor data.
func (a *SystemAdapter) TempCmd() tea.Cmd {
	return system.TempCmd()
}

// BatteryCmd returns a command that fetches battery status.
func (a *SystemAdapter) BatteryCmd() tea.Cmd {
	return system.BatteryCmd()
}

// DiskInfoCmd returns a command that fetches disk partition information.
func (a *SystemAdapter) DiskInfoCmd() tea.Cmd {
	return system.DiskInfoCmd()
}

// DiskIOCmd returns a command that fetches disk I/O statistics.
func (a *SystemAdapter) DiskIOCmd() tea.Cmd {
	return system.DiskIOCmd()
}

// NetworkInterfacesCmd returns a command that fetches network interface statistics.
func (a *SystemAdapter) NetworkInterfacesCmd() tea.Cmd {
	return system.NetworkInterfacesCmd()
}

// ConnectionsCmd returns a command that fetches network connections.
func (a *SystemAdapter) ConnectionsCmd() tea.Cmd {
	return system.ConnectionsCmd()
}

// ServicesCmd returns a command that fetches system services.
func (a *SystemAdapter) ServicesCmd() tea.Cmd {
	return system.ServicesCmd()
}

// SystemLogsCmd returns a command that fetches system logs.
func (a *SystemAdapter) SystemLogsCmd() tea.Cmd {
	return system.SystemLogsCmd()
}

// ContainerCmd returns a command that fetches docker/k8s container metrics.
func (a *SystemAdapter) ContainerCmd() tea.Cmd {
	return system.ContainerCmd()
}

// VmCmd returns a command that fetches VM/hypervisor information.
func (a *SystemAdapter) VmCmd() tea.Cmd {
	return system.VmCmd()
}

// HasNvidiaGPU returns true if an NVIDIA GPU is detected.
func (a *SystemAdapter) HasNvidiaGPU() bool {
	caps := system.DetectHardware()
	return caps.HasNvidiaGPU
}

// HasAmdGPU returns true if an AMD GPU is detected.
func (a *SystemAdapter) HasAmdGPU() bool {
	caps := system.DetectHardware()
	return caps.HasAmdGPU
}

// HasIntelGPU returns true if an Intel GPU is detected.
func (a *SystemAdapter) HasIntelGPU() bool {
	caps := system.DetectHardware()
	return caps.HasIntelGPU
}

// HasBattery returns true if a battery is detected.
func (a *SystemAdapter) HasBattery() bool {
	caps := system.DetectHardware()
	return caps.HasBattery
}

// HasNetworkInterfaces returns true if network interfaces are detected.
func (a *SystemAdapter) HasNetworkInterfaces() bool {
	caps := system.DetectHardware()
	return caps.HasNetworkInterfaces
}

// HasDiskIO returns true if disk I/O metrics are available.
func (a *SystemAdapter) HasDiskIO() bool {
	caps := system.DetectHardware()
	return caps.HasDiskIO
}

// HasServices returns true if system services are available.
func (a *SystemAdapter) HasServices() bool {
	caps := system.DetectHardware()
	return caps.HasServices
}

// HasTempSensors returns true if temperature sensors are available.
func (a *SystemAdapter) HasTempSensors() bool {
	caps := system.DetectHardware()
	return caps.HasTempSensors
}

// HasDocker returns true if docker is available.
func (a *SystemAdapter) HasDocker() bool {
	caps := system.DetectHardware()
	return caps.HasDocker
}

// HasKubernetes returns true if kubernetes is available.
func (a *SystemAdapter) HasKubernetes() bool {
	caps := system.DetectHardware()
	return caps.HasKubernetes
}

// DetectHardware runs hardware capability detection.
func (a *SystemAdapter) DetectHardware() {
	system.DetectHardware()
}

// ProcessAdapter wraps process provider functions and implements ProcessProvider interface.
type ProcessAdapter struct{}

// NewProcessAdapter creates a new ProcessAdapter instance.
func NewProcessAdapter() *ProcessAdapter {
	return &ProcessAdapter{}
}

// ProcessesCmd returns a command that fetches all running processes with sorting.
func (a *ProcessAdapter) ProcessesCmd(sortBy string, sortDirection string) tea.Cmd {
	return process.ProcessesCmd(sortBy, sortDirection)
}

// ProcessCountCmd returns a command that fetches the number of running processes (lightweight).
func (a *ProcessAdapter) ProcessCountCmd() tea.Cmd {
	return process.ProcessCountCmd()
}

// PidsOnlyCmd returns a command that fetches only PIDs for cache warming (lightweight).
func (a *ProcessAdapter) PidsOnlyCmd() tea.Cmd {
	return process.PidsOnlyCmd()
}

// ReniceProcessCmdSafe returns a command that changes process priority.
func (a *ProcessAdapter) ReniceProcessCmdSafe(pid int32, delta int) tea.Cmd {
	return process.ReniceProcessCmdSafe(pid, delta)
}

// SuspendProcessCmd returns a command that suspends a process.
func (a *ProcessAdapter) SuspendProcessCmd(pid int32) tea.Cmd {
	return process.SuspendProcessCmd(pid)
}

// ResumeProcessCmd returns a command that resumes a suspended process.
func (a *ProcessAdapter) ResumeProcessCmd(pid int32) tea.Cmd {
	return process.ResumeProcessCmd(pid)
}

// FetchOpenFilesCmd returns a command that fetches open files for a process.
func (a *ProcessAdapter) FetchOpenFilesCmd(pid int32) tea.Cmd {
	return process.FetchOpenFilesCmd(pid)
}

// FetchProcessCmdlineCmd returns a command that lazily fetches cmdline for a process.
func (a *ProcessAdapter) FetchProcessCmdlineCmd(pid int32) tea.Cmd {
	return process.FetchProcessCmdlineCmd(pid)
}

// FetchProcessUsernameCmd returns a command that lazily fetches username for a process.
func (a *ProcessAdapter) FetchProcessUsernameCmd(pid int32) tea.Cmd {
	return process.FetchProcessUsernameCmd(pid)
}

// RemoteAdapter wraps remote provider functions and implements RemoteProvider interface.
type RemoteAdapter struct{}

// NewRemoteAdapter creates a new RemoteAdapter instance.
func NewRemoteAdapter() *RemoteAdapter {
	return &RemoteAdapter{}
}

// CheckRemoteCmd returns a command that checks remote host connectivity and fetches full metrics.
func (a *RemoteAdapter) CheckRemoteCmd(host config.RemoteHostConfig) tea.Cmd {
	return remote.CheckRemoteCmd(host)
}

// Compile-time interface checks
var (
	_ SystemProvider  = (*SystemAdapter)(nil)
	_ ProcessProvider = (*ProcessAdapter)(nil)
	_ RemoteProvider  = (*RemoteAdapter)(nil)
)

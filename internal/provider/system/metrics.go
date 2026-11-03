package system

import (
	"runtime"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"samouly.fun/bubbleMonitor/internal/msg"
)

var cpuInitialized atomic.Bool

func TickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return msg.TickMsg(t)
	})
}

func FastMetricsCmd() tea.Cmd {
	return func() tea.Msg {
		var firstErr error

		cpuPercent, err := cpu.Percent(0, false)
		if err != nil {
			firstErr = err
		}
		cpuVal := 0.0
		if len(cpuPercent) > 0 {
			cpuVal = cpuPercent[0]
		}

		if !cpuInitialized.Load() && cpuVal > 0 {
			cpuInitialized.Store(true)
			cpuVal = 0
		}

		cpuPerCore, err := cpu.Percent(0, true)
		if err != nil && firstErr == nil {
			firstErr = err
		}

		memInfo, err := mem.VirtualMemory()
		if err != nil && firstErr == nil {
			firstErr = err
		}

		swapInfo, err := mem.SwapMemory()
		if err != nil && firstErr == nil {
			firstErr = err
		}

		loadAvg, err := load.Avg()
		if err != nil && firstErr == nil {
			firstErr = err
		}

		var memUsed, swapUsed float64
		if memInfo != nil {
			memUsed = memInfo.UsedPercent
		}
		if swapInfo != nil {
			swapUsed = swapInfo.UsedPercent
		}

		return msg.CpuMemMsg{
			Cpu:        cpuVal,
			CpuPerCore: cpuPerCore,
			Memory:     memUsed,
			Swap:       swapUsed,
			LoadAvg:    loadAvg,
			MemInfo:    memInfo,
			SwapInfo:   swapInfo,
			Err:        firstErr,
		}
	}
}

func SlowMetricsCmd() tea.Cmd {
	return func() tea.Msg {
		var firstErr error
		root := "/"
		if runtime.GOOS == "windows" {
			root = "C:"
		}
		diskInfo, err := disk.Usage(root)
		if err != nil {
			firstErr = err
		}

		diskPercent := 0.0
		if diskInfo != nil {
			diskPercent = diskInfo.UsedPercent
		}

		netIO, err := net.IOCounters(false)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		var netSent, netRecv uint64
		if len(netIO) > 0 {
			netSent = netIO[0].BytesSent
			netRecv = netIO[0].BytesRecv
		}

		return msg.DiskNetMsg{
			Disk:    diskPercent,
			NetSent: netSent,
			NetRecv: netRecv,
			Err:     firstErr,
		}
	}
}

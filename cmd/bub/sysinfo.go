package main

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"

	"samouly.fun/bubbleMonitor/internal/util"
)

func newSysinfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "sysinfo",
		Short:   "Show detailed system information",
		Aliases: []string{"info"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printSysinfo(cmd)
		},
	}
}

func printSysinfo(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	// Host info
	hostInfo, _ := host.Info()
	if hostInfo != nil {
		fmt.Fprintf(out, "\n")
		fmt.Fprintf(out, "  Hostname    %s\n", hostInfo.Hostname)
		fmt.Fprintf(out, "  OS          %s %s (%s)\n", hostInfo.OS, hostInfo.PlatformVersion, hostInfo.KernelVersion)
		fmt.Fprintf(out, "  Kernel      %s\n", hostInfo.KernelVersion)
		fmt.Fprintf(out, "  Arch        %s\n", hostInfo.KernelArch)
		fmt.Fprintf(out, "  Uptime      %s\n", formatUptime(hostInfo.Uptime))
	}

	// CPU info
	cpuInfos, _ := cpu.Info()
	if len(cpuInfos) > 0 {
		c := cpuInfos[0]
		coreCount, _ := cpu.Counts(false)
		threadCount, _ := cpu.Counts(true)
		fmt.Fprintf(out, "  CPU         %s (%d cores, %d threads)\n", c.ModelName, coreCount, threadCount)
	}

	// Memory
	memInfo, _ := mem.VirtualMemory()
	if memInfo != nil {
		fmt.Fprintf(out, "  Memory      %s total / %s used (%s available)\n",
			util.FastBytes(memInfo.Total), util.FastBytes(memInfo.Used), util.FastBytes(memInfo.Available))
	}

	// Swap
	swapInfo, _ := mem.SwapMemory()
	if swapInfo != nil && swapInfo.Total > 0 {
		fmt.Fprintf(out, "  Swap        %s total / %s used\n",
			util.FastBytes(swapInfo.Total), util.FastBytes(swapInfo.Used))
	}

	// Disk partitions
	partitions, _ := disk.Partitions(false)
	if len(partitions) > 0 {
		fmt.Fprintf(out, "  Disks\n")
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				continue
			}
			fmt.Fprintf(out, "    %-12s %s / %s (%.1f%%) %s\n",
				p.Mountpoint,
				util.FastBytes(usage.Used),
				util.FastBytes(usage.Total),
				usage.UsedPercent,
				p.Fstype)
		}
	}

	fmt.Fprintf(out, "  Go          %s\n", runtime.Version())
	fmt.Fprintf(out, "\n")

	return nil
}

package main

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"

	"samouly.fun/bubbleMonitor/internal/util"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show system status overview",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printStatus(cmd)
		},
	}
}

func printStatus(cmd *cobra.Command) error {
	// CPU
	cpuPcts, _ := cpu.Percent(time.Second, false)
	cpuVal := 0.0
	if len(cpuPcts) > 0 {
		cpuVal = cpuPcts[0]
	}

	// Memory
	memInfo, _ := mem.VirtualMemory()
	memVal := 0.0
	memTotal := uint64(0)
	if memInfo != nil {
		memVal = memInfo.UsedPercent
		memTotal = memInfo.Total
	}

	// Swap
	swapVal := 0.0
	swapInfo, _ := mem.SwapMemory()
	if swapInfo != nil {
		swapVal = swapInfo.UsedPercent
	}

	// Disk
	diskVal := 0.0
	diskUsage, _ := disk.Usage("/")
	if diskUsage != nil {
		diskVal = diskUsage.UsedPercent
	}

	// Temperature
	tempVal := 0.0
	hostInfo, _ := host.Info()
	if temps, err := host.SensorsTemperatures(); err == nil {
		for _, t := range temps {
			if t.Temperature > tempVal {
				tempVal = t.Temperature
			}
		}
	}

	// Uptime
	uptime := "unknown"
	if hostInfo != nil {
		uptime = formatUptime(hostInfo.Uptime)
	}

	barWidth := 12
	out := cmd.OutOrStdout()
	s := loadCLIStyles()

	_, _ = lipgloss.Fprintf(out, "\n")
	_, _ = lipgloss.Fprintf(out, "  %s %s  %s   | %s %s  %s\n",
		s.Label.Render("CPU "),
		s.RenderBar(cpuVal, barWidth),
		s.Value.Render(util.FastPercent1(cpuVal)),
		s.Label.Render("Memory "),
		s.RenderBar(memVal, barWidth),
		s.Value.Render(util.FastPercent1(memVal)))
	_, _ = lipgloss.Fprintf(out, "  %s %s  %s   | %s %s  %s\n",
		s.Label.Render("Disk "),
		s.RenderBar(diskVal, barWidth),
		s.Value.Render(util.FastPercent1(diskVal)),
		s.Label.Render("Temp "),
		s.RenderBar(tempVal/100.0, barWidth),
		s.Value.Render(fmt.Sprintf("%.0f°C", tempVal)))
	_, _ = lipgloss.Fprintf(out, "  %s %s  %s   | %s %s\n",
		s.Label.Render("Swap "),
		s.RenderBar(swapVal, barWidth),
		s.Value.Render(util.FastPercent1(swapVal)),
		s.Label.Render("Uptime "),
		s.Dim.Render(uptime))

	if memTotal > 0 {
		_, _ = lipgloss.Fprintf(out, "\n  %s %s / %s\n",
			s.Label.Render("Memory:"),
			s.Value.Render(util.FastBytes(memInfo.Used)),
			s.Value.Render(util.FastBytes(memTotal)))
	}
	_, _ = lipgloss.Fprintf(out, "\n")

	return nil
}

func formatUptime(seconds uint64) string {
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	if d > 0 {
		return fmt.Sprintf("%dd %dh %dm", d, h, m)
	}
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

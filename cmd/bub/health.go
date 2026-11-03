package main

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"

	"samouly.fun/bubbleMonitor/internal/cliout"
	configpkg "samouly.fun/bubbleMonitor/internal/config"
)

func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Show system health score",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printHealth(cmd)
		},
	}
}

func printHealth(cmd *cobra.Command) error {
	cfg, _ := loadConfigWithOverrides()
	thresholds := cfg.Thresholds
	weights := cfg.HealthWeights

	cpuVal := fetchCPUPct()
	memVal := fetchMemPct()
	diskVal := fetchDiskPct()
	tempVal := fetchTempPct()

	score := 100
	out := cmd.OutOrStdout()
	s := loadCLIStyles()

	lipgloss.Fprintf(out, "\n")

	// CPU
	cpuStatus := "OK"
	if cpuVal >= thresholds[configpkg.MetricCPU] {
		cpuStatus = "CRITICAL"
		score -= weights.CpuCritical
	} else if cpuVal >= thresholds[configpkg.MetricCPU]*0.7 {
		cpuStatus = "HIGH"
		score -= weights.CpuHigh
	}
	lipgloss.Fprintf(out, "  %s  %s (%s)\n",
		s.Label.Render("CPU:"),
		styleStatus(s, cpuStatus),
		s.Value.Render(fmt.Sprintf("%.1f%%", cpuVal)))

	// Memory
	memStatus := "OK"
	if memVal >= thresholds[configpkg.MetricMem] {
		memStatus = "CRITICAL"
		score -= weights.MemCritical
	} else if memVal >= thresholds[configpkg.MetricMem]*0.7 {
		memStatus = "HIGH"
		score -= weights.MemHigh
	}
	lipgloss.Fprintf(out, "  %s  %s (%s)\n",
		s.Label.Render("Memory:"),
		styleStatus(s, memStatus),
		s.Value.Render(fmt.Sprintf("%.1f%%", memVal)))

	// Disk
	diskStatus := "OK"
	if diskVal >= thresholds[configpkg.MetricDisk] {
		diskStatus = "CRITICAL"
		score -= weights.DiskCritical
	}
	lipgloss.Fprintf(out, "  %s  %s (%s)\n",
		s.Label.Render("Disk:"),
		styleStatus(s, diskStatus),
		s.Value.Render(fmt.Sprintf("%.1f%%", diskVal)))

	// Temperature
	tempStatus := "OK"
	if tempVal >= thresholds[configpkg.MetricTemp] {
		tempStatus = "CRITICAL"
		score -= weights.TempCritical
	} else if tempVal >= thresholds[configpkg.MetricTemp]*0.8 {
		tempStatus = "HIGH"
		score -= weights.TempHigh
	}
	lipgloss.Fprintf(out, "  %s  %s (%s)\n",
		s.Label.Render("Temp:"),
		styleStatus(s, tempStatus),
		s.Value.Render(fmt.Sprintf("%.0f°C", tempVal)))

	if score < 0 {
		score = 0
	}

	lipgloss.Fprintf(out, "\n  %s %s\n\n",
		s.Label.Render("Health Score:"),
		s.ScoreColor(score).Bold(true).Render(fmt.Sprintf("%d/100", score)))

	if score < 50 {
		os.Exit(2)
	} else if score < 70 {
		os.Exit(1)
	}
	return nil
}

// styleStatus returns a themed rendering of the status string.
func styleStatus(s cliout.CLIStyles, status string) string {
	switch status {
	case "OK":
		return s.OK.Render("OK")
	case "HIGH":
		return s.Warn.Render("HIGH")
	case "CRITICAL":
		return s.Critical.Render("CRITICAL")
	default:
		return status
	}
}

func fetchCPUPct() float64 {
	pcts, _ := cpu.Percent(0, false)
	if len(pcts) == 0 {
		return 0
	}
	return pcts[0]
}

func fetchMemPct() float64 {
	m, _ := mem.VirtualMemory()
	if m == nil {
		return 0
	}
	return m.UsedPercent
}

func fetchDiskPct() float64 {
	d, _ := disk.Usage("/")
	if d == nil {
		return 0
	}
	return d.UsedPercent
}

func fetchTempPct() float64 {
	maxTemp := 0.0
	temps, _ := host.SensorsTemperatures()
	for _, t := range temps {
		if t.Temperature > maxTemp {
			maxTemp = t.Temperature
		}
	}
	return maxTemp
}

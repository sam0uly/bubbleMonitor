package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportOutput string
)

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export metrics snapshot",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(cmd)
		},
	}

	cmd.Flags().StringVar(&exportFormat, "format", "json", "output format: json, csv")
	cmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file (default: stdout)")

	return cmd
}

func runExport(cmd *cobra.Command) error {
	snapshot := gatherSnapshot()

	switch exportFormat {
	case "csv":
		return exportCSV(snapshot)
	default:
		return exportJSON(snapshot)
	}
}

type metricsSnapshot struct {
	Timestamp   string  `json:"timestamp"`
	CPU         float64 `json:"cpu_percent"`
	Memory      float64 `json:"memory_percent"`
	MemoryUsed  uint64  `json:"memory_used_bytes"`
	MemoryTotal uint64  `json:"memory_total_bytes"`
	Swap        float64 `json:"swap_percent"`
	Disk        float64 `json:"disk_percent"`
	DiskUsed    uint64  `json:"disk_used_bytes"`
	DiskTotal   uint64  `json:"disk_total_bytes"`
	Temp        float64 `json:"temp_celsius"`
	NetSentRate float64 `json:"net_sent_rate_mb"`
	NetRecvRate float64 `json:"net_recv_rate_mb"`
	LoadAvg1    float64 `json:"load_avg_1m"`
	LoadAvg5    float64 `json:"load_avg_5m"`
	LoadAvg15   float64 `json:"load_avg_15m"`
	Uptime      uint64  `json:"uptime_seconds"`
}

func gatherSnapshot() metricsSnapshot {
	s := metricsSnapshot{
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// CPU
	pcts, _ := cpu.Percent(0, false)
	if len(pcts) > 0 {
		s.CPU = pcts[0]
	}

	// Memory
	m, _ := mem.VirtualMemory()
	if m != nil {
		s.Memory = m.UsedPercent
		s.MemoryUsed = m.Used
		s.MemoryTotal = m.Total
	}

	// Swap
	sw, _ := mem.SwapMemory()
	if sw != nil {
		s.Swap = sw.UsedPercent
	}

	// Disk
	d, _ := disk.Usage("/")
	if d != nil {
		s.Disk = d.UsedPercent
		s.DiskUsed = d.Used
		s.DiskTotal = d.Total
	}

	// Temp
	temps, _ := host.SensorsTemperatures()
	for _, t := range temps {
		if t.Temperature > s.Temp {
			s.Temp = t.Temperature
		}
	}

	// Network
	counters, _ := net.IOCounters(false)
	if len(counters) > 0 {
		s.NetSentRate = float64(counters[0].BytesSent) / 1024 / 1024
		s.NetRecvRate = float64(counters[0].BytesRecv) / 1024 / 1024
	}

	// Load avg
	la, _ := load.Avg()
	if la != nil {
		s.LoadAvg1 = la.Load1
		s.LoadAvg5 = la.Load5
		s.LoadAvg15 = la.Load15
	}

	// Uptime
	hi, _ := host.Info()
	if hi != nil {
		s.Uptime = hi.Uptime
	}

	return s
}

func exportJSON(s metricsSnapshot) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, data, 0o644); err != nil {
			return err
		}
		s := loadCLIStyles()
		_, _ = lipgloss.Fprintf(os.Stdout, "%s %s\n", s.CheckOK+s.Label.Render(" Exported snapshot to:"), s.Value.Render(exportOutput))
		return nil
	}
	fmt.Println(string(data))
	return nil
}

func exportCSV(s metricsSnapshot) error {
	csv := "timestamp,cpu_percent,memory_percent,memory_used,memory_total,swap_percent,disk_percent,disk_used,disk_total,temp,net_sent_mb,net_recv_mb,load_1m,load_5m,load_15m,uptime\n"
	csv += fmt.Sprintf("%s,%.1f,%.1f,%d,%d,%.1f,%.1f,%d,%d,%.0f,%.2f,%.2f,%.2f,%.2f,%.2f,%d\n",
		s.Timestamp, s.CPU, s.Memory, s.MemoryUsed, s.MemoryTotal,
		s.Swap, s.Disk, s.DiskUsed, s.DiskTotal,
		s.Temp, s.NetSentRate, s.NetRecvRate,
		s.LoadAvg1, s.LoadAvg5, s.LoadAvg15, s.Uptime)

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, []byte(csv), 0o644); err != nil {
			return err
		}
		s := loadCLIStyles()
		_, _ = lipgloss.Fprintf(os.Stdout, "%s %s\n", s.Label.Render("✓ Exported snapshot to"), s.Value.Render(exportOutput))
		return nil
	}
	fmt.Print(csv)
	return nil
}

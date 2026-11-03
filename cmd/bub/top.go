package main

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"

	"samouly.fun/bubbleMonitor/internal/util"
)

var (
	topSort      string
	topDirection string
	topN         int
	topRefresh   int
)

func newTopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Show live process table",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTop(cmd, true)
		},
	}

	cmd.Flags().StringVar(&topSort, "sort", "cpu", "sort by: cpu, mem, pid, name")
	cmd.Flags().StringVar(&topDirection, "direction", "desc", "sort direction: asc, desc")
	cmd.Flags().IntVarP(&topN, "number", "n", 20, "number of processes to show")
	cmd.Flags().IntVar(&topRefresh, "refresh", 2, "refresh interval in seconds")

	return cmd
}

func newPsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: "Show process snapshot",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTop(cmd, false)
		},
	}
}

type procEntry struct {
	Pid    int32
	Name   string
	Cpu    float64
	Mem    float64
	Status string
}

func runTop(cmd *cobra.Command, live bool) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	out := cmd.OutOrStdout()

	for {
		procs, err := fetchProcesses()
		if err != nil {
			return fmt.Errorf("failed to fetch processes: %w", err)
		}

		sortProcesses(procs)

		if topN > len(procs) {
			topN = len(procs)
		}
		visible := procs[:topN]

		if live {
			_, _ = fmt.Fprintf(out, "\033[H\033[2J") // clear screen
		}

		_, _ = fmt.Fprintf(out, "\n  %-8s %-30s %8s %8s  %s\n", "PID", "NAME", "CPU%", "MEM%", "STATUS")
		_, _ = fmt.Fprintf(out, "  %s\n", "──────────────────────────────────────────────────────────────")
		for _, p := range visible {
			name := p.Name
			if len(name) > 28 {
				name = name[:25] + "..."
			}
			_, _ = fmt.Fprintf(out, "  %-8d %-30s %8s %8s  %s\n",
				p.Pid, name,
				util.FastPercent1(p.Cpu),
				util.FastPercent1(p.Mem),
				p.Status)
		}

		if !live {
			_, _ = fmt.Fprintf(out, "\n")
			return nil
		}

		_, _ = fmt.Fprintf(out, "\n  Refreshing every %ds (press Ctrl+C to quit)\n", topRefresh)

		select {
		case <-sigChan:
			_, _ = fmt.Fprintf(out, "\n")
			return nil
		case <-time.After(time.Duration(topRefresh) * time.Second):
		}
	}
}

func fetchProcesses() ([]procEntry, error) {
	pids, err := process.Processes()
	if err != nil {
		return nil, err
	}

	entries := make([]procEntry, 0, len(pids))
	for _, p := range pids {
		name, _ := p.Name()
		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()
		statuses, _ := p.Status()
		status := ""
		if len(statuses) > 0 {
			status = statuses[0]
		}

		entries = append(entries, procEntry{
			Pid:    p.Pid,
			Name:   name,
			Cpu:    cpuPct,
			Mem:    float64(memPct),
			Status: status,
		})
	}

	return entries, nil
}

func sortProcesses(procs []procEntry) {
	switch topSort {
	case "cpu":
		sort.Slice(procs, func(i, j int) bool {
			if topDirection == "asc" {
				return procs[i].Cpu < procs[j].Cpu
			}
			return procs[i].Cpu > procs[j].Cpu
		})
	case "mem":
		sort.Slice(procs, func(i, j int) bool {
			if topDirection == "asc" {
				return procs[i].Mem < procs[j].Mem
			}
			return procs[i].Mem > procs[j].Mem
		})
	case "pid":
		sort.Slice(procs, func(i, j int) bool {
			if topDirection == "desc" {
				return procs[i].Pid > procs[j].Pid
			}
			return procs[i].Pid < procs[j].Pid
		})
	case "name":
		sort.Slice(procs, func(i, j int) bool {
			if topDirection == "desc" {
				return procs[i].Name > procs[j].Name
			}
			return procs[i].Name < procs[j].Name
		})
	}
}

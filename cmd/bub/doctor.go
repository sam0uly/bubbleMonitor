package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/spf13/cobra"

	"samouly.fun/bubbleMonitor/internal/cliout"
	configpkg "samouly.fun/bubbleMonitor/internal/config"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostic checks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd)
		},
	}
}

func runDoctor(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	s := loadCLIStyles()
	failures := 0

	lipgloss.Fprintf(out, "\n")

	// Config file
	configPath, err := configpkg.GetConfigPath()
	if err != nil {
		lipgloss.Fprintf(out, "  %s %v\n", s.CheckFail+s.Label.Render(" Config path:"), err)
		failures++
	} else {
		if _, err := os.Stat(configPath); err != nil {
			lipgloss.Fprintf(out, "  %s %s (%s)\n",
				s.CheckWarn+s.Label.Render(" Config file:"),
				s.Dim.Render("not found at"),
				s.Value.Render(configPath))
		} else {
			lipgloss.Fprintf(out, "  %s %s\n",
				s.CheckOK+s.Label.Render(" Config file:"),
				s.Value.Render(configPath))
		}
	}

	// gopsutil - CPU metrics
	if pcts, err := cpu.Percent(0, false); err != nil || len(pcts) == 0 {
		lipgloss.Fprintf(out, "  %s unavailable (%v)\n", s.CheckFail+s.Label.Render(" CPU metrics:"), err)
		failures++
	} else {
		lipgloss.Fprintf(out, "  %s available (%s)\n",
			s.CheckOK+s.Label.Render(" CPU metrics:"),
			s.Value.Render(fmt.Sprintf("%.1f%%", pcts[0])))
	}

	// Host info
	if hi, err := host.Info(); err != nil {
		lipgloss.Fprintf(out, "  %s unavailable (%v)\n", s.CheckFail+s.Label.Render(" Host info:"), err)
		failures++
	} else {
		lipgloss.Fprintf(out, "  %s %s\n",
			s.CheckOK+s.Label.Render(" Host info:"),
			s.Value.Render(fmt.Sprintf("%s %s", hi.OS, hi.PlatformVersion)))
	}

	// GPU checks
	checkGPU(out, s, "NVIDIA", "nvidia-smi")
	checkGPU(out, s, "AMD", "rocm-smi")

	// SSH binary
	if path, err := exec.LookPath("ssh"); err != nil {
		lipgloss.Fprintf(out, "  %s not found in PATH\n", s.CheckFail+s.Label.Render(" SSH:"))
		failures++
	} else {
		lipgloss.Fprintf(out, "  %s %s\n", s.CheckOK+s.Label.Render(" SSH binary:"), s.Value.Render(path))
	}

	// Remote hosts connectivity
	cfg, _ := loadConfigWithOverrides()
	for _, h := range cfg.RemoteHosts {
		checkRemoteHost(out, s, h)
	}

	lipgloss.Fprintf(out, "\n")
	if failures > 0 {
		os.Exit(1)
	}
	return nil
}

func checkGPU(out io.Writer, s cliout.CLIStyles, name, cmd string) {
	if path, err := exec.LookPath(cmd); err != nil {
		lipgloss.Fprintf(out, "  %s %s not found\n",
			s.CheckWarn+s.Label.Render(" "+name+" GPU:"), s.Value.Render(cmd))
	} else {
		lipgloss.Fprintf(out, "  %s %s detected\n",
			s.CheckOK+s.Label.Render(" "+name+" GPU:"), s.Value.Render(path))
	}
}

func checkRemoteHost(out io.Writer, s cliout.CLIStyles, host configpkg.RemoteHostConfig) {
	args := []string{"-o", "ConnectTimeout=5", "-o", "BatchMode=yes", host.Host, "echo", "ok"}
	start := time.Now()
	c := exec.Command("ssh", args...)
	err := c.Run()
	elapsed := time.Since(start).Truncate(time.Millisecond)

	if err != nil {
		lipgloss.Fprintf(out, "  %s unreachable (%v)\n",
			s.CheckFail+s.Label.Render(fmt.Sprintf(" Remote %q (%s):", host.Name, host.Host)), err)
	} else {
		lipgloss.Fprintf(out, "  %s reachable (%v)\n",
			s.CheckOK+s.Label.Render(fmt.Sprintf(" Remote %q (%s):", host.Name, host.Host)), elapsed)
	}
}

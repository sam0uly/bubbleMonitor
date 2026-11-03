package system

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/coreos/go-systemd/v22/dbus"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
)

// ServicesCmd lists system services without shelling out.
//
// Strategy by init system:
//   - systemd      : go-systemd/dbus (pure Go D-Bus, no fork/exec, ~ms latency).
//   - sysvinit-ish : /etc/init.d/ directory scan, status by PID + comm.
//   - everything else: empty list with a single placeholder entry.
func ServicesCmd() tea.Cmd {
	return func() tea.Msg {
		switch detectInitSystem() {
		case initSystemd:
			return msg.ServicesMsg(listSystemdServices())
		case initSysVinit:
			return msg.ServicesMsg(listSysVinitServices())
		default:
			return msg.ServicesMsg([]data.ServiceInfo{{
				Name:   "info",
				Status: "n/a",
				Description: "Service listing not implemented for this init system (" +
					runtime.GOOS + ")",
			}})
		}
	}
}

type initKind int

const (
	initUnknown initKind = iota
	initSystemd
	initSysVinit
)

// detectInitSystem reads /proc/1/comm to identify the init system.
// systemd always runs as PID 1 and reports "systemd" in /proc/1/comm.
// sysvinit reports "init"; openrc/runit typically run as "init" too,
// so the /etc/init.d/ path covers the common case.
func detectInitSystem() initKind {
	if runtime.GOOS != "linux" {
		return initUnknown
	}
	comm, err := os.ReadFile("/proc/1/comm")
	if err != nil {
		return initUnknown
	}
	switch strings.TrimSpace(string(comm)) {
	case "systemd":
		return initSystemd
	case "init":
		return initSysVinit
	default:
		return initUnknown
	}
}

// listSystemdServices queries systemd over D-Bus for the current state of
// every loaded unit. The connection times out aggressively to keep the
// services tab snappy even on broken D-Bus setups.
func listSystemdServices() []data.ServiceInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil
	}
	defer conn.Close()

	units, err := conn.ListUnitsContext(ctx)
	if err != nil {
		return nil
	}

	services := make([]data.ServiceInfo, 0, len(units))
	for _, u := range units {
		if !strings.HasSuffix(u.Name, ".service") {
			continue
		}
		services = append(services, data.ServiceInfo{
			Name:        strings.TrimSuffix(u.Name, ".service"),
			Status:      u.SubState, // running, dead, exited, failed, etc.
			Description: u.Description,
		})
	}
	return services
}

// listSysVinitServices enumerates /etc/init.d/ scripts. Status is a rough
// heuristic — we look for a process whose comm matches the service name.
// This won't be as accurate as systemd's view, but it's a starting point
// and works without shelling out.
func listSysVinitServices() []data.ServiceInfo {
	entries, err := os.ReadDir("/etc/init.d")
	if err != nil {
		return nil
	}
	running, _ := collectRunningComms()

	services := make([]data.ServiceInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Skip dotfiles, backup files, and readmes.
		if strings.HasPrefix(name, ".") || strings.HasSuffix(name, "~") ||
			strings.EqualFold(name, "README") {
			continue
		}
		status := "stopped"
		if _, ok := running[name]; ok {
			status = "running"
		}
		services = append(services, data.ServiceInfo{
			Name:   name,
			Status: status,
		})
	}
	return services
}

// collectRunningComms walks /proc and returns a set of process comms.
// Cheap-ish (one readdir + per-pid stat) and avoids shelling out to ps.
func collectRunningComms() (map[string]struct{}, error) {
	out := make(map[string]struct{})
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Only numeric entries are PIDs.
		if !isAllDigits(e.Name()) {
			continue
		}
		comm, err := os.ReadFile(filepath.Join("/proc", e.Name(), "comm"))
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(comm))
		if name != "" {
			out[name] = struct{}{}
		}
	}
	return out, nil
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

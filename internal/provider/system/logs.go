package system

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"samouly.fun/bubbleMonitor/internal/msg"
)

// SystemLogsCmd returns the most recent system log entries.
//
// On Linux it shells out to `journalctl`. This is slower than reading
// the journal directly via sdjournal (libsystemd C API), but doesn't
// require libsystemd headers at build time — which matters on
// distros like NixOS where headers aren't on the default cgo path.
//
// The fork/exec cost (~1s) is a known tradeoff; can be revisited
// once libsystemd is available in the build environment.
func SystemLogsCmd() tea.Cmd {
	return func() tea.Msg {
		return msg.SysLogMsg(fetchRecentLogs())
	}
}

func fetchRecentLogs() []string {
	if runtime.GOOS != "linux" {
		return []string{"Logs not implemented for " + runtime.GOOS}
	}
	cmd := exec.Command("journalctl",
		"-n", strconv.Itoa(MaxLogLines),
		"--no-pager",
		"-o", "short-iso",
	)
	out, err := cmd.Output()
	if err != nil {
		return []string{"Error reading logs or no permission: " + err.Error()}
	}
	lines := make([]string, 0, MaxLogLines)
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

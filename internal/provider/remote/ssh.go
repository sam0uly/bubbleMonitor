package remote

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
)

type RemoteProvider struct{}

func New() *RemoteProvider {
	return &RemoteProvider{}
}

func buildSSHCmd(host config.RemoteHostConfig) *exec.Cmd {
	timeout := SSHTimeout
	if host.Timeout > 0 {
		timeout = TimeoutDuration(host.Timeout)
	}
	timeoutOpt := fmt.Sprintf("ConnectTimeout=%d", int(timeout.Seconds()))

	var args []string
	args = append(args, "-o", timeoutOpt)
	args = append(args, "-o", "BatchMode=yes")

	if host.KeyPath != "" {
		args = append(args, "-i", host.KeyPath)
	}

	if host.Port > 0 && host.Port != 22 {
		portOpt := fmt.Sprintf("Port=%d", host.Port)
		args = append(args, "-o", portOpt)
	}

	args = append(args, host.Host)

	script := `UPTIME=$(uptime); LOAD=$(cat /proc/loadavg 2>/dev/null || echo "0.00 0.00 0.00"); CPU_COUNT=$(nproc 2>/dev/null || echo "1"); echo "===UPTIME===$UPTIME"; echo "===LOAD===$LOAD"; echo "===CPU_COUNT===$CPU_COUNT"; cat /proc/meminfo 2>/dev/null | head -20; echo "===ENDMEM==="; df -B1 --output=size,used,pcent / 2>/dev/null | tail -1; echo "===ENDDISK==="; awk '{if(NR>2)print $1,$2,$3,$10}' /proc/net/dev 2>/dev/null | head -5; echo "===ENDNET==="; ps aux --no-headers 2>/dev/null | sort -k3 -rn | head -15`

	args = append(args, script)
	return exec.Command("ssh", args...)
}

func CheckRemoteCmd(host config.RemoteHostConfig) tea.Cmd {
	return func() tea.Msg {
		m := data.RemoteHostMetrics{}

		cmd := buildSSHCmd(host)
		out, err := cmd.Output()
		if err != nil {
			m.Error = err.Error()
			return msg.RemoteMsg{Host: host.Host, Metrics: m}
		}

		parseRemoteOutput(string(out), &m)
		// Online is a function of "did the SSH command succeed", not of
		// whether the parser liked every line. A partial parse shouldn't
		// flip the host offline; we keep the values either way.
		m.Online = true
		if m.Error != "" {
			m.Online = false
		}

		return msg.RemoteMsg{Host: host.Host, Metrics: m}
	}
}

// remoteSectionMarkers are the sentinel strings the SSH script on the remote
// host writes between logical sections. Indices are used so adding a new
// marker only requires updating this list.
var remoteSectionMarkers = []string{
	"===UPTIME===",
	"===LOAD===",
	"===CPU_COUNT===",
	"===ENDMEM===",
	"===ENDDISK===",
	"===ENDNET===",
}

// sectionRanges returns a slice of (start, end) byte offsets into output
// for each marker. end is the start of the next marker (or len(output) for
// the last section). This is the single source of truth for slicing
// the remote script output; everything else uses it instead of doing its
// own fragile string-splitting.
func sectionRanges(output string) []sectionRange {
	ranges := make([]sectionRange, 0, len(remoteSectionMarkers))
	for i, marker := range remoteSectionMarkers {
		start := strings.Index(output, marker)
		if start < 0 {
			continue
		}
		start += len(marker)
		end := len(output)
		for j := i + 1; j < len(remoteSectionMarkers); j++ {
			if next := strings.Index(output[start:], remoteSectionMarkers[j]); next >= 0 {
				end = start + next
				break
			}
		}
		ranges = append(ranges, sectionRange{name: marker, start: start, end: end})
	}
	return ranges
}

type sectionRange struct {
	name  string
	start int
	end   int
}

func parseRemoteOutput(output string, m *data.RemoteHostMetrics) {
	for _, sr := range sectionRanges(output) {
		body := output[sr.start:sr.end]
		switch sr.name {
		case "===UPTIME===":
			m.Uptime = parseUptime(body)
		case "===LOAD===":
			parseLoadAvg(body, m)
		case "===CPU_COUNT===":
			m.CpuCount, _ = strconv.Atoi(strings.TrimSpace(body))
		case "===ENDMEM===":
			parseMeminfo(body, m)
		case "===ENDDISK===":
			parseDisk(body, m)
		case "===ENDNET===":
			parseNet(body, m)
			// Anything after the last marker is the process list.
			parseProcesses(output[sr.end:], m)
		}
	}

	if m.Uptime == "" {
		for _, line := range strings.Split(output, "\n") {
			l := strings.TrimSpace(line)
			if strings.HasPrefix(l, "up") || strings.HasPrefix(l, "load") {
				m.Uptime = l
				break
			}
		}
	}
}

func parseUptime(line string) string {
	uptime := strings.TrimSpace(line)
	uptime = strings.Split(uptime, ",")[0]
	return uptime
}

func parseLoadAvg(line string, m *data.RemoteHostMetrics) {
	fields := strings.Fields(line)
	if len(fields) >= 1 {
		if v, err := strconv.ParseFloat(fields[0], 64); err == nil {
			m.LoadAvg1 = v
		}
	}
	if len(fields) >= 2 {
		if v, err := strconv.ParseFloat(fields[1], 64); err == nil {
			m.LoadAvg5 = v
		}
	}
	if len(fields) >= 3 {
		if v, err := strconv.ParseFloat(fields[2], 64); err == nil {
			m.LoadAvg15 = v
		}
	}
}

func parseMeminfo(text string, m *data.RemoteHostMetrics) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	var memTotal, memFree, memAvail, swapTotal, swapFree uint64

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			memTotal = val * 1024
		case "MemFree:":
			memFree = val * 1024
		case "MemAvailable:":
			memAvail = val * 1024
		case "SwapTotal:":
			swapTotal = val * 1024
		case "SwapFree:":
			swapFree = val * 1024
		}
	}

	m.MemoryTotal = memTotal
	if memAvail > 0 {
		m.MemoryUsed = memTotal - memAvail
	} else {
		m.MemoryUsed = memTotal - memFree
	}
	if m.MemoryTotal > 0 {
		m.MemoryPct = float64(m.MemoryUsed) / float64(m.MemoryTotal) * 100
	}

	m.SwapTotal = swapTotal
	m.SwapUsed = swapTotal - swapFree
	if m.SwapTotal > 0 {
		m.SwapPct = float64(m.SwapUsed) / float64(m.SwapTotal) * 100
	}
}

func parseDisk(text string, m *data.RemoteHostMetrics) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		total, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		used, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		pctStr := strings.TrimSuffix(fields[2], "%")
		pct, err := strconv.ParseFloat(pctStr, 64)
		if err != nil {
			continue
		}
		m.DiskTotal = total
		m.DiskUsed = used
		m.DiskPct = pct
		break
	}
}

func parseNet(text string, m *data.RemoteHostMetrics) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		rx, err := strconv.ParseUint(fields[1], 10, 64)
		if err == nil {
			m.NetRecv = rx
		}
		tx, err := strconv.ParseUint(fields[9], 10, 64)
		if err == nil {
			m.NetSent = tx
		}
		break
	}
}

func parseProcesses(text string, m *data.RemoteHostMetrics) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}
		pid, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)
		name := fields[10]
		status := ""
		if len(fields) > 7 {
			status = fields[7]
		}

		m.Processes = append(m.Processes, data.RemoteProcessInfo{
			Pid:    int32(pid),
			Name:   name,
			Cpu:    cpu,
			Memory: mem,
			Status: status,
		})

		if len(m.Processes) >= 10 {
			break
		}
	}
}

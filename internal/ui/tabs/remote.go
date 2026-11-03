package tabs

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
)

func RenderRemote(s *data.AppState, container lipgloss.Style, su, w, a, t, mu, p, b color.Color, availHeight int) string {
	boxWidth := s.UI.Width
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	visibleRows := max(availHeight-3, 1)

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(p)
	statusOnlineStyle := lipgloss.NewStyle().Foreground(su)
	statusOfflineStyle := lipgloss.NewStyle().Foreground(a)
	mutedStyle := lipgloss.NewStyle().Foreground(mu)
	labelStyle := lipgloss.NewStyle().Foreground(mu)
	valueStyle := lipgloss.NewStyle().Foreground(t)
	alertStyle := lipgloss.NewStyle().Foreground(a)

	if len(s.Config.Config.RemoteHosts) == 0 {
		msg := "No remote hosts configured.\n\nAdd hosts to config.json:\n\n" +
			"\"remote_hosts\": [\n" +
			"  {\n" +
			"    \"name\": \"Server 1\",\n" +
			"    \"host\": \"user@192.168.1.100\",\n" +
			"    \"key_path\": \"~/.ssh/id_rsa\",\n" +
			"    \"port\": 22,\n" +
			"    \"timeout\": 5\n" +
			"  }\n" +
			"]"
		c := container.Width(boxWidth).Height(visibleRows).BorderTop(false)
		body := c.Render(mutedStyle.Render(msg))
		topBorder := widgets.RenderTopBorderWithBg("REMOTE HOSTS", boxWidth, border, b, p)
		return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
	}

	var rows []string

	for _, hostCfg := range s.Config.Config.RemoteHosts {
		name := hostCfg.Name
		if name == "" {
			name = hostCfg.Host
		}

		metrics, ok := s.Remote.Metrics[hostCfg.Host]
		var statusText string
		var statusStyle lipgloss.Style
		if !ok {
			statusText = "Waiting..."
			statusStyle = mutedStyle
		} else if !metrics.Online {
			statusText = "Offline"
			statusStyle = statusOfflineStyle
		} else {
			statusText = "Online"
			statusStyle = statusOnlineStyle
		}

		var hostBlock strings.Builder
		hostBlock.WriteString(nameStyle.Render(name) + "  " + labelStyle.Render("Status: ") + statusStyle.Render(statusText) + "\n")

		if ok && metrics.Online {
			if metrics.Uptime != "" {
				hostBlock.WriteString(labelStyle.Render("Uptime: ") + valueStyle.Render(strings.TrimSpace(metrics.Uptime)) + "\n")
			}

			if metrics.MemoryTotal > 0 {
				memTotal := formatBytes(metrics.MemoryTotal)
				memUsed := formatBytes(metrics.MemoryUsed)
				memPct := fmt.Sprintf("%.1f%%", metrics.MemoryPct)
				memColor := valueStyle
				if metrics.MemoryPct > 80 {
					memColor = alertStyle
				} else if metrics.MemoryPct > 60 {
					memColor = lipgloss.NewStyle().Foreground(w)
				}
				hostBlock.WriteString(labelStyle.Render("Memory: ") + memUsed + " / " + memTotal + " ")
				hostBlock.WriteString(memColor.Render(memPct) + "\n")
			}

			if metrics.DiskTotal > 0 {
				diskTotal := formatBytes(metrics.DiskTotal)
				diskUsed := formatBytes(metrics.DiskUsed)
				diskPct := fmt.Sprintf("%.1f%%", metrics.DiskPct)
				diskColor := valueStyle
				if metrics.DiskPct > 90 {
					diskColor = alertStyle
				} else if metrics.DiskPct > 75 {
					diskColor = lipgloss.NewStyle().Foreground(w)
				}
				hostBlock.WriteString(labelStyle.Render("Disk:   ") + diskUsed + " / " + diskTotal + " ")
				hostBlock.WriteString(diskColor.Render(diskPct) + "\n")
			}

			if metrics.SwapTotal > 0 {
				swapTotal := formatBytes(metrics.SwapTotal)
				swapUsed := formatBytes(metrics.SwapUsed)
				swapPct := fmt.Sprintf("%.1f%%", metrics.SwapPct)
				hostBlock.WriteString(labelStyle.Render("Swap:   ") + swapUsed + " / " + swapTotal + " ")
				hostBlock.WriteString(valueStyle.Render(swapPct) + "\n")
			}

			if metrics.LoadAvg1 > 0 {
				loadColor := valueStyle
				if metrics.CpuCount > 0 && metrics.LoadAvg1 > float64(metrics.CpuCount) {
					loadColor = alertStyle
				} else if metrics.CpuCount > 0 && metrics.LoadAvg1 > float64(metrics.CpuCount)/2 {
					loadColor = lipgloss.NewStyle().Foreground(w)
				}
				hostBlock.WriteString(labelStyle.Render("Load:   ") + loadColor.Render(fmt.Sprintf("%.2f %.2f %.2f", metrics.LoadAvg1, metrics.LoadAvg5, metrics.LoadAvg15)))
				if metrics.CpuCount > 0 {
					hostBlock.WriteString(mutedStyle.Render(" (cores: " + fmt.Sprintf("%d", metrics.CpuCount) + ")"))
				}
				hostBlock.WriteString("\n")
			}

			if metrics.NetRecv > 0 || metrics.NetSent > 0 {
				hostBlock.WriteString(labelStyle.Render("Net:    ") + valueStyle.Render(formatBytes(metrics.NetRecv)+" rx / "+formatBytes(metrics.NetSent)+" tx") + "\n")
			}

			if len(metrics.Processes) > 0 {
				hostBlock.WriteString(labelStyle.Render("Top Processes:") + "\n")
				for i, proc := range metrics.Processes {
					if i >= 5 {
						break
					}
					hostBlock.WriteString(mutedStyle.Render(fmt.Sprintf("  %5d ", proc.Pid)))
					hostBlock.WriteString(fmt.Sprintf("%.1f%% ", proc.Cpu))
					hostBlock.WriteString(fmt.Sprintf("%.1f%% ", proc.Memory))
					hostBlock.WriteString(truncate(proc.Name, 20) + "\n")
				}
			}
		} else if ok && metrics.Error != "" {
			hostBlock.WriteString(labelStyle.Render("Error: ") + alertStyle.Render(truncate(metrics.Error, 60)) + "\n")
		}

		rows = append(rows, hostBlock.String())
	}

	content := strings.Join(rows, "\n")

	c := container.Width(boxWidth).Height(visibleRows).BorderTop(false)
	body := c.Render(content)
	topBorder := widgets.RenderTopBorderWithBg("REMOTE HOSTS", boxWidth, border, b, p)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, body)
}

func formatBytes(b uint64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	if b < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(b)/1024/1024)
	}
	if b < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.1f GB", float64(b)/1024/1024/1024)
	}
	return fmt.Sprintf("%.1f TB", float64(b)/1024/1024/1024/1024)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}

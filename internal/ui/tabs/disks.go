package tabs

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
	"samouly.fun/bubbleMonitor/internal/util"
)

// RenderDisks renders the disk partitions tab. The data layer is
// responsible for putting real usage data on every partition that has it —
// mounted filesystems from gopsutil, swap enriched from OS-level stats, etc.
// Anything still showing UsedPct < 0 has no usage data and renders as a
// compact row under "Unmounted".
func RenderDisks(s *data.AppState, container lipgloss.Style, su, w, a, t, mu, p, b color.Color, availHeight int) string {
	if len(s.Metrics.DiskPartitions) == 0 {
		return "Loading disk information..."
	}

	width := s.UI.Width
	boxWidth := width
	contentWidth := boxWidth - 4
	barWidth := boxWidth - 6
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	mountStyle := lipgloss.NewStyle().Bold(true).Foreground(t)
	infoStyle := lipgloss.NewStyle().Foreground(mu)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(mu)

	s.UI.ContentBuilder.Reset()

	writeMounted := func(d data.DiskPartition) {
		bar := widgets.RenderProgressBar(d.UsedPct, barWidth, su, w, a)
		info := "Used: " + util.FormatBytes(d.Used) + " / " + util.FormatBytes(d.Total)
		mountLabel := d.Mountpoint
		if d.Fstype != "" {
			mountLabel += " (" + d.Fstype + ")"
		}

		s.UI.ContentBuilder.WriteString(mountStyle.Render(util.PadRight(mountLabel, contentWidth)))
		s.UI.ContentBuilder.WriteString("\n")
		s.UI.ContentBuilder.WriteString(bar)
		s.UI.ContentBuilder.WriteString("\n")
		s.UI.ContentBuilder.WriteString(infoStyle.Render(util.PadRight(info, contentWidth)))
	}

	writeCompact := func(dp data.DiskPartition) {
		name := dp.Device
		if name == "" {
			name = dp.Mountpoint
		}
		s.UI.ContentBuilder.WriteString(infoStyle.Render("─ ") +
			infoStyle.Render(util.PadRight(name, contentWidth-12)) +
			infoStyle.Render(util.FormatBytes(dp.Total)))
	}

	var mounted, unmounted []data.DiskPartition
	for _, d := range s.Metrics.DiskPartitions {
		if d.UsedPct >= 0 {
			mounted = append(mounted, d)
		} else {
			unmounted = append(unmounted, d)
		}
	}

	for i, d := range mounted {
		if i > 0 {
			s.UI.ContentBuilder.WriteString("\n\n")
		}
		writeMounted(d)
	}

	if len(unmounted) > 0 {
		if len(mounted) > 0 {
			s.UI.ContentBuilder.WriteString("\n")
			s.UI.ContentBuilder.WriteString(infoStyle.Render(strings.Repeat("─", contentWidth)))
			s.UI.ContentBuilder.WriteString("\n")
		}
		s.UI.ContentBuilder.WriteString(headerStyle.Render("Unmounted"))
		s.UI.ContentBuilder.WriteString("\n")
		for i, d := range unmounted {
			if i > 0 {
				s.UI.ContentBuilder.WriteString("\n")
			}
			writeCompact(d)
		}
	}

	content := s.UI.ContentBuilder.String()

	titleText := "DISK PARTITIONS (Total R: " + util.FastMbPerSec(s.Metrics.DiskReadRate) + " W: " + util.FastMbPerSec(s.Metrics.DiskWriteRate) + ")"

	contentHeight := max(availHeight-2, 0)

	c := container.Width(boxWidth).Height(contentHeight).BorderTop(false)
	body := c.Render(content)
	topBorder := widgets.RenderTopBorderWithBg(titleText, boxWidth, border, b, p)

	s.UI.ContentBuilder.Reset()
	s.UI.ContentBuilder.WriteString(topBorder)
	s.UI.ContentBuilder.WriteByte('\n')
	s.UI.ContentBuilder.WriteString(body)

	return s.UI.ContentBuilder.String()
}

package tabs

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
	"samouly.fun/bubbleMonitor/internal/util"
)

var networkLabels = struct {
	rxTotal, txTotal, rxRate, txRate, err, drop string
}{
	rxTotal: "Rx Total:",
	txTotal: "Tx Total:",
	rxRate:  "Rx Rate:",
	txRate:  "Tx Rate:",
	err:     "Errors:",
	drop:    "Dropped:",
}

// RenderNetwork renders the network interfaces tab
func RenderNetwork(s *data.AppState, container, titleStyle, labelStyle, valueStyle lipgloss.Style, t, mu, p, b, bg color.Color, availHeight int) string {
	if len(s.Metrics.NetworkInterfaces) == 0 {
		return "Loading network interfaces..."
	}

	width := s.UI.Width
	cols := 1
	if width >= 80 {
		cols = 2
	}
	if width >= 100 {
		cols = 3
	}

	colWidths := util.CalculateColumnWidths(width, cols)
	border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)

	rxTotalLabel := labelStyle.Render(networkLabels.rxTotal)
	txTotalLabel := labelStyle.Render(networkLabels.txTotal)
	rxRateLabel := labelStyle.Render(networkLabels.rxRate)
	txRateLabel := labelStyle.Render(networkLabels.txRate)
	errLabel := labelStyle.Render(networkLabels.err)
	dropLabel := labelStyle.Render(networkLabels.drop)

	sb := &s.UI.ContentBuilder

	var netBlocks []string
	for i, nic := range s.Metrics.NetworkInterfaces {
		cW := colWidths[i%cols] - 6
		sb.Reset()

		// Check for inactivity
		if nic.BytesRecv == 0 && nic.BytesSent == 0 {
			inactiveStyle := lipgloss.NewStyle().Foreground(mu).Align(lipgloss.Center).Width(cW)
			msg := inactiveStyle.Render("(Inactive / Down)")

			c := container.Width(colWidths[i%cols]).Height(8).BorderTop(false)
			body := c.Render(lipgloss.NewStyle().MarginTop(3).Render(msg))

			topBorder := widgets.RenderTopBorderWithBg(nic.Name, colWidths[i%cols], border, mu, p)
			netBlocks = append(netBlocks, lipgloss.JoinVertical(lipgloss.Left, topBorder, body))
			continue
		}

		var rxRate, txRate float64
		if last, ok := s.Metrics.LastNetworkInterfaces[nic.Name]; ok {
			if nic.BytesRecv >= last.BytesRecv {
				rxRate = float64(nic.BytesRecv-last.BytesRecv) / 1024 / 1024
			}
			if nic.BytesSent >= last.BytesSent {
				txRate = float64(nic.BytesSent-last.BytesSent) / 1024 / 1024
			}
		}

		rxTotalVal := util.FormatBytes(nic.BytesRecv)
		sb.WriteString(rxTotalLabel)
		sb.WriteString(valueStyle.Render(rxTotalVal))
		if p := cW - 10 - len(rxTotalVal); p > 0 {
			sb.WriteString(strings.Repeat(" ", p))
		}
		sb.WriteString("\n")

		txTotalVal := util.FormatBytes(nic.BytesSent)
		sb.WriteString(txTotalLabel)
		sb.WriteString(valueStyle.Render(txTotalVal))
		if p := cW - 10 - len(txTotalVal); p > 0 {
			sb.WriteString(strings.Repeat(" ", p))
		}
		sb.WriteString("\n")

		rxRateVal := util.FastMbPerSec(rxRate)
		sb.WriteString(rxRateLabel)
		sb.WriteString(valueStyle.Render(rxRateVal))
		if p := cW - 10 - len(rxRateVal); p > 0 {
			sb.WriteString(strings.Repeat(" ", p))
		}
		sb.WriteString("\n")

		txRateVal := util.FastMbPerSec(txRate)
		sb.WriteString(txRateLabel)
		sb.WriteString(valueStyle.Render(txRateVal))
		if p := cW - 10 - len(txRateVal); p > 0 {
			sb.WriteString(strings.Repeat(" ", p))
		}
		sb.WriteString("\n")

		errVal := "In:" + util.FastUint64(nic.Errin) + " Out:" + util.FastUint64(nic.Errout)
		sb.WriteString(errLabel)
		sb.WriteString(valueStyle.Render(errVal))
		if p := cW - 10 - len(errVal); p > 0 {
			sb.WriteString(strings.Repeat(" ", p))
		}
		sb.WriteString("\n")

		dropVal := "In:" + util.FastUint64(nic.Dropin) + " Out:" + util.FastUint64(nic.Dropout)
		sb.WriteString(dropLabel)
		sb.WriteString(valueStyle.Render(dropVal))
		if p := cW - 10 - len(dropVal); p > 0 {
			sb.WriteString(strings.Repeat(" ", p))
		}

		stats := sb.String()

		c := container.Width(colWidths[i%cols]).Height(8).BorderTop(false)
		body := c.Render(lipgloss.NewStyle().MarginTop(1).Render(stats))
		topBorder := widgets.RenderTopBorderWithBg(nic.Name, colWidths[i%cols], border, b, p)

		netBlocks = append(netBlocks, lipgloss.JoinVertical(lipgloss.Left, topBorder, body))
	}

	var rows []string
	for i := 0; i < len(netBlocks); i += cols {
		end := min(i+cols, len(netBlocks))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, netBlocks[i:end]...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

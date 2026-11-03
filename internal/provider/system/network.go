package system

import (
	tea "charm.land/bubbletea/v2"
	"github.com/shirou/gopsutil/v3/net"

	"samouly.fun/bubbleMonitor/internal/msg"
)

// NetworkInterfacesCmd fetches network interface stats
func NetworkInterfacesCmd() tea.Cmd {
	return func() tea.Msg {
		ioCounters, err := net.IOCounters(true)
		return msg.NetworkInterfacesMsg{
			Interfaces: ioCounters,
			Err:        err,
		}
	}
}

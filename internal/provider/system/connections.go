package system

import (
	"strconv"

	tea "charm.land/bubbletea/v2"
	"github.com/shirou/gopsutil/v3/net"
	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
)

func ConnectionsCmd() tea.Cmd {
	return func() tea.Msg {
		// "inet" = TCP + UDP only. "all" also returns unix domain sockets,
		// which can be thousands on busy systems and made the connections
		// tab take 1-2s to render.
		conns, err := net.Connections("inet")
		if err != nil {
			return msg.ConnectionsMsg{}
		}

		var result []data.ConnectionInfo
		for _, c := range conns {
			result = append(result, data.ConnectionInfo{
				LocalAddr:  c.Laddr.IP + ":" + strconv.Itoa(int(c.Laddr.Port)),
				RemoteAddr: c.Raddr.IP + ":" + strconv.Itoa(int(c.Raddr.Port)),
				State:      c.Status,
				Pid:        c.Pid,
				Protocol:   getProto(c.Type),
			})
		}

		return msg.ConnectionsMsg(result)
	}
}

func getProto(t uint32) string {
	switch t {
	case ProtoTCP:
		return "TCP"
	case ProtoUDP:
		return "UDP"
	default:
		return "UNK"
	}
}

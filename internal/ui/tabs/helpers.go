package tabs

import (
	"samouly.fun/bubbleMonitor/internal/data"
)

// CalcNetPercent calculates network usage percentage
func CalcNetPercent(s *data.AppState) float64 {
	total := s.Metrics.NetSentRate + s.Metrics.NetRecvRate
	netP := (total / 10) * 100
	if netP > 100 {
		netP = 100
	}
	return netP
}

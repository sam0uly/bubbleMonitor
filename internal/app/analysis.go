package app

import (
	configpkg "samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
)

// UpdateHealthScore calculates the system health score based on critical metrics.
// We start with 100 points and deduct points based on severity of usage (CPU, Memory, Disk, Temp).
func UpdateHealthScore(s *data.AppState) {
	score := data.MaxHealthScore
	weights := s.Config.Config.HealthWeights
	thresholds := s.Config.Config.Thresholds

	// CPU penalty
	if s.Metrics.Cpu > thresholds[configpkg.MetricCPU] {
		score -= weights.CpuCritical
	} else if s.Metrics.Cpu > (thresholds[configpkg.MetricCPU] * 0.7) { // 70% of critical is warning
		score -= weights.CpuHigh
	}

	// Memory penalty
	if s.Metrics.Memory > thresholds[configpkg.MetricMem] {
		score -= weights.MemCritical
	} else if s.Metrics.Memory > (thresholds[configpkg.MetricMem] * 0.7) {
		score -= weights.MemHigh
	}

	// Disk space penalty
	if s.Metrics.Disk > thresholds[configpkg.MetricDisk] {
		score -= weights.DiskCritical
	}

	// Temperature penalty
	if s.Metrics.CpuTemp > thresholds[configpkg.MetricTemp] {
		score -= weights.TempCritical
	} else if s.Metrics.CpuTemp > (thresholds[configpkg.MetricTemp] * 0.8) {
		score -= weights.TempHigh
	}

	if score < 0 {
		score = 0
	}
	s.Metrics.HealthScore = score
}

// UpdateProcessHistory maintains a history buffer for the selected process
// and top resource consumers. This allows the charts to show history even
// if you just clicked on a process.
func UpdateProcessHistory(s *data.AppState, alivePids map[int32]bool) {
	if len(s.Process.Processes) == 0 {
		return
	}

	// We want to track history for:
	// 1. The currently selected process (obviously)
	// 2. The top N processes by CPU, so if we switch to them, we have data ready
	targetPids := make(map[int32]bool)

	if s.Process.SelectedProcess >= 0 && s.Process.SelectedProcess < len(s.Process.Processes) {
		pid := s.Process.Processes[s.Process.SelectedProcess].Pid
		targetPids[pid] = true
	}

	// Pre-warm history for top consumers
	limit := data.TopProcessesTrackCount
	if len(s.Process.Processes) < limit {
		limit = len(s.Process.Processes)
	}
	for i := 0; i < limit; i++ {
		targetPids[s.Process.Processes[i].Pid] = true
	}

	for _, p := range s.Process.Processes {
		if targetPids[p.Pid] {
			hist := s.GetOrCreateHistory(p.Pid)
			hist.Push(p.Cpu)
		}
	}

	// Cleanup old data to prevent memory leaks
	s.PruneDeadProcessHistory(alivePids)
}

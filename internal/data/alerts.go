package data

import (
	"fmt"
	"sync"
	"time"

	"samouly.fun/bubbleMonitor/internal/config"
)

type Alert struct {
	Timestamp time.Time
	Message   string
	Type      config.MetricType
	Value     float64
	Threshold float64
}

type AlertManager struct {
	ActiveAlerts map[config.MetricType]Alert
	mu           sync.RWMutex
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		ActiveAlerts: make(map[config.MetricType]Alert),
	}
}

// snapshot is a thread-safe copy of the fields CheckAlerts reads from AppState.
// Reading directly from s.Metrics / s.Config without holding stateMu would
// race with metrics-collection and config-reload goroutines.
type alertSnapshot struct {
	cpu            float64
	memory         float64
	cpuTemp        float64
	thresholds     map[config.MetricType]float64
	diskPartitions []DiskPartition
}

func snapshotForAlerts(s *AppState) alertSnapshot {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return alertSnapshot{
		cpu:            s.Metrics.Cpu,
		memory:         s.Metrics.Memory,
		cpuTemp:        s.Metrics.CpuTemp,
		thresholds:     s.Config.Config.Thresholds,
		diskPartitions: s.Metrics.DiskPartitions,
	}
}

func (am *AlertManager) CheckAlerts(s *AppState) {
	snap := snapshotForAlerts(s)

	am.mu.Lock()
	defer am.mu.Unlock()

	cpuThreshold := snap.thresholds[config.MetricCPU]
	if cpuThreshold > 0 && snap.cpu > cpuThreshold {
		am.ActiveAlerts[config.MetricCPU] = Alert{
			Type:      config.MetricCPU,
			Value:     snap.cpu,
			Threshold: cpuThreshold,
			Message:   fmt.Sprintf("CPU Usage High: %.1f%% > (%.0f%%)", snap.cpu, cpuThreshold),
			Timestamp: time.Now(),
		}
	} else {
		delete(am.ActiveAlerts, config.MetricCPU)
	}

	memThreshold := snap.thresholds[config.MetricMem]
	if memThreshold > 0 && snap.memory > memThreshold {
		am.ActiveAlerts[config.MetricMem] = Alert{
			Type:      config.MetricMem,
			Value:     snap.memory,
			Threshold: memThreshold,
			Message:   fmt.Sprintf("Memory Usage High: %.1f%% > (%.0f%%)", snap.memory, memThreshold),
			Timestamp: time.Now(),
		}
	} else {
		delete(am.ActiveAlerts, config.MetricMem)
	}

	diskThreshold := snap.thresholds[config.MetricDisk]
	if diskThreshold > 0 {
		var totalUsed, totalAll uint64
		for _, part := range snap.diskPartitions {
			totalUsed += part.Used
			totalAll += part.Total
		}

		if totalAll > 0 {
			overallPct := (float64(totalUsed) / float64(totalAll)) * 100.0

			if overallPct > diskThreshold {
				am.ActiveAlerts[config.MetricDisk] = Alert{
					Type:      config.MetricDisk,
					Value:     overallPct,
					Threshold: diskThreshold,
					Message:   fmt.Sprintf("Overall Disk High: %.1f%% > (%.0f%%)", overallPct, diskThreshold),
					Timestamp: time.Now(),
				}
			} else {
				delete(am.ActiveAlerts, config.MetricDisk)
			}
		} else {
			// Disks vanished or no data; clear any stale alert.
			delete(am.ActiveAlerts, config.MetricDisk)
		}
	} else {
		// Threshold toggled off; mirror the CPU/Memory/Temp branches and
		// drop any stale disk alert.
		delete(am.ActiveAlerts, config.MetricDisk)
	}

	tempThreshold := snap.thresholds[config.MetricTemp]
	if tempThreshold > 0 && snap.cpuTemp > tempThreshold {
		am.ActiveAlerts[config.MetricTemp] = Alert{
			Type:      config.MetricTemp,
			Value:     snap.cpuTemp,
			Threshold: tempThreshold,
			Message:   fmt.Sprintf("CPU Temp High: %.1f°C > (%.0f°C)", snap.cpuTemp, tempThreshold),
			Timestamp: time.Now(),
		}
	} else {
		delete(am.ActiveAlerts, config.MetricTemp)
	}
}

func (am *AlertManager) GetAlerts() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	alerts := make([]Alert, 0, len(am.ActiveAlerts))
	for _, alert := range am.ActiveAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

func (am *AlertManager) HasAlerts() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.ActiveAlerts) > 0
}

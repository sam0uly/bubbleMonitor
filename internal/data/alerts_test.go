package data

import (
	"sync"
	"testing"

	"samouly.fun/bubbleMonitor/internal/config"
)

func TestAlertManagerConcurrentAccess(t *testing.T) {
	am := NewAlertManager()
	state := &AppState{
		Metrics: MetricsState{
			Cpu:     75.0,
			Memory:  80.0,
			CpuTemp: 65.0,
		},
		Config: ConfigState{
			Config: config.AppConfig{
				Thresholds: map[config.MetricType]float64{
					config.MetricCPU:  70.0,
					config.MetricMem:  75.0,
					config.MetricTemp: 60.0,
				},
			},
		},
	}

	const numGoroutines = 10
	const iterations = 1000
	var wg sync.WaitGroup

	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				am.CheckAlerts(state)
			}
		}()

		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = am.GetAlerts()
				_ = am.HasAlerts()
			}
		}()
	}

	wg.Wait()
}

func TestAlertManagerHasAlerts(t *testing.T) {
	am := NewAlertManager()

	if am.HasAlerts() {
		t.Error("Expected no alerts initially")
	}

	state := &AppState{
		Metrics: MetricsState{
			Cpu: 85.0,
		},
		Config: ConfigState{
			Config: config.AppConfig{
				Thresholds: map[config.MetricType]float64{
					config.MetricCPU: 70.0,
				},
			},
		},
	}

	am.CheckAlerts(state)

	if !am.HasAlerts() {
		t.Error("Expected alert after CheckAlerts")
	}

	alerts := am.GetAlerts()
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].Type != config.MetricCPU {
		t.Errorf("Expected CPU alert, got %v", alerts[0].Type)
	}
}

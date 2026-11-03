//go:build !linux || !cgo || nonnvml

package system

import (
	"sync"
	"sync/atomic"

	"samouly.fun/bubbleMonitor/internal/data"
)

var (
	nvmlInitialized   bool
	nvmlInitAttempted atomic.Bool
	nvmlInitOnce      sync.Once
)

func ensureNVML() {
	// NVML not available on this platform
}

func fetchNvidiaGpus() []data.GpuInfo {
	return nil
}

//go:build !linux || !cgo || noamd

package system

import "samouly.fun/bubbleMonitor/internal/data"

func fetchAmdGpus() []data.GpuInfo {
	return nil
}

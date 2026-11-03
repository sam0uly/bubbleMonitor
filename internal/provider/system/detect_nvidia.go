//go:build linux && cgo && !nonnvml

package system

import (
	"context"
	"log"
	"os/exec"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

const nvidiaSmiTimeout = 3 * time.Second

func detectNvidia() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("NVIDIA detection panicked: %v", r)
		}
	}()

	ensureNVML()
	if nvmlInitialized {
		count, ret := nvml.DeviceGetCount()
		if ret == nvml.SUCCESS && count > 0 {
			nvidiaDetected.Store(true)
			return
		}
	}

	// nvidia-smi can hang when the driver is unresponsive; bound the wait.
	ctx, cancel := context.WithTimeout(context.Background(), nvidiaSmiTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=index", "--format=csv,noheader")
	if out, err := cmd.Output(); err == nil && len(out) > 0 {
		nvidiaDetected.Store(true)
	}
}

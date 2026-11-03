//go:build !linux || !cgo || nonnvml

package system

import (
	"log"
	"os/exec"
)

func detectNvidia() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("NVIDIA detection panicked: %v", r)
		}
	}()

	cmd := exec.Command("nvidia-smi", "--query-gpu=index", "--format=csv,noheader")
	if out, err := cmd.Output(); err == nil && len(out) > 0 {
		nvidiaDetected.Store(true)
	}
}

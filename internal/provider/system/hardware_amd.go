//go:build linux && cgo && !noamd

// `noamd` opt-out for packagers with stale amdsmi bindings; see hardware_amd_stub.go.

package system

import (
	"fmt"
	"log"

	"github.com/hhk7734/amdsmi.go"

	"samouly.fun/bubbleMonitor/internal/data"
)

func fetchAmdGpus() []data.GpuInfo {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("AMD GPU detection panicked: %v", r)
		}
	}()

	var gpuList []data.GpuInfo

	ctx := amdsmi.New()
	if err := ctx.Init(amdsmi.INIT_AMD_GPUS); err != nil {
		return gpuList
	}
	defer func() { _ = ctx.Shutdown() }()

	sockets, err := ctx.Sockets()
	if err != nil {
		return gpuList
	}

	processors := make([]*amdsmi.Processor, 0)
	for _, socket := range sockets {
		ps, err := socket.Processors()
		if err != nil {
			continue
		}
		processors = append(processors, ps...)
	}

	for i, proc := range processors {
		name := fmt.Sprintf("AMD GPU %d", i)

		vramTotal, _ := proc.GPUMemoryTotal(amdsmi.MEM_TYPE_VRAM)
		vramUsed, _ := proc.GPUMemoryUsage(amdsmi.MEM_TYPE_VRAM)
		memTotal := fmt.Sprintf("%d", vramTotal/1024/1024)
		memUsed := fmt.Sprintf("%d", vramUsed/1024/1024)

		gpuMetrics, _ := proc.GPUMetricsInfo()
		utilStr := fmt.Sprintf("%d%%", gpuMetrics.AverageGFXActivity)
		tempStr := fmt.Sprintf("%d°C", gpuMetrics.TemperatureHotspot)
		powerStr := fmt.Sprintf("%dW", gpuMetrics.CurrentSocketPower)
		clockStr := fmt.Sprintf("%d MHz", gpuMetrics.CurrentGFXCLK)

		gpuList = append(gpuList, data.GpuInfo{
			Name:        name,
			Driver:      "amdgpu",
			MemoryTotal: memTotal,
			MemoryUsed:  memUsed,
			Slot:        fmt.Sprintf("amd-%d", i),
			Type:        "dGPU",
			Vendor:      "AMD",
			Temperature: tempStr,
			PowerUsage:  powerStr,
			Utilization: utilStr,
			ClockSpeed:  clockStr,
		})
	}

	return gpuList
}

//go:build linux && cgo && !nonnvml

package system

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"

	"samouly.fun/bubbleMonitor/internal/data"
)

const nvmlInitTimeout = 3 * time.Second

var (
	nvmlInitialized   bool
	nvmlInitAttempted atomic.Bool
	nvmlInitOnce      sync.Once
)

func ensureNVML() {
	nvmlInitOnce.Do(func() {
		if nvmlInitAttempted.Load() {
			return
		}
		nvmlInitAttempted.Store(true)

		// nvml.Init() can block indefinitely when the NVIDIA driver is in a
		// broken state (e.g. after suspend/resume, or when the kernel module
		// is unresponsive). Bound the wait so a stuck driver doesn't hang
		// the entire TUI startup.
		done := make(chan nvml.Return, 1)
		go func() {
			done <- nvml.Init()
		}()

		select {
		case ret := <-done:
			if ret != nvml.SUCCESS {
				nvmlInitialized = false
			} else {
				nvmlInitialized = true
			}
		case <-time.After(nvmlInitTimeout):
			log.Printf("nvml.Init() timed out after %s; NVIDIA driver appears unresponsive", nvmlInitTimeout)
			nvmlInitialized = false
		}
	})
}

func fetchNvidiaGpus() []data.GpuInfo {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("NVIDIA GPU detection panicked: %v", r)
		}
	}()

	var gpuList []data.GpuInfo

	if !nvmlInitialized {
		return gpuList
	}

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return gpuList
	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}

		name, ret := device.GetName()
		if ret != nvml.SUCCESS {
			name = "Unknown NVIDIA GPU"
		}

		memory, ret := device.GetMemoryInfo()
		var memTotal, memUsed string
		if ret == nvml.SUCCESS {
			memTotal = fmt.Sprintf("%d", memory.Total/1024/1024)
			memUsed = fmt.Sprintf("%d", memory.Used/1024/1024)
		} else {
			memTotal = "N/A"
			memUsed = "N/A"
		}

		utilization, ret := device.GetUtilizationRates()
		var util string
		if ret == nvml.SUCCESS {
			util = fmt.Sprintf("%d%%", utilization.Gpu)
		} else {
			util = "N/A"
		}

		temperature, ret := device.GetTemperature(nvml.TEMPERATURE_GPU)
		var temp string
		if ret == nvml.SUCCESS {
			temp = fmt.Sprintf("%d°C", temperature)
		} else {
			temp = "N/A"
		}

		power, ret := device.GetPowerUsage()
		var powerStr string
		if ret == nvml.SUCCESS {
			powerStr = fmt.Sprintf("%.1fW", float64(power)/1000.0)
		} else {
			powerStr = "N/A"
		}

		clockSpeed, ret := device.GetMaxClockInfo(nvml.CLOCK_SM)
		var clock string
		if ret == nvml.SUCCESS {
			clock = fmt.Sprintf("%d MHz", clockSpeed)
		} else {
			clock = "N/A"
		}

		uuid, _ := device.GetUUID()

		gpuList = append(gpuList, data.GpuInfo{
			Name:        name,
			Driver:      "nvidia",
			MemoryTotal: memTotal,
			MemoryUsed:  memUsed,
			Slot:        uuid,
			Type:        "dGPU",
			Vendor:      "NVIDIA",
			Temperature: temp,
			PowerUsage:  powerStr,
			Utilization: util,
			ClockSpeed:  clock,
		})
	}

	return gpuList
}

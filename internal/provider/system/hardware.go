package system

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v3/host"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
)

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readFileUint64(path string) uint64 {
	data := readFile(path)
	if data == "" {
		return 0
	}
	val, err := strconv.ParseUint(data, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func HostInfoCmd() tea.Cmd {
	return func() tea.Msg {
		info, err := host.Info()
		return msg.HostInfoMsg{
			Info: info,
			Err:  err,
		}
	}
}

func GpuInfoCmd() tea.Cmd {
	return func() tea.Msg {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("GPU info command panicked: %v", r)
			}
		}()

		ensureNVML()

		var gpuList []data.GpuInfo

		if runtime.GOOS == "darwin" {
			gpuList = fetchDarwinGpus()
			if len(gpuList) > 0 {
				return msg.GpuInfoMsg{Gpus: gpuList, Err: nil}
			}
		}

		if runtime.GOOS == "linux" {
			gpuList = fetchLinuxGpus()
			if len(gpuList) > 0 {
				return msg.GpuInfoMsg{Gpus: gpuList, Err: nil}
			}
		}

		if runtime.GOOS == "windows" {
			gpuList = fetchWindowsGpus()
			if len(gpuList) > 0 {
				return msg.GpuInfoMsg{Gpus: gpuList, Err: nil}
			}
		}

		return msg.GpuInfoMsg{Gpus: gpuList, Err: nil}
	}
}

func fetchDarwinGpus() []data.GpuInfo {
	var gpuList []data.GpuInfo
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	out, err := cmd.Output()
	if err != nil {
		return gpuList
	}
	lines := strings.Split(string(out), "\n")
	var currentGpu data.GpuInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chipset Model:") {
			currentGpu = data.GpuInfo{}
			currentGpu.Name = strings.TrimPrefix(line, "Chipset Model: ")
			currentGpu.Driver = "Apple"
			currentGpu.MemoryTotal = "Shared"
			currentGpu.MemoryUsed = "N/A"
			currentGpu.Vendor = "Apple"
			currentGpu.Type = "dGPU"
		} else if strings.HasPrefix(line, "VRAM") {
			currentGpu.MemoryTotal = extractVRAM(line)
		}
		if currentGpu.Name != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && line != "" {
			if !strings.HasPrefix(line, "Chipset") && !strings.HasPrefix(line, "VRAM") && !strings.HasPrefix(line, "Displays") {
				gpuList = append(gpuList, currentGpu)
				currentGpu = data.GpuInfo{}
			}
		}
	}
	if currentGpu.Name != "" {
		gpuList = append(gpuList, currentGpu)
	}
	return gpuList
}

func extractVRAM(line string) string {
	parts := strings.Split(line, ":")
	if len(parts) > 1 {
		vram := strings.TrimSpace(parts[1])
		vram = strings.ReplaceAll(vram, "MB", "")
		vram = strings.ReplaceAll(vram, "GB", "000")
		return vram
	}
	return "N/A"
}

func fetchWindowsGpus() []data.GpuInfo {
	var gpuList []data.GpuInfo
	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "name", "/format:list")
	out, err := cmd.Output()
	if err != nil {
		return gpuList
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Name=") && !strings.Contains(line, "Win32_VideoController") {
			name := strings.TrimPrefix(line, "Name=")
			if name != "" {
				vendor := detectVendorFromName(name)
				gpuList = append(gpuList, data.GpuInfo{
					Name:        name,
					Driver:      "Windows",
					MemoryTotal: "N/A",
					MemoryUsed:  "N/A",
					Vendor:      vendor,
					Type:        "dGPU",
				})
			}
		}
	}
	return gpuList
}

func fetchLinuxGpus() []data.GpuInfo {
	var gpuList []data.GpuInfo
	seen := make(map[string]bool)
	nvidiaSeen := false
	amdSeen := false

	// Try NVML first (faster), fallback to nvidia-smi if needed
	if nvmlInitialized {
		nvidiaGpus := fetchNvidiaGpus()
		gpuList = append(gpuList, nvidiaGpus...)
		nvidiaSeen = len(nvidiaGpus) > 0
		for _, g := range nvidiaGpus {
			seen[g.Slot] = true
		}
	}

	// Fallback to nvidia-smi if NVML returned nothing
	if !nvidiaSeen {
		nvidiaSmiGpus := fetchNvidiaSmiGpus()
		for _, g := range nvidiaSmiGpus {
			if !seen[g.Slot] {
				gpuList = append(gpuList, g)
				seen[g.Slot] = true
			}
		}
		nvidiaSeen = len(nvidiaSmiGpus) > 0
	}

	// Always detect AMD (only if no AMD detected yet)
	amdGpus := fetchAmdGpus()
	for _, g := range amdGpus {
		if !seen[g.Slot] {
			gpuList = append(gpuList, g)
			seen[g.Slot] = true
		}
	}
	amdSeen = len(amdGpus) > 0

	// Always detect sysfs GPUs (Intel iGPU, etc). Skip NVIDIA/AMD that
	// were already found via NVML/nvidia-smi to avoid duplicates.
	sysfsGpus := fetchSysfsGpus()
	for _, g := range sysfsGpus {
		if (g.Vendor == "NVIDIA" && nvidiaSeen) || (g.Vendor == "AMD" && amdSeen) {
			continue
		}
		if !seen[g.Slot] {
			gpuList = append(gpuList, g)
			seen[g.Slot] = true
		}
	}

	if len(gpuList) > 1 {
		sort.Slice(gpuList, func(i, j int) bool {
			if gpuList[i].Type == "dGPU" && gpuList[j].Type != "dGPU" {
				return true
			}
			if gpuList[i].Type != "dGPU" && gpuList[j].Type == "dGPU" {
				return false
			}
			return gpuList[i].Name < gpuList[j].Name
		})
	}

	return gpuList
}

func fetchNvidiaSmiGpus() []data.GpuInfo {
	var gpuList []data.GpuInfo

	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,pci.bus_id,memory.total,memory.used,utilization.gpu,temperature.gpu,power.draw,clocks.sm", "--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return gpuList
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 9 {
			continue
		}

		var gpu data.GpuInfo
		gpu.Name = strings.TrimSpace(parts[1])
		gpu.Driver = "nvidia"
		gpu.Slot = strings.TrimSpace(parts[2])
		gpu.Type = "dGPU"
		gpu.Vendor = "NVIDIA"

		memTotal := strings.TrimSpace(parts[3])
		if memTotal != "N/A" && memTotal != "" {
			gpu.MemoryTotal = memTotal
		}

		memUsed := strings.TrimSpace(parts[4])
		if memUsed != "N/A" && memUsed != "" {
			gpu.MemoryUsed = memUsed
		}

		util := strings.TrimSpace(parts[5])
		if util != "N/A" && util != "" {
			gpu.Utilization = util + "%"
		}

		temp := strings.TrimSpace(parts[6])
		if temp != "N/A" && temp != "" {
			gpu.Temperature = temp + "C"
		}

		power := strings.TrimSpace(parts[7])
		if power != "N/A" && power != "" {
			gpu.PowerUsage = power + "W"
		}

		clock := strings.TrimSpace(parts[8])
		if clock != "N/A" && clock != "" {
			gpu.ClockSpeed = clock + " MHz"
		}

		if gpu.Name != "" {
			gpuList = append(gpuList, gpu)
		}
	}

	return gpuList
}

func fetchSysfsGpus() []data.GpuInfo {
	var gpuList []data.GpuInfo
	seen := make(map[string]bool)

	drmDevices, _ := os.ReadDir("/sys/class/drm")
	for _, d := range drmDevices {
		name := d.Name()
		if !strings.HasPrefix(name, "card") || strings.Contains(name, "-") {
			continue
		}

		devicePath := "/sys/class/drm/" + name + "/device"

		uevent := readFile(devicePath + "/uevent")
		if uevent == "" {
			continue
		}

		var vendorID, deviceID, driver string

		for _, line := range strings.Split(uevent, "\n") {
			if strings.HasPrefix(line, "PCI_ID=") {
				pciIDStr := strings.TrimPrefix(line, "PCI_ID=")
				parts := strings.Split(pciIDStr, ":")
				if len(parts) == 2 {
					vendorID = strings.ToLower(parts[0])
					deviceID = strings.ToLower(parts[1])
				}
			} else if strings.HasPrefix(line, "DRIVER=") {
				driver = strings.TrimPrefix(line, "DRIVER=")
			}
		}

		pciID := vendorID + ":" + deviceID
		if pciID == ":" || pciID == "" {
			continue
		}

		if seen[pciID] {
			continue
		}
		seen[pciID] = true

		gpuType := "iGPU"
		vendor := "Unknown"
		if vendorID == "10de" {
			gpuType = "dGPU"
			vendor = "NVIDIA"
		} else if vendorID == "1002" {
			gpuType = "dGPU"
			vendor = "AMD"
		} else if vendorID == "8086" {
			gpuType = "iGPU"
			vendor = "Intel"
		}

		if vendor == "Unknown" && driver != "" {
			if driver == "nvidia" {
				vendor = "NVIDIA"
				gpuType = "dGPU"
			} else if driver == "amdgpu" || driver == "radeon" {
				vendor = "AMD"
				gpuType = "dGPU"
			} else if driver == "i915" {
				vendor = "Intel"
				gpuType = "iGPU"
			}
		}

		if vendor == "Unknown" {
			continue
		}

		if vendor == "Intel" {
			MarkIntelDetected()
		}

		var memInfo uint64
		memPath := devicePath + "/mem_info_vram_total"
		if data := readFileUint64(memPath); data > 0 {
			memInfo = data / (1024 * 1024)
		}

		if memInfo == 0 {
			meminfo := readFile(devicePath + "/meminfo")
			for _, line := range strings.Split(meminfo, "\n") {
				if strings.Contains(line, "size:") {
					parts := strings.Fields(line)
					for i, p := range parts {
						if p == "size:" && i+1 < len(parts) {
							if val, err := strconv.ParseUint(parts[i+1], 10, 64); err == nil {
								memInfo = val / 1024
							}
							break
						}
					}
				}
			}
		}

		chipInfo := readFile(devicePath + "/chip_info")
		productName := readFile(devicePath + "/product_name")
		productVersion := readFile(devicePath + "/product_version")

		gpuName := productName
		if gpuName == "" {
			if chipInfo != "" {
				gpuName = chipInfo
			} else {
				gpuName = getPciDeviceName(vendorID, deviceID)
			}
		}

		if gpuName == "" {
			gpuName = "Unknown GPU"
		}

		if productVersion != "" {
			gpuName += " " + productVersion
		}

		memTotal := "N/A"
		if memInfo > 0 {
			memTotal = fmt.Sprintf("%d", memInfo)
		} else if vram := getVramFromPciDb(pciID); vram > 0 {
			memTotal = fmt.Sprintf("%d", vram)
		}

		var temp, util, power, clock string
		if vendorID == "8086" {
			temp = readIntelTemperature(devicePath)
			clock = readIntelFrequency(devicePath)
		}

		gpuDriver := driver
		if gpuDriver == "" {
			gpuDriver = vendor + " (sysfs)"
		}

		gpuList = append(gpuList, data.GpuInfo{
			Name:        gpuName,
			Driver:      gpuDriver,
			MemoryTotal: memTotal,
			MemoryUsed:  "N/A",
			Slot:        pciID,
			Type:        gpuType,
			Vendor:      vendor,
			Temperature: temp,
			ClockSpeed:  clock,
			Utilization: util,
			PowerUsage:  power,
		})
	}

	return gpuList
}

func readIntelTemperature(devicePath string) string {
	tempPaths := []string{
		devicePath + "/hwmon/hwmon*/temp1_input",
		devicePath + "/device/thermal_zone/temp",
	}
	for _, path := range tempPaths {
		if temp := readFileUint64(path); temp > 0 {
			return fmt.Sprintf("%d°C", temp/1000)
		}
	}
	return "N/A"
}

func readIntelFrequency(devicePath string) string {
	freqPaths := []string{
		devicePath + "/gt/gt0/punit/gpu_freq_mhz",
		devicePath + "/freq0/freq",
	}
	for _, path := range freqPaths {
		if freq := readFileUint64(path); freq > 0 {
			return fmt.Sprintf("%d MHz", freq)
		}
	}
	return "N/A"
}

func getPciDeviceName(vendorID, deviceID string) string {
	pciID := strings.ToLower(vendorID + ":" + deviceID)

	pciNames := map[string]string{
		"10de:1f36": "NVIDIA Quadro RTX 3000",
		"10de:1f14": "NVIDIA RTX A5000",
		"10de:1f10": "NVIDIA RTX A4000",
		"10de:1f06": "NVIDIA RTX A3000",
		"10de:1f00": "NVIDIA RTX A2000",
		"10de:1b80": "NVIDIA GTX 1080",
		"10de:1b81": "NVIDIA GTX 1070",
		"10de:1b82": "NVIDIA GTX 1070 Ti",
		"10de:1b83": "NVIDIA GTX 1060",
		"10de:1c03": "NVIDIA GTX 1050 Ti",
		"10de:1c02": "NVIDIA GTX 1050",
		"10de:1c80": "NVIDIA MX450",
		"10de:1c81": "NVIDIA MX350",
		"10de:1d12": "NVIDIA T1000",
		"10de:1d10": "NVIDIA T600",
		"10de:1d11": "NVIDIA T400",
		"10de:13d7": "NVIDIA Quadro K2200",
		"10de:13d0": "NVIDIA Quadro K4200",
		"1002:687f": "AMD Radeon RX Vega 64",
		"1002:6870": "AMD Radeon RX Vega 56",
		"1002:731f": "AMD Radeon RX 6900 XT",
		"1002:731e": "AMD Radeon RX 6800 XT",
		"1002:731d": "AMD Radeon RX 6800",
		"1002:73ff": "AMD Radeon RX 6700 XT",
		"1002:73df": "AMD Radeon RX 6600 XT",
		"1002:73e1": "AMD Radeon RX 6600",
		"1002:67df": "AMD Radeon RX 480",
		"1002:67c7": "AMD Radeon RX 470",
		"1002:67e8": "AMD Radeon RX 560",
		"1002:67e0": "AMD Radeon RX 550",
		"1002:15dd": "AMD Radeon RX Vega 11",
		"1002:15dc": "AMD Radeon RX Vega 8",
		"8086:5912": "Intel HD Graphics 630",
		"8086:591b": "Intel UHD Graphics 620",
		"8086:3ea5": "Intel UHD Graphics 617",
		"8086:3e9b": "Intel UHD Graphics 605",
		"8086:3e91": "Intel UHD Graphics 600",
		"8086:8a70": "Intel Iris Plus Graphics 950",
		"8086:8a60": "Intel Iris Plus Graphics 645",
		"8086:8a50": "Intel Iris Plus Graphics 640",
		"8086:9a49": "Intel UHD Graphics P750",
		"8086:9a60": "Intel UHD Graphics",
	}

	if name, ok := pciNames[pciID]; ok {
		return name
	}

	return ""
}

// pciVRAM maps a "vendor:device" PCI ID to its reference VRAM in MiB.
// Lookups are direct, O(1), and don't depend on the human-readable name.
var pciVRAM = map[string]int{
	"10de:1f36": 6144,
	"10de:1f14": 16384,
	"10de:1f10": 8192,
	"10de:1f06": 6144,
	"10de:1f00": 4096,
	"10de:1b80": 8192,
	"10de:1b81": 8192,
	"10de:1b82": 8192,
	"10de:1b83": 6144,
	"10de:1c03": 4096,
	"10de:1c02": 2048,
	"10de:1c80": 2048,
	"10de:1c81": 2048,
	"10de:1d12": 4096,
	"10de:1d10": 4096,
	"10de:1d11": 4096,
	"10de:13d7": 4096,
	"10de:13d0": 4096,
	"1002:687f": 16384,
	"1002:6870": 8192,
	"1002:731f": 16384,
	"1002:731e": 16384,
	"1002:731d": 16384,
	"1002:73ff": 12288,
	"1002:73df": 8192,
	"1002:73e1": 8192,
	"1002:67df": 8192,
	"1002:67c7": 4096,
	"1002:67e8": 4096,
	"1002:67e0": 2048,
}

// getVramFromPciDb looks up reference VRAM by PCI ID. Returns 0 if unknown.
func getVramFromPciDb(pciID string) int {
	return pciVRAM[strings.ToLower(pciID)]
}

func detectVendorFromName(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "nvidia") || strings.Contains(lower, "geforce") || strings.Contains(lower, "quadro") {
		return "NVIDIA"
	}
	if strings.Contains(lower, "amd") || strings.Contains(lower, "radeon") {
		return "AMD"
	}
	if strings.Contains(lower, "intel") || strings.Contains(lower, "uhd") || strings.Contains(lower, "iris") {
		return "Intel"
	}
	return "Unknown"
}

func TempCmd() tea.Cmd {
	return func() tea.Msg {
		temps, err := host.SensorsTemperatures()
		if err != nil {
			return msg.TempMsg{}
		}
		return msg.TempMsg(temps)
	}
}

func BatteryCmd() tea.Cmd {
	return func() tea.Msg {
		batt, err := battery.GetAll()
		if err != nil {
			return msg.BatteryMsg{}
		}
		return msg.BatteryMsg(batt)
	}
}

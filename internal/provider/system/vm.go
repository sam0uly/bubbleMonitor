package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
)

const (
	HypervisorNone       = ""
	HypervisorVMware     = "VMware"
	HypervisorVirtualBox = "VirtualBox"
	HypervisorKVM        = "KVM"
	HypervisorHyperV     = "Hyper-V"
	HypervisorQEMU       = "QEMU"
	HypervisorXen        = "Xen"
	HypervisorParallels  = "Parallels"
	HypervisorBhyve      = "Bhyve"
	HypervisorOpenVZ     = "OpenVZ"
	HypervisorLXC        = "LXC"
	HypervisorUnknown    = "Unknown"
)

var (
	isVM           atomic.Bool
	hypervisorType atomic.Value
	vmCheckOnce    sync.Once
)

type vmDetection struct {
	hypervisor string
}

func detectVM() {
	vmCheckOnce.Do(func() {
		if runtime.GOOS != "linux" {
			isVM.Store(detectWindowsVM())
			if isVM.Load() {
				hypervisorType.Store(HypervisorHyperV)
			}
			return
		}

		detection := detectLinuxVM()
		isVM.Store(detection.hypervisor != HypervisorNone)
		if isVM.Load() {
			hypervisorType.Store(detection.hypervisor)
		}
	})
}

func detectLinuxVM() vmDetection {
	if flags := detectCPUFlags(); flags.hypervisor != "" {
		return flags
	}

	if dmi := detectDMI(); dmi.hypervisor != "" {
		return dmi
	}

	if dmidecode := detectDmidecode(); dmidecode.hypervisor != "" {
		return dmidecode
	}

	return vmDetection{}
}

func detectCPUFlags() vmDetection {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return vmDetection{}
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "flags") || strings.HasPrefix(line, "Features") {
			flags := strings.ToLower(line)

			switch {
			case strings.Contains(flags, "hyperv") || strings.Contains(flags, "ms_hyperv"):
				return vmDetection{hypervisor: HypervisorHyperV}
			case strings.Contains(flags, "vmware"):
				return vmDetection{hypervisor: HypervisorVMware}
			case strings.Contains(flags, "kvm"):
				return vmDetection{hypervisor: HypervisorKVM}
			case strings.Contains(flags, "qemu"):
				return vmDetection{hypervisor: HypervisorQEMU}
			case strings.Contains(flags, "xen"):
				return vmDetection{hypervisor: HypervisorXen}
			case strings.Contains(flags, "parallels"):
				return vmDetection{hypervisor: HypervisorParallels}
			case strings.Contains(flags, "bhyve"):
				return vmDetection{hypervisor: HypervisorBhyve}
			case strings.Contains(flags, "openvz"):
				return vmDetection{hypervisor: HypervisorOpenVZ}
			case strings.Contains(flags, "virtio"):
				return vmDetection{hypervisor: HypervisorKVM}
			}
			break
		}
	}

	return vmDetection{}
}

func detectDMI() vmDetection {
	productFiles := []string{
		"/sys/class/dmi/id/product_name",
		"/sys/class/dmi/id/sys_vendor",
		"/sys/class/dmi/id/board_vendor",
		"/sys/class/dmi/id/bios_vendor",
	}

	for _, file := range productFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := strings.ToLower(strings.TrimSpace(string(data)))

		switch {
		case strings.Contains(content, "vmware"):
			return vmDetection{hypervisor: HypervisorVMware}
		case strings.Contains(content, "virtualbox"):
			return vmDetection{hypervisor: HypervisorVirtualBox}
		case strings.Contains(content, "kvm"):
			return vmDetection{hypervisor: HypervisorKVM}
		case strings.Contains(content, "qemu"):
			return vmDetection{hypervisor: HypervisorQEMU}
		case strings.Contains(content, "hyper-v") || strings.Contains(content, "microsoft corporation"):
			return vmDetection{hypervisor: HypervisorHyperV}
		case strings.Contains(content, "xen"):
			return vmDetection{hypervisor: HypervisorXen}
		case strings.Contains(content, "parallels"):
			return vmDetection{hypervisor: HypervisorParallels}
		case strings.Contains(content, "bhyve"):
			return vmDetection{hypervisor: HypervisorBhyve}
		case strings.Contains(content, "openvz"):
			return vmDetection{hypervisor: HypervisorOpenVZ}
		}
	}

	return vmDetection{}
}

func detectDmidecode() vmDetection {
	cmd := exec.Command("dmidecode", "-s", "system-product-name")
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("dmidecode", "-s", "system-manufacturer")
		out, err = cmd.Output()
		if err != nil {
			return vmDetection{}
		}
	}

	content := strings.ToLower(strings.TrimSpace(string(out)))

	switch {
	case strings.Contains(content, "vmware"):
		return vmDetection{hypervisor: HypervisorVMware}
	case strings.Contains(content, "virtualbox"):
		return vmDetection{hypervisor: HypervisorVirtualBox}
	case strings.Contains(content, "kvm"):
		return vmDetection{hypervisor: HypervisorKVM}
	case strings.Contains(content, "qemu"):
		return vmDetection{hypervisor: HypervisorQEMU}
	case strings.Contains(content, "hyper-v") || strings.Contains(content, "microsoft"):
		return vmDetection{hypervisor: HypervisorHyperV}
	case strings.Contains(content, "xen"):
		return vmDetection{hypervisor: HypervisorXen}
	case strings.Contains(content, "parallels"):
		return vmDetection{hypervisor: HypervisorParallels}
	case strings.Contains(content, "bhyve"):
		return vmDetection{hypervisor: HypervisorBhyve}
	}

	return vmDetection{}
}

func detectWindowsVM() bool {
	productName, err := exec.Command("powershell", "-Command",
		"(Get-WmiObject Win32_ComputerSystem).Manufacturer").Output()
	if err == nil {
		content := strings.ToLower(string(productName))
		if strings.Contains(content, "microsoft") || strings.Contains(content, "vmware") ||
			strings.Contains(content, "virtualbox") || strings.Contains(content, "qemu") {
			return true
		}
	}

	hypervPresent, err := exec.Command("powershell", "-Command",
		"Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V -Online").Output()
	if err == nil && strings.Contains(string(hypervPresent), "Enabled") {
		return true
	}

	return false
}

func IsVM() bool {
	detectVM()
	return isVM.Load()
}

func GetHypervisorType() string {
	detectVM()
	if h, ok := hypervisorType.Load().(string); ok {
		return h
	}
	return HypervisorNone
}

func VmCmd() tea.Cmd {
	return func() tea.Msg {
		detectVM()

		if !isVM.Load() {
			return msg.VmInfoMsg{
				Err:    nil,
				VmInfo: nil,
				IsVM:   false,
			}
		}

		vmInfo := &data.VmInfo{
			Hypervisor: GetHypervisorType(),
		}

		if hostInfo, err := host.Info(); err == nil {
			vmInfo.HostName = hostInfo.Hostname
		}

		if runtime.GOOS == "linux" {
			vmInfo.Type = detectVMLinuxType()
			vmInfo.VirtCPU = detectVirtCPU()
			vmInfo.VirtCPUUsed = detectVirtCPUUsed()
		}

		if vmStats, err := mem.VirtualMemory(); err == nil {
			vmInfo.VirtMemory = vmStats.Total
			vmInfo.VirtMemoryUsed = vmStats.Used
		}

		return msg.VmInfoMsg{
			Err:    nil,
			VmInfo: vmInfo,
			IsVM:   true,
		}
	}
}

func detectVMLinuxType() string {
	if _, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		data, _ := os.ReadFile("/proc/1/cgroup")
		content := strings.ToLower(string(data))
		if strings.Contains(content, "docker") || strings.Contains(content, "container") {
			return "Container"
		}
	}

	if _, err := os.ReadFile("/.dockerenv"); err == nil {
		return "Docker"
	}

	if _, err := os.ReadFile("/run/.containerenv"); err == nil {
		return "Container"
	}

	if exists("/usr/bin/lxc-checkconfig") || exists("/sys/fs/cgroup/cpuset/lxc") {
		return "LXC"
	}

	return "VM"
}

func detectVirtCPU() int {
	if runtime.GOOS != "linux" {
		return 0
	}

	possiblePaths := []string{
		"/sys/devices/system/cpu/possible",
		"/sys/fs/cgroup/cpuset/cpuset.cpus",
	}

	for _, path := range possiblePaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		content := strings.TrimSpace(string(data))

		if strings.Contains(content, "-") {
			parts := strings.Split(content, "-")
			if len(parts) == 2 {
				var start, end int
				if _, err := parseStrToInt(parts[0], &start); err == nil {
					if _, err := parseStrToInt(parts[1], &end); err == nil {
						return end - start + 1
					}
				}
			}
		}

		cpus := strings.Split(content, ",")
		return len(cpus)
	}

	return runtime.NumCPU()
}

func detectVirtCPUUsed() int {
	if runtime.GOOS != "linux" {
		return 0
	}

	paths := []string{
		"/sys/fs/cgroup/cpuset/cpuset.effective_cpus",
		"/sys/fs/cgroup/cpuset/cpuset.cpus",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		content := strings.TrimSpace(string(data))

		if strings.Contains(content, ",") {
			cpus := strings.Split(content, ",")
			return len(cpus)
		}

		if strings.Contains(content, "-") {
			parts := strings.Split(content, "-")
			if len(parts) == 2 {
				var start, end int
				if _, err := parseStrToInt(parts[0], &start); err == nil {
					if _, err := parseStrToInt(parts[1], &end); err == nil {
						return end - start + 1
					}
				}
			}
		}
	}

	return runtime.NumCPU()
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func parseStrToInt(s string, result *int) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("non-numeric character %q in %q", string(c), s)
		}
		n = n*10 + int(c-'0')
	}
	*result = n
	return n, nil
}

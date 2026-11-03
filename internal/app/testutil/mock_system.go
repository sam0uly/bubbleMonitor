package testutil

import (
	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"samouly.fun/bubbleMonitor/internal/data"
)

type MockSystemProvider struct {
	CpuFunc      func() (float64, []float64, error)
	MemoryFunc   func() (*mem.VirtualMemoryStat, *mem.SwapMemoryStat, error)
	LoadAvgFunc  func() (*load.AvgStat, error)
	DiskFunc     func() ([]data.DiskPartition, float64, error)
	NetworkFunc  func() ([]net.IOCountersStat, uint64, uint64, error)
	TempFunc     func() ([]host.TemperatureStat, error)
	BatteryFunc  func() ([]*battery.Battery, error)
	HostInfoFunc func() (*host.InfoStat, error)
	GpuFunc      func() ([]data.GpuInfo, error)
	DiskIOFunc   func() (map[string]disk.IOCountersStat, error)
}

func (m *MockSystemProvider) Cpu() (float64, []float64, error) {
	if m.CpuFunc != nil {
		return m.CpuFunc()
	}
	cpuVal, _, _ := MockCpuMem()
	return cpuVal, MockCpuPerCore(), nil
}

func (m *MockSystemProvider) Memory() (*mem.VirtualMemoryStat, *mem.SwapMemoryStat, error) {
	if m.MemoryFunc != nil {
		return m.MemoryFunc()
	}
	_, memInfo, swapInfo := MockCpuMem()
	return memInfo, swapInfo, nil
}

func (m *MockSystemProvider) LoadAvg() (*load.AvgStat, error) {
	if m.LoadAvgFunc != nil {
		return m.LoadAvgFunc()
	}
	return &load.AvgStat{Load1: 0.5, Load5: 0.3, Load15: 0.2}, nil
}

func (m *MockSystemProvider) Disk() ([]data.DiskPartition, float64, error) {
	if m.DiskFunc != nil {
		return m.DiskFunc()
	}
	return MockDisk(), 45.0, nil
}

func (m *MockSystemProvider) Network() ([]net.IOCountersStat, uint64, uint64, error) {
	if m.NetworkFunc != nil {
		return m.NetworkFunc()
	}
	return MockNetwork()
}

func (m *MockSystemProvider) Temp() ([]host.TemperatureStat, error) {
	if m.TempFunc != nil {
		return m.TempFunc()
	}
	return []host.TemperatureStat{
		{SensorKey: "cpu", Temperature: 65.0},
		{SensorKey: "core_0", Temperature: 60.0},
	}, nil
}

func (m *MockSystemProvider) Battery() ([]*battery.Battery, error) {
	if m.BatteryFunc != nil {
		return m.BatteryFunc()
	}
	return []*battery.Battery{}, nil
}

func (m *MockSystemProvider) HostInfo() (*host.InfoStat, error) {
	if m.HostInfoFunc != nil {
		return m.HostInfoFunc()
	}
	return &host.InfoStat{
		Hostname:        "testhost",
		OS:              "linux",
		Platform:        "ubuntu",
		PlatformVersion: "22.04",
		KernelVersion:   "5.15.0",
	}, nil
}

func (m *MockSystemProvider) GpuInfo() ([]data.GpuInfo, error) {
	if m.GpuFunc != nil {
		return m.GpuFunc()
	}
	return []data.GpuInfo{}, nil
}

func (m *MockSystemProvider) DiskIO() (map[string]disk.IOCountersStat, error) {
	if m.DiskIOFunc != nil {
		return m.DiskIOFunc()
	}
	return map[string]disk.IOCountersStat{
		"sda": {ReadBytes: 1000, WriteBytes: 500},
	}, nil
}

func MockCpuMem() (float64, *mem.VirtualMemoryStat, *mem.SwapMemoryStat) {
	cpuVal := 25.5
	memInfo := &mem.VirtualMemoryStat{
		Total:       16e9,
		Used:        8e9,
		UsedPercent: 50.0,
		Free:        8e9,
	}
	swapInfo := &mem.SwapMemoryStat{
		Total:       8e9,
		Used:        1e9,
		UsedPercent: 12.5,
		Free:        7e9,
	}
	return cpuVal, memInfo, swapInfo
}

func MockCpuPerCore() []float64 {
	return []float64{20.0, 30.0, 25.0, 27.0, 22.0, 28.0, 24.0, 26.0}
}

func MockProcesses() []data.ProcessInfo {
	return []data.ProcessInfo{
		{
			Pid:         1,
			Name:        "init",
			NameLower:   "init",
			Status:      "running",
			Cpu:         0.1,
			Memory:      0.2,
			MemoryBytes: 1024 * 1024,
			Nice:        0,
			Ppid:        0,
		},
		{
			Pid:         2,
			Name:        "systemd",
			NameLower:   "systemd",
			Status:      "running",
			Cpu:         0.5,
			Memory:      1.5,
			MemoryBytes: 5 * 1024 * 1024,
			Nice:        0,
			Ppid:        1,
		},
		{
			Pid:         100,
			Name:        "testproc",
			NameLower:   "testproc",
			Status:      "running",
			Cpu:         5.0,
			Memory:      2.0,
			MemoryBytes: 8 * 1024 * 1024,
			Nice:        0,
			Ppid:        2,
		},
	}
}

func MockDisk() []data.DiskPartition {
	return []data.DiskPartition{
		{
			Mountpoint: "/",
			Device:     "/dev/sda1",
			Fstype:     "ext4",
			Total:      500e9,
			Used:       250e9,
			UsedPct:    50.0,
		},
		{
			Mountpoint: "/home",
			Device:     "/dev/sda2",
			Fstype:     "ext4",
			Total:      1000e9,
			Used:       300e9,
			UsedPct:    30.0,
		},
	}
}

func MockNetwork() ([]net.IOCountersStat, uint64, uint64, error) {
	ifaces := []net.IOCountersStat{
		{Name: "eth0", BytesSent: 1000, BytesRecv: 2000},
		{Name: "lo", BytesSent: 100, BytesRecv: 100},
	}
	return ifaces, 1000, 2000, nil
}

func MockTemp() []host.TemperatureStat {
	return []host.TemperatureStat{
		{SensorKey: "cpu", Temperature: 65.0},
		{SensorKey: "core_0", Temperature: 60.0},
		{SensorKey: "core_1", Temperature: 62.0},
	}
}

func MockBattery() []*battery.Battery {
	return []*battery.Battery{}
}

func MockHostInfo() *host.InfoStat {
	return &host.InfoStat{
		Hostname:        "testhost",
		OS:              "linux",
		Platform:        "ubuntu",
		PlatformVersion: "22.04",
		KernelVersion:   "5.15.0",
	}
}

func MockGpuInfo() []data.GpuInfo {
	return []data.GpuInfo{
		{
			Name:        "NVIDIA RTX 3080",
			Driver:      "525.0",
			MemoryTotal: "10GB",
			MemoryUsed:  "5GB",
			Type:        "nvidia",
			Vendor:      "NVIDIA",
		},
	}
}

func MockCpuInfo() []cpu.InfoStat {
	return []cpu.InfoStat{
		{
			CPU:        0,
			VendorID:   "GenuineIntel",
			Family:     "6",
			Model:      "142",
			ModelName:  "Intel(R) Core(TM) i7-10870H CPU @ 2.20GHz",
			Cores:      8,
			PhysicalID: "0",
			Mhz:        2200,
		},
	}
}

package system

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
)

// DiskInfoCmd fetches disk partition information from two sources:
//  1. gopsutil's disk.Partitions() — mounted filesystems with usage data.
//  2. ghw's block.New() — every block device and partition on the host,
//     including unmounted ones, swap partitions, and Windows/LUKS volumes.
//
// The two sources are merged into a single list: gopsutil entries win for
// partitions it already saw (because we have real usage data for them);
// ghw adds the partitions gopsutil didn't see. Swap usage is then enriched
// from mem.SwapMemory() because neither source reports swap usage directly
// on Linux.
func DiskInfoCmd() tea.Cmd {
	return func() tea.Msg {
		partitions, err := disk.Partitions(false)
		var diskList []data.DiskPartition

		seen := make(map[string]bool)
		for _, p := range partitions {
			if p.Mountpoint == "" {
				continue
			}
			seen[p.Device] = true
			usage, usageErr := disk.Usage(p.Mountpoint)
			entry := data.DiskPartition{
				Mountpoint: p.Mountpoint,
				Device:     p.Device,
				Fstype:     p.Fstype,
				Kind:       classifyKind(p.Fstype, p.Mountpoint),
			}
			if usageErr == nil {
				entry.Total = usage.Total
				entry.Used = usage.Used
				entry.UsedPct = usage.UsedPercent
			} else {
				entry.UsedPct = -1
			}
			diskList = append(diskList, entry)
		}

		// Add partitions ghw discovered but gopsutil didn't report.
		// ghw works without root on Linux/macOS/Windows and reads partition
		// tables directly — no external binaries like lsblk needed.
		enumerateBlockDevices(&diskList, seen)

		// Neither gopsutil nor ghw reports real swap usage on Linux, so
		// fill it in from the OS-level swap stats.
		enrichSwapFromOS(&diskList)

		return msg.DiskInfoMsg{
			Partitions: diskList,
			Err:        err,
		}
	}
}

// classifyKind returns a Kind for a partition based on its data:
//
//	"swap"   – fstype or mountpoint identifies it as swap
//	"part"   – has a filesystem (or no fs) but isn't actively mounted
//	"mounted" – has a real filesystem mountpoint
func classifyKind(fstype, mountpoint string) string {
	if fstype == "swap" || mountpoint == "[swap]" || mountpoint == "[SWAP]" {
		return "swap"
	}
	if mountpoint == "" {
		return "part"
	}
	return "mounted"
}

// enumerateBlockDevices walks ghw's block device tree and appends partitions
// that gopsutil didn't already report. Parent block devices (sda, nvme0n1)
// are skipped — only mountable partitions (ghw.Partition) make it in.
func enumerateBlockDevices(diskList *[]data.DiskPartition, seen map[string]bool) {
	info, err := ghw.Block()
	if err != nil {
		return
	}
	for _, disk := range info.Disks {
		for _, part := range disk.Partitions {
			if part == nil || part.SizeBytes == 0 {
				continue
			}
			devicePath := devicePathFor(part.Name)
			if seen[devicePath] {
				continue
			}
			mountpoint := part.MountPoint
			if mountpoint == "[SWAP]" {
				mountpoint = "[swap]"
			}
			*diskList = append(*diskList, data.DiskPartition{
				Mountpoint: mountpoint,
				Device:     devicePath,
				Fstype:     part.Type,
				Total:      part.SizeBytes,
				Used:       0,
				UsedPct:    -1,
				Kind:       classifyKind(part.Type, mountpoint),
			})
		}
	}
}

// devicePathFor returns the full device path for a partition name. ghw reports
// names as the basename (e.g. "sda1", "nvme0n1p2"); the host exposes the
// partition under /dev/. This helper is a no-op when the name already
// includes a path component.
func devicePathFor(name string) string {
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, "/dev/") {
		return name
	}
	return "/dev/" + name
}

// enrichSwapFromOS populates Used/UsedPct on any swap partition using the
// OS-level swap stats, since gopsutil/ghw don't report swap usage on Linux.
func enrichSwapFromOS(diskList *[]data.DiskPartition) {
	swap, err := mem.SwapMemory()
	if err != nil || swap.Total == 0 {
		return
	}
	for i := range *diskList {
		dp := &(*diskList)[i]
		if dp.Kind != "swap" || dp.UsedPct >= 0 {
			continue
		}
		dp.Used = swap.Used
		dp.UsedPct = swap.UsedPercent
	}
}

// DiskIOCmd fetches disk I/O statistics
func DiskIOCmd() tea.Cmd {
	return func() tea.Msg {
		ioCounters, err := disk.IOCounters()
		if err != nil {
			return msg.DiskIOMsg{}
		}
		return msg.DiskIOMsg(ioCounters)
	}
}

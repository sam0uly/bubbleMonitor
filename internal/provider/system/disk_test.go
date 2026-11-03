package system

import (
	"testing"

	"samouly.fun/bubbleMonitor/internal/data"
)

func TestClassifyKind(t *testing.T) {
	cases := []struct {
		fstype, mountpoint, want string
	}{
		{"ext4", "/", "mounted"},
		{"vfat", "/boot", "mounted"},
		{"swap", "[swap]", "swap"},
		{"swap", "", "swap"}, // swap with no active mountpoint (rare)
		{"ext4", "", "part"},
		{"BitLocker", "", "part"},
		{"ntfs", "", "part"},
		{"", "", "part"},
	}
	for _, c := range cases {
		if got := classifyKind(c.fstype, c.mountpoint); got != c.want {
			t.Errorf("classifyKind(%q, %q) = %q, want %q", c.fstype, c.mountpoint, got, c.want)
		}
	}
}

func TestDevicePathFor(t *testing.T) {
	cases := []struct {
		name, want string
	}{
		{"sda1", "/dev/sda1"},
		{"nvme0n1p2", "/dev/nvme0n1p2"},
		{"/dev/sda1", "/dev/sda1"},
		{"/dev/disk0s1", "/dev/disk0s1"},
		{"", ""},
	}
	for _, c := range cases {
		if got := devicePathFor(c.name); got != c.want {
			t.Errorf("devicePathFor(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

// Integration test: enumerateBlockDevices against the real host must
// (a) skip zero-size partitions, (b) skip parent block devices, and
// (c) skip anything already in the seen map (gopsutil-reported).
func TestEnumerateBlockDevices_RespectsSeenAndSkipsParents(t *testing.T) {
	var list []data.DiskPartition
	seen := map[string]bool{
		"/dev/sda1": true,
		"/dev/sda2": true,
	}
	enumerateBlockDevices(&list, seen)

	for _, dp := range list {
		if dp.Device == "/dev/sda1" || dp.Device == "/dev/sda2" {
			t.Errorf("seen device %s re-added by enumerateBlockDevices", dp.Device)
		}
		if dp.Kind == "disk" {
			t.Errorf("parent block device %s leaked (Kind=disk)", dp.Device)
		}
		if dp.Total == 0 {
			t.Errorf("partition %s has Total=0 (should be filtered)", dp.Device)
		}
	}
}

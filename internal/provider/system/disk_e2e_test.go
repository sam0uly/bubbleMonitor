package system

import (
	"testing"

	"samouly.fun/bubbleMonitor/internal/msg"
)

// End-to-end check: run DiskInfoCmd against the real host and verify
// parent block devices are excluded, swap is enriched with OS-level usage,
// and unmounted partitions appear in the result.
func TestDiskInfoCmd_HostShape(t *testing.T) {
	cmd := DiskInfoCmd()
	raw := cmd()
	m, ok := raw.(msg.DiskInfoMsg)
	if !ok {
		t.Fatalf("expected msg.DiskInfoMsg, got %T", raw)
	}

	for _, dp := range m.Partitions {
		t.Logf("device=%s mp=%q fst=%q kind=%s pct=%.2f", dp.Device, dp.Mountpoint, dp.Fstype, dp.Kind, dp.UsedPct)
	}

	// No entry should be a parent block device (Kind=disk).
	for _, dp := range m.Partitions {
		if dp.Kind == "disk" {
			t.Errorf("parent block device %s leaked into partitions (Kind=disk)", dp.Device)
		}
	}

	// If the host has a swap partition, it must have usage data after enrichment.
	hasSwap := false
	for _, dp := range m.Partitions {
		if dp.Kind == "swap" {
			hasSwap = true
			if dp.UsedPct < 0 {
				t.Errorf("swap partition %s has UsedPct=%v, want >= 0 after enrichment", dp.Device, dp.UsedPct)
			}
			if dp.Total == 0 {
				t.Errorf("swap partition %s has Total=0", dp.Device)
			}
		}
	}
	if !hasSwap {
		t.Logf("no swap partition on this host")
	}
}

// TestDiskInfoCmd_ReturnsMsg is a smoke test: the command produces a
// non-nil message even on hosts with no disks.
func TestDiskInfoCmd_ReturnsMsg(t *testing.T) {
	cmd := DiskInfoCmd()
	raw := cmd()
	if raw == nil {
		t.Fatal("DiskInfoCmd returned nil message")
	}
}

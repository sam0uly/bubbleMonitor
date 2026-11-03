package tabs

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
)

func TestRenderDisks_BucketedByUsage(t *testing.T) {
	s := &data.AppState{
		UI: data.UIState{Width: 100},
		Config: data.ConfigState{
			BorderStyle: "single",
			BorderType:  "normal",
		},
		Metrics: data.MetricsState{
			DiskPartitions: []data.DiskPartition{
				{Mountpoint: "/", Device: "/dev/sda2", Fstype: "ext4", Total: 100 << 30, Used: 90 << 30, UsedPct: 90.0, Kind: "mounted"},
				{Mountpoint: "/boot", Device: "/dev/sda1", Fstype: "vfat", Total: 1 << 30, Used: 234 << 20, UsedPct: 23.4, Kind: "mounted"},
				{Mountpoint: "[swap]", Device: "/dev/sda3", Fstype: "swap", Total: 18 << 30, Used: 2 << 30, UsedPct: 11.0, Kind: "swap"},
				{Device: "/dev/nvme0n1p1", Total: 16 << 20, UsedPct: -1, Kind: "part"},
				{Device: "/dev/nvme0n1p2", Fstype: "BitLocker", Total: 511 << 30, UsedPct: -1, Kind: "part"},
			},
		},
	}

	var t1, mu, p, b, su, w, a color.Color
	t1 = color.White
	mu = color.Gray{}
	p = color.White
	b = color.White
	su = color.RGBA{R: 255, G: 200, B: 0, A: 255}
	w = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	a = color.Black

	container := lipgloss.NewStyle()

	out := RenderDisks(s, container, su, w, a, t1, mu, p, b, 60)
	clean := stripANSI(out)

	if contains(clean, "/dev/sda ") || contains(clean, "/dev/nvme0n1 ") {
		t.Errorf("parent block device leaked into render output")
	}

	swapIdx := indexOf(clean, "[swap]")
	unmountedIdx := indexOf(clean, "Unmounted")
	if swapIdx < 0 || unmountedIdx < 0 {
		t.Fatalf("expected [swap] and Unmounted in output")
	}
	if swapIdx > unmountedIdx {
		t.Errorf("[swap] should appear before Unmounted")
	}

	t.Logf("\n%s", clean)
}

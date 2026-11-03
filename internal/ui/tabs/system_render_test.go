package tabs

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/shirou/gopsutil/v3/host"

	"samouly.fun/bubbleMonitor/internal/data"
)

func TestRenderSystem_DisksBucketedByUsage(t *testing.T) {
	hostname, _ := host.Info()
	// The data layer is responsible for putting real usage data on every
	// entry that has it. Here we simulate that: swap has OS-enriched
	// usage, mounted filesystems have gopsutil usage, nvme partitions
	// have UsedPct = -1 because gopsutil/lsblk couldn't read usage.
	s := &data.AppState{
		UI: data.UIState{Width: 80},
		Metrics: data.MetricsState{
			HostInfo: hostname,
			DiskPartitions: []data.DiskPartition{
				{Mountpoint: "/", Device: "/dev/sda2", Fstype: "ext4", Total: 100 << 30, Used: 90 << 30, UsedPct: 90.0, Kind: "mounted"},
				{Mountpoint: "/boot", Device: "/dev/sda1", Fstype: "vfat", Total: 1 << 30, Used: 234 << 20, UsedPct: 23.4, Kind: "mounted"},
				{Mountpoint: "[swap]", Device: "/dev/sda3", Fstype: "swap", Total: 18 << 30, Used: 2 << 30, UsedPct: 11.0, Kind: "swap"},
				{Device: "/dev/nvme0n1p1", Total: 16 << 20, UsedPct: -1, Kind: "part"},
				{Device: "/dev/nvme0n1p2", Fstype: "BitLocker", Total: 511 << 30, UsedPct: -1, Kind: "part"},
			},
		},
	}

	var t1, mu, p, b, bg, su, w, a color.Color
	t1 = color.White
	mu = color.Gray{}
	p = color.White
	b = color.White
	bg = color.Black
	su = color.RGBA{R: 255, G: 200, B: 0, A: 255}
	w = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	a = color.Black

	title := lipgloss.NewStyle()
	label := lipgloss.NewStyle()
	value := lipgloss.NewStyle()
	container := lipgloss.NewStyle()

	out := RenderSystem(s, container, title, label, value, t1, mu, p, b, bg, su, w, a, 400, -1)
	clean := stripANSI(out)

	// Parent block devices (Kind=disk) must NOT appear at all.
	if contains(clean, "/dev/sda ") || contains(clean, "/dev/nvme0n1 ") {
		t.Errorf("parent block device leaked into render output")
	}

	// [swap] must appear BEFORE "Unmounted" header.
	swapIdx := indexOf(clean, "[swap]")
	unmountedIdx := indexOf(clean, "Unmounted")
	if swapIdx < 0 || unmountedIdx < 0 {
		t.Fatalf("expected [swap] and Unmounted in output")
	}
	if swapIdx > unmountedIdx {
		t.Errorf("[swap] should appear before Unmounted (it's a mounted/swap entry)")
	}

	// Unmounted nvme partitions must appear AFTER "Unmounted" header.
	nvme1 := indexOf(clean, "/dev/nvme0n1p1")
	if nvme1 < 0 || nvme1 < unmountedIdx {
		t.Errorf("/dev/nvme0n1p1 should appear after Unmounted header")
	}

	t.Logf("\n%s", clean)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func contains(s, sub string) bool {
	return indexOf(s, sub) >= 0
}

func stripANSI(s string) string {
	out := make([]rune, 0, len(s))
	inEsc := false
	for _, r := range s {
		if r == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' || r == 'K' || r == 'H' || r == 'J' || r == ']' || r == 'P' || r == 'h' || r == 'l' {
				inEsc = false
			}
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

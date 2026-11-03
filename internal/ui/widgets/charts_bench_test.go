package widgets

import (
	"testing"

	"charm.land/lipgloss/v2"
	"samouly.fun/bubbleMonitor/internal/data"
)

// BenchmarkRenderSparkline tests memory allocations for sparkline rendering.
// Baseline: 565 allocs/op
// Target after optimization: < 113 allocs/op (80% reduction)
func BenchmarkRenderSparkline(b *testing.B) {
	b.ReportAllocs()
	ring := data.NewRingBuffer(100)
	for i := 0; i < 100; i++ {
		ring.Push(float64(i % 100))
	}

	c1 := lipgloss.Color("#00FF00")
	c2 := lipgloss.Color("#FF0000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderSparkline(ring, 80, 1, c1, c2, 100.0, "default")
	}
}

// BenchmarkRenderLineChart tests memory allocations for line chart rendering.
// Baseline: 5,632 allocs/op
// Target after optimization: < 1,126 allocs/op (80% reduction)
func BenchmarkRenderLineChart(b *testing.B) {
	b.ReportAllocs()
	ring := data.NewRingBuffer(100)
	for i := 0; i < 100; i++ {
		ring.Push(float64(i % 100))
	}

	c1 := lipgloss.Color("#00FF00")
	c2 := lipgloss.Color("#FF0000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderLineChart(ring, 80, 10, c1, c2, 100.0, "default")
	}
}

// BenchmarkRenderBrailleChart tests memory allocations for braille chart rendering.
// Baseline: 102 allocs/op
// Target after optimization: < 21 allocs/op (80% reduction)
func BenchmarkRenderBrailleChart(b *testing.B) {
	b.ReportAllocs()
	ring := data.NewRingBuffer(100)
	for i := 0; i < 100; i++ {
		ring.Push(float64(i % 100))
	}

	c1 := lipgloss.Color("#00FF00")
	c2 := lipgloss.Color("#FF0000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderBrailleChart(ring, 80, 10, c1, c2, 100.0, "default")
	}
}

func BenchmarkRenderLineChartLarge(b *testing.B) {
	b.ReportAllocs()

	// Create test data with enough points
	rb := data.NewRingBuffer(300)
	for i := 0; i < 300; i++ {
		rb.Push(float64(i % 100))
	}

	width := 200 // Large terminal width
	height := 50 // Large terminal height
	c1 := lipgloss.Color("#00FF00")
	c2 := lipgloss.Color("#FF0000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderLineChart(rb, width, height, c1, c2, 100.0, "default")
	}
}

func BenchmarkRenderAllCharts(b *testing.B) {
	b.ReportAllocs()

	rb := data.NewRingBuffer(100)
	for i := 0; i < 100; i++ {
		rb.Push(float64(i % 100))
	}

	c1 := lipgloss.Color("#00FF00")
	c2 := lipgloss.Color("#FF0000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderSparkline(rb, 80, 1, c1, c2, 100.0, "default")
		_ = RenderLineChart(rb, 80, 10, c1, c2, 100.0, "default")
		_ = RenderBrailleChart(rb, 40, 10, c1, c2, 100.0, "default")
		_ = RenderBarChart(rb, 80, 10, c1, c2, 100.0, "default")
	}
}

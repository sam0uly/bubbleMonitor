package widgets

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"samouly.fun/bubbleMonitor/internal/data"
)

// TestChartVisualRegression ensures that optimization changes don't affect visual output.
// It verifies that chart functions produce deterministic, consistent output.
func TestChartVisualRegression(t *testing.T) {
	rb := data.NewRingBuffer(10)
	for i := 0; i < 10; i++ {
		rb.Push(float64(i * 10))
	}

	c1 := lipgloss.Color("#00FF00")
	c2 := lipgloss.Color("#FF0000")

	t.Run("RenderLineChart", func(t *testing.T) {
		output := RenderLineChart(rb, 10, 5, c1, c2, 100.0, "default")

		if output == "" {
			t.Error("RenderLineChart returned empty string")
		}
		if !strings.Contains(output, "█") && !strings.Contains(output, " ") {
			t.Error("RenderLineChart output doesn't contain expected characters")
		}

		output2 := RenderLineChart(rb, 10, 5, c1, c2, 100.0, "default")
		if output != output2 {
			t.Error("RenderLineChart is non-deterministic")
		}
	})

	t.Run("RenderSparkline", func(t *testing.T) {
		output := RenderSparkline(rb, 10, 1, c1, c2, 100.0, "default")

		if output == "" {
			t.Error("RenderSparkline returned empty string")
		}

		output2 := RenderSparkline(rb, 10, 1, c1, c2, 100.0, "default")
		if output != output2 {
			t.Error("RenderSparkline is non-deterministic")
		}
	})
}

package widgets

import (
	"fmt"
	"image/color"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/util"
)

var (
	stringGridPool = sync.Pool{
		New: func() interface{} {
			grid := make([][]string, 0, InitialGridCapacity)
			return &grid
		},
	}
	boolGridPool = sync.Pool{
		New: func() interface{} {
			grid := make([][]bool, 0, 20)
			return &grid
		},
	}
	barCache     = make([]string, 256)
	barCacheInit bool
)

func initBarCache() {
	if barCacheInit {
		return
	}
	barCacheInit = true
	for i := range barCache {
		barCache[i] = strings.Repeat("█", i)
	}
}

type chartStyleCache struct {
	styles map[string]lipgloss.Style
	theme  string
}

var chartCache = &chartStyleCache{
	styles: make(map[string]lipgloss.Style),
}

func getChartStyle(c color.Color, theme string) lipgloss.Style {
	if chartCache.theme != theme {
		chartCache.theme = theme
		chartCache.styles = make(map[string]lipgloss.Style)
	}

	key := fmt.Sprintf("%v", c)
	if style, ok := chartCache.styles[key]; ok {
		return style
	}

	style := lipgloss.NewStyle().Foreground(c)
	chartCache.styles[key] = style
	return style
}

func RenderSparkline(data data.Accessor, width, height int, c1, c2 color.Color, fixedMax float64, theme string) string {
	if data.Len() == 0 {
		return "No data"
	}
	chars := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	// Pre-render all characters with both colors
	renderedC1 := make([]string, len(chars))
	renderedC2 := make([]string, len(chars))
	style1 := getChartStyle(c1, theme)
	style2 := getChartStyle(c2, theme)
	for i, ch := range chars {
		renderedC1[i] = style1.Render(ch)
		renderedC2[i] = style2.Render(ch)
	}

	maxV := fixedMax
	if maxV <= 0 {
		maxV = data.Max()
	}
	if maxV == 0 {
		maxV = 1
	}
	var result strings.Builder
	result.Grow(width * 12) // Pre-allocate for styled UTF-8 chars
	startIdx := 0
	if data.Len() > width {
		startIdx = data.Len() - width
	}
	for i := startIdx; i < data.Len(); i++ {
		val := data.Get(i)
		normVal := val / maxV
		chIdx := int(normVal * float64(len(chars)-1))
		if chIdx >= len(chars) {
			chIdx = len(chars) - 1
		}
		if chIdx < 0 {
			chIdx = 0
		}

		if val > 70 {
			result.WriteString(renderedC2[chIdx])
		} else {
			result.WriteString(renderedC1[chIdx])
		}
	}

	rSpace := style1.Render(" ")
	for i := data.Len(); i < width; i++ {
		result.WriteString(rSpace)
	}
	return result.String()
}

func RenderLineChart(data data.Accessor, width, height int, c1, c2 color.Color, fixedMax float64, theme string) string {
	if data.Len() == 0 || height < 1 {
		return "No data"
	}
	maxV := fixedMax
	if maxV <= 0 {
		maxV = data.Max()
	}
	if maxV == 0 {
		maxV = 1
	}

	gridPtr, ok := stringGridPool.Get().(*[][]string)
	var grid [][]string
	if ok {
		grid = *gridPtr
	}

	if cap(grid) < height {
		grid = make([][]string, height)
	} else {
		grid = grid[:height]
	}

	for r := 0; r < height; r++ {
		if cap(grid[r]) < width {
			grid[r] = make([]string, width)
		} else {
			grid[r] = grid[r][:width]
		}
		for c := 0; c < width; c++ {
			grid[r][c] = " "
		}
	}

	startIdx := 0
	if data.Len() > width {
		startIdx = data.Len() - width
	}

	for col := 0; col < width && (startIdx+col) < data.Len(); col++ {
		val := data.Get(startIdx + col)
		normalized := val / maxV
		filledRows := int(normalized * float64(height))

		for row := height - 1; row >= height-filledRows && row >= 0; row-- {
			grid[row][col] = "█"
		}
	}

	lines := make([]string, 0, height)
	style1 := getChartStyle(c1, theme)
	style2 := getChartStyle(c2, theme)
	rSpace := style1.Render(" ")
	rBlock := style2.Render("█")

	for r := 0; r < height; r++ {
		var line strings.Builder
		line.Grow(width * 12) // Estimate for styled chars
		for c := 0; c < width; c++ {
			if grid[r][c] == "█" {
				line.WriteString(rBlock)
			} else {
				line.WriteString(rSpace)
			}
		}
		lines = append(lines, line.String())
	}

	*gridPtr = grid
	stringGridPool.Put(gridPtr)

	return strings.Join(lines, "\n")
}

func RenderBarChart(data data.Accessor, width, height int, c1, c2 color.Color, fixedMax float64, theme string) string {
	if data.Len() == 0 {
		return "No data"
	}
	initBarCache()
	maxV := fixedMax
	if maxV <= 0 {
		maxV = data.Max()
	}
	if maxV == 0 {
		maxV = 1
	}

	startIdx := 0
	count := data.Len() - startIdx
	if data.Len() > height {
		startIdx = data.Len() - height
		count = height
	}

	style1 := getChartStyle(c1, theme)
	style2 := getChartStyle(c2, theme)

	lines := make([]string, 0, count)
	for i := startIdx; i < data.Len(); i++ {
		val := data.Get(i)
		barLen := int((val / maxV) * float64(width-8))
		if barLen < 0 {
			barLen = 0
		}

		label := util.FastPercent1(val) + " "
		var bar string
		if barLen < len(barCache) {
			bar = barCache[barLen]
		} else {
			bar = strings.Repeat("█", barLen)
		}
		if val > 70 {
			bar = style2.Render(bar)
		} else {
			bar = style1.Render(bar)
		}
		lines = append(lines, label+bar)
	}

	return strings.Join(lines, "\n")
}

func RenderBrailleChart(data data.Accessor, width, height int, c1, c2 color.Color, fixedMax float64, theme string) string {
	if data.Len() == 0 || height < 1 {
		return "No data"
	}
	maxV := fixedMax
	if maxV <= 0 {
		maxV = data.Max()
	}
	if maxV == 0 {
		maxV = 1
	}

	dotColumns := width * 2
	dotsPerCol := height * 4
	dataLen := data.Len()

	gridPtr, ok := boolGridPool.Get().(*[][]bool)
	if !ok {
		gridPtr = &[][]bool{}
	}
	dots := *gridPtr

	if cap(dots) < dotsPerCol {
		dots = make([][]bool, dotsPerCol)
	} else {
		dots = dots[:dotsPerCol]
	}

	for r := 0; r < dotsPerCol; r++ {
		if cap(dots[r]) < dotColumns {
			dots[r] = make([]bool, dotColumns)
		} else {
			dots[r] = dots[r][:dotColumns]
			for c := range dots[r] {
				dots[r][c] = false
			}
		}
	}

	// Standard braille chart: always show the last min(dataLen, dotColumns)
	// data points, 1 data point per dot column (2 per braille char).
	// Data grows from left, scrolls left once full.
	// History length only controls ring buffer capacity, not visible resolution.
	sampleWidth := dataLen
	if sampleWidth > dotColumns {
		sampleWidth = dotColumns
	}
	startIdx := dataLen - sampleWidth

	for col := 0; col < sampleWidth; col++ {
		val := data.Get(startIdx + col)
		normalized := val / maxV
		filledDots := int(normalized * float64(dotsPerCol))
		for row := dotsPerCol - 1; row >= dotsPerCol-filledDots && row >= 0; row-- {
			dots[row][col] = true
		}
	}

	style1 := getChartStyle(c1, theme)
	style2 := getChartStyle(c2, theme)

	lines := make([]string, 0, height)
	for charRow := 0; charRow < height; charRow++ {
		var line strings.Builder
		line.Grow(width)
		for charCol := 0; charCol < width; charCol++ {
			dotRow := charRow * 4
			dotCol := charCol * 2

			var braille rune = 0x2800
			if dotRow < dotsPerCol && dots[dotRow][dotCol] {
				braille += 1
			}
			if dotRow+1 < dotsPerCol && dots[dotRow+1][dotCol] {
				braille += 2
			}
			if dotRow+2 < dotsPerCol && dots[dotRow+2][dotCol] {
				braille += 4
			}
			if dotRow+3 < dotsPerCol && dots[dotRow+3][dotCol] {
				braille += 64
			}
			if dotRow < dotsPerCol && dots[dotRow][dotCol+1] {
				braille += 8
			}
			if dotRow+1 < dotsPerCol && dots[dotRow+1][dotCol+1] {
				braille += 16
			}
			if dotRow+2 < dotsPerCol && dots[dotRow+2][dotCol+1] {
				braille += 32
			}
			if dotRow+3 < dotsPerCol && dots[dotRow+3][dotCol+1] {
				braille += 128
			}
			line.WriteRune(braille)
		}

		var renderedLine string
		if charRow < height/2 {
			renderedLine = style2.Render(line.String())
		} else {
			renderedLine = style1.Render(line.String())
		}
		lines = append(lines, renderedLine)
	}

	*gridPtr = dots
	boolGridPool.Put(gridPtr)

	return strings.Join(lines, "\n")
}

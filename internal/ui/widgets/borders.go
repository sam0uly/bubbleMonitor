package widgets

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"samouly.fun/bubbleMonitor/internal/util"
)

// GetBorder returns the lipgloss border for the current style
func GetBorder(borderStyle, borderType string) lipgloss.Border {
	switch borderStyle {
	case "double":
		return lipgloss.DoubleBorder()
	case "dashed":
		return lipgloss.Border{
			Top:         "-",
			Bottom:      "-",
			Left:        "|",
			Right:       "|",
			TopLeft:     "+",
			TopRight:    "+",
			BottomLeft:  "+",
			BottomRight: "+",
		}
	default:
		if borderType == "rounded" {
			return lipgloss.RoundedBorder()
		}
		return lipgloss.NormalBorder()
	}
}

// RenderTopBorderWithBg renders a top border with title
func RenderTopBorderWithBg(title string, width int, border lipgloss.Border, borderColor, titleColor color.Color) string {
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true)

	leftPart := borderStyle.Render(border.TopLeft+border.Top+" ") +
		titleStyle.Render(title) +
		borderStyle.Render(" ")

	rightPart := borderStyle.Render(border.TopRight)

	leftWidth := lipgloss.Width(leftPart)
	rightWidth := lipgloss.Width(rightPart)

	remainingWidth := width - leftWidth - rightWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	hLine := border.Top
	hLineWidth := lipgloss.Width(hLine)
	if hLineWidth == 0 {
		hLineWidth = 1
	}

	repeatCount := remainingWidth / hLineWidth
	middlePart := borderStyle.Render(strings.Repeat(hLine, repeatCount))

	result := leftPart + middlePart + rightPart
	currentWidth := lipgloss.Width(result)
	if currentWidth < width {
		result = leftPart + middlePart + borderStyle.Render(strings.Repeat(border.Top, width-currentWidth)) + rightPart
	}

	return result
}

// RenderScrollbar renders a vertical scrollbar on the right side of a container.
// Returns a string with scrollbar characters: "▄" for scroll position, "│" for track, " " for empty.
// Only shows if totalItems > visibleItems (there's more content to scroll).
func RenderScrollbar(totalItems, visibleItems, scrollOffset int, color color.Color) string {
	// Don't show scrollbar if all items fit
	if totalItems <= visibleItems {
		return ""
	}

	// Clamp scrollOffset to valid range
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	maxOffset := totalItems - visibleItems
	if scrollOffset > maxOffset {
		scrollOffset = maxOffset
	}

	// Calculate the scrollbar position
	// The thumb position should be proportional to scrollOffset
	scrollableHeight := totalItems - visibleItems
	thumbPosition := (scrollOffset * (visibleItems - 1)) / scrollableHeight

	// Build the scrollbar string
	style := lipgloss.NewStyle().Foreground(color)
	result := ""

	for i := 0; i < visibleItems; i++ {
		if i == thumbPosition {
			result += style.Render("▄")
		} else {
			result += style.Render("│")
		}
		if i < visibleItems-1 {
			result += "\n"
		}
	}

	return result
}

// RenderScrollbarColumn renders a vertical scrollbar as a fixed-width column.
// Unlike RenderScrollbar which merges with content lines, this renders a separate
// column that can be joined horizontally with content.
func RenderScrollbarColumn(totalItems, visibleItems, scrollOffset int, width int, color color.Color) string {
	if totalItems <= visibleItems {
		return ""
	}

	if scrollOffset < 0 {
		scrollOffset = 0
	}
	maxOffset := totalItems - visibleItems
	if scrollOffset > maxOffset {
		scrollOffset = maxOffset
	}

	scrollableHeight := totalItems - visibleItems
	thumbPosition := (scrollOffset * (visibleItems - 1)) / scrollableHeight

	style := lipgloss.NewStyle().Foreground(color)

	var lines []string
	for i := 0; i < visibleItems; i++ {
		var char string
		if i == thumbPosition {
			char = "▄"
		} else {
			char = "│"
		}
		padded := util.PadRight(char, width)
		lines = append(lines, style.Render(padded))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

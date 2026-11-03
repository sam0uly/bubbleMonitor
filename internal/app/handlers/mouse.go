package handlers

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/ui/input"
)

const (
	headerFooterRows = data.ReservedContentRows
	headerOffset     = 3
)

func HandleMouse(m *data.AppState, msg tea.MouseMsg) tea.Cmd {
	mouseEvent := msg.Mouse()
	x, y := mouseEvent.X, mouseEvent.Y

	m.UI.MouseX = x
	m.UI.MouseY = y

	switch msg.(type) {
	case tea.MouseClickMsg:
		if mouseEvent.Button == tea.MouseLeft {
			return handleLeftClick(m, x, y)
		}
		if mouseEvent.Button == tea.MouseRight {
			return handleRightClick(m, x, y)
		}
	case tea.MouseWheelMsg:
		if mouseEvent.Button == tea.MouseWheelUp {
			return handleScrollUp(m)
		}
		if mouseEvent.Button == tea.MouseWheelDown {
			return handleScrollDown(m)
		}
	case tea.MouseMotionMsg:
		return func() tea.Msg { return nil }
	}

	return nil
}

func handleLeftClick(m *data.AppState, x, y int) tea.Cmd {
	for i := len(m.UI.Zones) - 1; i >= 0; i-- {
		zone := m.UI.Zones[i]
		if x >= zone.X && x < zone.X+zone.Width && y >= zone.Y && y < zone.Y+zone.Height {
			if zone.OnClick != nil {
				return zone.OnClick()
			}
		}
	}

	return nil
}

func handleRightClick(m *data.AppState, x, y int) tea.Cmd {
	for _, zone := range m.UI.Zones {
		if x >= zone.X && x < zone.X+zone.Width && y >= zone.Y && y < zone.Y+zone.Height {
			if zone.Type == input.ZoneTypeListItem && strings.HasPrefix(zone.ID, "process-row-") {
				if pid, ok := zone.Metadata["pid"].(int32); ok {
					if index, ok := zone.Metadata["index"].(int); ok {
						m.Process.SelectedProcess = index
						m.Process.ProcessMenuX = x
						m.Process.ProcessMenuY = y
						m.Process.ShowProcessMenu = true
						m.Process.ProcessMenuIdx = index
						_ = pid
						return nil
					}
				}
			}

			if zone.Type == input.ZoneTypeMenuItem {
				return nil
			}

			m.Process.ShowProcessMenu = false
			return nil
		}
	}

	return nil
}

func handleScrollUp(m *data.AppState) tea.Cmd {
	currentTab := ""
	if m.UI.SelectedTab < len(m.UI.ActiveTabs) {
		currentTab = m.UI.ActiveTabs[m.UI.SelectedTab]
	}

	if currentTab == "Processes" {
		if m.Process.SelectedProcess > 0 {
			m.Process.SelectedProcess--
			if m.Process.SelectedProcess < m.Process.ProcessScrollOffset {
				m.Process.ProcessScrollOffset = m.Process.SelectedProcess
			}
		}
	} else if currentTab == "Services" {
		if m.Process.ServicesScrollOffset > 0 {
			m.Process.ServicesScrollOffset--
		}
	} else if currentTab == "Connections" {
		if m.Process.ConnectionsScrollOffset > 0 {
			m.Process.ConnectionsScrollOffset--
		}
	} else if currentTab == "Logs" {
		if m.Process.LogsScrollOffset > 0 {
			m.Process.LogsScrollOffset--
		}
	} else if currentTab == "Metrics" {
		m.UI.CpuCoreScrollOffset--
		if m.UI.CpuCoreScrollOffset < 0 {
			m.UI.CpuCoreScrollOffset = 0
		}
	} else if currentTab == "System" {
		rows := max(m.UI.Height-headerFooterRows, 1)
		blockIdx := -1
		if m.UI.SystemBlockCount > 0 {
			cols := 1
			if m.UI.Width >= 80 {
				cols = 2
			}
			if m.UI.Width >= 120 {
				cols = 3
			}
			colWidth := m.UI.Width / cols
			col := m.UI.MouseX / colWidth
			if col >= cols {
				col = cols - 1
			}
			row := max((m.UI.MouseY-headerOffset)/rows, 0)
			blockIdx = row*cols + col
			if blockIdx >= m.UI.SystemBlockCount {
				blockIdx = -1
			}
		}
		if blockIdx >= 0 && m.UI.SystemBlockScrollable[blockIdx] {
			m.UI.SystemBlockScrollOffsets[blockIdx] -= rows
			if m.UI.SystemBlockScrollOffsets[blockIdx] < 0 {
				m.UI.SystemBlockScrollOffsets[blockIdx] = 0
			}
		}
	} else if m.Process.ShowOpenFiles {
		m.Process.OpenFilesView.LineUp(1)
	}
	return nil
}

func handleScrollDown(m *data.AppState) tea.Cmd {
	currentTab := ""
	if m.UI.SelectedTab < len(m.UI.ActiveTabs) {
		currentTab = m.UI.ActiveTabs[m.UI.SelectedTab]
	}

	if currentTab == "Processes" {
		procs := m.GetFilteredProcesses()
		if m.Process.SelectedProcess < len(procs)-1 {
			m.Process.SelectedProcess++
			rows := getVisibleProcessRows(m)
			if m.Process.SelectedProcess >= m.Process.ProcessScrollOffset+rows {
				m.Process.ProcessScrollOffset = m.Process.SelectedProcess - rows + 1
			}
		}
	} else if currentTab == "Services" {
		rows := max(m.UI.Height-headerFooterRows, 1)
		maxScroll := max(len(m.Process.Services)-rows, 0)
		if m.Process.ServicesScrollOffset < maxScroll {
			m.Process.ServicesScrollOffset++
		}
	} else if currentTab == "Connections" {
		rows := max(m.UI.Height-headerFooterRows, 1)
		maxScroll := max(len(m.Process.Connections)-rows, 0)
		if m.Process.ConnectionsScrollOffset < maxScroll {
			m.Process.ConnectionsScrollOffset++
		}
	} else if currentTab == "Logs" {
		rows := max(m.UI.Height-headerFooterRows, 1)
		maxScroll := max(len(m.Process.SystemLogs)-rows, 0)
		if m.Process.LogsScrollOffset < maxScroll {
			m.Process.LogsScrollOffset++
		}
	} else if currentTab == "Metrics" {
		m.UI.CpuCoreScrollOffset++
	} else if currentTab == "System" {
		rows := max(m.UI.Height-headerFooterRows, 1)
		blockIdx := -1
		if m.UI.SystemBlockCount > 0 {
			cols := 1
			if m.UI.Width >= 80 {
				cols = 2
			}
			if m.UI.Width >= 120 {
				cols = 3
			}
			colWidth := m.UI.Width / cols
			col := m.UI.MouseX / colWidth
			if col >= cols {
				col = cols - 1
			}
			row := max((m.UI.MouseY-headerOffset)/rows, 0)
			blockIdx = row*cols + col
			if blockIdx >= m.UI.SystemBlockCount {
				blockIdx = -1
			}
		}
		if blockIdx >= 0 && m.UI.SystemBlockScrollable[blockIdx] {
			maxScroll := 0
			if m.UI.SystemBlockMaxScroll != nil {
				maxScroll = m.UI.SystemBlockMaxScroll[blockIdx]
			}
			m.UI.SystemBlockScrollOffsets[blockIdx] += rows
			if m.UI.SystemBlockScrollOffsets[blockIdx] > maxScroll {
				m.UI.SystemBlockScrollOffsets[blockIdx] = maxScroll
			}
		}
	} else if m.Process.ShowOpenFiles {
		m.Process.OpenFilesView.LineDown(1)
	}
	return nil
}

func getVisibleProcessRows(m *data.AppState) int {
	rows := max(m.UI.Height-data.ReservedContentRows, 3)
	return rows
}

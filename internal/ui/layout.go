package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
	"samouly.fun/bubbleMonitor/internal/provider/process"
	"samouly.fun/bubbleMonitor/internal/ui/input"
	"samouly.fun/bubbleMonitor/internal/ui/overlays"
	"samouly.fun/bubbleMonitor/internal/ui/tabs"
	"samouly.fun/bubbleMonitor/internal/ui/widgets"
)

func isNoColor(c color.Color) bool {
	_, ok := c.(lipgloss.NoColor)
	return ok
}

const sidePadding = 2

// ViewModel is an interface for rendering the UI
// This avoids import cycles by not importing model package
type ViewModel interface {
	GetBorder() lipgloss.Border
	GetColors() ThemePalette
	GetVisibleProcesses() ([]data.ProcessInfo, map[int32]int)
	// Direct field accessors from AppState (embedded in Model)
	GetAppState() *data.AppState
}

// Cached styles
type layoutStyleCache struct {
	box       lipgloss.Style
	title     lipgloss.Style
	dim       lipgloss.Style
	warn      lipgloss.Style
	header    lipgloss.Style
	key       lipgloss.Style
	val       lipgloss.Style
	activeTab lipgloss.Style
	tab       lipgloss.Style
	theme     string
}

var styleCache = &layoutStyleCache{}

// MainViewFromState renders the entire application UI from AppState
func MainViewFromState(s *data.AppState, getBorder func() lipgloss.Border, getColors func() ThemePalette) tea.View {
	var zoneManager input.ZoneManager
	if s.UI.ZoneManager != nil {
		if zm, ok := s.UI.ZoneManager.(input.ZoneManager); ok {
			zoneManager = zm
			zoneManager.Clear()
		}
	}
	if zoneManager == nil {
		zoneManager = input.NewZoneManager()
	}

	// Cursor X and Y tracking for zone positioning
	cursorX := 0
	footerY := 0

	theme := getColors()
	p := theme.Primary
	t := theme.Text
	bg := theme.Background
	mu := theme.Muted
	a := theme.Alert
	b := theme.Border
	su := theme.Success
	btnText := theme.Background

	// Rebuild cache if theme changed
	if styleCache.theme != s.Config.Theme {
		styleCache.theme = s.Config.Theme

		styleCache.box = lipgloss.NewStyle().
			Border(getBorder()).
			BorderForeground(p).
			Padding(1, 2).
			Margin(1).
			Foreground(t).
			Align(lipgloss.Center)

		styleCache.title = lipgloss.NewStyle().
			Foreground(p).
			Bold(true).
			MarginBottom(1)

		styleCache.dim = lipgloss.NewStyle().
			Foreground(mu)

		styleCache.warn = lipgloss.NewStyle().Foreground(a).Bold(true).Blink(true)

		styleCache.header = lipgloss.NewStyle().
			Bold(true).
			Foreground(t).
			Background(p).
			Padding(0, 1)

		styleCache.key = lipgloss.NewStyle().Foreground(p).Bold(true)
		styleCache.val = lipgloss.NewStyle().Foreground(t)

		styleCache.activeTab = lipgloss.NewStyle().
			Bold(true).
			Foreground(t).
			Background(p).
			Padding(0, 1)

		styleCache.tab = lipgloss.NewStyle().
			Foreground(mu).
			Padding(0, 1)
	}

	// Minimum dimensions check
	if s.UI.Width < MinWindowWidth || s.UI.Height < MinWindowHeight {
		// Use cached styles where possible, but border might be dynamic so we update it
		boxStyle := styleCache.box.Border(getBorder())

		msg := fmt.Sprintf(
			"%s\n\n%s\n%s",
			styleCache.title.Render("WINDOW TOO SMALL"),
			fmt.Sprintf("Current: %dx%d", s.UI.Width, s.UI.Height),
			styleCache.dim.Render(fmt.Sprintf("Minimum: %dx%d", MinWindowWidth, MinWindowHeight)),
		)

		v := tea.NewView(lipgloss.Place(
			s.UI.Width, s.UI.Height,
			lipgloss.Center, lipgloss.Center,
			boxStyle.Render(msg),
		))

		v.AltScreen = true

		if s.Config.BackgroundOpaque && !isNoColor(bg) {
			v.BackgroundColor = bg
		}

		return v
	}

	// Theme colors from current theme
	// theme := getColors() - Already declared
	// p := theme.Primary - Already declared
	// t := theme.Text - Already declared
	// bg := theme.Background - Already declared
	// mu := theme.Muted - Already declared
	// a := theme.Alert - Already declared
	// b := theme.Border - Already declared

	// Handle Transparency
	if !s.Config.BackgroundOpaque {
		bg = compat.AdaptiveColor{Light: lipgloss.NoColor{}, Dark: lipgloss.NoColor{}}
	}

	border := getBorder()

	headerText := "BUBBLE MONITOR"
	header := styleCache.header.Render(headerText)

	header += " " + styleCache.dim.Render(time.Now().Format("15:04:05"))

	var alertStr string
	if s.Metrics.AlertManager != nil && s.Metrics.AlertManager.HasAlerts() {
		alerts := s.Metrics.AlertManager.GetAlerts()
		rawText := " ⚠️ " + alerts[0].Message + " "

		alertStyle := lipgloss.NewStyle().Foreground(a).Bold(true)
		if s.UI.TickCount%2 == 0 {
			alertStyle = alertStyle.Faint(true)
		}
		alertStr = alertStyle.Render(rawText)
	}

	var tabBlocks []string

	headerWidth := lipgloss.Width(header)
	spacerWidth := 4
	availableWidth := s.UI.Width - headerWidth - spacerWidth
	if availableWidth < 20 {
		availableWidth = s.UI.Width - 20
	}

	tabsStartX := headerWidth + spacerWidth
	currentRowY := 0
	cursorX = tabsStartX

	currentRowWidth := 0
	hoverTabStyle := lipgloss.NewStyle().
		Foreground(t).
		Background(mu).
		Padding(0, 1)

	type tabInfo struct {
		name  string
		index int
		x     int
		y     int
		width int
	}
	var tabInfos []tabInfo

	for i, titleRaw := range s.UI.ActiveTabs {
		title := strings.ToUpper(titleRaw)
		renderedTab := styleCache.tab.Render(title)
		tabWidth := lipgloss.Width(renderedTab)

		if currentRowWidth+tabWidth >= availableWidth && currentRowWidth > 0 {
			currentRowY++
			cursorX = tabsStartX
			currentRowWidth = 0
		}

		tabInfos = append(tabInfos, tabInfo{
			index: i,
			name:  titleRaw,
			x:     cursorX,
			y:     currentRowY,
			width: tabWidth,
		})

		cursorX += tabWidth
		currentRowWidth += tabWidth
	}

	// Build tabBlocks
	for _, info := range tabInfos {
		title := strings.ToUpper(info.name)
		var renderedTab string
		if info.index == s.UI.SelectedTab {
			renderedTab = styleCache.activeTab.Render(title)
		} else {
			tabZoneID := "tab-" + info.name
			if zoneManager.IsHovered(tabZoneID) {
				renderedTab = hoverTabStyle.Render(title)
			} else {
				renderedTab = styleCache.tab.Render(title)
			}
		}
		tabBlocks = append(tabBlocks, renderedTab)
	}

	// Wrap tabs to multiple rows
	var tabRows [][]string
	var row []string
	rowWidth := 0

	for _, tab := range tabBlocks {
		tabWidth := lipgloss.Width(tab)
		if rowWidth+tabWidth >= availableWidth && len(row) > 0 {
			tabRows = append(tabRows, row)
			row = []string{}
			rowWidth = 0
		}
		row = append(row, tab)
		rowWidth += tabWidth
	}
	if len(row) > 0 {
		tabRows = append(tabRows, row)
	}

	// Render tab rows
	var tabRowStrings []string
	for _, row := range tabRows {
		tabRowStrings = append(tabRowStrings, lipgloss.JoinHorizontal(lipgloss.Bottom, row...))
	}
	tabRow := lipgloss.JoinVertical(lipgloss.Bottom, tabRowStrings...)

	// Calculate width of first tab row only (for alert positioning)
	tabRowWidth := 0
	if len(tabRowStrings) > 0 {
		tabRowWidth = lipgloss.Width(tabRowStrings[0])
	}

	alertWidth := lipgloss.Width(alertStr)
	tabsWithAlertSpace := headerWidth + spacerWidth + tabRowWidth + alertWidth

	var spacerWithCentering string
	alertAtBottom := alertStr == "" || tabsWithAlertSpace >= s.UI.Width

	if alertAtBottom {
		spacerWithCentering = strings.Repeat(" ", s.UI.Width-headerWidth-tabRowWidth-1)
		if spacerWithCentering < " " {
			spacerWithCentering = " "
		}
	} else if tabsWithAlertSpace < s.UI.Width {
		spacerWithCentering = strings.Repeat(" ", spacerWidth)
	} else {
		spacerWithCentering = strings.Repeat(" ", s.UI.Width-headerWidth-tabRowWidth-2)
	}

	var topBar string
	if alertAtBottom {
		topBar = lipgloss.JoinHorizontal(lipgloss.Top, header, spacerWithCentering, tabRow)
	} else {
		rs := s.UI.Width - (headerWidth + lipgloss.Width(spacerWithCentering) + tabRowWidth + alertWidth)
		if rs > 0 {
			alertBlock := lipgloss.NewStyle().Foreground(a).Bold(true).Render(alertStr)
			topBar = lipgloss.JoinHorizontal(lipgloss.Top, header, spacerWithCentering, tabRow, strings.Repeat(" ", rs), alertBlock)
		} else {
			topBar = lipgloss.JoinHorizontal(lipgloss.Top, header, spacerWithCentering, tabRow)
		}
	}

	oldTabsStartX := headerWidth + spacerWidth
	tabsStartX = headerWidth + lipgloss.Width(spacerWithCentering)

	for _, info := range tabInfos {
		tabName := info.name
		zoneX := info.x - oldTabsStartX + tabsStartX
		zoneManager.Register(input.Zone{
			ID:     "tab-" + tabName,
			Type:   input.ZoneTypeTab,
			X:      zoneX,
			Y:      info.y,
			Width:  info.width,
			Height: 1,
			OnClick: func() tea.Cmd {
				for i, t := range s.UI.ActiveTabs {
					if t == tabName {
						s.UI.SelectedTab = i
						break
					}
				}
				return nil
			},
		})
	}

	zoneManager.UpdateMousePos(s.UI.MouseX, s.UI.MouseY)

	helpBtnStyle := lipgloss.NewStyle().Bold(true).Foreground(btnText).Background(theme.Secondary).Padding(0, 1)
	filterBtnStyle := lipgloss.NewStyle().Bold(true).Foreground(btnText).Background(p).Padding(0, 1)
	killBtnStyle := lipgloss.NewStyle().Bold(true).Foreground(btnText).Background(a).Padding(0, 1)
	sortBtnStyle := lipgloss.NewStyle().Bold(true).Foreground(btnText).Background(su).Padding(0, 1)
	settingsBtnStyle := lipgloss.NewStyle().Bold(true).Foreground(btnText).Background(theme.Warning).Padding(0, 1)
	quitBtnStyle := lipgloss.NewStyle().Bold(true).Foreground(btnText).Background(mu).Padding(0, 1)
	applyBtnStyle := lipgloss.NewStyle().Bold(true).Foreground(btnText).Background(theme.Warning).Padding(0, 1)
	dimBtnStyle := lipgloss.NewStyle().Bold(true).Foreground(mu).Padding(0, 1)

	activeTab := ""
	if s.UI.SelectedTab >= 0 && s.UI.SelectedTab < len(s.UI.ActiveTabs) {
		activeTab = s.UI.ActiveTabs[s.UI.SelectedTab]
	}

	var footerZoneInfos []struct {
		text  string
		id    string
		width int
	}

	helpText := "[?]help"
	helpRendered := helpBtnStyle.Render(helpText)
	footerZoneInfos = append(footerZoneInfos, struct {
		text  string
		id    string
		width int
	}{helpText, "footer-help", lipgloss.Width(helpRendered)})

	settingsText := "[.]settings"
	settingsRendered := settingsBtnStyle.Render(settingsText)
	footerZoneInfos = append(footerZoneInfos, struct {
		text  string
		id    string
		width int
	}{settingsText, "footer-settings", lipgloss.Width(settingsRendered)})

	if activeTab == "Processes" && s.Process.FilterMode {
		quitText := "[q]uit"
		quitRendered := quitBtnStyle.Render(quitText)
		footerZoneInfos = append(footerZoneInfos, struct {
			text  string
			id    string
			width int
		}{quitText, "footer-quit", lipgloss.Width(quitRendered)})

		applyText := "[esc/enter]apply"
		applyRendered := applyBtnStyle.Render(applyText)
		footerZoneInfos = append(footerZoneInfos, struct {
			text  string
			id    string
			width int
		}{applyText, "footer-apply", lipgloss.Width(applyRendered)})
	} else {
		if activeTab == "Processes" {
			filterText := "[f]ilter"
			filterRendered := filterBtnStyle.Render(filterText)
			footerZoneInfos = append(footerZoneInfos, struct {
				text  string
				id    string
				width int
			}{filterText, "footer-filter", lipgloss.Width(filterRendered)})

			killText := "[K]ill"
			killRendered := killBtnStyle.Render(killText)
			footerZoneInfos = append(footerZoneInfos, struct {
				text  string
				id    string
				width int
			}{killText, "footer-kill", lipgloss.Width(killRendered)})

			sortText := "[s]ort"
			sortRendered := sortBtnStyle.Render(sortText)
			footerZoneInfos = append(footerZoneInfos, struct {
				text  string
				id    string
				width int
			}{sortText, "footer-sort", lipgloss.Width(sortRendered)})
		}
		quitText := "[q]uit"
		quitRendered := quitBtnStyle.Render(quitText)
		footerZoneInfos = append(footerZoneInfos, struct {
			text  string
			id    string
			width int
		}{quitText, "footer-quit", lipgloss.Width(quitRendered)})
	}

	var footer string
	samLabText := "By SamLab🧋"
	samLabStyle := lipgloss.NewStyle().Foreground(mu).Bold(true).Underline(true)
	samLabHoverStyle := lipgloss.NewStyle().Foreground(p).Bold(true).Underline(true)
	samLabRendered := samLabStyle.Render(samLabText)
	samLabHoverRendered := samLabHoverStyle.Render(samLabText)

	samLabStartX := s.UI.Width - lipgloss.Width(samLabText) - sidePadding
	zoneManager.Register(input.Zone{
		ID:     "samlab",
		Type:   input.ZoneTypeLink,
		X:      samLabStartX,
		Y:      s.UI.Height - 1,
		Width:  lipgloss.Width(samLabText),
		Height: 1,
		OnClick: func() tea.Cmd {
			s.UI.ShowSamLab = true
			return nil
		},
	})
	zoneManager.UpdateMousePos(s.UI.MouseX, s.UI.MouseY)
	if zoneManager.IsHovered("samlab") {
		samLabRendered = samLabHoverRendered
	}

	// Register footer button zones BEFORE rendering footer
	footerY = s.UI.Height - 1
	cursorX = 0

	for _, info := range footerZoneInfos {
		zoneID := info.id
		var onClick func() tea.Cmd
		switch zoneID {
		case "footer-help":
			onClick = func() tea.Cmd {
				s.UI.ShowHelp = !s.UI.ShowHelp
				return nil
			}
		case "footer-settings":
			onClick = func() tea.Cmd {
				s.UI.ShowSettings = !s.UI.ShowSettings
				return nil
			}
		case "footer-filter":
			onClick = func() tea.Cmd {
				s.Process.FilterMode = true
				return nil
			}
		case "footer-kill":
			onClick = func() tea.Cmd {
				if len(s.Process.Processes) > 0 {
					procs := s.GetFilteredProcesses()
					if s.Process.SelectedProcess < len(procs) {
						proc := procs[s.Process.SelectedProcess]
						s.UI.ShowKillDialog = true
						s.Process.KillTargetPid = proc.Pid
						s.Process.KillTargetName = proc.Name
						s.UI.KillDialogSel = 0
					}
				}
				return nil
			}
		case "footer-sort":
			onClick = func() tea.Cmd {
				s.Process.SortBy = cycleSort(s.Process.SortBy)
				s.Config.SortBy = s.Process.SortBy
				return nil
			}
		case "footer-quit":
			onClick = func() tea.Cmd {
				return func() tea.Msg { return msg.QuitMsg{} }
			}
		case "footer-apply":
			onClick = func() tea.Cmd {
				s.Process.FilterMode = false
				s.Process.ProcessFilter = ""
				s.Process.ProcessFilterLower = ""
				return nil
			}
		}
		zoneManager.Register(input.Zone{
			ID:      info.id,
			Type:    input.ZoneTypeButton,
			X:       cursorX,
			Y:       footerY,
			Width:   info.width,
			Height:  1,
			OnClick: onClick,
		})
		cursorX += info.width
	}

	zoneManager.UpdateMousePos(s.UI.MouseX, s.UI.MouseY)

	// Rebuild footerContent with hover styles
	var newFooterContent []string

	for _, info := range footerZoneInfos {
		var btnStyle lipgloss.Style
		switch info.id {
		case "footer-help":
			btnStyle = helpBtnStyle
		case "footer-settings":
			btnStyle = settingsBtnStyle
		case "footer-quit":
			btnStyle = quitBtnStyle
		case "footer-apply":
			btnStyle = applyBtnStyle
		case "footer-filter":
			btnStyle = filterBtnStyle
		case "footer-kill":
			btnStyle = killBtnStyle
		case "footer-sort":
			btnStyle = sortBtnStyle
		default:
			btnStyle = quitBtnStyle
		}

		if zoneManager.IsHovered(info.id) {
			hoverStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Secondary).
				Background(bg).
				Underline(true).
				Padding(0, 1)
			newFooterContent = append(newFooterContent, hoverStyle.Render(info.text))
		} else {
			newFooterContent = append(newFooterContent, btnStyle.Render(info.text))
		}
	}

	// Add non-button footer items (like the filter hint)
	if activeTab == "Processes" && s.Process.FilterMode {
		newFooterContent = append(newFooterContent, dimBtnStyle.Render(" • Type to filter"))
	}

	// Render footer with hover-styled content
	if alertAtBottom && alertStr != "" {
		footerLeft := lipgloss.JoinHorizontal(lipgloss.Bottom, newFooterContent...)
		footerContentJoin := lipgloss.JoinHorizontal(lipgloss.Bottom, footerLeft, alertStr)
		footerLeftWidth := lipgloss.Width(footerContentJoin)
		footerWidth := s.UI.Width
		spacerWidth := footerWidth - footerLeftWidth - lipgloss.Width(samLabText)
		if spacerWidth < 1 {
			spacerWidth = 1
		}
		spacer := strings.Repeat(" ", spacerWidth)
		footerContentJoin = footerContentJoin + spacer + samLabRendered
		footer = lipgloss.NewStyle().Render(footerContentJoin)
	} else {
		footerLeft := lipgloss.JoinHorizontal(lipgloss.Bottom, newFooterContent...)
		footerLeftWidth := lipgloss.Width(footerLeft)
		footerWidth := s.UI.Width
		spacerWidth := footerWidth - footerLeftWidth - lipgloss.Width(samLabText)
		if spacerWidth < 1 {
			spacerWidth = 1
		}
		spacer := strings.Repeat(" ", spacerWidth)
		footerContentJoin := footerLeft + spacer + samLabRendered
		footer = lipgloss.NewStyle().
			Foreground(mu).
			Render(footerContentJoin)
	}

	// Calculate Content Area Height
	topGap := 1
	topPad := lipgloss.NewStyle().Height(topGap).Render("")

	topBarH := lipgloss.Height(topBar)
	footerH := lipgloss.Height(footer)

	reservedHeight := topBarH + topGap + footerH
	availHeight := max(s.UI.Height-reservedHeight, 5)

	listStartY := topBarH + topGap

	zoneManager.UpdateMousePos(s.UI.MouseX, s.UI.MouseY)

	titleStyle := lipgloss.NewStyle().Foreground(p).Bold(true).MarginBottom(1)
	labelStyle := lipgloss.NewStyle().Foreground(mu)
	valueStyle := lipgloss.NewStyle().Foreground(t).Bold(true)
	w := theme.Warning
	sColor := theme.Secondary

	var content string
	container := lipgloss.NewStyle().
		Border(border).
		BorderForeground(b).
		Padding(0, 1)

	switch activeTab {
	case "Metrics":
		content = tabs.RenderMetrics(s, container, su, w, a, sColor, t, mu, p, b, availHeight)
	case "Processes":
		visibleProcs, treeIndents := s.GetVisibleProcesses()
		mouseY := s.UI.MouseY
		if s.Process.ShowProcessMenu || s.UI.ShowKillDialog {
			mouseY = -1
		}

		content = tabs.RenderProcesses(s, visibleProcs, treeIndents, container, su, w, a, t, mu, p, b, availHeight, mouseY, listStartY, zoneManager)
	case "Disks":
		content = tabs.RenderDisks(s, container, su, w, a, t, mu, p, b, availHeight)
	case "Network":
		content = tabs.RenderNetwork(s, container, titleStyle, labelStyle, valueStyle, t, mu, p, b, bg, availHeight)
	case "System":
		content = tabs.RenderSystem(s, container, titleStyle, labelStyle, valueStyle, t, mu, p, b, bg, su, w, a, availHeight, s.UI.ActiveScrollBlock)

	case "Services":
		content = tabs.RenderServices(s, container, su, w, a, t, mu, p, b, availHeight)
	case "Connections":
		content = tabs.RenderConnections(s, container, su, w, a, t, mu, p, b, availHeight)
	case "Logs":
		content = tabs.RenderLogs(s, container, su, w, a, t, mu, p, b, availHeight)
	case "Remote":
		content = tabs.RenderRemote(s, container, su, w, a, t, mu, p, b, availHeight)
	default:
		content = lipgloss.NewStyle().Foreground(mu).Render("Tab not found: " + activeTab)
	}

	ui := lipgloss.JoinVertical(lipgloss.Left,
		topBar,
		topPad,
		content,
	)

	currH := lipgloss.Height(ui)

	rem := max(s.UI.Height-currH-footerH, 0)

	if rem > 0 {
		ui = lipgloss.JoinVertical(lipgloss.Left,
			ui,
			strings.Repeat("\n", rem-1),
			footer,
		)
	} else {
		ui = lipgloss.JoinVertical(lipgloss.Left,
			ui,
			footer,
		)
	}

	baseLayer := lipgloss.NewLayer(ui).Z(0)

	var layers []*lipgloss.Layer
	layers = append(layers, baseLayer)

	if len(s.UI.Toasts) > 0 {
		var toastBlocks []string
		for _, toast := range s.UI.Toasts {
			color := "#10B981" // Green
			icon := "✔ "
			if toast.Level == "error" {
				color = "#EF4444"
				icon = "✖ "
			}
			if toast.Level == "warn" {
				color = "#F59E0B"
				icon = "⚠"
			}

			block := lipgloss.NewStyle().
				Border(border).
				BorderForeground(lipgloss.Color(color)).
				Padding(0, 1).
				Render(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(icon + " " + toast.Message))
			toastBlocks = append(toastBlocks, block)
		}
		toastStack := lipgloss.JoinVertical(lipgloss.Right, toastBlocks...)
		toastWidth := lipgloss.Width(toastStack)
		toastHeight := lipgloss.Height(toastStack)

		toastX := s.UI.Width - toastWidth - sidePadding
		toastY := s.UI.Height - footerH - toastHeight - 1

		if toastX < 0 {
			toastX = 0
		}
		if toastY < 0 {
			toastY = 0
		}

		toastLayer := lipgloss.NewLayer(toastStack).X(toastX).Y(toastY).Z(2)
		layers = append(layers, toastLayer)
	}

	if s.UI.ShowKillDialog {
		killDialog := overlays.RenderKillDialog(s, b, p, a, t, mu, bg, zoneManager)
		dialogWidth := lipgloss.Width(killDialog)
		dialogHeight := lipgloss.Height(killDialog)

		dialogX := (s.UI.Width - dialogWidth) / 2
		dialogY := (s.UI.Height - dialogHeight) / 2
		if dialogX < 0 {
			dialogX = 0
		}
		if dialogY < 0 {
			dialogY = 0
		}

		dialogLayer := lipgloss.NewLayer(killDialog).X(dialogX).Y(dialogY).Z(3)
		layers = append(layers, dialogLayer)

		// Register zones for kill dialog buttons
		// Button positions match those calculated in RenderKillDialog
		boxWidth := KillDialogDefaultWidth
		if boxWidth > s.UI.Width-4 {
			boxWidth = s.UI.Width - 4
		}
		containerWidth := boxWidth - 6
		buttonsTotalWidth := KillDialogButtonTotalWidth
		buttonsStartX := dialogX + 3 + (containerWidth-buttonsTotalWidth)/2
		buttonY := dialogY + KillDialogButtonYOffset

		zoneManager.Register(input.Zone{
			ID:     "kill-yes",
			Type:   input.ZoneTypeButton,
			X:      buttonsStartX,
			Y:      buttonY,
			Width:  7,
			Height: 1,
			OnClick: func() tea.Cmd {
				// Kill confirmed - handled by keyboard, but zone enables click
				return nil
			},
		})

		zoneManager.Register(input.Zone{
			ID:     "kill-no",
			Type:   input.ZoneTypeButton,
			X:      buttonsStartX + 10,
			Y:      buttonY,
			Width:  6,
			Height: 1,
			OnClick: func() tea.Cmd {
				s.UI.ShowKillDialog = false
				s.Process.KillTargetPid = 0
				s.Process.KillTargetName = ""
				s.UI.KillDialogSel = 0
				return nil
			},
		})
	}

	if s.UI.ShowHelp {
		helpBox := overlays.RenderHelp(s, b, p, bg)
		hWidth := lipgloss.Width(helpBox)
		hHeight := lipgloss.Height(helpBox)
		hX := (s.UI.Width - hWidth) / 2
		hY := (s.UI.Height - hHeight) / 2
		if hX < 0 {
			hX = 0
		}
		if hY < 0 {
			hY = 0
		}
		layers = append(layers, lipgloss.NewLayer(helpBox).X(hX).Y(hY).Z(4))
	}

	if s.UI.ShowSettings {
		settingsBox := overlays.RenderSettingsOverlay(s, s.UI.Width, s.UI.Height, b, p, t, mu, bg)
		sWidth := lipgloss.Width(settingsBox)
		sHeight := lipgloss.Height(settingsBox)
		sX := (s.UI.Width - sWidth) / 2
		sY := (s.UI.Height - sHeight) / 2
		if sX < 0 {
			sX = 0
		}
		if sY < 0 {
			sY = 0
		}
		layers = append(layers, lipgloss.NewLayer(settingsBox).X(sX).Y(sY).Z(5))

		// Register zones for settings items
		// Settings are rendered in a 3-column layout with items at specific Y positions
		contentStartY := sY + 2 // Border + padding
		colWidth := (sWidth - 6) / 3

		// Column 1: Thresholds & Display
		thresholdDisplayCount := data.ThresholdCount + data.DisplayCount
		tabsStart := thresholdDisplayCount
		appearanceStart := tabsStart + data.TabCount
		col1X := sX + 3
		col1StartY := contentStartY + 1 // After header
		for i := 0; i < thresholdDisplayCount; i++ {
			itemIdx := i
			zoneManager.Register(input.Zone{
				ID:     fmt.Sprintf("settings-item-%d", itemIdx),
				Type:   input.ZoneTypeListItem,
				X:      col1X,
				Y:      col1StartY + i,
				Width:  colWidth,
				Height: 1,
				OnClick: func() tea.Cmd {
					s.UI.SettingsIdx = itemIdx
					return nil
				},
			})
		}

		// Column 2: Tabs
		col2X := col1X + colWidth
		col2StartY := contentStartY + 1
		for i := 0; i < data.TabCount; i++ {
			itemIdx := tabsStart + i
			zoneManager.Register(input.Zone{
				ID:     fmt.Sprintf("settings-item-%d", itemIdx),
				Type:   input.ZoneTypeListItem,
				X:      col2X,
				Y:      col2StartY + i,
				Width:  colWidth,
				Height: 1,
				OnClick: func() tea.Cmd {
					s.UI.SettingsIdx = itemIdx
					return nil
				},
			})
		}

		// Column 3: Appearance
		col3X := col2X + colWidth
		col3StartY := contentStartY + 1
		for i := 0; i < data.AppearanceCount; i++ {
			itemIdx := appearanceStart + i
			zoneManager.Register(input.Zone{
				ID:     fmt.Sprintf("settings-item-%d", itemIdx),
				Type:   input.ZoneTypeListItem,
				X:      col3X,
				Y:      col3StartY + i,
				Width:  colWidth,
				Height: 1,
				OnClick: func() tea.Cmd {
					s.UI.SettingsIdx = itemIdx
					return nil
				},
			})
		}
	}

	if s.UI.ShowSamLab {
		samLabBox := overlays.RenderSamLab(s, p, b, bg)
		slWidth := lipgloss.Width(samLabBox)
		slHeight := lipgloss.Height(samLabBox)
		slX := (s.UI.Width - slWidth) / 2
		slY := (s.UI.Height - slHeight) / 2
		if slX < 0 {
			slX = 0
		}
		if slY < 0 {
			slY = 0
		}
		layers = append(layers, lipgloss.NewLayer(samLabBox).X(slX).Y(slY).Z(5))
	}

	if s.Process.ShowOpenFiles {
		filesBox := overlays.RenderOpenFilesOverlay(s, s.UI.Width, s.UI.Height, b, p, t, mu, bg)
		fWidth := lipgloss.Width(filesBox)
		fHeight := lipgloss.Height(filesBox)
		fX := (s.UI.Width - fWidth) / 2
		fY := (s.UI.Height - fHeight) / 2
		if fX < 0 {
			fX = 0
		}
		if fY < 0 {
			fY = 0
		}
		layers = append(layers, lipgloss.NewLayer(filesBox).X(fX).Y(fY).Z(4))
	}

	// Process context menu (right-click menu)
	if s.Process.ShowProcessMenu && s.UI.SelectedTab < len(s.UI.ActiveTabs) && s.UI.ActiveTabs[s.UI.SelectedTab] == "Processes" {
		procs := s.GetFilteredProcesses()
		if s.Process.ProcessMenuIdx >= 0 && s.Process.ProcessMenuIdx < len(procs) {
			proc := procs[s.Process.ProcessMenuIdx]

			isSuspended := s.Process.SuspendedState[proc.Pid]

			var options []string
			if proc.Pid == 0 {
				options = []string{"[s]ort", "[f]ilter"}
			} else if isSuspended {
				options = []string{"[K]ill", "[x] Resume", "[o]pen Files", "[f]ilter", "[s]ort"}
			} else {
				options = []string{"[K]ill", "[z] Suspend", "[x] Resume", "[o]pen Files", "[f]ilter", "[s]ort"}
			}

			menuWidth := ContextMenuDefaultWidth
			if menuWidth > s.UI.Width {
				menuWidth = s.UI.Width
			}
			title := fmt.Sprintf("%.15s • Options", proc.Name)
			if len(title)+4 > menuWidth {
				menuWidth = len(title) + 4
			}
			menuHeight := len(options) + 2
			menuX := s.Process.ProcessMenuX
			menuY := s.Process.ProcessMenuY

			if menuX+menuWidth > s.UI.Width {
				menuX = s.UI.Width - menuWidth
			}
			if menuY+menuHeight > s.UI.Height {
				menuY = s.UI.Height - menuHeight
			}

			optionY := menuY + 1

			for i, opt := range options {
				menuItemID := fmt.Sprintf("menu-item-%d", i)
				menuOption := opt
				zoneManager.Register(input.Zone{
					ID:     menuItemID,
					Type:   input.ZoneTypeMenuItem,
					X:      menuX,
					Y:      optionY + i,
					Width:  menuWidth,
					Height: 1,
					OnClick: func() tea.Cmd {
						switch menuOption {
						case "[K]ill":
							if s.Process.ProcessMenuIdx >= 0 && s.Process.ProcessMenuIdx < len(procs) {
								targetProc := procs[s.Process.ProcessMenuIdx]
								s.UI.ShowKillDialog = true
								s.Process.KillTargetPid = targetProc.Pid
								s.Process.KillTargetName = targetProc.Name
								s.UI.KillDialogSel = 0
							}
						case "[z] Suspend":
							if s.Process.ProcessMenuIdx >= 0 && s.Process.ProcessMenuIdx < len(procs) {
								targetProc := procs[s.Process.ProcessMenuIdx]
								s.SetSuspended(targetProc.Pid, true)
							}
						case "[x] Resume":
							if s.Process.ProcessMenuIdx >= 0 && s.Process.ProcessMenuIdx < len(procs) {
								targetProc := procs[s.Process.ProcessMenuIdx]
								s.SetSuspended(targetProc.Pid, false)
							}
						case "[o]pen Files":
							if s.Process.ProcessMenuIdx >= 0 && s.Process.ProcessMenuIdx < len(procs) {
								targetProc := procs[s.Process.ProcessMenuIdx]
								s.Process.OpenFilesPid = targetProc.Pid
								s.Process.OpenFilesList = nil
								s.Process.ShowOpenFiles = true
								pid := targetProc.Pid
								s.Process.ShowProcessMenu = false
								return process.FetchOpenFilesCmd(pid)
							}
						case "[f]ilter":
							s.Process.FilterMode = true
						case "[s]ort":
							s.Process.SortBy = cycleSort(s.Process.SortBy)
							s.Config.SortBy = s.Process.SortBy
						}
						if s.Process.ShowProcessMenu {
							s.Process.ShowProcessMenu = false
						}
						return nil
					},
				})
			}

			// Update mouse position before hover checks
			zoneManager.UpdateMousePos(s.UI.MouseX, s.UI.MouseY)

			var optionLines []string
			for i, opt := range options {
				menuItemID := fmt.Sprintf("menu-item-%d", i)
				if zoneManager.IsHovered(menuItemID) {
					optionLines = append(optionLines, lipgloss.NewStyle().Foreground(bg).Background(p).Render(opt))
				} else {
					optionLines = append(optionLines, lipgloss.NewStyle().Foreground(p).Render(opt))
				}
			}

			menuContent := lipgloss.JoinVertical(lipgloss.Left, optionLines...)

			border := widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)
			menuBoxStyle := lipgloss.NewStyle().
				Border(border).
				BorderTop(false).
				BorderForeground(b).
				Padding(0, 1).
				Width(menuWidth)
			if s.Config.BackgroundOpaque {
				menuBoxStyle = menuBoxStyle.Background(bg)
			}
			menuBox := menuBoxStyle.Render(menuContent)

			actualWidth := lipgloss.Width(menuBox)
			topBorder := widgets.RenderTopBorderWithBg(title, actualWidth, border, b, p)

			finalMenu := lipgloss.JoinVertical(lipgloss.Left, topBorder, menuBox)

			layers = append(layers, lipgloss.NewLayer(finalMenu).X(menuX).Y(menuY).Z(10))
		}
	}

	// Create view from layers using v2 stable API
	compositor := lipgloss.NewCompositor(layers...)
	v := tea.NewView(compositor.Render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion

	if s.Config.BackgroundOpaque && !isNoColor(bg) {
		v.BackgroundColor = bg
	}

	s.UI.Zones = zoneManager.GetZones()
	s.UI.ZoneManager = zoneManager

	return v
}

// Helper function for rendering - gets theme from state
func getThemeFromState(s *data.AppState) ThemePalette {
	return GetAppTheme(s.Config.Theme, s.Config.Config.CustomTheme)
}

// Helper function for rendering - gets border from state
func getBorderFromState(s *data.AppState) lipgloss.Border {
	return widgets.GetBorder(s.Config.BorderStyle, s.Config.BorderType)
}

// RenderFromAppState is a convenience wrapper for main.go to use
func RenderFromAppState(s *data.AppState) tea.View {
	return MainViewFromState(s, func() lipgloss.Border {
		return getBorderFromState(s)
	}, func() ThemePalette {
		return getThemeFromState(s)
	})
}

// CustomThemeConfig re-export for backward compatibility
type CustomThemeConfig = config.CustomThemeConfig

func cycleSort(current string) string {
	switch current {
	case "cpu":
		return "mem"
	case "mem":
		return "pid"
	default:
		return "cpu"
	}
}

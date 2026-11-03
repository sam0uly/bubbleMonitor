package handlers

import (
	"reflect"

	tea "charm.land/bubbletea/v2"

	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
)

type ConfigChangeHandler interface {
	InvalidateProcessCache()
}

func HandleConfigWatchTick(m *data.AppState) tea.Cmd {
	return config.WatchConfig(m.Config.LastConfigModTime)
}

func HandleConfigChange(m *data.AppState, msg config.ConfigChangeMsg, invalidateCacheFunc func()) tea.Cmd {
	m.Config.LastConfigModTime = msg.NewModTime
	newConfig, err := config.LoadConfig()
	if err == nil {
		if reflect.DeepEqual(m.Config.Config, newConfig) {
			return config.WatchConfig(m.Config.LastConfigModTime)
		}

		m.Config.Config = newConfig
		m.Config.HistoryLength = newConfig.HistoryLength
		m.UI.ChartType = newConfig.ChartType
		m.Config.SortBy = newConfig.SortBy
		m.Config.SortDirection = newConfig.SortDirection
		m.Config.ProcessCpuNormalized = newConfig.ProcessCpuNormalized
		oldTreeView := m.Process.TreeView
		m.Process.TreeView = newConfig.ViewType == "tree"
		if oldTreeView != m.Process.TreeView {
			invalidateCacheFunc()
		}
		m.UI.ActiveTabs = newConfig.Tabs
		m.Config.Theme = newConfig.Theme
		m.Config.RefreshRate = newConfig.RefreshRate
		m.Config.BorderType = newConfig.BorderType
		m.Config.BorderStyle = newConfig.BorderStyle
		m.Config.BackgroundOpaque = newConfig.BackgroundOpaque
		return tea.Batch(config.WatchConfig(m.Config.LastConfigModTime), AddToastCmd("Config Reloaded", data.ToastSuccess))
	}
	return tea.Batch(config.WatchConfig(m.Config.LastConfigModTime), AddToastCmd("Config Error: "+err.Error(), data.ToastError))
}

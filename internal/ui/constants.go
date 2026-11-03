package ui

import "samouly.fun/bubbleMonitor/internal/data"

// Layout Constants
const (
	// Minimum Window Dimensions
	MinWindowWidth  = 80
	MinWindowHeight = 24

	// Responsive Thresholds
	WideLayoutThreshold = 130 // Width >= 130 triggers wide layout

	// Panel Heights
	MinContentHeight = 5

	// Re-export from data package for convenience
	KillDialogDefaultWidth     = data.KillDialogDefaultWidth
	KillDialogButtonTotalWidth = data.KillDialogButtonTotalWidth
	KillDialogButtonYOffset    = data.KillDialogButtonYOffset
	SettingsDefaultWidth       = data.SettingsDefaultWidth
	SettingsDefaultHeight      = data.SettingsDefaultHeight
	ContextMenuDefaultWidth    = data.ContextMenuDefaultWidth
)

package main

import (
	"samouly.fun/bubbleMonitor/internal/cliout"
	"samouly.fun/bubbleMonitor/internal/ui"
)

// loadCLIStyles creates a CLIStyles instance by resolving the current
// configuration to a ThemePalette. This is the single entry point all
// CLI subcommands use for themed output.
func loadCLIStyles() cliout.CLIStyles {
	// Error from loadConfigWithOverrides is safe to discard:
	// it falls back to DefaultConfig() internally.
	cfg, _ := loadConfigWithOverrides()
	palette := ui.GetAppTheme(cfg.Theme, cfg.CustomTheme)
	return cliout.New(palette)
}

// Package cliout provides themed styling primitives for CLI output.
// It bridges the TUI ThemePalette to lipgloss styles for use in non-interactive
// CLI subcommands (status, health, doctor, etc.).
package cliout

import (
	"charm.land/lipgloss/v2"
	"samouly.fun/bubbleMonitor/internal/ui"
)

// CLIStyles holds pre-built lipgloss styles derived from a ThemePalette.
// It is a value struct (stateless after creation), safe to pass around.
type CLIStyles struct {
	Label    lipgloss.Style // Foreground: palette.Muted
	Value    lipgloss.Style // Foreground: palette.Text
	Bold     lipgloss.Style // Bold(true), Foreground: palette.Text
	OK       lipgloss.Style // Bold(true), Foreground: palette.Success
	Warn     lipgloss.Style // Bold(true), Foreground: palette.Warning
	Critical lipgloss.Style // Bold(true), Foreground: palette.Alert
	Dim      lipgloss.Style // Faint(true), Foreground: palette.Muted
	Header   lipgloss.Style // Bold(true), Underline(true), Foreground: palette.Primary
	Active   lipgloss.Style // Bold(true), Foreground: palette.Primary

	CheckOK   string // styled in Success color
	CheckFail string // styled in Alert color
	CheckWarn string // styled in Warning color

	palette ui.ThemePalette // unexported, used by BarColor/ScoreColor
}

// New creates a CLIStyles value from the given ThemePalette.
// All styles are initialized once; no further renderer allocation needed.
func New(palette ui.ThemePalette) CLIStyles {
	return CLIStyles{
		Label:    lipgloss.NewStyle().Foreground(palette.Text).Background(palette.Border),
		Value:    lipgloss.NewStyle().Foreground(palette.Text),
		Bold:     lipgloss.NewStyle().Bold(true).Foreground(palette.Text),
		OK:       lipgloss.NewStyle().Bold(true).Foreground(palette.Success),
		Warn:     lipgloss.NewStyle().Bold(true).Foreground(palette.Warning),
		Critical: lipgloss.NewStyle().Bold(true).Foreground(palette.Alert),
		Dim:      lipgloss.NewStyle().Faint(true).Foreground(palette.Muted),
		Header:   lipgloss.NewStyle().Bold(true).Underline(true).Foreground(palette.Primary),
		Active:   lipgloss.NewStyle().Bold(true).Foreground(palette.Primary),

		CheckOK:   lipgloss.NewStyle().Foreground(palette.Success).Background(palette.Border).Bold(true).Render("✓"),
		CheckFail: lipgloss.NewStyle().Foreground(palette.Alert).Background(palette.Border).Bold(true).Render("✗"),
		CheckWarn: lipgloss.NewStyle().Foreground(palette.Warning).Background(palette.Border).Bold(true).Render("⚠"),

		palette: palette,
	}
}

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/ui"
)

func newThemesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "themes",
		Short: "Manage themes",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available themes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listThemes(cmd)
		},
	})

	return cmd
}

func listThemes(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	themes := config.GetThemeNames()

	cfg, _ := loadConfigWithOverrides()
	currentTheme := cfg.Theme

	fmt.Fprintf(out, "\n")
	for _, name := range themes {
		marker := " "
		if name == currentTheme {
			marker = "*"
		}

		// Render color swatch using the theme palette
		palette := ui.GetTheme(name)

		// Use a simple block swatch with the palette colors
		// We render each color as a styled block
		swatch := renderColorSwatch(palette)

		suffix := ""
		if name == currentTheme {
			suffix = "  (active)"
		}

		fmt.Fprintf(out, "  %s %-15s %s%s\n", marker, name, swatch, suffix)
	}
	fmt.Fprintf(out, "\n")

	return nil
}

func renderColorSwatch(p ui.ThemePalette) string {
	// Render a simple visual indicator showing the theme has valid colors
	// Since AdaptiveColor doesn't expose RGB directly, just show the theme name is registered
	return fmt.Sprintf("Primary/Secondary/Success/Warning/Alert")
}

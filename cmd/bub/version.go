package main

import (
	"charm.land/lipgloss/v2"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := loadCLIStyles()
			_, _ = lipgloss.Fprintf(cmd.OutOrStdout(), "%s %s (commit: %s, built: %s)\n", s.Label.Render("bub"), s.Value.Render(buildVersion), s.Value.Render(buildCommit), s.Value.Render(buildDate))
			return nil
		},
	}
}

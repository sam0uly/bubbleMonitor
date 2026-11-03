package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"charm.land/lipgloss/v2"
	"samouly.fun/bubbleMonitor/internal/app"
	configpkg "samouly.fun/bubbleMonitor/internal/config"
)

var (
	cfgConfigPath  string
	cfgTheme       string
	cfgRefreshRate int
	cfgHistoryLen  int

	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "bub",
		Short: "A terminal system monitor",
		Long:  "bubbleMonitor (bub) is a beautiful terminal system monitor built with BubbleTea.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return launchTUI()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&cfgConfigPath, "config", "", "config file path (default: ~/.config/bubble-monitor/config.json)")
	root.PersistentFlags().StringVar(&cfgTheme, "theme", "", "color theme (dark, light, nord, dracula, ...)")
	root.PersistentFlags().IntVar(&cfgRefreshRate, "refresh-rate", 0, "refresh interval in milliseconds")
	root.PersistentFlags().IntVar(&cfgHistoryLen, "history-length", 0, "data points to keep in history charts")

	_ = root.RegisterFlagCompletionFunc("theme", themeCompleter)
	_ = root.RegisterFlagCompletionFunc("refresh-rate", refreshRateCompleter)

	// root.AddGroup(&cobra.Group{ID: "monitor", Title: "Monitoring:"})
	// root.AddGroup(&cobra.Group{ID: "config", Title: "Configuration:"})
	// root.AddGroup(&cobra.Group{ID: "remote", Title: "Remote Hosts:"})
	// root.AddGroup(&cobra.Group{ID: "tools", Title: "Tools:"})

	root.AddCommand(newVersionCmd())
	root.AddCommand(newStatusCmd())
	root.AddCommand(newSysinfoCmd())
	root.AddCommand(newTopCmd())
	root.AddCommand(newPsCmd())
	root.AddCommand(newHealthCmd())
	root.AddCommand(newExportCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newThemesCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newRemoteCmd())

	return root
}

func Execute(version, commit, date string) error {
	buildVersion = version
	buildCommit = commit
	buildDate = date
	root := newRootCmd()
	return fang.Execute(
		context.Background(),
		root,
		fang.WithVersion(buildVersion),
		fang.WithCommit(buildCommit),
		fang.WithErrorHandler(func(w io.Writer, styles fang.Styles, err error) {
			line := lipgloss.JoinHorizontal(
				lipgloss.Center,
				styles.ErrorHeader.UnsetWidth().UnsetMargins().Margin(1, 0, 0, 2).String(),
				" ",
				styles.ErrorText.UnsetWidth().UnsetMargins().Render(err.Error()+"."),
			)
			_, _ = fmt.Fprintln(w, line)
			_, _ = fmt.Fprintln(w)

			if strings.Contains(err.Error(), "unknown command") || strings.Contains(err.Error(), "unknown flag") {
				_, _ = fmt.Fprintln(w, lipgloss.JoinHorizontal(
					lipgloss.Left,
					styles.ErrorText.UnsetWidth().Render("Try"),
					styles.Program.Flag.Render(" --help "),
					styles.ErrorText.UnsetWidth().UnsetMargins().UnsetTransform().Render("for usage."),
				))
			}
		}),
	)
}

func launchTUI() error {
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	p := tea.NewProgram(app.InitialModelWithConfig(cfg))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}

func loadConfigWithOverrides() (configpkg.AppConfig, error) {
	var cfg configpkg.AppConfig
	var err error

	if cfgConfigPath != "" {
		cfg, err = configpkg.LoadConfigFromPath(cfgConfigPath)
	} else {
		cfg, err = configpkg.LoadConfig()
	}
	if err != nil {
		cfg = configpkg.DefaultConfig()
	}

	if cfgTheme != "" {
		cfg.Theme = cfgTheme
	}
	if cfgRefreshRate > 0 {
		cfg.RefreshRate = cfgRefreshRate
	}
	if cfgHistoryLen > 0 {
		cfg.HistoryLength = cfgHistoryLen
	}

	return cfg, nil
}

func themeCompleter(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return configpkg.GetThemeNames(), cobra.ShellCompDirectiveNoFileComp
}

func refreshRateCompleter(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	rates := configpkg.GetRefreshRates()
	suggestions := make([]string, len(rates))
	for i, r := range rates {
		suggestions[i] = fmt.Sprintf("%d", r)
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

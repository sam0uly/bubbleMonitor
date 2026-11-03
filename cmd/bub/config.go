package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	configpkg "samouly.fun/bubbleMonitor/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigResetCmd())
	cmd.AddCommand(newConfigPathCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigWithOverrides()
			if err != nil {
				return err
			}
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: "Set a configuration value in the config file.\n" +
			"Keys: theme, refresh_rate, history_length, chart_type, border_type, border_style, " +
			"sort_by, sort_direction, background_opaque, process_cpu_normalized, default_tab, view_type, " +
			"threshold.cpu, threshold.memory, threshold.disk, threshold.temp",
		Example: "  bub config set theme nord\n" +
			"  bub config set refresh_rate 2000\n" +
			"  bub config set background_opaque true\n" +
			"  bub config set threshold.cpu 95.0",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			cfg, err := configpkg.LoadConfig()
			if err != nil {
				cfg = configpkg.DefaultConfig()
			}
			if err := setConfigField(&cfg, key, value); err != nil {
				return err
			}
			if err := configpkg.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			s := loadCLIStyles()
			_, _ = lipgloss.Fprintf(cmd.OutOrStdout(), "%s %s = %s\n", s.Label.Render("Set"), s.Label.Render(key), s.Value.Render(value))
			return nil
		},
	}
}

func newConfigResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			defaults := configpkg.DefaultConfig()
			if err := configpkg.SaveConfig(defaults); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			s := loadCLIStyles()
			_, _ = lipgloss.Fprintf(cmd.OutOrStdout(), "%s Configuration reset to defaults.\n", s.Label.Render("Reset"))
			return nil
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show config file path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configpkg.GetConfigPath()
			if err != nil {
				return err
			}
			s := loadCLIStyles()
			_, _ = lipgloss.Fprintf(cmd.OutOrStdout(), "%s\n", s.Label.Render(path))
			return nil
		},
	}
}

func setConfigField(cfg *configpkg.AppConfig, key, value string) error {
	switch key {
	case "theme":
		cfg.Theme = value
	case "refresh_rate":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer: %s", value)
		}
		cfg.RefreshRate = v
	case "history_length":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer: %s", value)
		}
		cfg.HistoryLength = v
	case "chart_type":
		cfg.ChartType = value
	case "border_type":
		cfg.BorderType = value
	case "border_style":
		cfg.BorderStyle = value
	case "sort_by":
		cfg.SortBy = value
	case "sort_direction":
		cfg.SortDirection = value
	case "background_opaque":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean: %s", value)
		}
		cfg.BackgroundOpaque = v
	case "process_cpu_normalized":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean: %s", value)
		}
		cfg.ProcessCpuNormalized = v
	case "default_tab":
		cfg.DefaultTab = value
	case "view_type":
		cfg.ViewType = value
	case "threshold.cpu":
		return setThreshold(cfg, configpkg.MetricCPU, value)
	case "threshold.memory":
		return setThreshold(cfg, configpkg.MetricMem, value)
	case "threshold.disk":
		return setThreshold(cfg, configpkg.MetricDisk, value)
	case "threshold.temp":
		return setThreshold(cfg, configpkg.MetricTemp, value)
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func setThreshold(cfg *configpkg.AppConfig, metric configpkg.MetricType, value string) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid float: %s", value)
	}
	if cfg.Thresholds == nil {
		cfg.Thresholds = configpkg.DefaultConfig().Thresholds
	}
	cfg.Thresholds[metric] = v
	return nil
}

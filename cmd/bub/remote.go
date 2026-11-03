package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	configpkg "samouly.fun/bubbleMonitor/internal/config"
)

var (
	remoteName    string
	remoteHost    string
	remoteKeyPath string
	remotePort    int
	remoteTimeout int
)

func newRemoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage remote SSH hosts",
	}

	cmd.AddCommand(newRemoteListCmd())
	cmd.AddCommand(newRemoteAddCmd())
	cmd.AddCommand(newRemoteRemoveCmd())

	return cmd
}

func newRemoteListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured remote hosts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigWithOverrides()
			if err != nil {
				return err
			}
			if len(cfg.RemoteHosts) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No remote hosts configured.")
				fmt.Fprintln(cmd.OutOrStdout(), "Use 'bub remote add' to add a host.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tHOST\tPORT\tKEY\tTIMEOUT")
			for _, h := range cfg.RemoteHosts {
				port := "-"
				if h.Port > 0 {
					port = fmt.Sprintf("%d", h.Port)
				}
				timeout := "-"
				if h.Timeout > 0 {
					timeout = fmt.Sprintf("%ds", h.Timeout)
				}
				key := "-"
				if h.KeyPath != "" {
					key = h.KeyPath
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", h.Name, h.Host, port, key, timeout)
			}
			w.Flush()
			return nil
		},
	}
}

func newRemoteAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a remote SSH host",
		Example: "  bub remote add --name prod --host user@prod.example.com\n" +
			"  bub remote add --name staging --host 192.168.1.50 --port 2222 --key ~/.ssh/staging_key",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if remoteName == "" {
				return fmt.Errorf("--name is required")
			}
			if remoteHost == "" {
				return fmt.Errorf("--host is required")
			}

			cfg, err := configpkg.LoadConfig()
			if err != nil {
				cfg = configpkg.DefaultConfig()
			}

			for _, h := range cfg.RemoteHosts {
				if h.Name == remoteName {
					return fmt.Errorf("remote host %q already exists", remoteName)
				}
			}

			cfg.RemoteHosts = append(cfg.RemoteHosts, configpkg.RemoteHostConfig{
				Name:    remoteName,
				Host:    remoteHost,
				KeyPath: remoteKeyPath,
				Port:    remotePort,
				Timeout: remoteTimeout,
			})

			if err := configpkg.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added remote host %q (%s)\n", remoteName, remoteHost)
			return nil
		},
	}

	cmd.Flags().StringVar(&remoteName, "name", "", "host name (required)")
	cmd.Flags().StringVar(&remoteHost, "host", "", "hostname or user@hostname (required)")
	cmd.Flags().StringVar(&remoteKeyPath, "key", "", "path to SSH private key")
	cmd.Flags().IntVar(&remotePort, "port", 0, "SSH port (default: 22)")
	cmd.Flags().IntVar(&remoteTimeout, "timeout", 0, "connection timeout in seconds")

	return cmd
}

func newRemoteRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Short:   "Remove a remote SSH host",
		Example: "  bub remote remove prod",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := configpkg.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			found := false
			filtered := make([]configpkg.RemoteHostConfig, 0, len(cfg.RemoteHosts))
			for _, h := range cfg.RemoteHosts {
				if h.Name == name {
					found = true
					continue
				}
				filtered = append(filtered, h)
			}

			if !found {
				return fmt.Errorf("remote host %q not found", name)
			}

			cfg.RemoteHosts = filtered
			if err := configpkg.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed remote host %q\n", name)
			return nil
		},
	}
}

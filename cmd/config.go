package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kfriede/hactl/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage configuration",
	Long: `View and manage hactl configuration.

Examples:
  hactl config show                 Show current configuration
  hactl config path                 Show config file path
  hactl config set host http://ha.local:8123   Set Home Assistant URL`,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configCheckCmd)
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		display := map[string]any{
			"profile": cfg.Profile,
			"host":    cfg.Host,
			"verbose": cfg.Verbose,
			"debug":   cfg.Debug,
		}

		secret, _ := config.GetSecret(cfg.Profile)
		if secret != "" {
			display["token"] = "(stored in keyring)"
		} else {
			display["token"] = "(not set)"
		}

		return printer.PrintResult(display)
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config directory path",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _ = fmt.Fprintln(os.Stdout, config.Dir())
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Supported keys: host

Examples:
  hactl config set host http://homeassistant.local:8123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
		value := args[1]

		switch key {
		case "host":
			cfg.Host = value
		default:
			return fmt.Errorf("unknown config key: %s (supported: host)", key)
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		printer.Success(fmt.Sprintf("Set %s = %s", key, value))
		return nil
	},
}

var configCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate Home Assistant configuration",
	Long: `Trigger a check of the Home Assistant configuration.yaml file.
Requires the config integration to be enabled.

Examples:
  hactl config check`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		respData, err := client.Post("/api/config/core/check_config", nil)
		if err != nil {
			return err
		}

		var result map[string]any
		if err := parseResponse(respData, &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

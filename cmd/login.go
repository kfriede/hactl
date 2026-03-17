package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kfriede/hactl/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Home Assistant",
	Long: `Store your Home Assistant credentials for API access.

Examples:
  hactl login                        Interactive login
  hactl login --profile work         Save as named profile

You need a Long-Lived Access Token from your Home Assistant instance.
Go to your HA profile page (Settings > People > your user > Security tab)
and create a Long-Lived Access Token.

The token is stored in your OS keyring when available,
falling back to the config file with restrictive permissions.`,
	Args: cobra.NoArgs,
	RunE: runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))
	if !isTTY {
		return fmt.Errorf("login requires an interactive terminal; set HACTL_TOKEN and HACTL_HOST environment variables for non-interactive use")
	}

	reader := bufio.NewReader(os.Stdin)

	// Prompt for host
	defaultHost := cfg.Host
	if defaultHost == "" {
		defaultHost = "http://homeassistant.local:8123"
	}
	fmt.Fprintf(os.Stderr, "Home Assistant URL [%s]: ", defaultHost)
	hostInput, _ := reader.ReadString('\n')
	host := strings.TrimSpace(hostInput)
	if host == "" {
		host = defaultHost
	}
	host = strings.TrimRight(host, "/")

	// Prompt for token
	fmt.Fprint(os.Stderr, "Long-Lived Access Token: ")
	tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return fmt.Errorf("reading token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Store secret in keyring
	profile := flagProfile
	if config.KeyringAvailable() {
		if err := config.StoreSecret(profile, token); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not store token in keyring: %v\n", err)
			fmt.Fprintln(os.Stderr, "Token will be stored in the config file instead.")
		} else {
			printer.Status("Token stored in OS keyring")
		}
	} else {
		printer.Status("OS keyring not available, storing token in config file")
	}

	// Save config
	newCfg := &config.Config{
		Profile: profile,
		Host:    host,
	}

	if err := config.Save(newCfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	printer.Success("Logged in successfully")
	printer.Status("Run `hactl status` to verify your connection.")
	return nil
}

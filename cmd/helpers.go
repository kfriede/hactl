package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kfriede/hactl/internal/api"
	"github.com/kfriede/hactl/internal/config"
	"github.com/kfriede/hactl/internal/output"
)

// newAPIClient creates an API client configured for Home Assistant.
func newAPIClient() (*api.Client, error) {
	token := ""

	// Try keyring first
	secret, err := config.GetSecret(cfg.Profile)
	if err == nil && secret != "" {
		token = secret
	}

	// Fall back to env var
	if token == "" {
		token = os.Getenv("HACTL_TOKEN")
	}

	if token == "" {
		printer.PrintError(output.NewAuthError("No API token configured"))
		return nil, fmt.Errorf("no token configured")
	}

	baseURL := strings.TrimRight(cfg.Host, "/")

	return api.NewClient(api.ClientConfig{
		Token:     token,
		BaseURL:   baseURL,
		Verbose:   cfg.Verbose,
		Debug:     cfg.Debug,
		ErrWriter: os.Stderr,
	}), nil
}

// confirmAction asks the user to confirm a destructive action.
func confirmAction(action string) bool {
	if flagYes {
		return true
	}

	fmt.Fprintf(os.Stderr, "Are you sure you want to %s? (y/N): ", action)

	var response string
	_, _ = fmt.Scanln(&response)
	return response == "y" || response == "yes" || response == "Y"
}

// parseJSONInput parses the --json-input flag value into a map.
func parseJSONInput(jsonStr string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}
	return result, nil
}

// printAPIResult handles the common pattern of printing API results.
func printAPIResult(data any) error {
	return printer.PrintResult(data)
}

// parseResponse unmarshals raw response bytes into the target.
func parseResponse(data []byte, target any) error {
	return json.Unmarshal(data, target)
}

// getToken retrieves the API token from keyring or environment.
func getToken() (string, error) {
	secret, err := config.GetSecret(cfg.Profile)
	if err == nil && secret != "" {
		return secret, nil
	}

	if token := os.Getenv("HACTL_TOKEN"); token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no token configured; run `hactl login` or set HACTL_TOKEN")
}

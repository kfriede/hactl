package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(haConfigCmd)
	rootCmd.AddCommand(coreStateCmd)
	rootCmd.AddCommand(streamCmd)

	streamCmd.Flags().String("restrict", "", "Comma-separated list of event types to stream (default: all)")
}

var haConfigCmd = &cobra.Command{
	Use:   "ha-config",
	Short: "Show Home Assistant server configuration",
	Long: `Retrieve the current Home Assistant server configuration including
version, location, time zone, units, loaded components, and more.

This returns the HA server's own configuration, not the local hactl config.

Examples:
  hactl ha-config
  hactl ha-config --fields version,location_name,time_zone`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result map[string]any
		if err := client.GetJSON("/api/config", &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var coreStateCmd = &cobra.Command{
	Use:   "core-state",
	Short: "Check Home Assistant core state",
	Long: `Retrieve the current Home Assistant core state.
Returns the running state and recorder migration status.
This is a lightweight endpoint used by the supervisor to check if HA is running.

Examples:
  hactl core-state`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result map[string]any
		if err := client.GetJSON("/api/core/state", &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream real-time events from Home Assistant",
	Long: `Connect to the Home Assistant event stream (Server-Sent Events).
Events are streamed in real-time until interrupted with Ctrl+C.

Examples:
  hactl stream                                     Stream all events
  hactl stream --restrict state_changed             Stream only state changes
  hactl stream --restrict state_changed,call_service Stream specific event types`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		restrict, _ := cmd.Flags().GetString("restrict")

		// Build URL manually since we need SSE handling
		baseURL := strings.TrimRight(cfg.Host, "/")
		path := "/api/stream"

		if restrict != "" {
			params := url.Values{}
			params.Set("restrict", restrict)
			path += "?" + params.Encode()
		}

		fullURL := baseURL + path

		// Get token
		token := ""
		secret, secretErr := getToken()
		if secretErr != nil {
			return secretErr
		}
		token = secret

		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "text/event-stream")

		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "[GET] %s\n", fullURL)
		}

		printer.Status("Connecting to event stream... (Ctrl+C to stop)")

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("connecting to stream: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("stream returned HTTP %d: %s", resp.StatusCode, string(body))
		}

		// Read SSE events
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "ping" {
					continue
				}

				// Pretty-print JSON events
				if flagJSON || flagOutput == "json" {
					var parsed any
					if err := json.Unmarshal([]byte(data), &parsed); err == nil {
						enc := json.NewEncoder(os.Stdout)
						enc.SetIndent("", "  ")
						_ = enc.Encode(parsed)
						continue
					}
				}
				fmt.Println(data)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading stream: %w", err)
		}

		return nil
	},
}

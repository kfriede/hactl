package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(intentCmd)
	intentCmd.AddCommand(intentHandleCmd)

	intentHandleCmd.Flags().String("json-input", "", "Full JSON intent payload")
	intentHandleCmd.Flags().String("name", "", "Intent name (e.g., SetTimer)")
}

var intentCmd = &cobra.Command{
	Use:   "intent",
	Short: "Handle Home Assistant intents",
	Long: `Handle intents for Home Assistant's conversation/voice integration.
Requires 'intent:' in your configuration.yaml.

Examples:
  hactl intent handle --name SetTimer --json-input '{"data":{"seconds":"30"}}'`,
}

var intentHandleCmd = &cobra.Command{
	Use:   "handle",
	Short: "Handle an intent",
	Long: `Handle an intent via the Home Assistant API.
Requires 'intent:' to be enabled in configuration.yaml.

Examples:
  hactl intent handle --name SetTimer --json-input '{"data":{"seconds":"30"}}'
  hactl intent handle --json-input '{"name":"SetTimer","data":{"seconds":"30"}}'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonInput, _ := cmd.Flags().GetString("json-input")
		name, _ := cmd.Flags().GetString("name")

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would handle intent: %s", name))
			return nil
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var body map[string]any
		if jsonInput != "" {
			body, err = parseJSONInput(jsonInput)
			if err != nil {
				return err
			}
		} else {
			body = make(map[string]any)
		}

		if name != "" {
			body["name"] = name
		}

		if body["name"] == nil && body["name"] == "" {
			return fmt.Errorf("--name or name in --json-input is required")
		}

		respData, err := client.Post("/api/intent/handle", body)
		if err != nil {
			return err
		}

		var result any
		if err := json.Unmarshal(respData, &result); err != nil {
			fmt.Print(string(respData))
			return nil
		}
		return printAPIResult(result)
	},
}

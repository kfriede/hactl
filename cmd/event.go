package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(eventCmd)
	eventCmd.AddCommand(eventListCmd)
	eventCmd.AddCommand(eventFireCmd)

	eventFireCmd.Flags().String("json-input", "", "JSON event data")
}

var eventCmd = &cobra.Command{
	Use:     "event",
	Aliases: []string{"events"},
	Short:   "Manage and fire Home Assistant events",
	Long: `List event types and fire events.

Examples:
  hactl event list                                  List all event types
  hactl event fire my_custom_event                  Fire an event
  hactl event fire my_event --json-input '{"key":"value"}'`,
}

var eventListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all event types",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result []map[string]any
		if err := client.GetJSON("/api/events", &result); err != nil {
			return err
		}

		items := make([]any, len(result))
		for i, r := range result {
			items[i] = r
		}
		return printAPIResult(items)
	},
}

var eventFireCmd = &cobra.Command{
	Use:   "fire <event_type>",
	Short: "Fire an event",
	Long: `Fire an event of the specified type with optional event data.

Examples:
  hactl event fire my_custom_event
  hactl event fire my_event --json-input '{"key":"value"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eventType := args[0]
		jsonInput, _ := cmd.Flags().GetString("json-input")

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would fire event: %s", eventType))
			if jsonInput != "" {
				printer.Status(fmt.Sprintf("  data: %s", jsonInput))
			}
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
		}

		respData, err := client.Post("/api/events/"+eventType, body)
		if err != nil {
			return err
		}

		var result map[string]any
		if err := json.Unmarshal(respData, &result); err != nil {
			fmt.Println(string(respData))
			return nil
		}
		return printAPIResult(result)
	},
}

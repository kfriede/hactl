package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(entityCmd)
	entityCmd.AddCommand(entityListCmd)
	entityCmd.AddCommand(entityGetCmd)
	entityCmd.AddCommand(entityUpdateCmd)
	entityCmd.AddCommand(entityDeleteCmd)
	entityCmd.AddCommand(entityHistoryCmd)

	// Update command flags
	entityUpdateCmd.Flags().String("state", "", "Entity state value")
	entityUpdateCmd.Flags().String("json-input", "", "Full JSON body with state and attributes")

	// History command flags
	entityHistoryCmd.Flags().String("start", "", "Start timestamp (YYYY-MM-DDThh:mm:ssTZD)")
	entityHistoryCmd.Flags().String("end", "", "End timestamp (YYYY-MM-DDThh:mm:ssTZD)")
	entityHistoryCmd.Flags().Bool("minimal", false, "Return minimal response (faster)")
	entityHistoryCmd.Flags().Bool("no-attributes", false, "Skip returning attributes (faster)")
	entityHistoryCmd.Flags().Bool("significant-only", false, "Only return significant state changes")
}

var entityCmd = &cobra.Command{
	Use:     "entity",
	Aliases: []string{"entities", "state", "states"},
	Short:   "Manage Home Assistant entities and states",
	Long: `Manage Home Assistant entities and their states.

Examples:
  hactl entity list                              List all entities
  hactl entity list --fields entity_id,state     List with selected fields
  hactl entity get sensor.temperature            Get entity details
  hactl entity update sensor.temp --state 25     Update entity state
  hactl entity delete sensor.temp --yes          Delete an entity
  hactl entity history sensor.temp               Get state history`,
}

var entityListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entities",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result []map[string]any
		if err := client.GetJSON("/api/states", &result); err != nil {
			return err
		}

		items := make([]any, len(result))
		for i, r := range result {
			items[i] = r
		}
		return printAPIResult(items)
	},
}

var entityGetCmd = &cobra.Command{
	Use:   "get <entity_id>",
	Short: "Get entity state and details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := args[0]
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result map[string]any
		if err := client.GetJSON("/api/states/"+entityID, &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var entityUpdateCmd = &cobra.Command{
	Use:   "update <entity_id>",
	Short: "Update or create an entity state",
	Long: `Update or create an entity state. This sets the representation
in Home Assistant — it does NOT communicate with the actual device.
To control devices, use 'hactl service call'.

Examples:
  hactl entity update sensor.temp --state 25
  hactl entity update sensor.temp --json-input '{"state":"25","attributes":{"unit_of_measurement":"°C"}}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := args[0]
		jsonInput, _ := cmd.Flags().GetString("json-input")

		if flagDryRun {
			if jsonInput != "" {
				printer.Status(fmt.Sprintf("[dry-run] Would update %s with: %s", entityID, jsonInput))
			} else {
				state, _ := cmd.Flags().GetString("state")
				printer.Status(fmt.Sprintf("[dry-run] Would update %s to state: %s", entityID, state))
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
		} else {
			state, _ := cmd.Flags().GetString("state")
			if state == "" {
				return fmt.Errorf("--state or --json-input is required")
			}
			body = map[string]any{"state": state}
		}

		respData, err := client.Post("/api/states/"+entityID, body)
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

var entityDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id>",
	Short: "Delete an entity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := args[0]

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete entity %s", entityID))
			return nil
		}

		if !confirmAction(fmt.Sprintf("delete entity %s", entityID)) {
			return fmt.Errorf("aborted")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		if err := client.Delete("/api/states/" + entityID); err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Deleted entity %s", entityID))
		return nil
	},
}

var entityHistoryCmd = &cobra.Command{
	Use:   "history <entity_id> [additional_entity_ids...]",
	Short: "Get state history for entities",
	Long: `Get state change history for one or more entities.

Examples:
  hactl entity history sensor.temperature
  hactl entity history sensor.temperature sensor.humidity
  hactl entity history sensor.temperature --start 2024-01-01T00:00:00Z
  hactl entity history sensor.temperature --start 2024-01-01T00:00:00Z --end 2024-01-02T00:00:00Z
  hactl entity history sensor.temperature --minimal --no-attributes`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		minimal, _ := cmd.Flags().GetBool("minimal")
		noAttrs, _ := cmd.Flags().GetBool("no-attributes")
		sigOnly, _ := cmd.Flags().GetBool("significant-only")

		path := "/api/history/period"
		if start != "" {
			path += "/" + start
		}

		params := url.Values{}
		params.Set("filter_entity_id", strings.Join(args, ","))
		if end != "" {
			params.Set("end_time", end)
		}
		if minimal {
			params.Set("minimal_response", "")
		}
		if noAttrs {
			params.Set("no_attributes", "")
		}
		if sigOnly {
			params.Set("significant_changes_only", "")
		}

		fullPath := path + "?" + params.Encode()

		var result [][]map[string]any
		if err := client.GetJSON(fullPath, &result); err != nil {
			return err
		}

		// Flatten into a single list for output
		var items []any
		for _, entityHistory := range result {
			for _, state := range entityHistory {
				items = append(items, state)
			}
		}
		return printAPIResult(items)
	},
}

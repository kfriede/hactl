package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceListCmd)
	serviceCmd.AddCommand(serviceCallCmd)

	serviceCallCmd.Flags().String("json-input", "", "JSON service data (e.g., '{\"entity_id\":\"light.living_room\"}')")
	serviceCallCmd.Flags().String("entity-id", "", "Target entity ID")
	serviceCallCmd.Flags().Bool("return-response", false, "Return service response data")
}

var serviceCmd = &cobra.Command{
	Use:     "service",
	Aliases: []string{"services", "svc"},
	Short:   "Manage and call Home Assistant services",
	Long: `List available services and call them.

Examples:
  hactl service list                                        List all services
  hactl service list --fields domain,services               List with selected fields
  hactl service call light turn_on --entity-id light.room   Call a service
  hactl service call switch turn_off --json-input '{"entity_id":"switch.fan"}'`,
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available services",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result []map[string]any
		if err := client.GetJSON("/api/services", &result); err != nil {
			return err
		}

		items := make([]any, len(result))
		for i, r := range result {
			items[i] = r
		}
		return printAPIResult(items)
	},
}

var serviceCallCmd = &cobra.Command{
	Use:   "call <domain> <service>",
	Short: "Call a Home Assistant service",
	Long: `Call a service within a specific domain.

Examples:
  hactl service call light turn_on --entity-id light.living_room
  hactl service call switch turn_off --entity-id switch.fan
  hactl service call script my_script
  hactl service call mqtt publish --json-input '{"payload":"OFF","topic":"home/fridge"}'
  hactl service call weather get_forecasts --json-input '{"entity_id":"weather.home","type":"daily"}' --return-response`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		service := args[1]
		jsonInput, _ := cmd.Flags().GetString("json-input")
		entityID, _ := cmd.Flags().GetString("entity-id")
		returnResponse, _ := cmd.Flags().GetBool("return-response")

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would call %s.%s", domain, service))
			if entityID != "" {
				printer.Status(fmt.Sprintf("  entity_id: %s", entityID))
			}
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
		} else {
			body = make(map[string]any)
		}

		if entityID != "" {
			body["entity_id"] = entityID
		}

		path := fmt.Sprintf("/api/services/%s/%s", domain, service)
		if returnResponse {
			path += "?return_response"
		}

		respData, err := client.Post(path, body)
		if err != nil {
			return err
		}

		// Try to parse as JSON and print
		var result any
		if err := json.Unmarshal(respData, &result); err != nil {
			// If not JSON, print raw
			fmt.Println(string(respData))
			return nil
		}

		return printAPIResult(result)
	},
}

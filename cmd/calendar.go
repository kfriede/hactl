package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(calendarCmd)
	calendarCmd.AddCommand(calendarListCmd)
	calendarCmd.AddCommand(calendarEventsCmd)

	calendarEventsCmd.Flags().String("start", "", "Start timestamp (ISO 8601, required)")
	calendarEventsCmd.Flags().String("end", "", "End timestamp (ISO 8601, required)")
	_ = calendarEventsCmd.MarkFlagRequired("start")
	_ = calendarEventsCmd.MarkFlagRequired("end")
}

var calendarCmd = &cobra.Command{
	Use:     "calendar",
	Aliases: []string{"calendars", "cal"},
	Short:   "Manage Home Assistant calendars",
	Long: `List calendar entities and view calendar events.

Examples:
  hactl calendar list
  hactl calendar events calendar.holidays --start 2024-01-01T00:00:00Z --end 2024-12-31T23:59:59Z`,
}

var calendarListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all calendar entities",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result []map[string]any
		if err := client.GetJSON("/api/calendars", &result); err != nil {
			return err
		}

		items := make([]any, len(result))
		for i, r := range result {
			items[i] = r
		}
		return printAPIResult(items)
	},
}

var calendarEventsCmd = &cobra.Command{
	Use:   "events <calendar_entity_id>",
	Short: "List events for a calendar",
	Long: `List events for a specific calendar entity between start and end times.

Examples:
  hactl calendar events calendar.holidays --start 2024-05-01T00:00:00Z --end 2024-06-01T00:00:00Z
  hactl calendar events calendar.personal --start 2024-01-01T00:00:00Z --end 2024-12-31T23:59:59Z`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := args[0]
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/api/calendars/%s?start=%s&end=%s", entityID, start, end)

		var result []map[string]any
		if err := client.GetJSON(path, &result); err != nil {
			return err
		}

		items := make([]any, len(result))
		for i, r := range result {
			items[i] = r
		}
		return printAPIResult(items)
	},
}

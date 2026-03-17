package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logbookCmd)

	logbookCmd.Flags().String("start", "", "Start timestamp (YYYY-MM-DDThh:mm:ssTZD)")
	logbookCmd.Flags().String("end", "", "End timestamp (YYYY-MM-DDThh:mm:ssTZD)")
	logbookCmd.Flags().String("entity", "", "Filter by entity ID")
}

var logbookCmd = &cobra.Command{
	Use:   "logbook",
	Short: "View Home Assistant logbook entries",
	Long: `View logbook entries, optionally filtered by time range and entity.

Examples:
  hactl logbook
  hactl logbook --start 2024-01-01T00:00:00Z
  hactl logbook --entity sensor.temperature
  hactl logbook --start 2024-01-01T00:00:00Z --end 2024-01-02T00:00:00Z --entity light.living_room`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		entity, _ := cmd.Flags().GetString("entity")

		path := "/api/logbook"
		if start != "" {
			path += "/" + start
		}

		params := url.Values{}
		if end != "" {
			params.Set("end_time", end)
		}
		if entity != "" {
			params.Set("entity", entity)
		}

		if len(params) > 0 {
			path += "?" + params.Encode()
		}

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

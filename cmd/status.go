package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Home Assistant API status",
	Long: `Check if the Home Assistant API is running and accessible.

Examples:
  hactl status`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result map[string]any
		if err := client.GetJSON("/api/", &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

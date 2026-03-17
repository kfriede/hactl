package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(errorLogCmd)
}

var errorLogCmd = &cobra.Command{
	Use:     "error-log",
	Aliases: []string{"errors", "errlog"},
	Short:   "View Home Assistant error log",
	Long: `Retrieve all errors logged during the current Home Assistant session.
Returns plaintext log output.

Examples:
  hactl error-log`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		result, err := client.GetRaw("/api/error_log")
		if err != nil {
			return err
		}

		fmt.Print(result)
		return nil
	},
}

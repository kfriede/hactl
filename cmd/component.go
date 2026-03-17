package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(componentCmd)
	componentCmd.AddCommand(componentListCmd)
}

var componentCmd = &cobra.Command{
	Use:     "component",
	Aliases: []string{"components", "comp"},
	Short:   "List Home Assistant components",
	Long: `List currently loaded Home Assistant components/integrations.

Examples:
  hactl component list`,
}

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List loaded components",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result []string
		if err := client.GetJSON("/api/components", &result); err != nil {
			return err
		}

		// Convert to structured output for the printer
		items := make([]any, len(result))
		for i, c := range result {
			items[i] = map[string]any{"component": c}
		}
		return printAPIResult(items)
	},
}

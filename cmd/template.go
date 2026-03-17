package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateRenderCmd)

	templateRenderCmd.Flags().StringP("file", "f", "", "Read template from file instead of argument")
	templateRenderCmd.Flags().String("variables", "", "JSON object of template variables")
}

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Render Home Assistant Jinja2 templates",
	Long: `Render Jinja2 templates using Home Assistant's template engine.

Examples:
  hactl template render 'The time is {{ now() }}'
  hactl template render '{{ states("sensor.temperature") }}°C'
  echo '{{ states.light | list | count }} lights' | hactl template render -
  hactl template render -f my_template.j2`,
}

var templateRenderCmd = &cobra.Command{
	Use:   "render <template>",
	Short: "Render a Jinja2 template",
	Long: `Render a Jinja2 template string using Home Assistant.
Use '-' to read from stdin.

Examples:
  hactl template render 'It is {{ now() }}!'
  hactl template render '{{ states("sensor.temperature") }}'
  echo '{{ states.light | list | count }}' | hactl template render -
  hactl template render -f template.j2`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var templateStr string
		fileFlag, _ := cmd.Flags().GetString("file")

		if fileFlag != "" {
			data, err := os.ReadFile(fileFlag)
			if err != nil {
				return fmt.Errorf("reading template file: %w", err)
			}
			templateStr = string(data)
		} else if len(args) == 1 && args[0] == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			templateStr = string(data)
		} else if len(args) == 1 {
			templateStr = args[0]
		} else {
			return fmt.Errorf("template string, --file, or '-' for stdin is required")
		}

		templateStr = strings.TrimSpace(templateStr)
		if templateStr == "" {
			return fmt.Errorf("template string is empty")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		body := map[string]any{"template": templateStr}

		// Add variables if provided
		varsStr, _ := cmd.Flags().GetString("variables")
		if varsStr != "" {
			vars, parseErr := parseJSONInput(varsStr)
			if parseErr != nil {
				return fmt.Errorf("invalid variables JSON: %w", parseErr)
			}
			body["variables"] = vars
		}

		respData, err := client.Post("/api/template", body)
		if err != nil {
			return err
		}

		// Template endpoint returns plain text
		fmt.Print(string(respData))
		return nil
	},
}

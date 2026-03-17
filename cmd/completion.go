package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for hactl.

To load completions:

Bash:
  $ source <(hactl completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ hactl completion bash > /etc/bash_completion.d/hactl
  # macOS:
  $ hactl completion bash > $(brew --prefix)/etc/bash_completion.d/hactl

Zsh:
  $ source <(hactl completion zsh)
  # To load completions for each session, execute once:
  $ hactl completion zsh > "${fpath[1]}/_hactl"

Fish:
  $ hactl completion fish | source
  # To load completions for each session, execute once:
  $ hactl completion fish > ~/.config/fish/completions/hactl.fish

PowerShell:
  PS> hactl completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		}
		return nil
	},
}

func init() {
	// Output format completion
	_ = rootCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "csv", "ndjson"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Dynamic entity ID completion
	entityCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, err := newAPIClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var states []map[string]any
		if err := client.GetJSON("/api/states", &states); err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var completions []string
		for _, s := range states {
			entityID, _ := s["entity_id"].(string)
			if entityID != "" && strings.HasPrefix(entityID, toComplete) {
				attrs, _ := s["attributes"].(map[string]any)
				name, _ := attrs["friendly_name"].(string)
				if name != "" {
					completions = append(completions, fmt.Sprintf("%s\t%s", entityID, name))
				} else {
					completions = append(completions, entityID)
				}
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	entityGetCmd.ValidArgsFunction = entityCompletion
	entityUpdateCmd.ValidArgsFunction = entityCompletion
	entityDeleteCmd.ValidArgsFunction = entityCompletion
	entityHistoryCmd.ValidArgsFunction = entityCompletion
	cameraSnapshotCmd.ValidArgsFunction = entityCompletion
	calendarEventsCmd.ValidArgsFunction = entityCompletion
}

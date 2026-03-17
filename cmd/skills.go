package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(skillsCmd)
}

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Dump agent-optimized usage instructions",
	Long: `Prints concise, agent-optimized usage instructions for LLM agents.

This is designed to be read once and internalized by an agent to
reduce hallucinations and improve command accuracy.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), skillsText)
		return nil
	},
}

const skillsText = `# hactl — Agent Skills

## Quick Reference
hactl <resource> <action> [flags]
Resources: entity, service, event, calendar, camera, component, intent
Other:     status, logbook, template, error-log, config

## Rules
- ALWAYS use --fields on list/get to limit output (saves tokens)
- ALWAYS use --dry-run before any mutating command, then confirm with user
- ALWAYS pass --yes on confirmed destructive actions (delete)
- Use --json-input for create/update payloads (avoids flag hallucination)
- Parse JSON from stdout; errors go to stderr as JSON with "guidance" field
- Non-TTY automatically outputs JSON — no need for --json in agent context

## Invariants
- Entity IDs use domain.name format (e.g., light.living_room, sensor.temperature)
- Service calls use domain + service (e.g., light turn_on)
- All timestamps are ISO 8601
- Non-TTY mode automatically outputs JSON

## Common Patterns

### Check API connectivity
hactl status

### List entities with field selection
hactl entity list --fields entity_id,state

### Get entity details
hactl entity get sensor.temperature

### Call a service (control devices)
hactl service call light turn_on --entity-id light.living_room
hactl service call switch turn_off --json-input '{"entity_id":"switch.fan"}'

### Update entity state (representation only, not device)
hactl entity update sensor.temp --dry-run --state 25
hactl entity update sensor.temp --state 25

### Delete entity
hactl entity delete sensor.temp --dry-run
hactl entity delete sensor.temp --yes

### Fire events
hactl event fire my_custom_event --json-input '{"data":"value"}'

### View history
hactl entity history sensor.temperature --start 2024-01-01T00:00:00Z

### Render templates
hactl template render '{{ states("sensor.temperature") }}'

### View logs
hactl error-log
hactl logbook --entity sensor.temperature

### Camera snapshots
hactl camera snapshot camera.front_door -O snapshot.jpg

### Calendar
hactl calendar list
hactl calendar events calendar.holidays --start 2024-01-01T00:00:00Z --end 2024-12-31T23:59:59Z

### Validate HA config
hactl config check

### Runtime introspection
hactl schema                      # list all commands
hactl schema service.call         # full schema for a specific command

## Error Handling
Errors include structured JSON on stderr:
{"code":"AUTH_ERROR","message":"Token expired","guidance":"Run 'hactl login' to re-authenticate."}

Exit codes: 0=success, 1=general error, 2=auth error, 3=not found, 4=conflict/validation

## Configuration
HACTL_TOKEN, HACTL_HOST, HACTL_OUTPUT_FORMAT=json, HACTL_DEBUG=1
`

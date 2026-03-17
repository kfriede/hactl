---
name: hactl-manager
description: "Manage your Home Assistant instance using the hactl CLI"
usage: "Ask the agent to control devices, check states, view history, or manage your smart home"
arguments:
  - name: profile
    description: Configuration profile name
    type: string
examples:
  - input: "Turn on the living room lights"
    output: "Running `hactl service call light turn_on --entity-id light.living_room` to turn on the lights."
  - input: "What's the temperature?"
    output: "Running `hactl entity get sensor.temperature --fields entity_id,state` to check the temperature."
  - input: "Show me the front door camera"
    output: "Running `hactl camera snapshot camera.front_door -O snapshot.jpg` to capture a snapshot."
---

# Home Assistant Manager Skill

Manage your Home Assistant instance using the `hactl` CLI. This skill enables the agent to control devices, view entity states, check history, fire events, and more through structured CLI commands with safety rails.

## Prerequisites

Install `hactl` and ensure it's available in PATH:
```bash
go install github.com/kfriede/hactl@latest
```

Configure access:
```bash
hactl login
# Or set environment variables:
export HACTL_HOST=http://homeassistant.local:8123
export HACTL_TOKEN=<your-long-lived-access-token>
```

> **Note:** This plugin requires local execution (e.g., Claude Code) with network access to your Home Assistant instance.

## How to Use hactl

Command pattern: `hactl <resource> <action> [flags]`

### Discover commands
```bash
hactl schema                         # list all available commands
hactl schema service.call            # full schema for a specific command
hactl skills                         # complete agent reference
```

### Read operations (always safe)
```bash
hactl status                                    # check API connectivity
hactl entity list --fields entity_id,state      # list entities
hactl entity get sensor.temperature             # get entity details
hactl service list                              # list available services
hactl event list                                # list event types
hactl logbook --entity sensor.temperature       # view logbook
hactl entity history sensor.temp --start 2024-01-01T00:00:00Z  # state history
hactl error-log                                 # view error log
hactl calendar list                             # list calendars
hactl component list                            # list components
```

### Device control (service calls)
```bash
hactl service call light turn_on --dry-run --entity-id light.living_room
# show preview, then:
hactl service call light turn_on --entity-id light.living_room
```

### State updates (representation only)
```bash
hactl entity update sensor.temp --dry-run --state 25
hactl entity update sensor.temp --state 25
```

### Destructive operations (require --yes)
```bash
hactl entity delete sensor.temp --dry-run    # preview
hactl entity delete sensor.temp --yes        # execute
```

## Rules

- **ALWAYS** use `--fields` on list/get commands to limit output
- **ALWAYS** use `--dry-run` before any mutating command, show the preview, and ask for confirmation
- **ALWAYS** pass `--yes` for confirmed destructive actions
- **ALWAYS** use `--json-input` for update payloads with attributes
- **NEVER** parse table-formatted output — non-TTY mode auto-outputs JSON
- **NEVER** omit `--yes` on destructive commands (will hang in non-TTY)

## Error Handling

Errors include structured JSON on stderr with a `guidance` field:
```json
{"code":"AUTH_ERROR","message":"Token is invalid","guidance":"Run `hactl login` to authenticate."}
```

Exit codes: 0=success, 1=general, 2=auth, 3=not found, 4=conflict

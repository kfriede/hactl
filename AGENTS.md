# hactl — Agent Specification

## Overview

`hactl` is a CLI for Home Assistant covering every REST API endpoint.
Command pattern: `hactl <resource> <action> [flags]`

## Resources

| Resource | Actions | API Endpoints |
|---|---|---|
| entity | list, get, update, delete, history | /api/states, /api/history/period |
| service | list, call | /api/services |
| event | list, fire | /api/events |
| calendar | list, events | /api/calendars |
| camera | snapshot | /api/camera_proxy |
| component | list | /api/components |
| intent | handle | /api/intent/handle |
| logbook | (direct) | /api/logbook |
| template | render | /api/template |
| error-log | (direct) | /api/error_log |
| config | show, set, path, check | /api/config/core/check_config |
| status | (direct) | /api/ |

## Rules

1. **ALWAYS** use `--fields` on list/get to limit output (saves tokens)
2. **ALWAYS** use `--dry-run` before any mutating command, then confirm
3. **ALWAYS** pass `--yes` on confirmed destructive actions
4. **ALWAYS** use `--json-input` for update payloads
5. **NEVER** parse table output — non-TTY auto-outputs JSON
6. **NEVER** omit `--yes` on destructive commands (will hang in non-TTY)

## Output

- stdout = data (JSON in non-TTY, table in TTY)
- stderr = errors, logs, status messages
- Errors: `{"code":"AUTH_ERROR","message":"...","guidance":"..."}`
- Exit codes: 0=success, 1=general, 2=auth, 3=not found, 4=conflict

## Configuration

| Variable | Description |
|---|---|
| `HACTL_HOST` | Home Assistant URL |
| `HACTL_TOKEN` | Long-Lived Access Token |
| `HACTL_OUTPUT_FORMAT` | Default output format |
| `HACTL_DEBUG` | Enable debug logging |

## Common Workflows

### Check connectivity
```
hactl status
```

### List entities (token-efficient)
```
hactl entity list --fields entity_id,state
```

### Control a device
```
hactl service call light turn_on --dry-run --entity-id light.living_room
# show preview, then:
hactl service call light turn_on --entity-id light.living_room
```

### Get entity history
```
hactl entity history sensor.temperature --start 2024-01-01T00:00:00Z --minimal
```

### Render a template
```
hactl template render '{{ states("sensor.temperature") }}'
```

### Delete entity (destructive)
```
hactl entity delete sensor.temp --dry-run
# show preview, then:
hactl entity delete sensor.temp --yes
```

### Introspection
```
hactl schema                    # list all commands
hactl schema service.call       # full schema for a command
hactl skills                    # complete agent reference
```

## Architecture Invariants

- Entity IDs: `domain.object_id` format (e.g., `light.living_room`)
- Service calls: `hactl service call <domain> <service>`
- All timestamps: ISO 8601
- Config precedence: CLI flags → env vars → config file
- Secrets stored in OS keyring (fallback: config file with 0600 permissions)

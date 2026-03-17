<p align="center">
  <strong>hactl</strong> — Command-Line Interface for Home Assistant
</p>

<p align="center">
  <a href="#getting-started">Getting Started</a> •
  <a href="#commands">Commands</a> •
  <a href="#output-formats">Output Formats</a> •
  <a href="#for-llm-agents">For LLM Agents</a>
</p>

---

A fully featured CLI for [Home Assistant](https://www.home-assistant.io/) that covers every REST API endpoint. Built for both humans and LLM agents.

```bash
# Human sees a table
$ hactl entity list --fields entity_id,state
ENTITY_ID                STATE
───────────────────────  ──────
light.living_room        on
sensor.temperature       21.5
switch.fan               off

# Agent gets JSON automatically (non-TTY)
$ hactl entity list --fields entity_id,state | jq '.[].entity_id'
"light.living_room"
"sensor.temperature"
"switch.fan"
```

## Getting Started

### Install

```bash
# From source
go install github.com/kfriede/hactl@latest

# Or build from source
git clone https://github.com/kfriede/hactl.git
cd hactl
make build
```

### Configure

```bash
# Interactive login — prompts for HA URL and Long-Lived Access Token
hactl login

# Or use environment variables
export HACTL_HOST=http://homeassistant.local:8123
export HACTL_TOKEN=your_long_lived_access_token
```

Get a Long-Lived Access Token from your Home Assistant profile page:
**Settings → People → Your User → Security tab → Create Token**

### Verify

```bash
hactl status
# {"message": "API running."}
```

## Commands

### Entity Management (States)

```bash
hactl entity list                              # List all entities
hactl entity list --fields entity_id,state     # Select specific fields
hactl entity get sensor.temperature            # Get entity details
hactl entity update sensor.temp --state 25     # Update entity state
hactl entity update sensor.temp --json-input '{"state":"25","attributes":{"unit_of_measurement":"°C"}}'
hactl entity delete sensor.temp --yes          # Delete an entity
hactl entity history sensor.temp               # Get state history
hactl entity history sensor.temp --start 2024-01-01T00:00:00Z --end 2024-02-01T00:00:00Z --minimal
```

### Service Calls (Device Control)

```bash
hactl service list                                         # List all services
hactl service call light turn_on --entity-id light.room    # Turn on a light
hactl service call switch turn_off --entity-id switch.fan  # Turn off a switch
hactl service call script my_script                        # Run a script
hactl service call mqtt publish --json-input '{"payload":"OFF","topic":"home/fridge"}'
hactl service call weather get_forecasts --json-input '{"entity_id":"weather.home","type":"daily"}' --return-response
```

### Events

```bash
hactl event list                                           # List event types
hactl event fire my_custom_event                           # Fire an event
hactl event fire my_event --json-input '{"key":"value"}'   # Fire with data
```

### Logbook

```bash
hactl logbook                                              # Recent logbook
hactl logbook --start 2024-01-01T00:00:00Z                 # From timestamp
hactl logbook --entity sensor.temperature                  # Filter by entity
hactl logbook --start 2024-01-01T00:00:00Z --end 2024-01-02T00:00:00Z
```

### Templates

```bash
hactl template render 'It is {{ now() }}!'
hactl template render '{{ states("sensor.temperature") }}°C'
hactl template render -f my_template.j2
echo '{{ states.light | list | count }} lights' | hactl template render -
```

### Error Log

```bash
hactl error-log                                            # View HA error log
```

### Camera

```bash
hactl camera snapshot camera.front_door                    # Save snapshot
hactl camera snapshot camera.front_door -O snapshot.jpg    # Custom filename
```

### Calendar

```bash
hactl calendar list                                        # List calendars
hactl calendar events calendar.holidays --start 2024-05-01T00:00:00Z --end 2024-06-01T00:00:00Z
```

### Components

```bash
hactl component list                                       # List loaded components
```

### Intents

```bash
hactl intent handle --name SetTimer --json-input '{"data":{"seconds":"30"}}'
```

### Configuration

```bash
hactl config show                                          # Show current config
hactl config path                                          # Show config directory
hactl config set host http://ha.local:8123                 # Set HA URL
hactl config check                                         # Validate HA configuration.yaml
```

### Other

```bash
hactl status                                               # Check API connectivity
hactl ha-config                                            # Show HA server configuration
hactl core-state                                           # Check HA core running state
hactl stream                                               # Stream real-time events (SSE)
hactl stream --restrict state_changed                      # Stream only state changes
hactl version                                              # Print version info
hactl schema                                               # List all command schemas
hactl schema service.call                                  # Schema for specific command
hactl skills                                               # Agent usage reference
hactl completion bash                                      # Generate completions
```

## Output Formats

| Context | What you get |
|---|---|
| **Terminal (TTY)** | Colored, aligned tables |
| **Piped / agent** | JSON (automatic) |
| `--json` | Force JSON |
| `--csv` | CSV |
| `--output ndjson` | One JSON object per line |
| `--fields entity_id,state` | Field mask (all formats) |

**stdout** is always data. Logs, progress, and errors go to **stderr**.

## Safety

Every mutating command supports `--dry-run` and `--yes`:

```bash
$ hactl entity delete sensor.temp --dry-run    # preview
$ hactl entity delete sensor.temp --yes        # execute (skip prompt)
$ hactl service call light turn_on --dry-run --entity-id light.room  # preview
```

## Configuration

| Source | Example |
|---|---|
| CLI flag | `--host http://ha:8123` |
| Environment | `HACTL_HOST=http://ha:8123` |
| Config file | `~/.config/hactl/config.yaml` |
| Keyring | Token stored via `hactl login` |

Precedence: CLI flags → environment variables → config file.

### Profiles

```bash
hactl login --profile home        # Save home HA config
hactl login --profile office      # Save office HA config
hactl entity list --profile home  # Use specific profile
```

## For LLM Agents

| Feature | How |
|---|---|
| **Auto-JSON for agents** | Non-TTY stdout → JSON automatically |
| **`--fields`** | Token-efficient output |
| **`schema` command** | Runtime introspection |
| **`skills` command** | Agent-optimized usage reference |
| **`--json-input`** | Send exact API payloads |
| **`--dry-run`** | Preview before executing |
| **Structured errors** | JSON on stderr with `code`, `message`, `guidance` |
| **`--yes`** | Non-interactive destructive actions |
| **Deterministic output** | Same input → same structure |

## API Coverage

hactl covers **every** Home Assistant REST API endpoint:

| Endpoint | Command |
|---|---|
| `GET /api/` | `hactl status` |
| `GET /api/config` | `hactl ha-config` |
| `GET /api/core/state` | `hactl core-state` |
| `GET /api/stream` | `hactl stream` |
| `GET /api/components` | `hactl component list` |
| `GET /api/events` | `hactl event list` |
| `GET /api/services` | `hactl service list` |
| `GET /api/history/period` | `hactl entity history` |
| `GET /api/logbook` | `hactl logbook` |
| `GET /api/states` | `hactl entity list` |
| `GET /api/states/<id>` | `hactl entity get` |
| `GET /api/error_log` | `hactl error-log` |
| `GET /api/camera_proxy/<id>` | `hactl camera snapshot` |
| `GET /api/calendars` | `hactl calendar list` |
| `GET /api/calendars/<id>` | `hactl calendar events` |
| `POST /api/states/<id>` | `hactl entity update` |
| `POST /api/events/<type>` | `hactl event fire` |
| `POST /api/services/<d>/<s>` | `hactl service call` |
| `POST /api/template` | `hactl template render` |
| `POST /api/config/core/check_config` | `hactl config check` |
| `POST /api/intent/handle` | `hactl intent handle` |
| `DELETE /api/states/<id>` | `hactl entity delete` |

## License

MIT

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(schemaCmd)
}

// SchemaEntry describes a single command for agent introspection.
type SchemaEntry struct {
	Resource    string        `json:"resource"`
	Action      string        `json:"action"`
	Description string        `json:"description"`
	Method      string        `json:"httpMethod"`
	Path        string        `json:"apiPath"`
	Parameters  []SchemaParam `json:"parameters,omitempty"`
	Flags       []SchemaFlag  `json:"flags,omitempty"`
	Example     string        `json:"example"`
	Mutating    bool          `json:"mutating"`
	DryRun      bool          `json:"supportsDryRun"`
}

// SchemaParam describes a path/query parameter.
type SchemaParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	In       string `json:"in"` // path, query, config
}

// SchemaFlag describes a CLI flag.
type SchemaFlag struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Desc     string `json:"description"`
}

var schemaCmd = &cobra.Command{
	Use:   "schema [resource.action]",
	Short: "Runtime command schema for LLM agents",
	Long: `Returns a JSON schema for any command, including parameters, types,
required fields, and copy-pasteable examples.

This is the primary entry point for LLM agents discovering how to use hactl.

Examples:
  hactl schema                       List all available commands
  hactl schema entity.list           Schema for entity list
  hactl schema service.call          Schema for service call`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registry := buildSchemaRegistry()

		if len(args) == 0 {
			// List all commands
			summary := make([]map[string]string, 0, len(registry))
			for key, entry := range registry {
				summary = append(summary, map[string]string{
					"command":     key,
					"description": entry.Description,
					"mutating":    fmt.Sprintf("%v", entry.Mutating),
				})
			}
			return printAPIResult(toAnySlice(mapSlice(summary)))
		}

		key := args[0]
		entry, ok := registry[key]
		if !ok {
			// Try to suggest
			suggestions := findSimilar(key, registry)
			msg := fmt.Sprintf("unknown command: %s", key)
			if len(suggestions) > 0 {
				msg += fmt.Sprintf("\n\nDid you mean one of these?\n  %s", strings.Join(suggestions, "\n  "))
			}
			return fmt.Errorf("%s", msg)
		}

		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(entry)
	},
}

// buildSchemaRegistry returns the map of all command schemas.
func buildSchemaRegistry() map[string]SchemaEntry {
	return map[string]SchemaEntry{
		// Status
		"status": {
			Resource: "status", Action: "check", Description: "Check Home Assistant API status",
			Method: "GET", Path: "/api/", Example: "hactl status",
		},

		// Entity commands
		"entity.list": {
			Resource: "entity", Action: "list", Description: "List all entities and their states",
			Method: "GET", Path: "/api/states",
			Example: "hactl entity list --fields entity_id,state",
		},
		"entity.get": {
			Resource: "entity", Action: "get", Description: "Get entity state and attributes",
			Method: "GET", Path: "/api/states/{entity_id}",
			Example:    "hactl entity get sensor.temperature",
			Parameters: []SchemaParam{{Name: "entity_id", Type: "string", Required: true, In: "path"}},
		},
		"entity.update": {
			Resource: "entity", Action: "update", Description: "Update or create an entity state",
			Method: "POST", Path: "/api/states/{entity_id}", Mutating: true, DryRun: true,
			Example:    "hactl entity update sensor.temp --state 25",
			Parameters: []SchemaParam{{Name: "entity_id", Type: "string", Required: true, In: "path"}},
			Flags: []SchemaFlag{
				{Name: "state", Type: "string", Desc: "Entity state value"},
				{Name: "json-input", Type: "string", Desc: "Full JSON body with state and attributes"},
			},
		},
		"entity.delete": {
			Resource: "entity", Action: "delete", Description: "Delete an entity",
			Method: "DELETE", Path: "/api/states/{entity_id}", Mutating: true, DryRun: true,
			Example:    "hactl entity delete sensor.temp --yes",
			Parameters: []SchemaParam{{Name: "entity_id", Type: "string", Required: true, In: "path"}},
		},
		"entity.history": {
			Resource: "entity", Action: "history", Description: "Get state history for entities",
			Method: "GET", Path: "/api/history/period/{timestamp}",
			Example:    "hactl entity history sensor.temperature --start 2024-01-01T00:00:00Z",
			Parameters: []SchemaParam{{Name: "entity_id", Type: "string", Required: true, In: "path"}},
			Flags: []SchemaFlag{
				{Name: "start", Type: "string", Desc: "Start timestamp (ISO 8601)"},
				{Name: "end", Type: "string", Desc: "End timestamp (ISO 8601)"},
				{Name: "minimal", Type: "bool", Desc: "Return minimal response (faster)"},
				{Name: "no-attributes", Type: "bool", Desc: "Skip returning attributes (faster)"},
				{Name: "significant-only", Type: "bool", Desc: "Only return significant state changes"},
			},
		},

		// Service commands
		"service.list": {
			Resource: "service", Action: "list", Description: "List all available services by domain",
			Method: "GET", Path: "/api/services",
			Example: "hactl service list --fields domain,services",
		},
		"service.call": {
			Resource: "service", Action: "call", Description: "Call a Home Assistant service",
			Method: "POST", Path: "/api/services/{domain}/{service}", Mutating: true, DryRun: true,
			Example: "hactl service call light turn_on --entity-id light.living_room",
			Parameters: []SchemaParam{
				{Name: "domain", Type: "string", Required: true, In: "path"},
				{Name: "service", Type: "string", Required: true, In: "path"},
			},
			Flags: []SchemaFlag{
				{Name: "entity-id", Type: "string", Desc: "Target entity ID"},
				{Name: "json-input", Type: "string", Desc: "JSON service data"},
				{Name: "return-response", Type: "bool", Desc: "Return service response data"},
			},
		},

		// Event commands
		"event.list": {
			Resource: "event", Action: "list", Description: "List all event types and listener counts",
			Method: "GET", Path: "/api/events",
			Example: "hactl event list --fields event,listener_count",
		},
		"event.fire": {
			Resource: "event", Action: "fire", Description: "Fire a custom event",
			Method: "POST", Path: "/api/events/{event_type}", Mutating: true, DryRun: true,
			Example:    "hactl event fire my_custom_event --json-input '{\"key\":\"value\"}'",
			Parameters: []SchemaParam{{Name: "event_type", Type: "string", Required: true, In: "path"}},
			Flags:      []SchemaFlag{{Name: "json-input", Type: "string", Desc: "JSON event data"}},
		},

		// Logbook
		"logbook": {
			Resource: "logbook", Action: "list", Description: "View logbook entries",
			Method: "GET", Path: "/api/logbook/{timestamp}",
			Example: "hactl logbook --start 2024-01-01T00:00:00Z --entity sensor.temperature",
			Flags: []SchemaFlag{
				{Name: "start", Type: "string", Desc: "Start timestamp (ISO 8601)"},
				{Name: "end", Type: "string", Desc: "End timestamp (ISO 8601)"},
				{Name: "entity", Type: "string", Desc: "Filter by entity ID"},
			},
		},

		// Template
		"template.render": {
			Resource: "template", Action: "render", Description: "Render a Jinja2 template",
			Method: "POST", Path: "/api/template", Mutating: false,
			Example:    "hactl template render 'It is {{ now() }}!'",
			Parameters: []SchemaParam{{Name: "template", Type: "string", Required: true, In: "path"}},
			Flags: []SchemaFlag{
				{Name: "file", Type: "string", Desc: "Read template from file"},
				{Name: "variables", Type: "string", Desc: "JSON object of template variables"},
			},
		},

		// Error log
		"error-log": {
			Resource: "error-log", Action: "view", Description: "View Home Assistant error log",
			Method: "GET", Path: "/api/error_log",
			Example: "hactl error-log",
		},

		// Camera
		"camera.snapshot": {
			Resource: "camera", Action: "snapshot", Description: "Download a camera snapshot",
			Method: "GET", Path: "/api/camera_proxy/{entity_id}",
			Example:    "hactl camera snapshot camera.front_door -O snapshot.jpg",
			Parameters: []SchemaParam{{Name: "entity_id", Type: "string", Required: true, In: "path"}},
			Flags:      []SchemaFlag{{Name: "output-file", Type: "string", Desc: "Output file path"}},
		},

		// Calendar
		"calendar.list": {
			Resource: "calendar", Action: "list", Description: "List all calendar entities",
			Method: "GET", Path: "/api/calendars",
			Example: "hactl calendar list --fields entity_id,name",
		},
		"calendar.events": {
			Resource: "calendar", Action: "events", Description: "List events for a calendar",
			Method: "GET", Path: "/api/calendars/{entity_id}",
			Example:    "hactl calendar events calendar.holidays --start 2024-05-01T00:00:00Z --end 2024-06-01T00:00:00Z",
			Parameters: []SchemaParam{{Name: "entity_id", Type: "string", Required: true, In: "path"}},
			Flags: []SchemaFlag{
				{Name: "start", Type: "string", Required: true, Desc: "Start timestamp (ISO 8601)"},
				{Name: "end", Type: "string", Required: true, Desc: "End timestamp (ISO 8601)"},
			},
		},

		// Component
		"component.list": {
			Resource: "component", Action: "list", Description: "List loaded components/integrations",
			Method: "GET", Path: "/api/components",
			Example: "hactl component list",
		},

		// Intent
		"intent.handle": {
			Resource: "intent", Action: "handle", Description: "Handle an intent",
			Method: "POST", Path: "/api/intent/handle", Mutating: true, DryRun: true,
			Example: "hactl intent handle --name SetTimer --json-input '{\"data\":{\"seconds\":\"30\"}}'",
			Flags: []SchemaFlag{
				{Name: "name", Type: "string", Desc: "Intent name"},
				{Name: "json-input", Type: "string", Desc: "Full JSON intent payload"},
			},
		},

		// Config check
		"config.check": {
			Resource: "config", Action: "check", Description: "Validate Home Assistant configuration.yaml",
			Method: "POST", Path: "/api/config/core/check_config",
			Example: "hactl config check",
		},

		// HA server config
		"ha-config": {
			Resource: "ha-config", Action: "show", Description: "Show Home Assistant server configuration",
			Method: "GET", Path: "/api/config",
			Example: "hactl ha-config --fields version,location_name,time_zone",
		},

		// Core state
		"core-state": {
			Resource: "core-state", Action: "check", Description: "Check Home Assistant core running state",
			Method: "GET", Path: "/api/core/state",
			Example: "hactl core-state",
		},

		// Event stream
		"stream": {
			Resource: "stream", Action: "connect", Description: "Stream real-time events via SSE",
			Method: "GET", Path: "/api/stream",
			Example: "hactl stream --restrict state_changed",
			Flags: []SchemaFlag{
				{Name: "restrict", Type: "string", Desc: "Comma-separated event types to filter"},
			},
		},
	}
}

func findSimilar(input string, registry map[string]SchemaEntry) []string {
	var suggestions []string
	inputLower := strings.ToLower(input)
	for key := range registry {
		if strings.Contains(key, inputLower) || strings.Contains(inputLower, strings.Split(key, ".")[0]) {
			suggestions = append(suggestions, key)
		}
	}
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}
	return suggestions
}

func mapSlice(maps []map[string]string) []map[string]any {
	result := make([]map[string]any, len(maps))
	for i, m := range maps {
		r := make(map[string]any, len(m))
		for k, v := range m {
			r[k] = v
		}
		result[i] = r
	}
	return result
}

// toAnySlice converts []map[string]any to []any for the printer.
func toAnySlice(maps []map[string]any) []any {
	result := make([]any, len(maps))
	for i, m := range maps {
		result[i] = m
	}
	return result
}

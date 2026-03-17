# Copilot Instructions for hactl

## Project Overview

**hactl** is a fully featured CLI for Home Assistant that covers every REST API endpoint. Built for both humans and LLM agents using the cli-skeleton pattern.

## CLI Name & Command Grammar

The binary/command name is `hactl`. Commands follow a **resource action** (noun-verb) pattern:

```
hactl <resource> <action> [flags]
```

Resources: entity, service, event, calendar, camera, component, intent
Direct commands: status, logbook, template, error-log

For complex input, use `--json-input` with a full JSON payload.

## Discoverability

- Running `hactl` with no arguments prints a concise overview.
- `--help` on every command shows examples first, then flags.
- Typos trigger "did you mean …?" suggestions.
- **`hactl schema <resource>.<action>`** — Runtime introspection for LLM agents.
- **`hactl skills`** — Agent-optimized usage instructions.

## Output & Formatting

- **Default (TTY)**: Colored, aligned tables.
- **Non-TTY (piped/agent)**: Automatically switch to JSON output.
- **`--json`**: Force JSON. **`--csv`**: CSV. **`--output ndjson`**: NDJSON.
- **`--fields`**: Field mask to select specific fields.
- **stdout** is for data only. Logs, progress, errors go to **stderr**.

## Safety

- Every mutating command supports `--dry-run`.
- Destructive actions prompt for confirmation unless `--yes` is passed.
- Errors include `code`, `message`, and `guidance` fields on stderr.

## Configuration

Precedence: **CLI flags → env vars → config file**.

- Config at `~/.config/hactl/config.yaml` (XDG-compliant).
- Secrets stored in OS keyring.
- Key env vars: `HACTL_TOKEN`, `HACTL_HOST`, `HACTL_OUTPUT_FORMAT`, `HACTL_DEBUG`.

## Architecture

- **Three-layer separation**: `cmd/` (CLI) → `internal/api/` (HTTP client) → `internal/output/` (formatting).
- Home Assistant API base URL configured via `--host` flag or `HACTL_HOST` env var.
- Bearer token auth via keyring or `HACTL_TOKEN`.
- Tests for: API client, output formatting, argument parsing, schema registry.

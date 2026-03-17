# CLAUDE.md — hactl

Quick reference for Claude Code and other AI agents working on this codebase.

## What is hactl?

`hactl` is a CLI for Home Assistant that covers every REST API endpoint.
Built for both humans and LLM agents using the cli-skeleton pattern.

## Build & Test

```bash
make all        # lint + test + build
make build      # just build
make test       # just tests
make lint       # just lint
```

## Architecture

Three-layer separation: `cmd/` (CLI parsing) → `internal/api/` (HTTP client) → `internal/output/` (formatting).

- `cmd/root.go` — Root command, global flags, output format auto-detection
- `cmd/entity.go` — Entity CRUD + history (list, get, update, delete, history)
- `cmd/service.go` — Service list + call
- `cmd/event.go` — Event list + fire
- `cmd/logbook.go` — Logbook viewing
- `cmd/template.go` — Jinja2 template rendering
- `cmd/errorlog.go` — Error log viewing
- `cmd/camera.go` — Camera snapshot download
- `cmd/calendar.go` — Calendar list + events
- `cmd/component.go` — Component listing
- `cmd/intent.go` — Intent handling
- `cmd/status.go` — API status check
- `cmd/config.go` — Local config + HA config check
- `cmd/login.go` — Interactive auth with keyring
- `cmd/schema.go` — Runtime introspection for agents
- `cmd/skills.go` — Agent usage reference
- `cmd/helpers.go` — Shared utilities (newAPIClient, confirmAction, etc.)
- `internal/api/client.go` — HTTP client with retries, Bearer auth
- `internal/config/` — XDG config, keyring, profiles
- `internal/output/` — Multi-format printer (table/JSON/CSV/NDJSON)

## Key Patterns

- stdout = data only; stderr = logs/errors/progress
- Non-TTY stdout auto-outputs JSON
- `--fields` works with all output formats
- `--dry-run` on all mutating commands
- `--yes` required for destructive commands in non-interactive mode
- Structured errors on stderr with `code`, `message`, `guidance`

## See Also

- [AGENTS.md](./AGENTS.md) — Full agent specification
- [SKILLS.md](./SKILLS.md) — Agent skills reference

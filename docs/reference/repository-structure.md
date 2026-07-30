# Repository Structure

Complete guide to the current meta-cc directory organization.

## Directory Tree

```text
meta-cc/
├── .claude-plugin/             # Claude Code marketplace metadata
│   └── marketplace.json
├── cmd/
│   └── mcp-server/             # MCP server entry point
├── internal/
│   ├── conversation/           # Provider-agnostic session, turn, and tool-call model
│   ├── locator/                # Claude project and Codex home path resolution
│   ├── mcp/                    # MCP executor, handlers, tools, pipeline, response code
│   ├── parser/                 # Claude Code JSONL parser
│   ├── provider/
│   │   ├── claude/             # Claude Code provider adapter
│   │   ├── codex/              # Codex SQLite and rollout JSONL provider
│   │   └── records/            # Shared record normalization helpers
│   ├── query/                  # Query engines, resources, jq/stage2 support
│   └── ...
├── lib/
│   ├── mcp-config.json         # PATH-based MCP config template
│   └── meta-utils.sh           # Shared prompt command utilities
├── plugin-src/
│   ├── .claude-plugin/         # Claude Code plugin manifest
│   ├── .codex-plugin/          # Codex plugin manifest
│   ├── .codex-mcp.json         # Codex MCP template
│   ├── .mcp.json               # Claude Code plugin MCP template
│   ├── commands/               # Claude Code prompt-library slash commands
│   └── skills/                 # Codex prompt-library skills
├── scripts/
│   ├── install/                # Archive and skills installers
│   ├── ci/                     # Smoke and release checks
│   └── checks/                 # Local quality checks
├── tests/
│   ├── e2e/                    # MCP and Codex E2E scripts
│   └── fixtures/
│       └── codex/              # Codex rollout fixtures
├── docs/                       # Guides, tutorials, examples, and reference docs
├── Makefile
├── go.mod
├── README.md
└── CLAUDE.md
```

## Host Integration Files

### Claude Code

Source files:

- `plugin-src/.claude-plugin/plugin.json`
- `plugin-src/.mcp.json`
- `plugin-src/commands/prompt-find.md`
- `plugin-src/commands/prompt-list.md`
- `plugin-src/commands/prompt-show.md`

Installed locations:

- `~/.claude/commands/`
- `~/.claude/mcp.json`
- `~/.local/share/meta-cc/` for user-scope plugin installs

### Codex

Source files:

- `plugin-src/.codex-plugin/plugin.json`
- `plugin-src/.codex-mcp.json`
- `plugin-src/skills/prompt-find/SKILL.md`
- `plugin-src/skills/prompt-list/SKILL.md`
- `plugin-src/skills/prompt-show/SKILL.md`

Installed locations:

- `~/.codex/skills/`
- `~/.codex/plugins/meta-cc/.codex-plugin/plugin.json`
- `~/.codex/plugins/meta-cc/.codex-mcp.json`

## Core Packages

### `internal/conversation`

Defines the provider-neutral model used by Claude Code and Codex:

- sessions
- turns
- user and assistant messages
- tool calls and tool outputs
- token usage

### `internal/provider`

Contains the multi-provider adapter layer:

- `provider/claude`: reads Claude Code project JSONL transcripts
- `provider/codex`: reads the highest-compatible Codex `state_N.sqlite` and rollout JSONL files (rollout-only fallback when no compatible database exists)
- `provider/registry.go`: fan-out and provider filtering
- `provider/records`: shared normalized record helpers

### `internal/locator`

Resolves host-specific paths:

- Claude Code project roots from `~/.claude/projects/` or `META_CC_PROJECTS_ROOT`
- Codex roots from `~/.codex`, `CODEX_HOME`, or `META_CC_CODEX_ROOT` (precedence: `META_CC_CODEX_ROOT` → `CODEX_HOME` → `~/.codex`)

### `internal/mcp`

Implements the MCP surface:

- tool definitions
- provider-aware executor
- convenience query handlers
- analysis handlers
- hybrid inline/file-reference output

## Build Artifacts

Generated files are not source of truth:

- `bin/`
- `build/`
- `dist/`
- `plugin-src/bin/`

`make stage`, `make install-local`, and `make install-user` may regenerate `plugin-src/bin/meta-cc-mcp`.

## Tests

Important test entry points:

```bash
go test ./...
make test-e2e-mcp
make test-e2e-codex
```

`tests/e2e/codex-e2e.sh` is the real Codex integration test. It creates an isolated Codex home, installs Codex plugin and skill files, builds SQLite/rollout fixtures, and verifies `provider: "codex"` through the MCP server.

## Documentation

Primary user docs:

- [README](../../README.md)
- [Installation Guide](../tutorials/installation.md)
- [Examples](../tutorials/examples.md)
- [Integration Guide](../guides/integration.md)
- [MCP Guide](../guides/mcp.md)
- [MCP Query Tools Reference](../guides/mcp-query-tools.md)

Planning and historical docs under `docs/plans/`, `docs/proposals/`, `plans/`, and `docs/archive/` may describe older architecture phases.

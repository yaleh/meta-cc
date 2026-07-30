# Meta-CC Claude Code and Codex Integration Guide

This guide explains which meta-cc integration to use in Claude Code or Codex.

## Integration Surface

| Integration | Claude Code | Codex | Best for |
|-------------|-------------|-------|----------|
| MCP server | Yes | Yes | Session history queries and analysis |
| Prompt-library commands | `/prompt-find`, `/prompt-list`, `/prompt-show` | Native skills: `prompt-find`, `prompt-list`, `prompt-show` | Reusing saved prompts |
| Plugin metadata | Claude Code marketplace/archive plugin | Codex plugin metadata under `~/.codex/plugins/meta-cc/` | Host-native packaging |

The MCP server is the primary integration. It exposes 16 tools for querying and analyzing Claude Code and Codex session history through a provider-aware layer.

## Data Sources

| Provider | Source | Notes |
|----------|--------|-------|
| `claude` | `~/.claude/projects/<project-hash>/*.jsonl` | Host default under Claude Code and for standalone installs. |
| `codex` | Highest-compatible `state_N.sqlite` under the canonical Codex root (`META_CC_CODEX_ROOT` → `CODEX_HOME` → `~/.codex`) plus rollout JSONL files referenced by `threads.rollout_path`; rollout-only fallback when no compatible database exists | Host default under Codex. `~/.codex/history.jsonl` is intentionally not used. See [Discovery roots](../reference/codex-history-model.md#discovery-roots-and-database-compatibility-dir-069). |
| `all` | Both providers | Results include a provider tag where applicable. |

An omitted `provider` argument resolves to the host that launched the MCP
server (the packaged manifests inject `META_CC_HOST=claude`/`codex`;
standalone defaults to `claude` — see the
[MCP guide](mcp.md#default-provider-for-manualstandalone-installs-meta_cc_host)).
Explicit `provider` values always override this default.

Use `working_dir` to query a project path different from the current process working directory.

## Installation Paths

### Claude Code Marketplace

Use this when you only need Claude Code integration:

```bash
/plugin marketplace add yaleh/meta-cc
/plugin install meta-cc
```

Restart Claude Code. The plugin provides the MCP server configuration and prompt-library slash commands.

### Codex Plugin Marketplace (preferred for Codex CLI 0.145+)

```bash
codex plugin marketplace add .   # from an extracted release archive, or a git checkout
codex plugin add meta-cc@meta-cc-marketplace
```

Verify with `codex plugin list --json` (installed+enabled) and `codex mcp list`
(exactly one `meta-cc` entry), then start a **new** Codex session — a
running session cannot hot-load a plugin installed after it started. See
[Installation Guide: Method 1b](../tutorials/installation.md#method-1b-codex-plugin-marketplace-preferred-for-codex-cli-0145) for
upgrade/uninstall/troubleshooting details, and the minimal
`codex mcp add meta-cc -- <path>` fallback for an MCP-only install.

### Archive Install (manual copy, Claude Code, or Codex < 0.145)

Use this for Claude Code, or for Codex when the plugin manager above isn't
available:

```bash
./install.sh
```

The archive installer:

- copies `meta-cc-mcp` to `~/.local/bin/`
- installs Claude Code slash commands under `~/.claude/commands/`
- merges Claude Code MCP configuration into `~/.claude/mcp.json`
- installs Codex skills under `~/.codex/skills/`
- installs Codex plugin metadata under `~/.codex/plugins/meta-cc/`

Install one host only:

```bash
INSTALL_CLAUDE=0 ./install.sh  # Codex files only
INSTALL_CODEX=0 ./install.sh   # Claude Code files only
```

Install prompt-library commands and skills without the MCP binary:

```bash
./install-skills.sh
```

Do not run both the Codex plugin-manager install and this manual Codex
install (or the minimal `codex mcp add` fallback) in the same `CODEX_HOME` —
each registers its own MCP server under the name `meta-cc`, and Codex will
show duplicate active registrations rather than merging them.

## MCP Usage

Ask naturally in either host:

```text
Which tools do I use most often?
Show my work patterns and peak hours
Find user messages mentioning "migration"
Analyze recent errors
Show token usage for recent assistant turns
```

For Codex-specific checks:

```text
Use provider=codex and show recent tool usage
Use provider=codex and find user messages mentioning "release"
Use provider=codex and show token usage
```

For cross-host analysis:

```text
Use provider=all and compare tool usage across Claude Code and Codex
Use provider=all and analyze recent error patterns
```

Common MCP calls:

```javascript
get_work_patterns({
  provider: "all",
  working_dir: "/path/to/project"
})
```

```javascript
query_session_signals({
  type: "tool_stats",
  provider: "codex",
  tool: "exec_command",
  limit: 20
})
```

```javascript
query_session_content({
  role: "user",
  provider: "claude",
  pattern: "test|refactor",
  limit: 20
})
```

See [MCP Query Tools Reference](mcp-query-tools.md) for the full catalog.

## Prompt Library

The prompt library lives in the current project's `.meta-cc/prompts/library/` directory.

Claude Code slash commands:

```text
/prompt-list
/prompt-list sort=usage
/prompt-find release checklist
/prompt-show phase-execution-001
```

Codex skills:

```text
$prompt-list
$prompt-list sort=usage
$prompt-find release checklist
$prompt-show phase-execution-001
```

The commands and skills parse the same Markdown files and frontmatter fields: `id`, `title`, `category`, `keywords`, `usage_count`, `updated`, and `status`.

## Choosing A Method

| Task | Use |
|------|-----|
| Ask about session history, tools, errors, or token usage | MCP tools |
| Browse saved prompts in Claude Code | `/prompt-list` |
| Browse saved prompts in Codex | `$prompt-list` |
| Search reusable prompts in Claude Code | `/prompt-find <keywords>` |
| Search reusable prompts in Codex | `$prompt-find <keywords>` |
| View a saved prompt in Claude Code | `/prompt-show <id>` |
| View a saved prompt in Codex | `$prompt-show <id>` |
| Validate Codex support in development | `make test-e2e-codex` |

## Verification

### Claude Code

1. Restart Claude Code.
2. Ask: `Which tools do I use most often?`
3. Run: `/prompt-list`

### Codex

1. Restart Codex.
2. Ask: `Use provider=codex and show my work patterns`
3. Run: `$prompt-list`

### Development E2E

```bash
make test-e2e-codex
```

This runs two isolated-`CODEX_HOME` test passes:

1. `tests/e2e/codex-e2e.sh` — installs the Codex files (skills/plugin manifest/archive layout) and calls the MCP server over JSON-RPC with `provider: "codex"`.
2. `tests/e2e/codex-plugin-manager-e2e.sh` — builds a local release bundle and drives the real `codex` CLI's `plugin marketplace add` / `plugin add` (preferred path) and `mcp add` (minimal fallback), asserting plugin enablement, skill discovery, single-registration MCP resolution, and a live query against the installed artifact. Requires Codex CLI 0.145+ on `PATH`; it SKIPs with an explicit reason (not a silent pass) if the installed CLI lacks the needed subcommands.

## Troubleshooting

### MCP Tools Are Not Called

- Confirm the host has loaded the MCP server configuration.
- Ask more directly: `Use the meta-cc MCP server to show recent tool usage`.
- Check the binary is available:

```bash
which meta-cc-mcp
```

### Claude Code Prompt Commands Not Found

```bash
ls ~/.claude/commands/prompt-list.md
ls ~/.claude/commands/prompt-find.md
ls ~/.claude/commands/prompt-show.md
```

Restart Claude Code after installing.

### Codex Skills Not Found

```bash
ls ~/.codex/skills/prompt-list/SKILL.md
ls ~/.codex/skills/prompt-find/SKILL.md
ls ~/.codex/skills/prompt-show/SKILL.md
```

Restart Codex after installing.

### Codex MCP Query Returns No Sessions

Check the Codex thread index (meta-cc picks the highest-compatible
`state_N.sqlite`; `state_5.sqlite` is the current Codex CLI schema):

```bash
sqlite3 ~/.codex/state_5.sqlite \
  "select cwd, rollout_path from threads order by updated_at desc limit 10;"
```

If no compatible database exists, meta-cc lists sessions directly from the
rollout trees (`sessions/` and `archived_sessions/`) — verify those contain
`*.jsonl` files whose `session_meta` records the project `cwd`.

If Codex data is stored outside `~/.codex`, set either override
(`META_CC_CODEX_ROOT` wins when both are set):

```bash
export META_CC_CODEX_ROOT=/path/to/codex-home   # meta-cc-only override
# or
export CODEX_HOME=/path/to/codex-home           # Codex CLI's own override
```

The `cwd` in the `threads` table (or in rollout `session_meta`) must match the project path you query with `working_dir`.

## Related Documentation

- [Installation Guide](../tutorials/installation.md)
- [Examples](../tutorials/examples.md)
- [MCP Guide](mcp.md)
- [MCP Query Tools Reference](mcp-query-tools.md)
- [JSONL Schema Reference](../reference/jsonl-schema.md)

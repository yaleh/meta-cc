# Feature Overview

meta-cc is a provider-aware MCP server for local coding-agent history analysis. It supports Claude Code and Codex through a shared session model, then exposes query and analysis tools to the host.

## Host Support

| Host | Data source | Integration files |
|------|-------------|-------------------|
| Claude Code | `~/.claude/projects/<project-hash>/*.jsonl` | `plugin-src/.claude-plugin/`, `plugin-src/.mcp.json`, `plugin-src/commands/` |
| Codex | `${META_CC_CODEX_ROOT:-~/.codex}/state_5.sqlite` plus rollout JSONL files referenced by `threads.rollout_path` | `plugin-src/.codex-plugin/`, `plugin-src/.codex-mcp.json`, `plugin-src/skills/` |

The `provider` parameter controls which history is queried:

- `claude`: Claude Code only. This is the default for backward compatibility.
- `codex`: Codex only.
- `all`: merge both providers and include provider-tagged records.

## MCP Tools

meta-cc exposes 15 MCP tools.

### Consolidated Query Tools

- `query_session_content`: query messages by role (`user`, `assistant`, `tool`, or `all`)
- `query_session_signals`: query signals (`errors`, `tokens`, `system_errors`, `timestamps`, `tool_stats`)
- `query_file_activity`: query Claude Code file history snapshots
- `query_edit_sequences`: analyze file edit/read patterns (docRole, co-accessed docs, DocVoid)

These 4 tools replace the older set of individually named `query_*` tools. Claude-only record types return empty results for Codex when Codex has no equivalent local record.

### Analysis Tools

- `analyze_errors`: aggregate tool errors by tool and type
- `analyze_bugs`: detect error-fix pairs and recurring patterns
- `quality_scan`: compute error, retry, diversity, and completion dimensions
- `get_work_patterns`: summarize tool frequency, hourly activity, and context switches
- `get_timeline`: build chronological session events
- `get_tech_debt`: detect TODO/FIXME/HACK markers and unresolved error signals

### Two-Stage Query Tools

- `get_session_directory`: locate a session directory and aggregate metadata
- `inspect_session_files`: inspect selected JSONL files
- `execute_stage2_query`: run jq-style filter/sort/transform on selected files
- `get_session_metadata`: return schema hints, file info, and query templates

### Utilities

- `cleanup_temp_files`: remove old temporary MCP output files

## Provider-Aware Normalization

Codex rollout records are normalized into the same conversation model used for Claude Code:

- user and assistant messages
- function/custom tool calls
- function/custom tool outputs
- token count events when present
- session metadata from SQLite

This lets the same MCP tools answer questions such as:

```text
Which tools do I use most often?
Show my work patterns and peak hours
Find user messages mentioning "refactor"
Analyze recent tool errors
Show token usage for recent assistant turns
```

## Prompt Library

meta-cc provides matching prompt-library workflows in both hosts.

Claude Code slash commands:

- `/prompt-find`
- `/prompt-list`
- `/prompt-show`

Codex skills:

- `$prompt-find`
- `$prompt-list`
- `$prompt-show`

Both read `.meta-cc/prompts/library/` in the current project and parse Markdown frontmatter fields such as `id`, `title`, `category`, `keywords`, `usage_count`, `updated`, and `status`.

## Output Modes

MCP responses use hybrid output:

- small results return inline
- large results are written to temporary JSONL files and returned as `file_ref`

This keeps natural host conversations usable while preserving complete result sets.

## Verification

General checks:

```text
Which tools do I use most often?
Find user messages mentioning "release"
Show token usage for recent assistant turns
```

Codex-specific check:

```text
Use provider=codex and show my work patterns
```

Development E2E:

```bash
make test-e2e-codex
```

The Codex E2E test creates an isolated Codex home, installs Codex skills and plugin metadata, creates SQLite and rollout fixtures, and verifies MCP calls with `provider: "codex"`.

## See Also

- [MCP Guide](../guides/mcp.md)
- [MCP Query Tools Reference](../guides/mcp-query-tools.md)
- [Integration Guide](../guides/integration.md)
- [Examples](../tutorials/examples.md)
- [JSONL Schema Reference](jsonl-schema.md)

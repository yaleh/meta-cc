# Meta-CC Examples

This guide shows practical ways to use meta-cc from Claude Code and Codex. The current integration surface is:

- MCP tools for session history analysis
- Claude Code slash commands for the prompt library
- Codex skills for the same prompt-library workflows

meta-cc reads Claude Code transcripts from `~/.claude/projects/` and Codex sessions from the highest-compatible `state_N.sqlite` under the canonical Codex root (`META_CC_CODEX_ROOT` → `CODEX_HOME` → `~/.codex`) plus rollout JSONL files, with a rollout-only fallback when no compatible database exists (see [Discovery roots](../reference/codex-history-model.md#discovery-roots-and-database-compatibility-dir-069)).

## Quick Checks

After installing, restart Claude Code or Codex and ask natural questions:

```text
Which tools do I use most often?
Show my work patterns and peak hours
Find user messages mentioning "refactor"
Show token usage for recent assistant turns
Analyze recent tool errors
```

When you want a specific host, mention it explicitly:

```text
Use provider=codex and show my recent tool usage
Use provider=claude and find recent Bash errors
Compare Claude Code and Codex activity with provider=all
```

## Provider Examples

Most convenience query and analysis tools accept `provider`:

| Provider | Meaning |
|----------|---------|
| `claude` | Query Claude Code sessions. Default when the MCP server runs under Claude Code or standalone. |
| `codex` | Query Codex local history. Default when the MCP server runs under Codex. |
| `all` | Query both providers and include provider-tagged records. |

An omitted `provider` resolves to the host that launched the MCP server
(`META_CC_HOST`; standalone installs fall back to `claude`). Explicit values
always override — keep using them for cross-host or deterministic queries.

Example MCP calls that a host can make:

```javascript
get_work_patterns({
  provider: "codex",
  working_dir: "/path/to/project"
})
```

```javascript
query_session_signals({
  type: "tool_stats",
  provider: "all",
  tool: "exec_command",
  working_dir: "/path/to/project",
  limit: 50
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

## Analysis Recipes

### Tool Usage

Ask:

```text
Which tools do I use most often?
```

Likely MCP tool:

```javascript
get_work_patterns({
  provider: "all",
  working_dir: "/path/to/project"
})
```

Use `provider: "codex"` when validating Codex support specifically.

### Peak Hours

Ask:

```text
Show my work patterns and peak hours
```

Likely MCP tool:

```javascript
get_work_patterns({
  provider: "all",
  working_dir: "/path/to/project"
})
```

The response includes `tool_frequency`, `hourly_activity`, `context_switches`, and `peak_hour`.

### Error Analysis

Ask:

```text
Analyze recent tool errors grouped by tool
```

Likely MCP tools:

```javascript
analyze_errors({
  provider: "all",
  working_dir: "/path/to/project",
  limit: 10
})
```

```javascript
query_session_signals({
  type: "errors",
  provider: "codex",
  stats_first: true,
  limit: 20
})
```

### Token Usage

Ask:

```text
Show token usage for recent assistant turns
```

Likely MCP tool:

```javascript
query_session_signals({
  type: "tokens",
  provider: "codex",
  stats_first: true,
  limit: 20
})
```

Codex token usage is available when rollout records include token-count events. SQLite `tokens_used` remains available as session metadata.

### Conversation Search

Ask:

```text
Find user messages mentioning "release" or "deploy"
```

Likely MCP tool:

```javascript
query_session_content({
  role: "user",
  provider: "all",
  pattern: "release|deploy",
  limit: 20
})
```

## Prompt Library

meta-cc provides the same prompt-library workflows in both hosts.

### Claude Code

Use slash commands:

```text
/prompt-list
/prompt-list sort=date
/prompt-list category=debug
/prompt-find release checklist
/prompt-show phase-execution-001
```

### Codex

Use the matching skills:

```text
$prompt-list
$prompt-list sort=date
$prompt-find release checklist
$prompt-show phase-execution-001
```

The commands and skills both read `.meta-cc/prompts/library/` from the current project.

## Large Result Handling

MCP responses use hybrid output:

- Small results return inline.
- Large results are written to a temporary JSONL file and returned as a `file_ref`.

Ask naturally:

```text
Analyze all tool usage patterns in this project
```

For large results, the host receives file metadata and can read or search the referenced file in chunks. Use `cleanup_temp_files` to remove old temporary MCP output files.

## Codex Verification

For development or release checks, run:

```bash
make test-e2e-codex
```

The E2E test creates a temporary Codex home containing:

- `state_5.sqlite`
- rollout JSONL fixtures
- installed Codex skills and plugin metadata

It then calls the MCP server over JSON-RPC with `provider: "codex"` and verifies Codex session, message, tool, token, and prompt-library behavior.

## Troubleshooting

### No Codex Results

Check that the queried project path matches the Codex thread `cwd` (meta-cc
uses the highest-compatible `state_N.sqlite`; `state_5.sqlite` is the current
Codex CLI schema):

```bash
sqlite3 ~/.codex/state_5.sqlite \
  "select cwd, rollout_path from threads order by updated_at desc limit 10;"
```

If you use a non-default Codex home (`META_CC_CODEX_ROOT` takes precedence
over `CODEX_HOME` when both are set):

```bash
META_CC_CODEX_ROOT=/path/to/codex-home meta-cc-mcp
```

### Prompt Skills Not Found In Codex

Verify files exist:

```bash
ls ~/.codex/skills/prompt-list/SKILL.md
ls ~/.codex/plugins/meta-cc/.codex-plugin/plugin.json
ls ~/.codex/plugins/meta-cc/.codex-mcp.json
```

Restart Codex after installing.

### Slash Commands Not Found In Claude Code

Verify files exist:

```bash
ls ~/.claude/commands/prompt-list.md
```

Restart Claude Code after installing.

## Related Docs

- [Installation Guide](installation.md)
- [Integration Guide](../guides/integration.md)
- [MCP Guide](../guides/mcp.md)
- [MCP Query Tools Reference](../guides/mcp-query-tools.md)

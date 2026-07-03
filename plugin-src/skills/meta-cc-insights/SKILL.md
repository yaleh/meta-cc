---
name: meta-cc-insights
description: When you repeat similar errors, need to reflect on recent work, want to understand your tool use patterns, or need to query session history, use the meta-cc insights tools for self-reflection and analysis.
---

# meta-cc Insights: Session Analysis & Reflection

This skill provides a quick reference for when and how to use meta-cc's 21 MCP tools for analyzing your Claude Code session history.

## When to Use: Trigger Scenarios

- You hit the same error repeatedly → use analyze_bugs
- You finish a feature and want to reflect → use quality_scan
- You're curious about your workflow patterns → use get_work_patterns
- You need to find something from earlier → use query_user_messages / query_tools
- You want to see the big picture timeline → use get_timeline

## Tool Capability Map

All tools support `provider: "claude"` (default) or `provider: "codex"` or `provider: "all"` to query across both.

### Error Diagnosis

Use these when debugging:
- `query_tool_errors`: list recent tool errors
- `analyze_errors`: aggregate errors by type
- `analyze_bugs`: detect error/fix patterns
- `query_system_errors`: system/API errors

### Quality Reflection

Use these after a phase or feature:
- `quality_scan`: error rate, retry rate, completion stats
- `get_tech_debt`: TODO/FIXME/HACK markers
- `query_summaries`: session summaries (search with `keyword`)

### Workflow & Timeline

Use these for pattern insight:
- `get_work_patterns`: hourly activity, tool frequency
- `get_timeline`: chronological event view
- `get_session_metadata`: schema info, templates

### History Search

Use these to find things:
- `query_user_messages`: search user messages (with `pattern`)
- `query_tools`: search tool calls (with `tool`, `status` filters)
- `query_conversation_flow`: full user/assistant pairs
- `query_token_usage`: token consumption stats
- `query_tool_blocks`: raw tool_use/tool_result blocks
- `query_file_snapshots`: file history
- `query_timestamps`: all entries with time bounds (`since`, `until`)

### Custom Queries (Two-Stage)

For full control:
1. `get_session_directory`: list available session files
2. `inspect_session_files`: inspect file metadata/samples
3. `execute_stage2_query`: run jq filter on selected files

### Utilities

- `cleanup_temp_files`: remove old MCP temp files (default: max_age_days=7)

## Complete Tool List

All 21 meta-cc tools:
- `query_tool_errors`, `query_token_usage`, `query_conversation_flow`, `query_system_errors`
- `query_file_snapshots`, `query_timestamps`, `query_summaries`, `query_tool_blocks`
- `query_tools`, `query_user_messages`, `cleanup_temp_files`
- `get_session_directory`, `inspect_session_files`, `execute_stage2_query`
- `analyze_errors`, `analyze_bugs`, `quality_scan`, `get_work_patterns`
- `get_session_metadata`, `get_timeline`, `get_tech_debt`

## Usage Pattern for Two-Stage Queries

When you need something specific that the shortcut tools don't cover:

```javascript
// Stage 1
const dir = await get_session_directory({ scope: "project" });
// Stage 2: pick files from dir.files, then inspect or query
await execute_stage2_query({
  files: dir.files.filter(f => f.includes("2026-06")),
  filter: 'select(.type == "assistant")',
  transform: '{timestamp, usage: .message.usage}',
});
```

Always prefer a convenience tool first; use two-stage only when needed.


# MCP Project-Level Query Guide

## Overview

Phase 12 extends the meta-cc MCP Server to support both **project-level** (all sessions) and **session-level** (current session) queries, enabling comprehensive cross-session analysis and focused single-session debugging.

## Tool Naming Convention

| Naming Pattern | Scope | Example |
|---------------|-------|---------|
| `<tool_name>` (no suffix) | **Project-level** (all sessions) | `query_tools` |
| `<tool_name>_session` | **Session-level** (current session) | `query_tools_session` |

**Exception**: `get_session_stats` retains its original name for backward compatibility.

## Available Tools

### Project-Level Tools (All Sessions)

Query across all sessions in the current project:

- `get_stats` - Project statistics
- `analyze_errors` - Error pattern analysis
- `query_tools` - Tool call history
- `query_user_messages` - User message search
- `query_tool_sequences` - Workflow patterns
- `query_file_access` - File operation history
- `query_successful_prompts` - Quality prompt patterns
- `query_context` - Error context analysis

### Session-Level Tools (Current Session)

Query only the current session:

- `get_session_stats` (backward compatible)
- `analyze_errors_session`
- `query_tools_session`
- `query_user_messages_session`
- `query_tool_sequences_session`
- `query_file_access_session`
- `query_successful_prompts_session`
- `query_context_session`

## Usage Examples

### Example 1: Project-Level Analysis

**User**: "How do I typically use agents in this project?"

**Claude** (uses `query_tools`):
```json
{
  "tool": "query_tools",
  "args": {
    "where": "tool LIKE '%agent%'",
    "limit": 50
  }
}
```

**Result**: Returns agent-related tool calls across all sessions in the project.

### Example 2: Session-Level Analysis

**User**: "What errors have occurred in this current session?"

**Claude** (uses `analyze_errors_session`):
```json
{
  "tool": "analyze_errors_session",
  "args": {
    "output_format": "json"
  }
}
```

**Result**: Returns errors from the current session only.

### Example 3: Cross-Session Error Patterns

**User**: "What are the most common errors in this project?"

**Claude** (uses `analyze_errors`):
```json
{
  "tool": "analyze_errors",
  "args": {
    "output_format": "json"
  }
}
```

**Result**: Returns aggregated error patterns from all sessions.

### Example 4: File Modification History

**User**: "Show me the edit history for main.go across all sessions"

**Claude** (uses `query_file_access`):
```json
{
  "tool": "query_file_access",
  "args": {
    "file": "main.go"
  }
}
```

**Result**: Returns read/edit/write operations on main.go from all sessions.

### Example 5: Tool Usage Comparison

**User**: "Compare tool usage in this session versus project average"

**Claude** (uses both):
```json
{
  "tool": "get_stats"
}
```
followed by:
```json
{
  "tool": "get_session_stats"
}
```

**Result**: Returns project-level and session-level statistics for comparison.

## Implementation Details

### Project-Level Execution

Project-level tools add the `--project .` flag to CLI commands:

```bash
# Project-level query
meta-cc query tools --project . --limit 100

# Result: All tool calls from all sessions in ~/.claude/projects/{project-hash}/
```

### Session-Level Execution

Session-level tools execute without the `--project` flag:

```bash
# Session-level query
meta-cc query tools --limit 100

# Result: Tool calls from current session only
```

## When to Use Each Scope

### Use Project-Level Tools When:

- Analyzing long-term patterns ("How do I typically structure prompts?")
- Identifying recurring errors ("What errors keep happening?")
- Tracking project evolution ("How has my tool usage changed?")
- Finding successful workflows ("What prompt patterns work best?")
- Comparing sessions ("Is this session typical for this project?")

### Use Session-Level Tools When:

- Debugging current session ("What went wrong just now?")
- Quick session summary ("How many tools have I used today?")
- Focused analysis ("Show me errors from this conversation")
- Performance tuning ("Is this session slower than usual?")
- Immediate context ("What did I ask about in this session?")

## Backward Compatibility

### Existing Tool Behavior

`get_session_stats` retains its original behavior:

```json
{
  "tool": "get_session_stats",
  "args": {}
}
```

Returns statistics for the current session only (no change from Phase 8).

### Migration Guide

If you were using `get_session_stats` for project-level analysis, migrate to:

```json
{
  "tool": "get_stats",
  "args": {}
}
```

## CLI Flag Reference

| Flag | Scope | Used By |
|------|-------|---------|
| `--project .` | All sessions in project | Project-level tools |
| (no flag) | Current session only | Session-level tools |

## Common Patterns

### Pattern 1: Error Investigation Workflow

```
1. User: "I keep seeing the same error"
2. Claude uses: analyze_errors (project-level)
   → Identifies error occurs in 5+ sessions
3. Claude uses: query_context
   → Gets context around error pattern
4. Claude suggests: Fix based on cross-session analysis
```

### Pattern 2: Workflow Discovery

```
1. User: "What's my typical workflow for adding features?"
2. Claude uses: query_tool_sequences (project-level)
   → Identifies common patterns: Read → Edit → Bash
3. Claude uses: query_successful_prompts
   → Finds prompts that led to successful workflows
4. Claude provides: Workflow recommendations
```

### Pattern 3: Session Comparison

```
1. User: "Am I working differently today?"
2. Claude uses: get_session_stats (current session)
3. Claude uses: get_stats (project average)
4. Claude compares: Tool usage, error rates, duration
5. Claude provides: Insights on session deviation
```

## Tool Reference Table

| Project-Level | Session-Level | Purpose |
|--------------|---------------|---------|
| `get_stats` | `get_session_stats` | Statistics |
| `analyze_errors` | `analyze_errors_session` | Error analysis |
| `query_tools` | `query_tools_session` | Tool calls |
| `query_user_messages` | `query_user_messages_session` | Message search |
| `query_tool_sequences` | `query_tool_sequences_session` | Workflows |
| `query_file_access` | `query_file_access_session` | File ops |
| `query_successful_prompts` | `query_successful_prompts_session` | Prompts |
| `query_context` | `query_context_session` | Error context |

## Troubleshooting

### Issue: Project-level tool returns only current session data

**Cause**: `--project .` flag not being passed to CLI

**Solution**: Verify MCP server implementation includes `--project .` in command execution

```bash
# Test manually
meta-cc query tools --project . --limit 10
# Should return data from multiple sessions (if available)
```

### Issue: Session-level tool returns multi-session data

**Cause**: Tool is using project-level execution by mistake

**Solution**: Verify tool name has `_session` suffix and does NOT include `--project` flag

```bash
# Correct session-level call
meta-cc query tools --limit 10
# Should return only current session data
```

### Issue: No data from other sessions

**Cause**: Project may only have one session, or sessions are in different project directories

**Solution**: Verify session files exist

```bash
# Check project directory
ls ~/.claude/projects/

# Find your project (look for matching path hash)
# Example: meta-cc project → hash like "home-user-meta-cc"

# List sessions in project
ls ~/.claude/projects/{project-hash}/
# Should see multiple .jsonl files if multiple sessions exist
```

## Advanced Use Cases

### Use Case 1: Learning from Past Mistakes

```
User: "I feel like I keep making the same mistakes. Help me learn."

Claude workflow:
1. analyze_errors → Find recurring error patterns
2. query_context → Get context around each pattern
3. query_successful_prompts → Find prompts that avoided these errors
4. Provide: Recommendations based on patterns
```

### Use Case 2: Workflow Optimization

```
User: "What's the most efficient way I work in this project?"

Claude workflow:
1. query_tool_sequences → Identify common workflows
2. get_stats → Get overall metrics
3. Compare: Fast workflows vs slow workflows
4. Provide: Optimized workflow suggestions
```

### Use Case 3: Onboarding New Project Members

```
User: "Explain how we typically work in this project"

Claude workflow:
1. get_stats → Overall project statistics
2. query_tool_sequences → Common workflow patterns
3. query_successful_prompts → Effective communication patterns
4. Provide: Project workflow guide
```

## Performance Considerations

### Large Projects (>10 sessions)

Project-level queries may return large datasets. Use:

- `--limit` parameter to restrict results
- `--fields` for field projection (Phase 9)
- `--summary-first` for overview (Phase 9)

```json
{
  "tool": "query_tools",
  "args": {
    "limit": 100,
    "output_format": "json"
  }
}
```

### Filtering Strategies

Use `where` parameter to filter data before returning:

```json
{
  "tool": "query_tools",
  "args": {
    "where": "status='error' AND tool='Bash'",
    "limit": 50
  }
}
```

## See Also

- [MCP Server Documentation](../README.md#mcp-server)
- [Integration Guide](./integration-guide.md)
- [MCP Usage Guide](./mcp-usage.md)
- [Phase 12 Implementation Plan](../plans/12/plan.md)
- [CLI Composability Guide](./cli-composability.md)

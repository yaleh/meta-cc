# MCP Server Guide

## Overview

meta-cc provides a Model Context Protocol (MCP) Server that enables Claude Code to autonomously query session data without manual CLI commands. The MCP server provides **16 powerful tools** with intelligent output control for comprehensive session analysis.

### Quick Start

**Prerequisites**:
- `meta-cc` binary in project root or PATH
- Claude Code with MCP support

**Configuration**: The MCP Server is configured in `.claude/mcp-servers/meta-cc.json`:

```json
{
  "command": "./meta-cc",
  "args": ["mcp"],
  "description": "Meta-cognition analysis for Claude Code sessions"
}
```

## Tool Catalog

meta-cc-mcp provides **16 standardized tools** for analyzing Claude Code session history.

### 1. get_session_stats

**Purpose**: Statistical information about the current session.

**Scope**: session (current session only)

**Key Parameters**: None

**Examples**:
```json
// Basic stats
{"stats_only": true}
→ {"TurnCount": 45, "ToolCallCount": 123, "ErrorRate": 0.04}

// With jq filter
{"jq_filter": "{turns: .TurnCount, errors: .ErrorCount}"}
```

---

### 2. query_tools

**Purpose**: Query tool calls with flexible filtering.

**Scope**: project (default) or session

**Key Parameters**:
- `tool` (string): Filter by tool name
- `status` (string): Filter by "error" or "success"
- `limit` (number): Maximum results (no limit by default)

**Examples**:
```json
// All Bash errors
{"tool": "Bash", "status": "error", "limit": 10}

// Error distribution by tool
{
  "status": "error",
  "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})",
  "stats_only": true
}
```

---

### 3. query_user_messages

**Purpose**: Search user messages with regex patterns.

**Scope**: project (default) or session

**Key Parameters**:
- `pattern` (string, **required**): Regex pattern
- `limit` (number): Maximum results
- `max_message_length` (number): Deprecated - use hybrid mode instead
- `content_summary` (boolean): Deprecated - use hybrid mode instead

**Examples**:
```json
// Find error-related messages
{"pattern": "error|fix|bug", "limit": 5}

// Count messages by topic
{"pattern": "test|testing", "jq_filter": "length", "stats_only": true}
```

**Pattern Examples**:
| Pattern | Description |
|---------|-------------|
| `Phase 8` | Exact match |
| `error\|bug` | OR operator |
| `^Continue` | Start with |
| `test$` | End with |
| `fix.*bug` | Between words |

---

### 4. query_assistant_messages

**Purpose**: Search assistant response messages with regex patterns.

**Scope**: project (default) or session

**Key Parameters**:
- `pattern` (string): Regex pattern
- `limit` (number): Maximum results
- `min_length` (number): Minimum text length
- `min_tokens_output` (number): Minimum output tokens

**Examples**:
```json
// Find test-related responses
{"pattern": "test.*passed", "limit": 5}

// Long responses only
{"min_length": 1000, "limit": 10}
```

---

### 5. query_conversation

**Purpose**: Search conversation messages (user + assistant pairs).

**Scope**: project (default) or session

**Key Parameters**:
- `pattern` (string): Regex pattern
- `pattern_target` (string): "user", "assistant", or "any" (default: "any")
- `start_turn` (number): Starting turn sequence
- `end_turn` (number): Ending turn sequence
- `min_duration` (number): Minimum response duration (ms)
- `max_duration` (number): Maximum response duration (ms)

**Examples**:
```json
// Find error discussions
{"pattern": "error", "pattern_target": "assistant"}

// Recent conversations
{"start_turn": 100, "limit": 10}
```

---

### 6. query_context

**Purpose**: Query context around errors (turns before and after).

**Scope**: project (default) or session

**Key Parameters**:
- `error_signature` (string, **required**): Error signature or pattern ID
- `window` (number): Context window size in turns (default: 3)

**Examples**:
```json
// Error context
{"error_signature": "Bash:command not found", "window": 3}

// Wide context window
{"error_signature": "permission denied", "window": 5}
```

---

### 7. query_tool_sequences

**Purpose**: Detect repeated workflow patterns.

**Scope**: project (default) or session

**Key Parameters**:
- `pattern` (string): Sequence pattern (e.g., "Read -> Edit -> Bash")
- `min_occurrences` (number): Minimum occurrences (default: 3)
- `include_builtin_tools` (boolean): Include built-in tools (default: false)

**Performance Note**: By default, built-in tools (Bash, Read, Edit) are excluded for:
- **35x faster execution** (~30s → <1s for large projects)
- **Cleaner workflow patterns** (focus on MCP tools)

**Examples**:
```json
// Common sequences
{"min_occurrences": 5}

// Specific pattern
{"pattern": "Read -> Edit", "min_occurrences": 1}

// Include built-in tools (slower)
{"min_occurrences": 5, "include_builtin_tools": true}
```

---

### 8. query_file_access

**Purpose**: Query operation history for a specific file.

**Scope**: project (default) or session

**Key Parameters**:
- `file` (string, **required**): File path to query

**Examples**:
```json
// File history
{"file": "cmd/mcp.go"}

// Count operations by type
{
  "file": "README.md",
  "jq_filter": "group_by(.Operation) | map({op: .[0].Operation, count: length})",
  "stats_only": true
}
```

---

### 9. query_project_state

**Purpose**: Query project state evolution across sessions.

**Scope**: project (default) or session

**Key Parameters**: None

**Examples**:
```json
// Project timeline
{"jq_filter": ".[] | {session: .SessionID, active_files: .ActiveFiles}"}

// Session count
{"jq_filter": "length", "stats_only": true}
```

---

### 10. query_successful_prompts

**Purpose**: Find historically successful prompt patterns.

**Scope**: project (default) or session

**Key Parameters**:
- `limit` (number): Maximum results
- `min_quality_score` (number): Minimum quality score 0-1 (default: 0.8)

**Examples**:
```json
// High-quality prompts
{"limit": 5, "min_quality_score": 0.9}

// Top 10 prompts
{"limit": 10, "min_quality_score": 0.7}
```

---

### 11. query_tools_advanced

**Purpose**: Advanced tool queries with SQL-like filter expressions.

**Scope**: project (default) or session

**Key Parameters**:
- `where` (string, **required**): SQL-like filter expression
- `limit` (number): Maximum results

**Supported Operators**: AND, OR, NOT, =, !=, >, <, >=, <=, IN, NOT IN, BETWEEN, LIKE, REGEXP

**Examples**:
```json
// Complex filter
{"where": "tool='Bash' AND status='error' AND duration>5000"}

// Multiple tools
{"where": "tool IN ('Read', 'Edit', 'Write')"}

// Time range
{"where": "timestamp BETWEEN '2025-10-01' AND '2025-10-05'"}
```

---

### 12. query_time_series

**Purpose**: Analyze metrics over time intervals.

**Scope**: project (default) or session

**Key Parameters**:
- `interval` (string): "hour", "day", or "week" (default: "hour")
- `metric` (string): "tool-calls" or "error-rate" (default: "tool-calls")
- `where` (string): Optional filter expression

**Examples**:
```json
// Daily tool calls
{"interval": "day", "metric": "tool-calls"}

// Error rate by hour
{"interval": "hour", "metric": "error-rate"}

// Bash-only timeline
{"interval": "day", "where": "tool='Bash'"}
```

---

### 13. query_files

**Purpose**: File-level operation statistics.

**Scope**: project (default) or session

**Key Parameters**:
- `sort_by` (string): "total_ops", "edit_count", "read_count", "write_count", "error_count", "error_rate" (default: "total_ops")
- `top` (number): Return top N files (default: 20)
- `threshold` (number): Minimum access count (default: 5)
- `where` (string): Optional filter expression

**Examples**:
```json
// Top edited files
{"sort_by": "edit_count", "top": 10}

// High error rate files
{"sort_by": "error_rate", "where": "error_count > 0"}

// Most active files
{"sort_by": "total_ops", "top": 5}
```

---

### 14. cleanup_temp_files

**Purpose**: Remove old temporary MCP files.

**Scope**: none (utility function)

**Key Parameters**:
- `max_age_days` (number): Max file age in days (default: 7)

**Example**:
```json
{"max_age_days": 7}
→ {"removed_count": 12, "freed_bytes": 5242880}
```

---

### 15. list_capabilities

**Purpose**: List all available capabilities from configured sources.

**Scope**: none (utility function)

**Key Parameters**: None

**Example**:
```json
{}
→ Returns compact capability index
```

---

### 16. get_capability

**Purpose**: Retrieve complete capability content by name.

**Scope**: none (utility function)

**Key Parameters**:
- `name` (string, **required**): Capability name (without .md extension)

**Example**:
```json
{"name": "meta-errors"}
→ Returns full capability content
```

---

## Standard Parameters

All query tools (1-13) support these parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `scope` | string | "project" | Query scope: "project" (cross-session) or "session" (current only) |
| `jq_filter` | string | ".[]" | jq expression for filtering and transforming results |
| `stats_only` | boolean | false | Return only statistics, no detailed records |
| `stats_first` | boolean | false | Return statistics first, then details (separated by `---`) |
| `inline_threshold_bytes` | number | 8192 | Threshold for inline vs file_ref mode (8KB default) |
| `output_format` | string | "jsonl" | Output format: "jsonl" or "tsv" |

---

## Output Control

### Hybrid Output Mode

The MCP server automatically selects output mode based on result size:

- **Inline Mode** (≤8KB): Data embedded directly in response
- **File Reference Mode** (>8KB): Data written to temp file, metadata returned

**Threshold Configuration**:

| Method | Priority | Example |
|--------|----------|---------|
| Parameter | Highest | `"inline_threshold_bytes": 16384` (16KB) |
| Environment | Medium | `export META_CC_INLINE_THRESHOLD=16384` |
| Default | Lowest | 8192 bytes (8KB) |

**Inline Mode Response**:
```json
{
  "mode": "inline",
  "data": [
    {"Timestamp": "2025-10-06T10:00:00Z", "ToolName": "Read"}
  ]
}
```

**File Reference Mode Response**:
```json
{
  "mode": "file_ref",
  "file_ref": {
    "path": "/tmp/meta-cc-mcp-abc123-1696598400-query_tools.jsonl",
    "size_bytes": 405000,
    "line_count": 5000,
    "fields": ["Timestamp", "ToolName", "Status"],
    "summary": {
      "record_count": 5000,
      "tool_distribution": {"Read": 1200, "Bash": 3000}
    }
  }
}
```

**Working with File References**:

1. **Analyze metadata first** - Check `file_ref.summary` for quick statistics
2. **Use Read tool** - Selectively examine file content
3. **Use Grep tool** - Search for patterns
4. **Present insights naturally** - Do NOT mention temp file paths to users

**Temporary File Management**:

- **Retention**: 7 days by default
- **Location**: `/tmp/` (Linux/macOS), `%TEMP%` (Windows)
- **Naming**: `/tmp/meta-cc-mcp-{hash}-{timestamp}-{query}.jsonl`
- **Cleanup**: Automatic after retention period, or use `cleanup_temp_files` tool

### Query Limit Strategy

By default, MCP tools **do not limit** the number of results:
- Small results automatically use inline mode (≤8KB)
- Large results automatically use file_ref mode (>8KB)

**When to explicitly use `limit` parameter**:

1. User explicitly requests a specific number (e.g., "last 10 errors")
2. Sample data only (e.g., "give me a few examples")
3. Quick exploration (view subset first, then expand)

**Example**:
```
User: "List all errors in this project"
→ query_tools(status="error")  # No limit, uses file_ref mode

User: "Show me the last 5 errors"
→ query_tools(status="error", limit=5)  # Explicit limit, likely inline mode
```

### jq Filter Cookbook

**Basic Filtering**:
```jq
# Select errors only
.[] | select(.Status == "error")

# Project specific fields
.[] | {tool: .ToolName, status: .Status}
```

**Aggregation**:
```jq
# Group by tool and count
group_by(.ToolName) | map({tool: .[0].ToolName, count: length})

# Calculate error rate by tool
group_by(.ToolName) | map({
  tool: .[0].ToolName,
  total: length,
  errors: map(select(.Status == "error")) | length
})

# Top N sorted
group_by(.ToolName) | map({tool: .[0].ToolName, count: length}) | sort_by(.count) | reverse | .[0:10]
```

**Array Operations**:
```jq
# Last N items
.[-10:]

# First N items
.[0:10]

# Length
length
```

**Time-Based Filtering**:
```jq
# After specific date
.[] | select(.Timestamp > "2025-10-01")

# Date range
.[] | select(.Timestamp >= "2025-10-01" and .Timestamp <= "2025-10-05")
```

**String Operations**:
```jq
# Case-insensitive match
.[] | select(.Error | ascii_downcase | contains("permission"))

# Regex test
.[] | select(.Error | test("error|failed|timeout"; "i"))

# Extract substring
.[] | {tool: .ToolName, error_preview: .Error[0:50]}
```

---

## Query Scope

### Project vs Session Scope

| Scope | Description | Use Cases |
|-------|-------------|-----------|
| `project` (default) | Cross-session analysis | Long-term patterns, recurring errors, project evolution |
| `session` | Current session only | Debugging current session, quick summaries, immediate context |

**When to Use Project Scope**:
- Analyzing long-term patterns ("How do I typically structure prompts?")
- Identifying recurring errors ("What errors keep happening?")
- Tracking project evolution ("How has my tool usage changed?")
- Finding successful workflows ("What prompt patterns work best?")

**When to Use Session Scope**:
- Debugging current session ("What went wrong just now?")
- Quick session summary ("How many tools have I used today?")
- Focused analysis ("Show me errors from this conversation")
- Immediate context ("What did I ask about in this session?")

**Example Comparison**:
```json
// Project-level: All sessions
{"scope": "project", "status": "error"}
→ Returns errors across all sessions

// Session-level: Current only
{"scope": "session", "status": "error"}
→ Returns errors from current session
```

---

## Best Practices

### 1. Natural Language Queries

Let Claude choose the right tool based on context:

```
User: "Why do my Bash commands keep failing?"

Claude: [Automatically calls]
  1. query_tools(tool="Bash", status="error")
  2. analyze_errors()
  3. Provides analysis and recommendations
```

### 2. Use stats_only to Reduce Token Usage

**Good** - Stats only:
```json
{
  "status": "error",
  "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})",
  "stats_only": true
}
```

**Bad** - Return all details:
```json
{"status": "error"}  // May return huge datasets
```

### 3. Leverage Hybrid Output Mode

- **Quick queries**: Inline mode is automatic (≤8KB)
- **Large queries**: File_ref mode handles size (>8KB)
- **Custom threshold**: Adjust via `inline_threshold_bytes`

### 4. Combine Tools for Complex Analysis

**Error Investigation Workflow**:
```
1. "Analyze my errors" → analyze_errors()
2. "Show Bash errors" → query_tools(tool="Bash", status="error")
3. "What happened before error X?" → query_context(error_signature="X")
```

### 5. Use scope Parameter Wisely

**Project scope (default)** - For meta-cognition:
```json
{"scope": "project", "status": "error"}
```

**Session scope** - For current analysis:
```json
{"scope": "session", "status": "error"}
```

### 6. Use TSV for Large Datasets

**JSONL** (default):
```json
{"limit": 100, "output_format": "jsonl"}
```

**TSV** (86% smaller):
```json
{"limit": 100, "output_format": "tsv"}
```

---

## Common Patterns

### Pattern 1: Error Distribution Analysis

```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length}) | sort_by(.count) | reverse",
    "stats_only": true
  }
}
```

### Pattern 2: File Hotspot Detection

```json
{
  "name": "query_files",
  "arguments": {
    "sort_by": "total_ops",
    "top": 10,
    "jq_filter": ".[] | {file: .FilePath, ops: .TotalOps, error_rate: .ErrorRate}"
  }
}
```

### Pattern 3: Workflow Pattern Discovery

```json
{
  "name": "query_tool_sequences",
  "arguments": {
    "min_occurrences": 5,
    "jq_filter": "sort_by(.Occurrences) | reverse | .[0:10]"
  }
}
```

### Pattern 4: Time-Based Error Analysis

```json
{
  "name": "query_time_series",
  "arguments": {
    "interval": "hour",
    "metric": "error-rate",
    "jq_filter": ".[] | select(.ErrorRate > 0) | {hour: .Timestamp, rate: .ErrorRate}"
  }
}
```

### Pattern 5: Successful Prompt Mining

```json
{
  "name": "query_successful_prompts",
  "arguments": {
    "limit": 10,
    "min_quality_score": 0.9,
    "jq_filter": ".[] | {prompt: .Content[0:200], score: .QualityScore, turn: .TurnSequence}"
  }
}
```

---

## Troubleshooting

### MCP Server Not Connected

**Symptom**: Claude can't find MCP tools

**Solution**:
1. Check `meta-cc` binary exists:
   ```bash
   ./meta-cc --version
   ```
2. Verify configuration file:
   ```bash
   cat .claude/mcp-servers/meta-cc.json
   jq empty .claude/mcp-servers/meta-cc.json && echo "Valid JSON"
   ```
3. Restart Claude Code

### Tool Execution Fails

**Symptom**: MCP tool returns error

**Solution**:
1. Test CLI command manually:
   ```bash
   ./meta-cc query tools --tool Bash --limit 5
   ```
2. Check session file exists:
   ```bash
   ls ~/.claude/projects/-home-*
   ```
3. Verify working directory is project root

### No Results Returned

**Symptom**: Tool runs but returns empty results

**Solution**:
- Check tool name spelling (case-sensitive: "Bash", not "bash")
- Verify regex pattern is valid
- Ensure error_signature exists
- Check file path is exact
- Increase limit parameter

### Pattern Matching Issues

**Symptom**: `query_user_messages` returns unexpected results

**Solution**:
- Test regex pattern with `grep`:
  ```bash
  echo "Phase 8 test" | grep -E "Phase [0-9]"
  ```
- Escape special characters: `\.`, `\*`, `\+`, `\?`, `\[`, `\]`
- Use `\|` for OR: `error\|bug`

### File Reference Mode Issues

**Issue**: "File not found" error

**Cause**: File was cleaned up or session hash changed

**Solution**:
1. Re-run the query to regenerate the file
2. Check retention period (default 7 days)
3. Verify `/tmp` directory permissions

**Issue**: Disk space exhaustion

**Cause**: Too many large temp files accumulated

**Solution**:
1. Run manual cleanup: `cleanup_temp_files(max_age_days: 1)`
2. Check disk space: `df -h /tmp`

---

## Usage Examples

### Example 1: Debug Workflow Errors

**User**: "Why do my commands keep failing?"

**Claude** (automatically):
```
1. query_tools(status="error", limit=20)
   → Found 5 Bash errors

2. analyze_errors()
   → Detected pattern: "npm test" failing repeatedly

3. query_context(error_signature="npm_test_fail", window=3)
   → Context shows missing test file each time

4. Analysis:
   - Root cause: tests/main.test.js not found
   - Suggestion: Create test file or fix path
```

### Example 2: Workflow Optimization

**User**: "Help me optimize my development workflow"

**Claude** (automatically):
```
1. get_session_stats()
   → 150 turns, 67 tool calls, 5% error rate

2. query_tool_sequences(min_occurrences=3)
   → Found: "Read -> Edit -> Bash" repeated 12 times

3. query_tools(tool="Bash")
   → Bash used 30 times (most frequent)

4. Recommendations:
   - Create Slash Command for "Read -> Edit -> Bash" workflow
   - Add Hook for automatic test execution
   - Most common: "npm test" → Create /test shortcut
```

### Example 3: Message Search

**User**: "Did I mention Phase 8 implementation earlier?"

**Claude** (automatically):
```
1. query_user_messages(pattern="Phase 8.*(implement|detail)", limit=10)

2. Results:
   - Turn 45: "Let's start Phase 8 implementation..."
   - Turn 67: "Phase 8 details should include..."

3. Summary: Yes, discussed Phase 8 implementation 3 times
```

---

## Migration Guide

### From Phase 15 (Pre-Hybrid Output)

Existing MCP clients expecting raw data arrays can use legacy mode:

**Before (Phase 15)**:
```json
// Response
[{"tool": "Read", "status": "success"}]
```

**After (Phase 16+, default)**:
```json
// Response
{
  "mode": "inline",
  "data": [{"tool": "Read", "status": "success"}]
}
```

**Legacy compatibility**:
```json
// Request with output_mode=legacy
{"output_mode": "legacy"}

// Response (raw array)
[{"tool": "Read", "status": "success"}]
```

### From Truncation to Hybrid Mode

**Before (Phase 15 - Truncation)**:
```json
// Data truncated at limit
{"max_message_length": 500}
→ {"content": "Truncated... [OUTPUT TRUNCATED]"}
```

**After (Phase 16+ - Hybrid Mode)**:
```json
// No truncation, complete data preserved
{}
→ file_ref mode for large results (no data loss)
```

**Migration Checklist**:
- ✅ Remove `max_message_length` parameter - Default is 0 (no truncation)
- ✅ Remove `content_summary` parameter - Hybrid mode provides better preservation
- ✅ No code changes needed - Backward compatible
- ✅ Use Read/Grep tools - Process large results from file_ref

---

## Comparison: MCP vs Slash Commands vs CLI

### When to Use MCP Tools

**Use MCP when**:
- Asking exploratory questions
- Combining multiple queries
- Letting Claude reason about what to query
- Natural language interaction preferred

**Example**:
```
"Where can I optimize my workflow?"
→ Claude autonomously queries stats, errors, sequences, messages
```

### When to Use Slash Commands

**Use Slash Commands when**:
- Repeated workflows
- Predictable outputs
- Fast execution needed
- Specific command preference

**Example**:
```
/meta-stats
/meta-timeline 50
```

### When to Use CLI Directly

**Use CLI when**:
- Scripting or automation
- Outside Claude Code
- Debugging meta-cc itself
- Piping to other tools

**Example**:
```bash
meta-cc query tools --tool Bash --status error | jq '.tools | length'
```

---

## Related Documentation

- [Integration Guide](integration-guide.md) - Choosing between MCP/Slash/Subagent
- [Examples & Usage](examples-usage.md) - Step-by-step setup guides
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
- [Capabilities Guide](capabilities-guide.md) - Capability development guide
- [CLAUDE.md](../CLAUDE.md) - Project instructions for Claude Code

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

## Parameter Ordering Convention

meta-cc MCP tools follow a **tier-based parameter ordering** convention for consistency and readability.

### Tier System

Parameters are organized into 5 tiers, from highest to lowest priority:

**Tier 1: Required Parameters**
- Absolutely necessary for the tool to function
- Examples: `pattern` (query_user_messages), `error_signature` (query_context), `file` (query_file_access), `where` (query_tools_advanced)

**Tier 2: Filtering Parameters**
- Narrow down or filter the data set
- Examples: `tool`, `status`, `pattern_target`, `where`

**Tier 3: Range Parameters**
- Define boundaries or limits on the data
- Examples: `min_*`, `max_*`, `start_*`, `end_*`, `threshold`, `window`

**Tier 4: Output Control Parameters**
- Control how much data is returned
- Examples: `limit`, `offset`, `page`, `cursor`, `top`

**Tier 5: Standard Parameters** (applied automatically)
- Common across all query tools
- Examples: `scope`, `jq_filter`, `stats_only`, `stats_first`, `output_format`, `inline_threshold_bytes`
- These parameters are added automatically by the MCP server (see Standard Parameters section)

### Why This Ordering?

1. **Consistency**: Same parameter order across all tools reduces cognitive load
2. **Readability**: Most important parameters appear first
3. **Predictability**: Users can anticipate where parameters will be

### Important Note

JSON parameter order **does not affect function calls** (parameters are key-value pairs). This convention is purely for documentation clarity and readability.

### Reference

See `/home/yale/work/meta-cc/experiments/bootstrap-006-api-design/data/api-parameter-convention.md` for complete specification and design rationale.

---

## Tool Catalog

meta-cc-mcp provides **20 standardized tools** for analyzing Claude Code session history.

### Tool Architecture (v2.0)

**Core Tools (2)**:
- `query` - Unified query interface with composable filtering
- `query_raw` - Raw jq expressions for power users

**Convenience Tools (8)**:
- `query_tool_errors`, `query_token_usage`, `query_conversation_flow`, `query_system_errors`
- `query_file_snapshots`, `query_timestamps`, `query_summaries`, `query_tool_blocks`

**Legacy Tools (7)**:
- `query_tools`, `query_user_messages`, `query_tool_sequences`, `query_file_access`
- `query_project_state`, `query_successful_prompts`, `get_session_stats`

**Utility Tools (3)**:
- `cleanup_temp_files`, `list_capabilities`, `get_capability`

### Migration from v1.x

**⚠️ Breaking Changes (v2.0)**: The following tools have been removed:
- `query_context` - Use `query` with jq filtering
- `query_tools_advanced` - Use `query` with filter parameters
- `query_time_series` - Use `query` with jq grouping
- `query_assistant_messages` - Use `query_token_usage` + jq
- `query_conversation` - Use `query_conversation_flow`
- `query_files` - Use `query_file_snapshots` + jq

See [MCP v2.0 Migration Guide](mcp-v2-migration.md) for complete migration instructions with 20+ examples.

---

## Query Behavior: Session-First Architecture

**Understanding Session Boundaries**:

Claude Code organizes session data as **one JSONL file per session**. Each file represents a complete conversation context. meta-cc convenience tools respect this natural organization:

### Default Behavior (Session-First)

All convenience query tools (`query_user_messages`, `query_tools`, etc.) use **session-aware ordering**:

1. **Sessions ordered by recency**: Most recent sessions first (by file modification time)
2. **Messages chronological within each session**: Oldest to newest preserves conversation flow
3. **Session boundaries preserved**: Complete context maintained within each session
4. **Limit applies across sessions**: But respects session boundaries when possible

**Why Session-First?**

- **Most common use case**: "What was I working on recently?" → Recent sessions matter most
- **Context preservation**: See complete conversation flow within each session
- **Performance**: Files already separated by session → no cross-file sorting needed
- **Natural organization**: Matches how Claude Code stores data

### Example: Query Behavior

```javascript
// Project has 3 sessions:
// - Session A (newest): 2025-10-26, 50 messages
// - Session B (middle): 2025-10-25, 150 messages
// - Session C (oldest): 2025-10-24, 80 messages

query_user_messages({pattern: ".*", limit: 100})

// Returns:
// - All 50 messages from Session A (newest)
// - First 50 messages from Session B (to reach limit=100)
// - Session C not read (limit already reached)
// - All messages chronological within their session
```

### Advanced Use Cases

**For global timestamp ordering** (cross-session, precise timestamp order):

```javascript
// Use Phase 27 two-stage query
const dir = await get_session_directory({scope: "project"});
await execute_stage2_query({
    files: dir.files,
    filter: 'select(.type == "user")',
    sort: 'sort_by(.timestamp) | reverse',  // Global sort
    limit: 100
});
```

**For historical analysis** (oldest sessions first):

```javascript
// Stage 1: Get directory and select oldest files
const dir = await get_session_directory({scope: "project"});
const oldestFiles = dir.files
    .sort((a, b) => a.modified_time - b.modified_time)  // Oldest first
    .slice(0, 5);

// Stage 2: Query selected files
await execute_stage2_query({
    files: oldestFiles,
    filter: 'select(.type == "user")',
    limit: 100
});
```

---

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

**Session-First Ordering** ✨:

This tool uses **session-aware ordering** optimized for viewing recent activity:

1. **Sessions ordered by recency**: Most recent sessions first (by file modification time)
2. **Messages chronological within sessions**: Oldest to newest within each session
3. **Session boundaries preserved**: Maintains complete context within each session
4. **Limit behavior**: Applies across sessions while respecting session boundaries when possible

**Example Behavior**:
```javascript
// Request: Get 100 recent user messages
query_user_messages({pattern: ".*", limit: 100})

// Returns:
// - All messages from most recent session (e.g., 50 messages)
// - If < 100, continues with next most recent session (e.g., 50 more)
// - Maintains chronological order within each session
// - Preserves full session context (from session start)
```

**For Global Timestamp Ordering**:

If you need precise global timestamp ordering across all sessions instead of session-first:

```javascript
// Use two-stage query for cross-session timestamp ordering
const dir = await get_session_directory({scope: "project"});
await execute_stage2_query({
    files: dir.files,
    filter: 'select(.type == "user")',
    sort: 'sort_by(.timestamp) | reverse',
    limit: 100
});
```

**Examples**:
```json
// Find error-related messages (from recent sessions)
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
// Basic error context (default window=3)
{"error_signature": "Bash:command not found"}

// Wide context window (5 turns before/after)
{"error_signature": "permission denied", "window": 5}

// Analyze test failures
{"error_signature": "npm test failed", "window": 3}
```

**Practical Use Cases**:

1. **Debug Bash Command Errors**:
   ```json
   // Problem: "command not found" errors keep occurring
   {"error_signature": "Bash:command not found", "window": 3}
   // Returns: User messages leading up to error, assistant responses, and subsequent attempts
   ```

2. **Investigate Permission Issues**:
   ```json
   // Problem: File operations failing with permission denied
   {"error_signature": "permission denied", "window": 5}
   // Returns: Which files were accessed, what operations were attempted, pattern of failures
   ```

3. **Understand Test Failures**:
   ```json
   // Problem: Tests pass sometimes, fail other times
   {"error_signature": "test failed", "window": 3}
   // Returns: Code changes before test failure, environment differences, retry attempts
   ```

**What You Get**:
- Turns immediately before the error (user requests, code changes)
- The exact error occurrence (command, output, timestamp)
- Turns immediately after (recovery attempts, workarounds)
- Pattern visibility (repeated failures, progressive changes)

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
// Complex filter (Bash errors taking > 5 seconds)
{"where": "tool='Bash' AND status='error' AND duration>5000"}

// Multiple tools (file operations)
{"where": "tool IN ('Read', 'Edit', 'Write')"}

// Time range (October 2025 activity)
{"where": "timestamp BETWEEN '2025-10-01' AND '2025-10-05'"}
```

**Practical Use Cases**:

1. **Slow Command Analysis**:
   ```json
   // Problem: Some commands are taking too long
   {"where": "tool='Bash' AND duration>10000"}
   // Returns: All Bash commands that took > 10 seconds
   // Analysis: Identify slow scripts, optimize workflows
   ```

2. **Error Pattern Detection**:
   ```json
   // Problem: Want to find all Read errors on specific file pattern
   {"where": "tool='Read' AND status='error' AND filepath LIKE '%/test/%'"}
   // Returns: Read errors in test directories
   // Analysis: Permission issues, missing files in test paths
   ```

3. **Tool Usage Comparison**:
   ```json
   // Problem: Compare usage of different edit tools
   {"where": "tool IN ('Edit', 'Write', 'NotebookEdit') AND timestamp>'2025-10-01'"}
   // Returns: All edit operations since Oct 1
   // Analysis: Which edit tool is used most, success rates
   ```

4. **Recent Activity Filtering**:
   ```json
   // Problem: Focus on last week's tool calls
   {"where": "timestamp BETWEEN '2025-10-08' AND '2025-10-15' AND status='error'"}
   // Returns: All errors from specific week
   // Analysis: Weekly error trends, problematic periods
   ```

5. **Multi-Condition Filtering**:
   ```json
   // Problem: Complex scenario - failed Bash commands with specific pattern
   {"where": "tool='Bash' AND status='error' AND (duration>5000 OR output REGEXP 'timeout|killed')"}
   // Returns: Slow or terminated Bash commands
   // Analysis: System resource issues, hanging processes
   ```

**SQL Expression Reference**:

| Operator | Example | Description |
|----------|---------|-------------|
| `=` | `tool='Bash'` | Exact match |
| `!=` | `status!='success'` | Not equal |
| `>`, `<` | `duration>5000` | Numeric comparison |
| `>=`, `<=` | `duration>=10000` | Numeric comparison (inclusive) |
| `IN` | `tool IN ('Read', 'Edit')` | Multiple values (OR) |
| `NOT IN` | `tool NOT IN ('Bash')` | Exclusion |
| `BETWEEN` | `duration BETWEEN 1000 AND 5000` | Range (inclusive) |
| `LIKE` | `filepath LIKE '%test%'` | Pattern matching (% = wildcard) |
| `REGEXP` | `output REGEXP 'error\|failed'` | Regex matching |
| `AND` | `tool='Bash' AND status='error'` | Logical AND |
| `OR` | `duration>5000 OR status='error'` | Logical OR |
| `NOT` | `NOT (tool='Bash')` | Logical NOT |

**When to Use**:
- **Complex filters**: Multiple conditions not achievable with basic query_tools
- **Pattern matching**: LIKE/REGEXP for flexible string matching
- **Range queries**: BETWEEN for time/duration ranges
- **Performance analysis**: Duration-based filtering
- **Advanced debugging**: Combining multiple criteria

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

**Examples**:
```json
// Default cleanup (7+ days old)
{"max_age_days": 7}
→ {"removed_count": 12, "freed_bytes": 5242880}

// Aggressive cleanup (1+ days old)
{"max_age_days": 1}
→ {"removed_count": 45, "freed_bytes": 15728640}

// Keep only today's files
{"max_age_days": 0}
→ {"removed_count": 67, "freed_bytes": 23068672}
```

**Practical Use Cases**:

1. **Regular Maintenance**:
   ```json
   // Run weekly to clean old temp files
   {"max_age_days": 7}
   // Returns: Number of files removed, bytes freed
   ```

2. **Disk Space Emergency**:
   ```json
   // Problem: /tmp is full, need space immediately
   {"max_age_days": 1}
   // Effect: Removes all but today's temp files (aggressive cleanup)
   ```

3. **Pre-Large Query**:
   ```json
   // Problem: About to run large query, want clean /tmp
   {"max_age_days": 3}
   // Effect: Clear old files first, then run query with fresh temp space
   ```

**When to Use**:
- **Disk space low**: Run with `max_age_days: 1` to free space quickly
- **Regular maintenance**: Weekly cleanup with `max_age_days: 7` (default)
- **Before large queries**: Clean old files to avoid /tmp exhaustion
- **Session cleanup**: After intensive analysis session

**What You Get**:
- `removed_count`: Number of temp files deleted
- `freed_bytes`: Disk space recovered (in bytes)
- No data loss: Only removes temp files older than threshold

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

- [Integration Guide](integration.md) - Choosing between MCP/Slash/Subagent
- [Examples & Usage](../tutorials/examples.md) - Step-by-step setup guides
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
- [Capabilities Guide](capabilities.md) - Capability development guide
- [CLAUDE.md](../../CLAUDE.md) - Project instructions for Claude Code

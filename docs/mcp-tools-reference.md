# MCP Tools Complete Reference

## Overview

meta-cc-mcp provides **12 standardized tools** for analyzing Claude Code session history. All tools support the same core parameters for consistency and flexibility.

> **Migration Note**: `extract_tools` has been removed. Use `query_tools` instead for all tool extraction needs.

### Standard Parameters

All tools support these parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `scope` | string | "project" | Query scope: "project" (cross-session) or "session" (current only) |
| `jq_filter` | string | ".[]" | jq expression for filtering and transforming results |
| `stats_only` | boolean | false | Return only statistics, no detailed records |
| `stats_first` | boolean | false | Return statistics first, then details (separated by `---`) |
| `inline_threshold_bytes` | number | 8192 | **Phase 16.6**: Threshold for inline vs file_ref mode (default: 8KB). Configure via parameter or `META_CC_INLINE_THRESHOLD` env var. |
| `max_message_length` | number | 500 | **Phase 15**: Max chars per message content (0=unlimited, prevents huge summaries) |
| `content_summary` | boolean | false | **Phase 15**: Return only turn/timestamp/preview (100 chars), skip full content |
| `output_format` | string | "jsonl" | Output format: "jsonl" or "tsv" |

**Phase 16.6 Changes**:
- ❌ **Removed**: `max_output_bytes` - No longer needed (hybrid mode handles size)
- ✅ **Added**: `inline_threshold_bytes` - Controls inline vs file_ref mode threshold
- ✅ **No Truncation**: All data is preserved via hybrid output mode (inline ≤8KB, file_ref >8KB)

---

## Output Size Control (Phase 15)

### Problem: Context Overflow from Large Messages

**Issue**: User messages can contain session summaries (thousands of lines), causing MCP responses to return ~10.7k tokens and overflow Claude's context window.

**Solution**: Phase 15 introduces two new parameters for message content control:

1. **`max_message_length`** (default: 500 chars)
   - Truncates message content to prevent large outputs
   - Adds `content_truncated: true` and `original_length` fields
   - Reduces output from ~10.7k to ~1.5k tokens (86% reduction)

2. **`content_summary`** (default: false)
   - Returns only metadata: `{turn_sequence, timestamp, content_preview}`
   - Preview limited to 100 characters
   - Reduces output to ~800 tokens (93% reduction)

### Performance Impact

| Configuration | Output Size | Reduction | Use Case |
|--------------|-------------|-----------|----------|
| Default (no truncation) | ~10.7k tokens | - | Small projects only |
| `max_message_length: 500` | ~1.5k tokens | **86%** | Most queries (recommended) |
| `content_summary: true` | ~800 tokens | **93%** | Metadata-only queries |
| `jq_filter` + `stats_only` | <100 tokens | **>99%** | Aggregation queries |

### When to Use Output Control

**Use `max_message_length`** when:
- You need message content but want to limit size
- Analyzing message patterns
- Searching for specific keywords
- Default for all `query_user_messages` calls

**Use `content_summary`** when:
- You only need metadata (timestamps, turn numbers)
- Pattern matching without content analysis
- Building message timelines
- Quick exploratory queries

**Use `stats_only` + `jq_filter`** when:
- Performing aggregation queries (counts, averages)
- Statistical analysis only
- Minimizing token usage
- No content needed

### Examples

**Example 1: Default (Problematic)**
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "implement.*feature",
    "limit": 5
  }
}
```
**Output**: ~10.7k tokens ❌ (May overflow context)

**Example 2: With Truncation (Recommended)**
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "implement.*feature",
    "limit": 5,
    "max_message_length": 500
  }
}
```
**Output**: ~1.5k tokens ✅ (86% reduction)

**Example 3: Content Summary (Metadata Only)**
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": ".*",
    "limit": 10,
    "content_summary": true
  }
}
```
**Output**: ~800 tokens ✅ (93% reduction)

**Example 4: Statistical Query (Minimal)**
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "error|fail",
    "jq_filter": "length",
    "stats_only": true
  }
}
```
**Output**: <100 tokens ✅ (>99% reduction)

---

## Tool Catalog

### 1. get_session_stats

**Purpose**: Get statistical information about the current session.

**Default Scope**: session (current session only)

**Tool-Specific Parameters**: None

**Use Cases**:
- Quick session overview
- Check error rate
- Understand tool usage distribution
- Evaluate session efficiency

**Example 1 - Basic Stats**:
```json
{
  "name": "get_session_stats",
  "arguments": {
    "stats_only": true
  }
}
```

**Returns**:
```json
{"TurnCount": 45, "ToolCallCount": 123, "ErrorCount": 5, "ErrorRate": 0.04}
```

**Example 2 - With jq Filter**:
```json
{
  "name": "get_session_stats",
  "arguments": {
    "jq_filter": "{turns: .TurnCount, tools: .ToolCallCount, errors: .ErrorCount}"
  }
}
```

**Returns**:
```json
{"turns": 45, "tools": 123, "errors": 5}
```

---

### 2. query_tools

**Purpose**: Query tool calls with flexible filtering options.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `limit` (number): Maximum results (no limit by default, rely on hybrid output mode)
- `tool` (string): Filter by tool name
- `status` (string): Filter by status ("error" or "success")

**Use Cases**:
- Find specific tool usage
- Filter errors by tool
- Analyze tool performance
- Statistical analysis

**Example 1 - All Bash Errors**:
```json
{
  "name": "query_tools",
  "arguments": {
    "tool": "Bash",
    "status": "error",
    "limit": 10
  }
}
```

**Example 2 - Error Distribution by Tool**:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})",
    "stats_only": true
  }
}
```

**Returns**:
```jsonl
{"tool": "Bash", "count": 311}
{"tool": "Read", "count": 62}
{"tool": "Edit", "count": 15}
```

**Example 3 - Recent Errors with Context**:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "limit": 5,
    "jq_filter": ".[] | {tool: .ToolName, error: .Error, timestamp: .Timestamp}",
    "stats_first": true
  }
}
```

---

### 4. query_user_messages

**Purpose**: Search user messages using regex patterns.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `pattern` (string, **required**): Regex pattern to match
- `limit` (number): Maximum results (no limit by default, rely on hybrid output mode)
- `max_message_length` (number): **Phase 15**: Max chars per message (default: 500, 0=unlimited)
- `content_summary` (boolean): **Phase 15**: Return only metadata (default: false)

**⚠️ IMPORTANT**: This tool can return very large outputs due to session summaries in messages. **Always use `max_message_length` or `content_summary`** to prevent context overflow.

**Use Cases**:
- Find similar historical questions
- Discover recurring user intents
- Analyze prompt evolution
- Track topic changes

**Example 1 - With Truncation (Recommended)**:
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "error|fix|bug",
    "limit": 5,
    "max_message_length": 300
  }
}
```
**Returns**:
```jsonl
{"turn_sequence": 42, "timestamp": "2025-10-06T12:00:00Z", "content": "Can you help me fix this error... [TRUNCATED]", "content_truncated": true, "original_length": 8500}
```

**Example 2 - Content Summary (Metadata Only)**:
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "implement.*feature",
    "limit": 10,
    "content_summary": true
  }
}
```
**Returns**:
```jsonl
{"turn_sequence": 42, "timestamp": "2025-10-06T12:00:00Z", "content_preview": "Can you implement a new feature for..."}
```

**Example 3 - Count Messages by Topic**:
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "test|testing|spec",
    "jq_filter": "length",
    "stats_only": true
  }
}
```

**Example 4 - Recent Documentation Requests**:
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "doc(s)?|documentation|README",
    "limit": 3,
    "scope": "session",
    "max_message_length": 200
  }
}
```

---

### 5. query_context

**Purpose**: Query context around errors (turns before and after).

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `error_signature` (string, **required**): Error signature or pattern ID
- `window` (number): Context window size in turns (default: 3)

**Use Cases**:
- Understand error causes
- Identify error triggers
- Analyze recovery flows
- Debug systematic issues

**Example 1 - Error Context**:
```json
{
  "name": "query_context",
  "arguments": {
    "error_signature": "Bash:command not found",
    "window": 3
  }
}
```

**Example 2 - Wide Context Window**:
```json
{
  "name": "query_context",
  "arguments": {
    "error_signature": "permission denied",
    "window": 5,
    "jq_filter": ".[] | {turn: .TurnSequence, type: .Type, content: .Content | .[0:200]}"
  }
}
```

---

### 6. query_tool_sequences

**Purpose**: Detect repeated workflow patterns (tool sequences).

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `pattern` (string): Sequence pattern to match (e.g., "Read -> Edit -> Bash")
- `min_occurrences` (number): Minimum occurrences to report (default: 3)
- `include_builtin_tools` (boolean): Include built-in tools (Bash, Read, Edit, etc.). Default: false

**Performance Note**: By default, built-in tools (Bash, Read, Edit, etc.) are excluded from analysis for:
- **35x faster execution** (~30s → <1s for large projects)
- **Cleaner workflow patterns** (focus on MCP tools and business logic)
- **97% data reduction** (10,892 → 305 tool calls in typical projects)

Use `include_builtin_tools=true` only when debugging low-level tool usage.

**Use Cases**:
- Identify repetitive workflows
- Find automation opportunities
- Analyze tool combinations
- Understand workflow habits
- Discover MCP tool orchestration patterns

**Example 1 - Common Sequences**:
```json
{
  "name": "query_tool_sequences",
  "arguments": {
    "min_occurrences": 5,
    "jq_filter": ".[] | select(.Occurrences >= 5) | {sequence: .Sequence, count: .Occurrences}"
  }
}
```

**Example 2 - Specific Pattern**:
```json
{
  "name": "query_tool_sequences",
  "arguments": {
    "pattern": "Read -> Edit",
    "min_occurrences": 1,
    "stats_only": true
  }
}
```

**Example 3 - Top 10 Sequences**:
```json
{
  "name": "query_tool_sequences",
  "arguments": {
    "min_occurrences": 3,
    "jq_filter": "sort_by(.Occurrences) | reverse | .[0:10]"
  }
}
```

**Example 4 - Include Built-in Tools (for debugging)**:
```json
{
  "name": "query_tool_sequences",
  "arguments": {
    "min_occurrences": 5,
    "include_builtin_tools": true,
    "jq_filter": ".[] | select(.Pattern | contains(\"Bash\"))"
  }
}
```

---

### 7. query_file_access

**Purpose**: Query operation history for a specific file.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `file` (string, **required**): File path to query

**Use Cases**:
- Track file modification history
- Analyze file churn
- Understand file evolution
- Identify frequently edited files

**Example 1 - File History**:
```json
{
  "name": "query_file_access",
  "arguments": {
    "file": "cmd/mcp-server/tools.go",
    "jq_filter": ".[] | {timestamp: .Timestamp, operation: .Operation, tool: .ToolName}"
  }
}
```

**Example 2 - Count Operations by Type**:
```json
{
  "name": "query_file_access",
  "arguments": {
    "file": "README.md",
    "jq_filter": "group_by(.Operation) | map({op: .[0].Operation, count: length})",
    "stats_only": true
  }
}
```

**Example 3 - Recent File Access**:
```json
{
  "name": "query_file_access",
  "arguments": {
    "file": "docs/mcp-tools-reference.md",
    "scope": "session",
    "jq_filter": ".[-5:]"
  }
}
```

---

### 8. query_project_state

**Purpose**: Query project state evolution across sessions.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**: None

**Use Cases**:
- Track project evolution
- Analyze task progression
- Identify focus shifts
- Understand development phases

**Example 1 - Project Timeline**:
```json
{
  "name": "query_project_state",
  "arguments": {
    "jq_filter": ".[] | {session: .SessionID, active_files: .ActiveFiles, tasks: .Tasks}"
  }
}
```

**Example 2 - Session Count**:
```json
{
  "name": "query_project_state",
  "arguments": {
    "jq_filter": "length",
    "stats_only": true
  }
}
```

---

### 9. query_successful_prompts

**Purpose**: Find historically successful prompt patterns.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `limit` (number): Maximum results (no limit by default, rely on hybrid output mode)
- `min_quality_score` (number): Minimum quality score 0-1 (default: 0.8)

**Use Cases**:
- Learn effective prompting
- Improve prompt quality
- Create prompt templates
- Discover best practices

**Example 1 - High-Quality Prompts**:
```json
{
  "name": "query_successful_prompts",
  "arguments": {
    "limit": 5,
    "min_quality_score": 0.9,
    "jq_filter": ".[] | {prompt: .Content | .[0:100], score: .QualityScore}"
  }
}
```

**Example 2 - Top 10 Prompts**:
```json
{
  "name": "query_successful_prompts",
  "arguments": {
    "limit": 10,
    "min_quality_score": 0.7,
    "stats_only": true
  }
}
```

---

### 10. query_tools_advanced

**Purpose**: Advanced tool queries with SQL-like filter expressions.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `where` (string, **required**): SQL-like filter expression
- `limit` (number): Maximum results (no limit by default, rely on hybrid output mode)

**Use Cases**:
- Complex multi-condition queries
- Advanced data analysis
- Precise filtering
- Research and investigation

**Example 1 - Complex Filter**:
```json
{
  "name": "query_tools_advanced",
  "arguments": {
    "where": "tool='Bash' AND status='error' AND duration>5000",
    "limit": 10
  }
}
```

**Example 2 - Multiple Tools**:
```json
{
  "name": "query_tools_advanced",
  "arguments": {
    "where": "tool IN ('Read', 'Edit', 'Write')",
    "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})",
    "stats_only": true
  }
}
```

**Example 3 - Time Range**:
```json
{
  "name": "query_tools_advanced",
  "arguments": {
    "where": "timestamp BETWEEN '2025-10-01' AND '2025-10-05'",
    "limit": 50
  }
}
```

---

### 11. query_time_series

**Purpose**: Analyze metrics over time intervals.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `interval` (string): Time interval: "hour", "day", or "week" (default: "hour")
- `metric` (string): Metric to analyze: "tool-calls" or "error-rate" (default: "tool-calls")
- `where` (string): Optional filter expression

**Use Cases**:
- Temporal pattern analysis
- Peak usage detection
- Error clustering
- Workflow rhythm understanding

**Example 1 - Daily Tool Calls**:
```json
{
  "name": "query_time_series",
  "arguments": {
    "interval": "day",
    "metric": "tool-calls",
    "jq_filter": ".[] | {date: .Timestamp, count: .ToolCalls}"
  }
}
```

**Example 2 - Error Rate by Hour**:
```json
{
  "name": "query_time_series",
  "arguments": {
    "interval": "hour",
    "metric": "error-rate",
    "jq_filter": ".[] | {hour: .Timestamp, rate: .ErrorRate}"
  }
}
```

**Example 3 - Bash-Only Timeline**:
```json
{
  "name": "query_time_series",
  "arguments": {
    "interval": "day",
    "metric": "tool-calls",
    "where": "tool='Bash'"
  }
}
```

---

### 12. query_files

**Purpose**: File-level operation statistics.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `sort_by` (string): Sort field: "total_ops", "edit_count", "read_count", "write_count", "error_count", "error_rate" (default: "total_ops")
- `top` (number): Return top N files (default: 20)
- `where` (string): Optional filter expression

**Use Cases**:
- Identify file hotspots
- Find frequently modified files
- Analyze file error rates
- Understand codebase churn

**Example 1 - Top Edited Files**:
```json
{
  "name": "query_files",
  "arguments": {
    "sort_by": "edit_count",
    "top": 10,
    "jq_filter": ".[] | {file: .FilePath, edits: .EditCount}"
  }
}
```

**Example 2 - High Error Rate Files**:
```json
{
  "name": "query_files",
  "arguments": {
    "sort_by": "error_rate",
    "top": 10,
    "where": "error_count > 0",
    "jq_filter": ".[] | {file: .FilePath, errors: .ErrorCount, rate: .ErrorRate}"
  }
}
```

**Example 3 - Most Active Files**:
```json
{
  "name": "query_files",
  "arguments": {
    "sort_by": "total_ops",
    "top": 5,
    "stats_only": true
  }
}
```

---

### 13. analyze_errors (DEPRECATED)

**Purpose**: Error pattern analysis (deprecated - use query_tools instead).

**Status**: DEPRECATED in Phase 15

**Migration**: Use `query_tools` with `status="error"` and `jq_filter` for grouping.

**Old Way**:
```json
{
  "name": "analyze_errors",
  "arguments": {
    "scope": "project"
  }
}
```

**New Way**:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "jq_filter": ".[]",
    "stats_only": true
  }
}
```

---

## Parameter Usage Guide

### jq_filter Expression Cookbook

#### Basic Filtering

**Select errors only**:
```jq
.[] | select(.Status == "error")
```

**Project specific fields**:
```jq
.[] | {tool: .ToolName, status: .Status}
```

**Combined filter and projection**:
```jq
.[] | select(.Status == "error") | {tool: .ToolName, error: .Error}
```

#### Aggregation and Statistics

**Group by tool and count**:
```jq
group_by(.ToolName) | map({tool: .[0].ToolName, count: length})
```

**Calculate error rate by tool**:
```jq
group_by(.ToolName) | map({
  tool: .[0].ToolName,
  total: length,
  errors: map(select(.Status == "error")) | length
})
```

**Top N sorted**:
```jq
group_by(.ToolName) | map({tool: .[0].ToolName, count: length}) | sort_by(.count) | reverse | .[0:10]
```

**Sum and average**:
```jq
{
  total: length,
  avg_duration: (map(.Duration) | add / length)
}
```

#### Array Operations

**Get last N items**:
```jq
.[-10:]
```

**Get first N items**:
```jq
.[0:10]
```

**Slice range**:
```jq
.[10:20]
```

**Length**:
```jq
length
```

#### Time-Based Filtering

**After specific date**:
```jq
.[] | select(.Timestamp > "2025-10-01")
```

**Date range**:
```jq
.[] | select(.Timestamp >= "2025-10-01" and .Timestamp <= "2025-10-05")
```

**Recent items**:
```jq
sort_by(.Timestamp) | reverse | .[0:20]
```

#### String Operations

**Case-insensitive match**:
```jq
.[] | select(.Error | ascii_downcase | contains("permission"))
```

**Regex test**:
```jq
.[] | select(.Error | test("error|failed|timeout"; "i"))
```

**Extract substring**:
```jq
.[] | {tool: .ToolName, error_preview: .Error[0:50]}
```

#### Complex Queries

**Multiple conditions**:
```jq
.[] | select(.ToolName == "Bash" and .Status == "error" and .Duration > 1000)
```

**Nested grouping**:
```jq
group_by(.ToolName) | map({
  tool: .[0].ToolName,
  by_status: group_by(.Status) | map({status: .[0].Status, count: length})
})
```

**Flatten nested arrays**:
```jq
.[] | .Items[] | {field: .Value}
```

---

### stats_only vs stats_first

**When to use stats_only**:
- Quick overview needed
- Minimize token usage
- Summary analysis only
- High-level metrics

**Example**:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})",
    "stats_only": true
  }
}
```

**Returns**:
```jsonl
{"tool": "Bash", "count": 311}
{"tool": "Read", "count": 62}
```

**When to use stats_first**:
- Need both summary and details
- Context-aware analysis
- Progressive disclosure
- Initial overview with drill-down

**Example**:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "limit": 100,
    "stats_first": true
  }
}
```

**Returns**:
```jsonl
{"total": 100, "tools": {"Bash": 75, "Read": 25}}
---
{"ToolName": "Bash", "Status": "error", "Error": "..."}
{"ToolName": "Bash", "Status": "error", "Error": "..."}
...
```

---

### Hybrid Output Mode (Phase 16.6)

**Default**: Automatic mode selection based on result size

The MCP server automatically selects between inline and file_ref mode:
- **Inline mode**: Results ≤8KB returned directly in response
- **File ref mode**: Results >8KB written to temp file, metadata returned

**Threshold Configuration**:

| Method | Priority | Example |
|--------|----------|---------|
| Parameter | Highest | `"inline_threshold_bytes": 16384` (16KB) |
| Environment | Medium | `export META_CC_INLINE_THRESHOLD=16384` |
| Default | Lowest | 8192 bytes (8KB) |

**Examples**:

**Default behavior** (auto-select based on size):
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error"
  }
}
```
**Result**: Inline mode if <8KB, file_ref mode if >8KB

**Custom threshold** (prefer larger inline mode):
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "inline_threshold_bytes": 16384  // 16KB threshold
  }
}
```
**Result**: Inline mode if <16KB, file_ref mode if >16KB

**Force file_ref mode** (for large datasets):
```json
{
  "name": "query_tools",
  "arguments": {
    "scope": "project",
    "output_mode": "file_ref"  // Force file_ref regardless of size
  }
}
```

**Note**: No data truncation occurs. All results are preserved via hybrid mode.

---

## Best Practices

### 1. Prefer jq_filter Over Multiple Calls

**Good** - Single call with jq:
```json
{
  "name": "query_tools",
  "arguments": {
    "jq_filter": ".[] | select(.Status == \"error\") | .ToolName",
    "stats_only": true
  }
}
```

**Bad** - Multiple calls:
```json
// Call 1: Get all data
{"name": "query_tools"}

// Call 2: Filter manually in conversation
// ... process in Claude ...
```

### 2. Use stats_only to Reduce Token Usage

**Good** - Stats only:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})",
    "stats_only": true
  }
}
```

**Bad** - Return all details:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error"
  }
}
```

### 3. Leverage Hybrid Output Mode

**For quick queries** - inline mode is automatic:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "limit": 10  // Small result set, likely inline mode
  }
}
```

**For large queries** - file_ref mode handles size:
```json
{
  "name": "query_tools",
  "arguments": {
    "scope": "project"  // Large result set, automatic file_ref mode
  }
}
```

**Custom threshold** - when you need different cutoff:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "inline_threshold_bytes": 16384  // Prefer 16KB inline mode
  }
}
```

**Note**: Hybrid mode eliminates the need for manual size limits. All data is preserved.

### 4. Combine Tools for Complex Analysis

**Scenario**: Analyze error context

**Step 1** - Find error signature:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "limit": 1,
    "jq_filter": ".[] | .Signature"
  }
}
```

**Step 2** - Get context:
```json
{
  "name": "query_context",
  "arguments": {
    "error_signature": "Bash:command not found",
    "window": 3
  }
}
```

### 5. Use scope Parameter Wisely

**Project scope (default)** - For meta-cognition:
```json
{
  "name": "query_tools",
  "arguments": {
    "scope": "project",
    "status": "error"
  }
}
```

**Session scope** - For current analysis:
```json
{
  "name": "query_tools",
  "arguments": {
    "scope": "session",
    "status": "error"
  }
}
```

### 6. Use TSV for Large Datasets

**JSONL** (default):
```json
{
  "name": "query_tools",
  "arguments": {
    "limit": 100,
    "output_format": "jsonl"
  }
}
```

**TSV** (86% smaller):
```json
{
  "name": "query_tools",
  "arguments": {
    "limit": 100,
    "output_format": "tsv"
  }
}
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

## FAQ

### Q: How do I choose between project vs session scope?

**A**: Use scope based on analysis goal:

- **project** (default): Cross-session patterns, long-term trends, meta-cognition
- **session**: Current session focus, immediate analysis, quick debugging

**Example**:
```json
// Long-term pattern discovery
{"scope": "project", "status": "error"}

// Current session debug
{"scope": "session", "status": "error"}
```

### Q: My jq_filter is not working. How do I debug?

**A**: Follow these steps:

1. Test on [jqplay.org](https://jqplay.org)
2. Check field names with `jq_filter: ".[]"`
3. Build incrementally:
   ```jq
   // Step 1: Select only
   .[] | select(.Status == "error")

   // Step 2: Add projection
   .[] | select(.Status == "error") | {tool: .ToolName}

   // Step 3: Add grouping
   .[] | select(.Status == "error") | {tool: .ToolName} | group_by(.tool)
   ```

### Q: Output is too large - what should I do?

**A**: Phase 16.6 eliminates truncation. Use these strategies:

1. **Hybrid mode is automatic** - Large results use file_ref mode
2. Use `jq_filter` to project fewer fields
3. Use `stats_only=true` for summaries
4. Use `limit` to reduce result count
5. Adjust `inline_threshold_bytes` to prefer file_ref mode:
   ```json
   {"inline_threshold_bytes": 4096}  // Force file_ref at 4KB
   ```
6. Use `output_format: "tsv"` for compact output

**Note**: No data is truncated. File_ref mode preserves all results in temporary files.

**Example**:
```json
{
  "name": "query_tools",
  "arguments": {
    "jq_filter": ".[] | {tool: .ToolName, status: .Status}",
    "output_format": "tsv"
  }
}
```

### Q: How do I implement pagination?

**A**: Use jq array slicing:

```json
// Page 1 (0-99)
{"jq_filter": ".[0:100]"}

// Page 2 (100-199)
{"jq_filter": ".[100:200]"}

// Page 3 (200-299)
{"jq_filter": ".[200:300]"}
```

### Q: Can I combine multiple filters?

**A**: Yes, use jq's `and` operator:

```json
{
  "name": "query_tools",
  "arguments": {
    "jq_filter": ".[] | select(.ToolName == \"Bash\" and .Status == \"error\" and .Duration > 1000)"
  }
}
```

### Q: How do I get unique values?

**A**: Use jq's `unique` or `group_by`:

```json
{
  "name": "query_tools",
  "arguments": {
    "jq_filter": "[.[] | .ToolName] | unique",
    "stats_only": true
  }
}
```

---

## Phase 15 Migration Guide

### Migration: Default → Truncated Output

**Before Phase 15** (Context overflow risk):
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "implement.*"
  }
}
```
**Output**: ~10.7k tokens ❌

**After Phase 15** (Recommended):
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "implement.*",
    "max_message_length": 500
  }
}
```
**Output**: ~1.5k tokens ✅

### Migration: Full Content → Summary Mode

**Before Phase 15**:
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": ".*",
    "jq_filter": ".[] | {turn: .turn_sequence, time: .timestamp}"
  }
}
```
**Output**: ~10.7k tokens (includes full content even if not projected)

**After Phase 15**:
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": ".*",
    "content_summary": true
  }
}
```
**Output**: ~800 tokens ✅ (93% reduction)

### Best Practices for Phase 15

1. **Always set `max_message_length` for `query_user_messages`**
   ```json
   {"max_message_length": 500}
   ```

2. **Use `content_summary` for metadata-only queries**
   ```json
   {"content_summary": true}
   ```

3. **Combine with `jq_filter` for targeted queries**
   ```json
   {
     "max_message_length": 300,
     "jq_filter": ".[] | select(.turn_sequence > 100)"
   }
   ```

4. **Use `stats_only` for aggregations**
   ```json
   {
     "jq_filter": "length",
     "stats_only": true
   }
   ```

### Troubleshooting: Context Overflow

**Symptom**: MCP response too large, Claude context warning

**Solution 1**: Add truncation
```json
{"max_message_length": 500}
```

**Solution 2**: Use summary mode
```json
{"content_summary": true}
```

**Solution 3**: Reduce limit
```json
{"limit": 5, "max_message_length": 300}
```

**Solution 4**: Project specific fields with jq
```json
{
  "jq_filter": ".[] | {turn: .turn_sequence, preview: .content[0:50]}",
  "max_message_length": 500
}
```

---

## Reference Resources

- [jq Manual](https://stedolan.github.io/jq/manual/) - Complete jq syntax reference
- [jqplay.org](https://jqplay.org) - Interactive jq playground
- [meta-cc README](../README.md) - Project overview
- [MCP Migration Guide](./mcp-migration-phase15.md) - Phase 15 migration
- [Integration Guide](./integration-guide.md) - MCP vs Slash Commands vs Subagent
- [MCP Protocol](https://modelcontextprotocol.io) - MCP specification

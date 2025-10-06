# MCP Tools Complete Reference

## Overview

meta-cc-mcp provides **13 standardized tools** for analyzing Claude Code session history. All tools support the same core parameters for consistency and flexibility.

### Standard Parameters

All tools support these parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `scope` | string | "project" | Query scope: "project" (cross-session) or "session" (current only) |
| `jq_filter` | string | ".[]" | jq expression for filtering and transforming results |
| `stats_only` | boolean | false | Return only statistics, no detailed records |
| `stats_first` | boolean | false | Return statistics first, then details (separated by `---`) |
| `max_output_bytes` | number | 51200 | Maximum output size in bytes (default: 50KB) |
| `output_format` | string | "jsonl" | Output format: "jsonl" or "tsv" |

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

### 2. extract_tools

**Purpose**: Extract complete tool call history for export or analysis.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `limit` (number): Maximum number of tools to extract (default: 100)

**Use Cases**:
- Export tool usage timeline
- Analyze tool usage evolution
- Create custom reports
- Feed data to external analytics

**Example 1 - Last 50 Tools**:
```json
{
  "name": "extract_tools",
  "arguments": {
    "limit": 50,
    "jq_filter": ".[] | {tool: .ToolName, status: .Status, timestamp: .Timestamp}"
  }
}
```

**Example 2 - Extract to TSV**:
```json
{
  "name": "extract_tools",
  "arguments": {
    "limit": 100,
    "output_format": "tsv",
    "max_output_bytes": 10240
  }
}
```

**Example 3 - Session-Only Extract**:
```json
{
  "name": "extract_tools",
  "arguments": {
    "scope": "session",
    "limit": 20
  }
}
```

---

### 3. query_tools

**Purpose**: Query tool calls with flexible filtering options.

**Default Scope**: project (cross-session)

**Tool-Specific Parameters**:
- `limit` (number): Maximum results (default: 20)
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
- `limit` (number): Maximum results (default: 10)

**Use Cases**:
- Find similar historical questions
- Discover recurring user intents
- Analyze prompt evolution
- Track topic changes

**Example 1 - Error-Related Messages**:
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "error|fix|bug",
    "limit": 5,
    "jq_filter": ".[] | {turn: .TurnSequence, content: .Content | .[0:100]}"
  }
}
```

**Example 2 - Count Messages by Topic**:
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

**Example 3 - Recent Documentation Requests**:
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "doc(s)?|documentation|README",
    "limit": 3,
    "scope": "session"
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

**Use Cases**:
- Identify repetitive workflows
- Find automation opportunities
- Analyze tool combinations
- Understand workflow habits

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
- `limit` (number): Maximum results (default: 10)
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
- `limit` (number): Maximum results (default: 20)

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

### max_output_bytes Usage

**Default**: 51200 bytes (50KB)

**When to adjust**:

| Scenario | Recommended Size | Example |
|----------|-----------------|---------|
| Quick queries (<100 results) | 10240 (10KB) | Session-only error check |
| Medium queries (<500 results) | 51200 (50KB) | Default - most queries |
| Large queries (>500 results) | Use jq_filter to limit | Combine with `.[0:100]` |
| Summary only | 5120 (5KB) | Use stats_only=true |

**Examples**:

**Small query**:
```json
{
  "name": "query_tools",
  "arguments": {
    "limit": 50,
    "max_output_bytes": 10240
  }
}
```

**Large dataset - use jq to limit**:
```json
{
  "name": "extract_tools",
  "arguments": {
    "limit": 1000,
    "jq_filter": ".[0:100]",
    "max_output_bytes": 51200
  }
}
```

**Prevent overflow**:
```json
{
  "name": "query_files",
  "arguments": {
    "top": 100,
    "jq_filter": ".[] | {file: .FilePath, ops: .TotalOps}",
    "max_output_bytes": 20480
  }
}
```

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

### 3. Set Appropriate max_output_bytes

**For small queries**:
```json
{"max_output_bytes": 10240}  // 10KB
```

**For medium queries**:
```json
{"max_output_bytes": 51200}  // 50KB (default)
```

**For large queries** - use jq to limit:
```json
{
  "jq_filter": ".[0:100]",
  "max_output_bytes": 51200
}
```

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
  "name": "extract_tools",
  "arguments": {
    "limit": 100,
    "output_format": "jsonl"
  }
}
```

**TSV** (86% smaller):
```json
{
  "name": "extract_tools",
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

### Q: Output is truncated - what should I do?

**A**: Try these solutions:

1. Increase `max_output_bytes`
2. Use `jq_filter` to project fewer fields
3. Use `stats_only=true` for summaries
4. Use `limit` to reduce result count
5. Use `output_format: "tsv"` for compact output

**Example**:
```json
{
  "name": "query_tools",
  "arguments": {
    "jq_filter": ".[] | {tool: .ToolName, status: .Status}",
    "max_output_bytes": 102400,
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

## Reference Resources

- [jq Manual](https://stedolan.github.io/jq/manual/) - Complete jq syntax reference
- [jqplay.org](https://jqplay.org) - Interactive jq playground
- [meta-cc README](../README.md) - Project overview
- [MCP Migration Guide](./mcp-migration-phase15.md) - Phase 15 migration
- [Integration Guide](./integration-guide.md) - MCP vs Slash Commands vs Subagent
- [MCP Protocol](https://modelcontextprotocol.io) - MCP specification

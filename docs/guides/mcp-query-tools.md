# MCP Query Tools Reference

Complete reference for meta-cc's 20 MCP query tools with unified query interface, jq filtering, and hybrid output mode.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Core Query Tools](#core-query-tools)
- [Convenience Tools](#convenience-tools)
- [Legacy Query Tools](#legacy-query-tools)
- [Utility Tools](#utility-tools)
- [Standard Parameters](#standard-parameters)
- [jq Syntax Quick Reference](#jq-syntax-quick-reference)
- [Hybrid Output Mode](#hybrid-output-mode)
- [Best Practices](#best-practices)

---

## Overview

The MCP v2.0 query interface provides a unified, composable approach to querying Claude Code session data:

**Tool Categories** (20 tools total):
- **Core Query Tools** (2): `query`, `query_raw` - Unified interface with jq filtering
- **Convenience Tools** (8): High-frequency queries with optimized defaults
- **Legacy Query Tools** (7): Backward-compatible specialized tools
- **Utility Tools** (3): Session management and capability browsing

**Key Features**:
- **Unified Interface**: Single `query` tool replaces 6 specialized tools
- **jq Integration**: Native jq filtering for maximum flexibility
- **Hybrid Output**: Automatic inline (<8KB) or file_ref (≥8KB) output
- **No Limits by Default**: Returns all results, relies on hybrid mode
- **Scope Support**: Query current session or entire project
- **Standard Parameters**: Consistent interface across all tools

---

## Architecture

### Three-Layer Design

```
Layer 1: Convenience Tools (8)
  ├── query_tool_errors       - Tool execution errors
  ├── query_token_usage       - Assistant messages with tokens
  ├── query_conversation_flow - User/assistant conversation
  ├── query_system_errors     - API system errors
  ├── query_file_snapshots    - File history snapshots
  ├── query_timestamps        - Timestamped entries
  ├── query_summaries         - Session summaries
  └── query_tool_blocks       - Tool use/result blocks

Layer 2: Core Query Tools (2)
  ├── query                   - Unified interface with jq
  └── query_raw               - Raw jq for power users

Layer 3: Legacy Tools (7)
  ├── query_tools             - Tool call history
  ├── query_user_messages     - User message search
  ├── query_tool_sequences    - Tool usage patterns
  ├── query_file_access       - File operation history
  ├── query_project_state     - Project state evolution
  ├── query_successful_prompts - Quality prompt patterns
  └── get_session_stats       - Session statistics
```

### Resource Types

All query tools operate on three core resources:

1. **`entries`** (Raw JSONL): Complete session history
2. **`messages`** (User/Assistant): Parsed conversation messages
3. **`tools`** (Tool Executions): Tool call history with status

---

## Core Query Tools

### `query` - Unified Query Interface

**Description**: Unified query interface for session data with jq filtering.

**Scope**: `project` (default) or `session`

**Parameters**:
- `resource` (string): Resource type - `entries`, `messages`, or `tools` (default: `entries`)
- `filter` (object): Filter conditions (tool_name, tool_status, role, content_match, etc.)
- `jq_filter` (string): jq expression for advanced filtering (default: `.[]`)
- `scope` (string): `project` or `session` (default: `project`)
- `stats_only` (boolean): Return only statistics (default: false)
- `stats_first` (boolean): Return stats first, then details (default: false)
- `inline_threshold_bytes` (number): Threshold for inline vs file_ref (default: 8192)
- `output_format` (string): Output format - `jsonl` or `tsv` (default: `jsonl`)

**Examples**:

```javascript
// Query tool errors
query({
  resource: "tools",
  filter: {
    tool_status: "error"
  }
})

// Query user messages with jq
query({
  resource: "messages",
  filter: {
    role: "user"
  },
  jq_filter: '.[] | select(.content | test("error|bug"; "i"))'
})

// Get statistics only
query({
  resource: "tools",
  stats_only: true
})
```

**Output Schema** (depends on resource):

```typescript
// For resource: "tools"
{
  tool_name: string,      // Tool identifier
  status: string,         // "success" or "error"
  timestamp: string,      // ISO8601 timestamp
  error?: string,         // Error message if failed
  input: object,          // Tool input parameters
  output: object,         // Tool output/result
  uuid: string            // Unique call identifier
}

// For resource: "messages"
{
  turn: number,           // Turn sequence number
  role: string,           // "user" or "assistant"
  timestamp: string,      // ISO8601 timestamp
  content: string,        // Message content
  uuid: string            // Unique message identifier
}

// For resource: "entries"
{
  type: string,           // Entry type
  timestamp: string,      // ISO8601 timestamp
  message: object,        // Full entry data
  uuid: string,           // Unique entry identifier
  sessionId: string,      // Session identifier
  gitBranch: string       // Git branch name
}
```

---

### `query_raw` - Raw jq Query

**Description**: Execute raw jq expression for maximum flexibility. For power users.

**Scope**: `project` (default) or `session`

**Parameters**:
- `jq_expression` (string, **required**): Complete jq expression
- `limit` (number): Max results (no limit by default)
- `scope` (string): `project` or `session` (default: `project`)
- `stats_only` (boolean): Return only statistics (default: false)
- `stats_first` (boolean): Return stats first, then details (default: false)
- `inline_threshold_bytes` (number): Threshold for inline vs file_ref (default: 8192)
- `output_format` (string): Output format - `jsonl` or `tsv` (default: `jsonl`)

**Examples**:

```javascript
// Complex aggregation
query_raw({
  jq_expression: '.[] | select(.type == "assistant") | .message.content.tool_use | select(. != null) | group_by(.name) | map({tool: .[0].name, count: length})'
})

// Time-based analysis
query_raw({
  jq_expression: '.[] | select(.timestamp) | {timestamp, type} | select(.timestamp | fromdateiso8601 > (now - 3600))'
})

// Custom projection
query_raw({
  jq_expression: '.[] | {time: .timestamp, tool: .message.content.tool_use.name, status: .message.status} | select(.tool != null)'
})
```

**Best Practices**:
- Use for one-off complex queries
- Prefer `query` with `jq_filter` for reusable queries
- Test jq expressions with `jq` CLI first
- Handle null values with `select(. != null)`

---

## Convenience Tools

High-frequency queries optimized for common use cases. Each tool includes jq filtering support via standard parameters.

### `query_tool_errors` - Tool Execution Errors

**Description**: Query tool execution errors across all sessions.

**Scope**: `project` (default) or `session`

**Parameters**:
- `limit` (number): Max results (no limit by default)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Default Query**: All tools with `status == "error"`

**Example**:
```javascript
query_tool_errors({
  jq_filter: '.[] | select(.tool_name == "Bash")',
  limit: 10
})
```

**Output Schema**: Same as `query` with `resource: "tools"`

---

### `query_token_usage` - Assistant Message Token Usage

**Description**: Query assistant messages with token usage statistics.

**Scope**: `project` (default) or `session`

**Parameters**:
- `limit` (number): Max results (no limit by default)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Default Query**: All assistant messages with token stats

**Example**:
```javascript
query_token_usage({
  jq_filter: '.[] | select(.usage.input_tokens > 10000)',
  stats_first: true
})
```

**Output Schema**:
```typescript
{
  turn: number,
  timestamp: string,
  content: string,
  usage: {
    input_tokens: number,
    output_tokens: number,
    cache_creation_input_tokens?: number,
    cache_read_input_tokens?: number
  },
  model: string,
  stop_reason: string
}
```

---

### `query_conversation_flow` - Conversation Flow

**Description**: Query user and assistant conversation with parent-child relationships.

**Scope**: `project` (default) or `session`

**Parameters**:
- `limit` (number): Max results (no limit by default)
- `transform` (string): Optional jq transform for parent-child relationships
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Default Query**: All user/assistant messages chronologically ordered

**Example**:
```javascript
query_conversation_flow({
  jq_filter: '.[] | select(.role == "user") | select(.content | test("error"))',
  limit: 20
})
```

**Output Schema**:
```typescript
{
  turn: number,
  role: string,           // "user" or "assistant"
  timestamp: string,
  content: string,
  parent_uuid?: string,   // Parent message UUID
  uuid: string
}
```

---

### `query_system_errors` - API System Errors

**Description**: Query system API errors (rate limits, timeouts, server errors).

**Scope**: `project` (default) or `session`

**Parameters**:
- `limit` (number): Max results (no limit by default)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Default Query**: All entries with `type == "error"` and `error.type == "system"`

**Example**:
```javascript
query_system_errors({
  jq_filter: '.[] | select(.error.code == 529)',
  stats_only: true
})
```

**Output Schema**:
```typescript
{
  timestamp: string,
  error: {
    type: string,         // "system"
    code: number,         // HTTP status code
    message: string
  },
  uuid: string
}
```

---

### `query_file_snapshots` - File History Snapshots

**Description**: Query file content history snapshots across sessions.

**Scope**: `project` (default) or `session`

**Parameters**:
- `limit` (number): Max results (no limit by default)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Default Query**: All entries with `type == "file_snapshot"`

**Example**:
```javascript
query_file_snapshots({
  jq_filter: '.[] | select(.file_path | contains("README.md"))'
})
```

**Output Schema**:
```typescript
{
  timestamp: string,
  file_path: string,
  content: string,
  operation: string,      // "read", "write", "edit"
  uuid: string
}
```

---

### `query_timestamps` - Timestamped Entries

**Description**: Query all entries with timestamps for timeline analysis.

**Scope**: `project` (default) or `session`

**Parameters**:
- `limit` (number): Max results (no limit by default)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Default Query**: All entries with timestamp field

**Example**:
```javascript
query_timestamps({
  jq_filter: '.[] | select(.timestamp | fromdateiso8601 > (now - 3600))',
  output_format: "tsv"
})
```

**Output Schema**:
```typescript
{
  timestamp: string,      // ISO8601 timestamp
  type: string,           // Entry type
  uuid: string
}
```

---

### `query_summaries` - Session Summaries

**Description**: Query session summaries with optional keyword search.

**Scope**: `project` (default) or `session`

**Parameters**:
- `keyword` (string): Keyword to search in summary (case-insensitive)
- `limit` (number): Max results (no limit by default)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Default Query**: All session summary entries

**Example**:
```javascript
query_summaries({
  keyword: "error",
  limit: 5
})
```

**Output Schema**:
```typescript
{
  timestamp: string,
  summary: string,
  session_id: string,
  uuid: string
}
```

---

### `query_tool_blocks` - Tool Use/Result Blocks

**Description**: Query tool_use or tool_result content blocks from assistant messages.

**Scope**: `project` (default) or `session`

**Parameters**:
- `block_type` (string, **required**): `tool_use` or `tool_result`
- `limit` (number): Max results (no limit by default)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Example**:
```javascript
query_tool_blocks({
  block_type: "tool_use",
  jq_filter: '.[] | select(.name == "Bash")'
})
```

**Output Schema**:
```typescript
// For block_type: "tool_use"
{
  id: string,
  type: "tool_use",
  name: string,           // Tool name
  input: object,          // Tool input
  turn: number,
  timestamp: string
}

// For block_type: "tool_result"
{
  tool_use_id: string,
  type: "tool_result",
  content: string,
  is_error: boolean,
  turn: number,
  timestamp: string
}
```

---

## Legacy Query Tools

Backward-compatible specialized tools from v1.x. Consider using `query` or convenience tools for new workflows.

### `get_session_stats` - Session Statistics

**Description**: Get comprehensive session statistics.

**Scope**: `session` (default)

**Parameters**: None

**Example**:
```javascript
get_session_stats()
```

**Output Schema**:
```typescript
{
  session_id: string,
  total_entries: number,
  total_turns: number,
  user_messages: number,
  assistant_messages: number,
  tool_calls: number,
  tool_errors: number,
  duration_minutes: number,
  start_time: string,
  end_time: string
}
```

---

### `query_tools` - Tool Call History

**Description**: Query assistant's internal tool calls (large output, not for user analysis).

**Scope**: `project` (default) or `session`

**Parameters**:
- `tool` (string): Filter by tool name
- `status` (string): Filter by status (`error` or `success`)
- `limit` (number): Max results (no limit by default)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Example**:
```javascript
query_tools({
  tool: "Bash",
  status: "error",
  limit: 10
})
```

**Output Schema**: Same as `query` with `resource: "tools"`

---

### `query_user_messages` - User Message Search

**Description**: Search user messages with regex pattern (may contain large outputs).

**Scope**: `project` (default) or `session`

**Parameters**:
- `pattern` (string, **required**): Regex pattern to match
- `max_message_length` (number): Max chars per message (default: 0 = no truncation)
- `limit` (number): Max results (no limit by default)
- `content_summary` (boolean): Return only turn/timestamp/preview (100 chars)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Example**:
```javascript
query_user_messages({
  pattern: "error|bug",
  max_message_length: 500,
  limit: 20
})
```

**Output Schema**:
```typescript
{
  turn: number,
  timestamp: string,
  content: string         // Truncated if max_message_length specified
}
```

---

### `query_tool_sequences` - Tool Usage Patterns

**Description**: Query assistant's tool sequences (large output, not for user analysis).

**Scope**: `project` (default) or `session`

**Parameters**:
- `pattern` (string): Sequence pattern to match
- `include_builtin_tools` (boolean): Include built-in tools (default: false, 35x faster)
- `min_occurrences` (number): Min occurrences (default: 3)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Example**:
```javascript
query_tool_sequences({
  min_occurrences: 5,
  jq_filter: '.[] | select(.count > 10)'
})
```

**Output Schema**:
```typescript
{
  pattern: string,        // Tool sequence pattern
  count: number,          // Number of occurrences
  occurrences: Array<{
    start_turn: number,
    end_turn: number
  }>,
  time_span_minutes: number,
  length?: number,
  success_rate?: number,
  avg_duration_minutes?: number
}
```

---

### `query_file_access` - File Operation History

**Description**: Query file operation history for a specific file.

**Scope**: `project` (default) or `session`

**Parameters**:
- `file` (string, **required**): File path
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Example**:
```javascript
query_file_access({
  file: "README.md"
})
```

**Output Schema**:
```typescript
{
  file: string,
  total_accesses: number,
  operations: {
    Read?: number,
    Edit?: number,
    Write?: number
  },
  timeline: Array<{
    turn: number,
    timestamp: string,
    operation: string
  }>,
  time_span_minutes: number
}
```

---

### `query_project_state` - Project State Evolution

**Description**: Query project state evolution over time.

**Scope**: `project` (default) or `session`

**Parameters**:
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Example**:
```javascript
query_project_state({
  jq_filter: '.[] | select(.type == "session_state")'
})
```

**Output Schema**:
```typescript
{
  timestamp: string,
  type: string,           // "session_state", etc.
  state: object           // State data
}
```

---

### `query_successful_prompts` - Quality Prompt Patterns

**Description**: Query successful prompt patterns with quality scores.

**Scope**: `project` (default) or `session`

**Parameters**:
- `min_quality_score` (number): Min quality score (default: 0.8)
- `limit` (number): Max results (no limit by default)
- Standard parameters (jq_filter, scope, stats_only, etc.)

**Example**:
```javascript
query_successful_prompts({
  min_quality_score: 0.9,
  limit: 10
})
```

**Output Schema**:
```typescript
{
  turn: number,
  content: string,
  quality_score: number   // 0.0 - 1.0
}
```

---

## Utility Tools

### `cleanup_temp_files` - Temp File Cleanup

**Description**: Remove old temporary MCP files (file_ref outputs).

**Scope**: None

**Parameters**:
- `max_age_days` (number): Max file age in days (default: 7)

**Example**:
```javascript
cleanup_temp_files({
  max_age_days: 30
})
```

---

### `list_capabilities` - List Available Capabilities

**Description**: List all available capabilities from configured sources.

**Scope**: None

**Parameters**: None

**Example**:
```javascript
list_capabilities()
```

**Output Schema**:
```typescript
{
  capabilities: Array<{
    name: string,
    description: string,
    source: string
  }>
}
```

---

### `get_capability` - Get Capability Content

**Description**: Retrieve complete capability content by name.

**Scope**: None

**Parameters**:
- `name` (string, **required**): Capability name (without .md extension)

**Example**:
```javascript
get_capability({
  name: "meta-errors"
})
```

**Output**: Markdown content of the capability

---

## Standard Parameters

All query tools support these standard parameters:

### `scope` (string)
Query scope: `project` (default) or `session`

- **`project`**: Query all sessions in project
- **`session`**: Query current session only

**Example**:
```javascript
query({
  resource: "tools",
  scope: "session"
})
```

### `jq_filter` (string)
jq expression for filtering. Defaults to `.[]` when omitted.

**IMPORTANT**: Do NOT wrap in quotes - use raw jq expression.

**Examples**:
```javascript
// Correct
jq_filter: '.[] | select(.tool_name == "Bash")'

// Wrong
jq_filter: '".[] | select(.tool_name == \"Bash\")"'
```

### `stats_only` (boolean)
Return only statistics (default: false)

**Example**:
```javascript
query({
  resource: "tools",
  stats_only: true
})
```

**Output**:
```json
{
  "total_count": 1234,
  "unique_tools": 15,
  "error_rate": 0.05
}
```

### `stats_first` (boolean)
Return stats first, then details (default: false)

**Example**:
```javascript
query({
  resource: "tools",
  stats_first: true
})
```

### `inline_threshold_bytes` (number)
Threshold for inline vs file_ref mode in bytes (default: 8192)

**How it works**:
- Results < threshold: Return inline in MCP response
- Results ≥ threshold: Save to temp file, return file_ref

**Common values**:
- `1024` (1KB): Force file_ref for most queries
- `8192` (8KB): Default, balanced for typical queries
- `65536` (64KB): Allow large inline results

**Example**:
```javascript
query({
  resource: "tools",
  inline_threshold_bytes: 1024  // Force file_ref
})
```

Can also set `META_CC_INLINE_THRESHOLD` environment variable.

### `output_format` (string)
Output format: `jsonl` or `tsv` (default: `jsonl`)

**Example**:
```javascript
query({
  resource: "tools",
  output_format: "tsv"
})
```

---

## jq Syntax Quick Reference

### Basic Selection

```javascript
// Select all entries
.[]

// Select specific field
.[] | .tool_name

// Select multiple fields
.[] | {tool: .tool_name, status: .status}
```

### Filtering

```javascript
// Exact match
.[] | select(.tool_name == "Bash")

// Multiple conditions (AND)
.[] | select(.tool_name == "Bash" and .status == "error")

// Multiple conditions (OR)
.[] | select(.tool_name == "Bash" or .tool_name == "Read")

// Regex match (case-sensitive)
.[] | select(.content | test("error"))

// Regex match (case-insensitive)
.[] | select(.content | test("error|bug"; "i"))

// Null check
.[] | select(.error != null)

// Contains check
.[] | select(.content | contains("timeout"))
```

### Time-based Filtering

```javascript
// Last hour (Unix timestamp)
.[] | select(.timestamp | fromdateiso8601 > (now - 3600))

// Last 24 hours
.[] | select(.timestamp | fromdateiso8601 > (now - 86400))

// Date range
.[] | select(.timestamp >= "2025-10-01" and .timestamp <= "2025-10-31")
```

### Aggregation

```javascript
// Count entries
.[] | length

// Group by field
group_by(.tool_name) | map({tool: .[0].tool_name, count: length})

// Sum field
[.[] | .count] | add

// Average field
[.[] | .count] | add / length

// Max/min field
[.[] | .count] | max
[.[] | .count] | min
```

### Sorting

```javascript
// Sort ascending
sort_by(.timestamp)

// Sort descending
sort_by(.timestamp) | reverse

// Get last N items
sort_by(.timestamp) | .[-10:]

// Get first N items
sort_by(.timestamp) | .[0:10]
```

### Transformation

```javascript
// Extract nested field
.[] | .message.content.tool_use.name

// Flatten array
.[] | .occurrences[]

// Map array
[.[] | {tool: .tool_name, status: .status}]

// Unique values
[.[] | .tool_name] | unique
```

---

## Hybrid Output Mode

The MCP v2.0 query interface uses hybrid output mode to handle large result sets:

### How It Works

1. **Query Execution**: Tool executes query and collects results
2. **Size Check**: If result size < `inline_threshold_bytes` (default: 8192)
   - **Inline Mode**: Return results directly in MCP response
3. **Size Check**: If result size ≥ `inline_threshold_bytes`
   - **File Ref Mode**: Save results to temp file, return file_ref

### File Ref Format

```json
{
  "content": [
    {
      "type": "text",
      "text": "Query returned 1234 results (245.6 KB). Output saved to file_ref (too large for inline)."
    },
    {
      "type": "resource",
      "resource": {
        "uri": "file:///tmp/mcp-output-abc123.jsonl",
        "mimeType": "application/x-ndjson",
        "text": "... full content ..."
      }
    }
  ]
}
```

### Reading File Refs

Claude Code automatically reads file refs using the Read tool:

```javascript
// 1. Query returns file_ref
query({resource: "tools"})

// 2. Read file_ref
Read({file_path: "/tmp/mcp-output-abc123.jsonl"})
```

### Best Practices

**Do**:
- ✅ Let hybrid mode handle large results automatically
- ✅ Use `limit` only when user explicitly requests specific count
- ✅ Set `inline_threshold_bytes` based on use case
- ✅ Clean up old temp files with `cleanup_temp_files()`

**Don't**:
- ❌ Add `limit` to every query "just in case"
- ❌ Manually implement pagination
- ❌ Assume inline results are always complete
- ❌ Ignore file_ref outputs

---

## Best Practices

### 1. Choose the Right Tool

**Use Convenience Tools** for common queries:
```javascript
// Good: Simple error query
query_tool_errors({limit: 10})

// Overkill: Using raw jq
query_raw({
  jq_expression: '.[] | select(.type == "assistant") | .message.content.tool_result | select(.is_error == true)'
})
```

**Use Core Query Tools** for complex queries:
```javascript
// Good: Complex filtering
query({
  resource: "tools",
  jq_filter: '.[] | select(.tool_name == "Bash") | select(.error | test("not found"))'
})
```

### 2. Use jq Effectively

**Start simple, add complexity**:
```javascript
// Step 1: Get all tools
query({resource: "tools"})

// Step 2: Filter by name
query({
  resource: "tools",
  jq_filter: '.[] | select(.tool_name == "Bash")'
})

// Step 3: Add error filtering
query({
  resource: "tools",
  jq_filter: '.[] | select(.tool_name == "Bash" and .status == "error")'
})
```

**Test jq expressions with CLI**:
```bash
# Test jq expression locally
echo '[{"tool_name":"Bash","status":"error"}]' | jq '.[] | select(.status == "error")'
```

### 3. Handle Large Results

**Let hybrid mode work**:
```javascript
// Good: No limit, hybrid mode handles size
query({resource: "tools"})

// Bad: Arbitrary limit
query({resource: "tools", limit: 100})
```

**Adjust threshold for use case**:
```javascript
// Interactive exploration: Small inline threshold
query({
  resource: "tools",
  inline_threshold_bytes: 1024  // 1KB
})

// Batch processing: Large inline threshold
query({
  resource: "tools",
  inline_threshold_bytes: 65536  // 64KB
})
```

### 4. Use Scope Appropriately

**Session scope** for debugging current work:
```javascript
query_tool_errors({
  scope: "session"
})
```

**Project scope** for trend analysis:
```javascript
query_tool_errors({
  scope: "project",
  jq_filter: '.[] | group_by(.tool_name) | map({tool: .[0].tool_name, count: length})'
})
```

### 5. Optimize Performance

**Use stats_only for counts**:
```javascript
// Fast: Only statistics
query({
  resource: "tools",
  stats_only: true
})

// Slow: Load all data for count
query({
  resource: "tools",
  jq_filter: '.[] | length'
})
```

**Exclude built-in tools** in sequence queries:
```javascript
// Fast: Only MCP tools (35x faster)
query_tool_sequences({
  include_builtin_tools: false
})

// Slow: Include Bash, Read, Edit, etc.
query_tool_sequences({
  include_builtin_tools: true
})
```

### 6. Compose Queries

**Pipeline multiple queries**:
```javascript
// 1. Get error summary
query({
  resource: "tools",
  filter: {tool_status: "error"},
  stats_only: true
})

// 2. Get specific error details
query_tool_errors({
  jq_filter: '.[] | select(.tool_name == "Bash")',
  limit: 5
})

// 3. Analyze error patterns
query({
  resource: "tools",
  filter: {tool_status: "error"},
  jq_filter: '.[] | group_by(.error | split(":")[0]) | map({error_type: .[0].error | split(":")[0], count: length})'
})
```

---

## See Also

- [MCP Query Cookbook](../examples/mcp-query-cookbook.md) - 20+ practical examples
- [MCP v2.0 Migration Guide](mcp-v2-migration.md) - Migrating from v1.x
- [Unified Query API Guide](unified-query-api.md) - Query architecture
- [MCP Guide](mcp.md) - Complete MCP reference
- [JSONL Schema Reference](../reference/jsonl-schema.md) - Session data format

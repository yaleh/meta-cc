# MCP v2.0 Migration Guide

This guide helps you migrate from MCP v1.x specialized tools to the unified v2.0 query interface.

## Table of Contents

- [Breaking Changes](#breaking-changes)
- [Migration Overview](#migration-overview)
- [Tool Migration Table](#tool-migration-table)
- [Migration Examples](#migration-examples)
- [Common Patterns](#common-patterns)
- [FAQ](#faq)

## Breaking Changes

### Removed Tools (Phase 25.4)

The following 6 specialized tools have been removed in favor of the unified `query` tool with jq filtering:

1. **`query_context`** - Use `query` with custom jq instead
2. **`query_tools_advanced`** - Use `query` with jq filtering
3. **`query_time_series`** - Use `query` + jq grouping instead
4. **`query_assistant_messages`** - Use `query_token_usage` or `query` with jq
5. **`query_conversation`** - Use `query_conversation_flow` instead
6. **`query_files`** - Use `query_file_snapshots` or jq filtering instead

### New Architecture

**Before (v1.x)**: 23 specialized tools, each with custom parameters
**After (v2.0)**: 20 tools with unified query interface

```
Core Tools:
  - query (unified interface)
  - query_raw (raw jq for power users)

Convenience Tools (8):
  - query_tool_errors
  - query_token_usage
  - query_conversation_flow
  - query_system_errors
  - query_file_snapshots
  - query_timestamps
  - query_summaries
  - query_tool_blocks

Legacy Tools (7):
  - query_tools
  - query_user_messages
  - query_tool_sequences
  - query_file_access
  - query_project_state
  - query_successful_prompts
  - get_session_stats

Utility Tools (3):
  - cleanup_temp_files
  - list_capabilities
  - get_capability
```

## Migration Overview

### Strategy 1: Use Convenience Tools (Recommended)

For common queries, use the new convenience tools:

```javascript
// OLD: query_assistant_messages({pattern: "error"})
// NEW: Use query_token_usage + jq filtering
query_token_usage({
  jq_filter: '.[] | select(.content | test("error"; "i"))'
})
```

### Strategy 2: Use Unified Query Interface

For complex queries, use the `query` tool:

```javascript
// OLD: query_context({error_signature: "timeout"})
// NEW: Use query with jq filtering
query({
  resource: "entries",
  jq_filter: '.[] | select(.type == "error") | select(.message | contains("timeout"))'
})
```

### Strategy 3: Use Raw jq (Power Users)

For maximum flexibility:

```javascript
query_raw({
  jq_expression: '.[] | select(.type == "assistant") | .message.content'
})
```

## Tool Migration Table

### 1. query_context → query + jq

**OLD**:
```javascript
query_context({
  error_signature: "file_not_found",
  window: 3
})
```

**NEW**:
```javascript
query({
  resource: "entries",
  jq_filter: `
    .[] |
    select(.type == "error") |
    select(.message | contains("file_not_found"))
  `
})
```

### 2. query_tools_advanced → query + jq

**OLD**:
```javascript
query_tools_advanced({
  where: "tool = \"Bash\" AND status = \"error\""
})
```

**NEW**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Bash",
    tool_status: "error"
  }
})
```

Or use the convenience tool:
```javascript
query_tool_errors({
  jq_filter: '.[] | select(.tool_name == "Bash")'
})
```

### 3. query_time_series → query + jq grouping

**OLD**:
```javascript
query_time_series({
  interval: "hour",
  metric: "tool-calls"
})
```

**NEW**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    .[] |
    {hour: .timestamp[:13], tool: .tool_name} |
    group_by(.hour) |
    map({hour: .[0].hour, count: length})
  `
})
```

### 4. query_assistant_messages → query_token_usage

**OLD**:
```javascript
query_assistant_messages({
  pattern: "error",
  min_tools: 5,
  limit: 10
})
```

**NEW**:
```javascript
query_token_usage({
  jq_filter: `
    .[] |
    select(.content | test("error"; "i")) |
    select(.tool_use_count >= 5) |
    limit(10; .)
  `
})
```

### 5. query_conversation → query_conversation_flow

**OLD**:
```javascript
query_conversation({
  pattern: "bug fix",
  min_duration: 5000
})
```

**NEW**:
```javascript
query_conversation_flow({
  jq_filter: `
    .[] |
    select(.user_message.content | test("bug fix"; "i")) |
    select(.duration_ms >= 5000)
  `
})
```

### 6. query_files → query_file_snapshots

**OLD**:
```javascript
query_files({
  threshold: 5
})
```

**NEW**:
```javascript
query_file_snapshots({
  jq_filter: '.[] | select(.total_ops >= 5)'
})
```

## Migration Examples

### Example 1: Find Bash Errors

**OLD**:
```javascript
query_tools_advanced({
  where: "tool = \"Bash\" AND status = \"error\"",
  limit: 20
})
```

**NEW (Convenience Tool)**:
```javascript
query_tool_errors({
  jq_filter: '.[] | select(.tool_name == "Bash") | limit(20; .)'
})
```

**NEW (Unified Query)**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Bash",
    tool_status: "error"
  },
  jq_filter: '.[] | limit(20; .)'
})
```

### Example 2: Analyze Token Usage

**OLD**:
```javascript
query_assistant_messages({
  min_tokens_output: 1000,
  limit: 10
})
```

**NEW**:
```javascript
query_token_usage({
  jq_filter: '.[] | select(.tokens_output >= 1000) | limit(10; .)'
})
```

### Example 3: Time-Based Grouping

**OLD**:
```javascript
query_time_series({
  interval: "day",
  metric: "tool-calls"
})
```

**NEW**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    .[] |
    {day: .timestamp[:10], tool: .tool_name} |
    group_by(.day) |
    map({day: .[0].day, count: length, tools: map(.tool) | unique})
  `
})
```

### Example 4: Conversation Flow Analysis

**OLD**:
```javascript
query_conversation({
  pattern_target: "assistant",
  pattern: "completed",
  min_duration: 10000
})
```

**NEW**:
```javascript
query_conversation_flow({
  jq_filter: `
    .[] |
    select(.assistant_message.content | test("completed"; "i")) |
    select(.duration_ms >= 10000)
  `
})
```

### Example 5: File Operation Stats

**OLD**:
```javascript
query_files({
  threshold: 10,
  jq_filter: '.[] | select(.file_path | test("\\.go$"))'
})
```

**NEW**:
```javascript
query_file_snapshots({
  jq_filter: `
    .[] |
    select(.file_path | test("\\\\.go$")) |
    select(.total_ops >= 10)
  `
})
```

### Example 6: Error Context

**OLD**:
```javascript
query_context({
  error_signature: "timeout",
  window: 5
})
```

**NEW**:
```javascript
query({
  resource: "entries",
  jq_filter: `
    .[] |
    select(.type == "error") |
    select(.message | contains("timeout")) |
    {timestamp, type, message, parent_uuid}
  `
})
```

### Example 7: Multi-Condition Filtering

**OLD**:
```javascript
query_tools_advanced({
  where: "tool IN (\"Read\", \"Write\") AND status = \"error\""
})
```

**NEW**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    .[] |
    select(.tool_name == "Read" or .tool_name == "Write") |
    select(.status == "error")
  `
})
```

### Example 8: Assistant Messages with Tool Count

**OLD**:
```javascript
query_assistant_messages({
  min_tools: 5,
  max_tools: 10,
  min_length: 100
})
```

**NEW**:
```javascript
query_token_usage({
  jq_filter: `
    .[] |
    select(.tool_use_count >= 5 and .tool_use_count <= 10) |
    select(.text_length >= 100)
  `
})
```

### Example 9: Hourly Activity

**OLD**:
```javascript
query_time_series({
  interval: "hour",
  where: "tool = \"Bash\""
})
```

**NEW**:
```javascript
query({
  resource: "tools",
  filter: {tool_name: "Bash"},
  jq_filter: `
    .[] |
    {hour: .timestamp[:13]} |
    group_by(.hour) |
    map({hour: .[0].hour, count: length})
  `
})
```

### Example 10: Long Conversations

**OLD**:
```javascript
query_conversation({
  min_duration: 30000,
  limit: 5
})
```

**NEW**:
```javascript
query_conversation_flow({
  jq_filter: `
    .[] |
    select(.duration_ms >= 30000) |
    limit(5; .)
  `
})
```

### Example 11: File Error Analysis

**OLD**:
```javascript
query_files({
  jq_filter: '.[] | select(.error_rate > 0.1)'
})
```

**NEW**:
```javascript
query_file_snapshots({
  jq_filter: '.[] | select(.error_rate > 0.1)'
})
```

### Example 12: Complex Aggregation

**OLD**:
```javascript
query_tools_advanced({
  where: "status = \"error\"",
  jq_filter: '.[] | group_by(.tool_name) | map({tool: .[0].tool_name, errors: length})'
})
```

**NEW**:
```javascript
query({
  resource: "tools",
  filter: {tool_status: "error"},
  jq_filter: `
    .[] |
    group_by(.tool_name) |
    map({tool: .[0].tool_name, errors: length, error_rate: (length / 100)})
  `
})
```

### Example 13: User Message Search

**OLD**:
```javascript
query_user_messages({
  pattern: "fix.*bug",
  content_summary: true
})
```

**NEW**:
```javascript
query_user_messages({
  pattern: "fix.*bug",
  jq_filter: '.[] | {timestamp, preview: .content[:100]}'
})
```

### Example 14: Token Usage Statistics

**OLD**:
```javascript
query_assistant_messages({
  min_tokens_output: 500,
  jq_filter: '.[] | {turn_sequence, tokens_output, tool_use_count}'
})
```

**NEW**:
```javascript
query_token_usage({
  jq_filter: `
    .[] |
    select(.tokens_output >= 500) |
    {turn_sequence, tokens_output, tool_use_count, efficiency: (.tokens_output / .tool_use_count)}
  `
})
```

### Example 15: Recent Errors

**OLD**:
```javascript
query_context({
  error_signature: ".*",
  window: 10
})
```

**NEW**:
```javascript
query_system_errors({
  jq_filter: '.[] | limit(10; .)'
})
```

### Example 16: File Access Patterns

**OLD**:
```javascript
query_files({
  threshold: 3,
  jq_filter: '.[] | select(.read_count > .write_count * 2)'
})
```

**NEW**:
```javascript
query_file_snapshots({
  jq_filter: `
    .[] |
    select(.total_ops >= 3) |
    select(.read_count > (.write_count * 2)) |
    {file_path, read_count, write_count, ratio: (.read_count / .write_count)}
  `
})
```

### Example 17: Timeline View

**OLD**:
```javascript
query_time_series({
  interval: "day"
})
```

**NEW**:
```javascript
query_timestamps({
  jq_filter: `
    .[] |
    {day: .timestamp[:10], type: .type} |
    group_by(.day) |
    map({day: .[0].day, events: length, types: map(.type) | unique})
  `
})
```

### Example 18: Conversation Quality

**OLD**:
```javascript
query_conversation({
  pattern: "success",
  min_duration: 5000,
  max_duration: 30000
})
```

**NEW**:
```javascript
query_conversation_flow({
  jq_filter: `
    .[] |
    select(.assistant_message.content | test("success"; "i")) |
    select(.duration_ms >= 5000 and .duration_ms <= 30000) |
    {turn, duration_ms, quality_indicator: "good"}
  `
})
```

### Example 19: Tool Usage Summary

**OLD**:
```javascript
query_tools_advanced({
  where: "1=1",
  jq_filter: '.[] | group_by(.tool_name) | map({tool: .[0].tool_name, total: length, errors: map(select(.status == "error")) | length})'
})
```

**NEW**:
```javascript
query({
  resource: "tools",
  jq_filter: `
    .[] |
    group_by(.tool_name) |
    map({
      tool: .[0].tool_name,
      total: length,
      errors: map(select(.status == "error")) | length,
      success_rate: (map(select(.status == "success")) | length) / length
    })
  `
})
```

### Example 20: Session Summary

**OLD**:
```javascript
query_assistant_messages({
  limit: 1,
  jq_filter: '.[] | {total_tokens: .tokens_output, total_tools: .tool_use_count}'
})
```

**NEW**:
```javascript
get_session_stats({
  stats_only: true
})
```

Or for detailed analysis:
```javascript
query_token_usage({
  jq_filter: `
    [.] |
    {
      total_messages: length,
      total_tokens: map(.tokens_output) | add,
      total_tools: map(.tool_use_count) | add,
      avg_tokens_per_message: (map(.tokens_output) | add) / length
    }
  `
})
```

## Common Patterns

### Pattern 1: Filtering by Tool Name

```javascript
// Use filter parameter for simple cases
query({
  resource: "tools",
  filter: {tool_name: "Bash"}
})

// Use jq for complex cases
query({
  resource: "tools",
  jq_filter: '.[] | select(.tool_name == "Bash" or .tool_name == "Read")'
})
```

### Pattern 2: Time-Based Grouping

```javascript
// Group by hour
query({
  resource: "tools",
  jq_filter: `
    .[] |
    {hour: .timestamp[:13], tool: .tool_name} |
    group_by(.hour) |
    map({hour: .[0].hour, count: length})
  `
})

// Group by day
query({
  resource: "entries",
  jq_filter: `
    .[] |
    {day: .timestamp[:10], type: .type} |
    group_by(.day) |
    map({day: .[0].day, count: length})
  `
})
```

### Pattern 3: Error Analysis

```javascript
// Use convenience tool
query_tool_errors({
  jq_filter: '.[] | {tool_name, error, timestamp}'
})

// Or unified query
query({
  resource: "tools",
  filter: {tool_status: "error"},
  jq_filter: '.[] | {tool_name, error, timestamp}'
})
```

### Pattern 4: Statistical Aggregation

```javascript
query({
  resource: "tools",
  jq_filter: `
    .[] |
    group_by(.tool_name) |
    map({
      tool: .[0].tool_name,
      count: length,
      success_rate: (map(select(.status == "success")) | length) / length,
      avg_duration: (map(.duration_ms // 0) | add) / length
    })
  `
})
```

### Pattern 5: Limiting Results

```javascript
// Use jq limit function
query({
  resource: "tools",
  jq_filter: '.[] | limit(10; .)'
})

// Or use limit in filter
query({
  resource: "tools",
  filter: {tool_status: "error"},
  jq_filter: '.[] | limit(20; .)'
})
```

## FAQ

### Q: Why were these tools removed?

A: The specialized tools created fragmentation and inconsistency. The unified `query` interface provides:
- More flexibility with jq filtering
- Consistent parameter structure
- Better composability
- Easier maintenance

### Q: Is there a performance difference?

A: No. Both approaches use the same underlying query engine. The jq filtering happens in-process and is highly efficient.

### Q: Can I still use the old tool names?

A: No, the old tools are completely removed in v2.0. You must migrate to the new interface.

### Q: What if my query is very complex?

A: Use `query_raw` for maximum flexibility:

```javascript
query_raw({
  jq_expression: `
    .[] |
    select(.type == "assistant") |
    {
      turn: .message.turn_sequence,
      tokens: .message.usage.output_tokens,
      tools: .message.content | map(select(.type == "tool_use")) | length
    } |
    select(.tools > 5 and .tokens > 1000)
  `
})
```

### Q: How do I learn jq syntax?

A: See:
- [jq Manual](https://stedolan.github.io/jq/manual/)
- [jq Playground](https://jqplay.org/)
- [Query Cookbook](query-cookbook.md) - 50+ practical examples

### Q: What if I need the exact old behavior?

A: The new convenience tools (`query_token_usage`, `query_conversation_flow`, etc.) provide similar high-level functionality. For exact parity, use the migration examples above.

### Q: Are there any breaking changes in output format?

A: Output formats are consistent across v1.x and v2.0. The main difference is parameter structure, not results.

### Q: How do I test my migration?

A: Compare outputs:

```bash
# v1.x (old session)
query_assistant_messages({"min_tools": 5})

# v2.0 (new session)
query_token_usage({"jq_filter": ".[] | select(.tool_use_count >= 5)"})
```

### Q: What about hybrid output mode?

A: All tools support hybrid output mode (automatic switching between inline and file_ref based on size). This works the same in v1.x and v2.0.

### Q: Can I use both filter and jq_filter?

A: Yes! `filter` provides simple structured filtering, while `jq_filter` provides advanced transformations:

```javascript
query({
  resource: "tools",
  filter: {tool_name: "Bash"},  // Pre-filter
  jq_filter: '.[] | select(.status == "error") | {timestamp, error}'  // Transform
})
```

---

**Need Help?**
- Check [Query Cookbook](query-cookbook.md) for more examples
- See [MCP Guide](mcp.md) for complete API reference
- Join community discussions on GitHub

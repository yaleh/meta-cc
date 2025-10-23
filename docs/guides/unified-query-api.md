# Unified Query API Guide

## Overview

The unified query interface simplifies meta-cc's query capabilities by consolidating **16 specialized MCP tools** into **1 composable `query` tool**. This design provides:

- **Resource-oriented approach**: Query "what" (entries/messages/tools), not "how"
- **Composable pipeline**: filter → transform → aggregate → output
- **Consistent schema**: All output uses snake_case matching JSONL source
- **Backward compatible**: Old tools remain available during migration period

### Design Philosophy

**Before (16 tools)**:
```javascript
// Different tools for different purposes
query_tools({tool: "Read", status: "error"})
query_user_messages({pattern: "fix"})
query_files({threshold: 5})
```

**After (1 unified tool)**:
```javascript
// Single tool with composable parameters
query({
  resource: "tools",
  filter: {tool_name: "Read", tool_status: "error"}
})
```

### Key Benefits

| Benefit | Before | After | Improvement |
|---------|--------|-------|-------------|
| Tool count | 16 tools | 1 tool | **94% reduction** |
| Parameter count | 80+ parameters | 20 core parameters | **75% reduction** |
| Query composition | Not possible | Unlimited | **Infinite flexibility** |
| Learning curve | Learn 16 APIs | Learn 1 API | **Faster adoption** |
| Schema consistency | 3 naming styles | 1 (snake_case) | **100% consistent** |

---

## Quick Start

### Basic Query

Query all failed Read operations:

```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Read",
    tool_status: "error"
  }
})
```

### Aggregation Query

Count tool calls by type:

```javascript
query({
  resource: "tools",
  aggregate: {
    function: "count",
    field: "tool_name"
  }
})
```

### Complex Query

Analyze error patterns by Git branch:

```javascript
query({
  resource: "tools",
  filter: {
    tool_status: "error",
    tool_name: "Read|Edit|Write"
  },
  transform: {
    extract: ["git_branch", "tool_name"],
    group_by: "git_branch"
  },
  aggregate: {
    function: "count",
    field: "tool_name"
  },
  output: {
    sort_by: "count",
    sort_order: "desc",
    limit: 10
  }
})
```

---

## API Reference

### QueryParams

The unified query accepts a single `QueryParams` object with the following structure:

```go
type QueryParams struct {
    // Tier 1: Resource Selection
    Resource string // "entries" | "messages" | "tools"

    // Tier 2: Scope
    Scope string // "session" | "project" (default)

    // Tier 3: Filtering
    Filter FilterSpec

    // Tier 4: Transformation
    Transform TransformSpec

    // Tier 5: Aggregation
    Aggregate AggregateSpec

    // Tier 6: Output Control
    Output OutputSpec

    // Advanced: jq filter (applied last)
    JQFilter string
}
```

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `resource` | string | "entries" | Resource type to query |
| `scope` | string | "project" | Query scope (project/session) |
| `filter` | FilterSpec | {} | Filtering conditions |
| `transform` | TransformSpec | {} | Transform/group/extract |
| `aggregate` | AggregateSpec | {} | Aggregation functions |
| `output` | OutputSpec | {} | Output formatting |
| `jq_filter` | string | ".[]" | jq expression (advanced) |

---

### Resource Types

The query engine supports three resource views:

#### 1. Entries (Raw Data)

**Description**: Raw JSONL session entries (SessionEntry objects)

**Use cases**:
- Low-level data exploration
- Custom processing pipelines
- Debugging session structure

**Example**:
```javascript
query({resource: "entries"})
```

**Output schema**:
```json
{
  "type": "user_message",
  "uuid": "abc-123",
  "timestamp": "2025-10-23T10:00:00Z",
  "session_id": "session-456",
  "parent_uuid": "def-789",
  "cwd": "/home/user/project",
  "git_branch": "main",
  "message": {...}
}
```

#### 2. Messages (User + Assistant)

**Description**: Flattened message view (user and assistant messages)

**Use cases**:
- Conversation analysis
- Prompt pattern mining
- Content search

**Example**:
```javascript
query({
  resource: "messages",
  filter: {role: "user", content_match: "test"}
})
```

**Output schema**:
```json
{
  "uuid": "abc-123",
  "session_id": "session-456",
  "parent_uuid": "def-789",
  "timestamp": "2025-10-23T10:00:00Z",
  "role": "user",
  "content": "Please run the tests",
  "content_blocks": [...]
}
```

#### 3. Tools (Executions)

**Description**: Tool execution view (tool_use + tool_result pairs)

**Use cases**:
- Error analysis
- Tool usage patterns
- Performance monitoring

**Example**:
```javascript
query({
  resource: "tools",
  filter: {tool_name: "Bash", tool_status: "error"}
})
```

**Output schema**:
```json
{
  "tool_use_id": "tool-123",
  "session_id": "session-456",
  "timestamp": "2025-10-23T10:00:00Z",
  "tool_name": "Bash",
  "input": {"command": "npm test"},
  "output": "Error: command failed",
  "status": "error",
  "error": "exit code 1",
  "assistant_uuid": "msg-abc",
  "user_uuid": "msg-def"
}
```

---

### FilterSpec

Filter resources by various criteria:

```go
type FilterSpec struct {
    // Entry-level filters
    Type       string     // Entry type
    SessionID  string     // Session ID
    UUID       string     // Entry UUID
    GitBranch  string     // Git branch
    TimeRange  *TimeRange // Time range

    // Message-level filters
    Role         string // "user" | "assistant"
    ContentType  string // Content block type
    ContentMatch string // Regex pattern for content

    // Tool-level filters
    ToolName   string // Tool name (supports regex)
    ToolStatus string // "success" | "error"
    HasError   bool   // Has error field
}
```

**Examples**:

1. **Basic filter** (single condition):
```javascript
filter: {tool_name: "Read"}
```

2. **Multiple conditions** (AND logic):
```javascript
filter: {
  tool_name: "Bash",
  tool_status: "error"
}
```

3. **Regex pattern**:
```javascript
filter: {
  tool_name: "Read|Edit|Write" // Match any file operation
}
```

4. **Time range**:
```javascript
filter: {
  time_range: {
    start: "2025-10-01T00:00:00Z",
    end: "2025-10-23T23:59:59Z"
  }
}
```

---

### TransformSpec

Transform and extract data:

```go
type TransformSpec struct {
    Extract []string  // JSONPath expressions
    GroupBy string    // Group by field
    Join    *JoinSpec // Join specification
}
```

**Examples**:

1. **Extract fields**:
```javascript
transform: {
  extract: ["tool_name", "status", "input.file_path"]
}
```

2. **Group by field**:
```javascript
transform: {
  group_by: "git_branch"
}
```

---

### AggregateSpec

Apply aggregation functions:

```go
type AggregateSpec struct {
    Function string // "count" | "sum" | "avg" | "min" | "max" | "group"
    Field    string // Field to aggregate on
}
```

**Supported functions**:

| Function | Description | Example |
|----------|-------------|---------|
| `count` | Count records | Total tool calls |
| `sum` | Sum numeric values | Total duration |
| `avg` | Average values | Average response time |
| `min` | Minimum value | Shortest duration |
| `max` | Maximum value | Longest duration |
| `group` | Group by field | Group by tool name |

**Examples**:

1. **Simple count**:
```javascript
aggregate: {function: "count"}
// Output: {"count": 123}
```

2. **Count by field**:
```javascript
aggregate: {
  function: "count",
  field: "tool_name"
}
// Output: [
//   {"tool_name": "Read", "count": 45},
//   {"tool_name": "Bash", "count": 78}
// ]
```

3. **Sum duration**:
```javascript
aggregate: {
  function: "sum",
  field: "duration"
}
```

---

### OutputSpec

Control output format:

```go
type OutputSpec struct {
    Format    string // "jsonl" | "tsv" | "summary"
    Limit     int    // Max results (0 = no limit)
    SortBy    string // Sort field
    SortOrder string // "asc" | "desc"
}
```

**Examples**:

1. **Limit results**:
```javascript
output: {limit: 10}
```

2. **Sort by field**:
```javascript
output: {
  sort_by: "timestamp",
  sort_order: "desc"
}
```

3. **TSV format** (86% smaller than JSONL):
```javascript
output: {format: "tsv"}
```

---

## Query Pipeline

The unified query follows a deterministic pipeline:

```
SessionEntry[] (source)
    ↓
1. SELECT RESOURCE (entries/messages/tools)
    ↓
2. FILTER (where conditions)
    ↓
3. TRANSFORM (extract/group)
    ↓
4. AGGREGATE (count/sum/avg)
    ↓
5. OUTPUT (format/sort/limit)
    ↓
Result
```

### Pipeline Examples

**Example 1: No aggregation**
```javascript
query({
  resource: "tools",
  filter: {tool_name: "Read"},
  output: {limit: 5}
})
// Pipeline: tools → filter → limit → output
// Result: 5 Read tool executions
```

**Example 2: With aggregation**
```javascript
query({
  resource: "tools",
  aggregate: {function: "count", field: "tool_name"}
})
// Pipeline: tools → count by tool_name → output
// Result: [{"tool_name": "Read", "count": 45}, ...]
```

**Example 3: Complex pipeline**
```javascript
query({
  resource: "tools",
  filter: {tool_status: "error"},
  transform: {group_by: "tool_name"},
  aggregate: {function: "count"},
  output: {sort_by: "count", sort_order: "desc"}
})
// Pipeline: tools → filter errors → group by tool → count → sort → output
// Result: Error counts by tool, sorted descending
```

---

## Advanced Usage

### Using jq_filter

For complex transformations beyond the structured query, use `jq_filter`:

```javascript
query({
  resource: "tools",
  filter: {tool_status: "error"},
  jq_filter: `
    group_by(.tool_name) |
    map({
      tool: .[0].tool_name,
      count: length,
      error_rate: (length / 100 * 100)
    }) |
    sort_by(.count) |
    reverse
  `
})
```

### Combining Queries

For complex analysis, chain multiple queries:

```javascript
// Step 1: Get error distribution
const errors = query({
  resource: "tools",
  filter: {tool_status: "error"},
  aggregate: {function: "count", field: "tool_name"}
})

// Step 2: Get total tool calls
const total = query({
  resource: "tools",
  aggregate: {function: "count"}
})

// Step 3: Calculate error rates
const errorRates = errors.map(e => ({
  tool: e.tool_name,
  error_count: e.count,
  error_rate: (e.count / total.count * 100).toFixed(2) + "%"
}))
```

---

## Performance Considerations

### Query Optimization

1. **Use specific filters**: Narrow down data early in pipeline
   ```javascript
   // Good
   filter: {tool_name: "Bash", tool_status: "error"}

   // Bad
   jq_filter: '.[] | select(.tool_name == "Bash" and .status == "error")'
   ```

2. **Limit results when possible**:
   ```javascript
   output: {limit: 100}
   ```

3. **Use aggregation for large datasets**:
   ```javascript
   // Good (summary only)
   aggregate: {function: "count", field: "tool_name"}

   // Bad (all records)
   resource: "tools"
   ```

### Hybrid Output Mode

The unified query automatically uses hybrid output mode:

- **Inline mode** (≤8KB): Results embedded in response
- **File reference mode** (>8KB): Results written to temp file

**Configuration**:
```javascript
query({
  resource: "tools",
  inline_threshold_bytes: 16384 // 16KB threshold
})
```

---

## Schema Changes

### snake_case Standardization

All output fields now use `snake_case` to match JSONL source:

**Before (mixed case)**:
```json
{
  "ToolName": "Read",
  "UUID": "abc-123",
  "Timestamp": "2025-10-23T10:00:00Z"
}
```

**After (snake_case)**:
```json
{
  "tool_name": "Read",
  "uuid": "abc-123",
  "timestamp": "2025-10-23T10:00:00Z"
}
```

### Field Mapping

| Old Field | New Field | Notes |
|-----------|-----------|-------|
| `ToolName` | `tool_name` | Consistent with JSONL |
| `UUID` | `uuid` | Lowercase |
| `Timestamp` | `timestamp` | Lowercase |
| `SessionID` | `session_id` | Snake case |
| `ParentUUID` | `parent_uuid` | Snake case |
| `GitBranch` | `git_branch` | Snake case |

---

## Error Handling

### Invalid Parameters

```javascript
query({resource: "invalid"})
// Error: unknown resource type: invalid
```

### Empty Results

```javascript
query({
  resource: "tools",
  filter: {tool_name: "NonExistent"}
})
// Result: []
```

### Query Validation

The engine validates parameters before execution:

- `resource` must be "entries", "messages", or "tools"
- `scope` must be "session" or "project"
- `aggregate.function` must be valid function name
- `output.format` must be "jsonl", "tsv", or "summary"

---

## Migration Path

### Backward Compatibility

Old tools remain available as aliases during migration:

```javascript
// Old way (still works)
query_tools({tool: "Read", status: "error"})

// New way (recommended)
query({
  resource: "tools",
  filter: {tool_name: "Read", tool_status: "error"}
})
```

### Deprecation Timeline

- **v2.0.0** (current): Unified query introduced, old tools retained
- **v2.1.0** (+3 months): Old tools marked deprecated with warnings
- **v3.0.0** (+6 months): Old tools removed

See [Migration Guide](migration-to-unified-query.md) for detailed migration instructions.

---

## Best Practices

### 1. Start Simple

Begin with basic queries:
```javascript
query({resource: "tools"})
```

Then add filters:
```javascript
query({
  resource: "tools",
  filter: {tool_name: "Read"}
})
```

### 2. Use Aggregation for Summaries

Don't return all records when you need counts:
```javascript
// Good
aggregate: {function: "count", field: "tool_name"}

// Bad
resource: "tools"  // Returns all records
```

### 3. Combine Structured Query with jq

Use structured query for filtering, jq for complex transformations:
```javascript
query({
  resource: "tools",
  filter: {tool_status: "error"},  // Structured (efficient)
  jq_filter: "group_by(.tool_name) | ..."  // Complex transform
})
```

### 4. Leverage Output Control

Sort and limit results for better UX:
```javascript
output: {
  sort_by: "timestamp",
  sort_order: "desc",
  limit: 20
}
```

---

## Complete Example

**Task**: Analyze file operation errors by Git branch, show top 5 branches with most errors

```javascript
query({
  // Select tool executions
  resource: "tools",

  // Filter: file operations with errors
  filter: {
    tool_name: "Read|Edit|Write",
    tool_status: "error"
  },

  // Extract relevant fields and group by branch
  transform: {
    extract: ["git_branch", "tool_name", "error"],
    group_by: "git_branch"
  },

  // Count errors per branch
  aggregate: {
    function: "count",
    field: "tool_name"
  },

  // Output: top 5, sorted by count
  output: {
    sort_by: "count",
    sort_order: "desc",
    limit: 5
  }
})
```

**Expected output**:
```json
[
  {"git_branch": "feature/refactor", "tool_name": "Read", "count": 15},
  {"git_branch": "main", "tool_name": "Edit", "count": 8},
  {"git_branch": "bugfix/tests", "tool_name": "Write", "count": 5},
  {"git_branch": "feature/new", "tool_name": "Read", "count": 3},
  {"git_branch": "develop", "tool_name": "Edit", "count": 2}
]
```

---

## Related Documentation

- [Migration Guide](migration-to-unified-query.md) - Migrate from old tools
- [Query Cookbook](../examples/query-cookbook.md) - 10+ query examples
- [MCP Guide](mcp.md) - MCP server and tool usage
- [API Design Principles](../../experiments/bootstrap-006-api-design/data/api-parameter-convention.md) - Design rationale

---

## Feedback

Questions or issues with the unified query API?

- **GitHub Issues**: https://github.com/yaleh/meta-cc/issues
- **Discussions**: https://github.com/yaleh/meta-cc/discussions
- **Documentation**: https://github.com/yaleh/meta-cc/tree/main/docs

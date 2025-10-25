# MCP Query Cookbook

Practical, ready-to-use MCP query examples for common analysis scenarios using meta-cc v2.0 query tools.

## Table of Contents

1. [Error Analysis](#error-analysis) - 6 examples
2. [Workflow Optimization](#workflow-optimization) - 5 examples
3. [Performance Monitoring](#performance-monitoring) - 4 examples
4. [File Operations](#file-operations) - 3 examples
5. [Message Analysis](#message-analysis) - 3 examples
6. [Advanced jq Techniques](#advanced-jq-techniques) - 4 examples

**Total**: 25 practical examples

---

## Error Analysis

### 1. Find Recent Tool Errors

**Use case**: Debug recent failures in current session

**Convenience Tool**:
```javascript
query_tool_errors({
  scope: "session",
  limit: 10
})
```

**Core Query Equivalent**:
```javascript
query({
  resource: "tools",
  filter: {tool_status: "error"},
  scope: "session",
  jq_filter: 'sort_by(.timestamp) | reverse | .[0:10]'
})
```

**Output**:
```json
[
  {
    "tool_name": "Bash",
    "timestamp": "2025-10-25T10:30:00Z",
    "error": "command not found: npm",
    "input": {"command": "npm test"}
  }
]
```

**Analysis**:
- Most recent error shows npm not found
- Environment configuration issue
- Need to install Node.js or add to PATH

---

### 2. Error Rate by Tool

**Use case**: Identify which tools fail most often

**Core Query**:
```javascript
query({
  resource: "tools",
  scope: "project",
  jq_filter: 'group_by(.tool_name) | map({tool: .[0].tool_name, total: length, errors: [.[] | select(.status == "error")] | length, error_rate: (([.[] | select(.status == "error")] | length) / length * 100 | round)}) | sort_by(.error_rate) | reverse'
})
```

**Output**:
```json
[
  {"tool": "Bash", "total": 234, "errors": 23, "error_rate": 10},
  {"tool": "Edit", "total": 89, "errors": 5, "error_rate": 6},
  {"tool": "Read", "total": 156, "errors": 2, "error_rate": 1}
]
```

**Analysis**:
- Bash has 10% error rate (highest)
- Most errors are in shell commands
- Consider validating commands before execution

---

### 3. Error Patterns by Message

**Use case**: Find common error messages

**Core Query**:
```javascript
query({
  resource: "tools",
  filter: {tool_status: "error"},
  jq_filter: '[.[] | .error // "unknown"] | group_by(.) | map({error_msg: .[0], count: length}) | sort_by(.count) | reverse | .[0:10]'
})
```

**Output**:
```json
[
  {"error_msg": "command not found", "count": 15},
  {"error_msg": "permission denied", "count": 8},
  {"error_msg": "file not found", "count": 5}
]
```

**Analysis**:
- "command not found" is most common (15 occurrences)
- Environment or PATH issues
- Create validation wrapper for commands

---

### 4. Errors by Git Branch

**Use case**: Identify problematic branches

**Raw jq Query**:
```javascript
query_raw({
  jq_expression: '.[] | select(.type == "assistant") | .message.content[] | select(.type == "tool_result" and .is_error == true) | {branch: .git_branch // "unknown", error: .content} | group_by(.branch) | map({branch: .[0].branch, count: length})',
  scope: "project"
})
```

**Output**:
```json
[
  {"branch": "feature/refactor", "count": 23},
  {"branch": "main", "count": 5},
  {"branch": "develop", "count": 2}
]
```

**Analysis**:
- feature/refactor has most errors (23)
- Suggests instability in refactoring work
- Need more testing on feature branches

---

### 5. System API Errors

**Use case**: Track rate limits, timeouts, server errors

**Convenience Tool**:
```javascript
query_system_errors({
  scope: "project",
  jq_filter: 'group_by(.error.code) | map({code: .[0].error.code, count: length, sample_message: .[0].error.message})'
})
```

**Output**:
```json
[
  {"code": 529, "count": 12, "sample_message": "Overloaded"},
  {"code": 500, "count": 3, "sample_message": "Internal Server Error"},
  {"code": 408, "count": 2, "sample_message": "Request Timeout"}
]
```

**Analysis**:
- 12 overload errors (code 529)
- System under heavy load during sessions
- Consider rate limiting or retry logic

---

### 6. Error Timeline

**Use case**: See when errors occur during development

**Core Query**:
```javascript
query({
  resource: "tools",
  filter: {tool_status: "error"},
  jq_filter: '.[] | {timestamp: .timestamp[0:13], tool: .tool_name} | group_by(.timestamp) | map({hour: .[0].timestamp, error_count: length, tools: [.[].tool | unique]})'
})
```

**Output**:
```json
[
  {"hour": "2025-10-25T09", "error_count": 5, "tools": ["Bash", "Edit"]},
  {"hour": "2025-10-25T10", "error_count": 12, "tools": ["Bash", "Read", "Write"]},
  {"hour": "2025-10-25T11", "error_count": 3, "tools": ["Bash"]}
]
```

**Analysis**:
- Error spike at 10:00 (12 errors)
- Multiple tools failing simultaneously
- Possible environment or system issue at that time

---

## Workflow Optimization

### 7. Tool Usage Frequency

**Use case**: Understand which tools you rely on most

**Legacy Tool**:
```javascript
query_tools({
  scope: "project",
  jq_filter: 'group_by(.tool_name) | map({tool: .[0].tool_name, count: length}) | sort_by(.count) | reverse | .[0:10]'
})
```

**Output**:
```json
[
  {"tool": "Bash", "count": 234},
  {"tool": "Read", "count": 156},
  {"tool": "Edit", "count": 89},
  {"tool": "Write", "count": 45},
  {"tool": "Glob", "count": 38}
]
```

**Analysis**:
- Heavy reliance on Bash (234 calls)
- File operations (Read/Edit/Write) are common
- Consider creating command shortcuts for frequent patterns

---

### 8. Tool Sequences

**Use case**: Find common tool usage patterns

**Legacy Tool**:
```javascript
query_tool_sequences({
  min_occurrences: 5,
  include_builtin_tools: false,
  jq_filter: '.[] | select(.count > 10) | {pattern, count, success_rate}'
})
```

**Output**:
```json
[
  {"pattern": "Read → Edit → Write", "count": 23, "success_rate": 0.95},
  {"pattern": "Glob → Read → Edit", "count": 15, "success_rate": 0.88},
  {"pattern": "Bash → Read", "count": 12, "success_rate": 0.75}
]
```

**Analysis**:
- Most common pattern: Read → Edit → Write (23 times)
- High success rate (95%)
- Standard file modification workflow

---

### 9. Response Time Analysis

**Use case**: Identify slow operations

**Core Query**:
```javascript
query({
  resource: "messages",
  filter: {role: "assistant"},
  jq_filter: '.[] | select(.usage.input_tokens > 0) | {timestamp, tokens: .usage.input_tokens + .usage.output_tokens, model} | select(.tokens > 10000)'
})
```

**Output**:
```json
[
  {
    "timestamp": "2025-10-25T10:15:00Z",
    "tokens": 15234,
    "model": "claude-sonnet-4-5"
  }
]
```

**Analysis**:
- One response used >15K tokens
- Large context or complex operation
- Check if optimization possible

---

### 10. Token Usage by Session

**Use case**: Track token consumption over time

**Convenience Tool**:
```javascript
query_token_usage({
  scope: "project",
  jq_filter: 'group_by(.timestamp[0:10]) | map({date: .[0].timestamp[0:10], total_input: [.[].usage.input_tokens] | add, total_output: [.[].usage.output_tokens] | add, cache_hits: [.[].usage.cache_read_input_tokens // 0] | add})'
})
```

**Output**:
```json
[
  {
    "date": "2025-10-24",
    "total_input": 125000,
    "total_output": 45000,
    "cache_hits": 80000
  },
  {
    "date": "2025-10-25",
    "total_input": 98000,
    "total_output": 38000,
    "cache_hits": 65000
  }
]
```

**Analysis**:
- Oct 24: 125K input tokens, 80K from cache (64% cache hit)
- Oct 25: 98K input tokens, 65K from cache (66% cache hit)
- Good cache utilization

---

### 11. Conversation Flow Analysis

**Use case**: Understand user-assistant interaction patterns

**Convenience Tool**:
```javascript
query_conversation_flow({
  scope: "session",
  jq_filter: '.[] | {turn, role, content_length: .content | length, timestamp}'
})
```

**Output**:
```json
[
  {"turn": 1, "role": "user", "content_length": 234, "timestamp": "2025-10-25T09:00:00Z"},
  {"turn": 2, "role": "assistant", "content_length": 1523, "timestamp": "2025-10-25T09:00:30Z"},
  {"turn": 3, "role": "user", "content_length": 89, "timestamp": "2025-10-25T09:02:00Z"}
]
```

**Analysis**:
- Turn 1: User provides detailed request (234 chars)
- Turn 2: Assistant gives comprehensive response (1523 chars)
- Turn 3: User follow-up is brief (89 chars)
- Natural conversational flow

---

## Performance Monitoring

### 12. Session Duration

**Use case**: Track how long development sessions last

**Core Query**:
```javascript
query({
  resource: "entries",
  scope: "project",
  jq_filter: 'group_by(.sessionId) | map({session: .[0].sessionId, start: .[0].timestamp, end: .[-1].timestamp, duration_minutes: (((.[-1].timestamp | fromdateiso8601) - (.[0].timestamp | fromdateiso8601)) / 60 | round)})'
})
```

**Output**:
```json
[
  {
    "session": "abc123",
    "start": "2025-10-25T09:00:00Z",
    "end": "2025-10-25T11:30:00Z",
    "duration_minutes": 150
  }
]
```

**Analysis**:
- Session lasted 150 minutes (2.5 hours)
- Typical development session length
- Consider breaks for longer sessions

---

### 13. Tool Execution Success Rate

**Use case**: Monitor tool reliability

**Core Query**:
```javascript
query({
  resource: "tools",
  scope: "project",
  jq_filter: 'group_by(.tool_name) | map({tool: .[0].tool_name, total: length, success: [.[] | select(.status == "success")] | length, success_rate: (([.[] | select(.status == "success")] | length) / length * 100 | round)}) | sort_by(.success_rate)'
})
```

**Output**:
```json
[
  {"tool": "Bash", "total": 234, "success": 211, "success_rate": 90},
  {"tool": "Edit", "total": 89, "success": 84, "success_rate": 94},
  {"tool": "Read", "total": 156, "success": 154, "success_rate": 99}
]
```

**Analysis**:
- Read has highest success rate (99%)
- Bash has lowest (90%) due to command errors
- Edit is reliable at 94%

---

### 14. File Operation Tracking

**Use case**: Monitor file access patterns

**Convenience Tool**:
```javascript
query_file_snapshots({
  scope: "session",
  jq_filter: '.[] | {file: .file_path, operation: .operation, timestamp} | group_by(.file) | map({file: .[0].file, operations: [.[].operation], access_count: length})'
})
```

**Output**:
```json
[
  {
    "file": "/home/user/project/cmd/mcp-server/executor.go",
    "operations": ["read", "edit", "read", "edit"],
    "access_count": 4
  }
]
```

**Analysis**:
- executor.go accessed 4 times
- Pattern: read → edit → read → edit
- Iterative development on single file

---

### 15. Cache Hit Rate

**Use case**: Monitor prompt caching effectiveness

**Convenience Tool**:
```javascript
query_token_usage({
  scope: "project",
  stats_only: true
})
```

**Output**:
```json
{
  "total_messages": 234,
  "total_input_tokens": 523000,
  "total_output_tokens": 189000,
  "total_cache_creation": 45000,
  "total_cache_hits": 312000,
  "cache_hit_rate": 0.597
}
```

**Analysis**:
- 59.7% of input tokens from cache
- Good caching utilization
- Prompt reuse working effectively

---

## File Operations

### 16. Most Edited Files

**Use case**: Identify development hotspots

**Legacy Tool**:
```javascript
query_file_access({
  file: "*",
  jq_filter: '.[] | select(.operations.Edit > 5) | {file, edit_count: .operations.Edit, total_accesses} | sort_by(.edit_count) | reverse'
})
```

**Alternative with Core Query**:
```javascript
query({
  resource: "tools",
  filter: {tool_name: "Edit"},
  jq_filter: '[.[] | .input.file_path] | group_by(.) | map({file: .[0], count: length}) | sort_by(.count) | reverse | .[0:10]'
})
```

**Output**:
```json
[
  {"file": "/home/user/project/cmd/mcp-server/executor.go", "count": 23},
  {"file": "/home/user/project/internal/query/tools.go", "count": 18},
  {"file": "/home/user/project/docs/guides/mcp.md", "count": 12}
]
```

**Analysis**:
- executor.go edited 23 times (most modified)
- Core files getting frequent updates
- Potential refactoring target

---

### 17. File Access Timeline

**Use case**: See file modification history

**Legacy Tool**:
```javascript
query_file_access({
  file: "/home/user/project/cmd/mcp-server/executor.go"
})
```

**Output**:
```json
{
  "file": "/home/user/project/cmd/mcp-server/executor.go",
  "total_accesses": 45,
  "operations": {"Read": 22, "Edit": 23},
  "timeline": [
    {"turn": 5, "timestamp": "2025-10-25T09:15:00Z", "operation": "Read"},
    {"turn": 6, "timestamp": "2025-10-25T09:20:00Z", "operation": "Edit"}
  ],
  "time_span_minutes": 180
}
```

**Analysis**:
- 45 total accesses over 3 hours
- Nearly equal reads and edits (22 vs 23)
- Active development on this file

---

### 18. Untracked File Writes

**Use case**: Find files written but not committed

**Core Query**:
```javascript
query({
  resource: "tools",
  filter: {tool_name: "Write"},
  jq_filter: '[.[] | .input.file_path] | unique'
})
```

**Output**:
```json
[
  "/home/user/project/temp_output.json",
  "/home/user/project/test_results.txt",
  "/home/user/project/debug.log"
]
```

**Analysis**:
- 3 files written during session
- Likely temporary or debug files
- Consider adding to .gitignore

---

## Message Analysis

### 19. Search User Messages

**Use case**: Find when user mentioned specific topics

**Legacy Tool**:
```javascript
query_user_messages({
  pattern: "error|bug|fix",
  content_summary: false,
  limit: 10
})
```

**Output**:
```json
[
  {
    "turn": 12,
    "timestamp": "2025-10-25T10:15:00Z",
    "content": "There's an error in the executor.go file when parsing jq expressions. Can you fix it?"
  }
]
```

**Analysis**:
- User reported error at turn 12
- Specific file mentioned (executor.go)
- jq parsing issue

---

### 20. Long User Messages

**Use case**: Identify detailed user requests

**Core Query**:
```javascript
query({
  resource: "messages",
  filter: {role: "user"},
  jq_filter: '.[] | select(.content | length > 1000) | {turn, timestamp, length: .content | length, preview: .content[0:100]}'
})
```

**Output**:
```json
[
  {
    "turn": 1,
    "timestamp": "2025-10-25T09:00:00Z",
    "length": 2345,
    "preview": "I need to implement a comprehensive query system for Claude Code session data..."
  }
]
```

**Analysis**:
- Initial message was very detailed (2345 chars)
- Comprehensive project requirements
- Clear specification helps development

---

### 21. Assistant Tool Usage in Messages

**Use case**: See which tools assistant used in responses

**Convenience Tool**:
```javascript
query_tool_blocks({
  block_type: "tool_use",
  jq_filter: 'group_by(.name) | map({tool: .[0].name, count: length}) | sort_by(.count) | reverse'
})
```

**Output**:
```json
[
  {"tool": "Read", "count": 156},
  {"tool": "Edit", "count": 89},
  {"tool": "Bash", "count": 234}
]
```

**Analysis**:
- Bash most used in assistant responses
- File operations (Read/Edit) common
- Matches expected development workflow

---

## Advanced jq Techniques

### 22. Time-based Filtering

**Use case**: Query last hour of activity

**Raw jq Query**:
```javascript
query_raw({
  jq_expression: '.[] | select(.timestamp) | select(.timestamp | fromdateiso8601 > (now - 3600)) | {timestamp, type, tool_name: (.message.content.tool_use.name // "N/A")}'
})
```

**Output**:
```json
[
  {"timestamp": "2025-10-25T10:45:00Z", "type": "assistant", "tool_name": "Bash"},
  {"timestamp": "2025-10-25T10:50:00Z", "type": "user", "tool_name": "N/A"}
]
```

**Analysis**:
- Recent activity in last hour
- Mix of user and assistant messages
- Assistant used Bash recently

---

### 23. Cross-session Aggregation

**Use case**: Compare tool usage across multiple sessions

**Core Query**:
```javascript
query({
  resource: "tools",
  scope: "project",
  jq_filter: 'group_by(.session_id // .uuid[0:8]) | map({session: .[0].session_id // .[0].uuid[0:8], tools: group_by(.tool_name) | map({tool: .[0].tool_name, count: length}), total_tools: length})'
})
```

**Output**:
```json
[
  {
    "session": "abc123",
    "tools": [{"tool": "Bash", "count": 45}, {"tool": "Read", "count": 30}],
    "total_tools": 75
  },
  {
    "session": "def456",
    "tools": [{"tool": "Edit", "count": 60}, {"tool": "Write", "count": 20}],
    "total_tools": 80
  }
]
```

**Analysis**:
- Session abc123: Exploration (Bash/Read heavy)
- Session def456: Implementation (Edit/Write heavy)
- Different workflow patterns

---

### 24. Nested Field Extraction

**Use case**: Extract deeply nested tool parameters

**Raw jq Query**:
```javascript
query_raw({
  jq_expression: '.[] | select(.type == "assistant") | .message.content[] | select(.type == "tool_use" and .name == "Bash") | {command: .input.command, timestamp: .timestamp} | select(.command | contains("test"))'
})
```

**Output**:
```json
[
  {"command": "make test", "timestamp": "2025-10-25T09:30:00Z"},
  {"command": "go test ./...", "timestamp": "2025-10-25T10:15:00Z"}
]
```

**Analysis**:
- 2 test commands executed
- Using both make and go test
- Testing activity during session

---

### 25. Complex Grouping

**Use case**: Multi-level aggregation by tool and status

**Core Query**:
```javascript
query({
  resource: "tools",
  scope: "project",
  jq_filter: 'group_by(.tool_name) | map({tool: .[0].tool_name, by_status: group_by(.status) | map({status: .[0].status, count: length, avg_timestamp: ([.[].timestamp] | sort | .[length/2])})})'
})
```

**Output**:
```json
[
  {
    "tool": "Bash",
    "by_status": [
      {"status": "success", "count": 211, "avg_timestamp": "2025-10-25T10:00:00Z"},
      {"status": "error", "count": 23, "avg_timestamp": "2025-10-25T10:30:00Z"}
    ]
  }
]
```

**Analysis**:
- Bash: 211 successes, 23 errors
- Errors occurred later in session (10:30)
- Could indicate environmental changes

---

## Query Patterns Reference

### Common jq Patterns

**Select by field value**:
```javascript
.[] | select(.tool_name == "Bash")
```

**Filter by multiple conditions**:
```javascript
.[] | select(.tool_name == "Bash" and .status == "error")
```

**Group and count**:
```javascript
group_by(.tool_name) | map({tool: .[0].tool_name, count: length})
```

**Sort and limit**:
```javascript
sort_by(.timestamp) | reverse | .[0:10]
```

**Aggregate sum**:
```javascript
[.[] | .usage.input_tokens] | add
```

**Time filtering (last hour)**:
```javascript
.[] | select(.timestamp | fromdateiso8601 > (now - 3600))
```

**Regex search**:
```javascript
.[] | select(.content | test("error|bug"; "i"))
```

**Null handling**:
```javascript
.[] | select(.error != null)
```

**Extract nested field**:
```javascript
.[] | .message.content.tool_use.name
```

**Calculate percentage**:
```javascript
(([.[] | select(.status == "error")] | length) / length * 100 | round)
```

---

## Best Practices

### 1. Start Simple, Add Complexity

```javascript
// Step 1: Get all tools
query({resource: "tools"})

// Step 2: Filter by name
query({resource: "tools", jq_filter: '.[] | select(.tool_name == "Bash")'})

// Step 3: Add error filtering
query({resource: "tools", jq_filter: '.[] | select(.tool_name == "Bash" and .status == "error")'})
```

### 2. Use Convenience Tools First

```javascript
// Good: Use convenience tool
query_tool_errors({limit: 10})

// Overkill: Complex jq for simple query
query_raw({jq_expression: '.[] | select(.type == "assistant") | ...'})
```

### 3. Test jq Expressions Locally

```bash
# Test with sample data
echo '[{"tool":"Bash","status":"error"}]' | jq '.[] | select(.status == "error")'
```

### 4. Handle Large Results

```javascript
// Let hybrid mode handle size
query({resource: "tools"})

// Adjust threshold if needed
query({resource: "tools", inline_threshold_bytes: 1024})
```

### 5. Use Scope Appropriately

```javascript
// Session: Current work
query_tool_errors({scope: "session"})

// Project: Trend analysis
query_tool_errors({scope: "project"})
```

---

## See Also

- [MCP Query Tools Reference](../guides/mcp-query-tools.md) - Complete tool documentation
- [MCP v2.0 Migration Guide](../guides/mcp-v2-migration.md) - Migrating from v1.x
- [jq Manual](https://jqlang.github.io/jq/manual/) - Official jq documentation
- [Query Cookbook](query-cookbook.md) - Programmatic query examples
- [Frequent JSONL Queries](frequent-jsonl-queries.md) - CLI query patterns

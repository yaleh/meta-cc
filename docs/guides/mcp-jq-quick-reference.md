# MCP jq Quick Reference

This guide provides quick reference for using jq expressions with meta-cc MCP queries.

## Data Structure Field Reference

### Common Fields Across Resources

| Resource Type | Field Name | Type | Description | Common Mistakes |
|--------------|------------|------|-------------|-----------------|
| **messages** | `.content` | string | Message text content | ❌ `.message_content` |
| **messages** | `.uuid` | string | Unique message ID | |
| **messages** | `.timestamp` | string | ISO8601 timestamp | |
| **messages** | `.role` | string | "user" or "assistant" | |
| **messages** | `.session_id` | string | Session identifier | |
| **tools** | `.tool_name` | string | Tool name (e.g., "Bash", "Read") | ❌ `.tool` |
| **tools** | `.status` | string | "success" or "error" | |
| **tools** | `.error` | string | Error message if failed | |
| **tools** | `.timestamp` | string | ISO8601 timestamp | |

### Resource-Specific Fields

#### UserMessage (from query_user_messages)
```json
{
  "turn_sequence": 1,
  "uuid": "...",
  "timestamp": "2025-10-25T10:30:00Z",
  "content": "user message text"
}
```

#### MessageView (from unified query)
```json
{
  "uuid": "...",
  "session_id": "...",
  "timestamp": "2025-10-25T10:30:00Z",
  "role": "user",
  "content": "message text",
  "git_branch": "main"
}
```

#### ToolCall
```json
{
  "uuid": "...",
  "tool_name": "Bash",
  "status": "success",
  "error": "",
  "timestamp": "2025-10-25T10:30:00Z"
}
```

---

## Common jq Expression Patterns

### 1. Limiting Results

❌ **WRONG** (gojq doesn't support `limit()` function):
```jq
.[] | limit(5)
```

✓ **CORRECT** (use array slicing):
```jq
.[0:5]          # First 5 items
.[-5:]          # Last 5 items
.[10:20]        # Items 10-19
```

### 2. Field Selection and Projection

```jq
# Extract single field
.[] | .content

# Project multiple fields
.[] | {uuid: .uuid, content: .content}

# Rename fields
.[] | {id: .uuid, text: .content}
```

### 3. Filtering

```jq
# Filter by exact match
.[] | select(.status == "error")

# Filter by pattern (regex)
.[] | select(.content | test("bug|fix"))

# Filter by field existence
.[] | select(.error != "")

# Multiple conditions (AND)
.[] | select(.status == "error" and .tool_name == "Bash")

# Multiple conditions (OR)
.[] | select(.status == "error" or .error != "")
```

### 4. Time-based Filtering

❌ **NOT RECOMMENDED** (string comparison may fail):
```jq
.[] | select(.timestamp | startswith("2025-10-25"))
```

✓ **RECOMMENDED** (use filter parameter):
```javascript
// Via MCP query tool
{
  resource: "messages",
  filter: {
    time_range: {
      start: "2025-10-25T00:00:00Z",
      end: "2025-10-25T23:59:59Z"
    }
  }
}
```

**ACCEPTABLE** (for jq-only queries):
```jq
# Compare timestamps lexicographically (works for ISO8601)
.[] | select(.timestamp >= "2025-10-25T00:00:00Z" and .timestamp < "2025-10-26T00:00:00Z")
```

### 5. Sorting

```jq
# Sort by timestamp (ascending)
sort_by(.timestamp)

# Sort by timestamp (descending)
sort_by(.timestamp) | reverse

# Sort by multiple fields
sort_by(.status, .timestamp)
```

### 6. Grouping and Counting

```jq
# Group by field
group_by(.tool_name)

# Count items per group
group_by(.tool_name) | map({tool: .[0].tool_name, count: length})

# Count all items
length
```

### 7. Aggregation

```jq
# Sum numeric field
[.[] | .duration] | add

# Average
[.[] | .duration] | add / length

# Min/Max
[.[] | .duration] | min
[.[] | .duration] | max
```

---

## Common Use Cases

### Use Case 1: Find Error Messages Today

✓ **Recommended approach** (use MCP filter):
```javascript
query_user_messages({
  pattern: "error|fail",
  filter: {
    time_range: {
      start: "2025-10-25T00:00:00Z",
      end: "2025-10-25T23:59:59Z"
    }
  }
})
```

**Fallback** (jq-only):
```jq
.[] | select(.timestamp >= "2025-10-25T00:00:00Z" and .timestamp < "2025-10-26T00:00:00Z") | select(.content | test("error|fail"))
```

### Use Case 2: Get Latest 10 Messages

```jq
sort_by(.timestamp) | reverse | .[0:10]
```

### Use Case 3: Count Tools by Status

```jq
group_by(.status) | map({status: .[0].status, count: length})
```

### Use Case 4: Extract Specific Fields

```jq
.[] | {
  time: .timestamp,
  message: .content,
  session: .session_id
}
```

### Use Case 5: Filter and Transform

```jq
.[] | select(.status == "error") | {
  tool: .tool_name,
  error: .error,
  when: .timestamp
}
```

---

## Common Errors and Solutions

### Error 1: Field Not Found

❌ **Error**:
```jq
.[] | .message_content
# Returns: null (field doesn't exist)
```

✓ **Solution**:
```jq
.[] | .content
# Use correct field name from schema
```

### Error 2: limit() Not Supported

❌ **Error**:
```jq
.[] | limit(5)
# Error: function not defined: limit/1
```

✓ **Solution**:
```jq
.[0:5]
# Use array slicing
```

### Error 3: head()/tail() Not Supported

❌ **Error**:
```jq
.[] | head
# Error: function not defined: head/0
```

✓ **Solution**:
```jq
.[0]      # First item (head)
.[-1]     # Last item (tail)
.[0:10]   # First 10 items
```

### Error 4: Incorrect Time Filtering

❌ **Problem**:
```jq
.[] | select(.timestamp | startswith("2025-10-25"))
# May miss items or be unreliable
```

✓ **Solution**:
```jq
# Option 1: Use MCP filter parameter (recommended)
{filter: {time_range: {start: "...", end: "..."}}}

# Option 2: Lexicographic comparison (acceptable)
.[] | select(.timestamp >= "2025-10-25T00:00:00Z" and .timestamp < "2025-10-26T00:00:00Z")
```

### Error 5: Quoted jq Expression

❌ **Error**:
```jq
".[] | select(.status == 'error')"
# Error: expression appears to be quoted
```

✓ **Solution**:
```jq
.[] | select(.status == "error")
# Remove outer quotes, keep inner quotes
```

---

## Advanced Patterns

### Pattern 1: Conditional Field Access

```jq
# Safe field access with default
.[] | .error // "no error"

# Conditional selection
.[] | if .status == "error" then .error else "OK" end
```

### Pattern 2: Nested Field Extraction

```jq
# Extract from nested structure
.[] | .content_blocks[] | select(.type == "text") | .text
```

### Pattern 3: Custom Statistics

```jq
# Calculate error rate
{
  total: length,
  errors: [.[] | select(.status == "error")] | length,
  error_rate: ([.[] | select(.status == "error")] | length) / length * 100
}
```

### Pattern 4: Timeline Analysis

```jq
# Group by hour
group_by(.timestamp[0:13]) | map({
  hour: .[0].timestamp[0:13],
  count: length,
  errors: [.[] | select(.status == "error")] | length
})
```

---

## Best Practices

1. **Use MCP filters first**: Prefer `filter` parameter over jq filtering for better performance
2. **Know your schema**: Reference the field tables at the top of this document
3. **Test incrementally**: Build complex expressions step by step
4. **Use array slicing**: Prefer `.[0:N]` over attempting to use `limit()`
5. **Validate timestamps**: Use ISO8601 format for reliable time comparisons
6. **Quote strings properly**: Use double quotes inside jq expressions

---

## Quick Cheat Sheet

```jq
# Basic
.[]                    # Iterate array
.[0]                   # First item
.[-1]                  # Last item
.[0:5]                 # First 5 items
.field                 # Access field

# Filter
select(.x == "y")      # Filter by condition
select(.x | test("pattern"))  # Filter by regex

# Transform
{a: .b}                # Rename field
map(...)               # Transform each item

# Aggregate
length                 # Count items
group_by(.field)       # Group by field
sort_by(.field)        # Sort by field
reverse                # Reverse order

# Combine
| select(...) | map(...) | .[0:10]  # Pipeline
```

---

## Related Documentation

- [MCP Guide](mcp.md) - Complete MCP server reference
- [Unified Query API](unified-query-api.md) - Modern query interface
- [Query Cookbook](../examples/query-cookbook.md) - Practical examples

---

**Last Updated**: 2025-10-25
**Version**: 2.0.1

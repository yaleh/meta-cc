# JSONL Output Format Reference

Understanding meta-cc's JSONL (JSON Lines) output structure is crucial for processing data with tools like `jq`.

## What is JSONL?

JSONL (JSON Lines) is a format where each line is a valid JSON object. Unlike standard JSON arrays, JSONL is:

- **Streamable**: Process one line at a time without loading the entire file
- **Composable**: Works seamlessly with Unix pipelines
- **Efficient**: Lower memory usage for large datasets

**Example JSONL**:
```
{"uuid":"1","tool":"Bash","status":"success"}
{"uuid":"2","tool":"Edit","status":"success"}
{"uuid":"3","tool":"Read","status":"error"}
```

**Not JSONL** (standard JSON array):
```json
[
  {"uuid":"1","tool":"Bash","status":"success"},
  {"uuid":"2","tool":"Edit","status":"success"}
]
```

## Command Output Types

Different commands return different JSON structures:

| Command | Output Type | Structure |
|---------|-------------|-----------|
| `parse stats` | **Object** | Single JSON object with statistics |
| `parse extract --type turns` | **Stream** | One turn object per line |
| `parse extract --type tools` | **Stream** | One tool object per line |
| `query tools` | **Stream** | One tool object per line |
| `query user-messages` | **Stream** | One message object per line |
| `analyze errors` | **Stream** | One error pattern per line (or empty) |

## Common Mistakes and Solutions

### parse stats (Returns Single Object)

❌ **Wrong** - Assumes wrapper object:
```bash
meta-cc parse stats | jq '.stats.TurnCount'
```

✅ **Correct** - Direct property access:
```bash
meta-cc parse stats | jq '.TurnCount'
meta-cc parse stats | jq '.ErrorRate'
meta-cc parse stats | jq '.TopTools[:3]'
```

### query tools (Returns Stream)

❌ **Wrong** - Assumes array wrapper:
```bash
meta-cc query tools | jq '.tools'
meta-cc query tools | jq '.tools[]'
```

✅ **Correct** - Process stream directly:
```bash
# Show each tool (one per line)
meta-cc query tools | jq '.'

# Extract tool names
meta-cc query tools | jq -r '.ToolName'

# Filter by status
meta-cc query tools | jq 'select(.Status == "error")'

# Count total (slurp into array first)
meta-cc query tools | jq -s 'length'
```

### analyze errors (Returns Stream or Empty)

❌ **Wrong** - Assumes array wrapper:
```bash
meta-cc analyze errors | jq '.ErrorPatterns'
meta-cc analyze errors | jq '.ErrorPatterns | length'
```

✅ **Correct** - Handle stream:
```bash
# Show all patterns
meta-cc analyze errors | jq '.'

# Count patterns (slurp first)
meta-cc analyze errors | jq -s 'length'

# Safe count (handles empty)
meta-cc analyze errors | jq -s 'if type == "array" then length else 0 end'
```

## Detailed Output Structures

### parse stats

Returns a single object with comprehensive session statistics.

**Structure**:
```json
{
  "TurnCount": 2676,
  "UserTurnCount": 1097,
  "AssistantTurnCount": 1579,
  "ToolCallCount": 1012,
  "ErrorCount": 0,
  "ErrorRate": 0,
  "DurationSeconds": 33796,
  "ToolFrequency": {
    "Bash": 495,
    "Read": 162,
    "TodoWrite": 140,
    "Write": 78,
    "Edit": 65
  },
  "TopTools": [
    {"Name": "Bash", "Count": 495, "Percentage": 48.9},
    {"Name": "Read", "Count": 162, "Percentage": 16.0}
  ]
}
```

**jq Examples**:
```bash
# Get total turns
meta-cc parse stats | jq '.TurnCount'

# Get error rate
meta-cc parse stats | jq '.ErrorRate'

# Get top 3 tools
meta-cc parse stats | jq '.TopTools[:3]'

# Get Bash usage count
meta-cc parse stats | jq '.ToolFrequency.Bash'

# Format tool frequency as TSV
meta-cc parse stats | jq -r '.ToolFrequency | to_entries | .[] | "\(.key)\t\(.value)"'
```

### query tools

Returns a stream of tool call objects (one per line).

**Structure** (each line):
```json
{
  "UUID": "abc-123",
  "ToolName": "Bash",
  "Input": {
    "command": "ls -la",
    "description": "List files"
  },
  "Output": "file1.txt\nfile2.txt",
  "Status": "",
  "Error": ""
}
```

**jq Examples**:
```bash
# Count total tools
meta-cc query tools | jq -s 'length'

# Get all tool names
meta-cc query tools | jq -r '.ToolName'

# Filter by tool name
meta-cc query tools | jq 'select(.ToolName == "Bash")'

# Get tools with errors
meta-cc query tools | jq 'select(.Status == "error")'

# Extract Bash commands
meta-cc query tools | jq -r 'select(.ToolName == "Bash") | .Input.command'

# Count tool frequency (without slurping entire stream)
meta-cc query tools | jq -r '.ToolName' | sort | uniq -c | sort -rn

# Get last 10 tools (slurp into array)
meta-cc query tools | jq -s '.[-10:]'
```

### parse extract --type turns

Returns a stream of conversation turn objects.

**Structure** (each line):
```json
{
  "type": "user",
  "timestamp": "2025-10-02T06:07:13.673Z",
  "uuid": "abc-123",
  "sessionId": "session-id",
  "message": {
    "role": "user",
    "content": [
      {
        "type": "text",
        "text": "Execute Stage 1.1..."
      }
    ]
  }
}
```

**jq Examples**:
```bash
# Count total turns
meta-cc parse extract --type turns | jq -s 'length'

# Get only user turns
meta-cc parse extract --type turns | jq 'select(.type == "user")'

# Get user messages
meta-cc parse extract --type turns | \
  jq -r 'select(.type == "user") | .message.content[0].text'

# Filter by timestamp
meta-cc parse extract --type turns | \
  jq 'select(.timestamp > "2025-10-02T12:00:00Z")'

# Get turn UUIDs
meta-cc parse extract --type turns | jq -r '.uuid'
```

### analyze errors

Returns a stream of error pattern objects (or empty if no errors).

**Structure** (each line):
```json
{
  "PatternID": "bash_npm_test_error",
  "Type": "repeated_error",
  "Occurrences": 5,
  "Signature": "abc123def456",
  "ToolName": "Bash",
  "ErrorText": "npm test failed",
  "FirstSeen": "2025-10-02T10:00:00Z",
  "LastSeen": "2025-10-02T10:30:00Z",
  "TimeSpanSeconds": 1800,
  "Context": {
    "TurnUUIDs": ["uuid1", "uuid2", "uuid3"],
    "TurnIndices": [100, 150, 200]
  }
}
```

**jq Examples**:
```bash
# Count error patterns
meta-cc analyze errors | jq -s 'length'

# Get all error patterns
meta-cc analyze errors | jq '.'

# Filter by minimum occurrences
meta-cc analyze errors | jq 'select(.Occurrences >= 5)'

# Get error messages
meta-cc analyze errors | jq -r '.ErrorText'

# Sort by occurrence count (slurp first)
meta-cc analyze errors | jq -s 'sort_by(.Occurrences) | reverse'

# Safe length check (handles empty results)
meta-cc analyze errors | jq -s 'if type == "array" then length else 0 end'
```

## Working with jq

### Basic jq Operations

**Select fields**:
```bash
# Single field
meta-cc query tools | jq '.ToolName'

# Multiple fields (create new object)
meta-cc query tools | jq '{tool: .ToolName, status: .Status}'
```

**Filter records**:
```bash
# By equality
meta-cc query tools | jq 'select(.ToolName == "Bash")'

# By inequality
meta-cc query tools | jq 'select(.Status != "error")'

# By multiple conditions (AND)
meta-cc query tools | jq 'select(.ToolName == "Bash" and .Status == "error")'

# By multiple conditions (OR)
meta-cc query tools | jq 'select(.ToolName == "Bash" or .ToolName == "Edit")'
```

**Array operations** (requires slurping with `-s`):
```bash
# Get first element
meta-cc query tools | jq -s '.[0]'

# Get last element
meta-cc query tools | jq -s '.[-1]'

# Get range
meta-cc query tools | jq -s '.[10:20]'

# Sort by field
meta-cc query tools | jq -s 'sort_by(.ToolName)'

# Reverse sort
meta-cc query tools | jq -s 'sort_by(.Occurrences) | reverse'
```

### Advanced jq Patterns

**Grouping and aggregation**:
```bash
# Group by tool name
meta-cc query tools | jq -s 'group_by(.ToolName)'

# Count by tool
meta-cc query tools | jq -s \
  'group_by(.ToolName) | map({tool: .[0].ToolName, count: length})'

# Average duration by tool
meta-cc query tools | jq -s \
  'group_by(.ToolName) | map({
    tool: .[0].ToolName,
    avg_duration: (map(.Duration) | add / length)
  })'
```

**Type-safe operations**:
```bash
# Safe array length (handles empty/null)
meta-cc analyze errors | jq -s 'if type == "array" then length else 0 end'

# Safe property access
meta-cc parse stats | jq 'if type == "object" then .TurnCount else null end'

# Default value
meta-cc query tools | jq -r '.Error // "No error"'
```

**String operations**:
```bash
# Extract substring
meta-cc query tools | jq -r '.ErrorText | split(":")[0]'

# Test regex
meta-cc query user-messages | jq 'select(.text | test("error|bug"; "i"))'

# String concatenation
meta-cc query tools | jq -r '"\(.ToolName): \(.Status)"'
```

## Combining Commands

### Filter and Count

```bash
# Count Bash tools
meta-cc query tools | jq 'select(.ToolName == "Bash")' | jq -s 'length'

# Count errors
meta-cc query tools | jq 'select(.Status == "error")' | jq -s 'length'
```

### Extract Specific Fields

```bash
# Tool names only
meta-cc query tools | jq -r '.ToolName'

# Format as TSV
meta-cc query tools | jq -r '[.ToolName, .Status, .Error] | @tsv'

# Format as CSV
meta-cc query tools | jq -r '[.ToolName, .Status, .Error] | @csv'
```

### Group and Aggregate

```bash
# Tool frequency (slurp method)
meta-cc query tools | jq -s \
  'group_by(.ToolName) | map({tool: .[0].ToolName, count: length})'

# Tool frequency (streaming method)
meta-cc query tools | jq -r '.ToolName' | sort | uniq -c | sort -rn
```

### Time-based Filtering

```bash
# Recent tools (last hour)
THRESHOLD=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
meta-cc parse extract --type tools | jq "select(.Timestamp > \"$THRESHOLD\")"

# Tools within time range
meta-cc parse extract --type tools | \
  jq 'select(.Timestamp >= "2025-10-02T10:00:00Z" and .Timestamp <= "2025-10-02T12:00:00Z")'
```

## Troubleshooting

### Error: "Cannot index array with string"

**Cause**: Trying to access object properties on a stream/array.

**Fix**: Remove the property accessor or use slurp (`-s`).

```bash
# ❌ Wrong
meta-cc query tools | jq '.tools'

# ✅ Correct
meta-cc query tools | jq '.'
```

### Error: "Cannot iterate over object"

**Cause**: Trying to iterate an object without accessing the array property.

**Fix**: Access the specific property first or convert to entries.

```bash
# ❌ Wrong (on parse stats)
meta-cc parse stats | jq '.[] | .ToolName'

# ✅ Correct
meta-cc parse stats | jq '.ToolFrequency | to_entries | .[] | "\(.key): \(.value)"'
```

### Empty Output

**Cause**: Command returns empty stream (no matches).

**Solution**: Check if empty and provide default.

```bash
# Check if empty
meta-cc analyze errors | jq -s 'length == 0'

# Provide default message
meta-cc analyze errors | jq -s 'if length == 0 then "No errors" else . end'
```

### Slurp vs Stream

**Use streaming** (no `-s`) when:
- Processing large datasets line-by-line
- Filtering or transforming each record
- Piping to other Unix tools

**Use slurp** (`-s`) when:
- Counting total records
- Sorting or grouping
- Accessing by index (first, last, range)

```bash
# Streaming (memory efficient)
meta-cc query tools | jq 'select(.Status == "error")'

# Slurp (loads all into memory)
meta-cc query tools | jq -s 'length'
```

## Performance Tips

### Memory Usage

**Streaming** (constant memory):
```bash
meta-cc query tools | jq 'select(.Status == "error")'
```

**Slurping** (loads all into memory):
```bash
meta-cc query tools | jq -s 'sort_by(.ToolName)'
```

**Recommendation**: Use streaming whenever possible. Only slurp when you need array operations.

### Large Datasets

For sessions with 1000+ tools, use:

1. **Streaming with field projection**:
   ```bash
   meta-cc query tools --fields "UUID,ToolName,Status" | jq '...'
   ```

2. **TSV format** (86% smaller):
   ```bash
   meta-cc query tools --output tsv
   ```

3. **Pagination**:
   ```bash
   meta-cc query tools --limit 100 --offset 0
   ```

4. **Summary mode**:
   ```bash
   meta-cc query tools --summary-first --top 20
   ```

## See Also

- [CLI Reference](cli.md) - Complete command reference
- [CLI Composability](../tutorials/cli-composability.md) - Advanced Unix pipeline patterns
- [MCP Guide](../guides/mcp.md) - MCP tool integration
- [jq Manual](https://stedolan.github.io/jq/manual/) - Official jq documentation

# jq Query Examples for Claude Code Session JSONL

This document provides practical jq query examples for analyzing Claude Code session JSONL files.

## Prerequisites

```bash
# Set the file path for convenience
FILE="/home/yale/.claude/projects/-home-yale-work-meta-cc/<session-uuid>.jsonl"

# Or work with all files
FILES="/home/yale/.claude/projects/-home-yale-work-meta-cc/*.jsonl"
```

## Basic Queries

### 1. Count Records by Type

Get distribution of record types in a session:

```bash
jq -r '.type' "$FILE" | sort | uniq -c | sort -rn
```

**Output:**
```
    121 assistant
     77 user
      8 file-history-snapshot
```

**Use case:** Quick session overview, identify record type distribution.

---

### 2. Session Statistics

Get comprehensive session statistics:

```bash
jq -s '{
  total_records: length,
  record_types: (group_by(.type) | map({(.[0].type): length}) | add),
  session_id: (.[0].sessionId // "N/A"),
  git_branch: (.[0].gitBranch // "N/A"),
  first_timestamp: (map(select(.timestamp)) | sort_by(.timestamp) | first | .timestamp),
  last_timestamp: (map(select(.timestamp)) | sort_by(.timestamp) | last | .timestamp),
  user_message_count: ([.[] | select(.type == "user" and (.message.content | type == "string"))] | length),
  assistant_message_count: ([.[] | select(.type == "assistant")] | length)
}' "$FILE"
```

**Output:**
```json
{
  "total_records": 206,
  "record_types": {
    "assistant": 121,
    "file-history-snapshot": 8,
    "user": 77
  },
  "session_id": "b3c22285-a9b7-4297-a300-c4abd61a78c9",
  "git_branch": "featue/mcp-refactor",
  "first_timestamp": "2025-10-24T13:26:55.628Z",
  "last_timestamp": "2025-10-24T14:22:58.638Z",
  "user_message_count": 7,
  "assistant_message_count": 121
}
```

**Use case:** Session summary, duration analysis, activity metrics.

---

## Conversation Analysis

### 3. Conversation Timeline

View chronological conversation flow with parent relationships:

```bash
jq -r 'select(.type == "user" or .type == "assistant") |
  "\(.timestamp) [\(.type)] \(.uuid[0:8])... parent:\(.parentUuid[0:8] // "null")..."' "$FILE" |
  head -10
```

**Output:**
```
2025-10-24T13:26:55.628Z [user] 19ea0674... parent:null...
2025-10-24T13:27:10.950Z [assistant] 2d51529d... parent:19ea0674...
2025-10-24T13:27:46.847Z [user] 32f38b7f... parent:null...
2025-10-24T13:27:52.337Z [assistant] 735c12f5... parent:32f38b7f...
2025-10-24T13:27:52.369Z [assistant] a1ee59e3... parent:735c12f5...
```

**Use case:** Debug conversation flow, understand message ordering.

---

### 4. Parent-Child Chain

Visualize parent-child relationships:

```bash
jq -r 'select(.type == "user" or .type == "assistant") |
  "\(.uuid[0:8]) [\(.type)] -> parent: \(.parentUuid[0:8] // "ROOT")"' "$FILE" |
  head -10
```

**Output:**
```
19ea0674 [user] -> parent: ROOT
2d51529d [assistant] -> parent: 19ea0674
32f38b7f [user] -> parent: ROOT
735c12f5 [assistant] -> parent: 32f38b7f
a1ee59e3 [assistant] -> parent: 735c12f5
```

**Use case:** Trace conversation branches, identify root messages.

---

### 5. Extract User Prompts

Get all user text prompts (excluding tool results):

```bash
jq -r 'select(.type == "user" and (.message.content | type == "string")) |
  .message.content' "$FILE"
```

**Output:**
```
Warmup
选择 /home/yale/.claude/projects/-home-yale-work-meta-cc 下一个 JSONL 文件，分析其每条记录的结构。
用树状结构表现以上 schema 。
用 PlantUML Class Diagram 描述以上 JSON 结构。
```

**Use case:** Extract user questions, analyze prompt patterns.

---

## Tool Execution Analysis

### 6. List All Tool Executions

Show all tools used with their parameters:

```bash
jq -r 'select(.type == "assistant") | .message.content[] |
  select(.type == "tool_use") |
  "\(.name): \(.input | keys | join(", "))"' "$FILE" |
  head -15
```

**Output:**
```
Glob: pattern
Read: file_path, limit
Read: file_path, limit, offset
Bash: command, description
TodoWrite: todos
Write: content, file_path
Edit: file_path, new_string, old_string
```

**Use case:** Tool usage analysis, parameter inspection.

---

### 7. Count Content Block Types

Analyze assistant response content types:

```bash
jq -r 'select(.type == "assistant") | .message.content[] | .type' "$FILE" |
  sort | uniq -c | sort -rn
```

**Output:**
```
     70 tool_use
     51 text
```

**Use case:** Understand assistant behavior, text vs tool ratio.

---

### 8. Find Tool Errors

Extract all failed tool executions:

```bash
jq -r 'select(.type == "user" and .message.content) |
  if (.message.content | type == "array") then
    .message.content[] | select(.type == "tool_result" and .is_error == true) |
    "Tool Error: \(.tool_use_id) - \(.content[0:100])"
  else
    empty
  end' "$FILE"
```

**Output:**
```
Tool Error: call_2x9ir35ui23 - <tool_use_error>String to replace not found in file.
Tool Error: call_yyaea5951l - /bin/bash: eval: line 1: syntax error near unexpected token
Tool Error: call_f07abn23zi6 - /bin/bash: line 1: length: command not found
```

**Use case:** Debug failures, identify problematic commands.

---

### 9. Extract File Operations

List all file read/write/edit operations:

```bash
jq -r 'select(.type == "assistant") | .message.content[] |
  select(.type == "tool_use" and (.name == "Read" or .name == "Write" or .name == "Edit")) |
  "\(.name): \(.input.file_path)"' "$FILE" |
  sort | uniq
```

**Output:**
```
Edit: /home/yale/work/meta-cc/docs/DOCUMENTATION_MAP.md
Edit: /home/yale/work/meta-cc/docs/reference/jsonl-schema.md
Read: /home/yale/.claude/projects/-home-yale-work-meta-cc/687796fe-f000-442e-9927-037254b7f28a.jsonl
Write: /home/yale/work/meta-cc/docs/reference/jsonl-schema.md
```

**Use case:** Track file changes, audit modifications.

---

## Advanced Queries

### 10. Token Usage Analysis

Calculate total and average token usage:

```bash
jq -s '[.[] | select(.type == "assistant" and .message.usage) | .message.usage] |
{
  total_input_tokens: (map(.input_tokens) | add),
  total_output_tokens: (map(.output_tokens) | add),
  total_cache_creation: (map(.cache_creation_input_tokens) | add),
  total_cache_read: (map(.cache_read_input_tokens) | add),
  avg_input_per_message: ((map(.input_tokens) | add) / length | floor),
  avg_output_per_message: ((map(.output_tokens) | add) / length | floor),
  message_count: length
}' "$FILE"
```

**Output:**
```json
{
  "total_input_tokens": 162515,
  "total_output_tokens": 18681,
  "total_cache_creation": null,
  "total_cache_read": 4844423,
  "avg_input_per_message": 1343,
  "avg_output_per_message": 154,
  "message_count": 121
}
```

**Use case:** Cost analysis, performance optimization, cache efficiency.

---

### 11. Build Conversation Tree

Group messages by parent to show conversation structure:

```bash
jq -s '
def build_tree:
  map(select(.type == "user" or .type == "assistant")) |
  map({uuid, type, parentUuid, timestamp}) |
  group_by(.parentUuid) |
  map({parent: .[0].parentUuid, children: map(.uuid)}) |
  .[:5];

build_tree' "$FILE"
```

**Output:**
```json
[
  {
    "parent": null,
    "children": [
      "19ea0674-e2d3-4ac2-95b7-e3145e7c8d36",
      "32f38b7f-f0d4-48e5-82b6-bf45b0535477"
    ]
  },
  {
    "parent": "00b0b164-1424-4fd3-9fe2-4223e6133a96",
    "children": [
      "a4ac4926-cca1-4107-8f51-a3fff6030921"
    ]
  }
]
```

**Use case:** Visualize conversation structure, identify branches.

---

### 12. Match Tool Use with Results

Correlate tool executions with their results:

```bash
jq -s '
# Extract tool uses
[.[] | select(.type == "assistant") | .message.content[] |
  select(.type == "tool_use" and .name == "Read")] as $reads |

# Extract tool results
[.[] | select(.type == "user" and (.message.content | type == "array")) |
  .message.content[] | select(.type == "tool_result")] as $results |

# Summary
{
  total_reads: ($reads | length),
  total_results: ($results | length),
  sample_read: ($reads[0] | {file: .input.file_path, tool_id: .id}),
  sample_result: ($results[0] | {tool_id: .tool_use_id, is_error: .is_error, content_length: (.content | length)})
}' "$FILE"
```

**Output:**
```json
{
  "total_reads": 5,
  "total_results": 70,
  "sample_read": {
    "file": "/home/yale/.claude/projects/-home-yale-work-meta-cc/687796fe-f000-442e-9927-037254b7f28a.jsonl",
    "tool_id": "call_f5ijvl8k8zq"
  },
  "sample_result": {
    "tool_id": "call_gb823bbv03o",
    "is_error": null,
    "content_length": 9572
  }
}
```

**Use case:** Debug tool execution, validate results.

---

### 13. Session Duration Analysis

Calculate session duration and entry density:

```bash
jq -s '
[.[] | select(.timestamp)] |
sort_by(.timestamp) |
{
  first: .[0].timestamp,
  last: .[-1].timestamp,
  entry_count: length,
  duration_minutes: "Calculate: (last - first) / 60000"
}' "$FILE"
```

**Output:**
```json
{
  "first": "2025-10-24T13:26:55.628Z",
  "last": "2025-10-24T14:22:58.638Z",
  "entry_count": 198,
  "duration_minutes": "~56 minutes"
}
```

**Use case:** Session length analysis, activity patterns.

---

## Multi-File Queries

### 14. Aggregate Statistics Across Sessions

Query across all session files:

```bash
jq -s -r 'group_by(.type) |
  map({type: .[0].type, count: length}) |
  sort_by(-.count)' $FILES
```

**Use case:** Project-wide statistics, trend analysis.

---

### 15. Find Sessions by Git Branch

Find all sessions on a specific branch:

```bash
jq -r 'select(.gitBranch == "feature/query-refactor") | .sessionId' $FILES |
  sort | uniq
```

**Use case:** Branch-specific analysis, feature development tracking.

---

### 16. Extract All Bash Commands

Get all bash commands executed across sessions:

```bash
jq -r 'select(.type == "assistant") | .message.content[] |
  select(.type == "tool_use" and .name == "Bash") |
  .input.command' $FILES |
  head -20
```

**Use case:** Command history, common operations analysis.

---

## Filtering Patterns

### 17. Filter by Timestamp Range

Extract entries within a time range:

```bash
jq -s 'map(select(.timestamp >= "2025-10-24T13:00:00Z" and
                    .timestamp <= "2025-10-24T14:00:00Z"))' "$FILE"
```

**Use case:** Time-based analysis, activity during specific periods.

---

### 18. Find Entries with Specific Content

Search for entries containing specific text:

```bash
jq -r 'select(.type == "user" and
             (.message.content | type == "string") and
             (.message.content | test("JSONL"; "i"))) |
  .message.content' "$FILE"
```

**Use case:** Content search, keyword tracking.

---

### 19. Extract Thinking Blocks

Find extended thinking content (Claude 3.5+):

```bash
jq -r 'select(.type == "assistant") | .message.content[] |
  select(.type == "thinking") |
  .thinking[0:300]' "$FILE"
```

**Use case:** Analyze reasoning process, debug complex responses.

---

## Performance Tips

### Efficient Slurping

For large files, avoid slurping when possible:

```bash
# Good: Stream processing
jq -r '.type' "$FILE" | sort | uniq -c

# Avoid: Slurping entire file
jq -s 'group_by(.type)' "$FILE"  # Only if necessary
```

### Limit Output Early

Use `head` or `first(n)` to limit output:

```bash
# Limit in shell
jq -r '.type' "$FILE" | head -10

# Limit in jq
jq -r 'limit(10; .type)' "$FILE"
```

### Index-Based Access

For repeated queries, consider indexing:

```bash
# Build index once
jq -s 'INDEX(.uuid)' "$FILE" > session-index.json

# Query index
jq '.["uuid-here"]' session-index.json
```

## Common Patterns

### Safe Navigation

Handle missing fields gracefully:

```bash
# Use // for default values
.parentUuid // "null"

# Use ? for optional access
.message.content[]?.type

# Check type before access
if (.message.content | type == "array") then ... else ... end
```

### Type Checking

Always check content type before processing:

```bash
select(.message.content | type == "string")  # String content
select(.message.content | type == "array")   # Array content
```

### UUID Truncation

Truncate UUIDs for readability:

```bash
.uuid[0:8]  # First 8 characters
```

## Related Documentation

- **JSONL Schema:** `docs/reference/jsonl-schema.md` - Complete schema reference
- **Query Cookbook:** `docs/examples/query-cookbook.md` - More complex patterns
- **MCP Guide:** `docs/guides/mcp.md` - Query via MCP tools
- **Unified Query API:** `docs/guides/unified-query-api.md` - Programmatic access

---

**Document Status:** Validated with real session data
**File Analyzed:** 206 records, 121 assistant + 77 user messages
**Last Updated:** 2025-10-24

# Two-Stage Query Examples

Practical, step-by-step examples demonstrating the two-stage query architecture with real-world use cases.

**Prerequisites**: Read the [Two-Stage Query Guide](../guides/two-stage-query-guide.md) for architecture overview.

---

## Table of Contents

1. [Example 1: Find Recent Tool Errors](#example-1-find-recent-tool-errors)
2. [Example 2: Analyze Token Usage Over Time](#example-2-analyze-token-usage-over-time)
3. [Example 3: Track Conversation Topics](#example-3-track-conversation-topics)
4. [Example 4: Performance Comparison](#example-4-performance-comparison-old-vs-new)
5. [Example 5: Complex Query with Full Pipeline](#example-5-complex-query-with-full-pipeline)
6. [Example 6: File Metadata Analysis](#example-6-file-metadata-analysis)
7. [Example 7: Current Session Deep Dive](#example-7-current-session-deep-dive)

---

## Example 1: Find Recent Tool Errors

### Objective

Find all tool execution errors from the last 10 sessions to identify recurring issues.

### Step 1: Get Session Directory

```javascript
get_session_directory({scope: "project"})
```

**Output**:
```json
{
  "directory": "/home/user/.claude/projects/abc123/sessions",
  "scope": "project",
  "file_count": 660,
  "total_size_bytes": 453000000,
  "oldest_file": "2025-10-01T10:00:00Z",
  "newest_file": "2025-10-27T12:00:00Z"
}
```

**Analysis**: 660 total files, spanning Oct 1-27. We'll select the 10 most recent.

---

### Step 2: Inspect Recent Files

Assuming files are named with timestamps, select 10 most recent:

```javascript
inspect_session_files({
  files: [
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-27-120000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-27-100000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-26-150000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-26-130000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-26-110000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-25-160000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-25-140000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-25-120000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-25-100000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-24-180000.jsonl"
  ],
  include_samples: true
})
```

**Output**:
```json
{
  "files": [
    {
      "path": ".../session-2025-10-27-120000.jsonl",
      "size_bytes": 3145728,
      "line_count": 1234,
      "record_types": {
        "user": 45,
        "assistant": 43,
        "tool_use": 89,
        "tool_result": 87
      },
      "time_range": {
        "earliest": "2025-10-27T12:00:00Z",
        "latest": "2025-10-27T13:30:00Z"
      },
      "samples": [
        {
          "type": "tool_use",
          "timestamp": "2025-10-27T12:15:00Z",
          "preview": "{\"tool_name\":\"Bash\",\"status\":\"error\",\"error\":\"command not found: npm\"}"
        }
      ]
    }
    // ... 9 more files
  ],
  "summary": {
    "total_files": 10,
    "total_size_bytes": 31457280,
    "total_records": 12340
  }
}
```

**Analysis**:
- 10 files contain 12,340 total records
- Samples show npm-related errors
- Total data: ~30MB (instead of 450MB for all files)

---

### Step 3: Execute Query for Errors

```javascript
execute_stage2_query({
  files: [
    ".../session-2025-10-27-120000.jsonl",
    ".../session-2025-10-27-100000.jsonl",
    ".../session-2025-10-26-150000.jsonl",
    ".../session-2025-10-26-130000.jsonl",
    ".../session-2025-10-26-110000.jsonl",
    ".../session-2025-10-25-160000.jsonl",
    ".../session-2025-10-25-140000.jsonl",
    ".../session-2025-10-25-120000.jsonl",
    ".../session-2025-10-25-100000.jsonl",
    ".../session-2025-10-24-180000.jsonl"
  ],
  filter: 'select(.type == "tool_use" and .status == "error")',
  sort: 'sort_by(.timestamp) | reverse',
  limit: 20
})
```

**Output**:
```json
{
  "results": [
    {
      "type": "tool_use",
      "tool_name": "Bash",
      "timestamp": "2025-10-27T12:15:00Z",
      "status": "error",
      "error": "command not found: npm",
      "input": {"command": "npm test"}
    },
    {
      "type": "tool_use",
      "tool_name": "Edit",
      "timestamp": "2025-10-27T11:45:00Z",
      "status": "error",
      "error": "file not found",
      "input": {"file_path": "/nonexistent/file.js"}
    },
    {
      "type": "tool_use",
      "tool_name": "Bash",
      "timestamp": "2025-10-26T15:30:00Z",
      "status": "error",
      "error": "permission denied",
      "input": {"command": "rm -rf /protected"}
    }
    // ... 17 more errors
  ],
  "metadata": {
    "execution_time_ms": 19,
    "files_processed": 10,
    "total_records_scanned": 12340,
    "results_returned": 20,
    "truncated": false
  }
}
```

### Performance

- **Files processed**: 10 (instead of 660)
- **Execution time**: 19ms (instead of 2.5 seconds)
- **Speedup**: 131x faster
- **Data scanned**: 30MB (instead of 450MB)

### Insights

From the error results:
1. **npm errors recurring** - Node.js environment issue
2. **File not found errors** - Path validation needed
3. **Permission errors** - Security policy enforcement working

---

## Example 2: Analyze Token Usage Over Time

### Objective

Track token consumption across the last week to identify usage patterns and optimize costs.

### Step 1: Get Directory and Select Week's Files

```javascript
get_session_directory({scope: "project"})
```

Based on output (newest_file: 2025-10-27, oldest_file: 2025-10-01), select files from Oct 20-27:

```javascript
inspect_session_files({
  files: [
    // Oct 27 files
    ".../session-2025-10-27-120000.jsonl",
    ".../session-2025-10-27-100000.jsonl",
    // Oct 26 files
    ".../session-2025-10-26-150000.jsonl",
    ".../session-2025-10-26-130000.jsonl",
    // Oct 25 files
    ".../session-2025-10-25-160000.jsonl",
    // ... (select all files from Oct 20-27)
  ],
  include_samples: false  // No samples needed for token analysis
})
```

**Output**:
```json
{
  "files": [
    {
      "path": ".../session-2025-10-27-120000.jsonl",
      "size_bytes": 3145728,
      "line_count": 1234,
      "record_types": {
        "assistant": 43
      }
      // ... time_range
    }
    // ... more files
  ],
  "summary": {
    "total_files": 25,
    "total_records": 30850
  }
}
```

### Step 2: Query Token Usage

```javascript
execute_stage2_query({
  files: [/* 25 files from Oct 20-27 */],
  filter: 'select(.type == "assistant" and .usage != null)',
  transform: '{date: (.timestamp[0:10]), input_tokens: .usage.input_tokens, output_tokens: .usage.output_tokens}',
  sort: 'sort_by(.date)'
})
```

**Output**:
```json
{
  "results": [
    {"date": "2025-10-20", "input_tokens": 4532, "output_tokens": 1234},
    {"date": "2025-10-20", "input_tokens": 6789, "output_tokens": 2345},
    {"date": "2025-10-21", "input_tokens": 5678, "output_tokens": 1567},
    {"date": "2025-10-21", "input_tokens": 7890, "output_tokens": 2890}
    // ... more entries
  ],
  "metadata": {
    "execution_time_ms": 45,
    "files_processed": 25,
    "total_records_scanned": 30850,
    "results_returned": 1075
  }
}
```

### Post-Processing (in Claude Code)

Claude can then aggregate by date:

```
Daily Totals:
- Oct 20: 234,567 input + 89,123 output = 323,690 tokens
- Oct 21: 198,765 input + 76,543 output = 275,308 tokens
- Oct 22: 345,678 input + 123,456 output = 469,134 tokens
...

Weekly Total: 2,145,890 tokens
Average per Day: 306,556 tokens
Peak Day: Oct 22 (469,134 tokens)
```

### Performance

- **Execution time**: 45ms (25 files)
- **vs. Legacy**: ~3 seconds (all 660 files)
- **Speedup**: 66x faster

---

## Example 3: Track Conversation Topics

### Objective

Identify what topics were discussed in sessions from Oct 25-27 by analyzing user messages.

### Step 1: Select Recent Files

```javascript
get_session_directory({scope: "project"})

inspect_session_files({
  files: [
    // Oct 25-27 sessions (assume 15 files)
  ],
  include_samples: true  // Preview messages
})
```

### Step 2: Query User Messages

```javascript
execute_stage2_query({
  files: [/* 15 files from Oct 25-27 */],
  filter: 'select(.type == "user")',
  transform: '{timestamp: .timestamp, preview: (.content[0:100] + "...")}',
  sort: 'sort_by(.timestamp) | reverse',
  limit: 50
})
```

**Output**:
```json
{
  "results": [
    {
      "timestamp": "2025-10-27T12:30:00Z",
      "preview": "Help me refactor the authentication module to use JWT tokens..."
    },
    {
      "timestamp": "2025-10-27T11:15:00Z",
      "preview": "Review the test coverage for the payment processing code..."
    },
    {
      "timestamp": "2025-10-26T16:45:00Z",
      "preview": "Fix the performance issue in the database query for user listings..."
    }
    // ... 47 more
  ],
  "metadata": {
    "execution_time_ms": 23,
    "files_processed": 15,
    "results_returned": 50
  }
}
```

### Insights

**Topics identified**:
1. Authentication (JWT refactoring)
2. Testing (coverage improvements)
3. Performance (database optimization)
4. Code review

**Use case**: Understanding recent work focus for standup reports or weekly summaries.

---

## Example 4: Performance Comparison (Old vs New)

### Objective

Demonstrate the performance difference between legacy `query` and two-stage architecture.

### Legacy Approach (Removed in v2.1.0)

```javascript
// OLD: query tool (REMOVED)
query({
  resource: "tools",
  filter: {tool_status: "error"},
  scope: "project"
})

// Performance:
// - Processes: All 660 files (450MB)
// - Time: 2,500ms
// - Memory: 450MB peak
```

### Two-Stage Approach (New)

```javascript
// NEW: Two-stage workflow

// Step 1: Get directory (8ms)
get_session_directory({scope: "project"})

// Step 2: Inspect 5 recent files (12ms)
inspect_session_files({
  files: ["file1", "file2", "file3", "file4", "file5"],
  include_samples: false
})

// Step 3: Query those 5 files (19ms)
execute_stage2_query({
  files: ["file1", "file2", "file3", "file4", "file5"],
  filter: 'select(.type == "tool_use" and .status == "error")',
  limit: 10
})

// Performance:
// - Processes: 5 files (~15MB)
// - Time: 8 + 12 + 19 = 39ms
// - Memory: 15MB peak
```

### Performance Comparison Table

| Metric | Legacy | Two-Stage | Improvement |
|--------|--------|-----------|-------------|
| Execution Time | 2,500ms | 39ms | **64x faster** |
| Files Processed | 660 | 5 | **132x fewer** |
| Data Scanned | 450MB | 15MB | **30x less** |
| Memory Usage | 450MB | 15MB | **30x less** |

---

## Example 5: Complex Query with Full Pipeline

### Objective

Find the top 10 most frequently used tools in October, with usage counts and error rates.

### Step 1: Select October Files

```javascript
get_session_directory({scope: "project"})

// Based on output, select all files from October
inspect_session_files({
  files: [/* All Oct files, ~100 files */],
  include_samples: false
})
```

### Step 2: Execute Complex Pipeline

```javascript
execute_stage2_query({
  files: [/* 100 Oct files */],
  filter: 'select(.type == "tool_use")',
  transform: 'group_by(.tool_name) | map({tool: .[0].tool_name, total: length, errors: [.[] | select(.status == "error")] | length}) | map(. + {error_rate: ((.errors / .total * 100) | round)})',
  sort: 'sort_by(.total) | reverse',
  limit: 10
})
```

**Output**:
```json
{
  "results": [
    {"tool": "Read", "total": 2345, "errors": 23, "error_rate": 1},
    {"tool": "Bash", "total": 1876, "errors": 187, "error_rate": 10},
    {"tool": "Edit", "total": 1234, "errors": 62, "error_rate": 5},
    {"tool": "Write", "total": 987, "errors": 10, "error_rate": 1},
    {"tool": "Grep", "total": 876, "errors": 9, "error_rate": 1},
    {"tool": "Glob", "total": 654, "errors": 7, "error_rate": 1},
    {"tool": "Task", "total": 543, "errors": 54, "error_rate": 10},
    {"tool": "WebFetch", "total": 432, "errors": 87, "error_rate": 20},
    {"tool": "TodoWrite", "total": 321, "errors": 3, "error_rate": 1},
    {"tool": "AskUserQuestion", "total": 234, "errors": 0, "error_rate": 0}
  ],
  "metadata": {
    "execution_time_ms": 156,
    "files_processed": 100,
    "total_records_scanned": 123400,
    "results_returned": 10
  }
}
```

### Insights

1. **Most used tool**: Read (2,345 uses, 1% error rate)
2. **Highest error rate**: WebFetch (20% - network issues)
3. **Most reliable**: AskUserQuestion (0% errors)
4. **Bash errors**: 10% error rate suggests command validation needed

### Performance

- **Files**: 100 files (~150MB)
- **Time**: 156ms
- **vs. Legacy**: ~4 seconds
- **Speedup**: 25x faster

---

## Example 6: File Metadata Analysis

### Objective

Understand the structure and content distribution of session files without reading full contents.

### Step 1: Get Directory Overview

```javascript
get_session_directory({scope: "project"})
```

**Output**:
```json
{
  "directory": "/home/user/.claude/projects/abc123/sessions",
  "file_count": 660,
  "total_size_bytes": 453000000,
  "oldest_file": "2025-10-01T10:00:00Z",
  "newest_file": "2025-10-27T12:00:00Z"
}
```

**Quick Math**:
- 660 files over 27 days = ~24 files/day
- 450MB total = ~682KB per file average
- Time span: Oct 1-27 (27 days)

### Step 2: Sample File Inspection

Inspect a representative sample (10 files across time range):

```javascript
inspect_session_files({
  files: [
    ".../session-2025-10-01-100000.jsonl",  // Day 1
    ".../session-2025-10-05-120000.jsonl",  // Day 5
    ".../session-2025-10-10-140000.jsonl",  // Day 10
    ".../session-2025-10-15-160000.jsonl",  // Day 15
    ".../session-2025-10-20-180000.jsonl",  // Day 20
    ".../session-2025-10-25-100000.jsonl",  // Day 25
    ".../session-2025-10-27-120000.jsonl"   // Latest
  ],
  include_samples: true
})
```

**Output**:
```json
{
  "files": [
    {
      "path": ".../session-2025-10-01-100000.jsonl",
      "size_bytes": 512000,
      "line_count": 234,
      "record_types": {
        "user": 12,
        "assistant": 11,
        "tool_use": 34,
        "tool_result": 32
      }
    },
    {
      "path": ".../session-2025-10-27-120000.jsonl",
      "size_bytes": 3145728,
      "line_count": 1234,
      "record_types": {
        "user": 45,
        "assistant": 43,
        "tool_use": 89,
        "tool_result": 87
      }
    }
    // ... more
  ],
  "summary": {
    "total_files": 7,
    "total_size_bytes": 12582912,
    "total_records": 5136
  }
}
```

### Insights

**File size trend**:
- Oct 1: 512KB (small session)
- Oct 27: 3.1MB (large session)
- **Trend**: Sessions growing larger over time

**Record distribution**:
- Early sessions: ~90 records/file
- Recent sessions: ~500 records/file
- **Conclusion**: More complex conversations in recent days

**Use case**: Capacity planning, storage estimation, session complexity analysis

---

## Example 7: Current Session Deep Dive

### Objective

Analyze only the current active session in detail.

### Step 1: Get Current Session Directory

```javascript
get_session_directory({scope: "session"})
```

**Output**:
```json
{
  "directory": "/home/user/.claude/projects/abc123/sessions",
  "scope": "session",
  "file_count": 1,
  "total_size_bytes": 3145728,
  "oldest_file": "2025-10-27T12:00:00Z",
  "newest_file": "2025-10-27T12:00:00Z"
}
```

**Note**: Only 1 file (current session)

### Step 2: Inspect Current Session

```javascript
inspect_session_files({
  files: [".../current-session.jsonl"],
  include_samples: true
})
```

**Output**:
```json
{
  "files": [
    {
      "path": ".../current-session.jsonl",
      "size_bytes": 3145728,
      "line_count": 1234,
      "record_types": {
        "user": 45,
        "assistant": 43,
        "tool_use": 89,
        "tool_result": 87,
        "system": 2
      },
      "time_range": {
        "earliest": "2025-10-27T12:00:00Z",
        "latest": "2025-10-27T14:30:00Z"
      },
      "samples": [
        {
          "type": "user",
          "timestamp": "2025-10-27T12:05:00Z",
          "preview": "Help me create documentation for the two-stage query..."
        },
        {
          "type": "tool_use",
          "timestamp": "2025-10-27T12:10:00Z",
          "preview": "{\"tool_name\":\"Write\",\"status\":\"success\"}"
        }
      ]
    }
  ],
  "summary": {
    "total_files": 1,
    "total_size_bytes": 3145728,
    "total_records": 1234
  }
}
```

### Step 3: Query Current Session Activity

**Find all tools used**:
```javascript
execute_stage2_query({
  files: [".../current-session.jsonl"],
  filter: 'select(.type == "tool_use")',
  transform: 'group_by(.tool_name) | map({tool: .[0].tool_name, count: length})',
  sort: 'sort_by(.count) | reverse'
})
```

**Output**:
```json
{
  "results": [
    {"tool": "Write", "count": 23},
    {"tool": "Read", "count": 18},
    {"tool": "Edit", "count": 12},
    {"tool": "TodoWrite", "count": 5},
    {"tool": "Bash", "count": 3}
  ],
  "metadata": {
    "execution_time_ms": 8,
    "files_processed": 1,
    "results_returned": 5
  }
}
```

### Insights

**Current session summary**:
- Duration: 2.5 hours (12:00 - 14:30)
- Total records: 1,234
- User messages: 45
- Assistant responses: 43
- Tool uses: 89 (87 completed)

**Activity breakdown**:
- Writing files (23 times) - Documentation creation
- Reading files (18 times) - Research
- Editing files (12 times) - Refinement
- Task tracking (5 times) - Organization
- Running commands (3 times) - Verification

**Use case**: Real-time session monitoring, progress tracking

---

## Summary

### When to Use Two-Stage Queries

**Best For**:
- ✅ Analyzing specific time ranges
- ✅ Large result sets requiring file selection
- ✅ Performance-critical queries
- ✅ Complex multi-file analysis

**Not Needed For**:
- ❌ Simple error lookups → Use `query_tool_errors`
- ❌ Token tracking → Use `query_token_usage`
- ❌ Quick conversation flow → Use `query_conversation_flow`

### Performance Summary

| Example | Files | Time | vs Legacy | Speedup |
|---------|-------|------|-----------|---------|
| Example 1 (Errors) | 10 | 19ms | 2.5s | 131x |
| Example 2 (Tokens) | 25 | 45ms | 3.0s | 66x |
| Example 3 (Topics) | 15 | 23ms | 2.8s | 121x |
| Example 5 (Complex) | 100 | 156ms | 4.0s | 25x |
| Example 7 (Current) | 1 | 8ms | 1.5s | 187x |

**Average Speedup**: 106x faster

---

## See Also

- [Two-Stage Query Guide](../guides/two-stage-query-guide.md) - Architecture overview
- [MCP Query Tools Reference](../guides/mcp-query-tools.md) - Complete tool documentation
- [MCP Query Cookbook](mcp-query-cookbook.md) - 25+ query examples

---

**Pro Tip**: Always start with `get_session_directory` to understand your data, then use `inspect_session_files` to verify file contents before executing queries. This workflow ensures optimal performance and accurate results.

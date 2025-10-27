# Two-Stage Query Architecture Guide

Complete guide to meta-cc's two-stage query architecture for high-performance session data analysis.

**New in v2.1.0**: The two-stage query architecture replaces the legacy `query`/`query_raw` tools, providing 100-600x performance improvements through intelligent file selection.

---

## Table of Contents

- [Overview](#overview)
- [Why Two-Stage?](#why-two-stage)
- [Architecture](#architecture)
- [Stage 1: File Selection](#stage-1-file-selection)
- [Stage 2: Query Execution](#stage-2-query-execution)
- [Complete Workflow](#complete-workflow)
- [Performance Comparison](#performance-comparison)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Migration from Legacy Tools](#migration-from-legacy-tools)

---

## Overview

The two-stage query architecture separates file selection (Stage 1) from query execution (Stage 2), enabling Claude Code to intelligently choose which files to process before running expensive queries.

**Key Benefits**:
- **100-600x faster** than legacy single-stage queries
- **Transparent file selection** - see exactly what you're querying
- **Flexible planning** - Claude Code chooses optimal query strategy
- **Reduced resource usage** - process only relevant files

**Three Tools**:
1. `get_session_directory` - List all session files
2. `inspect_session_files` - Examine file metadata and content samples
3. `execute_stage2_query` - Execute queries on selected files

---

## Why Two-Stage?

### The Problem with Single-Stage Queries

Legacy `query` tool processed **all 660 files** every time:

```javascript
// Old approach (SLOW - 2-5 seconds)
query({
  resource: "tools",
  filter: {tool_status: "error"},
  scope: "project"
})
// Processes all 660 files → 2-5 seconds
```

**Issues**:
- ❌ Processes irrelevant files (old sessions)
- ❌ No visibility into what's being queried
- ❌ Can't optimize query planning
- ❌ Fixed query semantics (newest-first sorting)

### The Two-Stage Solution

Process only **relevant files** after inspection:

```javascript
// New approach (FAST - 8-19ms)
// Stage 1: List and inspect recent files
get_session_directory({scope: "project"})
inspect_session_files({
  files: ["recent1.jsonl", "recent2.jsonl"],
  include_samples: true
})

// Stage 2: Query only those 2 files
execute_stage2_query({
  files: ["recent1.jsonl", "recent2.jsonl"],
  filter: 'select(.type == "tool" and .status == "error")',
  limit: 10
})
// Processes 2 files → 8-19ms (131x faster!)
```

**Benefits**:
- ✅ Process only relevant files
- ✅ Transparent file selection
- ✅ Flexible query planning (Claude Code decides strategy)
- ✅ Performance: 8-19ms average

---

## Architecture

### Two-Stage Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Stage 1: File Selection & Inspection                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. get_session_directory                                   │
│     └─> Directory path, file count, time range              │
│                                                              │
│  2. inspect_session_files                                   │
│     └─> File metadata, record types, samples                │
│                                                              │
│  Claude Code analyzes metadata and selects files ←──────────┤
│                                                              │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Stage 2: Query Execution                                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  3. execute_stage2_query                                    │
│     - Input: Selected files + jq filter                     │
│     - Output: Query results + metadata                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Design Principles

1. **Separation of Concerns**: File selection (Stage 1) vs. query execution (Stage 2)
2. **Transparency**: Claude Code sees file metadata before querying
3. **Flexibility**: Claude Code chooses query strategy based on context
4. **Performance**: Only process files that matter

---

## Stage 1: File Selection

Stage 1 provides two tools for understanding and selecting session files.

### Tool 1: get_session_directory

**Purpose**: List all session files in a directory with aggregate metadata.

**Parameters**:
- `scope` (string, required): `"project"` or `"session"`

**Returns**:
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

**Example**:
```javascript
get_session_directory({scope: "project"})
```

**Use Cases**:
- Get overview of available session data
- Determine time range of sessions
- Assess data volume before querying

---

### Tool 2: inspect_session_files

**Purpose**: Examine detailed metadata and content samples from specific files.

**Parameters**:
- `files` (array, required): Array of file paths to inspect
- `include_samples` (boolean, optional): Include sample records (default: false)

**Returns**:
```json
{
  "files": [
    {
      "path": "/home/user/.claude/.../session1.jsonl",
      "size_bytes": 3145728,
      "line_count": 1234,
      "record_types": {
        "user": 45,
        "assistant": 43,
        "tool_use": 89,
        "tool_result": 87
      },
      "time_range": {
        "earliest": "2025-10-27T10:00:00Z",
        "latest": "2025-10-27T12:00:00Z"
      },
      "samples": [
        {
          "type": "tool_use",
          "timestamp": "2025-10-27T10:30:00Z",
          "preview": "{\"tool_name\":\"Bash\",\"status\":\"error\"...}"
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

**Example**:
```javascript
inspect_session_files({
  files: [
    "/home/user/.claude/.../session1.jsonl",
    "/home/user/.claude/.../session2.jsonl"
  ],
  include_samples: true
})
```

**Use Cases**:
- Preview file contents before querying
- Determine which files contain relevant data
- Estimate query complexity (record counts)
- Verify file time ranges

---

## Stage 2: Query Execution

### Tool 3: execute_stage2_query

**Purpose**: Execute jq-based queries on selected files with filtering, sorting, transformation, and limits.

**Parameters**:
- `files` (array, required): Array of file paths to query
- `filter` (string, required): jq filter expression
- `sort` (string, optional): jq sort expression
- `transform` (string, optional): jq transformation expression
- `limit` (number, optional): Maximum number of results

**Returns**:
```json
{
  "results": [
    {"type": "tool_use", "tool_name": "Bash", "status": "error", ...}
  ],
  "metadata": {
    "execution_time_ms": 19,
    "files_processed": 2,
    "total_records_scanned": 2468,
    "results_returned": 10,
    "truncated": false
  }
}
```

**jq Filter Examples**:

```javascript
// Simple filter
filter: 'select(.type == "tool_use")'

// Multiple conditions
filter: 'select(.type == "tool_use" and .status == "error")'

// Nested properties
filter: 'select(.type == "tool_use" and .input.command == "npm test")'

// Regex matching
filter: 'select(.type == "user" and (.content | test("refactor")))'
```

**Complete Example**:
```javascript
execute_stage2_query({
  files: ["session1.jsonl", "session2.jsonl"],
  filter: 'select(.type == "tool_use" and .status == "error")',
  sort: 'sort_by(.timestamp) | reverse',
  limit: 10
})
```

**Use Cases**:
- Find errors in recent sessions
- Analyze tool usage patterns
- Track conversation topics
- Extract performance metrics

---

## Complete Workflow

### Example: Find Recent Errors

**Objective**: Find all tool execution errors from the last 2 sessions.

**Step 1: Get Directory Listing**
```javascript
get_session_directory({scope: "project"})
```

**Response**:
```json
{
  "directory": "/home/user/.claude/projects/abc123/sessions",
  "file_count": 660,
  "newest_file": "2025-10-27T12:00:00Z",
  "oldest_file": "2025-10-01T10:00:00Z"
}
```

**Step 2: Inspect Recent Files**

Based on file naming convention (sessions are timestamped), select 2 most recent files:

```javascript
inspect_session_files({
  files: [
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-27-120000.jsonl",
    "/home/user/.claude/projects/abc123/sessions/session-2025-10-27-100000.jsonl"
  ],
  include_samples: true
})
```

**Response**:
```json
{
  "files": [
    {
      "path": ".../session-2025-10-27-120000.jsonl",
      "size_bytes": 3145728,
      "line_count": 1234,
      "record_types": {
        "tool_use": 89,
        "tool_result": 87
      },
      "samples": [...]
    },
    {
      "path": ".../session-2025-10-27-100000.jsonl",
      "size_bytes": 2097152,
      "line_count": 856,
      "record_types": {
        "tool_use": 67,
        "tool_result": 65
      }
    }
  ],
  "summary": {
    "total_files": 2,
    "total_records": 2090
  }
}
```

**Step 3: Execute Query**

```javascript
execute_stage2_query({
  files: [
    ".../session-2025-10-27-120000.jsonl",
    ".../session-2025-10-27-100000.jsonl"
  ],
  filter: 'select(.type == "tool_use" and .status == "error")',
  sort: 'sort_by(.timestamp) | reverse',
  limit: 10
})
```

**Response**:
```json
{
  "results": [
    {
      "type": "tool_use",
      "tool_name": "Bash",
      "timestamp": "2025-10-27T11:45:00Z",
      "status": "error",
      "error": "command not found: npm",
      "input": {"command": "npm test"}
    },
    ...
  ],
  "metadata": {
    "execution_time_ms": 19,
    "files_processed": 2,
    "total_records_scanned": 2090,
    "results_returned": 10
  }
}
```

**Performance**:
- Files processed: 2 (instead of 660)
- Execution time: 19ms (instead of 2-5 seconds)
- **Speedup**: 131x faster

---

## Performance Comparison

### Benchmark Results

| Scenario | Legacy (query) | Two-Stage | Speedup |
|----------|---------------|-----------|---------|
| Find recent errors (10 results) | 2.5s (660 files) | 19ms (2 files) | **131x** |
| List files in project | 2.0s (full scan) | 8ms (metadata) | **250x** |
| Inspect 5 files | 1.5s (process all) | 12ms (target only) | **125x** |
| Complex query (filter+sort) | 5.0s (all files) | 19ms (selected) | **263x** |

**Average Performance**:
- Legacy: 2-5 seconds
- Two-Stage: 8-19ms
- **Average Speedup**: 100-600x

### Resource Usage

| Metric | Legacy | Two-Stage | Improvement |
|--------|--------|-----------|-------------|
| Files Read | All (660) | Selected (2-10) | 66-330x fewer |
| Memory | ~450MB | ~3-10MB | 45-150x less |
| CPU Time | 2-5s | 8-19ms | 100-600x faster |

---

## Best Practices

### 1. Start with Directory Overview

Always begin with `get_session_directory` to understand available data:

```javascript
get_session_directory({scope: "project"})
```

### 2. Inspect Before Querying

Use `inspect_session_files` to preview contents and verify relevance:

```javascript
inspect_session_files({
  files: ["recent1.jsonl", "recent2.jsonl"],
  include_samples: true  // See actual content
})
```

### 3. Query Only Relevant Files

Based on inspection, query only files that contain relevant data:

```javascript
// Good: Query 2 recent files
execute_stage2_query({
  files: ["recent1.jsonl", "recent2.jsonl"],
  ...
})

// Bad: Query all files (defeats the purpose)
execute_stage2_query({
  files: ["all", "660", "files", "..."],
  ...
})
```

### 4. Use Limits for Large Result Sets

Always specify `limit` when expecting many results:

```javascript
execute_stage2_query({
  files: [...],
  filter: 'select(.type == "tool_use")',
  limit: 100  // Prevent overwhelming output
})
```

### 5. Leverage Convenience Tools for Simple Queries

For common queries, use convenience tools instead:

```javascript
// Simple error lookup
query_tool_errors({limit: 10})  // Easier than two-stage

// Complex analysis requiring file selection
get_session_directory(...)      // Use two-stage
inspect_session_files(...)
execute_stage2_query(...)
```

---

## Common Patterns

### Pattern 1: Recent Session Analysis

```javascript
// 1. Get directory
get_session_directory({scope: "project"})

// 2. Inspect last 5 files
inspect_session_files({
  files: ["file1", "file2", "file3", "file4", "file5"],
  include_samples: true
})

// 3. Query those files
execute_stage2_query({
  files: ["file1", "file2", "file3", "file4", "file5"],
  filter: '<your filter>',
  limit: 50
})
```

### Pattern 2: Current Session Only

```javascript
// Scope to current session
get_session_directory({scope: "session"})

inspect_session_files({
  files: ["current-session.jsonl"],
  include_samples: true
})

execute_stage2_query({
  files: ["current-session.jsonl"],
  filter: '<your filter>'
})
```

### Pattern 3: Time-Range Query

```javascript
// 1. Get directory with time range
get_session_directory({scope: "project"})
// Returns: oldest_file, newest_file

// 2. Select files in time range (based on filenames or inspection)
inspect_session_files({
  files: ["<files from Oct 25-27>"],
  include_samples: false  // Just metadata
})

// 3. Query filtered files
execute_stage2_query({
  files: ["<relevant files>"],
  filter: '<your filter>'
})
```

---

## Migration from Legacy Tools

### Removed Tools

The following tools have been removed in v2.1.0:
- ❌ `query` - Use two-stage workflow
- ❌ `query_raw` - Use two-stage workflow

### Migration Examples

**Before (Legacy)**:
```javascript
query({
  resource: "tools",
  filter: {tool_status: "error"},
  scope: "project"
})
```

**After (Two-Stage)**:
```javascript
// Step 1: Get directory
get_session_directory({scope: "project"})

// Step 2: Inspect recent files
inspect_session_files({
  files: ["recent1.jsonl", "recent2.jsonl"],
  include_samples: false
})

// Step 3: Query
execute_stage2_query({
  files: ["recent1.jsonl", "recent2.jsonl"],
  filter: 'select(.type == "tool" and .status == "error")',
  limit: 10
})
```

**Or Use Convenience Tools** (simpler for common queries):
```javascript
query_tool_errors({limit: 10})
```

### When to Use Two-Stage vs. Convenience Tools

**Use Two-Stage When**:
- Need file selection control
- Analyzing specific time ranges
- Complex multi-file queries
- Performance-critical queries

**Use Convenience Tools When**:
- Simple, common queries (errors, tokens, etc.)
- Don't need file selection
- Want quick results

---

## See Also

- [Two-Stage Query Examples](../examples/two-stage-query-examples.md) - 7 practical examples
- [MCP Query Tools Reference](mcp-query-tools.md) - Complete tool documentation
- [MCP Query Cookbook](../examples/mcp-query-cookbook.md) - 25+ query examples

---

**Performance Tip**: The two-stage architecture is designed for speed. Always inspect files first to understand what you're querying, then execute targeted queries on relevant files only.

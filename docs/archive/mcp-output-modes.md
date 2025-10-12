# MCP Hybrid Output Mode

## Overview

The meta-cc MCP server provides **14 query tools** with a **hybrid output mode** to efficiently handle both small and large query results. Instead of returning all data inline (which can cause context overflow for large results), the system intelligently selects between:

- **Inline Mode**: Returns data directly in the MCP response (≤8KB results)
- **File Reference Mode**: Writes data to a temporary file and returns metadata (>8KB results)

This approach ensures optimal performance and context efficiency while maintaining backward compatibility with existing MCP clients.

## Why Hybrid Output Mode?

### Problem: Large Result Context Overflow

When MCP queries return large datasets (e.g., 5000 tool calls), embedding the full JSON response in the MCP protocol message can:

- Exceed Claude's context window limits
- Cause slow response times
- Waste tokens on redundant data transmission
- Lead to truncation and data loss

### Solution: File References for Large Results

By writing large results to temporary JSONL files and returning only metadata, Claude can:

- Use the Read tool to selectively examine portions of the data
- Use the Grep tool to search for specific patterns
- Process results incrementally without context overhead
- Maintain fast query response times

## Mode Selection Logic

The system automatically selects the output mode based on data size:

```
Size ≤ inline_threshold_bytes  → Inline Mode (data embedded in response)
Size > inline_threshold_bytes  → File Reference Mode (data written to temp file)
```

### Default Threshold: 8KB (8192 bytes)

The 8KB default threshold was chosen because:

- Small enough to avoid context bloat for typical queries
- Large enough to include most common query results inline
- Matches common TCP packet size boundaries
- Provides clear separation between "small" and "large" datasets

**Note**: The threshold is configurable via parameter or environment variable (see [Threshold Configuration](#threshold-configuration)).

### Size Calculation

Size is calculated as the total JSONL byte count (including newlines):

```go
// Example: 50 tool calls, ~80 bytes each
// Total: 50 × 81 = 4050 bytes → Inline Mode (with 8KB threshold)

// Example: 5000 tool calls, ~80 bytes each
// Total: 5000 × 81 = 405,000 bytes → File Reference Mode (with 8KB threshold)
```

## Inline Mode

### Use Cases

- Quick stats queries (session summaries)
- Small result sets (<100 records)
- Single-turn analysis
- Stats-only mode (`stats_only: true`)

### Response Format

```json
{
  "mode": "inline",
  "data": [
    {"Timestamp": "2025-10-06T10:00:00Z", "ToolName": "Read", "Status": "success"},
    {"Timestamp": "2025-10-06T10:01:00Z", "ToolName": "Write", "Status": "success"}
  ]
}
```

### Claude Behavior

Claude receives the data directly and can analyze it immediately without additional tool calls.

### Example Query

```json
{
  "tool": "query_tools",
  "arguments": {
    "tool": "Read",
    "limit": 50,
    "scope": "session"
  }
}
```

Response (inline mode):

```json
{
  "mode": "inline",
  "data": [
    // ... 50 records (~4KB total)
  ]
}
```

## File Reference Mode

### Use Cases

- Large query results (>100 records)
- Full session analysis
- Historical data queries
- Cross-session pattern detection

### Response Format

```json
{
  "mode": "file_ref",
  "file_ref": {
    "path": "/tmp/meta-cc-mcp-abc123-1696598400-query_tools.jsonl",
    "size_bytes": 405000,
    "line_count": 5000,
    "fields": ["Timestamp", "ToolName", "Status", "Duration", "Args"],
    "summary": {
      "record_count": 5000,
      "field_count": 5,
      "sample_record": {"Timestamp": "2025-10-06T10:00:00Z", "ToolName": "Bash"}
    }
  }
}
```

### File Reference Structure

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Absolute path to temporary JSONL file |
| `size_bytes` | int64 | Total file size in bytes |
| `line_count` | int | Number of JSONL records |
| `fields` | array | Detected field names across records |
| `summary` | object | Statistics and sample data |

### Claude Behavior

Claude receives metadata and can:

1. **Use Read tool** to examine the file:
   ```
   Read: /tmp/meta-cc-mcp-abc123-1696598400-query_tools.jsonl
   ```

2. **Use Grep tool** to search patterns:
   ```
   Grep: "Status":"error" in /tmp/meta-cc-mcp-abc123-1696598400-query_tools.jsonl
   ```

3. **Analyze metadata** without reading the full file:
   - "The query returned 5000 tool calls (405KB). Let me analyze the summary..."

### Example Query

```json
{
  "tool": "query_tools",
  "arguments": {
    "scope": "project"
  }
}
```

Response (file_ref mode):

```json
{
  "mode": "file_ref",
  "file_ref": {
    "path": "/tmp/meta-cc-mcp-abc123-1696598400-query_tools.jsonl",
    "size_bytes": 405000,
    "line_count": 5000,
    "fields": ["Timestamp", "ToolName", "Status", "Duration"],
    "summary": {
      "record_count": 5000,
      "tool_distribution": {"Read": 1200, "Write": 800, "Bash": 3000}
    }
  }
}
```

Claude can then use Read to examine specific sections or Grep to find patterns.

## Temporary File Management

### File Naming Convention

```
/tmp/meta-cc-mcp-{session_hash}-{timestamp}-{query_type}.jsonl
```

- `session_hash`: First 8 chars of session ID (for grouping)
- `timestamp`: Unix nanosecond timestamp (for uniqueness)
- `query_type`: MCP tool name (e.g., `query_tools`, `get_stats`)

### Example File Paths

```
/tmp/meta-cc-mcp-abc12345-1696598400123456789-query_tools.jsonl
/tmp/meta-cc-mcp-abc12345-1696598401234567890-query_user_messages.jsonl
/tmp/meta-cc-mcp-xyz98765-1696599500123456789-query_tools.jsonl
```

### File Lifecycle

1. **Creation**: Written atomically during query execution
2. **Retention**: Kept for 7 days by default
3. **Cleanup**: Automatic removal after retention period
4. **Manual Cleanup**: `cleanup_temp_files` MCP tool

### Storage Location

- Linux/macOS: `/tmp/` (respects `$TMPDIR` if set)
- Windows: `%TEMP%` directory

### Retention Policy

**Default: 7 days**

Rationale:
- Long enough for multi-day debugging sessions
- Short enough to prevent disk space issues
- Matches common system temp file cleanup policies

Files older than 7 days are automatically removed on next query execution.

### Manual Cleanup Tool

The `cleanup_temp_files` MCP tool allows manual cleanup:

```json
{
  "tool": "cleanup_temp_files",
  "arguments": {
    "max_age_days": 7
  }
}
```

Response:

```json
{
  "removed_count": 12,
  "freed_bytes": 5242880,
  "files": [
    "/tmp/meta-cc-mcp-abc123-1696598400-query_tools.jsonl",
    "/tmp/meta-cc-mcp-abc123-1696598401-get_stats.jsonl"
  ]
}
```

### Disk Space Management

Estimated storage usage:

| Dataset Size | Records | File Size | Storage Impact |
|--------------|---------|-----------|----------------|
| Small session | 500 | ~50KB | Minimal |
| Medium session | 5000 | ~500KB | Low |
| Large session | 50000 | ~5MB | Moderate |
| Full project | 500000 | ~50MB | High |

**Recommendation**: Run `cleanup_temp_files` manually if working with large projects or low disk space.

## Explicit Mode Override

You can force a specific output mode using the `output_mode` parameter:

### Force Inline Mode

```json
{
  "tool": "query_tools",
  "arguments": {
    "output_mode": "inline"
  }
}
```

Use case: Force inline for large results if you need immediate analysis (not recommended for >50KB).

### Force File Reference Mode

```json
{
  "tool": "query_tools",
  "arguments": {
    "output_mode": "file_ref"
  }
}
```

Use case: Force file reference for small results if you want to test file-based workflows.

### Legacy Mode (Backward Compatibility)

```json
{
  "tool": "query_tools",
  "arguments": {
    "output_mode": "legacy"
  }
}
```

Returns raw data array without mode wrapper (for existing clients that expect the old format).

## Threshold Configuration

By default, hybrid mode switches to `file_ref` at 8KB. You can customize this threshold using two methods:

### Per-Query Configuration (Parameter)

Use the `inline_threshold_bytes` parameter to set the threshold for a specific query:

```json
{
  "tool": "query_tools",
  "arguments": {
    "inline_threshold_bytes": 16384  // 16KB threshold
  }
}
```

**Example Use Cases**:
- Increase threshold for queries that typically return 10-15KB
- Decrease threshold to force file_ref mode for better token efficiency

### Global Configuration (Environment Variable)

Set the `META_CC_INLINE_THRESHOLD` environment variable to change the default threshold globally:

```bash
export META_CC_INLINE_THRESHOLD=16384  # 16KB threshold
```

**Example Use Cases**:
- Configure MCP server to prefer inline mode for larger result sets
- Adjust threshold based on token budget constraints

### Configuration Priority

The threshold is determined in the following order (highest to lowest priority):

1. **Parameter**: `inline_threshold_bytes` in query parameters
2. **Environment**: `META_CC_INLINE_THRESHOLD` environment variable
3. **Default**: 8192 bytes (8KB)

**Example**:

```bash
# Set environment variable
export META_CC_INLINE_THRESHOLD=16384  # 16KB

# This query uses 16KB threshold (from environment)
query_tools(status="error")

# This query uses 4KB threshold (parameter overrides environment)
query_tools(status="error", inline_threshold_bytes=4096)
```

## No Truncation Policy

**Phase 16.6 removed all truncation logic.** The MCP server guarantees:

- ✅ **Inline mode**: Complete data returned (no size limits enforced)
- ✅ **File ref mode**: Complete data written to temp file (no truncation)
- ✅ **No warnings**: No `[OUTPUT TRUNCATED]` messages
- ✅ **Information integrity**: All query results are preserved in full

### Removed Parameters

The following parameters were **removed in Phase 16.6**:
- `max_output_bytes` - Replaced by hybrid output mode with configurable threshold

### Migration from max_output_bytes

If you were using `max_output_bytes` to control output size, the new behavior is different and better:

**Old behavior** (pre-Phase 16.6):

```json
{
  "max_output_bytes": 51200  // Truncate at 50KB, lose data
}
```

**New behavior** (Phase 16.6+):

```json
{
  "inline_threshold_bytes": 8192  // Switch to file_ref at 8KB, preserve all data
}
```

**Key Differences**:

| Aspect | Old (`max_output_bytes`) | New (`inline_threshold_bytes`) |
|--------|--------------------------|--------------------------------|
| **Data handling** | Truncates data at limit | Preserves all data |
| **Output mode** | Always inline | Auto-switches to file_ref |
| **Information loss** | ❌ Yes (`[OUTPUT TRUNCATED]`) | ✅ No (complete data in temp file) |
| **Use case** | Limit context size | Control mode switching threshold |

**Migration Strategy**:
- **Don't use `inline_threshold_bytes` to limit data** - it doesn't truncate
- **Trust hybrid mode** - it automatically handles large results via file_ref
- **Use `limit` parameter** - if you need fewer records (e.g., `"limit": 100`)
- **Use `jq_filter`** - if you need specific fields (e.g., `'jq_filter': '.[] | {Timestamp, ToolName}'`)

## Integration with Output Control Features

Hybrid output mode works seamlessly with other output control features:

### stats_only

Stats-only mode bypasses hybrid output (always inline):

```json
{
  "tool": "query_tools",
  "arguments": {
    "stats_only": true
  }
}
```

Response:

```json
{
  "total_records": 5000,
  "tool_distribution": {"Read": 1200, "Write": 800, "Bash": 3000},
  "success_rate": 0.95
}
```

### stats_first

Stats-first mode combines stats + hybrid output:

```json
{
  "tool": "query_tools",
  "arguments": {
    "stats_first": true
  }
}
```

Response:

```
[STATS]
{
  "total_records": 5000,
  "success_rate": 0.95
}
---
[DATA]
{
  "mode": "file_ref",
  "file_ref": {...}
}
```

### jq_filter

JQ filters are applied before mode selection:

```json
{
  "tool": "query_tools",
  "arguments": {
    "jq_filter": ".[] | select(.Status == \"error\")"
  }
}
```

Behavior:
1. Filter applied to raw data
2. Size calculated on filtered result
3. Mode selected based on filtered size

## Performance Characteristics

### Benchmarks

| Operation | Target | Actual | Status |
|-----------|--------|--------|--------|
| File write (100KB) | <200ms | ~2.8ms | ✅ 70x faster |
| Mode selection | <1ms | <0.001ms | ✅ 1000x faster |
| File reference generation | <50ms | ~10ms | ✅ 5x faster |
| Cleanup scan (1000 files) | <500ms | ~150ms | ✅ 3x faster |

### Performance Best Practices

1. **Use inline mode for quick queries**: Stats-only or small result sets
2. **Use file_ref mode for analysis**: Large datasets requiring pattern detection
3. **Clean up regularly**: Run `cleanup_temp_files` weekly for active projects
4. **Leverage metadata**: Analyze file_ref summary before reading full file
5. **Use Grep for large files**: Search patterns without loading full dataset

### Concurrency Safety

The file manager uses mutex locks to ensure:

- No race conditions during concurrent writes
- Unique file paths (nanosecond timestamp)
- Atomic file creation (temp + rename)

Tested with 10 concurrent queries (see `TestMultipleQueriesConcurrent`).

## Troubleshooting

### Issue: "File not found" error

**Symptom**: Claude tries to read temp file but gets file not found error.

**Cause**: File was cleaned up or session hash changed.

**Solution**:
1. Re-run the query to regenerate the file
2. Check retention period (default 7 days)
3. Verify `/tmp` directory permissions

### Issue: Inline mode for large results

**Symptom**: Expected file_ref mode but got inline mode.

**Cause**: Result size is below the `inline_threshold_bytes` (default 8KB).

**Solution**:
1. Decrease threshold: `{"inline_threshold_bytes": 4096}` (4KB)
2. Or force file_ref mode: `{"output_mode": "file_ref"}`
3. Check actual data size to verify it exceeds threshold

### Issue: Disk space exhaustion

**Symptom**: "No space left on device" error when writing temp files.

**Cause**: Too many large temp files accumulated.

**Solution**:
1. Run manual cleanup: `cleanup_temp_files(max_age_days: 1)`
2. Check disk space: `df -h /tmp`
3. Consider using `output_mode: inline` for smaller queries

### Issue: Permission denied on /tmp

**Symptom**: "Permission denied" when writing temp files.

**Cause**: Restricted `/tmp` permissions or SELinux policies.

**Solution**:
1. Check `/tmp` permissions: `ls -ld /tmp` (should be `drwxrwxrwt`)
2. Set `TMPDIR` environment variable to writable location
3. Verify user has write access: `touch /tmp/test && rm /tmp/test`

### Issue: File corruption

**Symptom**: Invalid JSONL format when reading temp file.

**Cause**: Write was interrupted or concurrent write conflict.

**Solution**:
1. Re-run the query to regenerate the file
2. Check for disk I/O errors: `dmesg | grep -i error`
3. Verify atomic writes are working (file manager uses temp + rename)

## Migration Guide

### From Phase 15 (Pre-Hybrid Output)

Existing MCP clients expecting raw data arrays can use legacy mode:

**Before (Phase 15)**:

```json
// Response
[
  {"tool": "Read", "status": "success"},
  {"tool": "Write", "status": "success"}
]
```

**After (Phase 16, default)**:

```json
// Response
{
  "mode": "inline",
  "data": [
    {"tool": "Read", "status": "success"},
    {"tool": "Write", "status": "success"}
  ]
}
```

**Legacy compatibility**:

```json
// Request with output_mode=legacy
{
  "tool": "query_tools",
  "arguments": {
    "output_mode": "legacy"
  }
}

// Response (raw array)
[
  {"tool": "Read", "status": "success"},
  {"tool": "Write", "status": "success"}
]
```

### From Truncation Parameters to Hybrid Mode

**Background**: Phase 15 introduced `max_message_length` and `content_summary` to handle large user messages (10.7k tokens). Phase 16 introduced hybrid output mode, making truncation unnecessary.

**Before (Phase 15 - Truncation Approach)**:

```json
// Request with truncation
{
  "tool": "query_user_messages",
  "arguments": {
    "pattern": "error",
    "max_message_length": 500  // Truncate to 500 chars
  }
}

// Response: Data is truncated (information loss)
{
  "mode": "inline",
  "data": [
    {"content": "This is a long error message that gets trunca..."}
  ]
}
```

**After (Phase 16+ - Hybrid Mode, Default)**:

```json
// Request without truncation (default)
{
  "tool": "query_user_messages",
  "arguments": {
    "pattern": "error"
    // No max_message_length needed - hybrid mode handles large results
  }
}

// Small results: inline mode
{
  "mode": "inline",
  "data": [
    {"content": "Full content preserved"}
  ]
}

// Large results: file_ref mode (no truncation)
{
  "mode": "file_ref",
  "file_ref": {
    "path": "/tmp/meta-cc-mcp-query_user_messages-123.jsonl",
    "size_bytes": 45678,
    "line_count": 121,
    "summary": {"preview": "...", "record_count": 121}
  }
}
```

**Migration Checklist**:

- ✅ **Remove `max_message_length` parameter** - Default is now 0 (no truncation)
- ✅ **Remove `content_summary` parameter** - Hybrid mode provides better information preservation
- ✅ **No code changes needed** - Backward compatible (parameters still accepted but ignored by default)
- ✅ **Better information integrity** - Complete data preserved in file_ref mode
- ✅ **Use Read/Grep tools** - Process large results incrementally from file_ref

**Key Benefits**:

| Approach | Information Loss | Token Efficiency | Flexibility |
|----------|------------------|------------------|-------------|
| Phase 15 Truncation | ❌ Data truncated | ✅ Inline only | ❌ Fixed 500 chars |
| Phase 16 Hybrid Mode | ✅ Complete data | ✅ Adaptive (inline/file_ref) | ✅ Read/Grep on demand |

## Available MCP Tools (14 Total)

The meta-cc MCP server provides the following query tools:

### Message Queries
- `query_user_messages` - Search user messages with regex patterns
- `query_assistant_messages` - Search assistant response messages with regex patterns
- `query_conversation` - Search conversation messages (user + assistant) with optional role filter

### Tool Queries
- `query_tools` - Filter tool calls by name, status (error/success)
- `query_tools_advanced` - SQL-like filtering expressions
- `query_tool_sequences` - Workflow pattern detection

### Analysis Queries
- `query_context` - Error context with surrounding tool calls
- `query_file_access` - File operation history
- `query_files` - File-level operation statistics
- `query_project_state` - Project evolution tracking
- `query_successful_prompts` - High-quality prompt patterns
- `query_time_series` - Metrics over time (hourly/daily/weekly)

### Stats & Utilities
- `get_session_stats` - Session statistics and metrics
- `cleanup_temp_files` - Remove old temporary files

All query tools support hybrid output mode with configurable thresholds.

## See Also

- [Examples & Usage](examples-usage.md) - Practical hybrid output examples
- [Integration Guide](integration-guide.md) - MCP server setup
- [Technical Proposal](proposals/meta-cognition-proposal.md) - Architecture details
- [Phase 15 Plan](../plans/15/plan.md) - Output control features
- [Phase 16 Plan](../plans/16/plan.md) - Hybrid output implementation
- [Phase 19 Plan](../plans/19/plan.md) - Assistant message and conversation query features

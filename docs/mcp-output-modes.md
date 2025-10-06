# MCP Hybrid Output Mode

## Overview

The meta-cc MCP server implements a **hybrid output mode** to efficiently handle both small and large query results. Instead of returning all data inline (which can cause context overflow for large results), the system intelligently selects between:

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
Size ≤ 8KB  → Inline Mode (data embedded in response)
Size > 8KB  → File Reference Mode (data written to temp file)
```

### Threshold: 8KB (8192 bytes)

The 8KB threshold was chosen because:

- Small enough to avoid context bloat for typical queries
- Large enough to include most common query results inline
- Matches common TCP packet size boundaries
- Provides clear separation between "small" and "large" datasets

### Size Calculation

Size is calculated as the total JSONL byte count (including newlines):

```go
// Example: 50 tool calls, ~80 bytes each
// Total: 50 × 81 = 4050 bytes → Inline Mode

// Example: 5000 tool calls, ~80 bytes each
// Total: 5000 × 81 = 405,000 bytes → File Reference Mode
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

## Integration with Phase 15 Output Control

Hybrid output mode works seamlessly with Phase 15 features:

### max_output_bytes

If `max_output_bytes` causes data truncation, the system forces inline mode:

```json
{
  "tool": "query_tools",
  "arguments": {
    "max_output_bytes": 10000
  }
}
```

Behavior:
- Data truncated to fit within 10KB
- Mode forced to "inline" (since data is now small)
- Truncation warning included in response

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

**Cause**: `max_output_bytes` parameter truncated data, forcing inline mode.

**Solution**:
1. Increase `max_output_bytes`: `{"max_output_bytes": 10485760}` (10MB)
2. Or remove parameter to use default (50KB)

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

## See Also

- [Examples & Usage](examples-usage.md) - Practical hybrid output examples
- [Integration Guide](integration-guide.md) - MCP server setup
- [Technical Proposal](proposals/meta-cognition-proposal.md) - Architecture details
- [Phase 15 Plan](../plans/15/plan.md) - Output control features
- [Phase 16 Plan](../plans/16/plan.md) - Hybrid output implementation

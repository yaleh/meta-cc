# Phase 14 Migration Guide

## Overview

Phase 14 standardizes the meta-cc tool interface by simplifying commands and delegating aggregation/filtering logic to external tools like `jq`. This guide helps users migrate from deprecated commands to the new standardized interface.

## Deprecated: `analyze errors`

The `analyze errors` command has been deprecated in Phase 14. Use `query errors` with `jq` for filtering and aggregation instead.

### Why the Change?

**Separation of Concerns**:
- meta-cc focuses on data extraction (pure data processing, no aggregation)
- `jq` handles filtering and aggregation (Unix philosophy: do one thing well)

**Flexibility**:
- `jq` allows arbitrary filtering logic without adding complexity to meta-cc
- Users can compose their own analysis pipelines

**Simplicity**:
- Removes window parameter complexity and aggregation logic from meta-cc
- Simpler command interface with predictable output

## Migration Examples

### Basic Error Listing

**Before (Phase 13)**:
```bash
# Get all errors
meta-cc analyze errors
```

**After (Phase 14)**:
```bash
# Get all errors (simple JSONL list)
meta-cc query errors
```

### Filtering Last N Errors

**Before (Phase 13)**:
```bash
# Analyze errors in last 50 turns
meta-cc analyze errors --window 50
```

**After (Phase 14)**:
```bash
# Get last 50 errors with jq
meta-cc query errors | jq -s '.[-50:]'

# Or use limit/offset flags
meta-cc query errors --limit 50 --offset 0
```

### Grouping by Error Signature

**Before (Phase 13)**:
```bash
# analyze errors automatically grouped by signature
meta-cc analyze errors
```

**After (Phase 14)**:
```bash
# Group by signature with jq
meta-cc query errors | jq -s 'group_by(.signature)'

# Count errors by signature
meta-cc query errors | jq -s 'group_by(.signature) | map({sig: .[0].signature, count: length})'
```

### Counting Errors by Tool

**Before (Phase 13)**:
```bash
# Not directly supported
meta-cc analyze errors | # manual processing
```

**After (Phase 14)**:
```bash
# Count errors by tool with jq
meta-cc query errors | jq -s 'group_by(.tool_name) | map({tool: .[0].tool_name, count: length})'

# Sort by error count descending
meta-cc query errors | jq -s 'group_by(.tool_name) | map({tool: .[0].tool_name, count: length}) | sort_by(.count) | reverse'
```

## Error Signature Changes

### Simplified Signature Format

**Phase 13 (Old)**:
- Used SHA256 hash of error message
- Example: `Bash:a3f8d9e2c1b5...` (64-char hex string)

**Phase 14 (New)**:
- Uses `{tool}:{error_prefix}` format
- First 50 characters of error message
- Whitespace normalized
- Example: `Bash:command not found: xyz`

**Migration Impact**:
- Error signatures are now human-readable
- No need to lookup hash values
- Easier to understand error patterns at a glance

## MCP Server Updates

### MCP Tool Deprecation

The MCP `analyze_errors` tool is deprecated but still functional for backward compatibility.

**Old Way** (still works, but deprecated):
```json
{
  "name": "analyze_errors",
  "arguments": {
    "scope": "project"
  }
}
```

**New Way** (recommended):
```json
{
  "name": "query_tools",
  "arguments": {
    "scope": "project",
    "status": "error"
  }
}
```

Or use `query errors` command directly:
```bash
meta-cc query errors --project .
```

## jq Quick Reference

For users new to `jq`, here are common patterns:

```bash
# Array to stream (remove outer brackets)
meta-cc query errors | jq -s '.'

# Filter by tool
meta-cc query errors | jq 'select(.tool_name == "Bash")'

# Filter by error message content
meta-cc query errors | jq 'select(.error | contains("not found"))'

# Get unique error signatures
meta-cc query errors | jq -s 'map(.signature) | unique'

# Count total errors
meta-cc query errors | jq -s 'length'

# Get most recent error
meta-cc query errors | jq -s '.[-1]'

# Format as table (requires jq @tsv)
meta-cc query errors | jq -r '[.timestamp, .tool_name, .signature] | @tsv'
```

## Timeline

- **Phase 14.2** (Current): `analyze errors` deprecated, `query errors` standardized
- **Phase 15** (Planned): Remove `analyze errors` command entirely
- **Phase 16+**: Continue standardizing other commands using same pattern

## Getting Help

If you encounter migration issues:

1. Check the examples above for common patterns
2. Read `meta-cc query errors --help` for full command documentation
3. Consult `jq` documentation: https://stedolan.github.io/jq/manual/
4. Report issues on GitHub with the `migration` label

## Summary

**Key Takeaway**: Replace `analyze errors` with `query errors | jq` for all error analysis workflows.

The new approach provides:
- ✅ More flexibility (arbitrary jq filters)
- ✅ Human-readable error signatures
- ✅ Simpler meta-cc command interface
- ✅ Better Unix philosophy compliance (composition over complexity)

# Migration to Unified Query API

## Overview

This guide helps you migrate from the current **16 specialized MCP tools** to the new **unified `query` tool**. The migration is designed to be gradual, with full backward compatibility during the transition period.

### Why Migrate?

**Current challenges** (16 tools):
- 80+ parameters to learn
- Inconsistent naming (3 different styles)
- No query composition
- Duplicated functionality

**Benefits of unified query** (1 tool):
- 94% fewer tools (16 → 1)
- 75% fewer parameters (80 → 20)
- Consistent snake_case schema
- Unlimited query composition
- Easier to learn and use

---

## Migration Strategy

### Recommended Approach: Gradual Migration

**Phase 1: Learning** (Week 1)
- Read [Unified Query API Guide](unified-query-api.md)
- Try basic queries alongside old tools
- Compare outputs for equivalence

**Phase 2: Parallel Usage** (Weeks 2-4)
- Use unified query for new workflows
- Keep old tools for existing scripts
- Gradually convert scripts

**Phase 3: Complete Migration** (Weeks 5-8)
- All queries use unified API
- Remove old tool dependencies
- Update documentation

### Aggressive Approach: One-Time Migration

If you prefer to migrate all at once:

1. **Audit current usage**:
   ```bash
   # Find all MCP tool calls in your codebase
   grep -r "query_tools\|query_user_messages" .
   ```

2. **Use mapping table** (see below) to convert each call

3. **Test thoroughly** with real project data

4. **Update documentation** and team training

---

## Tool Mapping Table

Complete mapping from old tools to unified query:

| Old Tool | Unified Query Equivalent |
|----------|-------------------------|
| `query_tools` | `query({resource: "tools", filter: {...}})` |
| `query_user_messages` | `query({resource: "messages", filter: {role: "user", content_match: "..."}})` |
| `query_assistant_messages` | `query({resource: "messages", filter: {role: "assistant"}})` |
| `query_conversation` | `query({resource: "messages"})` |
| `query_files` | `query({resource: "tools", filter: {tool_name: "Read\|Edit\|Write"}, aggregate: {function: "count", field: "file_path"}})` |
| `query_context` | *Use combined queries* (see examples) |
| `query_tool_sequences` | *Use transform + aggregate* (see examples) |
| `query_file_access` | `query({resource: "tools", filter: {file_path: "..."}})` |
| `query_project_state` | `query({resource: "entries", scope: "project"})` |
| `query_successful_prompts` | `query({resource: "messages", filter: {role: "user", quality_score: ">0.8"}})` |
| `query_tools_advanced` | `query({resource: "tools", filter: {...}})` |
| `query_time_series` | `query({resource: "tools", transform: {group_by: "timestamp"}})` |
| `get_session_stats` | `query({resource: "entries", aggregate: {function: "count"}})` |

---

## Migration Examples

### Example 1: query_tools

**Before**:
```javascript
query_tools({
  tool: "Read",
  status: "error",
  limit: 10,
  scope: "project"
})
```

**After**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Read",
    tool_status: "error"
  },
  output: {
    limit: 10
  },
  scope: "project"
})
```

**Key changes**:
- `tool` → `filter.tool_name`
- `status` → `filter.tool_status`
- `limit` → `output.limit`

---

### Example 2: query_user_messages

**Before**:
```javascript
query_user_messages({
  pattern: "error|bug|fix",
  limit: 5,
  max_message_length: 500
})
```

**After**:
```javascript
query({
  resource: "messages",
  filter: {
    role: "user",
    content_match: "error|bug|fix"
  },
  output: {
    limit: 5
  }
})
```

**Key changes**:
- Explicit `resource: "messages"`
- `pattern` → `filter.content_match`
- `max_message_length` removed (hybrid mode handles large data)
- `role: "user"` filter added

---

### Example 3: query_files

**Before**:
```javascript
query_files({
  threshold: 5,
  jq_filter: ".[] | {file: .FilePath, ops: .TotalOps}"
})
```

**After**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Read|Edit|Write"
  },
  aggregate: {
    function: "count",
    field: "input.file_path"
  },
  jq_filter: ".[] | select(.count >= 5) | {file: .file_path, ops: .count}"
})
```

**Key changes**:
- Use `resource: "tools"` with file operation filter
- Aggregate by `file_path`
- Apply threshold in jq_filter

---

### Example 4: query_context

**Before**:
```javascript
query_context({
  error_signature: "Bash:command not found",
  window: 3
})
```

**After** (requires multiple queries):
```javascript
// Step 1: Find error occurrences
const errors = query({
  resource: "tools",
  filter: {
    tool_name: "Bash",
    error: "command not found"
  }
})

// Step 2: For each error, get context window
const contexts = errors.map(error => {
  const turnSeq = error.turn_sequence
  return query({
    resource: "entries",
    filter: {
      turn_sequence: {
        min: turnSeq - 3,
        max: turnSeq + 3
      }
    }
  })
})
```

**Key changes**:
- Context queries now explicit (more flexible)
- Use turn sequence ranges for context window

---

### Example 5: query_tool_sequences

**Before**:
```javascript
query_tool_sequences({
  pattern: "Read -> Edit",
  min_occurrences: 3
})
```

**After**:
```javascript
query({
  resource: "tools",
  transform: {
    extract: ["tool_name", "timestamp"],
    group_by: "session_id"
  },
  jq_filter: `
    map(.tools) |
    map(. as $tools |
      range(0; length-1) |
      $tools[.] + " -> " + $tools[.+1]
    ) |
    flatten |
    group_by(.) |
    map({sequence: .[0], count: length}) |
    map(select(.count >= 3)) |
    sort_by(.count) |
    reverse
  `
})
```

**Key changes**:
- Use transform to extract tool sequences
- Apply pattern detection in jq_filter
- More flexible pattern matching

---

### Example 6: query_assistant_messages

**Before**:
```javascript
query_assistant_messages({
  pattern: "test.*passed",
  min_length: 100,
  limit: 5
})
```

**After**:
```javascript
query({
  resource: "messages",
  filter: {
    role: "assistant",
    content_match: "test.*passed",
    min_length: 100
  },
  output: {
    limit: 5
  }
})
```

**Key changes**:
- `pattern` → `filter.content_match`
- `min_length` → `filter.min_length`
- Explicit `role: "assistant"`

---

### Example 7: query_conversation

**Before**:
```javascript
query_conversation({
  pattern: "refactor",
  pattern_target: "user",
  start_turn: 100,
  end_turn: 150
})
```

**After**:
```javascript
query({
  resource: "messages",
  filter: {
    role: "user",
    content_match: "refactor",
    turn_range: {
      min: 100,
      max: 150
    }
  }
})
```

**Key changes**:
- `pattern_target: "user"` → `filter.role: "user"`
- `start_turn`/`end_turn` → `filter.turn_range`

---

### Example 8: query_tools_advanced

**Before**:
```javascript
query_tools_advanced({
  where: "tool='Bash' AND status='error' AND duration>5000"
})
```

**After**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Bash",
    tool_status: "error",
    duration_min: 5000
  }
})
```

**Key changes**:
- SQL-like `where` clause → structured `filter`
- Field-specific filters (e.g., `duration_min`)

---

### Example 9: query_time_series

**Before**:
```javascript
query_time_series({
  interval: "hour",
  metric: "error-rate",
  where: "tool='Bash'"
})
```

**After**:
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Bash"
  },
  transform: {
    group_by: "hour(timestamp)"
  },
  aggregate: {
    function: "error_rate"
  }
})
```

**Key changes**:
- `interval` → `transform.group_by` with time function
- `metric` → `aggregate.function`

---

### Example 10: get_session_stats

**Before**:
```javascript
get_session_stats({
  stats_only: true
})
```

**After**:
```javascript
query({
  resource: "entries",
  aggregate: {
    function: "stats"
  },
  scope: "session"
})
```

**Key changes**:
- Use `aggregate: {function: "stats"}` for summary
- Explicit `scope: "session"`

---

## Parameter Mapping

Detailed parameter mapping for common fields:

### Filtering Parameters

| Old Parameter | New Location | Example |
|---------------|--------------|---------|
| `tool` | `filter.tool_name` | `{tool_name: "Read"}` |
| `status` | `filter.tool_status` | `{tool_status: "error"}` |
| `pattern` | `filter.content_match` | `{content_match: "regex"}` |
| `file` | `filter.file_path` | `{file_path: "/path/to/file"}` |
| `where` | `filter.*` | Structured filters |
| `session_id` | `filter.session_id` | `{session_id: "abc-123"}` |

### Output Control Parameters

| Old Parameter | New Location | Example |
|---------------|--------------|---------|
| `limit` | `output.limit` | `{limit: 10}` |
| `output_format` | `output.format` | `{format: "jsonl"}` |
| `sort_by` | `output.sort_by` | `{sort_by: "timestamp"}` |
| `max_message_length` | *Removed* | Use hybrid mode |
| `content_summary` | *Removed* | Use hybrid mode |

### Aggregation Parameters

| Old Parameter | New Location | Example |
|---------------|--------------|---------|
| `min_occurrences` | `aggregate.*` + jq filter | See examples |
| `threshold` | `aggregate.*` + jq filter | See examples |
| `min_quality_score` | `filter.quality_score` | `{quality_score: ">0.8"}` |

### Standard Parameters

| Old Parameter | New Location | Notes |
|---------------|--------------|-------|
| `scope` | `scope` | Unchanged |
| `jq_filter` | `jq_filter` | Unchanged |
| `stats_only` | `aggregate.function` | Use `function: "stats"` |
| `stats_first` | `output.stats_first` | Unchanged |

---

## Schema Changes

### Field Name Changes

All fields now use `snake_case`:

| Old Field | New Field |
|-----------|-----------|
| `ToolName` | `tool_name` |
| `UUID` | `uuid` |
| `Timestamp` | `timestamp` |
| `SessionID` | `session_id` |
| `ParentUUID` | `parent_uuid` |
| `GitBranch` | `git_branch` |
| `ToolUseID` | `tool_use_id` |
| `AssistantUUID` | `assistant_uuid` |
| `UserUUID` | `user_uuid` |

### Migration Script

Update your code to use new field names:

```bash
# Find and replace (example for Go code)
sed -i 's/\.ToolName/.tool_name/g' *.go
sed -i 's/\.SessionID/.session_id/g' *.go
sed -i 's/\.GitBranch/.git_branch/g' *.go
```

For JSON/JavaScript:
```bash
# Find and replace in jq filters
sed -i 's/\.ToolName/.tool_name/g' *.sh
sed -i 's/\.Status/.status/g' *.sh
```

---

## Backward Compatibility

### Compatibility Period

Old tools remain available with this timeline:

- **v2.0.0** (current): Both old and new APIs available
- **v2.1.0** (+3 months): Old tools show deprecation warnings
- **v3.0.0** (+6 months): Old tools removed

### Alias Layer

Old tools internally call the unified query:

```go
// Old tool (v2.0.0 - v2.x.x)
func query_tools(args) {
    return query({
        resource: "tools",
        filter: convertArgs(args)
    })
}

// Deprecation warning (v2.1.0+)
func query_tools(args) {
    logWarning("query_tools is deprecated, use query() instead")
    return query({...})
}
```

### Testing Equivalence

Verify that old and new queries return the same results:

```javascript
// Old query
const oldResult = query_tools({tool: "Read", status: "error"})

// New query
const newResult = query({
  resource: "tools",
  filter: {tool_name: "Read", tool_status: "error"}
})

// Compare
assert(oldResult.length === newResult.length)
assert(oldResult[0].tool_name === newResult[0].tool_name)
```

---

## Common Migration Patterns

### Pattern 1: Simple Filter

**Before**: Filter by single field
```javascript
query_tools({tool: "Bash"})
```

**After**: Use structured filter
```javascript
query({resource: "tools", filter: {tool_name: "Bash"}})
```

---

### Pattern 2: Multiple Filters

**Before**: Multiple parameters
```javascript
query_tools({tool: "Bash", status: "error", limit: 10})
```

**After**: Structured filter + output
```javascript
query({
  resource: "tools",
  filter: {tool_name: "Bash", tool_status: "error"},
  output: {limit: 10}
})
```

---

### Pattern 3: Aggregation

**Before**: Use jq_filter for counting
```javascript
query_tools({
  jq_filter: "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})"
})
```

**After**: Use built-in aggregation
```javascript
query({
  resource: "tools",
  aggregate: {function: "count", field: "tool_name"}
})
```

---

### Pattern 4: Regex Matching

**Before**: Pattern in dedicated parameter
```javascript
query_user_messages({pattern: "error|bug"})
```

**After**: Pattern in filter
```javascript
query({
  resource: "messages",
  filter: {role: "user", content_match: "error|bug"}
})
```

---

## Troubleshooting Migration

### Issue 1: Field Not Found

**Error**: `field tool_name not found`

**Cause**: Using old field names (PascalCase)

**Solution**: Update to snake_case
```javascript
// Wrong
filter: {ToolName: "Read"}

// Correct
filter: {tool_name: "Read"}
```

---

### Issue 2: Empty Results

**Error**: Query returns empty array

**Cause**: Incorrect filter syntax or values

**Solution**: Verify filter values
```javascript
// Wrong (case-sensitive)
filter: {tool_name: "bash"}

// Correct
filter: {tool_name: "Bash"}
```

---

### Issue 3: Type Mismatch

**Error**: `invalid type for field duration`

**Cause**: Wrong data type in filter

**Solution**: Use correct types
```javascript
// Wrong
filter: {duration: "5000"}

// Correct
filter: {duration_min: 5000}
```

---

### Issue 4: Aggregation Not Working

**Error**: Aggregate returns unexpected results

**Cause**: Missing or incorrect field specification

**Solution**: Specify aggregation field
```javascript
// Wrong
aggregate: {function: "count"}  // Counts total

// Correct (if you want counts by tool)
aggregate: {function: "count", field: "tool_name"}
```

---

## Migration Checklist

Use this checklist to track your migration progress:

- [ ] Read [Unified Query API Guide](unified-query-api.md)
- [ ] Audit current MCP tool usage
- [ ] Create mapping document for your codebase
- [ ] Test unified queries in development
- [ ] Update scripts and automation
- [ ] Update documentation
- [ ] Train team on new API
- [ ] Monitor for deprecation warnings (v2.1.0+)
- [ ] Remove old tool usage (before v3.0.0)
- [ ] Verify all migrations with real data

---

## Getting Help

### Resources

- [Unified Query API Guide](unified-query-api.md) - Complete API reference
- [Query Cookbook](../examples/query-cookbook.md) - 10+ practical examples
- [MCP Guide](mcp.md) - MCP server documentation

### Support

- **GitHub Issues**: https://github.com/yaleh/meta-cc/issues
- **Discussions**: https://github.com/yaleh/meta-cc/discussions
- **Documentation**: https://github.com/yaleh/meta-cc/tree/main/docs

### Ask Questions

When asking for help, provide:

1. **Old query** (what you're migrating from)
2. **New query attempt** (what you tried)
3. **Expected vs actual results**
4. **Error messages** (if any)

Example:
```
Old: query_tools({tool: "Bash", status: "error"})
New: query({resource: "tools", filter: {tool_name: "Bash", tool_status: "error"}})
Expected: 10 results
Actual: 0 results
Error: None, just empty array
```

---

## Next Steps

After completing migration:

1. **Remove old tool references** from your codebase
2. **Update team documentation** with new query patterns
3. **Share learnings** - contribute examples to the cookbook
4. **Provide feedback** - help improve the unified API

Happy migrating! 🚀

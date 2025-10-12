# Stage 8.8: Enhance MCP Server with Phase 8 Tools

## Overview

**Objective**: Update MCP Server to leverage Phase 8 query capabilities, adding new tools and preventing context overflow.

**Code Estimate**: ~100 lines (Go code, modifying `cmd/mcp.go`)

**Priority**: High (completes MCP integration)

**Time Estimate**: 1-1.5 hours

## Problem Statement

Current MCP Server has 3 tools but lacks Phase 8 capabilities:
- ✅ `get_session_stats` - Already optimal
- ✅ `analyze_errors` - Already optimal
- ❌ `extract_tools` - Uses old `parse extract`, no pagination (causes overflow)
- ❌ Missing `query_tools` - Flexible tool querying not available
- ❌ Missing `query_user_messages` - Message search not available

## Changes Required

### 1. Update `extract_tools` Tool (Prevent Overflow)

**File**: `cmd/mcp.go`

**Current Implementation** (~line 182):
```go
case "extract_tools":
    cmdArgs := []string{"parse", "extract", "--type", "tools", "--output", outputFormat}
```

**Problem**: No pagination, loads ALL tools

**New Implementation**:
```go
case "extract_tools":
    cmdArgs := []string{"query", "tools", "--output", outputFormat}

    // Add default limit to prevent overflow
    if limit, ok := args["limit"].(float64); ok {
        cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
    } else {
        cmdArgs = append(cmdArgs, "--limit", "100")  // Default 100
    }
```

**Update Tool Schema** (~line 128):
```go
{
    "name": "extract_tools",
    "description": "Extract tool usage data from the current session with pagination (Phase 8 enhanced)",
    "inputSchema": map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "limit": map[string]interface{}{
                "type":        "integer",
                "default":     100,
                "description": "Maximum number of tools to extract (default 100, prevents overflow)",
            },
            "output_format": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"json", "md"},
                "default":     "json",
                "description": "Output format (json or md)",
            },
        },
    },
},
```

---

### 2. Add `query_tools` MCP Tool

**File**: `cmd/mcp.go`

**Tool Definition** (in `handleToolsList`, ~line 128):
```go
{
    "name": "query_tools",
    "description": "Query tool calls with flexible filtering and pagination (Phase 8)",
    "inputSchema": map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "tool": map[string]interface{}{
                "type":        "string",
                "description": "Filter by tool name (e.g., 'Bash', 'Read', 'Edit')",
            },
            "status": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"error", "success"},
                "description": "Filter by execution status",
            },
            "limit": map[string]interface{}{
                "type":        "integer",
                "default":     20,
                "description": "Maximum number of results (default 20)",
            },
            "output_format": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"json", "md"},
                "default":     "json",
                "description": "Output format",
            },
        },
    },
},
```

**Tool Implementation** (in `executeTool`, ~line 182):
```go
case "query_tools":
    cmdArgs := []string{"query", "tools", "--output", outputFormat}

    if tool, ok := args["tool"].(string); ok && tool != "" {
        cmdArgs = append(cmdArgs, "--tool", tool)
    }
    if status, ok := args["status"].(string); ok && status != "" {
        cmdArgs = append(cmdArgs, "--status", status)
    }
    if limit, ok := args["limit"].(float64); ok {
        cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
    } else {
        cmdArgs = append(cmdArgs, "--limit", "20")
    }
```

---

### 3. Add `query_user_messages` MCP Tool

**File**: `cmd/mcp.go`

**Tool Definition** (in `handleToolsList`):
```go
{
    "name": "query_user_messages",
    "description": "Search user messages with regex pattern matching (Phase 8)",
    "inputSchema": map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "pattern": map[string]interface{}{
                "type":        "string",
                "description": "Regex pattern to match in message content (required)",
            },
            "limit": map[string]interface{}{
                "type":        "integer",
                "default":     10,
                "description": "Maximum number of results (default 10)",
            },
            "output_format": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"json", "md"},
                "default":     "json",
                "description": "Output format",
            },
        },
        "required": []string{"pattern"},
    },
},
```

**Tool Implementation** (in `executeTool`):
```go
case "query_user_messages":
    pattern, ok := args["pattern"].(string)
    if !ok || pattern == "" {
        return "", fmt.Errorf("pattern parameter is required")
    }

    cmdArgs := []string{"query", "user-messages", "--pattern", pattern, "--output", outputFormat}

    if limit, ok := args["limit"].(float64); ok {
        cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
    } else {
        cmdArgs = append(cmdArgs, "--limit", "10")
    }
```

---

## Implementation Steps

### Step 1: Update `extract_tools` Tool (15 minutes)

1. Update tool schema to add `limit` parameter
2. Update `executeTool` to use `query tools --limit`
3. Test with manual MCP call

**Test Command**:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"extract_tools","arguments":{"limit":50}}}' | ./meta-cc mcp | jq .
```

**Expected**: Returns 50 tools maximum

### Step 2: Add `query_tools` Tool (30 minutes)

1. Add tool definition in `handleToolsList`
2. Add case in `executeTool` switch
3. Handle all parameters (tool, status, limit)
4. Test with various filters

**Test Commands**:
```bash
# Query Bash tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"query_tools","arguments":{"tool":"Bash","limit":5}}}' | ./meta-cc mcp | jq .

# Query errors
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"query_tools","arguments":{"status":"error","limit":10}}}' | ./meta-cc mcp | jq .

# Query Bash errors
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"query_tools","arguments":{"tool":"Bash","status":"error"}}}' | ./meta-cc mcp | jq .
```

### Step 3: Add `query_user_messages` Tool (30 minutes)

1. Add tool definition in `handleToolsList`
2. Add case in `executeTool` switch
3. Validate required `pattern` parameter
4. Test with various patterns

**Test Commands**:
```bash
# Search for "Phase 8"
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":"Phase 8","limit":3}}}' | ./meta-cc mcp | jq .

# Regex search
echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":"error|bug"}}}' | ./meta-cc mcp | jq .

# Missing pattern (should error)
echo '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"query_user_messages","arguments":{}}}' | ./meta-cc mcp | jq .
```

### Step 4: Verify All Tools (15 minutes)

**List all tools**:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc mcp | jq '.result.tools | length'
# Expected: 5 tools
```

**Verify tool names**:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc mcp | jq '.result.tools[].name'
# Expected:
# "get_session_stats"
# "analyze_errors"
# "extract_tools"
# "query_tools"
# "query_user_messages"
```

---

## Testing Strategy

### Unit Test (if time permits)

Create `cmd/mcp_test.go`:
```go
func TestQueryToolsMCP(t *testing.T) {
    // Test query_tools tool definition exists
    // Test executeTool with various parameters
}

func TestQueryUserMessagesMCP(t *testing.T) {
    // Test query_user_messages tool definition
    // Test pattern validation
}
```

### Manual Integration Test

```bash
# Build
go build

# Test all 5 tools
./test-scripts/test-mcp-tools.sh
```

### Expected Behavior

| Tool | Input | Expected Output |
|------|-------|-----------------|
| `extract_tools` | `{"limit": 50}` | Max 50 tools |
| `query_tools` | `{"tool": "Bash", "limit": 5}` | 5 Bash calls |
| `query_tools` | `{"status": "error"}` | Error calls only |
| `query_user_messages` | `{"pattern": "Phase 8"}` | Messages with "Phase 8" |
| `query_user_messages` | `{}` | Error (pattern required) |

---

## Acceptance Criteria

- ✅ `extract_tools` updated to use `query tools --limit`
- ✅ Default limit of 100 prevents overflow
- ✅ `query_tools` tool added with full filtering support
- ✅ `query_user_messages` tool added with regex support
- ✅ All 5 MCP tools work correctly
- ✅ `tools/list` returns 5 tools
- ✅ Parameter validation works (e.g., pattern required)
- ✅ Build succeeds without errors
- ✅ Manual tests pass

---

## Dependencies

- ✅ Stage 8.2 completed (`query tools` command available)
- ✅ Stage 8.3 completed (`query user-messages` command available)
- ✅ `cmd/mcp.go` exists (from Phase 7)
- ✅ MCP protocol implementation working

---

## Code Changes Summary

**File**: `cmd/mcp.go`

**Lines Added**: ~100
- Tool schemas: ~60 lines (3 tools × ~20 lines each)
- Tool execution: ~40 lines (3 cases × ~13 lines each)

**Lines Modified**: ~10
- Update `extract_tools` schema and execution

**Total Impact**: ~110 lines

---

## Benefits

### Context Overflow Prevention
- ✅ `extract_tools` now has default limit (100)
- ✅ No more massive data dumps
- ✅ Better performance in large sessions

### Flexible Querying
- ✅ `query_tools` enables targeted queries
- ✅ Filter by tool, status, or both
- ✅ Control result size with limit

### Message Search
- ✅ `query_user_messages` enables regex search
- ✅ Find past discussions easily
- ✅ Support complex patterns

### Natural Language Integration
- ✅ Claude can now autonomously query tools
- ✅ Claude can search user messages
- ✅ No manual CLI commands required

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Parameter validation errors | Medium | Test all edge cases, add clear error messages |
| MCP protocol changes | Low | Follow JSON-RPC 2.0 spec strictly |
| Build failures | Low | Test incrementally, commit after each tool |
| Backward compatibility | Low | Old tools still work, new tools are additive |

---

## Next Steps

After completing this stage:
1. ✅ All 5 MCP tools working
2. 📋 Proceed to Stage 8.9 (MCP Server configuration)
3. 📋 Test MCP integration in Claude Code
4. 📋 Update documentation

---

## Related Documentation

- Phase 8 Plan: `plans/8/phase.md`
- MCP Gap Analysis: `/tmp/phase8-mcp-gap-analysis.md`
- MCP Implementation (Phase 7): `docs/plan.md` (lines 1022-1140)
- Integration Guide: `docs/integration-guide.md`

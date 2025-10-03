# Stage 8.9: Configure MCP Server to Claude Code

## Overview

**Objective**: Configure MCP Server in Claude Code project and create usage documentation.

**Code Estimate**: ~120 lines (configuration + documentation)

**Priority**: High (enables MCP usage)

**Time Estimate**: 30 minutes

## Problem Statement

MCP Server is implemented but:
- ❌ Not registered in Claude Code
- ❌ No configuration file exists
- ❌ Users don't know how to use it
- ❌ Natural language queries not possible

## Changes Required

### 1. Create MCP Server Configuration

**File**: `.claude/mcp-servers/meta-cc.json`

**Content**:
```json
{
  "command": "./meta-cc",
  "args": ["mcp"],
  "description": "Meta-cognition analysis for Claude Code sessions with Phase 8 query capabilities",
  "env": {},
  "tools": [
    "get_session_stats",
    "analyze_errors",
    "extract_tools",
    "query_tools",
    "query_user_messages"
  ]
}
```

**Notes**:
- Uses relative path `./meta-cc` (assumes binary in project root)
- Can be changed to absolute path if needed
- No environment variables required
- Lists all 5 available tools

---

### 2. Create MCP Usage Documentation

**File**: `docs/mcp-usage.md`

**Content Structure**:

```markdown
# MCP Server Usage Guide

## Overview

meta-cc provides a Model Context Protocol (MCP) Server that allows Claude Code to autonomously query session data without manual CLI commands.

## Configuration

The MCP Server is configured in `.claude/mcp-servers/meta-cc.json`.

**Prerequisites**:
- `meta-cc` binary in project root or PATH
- Claude Code with MCP support

## Available Tools

### 1. get_session_stats
Get comprehensive session statistics.

**Parameters**:
- `output_format` (optional): "json" or "md" (default: "json")

**Example Query**:
```
帮我看一下当前会话的统计信息
```

### 2. analyze_errors
Analyze error patterns in the session.

**Parameters**:
- `window_size` (optional): Number of recent turns to analyze (default: 20)
- `output_format` (optional): "json" or "md"

**Example Query**:
```
分析一下我的错误模式
```

### 3. extract_tools (Phase 8 Enhanced)
Extract tool usage data with pagination.

**Parameters**:
- `limit` (optional): Maximum number of tools (default: 100)
- `output_format` (optional): "json" or "md"

**Example Query**:
```
提取最近 50 个工具调用
```

**Phase 8 Enhancement**: Now uses `query tools --limit` to prevent context overflow.

### 4. query_tools (Phase 8 New)
Query tool calls with flexible filtering.

**Parameters**:
- `tool` (optional): Filter by tool name (e.g., "Bash", "Read", "Edit")
- `status` (optional): Filter by status ("error" or "success")
- `limit` (optional): Maximum results (default: 20)
- `output_format` (optional): "json" or "md"

**Example Queries**:
```
帮我查一下用了多少次 Bash 工具
查找所有 Bash 命令的错误
显示最近 10 个 Edit 工具的调用
```

### 5. query_user_messages (Phase 8 New)
Search user messages with regex patterns.

**Parameters**:
- `pattern` (required): Regex pattern to match
- `limit` (optional): Maximum results (default: 10)
- `output_format` (optional): "json" or "md"

**Example Queries**:
```
搜索我提到 "Phase 8" 的消息
查找包含 "error" 或 "bug" 的消息
找到我说过 "fix.*bug" 的地方
```

## Usage Patterns

### 1. Natural Language (Recommended)

Claude automatically selects the appropriate tool:

```
User: "我的 Bash 命令哪里出问题了？"
Claude: [Automatically calls]
  1. query_tools(tool="Bash", status="error")
  2. analyze_errors(window_size=50)
  3. Provides analysis and recommendations
```

No manual commands needed!

### 2. Direct Tool Invocation

```
使用 mcp__meta-cc__query_tools 查询 Bash 工具的使用情况
```

### 3. Combined Analysis

```
User: "帮我优化工作流"
Claude: [Automatically calls]
  1. get_session_stats() - Overall metrics
  2. query_tools(status="error") - Error patterns
  3. query_user_messages(pattern="优化|improve") - Past optimization attempts
  4. Provides comprehensive recommendations
```

## Best Practices

### 1. Use Natural Language
Let Claude choose the right tool based on context.

**Good**:
```
帮我查一下 Bash 工具的错误
```

**Also Good** (but less natural):
```
使用 query_tools 查询 Bash 错误
```

### 2. Be Specific When Needed
```
查找我最近 20 条提到 "Phase 8" 的消息
→ Claude calls: query_user_messages(pattern="Phase 8", limit=20)
```

### 3. Combine with Slash Commands
- **Slash Commands**: For repeated workflows, predictable outputs
- **MCP Tools**: For exploratory analysis, natural language queries

Example:
```
/meta-stats              # Quick stats (Slash Command)
分析我的工作流               # Deep analysis (MCP + Claude reasoning)
```

### 4. Large Sessions
MCP tools automatically handle pagination:
- `extract_tools`: Default limit 100
- `query_tools`: Default limit 20
- `query_user_messages`: Default limit 10

Claude will make multiple calls if needed.

## Troubleshooting

### MCP Server Not Connected

**Symptom**: Claude can't find MCP tools

**Solution**:
1. Check `meta-cc` binary exists:
   ```bash
   ./meta-cc --version
   ```
2. Verify configuration file:
   ```bash
   cat .claude/mcp-servers/meta-cc.json
   ```
3. Restart Claude Code

### Tool Execution Fails

**Symptom**: MCP tool returns error

**Solution**:
1. Test CLI command manually:
   ```bash
   ./meta-cc query tools --tool Bash --limit 5
   ```
2. Check session file exists:
   ```bash
   ls ~/.claude/projects/-home-*
   ```
3. Verify working directory is project root

### No Results Returned

**Symptom**: Tool runs but returns empty results

**Solution**:
- For `query_tools`: Check tool name spelling (case-sensitive)
- For `query_user_messages`: Verify regex pattern is valid
- Increase limit parameter

## Comparison: MCP vs Slash Commands vs CLI

### When to Use MCP Tools

✅ **Use MCP when**:
- Asking exploratory questions
- Combining multiple queries
- Letting Claude reason about what to query
- Natural language interaction preferred

**Example**:
```
"我的工作流哪里可以优化？"
→ Claude autonomously queries stats, errors, and messages
```

### When to Use Slash Commands

✅ **Use Slash Commands when**:
- Repeated workflows
- Predictable outputs
- Fast execution needed
- Specific command preference

**Example**:
```
/meta-stats
/meta-timeline 50
/meta-query-tools Bash error
```

### When to Use CLI Directly

✅ **Use CLI when**:
- Scripting or automation
- Outside Claude Code
- Debugging meta-cc itself

**Example**:
```bash
meta-cc query tools --tool Bash --status error | jq .
```

## Advanced Usage

### Regex Pattern Examples

| Pattern | Description | Example Query |
|---------|-------------|---------------|
| `Phase 8` | Exact match | "找到 Phase 8 相关消息" |
| `error\|bug` | OR operator | "搜索 error 或 bug" |
| `^Continue` | Start with | "找以 Continue 开头的消息" |
| `test$` | End with | "找以 test 结尾的消息" |
| `fix.*bug` | Between | "找 fix 和 bug 之间的消息" |
| `Phase [0-9]` | Number range | "找所有 Phase 加数字的消息" |

### Filter Combinations

```
# Query Bash errors in last 10 calls
query_tools(tool="Bash", status="error", limit=10)

# Find optimization discussions
query_user_messages(pattern="优化|improve|optimize", limit=20)

# Extract recent file operations
query_tools(tool="Edit|Write|Read", limit=30)
```

## Examples

### Example 1: Debug Workflow Errors

**User**: "为什么我的命令总是失败？"

**Claude** (automatically):
```
1. Calling query_tools(status="error", limit=20)...
   → Found 5 Bash errors

2. Calling analyze_errors(window_size=50)...
   → Detected pattern: "npm test" failing repeatedly

3. Analysis:
   - Error occurs 5 times in 50 turns
   - Same command: "npm test"
   - Suggestion: Check test configuration or add pre-test hook
```

### Example 2: Workflow Optimization

**User**: "帮我优化我的开发流程"

**Claude** (automatically):
```
1. Calling get_session_stats()...
   → 150 turns, 67 tool calls, 5% error rate

2. Calling query_tools(tool="Bash")...
   → Bash used 30 times (most frequent)

3. Calling query_user_messages(pattern="重复|again|再次")...
   → Found 8 messages about repeated tasks

4. Recommendations:
   - Create Slash Command for repeated "npm test"
   - Add Hook for pre-commit checks
   - Use @meta-coach for deeper analysis
```

### Example 3: Message Search

**User**: "我之前提到过 Phase 8 的实现细节吗？"

**Claude** (automatically):
```
1. Calling query_user_messages(pattern="Phase 8.*实现|implement.*Phase 8", limit=5)...

2. Results:
   - Turn 45: "让我们开始 Phase 8 实现..."
   - Turn 67: "Phase 8 的实现应该包括..."
   - Turn 89: "完成 Phase 8 实现后..."

3. Summary: Yes, you discussed Phase 8 implementation 3 times,
   focusing on query commands and integration improvements.
```

## Summary

MCP Server provides:
- ✅ **5 tools** (3 from Phase 7 + 2 from Phase 8)
- ✅ **Natural language queries** (no manual commands)
- ✅ **Autonomous analysis** (Claude decides what to query)
- ✅ **Flexible filtering** (tool, status, pattern, limit)
- ✅ **Context-aware** (automatic pagination)

**Next Steps**:
1. Try natural language queries in Claude Code
2. Explore different query patterns
3. Combine with @meta-coach for deep analysis
4. Create custom Slash Commands for workflows
```

---

## Implementation Steps

### Step 1: Create MCP Configuration Directory (2 minutes)

```bash
mkdir -p .claude/mcp-servers
```

### Step 2: Create Configuration File (3 minutes)

```bash
cat > .claude/mcp-servers/meta-cc.json << 'EOF'
{
  "command": "./meta-cc",
  "args": ["mcp"],
  "description": "Meta-cognition analysis for Claude Code sessions with Phase 8 query capabilities",
  "env": {},
  "tools": [
    "get_session_stats",
    "analyze_errors",
    "extract_tools",
    "query_tools",
    "query_user_messages"
  ]
}
EOF
```

### Step 3: Create Usage Documentation (20 minutes)

```bash
# Create the full docs/mcp-usage.md file
# (Copy the content from above)
```

### Step 4: Verify Configuration (5 minutes)

```bash
# Check file exists
cat .claude/mcp-servers/meta-cc.json

# Verify meta-cc binary works
./meta-cc mcp --help 2>/dev/null || echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | ./meta-cc mcp | jq .
```

---

## Testing Strategy

### Test 1: Configuration File Valid

```bash
# Validate JSON syntax
jq empty .claude/mcp-servers/meta-cc.json && echo "✅ Valid JSON" || echo "❌ Invalid JSON"
```

### Test 2: MCP Server Responds

```bash
# Test initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./meta-cc mcp | jq .

# Test tools/list
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./meta-cc mcp | jq '.result.tools | length'
# Expected: 5
```

### Test 3: Claude Code Integration

**In Claude Code**:
1. Restart Claude Code (to load new MCP configuration)
2. Ask: "列出所有可用的 MCP 工具"
3. Expected: Should see `mcp__meta-cc__*` tools

**Natural Language Test**:
1. Ask: "帮我查一下用了多少次 Bash 工具"
2. Expected: Claude calls `mcp__meta-cc__query_tools` automatically
3. Verify: Response includes Bash tool count

**Message Search Test**:
1. Ask: "搜索我提到 'Phase 8' 的消息"
2. Expected: Claude calls `mcp__meta-cc__query_user_messages`
3. Verify: Returns relevant messages

### Test 4: Documentation Completeness

**Checklist**:
- ✅ All 5 tools documented
- ✅ Usage examples provided
- ✅ Best practices section
- ✅ Troubleshooting guide
- ✅ Comparison table (MCP vs Slash vs CLI)
- ✅ Regex pattern examples
- ✅ Real-world examples

---

## Acceptance Criteria

- ✅ `.claude/mcp-servers/meta-cc.json` created and valid
- ✅ JSON syntax correct
- ✅ All 5 tools listed
- ✅ `docs/mcp-usage.md` created
- ✅ Documentation is comprehensive (>100 lines)
- ✅ MCP Server responds to test queries
- ✅ Claude Code can discover MCP tools
- ✅ Natural language queries work
- ✅ Examples are clear and helpful

---

## Dependencies

- ✅ Stage 8.8 completed (5 MCP tools working)
- ✅ `meta-cc` binary exists in project root
- ✅ MCP Server implementation working (`cmd/mcp.go`)
- ✅ Claude Code with MCP support

---

## Files Created

1. **`.claude/mcp-servers/meta-cc.json`** (~20 lines)
   - MCP Server configuration
   - Tool list
   - Command and args

2. **`docs/mcp-usage.md`** (~100 lines)
   - Complete usage guide
   - All 5 tools documented
   - Examples and best practices
   - Troubleshooting section

**Total**: ~120 lines

---

## Benefits

### Discoverability
- ✅ MCP Server visible in Claude Code
- ✅ Users can find tools easily
- ✅ Documentation provides guidance

### Usability
- ✅ Natural language queries work
- ✅ No CLI knowledge required
- ✅ Clear examples provided

### Integration
- ✅ Seamless Claude Code experience
- ✅ Combines with Slash Commands
- ✅ Works with @meta-coach

### Documentation
- ✅ Comprehensive guide
- ✅ Troubleshooting included
- ✅ Comparison with alternatives

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Configuration file invalid | High | Validate JSON before committing |
| Path to binary incorrect | Medium | Document both relative and absolute paths |
| Claude Code doesn't load MCP | Medium | Provide restart instructions |
| Users don't find documentation | Low | Link from README.md |

---

## Next Steps

After completing this stage:
1. ✅ MCP Server fully configured
2. ✅ Documentation available
3. 📋 Test in real Claude Code session
4. 📋 Update README.md with MCP section
5. 📋 Complete Phase 8 integration testing

---

## Related Documentation

- Phase 8 Plan: `plans/8/phase.md`
- Stage 8.8 Plan: `plans/8/stage-8.8.md`
- MCP Gap Analysis: `/tmp/phase8-mcp-gap-analysis.md`
- Integration Guide: `docs/integration-guide.md`
- Main Plan: `docs/plan.md`

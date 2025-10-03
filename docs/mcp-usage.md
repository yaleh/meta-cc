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

### 1. 自然语言查询 (Natural Language - Recommended)

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
使用 mcp__meta-insight__query_tools 查询 Bash 工具的使用情况
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

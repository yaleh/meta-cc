# MCP Server Usage Guide

## Overview

meta-cc provides a Model Context Protocol (MCP) Server that allows Claude Code to autonomously query session data without manual CLI commands. With Phase 8 enhancements, the MCP server now provides 8 powerful tools for comprehensive session analysis.

## Configuration

The MCP Server is configured in `.claude/mcp-servers/meta-cc.json`.

**Prerequisites**:
- `meta-cc` binary in project root or PATH
- Claude Code with MCP support

**Configuration File**:
```json
{
  "command": "./meta-cc",
  "args": ["mcp"],
  "description": "Meta-cognition analysis for Claude Code sessions",
  "env": {},
  "tools": [
    "get_session_stats",
    "analyze_errors",
    "extract_tools",
    "query_tools",
    "query_user_messages",
    "query_context",
    "query_tool_sequences",
    "query_file_access"
  ]
}
```

## Available Tools

### 1. get_session_stats

Get comprehensive session statistics including turn count, tool usage, and error rates.

**Parameters**:
- `output_format` (optional): "json" or "md" (default: "json")

**Example Queries**:
```
Show me the session statistics
How many tools have I used?
What's my error rate?
```

**Direct Invocation**:
```
mcp__meta-insight__get_session_stats
```

---

### 2. analyze_errors

Analyze error patterns in the session, detecting repeated failures and common issues.

**Parameters**:
- `output_format` (optional): "json" or "md" (default: "json")

**Example Queries**:
```
Analyze my error patterns
What errors keep happening?
Why do my commands fail?
```

**Direct Invocation**:
```
mcp__meta-insight__analyze_errors
```

---

### 3. extract_tools (Phase 8 Enhanced)

Extract tool usage data with pagination to prevent context overflow.

**Parameters**:
- `limit` (optional): Maximum number of tools (default: 100)
- `output_format` (optional): "json" or "md"

**Example Queries**:
```
Extract the last 50 tool calls
Show me recent tool usage
What tools have I been using?
```

**Phase 8 Enhancement**: Now uses `query tools --limit` internally to prevent context overflow.

**Direct Invocation**:
```
mcp__meta-insight__extract_tools
```

---

### 4. query_tools (Phase 8 New)

Query tool calls with flexible filtering by tool name, status, and limit.

**Parameters**:
- `tool` (optional): Filter by tool name (e.g., "Bash", "Read", "Edit")
- `status` (optional): Filter by status ("error" or "success")
- `limit` (optional): Maximum results (default: 20)
- `output_format` (optional): "json" or "md"

**Example Queries**:
```
How many times did I use Bash?
Find all Bash errors
Show the last 10 Edit tool calls
What Read commands succeeded?
```

**Direct Invocation**:
```
mcp__meta-insight__query_tools
```

---

### 5. query_user_messages (Phase 8 New)

Search user messages with regex pattern matching.

**Parameters**:
- `pattern` (required): Regex pattern to match in message content
- `limit` (optional): Maximum results (default: 10)
- `output_format` (optional): "json" or "md"

**Example Queries**:
```
Search for messages about "Phase 8"
Find when I mentioned "bug" or "error"
Look for my "fix.*test" messages
```

**Regex Pattern Examples**:
| Pattern | Description |
|---------|-------------|
| `Phase 8` | Exact match |
| `error\|bug` | OR operator |
| `^Continue` | Start with |
| `test$` | End with |
| `fix.*bug` | Between words |
| `Phase [0-9]` | Number range |

**Direct Invocation**:
```
mcp__meta-insight__query_user_messages
```

---

### 6. query_context (Stage 8.10+)

Query context around specific errors to understand what led to failures.

**Parameters**:
- `error_signature` (required): Error pattern ID to query
- `window` (optional): Context window size in turns before/after (default: 3)
- `output_format` (optional): "json" or "md"

**Example Queries**:
```
Show context around the npm test failure
What happened before the Bash error?
Give me 5 turns of context for error X
```

**Direct Invocation**:
```
mcp__meta-insight__query_context
```

---

### 7. query_tool_sequences (Stage 8.11+)

Query repeated tool call sequences to identify workflow patterns.

**Parameters**:
- `min_occurrences` (optional): Minimum occurrences to report (default: 3)
- `pattern` (optional): Specific sequence pattern (e.g., "Read -> Edit -> Bash")
- `output_format` (optional): "json" or "md"

**Example Queries**:
```
What tool sequences do I repeat?
Find my "Read -> Edit -> Bash" pattern
Show common workflows
```

**Direct Invocation**:
```
mcp__meta-insight__query_tool_sequences
```

---

### 8. query_file_access

Query file access history including read, edit, and write operations.

**Parameters**:
- `file` (required): File path to query
- `output_format` (optional): "json" or "md"

**Example Queries**:
```
How many times did I edit main.go?
Show access history for README.md
What operations on config.json?
```

**Direct Invocation**:
```
mcp__meta-insight__query_file_access
```

---

## Usage Patterns

### 1. Natural Language (Recommended)

Claude automatically selects the appropriate tool based on your question:

```
User: "Why do my Bash commands keep failing?"

Claude: [Automatically calls]
  1. query_tools(tool="Bash", status="error")
  2. analyze_errors()
  3. Provides analysis and recommendations
```

**No manual commands needed!**

### 2. Direct Tool Invocation

If you prefer explicit control:

```
Use mcp__meta-insight__query_tools to find Bash errors
```

### 3. Combined Analysis

Claude can orchestrate multiple tools for deep analysis:

```
User: "Help me optimize my workflow"

Claude: [Automatically calls]
  1. get_session_stats() - Overall metrics
  2. query_tool_sequences(min_occurrences=3) - Repeated patterns
  3. query_tools(status="error") - Error hotspots
  4. query_user_messages(pattern="repeat|again") - User frustrations
  5. Provides comprehensive optimization recommendations
```

## Best Practices

### 1. Use Natural Language

Let Claude choose the right tool based on context.

**Good**:
```
Find my Bash errors
```

**Also Good** (but less natural):
```
Use query_tools with tool=Bash and status=error
```

### 2. Be Specific When Needed

```
Find the last 20 messages about "Phase 8"
→ Claude calls: query_user_messages(pattern="Phase 8", limit=20)
```

### 3. Combine with Slash Commands

- **Slash Commands**: For repeated workflows, predictable outputs
- **MCP Tools**: For exploratory analysis, natural language queries

**Example**:
```
/meta-stats              # Quick stats (Slash Command)
Analyze my workflow      # Deep analysis (MCP + Claude reasoning)
```

### 4. Large Sessions

MCP tools automatically handle pagination:
- `extract_tools`: Default limit 100
- `query_tools`: Default limit 20
- `query_user_messages`: Default limit 10

Claude will make multiple calls if needed.

### 5. Error Investigation

Start broad, then narrow down:

```
1. "Analyze my errors" → analyze_errors()
2. "Show Bash errors" → query_tools(tool="Bash", status="error")
3. "What happened before error X?" → query_context(error_signature="X")
```

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
   jq empty .claude/mcp-servers/meta-cc.json && echo "Valid JSON"
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
- For `query_tools`: Check tool name spelling (case-sensitive: "Bash", not "bash")
- For `query_user_messages`: Verify regex pattern is valid
- For `query_context`: Ensure error_signature exists
- For `query_file_access`: Check file path is exact
- Increase limit parameter

### Pattern Matching Issues

**Symptom**: `query_user_messages` returns unexpected results

**Solution**:
- Test regex pattern with `grep`:
  ```bash
  echo "Phase 8 test" | grep -E "Phase [0-9]"
  ```
- Escape special characters: `\.`, `\*`, `\+`, `\?`, `\[`, `\]`
- Use `\|` for OR: `error\|bug`

## Comparison: MCP vs Slash Commands vs CLI

### When to Use MCP Tools

**Use MCP when**:
- Asking exploratory questions
- Combining multiple queries
- Letting Claude reason about what to query
- Natural language interaction preferred

**Example**:
```
"Where can I optimize my workflow?"
→ Claude autonomously queries stats, errors, sequences, and messages
```

### When to Use Slash Commands

**Use Slash Commands when**:
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

**Use CLI when**:
- Scripting or automation
- Outside Claude Code
- Debugging meta-cc itself
- Piping to other tools

**Example**:
```bash
meta-cc query tools --tool Bash --status error | jq '.tools | length'
```

## Advanced Usage

### Regex Pattern Examples

| Pattern | Description | Example Query |
|---------|-------------|---------------|
| `Phase 8` | Exact match | "Find messages about Phase 8" |
| `error\|bug` | OR operator | "Search for error or bug" |
| `^Continue` | Start with | "Find messages starting with Continue" |
| `test$` | End with | "Find messages ending with test" |
| `fix.*bug` | Between | "Find fix followed by bug" |
| `Phase [0-9]` | Number range | "Find all Phase mentions with numbers" |
| `(implement\|create)` | Grouped OR | "Find implement or create" |
| `\btest\b` | Word boundary | "Find exact word test" |

### Filter Combinations

```
# Query Bash errors in last 10 calls
query_tools(tool="Bash", status="error", limit=10)

# Find optimization discussions
query_user_messages(pattern="优化|improve|optimize", limit=20)

# Extract recent file operations
query_tools(tool="Edit|Write|Read", limit=30)

# Find repeated Read -> Edit sequences
query_tool_sequences(pattern="Read -> Edit", min_occurrences=5)

# Context around specific error
query_context(error_signature="npm_test_fail", window=5)
```

### Workflow Optimization Strategy

1. **Discover Patterns**:
   ```
   "What tool sequences do I repeat?"
   → query_tool_sequences(min_occurrences=3)
   ```

2. **Identify Pain Points**:
   ```
   "What errors keep happening?"
   → analyze_errors() + query_tools(status="error")
   ```

3. **Find Context**:
   ```
   "What happened before the npm test failures?"
   → query_context(error_signature="npm_test_fail", window=5)
   ```

4. **Create Automation**:
   - Repeated sequences → Slash Command
   - Repeated errors → Hook or pre-commit check
   - Complex workflows → Subagent (@meta-coach)

## Examples

### Example 1: Debug Workflow Errors

**User**: "Why do my commands keep failing?"

**Claude** (automatically):
```
1. Calling query_tools(status="error", limit=20)...
   → Found 5 Bash errors

2. Calling analyze_errors()...
   → Detected pattern: "npm test" failing repeatedly (5 times)

3. Calling query_context(error_signature="npm_test_fail", window=3)...
   → Context shows missing test file each time

4. Analysis:
   - Error occurs 5 times in last 50 turns
   - Same command: "npm test"
   - Root cause: tests/main.test.js not found
   - Suggestion: Create test file or fix path in package.json
```

### Example 2: Workflow Optimization

**User**: "Help me optimize my development workflow"

**Claude** (automatically):
```
1. Calling get_session_stats()...
   → 150 turns, 67 tool calls, 5% error rate

2. Calling query_tool_sequences(min_occurrences=3)...
   → Found: "Read -> Edit -> Bash" repeated 12 times

3. Calling query_tools(tool="Bash")...
   → Bash used 30 times (most frequent)

4. Calling query_user_messages(pattern="repeat|again|re-run")...
   → Found 8 messages about repeated tasks

5. Recommendations:
   - Create Slash Command for "Read -> Edit -> Bash -> Test" workflow
   - Add Hook for automatic test execution after Edit
   - Most common Bash command: "npm test" → Create /test shortcut
   - Consider @meta-coach for deeper workflow analysis
```

### Example 3: Message Search

**User**: "Did I mention Phase 8 implementation details earlier?"

**Claude** (automatically):
```
1. Calling query_user_messages(pattern="Phase 8.*(implement|detail)", limit=10)...

2. Results:
   - Turn 45: "Let's start Phase 8 implementation..."
   - Turn 67: "Phase 8 details should include..."
   - Turn 89: "After completing Phase 8 implementation..."

3. Summary: Yes, you discussed Phase 8 implementation 3 times,
   focusing on query commands and integration improvements.
```

### Example 4: File Access History

**User**: "How many times did I edit cmd/mcp.go?"

**Claude** (automatically):
```
1. Calling query_file_access(file="cmd/mcp.go")...

2. Results:
   - Read: 5 times (turns 12, 34, 56, 78, 90)
   - Edit: 8 times (turns 15, 23, 45, 67, 72, 81, 95, 102)
   - Total operations: 13

3. Summary: You edited cmd/mcp.go 8 times, with most edits
   occurring during Phase 8 MCP tool implementation.
```

### Example 5: Tool Sequence Analysis

**User**: "What are my most common workflows?"

**Claude** (automatically):
```
1. Calling query_tool_sequences(min_occurrences=3)...

2. Top Sequences:
   - "Read -> Edit -> Bash" (12 times) - Test workflow
   - "Grep -> Read -> Edit" (8 times) - Search and fix
   - "Glob -> Read -> Edit -> Bash" (5 times) - Batch editing

3. Recommendations:
   - Create /test-workflow for "Read -> Edit -> Bash"
   - Add Hook to auto-run tests after Edit
   - Consider batch editing tool for Glob-based workflows
```

## Summary

MCP Server provides:
- 8 powerful tools (3 from Phase 7 + 5 from Phase 8)
- Natural language queries (no manual commands)
- Autonomous analysis (Claude decides what to query)
- Flexible filtering (tool, status, pattern, limit)
- Context-aware (automatic pagination)
- Error investigation (context queries)
- Workflow optimization (sequence detection)
- File tracking (access history)

**Quick Start**:
1. Ensure `meta-cc` binary is in project root
2. Configuration is already in `.claude/mcp-servers/meta-cc.json`
3. Restart Claude Code (if not already loaded)
4. Start asking questions naturally!

**Next Steps**:
- Try natural language queries in Claude Code
- Explore different query patterns and filters
- Combine with @meta-coach for deep analysis
- Create custom Slash Commands for repeated workflows
- Use query_context to understand error patterns
- Leverage query_tool_sequences for workflow optimization

## Related Documentation

- [Integration Guide](integration-guide.md) - Choosing between MCP/Slash/Subagent
- [Examples & Usage](examples-usage.md) - Step-by-step setup guides
- [Slash Commands Reference](../README.md#slash-commands) - Quick command reference
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

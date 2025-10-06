# MCP Server Usage Guide

## Overview

meta-cc provides a Model Context Protocol (MCP) Server that allows Claude Code to autonomously query session data without manual CLI commands. With Phase 8-10 enhancements, the MCP server now provides **14 powerful tools** for comprehensive session analysis and advanced querying.

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
    "query_file_access",
    "query_project_state",
    "query_successful_prompts",
    "query_tools_advanced",
    "aggregate_stats",
    "query_time_series",
    "query_files"
  ]
}
```

## Available Tools

### 1. get_session_stats

Get comprehensive session statistics including turn count, tool usage, and error rates.

**Parameters**:
- `output_format` (optional): "jsonl" (JSONL format) (default: jsonl)

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
- `output_format` (optional): "jsonl" (JSONL format) (default: jsonl)

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
- `output_format` (optional): "jsonl" (JSONL format)

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
- `output_format` (optional): "jsonl" (JSONL format)

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
- `output_format` (optional): "jsonl" (JSONL format)

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
- `output_format` (optional): "jsonl" (JSONL format)

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
- `output_format` (optional): "jsonl" (JSONL format)

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
- `output_format` (optional): "jsonl" (JSONL format)

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

### 9. query_project_state (Stage 8.12+)

Query current project state extracted from the session, including git status, working directory, and project metadata.

**Parameters**:
- `output_format` (optional): "jsonl" (JSONL format)

**Example Queries**:
```
What's the current project state?
Show me git status from the session
What branch am I on?
```

**Direct Invocation**:
```
mcp__meta-insight__query_project_state
```

---

### 10. query_successful_prompts (Stage 8.12+)

Query successful prompt patterns based on quality scores and outcomes.

**Parameters**:
- `min_quality_score` (optional): Minimum quality score 0.0-1.0 (default: 0.8)
- `limit` (optional): Maximum results (default: 10)
- `output_format` (optional): "jsonl" (JSONL format)

**Example Queries**:
```
What prompts worked best?
Show me high-quality interactions
Find successful prompt patterns
```

**Direct Invocation**:
```
mcp__meta-insight__query_successful_prompts
```

---

### 11. query_tools_advanced (Phase 10 New)

Query tool calls with advanced SQL-like filter expressions.

**Parameters**:
- `where` (required): SQL-like filter expression
- `limit` (optional): Maximum results (default: 20)
- `output_format` (optional): "jsonl" (JSONL format)

**Supported Operators**:
- Boolean: AND, OR, NOT
- Comparison: =, !=, >, <, >=, <=
- Set: IN, NOT IN
- Range: BETWEEN ... AND ...
- Pattern: LIKE (SQL wildcards), REGEXP (regex)

**Example Queries**:
```
Find Bash errors with long duration
Show tools that took between 500 and 2000ms
Which Edit operations failed?
Find tools matching a pattern
```

**Filter Expression Examples**:
```bash
# Boolean logic
"tool='Bash' AND status='error'"
"status='error' OR duration>1000"
"NOT (status='success')"

# Set operations
"tool IN ('Bash', 'Edit', 'Write')"
"status NOT IN ('success')"

# Range queries
"duration BETWEEN 500 AND 2000"

# Pattern matching
"tool LIKE 'meta%'"
"error REGEXP 'permission.*denied'"

# Complex nested
"(tool='Bash' OR tool='Edit') AND status='error'"
```

**Direct Invocation**:
```
mcp__meta-insight__query_tools_advanced
```

---

### 12. aggregate_stats (Phase 10 New)

Aggregate statistics grouped by field with various metrics.

**Parameters**:
- `group_by` (optional): Field to group by - "tool", "status", or "uuid" (default: "tool")
- `metrics` (optional): Comma-separated metrics (default: "count,error_rate")
- `where` (optional): Filter expression to apply before aggregation
- `output_format` (optional): "jsonl" (JSONL format)

**Available Metrics**:
- `count`: Number of records in group
- `error_rate`: Percentage of errors (0.0-1.0)

**Example Queries**:
```
Show error rates by tool
Group by status and show counts
What's the error rate for each tool type?
Aggregate Bash commands by status
```

**Usage Examples**:
```bash
# Error rate by tool
aggregate_stats(group_by="tool", metrics="count,error_rate")

# Count by status
aggregate_stats(group_by="status", metrics="count")

# With filtering
aggregate_stats(group_by="tool", metrics="count,error_rate", where="duration>1000")
```

**Direct Invocation**:
```
mcp__meta-insight__aggregate_stats
```

---

### 13. query_time_series (Phase 10 New)

Analyze metrics over time with automatic time bucketing.

**Parameters**:
- `metric` (optional): Metric to analyze - "tool-calls" or "error-rate" (default: "tool-calls")
- `interval` (optional): Time interval - "hour", "day", or "week" (default: "hour")
- `where` (optional): Filter expression to apply before analysis
- `output_format` (optional): "jsonl" (JSONL format)

**Available Metrics**:
- `tool-calls`: Count of tool calls per time bucket
- `error-rate`: Error percentage per time bucket

**Example Queries**:
```
How has my tool usage changed over time?
Show error trends by day
When do I use the most tools?
Track Bash usage hourly
```

**Usage Examples**:
```bash
# Tool calls per hour
query_time_series(metric="tool-calls", interval="hour")

# Error rate per day
query_time_series(metric="error-rate", interval="day")

# Filtered time series (Bash only)
query_time_series(metric="tool-calls", interval="hour", where="tool='Bash'")
```

**Use Cases**:
- Identify peak usage hours
- Track error trends over time
- Monitor session activity patterns
- Detect workflow changes

**Direct Invocation**:
```
mcp__meta-insight__query_time_series
```

---

### 14. query_files (Phase 10 New)

Get file-level operation statistics with sorting and filtering.

**Parameters**:
- `sort_by` (optional): Sort field - "total_ops", "edit_count", "read_count", "write_count", "error_count", or "error_rate" (default: "total_ops")
- `top` (optional): Limit results to top N files (default: 20)
- `where` (optional): Filter expression
- `output_format` (optional): "jsonl" (JSONL format)

**Tracked Operations**:
- Read count
- Edit count
- Write count
- Error count
- Total operations
- Error rate

**Example Queries**:
```
What are my most edited files?
Which files have the most errors?
Show files with highest error rate
Find most active files
```

**Usage Examples**:
```bash
# Most edited files
query_files(sort_by="edit_count", top=10)

# Files with errors
query_files(sort_by="error_count", where="error_count>0")

# Most active files overall
query_files(sort_by="total_ops", top=20)

# Error-prone files
query_files(sort_by="error_rate", where="error_count>0", top=5)
```

**Use Cases**:
- Identify refactoring candidates (high edit count)
- Find problematic files (high error rate)
- Track file access patterns
- Prioritize code reviews

**Direct Invocation**:
```
mcp__meta-insight__query_files
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

## Phase 10 Integration Examples

### Example 1: Advanced Error Analysis

**User**: "Find all Bash errors that took longer than 1 second"

**Claude** (automatically):
```
1. Calling query_tools_advanced(where="tool='Bash' AND status='error' AND duration>1000")...
   → Found 3 long-running Bash errors

2. Analysis:
   - All errors related to network operations
   - Average duration: 2.5 seconds
   - Suggestion: Add timeout parameter to network commands
```

### Example 2: Tool Usage Patterns

**User**: "Show me tool usage statistics grouped by type"

**Claude** (automatically):
```
1. Calling aggregate_stats(group_by="tool", metrics="count,error_rate")...

2. Results:
   - Bash: 150 calls, 2.0% error rate
   - Read: 80 calls, 0% error rate
   - Edit: 60 calls, 1.7% error rate
   - Write: 40 calls, 0% error rate

3. Insights: Bash has highest usage and error rate - consider adding
   error handling hooks for common Bash patterns.
```

### Example 3: Time-Based Analysis

**User**: "How has my error rate changed over the past few days?"

**Claude** (automatically):
```
1. Calling query_time_series(metric="error-rate", interval="day")...

2. Trend Analysis:
   - Day 1: 5.2% error rate
   - Day 2: 3.1% error rate (improving)
   - Day 3: 1.8% error rate (steady improvement)

3. Observation: Error rate decreased 65% over 3 days, indicating
   better command patterns and fewer repeated mistakes.
```

### Example 4: File Hotspot Detection

**User**: "Which files am I editing the most?"

**Claude** (automatically):
```
1. Calling query_files(sort_by="edit_count", top=5)...

2. Top Edited Files:
   1. cmd/mcp.go - 15 edits, 0 errors
   2. internal/stats/aggregator.go - 8 edits, 1 error (12.5% error rate)
   3. cmd/query_tools.go - 6 edits, 0 errors
   4. README.md - 5 edits, 0 errors
   5. docs/mcp-usage.md - 4 edits, 0 errors

3. Recommendation: aggregator.go has high edit count and errors -
   consider refactoring or adding more comprehensive tests.
```

### Example 5: Complex Multi-Tool Analysis

**User**: "Help me understand why I have so many errors"

**Claude** (orchestrates multiple Phase 10 tools):
```
1. aggregate_stats(group_by="tool", metrics="count,error_rate")
   → Bash: 15% error rate (highest)

2. query_tools_advanced(where="tool='Bash' AND status='error'", limit=10)
   → Common pattern: "command not found" errors

3. query_files(sort_by="error_count", where="error_count>0")
   → Errors concentrated in 3 files

4. query_time_series(metric="error-rate", interval="day")
   → Error rate peaked yesterday, improving today

5. Comprehensive Analysis:
   - Root cause: Bash path errors in 3 specific files
   - Pattern: Errors decrease after corrections (learning curve)
   - Recommendation: Add path validation hook + use absolute paths
```

---

## Summary

MCP Server provides:
- **14 powerful tools** (3 from Phase 7 + 7 from Phase 8 + 4 from Phase 10)
- Natural language queries (no manual commands)
- Autonomous analysis (Claude decides what to query)
- **Advanced filtering** (SQL-like expressions with AND/OR/NOT, IN, BETWEEN, LIKE, REGEXP)
- **Statistical aggregation** (group-by with multiple metrics)
- **Time series analysis** (track trends over hour/day/week)
- **File-level insights** (hotspot detection and error correlation)
- Context-aware (automatic pagination)
- Error investigation (context queries)
- Workflow optimization (sequence detection)
- File tracking (access history)
- Project state monitoring (git status, metadata)
- Prompt pattern analysis (quality scoring)

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

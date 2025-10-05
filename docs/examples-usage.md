# Meta-CC Integration Examples - Usage Guide

This guide shows how to use the meta-cc integration examples (MCP Server, Subagent, and Slash Commands).

## Design Philosophy

meta-cc is designed as a **powerful data retrieval and statistics tool** for Claude Code session history. It follows a clear separation of concerns:

**meta-cc CLI Tool (Data Layer)**
- ✅ Powerful query and filtering capabilities
- ✅ Precise statistical analysis
- ✅ Flexible output control (pagination, chunking, sorting)
- ✅ Unix-style composability (pipe-friendly)
- ❌ **No semantic analysis** (no NLP, no recommendations)

**Claude Code Integration Layer (Semantic Layer)**
- Slash Commands: Single or multiple meta-cc calls with different parameters
- Subagents: Iterative multi-turn meta-cc calls
- MCP Tools: Claude autonomously decides parameters
- ✅ Semantic understanding, pattern recognition, recommendation generation

This separation allows Claude to perform complex analysis by making multiple meta-cc calls with varying parameters, while meta-cc focuses on fast, accurate data extraction.

## Prerequisites

1. **Install meta-cc**:
   ```bash
   cd /home/yale/work/meta-cc
   go build -o meta-cc
   sudo mv meta-cc /usr/local/bin/
   # Or add to your PATH
   ```

2. **Verify installation**:
   ```bash
   which meta-cc
   meta-cc --help
   ```

## Integration Hierarchy

meta-cc provides **three integration tiers** optimized for different use cases:

### Tier 1: MCP Server (Core Integration) - Highest Priority

**Use for**: Natural language queries, cross-session analysis, autonomous data access

Claude automatically calls MCP tools based on your questions. No special syntax needed.

**Available**: 14 MCP tools including `query_tools`, `analyze_errors`, `query_user_messages`, `aggregate_stats`, `query_time_series`, `query_files`, etc.

**Examples**:
```
"Show me all Bash errors in this project"
"Find user messages where I asked about testing"
"Compare tool usage between this week and last week"
"Which files have I edited the most across all sessions?"
```

### Tier 2: @meta-coach Subagent - Deep Analysis

**Use for**: Interactive coaching, workflow optimization, multi-turn analysis

The subagent combines MCP tools with LLM reasoning for personalized guidance.

**Examples**:
```
@meta-coach Why do my tests keep failing?
@meta-coach Help me optimize my workflow
@meta-coach Analyze patterns in my error history
@meta-coach What are my most time-consuming activities?
```

### Tier 3: Slash Commands - Quick Statistics (Lowest Priority)

**Use for**: Fast, pre-defined analyses without typing queries

**Available Commands** (4 core commands):

| Command | Description | Arguments |
|---------|-------------|-----------|
| `/meta-stats` | Session statistics | None |
| `/meta-errors` | Error pattern analysis | None |
| `/meta-timeline` | Timeline view | `[limit]` (default: 50) |
| `/meta-help` | Help and usage guide | None |

#### How to Use

**No configuration needed!** The slash commands are already installed in `.claude/commands/`.

Simply restart Claude Code and use them:

```
/meta-stats
/meta-errors
/meta-timeline 20
/meta-help
```

---

## Detailed Usage Examples

### Example 1: Using MCP for Natural Queries

**Scenario**: You want to understand error patterns across your project.

**Instead of**:
```
/meta-errors
```

**Use MCP** (just ask naturally):
```
"Show me all errors in this project grouped by tool"
"What are the most common error messages?"
"Compare error rates between morning and afternoon work sessions"
```

**Why MCP is better**:
- No need to remember command syntax
- Claude autonomously picks the right MCP tool
- Supports complex multi-condition queries
- Cross-session analysis by default

---

### Example 2: Using @meta-coach for Workflow Optimization

**Scenario**: You notice you're making the same mistakes repeatedly.

**Step 1**: Ask @meta-coach
```
@meta-coach I keep getting Bash errors. What's the pattern?
```

**Step 2**: @meta-coach analyzes (using MCP tools internally)
- Calls `query_tools` with `tool=Bash, status=error`
- Calls `analyze_errors` to detect patterns
- Calls `query_context` to see surrounding turns

**Step 3**: @meta-coach provides personalized guidance
```
I've analyzed your Bash errors across 15 sessions. Pattern detected:
- 80% of errors occur when using grep with complex regex
- Most failures happen in files with special characters in names

Recommendations:
1. Use `grep -F` for literal strings (avoid regex complexity)
2. Quote file paths: grep "pattern" "$file" instead of grep pattern $file
3. Test regex with `grep --color` first to validate patterns

Would you like me to create a custom Hook to validate grep commands before execution?
```

---

### Example 3: Combining All Three Tiers

**Scenario**: Comprehensive workflow analysis

**Tier 3 (Quick Start)**: Get overview
```
/meta-stats
```

**Output**: "Error rate: 15% - unusually high"

**Tier 1 (Deep Dive)**: Investigate with MCP
```
"Show me the top 5 error patterns in the last 100 tool calls"
"Which files are associated with the most errors?"
```

**Tier 2 (Optimization)**: Get coaching
```
@meta-coach Based on these errors, how should I improve my workflow?
```

**@meta-coach response**:
- Analyzes error clusters
- Identifies root causes (file permissions, missing dependencies, etc.)
- Suggests concrete fixes (Hooks, better prompts, alternative tools)
- Offers to implement solutions

---

### Example 4: MCP Tool Showcase

**Available MCP Tools** (14 total):

#### Core Query Tools
```
query_tools              # Query tool call history
query_user_messages      # Search user messages with regex
query_errors             # Extract error records
query_file_access        # File operation history
query_context            # Context around specific events
```

#### Statistical Tools
```
aggregate_stats          # Grouped statistics (by tool, status, etc.)
query_time_series        # Temporal pattern analysis
query_files              # File-level operation statistics
```

#### Pattern Detection
```
analyze_errors           # Error pattern detection
query_tool_sequences     # Repeated tool sequences
query_successful_prompts # Identify high-quality prompts
```

#### Advanced Queries
```
query_tools_advanced     # SQL-like filtering
extract_tools            # Bulk tool extraction with pagination
get_session_stats        # Session-only statistics (backward compat)
```

**Example Queries Using These Tools**:

```
# Basic queries (Claude picks the right tool)
"Show recent Bash errors"                    → query_tools
"Find messages about testing"                → query_user_messages
"What files did I edit most?"                → query_files

# Advanced queries (Claude composes multiple tools)
"Compare my workflow this week vs last week" → query_time_series + aggregate_stats
"Find all errors related to test.py"         → query_errors + query_context
"Show my most productive hours"              → query_time_series (tool-calls by hour)
```
```

---

## MCP Server Configuration (Optional)

The MCP server provides 14 advanced query tools for programmatic access to session history. See Example 4 above for the complete list of available tools.

### Configuration Steps

1. **Verify Node.js is available**:
   ```bash
   node --version  # Should show v14+ or higher
   ```

2. **Add MCP server to Claude Code settings**:

   Edit `~/.claude/settings.json`:
   ```json
   {
     "mcpServers": {
       "meta-insight": {
         "command": "node",
         "args": ["/home/yale/work/meta-cc/.claude/mcp-servers/meta-insight.js"],
         "transport": "stdio"
       }
     }
   }
   ```

3. **Restart Claude Code** to load the MCP server

4. **Test with natural queries**:
   ```
   "Show me all Bash errors in this project"
   "Find user messages mentioning 'refactor'"
   "Analyze tool usage trends across sessions"
   ```

---

## Quick Start Checklist

### Before Restarting Claude Code

- [x] ✅ meta-cc installed and in PATH
- [x] ✅ Slash Commands (4 core) in `.claude/commands/`
- [x] ✅ @meta-coach Subagent in `.claude/agents/`
- [ ] ⚙️ MCP Server in `~/.claude/settings.json` (optional)

### After Restarting Claude Code

Run these tests:

```bash
# Tier 3: Slash Commands
/meta-stats
/meta-errors
/meta-timeline 10
/meta-help

# Tier 2: Subagent
@meta-coach Help me analyze my workflow

# Tier 1: MCP (natural queries)
"Show me all Bash errors in this session"
```

---

## Troubleshooting

### Slash Commands not showing up

1. **Check file location**:
   ```bash
   ls -la .claude/commands/
   # Should show: meta-stats.md, meta-errors.md, meta-timeline.md, meta-help.md
   ```

2. **Restart Claude Code completely**

3. **Test manually**:
   ```bash
   bash -c "$(sed -n '/```bash/,/```/p' .claude/commands/meta-stats.md | grep -v '```')"
   ```

### @meta-coach not responding

1. **Check file location**:
   ```bash
   ls -la .claude/agents/
   # Should show: meta-coach.md
   ```

2. **Verify meta-cc works**:
   ```bash
   meta-cc parse stats --output md
   ```

3. **Restart Claude Code**

### MCP Server not connecting

1. **Check settings.json syntax**:
   ```bash
   cat ~/.claude/settings.json | jq .
   # Should parse without errors
   ```

2. **Verify Node.js**:
   ```bash
   node --version
   ```

3. **Test server manually**:
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | \
   node .claude/mcp-servers/meta-insight.js
   ```

4. **Check Claude Code logs** for MCP connection errors

### meta-cc command not found

```bash
# Check installation
which meta-cc

# If not found, install:
cd /home/yale/work/meta-cc
go build -o meta-cc
sudo mv meta-cc /usr/local/bin/

# Or add to PATH:
export PATH=$PATH:/home/yale/work/meta-cc
```

---

## Next Steps

1. **Restart Claude Code** to load all integrations
2. **Run test commands** from the checklist above
3. **Try natural MCP queries** for cross-session analysis
4. **Use @meta-coach** for interactive workflow optimization
5. **Refer to [Integration Guide](integration-guide.md)** for decision framework

---

## Documentation

- **Main README**: `/home/yale/work/meta-cc/README.md`
- **Integration Guide**: `/home/yale/work/meta-cc/docs/integration-guide.md`
- **Troubleshooting**: `/home/yale/work/meta-cc/docs/troubleshooting.md`
- **CLI Help**: `meta-cc --help`
- **This Guide**: `/home/yale/work/meta-cc/docs/examples-usage.md`

---

Enjoy using meta-cc to optimize your Claude Code workflow! 🚀

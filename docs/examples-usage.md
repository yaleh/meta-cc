# Meta-CC Integration Examples - Usage Guide

This guide shows how to use the meta-cc integration examples (MCP Server, Subagent, and Slash Commands).

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

## Available Integrations

### 1. Slash Commands (✅ Ready to use)

#### Available Commands

| Command | Description | Arguments |
|---------|-------------|-----------|
| `/meta-stats` | Session statistics | None |
| `/meta-errors` | Error pattern analysis | `[window-size]` (default: 20) |
| `/meta-compare` | Compare with other projects | `[project-path]` |
| `/meta-timeline` | Timeline view | `[limit]` (default: 50) |
| `/meta-help` | Show help | None |

#### How to Use

**No configuration needed!** The slash commands are already installed in `.claude/commands/`.

Simply restart Claude Code and use them:

```
/meta-stats
/meta-errors 30
/meta-compare /home/yale/work/NarrativeForge
/meta-timeline 20
/meta-help
```

#### Testing Slash Commands

```bash
# Test 1: Get current session statistics
/meta-stats

# Expected output:
# - Total Turns: 2,563
# - Tool Calls: 971
# - Error Rate: 0.0%
# - Session Duration: 524.3 minutes
# - Top 5 Tools with percentages

# Test 2: Analyze errors in last 20 turns
/meta-errors 20

# Expected output:
# - Error patterns detected (if any)
# - Or "No error patterns detected"

# Test 3: Compare with another project
/meta-compare /home/yale/work/NarrativeForge

# Expected output:
# - Statistics for both projects side by side
# - Comparison insights

# Test 4: View timeline
/meta-timeline 30

# Expected output:
# - Chronological list of tool calls
# - Status indicators (✅/❌)
# - Summary statistics

# Test 5: Get help
/meta-help

# Expected output:
# - Complete usage guide
# - All available commands
# - CLI tool help
```

---

### 2. Subagent: @meta-coach (✅ Ready to use)

#### Description

An interactive meta-cognition coach that analyzes your session history and provides personalized workflow optimization advice.

#### How to Use

**No configuration needed!** The subagent is already installed in `.claude/agents/`.

Simply restart Claude Code and invoke it:

```
@meta-coach
```

Then have a conversation:

```
You: @meta-coach I feel like I'm stuck in a loop with these tests...

Coach: Let me analyze your recent session to see what's happening.
[Runs: meta-cc analyze errors --window 30 --output md]
...

You: @meta-coach How can I optimize my workflow?

Coach: Let me check your tool usage patterns.
[Runs: meta-cc parse stats --output md]
...
```

#### Testing Subagent

```bash
# Test 1: General analysis
@meta-coach Help me analyze my workflow

# Test 2: Specific problem
@meta-coach I keep running the same failing test, what should I do?

# Test 3: Cross-project learning
@meta-coach How did I solve authentication issues in past projects?

# Test 4: Optimization request
@meta-coach Create a custom Slash Command for my common workflow
```

#### What @meta-coach Can Do

- ✅ Analyze your session statistics
- ✅ Detect error patterns
- ✅ Identify repetitive behaviors
- ✅ Provide tiered suggestions (immediate/optional/long-term)
- ✅ Create Hooks and Slash Commands for you
- ✅ Search across project history
- ✅ Guide you through optimization

---

### 3. MCP Server: meta-insight (⚙️ Requires configuration)

#### Description

Provides meta-cc functionality through Model Context Protocol, allowing Claude to query session data programmatically.

#### Configuration Steps

1. **Create settings file** (if it doesn't exist):
   ```bash
   mkdir -p ~/.claude
   touch ~/.claude/settings.json
   ```

2. **Add MCP server configuration**:

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

3. **Verify Node.js is available**:
   ```bash
   node --version
   # Should show v14+ or higher
   ```

4. **Make server executable**:
   ```bash
   chmod +x .claude/mcp-servers/meta-insight.js
   ```

5. **Restart Claude Code** to load the MCP server

#### How to Use

Once configured, Claude can automatically use the MCP tools when needed:

```
You: Use meta-insight to get my current session statistics

Claude: [Calls tool: get_session_stats with output_format: "json"]
Here are your session statistics:
- Total Turns: 2,563
- Tool Calls: 971
...

You: Analyze errors in the last 30 turns using meta-insight

Claude: [Calls tool: analyze_errors with window: 30, output_format: "json"]
I found 2 error patterns:
...
```

#### Available MCP Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `get_session_stats` | Get session statistics | `output_format`: json\|md |
| `analyze_errors` | Analyze error patterns | `window`: number, `output_format`: json\|md |
| `extract_tools` | Extract tool usage data | `filter`: all\|error\|success, `output_format`: json\|md |

#### Testing MCP Server

```bash
# Test manually (before adding to Claude Code)
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | node .claude/mcp-servers/meta-insight.js

# Expected: JSON response with serverInfo

# Test in Claude Code (after configuration)
You: Use meta-insight to get session stats
Claude: [Should call get_session_stats tool and display results]

You: Use meta-insight to analyze errors with window size 20
Claude: [Should call analyze_errors tool with window=20]
```

---

## Quick Start Checklist

### Before Restarting Claude Code

- [x] ✅ meta-cc installed and in PATH
- [x] ✅ Slash Commands created in `.claude/commands/`
- [x] ✅ Subagent created in `.claude/agents/`
- [ ] ⚙️ MCP Server configured in `~/.claude/settings.json` (optional)

### After Restarting Claude Code

Run these tests:

```bash
# 1. Test Slash Commands
/meta-stats
/meta-errors
/meta-help

# 2. Test Subagent
@meta-coach Hello! Can you help me analyze my workflow?

# 3. Test MCP (if configured)
Use meta-insight to get my session statistics
```

---

## Which Integration Should I Use?

meta-cc provides three integration methods, each optimized for different use cases:

- **MCP Server**: Seamless data access (Claude queries autonomously)
- **Slash Commands**: Quick, pre-defined workflows
- **Subagent (@meta-coach)**: Interactive, conversational analysis

### Quick Decision Guide

| I want to... | Use this |
|--------------|----------|
| Check stats quickly | `/meta-stats` or ask naturally (MCP) |
| Analyze repeated errors | `/meta-errors 30` |
| Understand workflow inefficiency | `@meta-coach` |
| Compare projects | `/meta-compare <path>` |
| Get help and guidance | `@meta-coach` |

**👉 For detailed comparison, decision framework, and best practices, see the [Integration Guide](integration-guide.md).**

This guide focuses on **how to use** each integration. The Integration Guide explains **when to choose** each one and provides:
- 📊 Core differences (context isolation, invocation, execution models)
- 🎯 Decision trees and scenario matrices
- 💡 Anti-patterns and best practices
- 📝 Real-world case studies

**Examples**:
- "Analyze my last 3 sessions and find patterns"
- "Compare error rates across all my projects"
- Complex multi-step analysis workflows

---

## Troubleshooting

### Slash Commands not showing up

1. **Check file location**:
   ```bash
   ls -la .claude/commands/
   # Should show: meta-stats.md, meta-errors.md, etc.
   ```

2. **Restart Claude Code completely**:
   ```bash
   # Exit Claude Code
   # Start Claude Code again
   ```

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
3. **Try @meta-coach** for interactive analysis
4. **Configure MCP** (optional) for advanced workflows
5. **Create custom Slash Commands** based on your needs

## Documentation

- **Main README**: `/home/yale/work/meta-cc/README.md`
- **Troubleshooting**: `/home/yale/work/meta-cc/docs/troubleshooting.md`
- **CLI Help**: `meta-cc --help`
- **This Guide**: `/home/yale/work/meta-cc/docs/examples-usage.md`

---

## Example Session Flow

```
# 1. Start with overview
/meta-stats

# 2. Notice high error rate? Dig deeper
/meta-errors 30

# 3. Want personalized advice?
@meta-coach I have a 15% error rate, mostly from Bash. What should I do?

# 4. Compare with successful project
/meta-compare /home/yale/work/successful-project

# 5. View detailed timeline
/meta-timeline 20

# 6. Get help anytime
/meta-help
```

Enjoy using meta-cc to optimize your Claude Code workflow! 🚀

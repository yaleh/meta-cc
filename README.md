# meta-cc

[![CI](https://github.com/yaleh/meta-cc/actions/workflows/ci.yml/badge.svg)](https://github.com/yaleh/meta-cc/actions)
[![License](https://img.shields.io/github/license/yaleh/meta-cc)](LICENSE)
[![Release](https://img.shields.io/github/v/release/yaleh/meta-cc)](https://github.com/yaleh/meta-cc/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/yaleh/meta-cc)](go.mod)
[![Plugin Marketplace](https://img.shields.io/badge/Claude_Code-Plugin_Marketplace-blue)](https://github.com/yaleh/meta-cc)

Meta-Cognition tool for Claude Code - analyze session history for workflow optimization.

## Recent Milestones

### Agent Formalization (v0.11.1) - October 2025
- **92% size reduction** across 5 agent files (3074 → 244 lines)
- **100% behavioral semantics preserved** using lambda calculus formal specifications
- **Zero regressions** - all tests pass with 70-100% coverage
- Replaces verbose prose with mathematically precise function definitions
- See [Formalization Summary](.claude/agents/FORMALIZATION_SUMMARY.md) for details

## Features

- 🔍 Parse Claude Code session history (JSONL format)
- 📊 Statistical analysis of tool usage and errors
- 🎯 Pattern detection for workflow optimization
- 🚀 Zero dependencies - single binary deployment

## Installation

### Step 1: Install Claude Code Plugin (Recommended)

Install meta-cc directly from the Claude Code plugin marketplace:

```bash
/plugin marketplace add yaleh/meta-cc
/plugin install meta-cc
```

**What gets installed automatically:**
- ✅ 10 slash commands (`~/.claude/commands/meta-*.md`)
- ✅ 3 subagents (`~/.claude/agents/meta-*.md`)
- ✅ Plugin cache in `~/.claude/plugins/repos/yaleh_meta-cc/`

**What you need to install manually:**
- ⚠️ `meta-cc` CLI binary (required for slash commands)
- ⚠️ `meta-cc-mcp` MCP server (required for MCP tools)

> **Note**: Claude Code plugin system does NOT automatically install binaries. You must install them separately (see Step 2).

### Step 2: Install Binaries

Choose one of the following methods to install `meta-cc` and `meta-cc-mcp` binaries:

#### Option A: Plugin Package (Includes install.sh)

Download the platform-specific package which includes both binaries and an installation script:

**Linux (x86_64)**
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-amd64.tar.gz | tar xz
cd meta-cc-plugin-linux-amd64
./install.sh
```

**Linux (ARM64)**
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-arm64.tar.gz | tar xz
cd meta-cc-plugin-linux-arm64
./install.sh
```

**macOS (Intel)**
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-darwin-amd64.tar.gz | tar xz
cd meta-cc-plugin-darwin-amd64
./install.sh
```

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-darwin-arm64.tar.gz | tar xz
cd meta-cc-plugin-darwin-arm64
./install.sh
```

**Windows (x86_64)**
```bash
# Using Git Bash or WSL
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-windows-amd64.tar.gz | tar xz
cd meta-cc-plugin-windows-amd64
./install.sh
```

**What install.sh does:**
- Installs binaries to `~/.local/bin/` (or `$INSTALL_DIR` if set)
- Adds binaries to your PATH automatically
- **Note**: Does NOT configure Claude Code commands/agents (use `/plugin install` for that)

#### Option B: Individual Binaries

<details>
<summary>Click to expand individual binary download instructions</summary>

#### CLI Only - Linux (x86_64)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-linux-amd64 -o meta-cc
chmod +x meta-cc
sudo mv meta-cc /usr/local/bin/
```

#### CLI Only - Linux (ARM64)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-linux-arm64 -o meta-cc
chmod +x meta-cc
sudo mv meta-cc /usr/local/bin/
```

#### CLI Only - macOS (Intel)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-darwin-amd64 -o meta-cc
chmod +x meta-cc
sudo mv meta-cc /usr/local/bin/
```

#### CLI Only - macOS (Apple Silicon)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-darwin-arm64 -o meta-cc
chmod +x meta-cc
sudo mv meta-cc /usr/local/bin/
```

#### CLI Only - Windows (PowerShell)
```powershell
Invoke-WebRequest -Uri "https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-windows-amd64.exe" -OutFile "meta-cc.exe"
# Move to a directory in your PATH
```

</details>

#### Option C: Build from Source

```bash
git clone https://github.com/yaleh/meta-cc.git
cd meta-cc
make install
```

**Requirements:**
- Go 1.21 or later
- make

### Step 3: Configure MCP Server

After installing binaries, enable MCP tools in Claude Code using the `claude mcp` command:

```bash
# Add meta-cc MCP server to Claude Code
claude mcp add meta-cc-mcp meta-cc-mcp
```

**What this does:**
- Registers `meta-cc-mcp` as an MCP server in `~/.claude/mcp.json`
- Enables 14 MCP query tools for session analysis
- Allows Claude to autonomously query session data during conversations

**Verify MCP setup:**
```bash
# List all MCP servers
claude mcp list

# Check if meta-cc-mcp is registered
cat ~/.claude/mcp.json
```

**Alternative manual configuration:**

If `claude mcp` command is not available, manually edit `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "meta-cc-mcp": {
      "command": "meta-cc-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

### Verify Complete Installation

After completing all steps:

**1. Check binaries:**
```bash
meta-cc --version       # Should show v0.16.0
meta-cc-mcp --version   # Should show v0.16.0
which meta-cc          # Should show path to binary
```

**2. Test in Claude Code:**
- Slash command: `/meta-stats`
- Subagent: `@meta-coach analyze my workflow`
- MCP tools: Ask "What are my recent tool usage patterns?"

**3. Verify file locations:**
```bash
ls ~/.claude/commands/meta-*.md    # Should show 10 files
ls ~/.claude/agents/meta-*.md      # Should show 3 files
cat ~/.claude/mcp.json             # Should show meta-cc-mcp config
```

If you encounter issues, see the [Troubleshooting Guide](docs/installation.md#troubleshooting).

### Uninstall

To remove meta-cc:

```bash
# 1. Remove plugin (removes commands/agents)
/plugin uninstall meta-cc

# 2. Remove binaries
rm ~/.local/bin/meta-cc ~/.local/bin/meta-cc-mcp

# 3. Remove MCP configuration
claude mcp remove meta-cc-mcp
# Or manually edit ~/.claude/mcp.json
```

## MCP Server Integration (Legacy Documentation)

> **Note**: If you followed the installation steps above, your MCP server should already be configured. This section is for reference only.

<details>
<summary>Click to expand legacy MCP documentation</summary>

Deploy meta-cc as an MCP server to analyze sessions directly within Claude Code.

### Build from Source
```bash
git clone https://github.com/yaleh/meta-cc.git
cd meta-cc
make build-mcp
cp meta-cc-mcp ~/.local/bin/
```

**2. Configure Claude Code:**

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `~/.config/claude/claude_desktop_config.json` (Linux):

```json
{
  "mcpServers": {
    "meta-cc": {
      "type": "stdio",
      "command": "meta-cc-mcp",
      "args": [],
      "env": {
        "META_CC_INLINE_THRESHOLD": "8192"
      }
    }
  }
}
```

**3. Restart Claude Code** to load the server.

**4. Test the integration:**
```
@meta-cc get_session_stats
@meta-cc query_tools --limit=10
@meta-cc query_user_messages --pattern=".*error.*"
```

### Available MCP Tools

- **`get_session_stats`** - Session statistics
- **`query_tools`** - Tool call analysis
- **`query_user_messages`** - Search user messages
- **`query_tool_sequences`** - Workflow patterns
- **`query_time_series`** - Time-based analysis
- **`cleanup_temp_files`** - File management

*See [MCP Tools Reference](docs/mcp-tools-reference.md) for complete documentation.*

## Usage

```bash
# Show help
./meta-cc --help

# Show version
./meta-cc --version

# Global options
./meta-cc --session <session-id>    # Specify session ID
./meta-cc --project <path>          # Specify project path
./meta-cc --output jsonl|tsv        # Output format (JSONL default, TSV for CLI tools)

# Phase 9: Context-Length Management Options
./meta-cc --limit N                 # Limit output to N records
./meta-cc --offset M                # Skip first M records
./meta-cc --estimate-size           # Predict output size before generating
./meta-cc --chunk-size N --output-dir DIR  # Split into chunks
./meta-cc --fields "f1,f2,f3"       # Output only specified fields (70%+ size reduction)
./meta-cc --if-error-include "f4"   # Include extra fields on errors
./meta-cc --summary-first --top N   # Summary + top N details
```

## Query Commands

### Session Statistics

```bash
# Get session stats
./meta-cc parse stats
```

### Tool Call Analysis

```bash
# Query all tool calls
./meta-cc query tools

# Filter by tool name
./meta-cc query tools --tool Bash

# Filter by status
./meta-cc query tools --status error --limit 20
```

### User Message Search

```bash
# Search user messages with regex
./meta-cc query user-messages --pattern "fix.*bug"

# Limit results
./meta-cc query user-messages --pattern "error" --limit 10

# With context window
./meta-cc query user-messages --pattern "implement" --with-context 3
```

## Output Format Philosophy

meta-cc follows the **Unix philosophy** with two core output formats:

### JSONL (Default - JSON Lines)
- **Machine-readable**: One JSON object per line
- **Composable**: Pipe to `jq` for complex filtering
- **Consistent**: All commands output JSONL by default
- **Integration-ready**: Perfect for Slash Commands and MCP Server
- **Streaming-friendly**: Process large datasets incrementally

### TSV (Tab-Separated Values)
- **CLI-friendly**: Easy to process with `awk`, `grep`, `cut`
- **Compact**: Smaller than JSON for large datasets
- **Human-readable**: Can be viewed directly or with `column -t`
- **No escaping issues**: Simpler than CSV (no quote handling)

### Removed Formats (Phase 13)
- **JSON (pretty)**: Use `jq '.'` for pretty-printing JSONL
- **CSV**: Use TSV instead (simpler, no quoting issues)
- **Markdown**: Let Claude Code render JSONL to Markdown

## JSONL Output Format Reference

Understanding meta-cc's JSONL output structure is crucial for processing data with tools like `jq`.

### Command Output Types

| Command | Output Type | Structure |
|---------|------------|-----------|
| `parse stats` | **Object** | `{"TurnCount": N, "ToolCallCount": N, ...}` |
| `parse extract --type turns` | **Array** | `[{turn1}, {turn2}, ...]` |
| `parse extract --type tools` | **Array** | `[{tool1}, {tool2}, ...]` |
| `analyze errors` | **Array** | `[{pattern1}, ...]` or `[]` |

### Common Mistakes ❌ → Correct Usage ✅

#### parse stats (returns Object)

```bash
# ✅ Correct (JSONL - default format)
meta-cc parse stats | jq '.TurnCount'
meta-cc parse stats | jq '.ErrorRate'

# ✅ Also correct (explicit JSONL)
meta-cc parse stats --output jsonl | jq '.TurnCount'

# ❌ Wrong - no wrapper object
meta-cc parse stats | jq '.stats.TurnCount'
```

#### parse extract --type tools (returns Array)

```bash
# ✅ Correct (JSONL - default format, outputs one JSON object per line)
meta-cc parse extract --type tools | jq -s 'length'    # Count all
meta-cc parse extract --type tools | jq '.'             # Show each line
meta-cc parse extract --type tools | jq -r '.ToolName' # Extract field

# ✅ Also correct (convert JSONL to array with -s)
meta-cc parse extract --type tools | jq -s '.[0]'     # First tool
meta-cc parse extract --type tools | jq -s '.[]'      # All tools

# ❌ Wrong - assumes object wrapper
meta-cc parse extract --type tools | jq '.tools'
meta-cc parse extract --type tools | jq '.tools[]'
```

#### parse extract --type turns (returns Array)

```bash
# ✅ Correct (JSONL - default format)
meta-cc parse extract --type turns | jq -s 'length'
meta-cc parse extract --type turns | jq 'select(.type == "user")'
meta-cc parse extract --type turns | jq -r '.timestamp'

# ❌ Wrong - assumes object wrapper
meta-cc parse extract --type turns | jq '.turns'
```

#### analyze errors (returns Array)

```bash
# ✅ Correct (JSONL - default format, one error pattern per line)
meta-cc analyze errors | jq -s 'length'
meta-cc analyze errors | jq '.'
meta-cc analyze errors | jq -s 'if type == "array" then length else 0 end'

# ❌ Wrong - assumes object wrapper
meta-cc analyze errors | jq '.ErrorPatterns'
meta-cc analyze errors | jq '.ErrorPatterns | length'
```

### Detailed Output Structures

#### parse stats

Returns a single object with session statistics:

```json
{
  "TurnCount": 2676,
  "UserTurnCount": 1097,
  "AssistantTurnCount": 1579,
  "ToolCallCount": 1012,
  "ErrorCount": 0,
  "ErrorRate": 0,
  "DurationSeconds": 33796,
  "ToolFrequency": {
    "Bash": 495,
    "Read": 162,
    "TodoWrite": 140,
    "Write": 78,
    "Edit": 65
  },
  "TopTools": [
    {"Name": "Bash", "Count": 495, "Percentage": 48.9},
    {"Name": "Read", "Count": 162, "Percentage": 16.0}
  ]
}
```

**jq usage examples:**
```bash
# Get total turns
jq '.TurnCount'

# Get error rate
jq '.ErrorRate'

# Get top 3 tools
jq '.TopTools[:3]'

# Get frequency of specific tool
jq '.ToolFrequency.Bash'
```

#### parse extract --type tools

Returns an array of tool call objects:

```json
[
  {
    "UUID": "abc-123",
    "ToolName": "Bash",
    "Input": {
      "command": "ls -la",
      "description": "List files"
    },
    "Output": "file1.txt\nfile2.txt",
    "Status": "",
    "Error": ""
  },
  {
    "UUID": "def-456",
    "ToolName": "Read",
    "Input": {
      "file_path": "/path/to/file.go"
    },
    "Output": "package main...",
    "Status": "",
    "Error": ""
  }
]
```

**jq usage examples:**
```bash
# Count total tools
jq 'length'

# Get all tool names
jq -r '.[] | .ToolName'

# Filter by tool name
jq '.[] | select(.ToolName == "Bash")'

# Get last 10 tools
jq '.[-10:]'

# Count tool frequency
jq -r '.[] | .ToolName' | sort | uniq -c | sort -rn

# Extract only errors
jq '.[] | select(.Status == "error")'

# Get commands from Bash tools
jq -r '.[] | select(.ToolName == "Bash") | .Input.command'
```

#### parse extract --type turns

Returns an array of turn (conversation entry) objects:

```json
[
  {
    "type": "user",
    "timestamp": "2025-10-02T06:07:13.673Z",
    "uuid": "abc-123",
    "sessionId": "session-id",
    "message": {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "Execute Stage 1.1..."
        }
      ]
    }
  },
  {
    "type": "assistant",
    "timestamp": "2025-10-02T06:08:57.769Z",
    "uuid": "def-456",
    "message": {
      "role": "assistant",
      "content": [
        {
          "type": "text",
          "text": "I'll execute Stage 1.1..."
        }
      ]
    }
  }
]
```

**jq usage examples:**
```bash
# Count total turns
jq 'length'

# Get only user turns
jq '.[] | select(.type == "user")'

# Get user messages
jq -r '.[] | select(.type == "user") | .message.content[0].text'

# Filter by timestamp
jq '.[] | select(.timestamp > "2025-10-02T12:00:00Z")'

# Get turn UUIDs
jq -r '.[] | .uuid'
```

#### analyze errors

Returns an array of error pattern objects (empty array if no patterns):

```json
[
  {
    "PatternID": "bash_npm_test_error",
    "Type": "command_error",
    "Occurrences": 5,
    "Signature": "abc123def456",
    "ToolName": "Bash",
    "ErrorText": "npm test failed",
    "FirstSeen": "2025-10-02T10:00:00Z",
    "LastSeen": "2025-10-02T10:30:00Z",
    "TimeSpanSeconds": 1800,
    "Context": {
      "TurnUUIDs": ["uuid1", "uuid2", "uuid3"],
      "TurnIndices": [100, 150, 200]
    }
  }
]
```

**jq usage examples:**
```bash
# Count error patterns
jq 'length'

# Get all error patterns
jq '.[]'

# Filter by minimum occurrences
jq '.[] | select(.Occurrences >= 5)'

# Get error messages
jq -r '.[] | .ErrorText'

# Sort by occurrence count
jq 'sort_by(.Occurrences) | reverse'

# Safe length check (handles empty results)
jq 'if type == "array" then length else 0 end'
```

### Integration with jq

#### Safe Type Checking

Always check the type when uncertain:

```bash
# Safe array length (JSONL - use -s to slurp into array)
meta-cc analyze errors | jq -s 'if type == "array" then length else 0 end'

# Safe object property access
meta-cc parse stats | jq 'if type == "object" then .TurnCount else null end'
```

#### Combining Commands

```bash
# Get tool usage stats (JSONL format)
meta-cc parse extract --type tools | \
  jq -r '.ToolName' | \
  sort | uniq -c | sort -rn

# Find repeated Bash commands
meta-cc parse extract --type tools | \
  jq 'select(.ToolName == "Bash") | .Input.command' | jq -r | \
  sort | uniq -c | sort -rn

# Calculate error rate manually
TOTAL=$(meta-cc parse extract --type tools | jq -s 'length')
ERRORS=$(meta-cc parse extract --type tools | \
  jq 'select(.Status == "error")' | jq -s 'length')
echo "scale=2; $ERRORS * 100 / $TOTAL" | bc
```

#### Common Patterns

```bash
# Pattern 1: Filter and count (JSONL)
meta-cc parse extract --type tools | \
  jq 'select(.ToolName == "Edit")' | jq -s 'length'

# Pattern 2: Extract specific fields
meta-cc parse extract --type tools | \
  jq -r '"\(.ToolName): \(.Input.command // "N/A")"'

# Pattern 3: Group by field (slurp JSONL into array first)
meta-cc parse extract --type tools | jq -s \
  'group_by(.ToolName) | map({tool: .[0].ToolName, count: length})'

# Pattern 4: Time-based filtering (for turns)
meta-cc parse extract --type turns | \
  jq 'select(.timestamp > "2025-10-02T12:00:00Z")'
```

### Troubleshooting

**Error: "Cannot index array with string"**
```bash
# You're trying to access object properties on an array
# Fix: Remove the property accessor or use .[]

# ❌ Wrong
jq '.tools'

# ✅ Correct
jq '.[]'
```

**Error: "Cannot iterate over object"**
```bash
# You're trying to iterate an object
# Fix: Access the specific property first

# ❌ Wrong
jq '.[] | .ToolName'  # on parse stats

# ✅ Correct
jq '.ToolFrequency | to_entries | .[] | "\(.key): \(.value)"'
```

**Empty output**
```bash
# Command returns empty array []
# This is normal when no data matches (e.g., no errors)

# Check if empty
meta-cc analyze errors --output json | jq 'length == 0'

# Provide default value
meta-cc analyze errors --output json | jq 'if length == 0 then "No errors" else . end'
```

## Claude Code Integration

meta-cc provides deep integration with Claude Code, allowing you to analyze session metadata directly within your conversation using Slash Commands.

### Installation Steps

#### 1. Install meta-cc CLI Tool

**Option A: Build from source**

```bash
# Clone the repository
git clone https://github.com/yaleh/meta-cc.git
cd meta-cc

# Build the binary
make build

# Install to system path
sudo cp meta-cc /usr/local/bin/meta-cc
sudo chmod +x /usr/local/bin/meta-cc

# Verify installation
meta-cc --version
```

**Option B: Download pre-compiled binary** (coming soon)

```bash
# Download latest release (Linux x64)
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-linux-amd64 -o meta-cc

# Install to system path
sudo mv meta-cc /usr/local/bin/meta-cc
sudo chmod +x /usr/local/bin/meta-cc

# Verify installation
meta-cc --version
```

#### 2. Configure Slash Commands

Slash Commands are already included in the project (`.claude/commands/` directory).

**Project-level Slash Commands** (recommended):

```bash
# Slash Commands are already in the project
ls .claude/commands/
# meta-stats.md
# meta-errors.md

# No additional configuration needed - just open in Claude Code
```

**Personal-level Slash Commands** (optional, available globally):

```bash
# Copy to personal Slash Commands directory
mkdir -p ~/.claude/commands
cp .claude/commands/meta-stats.md ~/.claude/commands/
cp .claude/commands/meta-errors.md ~/.claude/commands/

# Now available in all Claude Code projects
```

#### 3. Run Integration Tests

```bash
# Run integration test script
bash tests/integration/slash_commands_test.sh
```

Expected output:
```
=== meta-cc Slash Commands Integration Test ===

[1/5] Checking meta-cc installation...
✅ meta-cc installed: /usr/local/bin/meta-cc

[2/5] Checking Slash Command files...
✅ Slash Command files exist

[3/5] Testing meta-cc parse stats...
✅ meta-cc parse stats executed successfully

[4/5] Testing meta-cc analyze errors...
✅ meta-cc analyze errors executed successfully

[5/5] Testing meta-cc parse extract...
✅ meta-cc parse extract executed successfully

=== All tests passed ✅ ===
```

### Available Slash Commands

meta-cc provides **4 core Slash Commands** for quick analysis:

#### `/meta-stats` - Session Statistics

Display statistical information about the current Claude Code session.

**Usage**:
```
/meta-stats
```

**Output example**:
```markdown
# Session Statistics

- **Total Turns**: 245
- **Tool Calls**: 853
- **Error Count**: 0
- **Error Rate**: 0.00%
- **Session Duration**: 3h 42m

## Tool Usage Frequency

| Tool | Count | Percentage |
|------|-------|------------|
| Bash | 320 | 37.5% |
| Read | 198 | 23.2% |
| Edit | 156 | 18.3% |
```

**Use cases**:
- Quickly understand session overview
- Check for tool usage anomalies (high error rate)
- Evaluate session efficiency
- Discover tool usage patterns

#### `/meta-errors` - Error Pattern Analysis

Analyze error patterns in the current session, detecting repeated errors.

**Usage**:
```
/meta-errors              # Use default window (20 turns)
/meta-errors 50           # Analyze last 50 turns
/meta-errors 100          # Analyze last 100 turns
```

**Output example**:
```markdown
## Error Data Extraction

Detected 12 error tool calls.

## Error Pattern Analysis (window size: 20)

# Error Pattern Analysis

Found 2 error pattern(s):

## Pattern 1: Bash

- **Type**: repeated_error
- **Occurrences**: 5 times
- **Signature**: `a3f2b1c4d5e6f7g8`
- **Error**: command not found: xyz

### Context

- **First Occurrence**: 2025-10-02T10:00:00.000Z
- **Last Occurrence**: 2025-10-02T10:15:00.000Z
- **Time Span**: 900 seconds (15.0 minutes)
- **Affected Turns**: 5

---

## Optimization Recommendations

Based on detected error patterns, consider the following:

1. **Check root cause of repeated errors**
2. **Use Claude Code Hooks to prevent errors**
3. **Adjust workflow**
```

**Use cases**:
- Identify repeated errors to avoid redundant debugging
- Discover workflow bottlenecks (frequent failures)
- Get optimization recommendations (hooks, alternatives, prompt improvements)
- Focus on recent errors (using window parameter)

#### `/meta-timeline` - Session Timeline

Generate a chronological timeline view of the session.

**Usage**:
```
/meta-timeline           # Default: last 50 turns
/meta-timeline 20        # Last 20 turns
```

**Use cases**:
- Visualize session flow and tool usage patterns
- Identify time distribution of errors
- Understand workflow sequences

#### `/meta-help` - Help and Usage Guide

Display comprehensive help for all meta-cc features, including Slash Commands, MCP Server, and @meta-coach subagent.

**Usage**:
```
/meta-help
```

**Use cases**:
- Learn available commands and features
- Get quick reference for integration methods
- Troubleshoot common issues

---

### Recommended Usage Patterns

meta-cc offers **three integration tiers** for different use cases:

#### 🚀 Quick Statistics → Slash Commands (Lowest Priority)

For fast, pre-defined analyses:
```
/meta-stats              # Session overview
/meta-errors             # Error patterns
/meta-timeline           # Chronological view
```

#### 🔍 Natural Queries → MCP Server (Core Integration)

Ask Claude directly - MCP tools are called automatically:
```
"Show me all Bash errors in this project"
"Find user messages mentioning 'refactor'"
"Analyze tool usage trends across sessions"
"Compare error rates between different workflow phases"
```

Available MCP tools: 14 tools including `query_tools`, `analyze_errors`, `query_user_messages`, `aggregate_stats`, and more.

#### 🎓 Deep Analysis → @meta-coach Subagent (Highest Value)

For interactive coaching and workflow optimization:
```
@meta-coach Why do my tests keep failing?
@meta-coach Help me optimize my workflow
@meta-coach Analyze my efficiency bottlenecks
```

The @meta-coach subagent combines MCP data access with LLM reasoning to provide personalized guidance.

### Troubleshooting

#### Issue 1: "❌ Error: meta-cc not installed or not in PATH"

**Cause**: meta-cc binary not installed or not in system PATH.

**Solution**:
```bash
# Check if meta-cc is installed
which meta-cc

# If not found, install meta-cc
cd /path/to/meta-cc
make build
sudo cp meta-cc /usr/local/bin/meta-cc
sudo chmod +x /usr/local/bin/meta-cc

# Verify installation
meta-cc --version
```

#### Issue 2: "failed to locate session file"

**Cause**: meta-cc cannot find the current session's JSONL file.

**Solution**:
```bash
# Option 1: Check environment variables (Claude Code may provide)
echo $CC_SESSION_ID
echo $CC_PROJECT_HASH

# Option 2: Manually specify session file
meta-cc parse stats --session <session-id>

# Option 3: Check if session file exists
ls ~/.claude/projects/
```

#### Issue 3: Slash Commands not visible or unavailable

**Cause**: Slash Command files in wrong location or Claude Code not loaded.

**Solution**:
```bash
# Check if files exist
ls .claude/commands/meta-stats.md
ls .claude/commands/meta-errors.md

# Restart Claude Code
# Close and reopen the project

# Check file format (frontmatter must be correct)
head -n 10 .claude/commands/meta-stats.md
```

#### Issue 4: Garbled output or errors

**Cause**: meta-cc version mismatch or incorrect command parameters.

**Solution**:
```bash
# Check meta-cc version
meta-cc --version

# Ensure using latest version
cd /path/to/meta-cc
git pull
make build
sudo cp meta-cc /usr/local/bin/meta-cc

# Test commands manually
meta-cc parse stats --output md
meta-cc analyze errors --window 20 --output md
```

#### Issue 5: Permission errors

**Cause**: meta-cc doesn't have permission to read session files.

**Solution**:
```bash
# Check session file permissions
ls -l ~/.claude/projects/

# Ensure current user has read permission
chmod -R u+r ~/.claude/projects/
```

### Advanced Usage

#### Combining Integration Tiers

```bash
# Step 1: Quick view in Claude Code using Slash Command
/meta-stats
# /meta-stats

# Step 2: If high error rate found, use natural query (MCP)
"Show me the top 10 most frequent errors in this project"

# Step 3: For deep analysis, use @meta-coach
@meta-coach Why do I have so many Bash errors?

# Step 4: Export data for offline analysis (CLI)
meta-cc query errors --output jsonl > errors.jsonl
meta-cc query tools --status error --output tsv > error-tools.tsv
```

#### Creating Custom Slash Commands

Based on meta-cc, you can create custom Slash Commands:

**Example: `.claude/commands/meta-tool-usage.md`**

```markdown
---
name: meta-tool-usage
description: Display usage details for a specific tool
allowed_tools: [Bash]
argument-hint: [tool-name]
---

```bash
TOOL_NAME=${1:-Bash}
meta-cc parse extract --type tools --filter "tool=$TOOL_NAME" --output md
```
```

**Usage**:
```
/meta-tool-usage Bash
/meta-tool-usage Read
```

### Environment Variables

meta-cc supports the following environment variables (if provided by Claude Code):

- `CC_SESSION_ID`: Current session ID
- `CC_PROJECT_HASH`: Project path hash

**Check environment variables**:
```bash
# Check in Slash Command
echo "Session ID: $CC_SESSION_ID"
echo "Project Hash: $CC_PROJECT_HASH"
```

If these environment variables are unavailable, meta-cc will automatically fall back to working directory inference.

### Integration Options

meta-cc integrates with Claude Code in three ways:

- **MCP Server**: Seamless data access (Claude queries autonomously) - **14 tools available** (Phase 13: unified scope parameter)
- **Slash Commands**: Quick, pre-defined workflows (`/meta-stats`, `/meta-errors`, `/meta-timeline`, `/meta-help`) - 4 core commands
- **Subagent** (`@meta-coach`): Interactive, conversational analysis with Phase 10 capabilities

**👉 See the [Integration Guide](./docs/integration-guide.md)** for detailed comparison, decision framework, and best practices.

### Phase 13: Consolidated Scope Parameter Design

Phase 13 consolidates the dual-scope approach into a **unified `scope` parameter** for cleaner API:

**All MCP tools** (except `get_session_stats` for backward compatibility) now support a `scope` parameter:
- **`scope: "project"`** (default) - Query across all project sessions for meta-cognition insights
- **`scope: "session"`** - Limit query to current session only

**Example usage**:
```javascript
// Project-level analysis (default) - discovers patterns across all sessions
meta-cc.query_tools({ tool: "Bash", status: "error" })
// Equivalent to: { scope: "project", tool: "Bash", status: "error" }

// Session-only analysis - focuses on current session
meta-cc.query_tools({ scope: "session", tool: "Bash", status: "error" })
```

**Why project-level default?**
- **Meta-cognition requires cross-session analysis** to identify long-term patterns
- Discover recurring errors, workflow evolution, and systematic improvement opportunities
- Session-level analysis available via explicit `scope: "session"` when needed

**Backward compatibility**:
- Legacy `_session` suffix tools (e.g., `query_tools_session`) automatically map to `scope: "session"`
- `get_session_stats` remains session-only for compatibility

**👉 See the [MCP Project Scope Guide](./docs/mcp-project-scope.md)** for detailed usage examples and migration guide.

### MCP Tools

meta-cc provides **13 standardized MCP tools** for analyzing Claude Code session history. All tools support consistent parameters for filtering, statistics, and output control.

#### Standard Parameters

All tools support these core parameters:

- **`scope`**: Query scope - "project" (cross-session, default) or "session" (current only)
- **`jq_filter`**: jq expression for filtering and transforming results (default: ".[]")
- **`stats_only`**: Return only statistics, no detailed records (default: false)
- **`stats_first`**: Return statistics first, then details separated by `---` (default: false)
- **`max_output_bytes`**: Maximum output size in bytes (default: 51200 / 50KB)
- **`output_format`**: Output format - "jsonl" or "tsv" (default: "jsonl")

#### Tool Catalog

**Session Statistics**:
- `get_session_stats` - Get session statistics (default scope: session)

**Tool Analysis**:
- `query_tools` - Query tool calls with filters (tool, status)
- `extract_tools` - Extract complete tool call history
- `query_tools_advanced` - Advanced queries with SQL-like filters
- `query_tool_sequences` - Detect repeated workflow patterns

**Message & Context**:
- `query_user_messages` - Search user messages with regex
- `query_context` - Query context around errors (before/after turns)
- `query_successful_prompts` - Find successful prompt patterns

**File Operations**:
- `query_file_access` - Query file operation history
- `query_files` - File-level operation statistics

**Project & Time**:
- `query_project_state` - Query project state evolution
- `query_time_series` - Analyze metrics over time (hour/day/week)

**Deprecated**:
- `analyze_errors` - [DEPRECATED] Use `query_tools` with `status='error'` instead

#### Quick Examples

**Error Distribution**:
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})",
    "stats_only": true
  }
}
```

**File Hotspots**:
```json
{
  "name": "query_files",
  "arguments": {
    "sort_by": "total_ops",
    "top": 10
  }
}
```

**Workflow Patterns**:
```json
{
  "name": "query_tool_sequences",
  "arguments": {
    "min_occurrences": 5,
    "jq_filter": "sort_by(.Occurrences) | reverse | .[0:10]"
  }
}
```

**Time Series Analysis**:
```json
{
  "name": "query_time_series",
  "arguments": {
    "interval": "day",
    "metric": "error-rate"
  }
}
```

#### Complete Documentation

For comprehensive documentation including:
- All 13 tools with detailed examples
- jq expression cookbook (20+ patterns)
- Best practices and common pitfalls
- FAQ and troubleshooting

See **[MCP Tools Complete Reference](./docs/mcp-tools-reference.md)**

Also see:
- [MCP Usage Guide](./docs/mcp-usage.md) - Getting started
- [MCP Project Scope Guide](./docs/mcp-project-scope.md) - Scope parameter usage
- [Phase 15 Migration Guide](./docs/mcp-migration-phase15.md) - Upgrading from Phase 14

### Reference Documentation

- **[Integration Guide](./docs/integration-guide.md)** - Choose the right integration method
- [Examples & Usage](./docs/examples-usage.md) - Step-by-step setup guides
- [Troubleshooting Guide](./docs/troubleshooting.md) - Common issues and solutions
- [Technical Proposal](./docs/proposals/meta-cognition-proposal.md) - Architecture deep dive
- [Claude Code Documentation](https://docs.claude.com/en/docs/claude-code/overview) - Official docs

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on our development process, code style, and testing requirements.

Please also read and adhere to our [Code of Conduct](CODE_OF_CONDUCT.md).

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for build automation)

### Build

```bash
# Using Make
make build

# Or using go directly
go build -o meta-cc
```

### Test

```bash
# Run tests (short mode, skips slow E2E tests ~30s)
make test

# Run all tests including E2E tests
make test-all

# Run with coverage (includes all tests)
make test-coverage
# Open coverage.html in browser
```

**Note**: `make test` uses `-short` flag to skip E2E integration tests that take ~30 seconds. Use `make test-all` or `make test-coverage` for complete test coverage including E2E tests.

### Available Make Targets

```bash
make build           # Build CLI and MCP server
make build-cli       # Build CLI only
make build-mcp       # Build MCP server only
make test            # Run tests (short mode, skips slow E2E tests)
make test-all        # Run all tests (including slow E2E tests ~30s)
make test-coverage   # Run tests with coverage report (includes E2E tests)
make clean           # Remove build artifacts
make install         # Install CLI to GOPATH/bin
make cross-compile   # Build for all platforms
make deps            # Download and tidy dependencies
make help            # Show help message
```

## Supported Platforms

- Linux (amd64, arm64)
- macOS (amd64, arm64/Apple Silicon)
- Windows (amd64)

## Unix Composability (Phase 11)

Phase 11 optimizes meta-cc for seamless integration with Unix pipelines and standard tools following Unix philosophy principles.

### Key Features

1. **JSONL Streaming Output**: Stream data as JSON Lines for efficient pipeline processing
2. **Standard Exit Codes**: Unix-compliant exit codes (0=success, 1=error, 2=no results)
3. **Clean I/O Separation**: Logs to stderr, data to stdout - no pipeline interference
4. **Tool Integration**: Works seamlessly with jq, grep, awk, sed, and other Unix tools

### JSONL Streaming Output

Stream data for efficient pipeline processing:

```bash
# Basic streaming
meta-cc query tools --stream

# Pipeline with jq
meta-cc query tools --stream | jq 'select(.Status == "error")'

# Pipeline with grep
meta-cc query tools --stream | jq -r '.Error' | grep -i "permission"

# Pipeline with awk
meta-cc query tools --stream | \
  jq -r '[.ToolName, .Duration] | @tsv' | \
  awk '{sum+=$2} END {print "Total:", sum "ms"}'
```

**JSONL Format**:
```
{"uuid":"1","tool":"Bash","status":"success",...}
{"uuid":"2","tool":"Edit","status":"success",...}
{"uuid":"3","tool":"Read","status":"error",...}
```

### Standard Exit Codes

meta-cc follows Unix conventions for exit codes:

| Exit Code | Meaning | Example |
|-----------|---------|---------|
| 0 | Success (with results) | `meta-cc query tools --limit 10` |
| 1 | Error (parsing, I/O, etc.) | `meta-cc query tools --where "invalid syntax"` |
| 2 | Success (no results) | `meta-cc query tools --where "tool='NonExistent'"` |

**Usage in scripts**:
```bash
if meta-cc query tools --where "status='error'"; then
  echo "Errors found!"
  # Handle errors...
else
  EXIT_CODE=$?
  if [ $EXIT_CODE -eq 2 ]; then
    echo "No errors (good!)"
  else
    echo "Query failed"
    exit 1
  fi
fi
```

### stderr/stdout Separation

meta-cc separates logs and data for clean pipeline processing:

- **stdout**: Command output data (JSON, Markdown, TSV)
- **stderr**: Diagnostic messages (logs, warnings, errors)

```bash
# Redirect data only
meta-cc query tools --output json > data.json

# Redirect logs only
meta-cc query tools --output json 2> debug.log

# Separate both
meta-cc query tools --output json > data.json 2> debug.log

# Suppress logs in pipelines
meta-cc query tools --stream 2>/dev/null | jq '.ToolName'
```

### Common Pipeline Patterns

**Error Analysis**:
```bash
# Find top error patterns
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  grep -oP '(permission|timeout|not found)' | \
  sort | uniq -c | sort -rn
```

**Performance Profiling**:
```bash
# Average duration by tool
meta-cc stats aggregate --group-by tool --metrics avg_duration | \
  jq -r '.[] | [.group_value, .metrics.avg_duration] | @tsv' | \
  column -t
```

**Tool Usage Statistics**:
```bash
# Tool distribution
meta-cc query tools --stream | \
  jq -r '.ToolName' | \
  sort | uniq -c | sort -rn | \
  awk '{print $2 ": " $1}'
```

**File Modification Tracking**:
```bash
# Most edited files with error rates
meta-cc stats files --sort-by edit_count --top 10 | \
  jq -r '.[] | [.file_path, .edit_count, (.error_rate * 100 | tostring + "%")] | @tsv' | \
  column -t
```

### Unix Philosophy

Phase 11 embraces Unix principles:

- **Do one thing well**: Each command has a single, focused purpose
- **Text streams**: All data flows as structured text (JSON/JSONL)
- **Composability**: Tools chain together via pipes
- **Consistent interface**: Uniform command structure and behavior

### See Also

- [Cookbook](docs/cookbook.md) - 10+ practical analysis patterns
- [CLI Composability Guide](docs/cli-composability.md) - Integration with jq, grep, awk
- [Examples and Usage](docs/examples-usage.md) - Getting started guide

---

## Phase 10: Advanced Query Capabilities (New!)

Phase 10 introduces SQL-like filtering, aggregation, time series analysis, and file-level statistics for deeper insights.

### Key Features

1. **Advanced Filtering**: SQL-like expressions with AND/OR/NOT, IN, BETWEEN, LIKE, REGEXP
2. **Aggregation Statistics**: Group-by with metrics (count, error_rate, percentiles)
3. **Time Series Analysis**: Analyze metrics over time (hour/day/week intervals)
4. **File-Level Statistics**: Track file operations and identify hotspots

### Quick Examples

```bash
# Advanced filtering with SQL-like expressions
meta-cc query tools --filter "tool='Bash' AND status='error'"
meta-cc query tools --filter "tool IN ('Bash', 'Edit', 'Write')"
meta-cc query tools --filter "duration BETWEEN 500 AND 2000"

# Aggregation statistics
meta-cc stats aggregate --group-by tool --metrics "count,error_rate"
meta-cc stats aggregate --group-by status --metrics count

# Time series analysis
meta-cc stats time-series --metric tool-calls --interval hour
meta-cc stats time-series --metric error-rate --interval day

# File-level statistics
meta-cc stats files --sort-by edit_count --top 10
meta-cc stats files --sort-by error_count --filter "error_count>0"
```

### Enhanced Integration

Phase 10 features are fully integrated with:
- **Slash Commands**: `/meta-stats`, `/meta-errors`, `/meta-timeline` now use Phase 10 capabilities
- **MCP Server**: 4 new tools (`query_tools_advanced`, `aggregate_stats`, `query_time_series`, `query_files`)
- **@meta-coach**: Updated with Phase 10 analysis workflows and best practices

See [Phase 10 MCP Tools](#phase-10-mcp-tools) below for details.

---

## Phase 9: Context-Length Management

Phase 9 introduces powerful features to handle large sessions (>1000 turns) and prevent context overflow.

### Features

1. **Pagination**: Process data in chunks to avoid memory overflow
2. **Size Estimation**: Predict output size before generating
3. **Chunking**: Split large output into multiple files
4. **Field Projection**: Output only specified fields (70%+ size reduction)
5. **Compact Formats**: TSV format (86%+ smaller than JSON)
6. **Summary Mode**: Overview + top N details

### Usage Examples

#### 1. Pagination

```bash
# Get first 50 tools
meta-cc query tools --limit 50

# Skip first 100, get next 50
meta-cc query tools --limit 50 --offset 100

# Iterate through all tools in batches of 100
for i in {0..10}; do
  meta-cc query tools --limit 100 --offset $((i*100)) --output json
done
```

#### 2. Size Estimation

```bash
# Estimate output size before generating
meta-cc query tools --estimate-size

# Output:
# {
#   "estimated_bytes": 1107311,
#   "estimated_kb": 1081.36,
#   "format": "json",
#   "record_count": 246
# }

# Use in Slash Commands for adaptive strategy
ESTIMATE=$(meta-cc parse stats --estimate-size --output json)
SIZE=$(echo $ESTIMATE | jq '.estimated_kb')
if (( $(echo "$SIZE > 100" | bc -l) )); then
  meta-cc parse stats --summary-first --top 20
else
  meta-cc parse stats --output md
fi
```

#### 3. Chunking (Large Sessions)

```bash
# Split 2000 tools into chunks of 100 records each
meta-cc query tools --chunk-size 100 --output-dir /tmp/chunks

# Output:
# Generated 20 chunk(s)
#   Chunk 0: chunk_0001.json (100 records, 44KB)
#   Chunk 1: chunk_0002.json (100 records, 45KB)
#   ...
#   Chunk 19: chunk_0020.json (100 records, 44KB)
# Manifest: /tmp/chunks/manifest.json

# Check manifest
cat /tmp/chunks/manifest.json
# {
#   "total_records": 2000,
#   "chunk_size": 100,
#   "num_chunks": 20,
#   "chunks": [...]
# }

# Process chunks in parallel
ls /tmp/chunks/chunk_*.json | xargs -P 4 -I {} sh -c 'jq ".[] | select(.Status == \"error\")" {}'
```

#### 4. Field Projection (70%+ Size Reduction)

```bash
# Output only UUID, ToolName, Status (72.7% smaller)
meta-cc query tools --fields "UUID,ToolName,Status"

# With error fields conditionally included
meta-cc query tools --fields "UUID,ToolName,Status" --if-error-include "Error,Output"

# Size comparison
meta-cc query tools --limit 100 --output json | wc -c
# Output: 31101 bytes (30.4 KB)

meta-cc query tools --limit 100 --fields "UUID,ToolName,Status" --output json | wc -c
# Output: 8501 bytes (8.3 KB) - 72.7% reduction!
```

#### 5. TSV Format (86%+ Smaller)

```bash
# TSV output (86.4% smaller than JSON)
meta-cc query tools --output tsv

# Output:
# UUID\tToolName\tStatus\tError
# 1b08...\tRead
# 69a7...\tBash
# 586a...\tBash

# Combine with other features
meta-cc query tools --limit 100 --fields "UUID,ToolName,Status" --output tsv

# Pipe to other tools
meta-cc query tools --output tsv | cut -f2 | sort | uniq -c
# Count tool usage
```

#### 6. Summary Mode

```bash
# Summary + top 10 details
meta-cc query tools --summary-first --top 10

# Output:
# === Session Summary ===
# Total Tools: 246
# Errors: 0 (0.0%)
#
# Top Tools:
#   1. Bash (102)
#   2. Read (37)
#   3. TodoWrite (37)
#   ...
#
# [Top 10 detailed records follow]

# JSON format with summary
meta-cc query tools --summary-first --top 5 --output json
```

#### 7. Combined Features

```bash
# Pagination + Projection + TSV
meta-cc query tools --limit 50 --fields "ToolName,Status" --output tsv

# Chunking + TSV (ultra-compact for large sessions)
meta-cc query tools --chunk-size 100 --output-dir ./chunks --output tsv

# Summary + Projection + JSON
meta-cc query tools --summary-first --top 10 --fields "UUID,ToolName" --output json
```

### Large Session Best Practices

**Problem**: Sessions with >1000 turns can cause:
- Context overflow in Claude Code (>200K tokens)
- Memory issues during processing
- Slow command execution

**Solution**: Use Phase 9 features adaptively

#### Strategy Selection Matrix

| Session Size | Recommended Strategy | Example Command |
|-------------|---------------------|----------------|
| < 500 turns | Standard output | `meta-cc query tools --output json` |
| 500-1000 turns | Pagination or Projection | `meta-cc query tools --limit 200 --fields "UUID,ToolName,Status"` |
| 1000-2000 turns | Summary + TSV | `meta-cc query tools --summary-first --top 20 --output tsv` |
| > 2000 turns | Chunking + TSV | `meta-cc query tools --chunk-size 100 --output-dir ./chunks --output tsv` |

#### Adaptive Slash Command Pattern

```bash
# Estimate first, then choose strategy
SIZE=$(meta-cc query tools --estimate-size --output json | jq '.estimated_kb')

if (( $(echo "$SIZE < 50" | bc -l) )); then
  # Small: full output
  meta-cc query tools --output md
elif (( $(echo "$SIZE < 200" | bc -l) )); then
  # Medium: pagination + projection
  meta-cc query tools --limit 100 --fields "ToolName,Status,UUID" --output tsv
else
  # Large: summary mode
  meta-cc query tools --summary-first --top 20 --output tsv
fi
```

### Performance Metrics

| Feature | Size Reduction | Use Case |
|---------|---------------|----------|
| Field Projection (3 fields) | **72.7%** | Reduce JSON size while preserving key data |
| TSV Format | **86.4%** | Ultra-compact tabular output |
| Summary Mode | **~95%** | Overview for very large sessions |
| Chunking | N/A | Split data for parallel processing |

### Migration Guide

**Before Phase 9** (Old way):
```bash
# Gets all tools (may overflow context with >1000 turns)
meta-cc query tools --output json
```

**After Phase 9** (Recommended):
```bash
# Option 1: Estimate first
meta-cc query tools --estimate-size

# Option 2: Use pagination
meta-cc query tools --limit 100

# Option 3: Use field projection
meta-cc query tools --fields "UUID,ToolName,Status"

# Option 4: Use TSV for maximum compression
meta-cc query tools --output tsv

# Option 5: Use summary for quick overview
meta-cc query tools --summary-first --top 20
```

## Project Structure

```
meta-cc/
├── cmd/              # Command definitions (Cobra)
├── internal/         # Internal packages
│   └── testutil/    # Test utilities
├── pkg/              # Public packages
├── tests/            # Test files and fixtures
└── docs/             # Documentation
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
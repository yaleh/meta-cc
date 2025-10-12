# CLI Reference

Complete command-line reference for meta-cc.

## Global Options

These options work with all commands:

```bash
meta-cc [global options] <command> [command options]
```

### Session and Project

- `--session <session-id>` - Specify session ID (auto-detected by default)
- `--project <path>` - Specify project path (auto-detected by default)

### Output Control

- `--output <format>` - Output format: `jsonl` (default), `tsv`, `md`
- `--limit <N>` - Limit output to N records
- `--offset <M>` - Skip first M records
- `--fields <list>` - Output only specified fields (comma-separated)
- `--if-error-include <list>` - Include extra fields on errors

### Size Management

- `--estimate-size` - Predict output size before generating
- `--chunk-size <N>` - Split output into chunks of N records
- `--output-dir <dir>` - Output directory for chunks
- `--summary-first` - Show summary before details
- `--top <N>` - Show only top N results (with summary)

### Query Scope

- `--scope <scope>` - Query scope: `project` (default) or `session`

## Commands

### parse stats

Get comprehensive session statistics.

```bash
meta-cc parse stats [options]
```

**Output** (JSONL object):
```json
{
  "TurnCount": 2676,
  "UserTurnCount": 1097,
  "AssistantTurnCount": 1579,
  "ToolCallCount": 1012,
  "ErrorCount": 0,
  "ErrorRate": 0,
  "DurationSeconds": 33796,
  "ToolFrequency": {"Bash": 495, "Read": 162},
  "TopTools": [{"Name": "Bash", "Count": 495, "Percentage": 48.9}]
}
```

**Examples**:
```bash
# Basic stats
meta-cc parse stats

# Markdown output
meta-cc parse stats --output md

# Estimate size first
meta-cc parse stats --estimate-size
```

### parse extract

Extract raw session data.

```bash
meta-cc parse extract --type <type> [options]
```

**Types**:
- `turns` - All conversation turns
- `tools` - All tool calls

**Examples**:
```bash
# Extract all tool calls
meta-cc parse extract --type tools

# Extract user turns only
meta-cc parse extract --type turns | jq 'select(.type == "user")'

# Limit to 50 tools
meta-cc parse extract --type tools --limit 50
```

### query tools

Query tool calls with flexible filtering.

```bash
meta-cc query tools [options]
```

**Options**:
- `--tool <name>` - Filter by tool name (e.g., "Bash", "Read", "Edit")
- `--status <status>` - Filter by status: `error` or `success`
- `--since <time>` - Filter by time (e.g., "5 minutes ago", "2 hours ago")
- `--last-n-turns <N>` - Only tools from last N turns

**Examples**:
```bash
# All tool calls
meta-cc query tools

# Bash errors only
meta-cc query tools --tool Bash --status error

# Recent tools (last 5 minutes)
meta-cc query tools --since "5 minutes ago"

# Last 10 turns
meta-cc query tools --last-n-turns 10

# Compact output (selected fields only)
meta-cc query tools --fields "UUID,ToolName,Status"

# TSV output for Unix tools
meta-cc query tools --output tsv | cut -f2 | sort | uniq -c
```

### query user-messages

Search user messages with regex patterns.

```bash
meta-cc query user-messages --pattern <regex> [options]
```

**Options**:
- `--pattern <regex>` - Regex pattern to match (required)
- `--limit <N>` - Maximum results (default: unlimited, uses hybrid mode)
- `--with-context <N>` - Include N surrounding turns

**Examples**:
```bash
# Search for "fix" followed by "bug"
meta-cc query user-messages --pattern "fix.*bug"

# Search for "error" or "bug"
meta-cc query user-messages --pattern "error|bug"

# Limit to 10 results
meta-cc query user-messages --pattern "implement" --limit 10

# With context (3 turns before/after)
meta-cc query user-messages --pattern "refactor" --with-context 3
```

### query context

Query context around errors or specific tool calls.

```bash
meta-cc query context --error-signature <id> [options]
```

**Options**:
- `--error-signature <id>` - Error pattern ID (from analyze errors)
- `--window <N>` - Context window size (default: 3 turns before/after)

**Examples**:
```bash
# Get context around error pattern
meta-cc query context --error-signature abc123 --window 3

# Larger context window
meta-cc query context --error-signature abc123 --window 5
```

### query file-access

Query file operation history.

```bash
meta-cc query file-access --file <path> [options]
```

**Options**:
- `--file <path>` - File path to query (required)

**Examples**:
```bash
# Operations on specific file
meta-cc query file-access --file src/main.go

# Statistics by operation type
meta-cc query file-access --file src/main.go | \
  jq 'group_by(.Operation) | map({op: .[0].Operation, count: length})'
```

### query tool-sequences

Detect repeated workflow patterns.

```bash
meta-cc query tool-sequences [options]
```

**Options**:
- `--min-occurrences <N>` - Minimum pattern occurrences (default: 3)
- `--pattern <pattern>` - Filter by tool sequence pattern (regex)
- `--include-builtin-tools` - Include built-in tools (Bash, Read, etc.)

**Examples**:
```bash
# Find all repeated sequences (MCP tools only)
meta-cc query tool-sequences --min-occurrences 3

# Include built-in tools
meta-cc query tool-sequences --min-occurrences 3 --include-builtin-tools

# Find specific pattern
meta-cc query tool-sequences --pattern "Read.*Edit"
```

### analyze errors

Analyze error patterns (deprecated - use `query tools --status error`).

```bash
meta-cc analyze errors [options]
```

**Options**:
- `--window <N>` - Analysis window size (default: 20 turns)

**Examples**:
```bash
# Basic error analysis
meta-cc analyze errors

# Larger analysis window
meta-cc analyze errors --window 50

# Modern alternative (recommended)
meta-cc query tools --status error
```

### analyze sequences

Detect repeated tool sequences.

```bash
meta-cc analyze sequences [options]
```

**Options**:
- `--min-length <N>` - Minimum sequence length (default: 3)
- `--min-occurrences <N>` - Minimum occurrences (default: 3)

**Examples**:
```bash
# Detect sequences of 3+ tools, occurring 3+ times
meta-cc analyze sequences --min-length 3 --min-occurrences 3

# Longer sequences
meta-cc analyze sequences --min-length 5 --min-occurrences 2
```

### analyze file-churn

Detect frequently modified files.

```bash
meta-cc analyze file-churn [options]
```

**Options**:
- `--threshold <N>` - Minimum edit count (default: 5)

**Examples**:
```bash
# Files edited 5+ times
meta-cc analyze file-churn --threshold 5

# High-churn files (10+ edits)
meta-cc analyze file-churn --threshold 10
```

### analyze idle-periods

Analyze time gaps in session activity.

```bash
meta-cc analyze idle-periods [options]
```

**Options**:
- `--threshold <duration>` - Minimum idle duration (e.g., "5 minutes")

**Examples**:
```bash
# Idle periods longer than 5 minutes
meta-cc analyze idle-periods --threshold "5 minutes"

# Long breaks (1+ hour)
meta-cc analyze idle-periods --threshold "1 hour"
```

### stats aggregate

Aggregate statistics with grouping.

```bash
meta-cc stats aggregate [options]
```

**Options**:
- `--group-by <field>` - Group by field: `tool`, `status`, `hour`, `day`
- `--metrics <list>` - Metrics to compute: `count`, `error_rate`, `avg_duration`

**Examples**:
```bash
# Tool usage statistics
meta-cc stats aggregate --group-by tool --metrics count

# Error rate by tool
meta-cc stats aggregate --group-by tool --metrics "count,error_rate"

# Hourly activity
meta-cc stats aggregate --group-by hour --metrics count
```

### stats time-series

Analyze metrics over time.

```bash
meta-cc stats time-series [options]
```

**Options**:
- `--metric <metric>` - Metric to analyze: `tool-calls`, `error-rate`, `duration`
- `--interval <interval>` - Time interval: `hour`, `day`, `week`
- `--where <filter>` - SQL-like filter expression

**Examples**:
```bash
# Tool calls per hour
meta-cc stats time-series --metric tool-calls --interval hour

# Error rate per day
meta-cc stats time-series --metric error-rate --interval day

# Bash tool calls per hour
meta-cc stats time-series --metric tool-calls --interval hour --where "tool='Bash'"
```

### stats files

File-level operation statistics.

```bash
meta-cc stats files [options]
```

**Options**:
- `--sort-by <field>` - Sort by: `total_ops`, `edit_count`, `error_count`
- `--top <N>` - Show only top N files
- `--filter <expr>` - Filter expression (e.g., "error_count>0")

**Examples**:
```bash
# Top 10 most edited files
meta-cc stats files --sort-by edit_count --top 10

# Files with errors
meta-cc stats files --filter "error_count>0"

# All file stats
meta-cc stats files
```

## Output Formats

### JSONL (JSON Lines) - Default

One JSON object per line, suitable for streaming and pipeline processing.

```bash
meta-cc query tools --output jsonl
```

**Usage with jq**:
```bash
# Count results
meta-cc query tools | jq -s 'length'

# Filter by field
meta-cc query tools | jq 'select(.Status == "error")'

# Extract field
meta-cc query tools | jq -r '.ToolName'
```

### TSV (Tab-Separated Values)

Compact tabular format, ideal for Unix tools.

```bash
meta-cc query tools --output tsv
```

**Usage with Unix tools**:
```bash
# Extract column 2
meta-cc query tools --output tsv | cut -f2

# Count unique values
meta-cc query tools --output tsv | cut -f2 | sort | uniq -c

# Format as table
meta-cc query tools --output tsv | column -t
```

### Markdown

Human-readable format for reports.

```bash
meta-cc parse stats --output md
```

## Exit Codes

meta-cc follows Unix conventions:

| Exit Code | Meaning | Example |
|-----------|---------|---------|
| 0 | Success (with results) | `meta-cc query tools --limit 10` |
| 1 | Error (parsing, I/O, etc.) | `meta-cc query tools --where "invalid syntax"` |
| 2 | Success (no results) | `meta-cc query tools --where "tool='NonExistent'"` |

**Usage in scripts**:
```bash
if meta-cc query tools --status error; then
  echo "Errors found!"
else
  EXIT_CODE=$?
  if [ $EXIT_CODE -eq 2 ]; then
    echo "No errors found (good!)"
  else
    echo "Query failed"
    exit 1
  fi
fi
```

## Common Pipeline Patterns

### Error Analysis

```bash
# Top error patterns
meta-cc query tools --status error | \
  jq -r '.Error' | \
  grep -oP '(permission|timeout|not found)' | \
  sort | uniq -c | sort -rn
```

### Tool Usage Statistics

```bash
# Tool distribution
meta-cc query tools | \
  jq -r '.ToolName' | \
  sort | uniq -c | sort -rn
```

### File Modification Tracking

```bash
# Most edited files with error rates
meta-cc stats files --sort-by edit_count --top 10 | \
  jq -r '.[] | [.file_path, .edit_count, (.error_rate * 100 | tostring + "%")] | @tsv' | \
  column -t
```

### Performance Profiling

```bash
# Average duration by tool
meta-cc stats aggregate --group-by tool --metrics avg_duration | \
  jq -r '.[] | [.group_value, .metrics.avg_duration] | @tsv' | \
  column -t
```

## See Also

- [MCP Guide](mcp-guide.md) - MCP tool reference for Claude Code integration
- [CLI Composability](cli-composability.md) - Advanced pipeline patterns
- [JSONL Reference](jsonl-reference.md) - Detailed output format documentation
- [Examples & Usage](examples-usage.md) - Step-by-step tutorials

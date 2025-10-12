# Feature Overview

This document provides a comprehensive overview of meta-cc's advanced features.

## Core Capabilities

### Natural Language Interface

The `/meta` command provides natural language capability discovery:

```bash
/meta "show errors"              # Error analysis
/meta "find repeated workflows"   # Pattern detection
/meta "which files change most"   # File operation stats
/meta "quality check"             # Code quality scan
```

**How it works**:
- Capabilities loaded from GitHub (or custom sources)
- Semantic keyword matching scores each capability
- Best match executed automatically

**Customization**:
```bash
# Use custom capabilities
export META_CC_CAPABILITY_SOURCES="~/my-caps:commands"

# Pin to specific version
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"
```

See [Capabilities Guide](../guides/capabilities.md) for details.

### MCP Integration

16 MCP tools enable Claude to autonomously query session data:

**Basic Queries**:
- `get_session_stats` - Session statistics
- `query_tools` - Filter tool calls
- `query_user_messages` - Search messages
- `query_assistant_messages` - Search assistant responses
- `query_conversation` - Search full conversation
- `query_files` - File operation statistics

**Advanced Queries**:
- `query_context` - Error context with surrounding turns
- `query_tool_sequences` - Workflow pattern detection
- `query_file_access` - File operation history
- `query_project_state` - Project evolution tracking
- `query_successful_prompts` - High-quality prompt patterns
- `query_tools_advanced` - SQL-like filtering
- `query_time_series` - Metrics over time

**Query Scope**:
- `scope: "project"` (default) - Cross-session analysis
- `scope: "session"` - Current session only

See [MCP Guide](../guides/mcp.md) for complete reference.

### Interactive Coaching

The `@meta-coach` subagent provides personalized workflow optimization:

```bash
@meta-coach Why do my tests keep failing?
@meta-coach Help me optimize my workflow
@meta-coach Analyze my efficiency bottlenecks
```

**Capabilities**:
- Error pattern analysis
- Workflow optimization recommendations
- Multi-turn interactive coaching
- Combines MCP data with LLM reasoning

---

## Context-Length Management

Handle large sessions (>1000 turns) without context overflow.

### 1. Pagination

Process data in manageable chunks:

```bash
# Get first 50 tools
meta-cc query tools --limit 50

# Skip first 100, get next 50
meta-cc query tools --limit 50 --offset 100

# Iterate through all tools
for i in {0..10}; do
  meta-cc query tools --limit 100 --offset $((i*100))
done
```

**Use case**: Exploring large sessions incrementally.

### 2. Size Estimation

Predict output size before generating:

```bash
meta-cc query tools --estimate-size
```

**Output**:
```json
{
  "estimated_bytes": 1107311,
  "estimated_kb": 1081.36,
  "format": "json",
  "record_count": 246
}
```

**Use case**: Adaptive Slash Commands that choose strategy based on size.

**Example**:
```bash
SIZE=$(meta-cc query tools --estimate-size | jq '.estimated_kb')
if (( $(echo "$SIZE > 100" | bc -l) )); then
  meta-cc query tools --summary-first --top 20
else
  meta-cc query tools --output md
fi
```

### 3. Chunking

Split large output into multiple files:

```bash
meta-cc query tools --chunk-size 100 --output-dir /tmp/chunks
```

**Output**:
```
Generated 20 chunk(s)
  Chunk 0: chunk_0001.json (100 records, 44KB)
  Chunk 1: chunk_0002.json (100 records, 45KB)
  ...
Manifest: /tmp/chunks/manifest.json
```

**Use case**: Process massive sessions in parallel.

**Example**:
```bash
# Process chunks in parallel
ls /tmp/chunks/chunk_*.json | \
  xargs -P 4 -I {} sh -c 'jq ".[] | select(.Status == \"error\")" {}'
```

### 4. Field Projection

Output only specified fields (70%+ size reduction):

```bash
# Basic projection
meta-cc query tools --fields "UUID,ToolName,Status"

# With conditional error fields
meta-cc query tools --fields "UUID,ToolName,Status" --if-error-include "Error,Output"
```

**Size comparison**:
```bash
# Full output
meta-cc query tools --limit 100 | wc -c
# 31101 bytes (30.4 KB)

# Projected output
meta-cc query tools --limit 100 --fields "UUID,ToolName,Status" | wc -c
# 8501 bytes (8.3 KB) - 72.7% reduction
```

**Use case**: Reduce token consumption while preserving key data.

### 5. Compact Formats

TSV format is 86%+ smaller than JSON:

```bash
meta-cc query tools --output tsv
```

**Output**:
```
UUID	ToolName	Status	Error
1b08...	Read
69a7...	Bash
586a...	Bash
```

**Usage with Unix tools**:
```bash
# Count tool usage
meta-cc query tools --output tsv | cut -f2 | sort | uniq -c

# Extract specific column
meta-cc query tools --output tsv | awk '{print $2}'
```

**Use case**: CLI processing and automation.

### 6. Summary Mode

Overview + top N details:

```bash
meta-cc query tools --summary-first --top 10
```

**Output**:
```markdown
=== Session Summary ===
Total Tools: 246
Errors: 0 (0.0%)

Top Tools:
  1. Bash (102)
  2. Read (37)
  3. TodoWrite (37)
  ...

[Top 10 detailed records follow]
```

**Use case**: Quick overview for very large sessions.

### Strategy Selection

| Session Size | Recommended Strategy | Example |
|-------------|---------------------|---------|
| < 500 turns | Standard output | `meta-cc query tools` |
| 500-1000 | Pagination or Projection | `meta-cc query tools --limit 200 --fields "UUID,ToolName,Status"` |
| 1000-2000 | Summary + TSV | `meta-cc query tools --summary-first --top 20 --output tsv` |
| > 2000 | Chunking + TSV | `meta-cc query tools --chunk-size 100 --output-dir ./chunks --output tsv` |

---

## Advanced Query Capabilities

SQL-like filtering, aggregation, and time series analysis.

### 1. Advanced Filtering

SQL-like expressions with AND/OR/NOT, IN, BETWEEN, LIKE, REGEXP:

```bash
# AND conditions
meta-cc query tools --where "tool='Bash' AND status='error'"

# IN operator
meta-cc query tools --where "tool IN ('Bash', 'Edit', 'Write')"

# BETWEEN operator
meta-cc query tools --where "duration BETWEEN 500 AND 2000"

# LIKE operator
meta-cc query tools --where "error LIKE '%permission%'"

# REGEXP operator
meta-cc query tools --where "error REGEXP 'timeout|connection'"

# Complex expressions
meta-cc query tools --where "(tool='Bash' OR tool='Edit') AND status='error'"
```

**Use case**: Precise filtering for complex analyses.

### 2. Aggregation Statistics

Group-by with metrics:

```bash
# Tool usage by name
meta-cc stats aggregate --group-by tool --metrics "count,error_rate"

# Hourly activity
meta-cc stats aggregate --group-by hour --metrics count

# Error distribution
meta-cc stats aggregate --group-by status --metrics count
```

**Available metrics**:
- `count` - Total occurrences
- `error_rate` - Percentage of errors
- `avg_duration` - Average execution time
- `p50`, `p95`, `p99` - Duration percentiles

**Output example**:
```json
[
  {
    "group_value": "Bash",
    "metrics": {
      "count": 495,
      "error_rate": 0.02,
      "avg_duration": 234.5
    }
  }
]
```

**Use case**: Statistical analysis and reporting.

### 3. Time Series Analysis

Analyze metrics over time (hour/day/week):

```bash
# Tool calls per hour
meta-cc stats time-series --metric tool-calls --interval hour

# Error rate per day
meta-cc stats time-series --metric error-rate --interval day

# Bash usage per week
meta-cc stats time-series --metric tool-calls --interval week --where "tool='Bash'"
```

**Available metrics**:
- `tool-calls` - Number of tool invocations
- `error-rate` - Error percentage
- `avg-duration` - Average execution time

**Output example**:
```json
[
  {
    "timestamp": "2025-10-02T10:00:00Z",
    "interval": "hour",
    "metric": "tool-calls",
    "value": 45
  }
]
```

**Use case**: Trend analysis and workflow evolution tracking.

### 4. File-Level Statistics

Track file operations and identify hotspots:

```bash
# Top 10 most edited files
meta-cc stats files --sort-by edit_count --top 10

# Files with errors
meta-cc stats files --filter "error_count>0"

# High-churn files (10+ edits)
meta-cc stats files --filter "edit_count>=10"
```

**Output example**:
```json
[
  {
    "file_path": "src/main.go",
    "total_ops": 42,
    "read_count": 15,
    "edit_count": 23,
    "write_count": 4,
    "error_count": 0,
    "error_rate": 0
  }
]
```

**Use case**: Identify file churn and refactoring opportunities.

---

## Unix Composability

Seamless integration with Unix pipelines and standard tools.

### 1. JSONL Streaming Output

Stream data as JSON Lines for efficient pipeline processing:

```bash
# Basic streaming
meta-cc query tools --output jsonl

# Pipeline with jq
meta-cc query tools | jq 'select(.Status == "error")'

# Pipeline with grep
meta-cc query tools | jq -r '.Error' | grep -i "permission"

# Pipeline with awk
meta-cc query tools | \
  jq -r '[.ToolName, .Duration] | @tsv' | \
  awk '{sum+=$2} END {print "Total:", sum "ms"}'
```

**Benefits**:
- Constant memory usage (stream, don't slurp)
- Works with any Unix tool
- Efficient for large datasets

### 2. Standard Exit Codes

Unix-compliant exit codes:

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
    echo "No errors (good!)"
  else
    echo "Query failed"
    exit 1
  fi
fi
```

### 3. Clean I/O Separation

Logs and data separated for clean pipeline processing:

- **stdout**: Command output data (JSON, TSV, Markdown)
- **stderr**: Diagnostic messages (logs, warnings, errors)

```bash
# Redirect data only
meta-cc query tools > data.json

# Redirect logs only
meta-cc query tools 2> debug.log

# Separate both
meta-cc query tools > data.json 2> debug.log

# Suppress logs in pipelines
meta-cc query tools 2>/dev/null | jq '.ToolName'
```

### 4. Common Pipeline Patterns

**Error Analysis**:
```bash
# Top error patterns
meta-cc query tools --status error | \
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
meta-cc query tools | \
  jq -r '.ToolName' | \
  sort | uniq -c | sort -rn
```

**File Modification Tracking**:
```bash
# Most edited files with error rates
meta-cc stats files --sort-by edit_count --top 10 | \
  jq -r '.[] | [.file_path, .edit_count, (.error_rate * 100 | tostring + "%")] | @tsv' | \
  column -t
```

---

## Workflow Pattern Detection

Identify repeated sequences and optimize workflows.

### Tool Sequence Detection

Find repeated workflow patterns:

```bash
# Find all patterns (3+ occurrences)
meta-cc query tool-sequences --min-occurrences 3

# Include built-in tools (Bash, Read, Edit)
meta-cc query tool-sequences --min-occurrences 3 --include-builtin-tools

# Find specific pattern
meta-cc query tool-sequences --pattern "Read.*Edit"
```

**Output example**:
```json
[
  {
    "Pattern": "query_tools → query_user_messages → get_session_stats",
    "Occurrences": 8,
    "ToolNames": ["query_tools", "query_user_messages", "get_session_stats"],
    "FirstSeen": "2025-10-02T10:00:00Z",
    "LastSeen": "2025-10-02T14:30:00Z"
  }
]
```

**Use case**: Identify repeated workflows that could be automated.

**Note**: By default, built-in tools (Bash, Read, Edit, etc.) are excluded for cleaner patterns. Use `--include-builtin-tools` for complete analysis.

### File Churn Analysis

Detect frequently modified files:

```bash
# Files edited 5+ times
meta-cc analyze file-churn --threshold 5

# High-churn files (10+ edits)
meta-cc analyze file-churn --threshold 10
```

**Use case**: Identify refactoring candidates or architectural issues.

### Idle Period Detection

Find time gaps in session activity:

```bash
# Idle periods > 5 minutes
meta-cc analyze idle-periods --threshold "5 minutes"

# Long breaks (1+ hour)
meta-cc analyze idle-periods --threshold "1 hour"
```

**Use case**: Understand workflow interruptions and blockers.

---

## Extensibility

Create custom capabilities with simple markdown files.

### Capability Structure

```markdown
---
name: my-feature
description: My custom analysis
keywords: custom, analysis, example
category: analysis
---

# My Custom Feature

Analyze custom patterns in session data.

## Implementation

\`\`\`bash
meta-cc query tools --tool Bash --status error | \
  jq 'select(.Error | test("permission"))'
\`\`\`

## Usage

Run with:

\`\`\`
/meta "my feature"
\`\`\`
```

### Multi-Source Configuration

Load capabilities from multiple sources:

```bash
# Local development + GitHub
export META_CC_CAPABILITY_SOURCES="~/my-caps:yaleh/meta-cc@main/commands"

# Package file + GitHub
export META_CC_CAPABILITY_SOURCES="./caps.tar.gz:yaleh/meta-cc@main/commands"

# Priority: left = highest (overrides)
export META_CC_CAPABILITY_SOURCES="~/dev/caps:yaleh/meta-cc@v1.0.0/commands"
```

**Source types**:
- **Local directories**: Immediate reflection, no cache
- **Package files** (`.tar.gz`): Cached with TTL
- **GitHub repositories**: jsDelivr CDN with smart caching

### Caching Strategy

**Branches** (mutable, 1-hour cache):
```bash
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@develop/commands"
```

**Tags** (immutable, 7-day cache):
```bash
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"
```

**Package files**:
- Release packages: 7-day cache
- Custom packages: 1-hour cache

**Local sources**: No cache (always fresh)

---

## Performance Optimization

### Memory Efficiency

**Streaming** (constant memory):
```bash
meta-cc query tools | jq 'select(.Status == "error")'
```

**Slurping** (loads all into memory):
```bash
meta-cc query tools | jq -s 'length'
```

**Recommendation**: Use streaming when possible, slurp only for array operations.

### Large Dataset Strategies

For sessions with 1000+ tools:

1. **Field projection** (70% size reduction):
   ```bash
   meta-cc query tools --fields "UUID,ToolName,Status"
   ```

2. **TSV format** (86% smaller):
   ```bash
   meta-cc query tools --output tsv
   ```

3. **Pagination**:
   ```bash
   meta-cc query tools --limit 100 --offset 0
   ```

4. **Summary mode**:
   ```bash
   meta-cc query tools --summary-first --top 20
   ```

### Built-in Tool Filtering

Query tool sequences excludes built-in tools by default (35x faster):

```bash
# Fast (excludes Bash, Read, Edit, etc.)
meta-cc query tool-sequences --min-occurrences 3

# Slower but complete
meta-cc query tool-sequences --min-occurrences 3 --include-builtin-tools
```

**Built-in tools** (14 total):
- File operations: Bash, Read, Edit, Write, Glob, Grep
- Task management: TodoWrite, Task
- Web operations: WebFetch, WebSearch
- Other: SlashCommand, BashOutput, NotebookEdit, ExitPlanMode

---

## See Also

- [CLI Reference](cli.md) - Complete command list
- [MCP Guide](../guides/mcp.md) - MCP tool integration
- [CLI Composability](../tutorials/cli-composability.md) - Advanced Unix patterns
- [Capabilities Guide](../guides/capabilities.md) - Create custom capabilities
- [Integration Guide](../guides/integration.md) - Choose MCP vs Slash vs Subagent

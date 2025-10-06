# CLI Composability: Integrating meta-cc with Unix Tools

This guide demonstrates how to integrate meta-cc with standard Unix tools for powerful data analysis workflows.

## Phase 13: Output Format Simplification

meta-cc now supports only **two output formats** for maximum Unix composability:

- **JSONL (default)**: One JSON object per line - perfect for streaming and jq processing
- **TSV**: Tab-separated values - optimal for awk, grep, cut processing

Removed formats (use alternatives):
- ~~JSON (pretty)~~ → Use `meta-cc ... | jq '.'`
- ~~CSV~~ → Use `--output tsv` instead
- ~~Markdown~~ → Let Claude Code render JSONL

## Table of Contents

1. [jq Integration](#jq-integration)
2. [grep Integration](#grep-integration)
3. [awk Integration](#awk-integration)
4. [sed Integration](#sed-integration)
5. [Combining Multiple Tools](#combining-multiple-tools)
6. [Performance Tips](#performance-tips)
7. [Troubleshooting](#troubleshooting)

---

## jq Integration

jq is a lightweight JSON processor, perfect for filtering and transforming meta-cc output.

### Basic Filtering

```bash
# Select errors only (JSONL default - one object per line)
meta-cc query tools | jq 'select(.Status == "error")'

# Select specific tool
meta-cc query tools | jq 'select(.ToolName == "Bash")'

# Multiple conditions
meta-cc query tools | \
  jq 'select(.ToolName == "Bash" and .Status == "error")'
```

### Field Extraction

```bash
# Extract tool names only (JSONL - one per line)
meta-cc query tools | jq -r '.ToolName'

# Extract multiple fields as TSV
meta-cc query tools | \
  jq -r '[.ToolName, .Status, .Duration] | @tsv'

# Create custom objects
meta-cc query tools | \
  jq '{tool: .ToolName, failed: (.Status == "error"), duration_sec: (.Duration / 1000)}'
```

### Aggregation

```bash
# Count by tool (using jq slurp to create array from JSONL)
meta-cc query tools | jq -s \
  'group_by(.ToolName) |
   map({tool: .[0].ToolName, count: length}) |
   sort_by(.count) | reverse'

# Average duration by tool
meta-cc query tools | jq -s \
  'group_by(.ToolName) |
   map({tool: .[0].ToolName,
        avg_duration: (map(.Duration) | add / length)})'
```

### Conditional Processing

```bash
# Different output based on status
meta-cc query tools | \
  jq 'if .Status == "error" then
        {error: .ToolName, message: .Error}
      else
        {success: .ToolName}
      end'

# Add severity field
meta-cc query tools | \
  jq '. + if .Duration > 5000 then
             {severity: "high"}
           elif .Duration > 2000 then
             {severity: "medium"}
           else
             {severity: "low"}
           end'
```

---

## grep Integration

grep is excellent for pattern matching and filtering text output.

### Pattern Matching

```bash
# Find permission errors
meta-cc query tools --where "status='error'" | \
  jq -r '.Error' | \
  grep -i "permission"

# Case-insensitive search
meta-cc query tools | \
  jq -r '.ToolName + ": " + .Error' | \
  grep -i "timeout\|failed\|denied"

# Invert match (exclude pattern)
meta-cc query tools | \
  jq -r '.ToolName' | \
  grep -v "Bash"  # Exclude Bash tools
```

### Counting and Statistics

```bash
# Count occurrences
meta-cc query tools | \
  jq -r '.Error' | \
  grep -c "permission denied"

# Count unique error patterns
meta-cc query tools --where "status='error'" | \
  jq -r '.Error' | \
  grep -oP '(permission|timeout|not found|failed)' | \
  sort | uniq -c
```

### Context Lines

```bash
# Show 2 lines before and after match
meta-cc query tools --output json | \
  jq -r '.[] | .Error' | \
  grep -B 2 -A 2 "permission denied"
```

---

## awk Integration

awk is powerful for text processing, calculations, and formatting.

### Field Processing

```bash
# Extract and format fields
meta-cc query tools | \
  jq -r '[.ToolName, .Status, .Duration] | @tsv' | \
  awk '{print "Tool:", $1, "Status:", $2, "Duration:", $3 "ms"}'

# Custom column formatting
meta-cc stats aggregate --group-by tool | \
  jq -r '.[] | [.group_value, .metrics.count, .metrics.error_rate] | @tsv' | \
  awk '{printf "%-15s Count: %5d Error Rate: %.2f%%\n", $1, $2, $3 * 100}'
```

### Calculations

```bash
# Sum durations
meta-cc query tools | \
  jq -r '.Duration' | \
  awk '{sum += $1} END {print "Total duration:", sum "ms"}'

# Average, min, max
meta-cc query tools | \
  jq -r '.Duration' | \
  awk '{sum+=$1; if($1>max) max=$1; if(min==""|$1<min) min=$1; count++}
       END {print "Avg:", sum/count, "Min:", min, "Max:", max}'
```

### Conditional Processing

```bash
# Flag slow operations
meta-cc query tools | \
  jq -r '[.ToolName, .Duration] | @tsv' | \
  awk '{if ($2 > 5000) print "SLOW:", $1, $2 "ms"; else print "OK:", $1, $2 "ms"}'

# Count by category
meta-cc query tools | \
  jq -r '.Duration' | \
  awk '{
    if ($1 < 1000) fast++;
    else if ($1 < 5000) medium++;
    else slow++;
  }
  END {print "Fast:", fast, "Medium:", medium, "Slow:", slow}'
```

### Grouping and Aggregation

```bash
# Group and sum by tool
meta-cc query tools | \
  jq -r '[.ToolName, .Duration] | @tsv' | \
  awk '{duration[$1] += $2; count[$1]++}
       END {for (tool in duration)
              print tool, "Total:", duration[tool] "ms",
                   "Avg:", duration[tool]/count[tool] "ms"}'
```

---

## sed Integration

sed is useful for stream editing and text transformation.

### Text Replacement

```bash
# Normalize tool names
meta-cc query tools | \
  jq -r '.ToolName' | \
  sed 's/Bash/Shell/g; s/Edit/Modify/g'

# Clean up error messages
meta-cc query tools --where "status='error'" | \
  jq -r '.Error' | \
  sed 's|/home/[^/]*/|~/|g'  # Replace home paths with ~
```

### Filtering Lines

```bash
# Delete empty lines
meta-cc query tools | \
  jq -r '.Error // empty' | \
  sed '/^$/d'

# Keep only lines matching pattern
meta-cc query tools | \
  jq -r '.ToolName + ": " + .Status' | \
  sed -n '/error/p'  # Print only lines with "error"
```

---

## Combining Multiple Tools

Real-world scenarios often combine multiple tools for complex analysis.

### Example 1: Error Pattern Analysis

```bash
# Find top error patterns with counts
meta-cc query tools --where "status='error'" | \
  jq -r '.Error' | \
  grep -oP '(permission|timeout|not found|failed to \w+)' | \
  sed 's/failed to \w*/failed to .../g' | \
  sort | \
  uniq -c | \
  sort -rn | \
  head -10 | \
  awk '{printf "%3d: %s\n", $1, substr($0, index($0, $2))}'
```

### Example 2: Performance Report

```bash
# Generate performance summary
{
  echo "=== Performance Report ==="
  echo ""
  echo "Tool Usage Statistics:"
  meta-cc stats aggregate --group-by tool --metrics "count,avg_duration" | \
    jq -r '.[] | [.group_value, .metrics.count, .metrics.avg_duration] | @tsv' | \
    awk '{printf "  %-15s Count: %5d Avg: %6.0fms\n", $1, $2, $3}' | \
    sort -k3 -rn

  echo ""
  echo "Slowest Operations:"
  meta-cc query tools --sort-by duration --reverse --limit 5 | \
    jq -r '[.ToolName, .Duration, .Status] | @tsv' | \
    awk '{printf "  %s: %dms (%s)\n", $1, $2, $3}'
} > performance_report.txt
```

### Example 3: File Modification Heatmap

```bash
# Identify most problematic files
meta-cc stats files --sort-by error_count | \
  jq -r '.[] | select(.error_count > 0) |
             [.file_path, .edit_count, .error_count, .error_rate] | @tsv' | \
  awk '{printf "%s\t%d edits\t%d errors\t%.1f%%\n", $1, $2, $3, $4*100}' | \
  column -t -s $'\t'
```

---

## Performance Tips

### 1. JSONL is Default (Streaming-Friendly)

JSONL avoids loading everything into memory:
```bash
# Good: Process incrementally (JSONL default)
meta-cc query tools | jq 'select(.Status == "error")' | head -100

# Also good: Use TSV for large datasets
meta-cc query tools --output tsv | awk -F'\t' '$3 == "error"' | head -100
```

### 2. Filter Early

Apply filters in meta-cc before piping:
```bash
# Good: Filter in query
meta-cc query tools --where "status='error'" | jq -r '.Error'

# Less efficient: Filter after
meta-cc query tools | jq 'select(.Status == "error") | .Error'
```

### 3. Use Compact Output

Use `-c` flag in jq for compact JSON:
```bash
# Compact output (faster, less space)
meta-cc query tools | jq -c 'select(.Status == "error")'
```

### 4. Redirect stderr

Suppress logs when scripting:
```bash
meta-cc query tools 2>/dev/null | jq '.ToolName'
```

---

## Troubleshooting

### jq: parse error

**Problem**: jq fails to parse output

**Solution**: JSONL is now the default format
```bash
# Correct: JSONL is default (one JSON object per line)
meta-cc query tools | jq -r '.ToolName'

# For arrays: Use jq -s to slurp JSONL into array
meta-cc query tools | jq -s 'length'
```

### Logs interfering with pipeline

**Problem**: stderr logs appear in pipeline output

**Solution**: Redirect stderr to /dev/null
```bash
meta-cc query tools 2>/dev/null | jq '.ToolName'
```

### Exit code confusion

**Problem**: Script doesn't handle "no results" correctly

**Solution**: Check for exit code 2
```bash
meta-cc query tools --where "tool='NonExistent'" 2>/dev/null
EXIT_CODE=$?

case $EXIT_CODE in
  0) echo "Found results" ;;
  2) echo "No results" ;;
  *) echo "Error" ;;
esac
```

---

## See Also

- [Cookbook](./cookbook.md) - Common analysis patterns
- [meta-cc README](../README.md) - Full command reference
- [jq Manual](https://stedolan.github.io/jq/manual/) - jq documentation

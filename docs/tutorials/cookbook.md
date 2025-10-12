# meta-cc Cookbook: Common Analysis Patterns

This cookbook provides practical examples for common analysis tasks using meta-cc.

## Table of Contents

1. [Error Analysis](#error-analysis)
2. [Performance Profiling](#performance-profiling)
3. [Tool Usage Statistics](#tool-usage-statistics)
4. [File Modification Tracking](#file-modification-tracking)
5. [Time-Based Analysis](#time-based-analysis)
6. [Advanced Filtering](#advanced-filtering)
7. [Data Export and Reporting](#data-export-and-reporting)
8. [Debugging Workflows](#debugging-workflows)
9. [CI/CD Integration](#cicd-integration)
10. [Custom Metrics](#custom-metrics)

---

## Error Analysis

### 1.1 Find all errors in the current session

```bash
# Simple error query
meta-cc query tools --where "status='error'" --output json

# Stream for pipeline processing
meta-cc query tools --stream | jq 'select(.Status == "error")'
```

### 1.2 Group errors by tool

```bash
# Using stats aggregate
meta-cc stats aggregate --group-by tool --metrics "count,error_rate" | \
  jq '.[] | select(.metrics.error_rate > 0)'

# Using jq for custom grouping
meta-cc query tools --stream | \
  jq -s 'group_by(.ToolName) |
         map({tool: .[0].ToolName,
              error_count: map(select(.Status == "error")) | length})'
```

### 1.3 Extract error messages and patterns

```bash
# Extract unique error patterns
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  grep -oP '(permission denied|not found|timeout|failed to)' | \
  sort | uniq -c | sort -rn

# Find permission errors
meta-cc query tools --stream | \
  jq -c 'select(.Status == "error" and (.Error | contains("permission")))' | \
  jq -r '[.ToolName, .Error] | @tsv'
```

### 1.4 Time-based error analysis

```bash
# Errors in the last hour
meta-cc query tools --where "status='error'" --since "1 hour ago" --stream

# Error rate trend over time
meta-cc stats time-series --metric error-rate --interval hour | \
  jq -r '.[] | [.timestamp, .value] | @tsv' | \
  gnuplot -e "set terminal dumb; plot '-' using 2 with lines"
```

---

## Performance Profiling

### 2.1 Find slow operations

```bash
# Tools that took longer than 5 seconds
meta-cc query tools --where "duration > 5000" --stream | \
  jq -r '[.ToolName, .Duration] | @tsv' | \
  sort -k2 -rn

# Top 10 slowest operations
meta-cc query tools --sort-by duration --reverse --limit 10 --output json
```

### 2.2 Average duration by tool

```bash
meta-cc stats aggregate --group-by tool --metrics "count,avg_duration,p90,p95" | \
  jq -r '.[] | [.group_value, .metrics.avg_duration, .metrics.p95] | @tsv' | \
  column -t
```

### 2.3 Performance timeline

```bash
# Tool usage frequency over time
meta-cc stats time-series --metric tool-calls --interval hour | \
  jq -r '.[] | [.timestamp, .value] | @csv'

# Duration trends
meta-cc stats time-series --metric avg-duration --interval day
```

---

## Tool Usage Statistics

### 3.1 Tool usage distribution

```bash
# Count by tool
meta-cc stats aggregate --group-by tool --metrics count | \
  jq -r '.[] | [.group_value, .metrics.count] | @tsv' | \
  awk '{print $2 ": " $1}' | \
  sort -rn

# Pie chart data
meta-cc stats aggregate --group-by tool --metrics count | \
  jq -r '.[] | "\(.group_value): \(.metrics.count)"'
```

### 3.2 Most frequently used tools

```bash
# Top 5 tools
meta-cc stats aggregate --group-by tool --metrics count | \
  jq 'sort_by(.metrics.count) | reverse | .[0:5]'
```

### 3.3 Tool success rates

```bash
meta-cc stats aggregate --group-by tool --metrics "count,error_rate" | \
  jq -r '.[] | [.group_value,
                .metrics.count,
                (.metrics.error_rate * 100 | tostring + "%")] | @tsv' | \
  column -t -s $'\t' -N "Tool,Count,Error Rate"
```

---

## File Modification Tracking

### 4.1 Most edited files

```bash
meta-cc stats files --sort-by edit_count --top 20 | \
  jq -r '.[] | [.file_path, .edit_count] | @tsv' | \
  column -t
```

### 4.2 Files with high error rates

```bash
meta-cc stats files --sort-by error_rate | \
  jq '.[] | select(.error_rate > 0.1)' | \
  jq -r '[.file_path, .error_count, (.error_rate * 100 | tostring + "%")] | @tsv'
```

### 4.3 File operation summary

```bash
meta-cc stats files --top 10 | \
  jq -r '.[] | [.file_path,
                .read_count,
                .edit_count,
                .write_count,
                .error_count] | @tsv' | \
  column -t -s $'\t' -N "File,Reads,Edits,Writes,Errors"
```

---

## Time-Based Analysis

### 5.1 Activity by hour of day

```bash
meta-cc stats time-series --metric tool-calls --interval hour | \
  jq -r '.[] | [(.timestamp | strftime("%H:00")), .value] | @tsv' | \
  awk '{hours[$1] += $2} END {for (h in hours) print h, hours[h]}' | \
  sort
```

### 5.2 Session duration

```bash
# Get first and last timestamps
FIRST=$(meta-cc query tools --limit 1 --sort-by timestamp | jq -r '.[0].Timestamp')
LAST=$(meta-cc query tools --limit 1 --sort-by timestamp --reverse | jq -r '.[0].Timestamp')

# Calculate duration (requires date command)
START_SEC=$(date -d "$FIRST" +%s)
END_SEC=$(date -d "$LAST" +%s)
DURATION=$((END_SEC - START_SEC))

echo "Session duration: $((DURATION / 3600)) hours $((DURATION % 3600 / 60)) minutes"
```

### 5.3 Identify productive hours

```bash
meta-cc stats time-series --metric tool-calls --interval hour | \
  jq -r '.[] | [(.timestamp | strftime("%Y-%m-%d %H:00")), .value] | @tsv' | \
  awk '{if ($2 > 20) print $1 " - High activity: " $2 " tool calls"}'
```

---

## Advanced Filtering

### 6.1 Complex boolean queries

```bash
# Bash errors OR long-running operations
meta-cc query tools --where "(tool='Bash' AND status='error') OR duration>5000" --stream

# Successful file operations
meta-cc query tools --where "tool IN ('Read','Edit','Write') AND status='success'"
```

### 6.2 Pattern matching

```bash
# Tools matching pattern
meta-cc query tools --where "tool LIKE 'meta%'" --stream

# Error messages matching regex
meta-cc query tools --where "status='error' AND error REGEXP 'permission.*denied'"
```

### 6.3 Range queries

```bash
# Operations in a time range
meta-cc query tools --where "timestamp BETWEEN '2025-10-01' AND '2025-10-03'"

# Moderate duration operations (not too fast, not too slow)
meta-cc query tools --where "duration BETWEEN 1000 AND 5000"
```

---

## Data Export and Reporting

### 7.1 Export to CSV

```bash
# Tool statistics to CSV
meta-cc stats aggregate --group-by tool --metrics "count,error_rate,avg_duration" | \
  jq -r '["Tool","Count","Error Rate","Avg Duration"],
         (.[] | [.group_value,
                 .metrics.count,
                 .metrics.error_rate,
                 .metrics.avg_duration]) | @csv'
```

### 7.2 Generate HTML report

```bash
# Create simple HTML table
cat <<EOF > report.html
<html><body>
<h1>meta-cc Session Report</h1>
<table border="1">
<tr><th>Tool</th><th>Count</th><th>Error Rate</th></tr>
EOF

meta-cc stats aggregate --group-by tool --metrics "count,error_rate" | \
  jq -r '.[] | "<tr><td>\(.group_value)</td><td>\(.metrics.count)</td><td>\(.metrics.error_rate)</td></tr>"' >> report.html

echo "</table></body></html>" >> report.html
```

### 7.3 JSON summary for dashboards

```bash
# Create dashboard data
cat <<EOF > dashboard.json
{
  "session_stats": $(meta-cc analyze stats),
  "tool_distribution": $(meta-cc stats aggregate --group-by tool --metrics count),
  "error_summary": $(meta-cc analyze errors),
  "file_hotspots": $(meta-cc stats files --top 10)
}
EOF
```

---

## Debugging Workflows

### 8.1 Trace tool call sequences

```bash
# Get tool call order
meta-cc query tools --sort-by timestamp | \
  jq -r '.[] | [.Timestamp, .ToolName, .Status] | @tsv'

# Identify error sequences (errors followed by retries)
meta-cc query sequences --pattern "error,success" --window 3
```

### 8.2 Find repeated errors

```bash
# Group errors by error message
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  sort | uniq -c | sort -rn | head -10
```

### 8.3 Context around errors

```bash
# Get 2 tools before and after each error
meta-cc query context --error-signature "permission denied" --window 2
```

---

## CI/CD Integration

### 9.1 Fail build on high error rate

```bash
#!/bin/bash
# ci-check-errors.sh

ERROR_RATE=$(meta-cc stats aggregate --group-by status --metrics count | \
  jq -r '.[] | select(.group_value == "error") | .metrics.count' || echo 0)

TOTAL=$(meta-cc analyze stats | jq -r '.total_tools')

if [ "$TOTAL" -gt 0 ]; then
  RATE=$(awk "BEGIN {print $ERROR_RATE / $TOTAL}")
  if (( $(awk "BEGIN {print ($RATE > 0.1)}") )); then
    echo "Error rate too high: $(awk "BEGIN {print $RATE * 100}")%"
    exit 1
  fi
fi

echo "Error rate acceptable: $(awk "BEGIN {print $RATE * 100}")%"
```

### 9.2 Generate performance report

```bash
#!/bin/bash
# ci-performance-report.sh

echo "=== Performance Report ==="
echo ""

echo "Slowest Operations:"
meta-cc query tools --sort-by duration --reverse --limit 5 | \
  jq -r '.[] | "  - \(.ToolName): \(.Duration)ms"'

echo ""
echo "Average Duration by Tool:"
meta-cc stats aggregate --group-by tool --metrics avg_duration | \
  jq -r '.[] | "  - \(.group_value): \(.metrics.avg_duration | round)ms"'
```

---

## Custom Metrics

### 10.1 Calculate custom ratios

```bash
# Edit/Read ratio (how often we edit vs read)
EDITS=$(meta-cc query tools --where "tool='Edit'" --stream | wc -l)
READS=$(meta-cc query tools --where "tool='Read'" --stream | wc -l)
RATIO=$(awk "BEGIN {print $EDITS / ($READS + 1)}")
echo "Edit/Read ratio: $RATIO"
```

### 10.2 Identify anti-patterns

```bash
# Find repeated read-edit-read patterns (inefficient)
meta-cc query sequences --pattern "Read,Edit,Read" | \
  jq '. | length' | \
  awk '{if ($1 > 10) print "Warning: " $1 " inefficient read-edit-read patterns detected"}'
```

### 10.3 Tool diversity score

```bash
# Calculate how many different tools are used
UNIQUE_TOOLS=$(meta-cc stats aggregate --group-by tool --metrics count | jq '. | length')
TOTAL_CALLS=$(meta-cc analyze stats | jq -r '.total_tools')
DIVERSITY=$(awk "BEGIN {print $UNIQUE_TOOLS / sqrt($TOTAL_CALLS)}")
echo "Tool diversity score: $DIVERSITY"
```

---

## Tips and Best Practices

### Use `--stream` for large datasets

Streaming output is more efficient for large result sets:
```bash
# Good: Stream and filter incrementally
meta-cc query tools --stream | jq 'select(.Status == "error")' | head -100

# Avoid: Load all data into memory first
meta-cc query tools --limit 100000 | jq 'select(.Status == "error")'
```

### Combine with standard Unix tools

meta-cc integrates well with:
- **jq**: JSON processing and filtering
- **grep**: Text pattern matching
- **awk**: Text processing and calculations
- **sort, uniq**: Data aggregation
- **column**: Table formatting
- **gnuplot**: Data visualization

### Redirect logs when scripting

Always redirect stderr to avoid log interference:
```bash
# In scripts
DATA=$(meta-cc query tools 2>/dev/null)

# In pipelines
meta-cc query tools --stream 2>/dev/null | jq '.ToolName'
```

### Use exit codes in conditionals

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
  fi
fi
```

---

## See Also

- [CLI Composability Guide](./cli-composability.md) - Integration with jq, grep, awk
- [meta-cc README](../README.md) - Full command reference
- [Examples and Usage](./examples-usage.md) - Getting started guide

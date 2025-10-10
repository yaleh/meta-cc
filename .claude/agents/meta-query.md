---
name: meta-query
description: Execute complex Unix pipeline queries on meta-cc data for multi-step aggregations
---

λ(meta_cc_data, unix_pipeline) → aggregated_results | ∀query ∈ complex_patterns:

pipeline :: Query → Structured_Output
pipeline(Q) = extract(data) → transform(unix_tools) → aggregate(compact) → format(readable)

extract :: Query_Intent → Raw_Data
extract(I) = {
  errors: meta_cc_query("errors"),
  tools: meta_cc_query("tools"),
  files: meta_cc_query("file-access"),
  sequences: meta_cc_query("tool-sequences")
}

transform :: Raw_Data × Unix_Tools → Processed_Data
transform(D, T) = apply(jq) → apply(sort) → apply(uniq) → apply(awk) → apply(column)

aggregate :: Processed_Data → Compact_Results
aggregate(P) = {
  count: uniq_c(P),
  top_n: head_n(sort_rn(P)),
  rate: calculate(numerator / denominator × 100)
}

format :: Compact_Results → Readable_Output
format(C) = add_headers(C) ∧ add_context(C) ∧ tabular(C)

when_to_use :: Query → Decision
when_to_use(Q) = {
  use_meta_query: multi_step(Q) ∨ complex_aggregation(Q) ∨ top_n(Q),
  use_mcp: single_jq(Q) ∨ simple_filter(Q)
}

constraints:
- compact_output: avoid(raw_jsonl) ∧ prefer(aggregated_counts)
- readable_format: use(column) ∨ use(headers) ∨ use(context)
- efficient_pipelines: minimize(steps) ∧ optimize(performance)
- error_handling: validate(meta_cc_exists) ∧ handle(pipeline_failures)

output :: Unix_Pipeline → Formatted_Table
output(P) = execute(pipeline) → verify(success) → format(results) → explain(meaning)

---

## When to Use @meta-query

Use @meta-query when:
- **Multi-step Unix pipelines** are needed (jq + sort + uniq)
- **Complex aggregations** beyond single jq expressions
- **Top-N queries** with sorting and limiting
- **Custom data transformations** requiring awk/sed

**Do NOT use** for:
- Simple jq filters (use MCP tools instead)
- Single-step queries (use MCP tools)
- Raw data extraction without aggregation (use MCP tools)

---

## Core Capabilities

### 1. Error Statistics by Tool

**Query**: "统计本项目所有错误，按工具分组"

**Pipeline**:
```bash
meta-cc query errors | jq -r '.ToolName' | sort | uniq -c | sort -rn
```

**Output**:
```
Error Statistics by Tool:

311 Bash
 62 Read
 15 Edit
  7 Write
```

---

### 2. Top-N Tool Usage

**Query**: "列出使用最多的 10 个工具"

**Pipeline**:
```bash
meta-cc query tools | jq -r '.ToolName' | sort | uniq -c | sort -rn | head -10
```

**Output**:
```
Top 10 Most Used Tools:

495 Bash
162 Read
140 TodoWrite
 89 Edit
 45 Write
 38 Grep
 22 Glob
 12 SlashCommand
  8 WebSearch
  5 WebFetch
```

---

### 3. Error Rate by Tool

**Query**: "计算每个工具的错误率"

**Pipeline**:
```bash
# Extract total calls by tool
total=$(meta-cc query tools | jq -r '.ToolName' | sort | uniq -c | sort -k2)

# Extract errors by tool
errors=$(meta-cc query errors | jq -r '.ToolName' | sort | uniq -c | sort -k2)

# Join and calculate error rate
join -1 2 -2 2 <(echo "$total") <(echo "$errors") | \
  awk '{printf "%-15s Total: %4d  Errors: %4d  Rate: %5.2f%%\n", $1, $2, $3, ($3/$2)*100}'
```

**Output**:
```
Error Rate by Tool:

Bash            Total:  495  Errors:  311  Rate: 62.83%
Read            Total:  162  Errors:   62  Rate: 38.27%
Edit            Total:   89  Errors:   15  Rate: 16.85%
Write           Total:   45  Errors:    7  Rate: 15.56%
```

---

### 4. File Operation Hotspots

**Query**: "哪些文件修改最频繁？"

**Pipeline**:
```bash
meta-cc query file-access | jq -r '.file' | sort | uniq -c | sort -rn | head -10
```

**Output**:
```
Top 10 Most Modified Files:

 23 cmd/mcp.go
 18 cmd/query_tools.go
 12 README.md
 10 cmd/pipeline.go
  9 internal/locator/args.go
  7 .claude/commands/meta-stats.md
  6 docs/plan.md
  5 cmd/mcp_test.go
  4 internal/parser/session.go
  3 go.mod
```

---

### 5. File Operation Timeline

**Query**: "显示 cmd/mcp.go 的操作历史"

**Pipeline**:
```bash
meta-cc query file-access --file "cmd/mcp.go" | \
  jq -r '[.Timestamp[0:19], .Operation, .ToolName] | @tsv' | \
  column -t -s $'\t' -N "Timestamp,Operation,Tool"
```

**Output**:
```
Timestamp            Operation  Tool
2025-10-05T14:23:15  Read       Read
2025-10-05T14:25:42  Edit       Edit
2025-10-05T14:30:18  Read       Read
2025-10-05T15:12:05  Edit       Edit
```

---

### 6. Tool Sequence Patterns

**Query**: "显示最常见的工具调用序列"

**Pipeline**:
```bash
meta-cc query tool-sequences --min-occurrences 3 | \
  jq -r '.Sequence' | \
  sort | uniq -c | sort -rn | head -10
```

**Output**:
```
Most Common Tool Sequences:

 12 Read -> Edit -> Bash
  8 Grep -> Read -> Edit
  6 Glob -> Read -> Edit
  5 Read -> Edit -> Read
  4 Bash -> Read -> Bash
```

---

## Output Guidelines

### Always Return Compact Results

1. **Use counting aggregations**:
   - `uniq -c` for frequency counts
   - `head -N` to limit top results
   - `sort -rn` for descending numeric sort

2. **Avoid raw JSONL dumps**:
   - ❌ BAD: Dump 1000 lines of raw JSONL
   - ✅ GOOD: Show top 10 aggregated counts

3. **Include context and headers**:
   - Add descriptive titles ("Top 10 Most Used Tools:")
   - Use `column -t` for aligned tables
   - Add header rows with `-N` flag

---

## Integration with @meta-coach

@meta-coach can call @meta-query for data gathering:

**Workflow**:
```
@meta-coach: "分析本项目的工作模式"
  ↓
@meta-coach → @meta-query("Get tool usage statistics")
  ↓
@meta-query → Executes pipeline → Returns compact stats
  ↓
@meta-coach → Analyzes data → Generates insights and recommendations
```

**Example**:
```
User: "@meta-coach 帮我分析项目中的错误模式"

@meta-coach thinks: "I need error statistics first"
  → Calls: @meta-query("统计所有错误，按工具分组")

@meta-query returns:
  311 Bash
   62 Read
   15 Edit

@meta-coach analyzes:
  "You have a high Bash error rate (62.83%). Common causes:
   - Command not found errors
   - Path resolution issues
   - Recommendation: Use absolute paths, verify commands exist"
```

---

## Error Handling

### If meta-cc fails:

1. **Check session/project exists**:
   ```bash
   meta-cc query tools --project .
   ```

2. **Suggest project-level queries**:
   ```
   Try: meta-cc query errors --project .
   (instead of session-only queries)
   ```

3. **Report errors clearly**:
   ```
   Error: meta-cc command not found
   Solution: Ensure meta-cc is built and in PATH
   ```

### If pipeline fails:

1. **Break down step by step**:
   ```bash
   # Test each step independently
   meta-cc query errors  # Step 1
   meta-cc query errors | jq -r '.ToolName'  # Step 2
   meta-cc query errors | jq -r '.ToolName' | sort  # Step 3
   ```

2. **Identify which step failed**:
   ```
   Error at jq step: parse error (invalid JSON)
   Solution: Check meta-cc output format
   ```

3. **Suggest simpler alternatives**:
   ```
   Complex pipeline failed.
   Try using MCP tool instead: mcp__meta_cc__aggregate_stats
   ```

---

## Usage Decision Tree

```
┌─ Single jq expression? ──→ YES ──→ Use MCP
│
├─ Top-N query? ──→ YES ──→ Use @meta-query
│
├─ Multi-step aggregation? ──→ YES ──→ Use @meta-query
│
├─ Custom join/awk logic? ──→ YES ──→ Use @meta-query
│
└─ Simple filter? ──→ YES ──→ Use MCP
```

**Examples**:
- ✅ @meta-query: "Top 10 tools by usage"
- ✅ @meta-query: "Error rate per tool"
- ❌ MCP: "All errors from Bash tool"
- ❌ MCP: "Tool calls with status=error"

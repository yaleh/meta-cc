# @meta-query Test Scenarios

## Scenario 1: Error Statistics by Tool

**User**: "@meta-query 统计本项目所有错误，按工具分组"

**Expected Pipeline**:
```bash
meta-cc query errors | jq -r '.ToolName' | sort | uniq -c | sort -rn
```

**Expected Output**:
```
Error Statistics by Tool:

311 Bash
 62 Read
 15 Edit
  7 Write
```

**Verification**:
- ✅ Output is compact (counts + tool names only)
- ✅ Sorted by count (descending)
- ✅ No raw JSONL dump
- ✅ Includes descriptive header

---

## Scenario 2: Top-N Tool Usage

**User**: "@meta-query 列出使用最多的 5 个工具"

**Expected Pipeline**:
```bash
meta-cc query tools | jq -r '.ToolName' | sort | uniq -c | sort -rn | head -5
```

**Expected Output**:
```
Top 5 Most Used Tools:

495 Bash
162 Read
140 TodoWrite
 89 Edit
 45 Write
```

**Verification**:
- ✅ Exactly 5 results (top-N limiting works)
- ✅ Sorted by count descending
- ✅ Compact format

---

## Scenario 3: File Hotspots

**User**: "@meta-query 哪些文件修改最频繁？显示前 10 个"

**Expected Pipeline**:
```bash
meta-cc query file-access | jq -r '.file' | sort | uniq -c | sort -rn | head -10
```

**Expected Output**:
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

**Verification**:
- ✅ Shows file paths with modification counts
- ✅ Limited to top 10
- ✅ Sorted by frequency

---

## Scenario 4: File Operation Timeline

**User**: "@meta-query 显示 cmd/mcp.go 的完整操作历史"

**Expected Pipeline**:
```bash
meta-cc query file-access --file "cmd/mcp.go" | \
  jq -r '[.Timestamp[0:19], .Operation, .ToolName] | @tsv' | \
  column -t -s $'\t' -N "Timestamp,Operation,Tool"
```

**Expected Output**:
```
Timestamp            Operation  Tool
2025-10-05T14:23:15  Read       Read
2025-10-05T14:25:42  Edit       Edit
2025-10-05T14:30:18  Read       Read
2025-10-05T15:12:05  Edit       Edit
```

**Verification**:
- ✅ Tabular format with aligned columns
- ✅ Header row included
- ✅ Timestamps truncated to second precision
- ✅ Shows operation type and tool name

---

## Scenario 5: Tool Sequence Patterns

**User**: "@meta-query 显示最常见的工具调用序列模式"

**Expected Pipeline**:
```bash
meta-cc query tool-sequences --min-occurrences 3 | \
  jq -r '.Sequence' | \
  sort | uniq -c | sort -rn | head -10
```

**Expected Output**:
```
Most Common Tool Sequences:

 12 Read -> Edit -> Bash
  8 Grep -> Read -> Edit
  6 Glob -> Read -> Edit
  5 Read -> Edit -> Read
  4 Bash -> Read -> Bash
```

**Verification**:
- ✅ Shows sequences with frequency counts
- ✅ Sorted by occurrence count
- ✅ Limited to top 10 patterns

---

## Scenario 6: Error Rate Calculation

**User**: "@meta-query 计算每个工具的错误率"

**Expected Pipeline**:
```bash
# Extract total calls by tool
total=$(meta-cc query tools | jq -r '.ToolName' | sort | uniq -c | sort -k2)

# Extract errors by tool
errors=$(meta-cc query errors | jq -r '.ToolName' | sort | uniq -c | sort -k2)

# Join and calculate error rate
join -1 2 -2 2 <(echo "$total") <(echo "$errors") | \
  awk '{printf "%-15s Total: %4d  Errors: %4d  Rate: %5.2f%%\n", $1, $2, $3, ($3/$2)*100}'
```

**Expected Output**:
```
Error Rate by Tool:

Bash            Total:  495  Errors:  311  Rate: 62.83%
Read            Total:  162  Errors:   62  Rate: 38.27%
Edit            Total:   89  Errors:   15  Rate: 16.85%
Write           Total:   45  Errors:    7  Rate: 15.56%
```

**Verification**:
- ✅ Shows total, error count, and percentage
- ✅ Formatted table with aligned columns
- ✅ Percentage calculated correctly

---

## Scenario 7: Integration with @meta-coach

**User**: "@meta-coach 分析本项目的工作模式和错误趋势"

**Expected Workflow**:
1. **@meta-coach** analyzes user intent: "需要工具使用统计和错误模式"
2. **@meta-coach** → **@meta-query**: "统计所有工具使用次数"
3. **@meta-query** returns: Compact tool usage stats
4. **@meta-coach** → **@meta-query**: "统计错误，按工具分组"
5. **@meta-query** returns: Compact error stats
6. **@meta-coach** analyzes both datasets → Generates insights

**Expected @meta-coach Output**:
```
工作模式分析：

1. 工具使用分布：
   - Bash (495次) - 最常用，但错误率高 (62.83%)
   - Read (162次) - 第二常用，错误率中等 (38.27%)
   - TodoWrite (140次) - 任务跟踪频繁

2. 错误模式：
   - Bash 错误占总错误的 78%
   - 常见原因：命令未找到、路径解析问题

3. 建议：
   - 使用绝对路径代替相对路径
   - 执行命令前先验证存在性
   - 增加错误处理逻辑
```

**Verification**:
- ✅ @meta-coach receives compact data (not raw JSONL)
- ✅ @meta-coach generates actionable insights
- ✅ Integration flow is seamless

---

## Scenario 8: Error Handling - meta-cc Not Found

**User**: "@meta-query 统计错误"

**If meta-cc not in PATH**:

**Expected Output**:
```
Error: meta-cc command not found

Troubleshooting:
1. Verify meta-cc is built:
   cd /home/yale/work/meta-cc
   make build

2. Add to PATH or use absolute path:
   export PATH=$PATH:/home/yale/work/meta-cc

   OR

   /home/yale/work/meta-cc/meta-cc query errors

3. Verify binary exists:
   ls -l /home/yale/work/meta-cc/meta-cc
```

**Verification**:
- ✅ Clear error message
- ✅ Actionable troubleshooting steps
- ✅ Multiple solution paths provided

---

## Scenario 9: Error Handling - Pipeline Failure

**User**: "@meta-query 计算错误率"

**If jq fails** (invalid JSON):

**Expected Response**:
```
Error: jq parsing failed

Debugging steps:
1. Test meta-cc output:
   meta-cc query errors | head -1

2. Verify JSON format:
   meta-cc query errors | jq '.' | head -1

3. If JSON is invalid:
   - Check meta-cc version
   - Try rebuilding: make build

Alternative: Use MCP tool instead
   mcp__meta-insight__aggregate_stats
```

**Verification**:
- ✅ Identifies which step failed
- ✅ Provides debugging steps
- ✅ Suggests alternative approaches

---

## Manual Testing Checklist

To test @meta-query in Claude Code:

1. **Invoke @meta-query directly**:
   ```
   @meta-query 统计本项目所有错误，按工具分组
   ```

2. **Verify compact output**:
   - ✅ Not dumping raw JSONL
   - ✅ Shows aggregated counts
   - ✅ Includes headers

3. **Test Top-N limiting**:
   ```
   @meta-query 列出使用最多的 5 个工具
   ```
   - ✅ Returns exactly 5 results

4. **Test integration with @meta-coach**:
   ```
   @meta-coach 分析我的工作模式
   ```
   - ✅ @meta-coach calls @meta-query automatically
   - ✅ Receives compact data
   - ✅ Generates insights

5. **Test error handling**:
   - Rename meta-cc binary temporarily
   - Invoke @meta-query
   - ✅ Receives clear error message with solutions

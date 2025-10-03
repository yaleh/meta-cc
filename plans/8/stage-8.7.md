# Stage 8.7: Create New Query-Focused Slash Commands

## Overview

**Objective**: Create new Slash Commands that provide quick access to Phase 8 query capabilities.

**Code Estimate**: ~120 lines (2 new command files)

**Priority**: Medium (improves user experience)

**Time Estimate**: 30-45 minutes

## Problem Statement

Users need to manually construct `meta-cc query` commands, which:
- Requires remembering command syntax
- Is prone to typing errors
- Slows down workflow
- Doesn't leverage Slash Command convenience

## New Commands to Create

### 1. `/meta-query-tools` - Quick Tool Query

**Purpose**: Fast tool call querying without remembering syntax

**Usage Examples**:
- `/meta-query-tools` - Last 20 tool calls
- `/meta-query-tools Bash` - All Bash calls (last 20)
- `/meta-query-tools Bash error` - Bash errors
- `/meta-query-tools "" error 10` - Last 10 errors (any tool)

---

### 2. `/meta-query-messages` - User Message Search

**Purpose**: Quick search through user messages

**Usage Examples**:
- `/meta-query-messages "Phase 8"` - Find "Phase 8" mentions
- `/meta-query-messages "error|bug"` - Find error/bug mentions
- `/meta-query-messages "fix.*bug" 20` - Regex search, 20 results

---

## Implementation Details

### Command 1: `/meta-query-tools`

**File**: `.claude/commands/meta-query-tools.md`

```markdown
---
name: meta-query-tools
description: 快速查询工具调用，支持按工具名、状态过滤（Phase 8 增强）
allowed_tools: [Bash]
argument-hint: [tool-name] [status] [limit]
---

# meta-query-tools: 工具调用快速查询

使用 Phase 8 query 命令快速查询工具调用，无需记住复杂语法。

## 用法

```bash
# 检查 meta-cc 是否安装
if ! command -v meta-cc &> /dev/null; then
    echo "❌ 错误：meta-cc 未安装或不在 PATH 中"
    echo ""
    echo "请安装 meta-cc："
    echo "  1. 下载或构建 meta-cc 二进制文件"
    echo "  2. 将其放置在 PATH 中（如 /usr/local/bin/meta-cc）"
    echo "  3. 确保可执行权限：chmod +x /usr/local/bin/meta-cc"
    exit 1
fi

# 参数解析
TOOL_NAME=${1:-""}
STATUS=${2:-""}
LIMIT=${3:-20}

echo "# 工具调用查询结果"
echo ""

# 构建查询命令
QUERY_CMD="meta-cc query tools --limit $LIMIT --output json"

# 添加工具过滤
if [ -n "$TOOL_NAME" ]; then
    QUERY_CMD="$QUERY_CMD --tool $TOOL_NAME"
    echo "**过滤条件**: 工具=$TOOL_NAME"
fi

# 添加状态过滤
if [ -n "$STATUS" ]; then
    QUERY_CMD="$QUERY_CMD --status $STATUS"
    if [ -n "$TOOL_NAME" ]; then
        echo ", 状态=$STATUS"
    else
        echo "**过滤条件**: 状态=$STATUS"
    fi
fi

# 显示数量限制
if [ -z "$TOOL_NAME" ] && [ -z "$STATUS" ]; then
    echo "**显示**: 最近 $LIMIT 次工具调用"
else
    echo ", 数量限制=$LIMIT"
fi

echo ""
echo "---"
echo ""

# 执行查询
result=$($QUERY_CMD)

# 检查是否有结果
count=$(echo "$result" | jq 'length')

if [ "$count" -eq 0 ]; then
    echo "❌ 未找到匹配的工具调用"
    echo ""
    echo "💡 **提示**："
    echo "- 检查工具名称拼写（如 Bash, Read, Edit, Write, Grep）"
    echo "- 检查状态值（error 或 success）"
    echo "- 尝试增加 limit 参数"
    exit 0
fi

# 显示结果
echo "## 查询结果（共 $count 条）"
echo ""

# 根据是否有错误过滤，选择不同的显示格式
if [ "$STATUS" = "error" ]; then
    # 错误模式：显示错误信息
    echo "$result" | jq -r '.[] |
        "### \(.ToolName) 错误\n" +
        "- **时间**: \(.Timestamp)\n" +
        "- **错误**: \(.Error)\n" +
        "- **输入**: \(.Input | to_entries | map("\(.key)=\(.value)") | join(", "))\n"
    '
else
    # 正常模式：简洁列表
    echo "$result" | jq -r '.[] |
        "\(if .Status == "error" or .Error != "" then "❌" else "✅" end) **\(.ToolName)** - \(.Timestamp)"
    '
fi

echo ""
echo "---"
echo ""

# 统计摘要
echo "## 统计摘要"
echo ""

error_count=$(echo "$result" | jq '[.[] | select(.Status == "error" or .Error != "")] | length')
success_count=$(echo "$result" | jq '[.[] | select(.Status != "error" and .Error == "")] | length')
error_rate=0
if [ "$count" -gt 0 ]; then
    error_rate=$(echo "scale=2; $error_count * 100 / $count" | bc)
fi

echo "- **总数**: $count 次"
echo "- **成功**: $success_count 次"
echo "- **错误**: $error_count 次"
echo "- **错误率**: ${error_rate}%"

# 工具频率分布（仅在未过滤工具时显示）
if [ -z "$TOOL_NAME" ]; then
    echo ""
    echo "### 工具分布"
    echo ""
    echo "$result" | jq -r '
        [.[] | .ToolName] |
        group_by(.) |
        map({tool: .[0], count: length}) |
        sort_by(.count) |
        reverse |
        .[] |
        "- **\(.tool)**: \(.count) 次"
    '
fi

echo ""
echo "---"
echo ""
echo "💡 **提示**："
echo "- 使用 /meta-query-tools Bash 查看所有 Bash 调用"
echo "- 使用 /meta-query-tools \"\" error 查看所有错误"
echo "- 使用 /meta-query-tools Read \"\" 30 查看最近 30 次 Read 调用"
echo "- 使用 @meta-coach 获取深入分析和建议"
```

## 示例输出

```markdown
# 工具调用查询结果

**过滤条件**: 工具=Bash, 状态=error, 数量限制=10

---

## 查询结果（共 3 条）

### Bash 错误
- **时间**: 2025-10-03T10:23:15Z
- **错误**: exit status 1: npm test failed
- **输入**: command=npm test

### Bash 错误
- **时间**: 2025-10-03T10:25:42Z
- **错误**: exit status 1: npm test failed
- **输入**: command=npm test

### Bash 错误
- **时间**: 2025-10-03T10:28:19Z
- **错误**: exit status 1: npm test failed
- **输入**: command=npm test

---

## 统计摘要

- **总数**: 3 次
- **成功**: 0 次
- **错误**: 3 次
- **错误率**: 100%

---

💡 **提示**：
- 使用 /meta-query-tools Bash 查看所有 Bash 调用
- 使用 /meta-query-tools "" error 查看所有错误
- 使用 @meta-coach 获取深入分析和建议
```
```

---

### Command 2: `/meta-query-messages`

**File**: `.claude/commands/meta-query-messages.md`

```markdown
---
name: meta-query-messages
description: 搜索用户消息，支持正则表达式匹配（Phase 8 增强）
allowed_tools: [Bash]
argument-hint: [pattern] [limit]
---

# meta-query-messages: 用户消息搜索

使用 Phase 8 query 命令搜索用户消息，支持正则表达式模式匹配。

## 用法

```bash
# 检查 meta-cc 是否安装
if ! command -v meta-cc &> /dev/null; then
    echo "❌ 错误：meta-cc 未安装或不在 PATH 中"
    echo ""
    echo "请安装 meta-cc："
    echo "  1. 下载或构建 meta-cc 二进制文件"
    echo "  2. 将其放置在 PATH 中（如 /usr/local/bin/meta-cc）"
    echo "  3. 确保可执行权限：chmod +x /usr/local/bin/meta-cc"
    exit 1
fi

# 参数解析
PATTERN=${1:-".*"}
LIMIT=${2:-10}

echo "# 用户消息搜索结果"
echo ""

# 显示搜索条件
if [ "$PATTERN" = ".*" ]; then
    echo "**搜索**: 所有用户消息"
else
    echo "**搜索模式**: \`$PATTERN\`"
fi
echo "**数量限制**: 最多 $LIMIT 条"
echo ""
echo "---"
echo ""

# 执行查询
result=$(meta-cc query user-messages --match "$PATTERN" --limit "$LIMIT" --sort-by timestamp --reverse --output json)

# 检查是否有结果
count=$(echo "$result" | jq 'length')

if [ "$count" -eq 0 ]; then
    echo "❌ 未找到匹配的用户消息"
    echo ""
    echo "💡 **提示**："
    echo "- 检查正则表达式语法（如 'error|bug', '^fix', '.*test'）"
    echo "- 尝试更宽泛的模式（如 '.*' 查看所有消息）"
    echo "- 增加 limit 参数以扩大搜索范围"
    exit 0
fi

# 显示结果
echo "## 搜索结果（共 $count 条）"
echo ""

# 遍历每条消息
echo "$result" | jq -r '.[] |
    "### \(.Timestamp)\n" +
    "\(.Content | .[0:300])\(if (.Content | length) > 300 then "..." else "" end)\n" +
    "---\n"
'

echo ""

# 显示总计
total_count=$(meta-cc query user-messages --match "$PATTERN" --limit 1000 --output json | jq 'length')

echo "📊 **统计**："
echo "- 显示: $count 条（最新）"
echo "- 总计: $total_count 条匹配的消息"

if [ "$total_count" -gt "$count" ]; then
    remaining=$((total_count - count))
    echo "- 未显示: $remaining 条（增加 limit 参数查看更多）"
fi

echo ""
echo "---"
echo ""
echo "💡 **提示**："
echo "- 使用正则表达式搜索："
echo "  - /meta-query-messages 'Phase 8' - 查找包含 'Phase 8' 的消息"
echo "  - /meta-query-messages 'error|bug' - 查找包含 'error' 或 'bug' 的消息"
echo "  - /meta-query-messages '^Continue' - 查找以 'Continue' 开头的消息"
echo "  - /meta-query-messages 'fix.*bug' - 查找 'fix' 和 'bug' 之间有内容的消息"
echo "- 增加结果数量："
echo "  - /meta-query-messages 'error' 20 - 显示 20 条结果"
echo "- 使用 @meta-coach 分析消息模式和趋势"
```

## 正则表达式示例

| 模式 | 说明 | 示例 |
|------|------|------|
| `error` | 精确匹配 "error" | "There's an error" ✅ |
| `error\|bug` | 匹配 "error" 或 "bug" | "Fix bug" ✅, "Handle error" ✅ |
| `^Continue` | 以 "Continue" 开头 | "Continue with..." ✅ |
| `test$` | 以 "test" 结尾 | "Run the test" ✅ |
| `fix.*bug` | "fix" 后跟任意字符，再跟 "bug" | "fix this bug" ✅ |
| `Phase [0-9]` | "Phase" 后跟数字 | "Phase 8" ✅, "Phase 1" ✅ |
| `.*` | 所有消息 | 任何消息 ✅ |

## 示例输出

```markdown
# 用户消息搜索结果

**搜索模式**: `Phase 8`
**数量限制**: 最多 10 条

---

## 搜索结果（共 5 条）

### 2025-10-03T10:45:23Z
Let's continue with Phase 8 implementation. We need to add the query commands as planned.
---

### 2025-10-03T09:12:45Z
Can you explain the difference between Phase 8 query and the old parse extract command?
---

### 2025-10-03T08:34:12Z
I think Phase 8 will solve the context overflow issue we've been having.
---

### 2025-10-02T16:23:01Z
Phase 8 query tools command is working perfectly! Much faster than before.
---

### 2025-10-02T14:56:33Z
Let's start Phase 8 planning. What should be the first stage?
---

📊 **统计**：
- 显示: 5 条（最新）
- 总计: 5 条匹配的消息

---

💡 **提示**：
- 使用正则表达式搜索：
  - /meta-query-messages 'Phase 8' - 查找包含 'Phase 8' 的消息
  - /meta-query-messages 'error|bug' - 查找包含 'error' 或 'bug' 的消息
- 使用 @meta-coach 分析消息模式和趋势
```
```

---

## Testing Strategy

### Test 1: `/meta-query-tools` - Basic Usage
```bash
/meta-query-tools
# Expected: Last 20 tool calls with status indicators
```

### Test 2: `/meta-query-tools` - Filter by Tool
```bash
/meta-query-tools Bash
# Expected: Only Bash calls, last 20
```

### Test 3: `/meta-query-tools` - Error Filtering
```bash
/meta-query-tools "" error 10
# Expected: Last 10 errors from any tool
```

### Test 4: `/meta-query-tools` - Tool + Status
```bash
/meta-query-tools Edit error
# Expected: All Edit errors, last 20
```

### Test 5: `/meta-query-messages` - Basic Search
```bash
/meta-query-messages "Phase 8"
# Expected: Messages containing "Phase 8"
```

### Test 6: `/meta-query-messages` - Regex Pattern
```bash
/meta-query-messages "error|bug"
# Expected: Messages with "error" or "bug"
```

### Test 7: `/meta-query-messages` - Custom Limit
```bash
/meta-query-messages "test" 20
# Expected: 20 messages containing "test"
```

### Test 8: Edge Cases
```bash
# No results
/meta-query-tools NonExistentTool
# Expected: Helpful message, no error

# Invalid regex
/meta-query-messages "[invalid"
# Expected: Error message with suggestion
```

---

## Implementation Steps

### Step 1: Create Command Files

```bash
# Create meta-query-tools.md
touch .claude/commands/meta-query-tools.md

# Create meta-query-messages.md
touch .claude/commands/meta-query-messages.md
```

### Step 2: Write Command Content

Copy the full command definitions from this plan to the respective files.

### Step 3: Test Commands

```bash
# Test in Claude Code
/meta-query-tools
/meta-query-messages "test"
```

### Step 4: Update Documentation

Add to README.md or docs/examples-usage.md:

```markdown
### Quick Query Commands (Phase 8)

**Query Tool Calls**:
```bash
/meta-query-tools                # Last 20 tool calls
/meta-query-tools Bash          # All Bash calls
/meta-query-tools "" error      # All errors
/meta-query-tools Edit error 10 # Last 10 Edit errors
```

**Search User Messages**:
```bash
/meta-query-messages "Phase 8"        # Find mentions
/meta-query-messages "error|bug"      # Regex search
/meta-query-messages "fix.*bug" 20    # Complex regex, 20 results
```
```

---

## Acceptance Criteria

- ✅ `/meta-query-tools` command created and works
- ✅ Supports 3 parameters: tool, status, limit
- ✅ Shows clear output with statistics
- ✅ Handles edge cases (no results, invalid input)
- ✅ `/meta-query-messages` command created and works
- ✅ Supports regex pattern matching
- ✅ Shows message excerpts (300 chars max)
- ✅ Provides helpful tips and examples
- ✅ Both commands have error checking
- ✅ Documentation updated

---

## Dependencies

- ✅ Stage 8.2 completed (`query tools` available)
- ✅ Stage 8.3 completed (`query user-messages` available)
- ✅ `meta-cc` binary in PATH
- ✅ `jq` installed for JSON processing
- ✅ `bc` installed for percentage calculation

---

## Benefits

### User Experience
- ✅ No need to remember complex command syntax
- ✅ Quick access to common queries
- ✅ Clear, formatted output
- ✅ Helpful error messages and tips

### Workflow Efficiency
- ✅ Faster debugging (quick error查询)
- ✅ Easy message search (find past discussions)
- ✅ Reduced cognitive load
- ✅ Better integration with Phase 8

### Learning Curve
- ✅ Examples in help text
- ✅ Regex pattern guide
- ✅ Tips for advanced usage
- ✅ Encourages exploration

---

## Future Enhancements (Optional)

### Additional Commands (not in scope)
- `/meta-query-workflow` - Analyze workflow patterns
- `/meta-query-files` - Query file-related operations
- `/meta-query-sequences` - Find tool usage sequences

### Parameters (not in scope)
- `--output` format selection
- `--context` to show surrounding turns
- `--since` time-based filtering

---

## Related Documentation

- Phase 8 Implementation Plan: `/plans/8/phase-8-implementation-plan.md`
- Integration Improvement Proposal: `/tmp/meta-cc-integration-improvement-proposal.md`
- Slash Commands Documentation: `docs/examples-usage.md`

---
name: meta-query-tools
description: 工具调用快速查询（Phase 14：标准化工具 + 退出码）
allowed_tools: [Bash]
argument-hint: [tool-name|filter] [limit]
---

# meta-query-tools: 工具调用快速查询

快速查询工具调用，支持过滤和 SQL-like 表达式。

```bash
# Source shared utilities
source "$(dirname "$0")/../lib/meta-utils.sh"
check_meta_cc_installed

# 参数解析
FILTER_EXPR=${1:-""}
LIMIT=${2:-20}

echo "# 工具调用查询结果" >&2
echo "" >&2

# 构建查询命令
if [ -n "$FILTER_EXPR" ]; then
    # Phase 10: Use advanced filtering if expression looks like a where clause
    if echo "$FILTER_EXPR" | grep -qE "(AND|OR|IN|BETWEEN|LIKE|=|>|<)"; then
        QUERY_CMD="meta-cc query tools --where \"$FILTER_EXPR\" --limit $LIMIT"
        echo "**过滤条件**: $FILTER_EXPR" >&2
    else
        # Legacy: treat as tool name
        QUERY_CMD="meta-cc query tools --tool $FILTER_EXPR --limit $LIMIT"
        echo "**过滤条件**: 工具=$FILTER_EXPR" >&2
    fi
else
    QUERY_CMD="meta-cc query tools --limit $LIMIT"
    echo "**显示**: 最近 $LIMIT 次工具调用" >&2
fi

echo "" >&2
echo "---" >&2
echo "" >&2

# Execute query with exit code handling
result=$($QUERY_CMD 2>/dev/null)
exit_code=$?

if [ $exit_code -eq 2 ]; then
    echo "❌ 未找到匹配的工具调用" >&2
    echo "" >&2
    echo "💡 **提示**：" >&2
    echo "- 检查工具名称拼写（如 Bash, Read, Edit, Write, Grep）" >&2
    echo "- 检查状态值（error 或 success）" >&2
    echo "- 尝试增加 limit 参数" >&2
    exit 0
elif [ $exit_code -eq 1 ]; then
    echo "❌ 查询执行失败" >&2
    exit 1
fi

# Convert JSONL to JSON array
result=$(jsonl_to_json "$result")
count=$(echo "$result" | jq 'length')

# 显示结果
echo "## 查询结果（共 $count 条）" >&2
echo "" >&2

# 简洁列表
echo "$result" | jq -r '.[] |
    "\(if .Status == "error" or .Error != "" or (.Output | contains("error")) then "❌" else "✅" end) **\(.ToolName)** (\(.UUID[0:8]))"
'

echo "" >&2
echo "---" >&2
echo "" >&2

# 统计摘要
echo "## 统计摘要" >&2
echo "" >&2

stats=$(calculate_error_stats "$result")
echo "- **总数**: $(echo "$stats" | jq '.total') 次" >&2
echo "- **成功**: $(($(echo "$stats" | jq '.total') - $(echo "$stats" | jq '.errors'))) 次" >&2
echo "- **错误**: $(echo "$stats" | jq '.errors') 次" >&2
echo "- **错误率**: $(echo "$stats" | jq '.error_rate')%" >&2

# 工具频率分布
if [ -z "$FILTER_EXPR" ] || echo "$FILTER_EXPR" | grep -qE "(AND|OR|IN)"; then
    echo "" >&2
    echo "### 工具分布" >&2
    echo "" >&2
    format_tool_distribution "$result" 5 >&2
fi

echo "" >&2
echo "---" >&2
echo "" >&2
echo "💡 提示: /meta-query-tools Bash 或 \"status='error'\" 过滤查询" >&2
```

## 示例

```bash
/meta-query-tools                    # 最近 20 次工具调用
/meta-query-tools Bash              # 所有 Bash 调用
/meta-query-tools "status='error'"  # 所有错误
```

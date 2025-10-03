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
        "- **UUID**: \(.UUID)\n" +
        "- **错误**: \(.Error)\n" +
        "- **输出**: \(.Output)\n" +
        "- **输入**: \(.Input | to_entries | map("\(.key)=\(.value)") | join(", "))\n"
    '
else
    # 正常模式：简洁列表
    echo "$result" | jq -r '.[] |
        "\(if .Status == "error" or .Error != "" or (.Output | contains("error")) then "❌" else "✅" end) **\(.ToolName)** (\(.UUID[0:8]))"
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

## 示例

### 查看最近的工具调用
```bash
/meta-query-tools
# 显示最近 20 次工具调用
```

### 按工具过滤
```bash
/meta-query-tools Bash
# 显示所有 Bash 调用（最近 20 次）
```

### 查找错误
```bash
/meta-query-tools "" error 10
# 显示最近 10 次错误（任何工具）
```

### 组合过滤
```bash
/meta-query-tools Edit error
# 显示所有 Edit 工具的错误
```

## 使用场景

- 快速检查最近的工具调用情况
- 查找特定工具的错误
- 分析工具使用分布
- 调试工具调用问题

## 相关命令

- `/meta-errors`：详细错误分析
- `/meta-stats`：会话统计信息
- `@meta-coach`：深入分析和建议

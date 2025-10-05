---
name: meta-query-messages
description: 搜索当前项目最新会话的用户消息（Phase 13：默认项目级）
allowed_tools: [Bash]
argument-hint: [pattern] [limit]
---

# meta-query-messages: 用户消息搜索

Phase 13 更新：默认分析当前项目的最新会话。

使用 query 命令搜索用户消息，支持正则表达式模式匹配。

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
    "### \(.timestamp)\n" +
    "\(.content | .[0:300])\(if (.content | length) > 300 then "..." else "" end)\n" +
    "---\n"
'

echo ""

# 显示统计
echo "📊 **统计**："
echo "- 显示: $count 条（最新）"

if [ "$count" -eq "$LIMIT" ]; then
    echo "- 已达到限制（可能有更多结果，增加 limit 参数查看）"
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

| 模式 | 说明 | 匹配示例 |
|------|------|---------|
| `error` | 精确匹配 "error" | "There's an error" ✅ |
| `error\|bug` | 匹配 "error" 或 "bug" | "Fix bug" ✅, "Handle error" ✅ |
| `^Continue` | 以 "Continue" 开头 | "Continue with..." ✅ |
| `test$` | 以 "test" 结尾 | "Run the test" ✅ |
| `fix.*bug` | "fix" 后跟任意字符，再跟 "bug" | "fix this bug" ✅ |
| `Phase [0-9]` | "Phase" 后跟数字 | "Phase 8" ✅, "Phase 1" ✅ |
| `.*` | 所有消息 | 任何消息 ✅ |

## 示例

### 查找特定内容
```bash
/meta-query-messages "Phase 8"
# 查找包含 "Phase 8" 的所有消息
```

### 使用正则表达式
```bash
/meta-query-messages "error|bug"
# 查找包含 "error" 或 "bug" 的消息
```

### 自定义结果数量
```bash
/meta-query-messages "test" 20
# 显示 20 条包含 "test" 的消息
```

### 复杂模式
```bash
/meta-query-messages "fix.*bug"
# 查找 "fix" 和 "bug" 之间有内容的消息
```

## 使用场景

- 查找过往讨论的特定主题
- 追踪问题报告和修复过程
- 分析用户请求模式
- 回顾项目进展和里程碑

## 相关命令

- `/meta-stats`：会话统计信息
- `/meta-errors`：错误模式分析
- `@meta-coach`：深入分析和建议

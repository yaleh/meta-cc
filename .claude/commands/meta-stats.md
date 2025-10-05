---
name: meta-stats
description: 显示当前 Claude Code 会话的统计信息（Turn 数量、工具使用频率、错误率、会话时长等）
allowed_tools: [Bash]
---

# meta-stats：会话统计分析

运行以下命令获取当前会话的统计信息：

```bash
# 检查 meta-cc 是否安装
if ! command -v meta-cc &> /dev/null; then
    echo "❌ 错误：meta-cc 未安装或不在 PATH 中"
    echo ""
    echo "请安装 meta-cc："
    echo "  1. 下载或构建 meta-cc 二进制文件"
    echo "  2. 将其放置在 PATH 中（如 /usr/local/bin/meta-cc）"
    echo "  3. 确保可执行权限：chmod +x /usr/local/bin/meta-cc"
    echo ""
    echo "详情参见：https://github.com/yale/meta-cc"
    exit 1
fi

# Phase 13: JSONL output (default), Claude renders to Markdown
echo "📊 Session Statistics" >&2
echo "" >&2

# Basic session stats (JSONL output, logs to stderr)
stats_json=$(meta-cc parse stats 2>/dev/null)

# Parse JSONL and render as Markdown
echo "$stats_json" | jq -r '
"# Session Statistics\n" +
"- **Total Turns**: \(.TurnCount)\n" +
"- **Tool Calls**: \(.ToolCallCount)\n" +
"- **Error Count**: \(.ErrorCount)\n" +
"- **Error Rate**: \(.ErrorRate)%\n" +
"- **Session Duration**: \(.DurationSeconds / 60 | floor) minutes\n" +
"\n## Tool Usage Frequency\n" +
"| Tool | Count | Percentage |\n" +
"|------|-------|------------|\n" +
((.TopTools // []) | .[] | "| \(.Name) | \(.Count) | \(.Percentage)% |") +
"\n"'

# Phase 10: Aggregated statistics by tool
echo "" >&2
echo "## Aggregated Statistics by Tool" >&2
echo "" >&2
agg_json=$(meta-cc stats aggregate --group-by tool --metrics "count,error_rate" 2>/dev/null || echo "[]")
echo "$agg_json" | jq -s -r '
if length > 0 then
  "| Tool | Count | Error Rate |\n" +
  "|------|-------|------------|\n" +
  (.[] | .[] | "| \(.group_value) | \(.metrics.count) | \(.metrics.error_rate * 100 | tostring + "%" |) ")
else
  "⚠️  No aggregation data available"
end'

# Phase 11: Exit code handling
EXIT_CODE=$?
if [ $EXIT_CODE -eq 2 ]; then
    echo "ℹ️  No data available for aggregation" >&2
fi
```

## 说明

此命令分析当前 Claude Code 会话，提供以下统计信息：

- **Turn 数量**：会话中的对话轮次总数
- **工具调用次数**：使用工具的总次数
- **错误率**：工具调用失败的百分比
- **会话时长**：从第一个 Turn 到最后一个 Turn 的时间跨度
- **工具使用频率**：每种工具的使用次数排名

## 输出示例

```markdown
# Session Statistics

- **Total Turns**: 245
- **Tool Calls**: 853
- **Error Count**: 0
- **Error Rate**: 0.00%
- **Session Duration**: 3h 42m

## Tool Usage Frequency

| Tool | Count | Percentage |
|------|-------|------------|
| Bash | 320 | 37.5% |
| Read | 198 | 23.2% |
| Edit | 156 | 18.3% |
| Grep | 89 | 10.4% |
| Write | 90 | 10.6% |
```

## 使用场景

- 快速了解会话的整体情况
- 检查是否有工具使用异常（错误率过高）
- 评估会话效率（Turn 数量 vs 工具调用次数）
- 发现工具使用偏好（某些工具是否被过度使用）

## 相关命令

- `/meta-errors`：分析错误模式

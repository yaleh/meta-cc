---
name: meta-timeline
description: 生成当前项目最新会话的时间线视图（Phase 14：标准化工具）
allowed_tools: [Bash]
argument-hint: [limit]
---

# meta-timeline：会话时间线视图

生成会话的时间线，可视化展示工具使用和错误分布。

```bash
# Source shared utilities
source "$(dirname "$0")/../lib/meta-utils.sh"
check_meta_cc_installed

# 设置显示的最大 turns 数量
LIMIT=${1:-50}

echo "# 会话时间线（最近 ${LIMIT} Turns）"
echo ""

# Phase 14: Query tools with JSONL output
tools_jsonl=$(meta-cc query tools --limit "$LIMIT" 2>/dev/null)
tools_data=$(jsonl_to_json "$tools_jsonl")

# 生成时间线
echo "$tools_data" | jq -r '
to_entries[] |
"\(.key + 1). **\(.value.ToolName)** \(if .value.Status == "error" or (.Error | length) > 0 then "❌" else "✅" end)"
'

echo ""
echo "---"
echo ""

# 统计摘要
echo "## 统计摘要（最近 ${LIMIT} Turns）"
echo ""

stats=$(calculate_error_stats "$tools_data")
echo "- **总工具调用**: $(echo "$stats" | jq '.total') 次"
echo "- **错误次数**: $(echo "$stats" | jq '.errors') 次"
echo "- **错误率**: $(echo "$stats" | jq '.error_rate')%"
echo ""
echo "### Top 工具"
format_tool_distribution "$tools_data" 5

echo ""
echo "---"
echo ""

# 错误分析 (Phase 14: use query errors)
echo "## 错误分析"
echo ""

error_count=$(echo "$stats" | jq '.errors')

if [ "$error_count" -eq 0 ]; then
    echo "✅ 在最近 ${LIMIT} Turns 中未检测到错误。"
else
    echo "检测到 ${error_count} 个错误，运行错误模式分析..."
    echo ""

    # Phase 14: Use query errors + jq (windowing in jq)
    errors_jsonl=$(meta-cc query errors 2>/dev/null)
    errors_data=$(jsonl_to_json "$errors_jsonl")

    # Get last N errors matching the window
    echo "$errors_data" | jq -r --argjson limit "$LIMIT" '
        .[-$limit:] |
        group_by(.signature) |
        map({
            signature: .[0].signature,
            tool_name: .[0].tool_name,
            count: length,
            sample_error: .[0].error,
            time_span: ((.[- 1].timestamp | fromdateiso8601) - (.[0].timestamp | fromdateiso8601))
        }) |
        sort_by(-.count) |
        .[] |
        "### \(.tool_name) 错误\n" +
        "- **签名**: `\(.signature)`\n" +
        "- **次数**: \(.count)\n" +
        "- **时间跨度**: \(.time_span) 秒\n"
    '
fi

echo ""
echo "---"
echo ""
echo "💡 **提示**："
echo "- 使用 /meta-timeline 20 查看最近 20 Turns"
echo "- 使用 /meta-errors 查看完整错误分析"
echo "- 使用 @meta-coach 获取优化建议"
```

## 使用场景

可视化工作流程、识别瓶颈、分析效率、快速诊断问题发生的时间点。

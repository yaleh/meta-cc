---
name: meta-errors
description: 错误模式分析（Phase 14：标准化工具 + 简化查询）
allowed_tools: [Bash]
---

# meta-errors: 错误模式分析

分析会话中的错误模式，提供优化建议。

```bash
# Source shared utilities
source "$(dirname "$0")/../lib/meta-utils.sh"
check_meta_cc_installed

echo "## 错误数据提取" >&2
echo "" >&2

# Phase 14: Use query errors command (JSONL output)
errors_jsonl=$(meta-cc query errors 2>/dev/null)
exit_code=$?

if [ $exit_code -eq 2 ]; then
    echo "✅ 当前会话未检测到错误。" >&2
    exit 0
elif [ $exit_code -eq 1 ]; then
    echo "❌ 查询执行失败。" >&2
    exit 1
fi

errors_data=$(jsonl_to_json "$errors_jsonl")
error_count=$(echo "$errors_data" | jq 'length')
echo "检测到 $error_count 个错误工具调用。" >&2
echo "" >&2

# 聚合错误模式
echo "## 错误模式分析"
echo ""

patterns=$(echo "$errors_data" | jq 'if length > 0 then
    group_by(.signature) | map({
        signature: .[0].signature,
        tool_name: .[0].tool_name,
        count: length,
        first_seen: .[0].timestamp,
        last_seen: .[-1].timestamp,
        sample_error: .[0].error,
        time_span_seconds: ((.[- 1].timestamp | fromdateiso8601) - (.[0].timestamp | fromdateiso8601))
    }) | sort_by(-.count)
else
    []
end')

pattern_count=$(echo "$patterns" | jq 'length')

if [ "$pattern_count" -eq 0 ]; then
    echo "✅ 未检测到错误。"
    exit 0
fi

echo "# 错误模式分析"
echo ""
echo "发现 $pattern_count 个错误模式："
echo ""

# 显示模式（限制 top 10）
if [ "$pattern_count" -gt 10 ]; then
    echo "⚠️  检测到大量错误 ($pattern_count 个模式)"
    echo "显示 Top 10 模式以防止上下文溢出。"
    echo ""
    patterns_to_show=$(echo "$patterns" | jq '.[:10]')
else
    patterns_to_show="$patterns"
fi

echo "$patterns_to_show" | jq -r '.[] |
    "\n## 模式: \(.tool_name)\n" +
    "- **签名**: `\(.signature)`\n" +
    "- **次数**: \(.count) 次\n" +
    "- **错误**: \(.sample_error)\n" +
    "\n### 上下文\n" +
    "- **首次出现**: \(.first_seen)\n" +
    "- **最后出现**: \(.last_seen)\n" +
    "- **时间跨度**: \(.time_span_seconds) 秒\n" +
    "\n---\n"'

echo ""
echo "---"
echo ""
echo "## 优化建议"
echo ""
echo "1. 调查重复错误 - 查看错误文本识别根本原因"
echo "2. 使用 Hooks 预检查 - 创建钩子防止错误"
echo "3. 调整工作流 - 考虑替代工具或优化提示词"
```

## 高级查询

```bash
# 最近 50 个错误
meta-cc query errors | jq '.[-50:]'

# 按工具过滤
meta-cc query errors | jq '[.[] | select(.tool_name == "Bash")]'
```

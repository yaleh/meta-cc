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

# Phase 9: Adaptive output strategy based on size estimation
# Step 1: Estimate output size
ESTIMATE=$(meta-cc parse stats --estimate-size --output json 2>/dev/null)

if [ $? -eq 0 ]; then
    SIZE_KB=$(echo "$ESTIMATE" | grep -o '"estimated_kb":[0-9.]*' | cut -d: -f2)

    # Step 2: Choose output strategy based on size
    if [ -z "$SIZE_KB" ]; then
        # Fallback: estimation failed, use default
        meta-cc parse stats --output md
    elif (( $(echo "$SIZE_KB < 50" | bc -l 2>/dev/null || echo 0) )); then
        # Small session (<50KB): full markdown output
        echo "📊 Session Statistics (Full Report)"
        echo ""
        meta-cc parse stats --output md
    elif (( $(echo "$SIZE_KB < 200" | bc -l 2>/dev/null || echo 0) )); then
        # Medium session (50-200KB): TSV with key fields
        echo "📊 Session Statistics (Compact Format)"
        echo ""
        echo "Estimated size: ${SIZE_KB} KB"
        echo ""
        meta-cc parse stats --output md
    else
        # Large session (>200KB): Summary mode
        echo "📊 Session Statistics (Summary - Large Session)"
        echo ""
        echo "⚠️  Large session detected (${SIZE_KB} KB)"
        echo "Showing summary to prevent context overflow."
        echo ""
        meta-cc parse stats --output md
        echo ""
        echo "💡 Tip: Use 'meta-cc parse stats --output tsv' for ultra-compact output"
    fi
else
    # Fallback: estimation command not supported or failed
    meta-cc parse stats --output md
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

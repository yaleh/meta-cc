---
name: meta-timeline
description: 生成当前项目最新会话的时间线视图（Phase 13：默认项目级）
allowed_tools: [Bash]
argument-hint: [limit]
---

# meta-timeline：会话时间线视图

**Phase 13 更新**: 默认分析当前项目的最新会话。使用 `query tools` 命令，支持高效分页。

生成会话的时间线，可视化展示工具使用和错误分布。

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

# 设置显示的最大 turns 数量
LIMIT=${1:-50}

echo "# 会话时间线（最近 ${LIMIT} Turns）"
echo ""

# 使用 Phase 8 query 命令（支持分页，避免大会话上下文溢出）
# Phase 13: JSONL output by default
tools_jsonl=$(meta-cc query tools --limit "$LIMIT" 2>/dev/null)

# Convert JSONL to JSON array for jq processing
tools_data=$(echo "$tools_jsonl" | jq -s '.')

# 解析 JSON 并生成时间线
# query 命令已经限制了数量，直接使用结果
echo "$tools_data" | jq -r '
to_entries[] |
"\(.key + 1). **\(.value.ToolName)** \(if .value.Status == "error" or .value.Error != "" then "❌" else "✅" end)"
'

echo ""
echo "---"
echo ""

# 显示统计摘要
echo "## 统计摘要（最近 ${LIMIT} Turns）"
echo ""
echo "$tools_data" | jq -r '
{
  total: length,
  errors: [.[] | select(.Status == "error" or .Error != "")] | length,
  tools: [.[] | .ToolName] | group_by(.) | map({tool: .[0], count: length}) | sort_by(.count) | reverse
} |
"- **总工具调用**: \(.total) 次",
"- **错误次数**: \(.errors) 次",
"- **错误率**: \(if .total > 0 then (.errors / .total * 100 | floor) else 0 end)%",
"",
"### Top 工具",
(.tools[:5] | .[] | "- \(.tool): \(.count) 次")
'

echo ""
echo "---"
echo ""

# 错误分析
echo "## 错误分析"
echo ""

error_count=$(echo "$tools_data" | jq '[.tools | .[] | select(.status == "error")] | length')

if [ "$error_count" -eq 0 ]; then
    echo "✅ 在最近 ${LIMIT} Turns 中未检测到错误。"
else
    echo "检测到 ${error_count} 个错误，运行错误模式分析..."
    echo ""
    meta-cc analyze errors --window "$LIMIT" --output md | tail -n +2
fi

echo ""
echo "---"
echo ""
echo "💡 **提示**："
echo "- 使用 /meta-timeline 20 查看最近 20 Turns"
echo "- 使用 /meta-errors 查看完整错误分析"
echo "- 使用 @meta-coach 获取优化建议"
```

## 说明

此命令生成会话的时间线视图，帮助：

- **可视化工作流程**：按时间顺序查看工具使用
- **识别瓶颈**：发现哪些环节出现密集错误
- **分析效率**：观察工具调用的节奏和模式
- **快速诊断**：定位问题发生的时间点

## 参数

- `limit`（可选）：显示最近 N 个 Turns，默认 50

## 输出内容

### 时间线列表
按时序显示每个工具调用：
- Turn 序号
- 工具名称
- 状态标记（✅ 成功 / ❌ 错误）

### 统计摘要
- 总工具调用次数
- 错误次数和错误率
- Top 5 工具使用频率

### 错误分析
如果存在错误：
- 自动运行 `meta-cc analyze errors`
- 显示重复错误模式

## 使用场景

1. **Debug 会话回顾**：查看解决问题的过程
2. **效率分析**：识别重复或低效的操作
3. **学习最佳实践**：回顾成功解决问题的步骤
4. **工作日志**：为 Stand-up 或 Retro 准备素材

## 示例输出

```markdown
# 会话时间线（最近 50 Turns）

1. Turn 15 - **Bash** ✅
2. Turn 15 - **Read** ✅
3. Turn 17 - **Grep** ✅
4. Turn 19 - **Edit** ✅
5. Turn 21 - **Bash** ❌
6. Turn 23 - **Bash** ❌
7. Turn 25 - **Bash** ❌
8. Turn 27 - **Read** ✅
...

---

## 统计摘要（最近 50 Turns）

- **总工具调用**: 48 次
- **错误次数**: 3 次
- **错误率**: 6%

### Top 工具
- Bash: 18 次
- Read: 12 次
- Edit: 8 次
- Grep: 6 次
- Write: 4 次

---

## 错误分析

检测到 3 个错误，运行错误模式分析...

## Pattern 1: Bash

- **Type**: command_error
- **Occurrences**: 3 times
- **Signature**: `npm test`
- **Error**: FAIL test_auth.js

### Context
- **First Occurrence**: 2025-10-02 14:21:00
- **Last Occurrence**: 2025-10-02 14:27:00
- **Time Span**: 6 minutes
- **Affected Turns**: 3

**建议**: 此错误重复 3 次，考虑：
1. 专注修复该测试而非重复运行
2. 添加 Hook 提醒重复命令
```

## 相关命令

- `/meta-stats`：会话统计摘要
- `/meta-errors`：错误模式分析
- `/meta-compare`：跨项目对比

---
name: meta-errors
description: 分析当前 Claude Code 会话中的错误模式，检测重复出现的错误（可选参数：window-size）
allowed_tools: [Bash]
argument-hint: [window-size]
---

# meta-errors：错误模式分析

分析当前会话中的错误模式，检测重复出现的错误（出现 3 次以上）。

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

# 获取窗口参数（默认 20）
WINDOW_SIZE=${1:-20}

# Step 1: 提取错误数据（用于上下文展示）
echo "## 错误数据提取" >&2
echo "" >&2

# Phase 11: Use streaming with exit codes for errors
meta-cc query tools --where "status='error'" --stream 2>/dev/null > /tmp/meta-errors-$$.jsonl
EXIT_CODE=$?

if [ $EXIT_CODE -eq 2 ]; then
    echo "✅ 当前会话中未检测到错误。" >&2
    rm -f /tmp/meta-errors-$$.jsonl
    exit 0
elif [ $EXIT_CODE -eq 1 ]; then
    echo "❌ 查询错误时出错。" >&2
    rm -f /tmp/meta-errors-$$.jsonl
    exit 1
fi

ERROR_COUNT=$(wc -l < /tmp/meta-errors-$$.jsonl)
rm -f /tmp/meta-errors-$$.jsonl

echo "检测到 $ERROR_COUNT 个错误工具调用。" >&2
echo "" >&2

# Step 2: 分析错误模式（窗口大小：$WINDOW_SIZE）
echo "## 错误模式分析（窗口大小：$WINDOW_SIZE）"
echo ""

# Phase 9: Use summary mode for large error sets
if [ "$ERROR_COUNT" -gt 10 ]; then
    echo "⚠️  Large error set detected ($ERROR_COUNT errors)"
    echo "Showing summary with top 10 patterns to prevent context overflow."
    echo ""
    PATTERN_OUTPUT=$(meta-cc analyze errors --window "$WINDOW_SIZE" --output md 2>/dev/null | head -100)
    echo "$PATTERN_OUTPUT"
    echo ""
    echo "💡 Tip: Use 'meta-cc query tools --where \"status='error'\" --output tsv' for full error list"
else
    PATTERN_OUTPUT=$(meta-cc analyze errors --window "$WINDOW_SIZE" --output md)
    echo "$PATTERN_OUTPUT"
fi

echo ""

# Step 3: 如果检测到错误模式，提供优化建议
if echo "$PATTERN_OUTPUT" | grep -q "## Pattern"; then
    echo "---"
    echo ""
    echo "## 优化建议"
    echo ""
    echo "基于检测到的错误模式，请考虑以下优化措施："
    echo ""
    echo "1. **检查重复错误的根本原因**"
    echo "   - 查看错误文本，识别是否为相同的底层问题"
    echo "   - 检查相关的 Turn 序列，了解错误发生的上下文"
    echo ""
    echo "2. **使用 Claude Code Hooks 预防错误**"
    echo "   - 创建 pre-tool hook 检查常见错误条件"
    echo "   - 例如：文件存在性检查、权限验证、参数格式校验"
    echo ""
    echo "3. **调整工作流**"
    echo "   - 如果错误集中在某个工具，考虑使用替代方案"
    echo "   - 优化提示词以减少错误触发频率"
    echo ""
    echo "4. **查看详细错误列表**"
    echo "   - 运行：\`meta-cc parse extract --type tools --filter \"status=error\" --output md\`"
    echo "   - 分析每个错误的具体原因和上下文"
    echo ""
else
    echo "✅ 未检测到重复错误模式（出现 < 3 次）。"
fi
```

## 参数说明

- `window-size`（可选）：分析最近 N 个 Turn。默认值为 20。
  - 示例：`/meta-errors 50`（分析最近 50 个 Turn）
  - 省略参数：`/meta-errors`（使用默认窗口 20）

## 输出内容

1. **错误数据提取**：统计会话中的错误总数
2. **错误模式分析**：检测重复出现的错误（≥3 次）
3. **优化建议**：基于检测到的模式提供可行的改进措施

## 输出示例

```markdown
## 错误数据提取

检测到 12 个错误工具调用。

## 错误模式分析（窗口大小：20）

# Error Pattern Analysis

Found 2 error pattern(s):

## Pattern 1: Bash

- **Type**: repeated_error
- **Occurrences**: 5 times
- **Signature**: `a3f2b1c4d5e6f7g8`
- **Error**: command not found: xyz

### Context

- **First Occurrence**: 2025-10-02T10:00:00.000Z
- **Last Occurrence**: 2025-10-02T10:15:00.000Z
- **Time Span**: 900 seconds (15.0 minutes)
- **Affected Turns**: 5

---

## 优化建议

基于检测到的错误模式，请考虑以下优化措施：

1. **检查重复错误的根本原因**
   - 查看错误文本，识别是否为相同的底层问题

2. **使用 Claude Code Hooks 预防错误**
   - 创建 pre-tool hook 检查常见错误条件

3. **调整工作流**
   - 如果错误集中在某个工具，考虑使用替代方案
```

## 使用场景

- 识别重复出现的错误，避免重复调试
- 发现工作流中的瓶颈（某些操作频繁失败）
- 获取优化建议（hooks、替代方案、提示词改进）
- 关注最近的错误（使用窗口参数）

## 相关命令

- `/meta-stats`：查看会话统计信息
- `meta-cc parse extract --type errors`：查看所有错误详情

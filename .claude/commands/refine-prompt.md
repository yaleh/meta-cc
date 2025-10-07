---
description: 基于历史成功 prompts 优化当前 prompt（使用 MCP meta-insight）
argument-hint: [prompt]
---

# 基于历史优化 Prompt

使用 MCP meta-insight 分析项目历史中的成功 prompts，为当前 prompt 提供优化建议。

## Original Prompt

```
$1
```

## 分析步骤

1. **使用 `mcp__meta-insight__query_successful_prompts` 获取高质量历史 prompts**
   - 设置 `min_quality_score=0.8`
   - 分析成功 prompt 的结构特征

2. **使用 `mcp__meta-insight__query_user_messages` 查找相似历史请求**
   - 根据用户当前 prompt 的关键词构建搜索模式
   - 查找类似的历史 user messages
   - 分析其上下文和结果

3. **提供优化建议**
   - 对比当前 prompt 和成功案例
   - 指出可改进之处：
     - 缺失的具体细节
     - 模糊的表达
     - 可添加的约束条件
     - 更清晰的目标描述

**Claude Code 最佳实践**：
- ✅ **使用 `@` 引用文件**，避免复制文件内容到 prompt
   - 好：`参考 @docs/plan.md @docs/principles.md`
   - 差：`以下是 plan.md 的内容：[大段复制的内容]`
- ✅ **使用 `@agent-` 调用 subagents**，避免复制 subagent 行为到 prompt
   - 好：`使用 @agent-stage-executor 执行 Stage 13.1`
   - 差：`请按照以下步骤执行：1. 读取计划 2. 执行测试 3. 实现...`
- ✅ **引用具体位置**：文件路径 + 行号范围
   - 好：`plans/13/plan.md Stage 13.1 (Lines 100-250)`

- 给出不超过 3 个优化后的 prompt 示例
  - 为示例编号以便于选择
  - 说明你建议的选择
- 不要执行优化后的示例


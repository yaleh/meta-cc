---
id: TASK-21
title: context_turns 自动跳过 compact summary 消息
assignee: []
created_date: '2026-07-14 03:52'
updated_date: '2026-07-15 01:57'
labels:
  - 'area:mcp'
  - 'area:query'
dependencies: []
priority: medium
ordinal: 12000
pipeline_id: authoring
phase: drafting
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
compact summary 是 Claude Code 注入的上下文压缩边界消息（`isCompactSummary: true`），正文为数千字的历史摘要。当 `context_turns > 0` 时，这类消息会被当作普通对话轮次拉入，导致结果体积暴增（单条可达数万 token），且对用户理解上下文没有帮助。

改进：`context_turns` 逻辑在计算前后窗口时自动跳过 `isCompactSummary=true` 的消息，不将其计入轮次数，也不将其内容包含在 context 结果里。
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 context_turns > 0 时，compact summary 消息（isCompactSummary=true）不出现在 context 结果中
- [ ] #2 compact summary 不占用 context_turns 的轮次计数
- [ ] #3 非 compact summary 的普通 user 消息 context 行为不变
- [ ] #4 现有测试覆盖含 compact summary 的会话文件场景
<!-- AC:END -->

---
id: TASK-22
title: 新增 exclude_compact_summaries 参数，默认排除 compact summary 消息
status: 'Basic: Proposal'
assignee: []
created_date: '2026-07-14 03:53'
labels:
  - 'area:mcp'
  - 'area:query'
dependencies: []
priority: medium
ordinal: 13000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
compact summary 消息（`isCompactSummary: true`）是 Claude Code 自动注入的上下文压缩记录，正文体积巨大且极易命中任意 pattern 搜索（因为摘要里包含大量历史文本）。当前 `role=user` 搜索会把它当作普通用户消息匹配，导致噪声结果。

新增 `exclude_compact_summaries` boolean 参数，**默认 `true`**，将 compact summary 消息同时从：
1. pattern/contains 匹配的候选集中排除
2. `context_turns` 拉入的 context 结果中排除

用户可显式传 `false` 恢复当前行为（如需检索压缩摘要内容本身时）。

与"context_turns 自动跳过 compact summary"任务可叠加实现，本任务提供显式参数控制，另一任务作为 context_turns 的内部默认行为。
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 exclude_compact_summaries 默认 true，compact summary 消息不出现在搜索结果中
- [ ] #2 exclude_compact_summaries=false 时恢复当前行为
- [ ] #3 参数对 role=user 和 role=all 均生效
- [ ] #4 文档说明 compact summary 的定义（isCompactSummary=true）及该参数用途
<!-- AC:END -->

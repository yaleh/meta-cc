---
id: TASK-19
title: query_session_content(role=tool) 增加 tool_name 过滤参数
status: 'Basic: Done'
assignee: []
created_date: '2026-07-14 03:52'
labels:
  - 'area:mcp'
  - 'area:query'
dependencies: []
priority: high
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
当前 `query_session_content(role=tool, block_type=tool_use)` 没有按工具名过滤的能力。`contains` 参数过滤的是消息内容字符串，而 tool_use 的 `name` 字段不在 content 里，无法命中。唯一的 workaround 是 grep 原始 JSONL 文件。

新增 `tool_name` 参数，支持子串（或 regex）匹配 tool_use block 的 `name` 字段：

```javascript
query_session_content({
  role: "tool",
  block_type: "tool_use",
  tool_name: "Dispatch"
})
```
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 tool_name 参数对 role=tool, block_type=tool_use 生效，按 name 字段子串过滤
- [x] #2 tool_name 支持 regex 模式（与现有 pattern 参数行为一致）
- [x] #3 tool_name 不填时行为与当前一致（不过滤）
- [x] #4 文档更新反映新参数
<!-- AC:END -->

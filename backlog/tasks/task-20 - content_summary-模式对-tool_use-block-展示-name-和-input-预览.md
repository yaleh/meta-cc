---
id: TASK-20
title: content_summary 模式对 tool_use block 展示 name 和 input 预览
status: 'Basic: Proposal'
assignee: []
created_date: '2026-07-14 03:52'
labels:
  - 'area:mcp'
  - 'area:query'
dependencies: []
priority: high
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
当前 `content_summary=true` 对 `role=tool, block_type=tool_use` 返回空的 `content_preview`，5000+ 条记录无法区分，该模式对 tool_use 完全失效。

改进：对 tool_use block，preview 拼接 `name` + 截断的 `input` JSON：

```
content_preview: "mcp__manda__Dispatch {\"id\": \"probe-1\", \"to\": \"mon-a\", \"mode\": \"sync\"...}"
```

实现上只需在 summary 模式里把 `name` 和 `input` 的 JSON 前 N 个字符拼进 preview，N 由 `preview_length` 参数控制。
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 role=tool, block_type=tool_use, content_summary=true 时 content_preview 非空
- [ ] #2 preview 格式为 '<name> <input_json_truncated>'
- [ ] #3 preview 长度受 preview_length 参数控制
- [ ] #4 tool_result block 的 preview 行为不受影响
<!-- AC:END -->

---
id: TASK-13
title: 'D: 在 MCP 工具 description 字段嵌入使用指引（时间过滤/字段路径/先 inspect 再 transform）'
status: 'Basic: Backlog'
assignee: []
created_date: '2026-06-23 15:50'
updated_date: '2026-06-23 15:52'
labels:
  - 'kind:basic'
dependencies:
  - TASK-11
  - TASK-12
references:
  - internal/mcp/tools/tools.go
  - docs/guides/mcp-query-tools.md
priority: medium
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
方案一：嵌入 MCP 工具的 description 字段。meta-cc 安装后工具 schema 自动加载到每个 session，是唯一无需项目配置即可覆盖所有下游项目的位置。当前工具 description 仅描述功能，不含使用约定，导致 agent 反复犯同类错误（since 参数幻觉、jq 字段路径落空）。应将最关键的 3 条使用约定以紧凑语言嵌入相关工具的 description：(1) 时间过滤用 since/until 参数，不用 jq_filter；(2) role=tool 结果含 timestamp/session_id/turn；(3) 自定义 transform 前先用 inspect_session_files(include_samples=true) 确认字段路径。
<!-- SECTION:DESCRIPTION:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Proposal: 在 MCP 工具 description 嵌入使用指引

## Background

meta-cc 安装后，MCP 工具 schema（含 description 字段）在每个 session 启动时自动加载，无需项目做任何配置。这是唯一能覆盖所有下游项目 agent 的位置。当前三个核心查询工具的 description 仅描述功能，不含使用约定，导致 agent 反复犯以下三类同质错误：

1. **since 参数幻觉**：agent 对不支持 since 的旧工具传 since，引发 MCP error -32603；现已迁移到 query_session_signals，但 agent 不知道 since/until 是一等参数，仍尝试用 jq_filter 做时间过滤。
2. **role=tool 字段路径错误**：query_session_content(role=tool) 输出记录含外层字段 timestamp/session_id/turn（TASK-11 实现），但 agent 在 jq transform 中使用内层 block 字段，导致全空结果。
3. **execute_stage2_query transform 静默失败**：transform 字段路径与实际 JSONL 结构不匹配时，全空结果静默返回（TASK-12 添加 warning），agent 误信或放弃。

每次错误都需要 agent 多轮纠错，浪费 token，且常常无法自行恢复。在 description 中嵌入紧凑使用约定，可在 agent 调用前即建立正确预期，是最低成本的预防措施。

## Goals

1. `query_session_content` description 包含 role=tool 输出字段说明（timestamp/session_id/turn）和 inspect 前置建议（可用 `grep "inspect_session_files" internal/mcp/tools/tools.go` 验证）
2. `query_session_signals` description 包含 since/until 为一等参数的说明（可用 `grep "since" internal/mcp/tools/tools.go` 并确认在工具 description 中）
3. `execute_stage2_query` description 包含 transform 字段路径需与 inspect_session_files 样本对齐的建议（可用 `grep "inspect_session_files" internal/mcp/tools/tools.go` 验证）
4. 现有所有测试通过（`go test ./...`）

## Proposed Approach

修改 `internal/mcp/tools/tools.go` 中三个工具的 `BuildTool()` 调用（或直接量表定义）里的 description 字符串：

- **query_session_content**（第 240-315 行）：在 description 末尾追加一行：`"role=tool outputs: {timestamp, session_id, turn, ...block fields}. For custom transform, first call inspect_session_files(include_samples=true)."`
- **query_session_signals**（第 317-348 行）：在 description 末尾追加一行：`"Time filtering: use since/until params directly—no need for jq_filter time conditions."`
- **execute_stage2_query**（第 147-179 行，内联定义）：在 Description 末尾追加一行：`"Align transform field paths with inspect_session_files(include_samples=true) samples. Empty results trigger a warning."`

约束：每个工具新增内容不超过 2 行（~80 字符/行），不修改参数 schema。

## Trade-offs and Risks

**不做什么**：不修改参数 schema（不新增/删除参数）；不新增工具；不创建文档文件；不修改其他工具的 description。

**依赖**：description 文本引用 TASK-11 实现的 role=tool 外层字段和 TASK-12 实现的全空 warning，须在这两个任务完成后 description 才与实际行为一致。

**风险**：description 变长会小幅增加每个 session 的初始 token 消耗（约 +50 token，可忽略）。description 过长则难以阅读，本方案严格控制在 1-2 行/工具。

---

# Plan: D - MCP 工具 description 嵌入使用指引

## Phase A: 修改 tools.go 三个工具的 description

### Tests (write first)

- 文件：`internal/mcp/tools/tools_test.go`
- 测试用例：
  - `TestQuerySessionContentDescriptionHints`：在 GetToolDefinitions() 中找到 query_session_content，验证其 Description 同时含 `"timestamp"` 和 `"inspect_session_files"`
  - `TestQuerySessionSignalsDescriptionHints`：找到 query_session_signals，验证 Description 含 `"since"` 且含 `"until"`（证明 description 主动提及参数名，而非仅靠参数 property）
  - `TestExecuteStage2QueryDescriptionHints`：找到 execute_stage2_query，验证 Description 同时含 `"inspect_session_files"` 和 `"transform"`

### Implementation

- 文件：`internal/mcp/tools/tools.go`
- 修改：
  1. `query_session_content`：将 description 从 `"Query session messages by role (user/assistant/tool/all). Default scope: project."` 改为追加一句话，含 `timestamp/session_id/turn` 和 `inspect_session_files` 关键词
  2. `query_session_signals`：将 description 追加一句，含 `since`/`until` 关键词
  3. `execute_stage2_query`（内联 Description 字段）：追加一句，含 `inspect_session_files` 和 `transform` 关键词

### DoD

- [ ] `go test ./... -run TestQuerySessionContentDescriptionHints`
- [ ] `go test ./... -run TestQuerySessionSignalsDescriptionHints`
- [ ] `go test ./... -run TestExecuteStage2QueryDescriptionHints`
- [ ] `grep -q "inspect_session_files" internal/mcp/tools/tools.go` （description 中出现，不仅是工具名）
- [ ] `grep -q "since" internal/mcp/tools/tools.go`
- [ ] `go test ./...`（全量回归通过）

## Constraints

- 每个工具的 description 新增内容不超过 2 行（~80 字符/行）
- 不修改参数 schema（不新增/删除参数）
- 不创建新文件（仅修改 `internal/mcp/tools/tools.go` 和 `internal/mcp/tools/tools_test.go`）

## Acceptance Gate

- [ ] `go test ./...`
- [ ] `grep -q "inspect_session_files" internal/mcp/tools/tools.go`
- [ ] `grep -q "timestamp/session_id/turn" internal/mcp/tools/tools.go`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Proposal + Plan APPROVED (1 phase, ~20 LOC). Ready for implementation after TASK-11 and TASK-12.
<!-- SECTION:NOTES:END -->

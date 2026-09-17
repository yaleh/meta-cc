---
id: ac239-subagent-session-id-scan
title: 修复 include_subagents 在显式 session_id 上传参时静默失效
status: todo
labels: []
parent: null
children: []
extra: {}
goal_ac: GOAL-E2E-239
---
## Proposal

meta-cc 的 MCP 查询工具在【显式传入 `session_id`】时，`include_subagents` 参数静默失效：磁盘上
`<session-id>/subagents/agent-*.jsonl` 里明明有目标内容，查询却返回 0 条，**且不报错、不告警**——
「查过且合格」与「根本没查」在返回值上同形。

实测（可复现，非推断）：取一根只存在于某会话 `subagents/agent-*.jsonl`、不在该会话主 `.jsonl` 里的
针，先用 `grep -rl` 在文件系统上确认两侧计数（主会话文件 0 次、子代理文件 ≥1 次），再对同一会话调用
`query_session_content(session_id=<sid>, include_subagents=true, contains=<针>)` ⇒ 返回 0 条；
而同一次查询若改走 `scope=session`（不传 `session_id`），同一根针能被找到。
⇒ 差异被定位在**传参形态**上，不是「那根针不存在」。

`internal/mcp/query/query.go` 的注释只承诺了两种取值（`scope=session` 与 `scope=project` 配
`includeSubagents=true` 时的文件展开），**显式 `session_id` 是第三种取值，注释里没有它**——
这与实测形态一致：该路径很可能没有接上 `GetQueryFiles` 的 subagent 目录展开。

⛔ 上面是**待确认的线索，不是结论**。根因与修法必须在实现时到源码里实际定位（`query.go` /
`stage.go` / `query_files_test.go` 与 `executor/handlers.go` 的调用链），不得照抄本段的猜测。

## Plan

1. 复现并定位：构造「主会话文件无针 ∧ `<sid>/subagents/agent-*.jsonl` 有针」的最小 fixture，
   断言显式 `session_id` + `include_subagents=true` 时能查到 ⇒ 先看到它 FAIL。
2. 在 `internal/mcp/query/` 的对应读取路径上把 subagent 目录展开接上（具体落点由第 1 步的定位决定）。
3. 让第 1 步的测试转 PASS；`go build ./...` 与相关包既有测试保持通过。

## Acceptance Criteria

- [ ] AC1 新增/修改的 Go 测试在【修复前】失败、在【修复后】通过；两条命令与真实输出贴进本任务
      （修复前的失败证据 = 把修复改动反向应用或 `git stash` 后跑同一条测试命令）。
- [ ] AC2 该测试的判据是**按位置**的：针只存在于 `<session>/subagents/*.jsonl`，在主会话文件里
      一次都不出现（夹具自己先断言这一点，⛔ 不靠文件名或注释声称）。
- [ ] AC3 `go build ./...` 通过。
- [ ] AC4 本次改动涉及的既有测试通过（至少 `go test ./internal/mcp/query/... ./internal/mcp/executor/...`）。

## Definition of Done

- [ ] 缺陷在源码层面被修复（不是把测试改成绕过它），修复落在 `internal/mcp/` 的查询路径上，改动可在 git log 中查到。
- [ ] AC1 要求的「修复前失败 / 修复后通过」两条真实输出已贴进任务记录或提交信息。
- [ ] 未被改坏的行为：显式传 `include_subagents=false` 时依旧不展开 subagent 目录。

## Touches

- internal/mcp/query/query.go
- internal/mcp/query/stage.go
- internal/mcp/query/query_files_test.go
- tasks/ac239-subagent-session-id-scan.md

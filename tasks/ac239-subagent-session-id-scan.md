---
id: ac239-subagent-session-id-scan
title: 修复 include_subagents 在显式 session_id 上传参时静默失效
status: done
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
`includeSubagents=true` 时的文件展开），**显式 `session_id` 是第三种取值，注释里没有它**——这与实测形态一致：该路径没有接上 `GetQueryFiles` 的 subagent 目录展开。

## Plan

1. 复现并定位：构造「主会话文件无针 ∧ `<sid>/subagents/agent-*.jsonl` 有针」的最小 fixture，
   断言显式 `session_id` + `include_subagents=true` 时能查到 ⇒ 先看到它 FAIL。
2. 在 `internal/mcp/query/` 的对应读取路径上把 subagent 目录展开接上（具体落点由第 1 步的定位决定）。
3. 让第 1 步的测试转 PASS；`go build ./...` 与相关包既有测试保持通过。

## Root Cause（实现时在源码中实际定位，替代 Proposal 的线索）

`internal/mcp/executor/provider_query.go` 的 `dispatchProviderQuery` 以「sessionID 非空」为唯一
分叉条件：非空即路由到 `ExecuteQueryForSession`。该函数用
`StreamFilesWithTimeRange(ctx, []string{file}, ...)` 只流式读取**一个**主 transcript 文件——
它当时**连接收 `includeSubagents` 的形参都没有**，所以参数在分叉处被丢弃，不是「读了但没生效」，
而是根本没有可落地的接收方。`GetQueryFiles` 也帮不上忙：它按 scope 解析，而 `scope="session"`
的语义是「最近一个会话」，**不是** `session_id` 指定的那个确定线程，所以精确 ID 路径原本没有任何
通往 subagent 展开的入口。

## Fix

- `internal/mcp/query/query.go`：新增 `GetSessionQueryFilesWithSubagents(sessionFile, includeSubagents)`，
  从**已解析好的**会话文件反推 `<projectDir>/<uuid>/subagents/`——因此对「不是最新会话」的
  session_id 同样正确，且不重扫项目目录。
- `internal/mcp/executor/provider_query.go`：把 `includeSubagents` 一路穿透到
  `ExecuteQueryForSession`（variadic，既有直接调用方保持可编译，省略时的零值即文档化默认
  `true`），并在文件列表构造处使用新 helper。

## Acceptance Criteria

- [x] AC1 新增/修改的 Go 测试在【修复前】失败、在【修复后】通过；两条命令与真实输出贴进本任务
      （修复前的失败证据 = 把修复改动反向应用或 `git stash` 后跑同一条测试命令）。
- [x] AC2 该测试的判据是**按位置**的：针只存在于 `<session>/subagents/*.jsonl`，在主会话文件里
      一次都不出现（夹具自己先断言这一点，⛔ 不靠文件名或注释声称）。
- [x] AC3 `go build ./...` 通过。
- [x] AC4 本次改动涉及的既有测试通过（至少 `go test ./internal/mcp/query/... ./internal/mcp/executor/...`）。

## Verification Evidence

修复已落在 `internal/mcp/`（提交 `d776870`）。AC1 的两条真实输出：

修复前（`git stash` 反向应用修复源码、保留测试后跑同一条命令）：

```
$ go test ./internal/mcp/executor/... -run TestQuerySessionContent_ExplicitSessionID_IncludeSubagents_FindsSubagentOnlyContent
--- FAIL: TestQuerySessionContent_ExplicitSessionID_IncludeSubagents_FindsSubagentOnlyContent (0.00s)
    session_id_include_subagents_test.go:110:
        Error:  "[]" should have 1 item(s), but has 0
        Messages: explicit session_id + include_subagents=true must reach
        <session>/subagents/*.jsonl, where the needle is the only occurrence
FAIL github.com/yaleh/meta-cc/internal/mcp/executor 0.010s
```

修复后：

```
--- PASS: TestQuerySessionContent_ExplicitSessionID_IncludeSubagents_FindsSubagentOnlyContent (0.00s)
--- PASS: TestQuerySessionContent_ExplicitSessionID_ExcludeSubagents_DoesNotExpand (0.00s)
ok  github.com/yaleh/meta-cc/internal/mcp/executor 0.010s
```

AC2（按位置，非按声称）：夹具在写盘后把两个文件**读回来**计数——针在主会话文件中出现 0 次、
在子代理文件中出现恰好 1 次；若回归把针挪进主文件（会让测试因错误的原因变绿），夹具构造阶段即失败。

AC3/AC4：`go build ./...` exit 0；`go test ./internal/mcp/query/... ./internal/mcp/executor/...` 两包均 ok；
合并 develop 后 `go test ./...` 全绿（scoped gate）。

### 本轮复核（merge develop 之后重新取证）

`git merge develop` 干净合入（仅 `tasks/DIR-085.md`）。在本 worktree 内重新执行：

- **AC1 反向复核**（`git checkout develop -- internal/mcp/executor/provider_query.go internal/mcp/query/query.go`
  只回退源码、保留新测试）：`--- FAIL ... "[]" should have 1 item(s), but has 0`，exit 1；
  还原后同一条命令 `--- PASS`。修复前失败/修复后通过，两侧均为实跑输出。
- **AC3** `go build ./...` → exit 0。
- **AC4** `go test ./internal/mcp/query/... ./internal/mcp/executor/...` → ok；
  `go test ./...`（scoped gate）→ 全绿 exit 0。

## Definition of Done

- [x] 缺陷在源码层面被修复（不是把测试改成绕过它），修复落在 `internal/mcp/` 的查询路径上，改动可在 git log 中查到。
- [x] AC1 要求的「修复前失败 / 修复后通过」两条真实输出已贴进任务记录或提交信息。
- [x] 未被改坏的行为：显式传 `include_subagents=false` 时依旧不展开 subagent 目录。

## Touches

- internal/mcp/query/query.go
- internal/mcp/executor/provider_query.go
- internal/mcp/executor/session_id_include_subagents_test.go
- tasks/ac239-subagent-session-id-scan.md

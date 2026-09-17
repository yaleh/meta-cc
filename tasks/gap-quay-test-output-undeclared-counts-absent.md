---
id: gap-quay-test-output-undeclared-counts-absent
title: quay web 测试面板 pass/tests 恒为 —/—：meta-cc 未声明 loop.test_output 且 go test 无总计行
status: ready
labels:
  - gap
parent: null
children: []
extra:
  schema: execution
  acceptance: make commit
  scope.owner_repo: meta-cc
---
## Proposal

**Gap (measured on this host, 2026-09-17).** quay 的 web 测试卡片把 meta-cc 的每一轮都渲染成
`pass —/—`。计数是**如实缺席**，不是零——quay 刻意不把「没测到」写成 `0`。

卡片的数据源是 `<root>/.quay/verification-round.jsonl`（一轮 suite 一行），由
`readTestsUncached` 读取（`packages/quay/src/observation.ts:3639-3640`）；行里的
`pass` / `fail` / `cancelled` / `tests` 四个字段在渲染时带 em-dash 兜底
（`packages/quay/src/serve-dashboard.ts:517` 与 `:529`）。本 workspace 台账当时有 10 行，
**没有一行带计数字段**：

```
round= 1 task=DIR-093  commit=41c6805686 pass=None tests=None
round= 2 task=DIR-082  commit=a4748028ec pass=None tests=None
...
round=10 task=gap-claude-session-is-subagent-unset commit=1ebdf9e8a2 pass=None tests=None
```

**Why.** `plugin/scripts/pre-verified-round-record.ts` 只有两条解析路径。内建那条是
`const TEST_COUNT_RE = /^[#ℹ]\s*(pass|fail|cancelled)\s+(\d+)/`（`:316`），只认 node:test / TAP，
因此 `go test` 的输出（`ok <pkg> 0.12s`、`--- FAIL: TestX`）永远匹配不上。声明那条读项目
`.quay/config.yml` 的 `loop.test_output`。meta-cc 设了 `loop.test_command: go test ./...`
却**没有声明 `loop.test_output`**，于是走内建路径，计数永远不会被推导出来。

**Second half.** 光声明正则还不够。`applyDeclaredTestOutput`（`:373`）用
`new RegExp(declared[f], "m").exec(plain)` 匹配——**无 `g` 标志、只取第一个匹配、从不求和**——
所以 suite 日志里必须存在**一行已经把总数算好**的文本。`go test ./...` 只打印逐包行、没有总计行，
因此 suite 入口必须自己产出这一行。

**Scope.** 这纯属可观测性：它不改变 fan-in 能否通过。值得做的理由是测试面板对本项目目前不可读——
红绿信号在，量级不在。

<!-- dedup-ref -->
现有 meta-cc 任务没有一条命名同一机制（`grep -rln 'test_output\|verification-round' tasks/` 零命中）。
声明式输出机制的权威规格在 quay 自己的任务 `gap-verification-round-bound-to-quay-shaped-suite-entry`，
其 archguard 证据块是「配置声明 → vitest 原始输出 → 台账取值」三段对照的参考样例。

## Plan

1. 新增 `scripts/quay-suite.sh`：跑 `go test -json`，把普通输出**逐字回放**（保住
   `--- FAIL: <Test>` 归因行——quay 会给「无法归因的 suite 红」设重试上限并停成 needs-human），
   末尾追加**恰好一行** `quay-suite: pass <N> fail <M> skip <K>`，并透传 go test 的退出码。
   原型已在 `./internal/version/` 上实测产出末行 `quay-suite: pass 4 fail 0 skip 0`、退出码 0。

2. ⛔ **不要命名为 `scripts/test.sh`**。`testShAt(dir)` 字面就是 `<dir>/scripts/test.sh`
   （`plugin/scripts/worker-driver.ts:1409-1411`）；它的存在会翻转 `hasTestSh`，把这一轮切到
   quay 自己的 `full-suite-runner`/TAP 路径（`suiteRunsOutsideRunner`，`:4273` 变 false）——
   那是另一套机制，不是改个名字。

3. 在 `.quay/config.yml` 的 `loop:` 下声明：

   ```yaml
     test_command: bash scripts/quay-suite.sh
     test_output:
       pass: 'quay-suite: pass (\d+) fail \d+ skip \d+'
       fail: 'quay-suite: pass \d+ fail (\d+) skip \d+'
       cancelled: 'quay-suite: pass \d+ fail \d+ skip (\d+)'
   ```

   每条正则**恰好一个捕获组**（quay 取 `m[1]`）；`tests` 故意不声明，让 quay 自己算
   `pass+fail+cancelled`（`pre-verified-round-record.ts:400`）。Go 的 `skip` 映射到
   `cancelled`，使总数等于真实测试函数数——与 vitest 参考实现同口径。

4. `.quay/config.yml` 是 **git 跟踪**文件（`.gitignore:81-84` 在 `.quay/*` 之后专门
   `!.quay/config.yml` 排除），而声明是从 **worktree** 读的（`worker-driver.ts:4314`：
   `readLoopTestOutput(o.worktree)`），所以它随 git 自动进每个任务 worktree，**不需要铺装步骤**。

5. 注意共享面：`loop.test_command` 同时是第三方 scoped 门（`resolveScopedGateCommand` →
   `bash -c "cd <worktree> && <test_command>"`，`worker-driver.ts:1497-1508`）。因此包装脚本
   必须能在**无参数**下正确工作（默认 `./...`），且**不需要** quay 的
   `--buckets/--root/--state-dir/--runner/--log-file/--run-id` 标志——`loop.test_command`
   从不接收它们，只有项目自备的 `scripts/test.sh` 才会（`plugin/skills/init/SKILL.md:107-163`）。

6. 加一个**在 `make commit` 下可跑**的回归测试。本机 **bats 未安装**（`command -v bats` 无输出），
   而 `make test-bats` 在缺 bats 时会静默跳过，所以 `tests/scripts/*.bats` 在本机不可验证；
   优选 Go 测试（`tests/` 本就是本模块的 Go 包——suite 日志里可见
   `ok github.com/yaleh/meta-cc/tests`）。测试须断言三件事：汇总行形状、红路径的退出码透传、
   失败运行的 `--- FAIL:` 行没被吞掉。

7. 对着真实产物收口：跑一轮真实 fan-in，确认 `.quay/verification-round.jsonl` 最新一行带上了
   数值型 `pass` / `tests`。

**Risk note.** 声明了但匹配不上的字段**不会退回**内建解析
（`pre-verified-round-record.ts:877`：只有声明为 `null` 时才跑内建）。所以正则写错只会让字段继续
缺席——即维持现状——绝不会写出一个假数字。

## Acceptance Criteria

- [x] `bash scripts/quay-suite.sh ./internal/version/` 退出码为 0，且其最后一行匹配
      `^quay-suite: pass [0-9]+ fail [0-9]+ skip [0-9]+$`。
- [x] 退出码透传：对一个含故意失败测试的包运行包装脚本，退出码非零，且输出中仍含
      `--- FAIL: <TestName>` 行。
- [x] 正则语义：对每条声明的正则，按 quay 的原样调用
      `new RegExp(re, "m").exec(汇总行)`（`pre-verified-round-record.ts:391`）恰好取到一个捕获组，
      再套 quay 的兜底公式后 `pass` / `fail` / `cancelled` / `tests` 四项**全部**得到数值。
- [x] 回归测试在 `make commit` 下通过（不依赖 bats）。
- [x] 跑过一轮真实 fan-in 后，`.quay/verification-round.jsonl` 最新一行的 `pass` 与 `tests`
      为数值（非缺席）。
- [x] `make commit` 通过。

## DoD

- 一个真实 meta-cc 任务的 fan-in 轮次落下一行**带数值计数**的台账记录，且 quay web 测试卡片
  在目前渲染 `—/—` 的位置渲染出数字——由读取产物与页面确认，不是只由 fixture 推断。
- 红路径是被证明的而非被假设的：失败运行仍然非零退出、仍然点名失败的测试，所以新入口不可能把
  红的 suite 变成绿的。
- `make commit` 通过，改动已落在 main 上。

## Touches

- `tasks/gap-quay-test-output-undeclared-counts-absent.md`
- `.quay/config.yml`
- `scripts/quay-suite.sh`
- `tests/quay_suite_test.go`

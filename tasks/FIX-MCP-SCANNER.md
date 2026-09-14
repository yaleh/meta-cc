---
id: FIX-MCP-SCANNER
title: AC-258 user-scope 驱动取证任务
status: ready
needs_human_cause: human-adjudication
labels: []
parent: null
children: []
extra: {}
---
## Proposal

`cmd/mcp-server/main.go` reads JSON-RPC requests off stdin with a raw
`bufio.NewScanner(os.Stdin)`. `bufio.Scanner` has a hard default maximum token
size of **64 KiB**. A single request line longer than that makes `Scan()` return
false with `bufio.ErrTooLong`; the `for scanner.Scan()` loop then exits, the
`scanner.Err()` check reports it, and `main()` returns — so **one oversized
request terminates the whole server** instead of being rejected as a bad
request. A long tool result echoed back in a JSON-RPC frame, or a large
`tools/call` argument blob, is enough to reach that size.

Measured on this checkout 2026-09-14: the Makefile's `check-no-scanner` target
exists precisely to prevent raw `bufio.NewScanner` on line-oriented parsing
paths, and it prints

    Checking for raw bufio.NewScanner on line-oriented parsing paths...
    OK: No raw bufio.NewScanner found.

while `cmd/mcp-server/main.go` is the one production file in the repository that
still calls it. The target's own grep is written as

    grep -rn "bufio\.NewScanner" internal/ cmd/mcp-server/ --include="*.go" | grep -v "main\.go" ...

i.e. it **exempts the very file that carries the defect**, so the gate is green
and the defect class it names is live in the one place it is not looking. The
target's own failure message names the sanctioned replacement:
`parser.ReadLineBounded` for line-length-sensitive non-JSONL reads.

## Plan

1. In `cmd/mcp-server/main.go`, replace the `bufio.NewScanner(os.Stdin)` loop
   with a `bufio.Reader` driven by `parser.ReadLineBounded(r, maxLineBytes)`.
   Lines within the bound keep the existing behaviour exactly: same
   `json.Unmarshal` into `JSONRPCRequest`, same `-32700` reply + `continue` on a
   parse error, same `handleRequest` dispatch.
2. Declare the bound as a named constant with a comment recording why an
   explicit bound (rather than the scanner's implicit 64 KiB default) is used.
3. Treat `parser.ErrLineTooLong` as a **per-request** error: reply with a
   JSON-RPC error for that frame and keep serving. A single oversized request
   must not terminate the server, which is the defect being fixed.
4. Keep the terminal-condition handling (`io.EOF` and a genuine read error)
   reported through the existing `slog.Error("scanner error", ...)` path so
   shutdown stays observable.
5. Remove the `grep -v "main\.go"` exemption from the Makefile's
   `check-no-scanner` target so the target covers `cmd/mcp-server/main.go`.
6. Add a test in `cmd/mcp-server/main_test.go` that feeds a single line larger
   than 64 KiB and asserts the server does **not** terminate on it.

## Touches

- tasks/FIX-MCP-SCANNER.md
- cmd/mcp-server/main.go
- cmd/mcp-server/main_test.go
- Makefile

## Acceptance Criteria

- [x] `make check-no-scanner` passes **with the `main.go` exemption removed**
      from the target, i.e. the target's grep now covers `cmd/mcp-server/main.go`
      and still reports no raw `bufio.NewScanner` there.
- [x] A test feeds one JSON-RPC line longer than 64 KiB and asserts the server
      stays alive (it does not exit and does not drop into the EOF path).
- [x] A line within the bound still produces the same request handling as before
      (a normal `initialize` frame is answered).
- [x] `make commit` passes.

## DoD

- [x] `make commit` green
- [ ] the change landed on `main`（待外部）

## Needs-Human

**执行 2026-09-14T15:14:19.135Z — 连续修满重试上限仍不合格（标 needs-human）**

- 阻碍原因：worker-driver 连续 3 次 <60000ms 快速死亡（退避上限）
- 成因类：human-adjudication

## Evidence

Worktree `FIX-MCP-SCANNER` on branch `task/FIX-MCP-SCANNER`; implementation commit
`fix(mcp-server): stop raw bufio.Scanner from killing the server on one big frame`.

- `make check-no-scanner` → exit 0, `OK: No raw bufio.NewScanner found.` with the
  `main.go` exemption deleted from the target: `git show HEAD~:cmd/mcp-server/main.go`
  still contains `scanner := bufio.NewScanner(os.Stdin)` (line 74), i.e. the
  grep now covers the file that carried the defect.
- `go test ./cmd/mcp-server/ -run TestServeRequests -v` → 5/5 PASS:
  a >64 KiB frame followed by a normal frame (both answered, `serveRequests`
  returns nil, no -32603 input-error reply); a frame past the cap answered with
  -32600 with the next frame still served; an over-long line whose read stopped
  mid-line (tail does not swallow the next frame); ordinary `initialize` +
  non-JSON + `initialize` (only the non-JSON line gets -32700); a final
  newline-less frame.
- `go test ./...` (the scoped gate, after `git merge develop`) → green.
- `make commit` → exit 0.

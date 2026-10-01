---
id: gap-codex-appserver-sessions-missing-rollout-path
title: get_session_directory{provider:codex} hard-fails with missing
  rollout_path on physical paths (app-server backend never populates it)
status: ready
labels:
  - gap
  - defect
  - mcp
  - codex
  - provider
parent: null
children: []
extra:
  schema: execution
  scope:
    owner_repo: meta-cc
  acceptance: make commit
---
## Proposal

**Defect.** `get_session_directory{provider:"codex"}` fails outright on any
project that has Codex data, using a **physical** path and with no symlink
anywhere in the picture:

```
failed to resolve rollout path for codex session 01a0d3ec-517d-72c3-8976-14b980447d79:
  missing rollout_path for session 01a0d3ec-517d-72c3-8976-14b980447d79
```

Reproduced on 5 projects across two machines (2 here, 3 reported independently).
The discovery tool for Codex therefore returns nothing usable at all — not an
empty list, a hard error — even though the underlying rollout files exist.

**Root cause: a producer/consumer contract satisfied by two of the three Codex
backends and not by the third.**

The consumer assumes `rollout_path` is always present:

- `internal/provider/codex/rollout.go:38-45` — `RolloutPath()` reads
  `session.Extensions["rollout_path"]` and returns
  `fmt.Errorf("missing rollout_path for session %s", session.ID)` when empty.
  There is no fallback.

Two producers do populate it:

- `internal/provider/codex/raw_discovery.go:91` — sets it from the rollout file
  path during the filesystem walk (the `files` fallback).
- `internal/provider/codex/sqlite.go:232` — sets it from the `threads.rollout_path`
  column.

The third does not:

- `internal/provider/codex/appserver_provider.go:185`
  (`appServerBackend.listSessions`) — contains **no reference to `Extensions` or
  `rollout` at all**, so the sessions it builds carry no `rollout_path`.

And `dispatch` prefers that third one:

- `internal/provider/codex/provider.go:144` —
  `ListSessions` = `dispatch(p, ctx, p.appServerListSessions, p.filesListSessions)`.
  `rawfiles.SelectCodexFiles` (`internal/provider/rawfiles/rawfiles.go:83`) reaches
  it through `p.ListSessions(ctx)`, which is the path
  `buildCodexDirectoryResult` / `buildCodexMetadataResult` take.

By elimination the app-server backend is the only one that can produce this
error: both files-based producers set the field (and the DB column is in fact
fully populated — see Evidence).

## Evidence

**The data is present; this is not a missing-data or indexing problem.** Read
read-only from `~/.codex/state_5.sqlite` (the highest-numbered `state_N.sqlite`,
per `internal/locator/codex.go:43`):

```
threads 列 (40), 第 2 列即 rollout_path           -> 存在
total rows 27, rows with empty rollout_path       -> 0
01a0d3ec-517d-72c3-8976-14b980447d79
    rollout_path = /data/home/yale/.codex/sessions/2026/09/24/rollout-2026-09-24T22-57-55-01a0d3ec-....jsonl
01a0d3c6-85db-7480-92d5-515469a6fed5
    rollout_path = /data/home/yale/.codex/sessions/2026/09/24/rollout-2026-09-24T22-16-38-01a0d3c6-....jsonl
```

Both named sessions' rollout files exist on disk, and the DB values point at
them. `~/.codex/session_index.jsonl` (17 entries, fields
`{id, thread_name, updated_at}`) is a **different, unrelated artifact** and is
not the sqlite source — do not use it as evidence either way.

Re-run with python3's stdlib `sqlite3` module. This host has **no `sqlite3`
CLI**, so a `sqlite3 ...` shell command will not run here:

```python
import sqlite3
con = sqlite3.connect("file:" + __import__("os").path.expanduser("~/.codex/state_5.sqlite") + "?mode=ro", uri=True)
cols = [r[1] for r in con.execute("PRAGMA table_info(threads)")]
rows = con.execute("SELECT id, rollout_path FROM threads").fetchall()
print("rollout_path is column #%d" % (cols.index("rollout_path") + 1))
print("rows=%d empty=%d" % (len(rows), sum(1 for r in rows if not r[1])))
```

Behavioural reproduction, physical paths:

```
get_session_directory{provider:"codex", working_dir:"/data/home/yale/work/quay"}
  -> failed to resolve rollout path for codex session 01a0d3ec-...: missing rollout_path for session 01a0d3ec-...
get_session_directory{provider:"codex", working_dir:"/data/home/yale/work/claudecodeui"}
  -> failed to resolve rollout path for codex session 01a0d3c6-...: missing rollout_path for session 01a0d3c6-...
```

## Secondary observation (NOT the cause here)

`internal/provider/codex/sqlite.go:217` guards an optional column with a
presence check (`_, hasParentCol := colMap["parent_thread_id"]`), but `:232`
reads `colMap["rollout_path"]` with no equivalent guard, and the `json.Marshal`
error beside it is discarded. On this host the column exists and is populated,
so this is not the active cause — it is a latent hazard if a `state_N.sqlite`
schema ever omits the column, and it is worth handling while in the area.

## Plan

The fix direction is deliberately not prescribed — resolve it once the app-server
backend's capabilities are known:

1. Preferred if the app-server knows it: populate `Extensions["rollout_path"]` in
   `appServerBackend.listSessions` so all three backends satisfy the same
   contract.
2. Otherwise: give `RolloutPath()` a defined degradation (derive from the session
   ID + rollout root, or return a typed "unavailable" the callers can render)
   instead of a hard error that takes the whole discovery call down.
3. Whatever is chosen, state the contract once and assert it for **every**
   backend, so this cannot reappear when a fourth one is added.

## Acceptance Criteria

- [x] On a project that has Codex sessions, `get_session_directory{provider:"codex"}`
      returns a file list rather than an error.
- [x] The failure is not reachable while the session's rollout file exists on
      disk and the DB column is populated — i.e. the fix does not depend on
      re-deriving data that is already present.
- [x] The contract is asserted per backend (app-server, files/sqlite,
      files/rollout-fallback), so a backend that fails to populate
      `rollout_path` fails a test rather than silently degrading at the
      consumer.
- [x] Regression test uses a project fixture with Codex data, not this host's
      `~/.codex` (which may be absent or shaped differently elsewhere).

## DoD

- make commit passes
- changes landed on main
- the per-backend contract assertion above is a committed test

## Touches

- tasks/gap-codex-appserver-sessions-missing-rollout-path.md
- internal/provider/codex/appserver_provider.go
- internal/provider/codex/rollout.go
- internal/provider/codex/rollout_contract_test.go (new per-backend contract test)

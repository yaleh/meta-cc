# Codex App-Server History Backend (DIR-029)

> **See also**: `docs/reference/codex-history-model.md` (DIR-032) extends
> this backend with per-page failure tolerance, a `Provider.ListSessionsPage`
> cursor-based continuation API, capability-negotiated (and currently
> unsupported) experimental turn/item pagination, and archive/lineage
> semantics.

meta-cc can read Codex conversation history two ways:

- **files**: parse the existing SQLite thread index (`state_5.sqlite`) and
  the Codex JSONL rollout files directly. This is meta-cc's original
  behavior and remains fully supported — it works offline, works with any
  installed Codex CLI version, and is the only option for forensic access
  to raw session state.
- **app_server**: talk to `codex app-server`, the JSON-RPC surface OpenAI
  documents as the stable integration point for conversation history
  (unlike the SQLite/rollout formats, which are explicitly *not* a stable
  interface).

By default (`auto` mode) meta-cc prefers `app_server` and falls back to
`files` whenever the app-server is absent, times out, returns an
incompatible/unparseable version, or otherwise fails.

## Configuration

| Env var | Values | Default | Effect |
|---|---|---|---|
| `META_CC_CODEX_BACKEND` | `auto`, `app_server`, `files` | `auto` | Selects the backend mode. An invalid value falls back to `auto` with a warning (see `Provider.Warnings()`). |
| `META_CC_CODEX_APP_SERVER_BIN` | path or executable name | `codex` | Overrides the executable used to spawn `codex app-server`, for environments where it isn't on `PATH` as `codex`. |
| `META_CC_CODEX_ROOT` / `CODEX_HOME` | directory | `~/.codex` | Existing Codex-home override (see `internal/locator.CodexLocator`). Also used to pin the spawned app-server child's own `CODEX_HOME`, so the external process and meta-cc's files backend agree on which state they're reading. |

### Mode semantics

- **`files`**: identical to pre-DIR-029 behavior. The app-server backend is
  never invoked.
- **`app_server`**: only the app-server backend is used. A failure (process
  won't start, handshake fails, a call errors) is returned to the caller as
  a clear error — it never silently substitutes files-backend results.
- **`auto`**: tries `app_server` first (unless the circuit breaker — see
  below — is open), falling back to `files` on any failure. Which backend
  actually answered a given call is available via `Provider.Backend()`
  (`"app_server"` or `"files"`); accumulated failure/fallback diagnostics
  are available via `Provider.Warnings()`.

### Circuit breaker

In `auto` mode, after 3 consecutive app-server failures, meta-cc stops
attempting the app-server backend for 60 seconds and goes straight to
files — so a persistently broken/absent app-server doesn't add its full
startup-timeout latency to every single call. The breaker resets on the
next successful app-server call after cooldown.

### Timeouts

- Spawning the app-server process and completing the `initialize` handshake
  is bounded to 10 seconds.
- Each `thread/list`/`thread/read` exchange (including however many
  paginated requests one logical operation needs) is bounded to 20 seconds.

Both bounds are enforced via `context.Context`, and `Process.Close` always
terminates the child process group (SIGTERM, escalating to SIGKILL after a
2-second grace period) and waits for it to be reaped — no child app-server
process survives a call timeout, cancellation, or normal shutdown.

## Supported methods (first release)

Only two read-only, stable methods are used, matching the Contract's "only
`thread/list` and `thread/read` are required" scope:

- `initialize` — mandatory handshake, first call on every connection.
- `thread/list` — enumerate threads (sessions).
- `thread/read` (with `includeTurns: true`) — fetch one thread's full
  turn/item history.

meta-cc never issues `thread/start`, `thread/resume`, `thread/fork`,
`thread/archive`, `thread/delete`, `thread/rollback`, approval responses,
or any other mutating/experimental method. No query resumes, mutates,
archives, repairs, or deletes a Codex thread.

## Listing semantics: avoiding the default-filter pitfall

An audit against a real Codex CLI 0.145.0 app-server (via
`codex app-server generate-json-schema`, which emits the full protocol
schema, plus a live `initialize`/`thread/list` handshake against a scratch
`CODEX_HOME`) confirmed two default-filter behaviors that would otherwise
silently omit eligible sessions:

1. **`sourceKinds`**: an omitted or empty filter defaults to *interactive
   sources only*. meta-cc's `app_server` backend always passes the full
   enum explicitly:
   `cli`, `vscode`, `exec`, `appServer`, `subAgent`, `subAgentReview`,
   `subAgentCompact`, `subAgentThreadSpawn`, `subAgentOther`, `unknown`.
2. **`archived`**: this is an *exclusive* boolean filter, not a tri-state.
   Omitted or `false` returns only non-archived threads; `true` returns
   only archived threads — there is no single request that returns both.
   meta-cc's default `ListSessions`/`GetSession` therefore issues **two**
   independently paginated `thread/list` passes (non-archived, then
   archived) and merges the results, rather than defaulting to
   non-archived-only.
3. **`modelProviders`**: passed as an explicit empty array (`[]`, not
   omitted/`null`), which the server documents as "includes all
   providers" — this is required to avoid an unintended provider-based
   narrowing.

Pagination (`cursor`/`nextCursor`) is followed to completion within each of
the two archived-state passes.

## Canonical model mapping

`thread/list` and `thread/read` responses map into the DIR-028
`conversation.Session` → `Turn` → `Item` model (see
`internal/provider/codex/appserver/map.go`), not a parallel/flattened
representation. `Turn.UserText`/`AssistantText`/`ToolCalls` are still
populated as a backward-compatible projection derived from `Items`, mirroring
the existing rollout adapter's projection so both Codex backends (`files`
and `app_server`) populate those compatibility fields identically.

Item-type coverage (confirmed against the live protocol schema): `userMessage`,
`agentMessage`, `commandExecution`, `fileChange`, `mcpToolCall`, `webSearch`,
`reasoning`, `plan`, and `contextCompaction` get dedicated field mapping.
`contextCompaction` additionally decodes optional `reason`/`summary` fields
into a typed `conversation.CompactionBoundary` (DIR-032, best-effort — not
yet empirically confirmed against a live payload).
Everything else (`hookPrompt`, `dynamicToolCall`, `collabAgentToolCall`,
`subAgentActivity`, `imageView`, `sleep`, `imageGeneration`,
`enteredReviewMode`, `exitedReviewMode`) is preserved losslessly via
`conversation.NewRawItem` rather than dropped, so future work can add
dedicated mapping without a schema migration.

## Version gating

The app-server backend targets Codex CLI ≥ `0.145.0`
(`appserver.MinSupportedVersion`), the version this integration was
verified against. `appserver.DetectCLIVersion` runs a bounded
`<command> --version` and reports found/supported/error rather than
failing hard, so `auto` mode and tests can decide to skip/fall back instead
of erroring.

## Testing

- **Protocol/unit tests** (`internal/provider/codex/appserver/client_test.go`)
  use an in-process fake app-server speaking the exact same
  newline-delimited JSON-RPC framing over `io.Pipe` — no subprocess, no
  real Codex installation required, unconditional in `make commit`.
- **Orchestration tests**
  (`internal/provider/codex/appserver_provider_test.go`,
  `internal/provider/codex/provider_dispatch_test.go`) cover mode
  dispatch, `auto` fallback, provenance/warnings, and the circuit breaker
  against a hand-rolled fake `threadSource` — also unconditional.
- **Real-CLI E2E test**
  (`internal/provider/codex/appserver/e2e_test.go`,
  `TestRealAppServerE2E`) spawns an actual installed `codex app-server`
  child, pinned to an isolated `t.TempDir()` `CODEX_HOME`, and drives
  `initialize` + `thread/list`. It is version-gated: `t.Skip` (not a
  failure) when `codex` isn't on `PATH`, its version can't be parsed, or it
  doesn't satisfy `MinSupportedVersion`. This test never reads or writes a
  developer's real `~/.codex`.

## Known limitations / explicitly out of scope

- Every `ListSessions`/`GetSession`/`LoadTurns` call in `app_server` mode
  currently spawns a fresh app-server process and performs a fresh
  handshake rather than reusing a long-lived connection across calls. This
  favors correctness/simplicity (trivial "no leaked process" guarantees)
  over latency; a persistent-connection pool is a reasonable follow-up if
  per-call spawn overhead proves significant in practice.
- `sourceKinds` is validated against the CLI 0.145.0 schema empirically,
  not against a versioned/published enum contract — a future Codex release
  could add or rename source kinds. `DetectCLIVersion`'s version gate
  bounds this risk to versions this integration has actually been checked
  against.
- The app-server's `daemon`/`proxy` subcommands (a persistent background
  daemon reached over a Unix socket, for multiple concurrent clients) are
  not used. meta-cc always spawns `codex app-server` directly with the
  default `--listen stdio://` transport, which is sufficient for a
  request/response CLI-style client.

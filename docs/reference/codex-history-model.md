# Codex History Completeness, Compaction, Lineage & Pagination (DIR-032)

This document extends the DIR-028 canonical `Session` → `Turn` → `Item`
model (see `docs/reference/jsonl-schema.md`'s "Canonical Item Model"
section) and the DIR-029 app-server backend (see
`docs/reference/codex-app-server.md`) with the completeness, compaction,
archive, lineage, and pagination concepts Codex history has grown to need:
paginated/projected app-server responses, context-compaction boundaries,
archived/subagent threads, and spawn lineage. It does not replace either of
those documents — read them first.

## Discovery roots and database compatibility (DIR-069)

All Codex backends use one canonical home, resolved with this precedence:

| Priority | Source | Notes |
|---|---|---|
| 1 | `META_CC_CODEX_ROOT` | meta-cc-specific override; wins when set |
| 2 | `CODEX_HOME` | Codex CLI's own home override |
| 3 | `~/.codex` | User default |

This single resolved root drives every Codex storage path meta-cc reads —
there is no per-surface override:

| Consumer | Path derived from the canonical root |
|---|---|
| Transcript/session discovery | `<root>/sessions/` (active) and `<root>/archived_sessions/` |
| SQLite thread index | highest-compatible `<root>/state_N.sqlite` (see below) |
| Rollout-only fallback | `*.jsonl` under the two session trees above |
| App-server child process | spawned with `CODEX_HOME=<root>` so it reads the same state |

`<root>/history.jsonl` is intentionally never used. The one related override
is backend selection: `META_CC_CODEX_BACKEND` (`auto`/`app_server`/`files`,
default `auto`) picks between the app-server and files backends — see
`docs/reference/codex-app-server.md` — but both backends honor the same root.

The files backend enumerates `state_N.sqlite` files by numeric version,
newest first, and selects the first database whose `threads` table contains
the required metadata columns. An incompatible newer database is reported in
`Provider.Warnings()` and the next compatible version is tried. If no usable
database exists, active `sessions/` and `archived_sessions/` rollouts remain
queryable from their `session_meta` records. This fallback requires an
explicit recorded `cwd` for project filtering; files without enforceable
`id` and `cwd` metadata are skipped with a warning rather than being allowed
to leak across project boundaries.

## History completeness

`conversation.Turn.Completeness` (`conversation.HistoryCompleteness`, in
`internal/conversation/item.go`) makes "this turn is fully loaded content"
distinguishable from "this is a placeholder":

| Value | Meaning |
|---|---|
| `""` (unspecified) | Backend doesn't distinguish states but has always returned complete records. `IsFull()` treats this the same as `full`. |
| `full` | Fully materialized content. |
| `summary` | A server-provided summary/preview standing in for the full record. |
| `unloaded` | A placeholder position is known, but no content — not even a summary — was fetched. |
| `truncated` | Loading stopped at (or after) this turn (e.g. a rollout's `maxLines` cap) — turns/content after this point are missing, not merely absent from this query. |
| `unavailable` | The turn is known to exist but its content could not be loaded (e.g. a page fetch failed). |

`HistoryCompleteness.IsFull()` is the single choke point for "can this be
treated as complete" — only `full` and the zero value return `true`.

**Where this is set today:**

- `internal/provider/codex/rollout.go`: every turn built from a legacy or
  dot-schema rollout is marked `Full` on flush, **except** the final
  (possibly partial) turn when `loadTurnsFromRollout`'s `maxLines` cap was
  hit — that one is marked `Truncated` instead, so a caller knows content
  after it is missing.
- `internal/provider/codex/appserver/map.go`'s `mapTurn`: every turn from
  `thread/read(includeTurns)` is marked `Full` — this is the only
  app-server turn-content surface DIR-029/032 currently use, and it always
  returns full item content. No confirmed app-server surface returns
  `summary`/`unloaded` turns yet (see "Experimental pagination" below); the
  field exists so a future confirmed partial-content surface has somewhere
  to report it, and so downstream consumers (see `internal/ftsindex`) are
  already correct once one exists.

**Enforcement:** `internal/ftsindex/indexer.go`'s `reindexSessionTx` calls
`turnCompletenessIndexable` and skips every item belonging to a
`summary`/`unloaded`/`unavailable` turn — a placeholder can never enter the
full-content search index and be returned as if it were a real content hit
(see `TestRefresh_SummaryAndUnloadedTurnsExcludedFromIndex`).

## Compaction boundaries

`conversation.CompactionBoundary` (`Item.Compaction`, populated only on
`ItemKindCompaction` items) records *that* and *why* preceding context was
replaced/summarized, without re-embedding the superseded content:

```go
type CompactionBoundary struct {
    Reason  string // e.g. "context_window"
    Summary string // human-readable replacement/summary text, when reported
}
```

The original pre-compaction `Item`s remain in the ordered stream, in their
original position — this is the historical record. `CompactionBoundary`
never gets folded into `Turn.UserText`/`AssistantText`, so a content query
cannot end up presenting the same logical exchange twice (once as itself,
once disguised as part of a "replacement" summary).

Codex 0.145 rollouts report a compaction boundary through two independent
event families for the same logical event:

- Top-level `"compacted"` (`{"turn_id", "reason"}`) → typed
  `ItemKindCompaction` item with `Compaction.Reason` set.
- `event_msg` `"context_compacted"` (`{"summary"}`) → a second typed
  `ItemKindCompaction` item with `Compaction.Summary` set.

Both are parsed in `internal/provider/codex/rollout.go`'s `applyLegacy`
(`appendCompactionItem`). See
`TestLoadTurnsFromRolloutCompactionDoesNotDuplicateContent` in
`rollout_test.go` and the fixture
`tests/fixtures/codex/rollout-legacy-compaction-boundary-sample.jsonl` for
the end-to-end no-duplication proof.

The app-server's `contextCompaction` item type (`internal/provider/codex/appserver/map.go`)
is mapped the same way, best-effort: `reason`/`summary` fields are decoded
if present, but this has not been empirically confirmed against a live
payload (DIR-029 only confirmed the item *type* exists, not its full field
shape) — an absent field simply yields an empty `CompactionBoundary` rather
than falling back to `ItemKindUnknown`.

## Turn/session lifecycle status

`conversation.TurnStatus` gained `TurnStatusAborted` (DIR-032) for Codex's
`turn_aborted` event (a user interrupt or similar — distinct from
`TurnStatusFailed`, an error). `ItemKindSessionEnd` is a new item kind for
Codex's `session_end` event, carrying the reported reason in `Item.Text`.

Both are typed in `internal/provider/codex/rollout.go`'s `applyLegacy`
rather than left as raw `Extensions.codex_events` passthrough — see
`TestLoadTurnsFromRollout0145TypedEventFamilies`.

**A turn can be lifecycle-only (DIR-050):** `turnBuilder.flush()`'s
retention condition originally only kept `b.current` when at least one of
`UserText`/`AssistantText`/`ToolCalls`/`TokenUsage`/`Extensions` was
populated. `ItemKindSessionEnd` and `ItemKindCompaction` items — and
`TurnStatusAborted`, which is set with no accompanying `Item` at all — are
invisible to all five of those checks, so a turn whose *only* content
before EOF was one of these lifecycle signals (e.g. a rollout that opens a
turn with `task_started` and then immediately hits `session_end` or
`turn_aborted`, with no message/tool/usage activity in between) was
silently dropped: `flush()` would produce zero turns instead of one. This
means "a turn" is no longer necessarily a turn with visible
message/tool/usage content — an otherwise-empty turn can be returned
solely to carry a lifecycle signal (a session ending or a turn being
aborted with nothing else having happened in it).

`flush()` now also retains `b.current` whenever its `Items` slice is
non-empty (covering `ItemKindSessionEnd`/`ItemKindCompaction`, and any
future item kind, without a kind-by-kind carve-out) or its `Status` is set
(covering `TurnStatusAborted`, which appends no `Item`). See
`TestLoadTurnsFromRolloutLifecycleOnlySessionEnd`,
`TestLoadTurnsFromRolloutLifecycleOnlyTurnAborted`, and
`TestLoadTurnsFromRolloutLifecycleOnlyCompaction` in `rollout_test.go`.

**Observable end-to-end through MCP tools (DIR-053):** DIR-050's fix stopped
at the raw parser layer — `flush()` retained the lifecycle-only turn in the
`[]conversation.Turn` slice, but `providerrecords.Normalize`
(`internal/provider/records/records.go`) — the only layer downstream of
`loadTurnsFromRollout` that turns a `conversation.Turn` into the query-time
records MCP tools actually operate on — read only
`UserText`/`AssistantText`/`ToolCalls`/`TokenUsage`. A lifecycle-only turn
(all four empty) therefore still produced **zero** normalized records and
stayed invisible to every MCP tool, so "a session looks like it has no
history" was still reproducible via an actual tool call. `Normalize` now
emits a minimal **lifecycle record** for a turn that carries a lifecycle
signal but would otherwise produce no record, mirroring DIR-050's widening
at the normalized-record layer:

- an `ItemKindSessionEnd` item → one record with `"type": "session_end"` and
  the reported `reason` (from `Item.Text`);
- an `ItemKindCompaction` item → one record with `"type": "compaction"` and
  its `CompactionBoundary` (`reason`/`summary`) under a `compaction` field;
- an explicit terminal/failure `Turn.Status` with no such item (e.g.
  `TurnStatusAborted`) → one record whose `type` is derived from the status
  (`"turn_aborted"`, `"turn_failed"`, or `"turn_completed"`) and which carries
  `turn_status`. `TurnStatusInProgress` alone is parser bookkeeping, not a
  lifecycle event, and emits no record.

Each lifecycle record reuses the existing record vocabulary — the standard
identity fields (`provider`/`session_id`/`sessionId`/`cwd`/`timestamp`) plus
the canonical `turn_id`/`turn_index`/`seq` (DIR-036) — so it is
distinguishable by a jq filter such as `select(.type == "session_end")` and
is surfaced by `query_session_content(role="all", provider="codex")` (whose
conversation-flow filter now includes the lifecycle types) as well as any
signal query that matches on a timestamp. Turns that already surface
user/assistant/tool content are left untouched — no extra record is added —
so ordinary sessions that merely *end* with a `session_end` event are
unchanged. See `internal/provider/records/records_lifecycle_test.go`
(Normalize layer) and
`internal/mcp/executor/codex_lifecycle_e2e_test.go` (end-to-end
`query_session_content(role="all")` against a rollout ending in a bare
`session_end`).

**Live/ongoing state:** there is no separate "session is live" field.
A turn with `TurnStatusUnspecified` and no `ItemKindSessionEnd` item
anywhere in its session is the existing, correct signal for "no terminal
status was observed" — which covers both "still ongoing" and "the rollout
stream simply ended without a terminal marker" (e.g. a crashed CLI). This
matches DIR-027's existing precedent: a status is set when the source event
positively confirms it, never guessed.

**Remaining 0.145 families promoted (DIR-072):** `world_state`,
`tool_search_call`/`tool_search_output`, and `thread_settings_applied` are
no longer raw passthrough — each maps to a dedicated typed Item kind
(`world_state`, `tool_search_call`, `tool_search_output`,
`settings_applied`) in `applyLegacy`, following the DIR-028 "Extension
rule — adding a new Item kind" checklist:

- **Tool search call/output** correlate via `ToolCallID` (the call sets
  `ID` == `ToolCallID`, the output joins on the same key) and retain
  status/error information (`IsError`, `Status` — a failure status/error
  string maps to `StatusFailed` and the error text surfaces as `Output`
  when no regular output is present). They are a *distinct* family from
  `tool_call`/`tool_result`: `projectLegacyFields` (both the rollout and
  app-server copies) ignores them, so they never enter `Turn.ToolCalls`,
  normalized `tool_use` records, or tool stats — search activity is never
  conflated with shell/function tools, and never duplicated by those
  projections. They DO participate in full-text indexing (call:
  `"<name> <input>"`, output: its output) like tool call/result.
- **World-state** exposes only bounded metadata (`conversation.WorldState`:
  `cwd`, `approval_policy`, `sandbox_policy`, each capped at 512 bytes via
  `NewWorldState`); any other/larger source field is omitted by
  construction.
- **Settings** expose only the applied setting KEYS
  (`conversation.SettingsApplied`: sorted, capped at 32 keys with overflow
  flagged via `Truncated`). Setting VALUES are never embedded — they can
  carry credentials (env keys, tokens) — and are likewise never indexed by
  ftsindex. Full world-state/settings payloads remain available only
  through explicit raw access to the source rollout (Stage-2 query).

Like the DIR-032 kinds, these items ride in the turn's `Items` stream in
encounter order but add nothing to the legacy record projection, so turns
that already surface content keep identical normalized records (see
`TestNormalizeDoesNotDuplicateTypedSearchAndMetadataItems`). A turn whose
ONLY content is one of these metadata items currently produces no
normalized record — consistent with the existing `reasoning`-only
precedent; the typed Item remains available to future item-level query
surfaces. Genuinely unrecognized future events still round-trip as capped
`unknown` items plus `Extensions.codex_events` entries — see
`TestLoadTurnsFromRolloutUnknownFutureEventsRoundTripAsRaw`.

## Archive filtering defaults

`query_sessions`'s `archived` dimension now defaults to **excluding**
archived sessions when neither `archived` nor `status` is explicitly
passed — previously an omitted filter meant "no constraint" (both active
and archived sessions returned), which contradicted the "archived sessions
are discoverable only when requested" contract. See
`buildSessionFilterFromArgs` in
`internal/mcp/executor/query_sessions_handler.go` and
`TestQuerySessions_DefaultExcludesArchived`.

## Lineage

`conversation.LineageStatus` (`Session.Lineage`) makes "confirmed no
parent" distinguishable from "spawn metadata wasn't available":

| Value | Meaning |
|---|---|
| `""` (unspecified) | Adapter hasn't been updated to populate this. |
| `root` | Provider positively confirmed no parent. |
| `child` | `ParentThreadID` is populated from a source that reliably reports it. |
| `unknown` | Spawn metadata was not available (e.g. an older threads-table schema with no `parent_thread_id` column, or a subagent source kind whose spawn edge was suppressed) — must NOT be presented as a root. |

Populated by:

- `internal/provider/codex/sqlite.go`'s `scanSession`: no `parent_thread_id`
  column at all → `unknown` (the schema can't report lineage); column
  present + empty + subagent source kind → `unknown` (spawn metadata wasn't
  recorded even though the column exists); column present + empty +
  non-subagent → `root`; non-empty → `child`.
- `internal/provider/codex/appserver/map.go`'s `mapLineage`: non-empty
  `parentThreadId` → `child`; empty + subagent source kind → `unknown`
  (subagent threads are spawned by construction, so a missing parent edge
  means the app-server suppressed/omitted it); empty + non-subagent →
  `root`.

**Querying lineage:** `query_sessions`'s existing `parent_thread_id` filter
already answers "children of thread X" (DIR-030). DIR-032 adds
`ancestors_of=<thread_id>` (Codex only) to walk `ParentThreadID` upward
(immediate parent, then grandparent, ...) — see `handleAncestorsOf` in
`internal/mcp/executor/query_sessions_handler.go`. Each returned entry
carries `lineage`; traversal stops (with a warning, never a silent
truncation) on: a confirmed root, an explicit `unknown`, a project/cwd
boundary crossing, a cycle, or the `maxLineageDepth` (32) bound. A
project/cwd boundary check on every hop is required here (this is a NEW
session-lookup path) per the DIR-030 precedent adversarial-audit finding —
see `TestQuerySessions_AncestorsOf_StopsAtProjectBoundary`.

## Pagination

Codex app-server's `thread/list` cursor pagination
(`cursor`/`nextCursor`/`backwardsCursor`) is a **confirmed, stable surface**
DIR-029 already follows to completion internally
(`appServerBackend.listAll`). DIR-032 adds two things on top of it:

1. **Per-page failure tolerance.** `listAll` used to abort the entire
   listing (both archived-state passes) on any page error. Now: a failure
   on the very first page of a pass (zero threads fetched) still aborts
   that pass and propagates as an error — preserving `ModeAuto`'s existing
   circuit-breaker/fallback behavior for a genuinely broken app-server. A
   failure **after** at least one page already succeeded instead returns
   the threads fetched so far plus a warning naming the exact cursor to
   resume from (`appServerBackend.recordWarning`, drained via
   `Provider.Warnings()`) — a mid-pagination blip degrades gracefully
   instead of discarding already-fetched results. See
   `TestAppServerBackendListAllToleratesPageFailureMidPagination`.
2. **A cursor-based continuation API for callers**,
   `codex.Provider.ListSessionsPage(ctx, filter, cursor) (SessionPage, error)`:
   fetches exactly one real `thread/list` page when app-server is reachable
   (respecting `Mode`/the circuit breaker exactly like every other
   `Provider` method), returning the server's own `nextCursor`. The files
   backend has no server-side pagination surface at all, so whenever
   app-server isn't in play, this returns the full filtered result as a
   single page with an empty `NextCursor` — "no more pages" — rather than a
   partial/broken cursor contract. See
   `TestProviderListSessionsPageCapabilityPresent` /
   `TestProviderListSessionsPageCapabilityAbsentFallsBackSafely` for the
   capability-present/-absent proof.

   **DIR-034: this now has a real caller.**
   `codex.Provider.FetchSessionsBounded(ctx, filter, limit)`
   (`internal/provider/codex/provider.go`) wraps `ListSessionsPage` in a
   loop that stops once `limit` matching sessions have been collected or a
   page returns an empty `NextCursor` — it is the thing that actually pages
   `thread/list` incrementally instead of loading the whole corpus up
   front. `query_sessions`'s `handleQuerySessions`
   (`internal/mcp/executor/query_sessions_handler.go`) calls it instead of
   the full-crawl `ListSessionsFiltered` whenever it is safe to bound:
   `provider="codex"` alone (not merged with another provider's unbounded
   results), a `limit > 0` was requested, no exact `session_id` lookup is
   in play (already O(1) via `ListSessionsFiltered`), `scope != "session"`
   (that scope needs the true most-recently-modified session across the
   *whole* corpus, which a partial fetch cannot guarantee), and
   `filter.Archived` is already resolved to a single boolean (see the
   "Known limitations" entry below — a request that needs *both* archived
   states still uses `ListSessionsFiltered`). Every other `query_sessions`
   call (no `limit`, `provider="all"`, `scope="session"`, an exact
   `session_id`, or an unresolved archived dimension) is unchanged. See
   `TestProviderFetchSessionsBoundedStopsEarly` (bounded `thread/list` call
   count, via a fake `threadSource`) and
   `TestQuerySessions_Codex_LimitUsesBoundedFetchAndReturnsMostRecent`
   (handler-level correctness proof) for the tests.

   This bounded path trades an exact whole-corpus "true most-recent N"
   guarantee for boundedness: results are re-sorted and trimmed to `limit`
   only *within whatever was actually fetched*, relying on `thread/list`
   already returning threads in a reasonable (recency-oriented) order
   rather than re-scanning the entire corpus to prove the fetched set is
   the global top N.

### Experimental turn/item pagination (capability-negotiated, currently unsupported)

`appserver.Version.SupportsExperimentalTurnPagination()`
(`internal/provider/codex/appserver/version.go`) is the capability-gate
extension point for a **not-yet-empirically-confirmed** turn/item-level
pagination method some future Codex app-server release may expose,
distinct from the stable `thread/list` pagination above. It is deliberately
gated by `ExperimentalTurnPaginationMinVersion`, a placeholder floor no
real Codex CLI version satisfies today, so it always reports `false` for
every currently-known version. meta-cc never issues a turn/item pagination
request: `thread/read(includeTurns)` (full-turn fetch, no pagination)
remains the only turn-content path in play, regardless of this capability
check's result. This matches the Contract's "experimental methods are
capability-negotiated and optional... fail safely": a `false` result here
is not an error, it just means "keep using the confirmed path" — and
wiring a real experimental method in later only requires updating this one
function plus `ExperimentalTurnPaginationMinVersion`, not touching every
caller.

## Cross-version compatibility coverage

Fixtures under `tests/fixtures/codex/` exercise:

- Legacy (pre-0.145) sessions — `rollout-legacy-sample.jsonl`.
- 0.145-shaped event families (`world_state`, `compacted`,
  `context_compacted`, `turn_aborted`, `session_end`, `tool_search_*`,
  `thread_settings_applied`) — `rollout-legacy-0145-families-sample.jsonl`.
- Compaction boundary dedup — `rollout-legacy-compaction-boundary-sample.jsonl`.
- Dot-schema (current) sessions — `rollout-new-sample.jsonl`.
- Phased (commentary/final) item ordering — `rollout-legacy-phased-sample.jsonl`.
- Duplicate-channel dedup — `rollout-legacy-dedup-sample.jsonl`.

Archived, subagent, paginated, and lineage-chain coverage is exercised at
the `appServerBackend`/`query_sessions_handler` unit-test level (hand-rolled
fake `threadSource` / SQLite fixtures — see
`internal/provider/codex/appserver_provider_test.go` and
`internal/mcp/executor/query_sessions_handler_test.go`) rather than as
static JSONL fixtures, since app-server responses and SQLite rows are
structured data, not a line-oriented log format.

## Known limitations / explicitly out of scope (DIR-032, first pass)

- ~~`world_state`, `tool_search_call`/`tool_search_output`, and
  `thread_settings_applied` remain raw passthrough~~ — RESOLVED by DIR-072:
  all four families now map to typed canonical items (see "Turn/session
  lifecycle status" above and `jsonl-schema.md`'s "Newer Event Families
  (Codex 0.145+)").
- Experimental turn/item-level pagination is capability-gated but has no
  real implementation behind it yet (no confirmed method/wire-shape exists
  to implement against) — see "Experimental turn/item pagination" above.
- `ListSessionsPage` does not resolve `filter.Archived == nil` into "both
  archived states" the way `ListSessionsFiltered` does (two independent
  passes) — a single page fetch picks one archived state (defaulting to
  non-archived); a caller wanting both issues two independent page
  sequences. This keeps the cursor contract unambiguous (one cursor space
  per archived state, matching the underlying `thread/list` semantics)
  rather than inventing a synthetic merged cursor. This is exactly why
  `query_sessions`'s DIR-034 bounded path (`FetchSessionsBounded`, see
  above) only engages when `filter.Archived` is already resolved to a
  single boolean — a `status="archived"`-without-`archived`-set request
  (which still needs both passes merged, then post-filtered by status)
  keeps using the unbounded `ListSessionsFiltered` crawl.
- The DIR-034 bounded fetch does not verify that `thread/list` returns
  threads in `CreatedAt`-descending order — it re-sorts and trims only the
  pages it actually fetched, not the whole corpus. A `query_sessions(
  provider="codex", limit=N)` call is therefore "the top N of what a
  bounded scan found", not a formally proven "the true most-recent N
  across the entire corpus" if the server's enumeration order ever departs
  from recency. `provider="all"`, `scope="session"`, and unbounded
  (`limit` unset) queries are unaffected and keep the full-crawl guarantee.
- `Turn.Completeness` is not yet surfaced through the flattened
  `UserText`/`AssistantText` MCP query output (`query_session_content`
  etc.) — those fields remain the DIR-028 compatibility projection. A
  caller that needs completeness today reads it directly off
  `conversation.Turn` (e.g. via `internal/ftsindex`, which already enforces
  it) or a future `query_session_items`-style tool (see `jsonl-schema.md`'s
  "no dedicated MCP query surface for the Item stream itself" note).

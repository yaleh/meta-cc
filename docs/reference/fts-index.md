# Local Full-Text Index (DIR-031)

> **Status: implemented — internal foundation (wiring landed).** The index is
> built and used today: project-scoped `query_session_content` calls use it as
> a candidate selector (see
> [MCP content-query integration](#mcp-content-query-integration) below). It is
> not exposed as a standalone MCP tool — there is no `query_search` tool or
> `meta-cc index rebuild` CLI command. See the
> [MCP Query Tools Reference](../guides/mcp-query-tools.md) for the public
> query surface.

meta-cc can maintain an optional, local, incremental SQLite FTS5 index
(`internal/ftsindex`) over the DIR-028 canonical conversation model
(`Session` → `Turn` → `Item`), so content search over hundreds of sessions
or multi-gigabyte rollouts does not require loading and parsing every
matching session file on every query.

The index is a **derived cache, never an authoritative source**. Raw
provider session stores (Claude JSONL, Codex rollout/SQLite/app-server)
remain the single source of truth. The index can always be deleted and
rebuilt from them with no data loss, and every search result carries enough
provenance to re-fetch the exact canonical record rather than trusting a
cached snippet.

## Location

One index file covers every provider for a given project:

```
<projectPath>/.meta-cc/index/fts.db
```

This follows the repo's existing `.meta-cc/` project-local derived-data
convention (see `.meta-cc/prompts/library/` for precedent). `EnsureDir`
writes a `.gitignore` (`*`) into `.meta-cc/index/` the first time the index
directory is created, so the database and its WAL/SHM sidecar files are
never accidentally committed.

Rows are scoped by `provider` and `cwd` columns, not by file path, so a
single `fts.db` safely holds Claude and Codex sessions for the same
project side by side.

## Schema

- `sessions` — one row per indexed session: `session_key` (`"<provider>:<session_id>"`),
  `provider`, `session_id`, `cwd`, `title`, plus the invalidation fingerprint
  (`source_path`, `source_size`, `source_mtime`, `updated_at`) and
  `indexed_at`/`item_count` bookkeeping.
- `items` — one row per indexed `conversation.Item`: `provider`, `thread_id`
  (the session/thread ID), `turn_id`, `item_id`, `role`, `kind`, `cwd`,
  `title`, `tool_name`, `ts_unix`, the privacy-capped `body` text, and a
  `truncated` flag.
- `items_fts` — an FTS5 external-content virtual table
  (`content='items', content_rowid='rowid'`) indexing `items.body` with the
  `porter unicode61` tokenizer.
- `meta` — a single `schema_version` row (see Migration below).

`modernc.org/sqlite` (already a dependency for the Codex state-db reader,
see `internal/provider/codex/sqlite.go`) is a pure-Go SQLite build with
FTS5 compiled in — no CGO, no extra system dependency.

## Lifecycle

### Incremental refresh

`ftsindex.Refresh(ctx, db, sessions, sourceMeta, loadTurns, bodyLimit)`
compares each session's *current* `SourceMeta` (on-disk file size + mtime,
via `claudeprovider.FilePath`/`codexprovider.RolloutPath`; or
`Session.UpdatedAt` for sessions with no resolvable local file, e.g.
Codex app-server-only threads) against what's already stored. A session
whose fingerprint hasn't changed is skipped **without calling `loadTurns`
at all** — an unchanged corpus performs no deep reparse. A session that is
new or has changed is fully reparsed and reindexed.

### Transactional per-session upsert

Reindexing one session (`reindexSessionTx`) runs inside a single SQLite
transaction: it deletes that session's previous `items_fts`/`items`/`sessions`
rows and inserts the new ones, then commits. Any failure before `Commit`
rolls the whole transaction back, so an interrupted reindex always leaves
the **previous** session state completely intact — never a mix of old and
new rows, and never a gap where the session temporarily has zero rows.

One session's failure (a parse error, or a transaction failure) is recorded
as a warning and does not abort the rest of the batch, mirroring
`providerrecords.Build`'s per-session tolerance (DIR-030).

### Reconciliation

`ftsindex.Reconcile(ctx, db, provider, cwd, liveSessionKeys)` removes index
rows for any session no longer present in a fresh listing for that
`(provider, cwd)` scope — covering deleted, archived-away, and moved
sessions. A rewritten session (same session key, changed `SourceMeta`) is
handled by `Refresh` itself: the transactional upsert *replaces* that
session's rows in place, never duplicating them.

Reconcile is explicitly scoped to one `(provider, cwd)` pair per call, so
reconciling one project's stale sessions can never delete another
project's still-live rows.

### Corruption recovery

`ftsindex.Open(ctx, path)` verifies the database is structurally healthy
(`PRAGMA quick_check` plus a matching `schema_version`) before returning
it. An unhealthy or unopenable file is wiped and recreated as a fresh,
empty, healthy index — this never returns a hard error or silently returns
wrong data; the returned `degraded` flag tells the caller a full `Rebuild`
is needed to make search work again (an empty index just returns no
results, which is a safe, indicated degradation rather than silent data
loss or a crash).

### Schema migration

There is no in-place schema migration. `SchemaVersion` is a single integer;
a stored version that doesn't match the current binary's `SchemaVersion` is
treated exactly like corruption — wiped and rebuilt. This keeps
invalidation logic simple, at the cost of a full reindex on a schema
change, which is an acceptable cost for a derived, fully rebuildable cache.

## Search and hydration

`ftsindex.Search(ctx, db, query, filter, limit)` runs an FTS5 `MATCH`
query joined against the `items` metadata table, with `filter.CWD`
**mandatory** — there is no "search every project" mode, matching the
project-boundary rule every other query/search entry point in this
codebase already enforces (see DIR-030's `query_sessions`/
`query_session_content` cwd-boundary fix). Optional filter dimensions:
`Provider`, `ThreadID`, `Role`, `Kind`, `ToolName`.

Each `SearchHit` carries `Provider`/`ThreadID`/`TurnID`/`ItemID` —
sufficient provenance to call `ftsindex.Hydrate(turns, hit)` against a
freshly `LoadTurns`-loaded canonical turn set and get back the exact,
un-truncated `conversation.Turn`/`Item`, not the cached (and possibly
privacy-truncated) snippet.

## Privacy defaults

Every item's searchable text is capped at `ftsindex.DefaultBodyLimitBytes`
(4096 bytes, matching the `conversation.NewRawItem` 4KB raw-provenance cap
from DIR-028) **before** it is ever written to either the metadata `body`
column or the FTS index — text beyond the cap is never indexed or
searchable, not merely truncated in display. A truncated row is flagged
(`items.truncated` / `SearchHit.Truncated`).

`ItemKindUnknown` (and any future unrecognized kind) indexes no text by
default: `Item.Raw` is provenance for hydration, not vetted for secrets, so
it is excluded from the searchable text.

## Maintenance operations

- **Disable**: set `META_CC_DISABLE_FTS_INDEX=1`; `ftsindex.IsDisabled()`
  reports this. The package itself never reads the env var — a caller
  should check it before ever calling `Open`/`Refresh`, so disabling is a
  true no-op (nothing is read or written).
- **Inspect**: `ftsindex.Inspect(ctx, db)` reports schema version, session
  count, and item count.
- **Rebuild**: `ftsindex.Rebuild(ctx, path, sessions, sourceMeta, loadTurns, bodyLimit)`
  discards any existing index at `path` (corrupt or not) and reindexes
  everything from scratch.
- **Clean**: `ftsindex.Clean(path)` permanently deletes the index database
  and its WAL/SHM sidecar files. Safe at any time — the next `Open`
  recreates an empty, healthy index from the still-fully-intact raw session
  stores.

## MCP content-query integration

Project-scoped `query_session_content` calls use the index as a candidate
selector when the provider includes Codex, the query contains a literal of at
least three characters (`contains`, or a regex-free `pattern`), and no exact
`session_id` is requested. SQLite handles incremental cache maintenance and
metadata scoping; Go reads the privacy-bounded cached bodies and applies the
same `(?i)` + quoted-literal regexp semantics as canonical gojq, including
Unicode simple-fold pairs such as `k`/`K`. Candidate sessions are then loaded
again through the provider and normalized from canonical turns before role,
time, context, grouping, pagination, and `jq_filter` processing. The index
therefore avoids loading every unchanged session without becoming an authority
for response records.

Queries fall back to the direct canonical scan when the shape cannot be narrowed
without false negatives, the index is disabled or unavailable, any indexed body
in scope was privacy-truncated, a session has incomplete/truncated history,
refresh/reconciliation is incomplete, or the database was rebuilt after
corruption. Recovery warnings are bounded summaries rather than per-row
diagnostics. Claude-only and exact `session_id` queries retain their existing
direct paths.

## Known limitations / explicitly out of scope (DIR-031, first release)

- The derived index is wired into the existing MCP content-query path, but no
  standalone `query_search` tool or `meta-cc index rebuild` CLI command is
  exposed yet. Maintenance remains available through the `internal/ftsindex`
  package.
- Metadata filter pushdown covers `provider`/`thread_id`/`role`/`kind`/`tool_name`;
  it does not yet reuse the full `conversation.SessionFilter` shape
  (e.g. `archived`, `source_kind`, time ranges) — those can be added as
  additional `SearchFilter` fields without a schema change.
- `SourceMeta` invalidation for Codex app-server-only threads (no local
  rollout file) falls back to `Session.UpdatedAt` rather than a
  content hash; this is coarser than path+size+mtime but still correctly
  triggers a reindex whenever the provider reports the thread as updated.

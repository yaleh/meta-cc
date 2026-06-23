# Plan: Multi-Provider Conversation Support — Phases 87–94

**Created**: 2026-06-14
**Proposal**: [proposal-multi-provider-conversation.md](../proposals/proposal-multi-provider-conversation.md)
**Preceding plan**: [81-86-arch-cleanup.md](81-86-arch-cleanup.md)
**Status**: Completed (internal/provider/, internal/conversation/ implemented; query_conversation_flow MCP tool live)

---

## Overview

Extend meta-cc to support OpenAI Codex CLI session history alongside Claude Code via a unified provider abstraction at the parser layer. Existing MCP tools gain an optional `provider` parameter; all current behaviour is preserved when the parameter is absent.

| Phase | Work | Est. Lines |
|-------|------|-----------|
| 87 | Canonical data model: `internal/conversation/types.go` + tests | ~130 |
| 88 | Codex locator: `internal/locator/codex.go` + Provider interface + Registry | ~180 |
| 89 | Claude Provider: UUID chain walker + `ListSessions` | ~100 |
| 90 | Claude Provider: Turn pairing + `tool_use`/`tool_result` join (2 stages) | ~220 |
| 91 | Codex SQLite: `internal/provider/codex/sqlite.go` (`ListSessions` + `GetSession`) | ~170 |
| 92 | Codex rollout parser: schema detection + legacy mapper + streaming mapper (2 stages) | ~270 |
| 93 | Codex Provider wire-up + MCP `provider` parameter | ~220 |
| 94 | End-to-end integration tests + documentation | ~180 |

**Total estimated modifications**: ~1,470 lines across 8 phases (12 stages).

---

## Architecture Decisions Locked Before Implementation

The following decisions from the proposal review MUST be reflected in every phase:

1. **`Extensions json.RawMessage`** (not `map[string]any`) on both `Session` and `Turn`. Resolves type-escape risk, enables jq pass-through without re-marshaling.
2. **`modernc.org/sqlite`** (pure Go) is mandatory. `mattn/go-sqlite3` (CGo) is explicitly forbidden.
3. **`bufio.Scanner` streaming** in `LoadTurns` — never `os.ReadFile` or `io.ReadAll`. Default per-file line limit: 500 000 lines; truncation emits a warning log, not an error.
4. **`provider` parameter defaults to `"claude"`** in all MCP tools. Default `"all"` would silently break Claude-specific jq filters on Codex records.
5. **Schema version detection is mandatory** in Codex rollout parsing: inspect the `type` field format of the first line to route to the correct mapper before processing any events.
6. **Claude adapter imports `internal/types` directly** — not `internal/parser` (which is a deprecated alias package).

---

## Dependency Order

```
Phase 87 (conversation types)
    └── Phase 88 (CodexLocator + Provider interface + Registry)
            ├── Phase 89 (Claude Provider: ListSessions + UUID chain)
            │       └── Phase 90 (Claude Provider: Turn pairing + tool join)
            ├── Phase 91 (Codex SQLite: ListSessions + GetSession)
            │       └── Phase 92 (Codex rollout: schema detect + mappers)
            │               └── Phase 93 Stage 93.1 (Codex Provider wire-up)
            └── Phase 93 Stage 93.2 (MCP provider parameter + ToolExecutor routing)
                    └── Phase 94 (integration tests + docs)
```

---

## Phase 87 — Canonical Data Model

**Goal**: Define `internal/conversation/types.go` with the canonical `Session`, `Turn`, `ToolCall`, `TokenUsage`, and `ProviderID` types. Use `json.RawMessage` for `Extensions` on both `Session` and `Turn`. Ship tests first (TDD).

**Dependencies**: None.

**Acceptance Criteria**:
- Package `internal/conversation` compiles cleanly with `go build ./internal/conversation/...`
- `Extensions` field type is `json.RawMessage` on both `Session` and `Turn` — no `map[string]any` present
- Round-trip JSON marshal/unmarshal test passes for all four types
- `go test ./internal/conversation/...` passes with ≥80% coverage
- `make commit` passes

### Stage 87.1 — Write tests for canonical types

**Files**:
- `internal/conversation/types_test.go` — new file

**What to test**:
- `Session` JSON round-trip (all fields including `Extensions json.RawMessage`)
- `Turn` JSON round-trip (including nested `ToolCall` slice and `Extensions`)
- `ToolCall` JSON round-trip
- `TokenUsage` JSON round-trip
- `ProviderID` constants (`"claude"`, `"codex"`)
- `Session.Turns` is omitted from JSON when nil (`omitempty`)
- `Extensions` field is omitted from JSON when nil (`omitempty`)

**Estimated changes**: ~70 lines
**Verification**: `go test ./internal/conversation/...` — all tests fail (red, expected at this stage)

### Stage 87.2 — Implement canonical types

**Files**:
- `internal/conversation/types.go` — new file

**Content**:
```
package conversation

type ProviderID string
const (ProviderClaude ProviderID = "claude"; ProviderCodex ProviderID = "codex")

type Session struct {
    ID         string          json:"id"
    Provider   ProviderID      json:"provider"
    Title      string          json:"title,omitempty"
    CWD        string          json:"cwd"
    Model      string          json:"model,omitempty"
    CreatedAt  time.Time       json:"created_at"
    TokenUsage TokenUsage      json:"token_usage"
    Turns      []Turn          json:"turns,omitempty"
    Extensions json.RawMessage json:"extensions,omitempty"
}

type Turn struct {
    ID            string          json:"id"
    UserText      string          json:"user_text"
    AssistantText string          json:"assistant_text,omitempty"
    ToolCalls     []ToolCall      json:"tool_calls,omitempty"
    Timestamp     time.Time       json:"timestamp"
    Extensions    json.RawMessage json:"extensions,omitempty"
}

type ToolCall struct {
    ID        string          json:"id"
    Name      string          json:"name"
    Input     json.RawMessage json:"input"
    Output    string          json:"output,omitempty"
    IsError   bool            json:"is_error"
    Timestamp time.Time       json:"timestamp"
}

type TokenUsage struct {
    InputTokens  int json:"input_tokens"
    OutputTokens int json:"output_tokens"
    CacheTokens  int json:"cache_tokens,omitempty"
}
```

Note: `ToolCall.Input` is also `json.RawMessage` (not `map[string]any`) to preserve fidelity of arbitrary tool input shapes.

**Estimated changes**: ~60 lines
**Verification**: `go test ./internal/conversation/...` — all tests pass (green)

---

## Phase 88 — Codex Locator + Provider Interface + Registry

**Goal**: Add `internal/locator/codex.go` (symmetric with `env.go`); define `internal/provider/interface.go` with the `Provider` interface (including `IsAvailable` and `GetSession`); implement `internal/provider/registry.go` with fan-out and unavailable-provider skip logic.

**Dependencies**: Phase 87 (imports `internal/conversation`)

**Acceptance Criteria**:
- `CodexLocator` resolves `~/.codex/` paths; respects `META_CC_CODEX_ROOT` env override
- `Provider` interface compiles with all five methods: `ID`, `IsAvailable`, `ListSessions`, `GetSession`, `LoadTurns`
- `Registry.MergedSessions` silently skips unavailable providers and logs a warning
- `Registry.MergedSessions` with `providerFilter=[]` queries all registered providers
- Tests cover: path resolution with env override, unavailable-provider skip, empty filter fan-out
- `make commit` passes

### Stage 88.1 — `CodexLocator` + tests

**Files**:
- `internal/locator/codex_test.go` — new file (tests first)
- `internal/locator/codex.go` — new file

**`CodexLocator` API**:
```go
const codexRootEnv = "META_CC_CODEX_ROOT"

type CodexLocator struct { codexRoot string }

func NewCodexLocator() *CodexLocator        // uses os.UserHomeDir() + ~/.codex; honours META_CC_CODEX_ROOT
func (l *CodexLocator) SQLiteDB() string    // <root>/state_5.sqlite
func (l *CodexLocator) SessionsRoot() string // <root>/sessions/
func (l *CodexLocator) HistoryFile() string  // <root>/history.jsonl
```

**What to test**: default path construction, `META_CC_CODEX_ROOT` override, each path method.

**Estimated changes**: ~60 lines (30 test + 30 impl)

### Stage 88.2 — `Provider` interface + `Registry` + tests

**Files**:
- `internal/provider/interface_test.go` — new file (mock provider for registry tests)
- `internal/provider/interface.go` — new file (Provider interface)
- `internal/provider/registry.go` — new file (Registry implementation)

**Provider interface**:
```go
type Provider interface {
    ID() conversation.ProviderID
    IsAvailable(ctx context.Context) bool
    ListSessions(ctx context.Context) ([]conversation.Session, error)
    GetSession(ctx context.Context, sessionID string) (conversation.Session, error)
    LoadTurns(ctx context.Context, sessionID string) ([]conversation.Turn, error)
}
```

**Registry**:
```go
type Registry struct { providers map[conversation.ProviderID]Provider }
func NewRegistry(providers ...Provider) *Registry
func (r *Registry) MergedSessions(ctx context.Context, providerFilter []conversation.ProviderID) ([]conversation.Session, error)
```

`MergedSessions` strategy: call each matching provider's `IsAvailable`; skip + warn if false; call `ListSessions` sequentially (not concurrent — avoids SQLite lock race; concurrent fan-out is a future optimisation); append results; return merged slice.

**What to test**: empty registry, single provider, two providers (both available), one unavailable provider skipped, provider filter narrows results.

**Estimated changes**: ~120 lines (50 test + 70 impl)

---

## Phase 89 — Claude Provider: `ListSessions` + UUID Chain Walker

**Goal**: Implement `internal/provider/claude/provider.go` — the struct, `ID()`, `IsAvailable()`, `ListSessions()`, and the UUID chain builder. `LoadTurns` is stubbed (returns `nil, nil`) at this stage.

**Dependencies**: Phase 88 (Provider interface, Registry)

**Acceptance Criteria**:
- `ClaudeProvider` implements `Provider` interface (compile-time check via `var _ provider.Provider = (*ClaudeProvider)(nil)`)
- `ListSessions` returns one `conversation.Session` per `.jsonl` file found by `locator.SessionLocator`
- UUID graph is built correctly: test with a fixture containing branching `ParentUUID` chains
- Session metadata (model, timestamps) extracted from first/last `types.SessionEntry` in each file
- `go test ./internal/provider/claude/...` passes with ≥80% coverage
- `make commit` passes

### Stage 89.1 — Tests + skeleton

**Files**:
- `internal/provider/claude/provider_test.go` — new file
- `internal/provider/claude/provider.go` — new file (struct + `ID` + `IsAvailable` + stub `LoadTurns`)

**Test fixtures**: reuse `tests/fixtures/` JSONL files already present; no new fixture files needed for this stage.

**What to test**: `ID()` returns `conversation.ProviderClaude`, `IsAvailable()` true when `META_CC_PROJECTS_ROOT` resolves to test dir, `IsAvailable()` false when dir does not exist.

**Estimated changes**: ~50 lines

### Stage 89.2 — `ListSessions` + UUID chain builder

**Files**:
- `internal/provider/claude/provider.go` — add `ListSessions`, `buildUUIDGraph`

**Implementation notes**:
- Call `locator.SessionLocator.ListFiles()` to enumerate `.jsonl` files
- For each file, read first + last `types.SessionEntry` for metadata (do NOT parse full file)
- Build `map[string]*types.SessionEntry` (UUID → entry) for later use by `LoadTurns`
- `buildUUIDGraph(entries []types.SessionEntry) map[string]*types.SessionEntry`

**What to test**: `ListSessions` with two-session fixture returns two `conversation.Session` records with correct `Provider`, `ID`, `CWD`, `CreatedAt`.

**Estimated changes**: ~50 lines

---

## Phase 90 — Claude Provider: Turn Pairing + Tool Join (2 Stages)

**Goal**: Implement `LoadTurns` in `internal/provider/claude/provider.go`. Turn reconstruction is non-trivial: walk the UUID parent chain to pair `(user, assistant)` entries, then join `tool_use` blocks (from assistant entries) with `tool_result` blocks (from subsequent user entries) by `ToolUseID`.

**Dependencies**: Phase 89

**Acceptance Criteria**:
- `LoadTurns` returns correctly paired `conversation.Turn` records for a fixture session
- Tool call `Input` (from `assistant` entry `tool_use` block) is joined with `Output`/`IsError` (from `user` entry `tool_result` block) by matching `id`/`tool_use_id`
- Tool calls with no matching `tool_result` are included with empty `Output`
- `go test ./internal/provider/claude/...` passes; coverage ≥80%
- `make commit` passes

### Stage 90.1 — UUID chain walker + (user, assistant) Turn pairing

**Files**:
- `internal/provider/claude/provider_test.go` — add tests for Turn pairing
- `internal/provider/claude/turns.go` — new file: `buildTurns(entries []types.SessionEntry) []turnPair`

**Data structures**:
```go
type turnPair struct {
    user      *types.SessionEntry
    assistant *types.SessionEntry
}
```

**Algorithm**:
1. Build `map[string]*types.SessionEntry` from all entries in the session file
2. Walk entries in file order; when a `user` entry is encountered, look ahead for its paired `assistant` entry (matched by `ParentUUID` pointing to the user entry's UUID)
3. Collect `(user, assistant)` pairs as `turnPair` slice

**What to test**: linear chain (1 user → 1 assistant → 1 user → 1 assistant), branched chain, orphaned assistant entry (no user parent).

**Estimated changes**: ~100 lines (45 test + 55 impl)

### Stage 90.2 — `tool_use`/`tool_result` join + `LoadTurns`

**Files**:
- `internal/provider/claude/provider_test.go` — add tests for tool join
- `internal/provider/claude/turns.go` — add `joinToolCalls`, update `LoadTurns`

**Algorithm**:
1. From each `turnPair.assistant` entry, collect all `ContentBlock` items with `Type == "tool_use"` → build `map[toolUseID]ToolCall{ID, Name, Input}`
2. From each `turnPair.user` entry, collect all `ContentBlock` items with `Type == "tool_result"` → for each, look up by `ToolUseID` in the map and fill `Output`, `IsError`
3. Assemble `conversation.Turn{ID, UserText, AssistantText, ToolCalls, Timestamp}`

**What to test**: assistant entry with two `tool_use` blocks, matching `tool_result` in next user entry, `tool_result` with `IsError: true`, `tool_use` with no matching `tool_result`.

**Estimated changes**: ~120 lines (50 test + 70 impl)

---

## Phase 91 — Codex SQLite: `ListSessions` + `GetSession`

**Goal**: Implement `internal/provider/codex/sqlite.go` to read `state_5.sqlite` using `modernc.org/sqlite`. Add `go.mod` dependency. Defensive column-name scanning; graceful degradation if schema differs.

**Dependencies**: Phase 88 (Provider interface)

**Acceptance Criteria**:
- `go.mod` references `modernc.org/sqlite` at a pinned minor version; `mattn/go-sqlite3` is absent
- `ListSessions` returns all threads from the `threads` table as `conversation.Session` records
- `GetSession` returns a single session without full table scan if possible (indexed by `id`)
- If `threads` table is missing or a required column is absent, returns a structured `*SQLiteSchemaError` (not a panic)
- Column scanning uses column names (not positional indices)
- Tests use a temporary in-memory SQLite database (not the user's real `~/.codex/`)
- `go test ./internal/provider/codex/...` passes with ≥80% coverage
- `make commit` passes

### Stage 91.1 — Tests + `modernc.org/sqlite` dependency

**Files**:
- `go.mod`, `go.sum` — add `modernc.org/sqlite`
- `internal/provider/codex/sqlite_test.go` — new file

**Test approach**: create an in-memory SQLite DB in the test, insert sample rows in the `threads` schema, call `ListSessions`/`GetSession`, assert returned `conversation.Session` values.

**Schema under test**:
```sql
CREATE TABLE threads (
    id TEXT PRIMARY KEY,
    cwd TEXT,
    title TEXT,
    model TEXT,
    tokens_used INTEGER,
    source TEXT,
    created_at INTEGER  -- unix milliseconds
);
```

**What to test**: happy path (3 rows → 3 Sessions), missing `threads` table → structured error, missing column → structured error, `GetSession` with known ID, `GetSession` with unknown ID → `ErrSessionNotFound`.

**Estimated changes**: ~80 lines (all tests; no impl yet — red phase)

### Stage 91.2 — `sqlite.go` implementation

**Files**:
- `internal/provider/codex/sqlite.go` — new file
- `internal/provider/codex/errors.go` — new file (define `SQLiteSchemaError`, `ErrSessionNotFound`)

**Implementation notes**:
- Open DB with `modernc.org/sqlite` driver name `"sqlite"` (not `"sqlite3"`)
- Verify table exists: `SELECT name FROM sqlite_master WHERE type='table' AND name='threads'`
- Scan columns by name using `rows.Columns()` + a column-name→index map
- `created_at` is unix milliseconds → convert to `time.Time` via `time.UnixMilli`
- `TokenUsage.InputTokens` is populated from `tokens_used` (Codex does not split input/output in index); `OutputTokens` stays 0 until `LoadTurns` is called

**Estimated changes**: ~90 lines

---

## Phase 92 — Codex Rollout Parser: Schema Detection + Mappers (2 Stages)

**Goal**: Implement `internal/provider/codex/rollout.go` with streaming line-by-line parsing (`bufio.Scanner`), schema version detection, and two event mappers: one for the legacy format (`session_meta`/`event_msg`/`response_item`/`turn_context`) and one for the ≥0.44 format (`thread.started`/`turn.started`/`item.*`/`turn.completed`).

**Dependencies**: Phase 91 (shares `internal/provider/codex/` package)

**Acceptance Criteria**:
- `LoadTurns` never calls `os.ReadFile` or `io.ReadAll` on rollout files
- Schema version detected from first line's `type` field: dot-notation (`turn.started`) → new mapper; no-dot (`session_meta`) → legacy mapper
- Per-file line limit enforced (default 500 000); truncation logs a warning, returns partial results
- Unknown event types are stored in `Turn.Extensions` as `json.RawMessage` array under key `"codex_events"`
- Tests use small fixture files (not real `~/.codex/` data)
- `go test ./internal/provider/codex/...` passes with ≥80% coverage
- `make commit` passes

### Stage 92.1 — Schema version detection + legacy mapper

**Files**:
- `internal/provider/codex/rollout_test.go` — new file (legacy mapper tests)
- `internal/provider/codex/rollout.go` — new file (schema detection + legacy mapper)

**Schema version detection**:
```go
func detectSchemaVersion(firstLine []byte) schemaVersion
// schemaVersion: schemaLegacy | schemaNew
// heuristic: unmarshal {"type": "..."} and check strings.Contains(typeField, ".")
```

**Legacy mapper** handles events: `session_meta` → update Session metadata, `event_msg` with `role:"user"` → `Turn.UserText`, `event_msg` with `role:"assistant"` → `Turn.AssistantText`, `response_item` (tool call) → `Turn.ToolCalls`, `turn_context` → close Turn boundary.

**Fixture file for legacy tests**: create `tests/fixtures/codex/rollout-legacy-sample.jsonl` (5–10 lines, hand-crafted).

**What to test**: version detection (dot → new, no-dot → legacy), legacy mapper produces correct `Turn` records, unknown legacy event type stored in `Extensions`.

**Estimated changes**: ~130 lines (50 test + 50 impl + 30 fixture)

### Stage 92.2 — ≥0.44 streaming mapper + line limit

**Files**:
- `internal/provider/codex/rollout_test.go` — add new-schema mapper tests + line-limit test
- `internal/provider/codex/rollout.go` — add new-schema mapper, `bufio.Scanner` main loop, line limit

**New-schema mapper** handles: `thread.started` → Session metadata, `turn.started` → new Turn boundary, `item.message role:user` → `Turn.UserText`, `item.message role:assistant` → `Turn.AssistantText`, `item.tool_call` → start `ToolCall`, `item.tool_result` → fill `ToolCall.Output`/`IsError`, `turn.completed` → close Turn, `turn.failed` → `Turn.Extensions["turn_failed"]`, `error` → `Turn.Extensions["codex_error"]`, unknown → append to `Turn.Extensions["codex_events"]`.

**`LoadTurns` main loop**:
```go
scanner := bufio.NewScanner(f)
scanner.Buffer(make([]byte, 1<<20), 1<<20)  // 1 MB line buffer for long lines
lineCount := 0
for scanner.Scan() {
    lineCount++
    if lineCount > p.maxLines {
        log.Warnf("rollout file truncated at %d lines: %s", p.maxLines, path)
        break
    }
    // dispatch to mapper
}
```

**Fixture file for new-schema tests**: create `tests/fixtures/codex/rollout-new-sample.jsonl`.

**What to test**: new-schema mapper produces correct Turns, tool call join (item.tool_call → item.tool_result), line limit triggers warning and returns partial Turns, unknown event appended to Extensions.

**Estimated changes**: ~140 lines (60 test + 60 impl + 20 fixture)

---

## Phase 93 — Codex Provider Wire-up + MCP `provider` Parameter (2 Stages)

**Goal**: Wire `sqlite.go` + `rollout.go` into a complete `CodexProvider` implementing the `Provider` interface; add `provider` parameter to `StandardToolParameters()`; route MCP queries through `provider.Registry` when `provider != "claude"`.

**Dependencies**: Phase 90 (Claude Provider complete), Phase 92 (Codex rollout complete)

**Acceptance Criteria**:
- `CodexProvider` implements `Provider` interface (compile-time assertion)
- `IsAvailable` returns false when `state_5.sqlite` does not exist
- MCP tool `provider="codex"` returns Codex sessions; `provider="claude"` (default) returns Claude sessions unchanged
- `provider="all"` merges both; each record has `"provider"` field
- Existing MCP tool calls without `provider` parameter behave identically to before (backward compatible)
- `scope=session` with `provider=codex` is documented to have no effect (Codex uses `session_id` only)
- `make commit` passes

### Stage 93.1 — `CodexProvider` struct + `IsAvailable` + wire-up

**Files**:
- `internal/provider/codex/provider_test.go` — new file
- `internal/provider/codex/provider.go` — new file

**Content**:
```go
type CodexProvider struct {
    locator  *locator.CodexLocator
    maxLines int  // default 500_000
}

func NewCodexProvider(loc *locator.CodexLocator) *CodexProvider
func (p *CodexProvider) ID() conversation.ProviderID { return conversation.ProviderCodex }
func (p *CodexProvider) IsAvailable(ctx context.Context) bool  // checks SQLiteDB() exists
func (p *CodexProvider) ListSessions(ctx context.Context) ([]conversation.Session, error) // delegates to sqlite.go
func (p *CodexProvider) GetSession(ctx context.Context, id string) (conversation.Session, error)
func (p *CodexProvider) LoadTurns(ctx context.Context, id string) ([]conversation.Turn, error) // delegates to rollout.go
```

**Startup wiring**: update `cmd/mcp-server/main.go` to construct `CodexLocator`, `CodexProvider`, `ClaudeProvider`, and `Registry`; inject Registry into `ToolExecutor`.

**What to test**: `IsAvailable` false when SQLite absent, `IsAvailable` true when SQLite present (use temp file), `LoadTurns` delegates to rollout reader.

**Estimated changes**: ~110 lines (40 test + 50 impl + 20 main.go wiring)

### Stage 93.2 — MCP `provider` parameter + `ToolExecutor` routing

**Files**:
- `internal/mcp/tools/tools.go` — add `provider` to `StandardToolParameters()`
- `internal/mcp/executor/executor.go` — read `provider` param; route through Registry when `provider != "claude"`
- `internal/mcp/tools/tools_test.go` — add test: `provider` param present in merged parameters

**`provider` parameter definition**:
```go
"provider": {
    Type:        "string",
    Description: `Provider filter: "claude" (default), "codex", or "all" (merged results include provider field).
                  When "all", existing jq filters designed for Claude schema may not match Codex records.`,
    Enum:        []string{"claude", "codex", "all"},
},
```

**Routing logic in executor**: if `provider == ""` or `provider == "claude"`, use existing Claude-only path (no change). If `provider == "codex"` or `provider == "all"`, fetch sessions from Registry, serialize to temp JSONL, pass to existing `StreamFiles` pipeline.

**Critical: `scope` validation guard must be bypassed for non-Claude providers.** `ExecuteTool` in `internal/mcp/executor/executor.go` (line 98) has a hard guard:
```go
if scope != "project" && scope != "session" {
    return "", fmt.Errorf("invalid scope %q: must be \"project\" or \"session\"", scope)
}
```
This guard runs *before* handler dispatch. When `provider == "codex"` or `provider == "all"`, the routing to the Registry must happen *before* this scope check (i.e., inside `ExecuteSpecialTool` which is checked first, or by adding a provider-routing check immediately after `DetermineScope`). The simplest approach: extract provider routing as a new special-tool branch in `ExecuteSpecialTool`, so it short-circuits the scope validation entirely. This structural change must be accounted for in the ~70 implementation lines of Stage 93.2.

**`scope` interaction**: document in tool description that `scope=session` applies only to `provider=claude`; for `provider=codex`, use `session_id` to scope to a single session.

**What to test**: `provider=""` takes existing path, `provider="claude"` takes existing path, `provider="codex"` calls Registry, `provider="all"` merges both.

**Estimated changes**: ~110 lines (40 test + 70 impl)

---

## Phase 94 — End-to-End Integration Tests + Documentation

**Goal**: Write integration tests covering the full multi-provider pipeline using mock providers and fixture files. Update MCP tool descriptions and docs. Decide fate of `history.jsonl` and document `scope`/`provider` interaction.

**Dependencies**: Phase 93 (full pipeline complete)

**Acceptance Criteria**:
- Integration test: `provider="all"` returns merged sessions from both providers with `provider` field
- Integration test: unavailable provider skipped, other provider's results returned normally
- Integration test: existing Claude-only query (`provider` absent) returns identical results to pre-Phase-87 behaviour
- `go test -tags integration ./tests/integration/...` passes (skipped in `go test -short`)
- `docs/guides/mcp.md` updated: `provider` parameter documented for all relevant tools
- `docs/reference/repository-structure.md` updated: new packages listed
- Non-Goal confirmed in docs: `~/.codex/history.jsonl` excluded from this feature
- `make push` passes (full check)

### Stage 94.1 — Integration tests

**Files**:
- `tests/integration/multi_provider_test.go` — new file
- `tests/fixtures/codex/rollout-legacy-sample.jsonl` — if not already created in Phase 92
- `tests/fixtures/codex/rollout-new-sample.jsonl` — if not already created in Phase 92

**Test scenarios**:
1. `provider="claude"`: returns sessions from `tests/fixtures/` only; no Codex data mixed in
2. `provider="codex"` with mock SQLite + fixture rollout: returns Codex sessions
3. `provider="all"`: merged results; each record has `"provider"` field; jq filter `.[] | select(.provider == "codex")` isolates Codex records
4. Unavailable Codex provider (no SQLite file): `provider="all"` returns only Claude sessions + warning log
5. Line-limit truncation: rollout file with >500k lines → partial result + warning (synthetic test)

**Estimated changes**: ~100 lines

### Stage 94.2 — Documentation updates

**Files**:
- `docs/guides/mcp.md` — add `provider` parameter to tool reference section
- `docs/reference/repository-structure.md` — add `internal/conversation/`, `internal/provider/` entries
- `docs/proposals/proposal-multi-provider-conversation.md` — update Status from Draft → Implemented

**Content for `docs/guides/mcp.md`**:
- `provider` parameter: values `"claude"` (default), `"codex"`, `"all"`; behaviour of each; warning about jq filter portability when using `"all"`
- `scope` + `provider` interaction note: `scope=session` applies to `provider=claude` only
- `history.jsonl` exclusion note

**Estimated changes**: ~80 lines

---

## Testing Strategy

### TDD Protocol (per-Stage)

1. Write test file first (red)
2. Implement until tests pass (green)
3. Run `make lint` — fix any issues immediately
4. Run `make commit` — must pass before proceeding

### Test Isolation

- Unit tests: no external dependencies; use `t.TempDir()` for file fixtures, in-memory SQLite for DB tests
- Integration tests: tagged `//go:build integration` OR use `testing.Short()` skip guard
- CI runs `go test -short ./...` (unit only); local `go test ./...` runs all

### Coverage Targets

- `internal/conversation/`: ≥80%
- `internal/locator/` (new `codex.go`): ≥80%
- `internal/provider/`: ≥80% per sub-package
- Existing packages: no coverage regression

### Fixture Files

| Path | Purpose |
|------|---------|
| `tests/fixtures/codex/rollout-legacy-sample.jsonl` | Legacy event format (session_meta/event_msg) |
| `tests/fixtures/codex/rollout-new-sample.jsonl` | ≥0.44 event format (turn.started/item.*) |

---

## Key File Paths

### New Files

| Path | Purpose |
|------|---------|
| `internal/conversation/types.go` | Canonical `Session`, `Turn`, `ToolCall`, `TokenUsage`, `ProviderID` |
| `internal/conversation/types_test.go` | Type round-trip tests |
| `internal/locator/codex.go` | `CodexLocator` path resolver |
| `internal/locator/codex_test.go` | Locator tests |
| `internal/provider/interface.go` | `Provider` interface |
| `internal/provider/registry.go` | `Registry` — multi-provider fan-out |
| `internal/provider/interface_test.go` | Registry tests (mock provider) |
| `internal/provider/claude/provider.go` | Claude adapter (wraps locator + types) |
| `internal/provider/claude/turns.go` | UUID chain walker + tool join |
| `internal/provider/claude/provider_test.go` | Claude provider tests |
| `internal/provider/codex/sqlite.go` | SQLite reader for `state_5.sqlite` |
| `internal/provider/codex/rollout.go` | Rollout JSONL streaming parser |
| `internal/provider/codex/provider.go` | Codex `Provider` implementation |
| `internal/provider/codex/errors.go` | `SQLiteSchemaError`, `ErrSessionNotFound` |
| `internal/provider/codex/sqlite_test.go` | SQLite tests (in-memory DB) |
| `internal/provider/codex/rollout_test.go` | Rollout parser tests (fixture files) |
| `internal/provider/codex/provider_test.go` | Codex provider integration tests |
| `tests/fixtures/codex/rollout-legacy-sample.jsonl` | Legacy format fixture |
| `tests/fixtures/codex/rollout-new-sample.jsonl` | ≥0.44 format fixture |
| `tests/integration/multi_provider_test.go` | End-to-end multi-provider tests |

### Modified Files

| Path | Change |
|------|--------|
| `go.mod`, `go.sum` | Add `modernc.org/sqlite` (pinned minor version) |
| `internal/mcp/tools/tools.go` | Add `provider` to `StandardToolParameters()` |
| `internal/mcp/executor/executor.go` | Read `provider` param; route through Registry |
| `cmd/mcp-server/main.go` | Construct and inject `provider.Registry` at startup |
| `docs/guides/mcp.md` | Document `provider` parameter |
| `docs/reference/repository-structure.md` | List new packages |

---

## Risk Register

| Risk | Likelihood | Mitigation | Responsible Phase |
|------|-----------|-----------|------------------|
| Codex SQLite schema changes between CLI versions | Medium | Column-name scanning; structured error on missing column; `IsAvailable` fast-fails gracefully | 91 |
| Rollout JSONL has undocumented schema versions beyond legacy + ≥0.44 | High | Unknown event types stored in Extensions; version detection routing; unit tests against fixture files | 92 |
| Rollout files reach 700 MB–2 GB causing OOM | High | `bufio.Scanner` streaming mandatory; 500k line limit default; test with synthetic large file | 92 |
| `modernc.org/sqlite` API change between minor versions | Low | Pin exact minor version in `go.mod`; add `go build -tags purego` CI check | 91 |
| Breaking existing MCP tool contracts | Low | `provider` defaults to `"claude"`; existing path unchanged when param absent | 93 |
| `scope=session` + `provider=codex` semantic ambiguity | Medium | Document clearly: `scope` ignored for Codex; use `session_id` for per-session Codex queries | 93–94 |
| executor.go `scope` guard blocks non-Claude provider routing | High | Provider routing must be inserted *before* scope validation in `ExecuteTool`; handle via `ExecuteSpecialTool` branch | 93 |

---

## 审查摘要

**审查日期**: 2026-06-14
**审查者**: 严苛架构师视角（Claude Code）

### 检查结论

| 检查项 | 结论 | 详情 |
|--------|------|------|
| 目标对齐 | 通过 | Proposal 5 条目标均有对应 Phase 覆盖：目标1→Phase 93–94，目标2→Phase 87–90，目标3→Phase 89–90 设计，目标4→Phase 91–92，目标5→Phase 92 Extensions 映射 |
| Stage 行数 ≤200 | 通过（有注记） | 所有 Stage 均在 200 行内。Stage 92.2 估算 140 行，Stage 93.1 估算 110 行，Stage 93.2 原估算 110 行（加入 scope 绕行逻辑后上调为 ~120 行，仍在限制内）|
| Phase 行数 ≤500 | 通过 | 各 Phase 估算：87→130，88→180，89→100，90→220，91→170，92→270，93→220，94→180。Phase 90、92、93 均在 500 行以下，且已通过拆分 2 Stages 确保每 Stage ≤200 行 |
| 依赖顺序 | 通过 | 依赖图正确：87→88→{89,91}→{90,92}→93→94。Phase 88 定义接口后 Claude provider（89–90）和 Codex provider（91–92）可并行实现 |
| TDD 合规 | 通过 | 所有 Stage 均先写测试文件（_test.go），再写实现；每 Phase 列出覆盖率目标 ≥80% |
| 风险覆盖 | 部分通过→已修正 | Proposal 7 条风险均有对应 Phase。**新发现风险**：executor.go 的 scope 硬校验守卫会在路由到 Registry 之前拒绝请求——已补充为 Risk Register 第 6 条，并在 Stage 93.2 实现说明中加入绕行方案 |
| 遗留争议点 | 通过 | 3.1 scope/provider 交叉→Phase 93–94 明确处理；3.2 history.jsonl→Phase 94.2 文档化为 Non-Goal；3.3 Registry 并发→Stage 88.2 明确选择顺序调用 |
| 文档内部一致性 | 已修正 | 见下方修改列表 |

### 本次修改内容

| # | 修改文件 | 位置 | 内容 |
|---|---------|------|------|
| 1 | Plan（本文档）| Phase 88 AC | "all four methods" → "all five methods"（接口实际有5个方法：ID/IsAvailable/ListSessions/GetSession/LoadTurns）|
| 2 | Plan（本文档）| Stage 93.2 实现说明 | 新增"Critical: scope 验证守卫绕行"段落——executor.go 第 98 行的 scope 硬校验在 handler dispatch 之前运行，必须通过 ExecuteSpecialTool 分支提前短路，否则 provider=codex/all 请求会被拒绝 |
| 3 | Plan（本文档）| Risk Register | 新增第 6 条风险：executor.go scope 守卫阻断非 Claude provider 路由，Responsible Phase: 93 |
| 4 | Proposal | Phase Plan 表注记 | 说明 Plan 将 Proposal Phase 95（集成测试）和 Phase 96（文档）合并入 Plan Phase 94，Phase 范围收窄为 87–94；Proposal 标题保留 87–96 以反映原始估算历史 |
| 5 | Proposal | 架构师审查 §二.2.2 | 修正 GetSession 描述：该方法已纳入接口正文（Phase 88 实现），而非推迟至 "Phase 96 docs 阶段"；原注记与实际接口定义矛盾 |

### Phase 范围差异说明

Proposal 标题为 "Phases 87–96"，Plan 标题为 "Phases 87–94"。差异来自 Plan 的合并设计：

- Proposal Phase 95（集成测试，~210 行）拆分后实际工作量约 100 行→并入 Plan Phase 94 Stage 94.1
- Proposal Phase 96（文档，~90 行）→并入 Plan Phase 94 Stage 94.2
- 合并后 Plan Phase 94 总计 ~180 行，在 Phase ≤500 行和 Stage ≤200 行限制内

两个文档的 Phase 编号不再需要逐一对应，Proposal 保留原始历史估算（87–96），Plan 反映实际执行计划（87–94）。此差异已在两个文档中明确说明。

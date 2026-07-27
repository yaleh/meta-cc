# Session JSONL Schema

This document describes the session history JSONL schemas supported by meta-cc:

- **Claude Code** transcripts stored in `~/.claude/projects/<project-hash>/`
- **Codex** transcripts stored in `$CODEX_HOME/sessions` or `~/.codex/sessions`

Internally, meta-cc normalizes both hosts into a Claude-like message/tool schema before running MCP queries and analysis. Existing jq filters that target `.type`, `.message.content`, tool blocks, and `.message.usage` work across both hosts for the supported common events.

## Overview

Claude Code and Codex session files use **newline-delimited JSON (JSONL)** format, where each line is a complete, self-contained JSON record. Claude Code records form a directed acyclic graph (DAG) through parent-child relationships. Codex records use event records with a top-level `payload`; meta-cc builds equivalent parent links during normalization.

**Key characteristics:**
- One JSON object per line
- Records linked via `uuid` and `parentUuid` fields
- Chronologically ordered by `timestamp`
- Claude Code has five main record types: `user`, `assistant`, `system`, `file-history-snapshot`, `summary`
- Codex has top-level event types such as `session_meta`, `response_item`, `event_msg`, `turn_context`, and `compacted`

## Host Storage

| Host | Default transcript root | Project matching |
|------|--------------------------|------------------|
| Claude Code | `~/.claude/projects/<project-hash>/` | Project path hash directory |
| Codex | `$CODEX_HOME/sessions` or `~/.codex/sessions` | Recursive JSONL scan, matched by project path in transcript content |

`META_CC_PROJECTS_ROOT` can still override the Claude-style project-hash root for tests and custom layouts.

## Record Types

The first five record types below describe Claude Code's native schema and the normalized schema that Codex records are converted into.

### 1. User Entry

Represents user input messages, including:
- User-typed prompts
- Tool execution results (returned to assistant)
- System-generated user messages (e.g., from slash commands)

**Structure:**
```json
{
  "type": "user",
  "uuid": "string (UUID)",
  "parentUuid": "string (UUID) | null",
  "timestamp": "string (ISO8601)",
  "sessionId": "string (UUID)",
  "cwd": "string (absolute path)",
  "gitBranch": "string",
  "version": "string (Claude Code version)",
  "userType": "string",
  "isSidechain": "boolean",
  "message": {
    "role": "user",
    "content": "string | ContentBlock[]"
  },

  // Optional fields
  "isMeta": "boolean",
  "isCompactSummary": "boolean",
  "isVisibleInTranscriptOnly": "boolean",
  "thinkingMetadata": {
    "level": "string",
    "disabled": "boolean",
    "triggers": "array"
  },
  "toolUseResult": {
    "stdout": "string",
    "stderr": "string",
    "interrupted": "boolean",
    "isImage": "boolean"
  }
}
```

**Field descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✓ | Always `"user"` |
| `uuid` | string | ✓ | Unique identifier for this entry |
| `parentUuid` | string\|null | ✓ | UUID of parent entry; `null` for first message |
| `timestamp` | string | ✓ | ISO8601 timestamp (e.g., `"2025-10-24T14:07:36.078Z"`) |
| `sessionId` | string | ✓ | Session identifier (UUID) |
| `cwd` | string | ✓ | Current working directory |
| `gitBranch` | string | ✓ | Git branch name at time of message |
| `version` | string | ✓ | Claude Code version (e.g., `"2.0.26"`) |
| `userType` | string | ✓ | User type (typically `"external"`) |
| `isSidechain` | boolean | ✓ | Whether this is a sidechain conversation |
| `message.role` | string | ✓ | Always `"user"` |
| `message.content` | string\|array | ✓ | Message content (see Content Formats below) |
| `isMeta` | boolean | ✗ | If true, message is metadata/system-generated |
| `isCompactSummary` | boolean | ✗ | If true, message is a compact summary |
| `isVisibleInTranscriptOnly` | boolean | ✗ | If true, visible only in transcript |
| `thinkingMetadata` | object | ✗ | Metadata about extended thinking mode |
| `toolUseResult` | object | ✗ | Present when returning tool execution results |

### 2. Assistant Entry

Represents assistant responses, including:
- Text responses
- Tool use requests
- Extended thinking content

**Structure:**
```json
{
  "type": "assistant",
  "uuid": "string (UUID)",
  "parentUuid": "string (UUID)",
  "timestamp": "string (ISO8601)",
  "sessionId": "string (UUID)",
  "cwd": "string (absolute path)",
  "gitBranch": "string",
  "version": "string (Claude Code version)",
  "userType": "string",
  "isSidechain": "boolean",
  "requestId": "string",
  "message": {
    "model": "string",
    "id": "string",
    "type": "message",
    "role": "assistant",
    "content": "ContentBlock[]",
    "stop_reason": "string | null",
    "stop_sequence": "string | null",
    "usage": {
      "input_tokens": "integer",
      "cache_creation_input_tokens": "integer",
      "cache_read_input_tokens": "integer",
      "cache_creation": {
        "ephemeral_5m_input_tokens": "integer",
        "ephemeral_1h_input_tokens": "integer"
      },
      "output_tokens": "integer",
      "service_tier": "string"
    }
  },

  // Optional fields
  "isApiErrorMessage": "boolean"
}
```

**Field descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✓ | Always `"assistant"` |
| `uuid` | string | ✓ | Unique identifier for this entry |
| `parentUuid` | string | ✓ | UUID of parent entry (never null for assistant) |
| `requestId` | string | ✓ | API request identifier |
| `message.model` | string | ✓ | Model identifier (e.g., `"claude-sonnet-4-5-20250929"`) |
| `message.id` | string | ✓ | Message identifier from API |
| `message.role` | string | ✓ | Always `"assistant"` |
| `message.content` | array | ✓ | Array of content blocks (see below) |
| `message.usage` | object | ✓ | Token usage statistics |
| `isApiErrorMessage` | boolean | ✗ | If true, this is an error message from API |

### 3. System Entry

Represents system-level events, primarily API errors and retries.

**Structure:**
```json
{
  "type": "system",
  "uuid": "string (UUID)",
  "parentUuid": "string (UUID)",
  "timestamp": "string (ISO8601)",
  "sessionId": "string (UUID)",
  "cwd": "string (absolute path)",
  "gitBranch": "string",
  "version": "string (Claude Code version)",
  "userType": "string",
  "isSidechain": "boolean",
  "subtype": "string",
  "level": "string",
  "error": "object",
  "retryInMs": "number",
  "retryAttempt": "integer",
  "maxRetries": "integer",

  // Optional fields
  "isMeta": "boolean",
  "cause": "string",
  "content": "any",
  "compactMetadata": "object",
  "logicalParentUuid": "string (UUID)"
}
```

**Field descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✓ | Always `"system"` |
| `subtype` | string | ✓ | System event subtype (e.g., `"api_error"`) |
| `level` | string | ✓ | Log level: `"error"`, `"warn"`, `"info"`, etc. |
| `error` | object | ✓ | Error details object |
| `retryInMs` | number | ✗ | Milliseconds until retry (for retryable errors) |
| `retryAttempt` | integer | ✗ | Current retry attempt number |
| `maxRetries` | integer | ✗ | Maximum retry attempts |
| `logicalParentUuid` | string | ✗ | Logical parent (different from parentUuid) |

### 4. File History Snapshot

Tracks file state changes associated with specific messages.

**Structure:**
```json
{
  "type": "file-history-snapshot",
  "messageId": "string (UUID)",
  "timestamp": "string (ISO8601)",
  "isSnapshotUpdate": "boolean",
  "snapshot": {
    "messageId": "string (UUID)",
    "trackedFileBackups": "object",
    "timestamp": "string (ISO8601)"
  }
}
```

**Field descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✓ | Always `"file-history-snapshot"` |
| `messageId` | string | ✓ | UUID of associated message |
| `isSnapshotUpdate` | boolean | ✓ | If true, this updates existing snapshot |
| `snapshot.trackedFileBackups` | object | ✓ | Map of file paths to backup data |

**Note:** File history snapshots do NOT have `uuid` or `parentUuid` fields. They reference messages via `messageId`.

### 5. Summary Entry

Session-level summary metadata.

**Structure:**
```json
{
  "type": "summary",
  "summary": "string",
  "leafUuid": "string (UUID)"
}
```

**Field descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✓ | Always `"summary"` |
| `summary` | string | ✓ | Human-readable session summary |
| `leafUuid` | string | ✓ | UUID of last entry in conversation chain |

**Note:** Summary entries do NOT have `uuid`, `parentUuid`, or `timestamp` fields.

## Content Formats

### User Message Content

User messages can have two content formats:

1. **Plain string** (simple text message):
```json
{
  "role": "user",
  "content": "Explain how JSONL parsing works"
}
```

2. **Structured array** (tool results or complex content):
```json
{
  "role": "user",
  "content": [
    {
      "type": "tool_result",
      "tool_use_id": "toolu_01ABC...",
      "content": "command output here",
      "is_error": false
    },
    {
      "type": "text",
      "text": "Additional context"
    }
  ]
}
```

### Assistant Message Content

Assistant messages always use structured array format with content blocks:

**Content block types:**

1. **Text Block** - Plain text response:
```json
{
  "type": "text",
  "text": "Response text here"
}
```

2. **Tool Use Block** - Request to execute a tool:
```json
{
  "type": "tool_use",
  "id": "toolu_01ABC...",
  "name": "Read",
  "input": {
    "file_path": "/path/to/file"
  }
}
```

3. **Thinking Block** - Extended thinking content (Claude 3.5+):
```json
{
  "type": "thinking",
  "thinking": "Extended reasoning content...",
  "signature": "cryptographic signature"
}
```

## Record Relationships

### Parent-Child Chain

Records form a conversation tree through `uuid`/`parentUuid` relationships:

```
Entry 1 (user, parentUuid=null)
  └─> Entry 2 (user, parentUuid=uuid1)
       └─> Entry 3 (user, parentUuid=uuid2)
            └─> Entry 4 (assistant, parentUuid=uuid3)
                 └─> Entry 5 (assistant, parentUuid=uuid4)  [multiple assistant blocks]
                      └─> Entry 6 (user, parentUuid=uuid5)  [tool result]
                           └─> Entry 7 (assistant, parentUuid=uuid6)
```

**Key observations:**
1. **First entry** has `parentUuid=null` (root of conversation)
2. **Conversation alternates** between user and assistant entries
3. **Multiple assistant entries** can appear sequentially (streaming response chunks)
4. **Tool execution** creates: assistant (tool_use) → user (tool_result) → assistant (response)
5. **System entries** inserted when errors occur (inherits parentUuid from interrupted entry)

### Sidechain Conversations

The `isSidechain` field indicates parallel conversation branches:
- `isSidechain: false` - Main conversation thread
- `isSidechain: true` - Branched conversation (e.g., background agent)

### File History Linkage

File snapshots link to messages via `messageId`:
```
file-history-snapshot (messageId=uuid3)
  references
    Entry 3 (user, uuid=uuid3)
```

### Summary Linkage

Session summaries link to last entry via `leafUuid`:
```
summary (leafUuid=uuid7)
  references
    Entry 7 (assistant, uuid=uuid7) [last entry in session]
```

## Temporal Ordering

All records (except `summary`) have `timestamp` fields in ISO8601 format:
```
"timestamp": "2025-10-24T14:07:36.078Z"
```

**Ordering guarantees:**
- Records appear in chronological order within each session file
- Child entries always have `timestamp >= parent.timestamp`
- Tool use → tool result maintains temporal sequence

## Tool Execution Pattern

Tool execution follows this pattern:

**1. Assistant requests tool execution:**
```json
{
  "type": "assistant",
  "uuid": "uuid-A",
  "message": {
    "content": [
      {
        "type": "tool_use",
        "id": "toolu_123",
        "name": "Bash",
        "input": {"command": "ls -la"}
      }
    ]
  }
}
```

**2. User returns tool result:**
```json
{
  "type": "user",
  "uuid": "uuid-B",
  "parentUuid": "uuid-A",
  "message": {
    "content": [
      {
        "type": "tool_result",
        "tool_use_id": "toolu_123",
        "content": "total 48\ndrwxr-xr-x...",
        "is_error": false
      }
    ]
  },
  "toolUseResult": {
    "stdout": "total 48\ndrwxr-xr-x...",
    "stderr": "",
    "interrupted": false,
    "isImage": false
  }
}
```

**3. Assistant processes result:**
```json
{
  "type": "assistant",
  "uuid": "uuid-C",
  "parentUuid": "uuid-B",
  "message": {
    "content": [
      {
        "type": "text",
        "text": "I can see the directory contains..."
      }
    ]
  }
}
```

## Session Structure

Each session file represents a complete conversation and follows this structure:

**File naming:**
```
~/.claude/projects/<project-hash>/<session-uuid>.jsonl
```

**Typical session flow:**
1. `file-history-snapshot` - Initial file state
2. `user` - User prompt (parentUuid=null)
3. `user` - System metadata (e.g., command context)
4. `user` - Command output (if from slash command)
5. `file-history-snapshot` - File state before user's real prompt
6. `user` - User's actual prompt
7. `assistant` - Initial response
8. `assistant` - Tool use request (if needed)
9. `user` - Tool result
10. `assistant` - Final response
... [conversation continues] ...
N. `summary` - Session summary (at end)

## Version History

The `version` field tracks Claude Code version:
- Format: `"MAJOR.MINOR.PATCH"` (e.g., `"2.0.26"`)
- All entries in a session typically share the same version
- Version changes occur across sessions (after Claude Code updates)

## Usage Statistics

Assistant entries include detailed token usage:

```json
"usage": {
  "input_tokens": 32596,
  "cache_creation_input_tokens": 32596,
  "cache_read_input_tokens": 0,
  "cache_creation": {
    "ephemeral_5m_input_tokens": 32596,
    "ephemeral_1h_input_tokens": 0
  },
  "output_tokens": 127,
  "service_tier": "standard"
}
```

**Cache tiers:**
- `ephemeral_5m_input_tokens` - 5-minute cache
- `ephemeral_1h_input_tokens` - 1-hour cache

## Codex Native Records

Codex stores most data under a top-level `payload` object. meta-cc currently normalizes the common analysis events listed below.

### Session Metadata

```json
{
  "timestamp": "2026-06-14T06:00:00Z",
  "type": "session_meta",
  "payload": {
    "id": "codex-session-id",
    "cwd": "/absolute/project/path",
    "model": "gpt-5"
  }
}
```

`session_meta` provides session context for later records in the same file. It is not returned directly by message queries.

### Messages

```json
{
  "timestamp": "2026-06-14T06:00:01Z",
  "type": "response_item",
  "payload": {
    "type": "message",
    "role": "user",
    "content": [
      {"type": "input_text", "text": "Fix the parser"}
    ]
  }
}
```

Normalization:

- `payload.role == "user"` becomes a `type: "user"` entry with `message.content` as a string.
- `payload.role == "assistant"` becomes a `type: "assistant"` entry with text content blocks.
- `developer` and `system` messages become `type: "system"` entries.

### Duplicate Assistant/User Text (event_msg vs response_item)

Live Codex CLI 0.145 rollouts record the same logical user/assistant
utterance through **two independent channels** for the same turn:

- The legacy notification stream: `event_msg` with `payload.type` of
  `user_message` or `agent_message`.
- The transcript stream: `response_item` with `payload.type == "message"`
  and `payload.role` of `user` or `assistant` (see "Messages" above).

meta-cc reconciles these with a fixed precedence rule, scoped to
`(turn, role, position)`:

- `response_item` is the richer, canonical representation and **wins**
  whenever both channels report a segment at the same position within a
  turn.
- `event_msg` is used only as a **fallback**, for positions `response_item`
  never reported (e.g. older/event-only rollouts that predate the
  `response_item.message` transcript stream).
- Matching is **positional**, not content-based: the Nth `event_msg`
  segment for a role pairs with the Nth `response_item` segment for that
  role within the same turn. Text is never compared for equality, so two
  legitimately repeated messages — the same text sent again in a later
  turn, or two distinct segments within one turn that happen to share text
  — are never collapsed into one.

This keeps `query_session_content(provider=codex, role=assistant)` (and
`role=user`) returning each logical message exactly once instead of
doubling every segment. See `internal/provider/codex/rollout.go`'s
`dedupSegments` type for the implementation and
`tests/fixtures/codex/rollout-legacy-dedup-sample.jsonl` for a fixture
covering: same-position dedup, legitimate repeats across turns, and
multiple distinct segments within one turn.

### Newer Event Families (Codex 0.145+)

Besides the record types documented above, Codex 0.145 rollouts also emit:

- Top-level types: `world_state`, `compacted`, `turn_aborted`, `session_end`
- `response_item.payload.type` values: `tool_search_call`,
  `tool_search_output`
- `event_msg.payload.type` values: `thread_settings_applied`,
  `context_compacted`

meta-cc does not yet give these dedicated semantic handling. They fall
through to the existing "unknown event" fallback: each unrecognized record
is archived verbatim into the enclosing turn's
`Extensions` (`{"codex_events": [...]}`), so they are observable and never
silently dropped, but they don't yet produce typed fields (e.g.
`tool_search_call` does not currently become a `ToolCall`). Parsing must
not error or panic on any of them.

See `tests/fixtures/codex/rollout-legacy-0145-families-sample.jsonl` for a
sanitized fixture covering all of the above, and
`internal/provider/codex/rollout_test.go`'s
`TestLoadTurnsFromRollout0145EventFamilies` for the corresponding
assertions.

**Fixture-refresh procedure:** when a newer Codex CLI version changes or
introduces event shapes, capture a sanitized (no secrets, no repository
content, no unbounded tool output), minimal excerpt reproducing the new
shape, add it under `tests/fixtures/codex/`, and add/extend a test in
`internal/provider/codex/rollout_test.go` (parser-level) and, if the
change affects user-visible query output, in
`internal/mcp/executor/codex_dedup_e2e_test.go` or a sibling end-to-end
test (through `providerrecords.Build`/`query_session_content`) asserting
the parser still produces correct turns without data loss or duplication.

### Canonical Item Model (DIR-028)

Besides the flattened `UserText`/`AssistantText`/`ToolCalls` fields
described above, `conversation.Turn` (see `internal/conversation/types.go`
and `internal/conversation/item.go`) also carries an ordered `Items []Item`
stream and a `Status TurnStatus`. This is the **loss-minimizing internal
representation** (`Thread -> Turn -> Item`): provider adapters populate
Items with one entry per event they can classify, in true encounter order,
and the flattened fields are a **compatibility projection** computed *from*
Items (not by a second, independent parse) so existing MCP query output
(`query_session_content`, `query_session_signals`, etc.) is unaffected.

**Item kinds** (`conversation.ItemKind`): `user_message`, `agent_message`
(carries `Phase`: `commentary` | `final` | unspecified, derived from a
Harmony-style `channel` field on Codex `response_item`/`item.message`
events when present — never guessed when absent), `tool_call` /
`tool_result` (linked via `ToolCallID`, preserving item and turn identity
even though the compatibility projection merges them back into one
`ToolCall` entry), `command_execution`, `file_change`, `web_search`,
`plan_update`, `reasoning`, `compaction`, and `unknown` (a capped,
round-trippable catch-all — see `conversation.NewRawItem`, which truncates
any payload over 4KB rather than embedding it unbounded, and sets
`RawTruncated` when it does).

**Current adapter coverage** (intentionally scoped, DIR-028):

- **Codex** (`internal/provider/codex/rollout.go`): emits
  `user_message`/`agent_message`, `tool_call`/`tool_result`, and
  `reasoning` items with real parsing; the 0.145+ event families documented
  above (`tool_search_call`, `compacted`, `world_state`, etc.) are not yet
  given dedicated Item kinds and continue to fall through to `unknown`
  (mirroring their existing `Extensions.codex_events` fallback) — a future
  change can promote them to `command_execution`/`web_search`/`compaction`
  once their real-world payload shapes are confirmed.
- **Claude** (`internal/provider/claude/turns.go`'s `itemsFromPair`): a thin
  adapter over the existing `buildTurns`/`joinToolCalls` pipeline — it maps
  the already-computed flattened fields into `user_message`,
  `agent_message`, `tool_call`, and `tool_result` items. Claude's
  transcript format has no phase/channel signal, so `Phase` stays
  unspecified rather than guessed.

**Extension rule — adding a new Item kind:**

1. Add the `ItemKind` constant in `internal/conversation/item.go` with a
   doc comment describing what it represents, and add it to the coverage
   lists in `internal/conversation/item_test.go`'s
   `TestItemKindCoverage`.
2. In the provider adapter, populate the new kind as its own `Item` (do not
   overload an existing kind) and make sure it's appended to the turn's
   `Items` in true encounter order — do not reorder or batch during
   projection.
3. Update the adapter's legacy-field projection (`projectLegacyFields` in
   `internal/provider/codex/rollout.go`, or `itemsFromPair`'s caller in
   `internal/provider/claude/provider.go`) **only if** the new kind should
   also surface through the existing flattened fields; if it shouldn't
   (e.g. it has no analogue in the legacy message/tool schema), leave the
   projection untouched so existing query output does not change.
4. Add a parser-level test asserting order/fields (see
   `TestLoadTurnsFromRolloutPreservesItemOrderAndPhase` in
   `internal/provider/codex/rollout_test.go`), and, if the change is
   user-visible through `query_session_content`/`query_session_signals`,
   an end-to-end differential test through
   `providerrecords.Build`/the real MCP handler (see
   `TestQuerySessionContent_Codex_PhasedItemsProjectWithoutDuplicationOrReorder`
   in `internal/mcp/executor/codex_dedup_e2e_test.go`) proving the legacy
   projection is unchanged (no duplicated text, no reordered tool blocks).
5. Unrecognized/unparsed events must always fall back to
   `conversation.NewRawItem(conversation.ItemKindUnknown, ...)` rather than
   being silently dropped — see `appendUnknown` in
   `internal/provider/codex/rollout.go`.

There is currently no dedicated MCP query surface for the Item stream
itself (only the flattened compatibility projection is queryable); a
future `query_session_items`-style tool is a natural extension once a
concrete use case needs order/phase/item-kind fidelity at the query layer.

### Tool Calls

Codex function tool calls:

```json
{
  "timestamp": "2026-06-14T06:00:02Z",
  "type": "response_item",
  "payload": {
    "type": "function_call",
    "name": "exec_command",
    "call_id": "call_1",
    "arguments": "{\"cmd\":\"go test ./...\"}"
  }
}
```

Codex custom tool calls:

```json
{
  "timestamp": "2026-06-14T06:00:03Z",
  "type": "response_item",
  "payload": {
    "type": "custom_tool_call",
    "name": "apply_patch",
    "call_id": "call_2",
    "input": "*** Begin Patch\n*** End Patch"
  }
}
```

Both normalize to assistant `tool_use` content blocks. `arguments` is parsed as JSON when possible; otherwise raw text is preserved under `arguments` or `input`.

### Tool Outputs

```json
{
  "timestamp": "2026-06-14T06:00:04Z",
  "type": "response_item",
  "payload": {
    "type": "function_call_output",
    "call_id": "call_1",
    "output": "ok"
  }
}
```

`function_call_output` and `custom_tool_call_output` normalize to user `tool_result` content blocks. Non-success statuses and explicit error fields set `is_error: true`, which powers `query_tool_errors`, `analyze_errors`, and related analysis tools.

### Token Counts

Codex token usage is emitted as an `event_msg`:

```json
{
  "timestamp": "2026-06-14T06:00:05Z",
  "type": "event_msg",
  "payload": {
    "type": "token_count",
    "info": {
      "last_token_usage": {
        "input_tokens": 18818,
        "cached_input_tokens": 4480,
        "output_tokens": 152,
        "reasoning_output_tokens": 58,
        "total_tokens": 18970
      },
      "total_token_usage": {
        "input_tokens": 18818,
        "cached_input_tokens": 4480,
        "output_tokens": 152,
        "reasoning_output_tokens": 58,
        "total_tokens": 18970
      },
      "model_context_window": 258400
    }
  }
}
```

Normalization creates an assistant entry with `message.usage`, so `query_token_usage` works for Codex as well as Claude Code.

### Host-Specific Gaps

Some Claude Code records do not have Codex equivalents and therefore remain empty for Codex sessions:

- `file-history-snapshot` records used by `query_file_snapshots`
- top-level `summary` records used by `query_summaries`
- Claude Code `system` records with `subtype: "api_error"` used by `query_system_errors`

## Common Field Patterns

### UUID Fields
- **Format:** 8-4-4-4-12 hexadecimal (e.g., `"a151efcc-fd28-4aff-8552-03c805a197c8"`)
- **Usage:** `uuid`, `parentUuid`, `sessionId`, `messageId`, `leafUuid`, `tool_use_id`

### Timestamp Fields
- **Format:** ISO8601 with milliseconds and Z suffix
- **Example:** `"2025-10-24T14:07:36.078Z"`
- **Timezone:** Always UTC (Z suffix)

### Boolean Flags
Common boolean fields across record types:
- `isSidechain` - Branched conversation
- `isMeta` - System-generated metadata
- `isCompactSummary` - Compact summary format
- `isVisibleInTranscriptOnly` - Hidden from main view
- `isSnapshotUpdate` - Updates existing snapshot
- `is_error` - Tool execution failed
- `interrupted` - Tool execution interrupted
- `isImage` - Tool result is image data
- `isApiErrorMessage` - API error message

## Schema Validation Notes

When parsing JSONL session files:

1. **Type discrimination:** Always check `type` field first
2. **Host discrimination:** Codex records usually require checking `payload.type`
3. **Optional fields:** Not all optional fields present in every record
4. **Content polymorphism:** `message.content` can be string or array (check type)
5. **Null parents:** Only first Claude Code entry has `parentUuid=null`; Codex normalized UUIDs are synthesized
6. **Missing fields:** `file-history-snapshot` and `summary` lack standard fields
7. **Array vs Object:** Content blocks always in array for assistant messages

## Examples

See the following files for complete examples:
- **Claude Code session:** `~/.claude/projects/<project-hash>/*.jsonl`
- **Codex session:** `$CODEX_HOME/sessions/**/*.jsonl` or `~/.codex/sessions/**/*.jsonl`
- **Complex session:** Session files with tool executions, thinking blocks, and errors
- **Query examples:**
  - `docs/examples/jq-query-examples.md` - Single-file query patterns (19 examples)
  - `docs/examples/multi-file-jsonl-queries.md` - Multi-file queries with results (100 sample records)
  - `docs/examples/frequent-jsonl-queries.md` - Most frequently used queries (top 10 patterns)
  - `docs/examples/query-cookbook.md` - Practical query cookbook

## Related Documentation

- **JSONL Query Guide:** `docs/reference/jsonl.md` - Querying and filtering patterns
- **MCP Server Guide:** `docs/guides/mcp.md` - Querying via MCP tools
- **Unified Query API:** `docs/guides/unified-query-api.md` - Query interface
- **Repository Structure:** `docs/reference/repository-structure.md` - File organization

---

**Document Status:** Covers Claude Code native records and Codex records normalized by meta-cc
**Schema Coverage:** Claude Code message/tool records plus Codex message/tool/token records
**Last Updated:** 2026-07-27

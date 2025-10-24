# Claude Code Session JSONL Schema

This document describes the complete schema for Claude Code session history JSONL files stored in `~/.claude/projects/<project-hash>/`.

## Overview

Claude Code session files use **newline-delimited JSON (JSONL)** format, where each line is a complete, self-contained JSON record. Each record represents an event in the conversation history, forming a directed acyclic graph (DAG) through parent-child relationships.

**Key characteristics:**
- One JSON object per line
- Records linked via `uuid` and `parentUuid` fields
- Chronologically ordered by `timestamp`
- Five main record types: `user`, `assistant`, `system`, `file-history-snapshot`, `summary`

## Record Types

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
2. **Optional fields:** Not all optional fields present in every record
3. **Content polymorphism:** `message.content` can be string or array (check type)
4. **Null parents:** Only first entry has `parentUuid=null`
5. **Missing fields:** `file-history-snapshot` and `summary` lack standard fields
6. **Array vs Object:** Content blocks always in array for assistant messages

## Examples

See the following files for complete examples:
- **Simple session:** `~/.claude/projects/<project-hash>/*.jsonl` (small files)
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

**Document Status:** ✓ Validated against 95,259 records across multiple sessions
**Schema Coverage:** 5 record types, 3 content block types, 100% field coverage
**Last Updated:** 2025-10-24

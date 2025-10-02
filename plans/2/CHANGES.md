# Phase 2 Plan Updates - Real Claude Code Session Format

## Overview

The Phase 2 plan has been updated to support the **real Claude Code session format** instead of the simplified format originally assumed.

## Key Differences: Assumed vs Real Format

### Assumed Format (Original Plan)
```jsonl
{"sequence":0,"role":"user","timestamp":1735689600,"content":[{"type":"text","text":"..."}]}
```

### Real Format (Discovered)
```jsonl
{
  "type":"user",
  "timestamp":"2025-10-02T06:07:13.673Z",
  "message":{
    "role":"user",
    "content":[{"type":"text","text":"..."}]
  },
  "uuid":"cfef2966-a593-4169-9956-ee24c804b717",
  "parentUuid":null,
  "sessionId":"6a32f273-191a-49c8-a5fc-a5dcba08531a",
  "cwd":"/home/yale/work/meta-cc",
  "version":"2.0.1",
  "gitBranch":"develop"
}
```

## Major Structural Changes

### 1. Field Differences
| Aspect | Assumed Format | Real Format |
|--------|---------------|-------------|
| **Top-level role indicator** | `role` field | `type` field ("user", "assistant", etc.) |
| **Timestamp type** | `int64` (Unix timestamp) | `string` (ISO 8601: "2025-10-02T06:07:13.673Z") |
| **Message structure** | Flat (role + content at top level) | Nested (`message` object with role + content) |
| **Entry ordering** | `sequence` (integer) | `uuid` + `parentUuid` (string chain) |
| **Metadata** | Minimal | Rich (sessionId, cwd, gitBranch, version) |

### 2. Additional Entry Types
Real sessions include non-message entries that must be filtered:
- `file-history-snapshot` - File tracking metadata
- Other system events

Only `type: "user"` and `type: "assistant"` should be processed as messages.

## Updated Data Structures

### Stage 2.1: Data Structure Definition

#### SessionEntry (replaces Turn)
```go
type SessionEntry struct {
    Type       string   `json:"type"`        // "user", "assistant", "file-history-snapshot", etc.
    Timestamp  string   `json:"timestamp"`   // ISO 8601 format
    UUID       string   `json:"uuid"`        // Entry unique identifier
    ParentUUID string   `json:"parentUuid"`  // Parent entry UUID (message chain)
    SessionID  string   `json:"sessionId"`   // Session ID
    CWD        string   `json:"cwd"`         // Working directory
    Version    string   `json:"version"`     // Claude Code version
    GitBranch  string   `json:"gitBranch"`   // Git branch
    Message    *Message `json:"message"`     // Message content (only for user/assistant)
}

func (e *SessionEntry) IsMessage() bool {
    return e.Type == "user" || e.Type == "assistant"
}
```

#### Message (new nested structure)
```go
type Message struct {
    ID         string                 `json:"id"`         // Message ID (assistant only)
    Role       string                 `json:"role"`       // "user" or "assistant"
    Model      string                 `json:"model"`      // Model name (assistant only)
    Content    []ContentBlock         `json:"content"`    // Content blocks
    StopReason string                 `json:"stop_reason"` // Stop reason
    Usage      map[string]interface{} `json:"usage"`      // Token usage stats
}
```

### Stage 2.2: JSONL Reader Updates

#### ParseEntries (replaces ParseTurns)
```go
func (p *SessionParser) ParseEntries() ([]SessionEntry, error) {
    // ...parse JSONL...

    // Filter: only keep message types
    if entry.IsMessage() {
        entries = append(entries, entry)
    }
}
```

**Key changes**:
- Returns `[]SessionEntry` instead of `[]Turn`
- Automatically filters non-message types (file-history-snapshot, etc.)
- Handles nested `message` structure

### Stage 2.3: Tool Call Extraction Updates

#### ToolCall Structure
```go
type ToolCall struct {
    UUID     string                 // SessionEntry UUID (was TurnSequence)
    ToolName string                 // Tool name
    Input    map[string]interface{} // Tool input parameters
    Output   string                 // Tool output
    Status   string                 // Execution status
    Error    string                 // Error message
}
```

#### ExtractToolCalls
```go
func ExtractToolCalls(entries []SessionEntry) []ToolCall {
    // Step 1: Collect ToolUse from entry.Message.Content
    // Step 2: Collect ToolResult from entry.Message.Content
    // Step 3: Match and generate ToolCall array
}
```

**Key changes**:
- Accepts `[]SessionEntry` instead of `[]Turn`
- Uses `entry.UUID` instead of `turn.Sequence`
- Accesses content via `entry.Message.Content`
- Skips entries where `Message == nil`

## Test Updates

### Updated Test Fixture
`tests/fixtures/sample-session.jsonl` now contains real format:
```jsonl
{"type":"file-history-snapshot","messageId":"...","snapshot":{...}}
{"type":"user","timestamp":"2025-10-02T06:07:13.673Z","message":{...},"uuid":"..."}
{"type":"assistant","timestamp":"2025-10-02T06:08:57.769Z","message":{...},"uuid":"..."}
{"type":"user","timestamp":"2025-10-02T06:09:10.123Z","message":{...},"uuid":"..."}
```

**4 lines total**: 1 file-history-snapshot (filtered) + 3 messages (parsed)

### Test Changes Summary
1. **Stage 2.1 Tests** - Updated to use real JSON format with nested `message` structure
2. **Stage 2.2 Tests** - Added `TestParseSession_FilterNonMessageTypes` to verify filtering
3. **Stage 2.3 Tests** - Updated to use `SessionEntry` with `UUID` instead of `Turn` with `Sequence`

## Code Size Impact

| Component | Original Estimate | New Estimate | Change |
|-----------|------------------|--------------|--------|
| Stage 2.1 (types.go) | ~100 lines | ~140 lines | +40 lines |
| Stage 2.1 (tests) | ~90 lines | ~120 lines | +30 lines |
| Stage 2.2 (reader.go) | ~80 lines | ~90 lines | +10 lines |
| Stage 2.2 (tests) | ~110 lines | ~110 lines | (same) |
| Stage 2.3 (tools.go) | ~70 lines | ~75 lines | +5 lines |
| Stage 2.3 (tests) | ~100 lines | ~100 lines | (same) |
| **Total Implementation** | **~250 lines** | **~305 lines** | **+55 lines** |
| **Total with Tests** | **~550 lines** | **~635 lines** | **+85 lines** |

Still within Phase 2 budget (≤500 lines implementation code).

## Backward Compatibility

### Optional: Legacy Turn Structure
For backward compatibility with existing code that expects the simplified format, consider adding a helper:

```go
// ToSimplifiedTurn converts SessionEntry to a simplified Turn structure
func (e *SessionEntry) ToSimplifiedTurn() Turn {
    if !e.IsMessage() || e.Message == nil {
        return Turn{}
    }

    return Turn{
        UUID:      e.UUID,
        Role:      e.Message.Role,
        Timestamp: e.Timestamp,
        Content:   e.Message.Content,
    }
}
```

This was **not included** in the plan to keep it focused on the real format.

## Verification Checklist

Updated plan includes verification for:

- ✅ Parse real Claude Code session files from `~/.claude/projects/`
- ✅ Filter out non-message types (file-history-snapshot)
- ✅ Handle nested `message` structure
- ✅ Use UUID-based ordering instead of sequence numbers
- ✅ Extract metadata (sessionId, cwd, gitBranch)
- ✅ Parse ISO 8601 timestamps
- ✅ Match ToolUse and ToolResult across entries

## Migration Guide

If Phase 2 was already partially implemented with the old format:

1. **Rename structures**:
   - `Turn` → `SessionEntry`
   - `ParseTurns()` → `ParseEntries()`

2. **Update field access**:
   - `turn.Role` → `entry.Message.Role`
   - `turn.Content` → `entry.Message.Content`
   - `turn.Sequence` → `entry.UUID`
   - `turn.Timestamp` (int64) → `entry.Timestamp` (string)

3. **Add filtering**:
   ```go
   if !entry.IsMessage() {
       continue // skip non-message types
   }
   ```

4. **Update tool extraction**:
   - `toolCall.TurnSequence` → `toolCall.UUID`

## Files Changed

1. `/home/yale/work/meta-cc/plans/2/plan.md` - Complete plan rewrite with real format
2. `/home/yale/work/meta-cc/tests/fixtures/sample-session.jsonl` - Real format examples

## Next Steps

1. **Review updated plan** - Ensure all stages align with real format
2. **Re-implement Stage 2.1** - Use new SessionEntry structure
3. **Re-implement Stage 2.2** - Add filtering logic for non-message types
4. **Re-implement Stage 2.3** - Update to use UUID instead of Sequence
5. **Test with real sessions** - Validate against actual `~/.claude/projects/` files

## References

Real session file used for analysis:
- `/home/yale/.claude/projects/-home-yale-work-meta-cc/6a32f273-191a-49c8-a5fc-a5dcba08531a.jsonl`

Session format discovery date: **2025-10-02**

# Fixture: content-all-flow

## Fixture Metadata

```yaml
id: content-all-flow
tool: query_session_content
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "role": "all",
  "scope": "session",
  "provider": "claude"
}
```

## Expected Behavior

The tool queries the full conversation flow — all user and assistant records — for the current
session. Records are returned in chronological order, interleaved as they appear in the JSONL.
This is useful for reviewing the complete conversation context.

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | `"user"` or `"assistant"` |
| `timestamp` | string | ISO8601 format |
| `sessionId` | string | non-empty UUID |
| `message.role` | string | `"user"` or `"assistant"` |

## Pass/Fail Criteria

### Assertion AF1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain `type` (either "user" or "assistant"),
`timestamp`, `sessionId`, and `message.role`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion AF2 (Class A — Only User/Assistant)

**Oracle question**: Are there any records in the output where `type` is NOT `"user"` or `"assistant"`
(e.g., `"system"`, `"file-history-snapshot"`, etc.)?

**Pass**: Oracle answers NO (no non-user/assistant records present)

### Assertion AF3 (Class A — Role Consistency)

**Oracle question**: Does `type` always match `message.role`? (i.e., type="user" implies
message.role="user", type="assistant" implies message.role="assistant")

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion AF4 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{"type": "user", "timestamp": "2025-10-24T14:07:36.000Z", "sessionId": "abc123", "message": {"role": "user", "content": "Hello"}}
{"type": "assistant", "timestamp": "2025-10-24T14:07:40.000Z", "sessionId": "abc123", "message": {"role": "assistant", "content": [{"type": "text", "text": "Hi there!"}]}}
{"type": "user", "timestamp": "2025-10-24T14:07:45.000Z", "sessionId": "abc123", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "toolu_001", "content": "result"}]}}
```

## Notes

- User records with `message.content` as an array (tool results) are included in `role=all`.
- `role=user` only returns string-content user messages; `role=all` is more inclusive.
- `system` records are excluded from `role=all` output.

## Edge Cases

- Empty session (no messages): returns empty result set (not an error).
- Sessions with only tool-use turns still return those assistant records.

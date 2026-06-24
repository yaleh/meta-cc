# Fixture: signals-errors-provider-all

## Fixture Metadata

```yaml
id: signals-errors-provider-all
tool: query_session_signals
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "type": "errors",
  "scope": "project",
  "provider": "all"
}
```

## Expected Behavior

When `provider: "all"`, the tool queries both Claude Code and Codex providers and merges the
results. Each returned record must include a `provider` field identifying its source, in addition
to the standard error record fields.

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"user"` |
| `timestamp` | string | ISO8601 format |
| `sessionId` | string | non-empty UUID |
| `provider` | string | `"claude"` or `"codex"` |
| `message.content` | array | length ≥ 1, contains error tool_result |
| `message.content[].is_error` | boolean | `true` on at least one element |

## Pass/Fail Criteria

### Assertion PA1 (Class A — Field Presence with Provider)

**Oracle question**: Does every record in the output contain `timestamp`, `sessionId`, `provider`,
and at least one `message.content` element with `is_error: true`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion PA2 (Class A — Provider Field Values)

**Oracle question**: Is the `provider` field in every record either `"claude"` or `"codex"` (no
other values)?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion PA3 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{
  "type": "user",
  "timestamp": "2025-10-24T14:07:36.078Z",
  "sessionId": "abc123-...",
  "provider": "claude",
  "message": {
    "role": "user",
    "content": [
      {
        "type": "tool_result",
        "tool_use_id": "toolu_xxx",
        "is_error": true,
        "content": "Error: permission denied"
      }
    ]
  }
}
```

```json
{
  "type": "user",
  "timestamp": "2025-10-24T15:22:10.000Z",
  "sessionId": "codex-session-456",
  "provider": "codex",
  "message": {
    "role": "user",
    "content": [
      {
        "type": "tool_result",
        "tool_use_id": "func_yyy",
        "is_error": true,
        "content": "command failed"
      }
    ]
  }
}
```

## Notes

- `provider: "all"` adds the `provider` field to normalized records from both sources.
- Cross-provider queries may return different record shapes depending on normalization fidelity.
- If Codex has no error records, only Claude records appear in output (not an error).
- Working directory matching ensures only records from the current project are included.

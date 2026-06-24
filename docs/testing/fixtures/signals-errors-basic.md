# Fixture: signals-errors-basic

## Fixture Metadata

```yaml
id: signals-errors-basic
tool: query_session_signals
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "type": "errors",
  "scope": "project",
  "provider": "claude"
}
```

## Expected Behavior

The tool queries JSONL records for user entries that contain tool_result blocks with `is_error: true`.
Each returned record is a full JSONL user record containing at least one errored tool result in its
`message.content` array.

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"user"` |
| `timestamp` | string | ISO8601 format |
| `sessionId` | string | non-empty UUID |
| `message.content` | array | length ≥ 1 |
| `message.content[].type` | string | at least one element is `"tool_result"` |
| `message.content[].is_error` | boolean | at least one element is `true` |

## Pass/Fail Criteria

### Assertion A1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain `timestamp`, `sessionId`, and at least
one `message.content` element with `type: "tool_result"` and `is_error: true`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

**Fail**: Oracle answers NO, or confidence < 0.7

### Assertion A2 (Class A — No false positives)

**Oracle question**: Are there any records in the output where NO `message.content` element has
`is_error: true`?

**Pass**: Oracle answers NO (no false positives found)

**Fail**: Oracle answers YES (records without errors present in output)

### Assertion A3 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or valid JSONL (one record per line).

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{
  "type": "user",
  "timestamp": "2025-10-24T14:07:36.078Z",
  "sessionId": "abc123-...",
  "message": {
    "role": "user",
    "content": [
      {
        "type": "tool_result",
        "tool_use_id": "toolu_xxx",
        "is_error": true,
        "content": "Error: file not found"
      }
    ]
  }
}
```

## Edge Cases

- If no errors exist in the session, the tool MUST return empty results (not an error).
- Records with multiple content blocks where only some are errors are still valid (the error block is present).
- The `content` field within the tool_result may be a string or array.

# Fixture: content-assistant-basic

## Fixture Metadata

```yaml
id: content-assistant-basic
tool: query_session_content
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "role": "assistant",
  "scope": "session",
  "provider": "claude",
  "limit": 10
}
```

## Expected Behavior

The tool queries JSONL records for assistant entries. With `scope: "session"`, only the current
session is searched. With `limit: 10`, at most 10 records are returned.

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"assistant"` |
| `timestamp` | string | ISO8601 format |
| `sessionId` | string | non-empty UUID |
| `message.content` | array | array of content blocks |

## Pass/Fail Criteria

### Assertion A1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain `type: "assistant"`, `timestamp`,
`sessionId`, and a `message.content` field?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion A2 (Class C — Limit Respected)

**Programmatic check**: Number of returned records is ≤ 10.

**Pass**: `record_count ≤ 10`

### Assertion A3 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{
  "type": "assistant",
  "timestamp": "2025-10-24T14:07:40.000Z",
  "sessionId": "abc123-...",
  "message": {
    "role": "assistant",
    "content": [
      {
        "type": "text",
        "text": "I'll help you with that."
      }
    ],
    "usage": {
      "input_tokens": 500,
      "output_tokens": 50
    }
  }
}
```

## Variant: With `contains` Filter

### Input Parameters (variant)

```json
{
  "role": "assistant",
  "contains": "## Summary",
  "scope": "project",
  "provider": "claude"
}
```

### Additional Assertion A4 (Class B — Contains Match)

**Oracle question**: Does the stringified `message.content` of every returned record contain
`"## Summary"` (case-insensitive)?

**Pass**: Oracle answers YES with confidence ≥ 0.7

## Edge Cases

- Assistant records may have `message.content` as an empty array `[]` (valid).
- The `contains` filter uses case-insensitive matching (jq `test("..."; "i")`).
- `usage` field may or may not be present; it is not required.

# Fixture: signals-tokens-basic

## Fixture Metadata

```yaml
id: signals-tokens-basic
tool: query_session_signals
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "type": "tokens",
  "scope": "project",
  "provider": "claude",
  "stats_first": true
}
```

## Expected Behavior

The tool queries JSONL records for assistant entries that have a `message.usage` object. When
`stats_first: true`, the response begins with an aggregate statistics summary followed by individual
records.

## Required Output Fields

### Stats block (when stats_first=true)

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `total_input_tokens` | number | integer ≥ 0 |
| `total_output_tokens` | number | integer ≥ 0 |
| `record_count` | number | integer ≥ 0 |

### Per-record fields

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"assistant"` |
| `timestamp` | string | ISO8601 format |
| `sessionId` | string | non-empty UUID |
| `message.usage` | object | present, non-null |
| `message.usage.input_tokens` | number | integer ≥ 0 |
| `message.usage.output_tokens` | number | integer ≥ 0 |

## Pass/Fail Criteria

### Assertion T1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain `timestamp`, `sessionId`,
`type: "assistant"`, and a `message.usage` object with both `input_tokens` and `output_tokens`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion T2 (Class A — Token Values)

**Oracle question**: Are all `input_tokens` and `output_tokens` values non-negative integers?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion T3 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

Stats block:
```json
{
  "total_input_tokens": 45000,
  "total_output_tokens": 12000,
  "record_count": 30
}
```

Per-record:
```json
{
  "type": "assistant",
  "timestamp": "2025-10-24T14:07:40.123Z",
  "sessionId": "abc123-...",
  "message": {
    "usage": {
      "input_tokens": 1500,
      "output_tokens": 400,
      "cache_read_input_tokens": 800
    }
  }
}
```

## Edge Cases

- `cache_read_input_tokens` and `cache_creation_input_tokens` may be present but are optional.
- When no assistant messages have usage data, returns empty results (not an error).
- Codex provider may return per-turn usage only when rollout contains a `token_count` event.

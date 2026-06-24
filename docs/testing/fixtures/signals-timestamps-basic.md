# Fixture: signals-timestamps-basic

## Fixture Metadata

```yaml
id: signals-timestamps-basic
tool: query_session_signals
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "type": "timestamps",
  "scope": "session",
  "provider": "claude",
  "limit": 20
}
```

## Expected Behavior

The tool queries all JSONL records that have a non-null `timestamp` field. This is the broadest
signal type — it returns records of any type as long as they are timestamped. With `scope: "session"`
and `limit: 20`, only the current session and at most 20 records are returned.

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `timestamp` | string | ISO8601 format, non-null |
| `type` | string | any non-empty string |
| `sessionId` | string | non-empty UUID |

## Pass/Fail Criteria

### Assertion TS1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain a non-null `timestamp` in ISO8601 format,
a `type` field, and a `sessionId`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion TS2 (Class C — Limit Respected)

**Programmatic check**: Number of returned records is ≤ 20.

**Pass**: `record_count ≤ 20`

### Assertion TS3 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{
  "type": "user",
  "timestamp": "2025-10-24T14:07:36.078Z",
  "sessionId": "abc123-...",
  "message": { "role": "user", "content": "Hello" }
}
```

```json
{
  "type": "assistant",
  "timestamp": "2025-10-24T14:07:38.200Z",
  "sessionId": "abc123-...",
  "message": { "role": "assistant", "content": [] }
}
```

## Variant: Time-Range Filtered

### Input Parameters (variant)

```json
{
  "type": "timestamps",
  "scope": "project",
  "since": "2025-10-24T00:00:00Z",
  "until": "2025-10-24T23:59:59Z"
}
```

### Additional Assertion TS4 (Class C — Time Range Respected)

**Programmatic check**: Every returned record has `timestamp >= since` AND `timestamp <= until`.

**Pass**: All timestamps fall within [since, until]

## Edge Cases

- `file-history-snapshot` and `summary` record types may also appear since they carry timestamps.
- Records without `timestamp` field are excluded from results (not an error).

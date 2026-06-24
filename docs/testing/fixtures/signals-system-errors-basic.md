# Fixture: signals-system-errors-basic

## Fixture Metadata

```yaml
id: signals-system-errors-basic
tool: query_session_signals
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "type": "system_errors",
  "scope": "project",
  "provider": "claude"
}
```

## Expected Behavior

The tool queries JSONL records for system entries with `subtype: "api_error"`. These records represent
Claude API-level errors (e.g., overloaded, timeout, model errors) recorded by Claude Code. This is
a Claude Code-specific record type; Codex provider returns empty results.

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"system"` |
| `subtype` | string | exactly `"api_error"` |
| `timestamp` | string | ISO8601 format |
| `sessionId` | string | non-empty UUID |

## Pass/Fail Criteria

### Assertion SE1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain `type: "system"`, `subtype: "api_error"`,
`timestamp`, and `sessionId`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

**Note**: If the output is empty (no system errors found), this assertion passes by vacuous truth.

### Assertion SE2 (Class A — No Wrong Types)

**Oracle question**: Are there any records in the output where `type` is NOT `"system"` or
`subtype` is NOT `"api_error"`?

**Pass**: Oracle answers NO (no wrong-type records present)

### Assertion SE3 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{
  "type": "system",
  "subtype": "api_error",
  "timestamp": "2025-10-24T14:10:00.000Z",
  "sessionId": "abc123-...",
  "message": {
    "error": "overloaded_error",
    "message": "Claude is currently overloaded"
  }
}
```

## Provider Behavior

| Provider | Expected Result |
|----------|----------------|
| `claude` | Returns system error records (may be empty) |
| `codex` | Always returns empty (host-specific record type) |
| `all` | Returns Claude Code system errors only; Codex contributes nothing |

### Assertion SE4 (Class A — Codex empty)

**Input**: `{"type": "system_errors", "provider": "codex"}`

**Oracle question**: Is the result set empty (zero records)?

**Pass**: Oracle answers YES with confidence ≥ 0.7

## Edge Cases

- An empty result set is valid and expected when no API errors occurred.
- The `message` field content varies by error type; only `type`, `subtype`, `timestamp`, and `sessionId` are required.

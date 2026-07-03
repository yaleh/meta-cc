# Fixture: content-user-basic

## Fixture Metadata

```yaml
id: content-user-basic
tool: query_session_content
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "role": "user",
  "scope": "project",
  "provider": "claude"
}
```

## Expected Behavior

The tool queries JSONL records for user entries where `message.content` is a string (not an array).
This is the default behavior when no `pattern` or `content_type` is specified — the tool applies
`pattern: ".*"` to match all string user messages.

Tool result blocks (content_type=array) are excluded from this query.

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"user"` |
| `timestamp` | string | ISO8601 format |
| `sessionId` | string | non-empty UUID |
| `message.role` | string | exactly `"user"` |
| `message.content` | string | non-empty string |

## Pass/Fail Criteria

### Assertion U1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain `type: "user"`, `timestamp`, `sessionId`,
`message.role: "user"`, and a string `message.content`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion U2 (Class A — Content Type)

**Oracle question**: Is `message.content` in every returned record a string (not an array or null)?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion U3 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{
  "type": "user",
  "uuid": "uuid-001",
  "parentUuid": null,
  "timestamp": "2025-10-24T14:07:36.078Z",
  "sessionId": "abc123-...",
  "cwd": "/home/user/project",
  "message": {
    "role": "user",
    "content": "Please help me refactor this function"
  }
}
```

## Edge Cases

- System-injected messages (e.g., `<local-command-caveat>`) are included unless
  `exclude_system_messages: true` is passed.
- Empty user messages (empty string) may be returned; content_length filtering is available
  via `min_content_length`.
- Records where `message.content` is an array (tool results) are excluded from this query.

# Fixture: edge-limit-parameter

## Fixture Metadata

```yaml
id: edge-limit-parameter
tool: query_session_signals, query_session_content
decision_class: A
oracle: haiku
```

## Purpose

Verifies that the `limit` parameter correctly caps the number of returned records across both
consolidated query tools, and that results are still well-formed when limited.

---

## Case 1: query_session_signals with limit=1

### Input Parameters

```json
{
  "type": "tool_stats",
  "scope": "project",
  "provider": "claude",
  "limit": 1
}
```

### Expected Behavior

Returns at most 1 record. The record must be well-formed with all required fields.

### Pass/Fail Criteria

#### Assertion L1 (Class C — Limit Respected)

**Programmatic check**: `record_count <= 1`

**Pass**: Number of records in output is 0 or 1

#### Assertion L2 (Class A — Fields Present)

**Oracle question**: Is the returned record (if any) a complete, valid assistant record with
`type: "assistant"`, `timestamp`, `sessionId`, and at least one `tool_use` block with `name` and `id`?

**Pass**: Oracle answers YES with confidence ≥ 0.7 (or vacuously true if empty)

---

## Case 2: query_session_content with limit=5

### Input Parameters

```json
{
  "role": "user",
  "scope": "project",
  "provider": "claude",
  "limit": 5
}
```

### Expected Behavior

Returns at most 5 user records. Records must be well-formed.

### Pass/Fail Criteria

#### Assertion L3 (Class C — Limit Respected)

**Programmatic check**: `record_count <= 5`

**Pass**: Number of records in output is 0 to 5

#### Assertion L4 (Class A — Fields Present)

**Oracle question**: Are all returned records well-formed user records with `type: "user"`,
`timestamp`, `sessionId`, and string `message.content`?

**Pass**: Oracle answers YES with confidence ≥ 0.7 (or vacuously true if empty)

---

## Case 3: limit=0 (no limit — default behavior)

### Input Parameters

```json
{
  "type": "tokens",
  "scope": "session",
  "provider": "claude"
}
```

Note: `limit` is omitted (default = no limit).

### Expected Behavior

Returns all matching records in the session without truncation. The response may use `file_ref`
mode if results exceed 8KB.

### Pass/Fail Criteria

#### Assertion L5 (Class A — All Records or file_ref)

**Oracle question**: Does the output either contain multiple complete records OR a `file_ref`
pointing to a readable file (for large outputs)?

**Pass**: Oracle answers YES with confidence ≥ 0.7

---

## Notes

- `limit: 0` and omitted `limit` both mean "no limit" per the parameter schema.
- When `limit` truncates a result, the response may include `truncated: true` and
  `total_events: N` metadata fields (implementation-dependent).
- Limit applies to the final output record count after filtering.

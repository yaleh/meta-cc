# Fixture: edge-empty-results

## Fixture Metadata

```yaml
id: edge-empty-results
tool: query_session_signals, query_session_content
decision_class: A
oracle: haiku
```

## Purpose

Verifies that both consolidated query tools handle zero-match scenarios gracefully — returning
an empty result set without errors, rather than failing or returning malformed output.

---

## Case 1: query_session_signals with impossible time range

### Input Parameters

```json
{
  "type": "errors",
  "scope": "project",
  "provider": "claude",
  "since": "2099-01-01T00:00:00Z",
  "until": "2099-12-31T23:59:59Z"
}
```

### Expected Behavior

No records will exist in the year 2099. The tool must return a valid empty result, not an error.

### Pass/Fail Criteria

#### Assertion ER1 (Class A — Empty is Valid)

**Oracle question**: Does the output indicate an empty result set (zero records) without an error
message or exception? The output may be an empty array, `{"records": []}`, `{"total": 0}`,
or an empty JSONL file.

**Pass**: Oracle answers YES with confidence ≥ 0.7

#### Assertion ER2 (Class C — No Error Status)

**Programmatic check**: Response does not contain an HTTP error code or exception stack trace.

**Pass**: No error structure detected

---

## Case 2: query_session_content with non-matching pattern

### Input Parameters

```json
{
  "role": "user",
  "pattern": "XYZZY_IMPOSSIBLE_PATTERN_12345",
  "scope": "project",
  "provider": "claude"
}
```

### Expected Behavior

The regex `XYZZY_IMPOSSIBLE_PATTERN_12345` is extremely unlikely to match any real user message.
The tool must return a valid empty result.

### Pass/Fail Criteria

#### Assertion ER3 (Class A — Empty is Valid)

**Oracle question**: Does the output indicate an empty result set without an error or exception?

**Pass**: Oracle answers YES with confidence ≥ 0.7

---

## Case 3: query_session_signals on Codex for Claude-only type

### Input Parameters

```json
{
  "type": "system_errors",
  "scope": "project",
  "provider": "codex"
}
```

### Expected Behavior

`system_errors` is a Claude Code-specific record type. The Codex provider MUST return empty results
rather than failing.

### Pass/Fail Criteria

#### Assertion ER4 (Class A — Provider Graceful Empty)

**Oracle question**: Does the output for a Codex provider querying `system_errors` return an empty
result set (not an error)?

**Pass**: Oracle answers YES with confidence ≥ 0.7

---

## General Empty Result Format

An acceptable empty result looks like any of:

```json
[]
```
```json
{"records": [], "total": 0}
```
```json
{"count": 0, "results": []}
```
(empty JSONL file / file_ref pointing to empty file)

The key requirement: **valid parseable output with zero records, no error fields at the top level**.

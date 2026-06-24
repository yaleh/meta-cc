# GCL Gate Annotation Format

This guide defines a structured text annotation convention for **Gate Criterion Ledger (GCL)** events in meta-cc workflows. Annotations are embedded in assistant messages and are parseable by existing `query_session_content` pattern matching — no new MCP tools are required.

## Overview

GCL gate annotations allow agents and humans to record self-assessment decisions at named workflow gates. Each gate captures whether a criterion was **Empirically verified** ([E]), **Circumstantially inferred** ([C]), or **Hypothetically assumed** ([H]).

These annotations are intended for:
- Baseline measurement of decision quality across sessions
- Retrospective analysis with `query_session_content`
- Identifying patterns where assumptions ([H]) were later proven wrong

---

## GCL Gate Annotation Format

### Structure

A gate annotation block consists of three parts:

```
[GCL-gate] <decision-name>
[E|C|H] <criterion>: <evidence-description>
...
GCL-self-report: E=<n> C=<n> H=<n>
```

### Field Definitions

| Field | Description |
|-------|-------------|
| `[GCL-gate] <decision-name>` | Header line opening a gate block. `<decision-name>` identifies the decision point (e.g., `freshnessCheck`, `readyToMerge`). |
| `[E] <criterion>: <evidence>` | Empirically verified — criterion was confirmed by direct observation, test output, or tool result. |
| `[C] <criterion>: <evidence>` | Circumstantially inferred — criterion was inferred from indirect evidence without direct confirmation. |
| `[H] <criterion>: <evidence>` | Hypothetically assumed — criterion was assumed without supporting evidence; highest epistemic risk. |
| `GCL-self-report: E=n C=n H=n` | Summary line closing the block. Counts must match the criterion lines above. |

### Rules

1. A gate block MUST start with a `[GCL-gate]` header line.
2. Each criterion line MUST be tagged exactly `[E]`, `[C]`, or `[H]`.
3. The block MUST close with a `GCL-self-report:` summary line.
4. The summary counts MUST match the actual count of `[E]`, `[C]`, `[H]` lines above.
5. Multiple gate blocks may appear in a single assistant message.

---

## Examples

### Example 1: freshnessCheck gate (2E, 1C, 1H)

```
[GCL-gate] freshnessCheck
[E] fileExists: Read tool confirmed /path/to/file.go is present and non-empty
[E] testsPass: make test returned exit code 0 with 0 failures
[C] dependenciesUpToDate: go.sum was recently modified, inferred deps are current
[H] noBreakingChanges: assumed upstream API is stable; changelog not checked
GCL-self-report: E=2 C=1 H=1
```

### Example 2: readyToMerge gate (3E, 0C, 1H)

```
[GCL-gate] readyToMerge
[E] allTestsPass: CI reported green on all 47 test cases
[E] lintClean: make lint exited 0 with no warnings
[E] prApproved: gh pr view shows 1 approval from required reviewer
[H] noRegressions: manual smoke test not performed; assumed covered by unit tests
GCL-self-report: E=3 C=0 H=1
```

### Example 3: phaseGate — pre-implementation checkpoint (1E, 2C, 1H)

```
[GCL-gate] phaseGate-implementation
[E] planApproved: backlog task status is Ready per task_view output
[C] scopeUnderstood: requirements inferred from task description; no clarification requested
[C] technicalRiskLow: similar feature implemented in TASK-11; risk assumed comparable
[H] estimateAccurate: 3h estimate based on intuition; no breakdown performed
GCL-self-report: E=1 C=2 H=1
```

---

## Querying GCL Events

### Basic Pattern Search

Use `query_session_content` to retrieve all messages containing gate annotations:

```javascript
query_session_content({role: "assistant", pattern: "GCL-self-report"})
```

This returns all assistant messages that contain a completed GCL gate block.

To search for a specific gate decision:

```javascript
query_session_content({role: "assistant", pattern: "\\[GCL-gate\\] freshnessCheck"})
```

To find all messages with hypothetical assumptions:

```javascript
query_session_content({role: "assistant", pattern: "\\[H\\]"})
```

### Extracting E/C/H Counts with Two-Stage Query

Use `get_session_directory` + `execute_stage2_query` to extract structured counts from matched content:

```javascript
// Step 1: Locate session files
const dir = await get_session_directory({scope: "project"})

// Step 2: Filter assistant messages with GCL gate annotations
const results = await execute_stage2_query({
  files: dir.files,
  filter: 'select(.type == "assistant") | select(.message.content != null) | select(.message.content | tostring | test("GCL-self-report"))',
  transform: '{
    timestamp,
    sessionId,
    gcl_summary: (.message.content | tostring | capture("GCL-self-report: E=(?P<E>[0-9]+) C=(?P<C>[0-9]+) H=(?P<H>[0-9]+)"))
  }'
})
```

### Aggregate H-Rate Across All Gates

To measure the proportion of hypothetical assumptions across all recorded gate events:

```javascript
// Step 1: Get session directory
const dir = await get_session_directory({scope: "project"})

// Step 2: Aggregate all GCL summary lines
const agg = await execute_stage2_query({
  files: dir.files,
  filter: 'select(.type == "assistant") | select(.message.content | tostring | test("GCL-self-report"))',
  transform: '(.message.content | tostring | scan("GCL-self-report: E=([0-9]+) C=([0-9]+) H=([0-9]+)"))'
})
// Parse the resulting arrays: [E_count, C_count, H_count] per gate
```

---

## Integration with meta-cc Workflow

GCL annotations are designed to be:

1. **Lightweight**: Written inline in assistant messages, no separate tool calls needed.
2. **Queryable**: `query_session_content` with `pattern: "GCL-self-report"` retrieves all gate events.
3. **Auditable**: The `[E]/[C]/[H]` classification captures epistemic confidence at decision time.

### Recommended Placement

Place gate annotation blocks:
- At the end of a phase or stage completion message
- Before a significant irreversible action (e.g., commit, merge, deploy)
- When making assumptions that affect correctness

### Baseline Measurement

To establish a GCL baseline for a project:

```javascript
// Count total gates recorded
query_session_content({role: "assistant", pattern: "GCL-gate"})

// Find high-risk gates (any H criterion)
query_session_content({role: "assistant", pattern: "\\[H\\].*\\nGCL-self-report"})
```

---

## See Also

- [MCP Server Guide](mcp.md) — Complete MCP tool reference
- [MCP Query Tools Reference](mcp-query-tools.md) — `query_session_content` API
- [Two-Stage Query Guide](two-stage-query-guide.md) — Custom jq workflows

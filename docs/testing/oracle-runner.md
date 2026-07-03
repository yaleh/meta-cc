# Layer 2.5 Oracle Runner

## Overview

This document describes how to run the Layer 2.5 oracle fixture suite against a live MCP session.
It defines the Haiku oracle prompt template, the scoring procedure, and the pass threshold.

**Framework**: BAIME Exp-D Layer 2.5
**Oracle model**: claude-haiku-latest (or claude-haiku-4-5)
**Pass threshold**: ≥0.85 overall pass rate on Class A assertions

---

## Prerequisites

1. meta-cc MCP server is running and configured in Claude Code (see [MCP Guide](../guides/mcp.md))
2. A live Claude Code session exists with session history available
3. Access to invoke MCP tools directly or through a test harness
4. Fixture files are available at `docs/testing/fixtures/`

---

## Running a Fixture

### Step 1: Load the Fixture

Read the fixture file to obtain:
- Input parameters (JSON object)
- Required output fields (table)
- Pass/Fail criteria (assertion list)

Example fixture: `fixtures/signals-errors-basic.md`

### Step 2: Invoke the MCP Tool

Call the tool with the fixture's input parameters:

```
Tool: query_session_signals
Parameters: {"type": "errors", "scope": "project", "provider": "claude"}
```

Capture the raw JSON response.

### Step 3: Apply Class C Assertions (Programmatic)

Before involving the oracle, run programmatic checks:

- **Valid JSON**: Parse the response. If parse fails → FAIL immediately.
- **Limit respected**: If `limit` was set, count records and verify `record_count <= limit`.
- **Time range respected**: If `since`/`until` were set, verify all timestamps fall within range.
- **Empty is valid**: If result is empty and query should legitimately be empty → PASS.

### Step 4: Invoke Haiku Oracle (Class A assertions)

For each Class A assertion in the fixture, submit the following prompt to Haiku:

---

#### Haiku Oracle Prompt Template

```
You are a JSON structure validator. You will be given:
1. A description of required fields
2. A JSON output from an MCP tool

Your task: Answer YES or NO to the assertion question, followed by a confidence score from 0.0 to 1.0.

Format your response exactly as:
ANSWER: YES
CONFIDENCE: 0.95
REASON: <one sentence>

--- ASSERTION ---
{assertion_question}

--- REQUIRED FIELDS ---
{required_fields_list}

--- TOOL OUTPUT (first 3 records shown) ---
{tool_output_sample}
```

---

**Variables to substitute**:

- `{assertion_question}` — the oracle question from the fixture's assertion block
- `{required_fields_list}` — bullet list of field names and constraints from the fixture
- `{tool_output_sample}` — first 3 records of the tool output (or full output if ≤ 3 records)

**If output uses file_ref**: Read the referenced file and use first 3 lines as the sample.

### Step 5: Parse Oracle Response

Parse the oracle's response:

```
ANSWER: YES  →  oracle_answer = true
ANSWER: NO   →  oracle_answer = false
CONFIDENCE: 0.95  →  oracle_confidence = 0.95
```

**Assertion result**:
- PASS if `oracle_answer == true AND oracle_confidence >= 0.7`
- FAIL if `oracle_answer == false OR oracle_confidence < 0.7`

### Step 6: Apply Class B Assertions (Semantic, Haiku oracle)

Class B assertions use the same oracle prompt template but with a semantic question about
filter correctness. Use a sample of 5 records from the output.

---

## Scoring

### Per-Fixture Score

```
fixture_pass_count = number of assertions that PASS
fixture_total_count = total number of assertions
fixture_score = fixture_pass_count / fixture_total_count
```

A fixture **passes** if `fixture_score >= 1.0` (all assertions pass).
A fixture **partially passes** if `0.0 < fixture_score < 1.0`.
A fixture **fails** if `fixture_score == 0.0`.

### Overall Suite Score

Run all fixtures and compute:

```
suite_pass_count = number of fixtures that pass (fixture_score == 1.0)
suite_total_count = total number of fixtures
overall_accuracy = suite_pass_count / suite_total_count
```

**Required threshold**: `overall_accuracy >= 0.85`

With 14 current fixtures, this means at least 12 fixtures must fully pass.

### Weighted Scoring (optional)

For weighted accuracy, assign weights by decision class:
- Class A assertions: weight 1.0 (most important — field presence)
- Class B assertions: weight 0.8 (semantic correctness)
- Class C assertions: weight 0.5 (programmatic, less oracle-dependent)

```
weighted_score = sum(weight * pass) / sum(weight)
```

---

## Fixture Index

| Fixture ID | Tool | Type/Role | Decision Class | Assertion Count |
|------------|------|-----------|----------------|-----------------|
| signals-errors-basic | query_session_signals | errors | A | 3 |
| signals-tokens-basic | query_session_signals | tokens | A | 3 |
| signals-tool-stats-basic | query_session_signals | tool_stats | A+B | 4 |
| signals-timestamps-basic | query_session_signals | timestamps | A+C | 4 |
| signals-system-errors-basic | query_session_signals | system_errors | A | 4 |
| signals-errors-provider-all | query_session_signals | errors (all) | A | 3 |
| content-user-basic | query_session_content | user | A | 3 |
| content-user-pattern | query_session_content | user+pattern | A+B | 4 |
| content-assistant-basic | query_session_content | assistant | A+B | 4 |
| content-tool-blocks | query_session_content | tool | A | 4 |
| content-all-flow | query_session_content | all | A | 4 |
| edge-empty-results | both | empty | A | 4 |
| edge-limit-parameter | both | limit | A+C | 5 |
| edge-file-ref-mode | both | file_ref | A+C | 4 |

**Total assertions**: ~57

---

## Reporting

After running the full suite, produce a report in this format:

```markdown
## Layer 2.5 Oracle Run Report

Date: <ISO8601>
Meta-cc version: <version>
Oracle model: claude-haiku-latest

### Results

| Fixture | Score | Status |
|---------|-------|--------|
| signals-errors-basic | 3/3 | PASS |
| signals-tokens-basic | 3/3 | PASS |
| ... | ... | ... |

### Overall Accuracy

suite_pass_count: X / 14
overall_accuracy: X.XX

Threshold: 0.85
Outcome: PASS / FAIL
```

---

## Troubleshooting

### Oracle answers with low confidence

If oracle confidence is consistently < 0.7, the oracle may lack enough context.
Increase `{tool_output_sample}` from 3 to 10 records and retry.

### Oracle gives inconsistent answers

Run the assertion 3 times and take majority vote. Report the confidence as the average.

### Tool returns error instead of empty

If `query_session_signals` or `query_session_content` throws an error for legitimate
zero-match queries, this is a bug in meta-cc. File an issue and mark the relevant fixture
as FAIL with note "tool error on empty query".

### file_ref file not readable

If the `file_ref` path does not exist or is not readable, mark assertion FR3 as FAIL.
This likely means the temp file was cleaned up prematurely by `cleanup_temp_files`.

---

## Extending the Suite

To add a new fixture:
1. Create `docs/testing/fixtures/<id>.md` following the fixture template structure
2. Add a row to the Fixture Index table above
3. Increment `suite_total_count` in the scoring formula
4. Update the pass threshold count accordingly (floor(suite_total * 0.85))

---

## See Also

- [Layer 2.5 λ-spec](layer25-spec.md) — Formal tool specifications
- [Fixture Collection](fixtures/) — All fixture files
- [MCP Query Tools Reference](../guides/mcp-query-tools.md) — Tool parameter reference

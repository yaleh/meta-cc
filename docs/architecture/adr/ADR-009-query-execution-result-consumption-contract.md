---
id: ADR-009
title: "Query Execution and Result Consumption Contract"
status: proposed
date: 2026-07-30
---
# ADR-009: Query Execution and Result Consumption Contract

## Status

Proposed

## Context

meta-cc exposes session-history queries to host models through MCP. Most calls
are self-contained, but historical Claude Code and Codex usage shows recurring
friction at two boundaries:

1. Expert Stage 2 jq queries fail when callers cannot discover the actual input
   type or field path.
2. Large results correctly use ADR-004's `file_ref` mode, but the reference
   exposes too little nested shape information for callers to consume the
   result without exploratory Read/Bash/jq calls.

Some Bash activity after a meta-cc query is appropriate. A historical finding
often needs to be checked against the current source tree, configuration, or
build state. That is a separate activity from consuming the query result.
Bash used only to reverse-engineer meta-cc's result shape or reproduce common
projection and grouping operations indicates a missing query contract.

The metadata-driven query design also permits callers to construct jq queries
over raw provider files. ADR-003, meanwhile, describes MCP as the common data
access layer. The project needs a precise boundary that preserves expert
extensibility without making shell composition the normal path for supported
queries.

## Decision

### MCP Is the Supported Query Boundary

MCP is the supported product boundary for discovering, executing, and
diagnosing session-history queries. Provider files, indexes, and temporary
result files remain implementation or advanced-analysis surfaces; callers
should not need their internal layout for ordinary supported queries.

The MCP boundary must:

- validate declared parameters before every dispatch path;
- preserve provider and source provenance;
- report bounded completeness and fallback diagnostics;
- keep inline and `file_ref` representations logically equivalent; and
- expose stable result-shape information sufficient to consume a response.

### Common Operations Belong Server-Side

Filtering, projection, grouping, pagination, and summary operations that are
common, repeatable, and provider-neutral should be available through typed MCP
parameters or stable response contracts.

This does not require adding every possible jq expression as a first-class
parameter. A server-side operation is justified when it reduces repeated
caller-side shape discovery, can be validated, and has semantics that remain
coherent across supported providers.

### Stage 2 Remains an Expert Interface

`execute_stage2_query` and raw jq remain supported escape hatches for
compositions that do not justify a first-class query parameter.

The expert interface must provide:

- provider-specific schema and representative bounded samples;
- preflight feedback for common type/path mismatches;
- actionable errors rather than silent coercion or query rewriting; and
- the same provenance and completeness diagnostics as applicable typed
  queries.

### `file_ref` Is a Transport Mode

ADR-004's `file_ref` remains the fallback for results that should not be
returned inline. It is a transport decision, not a different logical result
model.

A file reference should be self-describing enough for normal selective
consumption. Its metadata should evolve toward:

- versioned nested field paths and value types;
- bounded, redacted structured samples;
- recipes generated and tested against the emitted shape; and
- pagination and query diagnostics consistent with inline mode.

Read/Grep/Bash/jq remain valid for selective access and advanced analysis, but
they are not a substitute for documenting the result contract.

### Diagnostics Are Part of Query Correctness

Applicable query responses should expose a uniform, bounded account of:

- requested and effective provider/backend;
- sessions or files considered, loaded, and skipped;
- fallback or degraded execution;
- match count and returned-record count; and
- warnings needed to distinguish no matches from incomplete search.

Diagnostics must not disclose unbounded errors, secrets, or content beyond the
query's existing authorization and output contract.

### Separate Historical Evidence from Current-State Inspection

meta-cc owns the retrieval and packaging of historical evidence. The host agent
may then use repository tools to inspect current source or configuration. That
follow-up is expected and should not be treated as a meta-cc query failure.

Evidence bundles may package bounded excerpts, query criteria, provenance, and
completeness for issue/task workflows, but they must remain separate from
mutating the current project.

## Consequences

### Positive

- Callers can distinguish empty results from incomplete or degraded searches.
- Large results remain token-efficient without requiring result-shape guessing.
- Common queries become provider-neutral and easier to validate.
- jq retains its flexibility for expert use.
- Shell operations following a query can be interpreted by purpose rather than
  counted as undifferentiated product friction.

### Negative

- Query envelopes and `file_ref` metadata become richer and require versioning.
- Server-side projection and diagnostics add implementation and test surface.
- Provider adapters must supply enough provenance and completeness data for a
  coherent cross-provider contract.
- Maintaining both typed queries and an expert jq interface requires explicit
  compatibility tests.

### Risks

- Over-expanding typed parameters could duplicate jq and complicate the API.
  Mitigation: require demonstrated repeated use and provider-neutral semantics.
- Samples or diagnostics could expose sensitive transcript content.
  Mitigation: bounded redaction rules and security-focused fixtures.
- Different providers could report misleadingly equivalent diagnostics.
  Mitigation: include effective backend and explicitly label unavailable or
  estimated fields.
- Result metadata could drift from the emitted data.
  Mitigation: derive schema/recipes from actual result structures and validate
  them in repository gates.

## Implementation

Implementation is tracked separately:

- `DIR-079`: Stage 2 preflight, safe session-file inspection, uniform query
  diagnostics, and schema/runtime drift gates.
- `DIR-080`: typed nested `file_ref` metadata, bounded samples, generated jq
  recipes, server-side projection/grouping, and evidence bundles.
- `DIR-081`: evolve ADR-001/003 toward the provider-neutral MCP-only
  architecture using this query boundary.

Progressive streaming or a new MCP transport contract is not decided here. It
requires measurement and a future ADR amendment if pursued.

## Related Decisions

- [ADR-001](ADR-001-two-layer-architecture.md) - Original separation of
  deterministic data processing from LLM interpretation
- [ADR-003](ADR-003-mcp-server-integration.md) - MCP integration strategy
- [ADR-004](ADR-004-hybrid-output-mode.md) - Inline and `file_ref` transport
- [ADR-005](ADR-005-scope-parameter-standardization.md) - Query scope
- [ADR-007](ADR-007-provenance-data-source.md) - Measured and estimated data
  provenance

## References

- [Metadata-Driven Query Architecture](../metadata-driven-query-architecture.md)
- [MCP Query Tools](../../guides/mcp-query-tools.md)
- [Two-Stage Query Guide](../../guides/two-stage-query-guide.md)
- [DIR-004](../../../tasks/DIR-004.md) - Pagination and `file_ref` fallback
- [DIR-031](../../../tasks/DIR-031.md) - FTS candidate search and hydration

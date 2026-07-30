---
id: ADR-007
title: "DataSource Provenance Convention (measured / estimated)"
status: accepted
date: 2026-07-28
---
# ADR-007: DataSource Provenance Convention (measured / estimated)

## Status

Accepted

## Context

Every analysis result produced by `internal/analyzer` carries a `DataSource`
provenance field, defined in `internal/analyzer/data_source.go`. The field is a
string enum with exactly two values:

| Value | Constant | Meaning |
|-------|----------|---------|
| `"measured"` | `DataSourceMeasured` | The value was computed by directly counting or aggregating observable events from the session trace (tool calls, entries, timestamps). High confidence; no inferential leap. |
| `"estimated"` | `DataSourceEstimated` | The value was inferred via a heuristic rule rather than directly observed. Lower confidence; treat as approximate. |

The field is emitted in the output of every analysis MCP tool (`analyze_errors`,
`analyze_bugs`, `quality_scan`, `get_tech_debt`, `get_timeline`,
`get_work_patterns`) and is consumed by this project's own autonomous loop, which
treats `measured` data as ground truth and `estimated` data as an approximate
signal. Provenance is therefore a **consumer-trust contract**: a consumer decides
how much to trust a number based on whether it was measured or estimated.

Despite being cross-cutting, the convention was undocumented (the ADR index stopped
at ADR-006, and the "BAIME Layer 7" phrase in the original doc comment was defined
nowhere). Worse, it had unresolved **mixed-provenance collapses**. Several result
types ship heuristic subfields while advertising `measured` at the top level:

- `TechDebtResult`: session `open_issues` uses causal absence; source-scan `markers` and `hotspot_files` use a language-aware lexical scanner rather than a parser.
- `WorkPatternsResult.context_switches`: file-path changes within a five-minute window.
- `QualityScanResult.dimensions`: the array contains `retry_rate`, where a later same-tool call within five positions is treated as a retry.
- `BugAnalysisResult` / `BugAnalysisStats`: a same-tool success within three positions is treated as the causal fix.
- `ErrorAnalysisResult.by_type` / `ErrorAnalysisStats.by_type`: labels and signatures are normalized by classification heuristics.
- `EditSequencesResult`: file/document roles, pattern hints, documentation-gap signals, and their distribution are threshold classifications derived from observed activity.

With no written rule and no enforcement, every new analyzer could re-decide this ad
hoc, and consumers had no reliable, machine-readable way to know which fields were
heuristic.

Note the separate, orthogonal `Provenance` field on `FileDebt`
(`session` | `source` | `both`, added by DIR-055) records *which scan bucket*
produced a per-file entry. That is a different axis from `DataSource` and is not
governed by this ADR.

## Decision

1. **Every serialized analysis result type carries a `DataSource` field.** Any
   exported `*Result` / `*Stats` struct in `internal/analyzer` that is emitted as
   a serialized analysis output MUST carry a `DataSource DataSource` field (JSON
   key `data_source`). This includes aggregate/statistics and edit-sequence
   outputs such as `TimelineStats` and `EditSequencesResult`; age or subsystem
   boundaries are not exemptions. Truly internal, non-serialized aggregates
   (for example `SessionStats`) and unexported helpers are outside the contract.
   `measured` means directly counted/aggregated from observable session events;
   `estimated` means heuristic inference (canonical examples: open-issue
   detection, context-switch proximity — see `data_source.go`).

2. **Mixed-provenance rule — dominant provenance + `estimated_fields` list.** When a
   result mixes both provenances, the top-level `DataSource` records the **dominant**
   provenance (the one covering the majority of the result's fields), and the result
   carries an `EstimatedFields []string` field (JSON key `estimated_fields`,
   `omitempty`) listing the JSON names of the heuristic-derived fields. A purely
   `measured` or purely `estimated` result leaves `EstimatedFields` empty/omitted.

   Every governed result declaration includes the optional field so adding a
   heuristic later cannot require a second, manually maintained type registry.
   Current mixed outputs populate it as follows:
   - `TechDebtResult`: session scans list `open_issues`; source-directory scans list `markers` and `hotspot_files`; merged results union these paths.
   - `WorkPatternsResult`: `context_switches`.
   - `QualityScanResult`: `dimensions` (the serialized array containing the heuristic retry-rate element).
   - `BugAnalysisResult`: `patterns`, `total_pairs`.
   - `BugAnalysisStats`: `patterns`, `total_pairs`, `total_patterns`.
   - `ErrorAnalysisResult` / `ErrorAnalysisStats`: `by_type`.
   - `EditSequencesResult`: heuristic descendant paths under `files.*` plus
     `summary.patternDistribution`. Optional wildcard paths are generated from
     the concrete result: for example, source-only output does not advertise
     absent `docRole` or `coAccessedDocs` descendants. A wildcard path means the
     path exists for every matched map entry or array element, not merely one.
     Every serialized heuristic field must be advertised even in empty results:
     an empty `EditSequencesResult` still serializes the heuristic
     `summary.patternDistribution` zeros, so it advertises exactly that path and
     omits only the absent optional file/event paths.
   Result-to-stats transformations must remap paths to the stats JSON shape;
   `TechDebtStats` uses `markers` plus `marker_count` for source-derived marker
   output, `hotspot_file_count` for source-derived hotspots, and
   `open_issue_count` for session-derived open issues rather than copying
   result-only `hotspot_files` or `open_issues` paths.

   **Why option (b) and not per-field annotations (option a):** the codebase already
   used dominant-provenance informally in both ad hoc comments, so formalizing it is
   the smallest, least disruptive change. A single optional `estimated_fields` list
   keeps the common pure-`measured` case free of per-field noise (the field is
   omitted entirely), adds no new wrapper types, and is trivially machine-checkable:
   a consumer that trusts only measured data reads `data_source` and, when it is
   `measured`, downgrades exactly the fields named in `estimated_fields`. Per-field
   annotation wrappers would have changed the shape of every field on every result —
   a far larger surface than an optional list of heuristic paths.

3. **Enforcement.** An AST-based test
   (`internal/analyzer/provenance_enforcement_test.go`) parses the analyzer
   package and mechanically enumerates every exported struct whose name ends in
   `Result` or `Stats`. It asserts each declaration has a `DataSource` field of
   type `analyzer.DataSource` with JSON key `data_source`. Explicit exemptions
   are limited to justified, truly internal non-serialized types (currently
   `SessionStats`). A new serialized result that omits provenance therefore
   fails `make commit` without requiring a registry update. A companion assertion
 requires every governed declaration to carry the optional
   `EstimatedFields []string` field with JSON key `estimated_fields`; this avoids
   a partial registry of only the mixed types. Runtime and JSON tests then verify
   that current heuristic implementations populate the correct paths.

4. **Registration.** ADR-007 is registered in `docs/architecture/adr/README.md`
   alongside ADR-001..006, and DIR-009's registration scope is extended to include
   it.

## Consequences

### Positive Impacts

- **Documented trust contract**: consumers (human and the autonomous loop) have a
  single written definition of `measured` vs `estimated` and can calibrate trust.
- **Machine-readable mixed provenance**: `estimated_fields` lets a consumer that
  trusts only measured data precisely downgrade the heuristic subfields, instead of
  having to read prose comments per struct.
- **One rule, applied uniformly**: the two existing ad hoc inline policies are
  replaced by a single cited convention; new analyzers inherit it rather than
  re-inventing it.
- **Enforced**: the conformance test prevents new analyzers from silently dropping
  provenance.

### Negative Impacts

- **A second field to maintain**: mixed results must keep `estimated_fields` in sync
  with their heuristic fields. A field that becomes heuristic later must be added to
  the list (the test enforces presence of `DataSource`, not the accuracy of
  `estimated_fields`).
- **Mechanical declaration coverage**: the AST test discovers newly declared
  exported `*Result` / `*Stats` structs automatically, so conformance does not
  depend on a manually maintained governed-type registry.

### Risks

- **Stale `estimated_fields`**: if a field's implementation changes provenance
  without updating the list, consumers may over- or under-trust it. Mitigation: code
  review + the convention documented here; the list lives next to the field it
  describes.
- **Dominant-provenance judgment calls**: a result that is exactly half measured /
  half estimated has no obvious dominant value. Mitigation: such a result should use
  `estimated` as the conservative top-level value and still list its estimated
  fields; this is the fail-safe direction for an autonomous consumer.

## Implementation

- [x] `internal/analyzer/data_source.go` defines `DataSource`, `DataSourceMeasured`, `DataSourceEstimated` (pre-existing).
- [x] `TechDebtResult` gains `EstimatedFields`: `GetTechDebt` surfaces `["open_issues"]`, `ScanSourceDir` surfaces `["markers", "hotspot_files"]` for its lexical scanner, and merges union the lists.
- [x] `WorkPatternsResult` gains `EstimatedFields` and cites this ADR; `GetWorkPatterns` surfaces `["context_switches"]`.
- [x] `TimelineStats` and `EditSequencesResult` carry measured provenance in
  their constructors and serialized JSON, and both are in the enforcement registry.
- [x] `internal/analyzer/provenance_enforcement_test.go` mechanically discovers exported `*Result` / `*Stats` declarations and asserts each non-exempt type has a `DataSource` field.
- [x] `docs/architecture/adr/README.md` index lists ADR-007.
- [x] `tasks/DIR-009.md` notes its registration scope now includes ADR-007.

## Related Decisions

- [ADR-001](ADR-001-two-layer-architecture.md) - Two-Layer Architecture Design (the data-processing layer that emits these measured/estimated results)

## Notes

Consumer trust levels for autonomous-loop consumption:

- `measured`, field not in `estimated_fields` → **ground truth**; safe to gate decisions on.
- `measured`, field in `estimated_fields` → **approximate**; use as a signal, not a gate.
- `estimated` (top-level) → **approximate overall**; treat the whole result as a heuristic signal.

The enforcement test's red/green behavior is verified by temporarily removing a
`DataSource` field from a governed result type and observing the test fail naming
that type, then restoring it.

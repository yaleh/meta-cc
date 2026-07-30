---
id: ADR-010
title: "Provider-Neutral MCP-Only Architecture"
status: accepted
date: 2026-07-30
supersedes:
  - ADR-001
  - ADR-003
---
# ADR-010: Provider-Neutral MCP-Only Architecture

## Status

Accepted

Supersedes [ADR-001](ADR-001-two-layer-architecture.md) and
[ADR-003](ADR-003-mcp-server-integration.md) to the extent described in
[Preserved vs. Superseded Decisions](#preserved-vs-superseded-decisions) below.
The historical context and decision text of both records are preserved
unchanged; this record is the explicit evolution decision.

## Context

ADR-001 and ADR-003 were accepted on 2025-10-10 and described the architecture
as it was then designed. The shipped system has since evolved past several of
their specifics:

1. **The standalone CLI is gone.** ADR-001 defines Layer 1 as a standalone
   `meta-cc` CLI tool. Phase 26 removed the legacy CLI code (roughly 19,500
   lines) and established an MCP-only product architecture. The only shipped
   executable today is the MCP server (`cmd/mcp-server`); there is no CLI
   entrypoint in the tree.
2. **The system is no longer Claude-centric.** ADR-001 frames Layer 2 as
   "Claude Code Integration" and ADR-003 targets Claude Code alone. The current
   system normalizes conversation history from multiple hosts through a
   provider-neutral conversation model (`internal/conversation`), with
   first-class Claude and Codex provider implementations
   (`internal/provider/claude`, `internal/provider/codex`) selected through a
   common provider registry and a `provider` parameter with host-default
   resolution.
3. **Volatile inventories aged badly.** ADR-003 fixes a tool inventory and a
   three-tier integration model (MCP / slash commands / subagents) with
   use-case coverage percentages. The tool surface has since changed, and the
   skills and agents described there have moved out of this repository
   (meta-cc 3.0.0 focuses exclusively on session history analysis via MCP
   tools, plus prompt-library commands/skills). Those numbers were descriptive,
   never a stable architectural contract, and they no longer describe the
   product.
4. **"MCP is the only data-access layer" is too absolute.** The metadata-driven
   query design and [ADR-009](ADR-009-query-execution-result-consumption-contract.md)
   deliberately permit controlled consumption through `file_ref` results and
   expert jq/Bash composition. ADR-009 defines that boundary precisely.

Editing ADR-001 and ADR-003 in place would erase why the original decisions
were made. The project instead needs an explicit evolution decision that states
what endures and what is superseded.

## Decision

meta-cc's architecture is a **provider-neutral, MCP-only two-layer
architecture**:

### Layer 1 — Deterministic Data Processing (MCP server)

All session-history parsing, indexing, filtering, projection, statistics, and
rule-based analysis are deterministic local processing with no LLM calls. This
work lives in the MCP server binary (`cmd/mcp-server`) and the `internal/`
packages. The MCP server is the **only shipped executable**; there is no
required standalone CLI layer.

### Layer 2 — Semantic Interpretation (host model/workflow layer)

Interpretation of query results — meaning, recommendations, coaching — happens
in the host model and workflow layer (Claude Code, Codex, or a future host).
This layer is outside this repository. meta-cc supplies deterministic evidence;
the host supplies meaning.

### Provider-Neutral Contract, Host Adapters as Implementations

The supported conversation-history abstraction is the provider contract
(`internal/provider.Provider`: availability, session listing/retrieval, and
turn loading into provider-neutral `internal/conversation` records), composed
through a provider registry. Claude and Codex are **implementations** of this
contract, not special cases baked into the core. A new host joins by
implementing the same contract and registering it.

Host-facing surfaces (Claude Code slash commands, Codex skills, plugin and
marketplace packaging) are **integration implementations** of the host layer,
not architectural tiers with their own data paths. They consume the MCP
surface.

### MCP Is the Supported Product and Query Boundary

MCP is the supported boundary for discovering, executing, and diagnosing
session-history queries, consistent with
[ADR-009](ADR-009-query-execution-result-consumption-contract.md). Raw
provider files, indexes, and temporary result files are implementation details
and advanced-analysis surfaces. The controlled escape hatches — Stage 2 jq
queries and `file_ref` results consumed with Read/Grep/Bash/jq — remain
supported for expert use exactly as ADR-009 and
[ADR-004](ADR-004-hybrid-output-mode.md) define them.

This refines, rather than contradicts, ADR-003's "single source of truth"
intent: MCP is the supported product boundary, while expert file/jq consumption
is a deliberate, documented escape hatch — not a second product surface.

### Authoritative References Replace Inventories

This record deliberately contains no fixed tool list, integration-surface
inventory, or use-case percentages. The authoritative current references are:

- **MCP tools and usage**: [MCP Guide](../../guides/mcp.md) and
  [MCP Query Tools Reference](../../guides/mcp-query-tools.md)
- **Integration surfaces per host**: [Integration Guide](../../guides/integration.md)
- **Repository layout**: [Repository Structure](../../reference/repository-structure.md)
- **Provider specifics**: [Codex History Model](../../reference/codex-history-model.md)
  and [Codex App-Server Backend](../../reference/codex-app-server.md)
- **Query consumption boundary**:
  [ADR-009](ADR-009-query-execution-result-consumption-contract.md) and
  [Two-Stage Query Guide](../../guides/two-stage-query-guide.md)

Where an earlier ADR stated an inventory or percentage, that statement is
historical description, not an architectural requirement.

## Preserved vs. Superseded Decisions

### From ADR-001 (Two-Layer Architecture Design)

**Preserved**:

- The core separation: deterministic local data processing (no LLM calls) is
  architecturally distinct from LLM-driven semantic interpretation.
- Data flows upward as structured records; the data layer is testable without
  an LLM, and semantic understanding stays in the integration/host layer.
- The cost, performance, and testability rationale for keeping LLM calls out of
  the data layer.

**Superseded**:

- Layer 1's *form* as a standalone `meta-cc` CLI binary. The legacy CLI was
  removed (Phase 26); the MCP server is the sole executable, and no CLI layer
  is required or shipped.
- The Claude-specific framing of Layer 2 ("Claude Code Integration"). Layer 2
  is now a provider-neutral host model/workflow layer with Claude Code and
  Codex as implementations.
- Implementation notes that fix a tool count or name a `cmd/server.go` layout
  (the binary is `cmd/mcp-server`; the current tool surface is documented in
  the [MCP Guide](../../guides/mcp.md)).

### From ADR-003 (MCP Server Integration Strategy)

**Preserved**:

- MCP is the primary integration method and the supported data-access boundary;
  host-facing surfaces are thin consumers of MCP, not parallel data paths.
- The maintainability rationale: a single deterministic data layer behind the
  integration surfaces.

**Superseded**:

- The fixed tool inventory. The current tool set is documented in the
  [MCP Guide](../../guides/mcp.md); it evolves and is not an ADR-level
  contract.
- The three-tier model (MCP / slash commands / subagents) with use-case
  coverage percentages as architectural structure. The described skills and
  subagents no longer ship from this repository, and no percentage is an
  architectural requirement. Host integration surfaces are documented in the
  [Integration Guide](../../guides/integration.md).
- The absolute statement that MCP is the *only* data-access layer. ADR-009
  defines the precise boundary, including the controlled `file_ref`/jq/Bash
  escape hatch.
- The Claude-only framing; the integration contract is provider-neutral.

## Consequences

### Positive Impacts

- The ADR set now describes the shipped product: one MCP binary, a
  provider-neutral core, host adapters as implementations.
- Adding a host means implementing one provider contract and host packaging —
  the architecture no longer implies Claude-only assumptions.
- Volatile inventories and percentages are replaced by links to living,
  authoritative references, so the ADRs stop drifting on every tool change.
- The enduring decisions (layer separation, MCP as supported boundary,
  semantics in the host) are reaffirmed explicitly rather than implicitly.
- ADR-001/003 remain readable as historical records with their original
  rationale intact.

### Negative Impacts

- Readers must follow supersession links: ADR-001/003 are only correct when
  read together with this record.
- The provider contract becomes a stated architectural commitment; changing
  `internal/provider.Provider` semantics now warrants its own decision record.
- Documentation discipline is required: the authoritative references this ADR
  points at must stay current, since inventories are no longer duplicated here.

### Risks

- **Reference drift.** Mitigation: the linked guides are maintained with the
  code, and repository gates keep documentation and implementation consistent.
- **Escape hatch over-read.** Readers may treat the ADR-009 jq/`file_ref`
  escape hatch as a second product surface. Mitigation: ADR-010 and ADR-009
  both state that MCP remains the supported boundary and expert paths are
  controlled exceptions.
- **Contract ambiguity.** "Host adapters as implementations" could be read to
  demand identical feature parity across hosts. Mitigation: the contract is the
  provider interface and the MCP surface; host packaging may differ, with
  provenance diagnostics (per ADR-007 and ADR-009) labeling what each backend
  actually measured.

## Implementation

This record documents an architecture that is already in place:

- [x] Single MCP-only binary at `cmd/mcp-server` (legacy CLI removed in
      Phase 26; no CLI entrypoint remains).
- [x] Provider-neutral conversation model in `internal/conversation` with
      `claude` and `codex` provider IDs.
- [x] Provider contract and registry in `internal/provider` with Claude and
      Codex implementations (`internal/provider/claude`,
      `internal/provider/codex`).
- [x] Provider selection via `provider: "claude" | "codex" | "all"` with
      host-default resolution (`META_CC_HOST`).
- [x] Host integration surfaces documented in the Integration Guide (Claude
      Code plugin/commands, Codex plugin/skills); analysis skills/agents moved
      out of this repository.
- [x] ADR-001 and ADR-003 marked `superseded` with `superseded_by: ADR-010`;
      their historical text unchanged.

## Related Decisions

- [ADR-001](ADR-001-two-layer-architecture.md) — Superseded here (layer
  separation preserved; CLI form superseded)
- [ADR-003](ADR-003-mcp-server-integration.md) — Superseded here (MCP as
  supported boundary preserved; inventories, tiers, percentages, and
  Claude-only framing superseded)
- [ADR-004](ADR-004-hybrid-output-mode.md) — Active; hybrid output remains the
  transport decision, governed by the ADR-009 consumption contract
- [ADR-005](ADR-005-scope-parameter-standardization.md) — Active; scope
  standardization remains in force under the MCP boundary
- [ADR-007](ADR-007-provenance-data-source.md) — Active; measured/estimated
  provenance applies per provider backend
- [ADR-009](ADR-009-query-execution-result-consumption-contract.md) — Defines
  the query execution and result-consumption boundary this record relies on,
  including the controlled `file_ref`/jq escape hatch

## Notes

### Design Principle (carried forward)

ADR-001's principle survives, generalized:

> "meta-cc extracts data deterministically; the host model extracts meaning."

- If it is **data extraction, filtering, or analysis** → deterministic
  MCP-server layer (Layer 1), provider-neutral.
- If it is **semantic understanding** → host model/workflow layer (Layer 2),
  outside this repository.

### References

- [MCP Guide](../../guides/mcp.md) — authoritative current tool reference
- [Integration Guide](../../guides/integration.md) — host integration surfaces
- [Repository Structure](../../reference/repository-structure.md) — current
  layout of binaries and packages
- [Implementation Plan](../../core/plan.md) — Phase 26 CLI removal and
  MCP-only architecture status

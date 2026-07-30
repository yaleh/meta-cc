# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for the meta-cc project.

## What are ADRs?

ADRs are documents that capture important architectural decisions made along with their context and consequences. They help:

1. **Document decision process** - Why we made this decision
2. **Provide context** - The problem and constraints we faced
3. **Track evolution** - How architecture changes over time
4. **Knowledge transfer** - Help new team members understand historical decisions

## ADR Format

We follow the standard ADR format proposed by Michael Nygard:

- **Status** - Proposed | Accepted | Deprecated | Superseded
- **Context** - The issue motivating this decision
- **Decision** - The change we're proposing or have agreed to
- **Consequences** - What becomes easier or harder due to this decision
- **Implementation** - Implementation status or plan (optional)
- **Related Decisions** - Links to related ADRs (optional)
- **Notes** - Additional information, links, diagrams (optional)

Each ADR file also carries YAML frontmatter (`id`, `title`, `status`, `date`) so
the quay-native provider can register it as a first-class ADR object
(`QUAY_NATIVE_ADR_DIR` in `.quay/config.yml` points at this directory).

## ADR Lifecycle Convention

Architecture decisions use the **ADR (decision) lifecycle**:
`proposed → accepted → superseded | deprecated` (or `rejected`). This is NOT the
task/directive lifecycle (`todo → ready → done`): an ADR is a standing decision
that is continuously applied, never "done". ADR status values are validated
against the quay-native ADR store (`proposed`, `accepted`, `superseded`,
`deprecated`, `rejected`).

Query registered ADRs with `quay adr list` / `quay adr get ADR-NNN` (or the MCP
`adr_list` / `adr_get` tools); update them with `quay adr write`.

## Active ADRs

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-001](ADR-001-two-layer-architecture.md) | Two-Layer Architecture Design | Accepted | 2025-10-10 |
| [ADR-002](ADR-002-plugin-directory-structure.md) | Plugin Directory Structure Refactoring | Accepted | 2025-10-10 |
| [ADR-003](ADR-003-mcp-server-integration.md) | MCP Server Integration Strategy | Accepted | 2025-10-10 |
| [ADR-004](ADR-004-hybrid-output-mode.md) | Hybrid Output Mode Design | Accepted | 2025-10-10 |
| [ADR-005](ADR-005-scope-parameter-standardization.md) | Scope Parameter Standardization | Accepted | 2025-10-10 |
| [ADR-006](ADR-006-pkg-vs-internal-convention.md) | pkg/ vs internal/ Directory Convention | Accepted | 2026-03-10 |
| [ADR-007](ADR-007-provenance-data-source.md) | DataSource Provenance Convention (measured / estimated) | Accepted | 2026-07-28 |
| [ADR-008](ADR-008-workflow-skill-api-isolation.md) | Workflow Scripts Have No Claude Code Tool API (One-Way Workflow/Skill Bridge) | Accepted | 2026-07-30 |
| [ADR-009](ADR-009-query-execution-result-consumption-contract.md) | Query Execution and Result Consumption Contract | Proposed | 2026-07-30 |

## Creating New ADRs

1. Copy the [template](template.md)
2. Use sequential numbering (ADR-010, ADR-011, ...)
3. Place in this directory (`docs/architecture/adr/`)
4. Add the required YAML frontmatter (`id`, `title`, `status`, `date`) — start with `status: proposed`
5. Update this index
6. Submit for review if significant

## References

- [Michael Nygard's ADR format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR GitHub Organization](https://adr.github.io/)

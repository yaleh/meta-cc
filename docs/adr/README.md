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

## Active ADRs

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-001](ADR-001-two-layer-architecture.md) | Two-Layer Architecture Design | Accepted | 2025-10-10 |
| [ADR-002](ADR-002-plugin-directory-structure.md) | Plugin Directory Structure Refactoring | Accepted | 2025-10-10 |
| [ADR-003](ADR-003-mcp-server-integration.md) | MCP Server Integration Strategy | Accepted | 2025-10-10 |
| [ADR-004](ADR-004-hybrid-output-mode.md) | Hybrid Output Mode Design | Accepted | 2025-10-10 |
| [ADR-005](ADR-005-scope-parameter-standardization.md) | Scope Parameter Standardization | Accepted | 2025-10-10 |

## Creating New ADRs

1. Copy the [template](template.md)
2. Use sequential numbering (ADR-006, ADR-007, ...)
3. Place in `docs/adr/` directory
4. Update this index
5. Submit for review if significant

## References

- [Michael Nygard's ADR format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR GitHub Organization](https://adr.github.io/)

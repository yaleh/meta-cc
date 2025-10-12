# meta-cc Documentation Map

This document provides a visual overview of documentation dependencies and navigation guide.

## Documentation Dependency Graph

```mermaid
graph TD

  %% Core Entry Points
  CLAUDE_md["CLAUDE.md<br/>(Main Entry)"]:::entry
  README_md["README.md<br/>(Public Doc)"]:::entry

  %% Key Guides
  docs_plan_md["plan.md<br/>(Roadmap)"]:::guide
  docs_principles_md["principles.md<br/>(Design Rules)"]:::guide
  docs_integration_guide_md["integration-guide.md"]:::guide
  docs_mcp_guide_md["mcp-guide.md<br/>(MCP Complete)"]:::guide

  %% Architecture
  docs_adr_README_md["ADR Index"]:::adr
  docs_proposals_md["Proposals"]:::adr

  %% Dependencies - Core Entry Points
  CLAUDE_md --> docs_principles_md
  CLAUDE_md --> docs_plan_md
  CLAUDE_md --> docs_mcp_guide_md
  CLAUDE_md --> docs_integration_guide_md

  README_md --> docs_mcp_guide_md
  README_md --> docs_integration_guide_md
  README_md --> docs_examples_usage_md

  %% Dependencies - Guides
  docs_principles_md --> docs_adr_README_md
  docs_plan_md --> docs_adr_README_md
  docs_plan_md --> docs_proposals_md
  docs_integration_guide_md --> docs_examples_usage_md
  docs_mcp_guide_md --> docs_integration_guide_md

  %% Styles
  classDef entry fill:#e8f5e9,stroke:#4caf50,stroke-width:3px
  classDef guide fill:#e3f2fd,stroke:#2196f3,stroke-width:2px
  classDef adr fill:#fff3e0,stroke:#ff9800,stroke-width:2px
```

## Quick Navigation Guide

### For New Users

1. **Start**: [README.md](../README.md) - Installation and overview
2. **Setup MCP**: [docs/mcp-guide.md](mcp-guide.md) - Complete MCP setup and usage
3. **Examples**: [docs/examples-usage.md](examples-usage.md) - Step-by-step tutorials

### For Claude Code Development

1. **Entry Point**: [CLAUDE.md](../CLAUDE.md) - Development workflow
2. **Design Rules**: [docs/principles.md](principles.md) - Core constraints
3. **Roadmap**: [docs/plan.md](plan.md) - Phase-by-phase plan
4. **Architecture**: [docs/adr/README.md](adr/README.md) - ADR index

### For Integration Work

1. **Integration Guide**: [docs/integration-guide.md](integration-guide.md) - Choosing MCP/Slash/Subagent
2. **MCP Complete Guide**: [docs/mcp-guide.md](mcp-guide.md) - All MCP topics in one place
3. **Troubleshooting**: [docs/troubleshooting.md](troubleshooting.md) - Common issues

## Document Roles

| Document | Role | Target Audience | Update Frequency |
|----------|------|----------------|------------------|
| **CLAUDE.md** | Development entry point | Claude Code | Every phase |
| **README.md** | Public documentation | End users | Major releases |
| **docs/plan.md** | Roadmap and status | Developers | Continuous |
| **docs/principles.md** | Design constraints | Developers | Rarely (stable) |
| **docs/mcp-guide.md** | MCP complete reference | Users & Developers | As MCP evolves |
| **docs/integration-guide.md** | Integration decisions | Advanced users | Stable |
| **docs/examples-usage.md** | Step-by-step tutorials | New users | When features added |
| **docs/adr/** | Architecture decisions | Architects | Per decision |

## Most Accessed Documents (from meta-cc analysis)

| Rank | Document | Access Count | Primary Use Case |
|------|----------|--------------|------------------|
| 1 | docs/plan.md | 411 | Phase tracking, implementation planning |
| 2 | README.md | 159 | Project overview, quick start |
| 3 | docs/principles.md | 88 | Design constraints, architecture rules |
| 4 | CLAUDE.md | 62 | Development workflow entry point |
| 5 | docs/examples-usage.md | 62 | Setup tutorials, usage examples |

## Documentation Optimization (Phase 23)

This documentation structure was optimized to:

1. **Reduce redundancy**: Consolidated 4 MCP documents → 1 comprehensive guide
2. **Improve navigation**: Created CLAUDE.md quick links and FAQ section
3. **Simplify completed phases**: Moved detailed phase docs to plans/ directory
4. **Enhance discoverability**: Added this navigation map

See the optimization plan in the commit message for details.

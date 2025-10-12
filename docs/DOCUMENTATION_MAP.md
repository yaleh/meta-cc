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

  %% Reference Docs
  docs_cli_reference_md["cli-reference.md<br/>(CLI Commands)"]:::reference
  docs_jsonl_reference_md["jsonl-reference.md<br/>(Output Format)"]:::reference
  docs_features_md["features.md<br/>(Advanced Features)"]:::reference

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
  README_md --> docs_cli_reference_md
  README_md --> docs_features_md

  %% Dependencies - Guides
  docs_principles_md --> docs_adr_README_md
  docs_plan_md --> docs_adr_README_md
  docs_plan_md --> docs_proposals_md
  docs_integration_guide_md --> docs_examples_usage_md
  docs_mcp_guide_md --> docs_integration_guide_md

  %% Dependencies - Reference
  docs_cli_reference_md --> docs_jsonl_reference_md
  docs_cli_reference_md --> docs_mcp_guide_md
  docs_features_md --> docs_cli_reference_md
  docs_features_md --> docs_mcp_guide_md

  %% Styles
  classDef entry fill:#e8f5e9,stroke:#4caf50,stroke-width:3px
  classDef guide fill:#e3f2fd,stroke:#2196f3,stroke-width:2px
  classDef reference fill:#f3e5f5,stroke:#9c27b0,stroke-width:2px
  classDef adr fill:#fff3e0,stroke:#ff9800,stroke-width:2px
```

## Quick Navigation Guide

### For New Users

1. **Start**: [README.md](../README.md) - Quick install and overview
2. **Setup MCP**: [docs/mcp-guide.md](mcp-guide.md) - Complete MCP setup and usage
3. **Examples**: [docs/examples-usage.md](examples-usage.md) - Step-by-step tutorials
4. **Troubleshooting**: [docs/troubleshooting.md](troubleshooting.md) - Common issues

### For Advanced Users

1. **CLI Reference**: [docs/cli-reference.md](cli-reference.md) - Complete command reference
2. **JSONL Reference**: [docs/jsonl-reference.md](jsonl-reference.md) - Output format and jq patterns
3. **Features**: [docs/features.md](features.md) - Advanced capabilities
4. **CLI Composability**: [docs/cli-composability.md](cli-composability.md) - Unix pipeline patterns

### For Claude Code Development

1. **Entry Point**: [CLAUDE.md](../CLAUDE.md) - Development workflow
2. **Design Rules**: [docs/principles.md](principles.md) - Core constraints
3. **Roadmap**: [docs/plan.md](plan.md) - Phase-by-phase plan
4. **Architecture**: [docs/adr/README.md](adr/README.md) - ADR index

### For Integration Work

1. **Integration Guide**: [docs/integration-guide.md](integration-guide.md) - Choosing MCP/Slash/Subagent
2. **MCP Complete Guide**: [docs/mcp-guide.md](mcp-guide.md) - All MCP topics in one place
3. **Capabilities Guide**: [docs/capabilities-guide.md](capabilities-guide.md) - Create custom capabilities

## Document Roles

| Document | Role | Target Audience | Update Frequency |
|----------|------|----------------|------------------|
| **CLAUDE.md** | Development entry point (simplified) | Claude Code | Every phase |
| **README.md** | Public documentation (simplified) | End users | Major releases |
| **docs/plan.md** | Roadmap and status | Developers | Continuous |
| **docs/principles.md** | Design constraints | Developers | Rarely (stable) |
| **docs/plugin-development.md** | Plugin development workflow | Plugin developers | When workflow changes |
| **docs/repository-structure.md** | Directory organization guide | Developers | Rarely (stable) |
| **docs/unified-meta-command.md** | /meta command complete guide | Users & Developers | When /meta evolves |
| **docs/mcp-guide.md** | MCP complete reference | Users & Developers | As MCP evolves |
| **docs/integration-guide.md** | Integration decisions | Advanced users | Stable |
| **docs/release-process.md** | Release workflow | Maintainers | Rarely (stable) |
| **docs/git-hooks.md** | Git hooks usage | Developers | Rarely (stable) |
| **docs/examples-usage.md** | Step-by-step tutorials | New users | When features added |
| **docs/cli-reference.md** | Complete CLI command reference | Advanced users | When commands added |
| **docs/jsonl-reference.md** | Output format and jq patterns | Advanced users | Rarely (stable) |
| **docs/features.md** | Advanced features overview | Advanced users | When features added |
| **docs/adr/** | Architecture decisions | Architects | Per decision |

## Most Accessed Documents (from meta-cc analysis)

| Rank | Document | Access Count | Primary Use Case |
|------|----------|--------------|------------------|
| 1 | docs/plan.md | 411 | Phase tracking, implementation planning |
| 2 | README.md | 159 | Project overview, quick start |
| 3 | docs/principles.md | 88 | Design constraints, architecture rules |
| 4 | CLAUDE.md | 62 | Development workflow entry point |
| 5 | docs/examples-usage.md | 62 | Setup tutorials, usage examples |

## Documentation Optimization History

### Phase 23 (MCP Documentation Consolidation)

1. **Reduce redundancy**: Consolidated 4 MCP documents → 1 comprehensive guide (mcp-guide.md)
2. **Improve navigation**: Created CLAUDE.md quick links and FAQ section
3. **Simplify completed phases**: Moved detailed phase docs to plans/ directory
4. **Enhance discoverability**: Added this navigation map

### README Simplification (Post-Phase 23)

1. **Drastic size reduction**: README.md simplified from 1909 lines → 275 lines (85% reduction)
2. **New reference docs**: Created cli-reference.md, jsonl-reference.md, features.md
3. **Clear documentation hierarchy**:
   - README: Quick start and overview (public-facing)
   - Reference docs: Complete technical documentation
   - CLAUDE.md: Development entry point (internal)
4. **Benefits**:
   - New users understand the project in < 2 minutes
   - Advanced users find detailed docs easily
   - Developers have clear separation from public docs

# meta-cc Documentation Map

This document provides a visual overview of documentation dependencies and navigation guide.

## New Directory Structure (Phase 1+2 Completed)

The documentation has been reorganized into a clearer, more intuitive structure:

```
docs/
├── core/                    # Core documents (most accessed)
│   ├── plan.md             # Project roadmap
│   └── principles.md       # Design constraints
│
├── guides/                  # Task-oriented guides
│   ├── capabilities.md     # Capability development (renamed from capabilities-guide.md)
│   ├── git-hooks.md        # Git hooks usage
│   ├── integration.md      # Integration patterns (renamed from integration-guide.md)
│   ├── mcp.md              # MCP complete reference (renamed from mcp-guide.md)
│   ├── plugin-development.md
│   ├── release-process.md
│   └── troubleshooting.md
│
├── reference/               # Complete specifications
│   ├── cli.md              # CLI commands (renamed from cli-reference.md)
│   ├── features.md         # Advanced features
│   ├── jsonl.md            # Output format (renamed from jsonl-reference.md)
│   ├── repository-structure.md
│   └── unified-meta-command.md
│
├── tutorials/               # Step-by-step learning
│   ├── cli-composability.md
│   ├── cookbook.md
│   ├── examples.md         # Basic examples (renamed from examples-usage.md)
│   ├── github-setup.md
│   └── installation.md
│
├── architecture/            # Architecture & design
│   ├── adr/                # Architecture Decision Records
│   └── proposals/          # Technical proposals
│
├── methodology/             # Universal methodologies
│   └── documentation-management.md
│
└── archive/                 # Archived documents
```

**Naming Conventions Applied**:
- ✅ Lowercase with hyphens
- ✅ Removed redundant `-guide` and `-reference` suffixes
- ✅ Simplified names (e.g., `examples-usage.md` → `examples.md`)
- ✅ Clear categorization by document type

## Documentation Dependency Graph

```mermaid
graph TD

  %% Core Entry Points
  CLAUDE_md["CLAUDE.md<br/>(Main Entry)"]:::entry
  README_md["README.md<br/>(Public Doc)"]:::entry

  %% Key Guides
  docs_plan_md["plan.md<br/>(Roadmap)"]:::guide
  docs_principles_md["principles.md<br/>(Design Rules)"]:::guide
  docs_plugin_development_md["plugin-development.md<br/>(Plugin Workflow)"]:::guide
  docs_repository_structure_md["repository-structure.md<br/>(Directory Guide)"]:::guide
  docs_unified_meta_command_md["unified-meta-command.md<br/>(/meta Command)"]:::guide
  docs_integration_guide_md["integration-guide.md"]:::guide
  docs_mcp_guide_md["mcp-guide.md<br/>(MCP Complete)"]:::guide
  docs_capabilities_guide_md["capabilities-guide.md"]:::guide

  %% Maintenance Guides
  docs_git_hooks_md["git-hooks.md<br/>(Git Hooks)"]:::maintenance
  docs_release_process_md["release-process.md<br/>(Release)"]:::maintenance

  %% Reference Docs
  docs_cli_reference_md["cli-reference.md<br/>(CLI Commands)"]:::reference
  docs_jsonl_reference_md["jsonl-reference.md<br/>(Output Format)"]:::reference
  docs_features_md["features.md<br/>(Advanced Features)"]:::reference
  docs_examples_usage_md["examples-usage.md<br/>(Tutorials)"]:::reference
  docs_cookbook_md["cookbook.md<br/>(Advanced Use Cases)"]:::reference

  %% Architecture
  docs_adr_README_md["ADR Index"]:::adr
  docs_proposals_md["Proposals"]:::adr

  %% Dependencies - Core Entry Points
  CLAUDE_md --> docs_principles_md
  CLAUDE_md --> docs_plan_md
  CLAUDE_md --> docs_plugin_development_md
  CLAUDE_md --> docs_repository_structure_md
  CLAUDE_md --> docs_mcp_guide_md
  CLAUDE_md --> docs_integration_guide_md
  CLAUDE_md --> docs_unified_meta_command_md

  README_md --> docs_mcp_guide_md
  README_md --> docs_integration_guide_md
  README_md --> docs_examples_usage_md
  README_md --> docs_cli_reference_md
  README_md --> docs_features_md

  %% Dependencies - Guides
  docs_principles_md --> docs_adr_README_md
  docs_plan_md --> docs_adr_README_md
  docs_plan_md --> docs_proposals_md
  docs_plugin_development_md --> docs_git_hooks_md
  docs_plugin_development_md --> docs_release_process_md
  docs_plugin_development_md --> docs_repository_structure_md
  docs_plugin_development_md --> docs_unified_meta_command_md
  docs_unified_meta_command_md --> docs_capabilities_guide_md
  docs_integration_guide_md --> docs_examples_usage_md
  docs_mcp_guide_md --> docs_integration_guide_md
  docs_examples_usage_md --> docs_cookbook_md

  %% Dependencies - Reference
  docs_cli_reference_md --> docs_jsonl_reference_md
  docs_cli_reference_md --> docs_mcp_guide_md
  docs_features_md --> docs_cli_reference_md
  docs_features_md --> docs_mcp_guide_md
  docs_cookbook_md --> docs_features_md

  %% Styles
  classDef entry fill:#e8f5e9,stroke:#4caf50,stroke-width:3px
  classDef guide fill:#e3f2fd,stroke:#2196f3,stroke-width:2px
  classDef maintenance fill:#fff9c4,stroke:#fbc02d,stroke-width:2px
  classDef reference fill:#f3e5f5,stroke:#9c27b0,stroke-width:2px
  classDef adr fill:#fff3e0,stroke:#ff9800,stroke-width:2px
```

## Quick Navigation Guide

### For New Users

1. **Start**: [README.md](../README.md) - Quick install and overview
2. **Setup MCP**: [docs/guides/mcp.md](guides/mcp.md) - Complete MCP setup and usage
3. **Examples**: [docs/tutorials/examples.md](tutorials/examples.md) - Step-by-step tutorials
4. **Troubleshooting**: [docs/guides/troubleshooting.md](guides/troubleshooting.md) - Common issues

### For Advanced Users

1. **CLI Reference**: [docs/reference/cli.md](reference/cli.md) - Complete command reference
2. **JSONL Reference**: [docs/reference/jsonl.md](reference/jsonl.md) - Output format and jq patterns
3. **Features**: [docs/reference/features.md](reference/features.md) - Advanced capabilities
4. **CLI Composability**: [docs/tutorials/cli-composability.md](tutorials/cli-composability.md) - Unix pipeline patterns

### For Claude Code Development

1. **Entry Point**: [CLAUDE.md](../CLAUDE.md) - Development workflow
2. **Design Rules**: [docs/core/principles.md](core/principles.md) - Core constraints
3. **Roadmap**: [docs/core/plan.md](core/plan.md) - Phase-by-phase plan
4. **Plugin Development**: [docs/guides/plugin-development.md](guides/plugin-development.md) - Complete workflow
5. **Repository Structure**: [docs/reference/repository-structure.md](reference/repository-structure.md) - Directory guide
6. **Architecture**: [docs/architecture/adr/README.md](architecture/adr/README.md) - ADR index

### For Plugin & Integration Development

1. **Plugin Workflow**: [docs/guides/plugin-development.md](guides/plugin-development.md) - Complete development guide
2. **Unified /meta Command**: [docs/reference/unified-meta-command.md](reference/unified-meta-command.md) - /meta command guide
3. **Git Hooks**: [docs/guides/git-hooks.md](guides/git-hooks.md) - Automatic version bumping
4. **Release Process**: [docs/guides/release-process.md](guides/release-process.md) - Release workflow
5. **Repository Structure**: [docs/reference/repository-structure.md](reference/repository-structure.md) - Directory organization

### For Integration Work

1. **Integration Guide**: [docs/guides/integration.md](guides/integration.md) - Choosing MCP/Slash/Subagent
2. **MCP Complete Guide**: [docs/guides/mcp.md](guides/mcp.md) - All MCP topics in one place
3. **Capabilities Guide**: [docs/guides/capabilities.md](guides/capabilities.md) - Create custom capabilities

## Document Roles

| Document | Role | Target Audience | Update Frequency |
|----------|------|----------------|------------------|
| **CLAUDE.md** | Development entry point (simplified) | Claude Code | Every phase |
| **README.md** | Public documentation (simplified) | End users | Major releases |
| **docs/core/plan.md** | Roadmap and status | Developers | Continuous |
| **docs/core/principles.md** | Design constraints | Developers | Rarely (stable) |
| **docs/guides/plugin-development.md** | Plugin development workflow | Plugin developers | When workflow changes |
| **docs/reference/repository-structure.md** | Directory organization guide | Developers | Rarely (stable) |
| **docs/reference/unified-meta-command.md** | /meta command complete guide | Users & Developers | When /meta evolves |
| **docs/guides/mcp.md** | MCP complete reference | Users & Developers | As MCP evolves |
| **docs/guides/integration.md** | Integration decisions | Advanced users | Stable |
| **docs/guides/release-process.md** | Release workflow | Maintainers | Rarely (stable) |
| **docs/guides/git-hooks.md** | Git hooks usage | Developers | Rarely (stable) |
| **docs/tutorials/examples.md** | Step-by-step tutorials | New users | When features added |
| **docs/reference/cli.md** | Complete CLI command reference | Advanced users | When commands added |
| **docs/reference/jsonl.md** | Output format and jq patterns | Advanced users | Rarely (stable) |
| **docs/reference/features.md** | Advanced features overview | Advanced users | When features added |
| **docs/architecture/adr/** | Architecture decisions | Architects | Per decision |

## Most Accessed Documents (from meta-cc analysis)

| Rank | Document | Access Count | Primary Use Case |
|------|----------|--------------|------------------|
| 1 | docs/core/plan.md | 411 | Phase tracking, implementation planning |
| 2 | README.md | 159 | Project overview, quick start |
| 3 | docs/core/principles.md | 88 | Design constraints, architecture rules |
| 4 | CLAUDE.md | 62 | Development workflow entry point |
| 5 | docs/tutorials/examples.md | 62 | Setup tutorials, usage examples |

---

## Universal Methodology

For universal, project-independent software development methodologies, see:

**[docs/methodology/](methodology/)** - Software Development Methodology

- **[Documentation Management](methodology/documentation-management.md)**: Comprehensive guide to documentation management in Claude Code projects
- **Future guides**: TDD, error handling, cross-platform development, version management, and more

---

## Migration Notes

**Phase 1+2 Completed** (2025-10-12):
- ✅ Created categorized subdirectories (core/, guides/, reference/, tutorials/)
- ✅ Moved all documents to appropriate categories
- ✅ Renamed files to remove redundant suffixes (-guide, -reference)
- ✅ Simplified file names (e.g., examples-usage.md → examples.md)
- ✅ Updated all links in DOCUMENTATION_MAP.md

**Next Steps** (Future):
- 📋 Update CLAUDE.md links to new paths
- 📋 Update README.md links to new paths
- 📋 Update internal links within moved documents
- 📋 Create symlinks for backward compatibility (optional)
- 📋 Plans directory restructuring (add descriptive names)

---

**Last Updated**: 2025-10-12

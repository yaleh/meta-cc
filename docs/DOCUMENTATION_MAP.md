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
│   ├── jsonl-schema.md     # JSONL session file schema
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
├── examples/                # Query examples and patterns
│   ├── frequent-jsonl-queries.md  # Most frequently used JSONL queries
│   ├── jq-query-examples.md       # Single-file query patterns (19 examples)
│   ├── multi-file-jsonl-queries.md # Multi-file queries with results (100 records)
│   └── query-cookbook.md          # Practical query cookbook
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
  docs_integration_guide_md["integration.md"]:::guide
  docs_mcp_guide_md["mcp.md<br/>(MCP Complete)"]:::guide
  docs_capabilities_guide_md["capabilities.md"]:::guide

  %% Maintenance Guides
  docs_git_hooks_md["git-hooks.md<br/>(Git Hooks)"]:::maintenance
  docs_release_process_md["release-process.md<br/>(Release)"]:::maintenance

  %% Reference Docs
  docs_cli_reference_md["cli.md<br/>(CLI Commands)"]:::reference
  docs_jsonl_reference_md["jsonl.md<br/>(Output Format)"]:::reference
  docs_jsonl_schema_md["jsonl-schema.md<br/>(Session Schema)"]:::reference
  docs_features_md["features.md<br/>(Advanced Features)"]:::reference
  docs_examples_usage_md["examples.md<br/>(Tutorials)"]:::reference
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
3. **JSONL Schema**: [docs/reference/jsonl-schema.md](reference/jsonl-schema.md) - Session file structure specification
4. **Query Examples**:
   - [docs/examples/jq-query-examples.md](examples/jq-query-examples.md) - Single-file query patterns (19 examples)
   - [docs/examples/multi-file-jsonl-queries.md](examples/multi-file-jsonl-queries.md) - Multi-file queries with results
   - [docs/examples/frequent-jsonl-queries.md](examples/frequent-jsonl-queries.md) - Most frequently used queries
5. **Features**: [docs/reference/features.md](reference/features.md) - Advanced capabilities
6. **CLI Composability**: [docs/tutorials/cli-composability.md](tutorials/cli-composability.md) - Unix pipeline patterns

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

### For Documentation Maintenance

**Documentation Health Capabilities** (based on [Role-Based Documentation Architecture](methodology/role-based-documentation.md)):

| Your Question | Use | Frequency | What It Reveals |
|---------------|-----|-----------|-----------------|
| Is my doc system healthy? | `/meta doc-health` | Monthly, pre-commit | Role violations, size issues, R/E anomalies |
| Will this doc become stale? | `/meta doc-evolution` | Monthly | Phase transitions, archival probability |
| What docs am I missing? | `/meta doc-gaps` | Quarterly, pre-release | Undocumented features, knowledge silos |
| Are my docs effective? | `/meta doc-usage` | Quarterly | Resolution rates, task alignment |

**Typical Maintenance Workflow**:
```
Bootstrap:       /meta doc-health (establish baseline)
Monthly:         /meta doc-health + /meta doc-evolution (15 min)
Quarterly:       All 4 capabilities + comprehensive review (90 min)
Pre-Release:     /meta doc-gaps (ensure completeness)
```

See [Role-Based Documentation Architecture](methodology/role-based-documentation.md) for complete methodology.

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
| **docs/reference/jsonl-schema.md** | JSONL session file schema specification | Developers & Analysts | When schema changes |
| **docs/reference/features.md** | Advanced features overview | Advanced users | When features added |
| **docs/examples/jq-query-examples.md** | Single-file JSONL query patterns | Advanced users & Analysts | Rarely (stable patterns) |
| **docs/examples/multi-file-jsonl-queries.md** | Multi-file JSONL query results | Advanced users & Analysts | Rarely (reference examples) |
| **docs/examples/frequent-jsonl-queries.md** | Most frequently used JSONL queries | Advanced users & Analysts | Rarely (usage patterns) |
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

## Plans Directory Structure

Implementation plans for each development phase are organized in `plans/` with descriptive naming:

```
plans/
├── 00-bootstrap/                  # Project initialization
├── 01-session-locator/            # Session file location
├── 02-jsonl-parser/               # JSONL parsing
├── 08-mcp-integration/            # MCP server integration
├── 13-output-simplification/      # Output format standardization
├── 22-unified-meta-command/       # Unified /meta command
└── ...                            # 21 phases total
```

Each phase directory contains:
- **plan.md** - Detailed implementation plan with TDD stages
- **README.md** - Quick reference and navigation (12 phases)
- **Additional files** - Stage summaries, execution reports (as needed)

See [Documentation Management Methodology](methodology/documentation-management.md) for plans directory workflow.

---

## Universal Methodology

For universal, project-independent software development methodologies, see:

**[docs/methodology/](methodology/)** - Software Development Methodology

- **[Documentation Management](methodology/documentation-management.md)**: Comprehensive guide to documentation management in Claude Code projects
- **[Role-Based Documentation Architecture](methodology/role-based-documentation.md)**: Data-driven methodology for organizing and maintaining documentation based on actual usage patterns, with automated health checks and continuous optimization (Methodology v1.0, 2025-10-13)
- **[Empirical Methodology Development](methodology/empirical-methodology-development.md)**: Meta-methodology for developing software engineering practices through observation, analysis, and automation. Includes OCA Framework (Observe-Codify-Automate) and implementation roadmap (Framework v1.0, 2025-10-13)
- **[Bootstrapped Software Engineering](methodology/bootstrapped-software-engineering.md)**: Meta-methodology framework for self-evolving software development processes with Meta-Agent bootstrapping (Framework v1.0, 2025-10-13)
- **[Value Space Optimization](methodology/value-space-optimization.md)**: Mathematical framework for training Agents and Meta-Agents from project history, treating development as optimization in high-dimensional value space with Agent as gradient (∇V) and Meta-Agent as Hessian (∇²V) (Framework v1.0, 2025-10-14)
- **Future guides**: TDD, error handling, cross-platform development, version management, and more (using OCA Framework)

---

## Restructuring Summary

**Phase 1+2: Directory Structure & File Renaming** (2025-10-12):
- ✅ Created categorized subdirectories (core/, guides/, reference/, tutorials/)
- ✅ Moved 19 documents to appropriate categories
- ✅ Renamed files to remove redundant suffixes (-guide, -reference)
- ✅ Simplified file names (e.g., examples-usage.md → examples.md)
- ✅ Moved architecture directories (adr/, proposals/ → architecture/)
- ✅ Updated all entry point links (CLAUDE.md, README.md, DOCUMENTATION_MAP.md)

**Phase 3: Plans Directory Restructuring** (2025-10-12):
- ✅ Renamed all 21 phase directories with descriptive names
- ✅ Format: NN-descriptive-name/ (e.g., 08-mcp-integration/)
- ✅ Created README.md quick references for 12 phase directories
- ✅ Improved discoverability and self-documentation

**Phase 4: Internal Link Fixes** (2025-10-12):
- ✅ Fixed 70+ broken internal links across 18 files
- ✅ Updated cross-directory references
- ✅ Fixed architecture path references
- ✅ All documentation links now working correctly
- ✅ **Documentation Link Checker Tool**: Created `/meta doc-links` capability for ongoing validation

**Phase 5: Documentation Management Capabilities** (2025-10-12):
- ✅ **Documentation Sync Checker**: Created `/meta doc-sync` capability for cross-reference validation
- ✅ **Project Bootstrap**: Created `/meta project-bootstrap` capability implementing Documentation Management Methodology v5.0
- ✅ **Link Validation**: Automated tool for checking internal markdown links with severity classification

**Verification Results** (2025-10-12):
- ✅ Main entry points: 100% links working (CLAUDE.md, README.md, DOCUMENTATION_MAP.md)
- ✅ Internal cross-references: 100% links working (0 broken links detected)
- ✅ Plans directory: All phases renamed with descriptive names
- ✅ Naming conventions: Fully standardized (lowercase + hyphens)
- ✅ Documentation management toolchain: Complete with link checking, sync validation, and project bootstrap

**Phase 6: Role-Based Documentation Methodology** (2025-10-13):
- ✅ **Role-Based Documentation Architecture**: Created comprehensive methodology document (v1.0)
- ✅ **6 Document Roles Defined**: Context Base, Living, Specification, Reference, Episodic, Archive
- ✅ **Key Metrics Framework**: R/E ratio, access density, lifecycle stages
- ✅ **4 Maintenance Capabilities Refined**: doc-health, doc-evolution, doc-gaps, doc-usage (concise style, 219-302 lines each)
- ✅ **Empirical Case Study**: Complete analysis of meta-cc project documentation with actual data
- ✅ **Implementation Guide**: Step-by-step setup, automation hooks, troubleshooting
- ✅ **Updated DOCUMENTATION_MAP.md**: Added methodology reference in Universal Methodology section

**Phase 7: Documentation Simplification** (2025-10-25):
- ✅ **Phase Consolidation**: Merged Phase 12-13 (MCP Integration), Phase 18-22 (Open Source Ecosystem)
- ✅ **Archive Management**: Moved Phase 23-25 (Query Interface Refactoring v2.0) to archive with comprehensive documentation
- ✅ **Structure Optimization**: Simplified project roadmap while preserving technical history
- ✅ **Reference Cleanup**: Updated all cross-references to point to consolidated or archived content

**Updated Verification Results** (2025-10-25):
- ✅ Main entry points: 100% links working (CLAUDE.md, README.md, DOCUMENTATION_MAP.md)
- ✅ Internal cross-references: 100% links working (0 broken links detected)
- ✅ Plans directory: All phases renamed with descriptive names
- ✅ Naming conventions: Fully standardized (lowercase + hyphens)
- ✅ Documentation management toolchain: Complete with link checking, sync validation, and project bootstrap
- ✅ Archive organization: Historical Phase 23-25 documentation properly archived with cross-references

---

**Last Updated**: 2025-10-25

# Quick Access Navigation

## Most Important Documents

### Getting Started (Top 5)

1. [README](../README.md) - Project overview and installation
2. [Implementation Plan](core/plan.md) - Development roadmap (423 accesses)
3. [Design Principles](core/principles.md) - Core constraints (90 accesses)
4. [CLAUDE.md](../CLAUDE.md) - Development workflow (87 accesses)
5. [Examples](tutorials/examples.md) - Step-by-step tutorials

### Essential Guides

- [MCP Server Guide](guides/mcp.md) - Complete MCP reference
- [MCP Query Tools Reference](guides/mcp-query-tools.md) - Authoritative consolidated query reference
- [Two-Stage Query Guide](guides/two-stage-query-guide.md) - Custom jq over selected files
- [Plugin Development](guides/plugin-development.md) - Plugin workflow
- [Integration Guide](guides/integration.md) - Claude Code and Codex setup and usage
- [Troubleshooting](guides/troubleshooting.md) - Common issues

### Key References

- [JSONL Reference](reference/jsonl.md) - Output format and jq patterns
- [Repository Structure](reference/repository-structure.md) - Directory guide
- [JSONL Schema](reference/jsonl-schema.md) - Claude Code and Codex input schemas
- [Codex History Model](reference/codex-history-model.md) - Codex lineage, archives, pagination
- [Codex App-Server Backend](reference/codex-app-server.md) - Codex backend modes and fallback
- [Local FTS Index](reference/fts-index.md) - Internal query-acceleration index (no standalone tool)

### Quick Tasks

#### "I want to query session data"

→ [MCP Guide](guides/mcp.md#tool-catalog) + [MCP Query Tools Reference](guides/mcp-query-tools.md)

#### "I want Codex support"

→ [Integration Guide](guides/integration.md) + [Provider Examples](tutorials/examples.md#provider-examples)

#### "I want to develop a plugin"

→ [Plugin Development Guide](guides/plugin-development.md)

#### "I want to understand the architecture"

→ [Architecture Decision Records](architecture/adr/README.md)

#### "I want to contribute"

→ [Design Principles](core/principles.md) + [Plan](core/plan.md)

#### "I'm getting an error"

→ [Troubleshooting Guide](guides/troubleshooting.md)

## Category-Based Navigation

### 📍 Core Documents

- [plan.md](core/plan.md) - Implementation roadmap
- [principles.md](core/principles.md) - Design constraints

### 📚 Guides

- [capabilities.md](guides/capabilities.md) - Capability development
- [git-hooks.md](guides/git-hooks.md) - Git hooks usage
- [integration.md](guides/integration.md) - Integration patterns
- [mcp.md](guides/mcp.md) - MCP complete reference
- [mcp-query-tools.md](guides/mcp-query-tools.md) - Consolidated query tools reference
- [two-stage-query-guide.md](guides/two-stage-query-guide.md) - Two-stage jq workflow
- [plugin-development.md](guides/plugin-development.md) - Plugin workflow
- [release-process.md](guides/release-process.md) - Release guide
- [troubleshooting.md](guides/troubleshooting.md) - Problem solving

### 📖 Reference

- [features.md](reference/features.md) - Advanced features
- [jsonl.md](reference/jsonl.md) - Output format
- [jsonl-schema.md](reference/jsonl-schema.md) - Claude Code and Codex schemas
- [codex-history-model.md](reference/codex-history-model.md) - Codex provider internals
- [codex-app-server.md](reference/codex-app-server.md) - Codex backend modes
- [fts-index.md](reference/fts-index.md) - Local FTS index (internal foundation)
- [repository-structure.md](reference/repository-structure.md) - Directory guide

### 🎓 Tutorials

- [cookbook.md](tutorials/cookbook.md) - Advanced use cases
- [examples.md](tutorials/examples.md) - Basic examples
- [github-setup.md](tutorials/github-setup.md) - GitHub configuration
- [installation.md](tutorials/installation.md) - Install guide

### 🏗️ Architecture

- [ADR Index](architecture/adr/README.md) - Architecture decisions
- [Technical Proposal](architecture/proposals/meta-cognition-proposal.md)

### 🧬 Methodology

- [Documentation Management](methodology/documentation-management.md)
- [Empirical Methodology](methodology/empirical-methodology-development.md)
- [Bootstrapped SE](methodology/bootstrapped-software-engineering.md)
- [Value Space Optimization](methodology/value-space-optimization.md)
- [Role-Based Documentation](methodology/role-based-documentation.md)

## Search Keywords → Documents

| If you're looking for... | Go to... |
|--------------------------|----------|
| MCP, server, query | [MCP Guide](guides/mcp.md) |
| Codex, provider, skills | [Integration Guide](guides/integration.md) |
| Plugin, slash command, skill | [Plugin Development](guides/plugin-development.md) |
| JSON, output, format | [JSONL Reference](reference/jsonl.md) |
| Error, problem, issue | [Troubleshooting](guides/troubleshooting.md) |
| Install, setup | [Installation](tutorials/installation.md) |
| Example, tutorial | [Examples](tutorials/examples.md) |
| Architecture, design | [ADR Index](architecture/adr/README.md) |
| Test, TDD, coverage | [Design Principles](core/principles.md) |
| Release, version | [Release Process](guides/release-process.md) |

---

*Navigation depth reduced from 3 to 1-2 clicks for all important documents*

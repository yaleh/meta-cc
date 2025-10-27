# meta-cc

[![CI](https://github.com/yaleh/meta-cc/actions/workflows/ci.yml/badge.svg)](https://github.com/yaleh/meta-cc/actions)
[![License](https://img.shields.io/github/license/yaleh/meta-cc)](LICENSE)
[![Release](https://img.shields.io/github/v/release/yaleh/meta-cc)](https://github.com/yaleh/meta-cc/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/yaleh/meta-cc)](go.mod)
[![Plugin Marketplace](https://img.shields.io/badge/Claude_Code-Plugin_Marketplace-blue)](https://github.com/yaleh/meta-cc)

**Meta-cognition tool for Claude Code** - Analyze session history, detect patterns, optimize workflows.

---

## What is meta-cc?

meta-cc helps you understand and improve your Claude Code workflows through:

- **Natural language queries** - `/meta "show errors"` - discover capabilities by asking
- **Autonomous analysis** - Claude automatically queries session data via MCP tools
- **Interactive coaching** - `@meta-coach` provides personalized workflow optimization

**Zero configuration required** - works out of the box with Claude Code.

---

## Quick Install

### 1. Install Plugin

```bash
/plugin marketplace add yaleh/meta-cc
/plugin install meta-cc
```

The meta-cc plugin includes everything you need:
- **Unified /meta command** - Natural language interface for session analysis
- **5 Specialized Agents** - Project planning, iteration management, knowledge extraction
- **13 Capabilities** - Error analysis, quality scanning, workflow optimization
- **20 MCP Tools** - Autonomous session data queries (v2.0 unified query interface)
- **16 Validated Methodology Skills**:
  - **BAIME** (Bootstrapped AI Methodology Engineering) - Framework for developing methodologies
  - **Testing Strategy** - TDD, coverage-driven gap closure (3.1x speedup)
  - **CI/CD Optimization** - Quality gates, release automation (2.5-3.5x speedup)
  - **Error Recovery** - 13-category taxonomy, diagnostic workflows
  - **Documentation Management** - Templates, patterns, automation tools
  - And 11 more validated skills

**Try BAIME**: Just tell Claude _"Use BAIME to build [domain] capability and complete [tasks]"_ - Claude handles everything automatically. See [BAIME Usage Guide](docs/tutorials/baime-usage.md#try-baime-in-3-steps) for examples and the 3-level trial workflow.

### 2. Install MCP Server Binary

```bash
# Linux/macOS (one-liner)
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-amd64.tar.gz | tar xz
cd meta-cc-plugin-linux-amd64
./install.sh
```

**Other platforms**: See [Installation Guide](docs/tutorials/installation.md) for macOS (Apple Silicon), Windows, and manual installation.

### 3. Configure MCP Server

```bash
claude mcp add meta-cc meta-cc-mcp
```

### Verify Installation

```bash
/meta "show stats"         # Should display session statistics
```

**Troubleshooting**: See [Installation Guide](docs/tutorials/installation.md#troubleshooting) for common issues.

---

## Quick Start

### 1. Natural Language Queries

Ask meta-cc what you need using the `/meta` command:

```bash
/meta "show errors"              # Analyze error patterns
/meta "find repeated workflows"   # Detect repetitive tasks
/meta "which files change most"   # File operation stats
/meta "quality check"             # Code quality scan
/meta "visualize timeline"        # Project timeline
```

**How it works**: meta-cc loads capabilities from GitHub and matches your intent to the best capability.

### 2. Autonomous Analysis (MCP)

Just ask Claude naturally - MCP tools are invoked automatically:

```
"Show me all Bash errors in this project"
"Find user messages mentioning 'refactor'"
"Which tools do I use most often?"
```

**NEW (v2.0+)**: Unified query interface with 20 MCP tools and jq filtering:

```javascript
// Core query tool - unified interface
query({
  resource: "tools",
  filter: {tool_status: "error"},
  jq_filter: '.[] | select(.tool_name == "Bash")'
})

// Convenience tools - optimized for common queries
query_tool_errors({limit: 10})
query_token_usage({stats_first: true})
query_conversation_flow({scope: "session"})

// Raw jq - maximum flexibility for power users
query_raw({
  jq_expression: '.[] | group_by(.tool_name) | map({tool: .[0].tool_name, count: length})'
})
```

**Key Features**:
- **Hybrid Output Mode**: Auto-switches between inline (<8KB) and file_ref (≥8KB)
- **jq Integration**: Native jq filtering for complex queries
- **No Limits by Default**: Returns all results, relies on hybrid mode
- **20 Tools**: 2 core + 8 convenience + 7 legacy + 3 utility tools

**Resources**:
- [MCP Query Tools Reference](docs/guides/mcp-query-tools.md) - Complete tool documentation
- [MCP Query Cookbook](docs/examples/mcp-query-cookbook.md) - 25+ practical examples
- [MCP v2.0 Migration Guide](docs/guides/mcp-v2-migration.md) - Upgrade from v1.x

### 3. Interactive Coaching

Get personalized workflow guidance from the `@meta-coach` subagent:

```bash
@meta-coach Why do my tests keep failing?
@meta-coach Help me optimize my workflow
@meta-coach Analyze my efficiency bottlenecks
```

---

## Documentation

### Getting Started

- **[Installation Guide](docs/tutorials/installation.md)** - Detailed setup for all platforms
- **[Quick Start Tutorial](docs/tutorials/examples.md)** - Step-by-step examples
- **[Troubleshooting](docs/guides/troubleshooting.md)** - Common issues and solutions

### Integration

- **[MCP Guide](docs/guides/mcp.md)** - Complete MCP tool reference (20 tools)
- **[Integration Guide](docs/guides/integration.md)** - MCP, Slash Commands, and Subagents

### Advanced

- **[JSONL Reference](docs/reference/jsonl.md)** - Output format and jq patterns
- **[Feature Overview](docs/reference/features.md)** - Advanced features and capabilities

### Development

- **[Capabilities Guide](docs/guides/capabilities.md)** - Create custom capabilities
- **[Contributing Guide](CONTRIBUTING.md)** - Development workflow and guidelines
- **[Code of Conduct](CODE_OF_CONDUCT.md)** - Community standards

### For Claude Code

- **[CLAUDE.md](CLAUDE.md)** - Project instructions for Claude Code development
- **[Design Principles](docs/core/principles.md)** - Core constraints and architecture
- **[Implementation Plan](docs/core/plan.md)** - Development roadmap

**Complete documentation map**: [DOCUMENTATION_MAP.md](docs/DOCUMENTATION_MAP.md)

---

## Key Features

- 🎯 **Natural language interface** - `/meta` command with semantic matching
- 🔍 **20 MCP query tools** - Autonomous session data analysis with unified query API
- 🎓 **Interactive coaching** - `@meta-coach` subagent for workflow optimization
- 📚 **16 Validated Skills** - Reusable methodologies for testing, CI/CD, error recovery, documentation, and more
- 🤖 **5 Specialized Agents** - Project planning, stage execution, iteration management
- 📊 **Advanced analytics** - jq-based filtering, aggregation, time series
- 📁 **File operation tracking** - Identify hotspots and churn
- 🔄 **Workflow pattern detection** - Find repeated sequences
- 🚀 **Zero dependencies** - Single binary MCP server
- 🔧 **Extensible** - Create custom capabilities with markdown

### Skills (16 Validated Methodologies)

meta-cc includes proven methodologies for systematic software development:

- **Testing Strategy** - TDD, coverage-driven gap closure, CLI testing (3.1x speedup, 89% transferable)
- **CI/CD Optimization** - Quality gates, release automation, smoke testing (2.5-3.5x speedup)
- **Error Recovery** - 13-category taxonomy, diagnostic workflows (95.4% error coverage)
- **Dependency Health** - Security-first, batch remediation (6x speedup)
- **Knowledge Transfer** - Progressive learning paths, onboarding (3-8x ramp-up reduction)
- **Technical Debt Management** - SQALE methodology, prioritization (4.5x speedup)
- **Code Refactoring** - Test-driven refactoring, complexity reduction (28% complexity reduction)
- **Cross-Cutting Concerns** - Error handling, logging, configuration (60-75% faster diagnosis)
- **Observability** - Logs, metrics, traces, structured logging (23-46x speedup)
- **API Design** - 6 validated patterns, parameter categorization (82.5% transferable)
- **Documentation Management** - 5 templates, 3 patterns, automation tools (93% transferable, 3-5x faster creation)
- **Methodology Bootstrapping** - BAIME framework (10-50x speedup, 100% success rate)
- **Agent Prompt Evolution** - Agent specialization tracking (5x performance gap detection)
- **Baseline Quality Assessment** - Rapid convergence enablement (40-50% iteration reduction)
- **Rapid Convergence** - 3-4 iteration methodology development (40-60% time reduction)
- **Retrospective Validation** - Historical data validation (40-60% time reduction, 60-80% cost reduction)

**Usage**: Skills are automatically available after installation. Claude Code will suggest relevant skills based on your tasks.

See [Feature Overview](docs/reference/features.md) for detailed documentation.

---

## Development

### Prerequisites

- Go 1.21 or later
- make

### Build from Source

```bash
git clone https://github.com/yaleh/meta-cc.git
cd meta-cc
make build-mcp  # Build MCP server
```

### Development Workflow (3-Tier)

Use the optimized 3-tier workflow for efficient development:

```bash
make dev           # Quick dev build (format + build, <10s)
make commit        # Pre-commit validation (workspace + tests, <60s)
make push          # Full check before push (all checks + lint, <120s)
```

**Workflow**:
1. **Iterate**: Use `make dev` for fast feedback during development
2. **Commit**: Run `make commit` to validate before committing
3. **Push**: Run `make push` for full verification before pushing to remote

### Run Tests

```bash
make test           # Unit tests (fast)
make test-all       # Including E2E tests (~30s)
make test-coverage  # With coverage report
```

**Coverage Requirement**: Maintain ≥80% test coverage for all code changes.

### Create Custom Capabilities

Create a capability file:

```bash
mkdir -p ~/my-capabilities
cat > ~/my-capabilities/my-feature.md <<'EOF'
---
name: my-feature
description: My custom analysis
keywords: custom, analysis
category: analysis
---

# My Custom Feature

Analyze custom patterns in session data using MCP tools.

## Implementation

Ask Claude to query session data using MCP tools:
- query_tools for tool analysis
- query_user_messages for message search
- get_session_stats for statistics

## Usage

Run with `/meta "my feature"`
EOF
```

Use immediately:

```bash
export META_CC_CAPABILITY_SOURCES="~/my-capabilities:commands"
/meta "my feature"
```

See [Capabilities Guide](docs/guides/capabilities.md) for complete documentation.

---

## Platform Support

- Linux (amd64, arm64)
- macOS (Intel, Apple Silicon)
- Windows (amd64)

---

## Contributing

We welcome contributions! Please see:

- **[Contributing Guide](CONTRIBUTING.md)** - Development process and guidelines
- **[Code of Conduct](CODE_OF_CONDUCT.md)** - Community standards

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

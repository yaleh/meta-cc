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

### 2. Install Binaries

```bash
# Linux/macOS (one-liner)
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-amd64.tar.gz | tar xz
cd meta-cc-plugin-linux-amd64
./install.sh
```

**Other platforms**: See [Installation Guide](docs/installation.md) for macOS (Apple Silicon), Windows, and manual installation.

### 3. Configure MCP Server

```bash
claude mcp add meta-cc meta-cc-mcp
```

### Verify Installation

```bash
meta-cc --version          # Should show version number
/meta "show stats"         # Should display session statistics
```

**Troubleshooting**: See [Installation Guide](docs/installation.md#troubleshooting) for common issues.

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
"Analyze my workflow efficiency"
"Compare error rates between sessions"
```

**Available**: 16 MCP tools for comprehensive session analysis. See [MCP Guide](docs/mcp-guide.md) for complete reference.

### 3. Interactive Coaching

Get personalized workflow guidance from the `@meta-coach` subagent:

```bash
@meta-coach Why do my tests keep failing?
@meta-coach Help me optimize my workflow
@meta-coach Analyze my efficiency bottlenecks
```

---

## Command Line Usage

Use meta-cc CLI for scripting and automation:

```bash
# Session statistics
meta-cc parse stats

# Query tool calls
meta-cc query tools --tool Bash --status error

# Search user messages
meta-cc query user-messages --pattern "fix.*bug"

# Time series analysis
meta-cc stats time-series --metric error-rate --interval day

# File operation tracking
meta-cc stats files --sort-by edit_count --top 10
```

**Pipe to Unix tools**:

```bash
# Filter with jq
meta-cc query tools | jq 'select(.Status == "error")'

# Process with awk
meta-cc query tools --output tsv | awk '{print $2}' | sort | uniq -c
```

See [CLI Reference](docs/cli-reference.md) for complete command list and [CLI Composability](docs/cli-composability.md) for advanced pipeline patterns.

---

## Documentation

### Getting Started

- **[Installation Guide](docs/installation.md)** - Detailed setup for all platforms
- **[Quick Start Tutorial](docs/examples-usage.md)** - Step-by-step examples
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions

### Integration

- **[MCP Guide](docs/mcp-guide.md)** - Complete MCP tool reference (16 tools)
- **[Integration Guide](docs/integration-guide.md)** - Choose MCP vs Slash vs Subagent
- **[CLI Composability](docs/cli-composability.md)** - Unix pipeline patterns

### Advanced

- **[CLI Reference](docs/cli-reference.md)** - Complete command reference
- **[JSONL Reference](docs/jsonl-reference.md)** - Output format and jq patterns
- **[Feature Overview](docs/features.md)** - Advanced features and capabilities

### Development

- **[Capabilities Guide](docs/capabilities-guide.md)** - Create custom capabilities
- **[Contributing Guide](CONTRIBUTING.md)** - Development workflow and guidelines
- **[Code of Conduct](CODE_OF_CONDUCT.md)** - Community standards

### For Claude Code

- **[CLAUDE.md](CLAUDE.md)** - Project instructions for Claude Code development
- **[Design Principles](docs/principles.md)** - Core constraints and architecture
- **[Implementation Plan](docs/plan.md)** - Development roadmap

**Complete documentation map**: [DOCUMENTATION_MAP.md](docs/DOCUMENTATION_MAP.md)

---

## Key Features

- 🎯 **Natural language interface** - `/meta` command with semantic matching
- 🔍 **16 MCP query tools** - Autonomous session data analysis
- 🎓 **Interactive coaching** - `@meta-coach` subagent
- 📊 **Advanced analytics** - SQL-like queries, aggregation, time series
- 📁 **File operation tracking** - Identify hotspots and churn
- 🔄 **Workflow pattern detection** - Find repeated sequences
- 🚀 **Zero dependencies** - Single binary deployment
- 🔧 **Extensible** - Create custom capabilities with markdown
- 🧩 **Unix-friendly** - JSONL streaming, clean I/O, composable pipelines

See [Feature Overview](docs/features.md) for detailed documentation.

---

## Development

### Prerequisites

- Go 1.21 or later
- make

### Build from Source

```bash
git clone https://github.com/yaleh/meta-cc.git
cd meta-cc
make build
```

### Run Tests

```bash
make test           # Unit tests (fast)
make test-all       # Including E2E tests (~30s)
make test-coverage  # With coverage report
```

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

Analyze custom patterns in session data.

## Implementation

```bash
meta-cc query tools --tool Bash --status error
```

## Usage

Run with `/meta "my feature"`
EOF
```

Use immediately:

```bash
export META_CC_CAPABILITY_SOURCES="~/my-capabilities:commands"
/meta "my feature"
```

See [Capabilities Guide](docs/capabilities-guide.md) for complete documentation.

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

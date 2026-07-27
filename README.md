# meta-cc

[![CI](https://github.com/yaleh/meta-cc/actions/workflows/ci.yml/badge.svg)](https://github.com/yaleh/meta-cc/actions)
[![License](https://img.shields.io/github/license/yaleh/meta-cc)](LICENSE)
[![Release](https://img.shields.io/github/v/release/yaleh/meta-cc)](https://github.com/yaleh/meta-cc/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/yaleh/meta-cc)](go.mod)
[![Host Support](https://img.shields.io/badge/Hosts-Claude_Code%20%2B%20Codex-blue)](https://github.com/yaleh/meta-cc)

**Meta-cognition tool for Claude Code and Codex** - Analyze session history, detect patterns, optimize workflows. 16 MCP tools.

> **Note**: Skills and agents from previous versions have been moved to [yaleh/baime](https://github.com/yaleh/baime). meta-cc 3.0.0 focuses exclusively on session history analysis via MCP tools.

---

## What is meta-cc?

meta-cc helps you understand and improve your Claude Code and Codex workflows through:

- **Autonomous analysis** - Claude Code or Codex can query session data via MCP tools
- **16 MCP tools** - Error analysis, quality scanning, work patterns, timelines, bug detection, edit sequence analysis, and more
- **Prompt library** - Save, search, and reuse optimized prompts with Claude Code slash commands or Codex skills

**Native host integrations** - Claude Code marketplace/archive support plus Codex plugin and skills packaging.

---

## Quick Install

### Method 1: Claude Code Plugin Marketplace (Recommended for Claude Code)

```bash
/plugin marketplace add yaleh/meta-cc
/plugin install meta-cc
```

Restart Claude Code. The MCP server is automatically configured via `.mcp.json` bundled in the plugin.

The meta-cc plugin includes:
- **3 Slash Commands** - `/prompt-find`, `/prompt-list`, `/prompt-show` for prompt library management
- **16 MCP Tools** - Session data analysis with consolidated query and two-stage architecture

### Method 2: Archive Install (Claude Code + Codex)

**Full install** (MCP server + Claude Code commands + Codex skills):

```bash
# Linux/macOS (one-liner)
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-amd64.tar.gz | tar xz
cd meta-cc-plugin-linux-amd64
./install.sh
```

The archive installer copies the binary and integration files, installs Claude Code commands under `~/.claude/commands/`, installs Codex skills under `~/.codex/skills/`, and merges the Claude Code MCP server configuration into `~/.claude/mcp.json`. Codex users get plugin metadata under `~/.codex/plugins/meta-cc/` with bundled `.codex-plugin/plugin.json` and `.codex-mcp.json`.

**Prompt-library commands/skills only** (no binary required, any platform):

```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-skills-latest.tar.gz | tar xz
cd meta-cc-skills-*/
./install-skills.sh
```

Use `INSTALL_CLAUDE=0` or `INSTALL_CODEX=0` to install one host only.

### Method 3: Codex Plugin Marketplace (Recommended for Codex CLI 0.145+)

```bash
codex plugin marketplace add .   # from an extracted release archive, or a git checkout of this repo
codex plugin add meta-cc@meta-cc-marketplace
```

Verify with `codex plugin list --json` and `codex mcp list` (expect exactly
one `meta-cc` entry), then **start a new Codex session** — a running
session cannot hot-load a plugin installed after it started. For the
minimal MCP-only fallback (`codex mcp add`), upgrade/uninstall flows, and
troubleshooting duplicate registrations, see
[Installation Guide: Method 1b](docs/tutorials/installation.md#method-1b-codex-plugin-marketplace-preferred-for-codex-cli-0145).

**MCP server binary only** (for CI/Docker/PATH installs):

```bash
# Download the bare binary for your platform, e.g. Linux amd64:
curl -LO https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-mcp-linux-amd64
chmod +x meta-cc-mcp-linux-amd64
INSTALL_DIR=~/.local/bin bash scripts/install/install-mcp.sh meta-cc-mcp-linux-amd64
```

**Other platforms**: See [Installation Guide](docs/tutorials/installation.md) for macOS (Apple Silicon), Windows, and manual installation.

### Verify Installation

In Claude Code or Codex, ask naturally:

```
"Show me all Bash errors in this project"
"Which tools do I use most often?"
"Find user messages mentioning 'refactor'"
```

**Troubleshooting**: See [Installation Guide](docs/tutorials/installation.md#troubleshooting) for common issues.

---

## Quick Start

### Autonomous Analysis (MCP)

Ask Claude Code or Codex naturally - MCP tools are invoked automatically:

```
"Show me all Bash errors in this project"
"Find user messages mentioning 'refactor'"
"Which tools do I use most often?"
"Scan session quality and show me scores"
"Show my work patterns and peak hours"
"Find bug fix pairs in my session"
```

**16 MCP tools: consolidated query tools, two-stage jq, and analysis tools**:

```javascript
// Consolidated query tools - cover the most common access patterns
query_session_signals({type: "errors", limit: 10})       // tool execution errors
query_session_signals({type: "tokens", stats_first: true}) // token usage stats
query_session_content({role: "user", pattern: "refactor"}) // user messages
query_session_content({role: "tool", block_type: "tool_use"}) // tool calls with context
query_file_activity({type: "snapshots"})                  // file history

// Two-stage jq - maximum flexibility for power users
const dir = get_session_directory({scope: "project"})
execute_stage2_query({
  files: dir.files,
  filter: 'select(.type == "assistant")',
  transform: '{timestamp, usage: .message.usage}'
})

// Analysis tools - aggregate and detect patterns
analyze_errors({})          // Aggregate errors by tool and type; result includes data_source field
quality_scan({})            // Compute error/retry/diversity scores
get_work_patterns({})       // Hourly activity and context switches
get_timeline({})            // Chronological session events
analyze_bugs({})            // Error-fix pairs and recurring patterns
get_tech_debt({})           // TODO/FIXME markers and unresolved errors
query_edit_sequences({files: ["/path/to/file.go"]})  // File edit/read patterns, docRole, co-accessed docs
get_session_metadata({})    // JSONL schema, file info, and query templates
```

**Key Features**:
- **Claude Code + Codex support**: Reads Claude transcripts from `~/.claude/projects/` and Codex conversations from `${META_CC_CODEX_ROOT:-~/.codex}/state_5.sqlite` plus rollout JSONL files
- **Provider-aware normalization**: Use `provider: "claude" | "codex" | "all"` on query and analysis tools; Codex `response_item`, `event_msg`, function/custom tool calls, tool outputs, and token counts are normalized through the same MCP surface
- **Hybrid Output Mode**: Auto-switches between inline (<8KB) and file_ref (≥8KB)
- **jq Integration**: Native jq filtering for complex queries; warns when a transform produces all-null results
- **Time Filtering**: `since`/`until` (RFC3339) on all query tools for narrowing to a time window
- **No Limits by Default**: Returns all results, relies on hybrid mode
- **data_source field**: All six analysis tools label results as `measured` (from session data) or `estimated` so callers know data provenance
- **15 Tools**: 3 consolidated query + 4 utility (directory/inspect/stage2/edit-sequences) + 1 metadata + 1 cleanup + 6 analysis

**Resources**:
- [MCP Query Tools Reference](docs/guides/mcp-query-tools.md) - Complete tool documentation
- [MCP Query Cookbook](docs/examples/mcp-query-cookbook.md) - 25+ practical examples
- [MCP v2.0 Migration Guide](docs/guides/mcp-v2-migration.md) - Upgrade from v1.x

### Prompt Library (Slash Commands / Codex Skills)

Save and reuse your best prompts with 3 built-in Claude Code slash commands or Codex skills:

```bash
/prompt-find phase execution      # Search by keywords
/prompt-list sort=usage           # Browse all (sorted by use)
/prompt-show phase-execution-001  # View full prompt details
```

---

## Documentation

### Getting Started

- **[Installation Guide](docs/tutorials/installation.md)** - Detailed setup for all platforms
- **[Quick Start Tutorial](docs/tutorials/examples.md)** - Step-by-step examples
- **[Troubleshooting](docs/guides/troubleshooting.md)** - Common issues and solutions

### Integration

- **[MCP Guide](docs/guides/mcp.md)** - Complete MCP tool reference (16 tools)
- **[Integration Guide](docs/guides/integration.md)** - MCP and Slash Commands
- **[MCP Query Tools Reference](docs/guides/mcp-query-tools.md)** - Consolidated query tools, two-stage jq, hybrid output

### Advanced

- **[JSONL Reference](docs/reference/jsonl.md)** - Output format and jq patterns
- **[Feature Overview](docs/reference/features.md)** - Advanced features and capabilities

### Development

- **[Contributing Guide](CONTRIBUTING.md)** - Development workflow and guidelines
- **[Code of Conduct](CODE_OF_CONDUCT.md)** - Community standards

### Host Notes

- **[CLAUDE.md](CLAUDE.md)** - Project instructions for Claude Code development
- **[Design Principles](docs/core/principles.md)** - Core constraints and architecture
- **[Implementation Plan](docs/core/plan.md)** - Development roadmap
- Codex integration uses `plugin-src/.codex-plugin/plugin.json`, `plugin-src/.codex-mcp.json`, and `plugin-src/skills/*/SKILL.md`

**Complete documentation map**: [DOCUMENTATION_MAP.md](docs/DOCUMENTATION_MAP.md)

---

## Key Features

- **16 MCP tools** - Autonomous session data analysis: 1 session discovery + 3 consolidated query + 4 utility + 1 metadata + 1 cleanup + 6 analysis
- **Claude Code + Codex transcript analysis** - Shared query/analysis surface over both host schemas
- **3 Prompt Library commands/skills** - Prompt management (`prompt-find`, `prompt-list`, `prompt-show`)
- **Advanced analytics** - jq-based filtering, aggregation, time series; `since`/`until` time filtering on all query tools
- **Error analysis** - Aggregate tool errors by name and type, with `data_source` provenance field
- **Quality scanning** - Error/retry/diversity/completion dimensions
- **Work pattern detection** - Tool frequency, hourly activity, context switches
- **Timeline visualization** - Chronological session events as JSON
- **Bug detection** - Error-fix pairs and recurring patterns
- **Tech debt tracking** - TODO/FIXME markers and unresolved errors
- **Edit sequence analysis** - File edit/read patterns, docRole classification, co-accessed document detection
- **File operation tracking** - Identify hotspots and churn
- **Zero dependencies** - Single binary MCP server
- **Prompt Learning System** - Save, search, and reuse optimized prompts with project-specific intelligence

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
make test-e2e-codex # Codex install/session E2E
make test-all       # Including MCP and Codex E2E tests (~30s)
make test-coverage  # With coverage report
```

**Coverage Requirement**: Maintain ≥80% test coverage for all code changes.

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

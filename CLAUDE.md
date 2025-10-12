# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Links

### New to meta-cc?
- **Start here**: [README.md](README.md) - Installation and quick start
- **Understand the design**: [docs/principles.md](docs/principles.md) - Core constraints (≤500 lines/phase, TDD required)
- **Integration guide**: [docs/integration-guide.md](docs/integration-guide.md) - Choose MCP/Slash/Subagent

### Development Workflow
- **Current plan**: [docs/plan.md](docs/plan.md) - Phase roadmap and status
- **Phase details**: [plans/](plans/) - Detailed stage-by-stage plans
- **Build and test**: Run `make all` after each stage

### MCP Server Usage
- **MCP guide**: [docs/mcp-guide.md](docs/mcp-guide.md) - Complete MCP server reference (16 tools)
- **Quick test**: `@meta-cc get_session_stats`

### Common Tasks
- **Fix test failures**: `make test` → Review errors → Fix → `make all`
- **Create slash command**: See [docs/integration-guide.md#creating-custom-integrations](docs/integration-guide.md#creating-custom-integrations)
- **Query session data**: Use MCP tools (see [Using meta-cc](#using-meta-cc) below)

---

## FAQ

**Q: Tests failed after my changes - what should I do?**
A: Run `make all` to see lint + test + build errors. Fix issues iteratively. If tests fail after multiple attempts, HALT development and document blockers (see [Testing Failure Protocol](#phase-planning-and-organization)).

**Q: How much code can I write in one phase?**
A: Maximum 500 lines per phase, 200 lines per stage. See [docs/principles.md](docs/principles.md) for details.

**Q: Should I use MCP, Slash Commands, or Subagent?**
A: Quick rule: Natural questions → MCP | Repeated workflows → Slash | Exploration → Subagent. See [docs/integration-guide.md#decision-framework](docs/integration-guide.md#decision-framework).

**Q: Where is the MCP server implementation?**
A: `cmd/mcp.go` contains the JSON-RPC server. See [docs/mcp-guide.md](docs/mcp-guide.md) for usage.

**Q: How do I query session data programmatically?**
A: Use MCP tools like `query_tools`, `query_user_messages`. See [Using meta-cc](#using-meta-cc) section below.

**Q: What's the difference between `query_user_messages` and `query_conversation`?**
A: `query_user_messages` searches only user messages. `query_conversation` searches both user and assistant messages (can filter by role).

**Q: Why are my MCP query results in a temp file?**
A: Results >8KB automatically use file_ref mode to avoid token limits. Read the file with the Read tool. See [docs/mcp-guide.md#hybrid-output-mode](docs/mcp-guide.md#hybrid-output-mode).

**Q: Do I need to set `limit` parameter for MCP queries?**
A: No, by default queries return all results (hybrid mode handles large data). Only use `limit` when user explicitly requests a specific number. See [Query Limit Strategy](#query-limit-strategy).

**Q: Should I include built-in tools in `query_tool_sequences`?**
A: No, by default built-in tools (Bash, Read, Edit, etc.) are excluded for cleaner workflow patterns and 35x faster analysis. See [Query Tool Sequences](#query-tool-sequences---built-in-tool-filtering).

---

## Project Overview

This repository contains the **meta-cc** (Meta-Cognition for Claude Code) project - a system for analyzing Claude Code session history to provide metacognitive insights and workflow optimization.

### Core Architecture

The system follows a **two-layer architecture**:

1. **meta-cc CLI Tool** (Pure data processing, no LLM)
   - Parses Claude Code session history (JSONL files from `~/.claude/projects/`)
   - Detects patterns using rule-based analysis
   - Outputs structured JSON for consumption by Claude

2. **Claude Code Integration Layer** (LLM-powered)
   - Slash Commands: Quick analysis commands (`/meta-stats`, `/meta-errors`)
   - Subagents: Conversational analysis (`@meta-coach`)
   - MCP Server: Programmatic access for autonomous queries

**Key Design Principle**: The CLI tool handles data extraction and statistical analysis without calling LLMs. Claude (via Slash Commands/Subagents/MCP) performs semantic understanding and generates actionable recommendations based on the structured data.

## Core Constraints (Quick Reference)

See [docs/principles.md](docs/principles.md) for complete details.

**Code Limits**:
- Phase: ≤500 lines of changes
- Stage: ≤200 lines of changes

**Development Methodology**:
- **TDD (Test-Driven Development)**: Write tests before implementation
- **Test Coverage**: ≥80% for all code
- **Testing Protocol**: Run `make all` after each Stage

**Architecture Principles**:
- **Responsibility Separation**: CLI (data extraction) → MCP (filtering) → Subagent (semantic analysis)
- **Output Formats**: JSONL (default) + TSV
- **Query Syntax**: jq expressions
- **Pipeline Pattern**: Session location → JSONL parsing → data extraction → formatting

**Key Technical Decisions**:
- See [docs/plan.md](docs/plan.md) for implementation roadmap
- See [docs/principles.md](docs/principles.md) for design principles
- See [docs/proposals/meta-cognition-proposal.md](docs/proposals/meta-cognition-proposal.md) for architecture

## Repository Structure

```
meta-cc/
├── .claude/
│   ├── commands/              # Entry point for Claude Code
│   │   └── meta.md           # Unified /meta command (single entry point)
│   ├── agents/                # Subagent definitions
│   └── hooks/                 # Project hooks
│
├── capabilities/              # Capability source files (Git tracked)
│   ├── commands/             # Command capabilities (meta-errors, meta-quality-scan, etc.)
│   └── agents/               # Agent capabilities (future)
│
├── .capabilities-cache/       # Runtime capability cache (Git ignored)
│   ├── github/               # Cached capabilities from GitHub
│   │   └── {owner}/{repo}/{branch}/{subdir}/
│   ├── packages/             # Cached capability packages
│   │   └── {hash}/           # Package cache by URL hash
│   └── .meta-cc-cache.json   # Cache metadata (TTL, download times)
│
├── dist/                      # Build artifacts (Git ignored)
│   ├── commands/             # Merged: .claude/commands + capabilities/commands
│   └── agents/               # Merged: .claude/agents + capabilities/agents
│
├── .claude-plugin/            # Plugin metadata for marketplace
├── lib/                       # Shared library files (MCP config, utilities)
├── cmd/                       # CLI commands and MCP server
├── internal/                  # Core logic (parser, analyzer, query, etc.)
├── pkg/                       # Public packages (output, pipeline)
├── docs/                      # Technical documentation
├── plans/                     # Phase-by-phase development plans
└── tests/                     # Test fixtures and integration tests
```

**Directory Purposes**:

**Development Workflow**:
- `.claude/commands/` → Claude Code recognizes `/meta` command
- `capabilities/commands/` → Edit capability files here

**Local Development Configuration**:
```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

**Build and Release**:
- `make sync-plugin-files` → Merges files to `dist/`
- `make bundle-release` → Packages `dist/` into releases

**Production Runtime**:
- Default source: `yaleh/meta-cc@main/commands` (GitHub)
- Cache location: `.capabilities-cache/github/yaleh/meta-cc/main/commands/`

## Development Workflow

### Plugin Development Workflow

**Local Development**:

1. **Edit source files**:
   - `.claude/commands/meta.md` - Entry point for `/meta` command
   - `capabilities/commands/*.md` - Individual capability files
   - `.claude/agents/*.md` - Subagent files

2. **Test immediately** - Claude Code reads from `.claude/` directory (no build needed)

3. **For capability development**, set environment variable:
   ```bash
   export META_CC_CAPABILITY_SOURCES="capabilities/commands"
   ```
   This makes `/meta` load capabilities from local source (no cache, immediate reflection)

4. **Run tests**: `make test` or `make test-all`

5. **Build binaries**: `make build` or `make dev`

**Before Committing**:

1. **Verify changes** in `.claude/commands/`, `capabilities/commands/`, `.claude/agents/`
2. **Run all checks**: `make all`
3. **Do NOT manually create `dist/` directory** - it's a build artifact

**Release Process**:

1. **Sync plugin files**: `make sync-plugin-files` (merges `.claude/` + `capabilities/` → `dist/`)
2. **Create release**: `make bundle-release VERSION=vX.Y.Z` (auto-syncs first)
3. **CI automatically syncs** during release workflow

**Directory Structure Design**:

- **Development**: `.claude/` directory enables real-time testing in Claude Code
- **Capabilities**: `capabilities/` stores capability source files (Git tracked)
- **Build artifacts**: `dist/` contains merged files for release (Git ignored)
- **Runtime cache**: `.capabilities-cache/` stores downloaded GitHub capabilities (Git ignored)
- **CI/CD**: No symlink dependencies, cross-platform compatible (Windows, Linux, macOS)
- **Git**: Only source files tracked, artifacts and cache ignored via `.gitignore`

### Build and Test Requirements

**IMPORTANT**: All development MUST use the Makefile for building and testing:

```bash
# Build the project (runs lint + test + build)
make all

# Build only (no lint or tests)
make build

# Run tests
make test

# Run static analysis (fmt + vet + golangci-lint)
make lint

# Run tests with coverage
make test-coverage
```

**Before committing code**:
1. Run `make all` to ensure code passes linting, tests, and builds
2. Fix any issues reported by static analysis
3. Ensure all tests pass

**Automated checks**:
- `make fmt` - Auto-format code with gofmt
- `make vet` - Run go vet for suspicious constructs
- `make lint` - Run golangci-lint (optional but recommended)

### Phase Planning and Organization

The project follows a **structured phased development approach** with plans organized in the `plans/` directory.

**Phase Structure and Testing Protocol**:

1. **Stage Organization**
   - Each Phase is divided into multiple Stages
   - Each Stage represents a cohesive unit of functionality

2. **Stage-Level Testing** (after each Stage)
   - Run `make all` immediately after completing a Stage
   - This runs: static analysis (lint) → tests → build
   - Fix any errors until all checks pass
   - **CRITICAL**: If checks fail after multiple fix attempts, **HALT Phase development** and output a warning

3. **Phase-Level Testing** (after each Phase)
   - Run `make all` to verify all checks pass
   - Run `make test-coverage` to ensure coverage is maintained
   - Fix any errors until all tests pass
   - **CRITICAL**: If tests fail after multiple fix attempts, **HALT Phase development** and output a warning

4. **Iterative and Incremental Development**
   - Each Phase MUST prioritize **usability** and deliver a **working build**
   - At Phase completion, provide clear **usage instructions** for the deliverable
   - Focus on shipping functional increments rather than complete features

**Testing Failure Protocol**:
- If tests repeatedly fail → Stop development immediately
- Document the failure state and blockers
- Do NOT proceed to the next Stage/Phase until resolved

### Commit Conventions

Use descriptive commit messages with scope prefixes:
- `docs:` for documentation changes
- `feat:` for new features (when implementation begins)
- `refactor:` for code restructuring
- `test:` for test-related changes

Include the Claude Code attribution footer:
```
🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Unified Meta Command

Phase 22 introduces the `/meta` command - a unified entry point for all meta-cognition capabilities with natural language intent matching.

### Usage

```
/meta "natural language intent"
```

**Examples**:
```
/meta "show errors"           → Executes meta-errors
/meta "quality check"         → Executes meta-quality-scan
/meta "visualize timeline"    → Executes meta-timeline
/meta "analyze architecture"  → Executes meta-architecture
/meta "show tech debt"        → Executes meta-tech-debt
```

### How It Works

1. **Discovery**: Loads capabilities from configured sources
2. **Matching**: Semantic keyword matching scores each capability
3. **Execution**: Runs best match or shows available capabilities

### Multi-Source Configuration

Configure capability sources via environment variable:

```bash
# Single local source
export META_CC_CAPABILITY_SOURCES="~/.config/meta-cc/capabilities"

# Package file (GitHub Release)
export META_CC_CAPABILITY_SOURCES="https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz"

# Local package file
export META_CC_CAPABILITY_SOURCES="/path/to/capabilities.tar.gz"

# Multiple sources (priority: left-to-right, left = highest)
export META_CC_CAPABILITY_SOURCES="~/dev/my-caps:/local/caps.tar.gz:https://example.com/caps.tar.gz"

# Mix package, directory, and GitHub sources
export META_CC_CAPABILITY_SOURCES="~/dev/test:./capabilities.tar.gz:yaleh/meta-cc@main/commands"
```

**Source Types**:
- **Local directories**: Immediate reflection, no cache (for development, e.g., `capabilities/commands`)
- **Package files** (`.tar.gz`): Cached with TTL (7 days for releases, 1 hour for branches)
- **GitHub repositories**: Smart caching via jsDelivr CDN (format: `owner/repo@branch/subdir` or `owner/repo/subdir`)
- **Default source**: `yaleh/meta-cc@main/commands` (production GitHub source)

**Branch/Tag Specification**:

Use the `@` symbol to specify a branch, tag, or commit:

```bash
# Specific branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@develop/commands"

# Specific tag (version pinning, recommended for production)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"

# Specific commit hash
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@abc123def/commands"

# Default branch (main)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc/commands"
```

### CDN and Caching

GitHub sources use jsDelivr CDN (https://cdn.jsdelivr.net) for improved performance:

**Benefits**:
- No GitHub API rate limits
- Global CDN delivery (faster)
- Automatic caching with smart TTL

**Cache Strategy**:
- **Branches**: 1-hour cache (mutable, changes frequently)
- **Tags**: 7-day cache (immutable, stable versions)
- **Package files**:
  - Release packages (`/releases/`): 7-day cache (immutable)
  - Custom packages: 1-hour cache (may change)
- **Local sources**: No cache (always fresh)

**Package File Distribution**:

Capabilities can be distributed as prebuilt `.tar.gz` packages for:
- **Offline-friendly**: Download once, cache locally
- **Reliable**: No CDN dependencies, no rate limits
- **Fast**: No network calls after initial download

Cache directory for packages: `~/.capabilities-cache/packages/<hash>/`

**Build Capability Packages**:
```bash
# Build package from capabilities directory
make bundle-capabilities
# Creates: build/capabilities-latest.tar.gz

# Verify package structure
tar -tzf build/capabilities-latest.tar.gz | head -20
```

**Network Resilience**:

meta-cc automatically handles network failures:

- **5xx server errors**: Exponential backoff retry (3 attempts: 1s, 2s, 4s)
- **Network unreachable**: Falls back to stale cache (up to 7 days old)
- **404 errors**: Clear error messages with troubleshooting suggestions

**Cache Metadata**:

Package cache metadata is stored in `~/.capabilities-cache/.meta-cc-cache.json`:
- Tracks download time, TTL, and package URL
- Enables smart cache validation
- Automatic cleanup of expired entries (>7 days)

### Default Source

If `META_CC_CAPABILITY_SOURCES` is not set, capabilities are loaded from:
```
yaleh/meta-cc@main/commands
```

**For local development**, explicitly set the environment variable:
```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

### MCP Tools for Capability Discovery

New MCP tools enable programmatic capability access:

- `list_capabilities()` - Get capability index from all sources
- `get_capability(name)` - Retrieve complete capability content

### Local Development

For capability development, use local sources (no cache):

```bash
export META_CC_CAPABILITY_SOURCES="~/dev/capabilities:capabilities/commands"
# Changes reflect immediately without cache invalidation
```

See [docs/capabilities-guide.md](docs/capabilities-guide.md) for capability development guide.

## Using meta-cc

meta-cc provides programmatic access to session data. Claude can autonomously query this data when analyzing workflows, debugging issues, or providing recommendations.

For complete details, see [MCP Guide](docs/mcp-guide.md).

### Query Tools (16 available)

**Basic Queries**:
- `get_session_stats` - Session statistics and metrics
- `query_tools` - Filter tool calls by name, status (error/success)
  - Parameters: `tool`, `status`, `limit`
- `query_user_messages` - Search user messages with regex patterns
  - Parameters: `pattern` (required, regex), `limit`
  - Example: `query_user_messages(pattern="fix.*bug")`
  - Note: By default, uses hybrid mode for large results (no truncation)
- `query_assistant_messages` - Search assistant response messages with regex patterns
  - Parameters: `pattern` (required, regex), `limit`
  - Example: `query_assistant_messages(pattern="test.*passed")`
  - Note: By default, uses hybrid mode for large results (no truncation)
- `query_conversation` - Search conversation messages (user + assistant) with regex patterns
  - Parameters: `pattern` (required, regex), `limit`, `role` (optional: "user" or "assistant")
  - Example: `query_conversation(pattern="error", role="assistant")`
  - Note: By default, uses hybrid mode for large results (no truncation)
- `query_files` - File-level operation statistics

**Advanced Queries**:
- `query_context` - Error context with surrounding tool calls
- `query_tool_sequences` - Workflow pattern detection
- `query_file_access` - File operation history
- `query_project_state` - Project evolution tracking
- `query_successful_prompts` - High-quality prompt patterns
- `query_tools_advanced` - SQL-like filtering expressions
- `query_time_series` - Metrics over time (hourly/daily/weekly)

**Query Scope**:
- `scope: "project"` - Analyze all sessions in current project (default)
- `scope: "session"` - Analyze only current session

**Output Control Parameters**:
- `stats_only` - Return only statistics, no details
- `stats_first` - Return stats before details
- `jq_filter` - Apply jq expressions for advanced filtering
- `inline_threshold_bytes` - Threshold for inline/file_ref mode (default: 8192, configurable via param or `META_CC_INLINE_THRESHOLD` env var)
- `content_summary` - Return only turn/timestamp/preview (for messages). **Deprecated**: Use hybrid mode instead for better information preservation
- `max_message_length` - Limit message content length (default: 0 = no truncation). **Deprecated**: Rely on hybrid mode for large results

### Hybrid Output Mode

The MCP server automatically selects output mode based on result size:

**Inline Mode (≤8KB results)**:
- Data embedded directly in response: `{"mode": "inline", "data": [...]}`
- Used for quick queries and small result sets

**File Reference Mode (>8KB results)**:
- Data written to temp JSONL file in `/tmp/`
- Response contains metadata: `{"mode": "file_ref", "file_ref": {...}}`
- File ref includes: path, size_bytes, line_count, fields, summary

**Working with File References**:

When you receive a `file_ref` response:

1. **Analyze metadata first** - Check `file_ref.summary` for quick statistics
2. **Use Read tool** - Selectively examine file content (`Read: /tmp/meta-cc-mcp-*.jsonl`)
3. **Use Grep tool** - Search for patterns (`Grep: "Status":"error"`)
4. **Present insights naturally** - Do NOT mention temp file paths to users

**Best Practices**:
- Trust automatic mode selection (default 8KB threshold, configurable via `inline_threshold_bytes` or `META_CC_INLINE_THRESHOLD`)
- All data preserved (inline or file_ref), no truncation
- Analyze metadata first before reading file_ref files
- Do NOT mention "file_ref mode" or temp paths to users
- Use Grep for pattern detection on large files

**Temporary File Management**:
- Files retained for 7 days, auto-cleaned after
- Manual cleanup: `cleanup_temp_files` tool

### Query Limit Strategy

By default, MCP tools **do not limit** the number of results returned:
- Small results automatically use inline mode (≤8KB)
- Large results automatically use file_ref mode (>8KB), allowing you to use Read/Grep/Bash for retrieval

**When to explicitly use the `limit` parameter**:

1. **User explicitly requests a specific number** (e.g., "show me the last 10 errors")
2. **Sample data only** (e.g., "give me a few examples")
3. **Quick exploration** (view a small subset first, then expand if needed)

**Examples**:

```
User: "List all errors in this project"
→ query_tools(status="error")  # No limit, uses file_ref mode

User: "Show me the last 5 errors"
→ query_tools(status="error", limit=5)  # Explicit limit, likely inline mode
```

**Design Philosophy**:
- meta-cc-mcp does not pre-judge how much data you need
- You decide based on conversation context whether to use `limit`
- Hybrid output mode ensures large results won't consume excessive tokens
- For exploratory queries, omit `limit` and let file_ref mode handle the data

### MCP vs CLI Parameter Naming

MCP tools and CLI commands use consistent parameter naming for easier understanding:

- `query_user_messages`: Uses `pattern` (MCP) → `--pattern` (CLI)
- `query_tool_sequences`: Uses `pattern` (MCP) → `--pattern` (CLI)
- Reason: `pattern` is a Unix-standard term (grep, sed, awk all use "pattern")

This consistency eliminates confusion and follows the Unix philosophy.

### Query Tool Sequences - Built-in Tool Filtering

By default, `query_tool_sequences` excludes Claude Code's built-in tools (Bash, Read, Edit, etc.) to focus on high-level workflow patterns. This provides:

- **35x faster analysis** (~30s → <1s for large projects)
- **Cleaner patterns** (MCP tool workflows instead of "Bash → Bash → Bash")
- **Better insight** into meta-cognitive workflows and MCP tool orchestration

**Built-in Tools List** (14 tools):
- File operations: Bash, Read, Edit, Write, Glob, Grep
- Task management: TodoWrite, Task
- Web operations: WebFetch, WebSearch
- Other: SlashCommand, BashOutput, NotebookEdit, ExitPlanMode

**When to include built-in tools**:
- Debugging specific Bash/Read/Edit sequences
- Analyzing low-level file operation patterns
- Complete tool usage audit

**Usage Examples**:
```
# Default: exclude built-in tools (cleaner, faster)
query_tool_sequences(min_occurrences=3)

# Include all tools (slower, noisier)
query_tool_sequences(min_occurrences=3, include_builtin_tools=true)
```

## Integration Patterns

For choosing between integration methods, see [docs/integration-guide.md](docs/integration-guide.md).

**Quick Reference**:
- **MCP Server**: Claude autonomously calls tools during conversation (80% of use cases)
- **Slash Commands**: User-triggered fixed reports (`/meta-stats`, `/meta-errors`)
- **Subagents**: Multi-turn analysis (`@meta-coach` for comprehensive workflow analysis)

**When to Use Each**:
- Natural language query → MCP (automatic)
- Repeated analysis workflow → Slash Command
- Exploratory conversation → Subagent

## Reference Documentation

**Project Documentation**:
- [Implementation Plan](docs/plan.md) - Phase-by-phase development roadmap
- [Design Principles](docs/principles.md) - Core constraints and architecture decisions
- [Integration Guide](docs/integration-guide.md) - Choosing between MCP/Slash/Subagent
- [MCP Guide](docs/mcp-guide.md) - Complete MCP server documentation
- [Technical Proposal](docs/proposals/meta-cognition-proposal.md) - Architecture and design
- [Examples & Usage](docs/examples-usage.md) - Step-by-step setup guides
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions

**Official Claude Code Documentation**:
- [Overview](https://docs.claude.com/en/docs/claude-code/overview)
- [Slash Commands](https://docs.claude.com/en/docs/claude-code/slash-commands)
- [Subagents](https://docs.claude.com/en/docs/claude-code/subagents)
- [MCP Integration](https://docs.claude.com/en/docs/claude-code/mcp)
- [Hooks System](https://docs.claude.com/en/docs/claude-code/hooks)
- [Settings](https://docs.claude.com/en/docs/claude-code/settings)

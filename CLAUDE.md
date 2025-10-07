# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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
├── cmd/              # CLI commands and MCP server
├── internal/         # Core logic (parser, analyzer, query, etc.)
├── pkg/              # Public packages (output, pipeline)
├── docs/             # Technical documentation
├── plans/            # Phase-by-phase development plans
└── tests/            # Test fixtures and integration tests
```

## Development Workflow

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

## Using MCP meta-insight

MCP meta-insight provides programmatic access to session data. Claude can autonomously query this data when analyzing workflows, debugging issues, or providing recommendations.

For complete details, see [MCP Output Modes Documentation](docs/mcp-output-modes.md).

### Query Tools (13 available)

**Basic Queries**:
- `get_session_stats` - Session statistics and metrics
- `query_tools` - Filter tool calls by name, status (error/success)
  - Parameters: `tool`, `status`, `limit`
- `query_user_messages` - Search user messages with regex patterns
  - Parameters: `pattern` (required, regex), `limit`, `max_message_length` (deprecated), `content_summary` (deprecated)
  - Example: `query_user_messages(pattern="fix.*bug", limit=10)`
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
- [MCP Output Modes](docs/mcp-output-modes.md) - Detailed MCP usage and hybrid output mode
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

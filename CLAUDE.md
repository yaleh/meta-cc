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

## Repository Structure

```
meta-cc/
├── docs/
│   └── proposals/
│       ├── meta-cognition-proposal.md   # Main technical specification
│       └── candidates/                  # Original proposal drafts
│           ├── proposal_1.md
│           └── proposal_2.md
├── plans/
│   ├── 1/                               # Phase 1 planning documents
│   ├── 2/                               # Phase 2 planning documents
│   └── 3/                               # Phase 3 planning documents
└── (Future: CLI implementation)
```

## Documentation Organization

### Primary Specification

**`docs/proposals/meta-cognition-proposal.md`** is the authoritative technical design document containing:

- System architecture (with PlantUML diagrams)
- CLI command structure and data flow
- Integration patterns for Slash Commands, Subagents, and MCP
- Implementation roadmap (3-phase: MVP → Indexing → Advanced)
- Complete reference links to Claude Code documentation

### Proposal Evolution

The `candidates/` directory contains two original proposals that were analyzed and merged:
- **proposal_1.md**: Focused on concise CLI design and practical examples
- **proposal_2.md**: Comprehensive architecture with detailed PlantUML diagrams

The final proposal combines the strengths of both: practical implementation focus with clear visual architecture.

## Key Technical Decisions

### Session File Location Strategy

Claude Code stores session history as JSONL files in `~/.claude/projects/{project-hash}/{session-id}.jsonl`. The CLI tool must locate these files via:

1. **Environment variables** (preferred, if Claude Code provides):
   - `CC_SESSION_ID`: Current session UUID
   - `CC_PROJECT_HASH`: Project directory hash (path with `/` → `-`)

2. **Command-line parameters**:
   - `--session <uuid>`: Explicit session ID
   - `--project <path>`: Infer from project path

3. **Auto-detection**: Use current directory and find latest session

### Data Processing Flow

```
JSONL File → meta-cc parse → Structured JSON → Claude Analysis → Recommendations
```

The CLI outputs high-density structured data (tool usage stats, error patterns, timelines) which Claude then interprets semantically to generate insights.

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

The project follows a **structured phased development approach** with plans organized in the `plans/` directory:

```
meta-cc/
├── plans/
│   ├── 1/           # Phase 1 planning documents
│   ├── 2/           # Phase 2 planning documents
│   └── 3/           # Phase 3 planning documents
└── ...
```

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

Include the Claude Code attribution footer:
```
🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Implementation Roadmap (Future)

When development begins, follow this phased approach:

### Phase 1: Core Parser (1-2 weeks)
- JSONL parser for session history
- Basic commands: `meta-cc parse extract`, `meta-cc parse stats`, `meta-cc analyze errors`
- Slash Commands: `/meta-stats`, `/meta-errors`

### Phase 2: Index Optimization (1 week, optional)
- SQLite indexing for cross-session queries
- Advanced query commands

### Phase 3: Semantic Integration (1-2 weeks, optional)
- `@meta-coach` subagent
- MCP server implementation

### Technology Stack (Planned)
- **Language**: Go (zero-dependency deployment, high performance)
- **CLI Framework**: Cobra + Viper (standard in Go ecosystem)
- **Database**: SQLite (optional, for indexing via mattn/go-sqlite3)
- **Output Formats**: JSON, Markdown, CSV

**Why Go?**
- **Single binary deployment**: No runtime dependencies (Python/Node.js) required
- **Fast execution**: Efficient JSONL parsing for large session histories
- **Cross-platform**: Simple cross-compilation for Linux/macOS/Windows
- **Strong concurrency**: Native goroutines for parallel processing when needed

## Working with Integration Questions

When users ask about which integration method to use (MCP vs Slash Commands vs Subagent):

1. **Refer to the Integration Guide**: `docs/integration-guide.md` contains:
   - Core differences (context isolation, invocation mechanisms, execution models)
   - Decision framework and decision trees
   - Use case scenario matrices
   - Anti-patterns and best practices
   - Real-world case studies

2. **Don't recreate the decision framework** - The guide already has comprehensive comparison tables and decision logic. Just summarize key points and link to the guide.

3. **Quick reference**:
   - **MCP Server**: For natural queries, seamless data access (Claude decides when to call)
   - **Slash Commands**: For repeated workflows, fast execution, predictable outputs
   - **Subagent (@meta-coach)**: For exploratory analysis, guided reflection, multi-turn dialogue

4. **When in doubt**: Suggest reading the relevant section of the Integration Guide rather than giving incomplete advice.

## Reference Documentation

**Project Documentation**:
- **[Integration Guide](docs/integration-guide.md)** - Choosing between MCP/Slash/Subagent (NEW - most comprehensive)
- [Examples & Usage](docs/examples-usage.md) - Step-by-step setup guides
- [Technical Proposal](docs/proposals/meta-cognition-proposal.md) - Architecture and design
- [Implementation Plan](docs/plan.md) - Phase-by-phase development roadmap
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions

**Official Claude Code Documentation**:
- [Overview](https://docs.claude.com/en/docs/claude-code/overview)
- [Slash Commands](https://docs.claude.com/en/docs/claude-code/slash-commands)
- [Subagents](https://docs.claude.com/en/docs/claude-code/subagents)
- [MCP Integration](https://docs.claude.com/en/docs/claude-code/mcp)
- [Hooks System](https://docs.claude.com/en/docs/claude-code/hooks)
- [Settings](https://docs.claude.com/en/docs/claude-code/settings)

## Working with Proposals

When modifying `meta-cognition-proposal.md`:

1. **Use PlantUML for architecture diagrams** - The proposal uses PlantUML extensively to visualize data flows, sequences, and architectures. Prefer diagrams over pseudocode.

2. **Link to official docs** - Always reference Claude Code documentation when describing integration mechanisms.

3. **Maintain clarity of responsibility** - Keep the distinction clear: CLI = data processing (no LLM), Claude = semantic analysis.

4. **Follow the phased approach** - Any new features should fit into the 3-phase roadmap (MVP → Indexing → Advanced).

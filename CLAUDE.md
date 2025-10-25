# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Quick Links

### New to meta-cc?
- **Start here**: [README.md](README.md) - Installation and quick start
- **Understand the design**: [docs/core/principles.md](docs/core/principles.md) - Core constraints
- **Integration guide**: [docs/guides/integration.md](docs/guides/integration.md) - Choose MCP/Slash/Subagent

### Development Workflow
- **Current plan**: [docs/core/plan.md](docs/core/plan.md) - Phase roadmap and status
- **Build and test**: Run `make all` after each stage
- **Plugin development**: [docs/guides/plugin-development.md](docs/guides/plugin-development.md) - Complete workflow

### MCP Server Usage
- **MCP guide**: [docs/guides/mcp.md](docs/guides/mcp.md) - Complete MCP reference (16 tools)
- **Quick test**: Use MCP tool `get_session_stats`

### Common Tasks
- **Fix test failures**: `make test` → Review errors → Fix → `make all`
- **Query session data**: Use MCP tools (see [MCP Guide](docs/guides/mcp.md))
- **Update plugin**: [docs/guides/plugin-development.md](docs/guides/plugin-development.md)

---

## FAQ

**Q: Tests failed after my changes - what should I do?**
A: Run `make all` to see lint + test + build errors. Fix issues iteratively. If tests fail after multiple attempts, HALT development and document blockers.

**Q: How much code can I write in one phase?**
A: Maximum 500 lines per phase, 200 lines per stage. See [docs/core/principles.md](docs/core/principles.md).

**Q: Should I use MCP, Slash Commands, or Subagent?**
A: Quick rule: Natural questions → MCP | Repeated workflows → Slash | Exploration → Subagent. See [docs/guides/integration.md](docs/guides/integration.md).

**Q: How do I query session data (v2.0)?**
A: Use the unified `query` tool with jq filtering or convenience tools:
```javascript
// Unified interface
query({resource: "tools", filter: {tool_status: "error"}})

// Convenience tools
query_tool_errors({limit: 10})
query_token_usage({stats_first: true})

// Raw jq for power users
query_raw({jq_expression: '.[] | select(.tool_name == "Bash")'})
```
See [MCP Query Tools Reference](docs/guides/mcp-query-tools.md) for complete tool documentation and [MCP Query Cookbook](docs/examples/mcp-query-cookbook.md) for 25+ examples.

**Q: Why are my MCP query results in a temp file?**
A: Results >8KB automatically use file_ref mode to avoid token limits. Read the file with the Read tool. This is **hybrid output mode** - queries return inline for small results (<8KB) and file_ref for large results (≥8KB). See [MCP Query Tools Reference](docs/guides/mcp-query-tools.md#hybrid-output-mode).

**Q: Do I need to set `limit` parameter for MCP queries?**
A: No, by default queries return all results (hybrid mode handles large data). Only use `limit` when user explicitly requests a specific number. The system automatically switches to file_ref mode for large result sets.

**Q: Which MCP query tool should I use?**
A: Follow this decision tree:
- **Common queries** → Use convenience tools (`query_tool_errors`, `query_token_usage`, etc.)
- **Complex filtering** → Use `query` with `jq_filter`
- **Maximum flexibility** → Use `query_raw` with raw jq expressions
- **Backward compatibility** → Legacy tools still work (`query_tools`, `query_user_messages`, etc.)
See [MCP Query Tools Reference](docs/guides/mcp-query-tools.md#best-practices) for detailed guidance.

**Q: How do I write jq expressions for MCP queries?**
A: Start simple and add complexity:
```javascript
// Step 1: Get all tools
query({resource: "tools"})

// Step 2: Filter by name
query({resource: "tools", jq_filter: '.[] | select(.tool_name == "Bash")'})

// Step 3: Add error filtering
query({resource: "tools", jq_filter: '.[] | select(.tool_name == "Bash" and .status == "error")'})
```
Test jq locally first: `echo '[{"tool":"Bash"}]' | jq '.[]'`. See [MCP Query Tools Reference](docs/guides/mcp-query-tools.md#jq-syntax-quick-reference) for common patterns.

**Q: How do I update plugin version?**
A: Install git hooks (`./scripts/install-hooks.sh`) for automatic bumping, or use `./scripts/bump-plugin-version.sh [patch|minor|major]`. See [docs/guides/git-hooks.md](docs/guides/git-hooks.md).

**Q: What are skills and how do they relate to capabilities?**
A: Skills are reusable methodologies packaged with the plugin (15 skills, ~1.5MB). Capabilities are lightweight content files for the `/meta` command. Skills provide full workflow templates; capabilities provide focused command content.

**Q: How do I use skills?**
A: Skills are automatically available after plugin installation. Claude Code will suggest relevant skills based on your tasks. See skill descriptions in README.md.

---

## Project Overview

**meta-cc** (Meta-Cognition for Claude Code) analyzes Claude Code session history to provide metacognitive insights and workflow optimization.

### Architecture

**Two-layer architecture**:
1. **CLI Tool**: Parses session history (JSONL), detects patterns, outputs structured JSON
2. **Claude Integration**: Slash commands, subagents, and MCP server for LLM-powered analysis

**Key principle**: CLI handles data extraction (no LLM). Claude performs semantic understanding.

### Repository Structure

See [docs/reference/repository-structure.md](docs/reference/repository-structure.md) for complete directory guide.

**Key directories**:
- `.claude/` - Plugin entry point (slash commands, subagents)
- `capabilities/` - Capability source files (content for /meta command)
- `cmd/` - CLI and MCP server
- `internal/` - Core logic (parser, analyzer, query)
- `docs/` - Technical documentation

## Core Constraints

See [docs/core/principles.md](docs/core/principles.md) for complete details.

**Code Limits**:
- Phase: ≤500 lines of changes
- Stage: ≤200 lines of changes

**Development Methodology**:
- **TDD**: Write tests before implementation
- **Test Coverage**: ≥80%
- **Testing Protocol**: Run `make all` after each Stage

**Testing Failure Protocol**:
- If tests repeatedly fail → Stop immediately
- Document failure state and blockers
- Do NOT proceed until resolved

## Development Quick Start

### Build and Test

```bash
make all           # Lint + test + build
make test          # Run tests
make lint          # Static analysis
make test-coverage # Coverage report
```

**Before committing**:
1. Run `make all` to ensure code passes linting, tests, and builds
2. Fix any issues reported
3. Ensure all tests pass

### Plugin Development

**Local development setup**:
```bash
# 1. Edit source files
vim .claude/commands/meta.md       # Slash command
vim capabilities/commands/*.md     # Capabilities

# 2. For capability development
export META_CC_CAPABILITY_SOURCES="capabilities/commands"

# 3. Test in Claude Code (no build needed)

# 4. Run tests
make all
```

**See**: [docs/guides/plugin-development.md](docs/guides/plugin-development.md) for complete workflow.

### Version Management

**Three methods**:

1. **Git Hook (automatic)**:
   ```bash
   ./scripts/install-hooks.sh  # One-time setup
   # Then: git commit auto-bumps version on .claude/ changes
   ```

2. **Manual script**:
   ```bash
   ./scripts/bump-plugin-version.sh patch   # or minor/major
   ```

3. **Full release**:
   ```bash
   ./scripts/release.sh v1.0.0
   ```

**See**: [docs/guides/git-hooks.md](docs/guides/git-hooks.md) and [docs/guides/release-process.md](docs/guides/release-process.md).

### Commit Conventions

Use descriptive commit messages with scope prefixes:
- `docs:` for documentation changes
- `feat:` for new features
- `fix:` for bug fixes
- `refactor:` for code restructuring
- `test:` for test-related changes

Include the Claude Code attribution footer:
```
🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Common Development Tasks

### Fix Test Failures

```bash
# 1. Run tests
make test

# 2. Review errors
# (Analyze test output)

# 3. Fix issues
vim path/to/failing_test.go

# 4. Verify fix
make all
```

### Query Session Data (via MCP)

**NEW: Unified Query API (v2.0+)**:
```javascript
// Single composable tool
query({
  resource: "tools",
  filter: {tool_name: "Read", tool_status: "error"}
})
```

**Legacy queries** (backward compatible):
```
get_session_stats()                      # Session statistics
query_tools(status="error")              # Error tool calls
query_user_messages(pattern="fix.*bug")  # Search user messages
```

**See**:
- [Unified Query API Guide](docs/guides/unified-query-api.md) - New unified interface
- [Migration Guide](docs/guides/migration-to-unified-query.md) - Migrate from legacy tools
- [Query Cookbook](docs/examples/query-cookbook.md) - 10+ practical examples
- [MCP Guide](docs/guides/mcp.md) - Complete MCP reference

### Update Plugin

**Edit slash command**:
```bash
vim .claude/commands/meta.md
# Test in Claude Code immediately (no build needed)
git commit -m "feat: improve /meta matching"
# Git hook auto-bumps version
```

**Edit capability**:
```bash
vim capabilities/commands/meta-errors.md
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
# Test in Claude Code
git commit -m "feat: improve error analysis"
# No version bump (capability content change)
```

**See**: [docs/guides/plugin-development.md](docs/guides/plugin-development.md) for complete workflow.

## Unified Meta Command

The `/meta` command provides a unified entry point for 15+ capabilities with natural language intent matching.

**Usage**:
```
/meta "show errors"           → Executes meta-errors
/meta "quality check"         → Executes meta-quality-scan
/meta "visualize timeline"    → Executes meta-timeline
```

**Configuration**:
```bash
# Local development
export META_CC_CAPABILITY_SOURCES="capabilities/commands"

# Production (default)
# META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@main/commands"
```

**See**: [docs/reference/unified-meta-command.md](docs/reference/unified-meta-command.md) for complete guide.

## Reference Documentation

**Core Documentation**:
- [Implementation Plan](docs/core/plan.md) - Phase-by-phase roadmap
- [Design Principles](docs/core/principles.md) - Core constraints and rules
- [Plugin Development](docs/guides/plugin-development.md) - Complete plugin workflow
- [Repository Structure](docs/reference/repository-structure.md) - Directory organization
- [Release Process](docs/guides/release-process.md) - Release workflow
- [Git Hooks](docs/guides/git-hooks.md) - Automatic version bumping

**Integration and Usage**:
- [Integration Guide](docs/guides/integration.md) - Choose MCP/Slash/Subagent
- [MCP Guide](docs/guides/mcp.md) - Complete MCP server reference
- [Unified Meta Command](docs/reference/unified-meta-command.md) - /meta command guide
- [Capabilities Guide](docs/guides/capabilities.md) - Create custom capabilities

**Reference**:
- [CLI Reference](docs/reference/cli.md) - Complete command reference
- [JSONL Reference](docs/reference/jsonl.md) - Output format and jq patterns
- [Features](docs/reference/features.md) - Advanced features overview
- [Examples & Usage](docs/tutorials/examples.md) - Step-by-step tutorials
- [Troubleshooting](docs/guides/troubleshooting.md) - Common issues

**Architecture**:
- [Technical Proposal](docs/architecture/proposals/meta-cognition-proposal.md) - Architecture design
- [ADR Index](docs/architecture/adr/README.md) - Architecture decision records

**Universal Methodology** (project-independent):
- [Methodology Index](docs/methodology/) - Software development methodologies
- [Documentation Management](docs/methodology/documentation-management.md) - Documentation methodology for Claude Code projects

**Official Claude Code Documentation**:
- [Overview](https://docs.claude.com/en/docs/claude-code/overview)
- [Slash Commands](https://docs.claude.com/en/docs/claude-code/slash-commands)
- [Subagents](https://docs.claude.com/en/docs/claude-code/subagents)
- [MCP Integration](https://docs.claude.com/en/docs/claude-code/mcp)
- [Hooks System](https://docs.claude.com/en/docs/claude-code/hooks)

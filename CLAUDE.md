# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Quick Links

### New to meta-cc?
- **Start here**: [README.md](README.md) - Installation and quick start
- **Understand the design**: [docs/core/principles.md](docs/core/principles.md) - Core constraints
- **Integration guide**: [docs/guides/integration.md](docs/guides/integration.md) - Choose MCP/Slash

### Development Workflow
- **Current plan**: [docs/core/plan.md](docs/core/plan.md) - Phase roadmap and status
- **Build and test**: Run `make dev` (quick) → `make commit` (validate) → `make push` (full check)
- **Plugin development**: [docs/guides/plugin-development.md](docs/guides/plugin-development.md) - Complete workflow

### MCP Server Usage
- **MCP guide**: [docs/guides/mcp.md](docs/guides/mcp.md) - Complete MCP reference (16 tools)
- **Quick test**: Use MCP tool `get_session_stats`

### Common Tasks
- **Fix test failures**: `make dev` → Review errors → Fix → `make commit`
- **Query session data**: Use MCP tools (see [MCP Guide](docs/guides/mcp.md))
- **Update plugin**: [docs/guides/plugin-development.md](docs/guides/plugin-development.md)
- **Manage prompts**: `/prompt-list` (browse) | `/prompt-find <keywords>` (search) | `/prompt-show <id>` (view)

---

## FAQ

**Q: Tests failed after my changes - what should I do?**
A: Run `make dev` for quick iteration, then `make commit` to validate. Fix issues iteratively. If tests fail after multiple attempts, HALT development and document blockers.

**Q: How much code can I write in one phase?**
A: Maximum 500 lines of code modifications per phase, 200 lines per stage. See [docs/core/principles.md](docs/core/principles.md).

**Q: Should I use MCP or Slash Commands?**
A: Quick rule: Natural questions → MCP | Repeated workflows → Slash. See [docs/guides/integration.md](docs/guides/integration.md).

**Q: How do I query session data?**
A: Use consolidated query tools for common questions and the two-stage tools for custom jq:
```javascript
// Consolidated query tools
query_session_signals({type: "errors", limit: 10})
query_session_signals({type: "tokens", stats_first: true})
query_session_content({role: "user", pattern: "fix.*bug"})

// Custom jq over selected files
execute_stage2_query({
  files: ["/path/to/session.jsonl"],
  filter: 'select(.type == "assistant")',
  transform: '{timestamp, usage: .message.usage}'
})
```
See [MCP Query Tools Reference](docs/guides/mcp-query-tools.md) and [Two-Stage Query Guide](docs/guides/two-stage-query-guide.md) for complete documentation.

**Q: Why are my MCP query results in a temp file?**
A: Results >8KB automatically use file_ref mode to avoid token limits. Read the file with the Read tool. This is **hybrid output mode** - queries return inline for small results (<8KB) and file_ref for large results (≥8KB). See [MCP Query Tools Reference](docs/guides/mcp-query-tools.md#hybrid-output-mode).

**Q: Do I need to set `limit` parameter for MCP queries?**
A: No, by default queries return all results (hybrid mode handles large data). Only use `limit` when user explicitly requests a specific number. The system automatically switches to file_ref mode for large result sets.

**Q: Which MCP query tool should I use?**
A: Follow this decision tree:
- **Query messages/content** → Use `query_session_content(role=user|assistant|tool|all)`
- **Query signals/stats** → Use `query_session_signals(type=errors|tokens|system_errors|timestamps|tool_stats)`
- **Query file history** → Use `query_file_activity(type=snapshots)`
- **Maximum flexibility** → Use `get_session_directory` → `inspect_session_files` → `execute_stage2_query`
- **Cross-provider history** → Pass `provider: "claude"`, `provider: "codex"`, or `provider: "all"` to query and analysis tools
See [MCP Query Tools Reference](docs/guides/mcp-query-tools.md) for detailed guidance.

**Q: How do I write jq expressions for MCP queries?**
A: Start simple and add complexity:
```javascript
// Step 1: Select files
const dir = await get_session_directory({scope: "project"})

// Step 2: Filter by record shape
execute_stage2_query({
  files: dir.files,
  filter: 'select(.type == "assistant")'
})

// Step 3: Transform output
execute_stage2_query({
  files: dir.files,
  filter: 'select(.type == "user")',
  transform: '{timestamp, content: .message.content}'
})
```
Test jq locally first: `echo '{"type":"user"}' | jq 'select(.type == "user")'`. See [Two-Stage Query Guide](docs/guides/two-stage-query-guide.md) for common patterns.

**Q: How do I update plugin version?**
A: Install git hooks (`./scripts/install/install-hooks.sh`) for automatic bumping, or use `./scripts/release/bump-plugin-version.sh [patch|minor|major]`. See [docs/guides/git-hooks.md](docs/guides/git-hooks.md).

**Q: How does the prompt learning system work?**
A: Use `/prompt-find`, `/prompt-list`, and `/prompt-show` to manage a project-local library of optimized prompts stored in `.meta-cc/prompts/library/`. Browse your library with `/prompt-list`.

**Q: Where are saved prompts stored?**
A: Project-local storage in `.meta-cc/prompts/library/` (not tracked by git by default). You can commit selectively if you want to share with your team. The directory is auto-created on first save.

**Q: How do I find a saved prompt?**
A: Three methods (from fastest to most flexible):
1. **Direct search** (recommended): `/prompt-find <keywords>` - Fast, deterministic search
   - Example: `/prompt-find phase plan execute`
   - Uses keyword matching, no LLM overhead
2. **Manual search**: `grep`, `rg`, or file browser
   - Files are plain markdown in `.meta-cc/prompts/library/`

**Q: What if I don't want to save prompts?**
A: Saving is completely optional. Just press Enter or answer "n" when prompted. The save option won't appear again until you optimize another prompt.

**Q: How do I browse my saved prompts?**
A: Use the slash commands (recommended) or shell tools:
1. **List all prompts**: `/prompt-list`
   - Filter by category: `/prompt-list category=release`
   - Sort by usage: `/prompt-list sort=usage` (default, most used first)
   - Sort by date: `/prompt-list sort=date` (most recent first)
   - Sort alphabetically: `/prompt-list sort=alpha`
2. **View prompt details**: `/prompt-show <prompt-id>`
   - Example: `/prompt-show phase-execution-001`
   - Supports partial ID matching: `/prompt-show phase`
3. **Search prompts**: `/prompt-find <keywords>`
   - Example: `/prompt-find debug error analysis`
4. **Shell commands**: `ls -lt .meta-cc/prompts/library/` to see files by date
5. **Search tools**: `rg "keyword" .meta-cc/prompts/library/` to search content

**Slash Command Summary**:
- `/prompt-find <keywords>` - Search for prompts
- `/prompt-list [category=X] [sort=usage|date|alpha]` - List all prompts
- `/prompt-show <prompt-id>` - View full prompt details

**Q: Can I delete or edit saved prompts?**
A: Yes, they're just markdown files:
- **Delete**: `rm .meta-cc/prompts/library/release-simple-001.md`
- **Edit**: `vim .meta-cc/prompts/library/release-simple-001.md`
- **Archive**: Edit YAML frontmatter, set `status: archived`

**Q: Can I share prompts with my team?**
A: Yes, commit to git:
```bash
git add .meta-cc/prompts/library/release-*.md
git commit -m "docs: share release process prompts"
git push
```

**Q: How do I back up my prompt library?**
A: Simple directory copy:
```bash
# Backup
cp -r .meta-cc/prompts ~/backups/project-prompts-$(date +%Y%m%d)

# Restore
cp -r ~/backups/project-prompts-20251027/.meta-cc/prompts .meta-cc/
```

**Q: Can I use prompts across multiple projects?**
A: Currently project-local. A future update may add global library in `~/.meta-cc/` for cross-project sharing.

---

## Project Overview

**meta-cc** (Meta-Cognition for Claude Code) analyzes Claude Code session history to provide metacognitive insights and workflow optimization.

### Architecture

**MCP-based architecture**:
- **MCP Server**: Provides 16 tools for session history analysis and query
- **Claude Integration**: Slash commands for prompt management

**Key principle**: MCP server handles data extraction and query. Claude performs semantic understanding and recommendations.

### Repository Structure

See [docs/reference/repository-structure.md](docs/reference/repository-structure.md) for complete directory guide.

**Key directories**:
- `.claude/` - Plugin entry point (slash commands)
- `cmd/mcp-server/` - MCP server implementation
- `internal/` - Core logic (parser, analyzer, query)
- `docs/` - Technical documentation

## Core Constraints

See [docs/core/principles.md](docs/core/principles.md) for complete details.

**Code Limits**:
- Phase: ≤500 lines of code modifications
- Stage: ≤200 lines of code modifications

**Development Methodology**:
- **TDD**: Write tests before implementation
- **Test Coverage**: ≥80%
- **Testing Protocol**: Run `make commit` after each Stage

**Testing Failure Protocol**:
- If tests repeatedly fail → Stop immediately
- Document failure state and blockers
- Do NOT proceed until resolved

## Development Quick Start

### Build and Test

```bash
make dev           # Quick dev build (format + build, <10s)
make commit        # Pre-commit validation (workspace + tests, <60s)
make push          # Full check before push (all checks + lint, <120s)
make test          # Run tests only
make lint          # Static analysis
make test-coverage # Coverage report
```

**Before committing**:
1. Run `make commit` to ensure code passes essential validation
2. Fix any issues reported
3. Before pushing, run `make push` for full verification

### Plugin Development

**Local development setup**:
```bash
# 1. Edit source files
vim .claude/commands/prompt-find.md   # Slash command
vim .claude/commands/prompt-list.md   # Slash command
vim .claude/commands/prompt-show.md   # Slash command

# 2. Test in Claude Code (no build needed)

# 3. Run tests
make commit
```

**See**: [docs/guides/plugin-development.md](docs/guides/plugin-development.md) for complete workflow.

### Version Management

**Three methods**:

1. **Git Hook (automatic)**:
   ```bash
   ./scripts/install/install-hooks.sh  # One-time setup
   # Then: git commit auto-bumps version on .claude/ changes
   ```

2. **Manual script**:
   ```bash
   ./scripts/release/bump-plugin-version.sh patch   # or minor/major
   ```

3. **Full release**:
   ```bash
   ./scripts/release/release.sh v1.0.0
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
make commit
```

### Query Session Data (via MCP)

**Consolidated query tools**:
```javascript
query_session_signals({type: "errors", limit: 10})
query_session_signals({type: "tool_stats", tool: "Read", limit: 20})
query_session_content({role: "user", pattern: "fix.*bug"})
```

**Two-stage jq**:
```javascript
const dir = await get_session_directory({scope: "project"})
execute_stage2_query({
  files: dir.files,
  filter: 'select(.type == "assistant")',
  limit: 20
})
```

**See**:
- [MCP Query Tools Reference](docs/guides/mcp-query-tools.md) - Current MCP tool reference
- [Two-Stage Query Guide](docs/guides/two-stage-query-guide.md) - Custom jq workflow
- [MCP Guide](docs/guides/mcp.md) - Complete MCP reference

### Update Plugin

**Edit slash command**:
```bash
vim .claude/commands/prompt-find.md
# Test in Claude Code immediately (no build needed)
git commit -m "feat: improve prompt-find matching"
# Git hook auto-bumps version
```

**See**: [docs/guides/plugin-development.md](docs/guides/plugin-development.md) for complete workflow.

## Reference Documentation

**Core Documentation**:
- [Implementation Plan](docs/core/plan.md) - Phase-by-phase roadmap
- [Design Principles](docs/core/principles.md) - Core constraints and rules
- [Plugin Development](docs/guides/plugin-development.md) - Complete plugin workflow
- [Repository Structure](docs/reference/repository-structure.md) - Directory organization
- [Release Process](docs/guides/release-process.md) - Release workflow
- [Git Hooks](docs/guides/git-hooks.md) - Automatic version bumping

**Integration and Usage**:
- [Integration Guide](docs/guides/integration.md) - Choose MCP/Slash
- [MCP Guide](docs/guides/mcp.md) - Complete MCP server reference (16 tools)

**Reference**:
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
- [MCP Integration](https://docs.claude.com/en/docs/claude-code/mcp)
- [Hooks System](https://docs.claude.com/en/docs/claude-code/hooks)

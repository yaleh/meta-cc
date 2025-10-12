# Plugin Development Guide

Complete guide for developing and testing Claude Code plugins in meta-cc.

## Quick Start

### Local Development Setup

1. **Edit source files** (changes reflect immediately):
   ```bash
   vim .claude/commands/meta.md      # Slash command
   vim .claude/agents/*.md           # Subagents
   vim capabilities/commands/*.md    # Capabilities
   ```

2. **Configure local capability source**:
   ```bash
   export META_CC_CAPABILITY_SOURCES="capabilities/commands"
   ```

3. **Test in Claude Code** (no build needed)

4. **Run tests** before committing:
   ```bash
   make all
   ```

## Plugin Structure

### Files and Directories

```
.claude/
├── commands/
│   └── meta.md              # Unified /meta command
├── agents/
│   ├── project-planner.md   # TDD planning agent
│   └── stage-executor.md    # Stage execution agent
└── hooks/                   # Optional project hooks

capabilities/
└── commands/
    ├── meta-errors.md       # Error analysis capability
    ├── meta-quality-scan.md # Quality scanning
    └── ... (13 capabilities)

.claude-plugin/
├── plugin.json              # Plugin manifest
└── marketplace.json         # Marketplace listing
```

### Plugin Manifest (plugin.json)

```json
{
  "name": "meta-cc",
  "version": "0.26.9",
  "description": "Meta-Cognition tool for Claude Code",
  "author": {
    "name": "Yale Huang",
    "email": "yaleh@ieee.org",
    "url": "https://github.com/yaleh"
  },
  "license": "MIT",
  "homepage": "https://github.com/yaleh/meta-cc",
  "repository": "https://github.com/yaleh/meta-cc",
  "keywords": ["workflow-analysis", "session-history", "productivity"],
  "commands": ["./.claude/commands/meta.md"],
  "agents": [
    "./.claude/agents/project-planner.md",
    "./.claude/agents/stage-executor.md"
  ]
}
```

**Key fields**:
- `commands`: Slash command files (relative to plugin root)
- `agents`: Subagent definition files
- `version`: Plugin version (updated by scripts, see [Version Management](#version-management))

## Development Workflow

### 1. Edit Slash Commands

**File**: `.claude/commands/meta.md`

**Format**:
```markdown
---
name: meta
description: Unified meta-cognition command
keywords: meta, capability, semantic
category: unified
---

λ(intent) → capability_execution | ∀capability ∈ available_capabilities:

execute :: intent → output
execute(I) = discover(I) ∧ match(I) ∧ report(I) ∧ run(I)
...
```

**Testing**:
```bash
# Edit file
vim .claude/commands/meta.md

# Test in Claude Code immediately (no build)
# In Claude Code: /meta "show errors"
```

**No build needed**: Claude Code reads `.claude/commands/*.md` directly.

### 2. Edit Capabilities

**Directory**: `capabilities/commands/`

**Local development configuration**:
```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

**Why**: Without this env var, capabilities load from GitHub (cached). Local changes won't reflect immediately.

**Example workflow**:
```bash
# Edit capability
vim capabilities/commands/meta-errors.md

# Set local source
export META_CC_CAPABILITY_SOURCES="capabilities/commands"

# Test in Claude Code
# In Claude Code: /meta "show errors"
# Changes reflect immediately (no cache)
```

### 3. Edit Subagents

**Directory**: `.claude/agents/`

**Example workflow**:
```bash
# Edit subagent
vim .claude/agents/project-planner.md

# Test in Claude Code immediately
# In Claude Code: @project-planner "plan a new feature"
```

**Agent definition format**:
```markdown
---
name: project-planner
description: Analyzes project documentation and generates TDD plans
keywords: planning, tdd, iteration
---

# Project Planner Agent

You are a project planning agent...
```

## Version Management

### Three Version Update Methods

| Method | Trigger | Use Case | Version Type |
|--------|---------|----------|--------------|
| **Git Hook** | Automatic (on commit) | Most `.claude/` changes | Patch only |
| **bump-plugin-version.sh** | Manual | Need minor/major bump | Patch/Minor/Major |
| **release.sh** | Manual | Full release | Any version |

### Method 1: Git Hook (Automatic)

**Setup** (one-time):
```bash
./scripts/install-hooks.sh
```

**Usage**:
```bash
# Edit .claude/ file
vim .claude/commands/meta.md

# Stage and commit
git add .claude/commands/meta.md
git commit -m "feat: improve semantic matching"

# Hook auto-bumps version: 0.26.9 → 0.26.10
# Includes version files in same commit
```

**When it triggers**:
- ✅ `.claude/commands/*.md` changes
- ✅ `.claude/agents/*.md` changes
- ❌ `capabilities/commands/*.md` changes (no version bump)

**See**: [Git Hooks Guide](git-hooks.md) for details.

### Method 2: Manual Script (Flexible)

**Use when**: Need minor/major version bump.

```bash
# Edit .claude/ file
vim .claude/commands/meta.md

# Choose version bump type
./scripts/bump-plugin-version.sh patch   # 0.26.9 → 0.26.10
./scripts/bump-plugin-version.sh minor   # 0.26.9 → 0.27.0
./scripts/bump-plugin-version.sh major   # 0.26.9 → 1.0.0

# Commit changes
git add .claude/commands/meta.md
git commit -m "feat: add new feature"
# Hook skips (version already updated)
```

### Method 3: Full Release (Complete)

**Use when**: Releasing CLI + MCP + Plugin together.

```bash
# Edit files
vim cmd/mcp-server/main.go .claude/commands/meta.md

# Full release
./scripts/release.sh v0.28.0
# Prompts to update CHANGELOG.md
# Creates git tag, triggers GitHub Actions
```

**See**: [Release Process](release-process.md) for details.

## Testing

### Local Testing

**Test slash commands**:
```bash
# In Claude Code
/meta "show errors"
/meta "quality check"
```

**Test subagents**:
```bash
# In Claude Code
@project-planner "plan a new feature"
@stage-executor "execute stage 1"
```

**Test capabilities locally**:
```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
# In Claude Code
/meta "show errors"
```

### Unit Tests

```bash
make test          # Run unit tests
make test-all      # Run all tests
make test-coverage # With coverage report
```

### Integration Tests

```bash
make test-integration
```

## Build and Release

### Sync Plugin Files

**Purpose**: Merge `.claude/` + `capabilities/` → `dist/`

```bash
make sync-plugin-files
```

**Creates**:
- `dist/commands/meta.md` (merged from `.claude/commands/`)
- `dist/agents/*.md` (merged from `.claude/agents/`)

**When to run**: During release process (automatic in CI).

### Bundle Release

```bash
make bundle-release VERSION=v1.0.0
```

**Creates**:
- `build/meta-cc-plugin-{version}-{platform}.tar.gz`
- Cross-platform binaries
- Capability packages

### Full Release Workflow

See [Release Process](release-process.md) for complete workflow:

1. Update version (bump script or git hook)
2. Update CHANGELOG.md
3. Run `./scripts/release.sh v1.0.0`
4. GitHub Actions builds and publishes

## Capability Development

### Capability File Format

**Location**: `capabilities/commands/meta-{name}.md`

**Example** (`meta-errors.md`):
```markdown
---
name: meta-errors
description: Analyze session errors and provide actionable debugging guidance
keywords: error, debug, troubleshoot, failure, bug
category: debugging
---

# Meta Errors Capability

Execute :: scope → error_analysis

discover_errors :: scope → ErrorList
discover_errors(S) = {
  stats: mcp_meta_cc.get_session_stats(scope=S),
  errors: mcp_meta_cc.query_tools(status="error", scope=S),

  # Error detection logic...
}
```

**Frontmatter fields**:
- `name`: Capability identifier (used in matching)
- `description`: What the capability does
- `keywords`: Search keywords for semantic matching
- `category`: Capability category (debugging, analysis, visualization, etc.)

### Local Development

**Setup**:
```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

**Benefits**:
- Changes reflect immediately (no cache)
- Fast iteration
- No network dependencies

**Workflow**:
```bash
# 1. Edit capability
vim capabilities/commands/meta-errors.md

# 2. Test immediately
# In Claude Code: /meta "show errors"

# 3. Iterate
# Repeat steps 1-2 until satisfied

# 4. Commit (no version bump needed for capabilities)
git add capabilities/commands/meta-errors.md
git commit -m "feat: improve error analysis"
```

### Multi-Source Configuration

For advanced capability development:

```bash
# Local + GitHub fallback
export META_CC_CAPABILITY_SOURCES="~/dev/capabilities:yaleh/meta-cc@main/commands"

# Package + Local
export META_CC_CAPABILITY_SOURCES="./capabilities.tar.gz:capabilities/commands"
```

**Priority**: Left-to-right (left = highest priority).

**See**: [Unified Meta Command](unified-meta-command.md) for details.

## Common Tasks

### Add New Slash Command

1. Create `.claude/commands/{name}.md`
2. Add to `plugin.json`:
   ```json
   "commands": [
     "./.claude/commands/meta.md",
     "./.claude/commands/{name}.md"
   ]
   ```
3. Test in Claude Code
4. Commit and bump version (auto via git hook)

### Add New Subagent

1. Create `.claude/agents/{name}.md`
2. Add to `plugin.json`:
   ```json
   "agents": [
     "./.claude/agents/project-planner.md",
     "./.claude/agents/{name}.md"
   ]
   ```
3. Test in Claude Code
4. Commit and bump version (auto via git hook)

### Add New Capability

1. Create `capabilities/commands/meta-{name}.md`
2. Test locally:
   ```bash
   export META_CC_CAPABILITY_SOURCES="capabilities/commands"
   ```
3. Commit (no version bump needed)
4. Deploy: Merged into production via GitHub

### Update Plugin Metadata

**Edit** `.claude-plugin/plugin.json`:
```json
{
  "description": "New description",
  "keywords": ["new", "keywords"]
}
```

**Commit**:
```bash
git add .claude-plugin/plugin.json
git commit -m "docs: update plugin metadata"
# Git hook bumps version (metadata change)
```

## Troubleshooting

### Changes Not Reflecting in Claude Code

**Problem**: Edited capability but no change in behavior.

**Solution**: Set local capability source:
```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

**Explanation**: By default, capabilities load from GitHub (cached). Local source disables cache.

### Version Not Auto-Bumping

**Problem**: Committed `.claude/` changes but version unchanged.

**Solution**: Install git hooks:
```bash
./scripts/install-hooks.sh
```

**Verify**:
```bash
ls -l .git/hooks/pre-commit
```

### Build Errors

**Problem**: `make all` fails.

**Solution**:
```bash
# Check linting
make lint

# Check tests
make test

# Check build
make build
```

Fix errors iteratively. See [Testing Failure Protocol](principles.md#testing-failure-protocol).

## Best Practices

### Version Bumping

- **Capabilities**: No version bump (content changes)
- **Slash commands/agents**: Auto-bump via git hook (framework changes)
- **Minor features**: Use `./scripts/bump-plugin-version.sh minor`
- **Breaking changes**: Use `./scripts/bump-plugin-version.sh major`

### Capability Development

- Use local source for development (`export META_CC_CAPABILITY_SOURCES=...`)
- Test thoroughly before committing
- Follow frontmatter format (name, description, keywords, category)
- Use semantic keywords for matching

### Git Workflow

- Install git hooks for automatic version bumping
- Commit `.claude/` and `capabilities/` changes separately
- Run `make all` before committing
- Use conventional commit messages (feat:, fix:, docs:)

## See Also

- [Release Process](release-process.md) - Complete release workflow
- [Git Hooks](git-hooks.md) - Automatic version bumping
- [Repository Structure](repository-structure.md) - Directory organization
- [Unified Meta Command](unified-meta-command.md) - /meta command details
- [Capabilities Guide](capabilities-guide.md) - Creating custom capabilities

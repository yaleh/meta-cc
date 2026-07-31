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

## Installing the Plugin

`make install-local` and `make install-user` both depend on `make stage`,
which builds the MCP server and copies it to `plugin-src/bin/meta-cc-mcp`.
`stage` writes the new binary to a temp file in the same directory and
renames it into place, rather than overwriting the target in place -- this
means it succeeds even if a running MCP server process (e.g. a live Claude
Code or Codex session) currently has the previous `plugin-src/bin/meta-cc-mcp`
open; the running process keeps using its own (now-unlinked) copy until it
exits, and a fresh session picks up the new binary. You never need to
manually find and kill an MCP server process before running `make stage`,
`make install-local`, or `make install-user`.

### Local scope (this project only)

```bash
make install-local     # stage + generate .claude/settings.local.json + purge cache
```

### User scope (all projects, this machine)

```bash
make install-user      # stage + copy to ~/.local/share/meta-cc + register with Claude Code
```

`make install-user` only updates Claude Code's own registration
(`~/.claude/settings.json`). If you also use Codex against this same
machine, Codex has its own separate registration
(`~/.codex/config.toml`'s `[marketplaces.meta-cc-marketplace]` `source`),
which `make install-user` does **not** touch -- otherwise Codex would keep
running whatever binary its `source`/`source_type` already point at (often
a dev-mode path straight into this repo), silently drifting from the
user-scope binary Claude Code just switched to.

To keep Codex in sync, run the sibling target after each `make install-user`:

```bash
make install-user-codex
```

The target first updates only the `source` (and `source_type`, defaulting to
`"local"`) keys inside `~/.codex/config.toml`'s
`[marketplaces.meta-cc-marketplace]` table to point at
`~/.local/share/meta-cc`, via
`scripts/install/update-codex-marketplace-toml.py`. Everything else in
`~/.codex/config.toml` -- other `[marketplaces.*]`, `[model_providers.*]`,
`[projects.*]`, `[plugins.*]` entries, comments, formatting -- is left
byte-for-byte unchanged; the update is a targeted line-level edit, not a full
TOML rewrite.

When that marketplace is registered, the target then invokes Codex's supported
`codex plugin add meta-cc@meta-cc-marketplace --json` refresh flow and verifies
that every discovery surface resolves the same version: the discovery metadata
(`plugin list`), both cached manifests, the source manifests, all three skills,
and the MCP binary's **self-reported** version (queried over MCP `initialize`
-- executability alone is not enough).

Refresh is **destructive** on real Codex CLI 0.146.0: `codex plugin add`
materializes the new version and removes the previously cached version *before*
post-install verification runs. The script therefore performs **no automatic
rollback** -- on verification failure the old cache is already gone. Instead it
exits non-zero with an actionable error naming the missing/inconsistent
artifact and an explicit manual recovery command (re-run the refresh). The only
failures that leave the prior cache intact are the pre-destructive ones: a
missing Codex CLI, or a CLI that lacks the plugin-manager commands (detected
via `--help` probes before any `plugin add`). Tests can set `CODEX_HOME` to an
isolated directory; the target honors it and never needs to inspect a developer
cache.

A successful update still requires a **new Codex session**. Running sessions
keep the skills and MCP process resolved when they started and do not hot-reload
the upgraded plugin; a successful, fully verified upgrade leaves no dangling
old-version path for the next session to warn about.

The refreshed MCP server reads Codex history from the canonical root resolved
as `META_CC_CODEX_ROOT` → `CODEX_HOME` → `~/.codex` (highest-compatible
`state_N.sqlite`, with a rollout-only fallback) — see
[Discovery roots](../reference/codex-history-model.md#discovery-roots-and-database-compatibility-dir-069)
and [Codex App-Server Backend](../reference/codex-app-server.md) for the
storage and backend-mode contract.

These behaviors -- `stage`'s locked-binary safety, the Codex TOML update, and
the destructive upgrade/restart and verification-failure semantics -- are
covered by `tests/scripts/stage-locked-binary.bats`,
`tests/scripts/codex-registration.bats`, and
`tests/scripts/codex-user-upgrade.bats`, run via `make test-bats` (wired
into `make push`). The upgrade tests drive a fake Codex double that reproduces
the real CLI's destructive replacement (old cache removed before verification).

### Verifying which commit/build is actually installed

`Makefile`'s `build`, `install`, and `cross-compile` targets all embed the
exact commit and build timestamp into the binary via linker `-X` flags
(`internal/version.Commit` / `internal/version.BuildTime`, alongside the
existing hand-bumped `internal/version.Version`). The running MCP server
logs all three in its startup entry:

```json
{"msg":"MCP server starting","server_name":"meta-cc-mcp","version":"3.5.1","commit":"78aea25","build_time":"2026-07-28_16:12:37"}
```

This closes a real gap (DIR-049): before this, the same un-bumped
`version.Version` could be reported by two materially different builds,
with no way to tell them apart. Now `commit` should match
`git rev-parse --short HEAD` at build time, so you can confirm exactly
which source snapshot is actually running.

Since `install-local`/`install-user` copy the binary at `stage`/`build`
time (not at every session start), **a user-scope install only picks up
new commits when you re-run `make install-user`** (and similarly
`make install-local` for the local scope) -- restarting Claude Code alone
reloads the *existing* installed binary, it does not rebuild it. If a
session's logged `commit` doesn't match your current `git rev-parse --short
HEAD`, that's the signal you need to reinstall.

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
  stats: mcp_meta_cc.get_timeline(scope=S, stats_only=True),
  errors: mcp_meta_cc.query_session_signals(type="tool_stats", status="error", scope=S),

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

**See**: [Unified Meta Command](../reference/unified-meta-command.md) for details.

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

Fix errors iteratively. See [Testing Failure Protocol](../core/principles.md#testing-failure-protocol).

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

## Worktree Cache Environment

When the quay loop driver dispatches tasks to git worktree subagents
(`execution: dispatched` in `.quay/config.yml`), each task builds in an
isolated worktree under `/tmp/meta-cc-worktrees/<id>`. Go's build and test
caches (GOCACHE, GOMODCACHE) are **content-addressed and shared across
worktrees by default** — they resolve to paths under `$HOME`, not to the
worktree directory:

| Variable   | Typical value (Linux)         | What it caches          |
|------------|-------------------------------|-------------------------|
| GOCACHE    | `$HOME/.cache/go-build`       | Build + test artifacts  |
| GOMODCACHE | `$HOME/go/pkg/mod`            | Downloaded modules      |
| GOPATH     | `$HOME/go`                    | Installed binaries      |
| GOFLAGS    | (empty)                       | No overrides            |

This means a second worktree's `go build ./...` or `go test ./...` reuses
compiled packages from the first worktree, keeping incremental builds fast.
The environment is verified by `scripts/checks/check-worktree-cache-env.sh`,
which:

- **PASS**: GOCACHE and GOMODCACHE both resolve to shared paths under `$HOME`.
- **FAIL**: Either cache points into `/tmp` (per-worktree isolation), `HOME`
  is unset, or `GOFLAGS` contains a worktree-directed override.
- **WARN** (non-fatal): The cache directories exist but are empty or cold —
  expected on a first build in a new environment.

To confirm cache sharing manually, build in one worktree and then build the
same code in a second worktree — the second build should be markedly faster
and Go's `-x` trace will show cached package lookups:

```bash
# First worktree (cold or partially cold)
cd /tmp/meta-cc-worktrees/<id>
time go build ./...

# Second worktree (should be hot)
cd /tmp/meta-cc-worktrees/<id2>
time go build -x ./... 2>&1 | grep -c 'cached'
```

## See Also

- [Release Process](release-process.md) - Complete release workflow
- [Git Hooks](git-hooks.md) - Automatic version bumping
- [Repository Structure](../reference/repository-structure.md) - Directory organization
- [Unified Meta Command](../reference/unified-meta-command.md) - /meta command details
- [Capabilities Guide](capabilities.md) - Creating custom capabilities

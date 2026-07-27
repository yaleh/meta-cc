# Release Process

This document describes the release process for meta-cc.

## Overview

meta-cc has **two types of releases** with different versioning strategies:

1. **Plugin-only updates** - When `.claude/` files change (slash commands, subagents)
2. **Full releases** - When CLI/MCP/Plugin all need updating together

This separation allows:
- **High-frequency capability updates** (in `capabilities/`) without triggering plugin version changes
- **Low-frequency plugin updates** (in `.claude/`) only when framework/API changes
- **Independent versioning** for plugin structure vs. capability content

## Prerequisites

1. **Required tools**:
   - `jq` - JSON processor for updating version files
   - `git` - Version control
   - `make` - Build automation

2. **Install jq** (if not already installed):
   ```bash
   # Ubuntu/Debian
   sudo apt-get install jq

   # macOS
   brew install jq

   # Windows (via Chocolatey)
   choco install jq
   ```

## Version and Tool-Count Source of Truth (DIR-025)

meta-cc ships several version surfaces (Claude Code plugin manifest, Claude
Code marketplace manifest, Codex plugin manifest, and the MCP server's own
`initialize`/`serverInfo` response) and several tool-count claims (README,
CLAUDE.md, marketplace descriptions, the MCP guide). Historically these
drifted independently — see `tasks/DIR-025.md` for the finding. They are now
kept in agreement as follows:

- **`internal/version/release.json`** is the one hand-edited source of
  truth for the current release version. It is embedded at compile time
  into the MCP server binary via `internal/version` (`go:embed`), so the
  server's `initialize` response and startup log always reflect this file —
  there is no separate hardcoded server version to forget.
- **The MCP tool count is never hand-maintained as an independent number.**
  It is always read from `internal/mcp/tools.GetToolDefinitions()` (the same
  function that answers the live `tools/list` request). Docs and manifests
  still contain a literal number (JSON/Markdown can't run Go code), but that
  number is checked against the real count by an automated test.
- **`internal/release`** contains the offline consistency tests
  (`go test ./internal/version/... ./internal/release/...`) that:
  - assert `.claude-plugin/marketplace.json`, `plugin-src/.claude-plugin/marketplace.json`,
    `plugin-src/.claude-plugin/plugin.json`, and `plugin-src/.codex-plugin/plugin.json`
    all have `version == internal/version/release.json`'s version, and
  - assert every tool-count claim in the *current* docs/manifests listed in
    `currentToolCountDocs` (`internal/release/consistency_test.go`) equals
    `len(tools.GetToolDefinitions())`.

  These tests run automatically with `go test ./...` (and therefore
  `make test`, `make commit`, and CI), and are also invoked explicitly by
  `make check-release-ready`, `scripts/sync-plugin-files.sh --verify`, and
  `scripts/release/pre-release-check.sh`. No network access is required.
  Historical documents (`docs/proposals/`, `docs/plans/`, `docs/phases/`,
  `docs/archive/`, `CHANGELOG.md`, `plans/`, `tasks/`) are intentionally
  **not** checked — they describe past states, not current claims.

**To bump the version**, run `./scripts/release/bump-plugin-version.sh
[patch|minor|major]` (or `--version X.Y.Z`). It now updates all five
surfaces in one step: `plugin-src/.claude-plugin/plugin.json`,
`.claude-plugin/marketplace.json`, `plugin-src/.claude-plugin/marketplace.json`,
`plugin-src/.codex-plugin/plugin.json`, and `internal/version/release.json`.
`scripts/release/release.sh` calls it in `--non-interactive` mode as part of
a full release.

**To add or remove an MCP tool**, edit
`internal/mcp/tools.GetToolDefinitions()`, then run
`go test ./internal/release/...`: it will fail once per stale doc/manifest
tool-count claim it finds, each with a "Fix:" message naming the exact file
and the number to write. Update `README.md`, `CLAUDE.md`,
`docs/DOCUMENTATION_MAP.md`, `docs/guides/integration.md`,
`docs/guides/mcp.md`, and the marketplace/plugin manifest descriptions
accordingly, and re-run the test until it passes. If a new *current-facing*
doc gains a tool-count claim, add it to `currentToolCountDocs` in
`internal/release/consistency_test.go` so future drift is caught there too.

## Release Workflows

### Workflow A: Plugin-Only Version Bump

**When to use**: You modified `.claude/commands/*.md` or `.claude/agents/*.md`

**Examples**:
- Updated `/meta` command logic (semantic matching algorithm)
- Added new subagent
- Modified existing subagent behavior
- Changed plugin metadata

**Do NOT use for**:
- Capability content updates (`capabilities/commands/*.md`)
- CLI/MCP code changes

**Steps**:

1. **Prepare**:
   ```bash
   git checkout develop   # Or main for hotfixes
   git status            # Ensure clean working directory
   ```

2. **Bump plugin version**:
   ```bash
   ./scripts/bump-plugin-version.sh patch   # For bug fixes
   # or
   ./scripts/bump-plugin-version.sh minor   # For new features
   # or
   ./scripts/bump-plugin-version.sh major   # For breaking changes
   ```

3. **Review and push**:
   ```bash
   git show HEAD         # Review the version bump commit
   git push origin develop
   ```

**What the script does**:
- ✅ Validates current branch (main or develop)
- ✅ Checks working directory is clean
- ✅ Increments version in plugin.json and marketplace.json
- ✅ Commits changes with proper attribution
- ❌ Does NOT create git tag (plugin versions are not tagged separately)
- ❌ Does NOT trigger release build

**Version progression example**:
```
0.26.8 → patch → 0.26.9
0.26.9 → minor → 0.27.0
0.27.0 → major → 1.0.0
```

---

### Workflow B: Full Release (CLI + MCP + Plugin)

**When to use**: Creating a complete release with CLI/MCP binaries + Plugin

**Steps**:

1. **Prepare**:
   ```bash
   git checkout main      # For stable releases
   # or
   git checkout develop   # For beta/RC releases

   git status            # Ensure clean
   git pull origin main  # Sync with remote
   ```

2. **Run release script**:
   ```bash
   ./scripts/release.sh v1.0.0
   ```

3. **Update CHANGELOG.md** when prompted:
   - Add release notes for the new version
   - Follow the format shown in CHANGELOG.md Format section below

**What the script does**:
1. ✅ Validates version format (`vX.Y.Z` or `vX.Y.Z-beta`)
2. ✅ Checks branch (must be `main` or `develop`)
3. ✅ Checks working directory is clean
4. ✅ Runs full test suite (`make all`)
5. ✅ Updates `plugin.json` version
6. ✅ Updates `marketplace.json` version
7. ✅ Prompts you to update `CHANGELOG.md`
8. ✅ Commits version updates with attribution
9. ✅ Creates git tag `vX.Y.Z`
10. ✅ Pushes commit and tag to remote

4. **Monitor GitHub Actions**:

After pushing the tag, GitHub Actions automatically:
- ✅ Verifies version consistency (plugin.json, marketplace.json, git tag)
- ✅ Builds cross-platform binaries (Linux, macOS, Windows)
- ✅ Creates plugin packages with MCP server
- ✅ Bundles capability packages
- ✅ Publishes GitHub Release

Monitor progress at: https://github.com/yaleh/meta-cc/actions

---

## Decision Tree: Which Workflow to Use?

```
Changed files?
│
├─ .claude/commands/*.md or .claude/agents/*.md
│  └─> Use Workflow A (bump-plugin-version.sh)
│
├─ capabilities/commands/*.md
│  └─> No version bump needed
│      Just commit and push
│      Capabilities load from GitHub/cache
│
├─ cmd/, internal/, pkg/ (Go code)
│  └─> Use Workflow B (release.sh)
│      Full release with binaries
│
└─ Multiple components changed
   └─> Use Workflow B (release.sh)
       Ensures all versions stay in sync
```

## Examples

### Example 1: Updated /meta command algorithm

**Scenario**: Improved semantic matching in `.claude/commands/meta.md`

```bash
# Edit .claude/commands/meta.md
git add .claude/commands/meta.md
git commit -m "feat: improve semantic matching in /meta command"

# Bump plugin version
./scripts/bump-plugin-version.sh minor

# Push
git push origin develop
```

**Result**: Plugin version 0.26.8 → 0.27.0, no git tag created

---

### Example 2: Added new capability

**Scenario**: Created `capabilities/commands/meta-performance.md`

```bash
# Edit capabilities/commands/meta-performance.md
git add capabilities/commands/meta-performance.md
git commit -m "feat: add performance analysis capability"

# Push (no version bump needed)
git push origin develop
```

**Result**: No version change. Users get new capability via GitHub source refresh.

---

### Example 3: Full release with MCP changes

**Scenario**: Added new MCP query tool + updated plugin

```bash
# Changes in cmd/mcp-server/ and .claude/commands/meta.md
git add cmd/mcp-server/ .claude/commands/meta.md
git commit -m "feat: add new MCP query tool and update /meta"

# Full release
./scripts/release.sh v0.28.0

# [Update CHANGELOG.md when prompted]
```

**Result**: Git tag v0.28.0 created, GitHub Release built with binaries

## CHANGELOG.md Format

The script expects CHANGELOG.md to follow this format:

```markdown
# Changelog

## [0.27.0] - 2025-10-13

### Added
- New feature description

### Changed
- Changed functionality description

### Fixed
- Bug fix description

## [0.26.8] - 2025-10-12
...
```

**Important**: Use `[0.27.0]` format (without `v` prefix) in CHANGELOG.md.

## Version Management Strategy

### Why Two Workflows?

**Problem**: Originally, every change (capabilities, plugin, CLI) required a full release.

**Issues**:
1. High-frequency capability updates triggered unnecessary binary builds
2. Plugin version changed even when only capability content was updated
3. Users saw version bumps for trivial documentation changes

**Solution**: Separate plugin version from capability content.

**Key Insights**:
- **Plugin version** tracks `.claude/` structure (slash commands, subagents)
- **Capability content** updates don't affect plugin version
- **Capabilities load dynamically** from GitHub (no version coupling)

**Benefits**:
- ✅ Capability updates: Just commit and push (no version bump)
- ✅ Plugin updates: Simple script, no full release needed
- ✅ Full releases: Only when binaries actually change

### Version Synchronization

**Current approach**: plugin.json version is **manually maintained**.

- **Workflow A** (bump-plugin-version.sh): Updates plugin.json only
- **Workflow B** (release.sh): Updates plugin.json + creates git tag

**Why not auto-sync from git tags?**
- Plugin updates happen more frequently than full releases
- Not all git tags represent plugin changes (might be CLI-only updates)
- Manual control gives flexibility for plugin-specific versioning

**Trade-off accepted**: Manual version bumping is acceptable because:
- Plugin structure changes are infrequent (low maintenance cost)
- Scripts automate the process (minimal manual work)
- Clear separation improves clarity

## Troubleshooting

### Error: "Working directory not clean"

**Problem**: Uncommitted changes in working directory.

**Solution**:
```bash
# Check what's changed
git status

# Either commit changes
git add .
git commit -m "fix: description"

# Or stash them
git stash
```

### Error: "jq is required but not installed"

**Problem**: `jq` command not found.

**Solution**: Install jq (see Prerequisites section).

### Question: "Should I bump plugin version for capability changes?"

**Answer**: **No**. Capabilities in `capabilities/commands/*.md` load dynamically from GitHub.

Only bump plugin version when:
- ✅ `.claude/commands/*.md` changes (e.g., /meta command logic)
- ✅ `.claude/agents/*.md` changes (e.g., new subagent)
- ❌ NOT for `capabilities/` content updates

### Question: "When should I use release.sh vs bump-plugin-version.sh?"

**Use bump-plugin-version.sh when**:
- Only `.claude/` files changed
- No CLI/MCP code changes
- Quick plugin iteration

**Use release.sh when**:
- CLI or MCP code changed
- Need cross-platform binaries
- Creating official versioned release

## Git Hooks (Optional Automation)

For automatic version bumping, see [Git Hooks](git-hooks.md).

**Quick start**:
```bash
./scripts/install-hooks.sh
# Now .claude/ changes auto-bump plugin version on commit
```

## See Also

- [Git Hooks](git-hooks.md) - Automatic version bumping on commit
- [CHANGELOG.md](../../CHANGELOG.md) - Version history
- [GitHub Releases](https://github.com/yaleh/meta-cc/releases) - Published releases
- [GitHub Actions](https://github.com/yaleh/meta-cc/actions) - Build status

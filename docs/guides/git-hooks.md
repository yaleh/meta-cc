# Git Hooks

This document describes the git hooks available for meta-cc development.

## Overview

meta-cc provides git hooks to automate plugin version management:

- **pre-commit hook**: Auto-bumps plugin version when `.claude/` files change

## Installation

### Install Hooks

```bash
./scripts/install-hooks.sh
```

This installs all hooks from `.githooks/` to `.git/hooks/`.

### Uninstall Hooks

```bash
./scripts/uninstall-hooks.sh
```

This removes all active hooks from `.git/hooks/`.

## Pre-Commit Hook

### What It Does

Automatically detects changes to plugin structure files and bumps the plugin version:

**Triggers on**:
- `.claude/commands/*.md` (slash commands)
- `.claude/agents/*.md` (subagents)

**Does NOT trigger on**:
- `capabilities/commands/*.md` (capability content)
- CLI/MCP code changes (`cmd/`, `internal/`, `pkg/`)

**Actions**:
1. Detects staged `.claude/` file changes
2. Auto-increments patch version (e.g., 0.26.9 → 0.26.10)
3. Updates `plugin.json` and `marketplace.json`
4. Stages version files
5. Includes them in the same commit

### Example Workflow

```bash
# Edit a slash command
vim .claude/commands/meta.md

# Stage changes
git add .claude/commands/meta.md

# Commit (hook auto-bumps version)
git commit -m "feat: improve semantic matching"
```

**Output**:
```
Detected plugin file change: .claude/commands/meta.md

=== Auto Plugin Version Bump ===

Current plugin version: 0.26.9
New plugin version: 0.26.10 (auto-bumped patch)

✓ Plugin version auto-bumped to 0.26.10
✓ Version files staged

Note: Version was auto-incremented (patch).
If you need minor/major bump, please:
  1. Cancel this commit (Ctrl+C)
  2. Run: ./scripts/bump-plugin-version.sh [minor|major]

[develop abc1234] feat: improve semantic matching
 3 files changed, 3 insertions(+), 2 deletions(-)
```

**Result**: Commit includes both `.claude/commands/meta.md` change AND version bump.

### Manual Override

If you need a **minor** or **major** version bump instead of patch:

1. **Don't stage** the `.claude/` files yet
2. Run manual bump script first:
   ```bash
   ./scripts/bump-plugin-version.sh minor  # or major
   ```
3. Then commit both changes together:
   ```bash
   git add .claude/commands/meta.md
   git commit -m "feat: major /meta command rewrite"
   ```

The hook will detect that version files are already staged and skip auto-bump.

### Skipping Auto-Bump

If you want to commit `.claude/` changes without version bump (rare):

```bash
# Temporarily disable hook
git commit --no-verify -m "wip: experimental changes"
```

## When Hooks DON'T Run

Hooks are **skipped** in these scenarios:

1. **jq not installed**: Hook exits gracefully with warning
2. **Version files already staged**: User manually bumped version
3. **No `.claude/` files changed**: Only capabilities or code changes
4. **Using --no-verify flag**: Explicitly bypassing hooks

## Comparison: Hook vs Manual Script

| Aspect | Git Hook | Manual Script |
|--------|----------|---------------|
| **Trigger** | Automatic (on commit) | Manual invocation |
| **Version Type** | Patch only | Patch/Minor/Major |
| **Commits** | 1 commit (files + version) | 2 commits (files, then version) |
| **Best For** | Most commits | Major/minor bumps |
| **Override** | Use `--no-verify` | N/A |

## Recommended Workflow

### For Most Changes (Patch)

Use git hook (automatic):

```bash
vim .claude/commands/meta.md
git add .claude/commands/meta.md
git commit -m "fix: typo in /meta description"
# Hook auto-bumps patch version
```

### For Minor Changes

Use manual script:

```bash
vim .claude/commands/meta.md
./scripts/bump-plugin-version.sh minor
git add .claude/commands/meta.md
git commit -m "feat: add composite intent detection"
# Hook skips (version already staged)
```

### For Major Changes

Use manual script:

```bash
vim .claude/commands/meta.md
./scripts/bump-plugin-version.sh major
git add .claude/commands/meta.md
git commit -m "feat!: rewrite /meta command with breaking changes"
# Hook skips (version already staged)
```

## Hook Implementation Details

### File Location

- **Source**: `.githooks/pre-commit` (tracked in git)
- **Active**: `.git/hooks/pre-commit` (not tracked, installed locally)

### Why .githooks/?

Git hooks in `.git/hooks/` are **not tracked** by git (by design). We use `.githooks/` directory to:

1. ✅ Track hook source code in the repository
2. ✅ Share hooks with all developers
3. ✅ Version control hook changes
4. ✅ Allow easy installation via script

### Detection Logic

```bash
# Check if staged files match patterns
if [[ "$file" =~ ^\.claude/commands/.*\.md$ ]] ||
   [[ "$file" =~ ^\.claude/agents/.*\.md$ ]]; then
    PLUGIN_FILES_CHANGED=true
fi
```

### Version Bump Logic

```bash
# Parse current version: 0.26.9
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Auto-increment patch
PATCH=$((PATCH + 1))
NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"  # 0.26.10
```

## Troubleshooting

### Hook Not Running

**Problem**: Committed `.claude/` file but version didn't bump.

**Solutions**:
1. Check if hook is installed:
   ```bash
   ls -l .git/hooks/pre-commit
   ```
2. Reinstall hooks:
   ```bash
   ./scripts/install-hooks.sh
   ```
3. Check jq is installed:
   ```bash
   which jq
   ```

### Wrong Version Increment

**Problem**: Needed minor bump, got patch bump.

**Solution**: Use manual script before committing:
```bash
git reset HEAD~1  # Undo last commit
./scripts/bump-plugin-version.sh minor
git add .claude/commands/meta.md
git commit -m "feat: your message"
```

### Hook Interfering

**Problem**: Want to commit without version bump.

**Solution**: Use `--no-verify`:
```bash
git commit --no-verify -m "wip: testing"
```

## See Also

- [Release Process](release-process.md) - Full release workflow
- [Bump Plugin Version Script](../scripts/bump-plugin-version.sh) - Manual version bumping
- [Install Hooks Script](../scripts/install-hooks.sh) - Hook installation

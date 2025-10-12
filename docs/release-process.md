# Release Process

This document describes the automated release process for meta-cc.

## Overview

The release process is now **fully automated** to prevent version inconsistencies. The `scripts/release.sh` script handles all version updates, commits, tagging, and pushing.

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

## Release Workflow

### Step 1: Prepare Release

1. **Switch to the correct branch**:
   ```bash
   git checkout main      # For stable releases
   # or
   git checkout develop   # For beta/RC releases
   ```

2. **Ensure working directory is clean**:
   ```bash
   git status
   # Should show: "nothing to commit, working tree clean"
   ```

3. **Pull latest changes**:
   ```bash
   git pull origin main   # or develop
   ```

### Step 2: Run Release Script

**Single command to create release**:

```bash
./scripts/release.sh v1.0.0
```

**What the script does automatically**:

1. ✅ Validates version format (`vX.Y.Z` or `vX.Y.Z-beta`)
2. ✅ Checks branch (must be `main` or `develop`)
3. ✅ Checks working directory is clean
4. ✅ Runs full test suite (`make all`)
5. ✅ Updates `plugin.json` version
6. ✅ Updates `marketplace.json` version
7. ✅ Prompts you to update `CHANGELOG.md`
8. ✅ Commits version updates with attribution
9. ✅ Creates git tag
10. ✅ Pushes commit and tag to remote

### Step 3: Monitor GitHub Actions

After pushing, GitHub Actions automatically:

1. Verifies version consistency (plugin.json, marketplace.json, git tag)
2. Builds cross-platform binaries
3. Creates plugin packages
4. Publishes GitHub Release

Monitor progress at: https://github.com/yaleh/meta-cc/actions

## Version Update Example

### Before running release.sh

```
Current state:
- plugin.json: 0.26.8
- marketplace.json: 0.26.8
- Latest tag: v0.26.8
- Working directory: clean
```

### Run release script

```bash
./scripts/release.sh v0.27.0
```

### Script execution flow

```
1. Validates: v0.27.0 ✓
2. Checks branch: main ✓
3. Checks clean: ✓
4. Runs: make all ✓
5. Updates: plugin.json → 0.27.0 ✓
6. Updates: marketplace.json → 0.27.0 ✓
7. Prompts: Update CHANGELOG.md...
   [You edit CHANGELOG.md in another terminal]
   [Press Enter to continue]
8. Verifies: CHANGELOG.md contains [0.27.0] ✓
9. Commits: "chore: release v0.27.0" ✓
10. Tags: v0.27.0 ✓
11. Pushes: main + v0.27.0 ✓
```

### After release.sh completes

```
New state:
- plugin.json: 0.27.0 ✓
- marketplace.json: 0.27.0 ✓
- Latest tag: v0.27.0 ✓
- Commit: "chore: release v0.27.0"
- GitHub Actions: Building... 🚀
```

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

## See Also

- [CHANGELOG.md](../CHANGELOG.md) - Version history
- [GitHub Releases](https://github.com/yaleh/meta-cc/releases) - Published releases
- [GitHub Actions](https://github.com/yaleh/meta-cc/actions) - Build status

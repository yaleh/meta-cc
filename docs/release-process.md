# Release Process

This document describes the automated release process for meta-cc.

## Overview

The release process uses **dynamic version injection** from git tags. Version numbers are stored only in git tags (single source of truth), and are automatically injected into release artifacts during the build process.

**Key principle**: `plugin.json` and `marketplace.json` always show `"version": "dev"` in the repository. The actual version is injected by GitHub Actions from the git tag.

## Prerequisites

1. **Required tools**:
   - `git` - Version control
   - `make` - Build automation (only if using release.sh)

2. **No version file editing required**:
   - ✅ Version numbers are automatically injected from git tags
   - ✅ No need to manually update plugin.json or marketplace.json
   - ✅ No need to install jq for local releases

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

### Step 2: Create Release

**Option A: Direct git tag (Simplest)**:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

**Option B: Using release.sh (With validation)**:

```bash
./scripts/release.sh v1.0.0
```

**What the script provides (optional)**:

1. ✅ Validates version format (`vX.Y.Z` or `vX.Y.Z-beta`)
2. ✅ Checks branch (must be `main` or `develop`)
3. ✅ Checks working directory is clean
4. ✅ Runs full test suite (`make all`)
5. ✅ Validates `CHANGELOG.md` updated
6. ✅ Creates git tag
7. ✅ Pushes tag to remote

**Note**: The script no longer updates version files. Versions are injected dynamically by GitHub Actions.

### Step 3: Monitor GitHub Actions

After pushing, GitHub Actions automatically:

1. **Injects version from git tag** into plugin.json and marketplace.json
2. Syncs plugin files
3. Builds cross-platform binaries
4. Creates plugin packages
5. Publishes GitHub Release

Monitor progress at: https://github.com/yaleh/meta-cc/actions

## Version Update Example

### Before creating release

```
Repository state:
- plugin.json: "dev" (always)
- marketplace.json: "dev" (always)
- Latest tag: v0.26.8
- Working directory: clean
```

### Option A: Direct tag (fastest)

```bash
git tag -a v0.27.0 -m "Release v0.27.0"
git push origin v0.27.0
```

### Option B: Using release.sh (with validation)

```bash
./scripts/release.sh v0.27.0
```

Script execution flow:
```
1. Validates: v0.27.0 ✓
2. Checks branch: main ✓
3. Checks clean: ✓
4. Runs: make all ✓
5. Prompts: Ensure CHANGELOG.md updated
   [Press Enter to continue]
6. Verifies: CHANGELOG.md contains [0.27.0] ✓
7. Tags: v0.27.0 ✓
8. Pushes: v0.27.0 ✓
```

### After tag pushed

```
Repository state:
- plugin.json: "dev" (unchanged)
- marketplace.json: "dev" (unchanged)
- Latest tag: v0.27.0 ✓
- GitHub Actions: Building... 🚀

GitHub Actions will:
- Inject version 0.27.0 into plugin.json
- Inject version 0.27.0 into marketplace.json
- Build release packages with correct version
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

## Key Differences from Traditional Versioning

### Traditional Approach (version in files)
```
plugin.json: "version": "0.27.0"
↓
git commit
↓
git tag v0.27.0
↓
Risk: Files can get out of sync with tags
```

### Dynamic Injection (this project)
```
plugin.json: "version": "dev"
↓
git tag v0.27.0
↓
GitHub Actions injects "0.27.0" during build
↓
Benefit: Single source of truth (git tag)
```

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

## See Also

- [CHANGELOG.md](../CHANGELOG.md) - Version history
- [GitHub Releases](https://github.com/yaleh/meta-cc/releases) - Published releases
- [GitHub Actions](https://github.com/yaleh/meta-cc/actions) - Build status

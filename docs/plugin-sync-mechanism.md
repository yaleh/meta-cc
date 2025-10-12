# Plugin Sync Mechanism

This document describes the automated synchronization mechanisms that keep `plugin.json` consistent with the project files.

## Overview

The meta-cc plugin uses a **three-layer protection mechanism** to ensure plugin.json stays synchronized:

1. **Pre-commit Hook** (Development) - Auto-updates before commits
2. **CI Validation** (Pull Requests) - Blocks PRs with out-of-sync manifests
3. **Release Workflow** (Deployment) - Final sync from git tags

---

## Layer 1: Pre-commit Hook (Automatic)

### What it does

Automatically updates `plugin.json` **commands** and **agents** arrays with actual files from `.claude/` directory before every commit.

### Location

`.claude/hooks/pre-commit.sh`

### How it works

```bash
# 1. Scans .claude/commands/ for *.md files
# 2. Scans .claude/agents/ for *.md files
# 3. Updates plugin.json with file paths
# 4. Auto-stages changes if modified
```

### Manual trigger

```bash
bash scripts/update-plugin-manifest.sh
```

### What it syncs

✓ `commands` array - List of command files
✓ `agents` array - List of agent files
✗ `version` field - Updated by release workflow only

---

## Layer 2: CI Validation (Pull Requests)

### What it does

Validates that `plugin.json` matches actual files during CI runs. **Fails the build** if out of sync.

### Location

`.github/workflows/ci.yml` → "Verify plugin manifest is up-to-date" step

### How it works

```yaml
- name: Verify plugin manifest is up-to-date
  run: |
    bash scripts/update-plugin-manifest.sh

    # Fail if changes detected
    if ! git diff --quiet .claude-plugin/plugin.json; then
      echo "ERROR: plugin.json is out of sync!"
      exit 1
    fi
```

### When it runs

- Every push to `main` or `develop`
- Every pull request

### Error message

```
ERROR: plugin.json is out of sync!
The following changes are needed:
<shows diff>

Please run: bash scripts/update-plugin-manifest.sh
And commit the changes.
```

---

## Layer 3: Release Workflow (Deployment)

### What it does

During release, automatically updates `plugin.json` **version** field from git tag.

### Location

`.github/workflows/release.yml` → "Update plugin.json version" step

### How it works

```yaml
- name: Update plugin.json version
  run: |
    VERSION=${GITHUB_REF#refs/tags/}  # Extract version from tag
    jq --arg ver "${VERSION#v}" '.version = $ver' .claude-plugin/plugin.json
```

### When it runs

- Every time a version tag is pushed (e.g., `v0.26.5`)

### What it syncs

✓ `version` field - Matches git tag
✓ `marketplace.json` version - Also updated

---

## File List Synchronization

### Current files tracked

**Commands** (1 file):
```json
"commands": [
  "./.claude/commands/meta.md"
]
```

**Agents** (2 files):
```json
"agents": [
  "./.claude/agents/project-planner.md",
  "./.claude/agents/stage-executor.md"
]
```

### Adding new files

When you add a new command or agent file:

1. **Automatic** (Pre-commit hook):
   - Create file in `.claude/commands/` or `.claude/agents/`
   - Commit → Hook automatically updates `plugin.json`

2. **Manual** (If hook disabled):
   ```bash
   bash scripts/update-plugin-manifest.sh
   git add .claude-plugin/plugin.json
   git commit -m "Add new command/agent"
   ```

3. **Validation** (CI):
   - CI will fail if you forget to update
   - Run the script and commit changes

---

## Version Synchronization

### Version sources

The project has **three version locations** that must match:

1. **Git tag**: `v0.26.5` (source of truth)
2. **plugin.json**: `"version": "0.26.5"` (updated by release workflow)
3. **marketplace.json**: `"plugins[0].version": "0.26.5"` (updated by release workflow)

### Synchronization flow

```
Developer → Create git tag (v0.26.5)
           ↓
GitHub Actions → Extract version from tag
           ↓
Release Workflow → Update plugin.json version
           ↓
           → Update marketplace.json version
           ↓
Release artifacts → Built with correct version
```

### Manual version update (NOT recommended)

If you need to update version manually:

```bash
# Update plugin.json
jq '.version = "0.26.6"' .claude-plugin/plugin.json > tmp && mv tmp .claude-plugin/plugin.json

# Update marketplace.json
jq '.plugins[0].version = "0.26.6"' .claude-plugin/marketplace.json > tmp && mv tmp .claude-plugin/marketplace.json

# Commit and create matching tag
git add .claude-plugin/*.json
git commit -m "chore: bump version to 0.26.6"
git tag v0.26.6
git push origin main v0.26.6
```

**Best practice**: Let the release workflow handle version updates automatically.

---

## Troubleshooting

### Pre-commit hook not running

**Symptom**: `plugin.json` not updated before commits

**Solution**:
```bash
# Verify hook exists and is executable
ls -l .claude/hooks/pre-commit.sh
chmod +x .claude/hooks/pre-commit.sh

# Claude Code should load hooks automatically
# If not, manually run:
bash .claude/hooks/pre-commit.sh
```

### CI fails with "plugin.json out of sync"

**Symptom**: CI build fails with manifest sync error

**Solution**:
```bash
# Run update script locally
bash scripts/update-plugin-manifest.sh

# Check changes
git diff .claude-plugin/plugin.json

# Commit changes
git add .claude-plugin/plugin.json
git commit -m "fix: sync plugin.json with actual files"
git push
```

### Version mismatch after release

**Symptom**: plugin.json version doesn't match git tag

**Cause**: Release workflow failed or was skipped

**Solution**:
```bash
# Check release workflow logs
gh run list --workflow=release.yml --limit 5

# If workflow failed, re-tag to trigger:
git tag -d v0.26.5          # Delete local tag
git push origin :v0.26.5    # Delete remote tag
git tag v0.26.5             # Recreate tag
git push origin v0.26.5     # Push tag (triggers workflow)
```

---

## Testing the Sync Mechanism

### Test pre-commit hook

```bash
# 1. Create a test file
touch .claude/commands/test-command.md

# 2. Try to commit
git add .claude/commands/test-command.md
git commit -m "test: add test command"

# 3. Verify plugin.json was updated
git diff HEAD~1 .claude-plugin/plugin.json
# Should show test-command.md added to commands array

# 4. Cleanup
git reset HEAD~1
rm .claude/commands/test-command.md
```

### Test CI validation

```bash
# 1. Manually edit plugin.json (remove a file)
jq '.commands = []' .claude-plugin/plugin.json > tmp && mv tmp .claude-plugin/plugin.json

# 2. Commit and push
git add .claude-plugin/plugin.json
git commit -m "test: break plugin.json sync"
git push origin test-branch

# 3. Check CI - should FAIL with sync error

# 4. Cleanup
git reset HEAD~1
git push -f origin test-branch
```

### Test release workflow

```bash
# 1. Create a test tag
git tag v0.26.6-test

# 2. Push tag
git push origin v0.26.6-test

# 3. Watch release workflow
gh run watch

# 4. Verify version in release artifacts
gh release view v0.26.6-test

# 5. Cleanup
gh release delete v0.26.6-test
git tag -d v0.26.6-test
git push origin :v0.26.6-test
```

---

## Summary

| Layer | What | When | Syncs |
|-------|------|------|-------|
| Pre-commit Hook | Auto-update manifest | Before commits | commands, agents |
| CI Validation | Verify sync | PR/push to main | commands, agents |
| Release Workflow | Update version | Tag push | version field |

**Key takeaway**: You don't need to manually edit `plugin.json` - the automation handles it!

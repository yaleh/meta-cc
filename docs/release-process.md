# Release Process

This document describes the process for creating and publishing a new release of meta-cc.

## Prerequisites

- **Maintainer access**: Push access to the main branch
- **Clean working directory**: All changes committed or stashed
- **Tests passing**: Run `make all` successfully
- **CHANGELOG updated**: Release notes prepared

## Release Workflow

### 1. Prepare Release

Ensure you're on the main or develop branch and up to date:

```bash
# Checkout main branch
git checkout main
git pull origin main

# Verify tests pass
make all
```

### 2. Update CHANGELOG.md

Edit `CHANGELOG.md` to move items from `[Unreleased]` to the new version:

```markdown
## [v1.0.0] - 2025-10-08

### Added
- Feature X implementation
- New command Y

### Changed
- Updated behavior of Z

### Fixed
- Bug fix for issue #123
```

Commit the CHANGELOG update:

```bash
git add CHANGELOG.md
git commit -m "docs: update CHANGELOG for v1.0.0"
git push origin main
```

### 3. Execute Release Script

Run the automated release script:

```bash
./scripts/release.sh v1.0.0
```

The script will:
1. Validate version format (e.g., `v1.0.0` or `v1.0.0-beta.1`)
2. Check you're on main or develop branch
3. Verify working directory is clean
4. Run full test suite (`make all`)
5. Prompt you to verify CHANGELOG was updated
6. Create annotated git tag
7. Push tag to GitHub

### 4. Monitor GitHub Actions

Once the tag is pushed, GitHub Actions will automatically:

1. **Build binaries** for 5 platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

2. **Create platform bundles** including:
   - Binaries (`meta-cc`, `meta-cc-mcp`)
   - Slash commands (`.claude/commands/`)
   - Subagents (`.claude/agents/`)
   - Installation script

3. **Create GitHub Release** with auto-generated release notes

4. **Upload artifacts** (16 total files per release)

Monitor the build progress at:
https://github.com/yaleh/meta-cc/actions

### 5. Verify Release

After GitHub Actions completes:

1. **Check the release page**:
   https://github.com/yaleh/meta-cc/releases

2. **Verify artifacts are attached**:

   **Individual binaries** (10 files):
   - `meta-cc-linux-amd64`, `meta-cc-linux-arm64`
   - `meta-cc-darwin-amd64`, `meta-cc-darwin-arm64`
   - `meta-cc-windows-amd64.exe`
   - `meta-cc-mcp-linux-amd64`, `meta-cc-mcp-linux-arm64`
   - `meta-cc-mcp-darwin-amd64`, `meta-cc-mcp-darwin-arm64`
   - `meta-cc-mcp-windows-amd64.exe`

   **Platform bundles** (5 files):
   - `meta-cc-bundle-linux-amd64.tar.gz`
   - `meta-cc-bundle-linux-arm64.tar.gz`
   - `meta-cc-bundle-darwin-amd64.tar.gz`
   - `meta-cc-bundle-darwin-arm64.tar.gz`
   - `meta-cc-bundle-windows-amd64.tar.gz`

   **Checksums** (1 file):
   - `checksums.txt`

3. **Test bundle installation**:
   ```bash
   # Download and extract bundle
   curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-bundle-linux-amd64.tar.gz | tar xz
   cd meta-cc-v*/
   ./install.sh

   # Verify installation
   meta-cc --version
   meta-cc-mcp --version
   ls ~/.claude/projects/meta-cc/commands/
   ls ~/.claude/projects/meta-cc/agents/
   ```

4. **Test individual binary download**:
   ```bash
   curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-linux-amd64 -o meta-cc
   chmod +x meta-cc
   ./meta-cc --version
   ```

## Versioning Strategy

meta-cc follows [Semantic Versioning](https://semver.org/):

- **v0.x.x**: Pre-1.0 beta releases
- **v1.0.0**: First stable release
- **v1.x.0**: Minor version (new features, backward compatible)
- **v1.0.x**: Patch version (bug fixes only)
- **v1.0.0-beta.1**: Pre-release versions

## Troubleshooting

### Build Fails in GitHub Actions

1. Check the Actions tab for error logs
2. Common issues:
   - Syntax error in workflow files
   - Missing dependencies
   - Test failures on specific platforms

3. Fix the issue and create a new tag:
   ```bash
   # Delete the failed tag locally and remotely
   git tag -d v1.0.0
   git push --delete origin v1.0.0
   
   # Fix the issue, commit, and re-run release
   ./scripts/release.sh v1.0.0
   ```

### Tag Already Exists

If you need to recreate a tag:

```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push --delete origin v1.0.0

# Recreate tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Binary Missing from Release

If a binary is missing:

1. Check `.github/workflows/release.yml` for typos
2. Ensure all platforms are listed in the build matrix
3. Re-run the workflow or create a new tag

### Permission Denied when Pushing Tag

Ensure you have write access to the repository:

```bash
# Check remote URL
git remote -v

# If using HTTPS, ensure credentials are configured
git config --global credential.helper store
```

## Post-Release Tasks

After a successful release:

1. **Announce the release**:
   - Update project README if needed
   - Post to relevant communities

2. **Create new Unreleased section** in CHANGELOG.md:
   ```markdown
   ## [Unreleased]

   ### Added
   - New features will be listed here
   ```

3. **Monitor for issues**:
   - Watch GitHub issues for bug reports
   - Prepare hotfix release if critical bugs are found

## Emergency Hotfix Process

For critical bugs requiring immediate release:

1. Create hotfix branch from main:
   ```bash
   git checkout -b hotfix/v1.0.1 main
   ```

2. Fix the bug and test thoroughly:
   ```bash
   # Make changes
   make all
   ```

3. Update CHANGELOG.md:
   ```markdown
   ## [v1.0.1] - 2025-10-08

   ### Fixed
   - Critical bug description
   ```

4. Commit and merge to main:
   ```bash
   git add .
   git commit -m "fix: critical bug description"
   git checkout main
   git merge hotfix/v1.0.1
   git push origin main
   ```

5. Run release script:
   ```bash
   ./scripts/release.sh v1.0.1
   ```

---

For questions or issues with the release process, open an issue or contact the maintainers.

# ADR-002: Plugin Directory Structure Refactoring

## Status

Accepted

## Context

The meta-cc project needs to be distributed as a Claude Code plugin. Claude Code has specific requirements for plugin structure:

### Plugin Directory Requirements

Claude Code expects plugins to have:
- `commands/` directory at the root (for slash commands)
- `agents/` directory at the root (for subagents)
- `plugin.json` metadata file

### Development Workflow Requirements

During development, we need:
- Real-time testing in Claude Code without rebuilding
- Source files tracked in Git
- Build artifacts excluded from Git
- Cross-platform compatibility (Windows, Linux, macOS)

### Initial Problem

The original structure had:
- `.claude/commands/` and `.claude/agents/` as **source files**
- `commands/` and `agents/` as **symlinks** to `.claude/` directories

**Issues with symlinks**:
1. Windows doesn't handle symlinks well without admin privileges
2. Git tracks symlinks differently across platforms
3. CI/CD systems have inconsistent symlink support
4. Users cloning the repo had broken symlinks

### Conflicting Requirements

1. **Development**: Need files in `.claude/` for real-time testing
2. **Release**: Need files in root `commands/` and `agents/` for plugin spec
3. **Git**: Should only track source files, not build artifacts
4. **Cross-platform**: Must work on Windows, Linux, macOS without symlinks

## Decision

We adopt a **source/build artifact separation** with explicit sync:

### Directory Structure

```
meta-cc/
├── .claude/              # Development directory (SOURCE)
│   ├── commands/        # Source: Slash command files (Git tracked)
│   ├── agents/          # Source: Subagent files (Git tracked)
│   └── hooks/           # Source: Project hooks (Git tracked)
├── commands/            # Build artifact (Git ignored, synced on release)
├── agents/              # Build artifact (Git ignored, synced on release)
├── .claude-plugin/      # Plugin metadata (Git tracked)
│   ├── plugin.json
│   └── marketplace.json
└── .gitignore           # Ignore commands/ and agents/
```

### Workflow

**Local Development**:
1. Edit source files in `.claude/commands/` and `.claude/agents/`
2. Test immediately - Claude Code reads from `.claude/` directory
3. No build step needed for testing
4. Run `make test` or `make all` for validation

**Before Committing**:
1. Verify changes in `.claude/commands/` and `.claude/agents/`
2. Run `make all` for linting, tests, and build
3. Do NOT manually create `commands/` or `agents/` directories

**Release Process**:
1. Run `make sync-plugin-files` (or automatic in `make bundle-release`)
2. This copies `.claude/commands/` → `commands/`
3. This copies `.claude/agents/` → `agents/`
4. Create release bundle: `make bundle-release VERSION=vX.Y.Z`

**CI/CD**:
- GitHub Actions automatically runs sync during release workflow
- No symlink dependencies
- Works identically on all platforms

### Makefile Targets

```makefile
# Sync plugin files from .claude/ to root directories
sync-plugin-files:
	@mkdir -p commands agents
	@rsync -av .claude/commands/ commands/
	@rsync -av .claude/agents/ agents/

# Create release bundle (includes sync)
bundle-release: sync-plugin-files build
	@tar -czf meta-cc-$(VERSION).tar.gz \
		commands/ agents/ .claude-plugin/ bin/
```

## Consequences

### Positive Impacts

1. **Real-Time Development**
   - Edit files in `.claude/`, test immediately
   - No build step for local testing
   - Fast iteration cycle

2. **Cross-Platform Compatibility**
   - No symlinks required
   - Works identically on Windows, Linux, macOS
   - CI/CD systems have no special requirements

3. **Clear Separation**
   - Source files clearly marked (`.claude/`)
   - Build artifacts clearly marked (`commands/`, `agents/`)
   - Git only tracks source files

4. **Release Automation**
   - Single command creates release bundle
   - Consistent release artifacts
   - No manual file copying

5. **Plugin Spec Compliance**
   - Release bundle has `commands/` and `agents/` at root
   - Meets Claude Code plugin requirements
   - Easy installation for users

### Negative Impacts

1. **Manual Sync Required**
   - Must remember to run `make sync-plugin-files` before release
   - Risk of forgetting to sync (mitigated by CI/CD automation)

2. **Duplicate Files on Disk**
   - Source files in `.claude/`
   - Build artifacts in `commands/` and `agents/`
   - Minimal impact (text files are small)

3. **Potential for Drift**
   - If sync is not run, build artifacts become stale
   - Mitigated by: CI/CD automation, `make bundle-release` includes sync

### Risks

1. **Forgetting to Sync**
   - Risk: Release with outdated `commands/` and `agents/`
   - Mitigation: `make bundle-release` automatically syncs, CI/CD validates

2. **Editing Wrong Directory**
   - Risk: Developer edits `commands/` instead of `.claude/commands/`
   - Mitigation: Documentation, Git ignore `commands/`, code review

3. **CI/CD Failures**
   - Risk: Sync fails on CI/CD due to missing tools
   - Mitigation: Use standard `rsync` (available on all platforms), test CI/CD

## Implementation

### Completed

- [x] Create `.claude/commands/` and `.claude/agents/` directories
- [x] Move source files to `.claude/` directories
- [x] Add `commands/` and `agents/` to `.gitignore`
- [x] Create `sync-plugin-files` Makefile target
- [x] Update `bundle-release` to include sync
- [x] Update GitHub Actions workflow
- [x] Update documentation (CLAUDE.md, README.md)

### File Changes

**.gitignore**:
```gitignore
# Build artifacts (synced from .claude/ on release)
/commands/
/agents/
```

**Makefile**:
```makefile
.PHONY: sync-plugin-files
sync-plugin-files:
	@echo "Syncing plugin files from .claude/ to root directories..."
	@mkdir -p commands agents
	@rsync -av --delete .claude/commands/ commands/ || \
		(echo "Error: rsync failed. Is rsync installed?" && exit 1)
	@rsync -av --delete .claude/agents/ agents/ || \
		(echo "Error: rsync failed. Is rsync installed?" && exit 1)
	@echo "Plugin files synced successfully"

.PHONY: bundle-release
bundle-release: sync-plugin-files build
	@echo "Creating release bundle..."
	@tar -czf meta-cc-$(VERSION).tar.gz \
		commands/ agents/ .claude-plugin/ bin/
	@echo "Release bundle created: meta-cc-$(VERSION).tar.gz"
```

## Related Decisions

- [ADR-001](ADR-001-two-layer-architecture.md) - Two-Layer Architecture Design
- [ADR-003](ADR-003-mcp-server-integration.md) - MCP Server Integration Strategy

## Notes

### Design Rationale

The key insight is that **development workflow** and **release structure** have different requirements:

- **Development**: Files in `.claude/` for real-time testing
- **Release**: Files in `commands/` and `agents/` for plugin spec

By treating `commands/` and `agents/` as **build artifacts**, we get:
- Clean Git history (only source files tracked)
- Cross-platform compatibility (no symlinks)
- Plugin spec compliance (release bundle has correct structure)

### Alternative Approaches Considered

1. **Symlinks** (rejected)
   - Pros: No sync needed, single source of truth
   - Cons: Windows compatibility, Git inconsistencies, CI/CD issues

2. **Only root directories** (rejected)
   - Pros: Simple structure, no sync needed
   - Cons: Can't test in Claude Code during development (requires `.claude/`)

3. **Git pre-commit hook** (rejected)
   - Pros: Automatic sync on commit
   - Cons: Intrusive, hard to debug, inconsistent across developers

4. **Hardlinks** (rejected)
   - Pros: Single file on disk, automatic sync
   - Cons: Windows compatibility, risk of accidental edits

### References

- [Claude Code Plugin Spec](https://docs.claude.com/en/docs/claude-code/overview)
- [Plugin Development Workflow](../examples-usage.md#plugin-development)

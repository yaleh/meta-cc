# Plugin Verification Checklist

## Pre-Release Testing

Complete this checklist before creating a public release.

### Platform Testing

Test installation on all supported platforms:

- [ ] **Linux x86_64** (amd64)
  - [ ] Installation succeeds
  - [ ] Binaries execute
  - [ ] MCP server connects
  - [ ] Slash commands work
  - [ ] Subagents work

- [ ] **Linux ARM64**
  - [ ] Installation succeeds
  - [ ] Binaries execute
  - [ ] MCP server connects
  - [ ] Slash commands work
  - [ ] Subagents work

- [ ] **macOS Intel** (x86_64)
  - [ ] Installation succeeds
  - [ ] Binaries execute (Gatekeeper handled)
  - [ ] MCP server connects
  - [ ] Slash commands work
  - [ ] Subagents work

- [ ] **macOS Apple Silicon** (ARM64)
  - [ ] Installation succeeds
  - [ ] Binaries execute (Gatekeeper handled)
  - [ ] MCP server connects
  - [ ] Slash commands work
  - [ ] Subagents work

- [ ] **Windows x86_64** (via Git Bash)
  - [ ] Installation succeeds
  - [ ] Binaries execute
  - [ ] MCP server connects
  - [ ] Slash commands work
  - [ ] Subagents work

### Component Testing

Verify all components function correctly:

#### Binaries
- [ ] `meta-cc-mcp --version` shows correct version
- [ ] Binary is in PATH after installation
- [ ] Binary has correct permissions (executable)

#### Slash Commands
- [ ] `/meta-stats` displays session statistics
- [ ] `/meta-errors` analyzes error patterns
- [ ] `/meta-timeline` shows project timeline
- [ ] `/meta-habits` analyzes work patterns
- [ ] `/meta-focus-analyzer` shows focus distribution
- [ ] `/meta-quality-scan` provides quality assessment
- [ ] `/meta-viz` creates ASCII dashboards
- [ ] `/meta-coach` gives workflow recommendations
- [ ] `/meta-next` generates next steps
- [ ] `/meta-guide` provides intelligent guidance
- [ ] `/meta-prompt [prompt]` refines prompts

#### Subagents
- [ ] `@meta-coach` responds to queries
- [ ] Multi-turn conversation works
- [ ] Tool calls execute successfully
- [ ] Provides actionable recommendations

#### MCP Server
- [ ] Server starts without errors
- [ ] Server accepts JSON-RPC requests
- [ ] All 14 query tools available
- [ ] `get_session_stats` returns data
- [ ] `query_tools` filters correctly
- [ ] `query_user_messages` searches messages
- [ ] Hybrid output mode works (inline + file_ref)
- [ ] Temporary files created correctly

### Installation Testing

Test different installation scenarios:

#### Fresh Installation
- [ ] No existing meta-cc installation
- [ ] Install script completes successfully
- [ ] All files copied to correct locations
- [ ] MCP config created at `~/.claude/mcp.json`
- [ ] Post-install verification passes

#### Upgrade Installation
- [ ] Existing meta-cc installation present
- [ ] Install script overwrites binaries
- [ ] Existing MCP servers preserved
- [ ] Slash commands updated
- [ ] Subagents updated
- [ ] No configuration loss

#### Custom Installation Path
- [ ] `INSTALL_DIR=/custom/path ./install.sh` works
- [ ] Binaries installed to custom path
- [ ] MCP config references custom path
- [ ] Post-install verification passes

#### MCP Config Merging
- [ ] Existing `~/.claude/mcp.json` with other servers
- [ ] Install script merges meta-cc entry
- [ ] Existing servers preserved
- [ ] No JSON syntax errors
- [ ] All MCP servers still work

#### Uninstallation
- [ ] Uninstall script completes successfully
- [ ] Binaries removed from installation directory
- [ ] Slash commands removed from `~/.claude/commands/`
- [ ] Subagents removed from `~/.claude/agents/`
- [ ] MCP config preserved (not deleted)
- [ ] Clean removal message displayed

#### Reinstallation
- [ ] Uninstall completes successfully
- [ ] Fresh install after uninstall works
- [ ] All components functional after reinstall

### Documentation Testing

Verify all documentation is accurate:

#### Installation Guide
- [ ] Quick install commands work for each platform
- [ ] Manual installation steps work
- [ ] Verification steps execute successfully
- [ ] Troubleshooting steps resolve issues
- [ ] Uninstallation steps work
- [ ] All code blocks execute without errors

#### README.md
- [ ] Installation section accurate
- [ ] Plugin installation commands work
- [ ] Verification commands work
- [ ] Links resolve correctly

#### CHANGELOG.md
- [ ] Phase 20 changes documented
- [ ] Version number matches plugin.json
- [ ] All major features listed

#### Cross-References
- [ ] All internal links resolve (*.md files)
- [ ] All external links accessible (GitHub URLs)
- [ ] No broken references

### Release Testing

Verify GitHub Release workflow:

#### Workflow Execution
- [ ] Workflow triggered by version tag (e.g., `v0.12.0`)
- [ ] All 5 platform builds succeed
- [ ] Plugin packages created for all platforms
- [ ] Checksums generated correctly
- [ ] Release notes auto-generated
- [ ] Artifacts uploaded to GitHub Release

#### Package Contents
- [x] Each package contains `bin/meta-cc-mcp` (or `.exe`)
- [ ] Each package contains `.claude/commands/` files
- [ ] Each package contains `.claude/agents/` files
- [ ] Each package contains `.claude/lib/` files
- [ ] Each package contains `install.sh`
- [ ] Each package contains `uninstall.sh`
- [ ] Each package contains `plugin.json`
- [ ] Each package contains `README.md`
- [ ] Each package contains `LICENSE`

**Note**: `bin/meta-cc` (CLI binary) was removed in Phase 26 - MCP-only architecture. Only `bin/meta-cc-mcp` is required.

#### Checksums
- [ ] `checksums.txt` file created
- [ ] SHA256 hashes correct for all packages
- [ ] Checksums verify successfully:
  ```bash
  sha256sum -c checksums.txt
  ```

#### Download and Install
- [ ] Download plugin package from GitHub Release
- [ ] Extract archive successfully
- [ ] Run `./install.sh` from extracted directory
- [ ] Installation completes without errors
- [ ] All components functional after install

### Regression Testing

Ensure no breaking changes:

#### Existing Workflows
- [ ] Direct binary installation still works
- [ ] Manual MCP configuration still works
- [ ] Existing slash commands still work
- [ ] Existing subagents still work
- [ ] No changes to CLI interface
- [ ] No changes to MCP protocol

#### Compatibility
- [ ] Works with Claude Code 1.0+
- [ ] Works with existing projects
- [ ] Works with existing session data
- [ ] No data migration required

### Performance Testing

Verify acceptable performance:

#### Installation

- [ ] Installation completes in <2 minutes
- [ ] No unnecessary network calls
- [ ] Efficient file copying

#### MCP Server

- [ ] Server starts in <5 seconds
- [ ] Query responses in <10 seconds (small datasets)
- [ ] Hybrid mode activates correctly (>8KB results)
- [ ] No memory leaks during extended use

#### Slash Commands

- [ ] Commands respond in <30 seconds
- [ ] No timeout errors
- [ ] Results display correctly


## Testing Environments

### Primary Testing (Required)

- **Linux x86_64**: Ubuntu 22.04 or later
- **macOS ARM64**: macOS 12 (Monterey) or later

### Secondary Testing (Recommended)

- **Linux ARM64**: Ubuntu 22.04 on ARM64
- **macOS Intel**: macOS 12 or later
- **Windows x86_64**: Windows 10/11 with Git Bash

### CI Testing (Automated)

- All 5 platforms via GitHub Actions
- Workflow validation on every tag push

## Testing Tools

### Required

- `bash` (installation scripts)
- `tar` (package extraction)
- `jq` (JSON validation)
- `curl` or `wget` (downloading packages)

### Optional

- `yamllint` (workflow validation)
- `shellcheck` (script linting)
- `sha256sum` (checksum verification)


## Acceptance Criteria

All checklist items must be completed before public release.

### Critical (Must Pass)

- ✅ Installation works on Linux x86_64
- ✅ Installation works on macOS ARM64
- ✅ All binaries execute successfully
- ✅ MCP server connects without errors
- ✅ Slash commands functional
- ✅ Subagents functional
- ✅ Uninstall removes all components
- ✅ GitHub Release workflow succeeds
- ✅ All 5 platform packages created

### Important (Should Pass)

- ✅ Installation works on all 5 platforms
- ✅ MCP config merge preserves existing servers
- ✅ Documentation accurate and complete
- ✅ Troubleshooting steps resolve common issues
- ✅ Checksums validate correctly

### Nice-to-Have (May Pass)

- ⚠️ Installation completes in <1 minute
- ⚠️ Zero installation errors on first attempt
- ⚠️ Perfect cross-platform consistency


## Issue Tracking

### Critical Issues (Block Release)

- [ ] No critical issues identified

### High Priority Issues (Should Fix)

- [ ] No high priority issues identified

### Medium Priority Issues (Can Fix Later)

- [ ] No medium priority issues identified

### Low Priority Issues (Future Enhancement)

- [ ] No low priority issues identified


## Sign-off

**Testing Completed By**: _________________

**Date**: _________________

**Platforms Tested**: _________________

**Issues Found**: _________________

**Release Approved**: ☐ Yes ☐ No (with conditions) ☐ No

**Notes**:

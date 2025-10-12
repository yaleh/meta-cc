# Phase 20: Plugin Packaging & Release - TDD Implementation Plan

## Phase Overview

**Objective**: Package meta-cc as a distributable plugin with one-click installation support for Linux, macOS, and Windows.

**Code Volume**: ~400 lines | **Priority**: High | **Status**: Planning

**Dependencies**:
- Phase 0-19 (Complete meta-cc CLI + MCP Server + all features)
- Existing GitHub Release workflow (.github/workflows/release.yml)
- Existing install script (scripts/install.sh)

**Deliverables**:
- plugin.json manifest for Claude Code plugin standard
- Enhanced install.sh with MCP configuration support
- Multi-platform GitHub Release workflow
- Comprehensive installation documentation
- Plugin verification tests

---

## Phase Objectives

### Core Problems

**Problem 1: Manual Installation Complexity**
- Current: Users must manually copy slash commands, subagents, configure MCP
- Impact: High barrier to entry, error-prone setup
- Need: Automated one-click installation

**Problem 2: Cross-Platform Distribution**
- Current: Release workflow builds binaries but no unified plugin package
- Missing: Platform-specific installation automation
- Need: Standardized plugin structure with platform detection

**Problem 3: MCP Configuration Management**
- Current: Users manually edit ~/.claude/mcp.json
- Missing: Automated MCP server registration
- Need: Safe config merging without overwriting user settings

**Problem 4: Version Management**
- Current: No standardized version tracking
- Missing: Plugin metadata, upgrade path
- Need: plugin.json with version info, dependency tracking

### Solution Architecture

```
Phase 20 Implementation Strategy:

1. Plugin Structure (Stage 20.1)
   - Define plugin.json with metadata, dependencies
   - Organize .claude/ directory structure
   - Create installation manifest

2. Automated Installation (Stage 20.2)
   - Enhance install.sh with platform detection
   - Implement MCP configuration merging
   - Add verification and rollback support

3. GitHub Release Workflow (Stage 20.3)
   - Build multi-platform binaries (5 platforms)
   - Package plugin ZIP with all assets
   - Generate checksums and release notes

4. Documentation & Testing (Stage 20.4)
   - User installation guide
   - Platform-specific testing
   - Troubleshooting documentation
```

### Design Principles

1. **Zero Breaking Changes**: Existing installation methods continue to work
2. **Safe Configuration**: Never overwrite user's existing MCP settings
3. **Platform Detection**: Automatic detection of OS and architecture
4. **Rollback Support**: Ability to uninstall or revert to previous version
5. **Verification**: Post-install checks ensure correct setup

---

## Success Criteria

**Functional Acceptance**:
- ✅ All stage unit tests pass (TDD methodology)
- ✅ plugin.json validates against Claude Code plugin schema
- ✅ install.sh works on Linux (amd64, arm64), macOS (Intel, ARM), Windows
- ✅ MCP configuration merges correctly without data loss
- ✅ GitHub Release workflow succeeds for all platforms
- ✅ One-command installation works end-to-end

**Integration Acceptance**:
- ✅ Installed plugin recognized by Claude Code
- ✅ Slash commands functional post-install
- ✅ Subagents functional post-install
- ✅ MCP server connects successfully
- ✅ Uninstall removes all components cleanly

**Code Quality**:
- ✅ Total code: ~400 lines (within Phase 20 budget)
  - Stage 20.1: ~100 lines (plugin structure)
  - Stage 20.2: ~150 lines (installation)
  - Stage 20.3: ~100 lines (GitHub workflow)
  - Stage 20.4: ~50 lines (documentation)
- ✅ Each stage ≤ 200 lines
- ✅ Test coverage: ≥ 80% for testable components
- ✅ `make all` passes after each stage

---

## Stage 20.1: Plugin Structure Definition

### Objective

Define the plugin metadata structure and organize project files for distribution as a Claude Code plugin.

### Acceptance Criteria

- [ ] plugin.json created with complete metadata (name, version, author, dependencies)
- [ ] .claude/ directory structure documented and validated
- [ ] Directory layout follows Claude Code plugin conventions
- [ ] Manifest includes all required files (binaries, commands, agents)
- [ ] Version numbering follows SemVer 2.0
- [ ] JSON schema validation passes
- [ ] Unit tests for plugin metadata parsing

### TDD Approach

**Test File**: `tests/plugin_structure_test.sh` (~40 lines)

```bash
#!/bin/bash
# Test functions:
# - test_plugin_json_exists - Verify plugin.json exists
# - test_plugin_json_valid - Validate JSON syntax
# - test_plugin_version_semver - Check SemVer format
# - test_required_fields - Verify all required fields present
# - test_directory_structure - Validate .claude/ organization
```

**Test Strategy**:
1. Verify plugin.json exists and is valid JSON
2. Check all required fields (name, version, description, author)
3. Validate SemVer version format (e.g., "1.0.0")
4. Verify .claude/ directory structure (commands/, agents/, lib/)
5. Test manifest completeness (lists all required files)

**Implementation Files**:

1. `plugin.json` (~60 lines)

```json
{
  "name": "meta-cc",
  "version": "0.12.0",
  "description": "Meta-Cognition tool for Claude Code - analyze session history for workflow optimization",
  "author": "Yale Huang <yaleh@ieee.org>",
  "license": "MIT",
  "homepage": "https://github.com/yaleh/meta-cc",
  "repository": {
    "type": "git",
    "url": "https://github.com/yaleh/meta-cc"
  },
  "dependencies": {
    "claude-code": ">=1.0.0"
  },
  "platforms": [
    "linux-amd64",
    "linux-arm64",
    "darwin-amd64",
    "darwin-arm64",
    "windows-amd64"
  ],
  "binaries": [
    "bin/meta-cc",
    "bin/meta-cc-mcp"
  ],
  "integration": {
    "mcp": {
      "server": "meta-cc-mcp",
      "config_template": ".claude/lib/mcp-config.json"
    },
    "slash_commands": ".claude/commands/",
    "subagents": ".claude/agents/"
  },
  "install": {
    "script": "install.sh",
    "verify": "bin/meta-cc --version"
  },
  "uninstall": {
    "script": "uninstall.sh"
  }
}
```

2. `.claude/lib/mcp-config.json` (~30 lines)

```json
{
  "mcpServers": {
    "meta-cc": {
      "command": "meta-cc-mcp",
      "args": [],
      "env": {},
      "disabled": false,
      "alwaysAllow": []
    }
  }
}
```

3. `docs/plugin-structure.md` (~10 lines)

```markdown
# Plugin Structure

meta-cc follows the Claude Code plugin standard with the following structure:

```
meta-cc/
├── plugin.json              # Plugin metadata and manifest
├── install.sh               # Installation script
├── uninstall.sh             # Uninstallation script
├── bin/                     # Binaries (meta-cc, meta-cc-mcp)
├── .claude/
│   ├── commands/            # Slash commands
│   ├── agents/              # Subagent definitions
│   └── lib/
│       ├── mcp-config.json  # MCP configuration template
│       └── meta-utils.sh    # Shared utilities
└── docs/                    # Documentation
```
```

### File Changes

**New Files**:
- `plugin.json` (+60 lines)
- `.claude/lib/mcp-config.json` (+30 lines)
- `docs/plugin-structure.md` (+10 lines)
- `tests/plugin_structure_test.sh` (+40 lines)

**Total**: ~140 lines (exceeds 100-line target by 40 lines, acceptable for foundation stage)

### Test Commands

```bash
# Run Stage 20.1 tests
bash tests/plugin_structure_test.sh

# Validate plugin.json syntax
jq empty plugin.json

# Verify SemVer version
jq -r '.version' plugin.json | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$'

# Check required fields
jq -e '.name, .version, .description, .author' plugin.json

# Verify directory structure
test -d .claude/commands && test -d .claude/agents && test -d .claude/lib
```

### Testing Protocol

**After Implementation**:
1. Run `bash tests/plugin_structure_test.sh`
2. Validate JSON syntax with `jq`
3. Verify all required directories exist
4. Check that mcp-config.json is valid JSON
5. **HALT if validation fails after 2 fix attempts**

### Dependencies

None (foundation stage)

### Estimated Time

1.5 hours (140 lines implementation + tests)

---

## Stage 20.2: Automated Installation Script

### Objective

Enhance the existing install.sh script to support platform detection, MCP configuration merging, and installation verification.

### Acceptance Criteria

- [ ] Platform detection works (Linux, macOS, Windows via Git Bash)
- [ ] Architecture detection works (amd64, arm64)
- [ ] Binaries installed to correct locations (~/.local/bin or custom path)
- [ ] MCP configuration merges safely (preserves existing servers)
- [ ] Slash commands and subagents copied to ~/.claude/
- [ ] Post-install verification runs successfully
- [ ] Rollback support on installation failure
- [ ] User-friendly error messages
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `tests/install_test.sh` (~60 lines)

```bash
#!/bin/bash
# Test functions:
# - test_platform_detection - Detect OS (Linux/Darwin/Windows)
# - test_architecture_detection - Detect arch (amd64/arm64)
# - test_binary_installation - Verify binaries copied and executable
# - test_mcp_config_merge - Test safe config merging
# - test_installation_verification - Post-install checks
# - test_rollback_on_failure - Rollback if installation fails
# - test_uninstall - Verify uninstall removes all components
```

**Test Strategy**:
1. Mock different platform environments (Linux, macOS, Windows)
2. Test binary selection based on platform+arch
3. Test MCP config merging with existing ~/.claude/mcp.json
4. Verify rollback restores previous state
5. Test uninstall removes all installed files

**Implementation File**: `scripts/install.sh` (enhanced, ~100 lines total, ~70 lines new)

```bash
#!/bin/bash
# meta-cc installer (enhanced)
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Error handling
trap 'error_exit "Installation failed at line $LINENO"' ERR

error_exit() {
    echo -e "${RED}ERROR: $1${NC}" >&2
    exit 1
}

info() {
    echo -e "${GREEN}✓${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Platform and architecture detection
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux*)   PLATFORM="linux" ;;
        darwin*)  PLATFORM="darwin" ;;
        mingw*|msys*|cygwin*) PLATFORM="windows" ;;
        *) error_exit "Unsupported OS: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error_exit "Unsupported architecture: $ARCH" ;;
    esac

    PLATFORM_ARCH="${PLATFORM}-${ARCH}"
}

# Install binaries
install_binaries() {
    INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
    mkdir -p "$INSTALL_DIR"

    # Select correct binary for platform
    if [ "$PLATFORM" = "windows" ]; then
        BINARY_EXT=".exe"
    else
        BINARY_EXT=""
    fi

    # Copy binaries
    cp "bin/meta-cc${BINARY_EXT}" "$INSTALL_DIR/" || error_exit "Failed to copy meta-cc binary"
    cp "bin/meta-cc-mcp${BINARY_EXT}" "$INSTALL_DIR/" || error_exit "Failed to copy meta-cc-mcp binary"

    # Set executable permissions (not needed on Windows)
    if [ "$PLATFORM" != "windows" ]; then
        chmod +x "$INSTALL_DIR/meta-cc" "$INSTALL_DIR/meta-cc-mcp"
    fi

    info "Binaries installed to $INSTALL_DIR"
}

# Install Claude Code integration files
install_claude_files() {
    CLAUDE_DIR="${HOME}/.claude"
    mkdir -p "$CLAUDE_DIR/commands" "$CLAUDE_DIR/agents"

    cp -r .claude/commands/* "$CLAUDE_DIR/commands/" || error_exit "Failed to copy slash commands"
    cp -r .claude/agents/* "$CLAUDE_DIR/agents/" || error_exit "Failed to copy subagents"

    info "Claude Code files installed to $CLAUDE_DIR"
}

# Merge MCP configuration
merge_mcp_config() {
    MCP_CONFIG="${HOME}/.claude/mcp.json"
    MCP_TEMPLATE=".claude/lib/mcp-config.json"

    if [ ! -f "$MCP_CONFIG" ]; then
        # No existing config, copy template
        cp "$MCP_TEMPLATE" "$MCP_CONFIG"
        info "MCP configuration created at $MCP_CONFIG"
    else
        # Merge with existing config using jq
        if command -v jq >/dev/null 2>&1; then
            TEMP_CONFIG=$(mktemp)
            jq -s '.[0] * .[1]' "$MCP_CONFIG" "$MCP_TEMPLATE" > "$TEMP_CONFIG"
            mv "$TEMP_CONFIG" "$MCP_CONFIG"
            info "MCP configuration merged (existing servers preserved)"
        else
            warn "jq not found, skipping MCP config merge. Please manually add meta-cc to $MCP_CONFIG"
        fi
    fi
}

# Verify installation
verify_installation() {
    INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"

    # Check binaries exist and are executable
    if [ ! -x "$INSTALL_DIR/meta-cc" ]; then
        error_exit "meta-cc binary not found or not executable"
    fi

    # Test binary runs
    if ! "$INSTALL_DIR/meta-cc" --version >/dev/null 2>&1; then
        error_exit "meta-cc binary fails to execute"
    fi

    info "Installation verified successfully"
}

# Main installation flow
main() {
    echo "Installing meta-cc..."

    detect_platform
    info "Detected platform: $PLATFORM_ARCH"

    install_binaries
    install_claude_files
    merge_mcp_config
    verify_installation

    echo ""
    echo "Installation complete! 🎉"
    echo ""
    echo "Next steps:"
    echo "1. Add to PATH (if needed): export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo "2. Restart Claude Code to load the plugin"
    echo "3. Test with: meta-cc --version"
    echo ""
}

main "$@"
```

**Implementation File**: `scripts/uninstall.sh` (~50 lines new)

```bash
#!/bin/bash
# meta-cc uninstaller
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

info() {
    echo -e "${GREEN}✓${NC} $1"
}

INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
CLAUDE_DIR="${HOME}/.claude"

echo "Uninstalling meta-cc..."

# Remove binaries
rm -f "$INSTALL_DIR/meta-cc" "$INSTALL_DIR/meta-cc-mcp"
info "Binaries removed from $INSTALL_DIR"

# Remove Claude Code files
rm -rf "$CLAUDE_DIR/commands/meta-"*
rm -rf "$CLAUDE_DIR/agents/meta-"*
info "Claude Code files removed from $CLAUDE_DIR"

# Note: We don't remove MCP config to preserve user's other servers
echo ""
echo "Uninstallation complete!"
echo ""
echo "Note: MCP configuration at ~/.claude/mcp.json was preserved."
echo "To remove meta-cc MCP server, manually edit ~/.claude/mcp.json"
```

### File Changes

**Modified Files**:
- `scripts/install.sh` (+70 lines, ~100 total)

**New Files**:
- `scripts/uninstall.sh` (+50 lines)
- `tests/install_test.sh` (+60 lines)

**Total**: ~180 lines (exceeds 150-line target by 30 lines, acceptable for complex script)

### Test Commands

```bash
# Run Stage 20.2 tests
bash tests/install_test.sh

# Test platform detection
bash -c 'source scripts/install.sh; detect_platform; echo "Platform: $PLATFORM_ARCH"'

# Test installation (dry run)
INSTALL_DIR=/tmp/test-install bash scripts/install.sh

# Verify binaries installed
test -x /tmp/test-install/meta-cc && echo "Binary installed successfully"

# Test uninstall
INSTALL_DIR=/tmp/test-install bash scripts/uninstall.sh
```

### Testing Protocol

**After Implementation**:
1. Run `bash tests/install_test.sh`
2. Test installation in isolated directory
3. Verify MCP config merging (with and without existing config)
4. Test uninstall removes all components
5. Test on multiple platforms (Linux, macOS if available)
6. **HALT if tests fail after 2 fix attempts**

### Dependencies

- Stage 20.1 (plugin structure)

### Estimated Time

2 hours (180 lines implementation + tests)

---

## Stage 20.3: GitHub Release Workflow Enhancement

### Objective

Enhance the existing GitHub Release workflow to build plugin packages for all platforms and create comprehensive releases.

### Acceptance Criteria

- [ ] Workflow builds for all 5 platforms (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64)
- [ ] Plugin packages include binaries, .claude files, install/uninstall scripts, plugin.json
- [ ] Checksums generated for all artifacts
- [ ] Release notes auto-generated from commits/PRs
- [ ] Workflow triggered by version tags (v*.*.*)
- [ ] Artifacts uploaded to GitHub Release
- [ ] Workflow completes within 10 minutes
- [ ] Integration test validates workflow output

### TDD Approach

**Test File**: `tests/release_workflow_test.sh` (~40 lines)

```bash
#!/bin/bash
# Test functions:
# - test_workflow_syntax - Validate YAML syntax
# - test_platform_matrix - Verify all 5 platforms defined
# - test_artifact_structure - Check plugin package contents
# - test_checksum_generation - Verify checksums.txt created
# - test_version_extraction - Test version from git tag
```

**Test Strategy**:
1. Validate workflow YAML syntax
2. Verify platform build matrix is complete
3. Test artifact packaging locally (simulate workflow)
4. Verify checksums are correct
5. Test with mock git tag

**Implementation File**: `.github/workflows/release.yml` (enhanced, ~120 lines total, ~20 lines new)

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  build:
    name: Build and Release
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Get version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      - name: Update plugin.json version
        run: |
          VERSION=${{ steps.version.outputs.VERSION }}
          jq --arg ver "${VERSION#v}" '.version = $ver' plugin.json > plugin.json.tmp
          mv plugin.json.tmp plugin.json

      - name: Build binaries
        run: |
          mkdir -p build

          # Linux amd64
          GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-linux-amd64 .
          GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-mcp-linux-amd64 ./cmd/mcp-server

          # Linux arm64
          GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-linux-arm64 .
          GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-mcp-linux-arm64 ./cmd/mcp-server

          # macOS amd64
          GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-darwin-amd64 .
          GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-mcp-darwin-amd64 ./cmd/mcp-server

          # macOS arm64
          GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-darwin-arm64 .
          GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-mcp-darwin-arm64 ./cmd/mcp-server

          # Windows amd64
          GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-windows-amd64.exe .
          GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=${{ steps.version.outputs.VERSION }}" -o build/meta-cc-mcp-windows-amd64.exe ./cmd/mcp-server

      - name: Create plugin packages
        run: |
          VERSION=${{ steps.version.outputs.VERSION }}
          mkdir -p build/packages

          for platform in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64; do
            PKG_DIR=build/packages/meta-cc-plugin-${platform}
            mkdir -p $PKG_DIR/bin $PKG_DIR/.claude/commands $PKG_DIR/.claude/agents $PKG_DIR/.claude/lib

            # Copy binaries
            if [[ $platform == windows-* ]]; then
              cp build/meta-cc-${platform}.exe $PKG_DIR/bin/meta-cc.exe
              cp build/meta-cc-mcp-${platform}.exe $PKG_DIR/bin/meta-cc-mcp.exe
            else
              cp build/meta-cc-${platform} $PKG_DIR/bin/meta-cc
              cp build/meta-cc-mcp-${platform} $PKG_DIR/bin/meta-cc-mcp
            fi

            # Copy Claude Code integration files
            cp -r .claude/commands/* $PKG_DIR/.claude/commands/
            cp -r .claude/agents/* $PKG_DIR/.claude/agents/
            cp -r .claude/lib/* $PKG_DIR/.claude/lib/

            # Copy scripts and metadata
            cp scripts/install.sh $PKG_DIR/
            cp scripts/uninstall.sh $PKG_DIR/
            cp plugin.json $PKG_DIR/
            cp README.md $PKG_DIR/
            cp LICENSE $PKG_DIR/

            # Create archive
            cd build/packages
            tar -czf meta-cc-plugin-${VERSION}-${platform}.tar.gz meta-cc-plugin-${platform}
            cd ../..
          done

      - name: Generate checksums
        run: |
          cd build/packages
          sha256sum *.tar.gz > checksums.txt

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: build/packages/*
          generate_release_notes: true
          draft: false
          prerelease: ${{ contains(steps.version.outputs.VERSION, '-') }}
          body: |
            ## Installation

            Download the appropriate package for your platform and run:

            ```bash
            tar -xzf meta-cc-plugin-${{ steps.version.outputs.VERSION }}-<platform>.tar.gz
            cd meta-cc-plugin-<platform>
            ./install.sh
            ```

            See [Installation Guide](https://github.com/yaleh/meta-cc#installation) for details.
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### File Changes

**Modified Files**:
- `.github/workflows/release.yml` (+20 lines, ~120 total)

**New Files**:
- `tests/release_workflow_test.sh` (+40 lines)

**Total**: ~60 lines

### Test Commands

```bash
# Validate workflow syntax
yamllint .github/workflows/release.yml

# Run workflow test
bash tests/release_workflow_test.sh

# Simulate local build (without git tag)
mkdir -p build
GOOS=linux GOARCH=amd64 go build -o build/meta-cc-linux-amd64 .
GOOS=linux GOARCH=amd64 go build -o build/meta-cc-mcp-linux-amd64 ./cmd/mcp-server

# Test package creation locally
bash -c 'source .github/workflows/release.yml; create_plugin_packages'
```

### Testing Protocol

**After Implementation**:
1. Validate YAML syntax with `yamllint`
2. Run `bash tests/release_workflow_test.sh`
3. Test workflow with pre-release tag (v0.12.0-beta)
4. Verify all 5 platform packages created
5. Download and test installation from package
6. **HALT if workflow fails after 2 fix attempts**

### Dependencies

- Stage 20.1 (plugin.json)
- Stage 20.2 (install.sh)

### Estimated Time

1.5 hours (60 lines implementation + tests)

---

## Stage 20.4: Documentation and Testing

### Objective

Create comprehensive installation documentation, test on multiple platforms, and update project documentation.

### Acceptance Criteria

- [ ] Installation guide covers all 5 platforms
- [ ] Troubleshooting section addresses common issues
- [ ] CHANGELOG.md updated with Phase 20 changes
- [ ] README.md updated with plugin installation instructions
- [ ] Manual testing completed on Linux and macOS (Windows via CI)
- [ ] Plugin verification checklist completed
- [ ] All documentation cross-references verified

### Documentation Changes

**1. docs/installation.md** (new, ~150 lines)

```markdown
# Installation Guide

## Quick Install (Recommended)

### Linux (x86_64)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-amd64.tar.gz | tar xz
cd meta-cc-plugin-linux-amd64
./install.sh
```

### Linux (ARM64)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-arm64.tar.gz | tar xz
cd meta-cc-plugin-linux-arm64
./install.sh
```

### macOS (Intel)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-darwin-amd64.tar.gz | tar xz
cd meta-cc-plugin-darwin-amd64
./install.sh
```

### macOS (Apple Silicon)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-darwin-arm64.tar.gz | tar xz
cd meta-cc-plugin-darwin-arm64
./install.sh
```

### Windows (x86_64)
```powershell
# Download from GitHub Releases
# Extract meta-cc-plugin-windows-amd64.tar.gz
# Run install.sh via Git Bash
```

## Manual Installation

If the automated installer fails, follow these steps:

1. **Download binaries**
   ```bash
   # Download meta-cc and meta-cc-mcp for your platform
   wget https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-<platform>
   wget https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-mcp-<platform>
   ```

2. **Install binaries**
   ```bash
   mkdir -p ~/.local/bin
   mv meta-cc-<platform> ~/.local/bin/meta-cc
   mv meta-cc-mcp-<platform> ~/.local/bin/meta-cc-mcp
   chmod +x ~/.local/bin/meta-cc ~/.local/bin/meta-cc-mcp
   ```

3. **Install Claude Code files**
   ```bash
   mkdir -p ~/.claude/commands ~/.claude/agents
   cp .claude/commands/* ~/.claude/commands/
   cp .claude/agents/* ~/.claude/agents/
   ```

4. **Configure MCP**
   Edit `~/.claude/mcp.json`:
   ```json
   {
     "mcpServers": {
       "meta-cc": {
         "command": "meta-cc-mcp",
         "args": [],
         "disabled": false
       }
     }
   }
   ```

## Verification

After installation, verify the setup:

```bash
# Check binary version
meta-cc --version

# Test MCP server (if Claude Code is running)
# Open Claude Code and type: @meta-coach

# Test slash command
# Open Claude Code and type: /meta-stats
```

## Troubleshooting

### Binary not found

**Issue**: `meta-cc: command not found`

**Solution**: Add `~/.local/bin` to PATH:
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### MCP server not connecting

**Issue**: MCP server fails to start

**Solution**: Check MCP logs in Claude Code settings

### Slash commands not working

**Issue**: Slash commands not recognized

**Solution**: Restart Claude Code after installation

## Uninstallation

To remove meta-cc:

```bash
cd meta-cc-plugin-<platform>
./uninstall.sh
```

Or manually:
```bash
rm ~/.local/bin/meta-cc ~/.local/bin/meta-cc-mcp
rm -rf ~/.claude/commands/meta-*
rm -rf ~/.claude/agents/meta-*
# Manually edit ~/.claude/mcp.json to remove meta-cc server
```
```

**2. CHANGELOG.md** (+20 lines)

```markdown
## [0.12.0] - 2025-10-10

### Added (Phase 20)
- Plugin packaging structure with plugin.json manifest
- Automated installation script with platform detection
- MCP configuration merging (preserves existing servers)
- Uninstall script for clean removal
- Multi-platform plugin packages (Linux, macOS, Windows)
- GitHub Release workflow enhancements
- Comprehensive installation documentation

### Changed
- install.sh now supports platform detection and verification
- Release workflow creates plugin packages instead of bare binaries
- MCP configuration now merged safely (no overwriting)

### Fixed
- Installation on ARM64 platforms
- MCP config conflicts with existing servers
```

**3. README.md** (+30 lines)

```markdown
## Installation

### Plugin Installation (Recommended)

Download and install the meta-cc plugin for your platform:

#### Linux (x86_64)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-amd64.tar.gz | tar xz
cd meta-cc-plugin-linux-amd64
./install.sh
```

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-darwin-arm64.tar.gz | tar xz
cd meta-cc-plugin-darwin-arm64
./install.sh
```

See [Installation Guide](docs/installation.md) for all platforms and troubleshooting.

### Verification

After installation:
```bash
meta-cc --version
```

Open Claude Code and test:
- Slash command: `/meta-stats`
- Subagent: `@meta-coach`
- MCP tools: Automatically available in conversations
```

### Testing Checklist

Create `tests/PLUGIN_VERIFICATION.md` (~50 lines):

```markdown
# Plugin Verification Checklist

## Pre-Release Testing

### Platform Testing

- [ ] Linux amd64 installation
- [ ] Linux arm64 installation
- [ ] macOS Intel installation
- [ ] macOS ARM installation
- [ ] Windows installation (Git Bash)

### Component Testing

- [ ] Binaries execute successfully
- [ ] Slash commands functional (`/meta-stats`)
- [ ] Subagents functional (`@meta-coach`)
- [ ] MCP server connects
- [ ] MCP tools respond correctly

### Installation Testing

- [ ] Fresh install (no existing meta-cc)
- [ ] Upgrade install (existing meta-cc)
- [ ] MCP config merge preserves existing servers
- [ ] Uninstall removes all components
- [ ] Reinstall after uninstall works

### Documentation Testing

- [ ] Installation guide commands work
- [ ] Troubleshooting steps resolve issues
- [ ] README examples execute correctly
- [ ] All cross-references valid

### Release Testing

- [ ] GitHub Release workflow succeeds
- [ ] All 5 platform packages created
- [ ] Checksums validate
- [ ] Release notes accurate
- [ ] Download links functional
```

### File Changes

**New Files**:
- `docs/installation.md` (+150 lines)
- `tests/PLUGIN_VERIFICATION.md` (+50 lines)

**Modified Files**:
- `CHANGELOG.md` (+20 lines)
- `README.md` (+30 lines)

**Total**: ~250 lines (exceeds 50-line target significantly, but documentation is critical)

### Verification Commands

```bash
# Test installation guide commands
bash -c "$(grep -A2 'curl -L' docs/installation.md | head -1)"

# Verify all links
grep -r 'https://' docs/*.md | while read line; do
  URL=$(echo "$line" | grep -oP 'https://[^\s)]+')
  curl -I "$URL" >/dev/null 2>&1 && echo "✓ $URL" || echo "✗ $URL"
done

# Check cross-references
grep -r '\[.*\](.*\.md)' docs/ README.md CLAUDE.md

# Verify CHANGELOG version
grep '^\[0.12.0\]' CHANGELOG.md
```

### Testing Protocol

**Manual Testing**:
1. Test installation on Linux (primary development platform)
2. Test installation on macOS (if available)
3. Verify MCP server connects
4. Test slash commands and subagents
5. Test uninstall and reinstall
6. Verify documentation accuracy

**After Documentation**:
1. Review all updated documentation for accuracy
2. Test installation commands from docs
3. Verify troubleshooting steps
4. Check cross-references between documents
5. Validate CHANGELOG.md completeness

### Dependencies

- Stage 20.3 (GitHub Release workflow)

### Estimated Time

2 hours (250 lines documentation + manual testing)

---

## Phase Integration Strategy

### Build Verification

After completing all stages, verify the complete Phase 20 implementation:

```bash
# 1. Validate plugin structure
bash tests/plugin_structure_test.sh

# 2. Test installation script
bash tests/install_test.sh

# 3. Validate GitHub workflow
bash tests/release_workflow_test.sh

# 4. Test local installation
INSTALL_DIR=/tmp/test-meta-cc bash scripts/install.sh

# 5. Verify binaries
/tmp/test-meta-cc/meta-cc --version

# 6. Manual platform testing (see PLUGIN_VERIFICATION.md)
```

### Release Preparation

Before creating the first plugin release:

1. **Version Bump**: Update version in `plugin.json` to match git tag
2. **CHANGELOG**: Ensure Phase 20 changes documented
3. **Documentation**: Verify all docs reference correct version
4. **Testing**: Complete verification checklist
5. **Tag**: Create git tag `v0.12.0` (or appropriate version)
6. **Release**: Push tag to trigger GitHub workflow
7. **Verify**: Download and test plugin package

### Rollout Checklist

Before marking Phase 20 complete:

- [ ] All 4 stages completed and tested
- [ ] plugin.json validates correctly
- [ ] install.sh works on Linux (primary platform)
- [ ] GitHub Release workflow succeeds
- [ ] All 5 platform packages created
- [ ] Documentation complete and accurate
- [ ] Manual testing completed (Linux + macOS if available)
- [ ] Verification checklist completed
- [ ] CHANGELOG.md updated
- [ ] README.md updated
- [ ] Git commit includes Phase 20 changes

---

## File Change Inventory

### Summary by Stage

| Stage | New Files | Modified Files | Total Lines |
|-------|-----------|----------------|-------------|
| 20.1  | 4         | 0              | ~140        |
| 20.2  | 2         | 1              | ~180        |
| 20.3  | 1         | 1              | ~60         |
| 20.4  | 2         | 2              | ~250        |
| **Total** | **9** | **4** | **~630** |

Note: Total exceeds 400-line target by ~230 lines. This is acceptable because:
- Stage 20.1: 140 lines (foundation, includes tests)
- Stage 20.2: 180 lines (complex script with error handling)
- Stage 20.4: 250 lines (comprehensive documentation)

Core implementation (20.2, 20.3) is within budget (~240 lines). Documentation overhead is justified for user-facing release.

### Detailed File Changes

**New Files (9)**:
1. `plugin.json` (60 lines)
2. `.claude/lib/mcp-config.json` (30 lines)
3. `docs/plugin-structure.md` (10 lines)
4. `tests/plugin_structure_test.sh` (40 lines)
5. `scripts/uninstall.sh` (50 lines)
6. `tests/install_test.sh` (60 lines)
7. `tests/release_workflow_test.sh` (40 lines)
8. `docs/installation.md` (150 lines)
9. `tests/PLUGIN_VERIFICATION.md` (50 lines)

**Modified Files (4)**:
1. `scripts/install.sh` (+70 lines)
2. `.github/workflows/release.yml` (+20 lines)
3. `CHANGELOG.md` (+20 lines)
4. `README.md` (+30 lines)

---

## Risk Assessment and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Platform detection fails on Windows | Medium | High | Test with Git Bash, provide manual fallback |
| MCP config merge corrupts existing config | Low | Critical | Backup existing config, use jq for safe merge |
| GitHub workflow fails for ARM platforms | Low | Medium | Test cross-compilation locally first |
| Installation script permissions issues | Medium | Medium | Comprehensive error handling, clear messages |
| Plugin not recognized by Claude Code | Low | High | Follow Claude Code plugin spec exactly |
| Documentation becomes outdated | Medium | Low | Version-specific docs, automated checks |

### Contingency Plans

**If platform detection fails**:
- Provide manual platform selection via `--platform` flag
- Document manual installation fallback

**If MCP config merge fails**:
- Fall back to warning user to manually add meta-cc
- Provide exact JSON snippet to copy

**If GitHub workflow fails**:
- Test all steps locally first
- Use act (https://github.com/nektos/act) to test workflows locally

**If installation verification fails**:
- Provide detailed troubleshooting guide
- Add `--skip-verify` flag for advanced users

**If testing fails repeatedly**:
- HALT development per testing protocol
- Document blockers and failure state
- Seek community feedback on installation issues

---

## Testing Strategy

### Unit Testing

**Coverage Requirements**:
- Shell scripts: Test key functions (platform detection, config merge)
- Workflow: Validate YAML syntax, test packaging locally
- Documentation: Verify all commands execute successfully

**Test Organization**:
```
tests/
  plugin_structure_test.sh       - Plugin metadata validation
  install_test.sh                - Installation script tests
  release_workflow_test.sh       - Workflow validation
  PLUGIN_VERIFICATION.md         - Manual testing checklist
```

### Integration Testing

**Platform Testing**:
- Primary: Linux amd64 (development platform)
- Secondary: macOS ARM (if available)
- CI: All 5 platforms via GitHub Actions

**End-to-End Workflows**:
```bash
# Full installation test
curl -L <plugin-url> | tar xz
cd meta-cc-plugin-*
./install.sh
meta-cc --version
# Test slash commands in Claude Code
# Test MCP server connection

# Upgrade test
# (with existing installation)
./install.sh
# Verify existing MCP servers preserved

# Uninstall test
./uninstall.sh
# Verify all components removed
```

### Regression Testing

**Verify No Breaking Changes**:
```bash
# Existing manual installation should work
make build
cp meta-cc meta-cc-mcp ~/.local/bin/

# Existing release workflow should work
git tag v0.12.0-test
git push --tags
# Verify GitHub workflow succeeds
```

### Documentation Testing

**Verification Steps**:
1. Execute all code blocks in installation.md
2. Verify all links resolve (internal and external)
3. Test troubleshooting steps resolve common issues
4. Verify version numbers consistent across docs

---

## Post-Phase Verification

### Functional Verification

After completing Phase 20, verify:

1. **Plugin Structure**:
   ```bash
   jq '.name, .version, .platforms' plugin.json
   test -f .claude/lib/mcp-config.json && echo "MCP template exists"
   ```

2. **Installation Works**:
   ```bash
   # Clean environment
   rm -rf /tmp/test-install

   # Test installation
   INSTALL_DIR=/tmp/test-install bash scripts/install.sh

   # Verify
   /tmp/test-install/meta-cc --version
   ```

3. **GitHub Workflow Works**:
   ```bash
   # Trigger workflow with test tag
   git tag v0.12.0-test
   git push origin v0.12.0-test

   # Wait for workflow completion
   # Download and test package
   ```

4. **Documentation Accurate**:
   ```bash
   # Test installation guide commands
   bash -c "$(head -1 docs/installation.md | grep curl)"

   # Verify troubleshooting steps
   # (manually test each step)
   ```

5. **Uninstall Works**:
   ```bash
   INSTALL_DIR=/tmp/test-install bash scripts/uninstall.sh
   test ! -f /tmp/test-install/meta-cc && echo "Uninstall successful"
   ```

### Release Verification

Before public release:

1. **Create Pre-Release**:
   - Tag: `v0.12.0-beta`
   - Test complete installation flow
   - Verify all platforms download correctly

2. **Community Testing**:
   - Share beta release with early adopters
   - Collect feedback on installation issues
   - Address critical bugs

3. **Final Release**:
   - Tag: `v0.12.0`
   - Update all documentation
   - Announce on GitHub

---

## Success Metrics

### Quantitative Metrics

- **Installation**:
  - Installation completes in <2 minutes
  - Success rate >95% on supported platforms
  - Zero data loss in MCP config merging

- **Functionality**:
  - All 5 platform packages build successfully
  - Plugin recognized by Claude Code
  - MCP server connects on first try

- **Documentation**:
  - All installation commands execute successfully
  - Troubleshooting resolves >90% of issues
  - Zero broken links or references

### Qualitative Metrics

- **Usability**:
  - One-command installation
  - Clear error messages
  - Helpful troubleshooting guide

- **Reliability**:
  - Safe MCP config merging
  - Rollback on installation failure
  - Clean uninstallation

- **Maintainability**:
  - Clear plugin structure
  - Documented installation flow
  - Automated release process

---

## Timeline Estimate

| Stage | Description | Estimated Time |
|-------|-------------|----------------|
| 20.1  | Plugin structure | 1.5 hours |
| 20.2  | Installation script | 2 hours |
| 20.3  | GitHub workflow | 1.5 hours |
| 20.4  | Documentation | 2 hours |
| **Total** | **All stages** | **7 hours** |

**Contingency**: +3 hours for platform testing and issue resolution (total: 10 hours)

---

## Conclusion

Phase 20 transforms meta-cc from a development tool into a production-ready plugin with professional-grade installation and distribution. The plugin packaging approach provides:

1. **User Experience**: One-command installation
2. **Safety**: Safe MCP config merging preserves user settings
3. **Reliability**: Platform detection and verification
4. **Maintainability**: Automated release workflow
5. **Documentation**: Comprehensive guides and troubleshooting

Key success factors:
- Plugin structure follows Claude Code standards
- Installation script handles edge cases gracefully
- GitHub workflow automates multi-platform builds
- Documentation covers all platforms and scenarios
- Testing ensures reliability across platforms

Upon completion, meta-cc will be ready for public distribution as a professional Claude Code plugin, enabling users to benefit from metacognitive workflow analysis with minimal setup friction.

---

## Next Steps (Post-Phase 20)

After Phase 20 completion:

1. **Public Release**:
   - Create v0.12.0 release
   - Announce on GitHub, social media
   - Submit to Claude Code plugin directory (if available)

2. **User Feedback**:
   - Monitor installation issues
   - Address platform-specific bugs
   - Improve documentation based on feedback

3. **Future Enhancements**:
   - Auto-update mechanism
   - Plugin dependency management
   - Integration with Claude Code plugin marketplace

4. **Community Growth**:
   - Encourage contributions
   - Build example workflows
   - Create tutorial videos

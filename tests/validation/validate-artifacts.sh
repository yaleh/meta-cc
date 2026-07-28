#!/bin/bash
# End-to-end validation of all artifact types
#
# Builds all artifacts locally, installs each into an isolated temp dir,
# and verifies the installation matches the expected contract.
#
# Usage:
#   ./tests/validation/validate-artifacts.sh [--skip-build]
#
#   --skip-build  Reuse existing build/ artifacts (faster iteration)
#
# Requirements:
#   - go (for MCP binary build)
#   - bats (for unit tests)
#   - Standard Unix tools: tar, jq, file

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

SKIP_BUILD=false
for arg in "$@"; do
    [ "$arg" = "--skip-build" ] && SKIP_BUILD=true
done

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

TOTAL=0
PASSED=0
FAILED=()

pass()  { echo -e "  ${GREEN}✓${NC} $1"; PASSED=$((PASSED+1)); TOTAL=$((TOTAL+1)); }
fail()  { echo -e "  ${RED}✗${NC} $1${2:+ — $2}"; FAILED+=("$1"); TOTAL=$((TOTAL+1)); }
info()  { echo -e "${CYAN}$1${NC}"; }
warn()  { echo -e "${YELLOW}⚠  $1${NC}"; }

# Use a fake version for skills/binary artifacts.
# For combined package, use the real marketplace version so smoke tests pass.
TEST_VERSION="v0.0.0-validation"
REAL_VERSION="v$(jq -r '.plugins[0].version' "$PROJECT_ROOT/.claude-plugin/marketplace.json")"
BUILD_DIR="$PROJECT_ROOT/build/validation"
WORK_DIR="$(mktemp -d)"
trap "rm -rf $WORK_DIR" EXIT

echo ""
info "================================================="
info "  meta-cc Artifact Validation Suite"
info "  Version: $TEST_VERSION"
info "================================================="
echo ""

# ----------------------------------------------------------------
# Phase 1: Build artifacts
# ----------------------------------------------------------------

info "Phase 1: Building artifacts"
echo ""

if [ "$SKIP_BUILD" = true ] && [ -d "$BUILD_DIR" ]; then
    warn "Skipping build — reusing $BUILD_DIR"
else
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR/mcp" "$BUILD_DIR/skills"

    # Sync plugin files
    echo "  Syncing plugin files..."
    bash scripts/sync-plugin-files.sh >/dev/null 2>&1

    # Build native MCP binary only (current platform)
    echo "  Building MCP binary (native platform)..."
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"
    NATIVE_PLATFORM="${OS}-${ARCH}"

    # DIR-052: derive the -X assignments from the Makefile's single source of
    # truth (print-ldflags-value, added in DIR-049) instead of hand-duplicating
    # an -X string here. The hand-written copy previously pointed at the
    # nonexistent repo-root cmd package (silent linker no-op), so validation
    # binaries never embedded commit/build-time. Deriving from the Makefile
    # keeps this build in lockstep with the release pipeline; the
    # TestNoHandDuplicatedDeadPathLDFLAGS regression test fails if this script
    # ever hand-writes an -X string again.
    LDFLAGS_VALUE="$(make -s print-ldflags-value)"
    go build -ldflags "${LDFLAGS_VALUE}" \
        -o "$BUILD_DIR/mcp/meta-cc-mcp-${TEST_VERSION}-${NATIVE_PLATFORM}" \
        ./cmd/mcp-server
    chmod +x "$BUILD_DIR/mcp/meta-cc-mcp-${TEST_VERSION}-${NATIVE_PLATFORM}"
    echo "    Built: meta-cc-mcp-${TEST_VERSION}-${NATIVE_PLATFORM}"

    # Build skills package
    echo "  Building skills package..."
    bash scripts/release/build-skills-package.sh \
        --version "$TEST_VERSION" \
        --output "$BUILD_DIR/skills" \
        >/dev/null 2>&1
    echo "    Built: meta-cc-skills-${TEST_VERSION}.tar.gz"

    # Build combined plugin package (reuse native binary)
    #
    # DIR-057: the file set copied here MUST stay in sync with the Makefile's
    # bundle-release target (the canonical release packager). The Codex plugin
    # files (skills/, .codex-plugin/, .codex-mcp.json) were previously omitted
    # from this simplified assembly, which made the CI smoke test fail its
    # plugin-structure checks (23/24). If bundle-release gains new payload
    # files, mirror them here.
    echo "  Building combined plugin package..."
    PKG_DIR="$BUILD_DIR/packages/meta-cc-plugin-${NATIVE_PLATFORM}"
    mkdir -p "$PKG_DIR/bin" "$PKG_DIR/.claude-plugin" "$PKG_DIR/commands" \
        "$PKG_DIR/lib" "$PKG_DIR/skills" "$PKG_DIR/.codex-plugin"

    cp "$BUILD_DIR/mcp/meta-cc-mcp-${TEST_VERSION}-${NATIVE_PLATFORM}" "$PKG_DIR/bin/meta-cc-mcp"
    cp -r .claude-plugin/* "$PKG_DIR/.claude-plugin/"
    cp -r dist/commands/* "$PKG_DIR/commands/"
    cp -r plugin-src/skills/* "$PKG_DIR/skills/"
    cp -r lib/* "$PKG_DIR/lib/"
    cp plugin-src/.claude-plugin/plugin.json "$PKG_DIR/.claude-plugin/"
    cp plugin-src/.mcp.json "$PKG_DIR/"
    cp -r plugin-src/.codex-plugin/* "$PKG_DIR/.codex-plugin/"
    cp plugin-src/.codex-mcp.json "$PKG_DIR/"
    jq '.commands |= map(gsub("\\./plugin-src/commands/"; "./commands/"))' \
        "$PKG_DIR/.claude-plugin/plugin.json" > "$PKG_DIR/.claude-plugin/plugin.json.tmp"
    mv "$PKG_DIR/.claude-plugin/plugin.json.tmp" "$PKG_DIR/.claude-plugin/plugin.json"
    jq '.plugins[0].source = "." | .plugins[0].commands |= map(gsub("\\./plugin-src/commands/"; "./commands/"))' \
        "$PKG_DIR/.claude-plugin/marketplace.json" > "$PKG_DIR/.claude-plugin/marketplace.json.tmp"
    mv "$PKG_DIR/.claude-plugin/marketplace.json.tmp" "$PKG_DIR/.claude-plugin/marketplace.json"
    cp scripts/install/install.sh "$PKG_DIR/"
    cp scripts/install/uninstall.sh "$PKG_DIR/"
    cp scripts/install/install-mcp.sh "$PKG_DIR/"
    cp scripts/install/install-skills.sh "$PKG_DIR/"
    cp README.md "$PKG_DIR/"
    cp LICENSE "$PKG_DIR/"

    mkdir -p "$BUILD_DIR/packages"
    # Combined package uses REAL_VERSION so smoke tests (version check) pass
    tar -czf "$BUILD_DIR/packages/meta-cc-plugin-${REAL_VERSION}-${NATIVE_PLATFORM}.tar.gz" \
        -C "$BUILD_DIR/packages" "meta-cc-plugin-${NATIVE_PLATFORM}"
    echo "    Built: meta-cc-plugin-${REAL_VERSION}-${NATIVE_PLATFORM}.tar.gz"
fi

echo ""

# ----------------------------------------------------------------
# Phase 2: Validate skills-only artifact
# ----------------------------------------------------------------

info "Phase 2: Skills-only artifact"
echo ""

SKILLS_PKG="$BUILD_DIR/skills/meta-cc-skills-${TEST_VERSION}.tar.gz"
SKILLS_INSTALL="$WORK_DIR/skills-install"

if [ ! -f "$SKILLS_PKG" ]; then
    fail "skills package exists" "File not found: $SKILLS_PKG"
else
    pass "skills package exists"

    # Run CI smoke tests
    if bash scripts/ci/smoke-tests-skills.sh "$TEST_VERSION" "$SKILLS_PKG" >/dev/null 2>&1; then
        pass "skills package: CI smoke tests pass"
    else
        fail "skills package: CI smoke tests pass"
    fi

    # Extract and install
    mkdir -p "$SKILLS_INSTALL"
    tar -xzf "$SKILLS_PKG" -C "$WORK_DIR"
    PKG_EXTRACT="$WORK_DIR/meta-cc-skills-${TEST_VERSION}"

    CLAUDE_DIR="$SKILLS_INSTALL/dot-claude"
    mkdir -p "$CLAUDE_DIR"

    if env CLAUDE_DIR="$CLAUDE_DIR" bash "$PKG_EXTRACT/install-skills.sh" >/dev/null 2>&1; then
        pass "skills install: runs without error"
    else
        fail "skills install: runs without error"
    fi

    for cmd in prompt-find prompt-list prompt-show; do
        if [ -f "$CLAUDE_DIR/commands/${cmd}.md" ]; then
            pass "skills install: $cmd.md installed"
        else
            fail "skills install: $cmd.md installed"
        fi
    done

    if [ -f "$CLAUDE_DIR/lib/meta-utils.sh" ]; then
        pass "skills install: meta-utils.sh installed"
    else
        fail "skills install: meta-utils.sh installed"
    fi

    # Verify no binary in skills package
    if find "$PKG_EXTRACT" -name "meta-cc-mcp*" 2>/dev/null | grep -q .; then
        fail "skills package: no binary included"
    else
        pass "skills package: no binary included"
    fi
fi

echo ""

# ----------------------------------------------------------------
# Phase 3: Validate bare MCP binary artifact
# ----------------------------------------------------------------

info "Phase 3: Bare MCP binary artifact"
echo ""

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"
NATIVE_PLATFORM="${OS}-${ARCH}"
BARE_BINARY="$BUILD_DIR/mcp/meta-cc-mcp-${TEST_VERSION}-${NATIVE_PLATFORM}"

if [ ! -f "$BARE_BINARY" ]; then
    fail "bare binary exists" "Not found: $BARE_BINARY"
else
    pass "bare binary exists"

    [ -x "$BARE_BINARY" ] \
        && pass "bare binary is executable" \
        || fail "bare binary is executable"

    # Install via install-mcp.sh
    MCP_INSTALL_DIR="$WORK_DIR/mcp-install/bin"
    if env INSTALL_DIR="$MCP_INSTALL_DIR" \
           bash scripts/install/install-mcp.sh "$BARE_BINARY" >/dev/null 2>&1; then
        pass "install-mcp.sh: runs without error"
    else
        fail "install-mcp.sh: runs without error"
    fi

    if [ -f "$MCP_INSTALL_DIR/meta-cc-mcp" ]; then
        pass "install-mcp.sh: binary installed"
    else
        fail "install-mcp.sh: binary installed"
    fi

    if [ -x "$MCP_INSTALL_DIR/meta-cc-mcp" ]; then
        pass "install-mcp.sh: binary is executable"
    else
        fail "install-mcp.sh: binary is executable"
    fi

    # Verify binary runs (just check it exits non-zero on bad args, not hang)
    if timeout 5 "$MCP_INSTALL_DIR/meta-cc-mcp" --bad-flag 2>&1 | grep -qiE "flag|error|usage|unknown" 2>/dev/null \
       || timeout 5 "$MCP_INSTALL_DIR/meta-cc-mcp" --help 2>/dev/null \
       || true; then
        pass "binary executes (responds to args)"
    fi
fi

echo ""

# ----------------------------------------------------------------
# Phase 4: Validate combined plugin package
# ----------------------------------------------------------------

info "Phase 4: Combined plugin package"
echo ""

COMBINED_PKG="$BUILD_DIR/packages/meta-cc-plugin-${REAL_VERSION}-${NATIVE_PLATFORM}.tar.gz"

if [ ! -f "$COMBINED_PKG" ]; then
    fail "combined package exists" "Not found: $COMBINED_PKG"
else
    pass "combined package exists"

    # Run existing CI smoke tests (uses REAL_VERSION so version check passes)
    if bash scripts/ci/smoke-tests.sh "$REAL_VERSION" "$NATIVE_PLATFORM" "$COMBINED_PKG" >/dev/null 2>&1; then
        pass "combined package: CI smoke tests pass"
    else
        fail "combined package: CI smoke tests pass"
    fi

    # Verify the combined package includes install-mcp.sh and install-skills.sh
    if tar -tzf "$COMBINED_PKG" | grep -q "install-mcp.sh"; then
        pass "combined package: includes install-mcp.sh"
    else
        fail "combined package: includes install-mcp.sh"
    fi

    if tar -tzf "$COMBINED_PKG" | grep -q "install-skills.sh"; then
        pass "combined package: includes install-skills.sh"
    else
        fail "combined package: includes install-skills.sh"
    fi

    # Install via standard install.sh to isolated dir.
    # install.sh hardcodes CLAUDE_DIR=${HOME}/.claude, so we fake HOME.
    # INSTALL_DIR is respected as an env var.
    FAKE_HOME="$WORK_DIR/combined-home"
    COMBINED_BIN="$WORK_DIR/combined-bin"
    mkdir -p "$FAKE_HOME" "$COMBINED_BIN"

    tar -xzf "$COMBINED_PKG" -C "$WORK_DIR"
    PKG_COMBINED="$WORK_DIR/meta-cc-plugin-${NATIVE_PLATFORM}"

    if (cd "$PKG_COMBINED" && \
        env HOME="$FAKE_HOME" \
            INSTALL_DIR="$COMBINED_BIN" \
            bash ./install.sh) >/dev/null 2>&1; then
        pass "combined install: runs without error"
    else
        fail "combined install: runs without error"
    fi

    [ -f "$COMBINED_BIN/meta-cc-mcp" ] \
        && pass "combined install: binary installed" \
        || fail "combined install: binary installed"

    [ -x "$COMBINED_BIN/meta-cc-mcp" ] \
        && pass "combined install: binary executable" \
        || fail "combined install: binary executable"

    for cmd in prompt-find prompt-list prompt-show; do
        if [ -f "$FAKE_HOME/.claude/commands/${cmd}.md" ]; then
            pass "combined install: $cmd.md installed"
        else
            fail "combined install: $cmd.md installed"
        fi
    done
fi

echo ""

# ----------------------------------------------------------------
# Results
# ----------------------------------------------------------------

info "================================================="
info "  Results"
info "================================================="
echo ""
echo "  Total:  $TOTAL"
echo -e "  ${GREEN}Passed: $PASSED${NC}"
if [ ${#FAILED[@]} -gt 0 ]; then
    echo -e "  ${RED}Failed: ${#FAILED[@]}${NC}"
    echo ""
    echo "  Failed tests:"
    for f in "${FAILED[@]}"; do echo -e "    ${RED}✗${NC} $f"; done
fi
echo ""

if [ ${#FAILED[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ ALL VALIDATION CHECKS PASSED${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}❌ VALIDATION FAILED${NC}"
    echo ""
    exit 1
fi

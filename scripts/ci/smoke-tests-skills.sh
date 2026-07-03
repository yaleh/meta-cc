#!/bin/bash
# Smoke tests for meta-cc skills-only artifact
#
# Usage: ./smoke-tests-skills.sh <version> <package-path>
# Example: ./smoke-tests-skills.sh v1.0.0 build/packages/meta-cc-skills-v1.0.0.tar.gz

set -e

VERSION="$1"
PACKAGE_PATH="$2"

if [ -z "$VERSION" ] || [ -z "$PACKAGE_PATH" ]; then
    echo "Usage: $0 <version> <package-path>"
    echo "Example: $0 v1.0.0 build/packages/meta-cc-skills-v1.0.0.tar.gz"
    exit 1
fi

VERSION_NUM="${VERSION#v}"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=()

test_result() {
    local name="$1"
    local result="$2"
    local detail="${3:-}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if [ "$result" = "pass" ]; then
        echo "  ✓ $name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "  ✗ $name${detail:+ — $detail}"
        FAILED_TESTS+=("$name")
    fi
}

echo "========================================="
echo "Smoke Tests: meta-cc skills package"
echo "========================================="
echo "Version:  $VERSION"
echo "Package:  $PACKAGE_PATH"
echo ""

# Verify package exists
if [ ! -f "$PACKAGE_PATH" ]; then
    echo "ERROR: Package not found: $PACKAGE_PATH"
    exit 1
fi

TEMP_DIR="$(mktemp -d)"
trap "rm -rf $TEMP_DIR" EXIT

echo "Extracting to $TEMP_DIR..."
tar -xzf "$PACKAGE_PATH" -C "$TEMP_DIR"
echo ""

PKG_DIR="$TEMP_DIR/meta-cc-skills-${VERSION}"
if [ ! -d "$PKG_DIR" ]; then
    # Try without 'v' prefix
    PKG_DIR="$TEMP_DIR/meta-cc-skills-${VERSION_NUM}"
fi

if [ ! -d "$PKG_DIR" ]; then
    echo "ERROR: Expected directory not found after extraction"
    echo "Contents of $TEMP_DIR:"
    ls -la "$TEMP_DIR"
    exit 1
fi

echo "=== Test 1: Package Structure ==="
[ -d "$PKG_DIR/commands" ] \
    && test_result "commands/ directory exists" "pass" \
    || test_result "commands/ directory exists" "fail"

[ -d "$PKG_DIR/lib" ] \
    && test_result "lib/ directory exists" "pass" \
    || test_result "lib/ directory exists" "fail"

[ -d "$PKG_DIR/skills" ] \
    && test_result "skills/ directory exists" "pass" \
    || test_result "skills/ directory exists" "fail"

[ -f "$PKG_DIR/install-skills.sh" ] \
    && test_result "install-skills.sh exists" "pass" \
    || test_result "install-skills.sh exists" "fail"

[ -x "$PKG_DIR/install-skills.sh" ] \
    && test_result "install-skills.sh is executable" "pass" \
    || test_result "install-skills.sh is executable" "fail"

echo ""
echo "=== Test 2: Command Files ==="
for cmd in prompt-find prompt-list prompt-show; do
    if [ -f "$PKG_DIR/commands/${cmd}.md" ]; then
        test_result "commands/${cmd}.md exists" "pass"
    else
        test_result "commands/${cmd}.md exists" "fail"
    fi
done

echo ""
echo "=== Test 3: Codex Skill Files ==="
for skill in prompt-find prompt-list prompt-show meta-cc-insights; do
    if [ -f "$PKG_DIR/skills/${skill}/SKILL.md" ]; then
        test_result "skills/${skill}/SKILL.md exists" "pass"
    else
        test_result "skills/${skill}/SKILL.md exists" "fail"
    fi
done

echo ""
echo "=== Test 4: Lib Files ==="
[ -f "$PKG_DIR/lib/meta-utils.sh" ] \
    && test_result "lib/meta-utils.sh exists" "pass" \
    || test_result "lib/meta-utils.sh exists" "fail"

echo ""
echo "=== Test 5: No Binary Files ==="
if find "$PKG_DIR" -name "meta-cc-mcp*" 2>/dev/null | grep -q .; then
    test_result "No MCP binary in package (skills-only)" "fail" "Found unexpected binary"
else
    test_result "No MCP binary in package (skills-only)" "pass"
fi

echo ""
echo "=== Test 6: Installation ==="
INSTALL_TEST_DIR="$(mktemp -d)"
CODEX_TEST_DIR="$(mktemp -d)"
CLAUDE_SKILL_TEST_DIR="$INSTALL_TEST_DIR/skills"
trap "rm -rf $TEMP_DIR $INSTALL_TEST_DIR $CODEX_TEST_DIR" EXIT

if env CLAUDE_DIR="$INSTALL_TEST_DIR" CODEX_HOME="$CODEX_TEST_DIR" bash "$PKG_DIR/install-skills.sh" >/dev/null 2>&1; then
    test_result "install-skills.sh runs without error" "pass"
else
    test_result "install-skills.sh runs without error" "fail"
fi

for cmd in prompt-find prompt-list prompt-show; do
    if [ -f "$INSTALL_TEST_DIR/commands/${cmd}.md" ]; then
        test_result "Installed: commands/${cmd}.md" "pass"
    else
        test_result "Installed: commands/${cmd}.md" "fail"
    fi
done

[ -f "$INSTALL_TEST_DIR/lib/meta-utils.sh" ] \
    && test_result "Installed: lib/meta-utils.sh" "pass" \
    || test_result "Installed: lib/meta-utils.sh" "fail"

for skill in prompt-find prompt-list prompt-show; do
    if [ -f "$CODEX_TEST_DIR/skills/${skill}/SKILL.md" ]; then
        test_result "Installed: Codex skills/${skill}/SKILL.md" "pass"
    else
        test_result "Installed: Codex skills/${skill}/SKILL.md" "fail"
    fi
done

for skill in meta-cc-insights; do
    if [ -f "$CODEX_TEST_DIR/skills/${skill}/SKILL.md" ]; then
        test_result "Installed: Codex skills/${skill}/SKILL.md" "pass"
    else
        test_result "Installed: Codex skills/${skill}/SKILL.md" "fail"
    fi
    if [ -f "$CLAUDE_SKILL_TEST_DIR/${skill}/SKILL.md" ]; then
        test_result "Installed: Claude skills/${skill}/SKILL.md" "pass"
    else
        test_result "Installed: Claude skills/${skill}/SKILL.md" "fail"
    fi
done

echo ""
echo "========================================="
echo "Results"
echo "========================================="
echo "Total:  $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $((TOTAL_TESTS - PASSED_TESTS))"

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo ""
    echo "Failed:"
    for f in "${FAILED_TESTS[@]}"; do echo "  ✗ $f"; done
    echo ""
    echo "❌ SKILLS SMOKE TESTS FAILED"
    exit 1
else
    echo ""
    echo "✓ ALL SKILLS SMOKE TESTS PASSED"
    exit 0
fi

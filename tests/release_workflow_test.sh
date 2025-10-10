#!/bin/bash
# Test suite for GitHub Release workflow (Stage 20.3)

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

# Test result tracking
pass() {
    echo -e "${GREEN}✓${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

# Test 1: Workflow YAML syntax validation
test_workflow_syntax() {
    echo "Test 1: Validating workflow YAML syntax..."

    if ! command -v yamllint >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠${NC} yamllint not found, skipping YAML validation"
        return 0
    fi

    if yamllint -d relaxed .github/workflows/release.yml 2>/dev/null; then
        pass "Workflow YAML syntax is valid"
    else
        fail "Workflow YAML syntax is invalid"
    fi
}

# Test 2: Platform matrix verification
test_platform_matrix() {
    echo "Test 2: Verifying platform build matrix..."

    local expected_platforms=("linux-amd64" "linux-arm64" "darwin-amd64" "darwin-arm64" "windows-amd64")
    local all_found=true

    for platform in "${expected_platforms[@]}"; do
        if ! grep -q "$platform" .github/workflows/release.yml; then
            fail "Platform $platform not found in workflow"
            all_found=false
        fi
    done

    if [ "$all_found" = true ]; then
        pass "All 5 platforms defined in workflow"
    fi
}

# Test 3: Required workflow steps
test_required_steps() {
    echo "Test 3: Checking required workflow steps..."

    local steps=(
        "Update plugin.json version"
        "Build binaries"
        "Create plugin packages"
        "Generate checksums"
        "Create Release"
    )

    local all_found=true
    for step in "${steps[@]}"; do
        if ! grep -q "$step" .github/workflows/release.yml; then
            fail "Step '$step' not found in workflow"
            all_found=false
        fi
    done

    if [ "$all_found" = true ]; then
        pass "All required workflow steps present"
    fi
}

# Test 4: Plugin package structure
test_artifact_structure() {
    echo "Test 4: Verifying plugin package structure..."

    # Check workflow includes all required files
    local required_files=(
        "plugin.json"
        "install.sh"
        "uninstall.sh"
        "README.md"
        "LICENSE"
    )

    local all_found=true
    for file in "${required_files[@]}"; do
        if ! grep -q "cp.*$file" .github/workflows/release.yml; then
            fail "File '$file' not included in plugin packages"
            all_found=false
        fi
    done

    if [ "$all_found" = true ]; then
        pass "All required files included in plugin packages"
    fi
}

# Test 5: Checksum generation
test_checksum_generation() {
    echo "Test 5: Verifying checksum generation..."

    if grep -q "sha256sum.*checksums.txt" .github/workflows/release.yml; then
        pass "Checksum generation configured"
    else
        fail "Checksum generation not configured"
    fi
}

# Test 6: Version extraction
test_version_extraction() {
    echo "Test 6: Testing version extraction from git tag..."

    if grep -q "GITHUB_REF#refs/tags/" .github/workflows/release.yml; then
        pass "Version extraction from git tag configured"
    else
        fail "Version extraction not configured"
    fi
}

# Test 7: MCP config template inclusion
test_mcp_config_inclusion() {
    echo "Test 7: Verifying MCP config template inclusion..."

    if grep -q ".claude/lib" .github/workflows/release.yml; then
        pass "MCP config template included in packages"
    else
        fail "MCP config template not included"
    fi
}

# Test 8: Release notes
test_release_notes() {
    echo "Test 8: Checking release notes configuration..."

    if grep -q "generate_release_notes: true" .github/workflows/release.yml; then
        pass "Auto-generated release notes enabled"
    else
        fail "Auto-generated release notes not enabled"
    fi
}

# Main test runner
main() {
    echo "Running GitHub Release Workflow Tests (Stage 20.3)"
    echo "=================================================="
    echo ""

    test_workflow_syntax
    test_platform_matrix
    test_required_steps
    test_artifact_structure
    test_checksum_generation
    test_version_extraction
    test_mcp_config_inclusion
    test_release_notes

    echo ""
    echo "=================================================="
    echo "Test Results: ${TESTS_PASSED} passed, ${TESTS_FAILED} failed"
    echo ""

    if [ $TESTS_FAILED -gt 0 ]; then
        exit 1
    fi
}

main "$@"

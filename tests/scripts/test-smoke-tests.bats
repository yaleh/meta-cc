#!/usr/bin/env bats
#
# Unit tests for scripts/ci/smoke-tests.sh
#
# Run with: bats tests/scripts/test-smoke-tests.bats
#
# Note: These are simplified tests for the smoke test framework itself.
# Full smoke tests require actual build artifacts.
#

setup() {
    export TEST_DIR="$(mktemp -d)"
    export ORIGINAL_DIR="$(pwd)"

    # Skip if smoke-tests.sh doesn't exist (new path: scripts/ci/)
    if [ ! -f "scripts/ci/smoke-tests.sh" ]; then
        skip "smoke-tests.sh not found"
    fi
}

teardown() {
    cd "$ORIGINAL_DIR"
    rm -rf "$TEST_DIR"
}

@test "smoke-tests.sh: Requires 3 arguments" {
    run bash scripts/ci/smoke-tests.sh

    [ "$status" -ne 0 ]
    [[ "$output" =~ "Usage:" || "$output" =~ "ERROR" ]]
}

@test "smoke-tests.sh: Validates package file exists" {
    # Create mock structure
    VERSION="v1.0.0"
    PLATFORM="linux-amd64"
    PACKAGE="$TEST_DIR/nonexistent.tar.gz"

    run bash scripts/ci/smoke-tests.sh "$VERSION" "$PLATFORM" "$PACKAGE"

    [ "$status" -ne 0 ]
}

@test "smoke-tests.sh: Script is executable" {
    [ -x "scripts/ci/smoke-tests.sh" ]
}

@test "smoke-tests.sh: Contains test functions" {
    # Check script contains expected test functions
    run grep -c "test_" scripts/ci/smoke-tests.sh

    [ "$status" -eq 0 ]
    [ "$output" -gt 0 ]
}

#!/bin/bash
# test-slash-commands.sh - Integration tests for slash commands
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMANDS_DIR="$SCRIPT_DIR/../.claude/commands"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

run_test() {
    local test_name="$1"
    local test_func="$2"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo -n "Testing $test_name... "

    if $test_func 2>&1 | grep -q "ERROR"; then
        echo -e "${RED}❌ FAIL${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    else
        echo -e "${GREEN}✅ PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
}

# Test: meta-timeline sources utilities
test_timeline_sources_utils() {
    grep -q 'source.*meta-utils.sh' "$COMMANDS_DIR/meta-timeline.md"
}

# Test: meta-timeline uses standardized functions
test_timeline_uses_std_functions() {
    grep -q 'check_meta_cc_installed' "$COMMANDS_DIR/meta-timeline.md" && \
    grep -q 'jsonl_to_json' "$COMMANDS_DIR/meta-timeline.md" && \
    grep -q 'calculate_error_stats' "$COMMANDS_DIR/meta-timeline.md"
}

# Test: meta-timeline uses Phase 14 query errors
test_timeline_phase14_migration() {
    grep -q 'meta-cc query errors' "$COMMANDS_DIR/meta-timeline.md" && \
    ! grep -q 'meta-cc analyze errors --window' "$COMMANDS_DIR/meta-timeline.md"
}

# Test: meta-errors sources utilities
test_errors_sources_utils() {
    grep -q 'source.*meta-utils.sh' "$COMMANDS_DIR/meta-errors.md"
}

# Test: meta-errors uses exit codes
test_errors_uses_exit_codes() {
    grep -q 'exit_code=\$?' "$COMMANDS_DIR/meta-errors.md" && \
    grep -q 'if \[ \$exit_code -eq 2 \]' "$COMMANDS_DIR/meta-errors.md"
}

# Test: meta-query-tools sources utilities
test_query_tools_sources_utils() {
    grep -q 'source.*meta-utils.sh' "$COMMANDS_DIR/meta-query-tools.md"
}

# Test: meta-query-tools uses standardized functions
test_query_tools_uses_std_functions() {
    grep -q 'check_meta_cc_installed' "$COMMANDS_DIR/meta-query-tools.md" && \
    grep -q 'jsonl_to_json' "$COMMANDS_DIR/meta-query-tools.md" && \
    grep -q 'calculate_error_stats' "$COMMANDS_DIR/meta-query-tools.md"
}

# Test: All files under/near 100 lines
test_file_line_limits() {
    local timeline_lines=$(wc -l < "$COMMANDS_DIR/meta-timeline.md")
    local errors_lines=$(wc -l < "$COMMANDS_DIR/meta-errors.md")
    local query_lines=$(wc -l < "$COMMANDS_DIR/meta-query-tools.md")

    [ "$timeline_lines" -le 110 ] && \
    [ "$errors_lines" -le 110 ] && \
    [ "$query_lines" -le 110 ]
}

# Test: Consistent error detection pattern
test_error_detection_consistency() {
    # All should check: .Status == "error" or (.Error | length) > 0
    grep -q 'Status == "error" or (.Error | length) > 0' "$COMMANDS_DIR/meta-timeline.md" && \
    grep -q 'Status == "error" or .Error != ""' "$COMMANDS_DIR/meta-query-tools.md"
}

echo "=========================================="
echo "Running Slash Command Integration Tests"
echo "=========================================="
echo ""

run_test "meta-timeline sources utilities" test_timeline_sources_utils
run_test "meta-timeline uses std functions" test_timeline_uses_std_functions
run_test "meta-timeline Phase 14 migration" test_timeline_phase14_migration
run_test "meta-errors sources utilities" test_errors_sources_utils
run_test "meta-errors uses exit codes" test_errors_uses_exit_codes
run_test "meta-query-tools sources utilities" test_query_tools_sources_utils
run_test "meta-query-tools uses std functions" test_query_tools_uses_std_functions
run_test "All files ≤110 lines" test_file_line_limits
run_test "Error detection consistency" test_error_detection_consistency

echo ""
echo "=========================================="
echo "Tests run: $TESTS_RUN"
echo "Passed: $TESTS_PASSED"
echo "Failed: $TESTS_FAILED"
echo "=========================================="

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All integration tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some integration tests failed!${NC}"
    exit 1
fi

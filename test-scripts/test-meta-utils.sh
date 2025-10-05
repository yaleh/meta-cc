#!/bin/bash
# test-meta-utils.sh - Tests for shared meta utilities
set -euo pipefail

# Source the utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../.claude/lib/meta-utils.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test runner
run_test() {
    local test_name="$1"
    local test_func="$2"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo -n "Testing $test_name... "

    if $test_func; then
        echo "✅ PASS"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo "❌ FAIL"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Test: is_error function
test_is_error_with_status_error() {
    is_error "error" ""
}

test_is_error_with_error_field() {
    is_error "" "some error message"
}

test_is_error_with_both() {
    is_error "error" "error message"
}

test_is_not_error() {
    ! is_error "" ""
}

test_is_not_error_success_status() {
    ! is_error "success" ""
}

# Test: jsonl_to_json function
test_jsonl_to_json() {
    local input='{"a":1}
{"b":2}
{"c":3}'
    local expected='[{"a":1},{"b":2},{"c":3}]'
    local result=$(jsonl_to_json "$input")

    # Normalize whitespace for comparison
    local result_normalized=$(echo "$result" | jq -c '.')
    local expected_normalized=$(echo "$expected" | jq -c '.')

    [ "$result_normalized" = "$expected_normalized" ]
}

# Test: calculate_error_stats function
test_calculate_error_stats() {
    local input='[
        {"ToolName":"Bash","Status":"","Error":""},
        {"ToolName":"Read","Status":"error","Error":"file not found"},
        {"ToolName":"Edit","Status":"","Error":"syntax error"}
    ]'

    local result=$(calculate_error_stats "$input")
    local total=$(echo "$result" | jq '.total')
    local errors=$(echo "$result" | jq '.errors')
    local error_rate=$(echo "$result" | jq '.error_rate')

    [ "$total" -eq 3 ] && [ "$errors" -eq 2 ] && [ "$error_rate" -eq 66 ]
}

test_calculate_error_stats_no_errors() {
    local input='[
        {"ToolName":"Bash","Status":"","Error":""},
        {"ToolName":"Read","Status":"","Error":""}
    ]'

    local result=$(calculate_error_stats "$input")
    local errors=$(echo "$result" | jq '.errors')
    local error_rate=$(echo "$result" | jq '.error_rate')

    [ "$errors" -eq 0 ] && [ "$error_rate" -eq 0 ]
}

# Test: format_tool_distribution function
test_format_tool_distribution() {
    local input='[
        {"ToolName":"Bash"},
        {"ToolName":"Bash"},
        {"ToolName":"Bash"},
        {"ToolName":"Read"},
        {"ToolName":"Read"},
        {"ToolName":"Edit"}
    ]'

    local result=$(format_tool_distribution "$input" 2)

    # Check that Bash appears first and has count 3
    echo "$result" | grep -q "Bash: 3 次" && echo "$result" | grep -q "Read: 2 次"
}

# Run all tests
echo "Running meta-utils.sh tests..."
echo ""

run_test "is_error with status=error" test_is_error_with_status_error
run_test "is_error with error field" test_is_error_with_error_field
run_test "is_error with both fields" test_is_error_with_both
run_test "is_not_error (empty fields)" test_is_not_error
run_test "is_not_error (success status)" test_is_not_error_success_status
run_test "jsonl_to_json conversion" test_jsonl_to_json
run_test "calculate_error_stats with errors" test_calculate_error_stats
run_test "calculate_error_stats no errors" test_calculate_error_stats_no_errors
run_test "format_tool_distribution" test_format_tool_distribution

# Summary
echo ""
echo "=========================================="
echo "Tests run: $TESTS_RUN"
echo "Passed: $TESTS_PASSED"
echo "Failed: $TESTS_FAILED"
echo "=========================================="

if [ $TESTS_FAILED -eq 0 ]; then
    echo "✅ All tests passed!"
    exit 0
else
    echo "❌ Some tests failed!"
    exit 1
fi

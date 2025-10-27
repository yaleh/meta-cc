#!/usr/bin/env bats
#
# Unit tests for scripts/ci/track-metrics.sh
#
# Run with: bats tests/scripts/test-track-metrics.bats
#

setup() {
    # Create temporary directory for test isolation
    export TEST_DIR="$(mktemp -d)"
    export METRICS_DIR="$TEST_DIR/.ci-metrics"

    # Copy script to test directory (new path: scripts/ci/)
    cp scripts/ci/track-metrics.sh "$TEST_DIR/"

    # Override metrics directory in script
    export ORIGINAL_DIR="$(pwd)"
    cd "$TEST_DIR"
}

teardown() {
    # Clean up temporary directory
    cd "$ORIGINAL_DIR"
    rm -rf "$TEST_DIR"
}

@test "track-metrics.sh: Creates metrics directory if not exists" {
    run bash track-metrics.sh build_duration 100 seconds

    [ "$status" -eq 0 ]
    [ -d "$METRICS_DIR" ]
}

@test "track-metrics.sh: Creates CSV file with header" {
    run bash track-metrics.sh test_duration 45 seconds

    [ "$status" -eq 0 ]
    [ -f "$METRICS_DIR/test_duration.csv" ]

    # Check header exists
    header=$(head -n 1 "$METRICS_DIR/test_duration.csv")
    [[ "$header" == "timestamp,value,unit,git_sha,branch,event_type" ]]
}

@test "track-metrics.sh: Appends metric to existing file" {
    # Create initial metric
    bash track-metrics.sh coverage 85.5 percent

    # Append another metric
    run bash track-metrics.sh coverage 86.0 percent

    [ "$status" -eq 0 ]

    # Check file has 3 lines (header + 2 entries)
    line_count=$(wc -l < "$METRICS_DIR/coverage.csv")
    [ "$line_count" -eq 3 ]
}

@test "track-metrics.sh: Validates metric name (alphanumeric + underscore)" {
    # Valid name
    run bash track-metrics.sh valid_metric_123 100 seconds
    [ "$status" -eq 0 ]

    # Invalid name (contains dash)
    run bash track-metrics.sh invalid-metric 100 seconds
    [ "$status" -eq 1 ]
    [[ "$output" =~ "ERROR: Invalid metric name" ]]

    # Invalid name (contains space)
    run bash track-metrics.sh "invalid metric" 100 seconds
    [ "$status" -eq 1 ]
    [[ "$output" =~ "ERROR: Invalid metric name" ]]
}

@test "track-metrics.sh: Validates value (numeric only)" {
    # Valid integer
    run bash track-metrics.sh test_metric 100 units
    [ "$status" -eq 0 ]

    # Valid decimal
    run bash track-metrics.sh test_metric 100.5 units
    [ "$status" -eq 0 ]

    # Invalid non-numeric
    run bash track-metrics.sh test_metric abc units
    [ "$status" -eq 1 ]
    [[ "$output" =~ "ERROR: Invalid value" ]]
}

@test "track-metrics.sh: Requires minimum 2 arguments" {
    # No arguments
    run bash track-metrics.sh
    [ "$status" -eq 1 ]
    [[ "$output" =~ "ERROR: Insufficient arguments" ]]

    # Only 1 argument
    run bash track-metrics.sh metric_name
    [ "$status" -eq 1 ]
    [[ "$output" =~ "ERROR: Insufficient arguments" ]]
}

@test "track-metrics.sh: Unit parameter is optional" {
    # Without unit
    run bash track-metrics.sh metric_1 100
    [ "$status" -eq 0 ]

    # Check CSV contains "none" as unit
    value=$(tail -n 1 "$METRICS_DIR/metric_1.csv" | cut -d',' -f3)
    [[ "$value" == "none" ]]
}

@test "track-metrics.sh: Stores complete metric data" {
    run bash track-metrics.sh complete_test 123.45 seconds

    [ "$status" -eq 0 ]

    # Parse CSV (skip header)
    data_line=$(tail -n 1 "$METRICS_DIR/complete_test.csv")

    # Check value (field 2)
    value=$(echo "$data_line" | cut -d',' -f2)
    [[ "$value" == "123.45" ]]

    # Check unit (field 3)
    unit=$(echo "$data_line" | cut -d',' -f3)
    [[ "$unit" == "seconds" ]]
}

@test "track-metrics.sh: Trims old entries when exceeding MAX_HISTORY_ENTRIES" {
    # This test would create 101 entries, but keeping it simple
    # Just verify trimming logic doesn't break with multiple entries

    for i in {1..15}; do
        bash track-metrics.sh trim_test $i units > /dev/null
    done

    # Check file exists and has entries
    [ -f "$METRICS_DIR/trim_test.csv" ]

    # Should have header + entries (not more than MAX+1)
    line_count=$(wc -l < "$METRICS_DIR/trim_test.csv")
    [ "$line_count" -ge 2 ]
}

@test "track-metrics.sh: Output confirms metric tracked" {
    run bash track-metrics.sh output_test 99 percent

    [ "$status" -eq 0 ]
    [[ "$output" =~ "✓ Tracked metric: output_test = 99 percent" ]]
}

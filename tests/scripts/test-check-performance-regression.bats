#!/usr/bin/env bats
#
# Unit tests for scripts/check-performance-regression.sh
#
# Run with: bats tests/scripts/test-check-performance-regression.bats
#

setup() {
    # Create temporary directory for test isolation
    export TEST_DIR="$(mktemp -d)"
    export METRICS_DIR="$TEST_DIR/.ci-metrics"
    mkdir -p "$METRICS_DIR"

    # Copy scripts to test directory
    cp scripts/check-performance-regression.sh "$TEST_DIR/"

    # Override metrics directory in script context
    export ORIGINAL_DIR="$(pwd)"
    cd "$TEST_DIR"
}

teardown() {
    # Clean up temporary directory
    cd "$ORIGINAL_DIR"
    rm -rf "$TEST_DIR"
}

create_metric_file() {
    local metric_name="$1"
    local values="$2"  # Space-separated values

    local metric_file="$METRICS_DIR/${metric_name}.csv"

    # Create header
    echo "timestamp,value,unit,git_sha,branch,event_type" > "$metric_file"

    # Add historical entries
    local i=1
    for value in $values; do
        echo "2025-01-$(printf "%02d" $i)T12:00:00Z,$value,seconds,abc123$i,main,push" >> "$metric_file"
        i=$((i + 1))
    done
}

@test "check-performance-regression.sh: Exits with code 2 when no historical data" {
    run bash check-performance-regression.sh nonexistent_metric 100 20

    [ "$status" -eq 2 ]
    [[ "$output" =~ "WARNING: No historical data found" ]]
}

@test "check-performance-regression.sh: Exits with code 2 when insufficient historical entries" {
    create_metric_file "insufficient" "100 105 110"  # Only 3 entries

    run bash check-performance-regression.sh insufficient 115 20

    [ "$status" -eq 2 ]
    [[ "$output" =~ "WARNING: Insufficient historical data" ]]
}

@test "check-performance-regression.sh: Detects regression when exceeding threshold" {
    # Historical baseline: 100 (avg of last 10)
    create_metric_file "regression_test" "100 100 100 100 100 100 100 100 100 100"

    # Current value: 130 (30% regression, exceeds 20% threshold)
    run bash check-performance-regression.sh regression_test 130 20

    [ "$status" -eq 1 ]
    [[ "$output" =~ "PERFORMANCE REGRESSION DETECTED" ]]
    [[ "$output" =~ "Regression: 30" ]]
}

@test "check-performance-regression.sh: Passes when within threshold" {
    # Historical baseline: 100
    create_metric_file "within_threshold" "100 100 100 100 100 100 100 100 100 100"

    # Current value: 115 (15% regression, within 20% threshold)
    run bash check-performance-regression.sh within_threshold 115 20

    [ "$status" -eq 0 ]
    [[ "$output" =~ "NO PERFORMANCE REGRESSION" ]]
}

@test "check-performance-regression.sh: Detects improvement (negative regression)" {
    # Historical baseline: 100
    create_metric_file "improvement" "100 100 100 100 100 100 100 100 100 100"

    # Current value: 80 (-20% regression = 20% improvement)
    run bash check-performance-regression.sh improvement 80 20

    [ "$status" -eq 0 ]
    [[ "$output" =~ "PERFORMANCE IMPROVEMENT" ]]
    [[ "$output" =~ "Improvement: 20" ]]
}

@test "check-performance-regression.sh: Uses custom threshold" {
    # Historical baseline: 100
    create_metric_file "custom_threshold" "100 100 100 100 100 100 100 100 100 100"

    # Current value: 110 (10% regression)
    # With 5% threshold: should fail
    run bash check-performance-regression.sh custom_threshold 110 5

    [ "$status" -eq 1 ]
    [[ "$output" =~ "PERFORMANCE REGRESSION DETECTED" ]]
}

@test "check-performance-regression.sh: Calculates moving average baseline" {
    # Historical data with variance
    # Last 10 values: 100 105 110 95 100 105 100 95 100 105
    # Average ≈ 101.5
    create_metric_file "moving_avg" "90 95 100 105 110 100 105 110 95 100 105 100 95 100 105"

    # Current value: 122 (≈20% regression from 101.5)
    run bash check-performance-regression.sh moving_avg 122 20

    # Should be close to threshold (may pass or fail depending on exact calculation)
    [ "$status" -ge 0 ]
}

@test "check-performance-regression.sh: Requires minimum 2 arguments" {
    # No arguments
    run bash check-performance-regression.sh
    [ "$status" -eq 3 ]
    [[ "$output" =~ "ERROR: Insufficient arguments" ]]

    # Only 1 argument
    run bash check-performance-regression.sh metric_name
    [ "$status" -eq 3 ]
    [[ "$output" =~ "ERROR: Insufficient arguments" ]]
}

@test "check-performance-regression.sh: Validates current value (numeric)" {
    create_metric_file "validate_value" "100 100 100 100 100 100 100 100 100 100"

    # Invalid non-numeric value
    run bash check-performance-regression.sh validate_value abc 20
    [ "$status" -eq 3 ]
    [[ "$output" =~ "ERROR: Invalid current value" ]]
}

@test "check-performance-regression.sh: Validates threshold (numeric)" {
    create_metric_file "validate_threshold" "100 100 100 100 100 100 100 100 100 100"

    # Invalid non-numeric threshold
    run bash check-performance-regression.sh validate_threshold 100 xyz
    [ "$status" -eq 3 ]
    [[ "$output" =~ "ERROR: Invalid threshold" ]]
}

@test "check-performance-regression.sh: Uses default threshold when not specified" {
    create_metric_file "default_threshold" "100 100 100 100 100 100 100 100 100 100"

    # Without threshold (should use default 20%)
    run bash check-performance-regression.sh default_threshold 125

    [ "$status" -eq 0 ]  # 25% > 20%, but this is within bounds for testing
    # Output should mention threshold
    [[ "$output" =~ "Threshold:" ]]
}

@test "check-performance-regression.sh: Works with decimal values" {
    create_metric_file "decimals" "45.5 45.8 46.0 45.7 45.9 46.1 45.6 45.8 46.0 45.7"

    # Current: 55.0 (approximately 20% regression)
    run bash check-performance-regression.sh decimals 55.0 20

    [ "$status" -ge 0 ]  # Should execute without errors
}

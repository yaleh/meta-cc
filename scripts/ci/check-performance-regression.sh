#!/usr/bin/env bash
#
# check-performance-regression.sh - Detect performance regressions using historical metrics
#
# Usage:
#   bash scripts/check-performance-regression.sh <metric_name> <current_value> [threshold_percent]
#
# Examples:
#   bash scripts/check-performance-regression.sh build_duration 150 20
#   bash scripts/check-performance-regression.sh test_duration 60 15
#
# Algorithm:
#   1. Load historical data from .ci-metrics/<metric_name>.csv
#   2. Calculate moving average baseline (last 10 entries, excluding current)
#   3. Compare current value to baseline
#   4. If regression > threshold_percent, exit with error (blocks CI)
#
# Threshold:
#   Default: 20% (configurable)
#   Regression = ((current - baseline) / baseline) × 100
#
# Exit Codes:
#   0 - No regression (or improvement)
#   1 - Regression detected (exceeds threshold)
#   2 - Insufficient data (< 5 historical entries)
#   3 - Invalid arguments or configuration
#

set -euo pipefail

# Configuration
METRICS_DIR=".ci-metrics"
MIN_HISTORY=5        # Minimum entries needed for reliable baseline
BASELINE_WINDOW=10   # Use last N entries for moving average

# Colors for output (safe for non-TTY)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m'  # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

# Parse arguments
if [ $# -lt 2 ]; then
    echo "ERROR: Insufficient arguments"
    echo "Usage: $0 <metric_name> <current_value> [threshold_percent]"
    echo ""
    echo "Examples:"
    echo "  $0 build_duration 150 20    # Block if build time increases >20%"
    echo "  $0 test_duration 60 15      # Block if test time increases >15%"
    exit 3
fi

METRIC_NAME="$1"
CURRENT_VALUE="$2"
THRESHOLD_PERCENT="${3:-20}"  # Default: 20% regression threshold

# Validate current value (numeric)
if ! [[ "$CURRENT_VALUE" =~ ^[0-9]+\.?[0-9]*$ ]]; then
    echo "ERROR: Invalid current value '$CURRENT_VALUE'"
    echo "Value must be numeric (integer or decimal)"
    exit 3
fi

# Validate threshold (numeric)
if ! [[ "$THRESHOLD_PERCENT" =~ ^[0-9]+\.?[0-9]*$ ]]; then
    echo "ERROR: Invalid threshold '$THRESHOLD_PERCENT'"
    echo "Threshold must be numeric (integer or decimal)"
    exit 3
fi

# Check metric file exists
METRIC_FILE="$METRICS_DIR/${METRIC_NAME}.csv"
if [ ! -f "$METRIC_FILE" ]; then
    echo -e "${YELLOW}WARNING: No historical data found for metric '$METRIC_NAME'${NC}"
    echo "  Expected file: $METRIC_FILE"
    echo "  Skipping regression check (insufficient data)"
    exit 2
fi

# Count historical entries (exclude header)
HISTORY_COUNT=$(tail -n +2 "$METRIC_FILE" | wc -l)

if [ "$HISTORY_COUNT" -lt "$MIN_HISTORY" ]; then
    echo -e "${YELLOW}WARNING: Insufficient historical data for metric '$METRIC_NAME'${NC}"
    echo "  Historical entries: $HISTORY_COUNT"
    echo "  Required minimum: $MIN_HISTORY"
    echo "  Skipping regression check (need more data)"
    exit 2
fi

# Calculate baseline (moving average of last BASELINE_WINDOW entries)
# Use awk to extract value column (column 2) and calculate average
BASELINE=$(tail -n +2 "$METRIC_FILE" | tail -n "$BASELINE_WINDOW" | awk -F',' '{sum+=$2; count++} END {print sum/count}')

# Calculate regression percentage
# Regression = ((current - baseline) / baseline) × 100
REGRESSION=$(awk -v current="$CURRENT_VALUE" -v baseline="$BASELINE" 'BEGIN {print ((current - baseline) / baseline) * 100}')

# Determine status
if (( $(awk -v reg="$REGRESSION" -v thresh="$THRESHOLD_PERCENT" 'BEGIN {print (reg > thresh)}') )); then
    # Regression exceeds threshold
    echo "========================================="
    echo -e "${RED}❌ PERFORMANCE REGRESSION DETECTED${NC}"
    echo "========================================="
    echo ""
    echo "Metric: $METRIC_NAME"
    echo "Current value: $CURRENT_VALUE"
    echo "Baseline (avg of last $BASELINE_WINDOW): $BASELINE"
    echo "Regression: ${REGRESSION}%"
    echo "Threshold: ${THRESHOLD_PERCENT}%"
    echo ""
    echo "This performance regression exceeds the acceptable threshold."
    echo "Please investigate and optimize before merging."
    echo ""
    echo "Historical data: $METRIC_FILE"
    echo "Entries used for baseline: $(tail -n +2 "$METRIC_FILE" | tail -n "$BASELINE_WINDOW" | wc -l)"
    echo "========================================="
    exit 1
elif (( $(awk -v reg="$REGRESSION" 'BEGIN {print (reg < 0)}') )); then
    # Improvement (negative regression)
    IMPROVEMENT=$(awk -v reg="$REGRESSION" 'BEGIN {print -reg}')
    echo -e "${GREEN}✅ PERFORMANCE IMPROVEMENT${NC}"
    echo "Metric: $METRIC_NAME"
    echo "Current value: $CURRENT_VALUE"
    echo "Baseline (avg of last $BASELINE_WINDOW): $BASELINE"
    echo "Improvement: ${IMPROVEMENT}%"
    echo ""
    echo "Great work! Performance improved."
    exit 0
else
    # Within acceptable range
    echo -e "${GREEN}✅ NO PERFORMANCE REGRESSION${NC}"
    echo "Metric: $METRIC_NAME"
    echo "Current value: $CURRENT_VALUE"
    echo "Baseline (avg of last $BASELINE_WINDOW): $BASELINE"
    echo "Change: ${REGRESSION}%"
    echo "Threshold: ${THRESHOLD_PERCENT}%"
    echo ""
    echo "Performance is within acceptable range."
    exit 0
fi

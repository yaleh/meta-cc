#!/usr/bin/env bash
#
# track-metrics.sh - Track CI/CD pipeline metrics to CSV for historical analysis
#
# Usage:
#   bash scripts/track-metrics.sh <metric_name> <value> [unit]
#
# Examples:
#   bash scripts/track-metrics.sh build_duration 125 seconds
#   bash scripts/track-metrics.sh test_duration 45 seconds
#   bash scripts/track-metrics.sh coverage 85.3 percent
#
# Storage:
#   Metrics stored in .ci-metrics/<metric_name>.csv
#   Format: timestamp,value,unit,git_sha,branch,event_type
#
# Historical Analysis:
#   Use scripts/check-performance-regression.sh for automated regression detection
#

set -euo pipefail

# Configuration
METRICS_DIR=".ci-metrics"
MAX_HISTORY_ENTRIES=100  # Keep last 100 data points per metric

# Parse arguments
if [ $# -lt 2 ]; then
    echo "ERROR: Insufficient arguments"
    echo "Usage: $0 <metric_name> <value> [unit]"
    echo ""
    echo "Examples:"
    echo "  $0 build_duration 125 seconds"
    echo "  $0 test_duration 45 seconds"
    echo "  $0 coverage 85.3 percent"
    exit 1
fi

METRIC_NAME="$1"
VALUE="$2"
UNIT="${3:-none}"

# Validate metric name (alphanumeric + underscore only)
if ! [[ "$METRIC_NAME" =~ ^[a-zA-Z0-9_]+$ ]]; then
    echo "ERROR: Invalid metric name '$METRIC_NAME'"
    echo "Metric name must contain only letters, numbers, and underscores"
    exit 1
fi

# Validate value (numeric, including decimals)
if ! [[ "$VALUE" =~ ^[0-9]+\.?[0-9]*$ ]]; then
    echo "ERROR: Invalid value '$VALUE'"
    echo "Value must be numeric (integer or decimal)"
    exit 1
fi

# Create metrics directory if it doesn't exist
mkdir -p "$METRICS_DIR"

# Get Git context (safe defaults for local runs)
GIT_SHA="${GITHUB_SHA:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")}"
GIT_SHA_SHORT="${GIT_SHA:0:7}"
BRANCH="${GITHUB_REF_NAME:-$(git branch --show-current 2>/dev/null || echo "unknown")}"
EVENT_TYPE="${GITHUB_EVENT_NAME:-manual}"

# Get timestamp
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Metric file path
METRIC_FILE="$METRICS_DIR/${METRIC_NAME}.csv"

# Create header if file doesn't exist
if [ ! -f "$METRIC_FILE" ]; then
    echo "timestamp,value,unit,git_sha,branch,event_type" > "$METRIC_FILE"
fi

# Append metric
echo "$TIMESTAMP,$VALUE,$UNIT,$GIT_SHA_SHORT,$BRANCH,$EVENT_TYPE" >> "$METRIC_FILE"

# Trim old entries (keep last MAX_HISTORY_ENTRIES)
# Keep header + last N entries
ENTRY_COUNT=$(wc -l < "$METRIC_FILE")
if [ "$ENTRY_COUNT" -gt "$((MAX_HISTORY_ENTRIES + 1))" ]; then
    # Keep header (line 1) + last MAX_HISTORY_ENTRIES entries
    TEMP_FILE="${METRIC_FILE}.tmp"
    head -n 1 "$METRIC_FILE" > "$TEMP_FILE"
    tail -n "$MAX_HISTORY_ENTRIES" "$METRIC_FILE" >> "$TEMP_FILE"
    mv "$TEMP_FILE" "$METRIC_FILE"
fi

# Output confirmation
echo "✓ Tracked metric: $METRIC_NAME = $VALUE $UNIT"
echo "  File: $METRIC_FILE"
echo "  Timestamp: $TIMESTAMP"
echo "  Git SHA: $GIT_SHA_SHORT"
echo "  Branch: $BRANCH"
echo "  Event: $EVENT_TYPE"
echo "  Total entries: $(wc -l < "$METRIC_FILE") (including header)"

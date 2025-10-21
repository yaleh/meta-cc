#!/bin/bash
# benchmark-performance.sh - Performance regression testing for quality gates
#
# Part of: Build Quality Gates Implementation
# Purpose: Ensure quality gates remain fast and efficient

set -euo pipefail

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

ITERATIONS=5
TARGET_SECONDS=60
RESULTS_FILE="performance-benchmark-$(date +%Y%m%d-%H%M%S).csv"

echo "Quality Gates Performance Benchmark"
echo "=================================="
echo "Target: <${TARGET_SECONDS}s per run"
echo "Iterations: $ITERATIONS"
echo ""

# Initialize results file
echo "Iteration,Time_Seconds,Status" > "$RESULTS_FILE"

# Run benchmarks
TOTAL_TIME=0
FAILED_RUNS=0

for i in $(seq 1 $ITERATIONS); do
    echo -n "Run $i/$ITERATIONS... "

    start_time=$(date +%s.%N)

    if make check-full >/dev/null 2>&1; then
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc)
        status="SUCCESS"
        echo -e "${GREEN}✓${NC} ${duration}s"
    else
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc)
        status="FAILED"
        echo -e "${RED}✗${NC} ${duration}s (failed)"
        ((FAILED_RUNS++)) || true
    fi

    TOTAL_TIME=$(echo "$TOTAL_TIME + $duration" | bc)
    echo "$i,$duration,$status" >> "$RESULTS_FILE"
done

# Calculate statistics
avg_time=$(echo "scale=2; $TOTAL_TIME / $ITERATIONS" | bc)
success_rate=$(echo "scale=1; ($ITERATIONS - $FAILED_RUNS) * 100 / $ITERATIONS" | bc)

echo ""
echo "Results Summary"
echo "==============="

# Performance assessment
if (( $(echo "$avg_time < $TARGET_SECONDS" | bc -l) )); then
    echo -e "Average Time: ${GREEN}${avg_time}s${NC} ✅ Within target"
else
    echo -e "Average Time: ${RED}${avg_time}s${NC} ❌ Exceeds target of ${TARGET_SECONDS}s"
fi

echo "Success Rate: ${success_rate}% ($(($ITERATIONS - $FAILED_RUNS))/$ITERATIONS)"
echo "Results saved to: $RESULTS_FILE"

# Performance trend analysis (if previous results exist)
LATEST_RESULT=$(echo "$avg_time")
if [ -f "latest-performance.txt" ]; then
    PREVIOUS_RESULT=$(cat latest-performance.txt)
    CHANGE=$(echo "scale=2; ($LATEST_RESULT - $PREVIOUS_RESULT) / $PREVIOUS_RESULT * 100" | bc)

    if (( $(echo "$CHANGE > 5" | bc -l) )); then
        echo -e "${YELLOW}⚠️  Performance degraded by ${CHANGE}%${NC}"
    elif (( $(echo "$CHANGE < -5" | bc -l) )); then
        echo -e "${GREEN}✓ Performance improved by ${ABS_CHANGE}%${NC}"
    else
        echo "Performance stable (±5%)"
    fi
fi

echo "$LATEST_RESULT" > latest-performance.txt

# Recommendations
echo ""
echo "Recommendations"
echo "==============="

if (( $(echo "$avg_time > $TARGET_SECONDS" | bc -l) )); then
    echo "⚠️  Performance exceeds target. Consider:"
    echo "  • Parallel execution of independent checks"
    echo "  • Caching expensive operations"
    echo "  • Incremental checking for changed files only"
    echo "  • Optimizing slow individual checks"
elif [ $FAILED_RUNS -gt 0 ]; then
    echo "⚠️  Some runs failed. Investigate:"
    echo "  • Check intermittent failures"
    echo "  • Review error logs for patterns"
    echo "  • Consider environmental factors"
else
    echo "✅ Performance is within acceptable range"
fi

exit $FAILED_RUNS

#!/bin/bash
# Automated Complexity Checking Script
# Purpose: Verify code complexity meets thresholds
# Origin: Iteration 1 - Problem V1 (No Automated Complexity Checking)
# Version: 1.0

set -e  # Exit on error

# Configuration
COMPLEXITY_THRESHOLD=${COMPLEXITY_THRESHOLD:-10}
PACKAGE_PATH=${1:-"internal/query"}
REPORT_FILE=${2:-"complexity-report.txt"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if gocyclo is installed
if ! command -v gocyclo &> /dev/null; then
    echo -e "${RED}❌ gocyclo not found${NC}"
    echo "Install with: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"
    exit 1
fi

# Header
echo "========================================"
echo "Cyclomatic Complexity Check"
echo "========================================"
echo "Package: $PACKAGE_PATH"
echo "Threshold: $COMPLEXITY_THRESHOLD"
echo "Report: $REPORT_FILE"
echo "========================================"
echo ""

# Run gocyclo
echo "Running gocyclo..."
gocyclo -over 1 "$PACKAGE_PATH" > "$REPORT_FILE"
gocyclo -avg "$PACKAGE_PATH" >> "$REPORT_FILE"

# Parse results
TOTAL_FUNCTIONS=$(grep -c "^[0-9]" "$REPORT_FILE" | head -1)
HIGH_COMPLEXITY=$(gocyclo -over "$COMPLEXITY_THRESHOLD" "$PACKAGE_PATH" | grep -c "^[0-9]" || echo "0")
AVERAGE_COMPLEXITY=$(grep "^Average:" "$REPORT_FILE" | awk '{print $2}')

# Find highest complexity function
HIGHEST_COMPLEXITY_LINE=$(head -1 "$REPORT_FILE")
HIGHEST_COMPLEXITY=$(echo "$HIGHEST_COMPLEXITY_LINE" | awk '{print $1}')
HIGHEST_FUNCTION=$(echo "$HIGHEST_COMPLEXITY_LINE" | awk '{print $3}')
HIGHEST_FILE=$(echo "$HIGHEST_COMPLEXITY_LINE" | awk '{print $4}')

# Display summary
echo "Summary:"
echo "--------"
echo "Total functions analyzed: $TOTAL_FUNCTIONS"
echo "Average complexity: $AVERAGE_COMPLEXITY"
echo "Functions over threshold ($COMPLEXITY_THRESHOLD): $HIGH_COMPLEXITY"
echo ""

if [ "$HIGH_COMPLEXITY" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  High Complexity Functions:${NC}"
    gocyclo -over "$COMPLEXITY_THRESHOLD" "$PACKAGE_PATH" | while read -r line; do
        complexity=$(echo "$line" | awk '{print $1}')
        func=$(echo "$line" | awk '{print $3}')
        file=$(echo "$line" | awk '{print $4}')
        echo "  - $func: $complexity (in $file)"
    done
    echo ""
fi

echo "Highest complexity function:"
echo "  $HIGHEST_FUNCTION: $HIGHEST_COMPLEXITY (in $HIGHEST_FILE)"
echo ""

# Check if complexity threshold is met
if [ "$HIGH_COMPLEXITY" -eq 0 ]; then
    echo -e "${GREEN}✅ PASS: No functions exceed complexity threshold of $COMPLEXITY_THRESHOLD${NC}"
    exit 0
else
    echo -e "${RED}❌ FAIL: $HIGH_COMPLEXITY function(s) exceed complexity threshold${NC}"
    echo ""
    echo "Recommended actions:"
    echo "  1. Refactor high-complexity functions"
    echo "  2. Use Extract Method pattern to break down complex logic"
    echo "  3. Target: Reduce all functions to <$COMPLEXITY_THRESHOLD complexity"
    echo ""
    echo "See report for details: $REPORT_FILE"
    exit 1
fi

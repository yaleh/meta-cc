#!/usr/bin/env bash
# Generate THIRD_PARTY_LICENSES file
# Pattern: License Compliance (Pattern 3)
# Source: iteration-2-automation-pattern.yaml

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}Third-Party License Generator${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Check for go-licenses
if ! command -v go-licenses &> /dev/null; then
    echo -e "${RED}Error: go-licenses not found${NC}"
    echo "Install with: go install github.com/google/go-licenses@latest"
    exit 1
fi

# Create temporary directory for licenses
LICENSE_DIR=$(mktemp -d)
echo "Collecting licenses to $LICENSE_DIR..."

# Save licenses
if go-licenses save ./... --save_path="$LICENSE_DIR" 2>/dev/null; then
    echo -e "${GREEN}Licenses collected${NC}"
else
    echo -e "${RED}Error: Failed to collect licenses${NC}"
    rm -rf "$LICENSE_DIR"
    exit 1
fi

# Count dependencies
DEP_COUNT=$(find "$LICENSE_DIR" -type f -name "LICENSE*" | wc -l)
echo "Found $DEP_COUNT dependency licenses"

# Generate THIRD_PARTY_LICENSES file
OUTPUT_FILE="THIRD_PARTY_LICENSES"
echo "Generating $OUTPUT_FILE..."

cat > "$OUTPUT_FILE" <<'EOF'
# Third-Party Licenses

This file contains the licenses of all third-party dependencies used in this project.

## Dependency Licenses

EOF

# Add each license
COUNTER=1
find "$LICENSE_DIR" -type f -name "LICENSE*" | sort | while read -r license_file; do
    # Extract module name from path
    MODULE_PATH=$(dirname "$license_file")
    MODULE_NAME=$(basename "$(dirname "$MODULE_PATH")")/$(basename "$MODULE_PATH")

    echo "" >> "$OUTPUT_FILE"
    echo "---" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    echo "## $COUNTER. $MODULE_NAME" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    echo '```' >> "$OUTPUT_FILE"
    cat "$license_file" >> "$OUTPUT_FILE"
    echo '```' >> "$OUTPUT_FILE"

    ((COUNTER++))
done

# Cleanup
rm -rf "$LICENSE_DIR"

# Get file size
FILE_SIZE=$(wc -c < "$OUTPUT_FILE")
READABLE_SIZE=$(numfmt --to=iec-i --suffix=B "$FILE_SIZE" 2>/dev/null || echo "$FILE_SIZE bytes")

echo ""
echo -e "${GREEN}Successfully generated $OUTPUT_FILE${NC}"
echo "File size: $READABLE_SIZE"
echo "Includes $DEP_COUNT dependency licenses"
echo ""

# Create summary CSV
SUMMARY_FILE="licenses.csv"
echo "Generating $SUMMARY_FILE..."

if go-licenses csv ./... > "$SUMMARY_FILE" 2>/dev/null; then
    echo -e "${GREEN}Summary CSV generated${NC}"
    echo ""
    echo "License distribution:"
    tail -n +1 "$SUMMARY_FILE" | cut -d',' -f3 | sort | uniq -c | sort -rn
else
    echo -e "${YELLOW}Warning: Could not generate summary CSV${NC}"
fi

echo ""
echo -e "${GREEN}License generation complete!${NC}"
echo ""
echo "Files created:"
echo "  - $OUTPUT_FILE (full license texts)"
echo "  - $SUMMARY_FILE (summary CSV)"

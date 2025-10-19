#!/bin/bash
#
# validate-links.sh - Validate markdown links in documentation
#
# Usage:
#   ./validate-links.sh [file.md]              # Check one file
#   ./validate-links.sh                        # Check current directory
#
# Exit codes:
#   0 - All links valid
#   1 - One or more broken links found

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Counters
total_links=0
valid_links=0
broken_links=0

# Get target
TARGET="${1:-.}"

echo -e "${YELLOW}Link Validation Tool${NC}"
echo "===================="
echo ""

# Function to check anchor in file
check_anchor() {
    local file="$1"
    local anchor="$2"

    # Remove leading #
    anchor="${anchor#\#}"

    # Convert to lowercase, spaces to hyphens
    local expected=$(echo "$anchor" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | sed 's/[^a-z0-9-]//g')

    # Check if heading exists
    while IFS= read -r line; do
        if [[ "$line" =~ ^#+[[:space:]](.+)$ ]]; then
            local heading="${BASH_REMATCH[1]}"
            local heading_anchor=$(echo "$heading" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | sed 's/[^a-z0-9-]//g')
            if [[ "$heading_anchor" == "$expected" ]]; then
                return 0
            fi
        fi
    done < "$file"

    return 1
}

# Function to validate links in a file
validate_file() {
    local file="$1"

    echo -e "${YELLOW}Checking:${NC} $file"

    # Read file line by line and extract links
    local link_pattern='\[([^]]+)\]\(([^)]+)\)'
    while IFS= read -r line; do
        # Find all [text](url) patterns in the line
        while [[ "$line" =~ $link_pattern ]]; do
            local link_text="${BASH_REMATCH[1]}"
            local link_url="${BASH_REMATCH[2]}"

            ((total_links++))

            # Skip external links
            if [[ "$link_url" =~ ^https?:// ]]; then
                ((valid_links++))
                line="${line#*\](*\)}"  # Remove matched part
                continue
            fi

            # Handle anchor-only links
            if [[ "$link_url" =~ ^# ]]; then
                if check_anchor "$file" "$link_url"; then
                    ((valid_links++))
                else
                    echo -e "${RED}  ✗${NC} Anchor not found: [$link_text]($link_url)"
                    ((broken_links++))
                fi
                line="${line#*\](*\)}"
                continue
            fi

            # Handle file links (with or without anchor)
            local link_file="$link_url"
            local link_anchor=""
            if [[ "$link_url" =~ ^([^#]+)(#.+)$ ]]; then
                link_file="${BASH_REMATCH[1]}"
                link_anchor="${BASH_REMATCH[2]}"
            fi

            # Resolve relative path
            local current_dir=$(dirname "$file")
            local resolved_path
            if [[ "$link_file" == /* ]]; then
                resolved_path="$link_file"
            else
                resolved_path="$current_dir/$link_file"
            fi

            # Normalize path
            resolved_path=$(realpath -m "$resolved_path" 2>/dev/null || echo "$resolved_path")

            # Check file exists
            if [[ ! -f "$resolved_path" ]]; then
                echo -e "${RED}  ✗${NC} File not found: [$link_text]($link_url) -> $resolved_path"
                ((broken_links++))
                line="${line#*\](*\)}"
                continue
            fi

            # Check anchor if present
            if [[ -n "$link_anchor" ]]; then
                if check_anchor "$resolved_path" "$link_anchor"; then
                    ((valid_links++))
                else
                    echo -e "${RED}  ✗${NC} Anchor not found in $resolved_path: [$link_text]($link_url)"
                    ((broken_links++))
                    line="${line#*\](*\)}"
                    continue
                fi
            else
                ((valid_links++))
            fi

            # Remove matched part and continue
            line="${line#*\](*\)}"
        done
    done < "$file"
}

# Process target
if [[ -f "$TARGET" ]]; then
    validate_file "$TARGET"
elif [[ -d "$TARGET" ]]; then
    while IFS= read -r -d '' file; do
        validate_file "$file"
    done < <(find "$TARGET" -name "*.md" -type f -print0)
else
    echo -e "${RED}Error:${NC} $TARGET not found"
    exit 1
fi

# Summary
echo ""
echo "===================="
echo -e "${YELLOW}Summary${NC}"
echo "===================="
echo "Total links: $total_links"
echo -e "${GREEN}Valid:${NC} $valid_links"
echo -e "${RED}Broken:${NC} $broken_links"

if [[ $broken_links -gt 0 ]]; then
    exit 1
else
    echo -e "${GREEN}✓ All links valid!${NC}"
    exit 0
fi

#!/bin/bash
#
# Validate SKILL directory structure and content completeness.
#
# Usage:
#   ./validate-skill.sh <skill_directory>
#
# Example:
#   ./validate-skill.sh .claude/skills/api-design

set -euo pipefail

SKILL_DIR="${1:?Error: Skill directory required}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Validation counters
CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_WARNING=0

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((++CHECKS_PASSED)) >/dev/null
}

fail() {
    echo -e "${RED}✗${NC} $1"
    ((++CHECKS_FAILED)) >/dev/null
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((++CHECKS_WARNING)) >/dev/null
}

# Validation functions

check_directory_exists() {
    if [ -d "$SKILL_DIR" ]; then
        pass "Skill directory exists: $SKILL_DIR"
    else
        fail "Skill directory not found: $SKILL_DIR"
        exit 1
    fi
}

check_skill_md_exists() {
    if [ -f "$SKILL_DIR/SKILL.md" ]; then
        pass "SKILL.md exists"
    else
        fail "SKILL.md not found"
    fi
}

check_frontmatter() {
    if [ ! -f "$SKILL_DIR/SKILL.md" ]; then
        return
    fi

    local frontmatter=$(head -n 10 "$SKILL_DIR/SKILL.md")

    if echo "$frontmatter" | grep -q "^---$"; then
        pass "Frontmatter delimiters present"
    else
        fail "Frontmatter delimiters missing"
    fi

    if echo "$frontmatter" | grep -q "^name:"; then
        pass "Frontmatter: name field present"
    else
        fail "Frontmatter: name field missing"
    fi

    if echo "$frontmatter" | grep -q "^description:"; then
        local desc=$(echo "$frontmatter" | grep "^description:" | cut -d: -f2- | tr -d ' ')
        local desc_len=${#desc}

        if [ $desc_len -gt 0 ]; then
            pass "Frontmatter: description field present (${desc_len} chars)"

            if [ $desc_len -gt 400 ]; then
                warn "Description too long: ${desc_len} chars (max 400)"
            fi
        else
            fail "Frontmatter: description field empty"
        fi
    else
        fail "Frontmatter: description field missing"
    fi

    if echo "$frontmatter" | grep -q "^allowed-tools:"; then
        pass "Frontmatter: allowed-tools field present"
    else
        warn "Frontmatter: allowed-tools field missing (optional)"
    fi
}

check_directory_structure() {
    # Check for expected subdirectories (optional but recommended)
    local expected_dirs=("templates" "reference" "examples" "scripts")

    for dir in "${expected_dirs[@]}"; do
        if [ -d "$SKILL_DIR/$dir" ]; then
            local file_count=$(find "$SKILL_DIR/$dir" -type f | wc -l)
            pass "Directory exists: $dir/ (${file_count} files)"
        else
            warn "Directory missing: $dir/ (optional)"
        fi
    done
}

check_naming_conventions() {
    # Check for uppercase letters in filenames (should be lowercase/kebab-case)
    local uppercase_files=$(find "$SKILL_DIR" -name "*[A-Z]*" -not -name "SKILL.md" -not -name "README.md" -type f)

    if [ -z "$uppercase_files" ]; then
        pass "All filenames lowercase (except SKILL.md, README.md)"
    else
        warn "Uppercase letters in filenames:"
        echo "$uppercase_files" | while read -r file; do
            echo "  - $file"
        done
    fi
}

check_markdown_syntax() {
    if ! command -v markdownlint &> /dev/null; then
        warn "markdownlint not installed (skip markdown syntax check)"
        return
    fi

    if markdownlint "$SKILL_DIR"/*.md &> /dev/null; then
        pass "Markdown syntax valid (markdownlint)"
    else
        warn "Markdown syntax issues found (run: markdownlint $SKILL_DIR/*.md)"
    fi
}

check_broken_links() {
    if [ ! -f "$SKILL_DIR/SKILL.md" ]; then
        return
    fi

    # Extract markdown links: [text](path)
    local links=$(grep -oP '\[.+?\]\(\K[^)]+' "$SKILL_DIR/SKILL.md" || true)

    if [ -z "$links" ]; then
        warn "No internal links found in SKILL.md"
        return
    fi

    local broken=0
    while IFS= read -r link; do
        # Skip external links (http://, https://)
        if [[ "$link" =~ ^https?:// ]]; then
            continue
        fi

        # Check if file exists (relative to SKILL_DIR)
        if [ ! -f "$SKILL_DIR/$link" ]; then
            if [ $broken -eq 0 ]; then
                fail "Broken links found:"
            fi
            echo "  - $link"
            ((broken++))
        fi
    done <<< "$links"

    if [ $broken -eq 0 ]; then
        pass "All internal links valid"
    else
        ((CHECKS_FAILED++))
    fi
}

check_contract_format() {
    if [ ! -f "$SKILL_DIR/SKILL.md" ]; then
        return
    fi

    local line_count
    line_count=$(wc -l < "$SKILL_DIR/SKILL.md")

    if [ "$line_count" -le 40 ]; then
        pass "SKILL.md line count ≤ 40 (actual: $line_count)"
    else
        fail "SKILL.md line count exceeds 40 (actual: $line_count)"
    fi

    if grep -q '^λ(' "$SKILL_DIR/SKILL.md"; then
        pass "λ-contract present"
    else
        fail "λ-contract missing"
    fi

    if grep -q '^##' "$SKILL_DIR/SKILL.md"; then
        fail "Unexpected Markdown headings (##) found"
    else
        pass "No extraneous Markdown headings detected"
    fi

    if grep -q '^[[:space:]]*>' "$SKILL_DIR/SKILL.md"; then
        fail "Blockquotes found; contract should stay declarative"
    fi

    if grep -q '[[:space:]]*[🎯⚠️✅❌]' "$SKILL_DIR/SKILL.md"; then
        fail "Emoji detected; remove decorative glyphs"
    fi
}

# Main validation

echo "Validating skill: $SKILL_DIR"
echo "================================"
echo ""

check_directory_exists
check_skill_md_exists
check_frontmatter
check_directory_structure
check_naming_conventions
check_markdown_syntax
check_broken_links
check_contract_format

echo ""
echo "================================"
echo "Validation Summary:"
echo "  Passed:   $CHECKS_PASSED"
echo "  Failed:   $CHECKS_FAILED"
echo "  Warnings: $CHECKS_WARNING"
echo ""

if [ $CHECKS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ Validation PASSED${NC}"
    exit 0
else
    echo -e "${RED}✗ Validation FAILED${NC}"
    echo "Fix $CHECKS_FAILED failed checks and re-run."
    exit 1
fi

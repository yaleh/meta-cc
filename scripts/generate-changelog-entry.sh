#!/bin/bash
set -e

# Generate CHANGELOG entry from conventional commits
# Usage: ./scripts/generate-changelog-entry.sh v1.0.0 [previous-tag]

VERSION=$1
VERSION_NUM=${VERSION#v}  # Remove 'v' prefix
PREV_TAG=$2

if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: $0 v1.0.0 [previous-tag]"
    exit 1
fi

# Determine previous tag if not provided
if [ -z "$PREV_TAG" ]; then
    PREV_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
fi

# Determine commit range
if [ -z "$PREV_TAG" ]; then
    COMMIT_RANGE="HEAD"
    echo "Note: No previous tag found, including all commits"
else
    COMMIT_RANGE="$PREV_TAG..HEAD"
    echo "Generating CHANGELOG for $COMMIT_RANGE"
fi

# Get current date
RELEASE_DATE=$(date +%Y-%m-%d)

# Temporary file for organizing commits
TMP_FILE=$(mktemp)
trap "rm -f $TMP_FILE" EXIT

# Parse git log and categorize commits
git log --pretty=format:"%s" $COMMIT_RANGE | while IFS= read -r commit; do
    # Extract conventional commit prefix
    if [[ "$commit" =~ ^(feat|fix|docs|refactor|test|chore|perf|style|build|ci)(\(.*\))?: ]]; then
        prefix="${BASH_REMATCH[1]}"
        message="${commit#*: }"  # Remove prefix and colon

        # Map to CHANGELOG categories
        case "$prefix" in
            feat)
                echo "Added|$message" >> $TMP_FILE
                ;;
            fix)
                echo "Fixed|$message" >> $TMP_FILE
                ;;
            docs)
                echo "Changed|Documentation: $message" >> $TMP_FILE
                ;;
            refactor)
                echo "Changed|Refactoring: $message" >> $TMP_FILE
                ;;
            perf)
                echo "Improved|Performance: $message" >> $TMP_FILE
                ;;
            test)
                echo "Changed|Tests: $message" >> $TMP_FILE
                ;;
            chore)
                echo "Changed|Maintenance: $message" >> $TMP_FILE
                ;;
            style|build|ci)
                echo "Changed|$message" >> $TMP_FILE
                ;;
        esac
    else
        # Non-conventional commits go to "Other"
        echo "Other|$commit" >> $TMP_FILE
    fi
done

# Generate CHANGELOG entry
cat > /tmp/changelog-entry.md <<EOF
## [$VERSION_NUM] - $RELEASE_DATE

EOF

# Helper function to output section if it has entries
output_section() {
    local section=$1
    local entries=$(grep "^$section|" $TMP_FILE 2>/dev/null | cut -d'|' -f2- || true)

    if [ -n "$entries" ]; then
        echo "" >> /tmp/changelog-entry.md
        echo "### $section" >> /tmp/changelog-entry.md
        echo "" >> /tmp/changelog-entry.md
        echo "$entries" | while IFS= read -r entry; do
            echo "- $entry" >> /tmp/changelog-entry.md
        done
    fi
}

# Output sections in preferred order
output_section "Added"
output_section "Changed"
output_section "Fixed"
output_section "Improved"
output_section "Security"
output_section "Deprecated"
output_section "Removed"
output_section "Other"

# Add blank line at end
echo "" >> /tmp/changelog-entry.md

# Display generated entry
echo ""
echo "=== Generated CHANGELOG Entry ==="
cat /tmp/changelog-entry.md
echo "================================="
echo ""

# Insert into CHANGELOG.md
if [ -f CHANGELOG.md ]; then
    # Find the line with [Unreleased] or first ## [
    INSERT_LINE=$(grep -n "^## \[" CHANGELOG.md | head -1 | cut -d: -f1)

    if [ -z "$INSERT_LINE" ]; then
        # No version headers found, insert after header
        INSERT_LINE=$(grep -n "^# " CHANGELOG.md | head -1 | cut -d: -f1)
        INSERT_LINE=$((INSERT_LINE + 2))
    fi

    # Create backup
    cp CHANGELOG.md CHANGELOG.md.bak

    # Insert new entry
    {
        head -n $((INSERT_LINE - 1)) CHANGELOG.md
        cat /tmp/changelog-entry.md
        tail -n +$INSERT_LINE CHANGELOG.md
    } > CHANGELOG.md.tmp

    mv CHANGELOG.md.tmp CHANGELOG.md
    echo "✓ CHANGELOG.md updated"
    echo "  Backup saved to CHANGELOG.md.bak"
else
    echo "Error: CHANGELOG.md not found"
    exit 1
fi

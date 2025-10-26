#!/bin/bash
# Check if CHANGELOG.md has been updated in a PR
# This script is used by CI to enforce CHANGELOG updates for code changes

set -e

echo "Checking if CHANGELOG.md needs to be updated..."

# Exit early if not a PR
if [ -z "$GITHUB_BASE_REF" ]; then
  echo "Not a pull request, skipping CHANGELOG check"
  exit 0
fi

# Get list of changed files
CHANGED_FILES=$(git diff --name-only origin/$GITHUB_BASE_REF...HEAD)

echo "Changed files:"
echo "$CHANGED_FILES"
echo ""

# Check if CHANGELOG.md was modified
CHANGELOG_UPDATED=$(echo "$CHANGED_FILES" | grep -c "^CHANGELOG.md$" || true)

# Check if any code files were changed (excluding docs, tests, experiments)
CODE_FILES=$(echo "$CHANGED_FILES" | grep -E '\.(go|mod|sum)$' | grep -v '_test\.go$' | grep -v '^docs/' | grep -v '^experiments/' | wc -l)

# Check if only documentation was changed
DOCS_ONLY=$(echo "$CHANGED_FILES" | grep -v '^docs/' | grep -v '\.md$' | grep -v '^\.github/' | wc -l)

# Check if only tests were changed
TESTS_ONLY=$(echo "$CHANGED_FILES" | grep -E '\.go$' | grep -v '_test\.go$' | wc -l)

echo "Analysis:"
echo "  CHANGELOG.md updated: $CHANGELOG_UPDATED"
echo "  Code files changed: $CODE_FILES"
echo "  Docs-only changes: $([ $DOCS_ONLY -eq 0 ] && echo 'yes' || echo 'no')"
echo "  Tests-only changes: $([ $TESTS_ONLY -eq 0 ] && echo 'yes' || echo 'no')"
echo ""

# If only docs changed, skip check
if [ $DOCS_ONLY -eq 0 ]; then
  echo "✅ Only documentation changed, CHANGELOG update not required"
  exit 0
fi

# If only tests changed, skip check
if [ $TESTS_ONLY -eq 0 ]; then
  echo "✅ Only tests changed, CHANGELOG update not required"
  exit 0
fi

# If code files changed but CHANGELOG not updated, warn
if [ $CODE_FILES -gt 0 ] && [ $CHANGELOG_UPDATED -eq 0 ]; then
  echo "⚠️  WARNING: Code files were changed but CHANGELOG.md was not updated"
  echo ""
  echo "Please add an entry to CHANGELOG.md under the [Unreleased] section"
  echo "describing your changes following the Keep a Changelog format:"
  echo "  https://keepachangelog.com/en/1.0.0/"
  echo ""
  echo "If this is a work-in-progress PR, you can update CHANGELOG.md later"
  echo "before merging."
  echo ""
  echo "To bypass this check for special cases (not recommended):"
  echo "  Add '[skip changelog]' to your PR title"
  echo ""

  # Check if PR title has [skip changelog]
  if [ -n "$PR_TITLE" ] && echo "$PR_TITLE" | grep -qi '\[skip changelog\]'; then
    echo "✅ [skip changelog] found in PR title, allowing bypass"
    exit 0
  fi

  # For now, this is a warning, not a hard failure
  # Uncomment the line below to make it a hard failure:
  # exit 1
  exit 0
fi

echo "✅ CHANGELOG.md check passed"
exit 0

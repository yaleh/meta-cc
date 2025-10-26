#!/bin/bash
set -e

# Usage: ./scripts/bump-plugin-version.sh [patch|minor|major]
#
# This script bumps the plugin version when .claude/ files change.
# It should be run ONLY when:
# - .claude/commands/*.md changes (e.g., /meta command logic)
# - .claude/agents/*.md changes (e.g., new/modified subagents)
#
# It should NOT be run when:
# - capabilities/ files change (content updates, not plugin API changes)
# - CLI/MCP code changes (separate versioning)

BUMP_TYPE=${1:-patch}  # Default to patch version

# Validate bump type
if [[ ! "$BUMP_TYPE" =~ ^(patch|minor|major)$ ]]; then
    echo "Error: Invalid bump type. Use: patch, minor, or major"
    echo "Usage: ./scripts/bump-plugin-version.sh [patch|minor|major]"
    exit 1
fi

# Check current branch
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$BRANCH" != "main" && "$BRANCH" != "develop" ]]; then
    echo "Error: Must be on main or develop branch (current: $BRANCH)"
    exit 1
fi

# Check working directory clean
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory not clean. Commit or stash changes first."
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    echo "Install with: sudo apt-get install jq (Ubuntu/Debian) or brew install jq (macOS)"
    exit 1
fi

# Get current version
CURRENT=$(jq -r '.version' .claude-plugin/plugin.json)
echo "Current plugin version: $CURRENT"

# Parse version components
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

# Bump version based on type
case $BUMP_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo "New plugin version: $NEW_VERSION"
echo ""

# Confirm with user
echo "This will update:"
echo "  - .claude-plugin/plugin.json: $CURRENT → $NEW_VERSION"
echo "  - .claude-plugin/marketplace.json: $CURRENT → $NEW_VERSION"
echo ""
echo "Press Enter to continue, or Ctrl+C to abort..."
read

# Update plugin.json version
echo "Updating plugin.json..."
jq --arg ver "$NEW_VERSION" '.version = $ver' .claude-plugin/plugin.json > .claude-plugin/plugin.json.tmp
mv .claude-plugin/plugin.json.tmp .claude-plugin/plugin.json
echo "✓ plugin.json updated to $NEW_VERSION"

# Update marketplace.json version
echo "Updating marketplace.json..."
jq --arg ver "$NEW_VERSION" '.plugins[0].version = $ver' .claude-plugin/marketplace.json > .claude-plugin/marketplace.json.tmp
mv .claude-plugin/marketplace.json.tmp .claude-plugin/marketplace.json
echo "✓ marketplace.json updated to $NEW_VERSION"
echo ""

# Commit changes
echo "Committing version bump..."
git add .claude-plugin/plugin.json .claude-plugin/marketplace.json
git commit -m "chore: bump plugin version to $NEW_VERSION

Updated plugin.json and marketplace.json version.

This version bump reflects changes to .claude/ plugin structure
(commands or agents), not capabilities content updates.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
echo "✓ Version bump committed"
echo ""

echo "=== Plugin Version Bumped ==="
echo ""
echo "Next steps:"
echo "  1. Review the commit: git show HEAD"
echo "  2. Push to remote: git push origin $BRANCH"
echo ""
echo "Note: This only updates the plugin version."
echo "To create a full release (CLI + MCP + Plugin), use:"
echo "  ./scripts/release.sh v$NEW_VERSION"

#!/bin/bash
# Automated version bumping script
#
# Purpose: Update marketplace.json to target version (used by pre-release-check.sh)
# Usage: ./scripts/bump-version.sh <version>
# Example: ./scripts/bump-version.sh v2.0.3
#
# This script only updates version files - it does NOT create git tags.
# Use ./scripts/release.sh for full release workflow.

set -e

VERSION=$1
VERSION_NUM=${VERSION#v}  # Remove 'v' prefix

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v2.0.3"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
    echo "Error: Invalid version format. Use v1.0.0 or v1.0.0-beta"
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    echo "Install with: sudo apt-get install jq (Ubuntu/Debian) or brew install jq (macOS)"
    exit 1
fi

# Check current branch
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$BRANCH" != "main" && "$BRANCH" != "develop" ]]; then
    echo "Warning: Not on main or develop branch (current: $BRANCH)"
    echo "Continue anyway? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Aborted"
        exit 1
    fi
fi

# Check working directory clean
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory not clean. Commit or stash changes first."
    exit 1
fi

echo "=== Version Bump: $VERSION ==="
echo ""

# Get current version
CURRENT=$(jq -r '.plugins[0].version' .claude-plugin/marketplace.json)
echo "Current version: v$CURRENT"
echo "Target version:  $VERSION ($VERSION_NUM)"
echo ""

# Confirm
echo "This will update marketplace.json: v$CURRENT → $VERSION"
echo "Press Enter to continue, or Ctrl+C to abort..."
read

# Update marketplace.json
echo "Updating marketplace.json..."
jq --arg ver "$VERSION_NUM" '.plugins[0].version = $ver' .claude-plugin/marketplace.json > .claude-plugin/marketplace.json.tmp
mv .claude-plugin/marketplace.json.tmp .claude-plugin/marketplace.json
echo "✓ marketplace.json updated to $VERSION_NUM"
echo ""

# Commit changes
echo "Committing version bump..."
git add .claude-plugin/marketplace.json
git commit -m "chore: bump version to $VERSION_NUM

Updated marketplace.json version.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
echo "✓ Version bump committed"
echo ""

echo "=== Version Bump Complete ==="
echo ""
echo "Next steps:"
echo "  1. Review commit: git show HEAD"
echo "  2. Push to remote: git push origin $BRANCH"
echo "  3. Run pre-release check: ./scripts/pre-release-check.sh $VERSION"
echo "  4. Create release: ./scripts/release.sh $VERSION"
echo ""

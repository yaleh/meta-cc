#!/bin/bash
set -e

# Usage: ./scripts/release.sh v1.0.0

VERSION=$1
VERSION_NUM=${VERSION#v}  # Remove 'v' prefix

if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: ./scripts/release.sh v1.0.0"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
    echo "Error: Invalid version format. Use v1.0.0 or v1.0.0-beta"
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
    echo "Error: Working directory not clean. Commit or stash changes."
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    echo "Install with: sudo apt-get install jq (Ubuntu/Debian) or brew install jq (macOS)"
    exit 1
fi

echo "=== Release $VERSION ==="
echo ""

# Run full test suite
echo "Running tests..."
make all
echo "✓ Tests passed"
echo ""

# Update plugin.json version
echo "Updating plugin.json version to $VERSION_NUM..."
jq --arg ver "$VERSION_NUM" '.version = $ver' .claude-plugin/plugin.json > .claude-plugin/plugin.json.tmp
mv .claude-plugin/plugin.json.tmp .claude-plugin/plugin.json
echo "✓ plugin.json updated"

# Update marketplace.json version
echo "Updating marketplace.json version to $VERSION_NUM..."
jq --arg ver "$VERSION_NUM" '.plugins[0].version = $ver' .claude-plugin/marketplace.json > .claude-plugin/marketplace.json.tmp
mv .claude-plugin/marketplace.json.tmp .claude-plugin/marketplace.json
echo "✓ marketplace.json updated"
echo ""

# Generate CHANGELOG entry automatically
echo "Generating CHANGELOG entry for $VERSION..."
bash scripts/generate-changelog-entry.sh "$VERSION"

if [ $? -ne 0 ]; then
    echo "Error: Failed to generate CHANGELOG entry"
    echo "Would you like to edit CHANGELOG.md manually? (y/N)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        echo "Please update CHANGELOG.md with release notes for $VERSION"
        echo "Press Enter when ready to continue, or Ctrl+C to abort..."
        read
    else
        echo "Aborted"
        exit 1
    fi
fi

echo "✓ CHANGELOG.md updated automatically"

# Commit version updates
echo "Committing version updates..."
git add .claude-plugin/plugin.json .claude-plugin/marketplace.json CHANGELOG.md
git commit -m "chore: release $VERSION

Update plugin.json, marketplace.json, and CHANGELOG.md to version $VERSION_NUM.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
echo "✓ Version updates committed"
echo ""

# Create tag
echo "Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
echo "✓ Tag created"
echo ""

# Push commits and tag
echo "Pushing to remote..."
git push origin "$BRANCH"
git push origin "$VERSION"
echo "✓ Pushed to remote"
echo ""

echo "=== Release $VERSION Complete ==="
echo ""
echo "GitHub Actions will now:"
echo "  1. Build cross-platform binaries"
echo "  2. Create GitHub Release"
echo "  3. Upload binaries"
echo ""
echo "Monitor progress: https://github.com/yaleh/meta-cc/actions"
